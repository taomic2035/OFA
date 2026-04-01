/**
 * OFA Node.js SDK - 离线模块
 * 支持离线能力等级 L1-L4
 */

import { v4 as uuidv4 } from 'uuid';

/**
 * 离线能力等级
 */
export enum OfflineLevel {
  NONE = 0,   // 不支持离线
  L1 = 1,     // 完全离线 (本地执行)
  L2 = 2,     // 局域网协作
  L3 = 3,     // 弱网同步
  L4 = 4,     // 在线模式
}

/**
 * 任务状态
 */
export enum TaskStatus {
  PENDING = 'pending',
  RUNNING = 'running',
  COMPLETED = 'completed',
  FAILED = 'failed',
  CANCELLED = 'cancelled',
}

/**
 * 离线任务
 */
export interface LocalTask {
  id: string;
  skillId: string;
  input: unknown;
  output?: unknown;
  status: TaskStatus;
  error?: string;
  createdAt: number;
  completedAt?: number;
  retryCount: number;
  maxRetries: number;
  syncPending: boolean;
}

/**
 * 本地调度器
 */
export class LocalScheduler {
  private workerCount: number;
  private offlineLevel: OfflineLevel;
  private skills: Map<string, { handler: Function; offlineCapable: boolean }> = new Map();
  private tasks: Map<string, LocalTask> = new Map();
  private taskQueue: LocalTask[] = [];
  private running = false;
  private pendingCount = 0;
  private completedCount = 0;

  constructor(workerCount: number = 4, offlineLevel: OfflineLevel = OfflineLevel.L1) {
    this.workerCount = workerCount;
    this.offlineLevel = offlineLevel;
  }

  /**
   * 启动调度器
   */
  start(): void {
    if (this.running) return;
    this.running = true;
    console.log(`Local scheduler started with ${this.workerCount} workers, level: ${this.offlineLevel}`);

    // 启动工作任务
    for (let i = 0; i < this.workerCount; i++) {
      this.workerLoop(i);
    }
  }

  /**
   * 停止调度器
   */
  stop(): void {
    this.running = false;
    console.log('Local scheduler stopped');
  }

  private async workerLoop(workerId: number): Promise<void> {
    while (this.running) {
      const task = this.taskQueue.shift();
      if (task) {
        await this.executeTask(task);
      } else {
        await new Promise(r => setTimeout(r, 100));
      }
    }
  }

  private async executeTask(task: LocalTask): Promise<void> {
    task.status = TaskStatus.RUNNING;

    try {
      const skillInfo = this.skills.get(task.skillId);
      if (!skillInfo) {
        throw new Error(`Skill not found: ${task.skillId}`);
      }

      if (!skillInfo.offlineCapable) {
        throw new Error('Skill does not support offline execution');
      }

      const output = await skillInfo.handler(task.input);

      task.output = output;
      task.status = TaskStatus.COMPLETED;
      task.completedAt = Date.now();
      task.syncPending = this.offlineLevel !== OfflineLevel.L4;

      this.completedCount++;
      this.pendingCount--;

      console.log(`Task ${task.id} completed: ${task.skillId}`);

    } catch (error) {
      task.status = TaskStatus.FAILED;
      task.error = String(error);

      if (task.retryCount < task.maxRetries) {
        task.retryCount++;
        task.status = TaskStatus.PENDING;
        this.taskQueue.push(task);
        console.warn(`Task ${task.id} retry ${task.retryCount}`);
      } else {
        this.pendingCount--;
        console.error(`Task ${task.id} failed: ${error}`);
      }
    }
  }

  /**
   * 注册技能
   */
  registerSkill(skillId: string, handler: Function, offlineCapable: boolean = true): void {
    this.skills.set(skillId, { handler, offlineCapable });
    console.log(`Registered local skill: ${skillId} (offline: ${offlineCapable})`);
  }

  /**
   * 注销技能
   */
  unregisterSkill(skillId: string): void {
    this.skills.delete(skillId);
  }

  /**
   * 提交任务
   */
  submitTask(skillId: string, input: unknown): string {
    const task: LocalTask = {
      id: `local-${uuidv4().split('-')[0]}`,
      skillId,
      input,
      status: TaskStatus.PENDING,
      createdAt: Date.now(),
      retryCount: 0,
      maxRetries: 3,
      syncPending: true,
    };

    this.tasks.set(task.id, task);
    this.taskQueue.push(task);
    this.pendingCount++;

    console.log(`Task submitted: ${task.id} -> ${skillId}`);
    return task.id;
  }

