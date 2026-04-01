"""
OFA Python Agent SDK
Sprint 29: Multi-platform SDK Extension
v0.10.0: Offline + P2P + Constraint support
"""

from .agent import OFAAgent, AgentConfig, AgentState
from .skills import Skill, SkillExecutor, SkillRegistry
from .connection import Connection, ConnectionType
from .protocol import Message, MessageType, Protocol
from .builtin import EchoSkill, TextProcessSkill, CalculatorSkill, JSONSkill

# 新增模块
from .offline import (
    OfflineLevel,
    OfflineCache,
    LocalScheduler,
    OfflineManager,
    TaskStatus,
    OfflineTask,
)
from .p2p import (
    P2PClient,
    P2PDiscovery,
    P2PMessage,
    P2PMessageType,
    PeerInfo,
)
from .constraint import (
    ConstraintType,
    ConstraintResult,
    ConstraintRule,
    ConstraintEngine,
    ConstraintClient,
    ConstraintViolationError,
)

__version__ = "0.10.0"
__all__ = [
    # Core
    "OFAAgent",
    "AgentConfig",
    "AgentState",
    "Skill",
    "SkillExecutor",
    "SkillRegistry",
    "Connection",
    "ConnectionType",
    "Message",
    "MessageType",
    "Protocol",
    # Builtin skills
    "EchoSkill",
    "TextProcessSkill",
    "CalculatorSkill",
    "JSONSkill",
    # Offline
    "OfflineLevel",
    "OfflineCache",
    "LocalScheduler",
    "OfflineManager",
    "TaskStatus",
    "OfflineTask",
    # P2P
    "P2PClient",
    "P2PDiscovery",
    "P2PMessage",
    "P2PMessageType",
    "PeerInfo",
    # Constraint
    "ConstraintType",
    "ConstraintResult",
    "ConstraintRule",
    "ConstraintEngine",
    "ConstraintClient",
    "ConstraintViolationError",
]