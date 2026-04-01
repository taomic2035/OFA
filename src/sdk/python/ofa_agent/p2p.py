"""
OFA P2P Module
Agent间直接通信
"""

import asyncio
import json
import logging
import socket
import threading
import time
import uuid
from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Dict, List, Optional, Callable

logger = logging.getLogger(__name__)


class P2PMessageType(Enum):
    """P2P消息类型"""
    DATA = "data"
    BROADCAST = "broadcast"
    REQUEST = "request"
    RESPONSE = "response"
    DISCOVERY = "discovery"
    HEARTBEAT = "heartbeat"


@dataclass
class PeerInfo:
    """设备信息"""
    id: str
    name: str
    address: str
    port: int
    type: str = "unknown"
    online: bool = True
    latency_ms: int = 0
    last_seen: float = field(default_factory=time.time)


@dataclass
class P2PMessage:
    """P2P消息"""
    type: P2PMessageType
    from_id: str
    to_id: Optional[str]
    data: Any
    timestamp: float = field(default_factory=time.time)
    msg_id: str = field(default_factory=lambda: uuid.uuid4().hex[:8])

    def to_json(self) -> str:
        return json.dumps({
            "type": self.type.value,
            "from": self.from_id,
            "to": self.to_id,
            "data": self.data,
            "timestamp": self.timestamp,
            "msg_id": self.msg_id,
        })

    @classmethod
    def from_json(cls, json_str: str) -> 'P2PMessage':
        obj = json.loads(json_str)
        return cls(
            type=P2PMessageType(obj["type"]),
            from_id=obj["from"],
            to_id=obj.get("to"),
            data=obj["data"],
            timestamp=obj["timestamp"],
            msg_id=obj["msg_id"],
        )


