/**
 * @file agent_napi.cpp
 * @brief OFA Agent NAPI 接口层 - OpenHarmony ArkTS/JS 绑定
 */

#include "ofa/agent.h"
#include <napi/native_api.h>
#include <napi/native_node_api.h>

// NAPI 辅助函数
static napi_value CreateNAPIError(napi_env env, const char* message) {
    napi_value error;
    napi_create_string_utf8(env, message, NAPI_AUTO_LENGTH, &error);
    return error;
}

static napi_value CreateNAPIUndefined(napi_env env) {
    napi_value undefined;
    napi_get_undefined(env, &undefined);
    return undefined;
}

static napi_value CreateNAPIBoolean(napi_env env, bool value) {
    napi_value result;
    napi_get_boolean(env, value, &result);
    return result;
}

static napi_value CreateNAPINumber(napi_env env, int32_t value) {
    napi_value result;
    napi_create_int32(env, value, &result);
    return result;
}

static napi_value CreateNAPIString(napi_env env, const char* str) {
    napi_value result;
    napi_create_string_utf8(env, str, NAPI_AUTO_LENGTH, &result);
    return result;
}

// Agent 类包装
class OFAAgentWrapper {
public:
    OFA_Agent* agent = nullptr;
    OFA_LocalScheduler* scheduler = nullptr;
    OFA_OfflineCache* cache = nullptr;

    ~OFAAgentWrapper() {
        if (agent) {
            OFA_Agent_Stop(agent);
            OFA_Agent_Destroy(agent);
        }
        if (scheduler) {
            OFA_LocalScheduler_Destroy(scheduler);
        }
        if (cache) {
            OFA_OfflineCache_Destroy(cache);
        }
    }
};

// 创建 Agent
static napi_value NAPI_Agent_Create(napi_env env, napi_callback_info info) {
    size_t argc = 1;
    napi_value args[1];
    napi_get_cb_info(env, info, &argc, args, nullptr, nullptr);

    if (argc < 1) {
        napi_throw_error(env, nullptr, "Config object required");
        return CreateNAPIUndefined(env);
    }

    // 解析配置对象
    napi_value config = args[0];

    OFA_AgentConfig agent_config = {};

    // name
    napi_value name_val;
    if (napi_get_named_property(env, config, "name", &name_val) == napi_ok) {
        size_t name_len;
        napi_get_value_string_utf8(env, name_val, nullptr, 0, &name_len);
        char* name = new char[name_len + 1];
        napi_get_value_string_utf8(env, name_val, name, name_len + 1, &name_len);
        agent_config.name = name;
    }

    // centerAddr
    napi_value center_val;
    if (napi_get_named_property(env, config, "centerAddr", &center_val) == napi_ok) {
        size_t center_len;
        napi_get_value_string_utf8(env, center_val, nullptr, 0, &center_len);
        char* center = new char[center_len + 1];
        napi_get_value_string_utf8(env, center_val, center, center_len + 1, &center_len);
        agent_config.center_addr = center;
    }

    // type
    napi_value type_val;
    if (napi_get_named_property(env, config, "type", &type_val) == napi_ok) {
        int32_t type;
        napi_get_value_int32(env, type_val, &type);
        agent_config.type = static_cast<OFA_AgentType>(type);
    }

    // offlineLevel
    napi_value offline_val;
    if (napi_get_named_property(env, config, "offlineLevel", &offline_val) == napi_ok) {
        int32_t offline;
        napi_get_value_int32(env, offline_val, &offline);
        agent_config.offline = static_cast<OFA_OfflineLevel>(offline);
    }

    // p2pEnabled
    napi_value p2p_val;
    if (napi_get_named_property(env, config, "p2pEnabled", &p2p_val) == napi_ok) {
        bool p2p;
        napi_get_value_bool(env, p2p_val, &p2p);
        agent_config.p2p_enabled = p2p;
    }

    // autoSync
    napi_value sync_val;
    if (napi_get_named_property(env, config, "autoSync", &sync_val) == napi_ok) {
        bool sync;
        napi_get_value_bool(env, sync_val, &sync);
        agent_config.auto_sync = sync;
    }

    // 创建 Agent
    OFA_Agent* agent = OFA_Agent_Create(&agent_config);
    if (!agent) {
        napi_throw_error(env, nullptr, "Failed to create agent");
        return CreateNAPIUndefined(env);
    }

    // 创建包装对象
    auto* wrapper = new OFAAgentWrapper();
    wrapper->agent = agent;

    // 如果支持离线，创建本地调度器和缓存
    if (agent_config.offline >= OFA_OFFLINE_L1) {
        wrapper->scheduler = OFA_LocalScheduler_Create(4, agent_config.offline);
        wrapper->cache = OFA_OfflineCache_Create(10 * 1024 * 1024);
    }

    // 创建 JS 对象
    napi_value obj;
    napi_create_object(env, &obj);

    // 存储 wrapper 指针
    napi_value external;
    napi_create_external(env, wrapper, [](napi_env, void* data, void*) {
        delete static_cast<OFAAgentWrapper*>(data);
    }, nullptr, &external);
    napi_set_named_property(env, obj, "_internal", external);

    // 绑定方法
    napi_set_named_property(env, obj, "start", nullptr); // 后续绑定
    napi_set_named_property(env, obj, "stop", nullptr);
    napi_set_named_property(env, obj, "isOnline", nullptr);
    napi_set_named_property(env, obj, "executeLocal", nullptr);

    // 清理临时字符串
    if (agent_config.name) delete[] agent_config.name;
    if (agent_config.center_addr) delete[] agent_config.center_addr;

    return obj;
}

