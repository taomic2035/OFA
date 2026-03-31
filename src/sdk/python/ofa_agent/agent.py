"""
OFA Agent Core
Sprint 29: Python Agent SDK
"""

import asyncio
import json
import logging
import time
import uuid
from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Dict, List, Optional, Callable, Awaitable

logger = logging.getLogger(__name__)


class AgentState(Enum):
    """Agent状态"""
    INITIALIZING = "initializing"
    CONNECTING = "connecting"
    ONLINE = "online"
    BUSY = "busy"
    OFFLINE = "offline"
    ERROR = "error"


@dataclass
class AgentConfig:
    """Agent配置"""
    agent_id: str = ""
    name: str = "Python Agent"
    center_url: str = "localhost:9090"
    connection_type: str = "grpc"  # grpc, http, websocket
    heartbeat_interval: float = 30.0
    reconnect_interval: float = 5.0
    max_reconnect_attempts: int = 10
    skills: List[str] = field(default_factory=list)
    metadata: Dict[str, Any] = field(default_factory=dict)
    tls_enabled: bool = False
    timeout: float = 30.0

    def __post_init__(self):
        if not self.agent_id:
            self.agent_id = f"python-agent-{uuid.uuid4().hex[:8]}"


@dataclass
class AgentInfo:
    """Agent信息"""
    id: str
    name: str
    type: str = "python"
    version: str = "8.1.0"
    platform: str = "python"
    skills: List[str] = field(default_factory=list)
    state: AgentState = AgentState.OFFLINE
    metadata: Dict[str, Any] = field(default_factory=dict)
    last_heartbeat: float = 0.0
    created_at: float = field(default_factory=time.time)