  /**
   * 获取任务
   */
  getTask(taskId: string): LocalTask | undefined {
    return this.tasks.get(taskId);
  }

  /**
   * 取消任务
   */
  cancelTask(taskId: string): boolean {
    const task = this.tasks.get(taskId);
    if (task && task.status === TaskStatus.PENDING) {
      task.status = TaskStatus.CANCELLED;
      this.pendingCount--;
      return true;
    }
    return false;
  }

  /**
   * 列出待处理任务
   */
  listPendingTasks(): string[] {
    const result: string[] = [];
    this.tasks.forEach((task, id) => {
      if (task.status === TaskStatus.PENDING) {
        result.push(id);
      }
    });
    return result;
  }

  /**
   * 列出已注册技能
   */
  listSkills(): string[] {
    return Array.from(this.skills.keys());
  }

  getPendingCount(): number { return this.pendingCount; }
  getCompletedCount(): number { return this.completedCount; }
  getOfflineLevel(): OfflineLevel { return this.offlineLevel; }
}

/**
 * 离线缓存
 */
export class OfflineCache {
  private cache: Map<string, { data: unknown; timestamp: number; expiry?: number; synced: boolean }> = new Map();
  private pendingSync: Set<string> = new Set();
  private hits = 0;
  private misses = 0;
  private maxSize: number;
  private currentSize = 0;

  constructor(maxSize: number = 10 * 1024 * 1024) {
    this.maxSize = maxSize;
  }

  /**
   * 存储数据
   */
  put(key: string, data: unknown, expiryMs: number = 0): void {
    const timestamp = Date.now();
    const expiry = expiryMs > 0 ? timestamp + expiryMs : undefined;

    // 检查容量 (简化)
    const size = this.estimateSize(data);
    if (this.currentSize + size > this.maxSize) {
      this.evictIfNeeded(size);
    }

    this.cache.set(key, { data, timestamp, expiry, synced: false });
    this.currentSize += size;
    this.pendingSync.add(key);
  }

  /**
   * 获取数据
   */
  get(key: string): unknown | undefined {
    const entry = this.cache.get(key);

    if (!entry) {
      this.misses++;
      return undefined;
    }

    // 检查过期
    if (entry.expiry && Date.now() > entry.expiry) {
      this.remove(key);
      this.misses++;
      return undefined;
    }

    this.hits++;
    return entry.data;
  }

  /**
   * 删除数据
   */
  remove(key: string): void {
    const entry = this.cache.get(key);
    if (entry) {
      this.currentSize -= this.estimateSize(entry.data);
      this.cache.delete(key);
      this.pendingSync.delete(key);
    }
  }

  /**
   * 清空缓存
   */
  clear(): void {
    this.cache.clear();
    this.pendingSync.clear();
    this.currentSize = 0;
  }

  /**
   * 获取待同步键列表
   */
  getPendingKeys(): string[] {
    return Array.from(this.pendingSync);
  }

  /**
   * 标记已同步
   */
  markSynced(key: string): void {
    const entry = this.cache.get(key);
    if (entry) {
      entry.synced = true;
    }
    this.pendingSync.delete(key);
  }

  /**
   * 获取命中率
   */
  hitRate(): number {
    const total = this.hits + this.misses;
    return total > 0 ? this.hits / total : 0;
  }

  /**
   * 获取待同步数量
   */
  getPendingCount(): number {
    return this.pendingSync.size;
  }

  private estimateSize(data: unknown): number {
    try {
      return JSON.stringify(data).length * 2;
    } catch {
      return 1024; // 默认 1KB
    }
  }

  private evictIfNeeded(needed: number): void {
    const now = Date.now();

    // 清理过期项
    this.cache.forEach((entry, key) => {
      if (entry.expiry && now > entry.expiry) {
        this.remove(key);
      }
    });

    // 清理最旧的已同步项
    while (this.currentSize + needed > this.maxSize && this.cache.size > 0) {
      let oldestKey: string | null = null;
      let oldestTime = Infinity;

      this.cache.forEach((entry, key) => {
        if (entry.synced && entry.timestamp < oldestTime) {
          oldestTime = entry.timestamp;
          oldestKey = key;
        }
      });

      if (oldestKey) {
        this.remove(oldestKey);
      } else {
        break;
      }
    }
  }
}

