// OFAiOSAgent.swift
// OFA iOS SDK - Agent Implementation for iPhone/iPad/Mac (v9.8.0)
// Supports multi-device Apple ecosystem integration
// Aligned with Android SDK architecture

import Foundation
import Combine

/// Device type for Apple devices
public enum AppleDeviceType: String, Codable {
    case iPhone = "iphone"
    case iPad = "ipad"
    case mac = "mac"
    case watch = "watch"
    case unknown = "unknown"
}

/// Agent status
public enum AgentStatus: String, Codable {
    case online = "online"
    case offline = "offline"
    case busy = "busy"
    case error = "error"
}

/// Agent profile representing device capabilities
public struct AgentProfile: Codable {
    public let agentId: String
    public let deviceType: AppleDeviceType
    public let deviceName: String
    public let osVersion: String
    public let capabilities: [String]
    public let identityId: String?
    public var status: AgentStatus
    public var lastHeartbeat: Date

    public init(
        agentId: String,
        deviceType: AppleDeviceType,
        deviceName: String,
        osVersion: String = "",
        capabilities: [String] = [],
        identityId: String? = nil
    ) {
        self.agentId = agentId
        self.deviceType = deviceType
        self.deviceName = deviceName
        self.osVersion = osVersion
        self.capabilities = capabilities
        self.identityId = identityId
        self.status = .offline
        self.lastHeartbeat = Date()
    }
}

/// Agent mode for operation
public enum AgentMode: String, Codable {
    case standalone = "standalone"  // Fully independent
    case sync = "sync"              // Sync with Center periodically
}

/// OFA iOS Agent - Main entry point for iPhone/iPad/Mac (v9.8.0)
public class OFAiOSAgent: ObservableObject {

    // MARK: - Published Properties

    @Published public var status: AgentStatus = .offline
    @Published public var connectedToCenter: Bool = false
    @Published public var currentIdentityId: String?
    @Published public var currentScene: SceneType = .unknown
    @Published public var errorCount: Int = 0

    // MARK: - Core Components (v9.8.0 - Aligned with Android)

    public let profile: AgentProfile
    private let modeManager: AgentModeManager
    private let centerConnection: CenterConnection
    private let identityManager: IdentityManager
    private let sceneDetector: SceneDetector
    private let audioPlayer: AudioPlayer

    // MARK: - New Components (v9.8.0)

    private let errorHandler: ErrorHandler
    private let connectionRecovery: ConnectionRecoveryManager
    private let distributedOrchestrator: DistributedOrchestrator
    private let memoryManager: UserMemoryManager

    // MARK: - Configuration

    private let config: AgentConfig
    private var cancellables = Set<AnyCancellable>()

    // MARK: - Initialization

    public init(config: AgentConfig) {
        self.config = config

        // Determine device type
        let deviceType = Self.detectDeviceType()

        // Create profile
        self.profile = AgentProfile(
            agentId: config.agentId ?? Self.generateAgentId(),
            deviceType: deviceType,
            deviceName: Self.getDeviceName(),
            osVersion: Self.getOSVersion(),
            capabilities: Self.getCapabilities(for: deviceType),
            identityId: nil
        )

        // Initialize core components
        self.modeManager = AgentModeManager(mode: config.mode)
        self.centerConnection = CenterConnection(centerAddress: config.centerAddress)
        self.identityManager = IdentityManager()
        self.sceneDetector = SceneDetector()
        self.audioPlayer = AudioPlayer()

        // Initialize new components (v9.8.0)
        self.connectionRecovery = ConnectionRecoveryManager(centerConnection: centerConnection)
        self.distributedOrchestrator = DistributedOrchestrator()
        self.memoryManager = UserMemoryManager()

        // Setup bindings
        setupBindings()
    }

    // MARK: - Public Methods

