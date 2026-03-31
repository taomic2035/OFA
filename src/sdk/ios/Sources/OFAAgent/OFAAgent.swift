import Foundation
import GRPC
import NIOCore
import SwiftProtobuf

/// OFA Agent for iOS devices
public final class OFAAgent: @unchecked Sendable {

    // MARK: - Types

    /// Agent type enumeration
    public enum AgentType: Int, Sendable {
        case full = 1
        case mobile = 2
        case lite = 3
        case iot = 4
        case edge = 5
    }

    /// Agent status
    public enum Status: Int, Sendable {
        case unknown = 0
        case online = 1
        case busy = 2
        case idle = 3
        case offline = 4
    }

    /// Connection state
    public enum ConnectionState: Sendable {
        case disconnected
        case connecting
        case connected
        case reconnecting
        case error(Error)
    }

    // MARK: - Properties

    /// Agent identifier
    public let agentId: String

    /// Agent name
    public let name: String

    /// Agent type
    public let type: AgentType

    /// Center server address
    public let centerAddress: String

    /// Center server port
    public let centerPort: Int

    /// Current connection state
    public private(set) var connectionState: ConnectionState = .disconnected

    /// Registered skills
    private var skills: [String: any SkillExecutor] = [:]

    /// gRPC channel
    private var channel: GRPCChannel?

    /// gRPC client
    private var client: AgentServiceClient?

    /// Bidirectional stream
    private var call: AgentServiceConnectCall?

    /// Heartbeat timer
    private var heartbeatTimer: Timer?

    /// Heartbeat interval in seconds
    public var heartbeatInterval: TimeInterval = 30.0

    /// Connection delegate
    public weak var delegate: OFAAgentDelegate?

    /// Task delegate
    public weak var taskDelegate: OFAAgentTaskDelegate?

    /// Lock for thread safety
    private let lock = NSLock()

    // MARK: - Initialization

    /// Initialize agent with configuration
    public init(
        agentId: String? = nil,
        name: String? = nil,
        type: AgentType = .mobile,
        centerAddress: String,
        centerPort: Int = 9090
    ) {
        self.agentId = agentId ?? UUID().uuidString
        self.name = name ?? UIDevice.current.name
        self.type = type
        self.centerAddress = centerAddress
        self.centerPort = centerPort

        // Register built-in skills
        registerBuiltInSkills()
    }

    deinit {
        disconnect()
    }

    // MARK: - Skill Management

    /// Register a skill executor
    public func registerSkill(_ skill: any SkillExecutor) {
        lock.lock()
        defer { lock.unlock() }
        skills[skill.skillId] = skill
    }

    /// Unregister a skill
    public func unregisterSkill(_ skillId: String) {
        lock.lock()
        defer { lock.unlock() }
        skills.removeValue(forKey: skillId)
    }

    /// Get registered skill IDs
    public func getRegisteredSkills() -> [String] {
        lock.lock()
        defer { lock.unlock() }
        return Array(skills.keys)
    }

    private func registerBuiltInSkills() {
        // Built-in skills can be added here
    }

    // MARK: - Connection

    /// Connect to OFA Center
    public func connect() async throws {
        await updateConnectionState(.connecting)

        do {
            // Create gRPC channel
            let group = MultiThreadedEventLoopGroup(numberOfThreads: 1)
            let channel = try GRPCChannelPool.with(
                target: .host(centerAddress, port: centerPort),
                transportSecurity: .plaintext,
                eventLoopGroup: group
            )

            self.channel = channel
            self.client = AgentServiceClient(channel: channel)

            // Start bidirectional stream
            try await startStream()

            // Send registration
            try await sendRegistration()

            await updateConnectionState(.connected)
            startHeartbeat()

        } catch {
            await updateConnectionState(.error(error))
            throw error
        }
    }

    /// Disconnect from Center
    public func disconnect() {
        stopHeartbeat()

        call?.response.cancel()
        call = nil
        client = nil

        channel?.close()
        channel = nil

        Task {
            await updateConnectionState(.disconnected)
        }
    }

    private func startStream() async throws {
        guard let client = client else {
            throw OFAError.notConnected
        }

        let call = client.connect { [weak self] message in
            self?.handleCenterMessage(message)
        }

        self.call = call
    }

    private func sendRegistration() async throws {
        let request = Ofa_AgentMessage.with {
            $0.msgID = UUID().uuidString
            $0.register = Ofa_RegisterRequest.with {
                $0.agentID = agentId
                $0.name = name
                $0.type = Ofa_AgentType(rawValue: type.rawValue) ?? .mobile
                $0.deviceInfo = getDeviceInfo()
                $0.capabilities = getCapabilities()
            }
        }

        try await call?.requestStream.send(request)
    }

    private func handleCenterMessage(_ message: Ofa_CenterMessage) {
        switch message.payload {
        case .task(let task):
            handleTaskAssignment(task)
        case .config(let config):
            handleConfigUpdate(config)
        case .broadcast(let broadcast):
            handleBroadcast(broadcast)
        case .message(let msg):
            handleMessage(msg)
        default:
            break
        }
    }

    // MARK: - Heartbeat

