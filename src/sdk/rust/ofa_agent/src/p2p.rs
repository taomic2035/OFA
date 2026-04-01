//! OFA Rust SDK - P2P Module
//! Agent 间直接通信

use std::collections::HashMap;
use std::net::{SocketAddr, TcpListener, TcpStream, UdpSocket};
use std::sync::Arc;
use std::time::{Duration, SystemTime, UNIX_EPOCH};
use tokio::sync::RwLock;
use serde::{Deserialize, Serialize};

/// P2P 消息类型
#[derive(Debug, Clone, Copy, Serialize, Deserialize)]
pub enum P2PMessageType {
    Data,
    Broadcast,
    Request,
    Response,
    Discovery,
    Heartbeat,
}

/// P2P 消息
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct P2PMessage {
    pub msg_type: P2PMessageType,
    pub from_id: String,
    pub to_id: Option<String>,
    pub data: Option<Vec<u8>>,
    pub timestamp: u64,
    pub msg_id: String,
}

impl P2PMessage {
    pub fn new(msg_type: P2PMessageType, from_id: String, to_id: Option<String>, data: Option<Vec<u8>>) -> Self {
        Self {
            msg_type,
            from_id,
            to_id,
            data,
            timestamp: SystemTime::now()
                .duration_since(UNIX_EPOCH)
                .unwrap()
                .as_millis() as u64,
            msg_id: uuid::Uuid::new_v4().to_string()[..8].to_string(),
        }
    }

    pub fn to_json(&self) -> Result<String, serde_json::Error> {
        serde_json::to_string(self)
    }

    pub fn from_json(json: &str) -> Result<Self, serde_json::Error> {
        serde_json::from_str(json)
    }
}

/// 设备信息
#[derive(Debug, Clone)]
pub struct PeerInfo {
    pub id: String,
    pub name: String,
    pub address: String,
    pub port: u16,
    pub online: bool,
    pub last_seen: u64,
    pub latency_ms: u32,
}

impl PeerInfo {
    pub fn new(id: String, name: String, address: String, port: u16) -> Self {
        Self {
            id,
            name,
            address,
            port,
            online: true,
            last_seen: SystemTime::now()
                .duration_since(UNIX_EPOCH)
                .unwrap()
                .as_millis() as u64,
            latency_ms: 0,
        }
    }
}

/// P2P 客户端
pub struct P2PClient {
    agent_id: String,
    port: u16,
    peers: Arc<RwLock<HashMap<String, PeerInfo>>>,
    running: Arc<std::sync::atomic::AtomicBool>,
    message_handlers: Arc<RwLock<Vec<Box<dyn Fn(&P2PMessage) + Send + Sync>>>>,
}

impl P2PClient {
    pub fn new(agent_id: String, port: Option<u16>) -> Self {
        Self {
            agent_id,
            port: port.unwrap_or_else(|| 9000 + (rand::random::<u16>() % 1000)),
            peers: Arc::new(RwLock::new(HashMap::new())),
            running: Arc::new(std::sync::atomic::AtomicBool::new(false)),
            message_handlers: Arc::new(RwLock::new(Vec::new())),
        }
    }

    /// 启动 P2P 服务
    pub async fn start(&self) -> Result<(), Box<dyn std::error::Error>> {
        if self.running.load(std::sync::atomic::Ordering::SeqCst) {
            return Ok(());
        }
        self.running.store(true, std::sync::atomic::Ordering::SeqCst);

        let addr = format!("0.0.0.0:{}", self.port);
        let listener = TcpListener::bind(&addr)?;

        log::info!("P2P server started on port {}", self.port);

        // 启动监听任务
        let running = self.running.clone();
        let peers = self.peers.clone();
        let handlers = self.message_handlers.clone();

        tokio::spawn(async move {
            for stream in listener.incoming() {
                if !running.load(std::sync::atomic::Ordering::SeqCst) {
                    break;
                }

                if let Ok(stream) = stream {
                    let peers = peers.clone();
                    let handlers = handlers.clone();

                    tokio::spawn(async move {
                        if let Ok(data) = Self::read_stream(&stream) {
                            if let Ok(msg) = P2PMessage::from_json(&data) {
                                // 更新设备状态
                                let mut peers_write = peers.write().await;
                                if let Some(peer) = peers_write.get_mut(&msg.from_id) {
                                    peer.last_seen = SystemTime::now()
                                        .duration_since(UNIX_EPOCH)
                                        .unwrap()
                                        .as_millis() as u64;
                                    peer.online = true;
                                }

                                // 调用处理器
                                let handlers_read = handlers.read().await;
                                for handler in handlers_read.iter() {
                                    handler(&msg);
                                }
                            }
                        }
                    });
                }
            }
        });

        Ok(())
    }

