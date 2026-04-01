import Foundation
import Network

/// P2P 消息类型
public enum P2PMessageType: String, Codable, Sendable {
    case data
    case broadcast
    case request
    case response
    case discovery
    case heartbeat
}

/// P2P 消息
public struct P2PMessage: Codable, Sendable {
    public let type: P2PMessageType
    public let fromId: String
    public let toId: String?
    public let data: Data?
    public let timestamp: Date
    public let msgId: String

    public init(type: P2PMessageType, fromId: String, toId: String? = nil, data: Data? = nil) {
        self.type = type
        self.fromId = fromId
        self.toId = toId
        self.data = data
        self.timestamp = Date()
        self.msgId = UUID().uuidString.prefix(8).lowercased()
    }

    public func toJson() -> Data? {
        try? JSONEncoder().encode(self)
    }

    public static func fromJson(_ data: Data) -> P2PMessage? {
        try? JSONDecoder().decode(P2PMessage.self, from: data)
    }
}

/// 设备信息
public struct PeerInfo: Codable, Sendable, Identifiable {
    public let id: String
    public let name: String
    public let address: String
    public let port: Int
    public var online: Bool
    public var lastSeen: Date
    public var latencyMs: Int

    public init(id: String, name: String, address: String, port: Int) {
        self.id = id
        self.name = name
        self.address = address
        self.port = port
        self.online = true
        self.lastSeen = Date()
        self.latencyMs = 0
    }
}

/// P2P 客户端
public actor P2PClient {
    private let agentId: String
    private let port: Int
    private var peers: [String: PeerInfo] = [:]
    private var running = false

    private var listener: NWListener?
    private var connections: [NWConnection] = []

    private var messageHandlers: [(P2PMessage) -> Void] = []
    private var peerHandlers: [(String, String, Bool) -> Void] = []

    public init(agentId: String, port: Int = 0) {
        self.agentId = agentId
        self.port = port > 0 ? port : Int.random(in: 9000...9999)
    }

    /// 启动 P2P 服务
    public func start() async throws {
        guard !running else { return }
        running = true

        let parameters = NWParameters.tcp
        parameters.allowLocalEndpointReuse = true

        listener = try NWListener(using: parameters, on: NWEndpoint.Port(rawValue: UInt16(port))!)

        listener?.newConnectionHandler = { [weak self] connection in
            Task {
                await self?.handleConnection(connection)
            }
        }

        listener?.start(queue: .global())
        print("P2P server started on port \(self.port)")
    }

    /// 停止 P2P 服务
    public func stop() {
        running = false
        listener?.cancel()
        connections.forEach { $0.cancel() }
        connections.removeAll()
        print("P2P server stopped")
    }

    private func handleConnection(_ connection: NWConnection) {
        connection.start(queue: .global())
        connections.append(connection)

        connection.receive(minimumIncompleteLength: 1, maximumLength: 65536) { [weak self] data, _, _, error in
            if let data = data, let message = P2PMessage.fromJson(data) {
                Task {
                    await self?.handleMessage(message)
                }
            }
            if let error = error {
                print("Connection error: \(error)")
            }
        }
    }

    private func handleMessage(_ message: P2PMessage) {
        // 更新设备状态
        if var peer = peers[message.fromId] {
            peer.lastSeen = Date()
            peer.online = true
            peers[message.fromId] = peer
        }

        // 通知处理器
        for handler in messageHandlers {
            handler(message)
        }
    }

    /// 发送消息
    public func send(peerId: String, data: Data) async -> Bool {
        guard let peer = peers[peerId], peer.online else {
            print("Peer not found or offline: \(peerId)")
            return false
        }

        let message = P2PMessage(
            type: .data,
            fromId: agentId,
            toId: peerId,
            data: data
        )

        return await sendToPeer(peer, message: message)
    }

    private func sendToPeer(_ peer: PeerInfo, message: P2PMessage) async -> Bool {
        guard let jsonData = message.toJson() else { return false }

        let endpoint = NWEndpoint.Host(peer.address)
        let port = NWEndpoint.Port(rawValue: UInt16(peer.port))!

        let connection = NWConnection(host: endpoint, port: port, using: .tcp)
        connection.start(queue: .global())

        return await withCheckedContinuation { continuation in
            connection.send(content: jsonData, completion: .contentProcessed { error in
                if let error = error {
                    print("Send failed: \(error)")
                    continuation.resume(returning: false)
                } else {
                    continuation.resume(returning: true)
                }
                connection.cancel()
            })
        }
    }

    /// 广播消息
    public func broadcast(data: Data) async -> [String: Bool] {
        var results: [String: Bool] = [:]

        for (peerId, peer) in peers where peer.online {
            results[peerId] = await send(peerId: peerId, data: data)
        }

        return results
    }

    /// 添加设备
    public func addPeer(_ peer: PeerInfo) {
        peers[peer.id] = peer

        for handler in peerHandlers {
            handler(peer.id, peer.name, true)
        }

        print("Peer added: \(peer.id)")
    }

    /// 移除设备
    public func removePeer(_ peerId: String) {
        if let peer = peers.removeValue(forKey: peerId) {
            for handler in peerHandlers {
                handler(peerId, peer.name, false)
            }
            print("Peer removed: \(peerId)")
        }
    }

    /// 获取设备列表
    public func getPeers() -> [PeerInfo] {
        Array(peers.values)
    }

    /// 获取在线设备
    public func getOnlinePeers() -> [PeerInfo] {
        peers.values.filter { $0.online }
    }

    /// 获取端口
    public func getPort() -> Int {
        port
    }

    /// 添加消息处理器
    public func onMessage(_ handler: @escaping (P2PMessage) -> Void) {
        messageHandlers.append(handler)
    }

    /// 添加设备变化处理器
    public func onPeerChange(_ handler: @escaping (String, String, Bool) -> Void) {
        peerHandlers.append(handler)
    }

    /// 检查设备状态
    public func checkPeersStatus() {
        let now = Date()
        for (id, peer) in peers {
            if now.timeIntervalSince(peer.lastSeen) > 30 {
                peers[id]?.online = false
                for handler in peerHandlers {
                    handler(id, peer.name, false)
                }
            }
        }
    }

    /// 获取统计信息
    public func getStats() -> P2PStats {
        let online = peers.values.filter { $0.online }.count
        return P2PStats(
            agentId: agentId,
            port: port,
            peersTotal: peers.count,
            peersOnline: online
        )
    }
}

/// P2P 统计信息
public struct P2PStats: Sendable {
    public let agentId: String
    public let port: Int
    public let peersTotal: Int
    public let peersOnline: Int
}