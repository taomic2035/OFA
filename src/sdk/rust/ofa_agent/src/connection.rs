//! Connection Module
//! Sprint 29: Rust Agent SDK

use async_trait::async_trait;
use crate::config::AgentConfig;
use crate::message::Message;
use crate::error::{Error, Result};

/// 连接trait
#[async_trait]
pub trait Connection: Send + Sync {
    /// 连接
    async fn connect(&self, config: &AgentConfig) -> Result<()>;
    /// 断开
    async fn disconnect(&self) -> Result<()>;
    /// 发送
    async fn send(&self, msg: &Message) -> Result<()>;
    /// 接收
    async fn receive(&self) -> Result<Option<Message>>;
    /// 是否已连接
    fn is_connected(&self) -> bool;
}

/// 连接类型
#[derive(Debug, Clone, Copy)]
pub enum ConnectionType {
    Grpc,
    Http,
    WebSocket,
}

impl ConnectionType {
    /// 从字符串解析
    pub fn from_str(s: &str) -> Option<Self> {
        match s.to_lowercase().as_str() {
            "grpc" => Some(Self::Grpc),
            "http" => Some(Self::Http),
            "websocket" | "ws" => Some(Self::WebSocket),
            _ => None,
        }
    }

    /// 创建连接
    pub fn create_connection(config: &AgentConfig) -> Result<Box<dyn Connection>> {
        let conn_type = Self::from_str(&config.connection_type)
            .ok_or_else(|| Error::Config(format!("Unknown connection type: {}", config.connection_type)))?;

        match conn_type {
            ConnectionType::Grpc => Ok(Box::new(GrpcConnection::new())),
            ConnectionType::Http => Ok(Box::new(HttpConnection::new())),
            ConnectionType::WebSocket => Ok(Box::new(WebSocketConnection::new())),
        }
    }
}

/// HTTP连接实现
pub struct HttpConnection {
    connected: std::sync::Arc<tokio::sync::RwLock<bool>>,
    base_url: std::sync::Arc<tokio::sync::RwLock<String>>,
    message_queue: std::sync::Arc<tokio::sync::Mutex<Vec<Message>>>,
}

impl HttpConnection {
    pub fn new() -> Self {
        Self {
            connected: std::sync::Arc::new(tokio::sync::RwLock::new(false)),
            base_url: std::sync::Arc::new(tokio::sync::RwLock::new(String::new())),
            message_queue: std::sync::Arc::new(tokio::sync::Mutex::new(Vec::new())),
        }
    }
}

#[async_trait]
impl Connection for HttpConnection {
    async fn connect(&self, config: &AgentConfig) -> Result<()> {
        *self.base_url.write().await = format!("http://{}", config.center_url);
        *self.connected.write().await = true;
        Ok(())
    }

    async fn disconnect(&self) -> Result<()> {
        *self.connected.write().await = false;
        Ok(())
    }

    async fn send(&self, msg: &Message) -> Result<()> {
        // 简化实现
        Ok(())
    }

    async fn receive(&self) -> Result<Option<Message>> {
        let mut queue = self.message_queue.lock().await;
        Ok(queue.pop())
    }

    fn is_connected(&self) -> bool {
        futures::executor::block_on(async { *self.connected.read().await })
    }
}

/// gRPC连接实现
pub struct GrpcConnection {
    connected: std::sync::Arc<tokio::sync::RwLock<bool>>,
}

impl GrpcConnection {
    pub fn new() -> Self {
        Self {
            connected: std::sync::Arc::new(tokio::sync::RwLock::new(false)),
        }
    }
}

#[async_trait]
impl Connection for GrpcConnection {
    async fn connect(&self, _config: &AgentConfig) -> Result<()> {
        // 实际实现需要gRPC客户端
        *self.connected.write().await = true;
        Ok(())
    }

    async fn disconnect(&self) -> Result<()> {
        *self.connected.write().await = false;
        Ok(())
    }

    async fn send(&self, _msg: &Message) -> Result<()> {
        Ok(())
    }

    async fn receive(&self) -> Result<Option<Message>> {
        Ok(None)
    }

    fn is_connected(&self) -> bool {
        futures::executor::block_on(async { *self.connected.read().await })
    }
}

/// WebSocket连接实现
pub struct WebSocketConnection {
    connected: std::sync::Arc<tokio::sync::RwLock<bool>>,
}

impl WebSocketConnection {
    pub fn new() -> Self {
        Self {
            connected: std::sync::Arc::new(tokio::sync::RwLock::new(false)),
        }
    }
}

#[async_trait]
impl Connection for WebSocketConnection {
    async fn connect(&self, _config: &AgentConfig) -> Result<()> {
        *self.connected.write().await = true;
        Ok(())
    }

    async fn disconnect(&self) -> Result<()> {
        *self.connected.write().await = false;
        Ok(())
    }

    async fn send(&self, _msg: &Message) -> Result<()> {
        Ok(())
    }

    async fn receive(&self) -> Result<Option<Message>> {
        Ok(None)
    }

    fn is_connected(&self) -> bool {
        futures::executor::block_on(async { *self.connected.read().await })
    }
}