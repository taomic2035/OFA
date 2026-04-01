/**
 * @file offline.hpp
 * @brief Offline Support for OFA C++ SDK
 * @version 8.1.0
 * Sprint 30: Offline Capability Enhancement
 */

#ifndef OFA_OFFLINE_HPP
#define OFA_OFFLINE_HPP

#include <string>
#include <vector>
#include <map>
#include <queue>
#include <memory>
#include <functional>
#include <atomic>
#include <mutex>
#include <condition_variable>
#include <future>
#include <chrono>
#include <fstream>

#include "types.hpp"
#include "error.hpp"

namespace ofa {

/**
 * @brief 离线能力级别
 */
enum class OfflineLevel {
    L1_Offline,     // 完全离线模式
    L2_LAN,         // 局域网协作模式
    L3_WeakNet,     // 弱网同步模式
    L4_Online       // 在线模式
};

/**
 * @brief 本地任务状态
 */
enum class LocalTaskStatus {
    Pending,
    Running,
    Success,
    Failed,
    Cancelled,
    Retrying
};

/**
 * @brief 本地任务
 */
struct LocalTask {
    std::string taskId;
    std::string skillId;
    std::string operation;
    json input;
    json result;
    LocalTaskStatus status = LocalTaskStatus::Pending;
    std::string error;
    int retryCount = 0;
    int maxRetries = 3;
    int64_t createdAt;
    int64_t updatedAt;
    int64_t completedAt;
    bool syncPending = false;  // 需要同步到Center

    json toJson() const {
        json j;
        j["taskId"] = taskId;
        j["skillId"] = skillId;
        j["operation"] = operation;
        j["input"] = input;
        j["result"] = result;
        j["status"] = static_cast<int>(status);
        j["error"] = error;
        j["retryCount"] = retryCount;
        j["maxRetries"] = maxRetries;
        j["createdAt"] = createdAt;
        j["updatedAt"] = updatedAt;
        j["completedAt"] = completedAt;
        j["syncPending"] = syncPending;
        return j;
    }

    static LocalTask fromJson(const json& j) {
        LocalTask task;
        task.taskId = j["taskId"].get<std::string>();
        task.skillId = j["skillId"].get<std::string>();
        task.operation = j["operation"].get<std::string>();
        task.input = j["input"];
        task.result = j.value("result", json::object());
        task.status = static_cast<LocalTaskStatus>(j["status"].get<int>());
        task.error = j.value("error", "");
        task.retryCount = j.value("retryCount", 0);
        task.maxRetries = j.value("maxRetries", 3);
        task.createdAt = j["createdAt"].get<int64_t>();
        task.updatedAt = j["updatedAt"].get<int64_t>();
        task.completedAt = j.value("completedAt", 0);
        task.syncPending = j.value("syncPending", false);
        return task;
    }
};

/**
 * @brief 离线缓存数据
 */
struct CacheEntry {
    std::string key;
    json data;
    int64_t createdAt;
    int64_t expiresAt;
    bool syncPending;
    std::string source;  // 数据来源

    bool isExpired() const {
        if (expiresAt == 0) return false;
        auto now = std::chrono::duration_cast<std::chrono::seconds>(
            std::chrono::system_clock::now().time_since_epoch()
        ).count();
        return now > expiresAt;
    }
};

/**
 * @brief 本地调度器
 */
class LocalScheduler {
public:
    using TaskHandler = std::function<json(const std::string&, const std::string&, const json&)>;
    using TaskCallback = std::function<void(const LocalTask&)>;

    explicit LocalScheduler(size_t maxConcurrent = 4);
    ~LocalScheduler();

    /**
     * @brief 注册任务处理器
     */
    void registerHandler(const std::string& skillId, TaskHandler handler);

    /**
     * @brief 提交任务
     */
    std::string submit(const std::string& skillId,
                       const std::string& operation,
                       const json& input,
                       int maxRetries = 3);

    /**
     * @brief 取消任务
     */
    bool cancel(const std::string& taskId);

