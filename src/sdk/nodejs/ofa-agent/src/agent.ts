/**
 * Agent Core Module
 * Sprint 29: Node.js Agent SDK
 */

import { v4 as uuidv4 } from 'uuid';
import { AgentConfig, DefaultAgentConfig } from './config';
import { Connection, createConnection } from './connection';
import { SkillExecutor, SkillRegistry } from './skills';
import { Message, MessageType, Task, TaskResult } from './message';

/**
 * Agent状态
 */
export enum AgentState {
  INITIALIZING = 'initializing',
  CONNECTING = 'connecting',
  ONLINE = 'online',
  BUSY = 'busy',
  OFFLINE = 'offline',
  ERROR = 'error',
}

/**
 * Agent信息
 */
export interface AgentInfo {
  id: string;
  name: string;
  type: string;
  version: string;
  platform: string;
  skills: string[];
  state: AgentState;
  metadata: Record<string, unknown>;
  lastHeartbeat: number;
}

/**
 * Agent统计
 */
export interface AgentStats {
  tasksExecuted: number;
  tasksSuccess: number;
  tasksFailed: number;
  messagesSent: number;
  messagesReceived: number;
}

/**
 * OFA Agent
 */
export class OFAAgent {
  private config: AgentConfig;
  private _state: AgentState = AgentState.INITIALIZING;
  private _connection: Connection | null = null;
  private _skillExecutor: SkillExecutor;
  private _skillRegistry: SkillRegistry;
  private _stats: AgentStats;
  private _running: boolean = false;
  private _heartbeatTimer?: NodeJS.Timeout;
  private _messageHandlers: Map<string, Set<(msg: Message) => void>> = new Map();

  /**
   * Agent ID
   */
  public readonly id: string;

  /**
   * Agent信息
   */
  public readonly info: AgentInfo;

  constructor(config?: Partial<AgentConfig>) {
    this.config = { ...DefaultAgentConfig(), ...config };
    this.id = this.config.agentId || `nodejs-agent-${uuidv4().split('-')[0]}`;

    this.info = {
      id: this.id,
      name: this.config.name,
      type: 'nodejs',
      version: '8.1.0',
      platform: 'nodejs',
      skills: [],
      state: AgentState.INITIALIZING,
      metadata: {},
      lastHeartbeat: 0,
    };

    this._skillExecutor = new SkillExecutor();
    this._skillRegistry = new SkillRegistry();
    this._stats = {
      tasksExecuted: 0,
      tasksSuccess: 0,
      tasksFailed: 0,
      messagesSent: 0,
      messagesReceived: 0,
    };
  }

  /**
   * 获取当前状态
   */
  get state(): AgentState {
    return this._state;
  }

  /**
   * 是否在线
   */
  get isOnline(): boolean {
    return this._state === AgentState.ONLINE;
  }

  /**
   * 连接到Center
   */
  async connect(): Promise<boolean> {
    this._state = AgentState.CONNECTING;
    console.log(`Connecting to ${this.config.centerUrl}`);

    try {
      this._connection = createConnection(this.config);
      await this._connection.connect();
      await this._register();
      this._state = AgentState.ONLINE;
      this._running = true;
      this._startHeartbeat();

      console.log(`Agent connected: ${this.id}`);
      return true;
    } catch (error) {
      this._state = AgentState.ERROR;
      console.error('Connection failed:', error);
      return false;
    }
  }

  /**
   * 断开连接
   */
  async disconnect(): Promise<void> {
    this._running = false;

    if (this._heartbeatTimer) {
      clearInterval(this._heartbeatTimer);
    }

    if (this._connection) {
      await this._connection.disconnect();
    }

    this._state = AgentState.OFFLINE;
    console.log(`Agent disconnected: ${this.id}`);
  }

  /**
   * 注册到Center
   */
  private async _register(): Promise<void> {
    const msg: Message = {
      id: uuidv4(),
      type: MessageType.REGISTER,
      from: this.id,
      to: '',
      subject: '',
      data: { ...this.info },
      timestamp: Date.now(),
    };
    await this._connection!.send(msg);
  }

  /**
   * 启动心跳
   */
  private _startHeartbeat(): void {
    this._heartbeatTimer = setInterval(async () => {
      if (!this._running || !this._connection) return;

      const msg: Message = {
        id: uuidv4(),
        type: MessageType.HEARTBEAT,
        from: this.id,
        to: '',
        subject: '',
        data: { state: this._state },
        timestamp: Date.now(),
      };

      try {
        await this._connection.send(msg);
        this.info.lastHeartbeat = Date.now();
      } catch (error) {
        console.error('Heartbeat failed:', error);
        await this._reconnect();
      }
    }, this.config.heartbeatInterval * 1000);
  }

