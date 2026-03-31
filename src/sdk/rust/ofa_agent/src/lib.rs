//! OFA Rust Agent SDK
//! Sprint 29: Multi-platform SDK Extension

pub mod agent;
pub mod config;
pub mod skills;
pub mod connection;
pub mod protocol;
pub mod message;
pub mod error;

pub use agent::{Agent, AgentState, AgentInfo};
pub use config::AgentConfig;
pub use skills::{Skill, SkillExecutor, SkillRegistry};
pub use connection::{Connection, ConnectionType};
pub use protocol::Protocol;
pub use message::{Message, MessageType, Task, TaskResult};
pub use error::{Error, Result};

/// SDK版本
pub const VERSION: &str = "8.1.0";