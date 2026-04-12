// Distributed.swift
// OFA iOS SDK - Distributed Components (v9.8.0)
// Aligned with Android SDK Distributed Package

import Foundation
import Combine

// MARK: - Device Role

/// Device role in the distributed system
public enum DeviceRole: String, Codable {
    case source = "source"       // Data source device (watch, sensors)
    case display = "display"     // Display device (phone, tablet)
    case executor = "executor"   // Execution device (phone, speaker)
    case coordinator = "coordinator"  // Coordinating device
    case all = "all"             // All roles
}

// MARK: - EventBus

/// Event types for cross-device communication
public enum EventType: String, Codable {
    case sceneChange = "scene_change"
    case healthAlert = "health_alert"
    case deviceStatus = "device_status"
    case identitySync = "identity_sync"
    case taskAssign = "task_assign"
    case taskComplete = "task_complete"
    case messageRoute = "message_route"
    case notification = "notification"
    case audioStream = "audio_stream"
    case chatMessage = "chat_message"
}

/// Event for EventBus
public struct Event: Codable {
    public let id: String
    public let type: EventType
    public let sourceAgentId: String
    public let targetAgentIds: [String]?
    public let payload: [String: AnyCodable]
    public let timestamp: Date
    public let priority: EventPriority

    public init(
        type: EventType,
        sourceAgentId: String,
        targetAgentIds: [String]? = nil,
        payload: [String: Any] = [:],
        priority: EventPriority = .normal
    ) {
        self.id = UUID().uuidString
        self.type = type
        self.sourceAgentId = sourceAgentId
        self.targetAgentIds = targetAgentIds
        self.payload = payload.mapValues { AnyCodable($0) }
        self.timestamp = Date()
        self.priority = priority
    }
}

/// Event priority
public enum EventPriority: Int, Codable {
    case low = 0
    case normal = 1
    case high = 2
    case urgent = 3
}

/// Event listener protocol
public protocol EventListener: AnyObject {
    func onEvent(event: Event)
}

/// EventBus - publish/subscribe for cross-device events
public class EventBus: ObservableObject {

    // MARK: - Published Properties

    @Published public var eventCount: Int = 0
    @Published public var lastEventTime: Date?

    // MARK: - Private Properties

    private var listeners: [EventType: [EventListener]] = [:]
    private var eventHistory: [Event] = []
    private let maxHistorySize: Int = 100
    private var cancellables = Set<AnyCancellable>()

    // MARK: - Initialization

    public init() {
        // Initialize empty listener map
        for type in [EventType.sceneChange, .healthAlert, .deviceStatus, .identitySync,
                     .taskAssign, .taskComplete, .messageRoute, .notification,
                     .audioStream, .chatMessage] {
            listeners[type] = []
        }
    }

    // MARK: - Public Methods

    /// Subscribe to events of a specific type
    public func subscribe(type: EventType, listener: EventListener) {
        if listeners[type] == nil {
            listeners[type] = []
        }
        listeners[type]?.append(listener)
    }

    /// Unsubscribe from events
    public func unsubscribe(type: EventType, listener: EventListener) {
        listeners[type]?.removeAll { $0 == listener }
    }

    /// Publish an event
    public func publish(event: Event) {
        // Add to history
        eventHistory.append(event)
        if eventHistory.count > maxHistorySize {
            eventHistory.removeFirst()
        }

        eventCount += 1
        lastEventTime = event.timestamp

        // Notify listeners
        if let typeListeners = listeners[event.type] {
            for listener in typeListeners {
                listener.onEvent(event: event)
            }
        }

        // Also notify wildcard listeners (if implemented)
        print("EventBus: Published \(event.type.rawValue) from \(event.sourceAgentId)")
    }

    /// Publish event to specific targets
    public func publishTo(event: Event, targets: [String]) {
        var targetedEvent = event
        // Would modify event to include targets
        publish(event: event)
    }

    /// Get event history
    public func getHistory(limit: Int = 50) -> [Event] {
        return eventHistory.suffix(limit)
    }

    /// Get events by type
    public func getEventsByType(type: EventType, limit: Int = 20) -> [Event] {
        return eventHistory.filter { $0.type == type }.suffix(limit)
    }