/**
 * 离线管理器
 */
export class OfflineManager {
  private level: OfflineLevel;
  public scheduler: LocalScheduler;
  public cache: OfflineCache;
  private offlineMode: boolean;
  private syncCallback?: (key: string, data: unknown) => Promise<void>;

  constructor(level: OfflineLevel = OfflineLevel.L1) {
    this.level = level;
    this.scheduler = new LocalScheduler(4, level);
    this.cache = new OfflineCache();
    this.offlineMode = level === OfflineLevel.L1;
  }

  /**
   * 启动离线管理器
   */
  start(): void {
    this.scheduler.start();
    console.log(`Offline manager started at level ${this.level}`);
  }

  /**
   * 停止离线管理器
   */
  stop(): void {
    this.scheduler.stop();
    this.cache.clear();
    console.log('Offline manager stopped');
  }

  /**
   * 设置离线模式
   */
  setOfflineMode(offline: boolean): void {
    this.offlineMode = offline;
    console.log(`Offline mode: ${offline}`);
  }

  /**
   * 是否处于离线模式
   */
  isOffline(): boolean {
    return this.offlineMode;
  }

  /**
   * 获取离线等级
   */
  getLevel(): OfflineLevel {
    return this.level;
  }

  /**
   * 注册技能
   */
  registerSkill(skillId: string, handler: Function, offlineCapable: boolean = true): void {
    this.scheduler.registerSkill(skillId, handler, offlineCapable);
  }

  /**
   * 本地执行任务
   */
  executeLocal(skillId: string, input: unknown): string {
    return this.scheduler.submitTask(skillId, input);
  }

  /**
   * 同步执行
   */
  async executeSync(skillId: string, input: unknown, timeout: number = 30000): Promise<unknown> {
    const taskId = this.scheduler.submitTask(skillId, input);

    const startTime = Date.now();
    while (true) {
      const task = this.scheduler.getTask(taskId);
      if (task) {
        switch (task.status) {
          case TaskStatus.COMPLETED:
            return task.output;
          case TaskStatus.FAILED:
            throw new Error(task.error || 'Task failed');
          case TaskStatus.CANCELLED:
            throw new Error('Task cancelled');
        }
      }

      if (Date.now() - startTime > timeout) {
        throw new Error('Timeout');
      }

      await new Promise(r => setTimeout(r, 100));
    }
  }

  /**
   * 缓存数据
   */
  cacheData(key: string, data: unknown, expiryMs: number = 0): void {
    this.cache.put(key, data, expiryMs);
  }

  /**
   * 获取缓存数据
   */
  getCachedData(key: string): unknown | undefined {
    return this.cache.get(key);
  }

  /**
   * 获取待同步键列表
   */
  getPendingSyncKeys(): string[] {
    return this.cache.getPendingKeys();
  }

  /**
   * 立即同步
   */
  async syncNow(): Promise<boolean> {
    const pending = this.cache.getPendingKeys();
    if (pending.length === 0) return true;

    if (!this.syncCallback) {
      console.warn('No sync callback configured');
      return false;
    }

    for (const key of pending) {
      const data = this.cache.get(key);
      if (data !== undefined) {
        try {
          await this.syncCallback(key, data);
          this.cache.markSynced(key);
        } catch (error) {
          console.error(`Sync failed for ${key}:`, error);
          return false;
        }
      }
    }

    return true;
  }

  /**
   * 设置同步回调
   */
  setSyncCallback(callback: (key: string, data: unknown) => Promise<void>): void {
    this.syncCallback = callback;
  }

  /**
   * 获取任务
   */
  getTask(taskId: string): LocalTask | undefined {
    return this.scheduler.getTask(taskId);
  }

  /**
   * 获取统计信息
   */
  getStats(): {
    offlineMode: boolean;
    level: number;
    pendingTasks: number;
    completedTasks: number;
    pendingSync: number;
    cacheHitRate: number;
  } {
    return {
      offlineMode: this.offlineMode,
      level: this.level,
      pendingTasks: this.scheduler.getPendingCount(),
      completedTasks: this.scheduler.getCompletedCount(),
      pendingSync: this.cache.getPendingCount(),
      cacheHitRate: this.cache.hitRate(),
    };
  }
}