    private func startHeartbeat() {
        stopHeartbeat()

        heartbeatTimer = Timer.scheduledTimer(
            withTimeInterval: heartbeatInterval,
            repeats: true
        ) { [weak self] _ in
            Task {
                try? await self?.sendHeartbeat()
            }
        }
    }

    private func stopHeartbeat() {
        heartbeatTimer?.invalidate()
        heartbeatTimer = nil
    }

    private func sendHeartbeat() async throws {
        let request = Ofa_AgentMessage.with {
            $0.msgID = UUID().uuidString
            $0.heartbeat = Ofa_HeartbeatRequest.with {
                $0.agentID = agentId
                $0.status = Ofa_AgentStatus.online
                $0.resources = getResourceUsage()
            }
        }

        try await call?.requestStream.send(request)
    }

    // MARK: - Task Handling

    private func handleTaskAssignment(_ task: Ofa_TaskAssignment) {
        Task { [weak self] in
            guard let self = self else { return }

            await self.taskDelegate?.agent(self, didReceiveTask: task.taskID, skillId: task.skillID)

            do {
                guard let executor = self.skills[task.skillID] else {
                    throw OFAError.skillNotFound(task.skillID)
                }

                let output = try await executor.execute(task.input)

                let result = Ofa_AgentMessage.with {
                    $0.msgID = UUID().uuidString
                    $0.taskResult = Ofa_TaskResult.with {
                        $0.taskID = task.taskID
                        $0.status = .completed
                        $0.output = output
                    }
                }

                try await self.call?.requestStream.send(result)
                await self.taskDelegate?.agent(self, didCompleteTask: task.taskID)

            } catch {
                let result = Ofa_AgentMessage.with {
                    $0.msgID = UUID().uuidString
                    $0.taskResult = Ofa_TaskResult.with {
                        $0.taskID = task.taskID
                        $0.status = .failed
                        $0.error = error.localizedDescription
                    }
                }

                try? await self.call?.requestStream.send(result)
                await self.taskDelegate?.agent(self, didFailTask: task.taskID, error: error)
            }
        }
    }

    private func handleConfigUpdate(_ config: Ofa_ConfigUpdate) {
        // Handle configuration updates
    }

    private func handleBroadcast(_ broadcast: Ofa_BroadcastMessage) {
        // Handle broadcast messages
    }

    private func handleMessage(_ message: Ofa_Message) {
        // Handle direct messages
    }

    // MARK: - Device Info

    private func getDeviceInfo() -> Ofa_DeviceInfo {
        let device = UIDevice.current

        return Ofa_DeviceInfo.with {
            $0.os = "ios"
            $0.osVersion = device.systemVersion
            $0.model = device.model
            $0.manufacturer = "Apple"
            $0.cpuCores = Int32(ProcessInfo.processInfo.processorCount)
            $0.totalMemory = Int64(ProcessInfo.processInfo.physicalMemory)
            $0.arch = "arm64"
        }
    }

    private func getResourceUsage() -> Ofa_ResourceUsage {
        var memoryUsed: Double = 0

        var info = mach_task_basic_info()
        var count = mach_msg_type_number_t(MemoryLayout<mach_task_basic_info>.size) / 4

        let result = withUnsafeMutablePointer(to: &info) {
            $0.withMemoryRebound(to: integer_t.self, capacity: Int(count)) {
                task_info(mach_task_self_, task_flavor_t(MACH_TASK_BASIC_INFO), $0, &count)
            }
        }

        if result == KERN_SUCCESS {
            memoryUsed = Double(info.resident_size) / Double(ProcessInfo.processInfo.physicalMemory) * 100
        }

        return Ofa_ResourceUsage.with {
            $0.memoryUsage = memoryUsed
            $0.batteryLevel = Int32(UIDevice.current.batteryLevel * 100)
            $0.networkType = "wifi" // Would need actual detection
        }
    }

    private func getCapabilities() -> [Ofa_Capability] {
        lock.lock()
        defer { lock.unlock() }

        return skills.map { (_, skill) in
            Ofa_Capability.with {
                $0.id = skill.skillId
                $0.name = skill.skillName
                $0.category = skill.category
            }
        }
    }

    // MARK: - State Management

    private func updateConnectionState(_ state: ConnectionState) async {
        connectionState = state
        await delegate?.agent(self, didChangeConnectionState: state)
    }
}

// MARK: - Delegate Protocols

/// Connection and lifecycle delegate
public protocol OFAAgentDelegate: AnyObject, Sendable {
    func agent(_ agent: OFAAgent, didChangeConnectionState state: OFAAgent.ConnectionState)
}

/// Task execution delegate
public protocol OFAAgentTaskDelegate: AnyObject, Sendable {
    func agent(_ agent: OFAAgent, didReceiveTask taskId: String, skillId: String)
    func agent(_ agent: OFAAgent, didCompleteTask taskId: String)
    func agent(_ agent: OFAAgent, didFailTask taskId: String, error: Error)
}

// MARK: - Errors

public enum OFAError: Error, Sendable {
    case notConnected
    case skillNotFound(String)
    case executionFailed(String)
    case invalidResponse
}