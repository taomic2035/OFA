//! Message Module
//! Sprint 29: Rust Agent SDK

use serde::{Deserialize, Serialize};

/// 消息类型
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum MessageType {
    Register,
    Heartbeat,
    Task,
    TaskResult,
    Message,
    Broadcast,
    Discovery,
    Error,
    Ack,
}

impl Default for MessageType {
    fn default() -> Self {
        Self::Message
    }
}

/// 消息
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct Message {
    pub id: String,
    pub msg_type: MessageType,
    pub from: String,
    pub to: String,
    pub subject: String,
    pub data: serde_json::Value,
    pub timestamp: u64,
}

impl Message {
    pub fn new(msg_type: MessageType, from: String, data: serde_json::Value) -> Self {
        Self {
            id: uuid::Uuid::new_v4().to_string(),
            msg_type,
            from,
            data,
            timestamp: std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .map(|d| d.as_millis() as u64)
                .unwrap_or(0),
            ..Default::default()
        }
    }
}

/// 任务
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Task {
    pub task_id: String,
    pub skill_id: String,
    pub operation: String,
    pub input: serde_json::Value,
    pub priority: i32,
    pub timeout: u32,
}

impl Default for Task {
    fn default() -> Self {
        Self {
            task_id: uuid::Uuid::new_v4().to_string(),
            skill_id: String::new(),
            operation: String::new(),
            input: serde_json::json!({}),
            priority: 0,
            timeout: 30,
        }
    }
}

/// 任务结果
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TaskResult {
    pub task_id: String,
    pub success: bool,
    pub result: Option<serde_json::Value>,
    pub error: Option<String>,
    pub agent_id: String,
    pub duration_ms: u64,
}