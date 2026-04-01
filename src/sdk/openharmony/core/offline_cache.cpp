/**
 * @file offline_cache.cpp
 * @brief OFA 离线缓存实现 - 用于离线数据存储和同步
 */

#include "ofa/agent.h"
#include <string>
#include <vector>
#include <unordered_map>
#include <mutex>
#include <atomic>
#include <chrono>
#include <cstring>

namespace ofa {

// 缓存项
struct CacheItem {
    std::string key;
    std::vector<uint8_t> data;
    int64_t timestamp;
    int64_t expiry;
    bool synced;

    CacheItem() : timestamp(0), expiry(0), synced(false) {}
};

// 离线缓存实现
struct OfflineCacheImpl {
    size_t max_size = 10 * 1024 * 1024; // 10MB
    std::atomic<size_t> current_size{0};

    std::unordered_map<std::string, CacheItem> items;
    std::mutex mutex;

    // 待同步队列
    std::vector<std::string> pending_sync;
    std::mutex sync_mutex;

    // 统计
    std::atomic<int64_t> hits{0};
    std::atomic<int64_t> misses{0};

    OfflineCacheImpl() = default;

    bool put(const std::string& key, const uint8_t* data, size_t len, int64_t expiry_ms = 0) {
        std::lock_guard<std::mutex> lock(mutex);

        // 检查容量
        if (current_size + len > max_size) {
            // 清理过期或最旧的项
            evictIfNeeded(len);
        }

        CacheItem item;
        item.key = key;
        item.data.assign(data, data + len);
        item.timestamp = std::chrono::duration_cast<std::chrono::milliseconds>(
            std::chrono::system_clock::now().time_since_epoch()
        ).count();
        item.expiry = expiry_ms > 0 ? item.timestamp + expiry_ms : 0;
        item.synced = false;

        // 如果已存在，更新大小计算
        auto it = items.find(key);
        if (it != items.end()) {
            current_size -= it->second.data.size();
        }

        items[key] = item;
        current_size += len;

        // 添加到待同步队列
        {
            std::lock_guard<std::mutex> sync_lock(sync_mutex);
            if (std::find(pending_sync.begin(), pending_sync.end(), key) == pending_sync.end()) {
                pending_sync.push_back(key);
            }
        }

        return true;
    }

    bool get(const std::string& key, std::vector<uint8_t>& out_data) {
        std::lock_guard<std::mutex> lock(mutex);

        auto it = items.find(key);
        if (it == items.end()) {
            misses++;
            return false;
        }

        // 检查过期
        if (it->second.expiry > 0) {
            int64_t now = std::chrono::duration_cast<std::chrono::milliseconds>(
                std::chrono::system_clock::now().time_since_epoch()
            ).count();
            if (now > it->second.expiry) {
                current_size -= it->second.data.size();
                items.erase(it);
                misses++;
                return false;
            }
        }

        out_data = it->second.data;
        hits++;
        return true;
    }

    bool remove(const std::string& key) {
        std::lock_guard<std::mutex> lock(mutex);

        auto it = items.find(key);
        if (it == items.end()) {
            return false;
        }

        current_size -= it->second.data.size();
        items.erase(it);

        // 从待同步队列移除
        {
            std::lock_guard<std::mutex> sync_lock(sync_mutex);
            pending_sync.erase(
                std::remove(pending_sync.begin(), pending_sync.end(), key),
                pending_sync.end()
            );
        }

        return true;
    }

    void clear() {
        std::lock_guard<std::mutex> lock(mutex);
        items.clear();
        current_size = 0;

        std::lock_guard<std::mutex> sync_lock(sync_mutex);
        pending_sync.clear();
    }

    std::vector<std::string> getPendingSync() {
        std::lock_guard<std::mutex> sync_lock(sync_mutex);
        return pending_sync;
    }

    void markSynced(const std::string& key) {
        std::lock_guard<std::mutex> lock(mutex);

        auto it = items.find(key);
        if (it != items.end()) {
            it->second.synced = true;
        }

        std::lock_guard<std::mutex> sync_lock(sync_mutex);
        pending_sync.erase(
            std::remove(pending_sync.begin(), pending_sync.end(), key),
            pending_sync.end()
        );
    }

    size_t size() const {
        return current_size;
    }

    size_t count() {
        std::lock_guard<std::mutex> lock(mutex);
        return items.size();
    }

