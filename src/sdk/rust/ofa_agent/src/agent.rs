//! Agent Core Module
//! Sprint 29: Rust Agent SDK

use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, Instant};
use tokio::sync::{Mutex, RwLock};
use tokio::time::{interval, sleep};
use tracing::{info, warn, error, debug};
use uuid::Uuid;

use crate::config::AgentConfig;
use crate::connection::{Connection, ConnectionType};
use crate::skills::{SkillExecutor, SkillRegistry};
use crate::message::{Message, MessageType, Task, TaskResult};
use crate::error::{Error, Result};

/// Agent状态
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum AgentState {
    Initializing,
    Connecting,
    Online,
    Busy,
    Offline,
    Error,
}

impl AgentState {
    pub fn as_str(&self) -> &'static str {
        match self {
            AgentState::Initializing => "initializing",
            AgentState::Connecting => "connecting",
            AgentState::Online => "online",
            AgentState::Busy => "busy",
            AgentState::Offline => "offline",
            AgentState::Error => "error",
        }
    }
}

/// Agent信息
#[derive(Debug, Clone)]
pub struct AgentInfo {
    pub id: String,
    pub name: String,
    pub agent_type: String,
    pub version: String,
    pub platform: String,
    pub skills: Vec<String>,
    pub state: AgentState,
    pub metadata: HashMap<String, String>,
    pub last_heartbeat: Instant,
}

/// Agent统计
#[derive(Debug, Clone, Default)]
pub struct AgentStats {
    pub tasks_executed: u64,
    pub tasks_success: u64,
    pub tasks_failed: u64,
    pub messages_sent: u64,
    pub messages_received: u64,
    pub uptime_seconds: u64,
}

/// OFA Agent
pub struct Agent {
    config: AgentConfig,
    info: Arc<RwLock<AgentInfo>>,
    state: Arc<RwLock<AgentState>>,
    connection: Arc<Mutex<Option<Box<dyn Connection>>>>,
    skill_executor: Arc<SkillExecutor>,
    skill_registry: Arc<SkillRegistry>,
    stats: Arc<RwLock<AgentStats>>,
    running: Arc<RwLock<bool>>,
}

impl Agent {
    /// 创建新Agent
    pub fn new(config: AgentConfig) -> Self {
        let id = if config.agent_id.is_empty() {
            format!("rust-agent-{}", Uuid::new_v4().to_string().split('-').next().unwrap_or("unknown"))
        } else {
            config.agent_id.clone()
        };

        let info = AgentInfo {
            id,
            name: config.name.clone(),
            agent_type: "rust".to_string(),
            version: crate::VERSION.to_string(),
            platform: "rust".to_string(),
            skills: config.skills.clone(),
            state: AgentState::Initializing,
            metadata: config.metadata.clone(),
            last_heartbeat: Instant::now(),
        };

        Self {
            config,
            info: Arc::new(RwLock::new(info)),
            state: Arc::new(RwLock::new(AgentState::Initializing)),
            connection: Arc::new(Mutex::new(None)),
            skill_executor: Arc::new(SkillExecutor::new()),
            skill_registry: Arc::new(SkillRegistry::new()),
            stats: Arc::new(RwLock::new(AgentStats::default())),
            running: Arc::new(RwLock::new(false)),
        }
    }

    /// 获取Agent ID
    pub fn id(&self) -> String {
        self.info.read().await.id.clone()
    }

    /// 获取状态
    pub fn state(&self) -> AgentState {
        *self.state.read().await
    }

    /// 是否在线
    pub fn is_online(&self) -> bool {
        self.state() == AgentState::Online
    }

    /// 连接到Center
    pub async fn connect(&self) -> Result<()> {
        *self.state.write().await = AgentState::Connecting;
        info!("Connecting to {}", self.config.center_url);

        // 创建连接
        let conn = ConnectionType::create_connection(&self.config)?;
        conn.connect(&self.config).await?;

        // 注册
        self.register(&conn).await?;

        // 保存连接
        *self.connection.lock().await = Some(conn);

        *self.state.write().await = AgentState::Online;
        *self.running.write().await = true;

        info!("Agent connected: {}", self.id());
        Ok(())
    }

    /// 断开连接
    pub async fn disconnect(&self) -> Result<()> {
        *self.running.write().await = false;

        if let Some(conn) = self.connection.lock().await.take() {
            conn.disconnect().await?;
        }

        *self.state.write().await = AgentState::Offline;
        info!("Agent disconnected: {}", self.id());
        Ok(())
    }

