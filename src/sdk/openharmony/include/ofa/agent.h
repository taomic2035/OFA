/**
 * @file agent.h
 * @brief OFA Agent SDK for OpenHarmony
 *
 * OpenHarmony 设备的 OFA Agent SDK，支持离线运行和 P2P 通信
 */

#ifndef OFA_AGENT_H
#define OFA_AGENT_H

#include <stdint.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

/* 版本信息 */
#define OFA_SDK_VERSION_MAJOR 0
#define OFA_SDK_VERSION_MINOR 9
#define OFA_SDK_VERSION_PATCH 0
#define OFA_SDK_VERSION "0.9.0"

/* Agent 类型 */
typedef enum {
    OFA_AGENT_TYPE_FULL = 0,    /* 全功能 Agent */
    OFA_AGENT_TYPE_LITE = 1,    /* 轻量 Agent (手表等) */
    OFA_AGENT_TYPE_EDGE = 2,    /* 边缘 Agent (IoT) */
    OFA_AGENT_TYPE_SENSOR = 3   /* 传感器 Agent */
} OFA_AgentType;

/* 离线能力等级 */
typedef enum {
    OFA_OFFLINE_NONE = 0,       /* 不支持离线 */
    OFA_OFFLINE_L1 = 1,         /* 完全离线 (本地执行) */
    OFA_OFFLINE_L2 = 2,         /* 局域网协作 */
    OFA_OFFLINE_L3 = 3,         /* 弱网同步 */
    OFA_OFFLINE_L4 = 4          /* 在线模式 */
} OFA_OfflineLevel;

/* Agent 配置 */
typedef struct {
    const char* name;           /* Agent 名称 */
    const char* center_addr;    /* Center 地址 (可选) */
    OFA_AgentType type;         /* Agent 类型 */
    OFA_OfflineLevel offline;   /* 离线能力等级 */
    bool p2p_enabled;           /* 启用 P2P */
    bool auto_sync;             /* 自动同步 */
    int heartbeat_interval_ms;  /* 心跳间隔 */
    void* user_data;            /* 用户数据 */
} OFA_AgentConfig;

/* Agent 句柄 */
typedef struct OFA_Agent OFA_Agent;

/* 结果状态 */
typedef enum {
    OFA_OK = 0,
    OFA_ERROR = -1,
    OFA_ERROR_OFFLINE = -2,
    OFA_ERROR_TIMEOUT = -3,
    OFA_ERROR_CONSTRAINT = -4,
    OFA_ERROR_UNAUTHORIZED = -5
} OFA_Result;

/* ==================== Agent 生命周期 ==================== */

/**
 * 创建 Agent 实例
 * @param config 配置参数
 * @return Agent 句柄，失败返回 NULL
 */
OFA_Agent* OFA_Agent_Create(const OFA_AgentConfig* config);

/**
 * 启动 Agent
 * @param agent Agent 句柄
 * @return OFA_OK 成功
 */
OFA_Result OFA_Agent_Start(OFA_Agent* agent);

/**
 * 停止 Agent
 * @param agent Agent 句柄
 * @return OFA_OK 成功
 */
OFA_Result OFA_Agent_Stop(OFA_Agent* agent);

/**
 * 销毁 Agent
 * @param agent Agent 句柄
 */
void OFA_Agent_Destroy(OFA_Agent* agent);

/* ==================== 离线模式 ==================== */

/**
 * 设置离线模式
 * @param agent Agent 句柄
 * @param offline 是否离线
 * @return OFA_OK 成功
 */
OFA_Result OFA_Agent_SetOfflineMode(OFA_Agent* agent, bool offline);

/**
 * 检查是否在线
 * @param agent Agent 句柄
 * @return true 在线
 */
bool OFA_Agent_IsOnline(const OFA_Agent* agent);

/**
 * 获取当前离线等级
 * @param agent Agent 句柄
 * @return 离线等级
 */
OFA_OfflineLevel OFA_Agent_GetOfflineLevel(const OFA_Agent* agent);

/* ==================== 任务执行 ==================== */

/* 任务句柄 */
typedef struct OFA_Task OFA_Task;

/* 技能处理器 */
typedef OFA_Result (*OFA_SkillHandler)(
    const uint8_t* input,
    size_t input_len,
    uint8_t** output,
    size_t* output_len,
    void* user_data
);

/* 技能定义 */
typedef struct {
    const char* id;             /* 技能 ID */
    const char* name;           /* 技能名称 */
    const char* category;       /* 分类 */
    bool offline_capable;       /* 是否支持离线 */
    OFA_SkillHandler handler;   /* 处理函数 */
    void* user_data;            /* 用户数据 */
} OFA_Skill;

/**
 * 注册技能
 * @param agent Agent 句柄
 * @param skill 技能定义
 * @return OFA_OK 成功
 */
OFA_Result OFA_Agent_RegisterSkill(OFA_Agent* agent, const OFA_Skill* skill);

/**
 * 本地执行任务 (支持离线)
 * @param agent Agent 句柄
 * @param skill_id 技能 ID
 * @param input 输入数据
 * @param input_len 输入长度
 * @param output 输出数据 (需调用 OFA_Free 释放)
 * @param output_len 输出长度
 * @return OFA_OK 成功
 */