class P2PClient:
    """P2P客户端"""

    def __init__(self, agent_id: str, port: int = 0):
        self.agent_id = agent_id
        self.port = port or self._find_free_port()
        self._peers: Dict[str, PeerInfo] = {}
        self._message_handlers: List[Callable] = []
        self._peer_callbacks: List[Callable] = []
        self._running = False
        self._server_socket: Optional[socket.socket] = None
        self._server_thread: Optional[threading.Thread] = None
        self._lock = threading.RLock()

        logger.info(f"P2P client initialized on port {self.port}")

    def _find_free_port(self) -> int:
        """查找可用端口"""
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            s.bind(('', 0))
            return s.getsockname()[1]

    def start(self):
        """启动P2P服务"""
        if self._running:
            return

        self._running = True
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        self._server_socket.bind(('0.0.0.0', self.port))
        self._server_socket.listen(10)

        self._server_thread = threading.Thread(target=self._accept_loop, daemon=True)
        self._server_thread.start()

        logger.info(f"P2P server started on port {self.port}")

    def stop(self):
        """停止P2P服务"""
        self._running = False
        if self._server_socket:
            self._server_socket.close()
        if self._server_thread:
            self._server_thread.join(timeout=2)
        logger.info("P2P server stopped")

    def _accept_loop(self):
        """接受连接循环"""
        while self._running:
            try:
                conn, addr = self._server_socket.accept()
                threading.Thread(target=self._handle_connection, args=(conn, addr), daemon=True).start()
            except Exception as e:
                if self._running:
                    logger.error(f"Accept error: {e}")

    def _handle_connection(self, conn: socket.socket, addr: tuple):
        """处理连接"""
        try:
            data = b""
            while True:
                chunk = conn.recv(4096)
                if not chunk:
                    break
                data += chunk

            if data:
                msg = P2PMessage.from_json(data.decode('utf-8'))
                self._handle_message(msg)

        except Exception as e:
            logger.error(f"Connection error: {e}")
        finally:
            conn.close()

    def _handle_message(self, msg: P2PMessage):
        """处理接收的消息"""
        logger.debug(f"Received P2P message from {msg.from_id}: {msg.type.value}")

        # 更新设备信息
        with self._lock:
            if msg.from_id in self._peers:
                peer = self._peers[msg.from_id]
                peer.last_seen = time.time()
                peer.online = True

        # 调用处理器
        for handler in self._message_handlers:
            try:
                handler(msg)
            except Exception as e:
                logger.error(f"Message handler error: {e}")

    def send(self, peer_id: str, data: Any) -> bool:
        """发送消息到指定设备"""
        with self._lock:
            peer = self._peers.get(peer_id)
            if not peer or not peer.online:
                logger.warning(f"Peer not found or offline: {peer_id}")
                return False

        msg = P2PMessage(
            type=P2PMessageType.DATA,
            from_id=self.agent_id,
            to_id=peer_id,
            data=data,
        )

        return self._send_to_peer(peer, msg)

    def _send_to_peer(self, peer: PeerInfo, msg: P2PMessage) -> bool:
        """发送到设备"""
        try:
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            sock.settimeout(5.0)
            sock.connect((peer.address, peer.port))
            sock.sendall(msg.to_json().encode('utf-8'))
            sock.close()
            return True
        except Exception as e:
            logger.error(f"Send to {peer.id} failed: {e}")
            with self._lock:
                peer.online = False
            return False

    def broadcast(self, data: Any) -> Dict[str, bool]:
        """广播消息"""
        msg = P2PMessage(
            type=P2PMessageType.BROADCAST,
            from_id=self.agent_id,
            to_id=None,
            data=data,
        )

        results = {}
        with self._lock:
            peers = [p for p in self._peers.values() if p.online]

        for peer in peers:
            results[peer.id] = self._send_to_peer(peer, msg)

        return results

    def request(self, peer_id: str, data: Any, timeout: float = 5.0) -> Optional[Any]:
        """请求-响应模式"""
        with self._lock:
            peer = self._peers.get(peer_id)
            if not peer or not peer.online:
                return None

        msg = P2PMessage(
            type=P2PMessageType.REQUEST,
            from_id=self.agent_id,
            to_id=peer_id,
            data=data,
        )

        # 创建响应等待
        response_future: threading.Event = threading.Event()
        response_data: Any = None

        def response_handler(m: P2PMessage):
            if m.type == P2PMessageType.RESPONSE and m.msg_id == msg.msg_id:
                nonlocal response_data
                response_data = m.data
                response_future.set()

        self._message_handlers.append(response_handler)

        if self._send_to_peer(peer, msg):
            response_future.wait(timeout)
            self._message_handlers.remove(response_handler)
            return response_data

        self._message_handlers.remove(response_handler)
        return None

    def add_peer(self, peer: PeerInfo):
        """添加设备"""
        with self._lock:
            self._peers[peer.id] = peer

        for callback in self._peer_callbacks:
            try:
                callback(peer.id, peer.name, True)
            except Exception as e:
                logger.error(f"Peer callback error: {e}")

        logger.info(f"Peer added: {peer.id}")

    def remove_peer(self, peer_id: str):
        """移除设备"""
        with self._lock:
            peer = self._peers.pop(peer_id, None)

        if peer:
            for callback in self._peer_callbacks:
                try:
                    callback(peer_id, peer.name, False)
                except Exception as e:
                    logger.error(f"Peer callback error: {e}")

            logger.info(f"Peer removed: {peer_id}")

    def get_peer(self, peer_id: str) -> Optional[PeerInfo]:
        """获取设备信息"""
        with self._lock:
            return self._peers.get(peer_id)

    def list_peers(self) -> List[PeerInfo]:
        """列出所有设备"""
        with self._lock:
            return list(self._peers.values())

    def online_peers(self) -> List[PeerInfo]:
        """列出在线设备"""
        with self._lock:
            return [p for p in self._peers.values() if p.online]

    def on_message(self, handler: Callable):
        """注册消息处理器"""
        self._message_handlers.append(handler)

    def on_peer_change(self, callback: Callable):
        """注册设备变化回调"""
        self._peer_callbacks.append(callback)

    def check_peers_status(self, timeout: float = 2.0):
        """检查设备状态"""
        with self._lock:
            peers = list(self._peers.values())

        now = time.time()
        for peer in peers:
            # 检查最后活跃时间
            if now - peer.last_seen > 30:  # 30秒未活跃
                peer.online = False
                for callback in self._peer_callbacks:
                    try:
                        callback(peer.id, peer.name, False)
                    except Exception as e:
                        logger.error(f"Peer callback error: {e}")

    def get_stats(self) -> Dict[str, Any]:
        """获取统计信息"""
        with self._lock:
            online = len([p for p in self._peers.values() if p.online])
            total = len(self._peers)
        return {
            "agent_id": self.agent_id,
            "port": self.port,
            "peers_total": total,
            "peers_online": online,
        }