    /// Clear event history
    public func clearHistory() {
        eventHistory.removeAll()
        eventCount = 0
    }
}

// MARK: - CrossDeviceRouter

/// Routing rule for message routing
public struct RoutingRule: Codable {
    public let id: String
    public let name: String
    public let sceneCondition: String?      // Scene type to match
    public let messageTypes: [String]?       // Message types to route
    public let targetRole: DeviceRole        // Target device role
    public let targetDeviceId: String?       // Specific device ID
    public let action: RoutingAction
    public let priority: Int
    public let enabled: Bool

    public init(
        name: String,
        sceneCondition: String? = nil,
        messageTypes: [String]? = nil,
        targetRole: DeviceRole = .display,
        targetDeviceId: String? = nil,
        action: RoutingAction = .route,
        priority: Int = 0,
        enabled: Bool = true
    ) {
        self.id = UUID().uuidString
        self.name = name
        self.sceneCondition = sceneCondition
        self.messageTypes = messageTypes
        self.targetRole = targetRole
        self.targetDeviceId = targetDeviceId
        self.action = action
        self.priority = priority
        self.enabled = enabled
    }
}

/// Routing action
public enum RoutingAction: String, Codable {
    case route = "route"       // Route to target device
    case filter = "filter"     // Filter out message
    case delay = "delay"       // Delay delivery
    case broadcast = "broadcast"  // Broadcast to all devices
    case block = "block"       // Block message
}

/// CrossDeviceRouter - routes messages between devices based on rules
public class CrossDeviceRouter: ObservableObject {

    // MARK: - Published Properties

    @Published public var activeRules: [RoutingRule] = []
    @Published public var routingCount: Int = 0

    // MARK: - Private Properties

    private var eventBus: EventBus
    private var sceneDetector: SceneDetector?
    private var pendingMessages: [PendingMessage] = []
    private var connectedDevices: [String: DeviceInfo] = [:]

    // MARK: - Initialization

    public init(eventBus: EventBus) {
        self.eventBus = eventBus
        self.activeRules = Self.defaultRules()

        // Subscribe to events
        eventBus.subscribe(type: .sceneChange, listener: self)
        eventBus.subscribe(type: .messageRoute, listener: self)
    }

    // MARK: - Public Methods

    /// Set scene detector for scene-aware routing
    public func setSceneDetector(detector: SceneDetector) {
        self.sceneDetector = detector
    }

    /// Add or update routing rule
    public func addRule(rule: RoutingRule) {
        // Remove existing rule with same ID
        activeRules.removeAll { $0.id == rule.id }
        activeRules.append(rule)
        // Sort by priority
        activeRules.sort { $0.priority > $1.priority }
    }

    /// Remove routing rule
    public func removeRule(ruleId: String) {
        activeRules.removeAll { $0.id == ruleId }
    }

    /// Route a message
    public func route(message: Message) -> RoutingResult {
        routingCount += 1

        // Get current scene
        let currentScene = sceneDetector?.currentScene.rawValue ?? "unknown"

        // Find matching rule
        for rule in activeRules where rule.enabled {
            // Check scene condition
            if let sceneCond = rule.sceneCondition, sceneCond != currentScene {
                continue
            }

            // Check message type
            if let types = rule.messageTypes, !types.contains(message.type) {
                continue
            }

            // Execute routing action
            return executeAction(rule: rule, message: message)
        }

        // No matching rule - default routing
        return RoutingResult(action: .route, target: .display, delivered: false)
    }

    /// Register connected device
    public func registerDevice(device: DeviceInfo) {
        connectedDevices[device.agentId] = device
    }

    /// Unregister device
    public func unregisterDevice(deviceId: String) {
        connectedDevices.removeValue(forKey: deviceId)
    }

    /// Get devices by role
    public func getDevicesByRole(role: DeviceRole) -> [DeviceInfo] {
        return connectedDevices.values.filter { $0.role == role }
    }

    /// Get all connected devices
    public func getAllDevices() -> [DeviceInfo] {
        return Array(connectedDevices.values)
    }

    // MARK: - Private Methods