OFA_Result OFA_Agent_ExecuteLocal(
    OFA_Agent* agent,
    const char* skill_id,
    const uint8_t* input,
    size_t input_len,
    uint8_t** output,
    size_t* output_len
);

/**
 * 提交任务到 Center (需要在线)
 * @param agent Agent 句柄
 * @param skill_id 技能 ID
 * @param target_agent 目标 Agent (可选)
 * @param input 输入数据
 * @param input_len 输入长度
 * @param task_id 返回任务 ID
 * @return OFA_OK 成功
 */
OFA_Result OFA_Agent_SubmitTask(
    OFA_Agent* agent,
    const char* skill_id,
    const char* target_agent,
    const uint8_t* input,
    size_t input_len,
    char** task_id
);

/* ==================== P2P 通信 ==================== */

/* P2P 消息类型 */
typedef enum {
    OFA_P2P_MSG_DATA = 0,       /* 数据消息 */
    OFA_P2P_MSG_BROADCAST = 1,  /* 广播消息 */
    OFA_P2P_MSG_REQUEST = 2,    /* 请求消息 */
    OFA_P2P_MSG_RESPONSE = 3    /* 响应消息 */
} OFA_P2PMsgType;

/* P2P 消息 */
typedef struct {
    OFA_P2PMsgType type;        /* 消息类型 */
    const char* from;           /* 发送方 */
    const char* to;             /* 接收方 */
    const uint8_t* data;        /* 数据 */
    size_t data_len;            /* 数据长度 */
} OFA_P2PMessage;

/* P2P 消息回调 */
typedef void (*OFA_P2PMessageCallback)(
    const OFA_P2PMessage* msg,
    void* user_data
);

/* P2P 发现回调 */
typedef void (*OFA_PeerCallback)(
    const char* peer_id,
    const char* peer_name,
    bool added,
    void* user_data
);

/**
 * 启动 P2P 设备发现
 * @param agent Agent 句柄
 * @param callback 发现回调
 * @param user_data 用户数据
 * @return OFA_OK 成功
 */
OFA_Result OFA_P2P_StartDiscovery(
    OFA_Agent* agent,
    OFA_PeerCallback callback,
    void* user_data
);

/**
 * 停止 P2P 设备发现
 * @param agent Agent 句柄
 */
void OFA_P2P_StopDiscovery(OFA_Agent* agent);

/**
 * 发送 P2P 消息
 * @param agent Agent 句柄
 * @param peer_id 目标设备 ID
 * @param data 数据
 * @param data_len 数据长度
 * @return OFA_OK 成功
 */
OFA_Result OFA_P2P_Send(
    OFA_Agent* agent,
    const char* peer_id,
    const uint8_t* data,
    size_t data_len
);

/**
 * 广播消息到所有附近设备
 * @param agent Agent 句柄
 * @param data 数据
 * @param data_len 数据长度
 * @return OFA_OK 成功
 */
OFA_Result OFA_P2P_Broadcast(
    OFA_Agent* agent,
    const uint8_t* data,
    size_t data_len
);

/**
 * 设置 P2P 消息监听
 * @param agent Agent 句柄
 * @param callback 消息回调
 * @param user_data 用户数据
 */
void OFA_P2P_SetMessageCallback(
    OFA_Agent* agent,
    OFA_P2PMessageCallback callback,
    void* user_data
);

/* ==================== 约束检查 ==================== */

/* 约束类型 */
typedef enum {
    OFA_CONSTRAINT_NONE = 0,
    OFA_CONSTRAINT_PRIVACY = 1,     /* 隐私保护 */
    OFA_CONSTRAINT_FINANCIAL = 2,   /* 财产相关 */
    OFA_CONSTRAINT_SECURITY = 4,    /* 安全敏感 */
    OFA_CONSTRAINT_AUTH_REQUIRED = 8 /* 需要授权 */
} OFA_ConstraintType;

/* 约束检查结果 */
typedef struct {
    bool allowed;                /* 是否允许 */
    OFA_ConstraintType violated; /* 违反的约束 */
    const char* reason;          /* 原因说明 */
} OFA_ConstraintResult;

/**
 * 检查交互约束
 * @param agent Agent 句柄
 * @param action 操作类型
 * @param data 数据
 * @param data_len 数据长度
 * @return 约束检查结果
 */
OFA_ConstraintResult OFA_Constraint_Check(
    OFA_Agent* agent,
    const char* action,
    const uint8_t* data,
    size_t data_len
);

/* ==================== 工具函数 ==================== */

/**
 * 释放内存
 * @param ptr 指针
 */
void OFA_Free(void* ptr);

/**
 * 获取版本字符串
 * @return 版本字符串
 */
const char* OFA_GetVersion(void);

/**
 * 获取错误描述
 * @param result 错误码
 * @return 错误描述
 */
const char* OFA_GetErrorString(OFA_Result result);

#ifdef __cplusplus
}
#endif

#endif /* OFA_AGENT_H */