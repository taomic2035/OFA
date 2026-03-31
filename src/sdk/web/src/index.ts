/**
 * OFA Web Agent SDK
 * TypeScript/JavaScript implementation for browser-based agents
 *
 * @version 2.0.0
 * @author OFA Team
 */

// Types and Interfaces

export interface AgentConfig {
  centerUrl: string;
  wsUrl?: string;
  agentId?: string;
  agentName?: string;
  agentType?: string;
  capabilities?: SkillCapability[];
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
  heartbeatInterval?: number;
}

export interface SkillCapability {
  id: string;
  name: string;
  operations: string[];
  description?: string;
}

export interface Task {
  id: string;
  skillId: string;
  operation: string;
  params: Record<string, unknown>;
  priority?: number;
  timeout?: number;
}

export interface TaskResult {
  taskId: string;
  success: boolean;
  data?: unknown;
  error?: string;
  duration?: number;
}

export interface Message {
  id: string;
  fromAgent: string;
  toAgent: string;
  type: string;
  payload: unknown;
  timestamp: number;
}

export interface AgentStatus {
  id: string;
  name: string;
  type: string;
  state: AgentState;
  capabilities: SkillCapability[];
  lastSeen: number;
  load: number;
}

export type AgentState = 'online' | 'offline' | 'busy' | 'idle';

export type TaskHandler = (task: Task) => Promise<TaskResult>;
export type MessageHandler = (message: Message) => void;
export type StatusHandler = (status: AgentStatus) => void;
export type ErrorHandler = (error: Error) => void;

// Skill Interface

export interface Skill {
  id: string;
  name: string;
  execute(operation: string, params: Record<string, unknown>): Promise<unknown>;
}

// Built-in Skills

export class EchoSkill implements Skill {
  id = 'echo';
  name = 'Echo';

  async execute(operation: string, params: Record<string, unknown>): Promise<unknown> {
    return { message: params.message || 'Hello' };
  }
}

export class TextProcessSkill implements Skill {
  id = 'text.process';
  name = 'Text Process';

  async execute(operation: string, params: Record<string, unknown>): Promise<unknown> {
    const text = String(params.text || '');

    switch (operation) {
      case 'uppercase':
        return { result: text.toUpperCase() };
      case 'lowercase':
        return { result: text.toLowerCase() };
      case 'reverse':
        return { result: text.split('').reverse().join('') };
      case 'length':
        return { result: text.length };
      default:
        throw new Error(`Unknown operation: ${operation}`);
    }
  }
}

export class JSONProcessSkill implements Skill {
  id = 'json.process';
  name = 'JSON Process';

  async execute(operation: string, params: Record<string, unknown>): Promise<unknown> {
    const data = params.data;

    switch (operation) {
      case 'get_keys':
        if (typeof data === 'object' && data !== null) {
          return { result: Object.keys(data) };
        }
        throw new Error('Input must be an object');
      case 'get_values':
        if (typeof data === 'object' && data !== null) {
          return { result: Object.values(data) };
        }
        throw new Error('Input must be an object');
      case 'pretty':
        return { result: JSON.stringify(data, null, 2) };
      default:
        throw new Error(`Unknown operation: ${operation}`);
    }
  }
}

// Web Agent Class

export class OFAWebAgent {
  private config: AgentConfig;
  private ws: WebSocket | null = null;
  private skills: Map<string, Skill> = new Map();
  private taskHandlers: Map<string, TaskHandler> = new Map();
  private messageHandlers: MessageHandler[] = [];
  private statusHandlers: StatusHandler[] = [];
  private errorHandlers: ErrorHandler[] = [];

  private agentId: string;
  private agentName: string;
  private agentType: string;
  private state: AgentState = 'offline';
  private reconnectAttempts = 0;
  private heartbeatTimer: number | null = null;

