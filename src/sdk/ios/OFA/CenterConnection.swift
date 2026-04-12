// CenterConnection.swift
// OFA iOS SDK - WebSocket Connection to Center (v8.1.0)

import Foundation
import Combine

/// WebSocket connection state
public enum ConnectionState: String, Codable {
    case disconnected = "disconnected"
    case connecting = "connecting"
    case connected = "connected"
    case reconnecting = "reconnecting"
    case error = "error"
}

/// WebSocket message types
public enum MessageType: String, Codable {
    case register = "Register"
    case registerAck = "RegisterAck"
    case heartbeat = "Heartbeat"
    case stateUpdate = "StateUpdate"
    case taskAssign = "TaskAssign"
    case taskResult = "TaskResult"
    case syncRequest = "SyncRequest"
    case syncResponse = "SyncResponse"
    case behaviorReport = "BehaviorReport"
    case emotionUpdate = "EmotionUpdate"
    case error = "Error"
    case audioStream = "audio_stream"
    case audioChunk = "audio_chunk"
    case audioEnd = "audio_end"
    case ttsRequest = "tts_request"
    case chat = "chat"
    case chatStream = "chat_stream"
}

/// WebSocket message
public struct WebSocketMessage: Codable {
    public let type: MessageType
    public let payload: [String: AnyCodable]
    public let timestamp: Date
    public let messageId: String?

    public init(type: MessageType, payload: [String: AnyCodable], messageId: String? = nil) {
        self.type = type
        self.payload = payload
        self.timestamp = Date()
        self.messageId = messageId
    }
}

/// Center connection manager
public class CenterConnection: ObservableObject {

    // MARK: - Published Properties

    @Published public var connectionState: ConnectionState = .disconnected
    @Published public var sessionId: String?

    // MARK: - Publishers

    public let statusPublisher: AnyPublisher<AgentStatus, Never>
    public let connectedPublisher: AnyPublisher<Bool, Never>
    public let messagePublisher: AnyPublisher<WebSocketMessage, Never>

    // MARK: - Private Properties

    private let centerAddress: String
    private var webSocket: URLSessionWebSocketTask?
    private var urlSession: URLSession
    private var heartbeatTimer: Timer?
    private var reconnectAttempts: Int = 0
    private let maxReconnectAttempts: Int = 3

    private var statusSubject = CurrentValueSubject<AgentStatus, Never>(.offline)
    private var connectedSubject = CurrentValueSubject<Bool, Never>(false)
    private var messageSubject = PassthroughSubject<WebSocketMessage, Never>()

    private var cancellables = Set<AnyCancellable>()

    // MARK: - Initialization

    public init(centerAddress: String?) {
        self.centerAddress = centerAddress ?? ""
        self.urlSession = URLSession(configuration: .default)

        statusPublisher = statusSubject.eraseToAnyPublisher()
        connectedPublisher = connectedSubject.eraseToAnyPublisher()
        messagePublisher = messageSubject.eraseToAnyPublisher()
    }

    // MARK: - Public Methods

    /// Connect to Center server
    public func connect(address: String) async throws {
        connectionState = .connecting

        let wsURL = URL(string: address)!
        var request = URLRequest(url: wsURL)
        request.setValue("websocket", forHTTPHeaderField: "Upgrade")
        request.setValue("Upgrade", forHTTPHeaderField: "Connection")

        webSocket = urlSession.webSocketTask(with: request)
        webSocket?.resume()

        // Start receiving messages
        receiveMessages()

        connectionState = .connected
        connectedSubject.send(true)
        statusSubject.send(.online)

        reconnectAttempts = 0
    }

    /// Register agent with Center
    public func register(profile: AgentProfile) async throws {
        let payload: [String: AnyCodable] = [
            "agent_id": AnyCodable(profile.agentId),
            "device_type": AnyCodable(profile.deviceType.rawValue),
            "device_name": AnyCodable(profile.deviceName),
            "capabilities": AnyCodable(profile.capabilities),
            "identity_id": AnyCodable(profile.identityId ?? "")
        ]

        let message = WebSocketMessage(type: .register, payload: payload)
        try await send(message)
    }

    /// Send heartbeat
    public func sendHeartbeat(profile: AgentProfile) async throws {
        let payload: [String: AnyCodable] = [
            "agent_id": AnyCodable(profile.agentId),
            "status": AnyCodable(profile.status.rawValue),
            "timestamp": AnyCodable(Date().ISO8601Format())
        ]

        let message = WebSocketMessage(type: .heartbeat, payload: payload)
        try await send(message)
    }

