"""
OFA Python SDK
"""

import asyncio
import json
import time
import uuid
from typing import Dict, Any, Optional, Callable, Awaitable

import grpc
import ofa_pb2
import ofa_pb2_grpc


class Skill:
    """Base class for skills"""

    @property
    def id(self) -> str:
        raise NotImplementedError

    @property
    def name(self) -> str:
        raise NotImplementedError

    @property
    def version(self) -> str:
        return "1.0.0"

    @property
    def category(self) -> str:
        return "general"

    async def execute(self, input_data: Dict[str, Any]) -> Dict[str, Any]:
        raise NotImplementedError


class Message:
    """Represents a message"""

    def __init__(self, msg_id: str, from_agent: str, to_agent: str,
                 action: str, payload: bytes, timestamp: int):
        self.msg_id = msg_id
        self.from_agent = from_agent
        self.to_agent = to_agent
        self.action = action
        self.payload = payload
        self.timestamp = timestamp

    @property
    def payload_json(self) -> Dict[str, Any]:
        return json.loads(self.payload) if self.payload else {}


AGENT_TYPES = {
    "unknown": 0,
    "full": 1,
    "mobile": 2,
    "lite": 3,
    "iot": 4,
    "edge": 5,
}


class Agent:
    """OFA Agent Client"""

    def __init__(self, center_addr: str, name: str, agent_type: str = "full"):
        self.center_addr = center_addr
        self.name = name
        self.agent_type = AGENT_TYPES.get(agent_type, 1)

        self._channel: Optional[grpc.aio.Channel] = None
        self._stub: Optional[ofa_pb2_grpc.AgentServiceStub] = None
        self._stream = None

        self._agent_id: Optional[str] = None
        self._token: Optional[str] = None

        self._skills: Dict[str, Skill] = {}
        self._message_handler: Optional[Callable[[Message], Awaitable[None]]] = None

        self._running = False
        self._tasks: Dict[str, asyncio.Task] = {}

    @property
    def id(self) -> str:
        return self._agent_id

    def register_skill(self, skill: Skill):
        """Register a skill"""
        self._skills[skill.id] = skill

    def set_message_handler(self, handler: Callable[[Message], Awaitable[None]]):
        """Set message handler"""
        self._message_handler = handler

    async def connect(self):
        """Connect to Center"""
        self._channel = grpc.aio.insecure_channel(self.center_addr)
        self._stub = ofa_pb2_grpc.AgentServiceStub(self._channel)

        self._stream = self._stub.Connect()

        # Send registration
        await self._stream.send(ofa_pb2.AgentMessage(
            msg_id=str(uuid.uuid4()),
            timestamp=int(time.time()),
            register=ofa_pb2.RegisterRequest(
                name=self.name,
                type=self.agent_type,
                capabilities=self._get_capabilities(),
            )
        ))

        # Wait for response
        response = await self._stream.recv()
        if response.register and response.register.success:
            self._agent_id = response.register.agent_id
            self._token = response.register.token
            self._running = True

            # Start background tasks
            self._tasks['receive'] = asyncio.create_task(self._receive_loop())
            self._tasks['heartbeat'] = asyncio.create_task(
                self._heartbeat_loop(response.register.heartbeat_interval_ms / 1000)
            )
        else:
            raise Exception(f"Registration failed: {response.register.error if response.register else 'No response'}")

    async def disconnect(self):
        """Disconnect from Center"""
        self._running = False

        for task in self._tasks.values():
            task.cancel()

        if self._stream:
            await self._stream.done()

        if self._channel:
            await self._channel.close()

    async def send_message(self, to_agent: str, action: str, payload: Dict[str, Any]):
        """Send message to another agent"""
        stub = ofa_pb2_grpc.MessageServiceStub(self._channel)
        await stub.SendMessage(ofa_pb2.SendMessageRequest(
            message=ofa_pb2.Message(
                from_agent=self._agent_id,
                to_agent=to_agent,
                action=action,
                payload=json.dumps(payload).encode(),
                timestamp=int(time.time()),
            )
        ))

    def _get_capabilities(self) -> list:
        """Get capabilities from registered skills"""
        return [
            ofa_pb2.Capability(
                id=skill.id,
                name=skill.name,
                version=skill.version,
                category=skill.category,
            )
            for skill in self._skills.values()
        ]

    async def _receive_loop(self):
        """Receive messages from Center"""
        while self._running:
            try:
                msg = await self._stream.recv()

                if msg.task:
                    asyncio.create_task(self._handle_task(msg.task))

                if msg.message:
                    await self._handle_message(msg.message)

                if msg.cancel_task:
                    # Handle task cancellation
                    pass

            except Exception as e:
                if self._running:
                    print(f"Receive error: {e}")

    async def _heartbeat_loop(self, interval: float):
        """Send heartbeats to Center"""
        while self._running:
            try:
                await self._stream.send(ofa_pb2.AgentMessage(
                    msg_id=str(uuid.uuid4()),
                    timestamp=int(time.time()),
                    heartbeat=ofa_pb2.HeartbeatRequest(
                        agent_id=self._agent_id,
                        status=1,  # Online
                    )
                ))
            except Exception as e:
                print(f"Heartbeat error: {e}")

            await asyncio.sleep(interval)

    async def _handle_task(self, task):
        """Handle incoming task"""
        skill = self._skills.get(task.skill_id)
        if not skill:
            await self._send_task_result(
                task.task_id,
                status=4,  # Failed
                error=f"Skill not found: {task.skill_id}"
            )
            return

        try:
            input_data = json.loads(task.input) if task.input else {}
            result = await skill.execute(input_data)

            await self._send_task_result(
                task.task_id,
                status=3,  # Completed
                output=json.dumps(result).encode()
            )
        except Exception as e:
            await self._send_task_result(
                task.task_id,
                status=4,  # Failed
                error=str(e)
            )

    async def _handle_message(self, msg):
        """Handle incoming message"""
        if self._message_handler:
            message = Message(
                msg_id=msg.msg_id,
                from_agent=msg.from_agent,
                to_agent=msg.to_agent,
                action=msg.action,
                payload=msg.payload,
                timestamp=msg.timestamp,
            )

            try:
                await self._message_handler(message)
                success = True
                error = ""
            except Exception as e:
                success = False
                error = str(e)

            await self._stream.send(ofa_pb2.AgentMessage(
                msg_id=str(uuid.uuid4()),
                timestamp=int(time.time()),
                message_response=ofa_pb2.MessageResponse(
                    msg_id=msg.msg_id,
                    success=success,
                    error=error,
                )
            ))

    async def _send_task_result(self, task_id: str, status: int,
                                 output: bytes = None, error: str = ""):
        """Send task result to Center"""
        await self._stream.send(ofa_pb2.AgentMessage(
            msg_id=str(uuid.uuid4()),
            timestamp=int(time.time()),
            task_result=ofa_pb2.TaskResult(
                task_id=task_id,
                status=status,
                output=output or b"",
                error=error,
            )
        ))