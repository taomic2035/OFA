/**
 * @file p2p.hpp
 * @brief P2P Communication for OFA C++ SDK
 * @version 8.1.0
 * Sprint 30: P2P Enhancement
 */

#ifndef OFA_P2P_HPP
#define OFA_P2P_HPP

#include <string>
#include <vector>
#include <map>
#include <memory>
#include <functional>
#include <atomic>
#include <mutex>
#include <future>
#include <chrono>
#include <thread>

#include "types.hpp"
#include "error.hpp"

namespace ofa {

/**
 * @brief 对端节点信息
 */
struct PeerInfo {
    std::string id;
    std::string name;
    std::string address;       // IP地址
    uint16_t port;             // 端口
    std::vector<std::string> skills;
    std::map<std::string, std::string> metadata;
    int64_t lastSeen;          // 最后发现时间
    bool isOnline;

    json toJson() const {
        json j;
        j["id"] = id;
        j["name"] = name;
        j["address"] = address;
        j["port"] = port;
        j["skills"] = skills;
        j["metadata"] = metadata;
        j["lastSeen"] = lastSeen;
        j["isOnline"] = isOnline;
        return j;
    }

    static PeerInfo fromJson(const json& j) {
        PeerInfo peer;
        peer.id = j["id"].get<std::string>();
        peer.name = j.value("name", "");
        peer.address = j["address"].get<std::string>();
        peer.port = j["port"].get<uint16_t>();
        peer.skills = j.value("skills", std::vector<std::string>{});
        peer.metadata = j.value("metadata", std::map<std::string, std::string>{});
        peer.lastSeen = j.value("lastSeen", 0);
        peer.isOnline = j.value("isOnline", true);
        return peer;
    }
};

/**
 * @brief P2P消息类型
 */
enum class P2PMessageType {
    Discovery,      // 发现请求
    DiscoveryAck,   // 发现响应
    TaskRequest,    // 任务请求
    TaskResponse,   // 任务响应
    DataSync,       // 数据同步
    Heartbeat,      // 心跳
    Custom          // 自定义消息
};

/**
 * @brief P2P消息
 */
struct P2PMessage {
    std::string id;
    P2PMessageType type;
    std::string from;
    std::string to;
    json data;
    int64_t timestamp;

    json toJson() const {
        json j;
        j["id"] = id;
        j["type"] = static_cast<int>(type);
        j["from"] = from;
        j["to"] = to;
        j["data"] = data;
        j["timestamp"] = timestamp;
        return j;
    }

    static P2PMessage fromJson(const json& j) {
        P2PMessage msg;
        msg.id = j["id"].get<std::string>();
        msg.type = static_cast<P2PMessageType>(j["type"].get<int>());
        msg.from = j["from"].get<std::string>();
        msg.to = j.value("to", "");
        msg.data = j.value("data", json::object());
        msg.timestamp = j["timestamp"].get<int64_t>();
        return msg;
    }
};

/**
 * @brief P2P消息处理器
 */
using P2PMessageHandler = std::function<void(const P2PMessage&)>;

/**
 * @brief P2P客户端
 */
class P2PClient {
public:
    explicit P2PClient(const std::string& agentId,
                       uint16_t listenPort = 0,
                       uint16_t discoveryPort = 7890);
    ~P2PClient();

    /**
     * @brief 启动P2P服务
     */
    std::future<bool> start();

    /**
     * @brief 停止P2P服务
     */
    void stop();

    /**
     * @brief 开始发现对端
     */
    void startDiscovery(const std::vector<std::string>& skills = {});

    /**
     * @brief 停止发现
     */
    void stopDiscovery();

    /**
     * @brief 获取发现的对端列表
     */
    std::vector<PeerInfo> getPeers() const;

    /**
     * @brief 获取特定对端
     */
    PeerInfo getPeer(const std::string& peerId) const;

    /**
     * @brief 发送消息到对端
     */
    std::future<bool> send(const std::string& peerId, const P2PMessage& msg);

    /**
     * @brief 广播消息
     */
    std::future<void> broadcast(const P2PMessage& msg);

    /**
     * @brief 发送任务请求到对端
     */
    std::future<json> sendTaskRequest(const std::string& peerId,
                                       const std::string& skillId,
                                       const std::string& operation,
                                       const json& input,
                                       int timeoutMs = 5000);

    /**
     * @brief 注册消息处理器
     */
    void onMessage(P2PMessageType type, P2PMessageHandler handler);

    /**
     * @brief 设置对端在线状态回调
     */
    void onPeerOnline(std::function<void(const PeerInfo&)> callback);
    void onPeerOffline(std::function<void(const PeerInfo&)> callback);

    /**
     * @brief 获取本地端口
     */
    uint16_t localPort() const { return listenPort_; }

    /**
     * @brief 是否运行中
     */
    bool isRunning() const { return running_; }

private:
    void tcpListen();
    void udpDiscovery();
    void handleConnection();
    void processMessage(const P2PMessage& msg);
    void checkPeerTimeout();
    void sendDiscoveryBroadcast();
    std::string generateId();

    std::string agentId_;
    uint16_t listenPort_;
    uint16_t discoveryPort_;

    std::map<std::string, PeerInfo> peers_;
    std::map<P2PMessageType, std::vector<P2PMessageHandler>> handlers_;
    std::function<void(const PeerInfo&)> onPeerOnline_;
    std::function<void(const PeerInfo&)> onPeerOffline_;

    std::atomic<bool> running_{false};
    std::atomic<bool> discovering_{false};
    std::thread tcpThread_;
    std::thread udpThread_;
    std::thread timeoutThread_;
    mutable std::mutex mutex_;

    // 简化实现使用socket句柄存储
    int tcpSocket_ = -1;
    int udpSocket_ = -1;
};

/**
 * @brief P2P发现服务
 */
class P2PDiscovery {
public:
    explicit P2PDiscovery(uint16_t port = 7890);
    ~P2PDiscovery();

    /**
     * @brief 启动发现服务
     */
    bool start();

    /**
     * @brief 停止发现服务
     */
    void stop();

    /**
     * @brief 广播发现消息
     */
    void broadcast(const PeerInfo& selfInfo);

    /**
     * @brief 设置发现回调
     */
    void onDiscovery(std::function<void(const PeerInfo&)> callback);

private:
    void listenLoop();

    uint16_t port_;
    std::atomic<bool> running_{false};
    std::thread thread_;
    std::function<void(const PeerInfo&)> onDiscovery_;
    int socket_ = -1;
};

} // namespace ofa

#endif // OFA_P2P_HPP