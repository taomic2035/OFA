//! Agent Configuration
//! Sprint 29: Rust Agent SDK

use std::collections::HashMap;

/// Agent配置
#[derive(Debug, Clone)]
pub struct AgentConfig {
    /// Agent ID
    pub agent_id: String,
    /// Agent名称
    pub name: String,
    /// Center地址
    pub center_url: String,
    /// 连接类型
    pub connection_type: String,
    /// 心跳间隔(秒)
    pub heartbeat_interval: u32,
    /// 重连间隔(秒)
    pub reconnect_interval: u32,
    /// 最大重连次数
    pub max_reconnect_attempts: u32,
    /// 技能列表
    pub skills: Vec<String>,
    /// 元数据
    pub metadata: HashMap<String, String>,
    /// 启用TLS
    pub tls_enabled: bool,
    /// 超时(秒)
    pub timeout: u32,
}

impl Default for AgentConfig {
    fn default() -> Self {
        Self {
            agent_id: String::new(),
            name: "Rust Agent".to_string(),
            center_url: "localhost:9090".to_string(),
            connection_type: "grpc".to_string(),
            heartbeat_interval: 30,
            reconnect_interval: 5,
            max_reconnect_attempts: 10,
            skills: Vec::new(),
            metadata: HashMap::new(),
            tls_enabled: false,
            timeout: 30,
        }
    }
}

impl AgentConfig {
    /// 创建配置构建器
    pub fn builder() -> AgentConfigBuilder {
        AgentConfigBuilder::default()
    }
}

/// 配置构建器
#[derive(Debug, Default)]
pub struct AgentConfigBuilder {
    config: AgentConfig,
}

impl AgentConfigBuilder {
    pub fn agent_id(mut self, id: impl Into<String>) -> Self {
        self.config.agent_id = id.into();
        self
    }

    pub fn name(mut self, name: impl Into<String>) -> Self {
        self.config.name = name.into();
        self
    }

    pub fn center_url(mut self, url: impl Into<String>) -> Self {
        self.config.center_url = url.into();
        self
    }

    pub fn connection_type(mut self, conn_type: impl Into<String>) -> Self {
        self.config.connection_type = conn_type.into();
        self
    }

    pub fn heartbeat_interval(mut self, interval: u32) -> Self {
        self.config.heartbeat_interval = interval;
        self
    }

    pub fn skill(mut self, skill: impl Into<String>) -> Self {
        self.config.skills.push(skill.into());
        self
    }

    pub fn skills(mut self, skills: Vec<String>) -> Self {
        self.config.skills = skills;
        self
    }

    pub fn metadata(mut self, key: impl Into<String>, value: impl Into<String>) -> Self {
        self.config.metadata.insert(key.into(), value.into());
        self
    }

    pub fn tls(mut self, enabled: bool) -> Self {
        self.config.tls_enabled = enabled;
        self
    }

    pub fn timeout(mut self, timeout: u32) -> Self {
        self.config.timeout = timeout;
        self
    }

    pub fn build(self) -> AgentConfig {
        self.config
    }
}