    /// 注册到Center
    async fn register(&self, conn: &Box<dyn Connection>) -> Result<()> {
        let info = self.info.read().await;
        let msg = Message {
            msg_type: MessageType::Register,
            from: info.id.clone(),
            data: serde_json::to_value(&*info)?,
            ..Default::default()
        };
        conn.send(&msg).await?;
        Ok(())
    }

    /// 注册技能
    pub fn register_skill(&self, skill_id: &str, handler: impl crate::skills::SkillHandler + 'static) {
        self.skill_executor.register(skill_id, Arc::new(handler));
        self.info.write().await.skills.push(skill_id.to_string());
        info!("Skill registered: {}", skill_id);
    }

    /// 执行任务
    pub async fn execute_task(&self, task: Task) -> Result<TaskResult> {
        *self.state.write().await = AgentState::Busy;

        let start = Instant::now();
        let result = self.skill_executor.execute(&task).await;

        let stats = self.stats.write().await;
        stats.tasks_executed += 1;
        if result.is_ok() {
            stats.tasks_success += 1;
        } else {
            stats.tasks_failed += 1;
        }

        *self.state.write().await = AgentState::Online;

        result
    }

    /// 发送消息
    pub async fn send_message(&self, target: &str, msg_type: MessageType, data: serde_json::Value) -> Result<()> {
        let conn = self.connection.lock().await;
        if let Some(ref c) = *conn {
            let msg = Message {
                msg_type,
                from: self.id(),
                to: target.to_string(),
                data,
                ..Default::default()
            };
            c.send(&msg).await?;

            self.stats.write().await.messages_sent += 1;
        }
        Ok(())
    }

    /// 广播消息
    pub async fn broadcast(&self, msg_type: MessageType, data: serde_json::Value) -> Result<()> {
        self.send_message("", msg_type, data).await
    }

    /// 运行Agent
    pub async fn run(&self) -> Result<()> {
        self.connect().await?;

        // 心跳循环
        let heartbeat_task = self.heartbeat_loop();

        // 消息处理循环
        let message_task = self.message_loop();

        tokio::try_join!(heartbeat_task, message_task)?;

        Ok(())
    }

    /// 心跳循环
    async fn heartbeat_loop(&self) -> Result<()> {
        let mut ticker = interval(Duration::from_secs(self.config.heartbeat_interval as u64));

        loop {
            if !*self.running.read().await {
                break;
            }

            ticker.tick().await;

            if let Some(ref conn) = *self.connection.lock().await {
                let msg = Message {
                    msg_type: MessageType::Heartbeat,
                    from: self.id(),
                    data: serde_json::json!({
                        "state": self.state().as_str(),
                    }),
                    ..Default::default()
                };
                conn.send(&msg).await?;
                self.info.write().await.last_heartbeat = Instant::now();
            }
        }

        Ok(())
    }

    /// 消息处理循环
    async fn message_loop(&self) -> Result<()> {
        loop {
            if !*self.running.read().await {
                break;
            }

            if let Some(ref conn) = *self.connection.lock().await {
                if let Some(msg) = conn.receive().await? {
                    self.handle_message(msg).await?;
                    self.stats.write().await.messages_received += 1;
                }
            }

            sleep(Duration::from_millis(100)).await;
        }

        Ok(())
    }

    /// 处理消息
    async fn handle_message(&self, msg: Message) -> Result<()> {
        match msg.msg_type {
            MessageType::Task => {
                let task: Task = serde_json::from_value(msg.data)?;
                let result = self.execute_task(task).await?;

                if let Some(ref conn) = *self.connection.lock().await {
                    let result_msg = Message {
                        msg_type: MessageType::TaskResult,
                        from: self.id(),
                        data: serde_json::to_value(result)?,
                        ..Default::default()
                    };
                    conn.send(&result_msg).await?;
                }
            }
            _ => {
                debug!("Received message: {:?}", msg.msg_type);
            }
        }
        Ok(())
    }

    /// 获取统计
    pub fn stats(&self) -> AgentStats {
        self.stats.read().await.clone()
    }

    /// 获取信息
    pub fn info(&self) -> AgentInfo {
        self.info.read().await.clone()
    }
}