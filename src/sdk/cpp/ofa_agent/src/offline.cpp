/**
 * @file offline.cpp
 * @brief Offline Support Implementation
 * Sprint 30: Offline Capability Enhancement
 */

#include "ofa/offline.hpp"
#include <chrono>
#include <thread>
#include <random>
#include <sstream>
#include <fstream>
#include <algorithm>

namespace ofa {

// Helper: generate UUID
static std::string generateId() {
    std::random_device rd;
    std::mt19937 gen(rd());
    std::uniform_int_distribution<> dis(0, 15);

    std::stringstream ss;
    ss << std::hex;
    for (int i = 0; i < 8; i++) ss << dis(gen);
    ss << "-";
    for (int i = 0; i < 4; i++) ss << dis(gen);
    return ss.str();
}

// Helper: current timestamp
static int64_t nowMs() {
    return std::chrono::duration_cast<std::chrono::milliseconds>(
        std::chrono::system_clock::now().time_since_epoch()
    ).count();
}

// ==================== LocalScheduler ====================

LocalScheduler::LocalScheduler(size_t maxConcurrent)
    : maxConcurrent_(maxConcurrent) {
}

LocalScheduler::~LocalScheduler() {
    stop();
}

void LocalScheduler::registerHandler(const std::string& skillId, TaskHandler handler) {
    std::lock_guard<std::mutex> lock(mutex_);
    handlers_[skillId] = handler;
}

std::string LocalScheduler::submit(const std::string& skillId,
                                    const std::string& operation,
                                    const json& input,
                                    int maxRetries) {
    std::lock_guard<std::mutex> lock(mutex_);

    LocalTask task;
    task.taskId = "local-" + generateId();
    task.skillId = skillId;
    task.operation = operation;
    task.input = input;
    task.status = LocalTaskStatus::Pending;
    task.maxRetries = maxRetries;
    task.createdAt = nowMs();
    task.updatedAt = nowMs();
    task.syncPending = true;

    pendingQueue_.push(task);
    stats_.totalTasks++;

    cv_.notify_one();
    return task.taskId;
}

bool LocalScheduler::cancel(const std::string& taskId) {
    std::lock_guard<std::mutex> lock(mutex_);

    // Check pending queue
    std::queue<LocalTask> newQueue;
    bool found = false;
    while (!pendingQueue_.empty()) {
        auto task = pendingQueue_.front();
        pendingQueue_.pop();
        if (task.taskId == taskId) {
            task.status = LocalTaskStatus::Cancelled;
            task.updatedAt = nowMs();
            completedTasks_[taskId] = task;
            found = true;
        } else {
            newQueue.push(task);
        }
    }
    pendingQueue_ = newQueue;

    // Check active tasks
    auto it = activeTasks_.find(taskId);
    if (it != activeTasks_.end()) {
        it->second.status = LocalTaskStatus::Cancelled;
        return true;
    }

    return found;
}

LocalTask LocalScheduler::getTask(const std::string& taskId) const {
    std::lock_guard<std::mutex> lock(mutex_);

    auto it = activeTasks_.find(taskId);
    if (it != activeTasks_.end()) {
        return it->second;
    }

    auto it2 = completedTasks_.find(taskId);
    if (it2 != completedTasks_.end()) {
        return it2->second;
    }

    throw OfAException("Task not found: " + taskId);
}

std::vector<LocalTask> LocalScheduler::getPendingTasks() const {
    std::lock_guard<std::mutex> lock(mutex_);

    std::vector<LocalTask> tasks;
    std::queue<LocalTask> temp = pendingQueue_;
    while (!temp.empty()) {
        tasks.push_back(temp.front());
        temp.pop();
    }
    return tasks;
}

std::vector<LocalTask> LocalScheduler::getSyncPendingTasks() const {
    std::lock_guard<std::mutex> lock(mutex_);

    std::vector<LocalTask> tasks;
    for (const auto& pair : completedTasks_) {
        if (pair.second.syncPending) {
            tasks.push_back(pair.second);
        }
    }
    return tasks;
}

void LocalScheduler::onComplete(TaskCallback callback) {
    onComplete_ = callback;
}

void LocalScheduler::start() {
    if (running_) return;
    running_ = true;
    workerThread_ = std::thread(&LocalScheduler::processLoop, this);
}

void LocalScheduler::stop() {
    if (!running_) return;
    running_ = false;
    cv_.notify_all();
    if (workerThread_.joinable()) {
        workerThread_.join();
    }
}

LocalScheduler::Stats LocalScheduler::stats() const {
    std::lock_guard<std::mutex> lock(mutex_);
    return stats_;
}

void LocalScheduler::processLoop() {
    while (running_) {
        LocalTask task;
        {
            std::unique_lock<std::mutex> lock(mutex_);
            cv_.wait_for(lock, std::chrono::milliseconds(100), [this]() {
                return !pendingQueue_.empty() || !running_;
            });

            if (!running_) break;

            if (pendingQueue_.empty()) continue;

            // Check concurrency limit
            if (activeTasks_.size() >= maxConcurrent_) continue;

            task = pendingQueue_.front();
            pendingQueue_.pop();
            activeTasks_[task.taskId] = task;
        }

        executeTask(task);
    }
}

void LocalScheduler::executeTask(LocalTask& task) {
    task.status = LocalTaskStatus::Running;
    task.updatedAt = nowMs();

    try {
        auto it = handlers_.find(task.skillId);
        if (it == handlers_.end()) {
            throw OfAException("No handler for skill: " + task.skillId);
        }

        task.result = it->second(task.skillId, task.operation, task.input);
        task.status = LocalTaskStatus::Success;
        task.completedAt = nowMs();
        stats_.successTasks++;

        {
            std::lock_guard<std::mutex> lock(mutex_);
            activeTasks_.erase(task.taskId);
            completedTasks_[task.taskId] = task;
        }

        if (onComplete_) {
            onComplete_(task);
        }

    } catch (const std::exception& e) {
        task.error = e.what();
        task.updatedAt = nowMs();

        if (task.retryCount < task.maxRetries) {
            retryTask(task);
        } else {
            task.status = LocalTaskStatus::Failed;
            task.completedAt = nowMs();
            stats_.failedTasks++;

            {
                std::lock_guard<std::mutex> lock(mutex_);
                activeTasks_.erase(task.taskId);
                completedTasks_[task.taskId] = task;
            }

            if (onComplete_) {
                onComplete_(task);
            }
        }
    }

    cv_.notify_one();
}

void LocalScheduler::retryTask(LocalTask& task) {
    task.status = LocalTaskStatus::Retrying;
    task.retryCount++;
    stats_.retryTasks++;
    task.updatedAt = nowMs();

    // Add back to queue with delay
    std::this_thread::sleep_for(std::chrono::seconds(1));

    {
        std::lock_guard<std::mutex> lock(mutex_);
        activeTasks_.erase(task.taskId);
        pendingQueue_.push(task);
    }

    cv_.notify_one();
}

// ==================== OfflineCache ====================

OfflineCache::OfflineCache(const std::string& dbPath)
    : dbPath_(dbPath) {
    if (!dbPath_.empty()) {
        load();
    }
}

OfflineCache::~OfflineCache() {
    if (!dbPath_.empty()) {
        persist();
    }
}

void OfflineCache::put(const std::string& key,
                        const json& data,
                        int64_t ttlSeconds,
                        bool syncPending) {
    std::lock_guard<std::mutex> lock(mutex_);

    CacheEntry entry;
    entry.key = key;
    entry.data = data;
    entry.createdAt = nowMs();
    entry.expiresAt = ttlSeconds > 0 ? entry.createdAt + ttlSeconds * 1000 : 0;
    entry.syncPending = syncPending;
    entry.source = "local";

    cache_[key] = entry;
}

json OfflineCache::get(const std::string& key) const {
    std::lock_guard<std::mutex> lock(mutex_);

    auto it = cache_.find(key);
    if (it == cache_.end()) {
        return json::object();
    }

    if (it->second.isExpired()) {
        return json::object();
    }

    return it->second.data;
}

bool OfflineCache::exists(const std::string& key) const {
    std::lock_guard<std::mutex> lock(mutex_);

    auto it = cache_.find(key);
    if (it == cache_.end()) return false;

    return !it->second.isExpired();
}

void OfflineCache::remove(const std::string& key) {
    std::lock_guard<std::mutex> lock(mutex_);
    cache_.erase(key);
}

void OfflineCache::clearExpired() {
    std::lock_guard<std::mutex> lock(mutex_);

    auto now = std::chrono::duration_cast<std::chrono::seconds>(
        std::chrono::system_clock::now().time_since_epoch()
    ).count();

    for (auto it = cache_.begin(); it != cache_.end(); ) {
        if (it->second.isExpired()) {
            it = cache_.erase(it);
        } else {
            ++it;
        }
    }
}

std::vector<CacheEntry> OfflineCache::getSyncPending() const {
    std::lock_guard<std::mutex> lock(mutex_);

    std::vector<CacheEntry> entries;
    for (const auto& pair : cache_) {
        if (pair.second.syncPending && !pair.second.isExpired()) {
            entries.push_back(pair.second);
        }
    }
    return entries;
}

void OfflineCache::markSynced(const std::string& key) {
    std::lock_guard<std::mutex> lock(mutex_);

    auto it = cache_.find(key);
    if (it != cache_.end()) {
        it->second.syncPending = false;
    }
}

size_t OfflineCache::size() const {
    std::lock_guard<std::mutex> lock(mutex_);
    return cache_.size();
}

void OfflineCache::persist() {
    if (dbPath_.empty()) return;

    std::lock_guard<std::mutex> lock(mutex_);

    json j = json::array();
    for (const auto& pair : cache_) {
        json entry;
        entry["key"] = pair.second.key;
        entry["data"] = pair.second.data;
        entry["createdAt"] = pair.second.createdAt;
        entry["expiresAt"] = pair.second.expiresAt;
        entry["syncPending"] = pair.second.syncPending;
        entry["source"] = pair.second.source;
        j.push_back(entry);
    }

    std::ofstream file(dbPath_);
    if (file.is_open()) {
        file << j.dump(2);
        file.close();
    }
}

void OfflineCache::load() {
    if (dbPath_.empty()) return;

    std::lock_guard<std::mutex> lock(mutex_);

    std::ifstream file(dbPath_);
    if (!file.is_open()) return;

    try {
        json j;
        file >> j;
        file.close();

        for (const auto& entry : j) {
            CacheEntry ce;
            ce.key = entry["key"].get<std::string>();
            ce.data = entry["data"];
            ce.createdAt = entry["createdAt"].get<int64_t>();
            ce.expiresAt = entry["expiresAt"].get<int64_t>();
            ce.syncPending = entry["syncPending"].get<bool>();
            ce.source = entry["source"].get<std::string>();

            if (!ce.isExpired()) {
                cache_[ce.key] = ce;
            }
        }
    } catch (...) {
        // Invalid JSON, ignore
    }
}

// ==================== OfflineManager ====================

OfflineManager::OfflineManager(const std::string& cachePath)
    : cache_(cachePath) {
}

OfflineManager::~OfflineManager() {
    stop();
}

void OfflineManager::setLevel(OfflineLevel level) {
    level_ = level;
}

bool OfflineManager::isOnline() const {
    // Simple check: try to resolve a known host
    // In production, implement proper network detection
    return level_ == OfflineLevel::L4_Online;
}

void OfflineManager::autoSwitch() {
    // Auto-detect network and switch mode
    // For now, keep current level
    // In production: ping center, check connectivity
}

LocalScheduler::TaskHandler OfflineManager::submitTask(const std::string& skillId,
                                                         const std::string& operation,
                                                         const json& input) {
    // Return empty handler, actual submit happens in scheduler
    std::string taskId = scheduler_.submit(skillId, operation, input);

    // Cache the task input
    cache_.put("task:" + taskId, input, 0, true);

    return [](const std::string&, const std::string&, const json&) { return json::object(); };
}

std::future<bool> OfflineManager::sync() {
    return std::async(std::launch::async, [this]() {
        // Sync pending tasks and cache to center
        auto pendingTasks = scheduler_.getSyncPendingTasks();
        auto pendingCache = cache_.getSyncPending();

        // In production: send to center via gRPC/REST
        // For now, just mark as synced

        for (const auto& task : pendingTasks) {
            // Would send to center here
            // Mark as synced
        }

        for (const auto& entry : pendingCache) {
            cache_.markSynced(entry.key);
        }

        return true;
    });
}

void OfflineManager::registerOfflineSkill(const std::string& skillId,
                                           LocalScheduler::TaskHandler handler) {
    scheduler_.registerHandler(skillId, handler);
}

void OfflineManager::start() {
    if (running_) return;
    running_ = true;
    scheduler_.start();
    monitorThread_ = std::thread(&OfflineManager::networkMonitor, this);
}

void OfflineManager::stop() {
    if (!running_) return;
    running_ = false;
    scheduler_.stop();
    if (monitorThread_.joinable()) {
        monitorThread_.join();
    }
    cache_.persist();
}

void OfflineManager::networkMonitor() {
    while (running_) {
        std::this_thread::sleep_for(std::chrono::seconds(5));

        autoSwitch();

        // If online, try to sync
        if (level_ == OfflineLevel::L4_Online) {
            sync().get();
        }
    }
}

} // namespace ofa