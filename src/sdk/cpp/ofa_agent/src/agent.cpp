/**
 * @file agent.cpp
 * @brief Agent Implementation
 * Sprint 29: C++ Agent SDK
 */

#include "ofa/agent.hpp"
#include "ofa/error.hpp"
#include <chrono>
#include <thread>
#include <random>
#include <sstream>

namespace ofa {

// 生成UUID
static std::string generateUUID() {
    std::random_device rd;
    std::mt19937 gen(rd());
    std::uniform_int_distribution<> dis(0, 15);
    std::uniform_int_distribution<> dis2(8, 11);

    std::stringstream ss;
    ss << std::hex;
    for (int i = 0; i < 8; i++) ss << dis(gen);
    ss << "-";
    for (int i = 0; i < 4; i++) ss << dis(gen);
    ss << "-4"; // UUID v4
    for (int i = 0; i < 3; i++) ss << dis(gen);
    ss << "-";
    ss << dis2(gen); // 8, 9, a, or b
    for (int i = 0; i < 3; i++) ss << dis(gen);
    ss << "-";
    for (int i = 0; i < 12; i++) ss << dis(gen);
    return ss.str();
}

Agent::Agent(const AgentConfig& config)
    : config_(config)
    , skillExecutor_(std::make_shared<SkillExecutor>())
    , skillRegistry_(std::make_shared<SkillRegistry>())
{
    info_.id = config_.agentId.empty() ? "cpp-agent-" + generateUUID().substr(0, 8) : config_.agentId;
    info_.name = config_.name;
    info_.type = "cpp";
    info_.version = Version::STRING;
    info_.platform = "cpp";
    info_.skills = config_.skills;
    info_.state = AgentState::Initializing;
    info_.lastHeartbeat = 0;
}

Agent::~Agent() {
    if (running_) {
        running_ = false;
    }
}

std::future<bool> Agent::connect() {
    return std::async(std::launch::async, [this]() {
        state_ = AgentState::Connecting;

        try {
            connection_ = createConnection(config_);
            connection_->connect(config_).get();

            // 注册
            Message msg;
            msg.id = generateUUID();
            msg.type = MessageType::Register;
            msg.from = info_.id;
            msg.data = {
                {"id", info_.id},
                {"name", info_.name},
                {"type", info_.type},
                {"version", info_.version},
                {"platform", info_.platform},
                {"skills", info_.skills}
            };
            msg.timestamp = std::chrono::duration_cast<std::chrono::milliseconds>(
                std::chrono::system_clock::now().time_since_epoch()
            ).count();

            connection_->send(msg).get();

            state_ = AgentState::Online;
            running_ = true;

            // 启动心跳线程
            std::thread(&Agent::heartbeatLoop, this).detach();

            return true;
        } catch (const std::exception& e) {
            state_ = AgentState::Error;
            return false;
        }
    });
}

std::future<void> Agent::disconnect() {
    running_ = false;

    return std::async(std::launch::async, [this]() {
        if (connection_) {
            connection_->disconnect().get();
        }
        state_ = AgentState::Offline;
    });
}

void Agent::registerSkill(const std::string& skillId, SkillHandler handler) {
    skillExecutor_->registerSkill(skillId, handler);
    info_.skills.push_back(skillId);
}

void Agent::unregisterSkill(const std::string& skillId) {
    skillExecutor_->unregisterSkill(skillId);
    auto it = std::find(info_.skills.begin(), info_.skills.end(), skillId);
    if (it != info_.skills.end()) {
        info_.skills.erase(it);
    }
}

std::future<TaskResult> Agent::executeTask(const Task& task) {
    return std::async(std::launch::async, [this, task]() {
        state_ = AgentState::Busy;

        auto start = std::chrono::steady_clock::now();
        TaskResult result;
        result.taskId = task.taskId;
        result.agentId = info_.id;

        try {
            result.result = skillExecutor_->execute(task.skillId, task.operation, task.input);
            result.success = true;
            stats_.tasksSuccess++;
        } catch (const std::exception& e) {
            result.error = e.what();
            result.success = false;
            stats_.tasksFailed++;
        }

        auto end = std::chrono::steady_clock::now();
        result.durationMs = std::chrono::duration_cast<std::chrono::milliseconds>(end - start).count();

        stats_.tasksExecuted++;
        state_ = AgentState::Online;

        return result;
    });
}

std::future<void> Agent::sendMessage(const std::string& target, MessageType type, const json& data) {
    return std::async(std::launch::async, [this, target, type, data]() {
        if (!connection_) throw ConnectionException("Not connected");

        Message msg;
        msg.id = generateUUID();
        msg.type = type;
        msg.from = info_.id;
        msg.to = target;
        msg.data = data;
        msg.timestamp = std::chrono::duration_cast<std::chrono::milliseconds>(
            std::chrono::system_clock::now().time_since_epoch()
        ).count();

        connection_->send(msg).get();
        stats_.messagesSent++;
    });
}

std::future<void> Agent::broadcast(MessageType type, const json& data) {
    return sendMessage("", type, data);
}

void Agent::onMessage(const std::string& type, MessageHandler handler) {
    std::lock_guard<std::mutex> lock(mutex_);
    messageHandlers_[type].push_back(handler);
}

std::future<void> Agent::run() {
    return std::async(std::launch::async, [this]() {
        connect().get();

        while (running_ && connection_) {
            try {
                auto msg = connection_->receive().get();
                if (msg) {
                    handleMessage(*msg);
                    stats_.messagesReceived++;
                }
            } catch (const std::exception& e) {
                // 错误处理
            }

            std::this_thread::sleep_for(std::chrono::milliseconds(100));
        }
    });
}

AgentStats Agent::stats() const {
    return stats_;
}

void Agent::heartbeatLoop() {
    while (running_) {
        std::this_thread::sleep_for(std::chrono::seconds(config_.heartbeatInterval));

        if (!running_ || !connection_) break;

        try {
            Message msg;
            msg.id = generateUUID();
            msg.type = MessageType::Heartbeat;
            msg.from = info_.id;
            msg.data = {{"state", Message::typeToString(MessageType::Heartbeat)}};
            msg.timestamp = std::chrono::duration_cast<std::chrono::milliseconds>(
                std::chrono::system_clock::now().time_since_epoch()
            ).count();

            connection_->send(msg).get();
            info_.lastHeartbeat = msg.timestamp;
        } catch (const std::exception& e) {
            // 重连逻辑
        }
    }
}

void Agent::handleMessage(const Message& msg) {
    if (msg.type == MessageType::Task) {
        Task task = Task::fromJson(msg.data);
        auto result = executeTask(task).get();

        Message response;
        response.id = generateUUID();
        response.type = MessageType::TaskResult;
        response.from = info_.id;
        response.data = result.toJson();

        if (connection_) {
            connection_->send(response).get();
        }
    }

    // 自定义消息处理
    std::lock_guard<std::mutex> lock(mutex_);
    std::string typeStr = Message::typeToString(msg.type);
    auto it = messageHandlers_.find(typeStr);
    if (it != messageHandlers_.end()) {
        for (auto& handler : it->second) {
            handler(msg.data);
        }
    }
}

} // namespace ofa