    fn read_stream(stream: &TcpStream) -> Result<String, std::io::Error> {
        use std::io::Read;
        let mut data = String::new();
        let mut reader = stream;
        reader.read_to_string(&mut data)?;
        Ok(data)
    }

    /// 停止 P2P 服务
    pub async fn stop(&self) {
        self.running.store(false, std::sync::atomic::Ordering::SeqCst);
        log::info!("P2P server stopped");
    }

    /// 发送消息
    pub async fn send(&self, peer_id: &str, data: Vec<u8>) -> bool {
        let peers = self.peers.read().await;
        let peer = match peers.get(peer_id) {
            Some(p) if p.online => p.clone(),
            _ => {
                log::warn!("Peer not found or offline: {}", peer_id);
                return false;
            }
        };
        drop(peers);

        let msg = P2PMessage::new(
            P2PMessageType::Data,
            self.agent_id.clone(),
            Some(peer_id.to_string()),
            Some(data),
        );

        self.send_to_peer(&peer, &msg).await
    }

    async fn send_to_peer(&self, peer: &PeerInfo, msg: &P2PMessage) -> bool {
        let addr = format!("{}:{}", peer.address, peer.port);

        match TcpStream::connect_timeout(&addr.parse().unwrap(), Duration::from_secs(5)) {
            Ok(mut stream) => {
                use std::io::Write;
                if let Ok(json) = msg.to_json() {
                    stream.write_all(json.as_bytes()).is_ok()
                } else {
                    false
                }
            }
            Err(e) => {
                log::error!("Send to {} failed: {}", peer.id, e);
                let mut peers = self.peers.write().await;
                if let Some(p) = peers.get_mut(&peer.id) {
                    p.online = false;
                }
                false
            }
        }
    }

    /// 广播消息
    pub async fn broadcast(&self, data: Vec<u8>) -> HashMap<String, bool> {
        let mut results = HashMap::new();
        let peers = self.peers.read().await;

        for (peer_id, peer) in peers.iter() {
            if peer.online {
                results.insert(peer_id.clone(), self.send(peer_id, data.clone()).await);
            }
        }

        results
    }

    /// 添加设备
    pub async fn add_peer(&self, peer: PeerInfo) {
        self.peers.write().await.insert(peer.id.clone(), peer.clone());
        log::info!("Peer added: {}", peer.id);
    }

    /// 移除设备
    pub async fn remove_peer(&self, peer_id: &str) {
        if self.peers.write().await.remove(peer_id).is_some() {
            log::info!("Peer removed: {}", peer_id);
        }
    }

    /// 获取设备列表
    pub async fn get_peers(&self) -> Vec<PeerInfo> {
        self.peers.read().await.values().cloned().collect()
    }

    /// 获取在线设备
    pub async fn get_online_peers(&self) -> Vec<PeerInfo> {
        self.peers
            .read()
            .await
            .values()
            .filter(|p| p.online)
            .cloned()
            .collect()
    }

    /// 获取端口
    pub fn get_port(&self) -> u16 {
        self.port
    }

    /// 添加消息处理器
    pub async fn on_message<F>(&self, handler: F)
    where
        F: Fn(&P2PMessage) + Send + Sync + 'static,
    {
        self.message_handlers.write().await.push(Box::new(handler));
    }
}