import Foundation

/// 本地任务调度器
public actor LocalScheduler {
    private let workerCount: Int
    private let offlineLevel: OfflineLevel
    private var skills: [String: any SkillExecutor] = [:]
    private var skillMetadata: [String: SkillMetadata] = [:]
    private var tasks: [String: LocalTask] = [:]
    private var taskQueue: [LocalTask] = []
    private var running = false
    private var pendingCount = 0
    private var completedCount = 0

    public init(workerCount: Int = 4, offlineLevel: OfflineLevel = .l1) {
        self.workerCount = workerCount
        self.offlineLevel = offlineLevel
    }

    /// 启动调度器
    public func start() {
        guard !running else { return }
        running = true
        print("Local scheduler started with \(workerCount) workers, level: \(offlineLevel.rawValue)")

        // 启动工作任务
        for _ in 0..<workerCount {
            Task {
                await workerLoop()
            }
        }
    }

    /// 停止调度器
    public func stop() {
        running = false
        print("Local scheduler stopped")
    }

    private func workerLoop() async {
        while running {
            if let task = taskQueue.first {
                taskQueue.removeFirst()
                await executeTask(task)
            } else {
                try? await Task.sleep(nanoseconds: 100_000_000) // 100ms
            }
        }
    }

    private func executeTask(_ task: LocalTask) async {
        var updatedTask = task
        updatedTask.status = .running
        tasks[task.id] = updatedTask

        do {
            guard let executor = skills[task.skillId] else {
                throw OFAError.skillNotFound(task.skillId)
            }

            // 检查是否支持离线
            if let meta = skillMetadata[task.skillId], !meta.offlineCapable {
                throw OFAError.executionFailed("Skill does not support offline execution")
            }

            let output = try await executor.execute(updatedTask.input ?? Data())

            updatedTask.output = output
            updatedTask.status = .completed
            updatedTask.completedAt = Date()
            updatedTask.syncPending = offlineLevel != .l4

            tasks[task.id] = updatedTask
            completedCount += 1
            pendingCount -= 1

            await notifyTaskCompleted(task.id, output)

        } catch {
            updatedTask.status = .failed
            updatedTask.error = error.localizedDescription

            if updatedTask.canRetry {
                updatedTask.retryCount += 1
                updatedTask.status = .pending
                taskQueue.append(updatedTask)
                tasks[task.id] = updatedTask
                print("Task \(task.id) retry \(updatedTask.retryCount)")
            } else {
                tasks[task.id] = updatedTask
                pendingCount -= 1
                await notifyTaskFailed(task.id, error.localizedDescription)
            }
        }
    }

    /// 注册技能
    public func registerSkill(
        _ skill: any SkillExecutor,
        offlineCapable: Bool = true,
        category: String = "general"
    ) {
        skills[skill.skillId] = skill
        skillMetadata[skill.skillId] = SkillMetadata(
            offlineCapable: offlineCapable,
            category: category
        )
        print("Registered local skill: \(skill.skillId) (offline: \(offlineCapable))")
    }

    /// 注销技能
    public func unregisterSkill(_ skillId: String) {
        skills.removeValue(forKey: skillId)
        skillMetadata.removeValue(forKey: skillId)
    }

    /// 提交任务
    public func submitTask(skillId: String, input: Data? = nil) -> String {
        let task = LocalTask(skillId: skillId, input: input)
        tasks[task.id] = task
        taskQueue.append(task)
        pendingCount += 1

        print("Task submitted: \(task.id) -> \(skillId)")
        return task.id
    }

    /// 获取任务
    public func getTask(_ taskId: String) -> LocalTask? {
        tasks[taskId]
    }

    /// 取消任务
    public func cancelTask(_ taskId: String) -> Bool {
        guard var task = tasks[taskId], task.status == .pending else {
            return false
        }
        task.status = .cancelled
        tasks[taskId] = task
        pendingCount -= 1
        return true
    }

    /// 列出待处理任务
    public func listPendingTasks() -> [String] {
        tasks.filter { $0.value.status == .pending }.map { $0.key }
    }

    /// 列出已注册技能
    public func listSkills() -> [String] {
        Array(skills.keys)
    }

    /// 获取待处理任务数
    public func getPendingCount() -> Int { pendingCount }

    /// 获取已完成任务数
    public func getCompletedCount() -> Int { completedCount }

    /// 获取离线等级
    public func getOfflineLevel() -> OfflineLevel { offlineLevel }

    // MARK: - Notifications

    private var taskCompletionHandlers: [(String, Data?) -> Void] = []
    private var taskFailureHandlers: [(String, String) -> Void] = []

    public func onTaskCompleted(_ handler: @escaping (String, Data?) -> Void) {
        taskCompletionHandlers.append(handler)
    }

    public func onTaskFailed(_ handler: @escaping (String, String) -> Void) {
        taskFailureHandlers.append(handler)
    }

    private func notifyTaskCompleted(_ taskId: String, _ output: Data?) {
        for handler in taskCompletionHandlers {
            handler(taskId, output)
        }
    }

    private func notifyTaskFailed(_ taskId: String, _ error: String) {
        for handler in taskFailureHandlers {
            handler(taskId, error)
        }
    }
}

/// 技能元数据
private struct SkillMetadata {
    let offlineCapable: Bool
    let category: String
}