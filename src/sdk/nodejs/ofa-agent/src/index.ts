/**
 * OFA Node.js Agent SDK
 * Sprint 29: Multi-platform SDK Extension
 */

export { OFAAgent, AgentConfig, AgentState, AgentInfo } from './agent';
export { Skill, SkillExecutor, SkillRegistry, SkillInfo } from './skills';
export { Connection, ConnectionType, HttpConnection, WebSocketConnection, GrpcConnection } from './connection';
export { Message, MessageType, Task, TaskResult } from './message';
export { Protocol } from './protocol';
export { registerBuiltinSkills, EchoSkill, TextProcessSkill, CalculatorSkill, JSONSkill } from './builtin';

export const VERSION = '8.1.0';