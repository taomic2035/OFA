/**
 * @file agent.cpp
 * @brief OFA Agent 核心实现
 */

#include "ofa/agent.h"
#include <cstring>
#include <string>
#include <unordered_map>
#include <vector>
#include <mutex>
#include <atomic>
#include <thread>
#include <chrono>
#include <random>

namespace ofa {

// 内部实现结构
struct AgentImpl {
    // 配置
    std::string name;
    std::string center_addr;
    OFA_AgentType type;
    OFA_OfflineLevel offline_level;
    bool p2p_enabled;
    bool auto_sync;
    int heartbeat_interval_ms;
    void* user_data;

    // 状态
    std::atomic<bool> running{false};
    std::atomic<bool> offline_mode{false};
    std::atomic<bool> online{false};

    // 技能注册表
    std::unordered_map<std::string, OFA_Skill> skills;
    std::mutex skills_mutex;

    // 离线缓存
    std::vector<std::vector<uint8_t>> pending_sync;
    std::mutex sync_mutex;
    size_t cache_size{0};
    size_t cache_max_size{10 * 1024 * 1024}; // 10MB

    // P2P
    OFA_PeerCallback peer_callback{nullptr};
    void* peer_user_data{nullptr};
    OFA_P2PMessageCallback msg_callback{nullptr};
    void* msg_user_data{nullptr};
    std::vector<std::string> discovered_peers;
    std::mutex p2p_mutex;

    // Agent ID
    std::string agent_id;

    AgentImpl() {
        // 生成随机 Agent ID
        std::random_device rd;
        std::mt19937 gen(rd());
        std::uniform_int_distribution<> dis(0, 999999);
        agent_id = "oh-agent-" + std::to_string(dis(gen));
    }
};

} // namespace ofa

// C 接口实现

OFA_Agent* OFA_Agent_Create(const OFA_AgentConfig* config) {
    if (!config) return nullptr;

    auto* impl = new ofa::AgentImpl();
    impl->name = config->name ? config->name : "openharmony-agent";
    impl->center_addr = config->center_addr ? config->center_addr : "";
    impl->type = config->type;
    impl->offline_level = config->offline;
    impl->p2p_enabled = config->p2p_enabled;
    impl->auto_sync = config->auto_sync;
    impl->heartbeat_interval_ms = config->heartbeat_interval_ms;
    impl->user_data = config->user_data;

    // 根据离线等级设置模式
    if (config->offline == OFA_OFFLINE_L1) {
        impl->offline_mode = true;
    }

    return reinterpret_cast<OFA_Agent*>(impl);
}

OFA_Result OFA_Agent_Start(OFA_Agent* agent) {
    if (!agent) return OFA_ERROR;

    auto* impl = reinterpret_cast<ofa::AgentImpl*>(agent);
    if (impl->running) return OFA_ERROR;

    impl->running = true;

    // 如果有 Center 地址且不在离线模式，尝试连接
    if (!impl->center_addr.empty() && !impl->offline_mode) {
        // TODO: 实际连接逻辑
        impl->online = true;
    }

    return OFA_OK;
}

OFA_Result OFA_Agent_Stop(OFA_Agent* agent) {
    if (!agent) return OFA_ERROR;

    auto* impl = reinterpret_cast<ofa::AgentImpl*>(agent);
    impl->running = false;
    impl->online = false;

    return OFA_OK;
}

void OFA_Agent_Destroy(OFA_Agent* agent) {
    if (!agent) return;

    auto* impl = reinterpret_cast<ofa::AgentImpl*>(agent);
    impl->running = false;
    delete impl;
}

OFA_Result OFA_Agent_SetOfflineMode(OFA_Agent* agent, bool offline) {
    if (!agent) return OFA_ERROR;

    auto* impl = reinterpret_cast<ofa::AgentImpl*>(agent);
    impl->offline_mode = offline;

    return OFA_OK;
}

bool OFA_Agent_IsOnline(const OFA_Agent* agent) {
    if (!agent) return false;

    auto* impl = reinterpret_cast<const ofa::AgentImpl*>(agent);
    return impl->online;
}

OFA_OfflineLevel OFA_Agent_GetOfflineLevel(const OFA_Agent* agent) {
    if (!agent) return OFA_OFFLINE_NONE;

    auto* impl = reinterpret_cast<const ofa::AgentImpl*>(agent);
    return impl->offline_level;
}

OFA_Result OFA_Agent_RegisterSkill(OFA_Agent* agent, const OFA_Skill* skill) {
    if (!agent || !skill || !skill->id) return OFA_ERROR;

    auto* impl = reinterpret_cast<ofa::AgentImpl*>(agent);
    std::lock_guard<std::mutex> lock(impl->skills_mutex);

    impl->skills[skill->id] = *skill;

    return OFA_OK;
}