    private func executeAction(rule: RoutingRule, message: Message) -> RoutingResult {
        switch rule.action {
        case .route:
            // Find target device
            if let targetId = rule.targetDeviceId {
                if connectedDevices[targetId] != nil {
                    deliverTo(message: message, deviceId: targetId)
                    return RoutingResult(action: .route, target: rule.targetRole, delivered: true)
                }
            } else {
                // Route by role
                let targets = getDevicesByRole(role: rule.targetRole)
                if let target = targets.first {
                    deliverTo(message: message, deviceId: target.agentId)
                    return RoutingResult(action: .route, target: rule.targetRole, delivered: true)
                }
            }
            return RoutingResult(action: .route, target: rule.targetRole, delivered: false)

        case .filter:
            // Filter out message
            return RoutingResult(action: .filter, target: .all, delivered: false)

        case .delay:
            // Delay delivery
            let pending = PendingMessage(message: message, delaySeconds: 60)
            pendingMessages.append(pending)
            return RoutingResult(action: .delay, target: rule.targetRole, delivered: false)

        case .broadcast:
            // Broadcast to all devices
            for device in connectedDevices.values {
                deliverTo(message: message, deviceId: device.agentId)
            }
            return RoutingResult(action: .broadcast, target: .all, delivered: true)

        case .block:
            // Block message
            return RoutingResult(action: .block, target: .all, delivered: false)
        }
    }

    private func deliverTo(message: Message, deviceId: String) {
        // Would use CenterConnection or P2P to deliver
        let event = Event(
            type: .messageRoute,
            sourceAgentId: message.sourceAgentId,
            targetAgentIds: [deviceId],
            payload: ["message": AnyCodable(message)]
        )
        eventBus.publish(event: event)
    }

    // MARK: - Default Rules

    private static func defaultRules() -> [RoutingRule] {
        var rules: [RoutingRule] = []

        // Running scene - route urgent messages to watch
        rules.append(RoutingRule(
            name: "Running-UrgentToWatch",
            sceneCondition: "running",
            messageTypes: ["urgent", "health"],
            targetRole: .display,
            action: .route,
            priority: 10
        ))

        // Meeting scene - block non-urgent calls
        rules.append(RoutingRule(
            name: "Meeting-BlockCalls",
            sceneCondition: "meeting",
            messageTypes: ["call", "social"],
            targetRole: .all,
            action: .filter,
            priority: 8
        ))

        // Health alert - broadcast to all devices
        rules.append(RoutingRule(
            name: "HealthAlert-Broadcast",
            sceneCondition: "health_alert",
            messageTypes: ["health", "alert"],
            targetRole: .all,
            action: .broadcast,
            priority: 15
        ))

        // Driving scene - delay non-urgent messages
        rules.append(RoutingRule(
            name: "Driving-DelayMessages",
            sceneCondition: "driving",
            messageTypes: ["social", "notification"],
            targetRole: .display,
            action: .delay,
            priority: 7
        ))

        return rules
    }
}

// MARK: - Supporting Types

/// Device info for routing
public struct DeviceInfo: Codable {
    public let agentId: String
    public let deviceType: AppleDeviceType
    public let role: DeviceRole
    public var isOnline: Bool
    public var capabilities: [String]
    public var lastSeen: Date

    public init(
        agentId: String,
        deviceType: AppleDeviceType,
        role: DeviceRole = .display,
        capabilities: [String] = []
    ) {
        self.agentId = agentId
        self.deviceType = deviceType
        self.role = role
        self.isOnline = true
        self.capabilities = capabilities
        self.lastSeen = Date()
    }
}

/// Message for routing
public struct Message: Codable {
    public let id: String
    public let type: String
    public let sourceAgentId: String
    public let content: String
    public let metadata: [String: AnyCodable]
    public let priority: MessagePriority
    public let timestamp: Date

    public init(
        type: String,
        sourceAgentId: String,
        content: String,
        priority: MessagePriority = .normal
    ) {
        self.id = UUID().uuidString
        self.type = type
        self.sourceAgentId = sourceAgentId
        self.content = content
        self.metadata = [:]
        self.priority = priority
        self.timestamp = Date()
    }
}

/// Message priority
public enum MessagePriority: Int, Codable {
    case low = 0
    case normal = 1
    case high = 2
    case urgent = 3
}

