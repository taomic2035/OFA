/**
 * @file agent.hpp
 * @brief OFA C++ Agent SDK
 * @version 8.1.0
 * Sprint 29: Multi-platform SDK Extension
 */

#ifndef OFA_AGENT_HPP
#define OFA_AGENT_HPP

#include <string>
#include <memory>
#include <functional>
#include <vector>
#include <map>
#include <atomic>
#include <mutex>
#include <future>

#include "config.hpp"
#include "skills.hpp"
#include "connection.hpp"
#include "message.hpp"
#include "types.hpp"

namespace ofa {

/**
 * @brief Agent状态
 */
enum class AgentState {
    Initializing,
    Connecting,
    Online,
    Busy,
    Offline,
    Error
};

/**
 * @brief Agent信息
 */
struct AgentInfo {
    std::string id;
    std::string name;
    std::string type;
    std::string version;
    std::string platform;
    std::vector<std::string> skills;
    AgentState state;
    std::map<std::string, std::string> metadata;
    int64_t lastHeartbeat;
};

/**
 * @brief Agent统计
 */
struct AgentStats {
    uint64_t tasksExecuted = 0;
    uint64_t tasksSuccess = 0;
    uint64_t tasksFailed = 0;
    uint64_t messagesSent = 0;
    uint64_t messagesReceived = 0;
};

/**
 * @brief OFA Agent
 */
class Agent {
public:
    /**
     * @brief 构造函数
     */
    explicit Agent(const AgentConfig& config);

    /**
     * @brief 析构函数
     */
    ~Agent();

    // 禁止拷贝
    Agent(const Agent&) = delete;
    Agent& operator=(const Agent&) = delete;

    /**
     * @brief 获取Agent ID
     */
    const std::string& id() const { return info_.id; }

    /**
     * @brief 获取状态
     */
    AgentState state() const { return state_; }

    /**
     * @brief 是否在线
     */
    bool isOnline() const { return state_ == AgentState::Online; }

    /**
     * @brief 获取信息
     */
    const AgentInfo& info() const { return info_; }

    /**
     * @brief 连接到Center
     */
    std::future<bool> connect();

    /**
     * @brief 断开连接
     */
    std::future<void> disconnect();

    /**
     * @brief 注册技能
     */
    void registerSkill(const std::string& skillId, SkillHandler handler);

    /**
     * @brief 注销技能
     */
    void unregisterSkill(const std::string& skillId);

    /**
     * @brief 执行任务
     */
    std::future<TaskResult> executeTask(const Task& task);

    /**
     * @brief 发送消息
     */
    std::future<void> sendMessage(const std::string& target,
                                   MessageType type,
                                   const json& data);

    /**
     * @brief 广播消息
     */
    std::future<void> broadcast(MessageType type, const json& data);

    /**
     * @brief 注册消息处理器
     */
    void onMessage(const std::string& type, MessageHandler handler);

    /**
     * @brief 运行Agent
     */
    std::future<void> run();

    /**
     * @brief 获取统计
     */
    AgentStats stats() const;

private:
    void heartbeatLoop();
    void messageLoop();
    void handleMessage(const Message& msg);

    AgentConfig config_;
    AgentInfo info_;
    std::atomic<AgentState> state_{AgentState::Initializing};
    std::unique_ptr<Connection> connection_;
    std::shared_ptr<SkillExecutor> skillExecutor_;
    std::shared_ptr<SkillRegistry> skillRegistry_;
    AgentStats stats_;
    std::atomic<bool> running_{false};

    std::map<std::string, std::vector<MessageHandler>> messageHandlers_;
    mutable std::mutex mutex_;
};

} // namespace ofa

#endif // OFA_AGENT_HPP