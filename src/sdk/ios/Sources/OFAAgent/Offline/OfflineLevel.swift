import Foundation

/// 离线能力等级
public enum OfflineLevel: Int, Sendable, Codable {
    /// 不支持离线
    case none = 0
    /// 完全离线 (本地执行)
    case l1 = 1
    /// 局域网协作
    case l2 = 2
    /// 弱网同步
    case l3 = 3
    /// 在线模式
    case l4 = 4

    public var description: String {
        switch self {
        case .none: return "No offline support"
        case .l1: return "Full offline (local execution)"
        case .l2: return "LAN collaboration"
        case .l3: return "Weak network sync"
        case .l4: return "Online mode"
        }
    }
}

/// 任务状态
public enum TaskStatus: String, Sendable, Codable {
    case pending
    case running
    case completed
    case failed
    case cancelled
}

/// 离线任务
public struct LocalTask: Sendable, Codable {
    public let id: String
    public let skillId: String
    public let input: Data?
    public var output: Data?
    public var status: TaskStatus
    public var error: String?
    public let createdAt: Date
    public var completedAt: Date?
    public var retryCount: Int
    public let maxRetries: Int
    public var syncPending: Bool

    public init(skillId: String, input: Data? = nil) {
        self.id = "local-" + UUID().uuidString.prefix(8)
        self.skillId = skillId
        self.input = input
        self.output = nil
        self.status = .pending
        self.error = nil
        self.createdAt = Date()
        self.completedAt = nil
        self.retryCount = 0
        self.maxRetries = 3
        self.syncPending = true
    }

    public var duration: TimeInterval {
        if let completedAt = completedAt {
            return completedAt.timeIntervalSince(createdAt)
        }
        return Date().timeIntervalSince(createdAt)
    }

    public var canRetry: Bool {
        retryCount < maxRetries
    }
}