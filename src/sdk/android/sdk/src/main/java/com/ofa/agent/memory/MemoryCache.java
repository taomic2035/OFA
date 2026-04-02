package com.ofa.agent.memory;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.ArrayList;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Memory Cache - L1内存缓存层
 * 使用LRU策略缓存热数据
 */
public class MemoryCache {

    private static final int DEFAULT_MAX_SIZE = 100;
    private static final int MAX_RECENT_KEYS = 50;

    // LRU缓存
    private final LinkedHashMap<String, CacheEntry> cache;
    private final int maxSize;

    // 最近访问的keys（用于快速判断热点）
    private final LinkedHashMap<String, Boolean> recentKeys;

    /**
     * 缓存条目
     */
    private static class CacheEntry {
        final String key;
        final String value;
        final MemoryEntry entry;
        long lastAccess;

        CacheEntry(@NonNull String key, @NonNull String value, @NonNull MemoryEntry entry) {
            this.key = key;
            this.value = value;
            this.entry = entry;
            this.lastAccess = System.currentTimeMillis();
        }

        void touch() {
            this.lastAccess = System.currentTimeMillis();
        }
    }

    public MemoryCache() {
        this(DEFAULT_MAX_SIZE);
    }

    public MemoryCache(int maxSize) {
        this.maxSize = maxSize;
        this.cache = new LinkedHashMap<>(maxSize, 0.75f, true); // accessOrder=true
        this.recentKeys = new LinkedHashMap<>(MAX_RECENT_KEYS, 0.75f, true);
    }

    // ===== 存取操作 =====

    /**
     * 存入缓存
     */
    public void put(@NonNull String key, @NonNull MemoryEntry entry) {
        String compositeKey = makeCompositeKey(key, entry.getValue());
        synchronized (cache) {
            cache.put(compositeKey, new CacheEntry(key, entry.getValue(), entry));
            recentKeys.put(key, true);
            trimIfNeeded();
        }
    }

    /**
     * 获取推荐值（最高分）
     */
    @Nullable
    public String getRecommendedValue(@NonNull String key) {
        synchronized (cache) {
            String topKey = null;
            float topScore = -1;

            for (Map.Entry<String, CacheEntry> e : cache.entrySet()) {
                CacheEntry entry = e.getValue();
                if (entry.key.equals(key)) {
                    float score = entry.entry.calculateRecommendationScore();
                    if (score > topScore) {
                        topScore = score;
                        topKey = entry.value;
                    }
                }
            }
            return topKey;
        }
    }

    /**
     * 获取推荐列表
     */
    @NonNull
    public List<String> getRecommendedValues(@NonNull String key, int limit) {
        List<MemoryEntry> entries = getEntries(key);
        if (entries.isEmpty()) return Collections.emptyList();

        // 按分数排序
        entries.sort((a, b) -> Float.compare(
                b.calculateRecommendationScore(),
                a.calculateRecommendationScore()));

        List<String> values = new ArrayList<>();
        for (int i = 0; i < Math.min(limit, entries.size()); i++) {
            values.add(entries.get(i).getValue());
        }
        return values;
    }

    /**
     * 获取所有条目
     */
    @NonNull
    public List<MemoryEntry> getEntries(@NonNull String key) {
        List<MemoryEntry> entries = new ArrayList<>();
        synchronized (cache) {
            for (Map.Entry<String, CacheEntry> e : cache.entrySet()) {
                CacheEntry entry = e.getValue();
                if (entry.key.equals(key)) {
                    entry.touch();
                    entries.add(entry.entry);
                }
            }
        }
        return entries;
    }

    /**
     * 获取单个条目
     */
    @Nullable
    public MemoryEntry getEntry(@NonNull String key, @NonNull String value) {
        String compositeKey = makeCompositeKey(key, value);
        synchronized (cache) {
            CacheEntry entry = cache.get(compositeKey);
            if (entry != null) {
                entry.touch();
                return entry.entry;
            }
        }
        return null;
    }

    /**
     * 获取上次使用的值
     */
    @Nullable
    public String getLastUsedValue(@NonNull String key) {
        String lastValue = null;
        long lastTime = 0;

        synchronized (cache) {
            for (Map.Entry<String, CacheEntry> e : cache.entrySet()) {
                CacheEntry entry = e.getValue();
                if (entry.key.equals(key) && entry.entry.getTimestamp() > lastTime) {
                    lastTime = entry.entry.getTimestamp();
                    lastValue = entry.value;
                }
            }
        }
        return lastValue;
    }

    // ===== 缓存管理 =====

    /**
     * 检查是否命中缓存
     */
    public boolean hasKey(@NonNull String key) {
        synchronized (cache) {
            for (CacheEntry entry : cache.values()) {
                if (entry.key.equals(key)) {
                    return true;
                }
            }
        }
        return false;
    }

    /**
     * 检查是否是热点key
     */
    public boolean isHotKey(@NonNull String key) {
        synchronized (recentKeys) {
            return recentKeys.containsKey(key);
        }
    }

    /**
     * 删除指定条目
     */
    public void remove(@NonNull String key, @NonNull String value) {
        String compositeKey = makeCompositeKey(key, value);
        synchronized (cache) {
            cache.remove(compositeKey);
        }
    }

    /**
     * 删除指定key的所有条目
     */
    public void removeAll(@NonNull String key) {
        synchronized (cache) {
            cache.entrySet().removeIf(e -> e.getValue().key.equals(key));
            recentKeys.remove(key);
        }
    }

    /**
     * 清空缓存
     */
    public void clear() {
        synchronized (cache) {
            cache.clear();
            recentKeys.clear();
        }
    }

    /**
     * 获取缓存大小
     */
    public int size() {
        return cache.size();
    }

    // ===== 内部方法 =====

    private String makeCompositeKey(@NonNull String key, @NonNull String value) {
        return key + "||" + value;
    }

    private void trimIfNeeded() {
        while (cache.size() > maxSize) {
            // LinkedHashMap的迭代顺序就是LRU顺序
            String firstKey = cache.keySet().iterator().next();
            cache.remove(firstKey);
        }

        while (recentKeys.size() > MAX_RECENT_KEYS) {
            String firstKey = recentKeys.keySet().iterator().next();
            recentKeys.remove(firstKey);
        }
    }

    /**
     * 获取缓存统计
     */
    @NonNull
    public CacheStats getStats() {
        synchronized (cache) {
            return new CacheStats(cache.size(), recentKeys.size());
        }
    }

    public static class CacheStats {
        public final int cacheSize;
        public final int recentKeyCount;

        public CacheStats(int cacheSize, int recentKeyCount) {
            this.cacheSize = cacheSize;
            this.recentKeyCount = recentKeyCount;
        }
    }
}