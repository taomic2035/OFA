"""
Connection Module
Sprint 29: Python Agent SDK
"""

import asyncio
import json
import logging
import aiohttp
import websockets
from typing import Any, Dict, Optional
from dataclasses import dataclass

logger = logging.getLogger(__name__)


class ConnectionType:
    """连接类型"""
    GRPC = "grpc"
    HTTP = "http"
    WEBSOCKET = "websocket"


class Connection:
    """连接基类"""

    def __init__(self, config):
        self.config = config
        self.connected = False

    async def connect(self) -> None:
        raise NotImplementedError

    async def disconnect(self) -> None:
        raise NotImplementedError

    async def send(self, msg: Dict[str, Any]) -> None:
        raise NotImplementedError

    async def receive(self) -> Optional[Dict[str, Any]]:
        raise NotImplementedError


class HTTPConnection(Connection):
    """HTTP连接"""

    def __init__(self, config):
        super().__init__(config)
        self._session: Optional[aiohttp.ClientSession] = None
        self._base_url = f"http://{config.center_url}"
        self._poll_interval = 1.0
        self._message_queue: asyncio.Queue = asyncio.Queue()

    async def connect(self) -> None:
        self._session = aiohttp.ClientSession(
            timeout=aiohttp.ClientTimeout(total=self.config.timeout)
        )

        # 注册Agent
        register_url = f"{self._base_url}/api/v1/agents/register"
        async with self._session.post(register_url, json={
            "id": self.config.agent_id,
            "name": self.config.name,
            "type": "python",
        }) as resp:
            if resp.status == 200:
                self.connected = True
                logger.info(f"HTTP connection established: {self._base_url}")
                # 启动消息轮询
                asyncio.create_task(self._poll_messages())
            else:
                raise ConnectionError(f"Registration failed: {resp.status}")

    async def disconnect(self) -> None:
        self.connected = False
        if self._session:
            await self._session.close()
            self._session = None

    async def send(self, msg: Dict[str, Any]) -> None:
        if not self._session or not self.connected:
            raise ConnectionError("Not connected")

        msg_type = msg.get("type", "message")
        url = f"{self._base_url}/api/v1/{msg_type}"

        async with self._session.post(url, json=msg) as resp:
            if resp.status != 200:
                logger.error(f"Send failed: {resp.status}")

    async def receive(self) -> Optional[Dict[str, Any]]:
        try:
            msg = await asyncio.wait_for(
                self._message_queue.get(),
                timeout=0.1
            )
            return msg
        except asyncio.TimeoutError:
            return None

    async def _poll_messages(self) -> None:
        """轮询消息"""
        url = f"{self._base_url}/api/v1/agents/{self.config.agent_id}/messages"

        while self.connected:
            try:
                async with self._session.get(url) as resp:
                    if resp.status == 200:
                        messages = await resp.json()
                        for msg in messages:
                            await self._message_queue.put(msg)
                await asyncio.sleep(self._poll_interval)
            except Exception as e:
                logger.error(f"Poll error: {e}")
                await asyncio.sleep(self._poll_interval)


class WebSocketConnection(Connection):
    """WebSocket连接"""

    def __init__(self, config):
        super().__init__(config)
        self._ws: Optional[websockets.WebSocketClientProtocol] = None
        self._url = f"ws://{config.center_url}/ws"
        self._message_queue: asyncio.Queue = asyncio.Queue()

    async def connect(self) -> None:
        self._ws = await websockets.connect(
            self._url,
            ping_interval=30,
            ping_timeout=10
        )
        self.connected = True
        logger.info(f"WebSocket connection established: {self._url}")

        # 发送注册消息
        await self.send({
            "type": "register",
            "id": self.config.agent_id,
            "name": self.config.name,
        })

        # 启动消息接收
        asyncio.create_task(self._receive_loop())

    async def disconnect(self) -> None:
        self.connected = False
        if self._ws:
            await self._ws.close()
            self._ws = None

    async def send(self, msg: Dict[str, Any]) -> None:
        if not self._ws or not self.connected:
            raise ConnectionError("Not connected")

        await self._ws.send(json.dumps(msg))

    async def receive(self) -> Optional[Dict[str, Any]]:
        try:
            msg = await asyncio.wait_for(
                self._message_queue.get(),
                timeout=0.1
            )
            return msg
        except asyncio.TimeoutError:
            return None

    async def _receive_loop(self) -> None:
        """接收循环"""
        while self.connected and self._ws:
            try:
                data = await self._ws.recv()
                msg = json.loads(data)
                await self._message_queue.put(msg)
            except websockets.ConnectionClosed:
                self.connected = False
                logger.info("WebSocket closed")
                break
            except Exception as e:
                logger.error(f"Receive error: {e}")


class GRPCConnection(Connection):
    """gRPC连接（模拟实现）"""

    def __init__(self, config):
        super().__init__(config)
        self._channel: Optional[Any] = None
        self._stub: Optional[Any] = None
        self._message_queue: asyncio.Queue = asyncio.Queue()

    async def connect(self) -> None:
        # gRPC需要grpcio库
        try:
            import grpc
            from grpc import aio

            # 创建channel
            self._channel = aio.insecure_channel(self.config.center_url)
            self.connected = True
            logger.info(f"gRPC connection established: {self.config.center_url}")

        except ImportError:
            logger.warning("grpcio not installed, using HTTP fallback")
            # 回退到HTTP
            http_conn = HTTPConnection(self.config)
            await http_conn.connect()
            self._session = http_conn._session
            self.connected = True

    async def disconnect(self) -> None:
        self.connected = False
        if self._channel:
            await self._channel.close()

    async def send(self, msg: Dict[str, Any]) -> None:
        if not self.connected:
            raise ConnectionError("Not connected")

        # 模拟发送
        logger.debug(f"Sending: {msg}")

    async def receive(self) -> Optional[Dict[str, Any]]:
        try:
            msg = await asyncio.wait_for(
                self._message_queue.get(),
                timeout=0.1
            )
            return msg
        except asyncio.TimeoutError:
            return None