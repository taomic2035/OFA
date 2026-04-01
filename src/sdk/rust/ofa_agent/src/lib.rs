//! OFA Rust Agent SDK
//! Sprint 29: Multi-platform SDK Extension
//! v0.10.0: Offline + P2P + Constraint support

pub mod agent;
pub mod config;
pub mod skills;
pub mod connection;
pub mod protocol;
pub mod message;
pub mod error;

// 新增模块
pub mod offline;
pub mod p2p;
pub mod constraint;

pub use agent::{Agent, AgentState, AgentInfo};
pub use config::AgentConfig;
pub use skills::{Skill, SkillExecutor, SkillRegistry};
pub use connection::{Connection, ConnectionType};
pub use protocol::Protocol;
pub use message::{Message, MessageType, Task, TaskResult};
pub use error::{Error, Result};

// 离线模块导出
pub use offline::{
    OfflineLevel, TaskStatus, LocalTask, LocalScheduler,
    OfflineCache, OfflineManager,
};

// P2P 模块导出
pub use p2p::{
    P2PMessageType, P2PMessage, PeerInfo, P2PClient,
};

// 约束模块导出
pub use constraint::{
    ConstraintType, ConstraintResult, ConstraintRule, ConstraintChecker,
};

/// SDK版本
pub const VERSION: &str = "0.10.0";