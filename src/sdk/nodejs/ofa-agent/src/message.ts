/**
 * Message Module
 * Sprint 29: Node.js Agent SDK
 */

import { v4 as uuidv4 } from 'uuid';

/**
 * 消息类型
 */
export enum MessageType {
  REGISTER = 'register',
  HEARTBEAT = 'heartbeat',
  TASK = 'task',
  TASK_RESULT = 'task_result',
  MESSAGE = 'message',
  BROADCAST = 'broadcast',
  DISCOVERY = 'discovery',
  ERROR = 'error',
  ACK = 'ack',
}

/**
 * 消息
 */
export interface Message {
  id: string;
  type: MessageType;
  from: string;
  to: string;
  subject: string;
  data: unknown;
  timestamp: number;
}

/**
 * 创建消息
 */
export function createMessage(type: MessageType, from: string, data: unknown): Message {
  return {
    id: uuidv4(),
    type,
    from,
    to: '',
    subject: '',
    data,
    timestamp: Date.now(),
  };
}

/**
 * 任务
 */
export interface Task {
  taskId: string;
  skillId: string;
  operation: string;
  input: unknown;
  priority: number;
  timeout: number;
}

/**
 * 创建任务
 */
export function createTask(skillId: string, operation: string, input: unknown): Task {
  return {
    taskId: uuidv4(),
    skillId,
    operation,
    input,
    priority: 0,
    timeout: 30,
  };
}

/**
 * 任务结果
 */
export interface TaskResult {
  taskId: string;
  success: boolean;
  result?: unknown;
  error?: string;
  agentId: string;
  durationMs: number;
}