  /**
   * 重连
   */
  private async _reconnect(): Promise<void> {
    for (let i = 0; i < this.config.maxReconnectAttempts; i++) {
      this._state = AgentState.CONNECTING;
      console.log(`Reconnecting (attempt ${i + 1})`);

      try {
        await this._connection?.connect();
        await this._register();
        this._state = AgentState.ONLINE;
        this._running = true;
        console.log('Reconnected successfully');
        return;
      } catch (error) {
        console.error('Reconnect failed:', error);
        await new Promise(r => setTimeout(r, this.config.reconnectInterval * 1000));
      }
    }

    this._state = AgentState.ERROR;
    console.error('Max reconnect attempts reached');
  }

  /**
   * 注册技能
   */
  registerSkill(skillId: string, handler: (operation: string, input: unknown) => Promise<unknown>): void {
    this._skillExecutor.register(skillId, handler);
    this.info.skills.push(skillId);
    console.log(`Skill registered: ${skillId}`);
  }

  /**
   * 注销技能
   */
  unregisterSkill(skillId: string): void {
    this._skillExecutor.unregister(skillId);
    this.info.skills = this.info.skills.filter(s => s !== skillId);
  }

  /**
   * 执行任务
   */
  async executeTask(task: Task): Promise<TaskResult> {
    this._state = AgentState.BUSY;

    const start = Date.now();
    let result: TaskResult;

    try {
      const output = await this._skillExecutor.execute(task.skillId, task.operation, task.input);
      result = {
        taskId: task.taskId,
        success: true,
        result: output,
        agentId: this.id,
        durationMs: Date.now() - start,
      };
      this._stats.tasksSuccess++;
    } catch (error) {
      result = {
        taskId: task.taskId,
        success: false,
        error: String(error),
        agentId: this.id,
        durationMs: Date.now() - start,
      };
      this._stats.tasksFailed++;
    }

    this._stats.tasksExecuted++;
    this._state = AgentState.ONLINE;
    return result;
  }

  /**
   * 发送消息
   */
  async sendMessage(target: string, type: MessageType, data: unknown): Promise<void> {
    if (!this._connection) throw new Error('Not connected');

    const msg: Message = {
      id: uuidv4(),
      type,
      from: this.id,
      to: target,
      subject: '',
      data,
      timestamp: Date.now(),
    };

    await this._connection.send(msg);
    this._stats.messagesSent++;
  }

  /**
   * 广播消息
   */
  async broadcast(type: MessageType, data: unknown): Promise<void> {
    await this.sendMessage('', type, data);
  }

  /**
   * 注册消息处理器
   */
  onMessage(type: string, handler: (msg: Message) => void): void {
    if (!this._messageHandlers.has(type)) {
      this._messageHandlers.set(type, new Set());
    }
    this._messageHandlers.get(type)!.add(handler);
  }

  /**
   * 运行Agent
   */
  async run(): Promise<void> {
    await this.connect();

    while (this._running && this._connection) {
      try {
        const msg = await this._connection.receive();
        if (msg) {
          await this._handleMessage(msg);
          this._stats.messagesReceived++;
        }
      } catch (error) {
        console.error('Message handling error:', error);
      }

      await new Promise(r => setTimeout(r, 100));
    }
  }

  /**
   * 处理消息
   */
  private async _handleMessage(msg: Message): Promise<void> {
    if (msg.type === MessageType.TASK) {
      const task = msg.data as Task;
      const result = await this.executeTask(task);

      await this._connection?.send({
        id: uuidv4(),
        type: MessageType.TASK_RESULT,
        from: this.id,
        to: '',
        subject: '',
        data: result,
        timestamp: Date.now(),
      });
    }

    // 调用自定义处理器
    const handlers = this._messageHandlers.get(msg.type);
    if (handlers) {
      for (const handler of handlers) {
        try {
          handler(msg);
        } catch (error) {
          console.error('Handler error:', error);
        }
      }
    }
  }

  /**
   * 获取统计
   */
  getStats(): AgentStats {
    return { ...this._stats };
  }
}

// 导出配置
export { AgentConfig, DefaultAgentConfig } from './config';