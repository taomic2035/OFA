/**
 * OFA Node.js Agent SDK
 * Sprint 29: Multi-platform SDK Extension
 * v0.10.0: Offline + P2P + Constraint support
 */

export { OFAAgent, AgentConfig, AgentState, AgentInfo } from './agent';
export { Skill, SkillExecutor, SkillRegistry, SkillInfo } from './skills';
export { Connection, ConnectionType, HttpConnection, WebSocketConnection, GrpcConnection } from './connection';
export { Message, MessageType, Task, TaskResult } from './message';
export { Protocol } from './protocol';
export { registerBuiltinSkills, EchoSkill, TextProcessSkill, CalculatorSkill, JSONSkill } from './builtin';

// 离线模块
export {
  OfflineLevel,
  TaskStatus,
  LocalTask,
  LocalScheduler,
  OfflineCache,
  OfflineManager,
} from './offline';

// P2P 模块
export {
  P2PMessageType,
  P2PMessage,
  PeerInfo,
  P2PClient,
} from './p2p';

// 约束模块
export {
  ConstraintType,
  ConstraintResult,
  ConstraintRule,
  ConstraintChecker,
  ConstraintClient,
} from './constraint';

export const VERSION = '0.10.0';