// 启动 Agent
static napi_value NAPI_Agent_Start(napi_env env, napi_callback_info info) {
    // TODO: 实现完整版本
    return CreateNAPIBoolean(env, true);
}

// 停止 Agent
static napi_value NAPI_Agent_Stop(napi_env env, napi_callback_info info) {
    // TODO: 实现完整版本
    return CreateNAPIBoolean(env, true);
}

// 执行本地任务
static napi_value NAPI_Agent_ExecuteLocal(napi_env env, napi_callback_info info) {
    size_t argc = 2;
    napi_value args[2];
    napi_get_cb_info(env, info, &argc, args, nullptr, nullptr);

    if (argc < 2) {
        napi_throw_error(env, nullptr, "skillId and input required");
        return CreateNAPIUndefined(env);
    }

    // 获取 skillId
    size_t skill_len;
    napi_get_value_string_utf8(env, args[0], nullptr, 0, &skill_len);
    char* skill_id = new char[skill_len + 1];
    napi_get_value_string_utf8(env, args[0], skill_id, skill_len + 1, &skill_len);

    // 获取 input (假设是 ArrayBuffer)
    uint8_t* input_data = nullptr;
    size_t input_len = 0;
    napi_get_arraybuffer_info(env, args[1], reinterpret_cast<void**>(&input_data), &input_len);

    // TODO: 从 this 获取 wrapper 并执行
    uint8_t* output = nullptr;
    size_t output_len = 0;

    // 创建输出 ArrayBuffer
    napi_value output_buffer;
    // napi_create_arraybuffer(env, output_len, reinterpret_cast<void**>(&output), &output_buffer);

    delete[] skill_id;
    return CreateNAPIUndefined(env);
}

// 获取版本
static napi_value NAPI_GetVersion(napi_env env, napi_callback_info info) {
    return CreateNAPIString(env, OFA_GetVersion());
}

// 离线等级常量
static napi_value DefineOfflineLevels(napi_env env) {
    napi_value obj;
    napi_create_object(env, &obj);

    napi_set_named_property(env, obj, "NONE", CreateNAPINumber(env, OFA_OFFLINE_NONE));
    napi_set_named_property(env, obj, "L1", CreateNAPINumber(env, OFA_OFFLINE_L1));
    napi_set_named_property(env, obj, "L2", CreateNAPINumber(env, OFA_OFFLINE_L2));
    napi_set_named_property(env, obj, "L3", CreateNAPINumber(env, OFA_OFFLINE_L3));
    napi_set_named_property(env, obj, "L4", CreateNAPINumber(env, OFA_OFFLINE_L4));

    return obj;
}

// Agent 类型常量
static napi_value DefineAgentTypes(napi_env env) {
    napi_value obj;
    napi_create_object(env, &obj);

    napi_set_named_property(env, obj, "FULL", CreateNAPINumber(env, OFA_AGENT_TYPE_FULL));
    napi_set_named_property(env, obj, "LITE", CreateNAPINumber(env, OFA_AGENT_TYPE_LITE));
    napi_set_named_property(env, obj, "EDGE", CreateNAPINumber(env, OFA_AGENT_TYPE_EDGE));
    napi_set_named_property(env, obj, "SENSOR", CreateNAPINumber(env, OFA_AGENT_TYPE_SENSOR));

    return obj;
}

// 模块导出
EXTERN_C_START
static napi_value Init(napi_env env, napi_value exports) {
    // 导出函数
    napi_property_descriptor desc[] = {
        {"createAgent", nullptr, NAPI_Agent_Create, nullptr, nullptr, nullptr, napi_default, nullptr},
        {"getVersion", nullptr, NAPI_GetVersion, nullptr, nullptr, nullptr, napi_default, nullptr},
    };

    napi_define_properties(env, exports, sizeof(desc) / sizeof(desc[0]), desc);

    // 导出常量
    napi_set_named_property(env, exports, "OfflineLevel", DefineOfflineLevels(env));
    napi_set_named_property(env, exports, "AgentType", DefineAgentTypes(env));

    return exports;
}
EXTERN_C_END

// 模块定义
static napi_module module = {
    .nm_version = 1,
    .nm_flags = NAPI_MODULE_VERSION,
    .nm_filename = nullptr,
    .nm_register_func = Init,
    .nm_modname = "ofa",
    .nm_priv = nullptr,
    .reserved = {nullptr}
};

NAPI_MODULE(ofa, module)