class OFAAgent:
    """OFA Python Agent"""

    def __init__(self, config: Optional[AgentConfig] = None):
        self.config = config or AgentConfig()
        self.info = AgentInfo(
            id=self.config.agent_id,
            name=self.config.name,
            skills=self.config.skills,
            metadata=self.config.metadata,
        )
        self.state = AgentState.INITIALIZING
        self._connection: Optional[Any] = None
        self._skill_registry: Dict[str, Callable] = {}
        self._task_handlers: Dict[str, Callable] = {}
        self._running = False
        self._heartbeat_task: Optional[asyncio.Task] = None
        self._message_handlers: Dict[str, List[Callable]] = {}

        logger.info(f"OFA Agent initialized: {self.info.id}")

    async def connect(self) -> bool:
        """连接到Center"""
        self.state = AgentState.CONNECTING
        logger.info(f"Connecting to {self.config.center_url}")

        try:
            # 根据连接类型创建连接
            if self.config.connection_type == "grpc":
                from .connection import GRPCConnection
                self._connection = GRPCConnection(self.config)
            elif self.config.connection_type == "http":
                from .connection import HTTPConnection
                self._connection = HTTPConnection(self.config)
            elif self.config.connection_type == "websocket":
                from .connection import WebSocketConnection
                self._connection = WebSocketConnection(self.config)

            await self._connection.connect()
            await self._register()
            self.state = AgentState.ONLINE
            self._running = True

            # 启动心跳
            self._heartbeat_task = asyncio.create_task(self._heartbeat_loop())

            logger.info(f"Agent connected: {self.info.id}")
            return True

        except Exception as e:
            self.state = AgentState.ERROR
            logger.error(f"Connection failed: {e}")
            return False

    async def disconnect(self) -> None:
        """断开连接"""
        self._running = False

        if self._heartbeat_task:
            self._heartbeat_task.cancel()
            try:
                await self._heartbeat_task
            except asyncio.CancelledError:
                pass

        if self._connection:
            await self._connection.disconnect()

        self.state = AgentState.OFFLINE
        logger.info(f"Agent disconnected: {self.info.id}")

    async def _register(self) -> None:
        """注册到Center"""
        register_msg = {
            "type": "register",
            "agent_id": self.info.id,
            "name": self.info.name,
            "type": self.info.type,
            "version": self.info.version,
            "platform": self.info.platform,
            "skills": self.info.skills,
            "metadata": self.info.metadata,
        }
        await self._connection.send(register_msg)

    async def _heartbeat_loop(self) -> None:
        """心跳循环"""
        while self._running:
            try:
                heartbeat_msg = {
                    "type": "heartbeat",
                    "agent_id": self.info.id,
                    "state": self.state.value,
                    "timestamp": time.time(),
                }
                await self._connection.send(heartbeat_msg)
                self.info.last_heartbeat = time.time()
                await asyncio.sleep(self.config.heartbeat_interval)
            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error(f"Heartbeat failed: {e}")
                await self._reconnect()

    async def _reconnect(self) -> None:
        """重连"""
        for attempt in range(self.config.max_reconnect_attempts):
            self.state = AgentState.CONNECTING
            logger.info(f"Reconnecting (attempt {attempt + 1})")

            try:
                await self._connection.connect()
                await self._register()
                self.state = AgentState.ONLINE
                self._running = True
                logger.info("Reconnected successfully")
                return
            except Exception as e:
                logger.error(f"Reconnect failed: {e}")
                await asyncio.sleep(self.config.reconnect_interval)

        self.state = AgentState.ERROR
        logger.error("Max reconnect attempts reached")

    def register_skill(self, skill_id: str, handler: Callable) -> None:
        """注册技能"""
        self._skill_registry[skill_id] = handler
        self.info.skills.append(skill_id)
        logger.info(f"Skill registered: {skill_id}")

    def unregister_skill(self, skill_id: str) -> None:
        """注销技能"""
        self._skill_registry.pop(skill_id, None)
        self.info.skills = [s for s in self.info.skills if s != skill_id]

    def on_message(self, msg_type: str, handler: Callable) -> None:
        """注册消息处理器"""
        if msg_type not in self._message_handlers:
            self._message_handlers[msg_type] = []
        self._message_handlers[msg_type].append(handler)

    async def execute_task(self, task: Dict[str, Any]) -> Dict[str, Any]:
        """执行任务"""
        skill_id = task.get("skill_id")
        operation = task.get("operation")
        input_data = task.get("input", {})

        self.state = AgentState.BUSY

        try:
            handler = self._skill_registry.get(skill_id)
            if not handler:
                raise ValueError(f"Skill not found: {skill_id}")

            # 支持同步和异步处理器
            if asyncio.iscoroutinefunction(handler):
                result = await handler(operation, input_data)
            else:
                result = handler(operation, input_data)

            self.state = AgentState.ONLINE
            return {
                "success": True,
                "result": result,
                "agent_id": self.info.id,
            }

        except Exception as e:
            self.state = AgentState.ERROR
            logger.error(f"Task execution failed: {e}")
            return {
                "success": False,
                "error": str(e),
                "agent_id": self.info.id,
            }

    async def send_message(self, target: str, msg_type: str, data: Any) -> None:
        """发送消息"""
        msg = {
            "type": "message",
            "from": self.info.id,
            "to": target,
            "msg_type": msg_type,
            "data": data,
            "timestamp": time.time(),
        }
        await self._connection.send(msg)

    async def broadcast(self, msg_type: str, data: Any) -> None:
        """广播消息"""
        msg = {
            "type": "broadcast",
            "from": self.info.id,
            "msg_type": msg_type,
            "data": data,
            "timestamp": time.time(),
        }
        await self._connection.send(msg)

    async def run(self) -> None:
        """运行Agent"""
        await self.connect()

        try:
            while self._running:
                msg = await self._connection.receive()
                if msg:
                    await self._handle_message(msg)
        except asyncio.CancelledError:
            pass
        finally:
            await self.disconnect()

    async def _handle_message(self, msg: Dict[str, Any]) -> None:
        """处理消息"""
        msg_type = msg.get("type")

        # 任务消息
        if msg_type == "task":
            result = await self.execute_task(msg)
            await self._connection.send({
                "type": "task_result",
                "task_id": msg.get("task_id"),
                **result,
            })

        # 其他消息
        handlers = self._message_handlers.get(msg_type, [])
        for handler in handlers:
            try:
                if asyncio.iscoroutinefunction(handler):
                    await handler(msg)
                else:
                    handler(msg)
            except Exception as e:
                logger.error(f"Message handler failed: {e}")

    @property
    def is_online(self) -> bool:
        """是否在线"""
        return self.state == AgentState.ONLINE

    def get_info(self) -> Dict[str, Any]:
        """获取Agent信息"""
        return {
            "id": self.info.id,
            "name": self.info.name,
            "type": self.info.type,
            "version": self.info.version,
            "platform": self.info.platform,
            "skills": self.info.skills,
            "state": self.state.value,
            "metadata": self.info.metadata,
            "last_heartbeat": self.info.last_heartbeat,
        }


async def create_agent(config: Optional[AgentConfig] = None) -> OFAAgent:
    """创建Agent实例"""
    agent = OFAAgent(config)
    return agent


def run_agent(config: Optional[AgentConfig] = None) -> None:
    """运行Agent（同步入口）"""
    agent = OFAAgent(config)
    asyncio.run(agent.run())