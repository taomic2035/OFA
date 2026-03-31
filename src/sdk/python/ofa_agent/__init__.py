"""
OFA Python Agent SDK
Sprint 29: Multi-platform SDK Extension
"""

from .agent import OFAAgent, AgentConfig, AgentState
from .skills import Skill, SkillExecutor, SkillRegistry
from .connection import Connection, ConnectionType
from .protocol import Message, MessageType, Protocol
from .builtin import EchoSkill, TextProcessSkill, CalculatorSkill, JSONSkill

__version__ = "8.1.0"
__all__ = [
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
    "EchoSkill",
    "TextProcessSkill",
    "CalculatorSkill",
    "JSONSkill",
]