    /**
     * @brief 获取任务状态
     */
    LocalTask getTask(const std::string& taskId) const;

    /**
     * @brief 获取所有待处理任务
     */
    std::vector<LocalTask> getPendingTasks() const;

    /**
     * @brief 获取所有需要同步的任务
     */
    std::vector<LocalTask> getSyncPendingTasks() const;

    /**
     * @brief 设置任务完成回调
     */
    void onComplete(TaskCallback callback);

    /**
     * @brief 启动调度器
     */
    void start();

    /**
     * @brief 停止调度器
     */
    void stop();

    /**
     * @brief 获取统计
     */
    struct Stats {
        uint64_t totalTasks = 0;
        uint64_t successTasks = 0;
        uint64_t failedTasks = 0;
        uint64_t retryTasks = 0;
    };
    Stats stats() const;

private:
    void processLoop();
    void executeTask(LocalTask& task);
    void retryTask(LocalTask& task);

    size_t maxConcurrent_;
    std::map<std::string, TaskHandler> handlers_;
    std::queue<LocalTask> pendingQueue_;
    std::map<std::string, LocalTask> activeTasks_;
    std::map<std::string, LocalTask> completedTasks_;
    TaskCallback onComplete_;

    std::atomic<bool> running_{false};
    std::thread workerThread_;
    std::condition_variable cv_;
    mutable std::mutex mutex_;
    Stats stats_;
};

/**
 * @brief 离线缓存
 */
class OfflineCache {
public:
    explicit OfflineCache(const std::string& dbPath = "");
    ~OfflineCache();

    /**
     * @brief 存储数据
     */
    void put(const std::string& key,
             const json& data,
             int64_t ttlSeconds = 0,
             bool syncPending = false);

    /**
     * @brief 获取数据
     */
    json get(const std::string& key) const;

    /**
     * @brief 检查是否存在
     */
    bool exists(const std::string& key) const;

    /**
     * @brief 删除数据
     */
    void remove(const std::string& key);

    /**
     * @brief 清除过期数据
     */
    void clearExpired();

    /**
     * @brief 获取所有需要同步的数据
     */
    std::vector<CacheEntry> getSyncPending() const;

    /**
     * @brief 标记已同步
     */
    void markSynced(const std::string& key);

    /**
     * @brief 获取缓存大小
     */
    size_t size() const;

    /**
     * @brief 持久化到文件
     */
    void persist();

    /**
     * @brief 从文件加载
     */
    void load();

private:
    std::string dbPath_;
    std::map<std::string, CacheEntry> cache_;
    mutable std::mutex mutex_;
};

/**
 * @brief 离线管理器
 */
class OfflineManager {
public:
    explicit OfflineManager(const std::string& cachePath = "");
    ~OfflineManager();

    /**
     * @brief 设置离线级别
     */
    void setLevel(OfflineLevel level);
    OfflineLevel level() const { return level_; }

    /**
     * @brief 检查网络状态
     */
    bool isOnline() const;

    /**
     * @brief 自动切换离线模式
     */
    void autoSwitch();

    /**
     * @brief 获取调度器
     */
    LocalScheduler& scheduler() { return scheduler_; }

    /**
     * @brief 获取缓存
     */
    OfflineCache& cache() { return cache_; }

    /**
     * @brief 提交离线任务
     */
    std::string submitTask(const std::string& skillId,
                           const std::string& operation,
                           const json& input);

    /**
     * @brief 同步到Center
     */
    std::future<bool> sync();

    /**
     * @brief 注册离线技能处理器
     */
    void registerOfflineSkill(const std::string& skillId,
                              LocalScheduler::TaskHandler handler);

    /**
     * @brief 启动
     */
    void start();

    /**
     * @brief 停止
     */
    void stop();

private:
    OfflineLevel level_ = OfflineLevel::L4_Online;
    LocalScheduler scheduler_;
    OfflineCache cache_;
    std::atomic<bool> running_{false};
    std::thread monitorThread_;

    void networkMonitor();
};

} // namespace ofa

#endif // OFA_OFFLINE_HPP