/// Routing result
public struct RoutingResult {
    public let action: RoutingAction
    public let target: DeviceRole
    public let delivered: Bool
}

/// Pending message for delayed delivery
public struct PendingMessage {
    public let message: Message
    public let delaySeconds: Int
    public let queuedAt: Date

    public init(message: Message, delaySeconds: Int) {
        self.message = message
        self.delaySeconds = delaySeconds
        self.queuedAt = Date()
    }
}

// MARK: - EventListener Implementation

extension CrossDeviceRouter: EventListener {
    public func onEvent(event: Event) {
        switch event.type {
        case .sceneChange:
            // Update routing based on new scene
            if let scene = event.payload["scene"]?.value as? String {
                print("CrossDeviceRouter: Scene changed to \(scene)")
            }

        case .messageRoute:
            // Process routed message
            if let message = event.payload["message"]?.value as? Message {
                route(message: message)
            }

        default:
            break
        }
    }
}

// MARK: - Distributed Orchestrator

/// Distributed orchestrator - coordinates all distributed components
public class DistributedOrchestrator: ObservableObject {

    // MARK: - Published Properties

    @Published public var isInitialized: Bool = false
    @Published public var connectedDeviceCount: Int = 0

    // MARK: - Components

    public let eventBus: EventBus
    public let router: CrossDeviceRouter
    private var sceneDetector: SceneDetector?
    private var centerConnection: CenterConnection?

    // MARK: - Initialization

    public init() {
        self.eventBus = EventBus()
        self.router = CrossDeviceRouter(eventBus: eventBus)
    }

    /// Initialize with scene detector and center connection
    public func initialize(
        sceneDetector: SceneDetector,
        centerConnection: CenterConnection
    ) {
        self.sceneDetector = sceneDetector
        self.centerConnection = centerConnection

        router.setSceneDetector(detector: sceneDetector)

        // Subscribe scene detector to event bus
        eventBus.subscribe(type: .sceneChange, listener: SceneEventListener(sceneDetector: sceneDetector))

        isInitialized = true
        print("DistributedOrchestrator initialized")
    }

    /// Handle scene change from detector
    public func handleSceneChange(scene: SceneState) {
        let event = Event(
            type: .sceneChange,
            sourceAgentId: scene.agentId,
            payload: [
                "scene": AnyCodable(scene.type.rawValue),
                "confidence": AnyCodable(scene.confidence),
                "active": AnyCodable(scene.active)
            ],
            priority: .high
        )
        eventBus.publish(event: event)
    }

    /// Route message to appropriate device
    public func routeMessage(message: Message) -> RoutingResult {
        return router.route(message: message)
    }

    /// Broadcast to all devices
    public func broadcast(message: Message) {
        let event = Event(
            type: .messageRoute,
            sourceAgentId: message.sourceAgentId,
            targetAgentIds: nil,  // All devices
            payload: ["message": AnyCodable(message)],
            priority: EventPriority(rawValue: message.priority.rawValue) ?? .normal
        )
        eventBus.publish(event: event)
    }

    /// Register local device
    public func registerLocalDevice(profile: AgentProfile) {
        let device = DeviceInfo(
            agentId: profile.agentId,
            deviceType: profile.deviceType,
            role: determineRole(for: profile.deviceType),
            capabilities: profile.capabilities
        )
        router.registerDevice(device: device)
        connectedDeviceCount = router.getAllDevices().count
    }

    private func determineRole(for deviceType: AppleDeviceType) -> DeviceRole {
        switch deviceType {
        case .watch:
            return .source  // Watch provides health data
        case .iPhone:
            return .executor  // Phone executes actions
        case .iPad:
            return .display  // Tablet displays content
        case .mac:
            return .coordinator  // Mac coordinates
        default:
            return .all
        }
    }
}

// MARK: - Scene Event Listener

/// Scene event listener that triggers actions
class SceneEventListener: EventListener {
    private weak var sceneDetector: SceneDetector?

    init(sceneDetector: SceneDetector) {
        self.sceneDetector = sceneDetector
    }

    func onEvent(event: Event) {
        // Handle scene-related events
        if event.type == .sceneChange {
            // Could trigger additional actions
            print("SceneEventListener: Received scene change event")
        }
    }
}