OFA_Result OFA_Agent_ExecuteLocal(
    OFA_Agent* agent,
    const char* skill_id,
    const uint8_t* input,
    size_t input_len,
    uint8_t** output,
    size_t* output_len
) {
    if (!agent || !skill_id || !output || !output_len) return OFA_ERROR;

    auto* impl = reinterpret_cast<ofa::AgentImpl*>(agent);

    std::lock_guard<std::mutex> lock(impl->skills_mutex);

    auto it = impl->skills.find(skill_id);
    if (it == impl->skills.end()) {
        return OFA_ERROR; // 技能不存在
    }

    const OFA_Skill& skill = it->second;

    // 检查离线能力
    if (impl->offline_mode && !skill.offline_capable) {
        return OFA_ERROR_OFFLINE;
    }

    // 执行技能
    if (skill.handler) {
        return skill.handler(input, input_len, output, output_len, skill.user_data);
    }

    return OFA_ERROR;
}

OFA_Result OFA_Agent_SubmitTask(
    OFA_Agent* agent,
    const char* skill_id,
    const char* target_agent,
    const uint8_t* input,
    size_t input_len,
    char** task_id
) {
    if (!agent || !skill_id) return OFA_ERROR;

    auto* impl = reinterpret_cast<ofa::AgentImpl*>(agent);

    // 需要在线
    if (!impl->online && !impl->offline_mode) {
        return OFA_ERROR_OFFLINE;
    }

    // TODO: 实际提交到 Center
    if (task_id) {
        std::random_device rd;
        std::mt19937 gen(rd());
        std::uniform_int_distribution<> dis(0, 999999);
        std::string tid = "task-" + std::to_string(dis(gen));
        *task_id = strdup(tid.c_str());
    }

    return OFA_OK;
}

// P2P 实现

OFA_Result OFA_P2P_StartDiscovery(
    OFA_Agent* agent,
    OFA_PeerCallback callback,
    void* user_data
) {
    if (!agent) return OFA_ERROR;

    auto* impl = reinterpret_cast<ofa::AgentImpl*>(agent);

    std::lock_guard<std::mutex> lock(impl->p2p_mutex);
    impl->peer_callback = callback;
    impl->peer_user_data = user_data;

    // TODO: 实际设备发现 (通过分布式软总线)

    return OFA_OK;
}

void OFA_P2P_StopDiscovery(OFA_Agent* agent) {
    if (!agent) return;

    auto* impl = reinterpret_cast<ofa::AgentImpl*>(agent);
    std::lock_guard<std::mutex> lock(impl->p2p_mutex);
    impl->peer_callback = nullptr;
}

OFA_Result OFA_P2P_Send(
    OFA_Agent* agent,
    const char* peer_id,
    const uint8_t* data,
    size_t data_len
) {
    if (!agent || !peer_id || !data) return OFA_ERROR;

    // TODO: 实际发送逻辑

    return OFA_OK;
}

OFA_Result OFA_P2P_Broadcast(
    OFA_Agent* agent,
    const uint8_t* data,
    size_t data_len
) {
    if (!agent || !data) return OFA_ERROR;

    auto* impl = reinterpret_cast<ofa::AgentImpl*>(agent);

    std::lock_guard<std::mutex> lock(impl->p2p_mutex);
    for (const auto& peer : impl->discovered_peers) {
        // TODO: 发送到每个设备
    }

    return OFA_OK;
}

void OFA_P2P_SetMessageCallback(
    OFA_Agent* agent,
    OFA_P2PMessageCallback callback,
    void* user_data
) {
    if (!agent) return;

    auto* impl = reinterpret_cast<ofa::AgentImpl*>(agent);
    std::lock_guard<std::mutex> lock(impl->p2p_mutex);
    impl->msg_callback = callback;
    impl->msg_user_data = user_data;
}

// 约束检查

OFA_ConstraintResult OFA_Constraint_Check(
    OFA_Agent* agent,
    const char* action,
    const uint8_t* data,
    size_t data_len
) {
    OFA_ConstraintResult result;
    result.allowed = true;
    result.violated = OFA_CONSTRAINT_NONE;
    result.reason = nullptr;

    if (!agent || !action) {
        result.allowed = false;
        return result;
    }

    // 检查敏感操作
    // TODO: 实现完整约束检查

    // 检查支付相关
    if (strstr(action, "payment") != nullptr) {
        result.allowed = false;
        result.violated = OFA_CONSTRAINT_FINANCIAL;
        result.reason = "Payment operations require online mode";
    }

    // 检查隐私数据
    if (data && data_len > 0) {
        // 简单检查身份证号模式
        std::string data_str(reinterpret_cast<const char*>(data), data_len);
        if (data_str.find("\"idcard\"") != std::string::npos ||
            data_str.find("\"id_card\"") != std::string::npos) {
            result.allowed = false;
            result.violated = OFA_CONSTRAINT_PRIVACY;
            result.reason = "Data contains sensitive personal information";
        }
    }

    return result;
}

// 工具函数

void OFA_Free(void* ptr) {
    if (ptr) {
        free(ptr);
    }
}

const char* OFA_GetVersion(void) {
    return OFA_SDK_VERSION;
}

const char* OFA_GetErrorString(OFA_Result result) {
    switch (result) {
        case OFA_OK: return "Success";
        case OFA_ERROR: return "General error";
        case OFA_ERROR_OFFLINE: return "Operation requires online mode";
        case OFA_ERROR_TIMEOUT: return "Operation timed out";
        case OFA_ERROR_CONSTRAINT: return "Operation violates constraints";
        case OFA_ERROR_UNAUTHORIZED: return "Unauthorized operation";
        default: return "Unknown error";
    }
}