    double hitRate() const {
        int64_t total = hits + misses;
        if (total == 0) return 0.0;
        return static_cast<double>(hits) / total;
    }

private:
    void evictIfNeeded(size_t needed) {
        // 先清理过期项
        int64_t now = std::chrono::duration_cast<std::chrono::milliseconds>(
            std::chrono::system_clock::now().time_since_epoch()
        ).count();

        for (auto it = items.begin(); it != items.end(); ) {
            if (it->second.expiry > 0 && now > it->second.expiry) {
                current_size -= it->second.data.size();
                it = items.erase(it);
            } else {
                ++it;
            }
        }

        // 如果仍不够，清理最旧的已同步项
        while (current_size + needed > max_size && !items.empty()) {
            // 找最旧的已同步项
            auto oldest = items.begin();
            for (auto it = items.begin(); it != items.end(); ++it) {
                if (it->second.synced && it->second.timestamp < oldest->second.timestamp) {
                    oldest = it;
                }
            }

            if (oldest->second.synced) {
                current_size -= oldest->second.data.size();
                items.erase(oldest);
            } else {
                // 没有已同步项可清理，清理最旧的未同步项
                oldest = items.begin();
                for (auto it = items.begin(); it != items.end(); ++it) {
                    if (it->second.timestamp < oldest->second.timestamp) {
                        oldest = it;
                    }
                }
                current_size -= oldest->second.data.size();
                items.erase(oldest);
            }
        }
    }
};

} // namespace ofa

// C 接口实现

struct OFA_OfflineCache {
    ofa::OfflineCacheImpl impl;
};

OFA_OfflineCache* OFA_OfflineCache_Create(size_t max_size) {
    auto* cache = new OFA_OfflineCache();
    cache->impl.max_size = max_size > 0 ? max_size : 10 * 1024 * 1024;
    return cache;
}

void OFA_OfflineCache_Destroy(OFA_OfflineCache* cache) {
    if (cache) {
        delete cache;
    }
}

OFA_Result OFA_OfflineCache_Put(
    OFA_OfflineCache* cache,
    const char* key,
    const uint8_t* data,
    size_t data_len,
    int64_t expiry_ms
) {
    if (!cache || !key || !data) return OFA_ERROR;

    if (cache->impl.put(key, data, data_len, expiry_ms)) {
        return OFA_OK;
    }
    return OFA_ERROR;
}

OFA_Result OFA_OfflineCache_Get(
    OFA_OfflineCache* cache,
    const char* key,
    uint8_t** data,
    size_t* data_len
) {
    if (!cache || !key || !data || !data_len) return OFA_ERROR;

    std::vector<uint8_t> out;
    if (cache->impl.get(key, out)) {
        *data_len = out.size();
        *data = static_cast<uint8_t*>(malloc(out.size()));
        memcpy(*data, out.data(), out.size());
        return OFA_OK;
    }

    *data = nullptr;
    *data_len = 0;
    return OFA_ERROR;
}

OFA_Result OFA_OfflineCache_Remove(
    OFA_OfflineCache* cache,
    const char* key
) {
    if (!cache || !key) return OFA_ERROR;

    if (cache->impl.remove(key)) {
        return OFA_OK;
    }
    return OFA_ERROR;
}

void OFA_OfflineCache_Clear(OFA_OfflineCache* cache) {
    if (cache) {
        cache->impl.clear();
    }
}

size_t OFA_OfflineCache_Size(const OFA_OfflineCache* cache) {
    if (!cache) return 0;
    return cache->impl.size();
}

size_t OFA_OfflineCache_Count(const OFA_OfflineCache* cache) {
    if (!cache) return 0;
    return cache->impl.count();
}

size_t OFA_OfflineCache_PendingCount(const OFA_OfflineCache* cache) {
    if (!cache) return 0;
    return cache->impl.getPendingSync().size();
}

OFA_Result OFA_OfflineCache_GetPendingKeys(
    OFA_OfflineCache* cache,
    char*** keys,
    size_t* count
) {
    if (!cache || !keys || !count) return OFA_ERROR;

    auto pending = cache->impl.getPendingSync();
    *count = pending.size();

    if (pending.empty()) {
        *keys = nullptr;
        return OFA_OK;
    }

    *keys = static_cast<char**>(malloc(sizeof(char*) * pending.size()));
    for (size_t i = 0; i < pending.size(); i++) {
        (*keys)[i] = strdup(pending[i].c_str());
    }

    return OFA_OK;
}

OFA_Result OFA_OfflineCache_MarkSynced(
    OFA_OfflineCache* cache,
    const char* key
) {
    if (!cache || !key) return OFA_ERROR;

    cache->impl.markSynced(key);
    return OFA_OK;
}

double OFA_OfflineCache_HitRate(const OFA_OfflineCache* cache) {
    if (!cache) return 0.0;
    return cache->impl.hitRate();
}