  /**
   * Create a new Web Agent
   */
  constructor(config: AgentConfig) {
    this.config = {
      reconnectInterval: 5000,
      maxReconnectAttempts: 10,
      heartbeatInterval: 30000,
      agentType: 'web',
      ...config
    };

    this.agentId = config.agentId || this.generateId();
    this.agentName = config.agentName || `WebAgent-${this.agentId.slice(0, 8)}`;
    this.agentType = config.agentType || 'web';

    // Register built-in skills
    this.registerSkill(new EchoSkill());
    this.registerSkill(new TextProcessSkill());
    this.registerSkill(new JSONProcessSkill());

    // Register custom capabilities
    if (config.capabilities) {
      for (const cap of config.capabilities) {
        this.taskHandlers.set(cap.id, this.handleTask.bind(this));
      }
    }
  }

  /**
   * Generate unique agent ID
   */
  private generateId(): string {
    return 'web-' + Date.now().toString(36) + '-' + Math.random().toString(36).slice(2, 9);
  }

  /**
   * Connect to Center
   */
  async connect(): Promise<void> {
    const wsUrl = this.config.wsUrl || this.config.centerUrl.replace('http', 'ws') + '/ws';

    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
          this.state = 'online';
          this.reconnectAttempts = 0;
          this.sendRegister();
          this.startHeartbeat();
          this.notifyStatus();
          resolve();
        };

        this.ws.onmessage = (event) => {
          this.handleMessage(event.data);
        };

        this.ws.onerror = (error) => {
          this.notifyError(new Error('WebSocket error'));
          reject(error);
        };