class P2PDiscovery:
    """P2P设备发现"""

    def __init__(self, agent_id: str, p2p_client: P2PClient):
        self.agent_id = agent_id
        self.p2p_client = p2p_client
        self._discovering = False
        self._discovery_thread: Optional[threading.Thread] = None
        self._broadcast_port = 9999  # 发现广播端口
        self._discovery_interval = 5.0  # 发现间隔

    def start_discovery(self):
        """启动设备发现"""
        if self._discovering:
            return

        self._discovering = True
        self._discovery_thread = threading.Thread(target=self._discovery_loop, daemon=True)
        self._discovery_thread.start()
        logger.info("P2P discovery started")

    def stop_discovery(self):
        """停止设备发现"""
        self._discovering = False
        if self._discovery_thread:
            self._discovery_thread.join(timeout=2)
        logger.info("P2P discovery stopped")

    def _discovery_loop(self):
        """发现循环"""
        # 创建UDP广播socket
        sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        sock.setsockopt(socket.SOL_SOCKET, socket.SO_BROADCAST, 1)
        sock.bind(('0.0.0.0', self._broadcast_port))

        while self._discovering:
            try:
                # 发送发现广播
                discovery_msg = json.dumps({
                    "type": "discovery",
                    "agent_id": self.agent_id,
                    "port": self.p2p_client.port,
                    "timestamp": time.time(),
                }).encode('utf-8')

                sock.sendto(discovery_msg, ('255.255.255.255', self._broadcast_port))

                # 接收其他设备的广播
                sock.settimeout(1.0)
                try:
                    data, addr = sock.recvfrom(4096)
                    msg = json.loads(data.decode('utf-8'))

                    if msg["type"] == "discovery" and msg["agent_id"] != self.agent_id:
                        peer = PeerInfo(
                            id=msg["agent_id"],
                            name=msg["agent_id"],
                            address=addr[0],
                            port=msg["port"],
                            type="p2p",
                            online=True,
                            last_seen=time.time(),
                        )
                        self.p2p_client.add_peer(peer)
                except socket.timeout:
                    pass

                # 检查设备状态
                self.p2p_client.check_peers_status()

                time.sleep(self._discovery_interval)

            except Exception as e:
                logger.error(f"Discovery error: {e}")

        sock.close()

    def manual_discovery(self, addresses: List[str]):
        """手动发现指定地址的设备"""
        for addr in addresses:
            try:
                # 尝试连接并获取设备信息
                parts = addr.split(':')
                host = parts[0]
                port = int(parts[1]) if len(parts) > 1 else self._broadcast_port

                sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                sock.settimeout(2.0)
                sock.connect((host, port))

                # 发送发现请求
                msg = P2PMessage(
                    type=P2PMessageType.DISCOVERY,
                    from_id=self.agent_id,
                    to_id=None,
                    data={"request": "info"},
                )
                sock.sendall(msg.to_json().encode('utf-8'))
                sock.close()

            except Exception as e:
                logger.warning(f"Manual discovery failed for {addr}: {e}")