/**
 * @file p2p.cpp
 * @brief P2P Communication Implementation
 * Sprint 30: P2P Enhancement
 */

#include "ofa/p2p.hpp"
#include <chrono>
#include <thread>
#include <random>
#include <sstream>

// Platform-specific socket includes
#ifdef _WIN32
    #include <winsock2.h>
    #include <ws2tcpip.h>
    #pragma comment(lib, "ws2_32.lib")
    #define CLOSE_SOCKET closesocket
    #define SOCKET_TYPE SOCKET
    #define INVALID_SOCKET_VAL INVALID_SOCKET
#else
    #include <sys/socket.h>
    #include <netinet/in.h>
    #include <arpa/inet.h>
    #include <unistd.h>
    #define CLOSE_SOCKET close
    #define SOCKET_TYPE int
    #define INVALID_SOCKET_VAL -1
#endif

namespace ofa {

// Initialize Winsock on Windows
#ifdef _WIN32
static bool winsockInitialized = false;
static void initWinsock() {
    if (!winsockInitialized) {
        WSADATA wsaData;
        WSAStartup(MAKEWORD(2, 2), &wsaData);
        winsockInitialized = true;
    }
}
#endif

// Helper functions
static std::string generateId() {
    std::random_device rd;
    std::mt19937 gen(rd());
    std::uniform_int_distribution<> dis(0, 15);

    std::stringstream ss;
    ss << std::hex;
    for (int i = 0; i < 16; i++) ss << dis(gen);
    return ss.str();
}

static int64_t nowMs() {
    return std::chrono::duration_cast<std::chrono::milliseconds>(
        std::chrono::system_clock::now().time_since_epoch()
    ).count();
}

// ==================== P2PClient ====================

P2PClient::P2PClient(const std::string& agentId,
                     uint16_t listenPort,
                     uint16_t discoveryPort)
    : agentId_(agentId)
    , listenPort_(listenPort)
    , discoveryPort_(discoveryPort) {
#ifdef _WIN32
    initWinsock();
#endif
}

P2PClient::~P2PClient() {
    stop();
}

std::string P2PClient::generateId() {
    return ::ofa::generateId();
}

std::future<bool> P2PClient::start() {
    return std::async(std::launch::async, [this]() {
        if (running_) return true;

        // Create TCP socket
        tcpSocket_ = socket(AF_INET, SOCK_STREAM, 0);
        if (tcpSocket_ == INVALID_SOCKET_VAL) {
            return false;
        }

        // Bind to port
        struct sockaddr_in addr;
        addr.sin_family = AF_INET;
        addr.sin_addr.s_addr = INADDR_ANY;
        addr.sin_port = htons(listenPort_ == 0 ? 7891 : listenPort_);

        if (bind(tcpSocket_, (struct sockaddr*)&addr, sizeof(addr)) < 0) {
            CLOSE_SOCKET(tcpSocket_);
            return false;
        }

        listen(tcpSocket_, 5);

        // Get actual port
        if (listenPort_ == 0) {
            socklen_t len = sizeof(addr);
            getsockname(tcpSocket_, (struct sockaddr*)&addr, &len);
            listenPort_ = ntohs(addr.sin_port);
        }

        // Create UDP socket for discovery
        udpSocket_ = socket(AF_INET, SOCK_DGRAM, 0);
        if (udpSocket_ == INVALID_SOCKET_VAL) {
            CLOSE_SOCKET(tcpSocket_);
            return false;
        }

        // Enable broadcast
        int broadcast = 1;
        setsockopt(udpSocket_, SOL_SOCKET, SO_BROADCAST,
                   (const char*)&broadcast, sizeof(broadcast));

        running_ = true;
        tcpThread_ = std::thread(&P2PClient::tcpListen, this);

        return true;
    });
}

void P2PClient::stop() {
    if (!running_) return;
    running_ = false;
    discovering_ = false;

    if (tcpSocket_ != INVALID_SOCKET_VAL) {
        CLOSE_SOCKET(tcpSocket_);
        tcpSocket_ = INVALID_SOCKET_VAL;
    }

    if (udpSocket_ != INVALID_SOCKET_VAL) {
        CLOSE_SOCKET(udpSocket_);
        udpSocket_ = INVALID_SOCKET_VAL;
    }

    if (tcpThread_.joinable()) {
        tcpThread_.join();
    }
    if (udpThread_.joinable()) {
        udpThread_.join();
    }
    if (timeoutThread_.joinable()) {
        timeoutThread_.join();
    }
}

void P2PClient::startDiscovery(const std::vector<std::string>& skills) {
    if (discovering_) return;
    discovering_ = true;
    udpThread_ = std::thread(&P2PClient::udpDiscovery, this);
    timeoutThread_ = std::thread(&P2PClient::checkPeerTimeout, this);
}

void P2PClient::stopDiscovery() {
    discovering_ = false;
}

std::vector<PeerInfo> P2PClient::getPeers() const {
    std::lock_guard<std::mutex> lock(mutex_);
    std::vector<PeerInfo> result;
    for (const auto& pair : peers_) {
        if (pair.second.isOnline) {
            result.push_back(pair.second);
        }
    }
    return result;
}

PeerInfo P2PClient::getPeer(const std::string& peerId) const {
    std::lock_guard<std::mutex> lock(mutex_);
    auto it = peers_.find(peerId);
    if (it == peers_.end()) {
        throw OfAException("Peer not found: " + peerId);
    }
    return it->second;
}

std::future<bool> P2PClient::send(const std::string& peerId, const P2PMessage& msg) {
    return std::async(std::launch::async, [this, peerId, msg]() {
        std::lock_guard<std::mutex> lock(mutex_);

        auto it = peers_.find(peerId);
        if (it == peers_.end() || !it->second.isOnline) {
            return false;
        }

        // Create socket and connect
        SOCKET_TYPE sock = socket(AF_INET, SOCK_STREAM, 0);
        if (sock == INVALID_SOCKET_VAL) {
            return false;
        }

        struct sockaddr_in addr;
        addr.sin_family = AF_INET;
        addr.sin_port = htons(it->second.port);

#ifdef _WIN32
        InetPtonA(AF_INET, it->second.address.c_str(), &addr.sin_addr);
#else
        inet_pton(AF_INET, it->second.address.c_str(), &addr.sin_addr);
#endif

        if (connect(sock, (struct sockaddr*)&addr, sizeof(addr)) < 0) {
            CLOSE_SOCKET(sock);
            return false;
        }

        // Send message
        std::string data = msg.toJson().dump();
        send(sock, data.c_str(), (int)data.size(), 0);

        CLOSE_SOCKET(sock);
        return true;
    });
}

std::future<void> P2PClient::broadcast(const P2PMessage& msg) {
    return std::async(std::launch::async, [this, msg]() {
        auto peers = getPeers();
        for (const auto& peer : peers) {
            send(peer.id, msg).get();
        }
    });
}

std::future<json> P2PClient::sendTaskRequest(const std::string& peerId,
                                              const std::string& skillId,
                                              const std::string& operation,
                                              const json& input,
                                              int timeoutMs) {
    return std::async(std::launch::async, [this, peerId, skillId, operation, input, timeoutMs]() {
        P2PMessage msg;
        msg.id = generateId();
        msg.type = P2PMessageType::TaskRequest;
        msg.from = agentId_;
        msg.to = peerId;
        msg.data = {
            {"skillId", skillId},
            {"operation", operation},
            {"input", input}
        };
        msg.timestamp = nowMs();

        // In production: send and wait for response
        // For now: simulate
        send(peerId, msg).get();

        // Would wait for TaskResponse
        return json{{"status", "sent"}};
    });
}

void P2PClient::onMessage(P2PMessageType type, P2PMessageHandler handler) {
    std::lock_guard<std::mutex> lock(mutex_);
    handlers_[type].push_back(handler);
}

void P2PClient::onPeerOnline(std::function<void(const PeerInfo&)> callback) {
    onPeerOnline_ = callback;
}

void P2PClient::onPeerOffline(std::function<void(const PeerInfo&)> callback) {
    onPeerOffline_ = callback;
}

void P2PClient::tcpListen() {
    while (running_) {
        struct sockaddr_in clientAddr;
        socklen_t clientLen = sizeof(clientAddr);

        // Set non-blocking with timeout
        fd_set readfds;
        FD_ZERO(&readfds);
        FD_SET(tcpSocket_, &readfds);

        struct timeval tv;
        tv.tv_sec = 1;
        tv.tv_usec = 0;

        int activity = select((int)tcpSocket_ + 1, &readfds, nullptr, nullptr, &tv);

        if (activity <= 0 || !running_) continue;

        SOCKET_TYPE clientSock = accept(tcpSocket_, (struct sockaddr*)&clientAddr, &clientLen);
        if (clientSock == INVALID_SOCKET_VAL) continue;

        // Handle connection in thread
        std::thread(&P2PClient::handleConnection, this, clientSock).detach();
    }
}

void P2PClient::handleConnection(SOCKET_TYPE clientSock) {
    char buffer[4096];
    int bytesRead = recv(clientSock, buffer, sizeof(buffer) - 1, 0);

    if (bytesRead > 0) {
        buffer[bytesRead] = '\0';

        try {
            json j = json::parse(std::string(buffer, bytesRead));
            P2PMessage msg = P2PMessage::fromJson(j);
            processMessage(msg);
        } catch (...) {
            // Invalid message
        }
    }

    CLOSE_SOCKET(clientSock);
}

void P2PClient::processMessage(const P2PMessage& msg) {
    // Handle discovery
    if (msg.type == P2PMessageType::Discovery) {
        PeerInfo peer = PeerInfo::fromJson(msg.data);
        peer.lastSeen = nowMs();
        peer.isOnline = true;

        std::lock_guard<std::mutex> lock(mutex_);
        bool isNew = peers_.find(peer.id) == peers_.end();
        peers_[peer.id] = peer;

        if (isNew && onPeerOnline_) {
            onPeerOnline_(peer);
        }

        // Send ack
        P2PMessage ack;
        ack.id = generateId();
        ack.type = P2PMessageType::DiscoveryAck;
        ack.from = agentId_;
        ack.data = {{}};  // Self info would go here
        send(peer.id, ack);
    }

    // Handle task request
    if (msg.type == P2PMessageType::TaskRequest) {
        std::lock_guard<std::mutex> lock(mutex_);
        auto it = handlers_.find(P2PMessageType::TaskRequest);
        if (it != handlers_.end()) {
            for (auto& handler : it->second) {
                handler(msg);
            }
        }
    }

    // Call registered handlers
    std::lock_guard<std::mutex> lock(mutex_);
    auto it = handlers_.find(msg.type);
    if (it != handlers_.end()) {
        for (auto& handler : it->second) {
            handler(msg);
        }
    }
}

void P2PClient::checkPeerTimeout() {
    while (running_) {
        std::this_thread::sleep_for(std::chrono::seconds(5));

        auto now = nowMs();
        std::lock_guard<std::mutex> lock(mutex_);

        for (auto& pair : peers_) {
            if (now - pair.second.lastSeen > 30000) {  // 30s timeout
                if (pair.second.isOnline) {
                    pair.second.isOnline = false;
                    if (onPeerOffline_) {
                        onPeerOffline_(pair.second);
                    }
                }
            }
        }
    }
}

void P2PClient::udpDiscovery() {
    while (discovering_ && running_) {
        sendDiscoveryBroadcast();
        std::this_thread::sleep_for(std::chrono::seconds(5));
    }
}

void P2PClient::sendDiscoveryBroadcast() {
    P2PMessage msg;
    msg.id = generateId();
    msg.type = P2PMessageType::Discovery;
    msg.from = agentId_;
    msg.timestamp = nowMs();

    // Self info
    PeerInfo selfInfo;
    selfInfo.id = agentId_;
    selfInfo.address = "0.0.0.0";  // Would get actual IP
    selfInfo.port = listenPort_;
    msg.data = selfInfo.toJson();

    std::string data = msg.toJson().dump();

    struct sockaddr_in addr;
    addr.sin_family = AF_INET;
    addr.sin_port = htons(discoveryPort_);
#ifdef _WIN32
    InetPtonA(AF_INET, "255.255.255.255", &addr.sin_addr);
#else
    inet_pton(AF_INET, "255.255.255.255", &addr.sin_addr);
#endif

    sendto(udpSocket_, data.c_str(), (int)data.size(), 0,
           (struct sockaddr*)&addr, sizeof(addr));
}

// ==================== P2PDiscovery ====================

P2PDiscovery::P2PDiscovery(uint16_t port)
    : port_(port) {
#ifdef _WIN32
    initWinsock();
#endif
}

P2PDiscovery::~P2PDiscovery() {
    stop();
}

bool P2PDiscovery::start() {
    socket_ = socket(AF_INET, SOCK_DGRAM, 0);
    if (socket_ == INVALID_SOCKET_VAL) {
        return false;
    }

    int broadcast = 1;
    setsockopt(socket_, SOL_SOCKET, SO_BROADCAST,
               (const char*)&broadcast, sizeof(broadcast));

    struct sockaddr_in addr;
    addr.sin_family = AF_INET;
    addr.sin_addr.s_addr = INADDR_ANY;
    addr.sin_port = htons(port_);

    if (bind(socket_, (struct sockaddr*)&addr, sizeof(addr)) < 0) {
        CLOSE_SOCKET(socket_);
        return false;
    }

    running_ = true;
    thread_ = std::thread(&P2PDiscovery::listenLoop, this);
    return true;
}

void P2PDiscovery::stop() {
    running_ = false;
    if (socket_ != INVALID_SOCKET_VAL) {
        CLOSE_SOCKET(socket_);
    }
    if (thread_.joinable()) {
        thread_.join();
    }
}

void P2PDiscovery::broadcast(const PeerInfo& selfInfo) {
    P2PMessage msg;
    msg.id = generateId();
    msg.type = P2PMessageType::Discovery;
    msg.from = selfInfo.id;
    msg.data = selfInfo.toJson();
    msg.timestamp = nowMs();

    std::string data = msg.toJson().dump();

    struct sockaddr_in addr;
    addr.sin_family = AF_INET;
    addr.sin_port = htons(port_);
#ifdef _WIN32
    InetPtonA(AF_INET, "255.255.255.255", &addr.sin_addr);
#else
    inet_pton(AF_INET, "255.255.255.255", &addr.sin_addr);
#endif

    sendto(socket_, data.c_str(), (int)data.size(), 0,
           (struct sockaddr*)&addr, sizeof(addr));
}

void P2PDiscovery::onDiscovery(std::function<void(const PeerInfo&)> callback) {
    onDiscovery_ = callback;
}

void P2PDiscovery::listenLoop() {
    char buffer[4096];
    struct sockaddr_in clientAddr;
    socklen_t clientLen = sizeof(clientAddr);

    while (running_) {
        fd_set readfds;
        FD_ZERO(&readfds);
        FD_SET(socket_, &readfds);

        struct timeval tv;
        tv.tv_sec = 1;
        tv.tv_usec = 0;

        int activity = select(socket_ + 1, &readfds, nullptr, nullptr, &tv);

        if (activity <= 0 || !running_) continue;

        int bytesRead = recvfrom(socket_, buffer, sizeof(buffer) - 1, 0,
                                 (struct sockaddr*)&clientAddr, &clientLen);

        if (bytesRead > 0) {
            buffer[bytesRead] = '\0';

            try {
                json j = json::parse(std::string(buffer, bytesRead));
                P2PMessage msg = P2PMessage::fromJson(j);

                if (msg.type == P2PMessageType::Discovery) {
                    PeerInfo peer = PeerInfo::fromJson(msg.data);

                    // Get sender address
#ifdef _WIN32
                    char ipStr[INET_ADDRSTRLEN];
                    InetNtopA(AF_INET, &clientAddr.sin_addr, ipStr, INET_ADDRSTRLEN);
#else
                    char ipStr[INET_ADDRSTRLEN];
                    inet_ntop(AF_INET, &clientAddr.sin_addr, ipStr, INET_ADDRSTRLEN);
#endif

                    peer.address = ipStr;
                    peer.lastSeen = nowMs();
                    peer.isOnline = true;

                    if (onDiscovery_) {
                        onDiscovery_(peer);
                    }
                }
            } catch (...) {
                // Invalid message
            }
        }
    }
}

} // namespace ofa