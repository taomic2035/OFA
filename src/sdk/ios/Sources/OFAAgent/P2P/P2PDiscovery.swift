import Foundation
import Network

/// P2P 设备发现
public actor P2PDiscovery {
    private let agentId: String
    private let p2pClient: P2PClient

    private var discovering = false
    private var browser: NWBrowser?
    private var advertiser: NWBrowser?

    private let bonjourType = "_ofa._tcp"
    private let discoveryInterval: TimeInterval = 5.0

    public init(agentId: String, p2pClient: P2PClient) {
        self.agentId = agentId
        self.p2pClient = p2pClient
    }

    /// 启动设备发现
    public func startDiscovery() async {
        guard !discovering else { return }
        discovering = true

        // 启动 Bonjour 浏览
        let parameters = NWParameters()
        parameters.includePeerToPeer = true

        browser = NWBrowser(for: .bonjour(type: bonjourType, domain: nil), using: parameters)

        browser?.stateUpdateHandler = { state in
            switch state {
            case .ready:
                print("Discovery browser ready")
            case .failed(let error):
                print("Discovery browser failed: \(error)")
            default:
                break
            }
        }

        browser?.browseResultsChangedHandler = { [weak self] results, changes in
            Task {
                await self?.handleBrowseResults(results, changes: changes)
            }
        }

        browser?.start(queue: .global())
        print("P2P discovery started")
    }

    /// 停止设备发现
    public func stopDiscovery() {
        discovering = false
        browser?.cancel()
        browser = nil
        print("P2P discovery stopped")
    }

    private func handleBrowseResults(_ results: Set<NWBrowser.Result>, changes: [NWBrowser.Result.Change]) {
        for change in changes {
            switch change {
            case .added(let result):
                handleDiscoveredPeer(result)
            case .removed(let result):
                handleRemovedPeer(result)
            default:
                break
            }
        }
    }

    private func handleDiscoveredPeer(_ result: NWBrowser.Result) {
        guard case .service(let name, let type, let domain, _) = result.endpoint else {
            return
        }

        // 解析服务
        let parameters = NWParameters()
        let connection = NWConnection(to: result.endpoint, using: parameters)

        connection.stateUpdateHandler = { [weak self] state in
            if case .ready = state {
                // 获取地址和端口
                if let endpoint = connection.currentPath?.remoteEndpoint,
                   case .hostPort(let host, let port) = endpoint {
                    let address = String(describing: host)
                    let portInt = Int(port.rawValue)

                    let peerId = name.replacingOccurrences(of: "OFA-", with: "")
                    if peerId != self?.agentId {
                        let peer = PeerInfo(
                            id: peerId,
                            name: name,
                            address: address,
                            port: portInt
                        )
                        Task {
                            await self?.p2pClient.addPeer(peer)
                        }
                    }
                }
            }
        }

        connection.start(queue: .global())
    }

    private func handleRemovedPeer(_ result: NWBrowser.Result) {
        guard case .service(let name, _, _, _) = result.endpoint else {
            return
        }

        let peerId = name.replacingOccurrences(of: "OFA-", with: "")
        Task {
            await p2pClient.removePeer(peerId)
        }
    }

    /// 手动发现指定地址
    public func manualDiscovery(addresses: [String]) async {
        for address in addresses {
            let parts = address.split(separator: ":")
            guard parts.count >= 1 else { continue }

            let host = String(parts[0])
            let port = parts.count > 1 ? Int(String(parts[1])) ?? 9090 : 9090

            // 尝试连接
            let endpoint = NWEndpoint.Host(host)
            guard let nwPort = NWEndpoint.Port(rawValue: UInt16(port)) else { continue }

            let connection = NWConnection(host: endpoint, port: nwPort, using: .tcp)
            connection.start(queue: .global())

            // 发送发现请求
            let message = P2PMessage(
                type: .discovery,
                fromId: agentId,
                data: "request_info".data(using: .utf8)
            )

            if let data = message.toJson() {
                connection.send(content: data, completion: .contentProcessed { _ in
                    connection.cancel()
                })
            }
        }
    }

    /// 广播自身存在
    public func broadcastPresence(port: Int) {
        // 创建 Bonjour 服务
        // 注意: 实际实现需要使用 NSNetService 或 NWListener 的 Bonjour 功能
        print("Broadcasting presence on port \(port)")
    }
}