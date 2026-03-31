/**
 * Skills Module
 * Sprint 29: Node.js Agent SDK
 */

/**
 * 技能信息
 */
export interface SkillInfo {
  id: string;
  name: string;
  description: string;
  operations: string[];
  version: string;
  author: string;
  tags: string[];
}

/**
 * 技能处理器
 */
export type SkillHandler = (operation: string, input: unknown) => Promise<unknown>;

/**
 * 技能执行器
 */
export class SkillExecutor {
  private skills: Map<string, SkillHandler> = new Map();
  private stats: Map<string, { invocations: number; successes: number; failures: number }> = new Map();

  /**
   * 注册技能
   */
  register(skillId: string, handler: SkillHandler): void {
    this.skills.set(skillId, handler);
    this.stats.set(skillId, { invocations: 0, successes: 0, failures: 0 });
    console.log(`Skill registered: ${skillId}`);
  }

  /**
   * 注销技能
   */
  unregister(skillId: string): void {
    this.skills.delete(skillId);
    this.stats.delete(skillId);
  }

  /**
   * 执行技能
   */
  async execute(skillId: string, operation: string, input: unknown): Promise<unknown> {
    const handler = this.skills.get(skillId);
    if (!handler) {
      throw new Error(`Skill not found: ${skillId}`);
    }

    const stat = this.stats.get(skillId)!;
    stat.invocations++;

    try {
      const result = await handler(operation, input);
      stat.successes++;
      return result;
    } catch (error) {
      stat.failures++;
      throw error;
    }
  }

  /**
   * 列出技能
   */
  listSkills(): string[] {
    return Array.from(this.skills.keys());
  }

  /**
   * 获取统计
   */
  getStats(skillId: string): { invocations: number; successes: number; failures: number } | undefined {
    return this.stats.get(skillId);
  }
}

/**
 * 技能注册表
 */
export class SkillRegistry {
  private skills: Map<string, SkillInfo> = new Map();
  private categories: Map<string, Set<string>> = new Map();

  /**
   * 注册技能
   */
  register(info: SkillInfo): void {
    this.skills.set(info.id, info);

    for (const tag of info.tags) {
      if (!this.categories.has(tag)) {
        this.categories.set(tag, new Set());
      }
      this.categories.get(tag)!.add(info.id);
    }
  }

  /**
   * 获取技能
   */
  get(skillId: string): SkillInfo | undefined {
    return this.skills.get(skillId);
  }

  /**
   * 搜索技能
   */
  search(query: string): SkillInfo[] {
    const q = query.toLowerCase();
    return Array.from(this.skills.values()).filter(
      info =>
        info.name.toLowerCase().includes(q) ||
        info.description.toLowerCase().includes(q) ||
        info.id.toLowerCase().includes(q)
    );
  }

  /**
   * 按分类获取
   */
  byCategory(category: string): SkillInfo[] {
    const ids = this.categories.get(category);
    if (!ids) return [];

    return Array.from(ids)
      .map(id => this.skills.get(id))
      .filter((info): info is SkillInfo => info !== undefined);
  }

  /**
   * 列出所有
   */
  listAll(): SkillInfo[] {
    return Array.from(this.skills.values());
  }
}