    /// Initialize the agent
    public func initialize() async throws {
        // Initialize core components
        try await identityManager.initialize()
        sceneDetector.initialize()
        audioPlayer.initialize()

        // Initialize memory manager
        try await memoryManager.initialize()

        // Initialize distributed orchestrator
        distributedOrchestrator.initialize(
            sceneDetector: sceneDetector,
            centerConnection: centerConnection
        )

        // Register local device
        distributedOrchestrator.registerLocalDevice(profile: profile)

        // Connect to Center if in sync mode
        if modeManager.mode == .sync && config.centerAddress != nil {
            try await connectCenter()
        }

        status = .online
        print("OFAiOSAgent initialized (v9.8.0)")
    }

    /// Connect to Center server with error handling
    public func connectCenter() async throws {
        guard let address = config.centerAddress else {
            throw OFAError(
                code: "CONFIG_ERROR",
                message: "Center address not configured",
                category: .validation,
                severity: .high,
                strategy: .none
            )
        }

        // Use retry executor for connection
        let retryExecutor = RetryExecutor(config: .aggressiveConfig)
        let circuitBreaker = CircuitBreaker.defaultBreaker(name: "center_connection")

        do {
            try await retryExecutor.execute(
                operation: { attempt in
                    try await self.centerConnection.connect(address: address)
                    try await self.centerConnection.register(profile: self.profile)
                },
                circuitBreaker: circuitBreaker
            )

            connectedToCenter = true
            status = .online

            // Start heartbeat
            startHeartbeat()

            print("Connected to Center successfully")
        } catch {
            ErrorHandler.notifyError(error: ErrorHandler.categorizeError(error))
            errorCount += 1
            throw error
        }
    }

    /// Disconnect from Center
    public func disconnect() async {
        connectionRecovery.stopRecovery()
        centerConnection.disconnect()
        connectedToCenter = false
        status = .offline
    }

    /// Set agent mode
    public func setMode(_ mode: AgentMode) {
        modeManager.setMode(mode)

        if mode == .sync && !connectedToCenter {
            Task {
                try? await connectCenter()
            }
        }
    }

    /// Sync with Center (v9.8.0 - enhanced)
    public func syncWithCenter() async throws {
        guard connectedToCenter else {
            throw OFAError(
                code: "NOT_CONNECTED",
                message: "Not connected to Center",
                category: .connection,
                severity: .medium,
                strategy: .backoffRetry
            )
        }

        // Sync identity
        if let identity = identityManager.currentIdentity {
            try await centerConnection.syncIdentity(identity)
        }

        // Sync behaviors
        let behaviors = await identityManager.getPendingBehaviors()
        try await centerConnection.reportBehaviors(behaviors)

        // Sync memory (new in v9.8.0)
        let importantMemories = await memoryManager.getByCategory(category: .identity)
        for memory in importantMemories where memory.importance > 0.8 {
            // Would send to Center for backup
        }
    }

    /// Get current status
    public func getStatus() -> AgentStatus {
        return status
    }

    /// Get memory statistics (new in v9.8.0)
    public func getMemoryStats() -> MemoryStats {
        return memoryManager.getStats()
    }

    /// Get distributed orchestrator (new in v9.8.0)
    public func getDistributedOrchestrator() -> DistributedOrchestrator {
        return distributedOrchestrator
    }

    /// Get connection recovery status (new in v9.8.0)
    public func getConnectionRecoveryStatus() -> (isRecovering: Bool, circuitState: CircuitState) {
        return (connectionRecovery.isRecovering, connectionRecovery.circuitState)
    }

    /// Store memory (new in v9.8.0)
    public func storeMemory(key: String, value: Any, category: MemoryCategory = .general) async throws {
        let entry = MemoryEntry(
            key: key,
            value: value,
            level: .l2,
            category: category
        )
        try await memoryManager.store(entry: entry)
    }

    /// Retrieve memory (new in v9.8.0)
    public func retrieveMemory(key: String) async -> Any? {
        let entry = await memoryManager.retrieve(key: key)
        return entry?.value.value
    }