    /// Sync identity with Center
    public func syncIdentity(_ identity: PersonalIdentity) async throws {
        let payload: [String: AnyCodable] = [
            "identity": AnyCodable(identity)
        ]

        let message = WebSocketMessage(type: .syncRequest, payload: payload)
        try await send(message)
    }

    /// Report behaviors to Center
    public func reportBehaviors(_ behaviors: [BehaviorObservation]) async throws {
        let payload: [String: AnyCodable] = [
            "behaviors": AnyCodable(behaviors)
        ]

        let message = WebSocketMessage(type: .behaviorReport, payload: payload)
        try await send(message)
    }

    /// Disconnect from Center
    public func disconnect() {
        heartbeatTimer?.invalidate()
        webSocket?.cancel(with: .normalClosure, reason: nil)
        webSocket = nil

        connectionState = .disconnected
        connectedSubject.send(false)
        statusSubject.send(.offline)
        sessionId = nil
    }

    /// Send WebSocket message
    public func send(_ message: WebSocketMessage) async throws {
        guard let webSocket = webSocket else {
            throw OFAError.connectionError("WebSocket not connected")
        }

        let encoder = JSONEncoder()
        let data = try encoder.encode(message)
        webSocket.send(.data(data))
    }

    // MARK: - Private Methods

    private func receiveMessages() {
        guard let webSocket = webSocket else { return }

        webSocket.receive { result in
            switch result {
            case .success(let message):
                switch message {
                case .data(let data):
                    self.handleReceivedData(data)
                case .string(let string):
                    self.handleReceivedString(string)
                @unknown default:
                    break
                }

                // Continue receiving
                self.receiveMessages()

            case .failure(let error):
                print("WebSocket receive error: \(error)")
                self.handleConnectionError(error)
            }
        }
    }

    private func handleReceivedData(_ data: Data) {
        do {
            let decoder = JSONDecoder()
            let message = try decoder.decode(WebSocketMessage.self, from: data)
            handleReceivedMessage(message)
        } catch {
            print("Failed to decode message: \(error)")
        }
    }

    private func handleReceivedString(_ string: String) {
        do {
            let data = string.data(using: .utf8)!
            let decoder = JSONDecoder()
            let message = try decoder.decode(WebSocketMessage.self, from: data)
            handleReceivedMessage(message)
        } catch {
            print("Failed to decode string message: \(error)")
        }
    }

    private func handleReceivedMessage(_ message: WebSocketMessage) {
        messageSubject.send(message)

        switch message.type {
        case .registerAck:
            if let sessionId = message.payload["session_id"]?.value as? String {
                self.sessionId = sessionId
                print("Registered with session: \(sessionId)")
            }

        case .stateUpdate:
            print("Received state update")

        case .error:
            if let errorMsg = message.payload["message"]?.value as? String {
                print("Received error: \(errorMsg)")
            }

        default:
            break
        }
    }

    private func handleConnectionError(_ error: Error) {
        connectionState = .error
        statusSubject.send(.error)
        connectedSubject.send(false)

        // Attempt reconnect
        attemptReconnect()
    }

    private func attemptReconnect() {
        if reconnectAttempts >= maxReconnectAttempts {
            print("Max reconnect attempts reached")
            connectionState = .disconnected
            return
        }

        reconnectAttempts += 1
        connectionState = .reconnecting

        DispatchQueue.main.asyncAfter(deadline: .now() + 5.0) {
            Task {
                try? await self.connect(address: self.centerAddress)
            }
        }
    }
}

// MARK: - Helper Types

/// AnyCodable for encoding/decoding arbitrary JSON values
public struct AnyCodable: Codable {
    public let value: Any

    public init(_ value: Any) {
        self.value = value
    }

    public init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()

        if let string = try? container.decode(String.self) {
            value = string
        } else if let int = try? container.decode(Int.self) {
            value = int
        } else if let double = try? container.decode(Double.self) {
            value = double
        } else if let bool = try? container.decode(Bool.self) {
            value = bool
        } else if let array = try? container.decode([AnyCodable].self) {
            value = array.map { $0.value }
        } else if let dict = try? container.decode([String: AnyCodable].self) {
            value = dict.mapValues { $0.value }
        } else {
            value = ""
        }
    }

    public func encode(to encoder: Encoder) throws {
        var container = encoder.singleValueContainer()

        if let string = value as? String {
            try container.encode(string)
        } else if let int = value as? Int {
            try container.encode(int)
        } else if let double = value as? Double {
            try container.encode(double)
        } else if let bool = value as? Bool {
            try container.encode(bool)
        }
    }
}