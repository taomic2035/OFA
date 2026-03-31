/**
 * Built-in Skills
 * Sprint 29: Node.js Agent SDK
 */

import { SkillInfo, SkillHandler } from './skills';

/**
 * Echo技能
 */
export class EchoSkill implements SkillHandler {
  info: SkillInfo = {
    id: 'echo',
    name: 'Echo',
    description: 'Echo input',
    operations: ['echo', 'ping'],
    version: '1.0.0',
    author: 'OFA',
    tags: ['utility', 'test'],
  };

  async execute(operation: string, input: unknown): Promise<unknown> {
    if (operation === 'ping') {
      return { pong: true, timestamp: Date.now() };
    }
    return input;
  }
}

/**
 * 文本处理技能
 */
export class TextProcessSkill implements SkillHandler {
  info: SkillInfo = {
    id: 'text.process',
    name: 'Text Process',
    description: 'Text processing operations',
    operations: ['uppercase', 'lowercase', 'reverse', 'length', 'trim', 'split'],
    version: '1.0.0',
    author: 'OFA',
    tags: ['text', 'utility'],
  };

  async execute(operation: string, input: unknown): Promise<unknown> {
    const data = input as { text?: string; separator?: string };
    const text = data.text || '';

    switch (operation) {
      case 'uppercase':
        return text.toUpperCase();
      case 'lowercase':
        return text.toLowerCase();
      case 'reverse':
        return text.split('').reverse().join('');
      case 'length':
        return text.length;
      case 'trim':
        return text.trim();
      case 'split':
        return text.split(data.separator || ' ');
      default:
        return text;
    }
  }
}

/**
 * 计算器技能
 */
export class CalculatorSkill implements SkillHandler {
  info: SkillInfo = {
    id: 'calculator',
    name: 'Calculator',
    description: 'Math calculations',
    operations: ['add', 'sub', 'mul', 'div', 'pow', 'sqrt', 'mod', 'abs'],
    version: '1.0.0',
    author: 'OFA',
    tags: ['math', 'utility'],
  };

  async execute(operation: string, input: unknown): Promise<unknown> {
    const data = input as { a?: number; b?: number };
    const a = data.a ?? 0;
    const b = data.b ?? 0;

    switch (operation) {
      case 'add':
        return a + b;
      case 'sub':
        return a - b;
      case 'mul':
        return a * b;
      case 'div':
        if (b === 0) throw new Error('Division by zero');
        return a / b;
      case 'pow':
        return Math.pow(a, b);
      case 'sqrt':
        return Math.sqrt(a);
      case 'mod':
        return a % b;
      case 'abs':
        return Math.abs(a);
      default:
        throw new Error(`Unknown operation: ${operation}`);
    }
  }
}

/**
 * JSON处理技能
 */
export class JSONSkill implements SkillHandler {
  info: SkillInfo = {
    id: 'json.process',
    name: 'JSON Process',
    description: 'JSON data processing',
    operations: ['parse', 'stringify', 'get_keys', 'get_values', 'get', 'set', 'validate'],
    version: '1.0.0',
    author: 'OFA',
    tags: ['json', 'data'],
  };

  async execute(operation: string, input: unknown): Promise<unknown> {
    const data = input as { data?: unknown; path?: string; value?: unknown; indent?: number };
    const obj = data.data;

    switch (operation) {
      case 'parse':
        return typeof obj === 'string' ? JSON.parse(obj) : obj;
      case 'stringify':
        return JSON.stringify(obj, null, data.indent ?? 2);
      case 'get_keys':
        return obj && typeof obj === 'object' ? Object.keys(obj) : [];
      case 'get_values':
        return obj && typeof obj === 'object' ? Object.values(obj) : [];
      case 'validate':
        try {
          if (typeof obj === 'string') JSON.parse(obj);
          else JSON.stringify(obj);
          return { valid: true };
        } catch {
          return { valid: false };
        }
      default:
        return obj;
    }
  }
}

/**
 * 内置技能列表
 */
export const BUILTIN_SKILLS = [
  new EchoSkill(),
  new TextProcessSkill(),
  new CalculatorSkill(),
  new JSONSkill(),
];

/**
 * 注册内置技能
 */
export function registerBuiltinSkills(executor: { register: (id: string, handler: SkillHandler) => void }): void {
  for (const skill of BUILTIN_SKILLS) {
    const handler = skill.execute.bind(skill);
    executor.register(skill.info.id, handler);
  }
}