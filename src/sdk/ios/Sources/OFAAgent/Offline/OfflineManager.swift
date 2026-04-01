import Foundation

/// 离线管理器
public actor OfflineManager {
    private let level: OfflineLevel
    public let scheduler: LocalScheduler
    public let cache: OfflineCache

    private var offlineMode: Bool
    private var syncCallback: ((String, Data) async throws -> Void)?

    public init(level: OfflineLevel = .l1) {
        self.level = level
        self.scheduler = LocalScheduler(workerCount: 4, offlineLevel: level)
        self.cache = OfflineCache()
        self.offlineMode = level == .l1
    }

    /// 启动离线管理器
    public func start() async {
        await scheduler.start()
        print("Offline manager started at level \(level.rawValue)")
    }

    /// 停止离线管理器
    public func stop() async {
        await scheduler.stop()
        await cache.clear()
        print("Offline manager stopped")
    }

    /// 设置离线模式
    public func setOfflineMode(_ offline: Bool) {
        offlineMode = offline
        print("Offline mode: \(offline)")
    }

    /// 是否处于离线模式
    public func isOffline() -> Bool {
        offlineMode
    }

    /// 获取离线等级
    public func getLevel() -> OfflineLevel {
        level
    }

    /// 注册技能
    public func registerSkill(_ skill: any SkillExecutor, offlineCapable: Bool = true) async {
        await scheduler.registerSkill(skill, offlineCapable: offlineCapable)
    }

    /// 本地执行任务
    public func executeLocal(skillId: String, input: Data? = nil) async -> String {
        await scheduler.submitTask(skillId: skillId, input: input)
    }

    /// 同步执行
    public func executeSync(skillId: String, input: Data? = nil, timeout: TimeInterval = 30) async throws -> Data {
        let taskId = await scheduler.submitTask(skillId: skillId, input: input)

        // 等待完成
        let startTime = Date()
        while true {
            if let task = await scheduler.getTask(taskId) {
                switch task.status {
                case .completed:
                    return task.output ?? Data()
                case .failed:
                    throw OFAError.executionFailed(task.error ?? "Unknown error")
                case .cancelled:
                    throw OFAError.executionFailed("Task cancelled")
                default:
                    break
                }
            }

            if Date().timeIntervalSince(startTime) > timeout {
                throw OFAError.executionFailed("Timeout")
            }

            try await Task.sleep(nanoseconds: 100_000_000) // 100ms
        }
    }

    /// 缓存数据
    public func cacheData(_ key: String, data: Data, expiryMs: Int = 0) async {
        await cache.put(key, data: data, expiryMs: expiryMs)
    }

    /// 获取缓存数据
    public func getCachedData(_ key: String) async -> Data? {
        await cache.get(key)
    }

    /// 获取待同步键列表
    public func getPendingSyncKeys() async -> [String] {
        await cache.getPendingKeys()
    }

    /// 立即同步
    public func syncNow() async -> Bool {
        let pending = await cache.getPendingKeys()
        if pending.isEmpty { return true }

        guard let callback = syncCallback else {
            print("No sync callback configured")
            return false
        }

        for key in pending {
            if let data = await cache.get(key) {
                do {
                    try await callback(key, data)
                    await cache.markSynced(key)
                } catch {
                    print("Sync failed for \(key): \(error)")
                    return false
                }
            }
        }

        return true
    }

    /// 设置同步回调
    public func setSyncCallback(_ callback: @escaping (String, Data) async throws -> Void) {
        syncCallback = callback
    }

    /// 获取任务
    public func getTask(_ taskId: String) async -> LocalTask? {
        await scheduler.getTask(taskId)
    }

    /// 获取统计信息
    public func getStats() async -> OfflineStats {
        OfflineStats(
            offlineMode: offlineMode,
            level: level.rawValue,
            pendingTasks: await scheduler.getPendingCount(),
            completedTasks: await scheduler.getCompletedCount(),
            pendingSync: await cache.getPendingCount(),
            cacheHitRate: await cache.hitRate()
        )
    }
}

/// 离线统计信息
public struct OfflineStats: Sendable {
    public let offlineMode: Bool
    public let level: Int
    public let pendingTasks: Int
    public let completedTasks: Int
    public let pendingSync: Int
    public let cacheHitRate: Double
}