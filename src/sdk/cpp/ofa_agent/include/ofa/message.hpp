/**
 * @file message.hpp
 * @brief Message Types
 * Sprint 29: C++ Agent SDK
 */

#ifndef OFA_MESSAGE_HPP
#define OFA_MESSAGE_HPP

#include <string>
#include "types.hpp"

namespace ofa {

/**
 * @brief 消息类型
 */
enum class MessageType {
    Register,
    Heartbeat,
    Task,
    TaskResult,
    Message,
    Broadcast,
    Discovery,
    Error,
    Ack
};

/**
 * @brief 消息
 */
struct Message {
    std::string id;
    MessageType type;
    std::string from;
    std::string to;
    std::string subject;
    json data;
    int64_t timestamp;

    Message() : type(MessageType::Message), timestamp(0) {}

    json toJson() const {
        return {
            {"id", id},
            {"type", typeToString(type)},
            {"from", from},
            {"to", to},
            {"subject", subject},
            {"data", data},
            {"timestamp", timestamp}
        };
    }

    static Message fromJson(const json& j) {
        Message msg;
        msg.id = j.value("id", "");
        msg.type = stringToType(j.value("type", "message"));
        msg.from = j.value("from", "");
        msg.to = j.value("to", "");
        msg.subject = j.value("subject", "");
        msg.data = j.value("data", json::object());
        msg.timestamp = j.value("timestamp", 0LL);
        return msg;
    }

    static std::string typeToString(MessageType type) {
        switch (type) {
            case MessageType::Register: return "register";
            case MessageType::Heartbeat: return "heartbeat";
            case MessageType::Task: return "task";
            case MessageType::TaskResult: return "task_result";
            case MessageType::Message: return "message";
            case MessageType::Broadcast: return "broadcast";
            case MessageType::Discovery: return "discovery";
            case MessageType::Error: return "error";
            case MessageType::Ack: return "ack";
            default: return "unknown";
        }
    }

    static MessageType stringToType(const std::string& str) {
        if (str == "register") return MessageType::Register;
        if (str == "heartbeat") return MessageType::Heartbeat;
        if (str == "task") return MessageType::Task;
        if (str == "task_result") return MessageType::TaskResult;
        if (str == "message") return MessageType::Message;
        if (str == "broadcast") return MessageType::Broadcast;
        if (str == "discovery") return MessageType::Discovery;
        if (str == "error") return MessageType::Error;
        if (str == "ack") return MessageType::Ack;
        return MessageType::Message;
    }
};

/**
 * @brief 任务
 */
struct Task {
    std::string taskId;
    std::string skillId;
    std::string operation;
    json input;
    int32_t priority = 0;
    uint32_t timeout = 30;

    json toJson() const {
        return {
            {"task_id", taskId},
            {"skill_id", skillId},
            {"operation", operation},
            {"input", input},
            {"priority", priority},
            {"timeout", timeout}
        };
    }

    static Task fromJson(const json& j) {
        Task task;
        task.taskId = j.value("task_id", "");
        task.skillId = j.value("skill_id", "");
        task.operation = j.value("operation", "");
        task.input = j.value("input", json::object());
        task.priority = j.value("priority", 0);
        task.timeout = j.value("timeout", 30u);
        return task;
    }
};

/**
 * @brief 任务结果
 */
struct TaskResult {
    std::string taskId;
    bool success = false;
    json result;
    std::string error;
    std::string agentId;
    uint64_t durationMs = 0;

    json toJson() const {
        json j = {
            {"task_id", taskId},
            {"success", success},
            {"agent_id", agentId},
            {"duration_ms", durationMs}
        };
        if (success && !result.is_null()) {
            j["result"] = result;
        }
        if (!success && !error.empty()) {
            j["error"] = error;
        }
        return j;
    }
};

} // namespace ofa

#endif // OFA_MESSAGE_HPP