    /// Handle error with recovery (new in v9.8.0)
    public func handleError(error: Error) async {
        let categorized = ErrorHandler.categorizeError(error)
        ErrorHandler.notifyError(error: categorized)
        errorCount += 1

        // Attempt recovery based on strategy
        switch categorized.strategy {
        case .backoffRetry:
            connectionRecovery.startRecovery()

        case .gracefulDegrade:
            // Use cached data
            status = .busy

        case .circuitBreak:
            // Wait for circuit to recover
            status = .error

        default:
            break
        }
    }

    // MARK: - Device Detection

    private static func detectDeviceType() -> AppleDeviceType {
        #if os(iOS)
        if UIDevice.current.userInterfaceIdiom == .pad {
            return .iPad
        } else if UIDevice.current.userInterfaceIdiom == .phone {
            return .iPhone
        }
        #elseif os(macOS)
        return .mac
        #elseif os(watchOS)
        return .watch
        #endif
        return .unknown
    }

    private static func getDeviceName() -> String {
        #if os(iOS)
        return UIDevice.current.name
        #elseif os(macOS)
        return Host.current().localizedName ?? "Mac"
        #else
        return "Unknown"
        #endif
    }

    private static func getOSVersion() -> String {
        #if os(iOS)
        return UIDevice.current.systemVersion
        #elseif os(macOS)
        let version = ProcessInfo.processInfo.operatingSystemVersion
        return "\(version.majorVersion).\(version.minorVersion).\(version.patchVersion)"
        #else
        return "Unknown"
        #endif
    }

    private static func getCapabilities(for deviceType: AppleDeviceType) -> [String] {
        var capabilities: [String] = []

        #if os(iOS)
        capabilities.append("voice")
        capabilities.append("display")
        capabilities.append("camera")
        capabilities.append("location")
        capabilities.append("health")  // If HealthKit available

        if deviceType == .iPad {
            capabilities.append("multitasking")
            capabilities.append("stylus")
        }
        #elseif os(macOS)
        capabilities.append("voice")
        capabilities.append("display")
        capabilities.append("keyboard")
        capabilities.append("mouse")
        capabilities.append("file_system")
        #elseif os(watchOS)
        capabilities.append("health")
        capabilities.append("heart_rate")
        capabilities.append("activity")
        #endif

        return capabilities
    }

    private static func generateAgentId() -> String {
        return "agent_" + UUID().uuidString.lowercased()
    }

    // MARK: - Private Methods

    private func setupBindings() {
        // Bind status changes
        centerConnection.statusPublisher
            .receive(on: DispatchQueue.main)
            .assign(to: &$status)

        // Bind identity changes
        identityManager.identityPublisher
            .receive(on: DispatchQueue.main)
            .assign(to: &$currentIdentityId)

        // Bind connection state
        centerConnection.connectedPublisher
            .receive(on: DispatchQueue.main)
            .assign(to: &$connectedToCenter)
    }

    private func startHeartbeat() {
        Timer.scheduledTimer(withTimeInterval: 30.0, repeats: true) { _ in
            Task {
                try? await self.sendHeartbeat()
            }
        }
    }

    private func sendHeartbeat() async throws {
        try await centerConnection.sendHeartbeat(profile: profile)
    }
}

// MARK: - Agent Configuration

public struct AgentConfig {
    public var agentId: String?
    public var centerAddress: String?
    public var mode: AgentMode
    public var heartbeatInterval: TimeInterval
    public var syncInterval: TimeInterval
    public var enableCache: Bool

    public init(
        agentId: String? = nil,
        centerAddress: String? = nil,
        mode: AgentMode = .sync,
        heartbeatInterval: TimeInterval = 30.0,
        syncInterval: TimeInterval = 300.0,
        enableCache: Bool = true
    ) {
        self.agentId = agentId
        self.centerAddress = centerAddress
        self.mode = mode
        self.heartbeatInterval = heartbeatInterval
        self.syncInterval = syncInterval
        self.enableCache = enableCache
    }
}

// Note: OFAError is now defined in ErrorHandler.swift (v9.8.0)