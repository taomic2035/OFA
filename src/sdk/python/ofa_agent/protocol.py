"""
Protocol Module
Sprint 29: Python Agent SDK
"""

import json
import time
import uuid
from enum import Enum
from typing import Any, Dict, Optional
from dataclasses import dataclass, field


class MessageType(Enum):
    """消息类型"""
    REGISTER = "register"
    HEARTBEAT = "heartbeat"
    TASK = "task"
    TASK_RESULT = "task_result"
    MESSAGE = "message"
    BROADCAST = "broadcast"
    DISCOVERY = "discovery"
    ERROR = "error"
    ACK = "ack"


@dataclass
class Message:
    """消息"""
    id: str = field(default_factory=lambda: uuid.uuid4().hex)
    type: MessageType = MessageType.MESSAGE
    from_agent: str = ""
    to_agent: str = ""
    subject: str = ""
    data: Any = None
    timestamp: float = field(default_factory=time.time)
    metadata: Dict[str, Any] = field(default_factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        return {
            "id": self.id,
            "type": self.type.value,
            "from": self.from_agent,
            "to": self.to_agent,
            "subject": self.subject,
            "data": self.data,
            "timestamp": self.timestamp,
            "metadata": self.metadata,
        }

    def to_json(self) -> str:
        return json.dumps(self.to_dict())

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "Message":
        return cls(
            id=data.get("id", uuid.uuid4().hex),
            type=MessageType(data.get("type", "message")),
            from_agent=data.get("from", ""),
            to_agent=data.get("to", ""),
            subject=data.get("subject", ""),
            data=data.get("data"),
            timestamp=data.get("timestamp", time.time()),
            metadata=data.get("metadata", {}),
        )

    @classmethod
    def from_json(cls, json_str: str) -> "Message":
        return cls.from_dict(json.loads(json_str))


@dataclass
class TaskMessage:
    """任务消息"""
    task_id: str = field(default_factory=lambda: uuid.uuid4().hex)
    skill_id: str = ""
    operation: str = ""
    input_data: Dict[str, Any] = field(default_factory=dict)
    priority: int = 0
    timeout: float = 30.0
    callback: Optional[str] = None

    def to_dict(self) -> Dict[str, Any]:
        return {
            "type": "task",
            "task_id": self.task_id,
            "skill_id": self.skill_id,
            "operation": self.operation,
            "input": self.input_data,
            "priority": self.priority,
            "timeout": self.timeout,
            "callback": self.callback,
        }


@dataclass
class TaskResult:
    """任务结果"""
    task_id: str = ""
    success: bool = False
    result: Any = None
    error: str = ""
    agent_id: str = ""
    duration: float = 0.0

    def to_dict(self) -> Dict[str, Any]:
        return {
            "type": "task_result",
            "task_id": self.task_id,
            "success": self.success,
            "result": self.result,
            "error": self.error,
            "agent_id": self.agent_id,
            "duration": self.duration,
        }


class Protocol:
    """协议处理"""

    VERSION = "8.1.0"
    MAGIC = "OFA"
    HEADER_SIZE = 16

    @staticmethod
    def encode(msg: Message) -> bytes:
        """编码消息"""
        json_data = msg.to_json().encode("utf-8")
        header = Protocol._make_header(len(json_data), msg.type.value)
        return header + json_data

    @staticmethod
    def decode(data: bytes) -> Optional[Message]:
        """解码消息"""
        if len(data) < Protocol.HEADER_SIZE:
            return None

        header = data[:Protocol.HEADER_SIZE]
        body = data[Protocol.HEADER_SIZE:]

        msg_type, length = Protocol._parse_header(header)
        if length != len(body):
            return None

        return Message.from_json(body.decode("utf-8"))

    @staticmethod
    def _make_header(length: int, msg_type: str) -> bytes:
        """创建头部"""
        # 4字节魔数 + 4字节类型 + 4字节长度 + 4字节版本
        magic = Protocol.MAGIC.encode("utf-8")
        type_bytes = msg_type.encode("utf-8").ljust(4, b"\0")[:4]
        length_bytes = length.to_bytes(4, "big")
        version = Protocol.VERSION.encode("utf-8").ljust(4, b"\0")[:4]
        return magic + type_bytes + length_bytes + version

    @staticmethod
    def _parse_header(header: bytes) -> tuple:
        """解析头部"""
        magic = header[:4].decode("utf-8")
        if magic != Protocol.MAGIC:
            raise ValueError("Invalid magic number")

        msg_type = header[4:8].decode("utf-8").rstrip("\0")
        length = int.from_bytes(header[8:12], "big")
        version = header[12:16].decode("utf-8").rstrip("\0")

        return msg_type, length


class BinaryProtocol:
    """二进制协议（轻量级）"""

    FRAME_START = 0xAA
    FRAME_END = 0x55

    @staticmethod
    def encode(msg_type: int, data: bytes) -> bytes:
        """编码"""
        length = len(data)
        frame = bytearray()
        frame.append(BinaryProtocol.FRAME_START)
        frame.append(msg_type)
        frame.extend(length.to_bytes(2, "big"))
        frame.extend(data)
        frame.append(BinaryProtocol.FRAME_END)
        return bytes(frame)

    @staticmethod
    def decode(data: bytes) -> Optional[tuple]:
        """解码"""
        if len(data) < 5:
            return None
        if data[0] != BinaryProtocol.FRAME_START:
            return None

        msg_type = data[1]
        length = int.from_bytes(data[2:4], "big")
        payload = data[4:4+length]

        if len(data) < 4 + length + 1:
            return None
        if data[4+length] != BinaryProtocol.FRAME_END:
            return None

        return msg_type, payload