        this.ws.onclose = () => {
          this.state = 'offline';
          this.stopHeartbeat();
          this.notifyStatus();
          this.attemptReconnect();
        };
      } catch (error) {
        reject(error);
      }
    });
  }

  /**
   * Disconnect from Center
   */
  async disconnect(): Promise<void> {
    this.stopHeartbeat();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.state = 'offline';
    this.notifyStatus();
  }

  /**
   * Send registration message
   */
  private sendRegister(): void {
    const capabilities = Array.from(this.skills.values()).map(skill => ({
      id: skill.id,
      name: skill.name,
      operations: this.getSkillOperations(skill.id)
    }));

    this.send({
      type: 'register',
      agentId: this.agentId,
      agentName: this.agentName,
      agentType: this.agentType,
      capabilities
    });
  }

  /**
   * Get operations for a skill
   */
  private getSkillOperations(skillId: string): string[] {
    // Default operations for built-in skills
    const defaultOps: Record<string, string[]> = {
      'echo': ['execute'],
      'text.process': ['uppercase', 'lowercase', 'reverse', 'length'],
      'json.process': ['get_keys', 'get_values', 'pretty']
    };
    return defaultOps[skillId] || [];
  }

  /**
   * Start heartbeat timer
   */
  private startHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
    }

    this.heartbeatTimer = window.setInterval(() => {
      if (this.state === 'online' && this.ws) {
        this.send({
          type: 'heartbeat',
          agentId: this.agentId,
          timestamp: Date.now()
        });
      }
    }, this.config.heartbeatInterval!);
  }

  /**
   * Stop heartbeat timer
   */
  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  /**
   * Attempt to reconnect
   */
  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.config.maxReconnectAttempts!) {
      this.notifyError(new Error('Max reconnect attempts reached'));
      return;
    }

    this.reconnectAttempts++;
    console.log(`Attempting reconnect (${this.reconnectAttempts}/${this.config.maxReconnectAttempts})`);

    setTimeout(() => {
      this.connect().catch(err => {
        this.notifyError(err);
      });
    }, this.config.reconnectInterval!);
  }

  /**
   * Send message to Center
   */
  private send(data: unknown): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data));
    }
  }

  /**
   * Handle incoming message
   */
  private handleMessage(data: string): void {
    try {
      const msg = JSON.parse(data);

      switch (msg.type) {
        case 'task':
          this.handleTaskMessage(msg);
          break;
        case 'message':
          this.handleAgentMessage(msg);
          break;
        case 'status':
          this.handleStatusUpdate(msg);
          break;
        default:
          console.log('Unknown message type:', msg.type);
      }
    } catch (error) {
      this.notifyError(new Error('Failed to parse message'));
    }
  }

  /**
   * Handle task message
   */
  private async handleTaskMessage(msg: Task): Promise<void> {
    this.state = 'busy';
    this.notifyStatus();

    const startTime = Date.now();
    let result: TaskResult;

    try {
      const data = await this.handleTask(msg);
      result = {
        taskId: msg.id,
        success: true,
        data,
        duration: Date.now() - startTime
      };
    } catch (error) {
      result = {
        taskId: msg.id,
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
        duration: Date.now() - startTime
      };
    }

    this.send({
      type: 'task_result',
      ...result
    });

    this.state = 'idle';
    this.notifyStatus();
  }

  /**
   * Handle task execution
   */
  private async handleTask(task: Task): Promise<unknown> {
    const skill = this.skills.get(task.skillId);
    if (skill) {
      return skill.execute(task.operation, task.params);
    }

    // Check custom handler
    const handler = this.taskHandlers.get(task.skillId);
    if (handler) {
      const result = await handler(task);
      return result.data;
    }

    throw new Error(`Unknown skill: ${task.skillId}`);
  }

  /**
   * Handle agent message
   */
  private handleAgentMessage(msg: Message): void {
    for (const handler of this.messageHandlers) {
      handler(msg);
    }
  }

  /**
   * Handle status update
   */
  private handleStatusUpdate(msg: AgentStatus): void {
    for (const handler of this.statusHandlers) {
      handler(msg);
    }
  }

  /**
   * Notify status handlers
   */
  private notifyStatus(): void {
    const status: AgentStatus = {
      id: this.agentId,
      name: this.agentName,
      type: this.agentType,
      state: this.state,
      capabilities: Array.from(this.skills.values()).map(s => ({
        id: s.id,
        name: s.name,
        operations: this.getSkillOperations(s.id)
      })),
      lastSeen: Date.now(),
      load: this.state === 'busy' ? 100 : 0
    };

    for (const handler of this.statusHandlers) {
      handler(status);
    }
  }

  /**
   * Notify error handlers
   */
  private notifyError(error: Error): void {
    for (const handler of this.errorHandlers) {
      handler(error);
    }
  }

  // Public API

  /**
   * Register a skill
   */
  registerSkill(skill: Skill): void {
    this.skills.set(skill.id, skill);
  }

  /**
   * Unregister a skill
   */
  unregisterSkill(skillId: string): void {
    this.skills.delete(skillId);
  }

  /**
   * Register task handler
   */
  onTask(skillId: string, handler: TaskHandler): void {
    this.taskHandlers.set(skillId, handler);
  }

  /**
   * Register message handler
   */
  onMessage(handler: MessageHandler): void {
    this.messageHandlers.push(handler);
  }

  /**
   * Register status handler
   */
  onStatus(handler: StatusHandler): void {
    this.statusHandlers.push(handler);
  }

  /**
   * Register error handler
   */
  onError(handler: ErrorHandler): void {
    this.errorHandlers.push(handler);
  }

  /**
   * Send message to another agent
   */
  sendMessage(toAgent: string, payload: unknown): void {
    this.send({
      type: 'agent_message',
      fromAgent: this.agentId,
      toAgent,
      payload,
      timestamp: Date.now()
    });
  }

  /**
   * Get agent ID
   */
  getAgentId(): string {
    return this.agentId;
  }

  /**
   * Get agent name
   */
  getAgentName(): string {
    return this.agentName;
  }

  /**
   * Get agent state
   */
  getState(): AgentState {
    return this.state;
  }

  /**
   * Get registered skills
   */
  getSkills(): Skill[] {
    return Array.from(this.skills.values());
  }
}

// Factory function for easy creation

export function createWebAgent(config: AgentConfig): OFAWebAgent {
  return new OFAWebAgent(config);
}

// Default export

export default OFAWebAgent;