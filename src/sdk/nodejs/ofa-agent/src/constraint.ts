/**
 * OFA Node.js SDK - 约束模块
 * Agent 交互约束检查
 */

/**
 * 约束类型
 */
export enum ConstraintType {
  NONE = 0,
  PRIVACY = 1,
  FINANCIAL = 2,
  SECURITY = 4,
  AUTH_REQUIRED = 8,
  LOCATION = 16,
  PERSONAL = 32,
  DEVICE = 64,
}

/**
 * 约束检查结果
 */
export interface ConstraintResult {
  allowed: boolean;
  violated: ConstraintType;
  reason?: string;
  requiresAuth: boolean;
  suggestions: string[];
}

/**
 * 约束规则
 */
export interface ConstraintRule {
  name: string;
  type: ConstraintType;
  actionPattern?: RegExp;
  dataPattern?: RegExp;
  offlineRestricted: boolean;
  requiresAuth: boolean;
  message: string;
}

/**
 * 约束检查器
 */
export class ConstraintChecker {
  private rules: ConstraintRule[] = [];
  private offlineRestrictedActions: Set<string> = new Set();
  private sensitiveFields: Set<string> = new Set();
  private offlineMode = false;

  constructor() {
    this.loadDefaultRules();
    this.loadSensitiveFields();
  }

  private loadDefaultRules(): void {
    // 财务操作
    this.addRule({
      name: 'financial_operations',
      type: ConstraintType.FINANCIAL,
      actionPattern: /(payment|transfer|withdraw|pay)/i,
      offlineRestricted: true,
      requiresAuth: true,
      message: 'Financial operations require online mode and authorization',
    });

    // 隐私数据
    this.addRule({
      name: 'privacy_data',
      type: ConstraintType.PRIVACY,
      dataPattern: /(idcard|id_card|身份证|passport|护照)/i,
      offlineRestricted: false,
      requiresAuth: false,
      message: 'Data contains sensitive personal information',
    });

    // 位置信息
    this.addRule({
      name: 'location_data',
      type: ConstraintType.LOCATION,
      dataPattern: /(location|gps|latitude|longitude|经纬度)/i,
      offlineRestricted: false,
      requiresAuth: true,
      message: 'Location data sharing requires authorization',
    });

    // 安全操作
    this.addRule({
      name: 'security_operations',
      type: ConstraintType.SECURITY,
      actionPattern: /(delete|password|auth|login|logout)/i,
      offlineRestricted: true,
      requiresAuth: true,
      message: 'Security operations require online mode and authorization',
    });
  }

  private loadSensitiveFields(): void {
    const fields = [
      'idcard', 'id_card', '身份证', 'passport', '护照',
      'phone', 'mobile', '电话', '手机',
      'email', '邮箱',
      'address', '地址',
      'bank_account', '银行卡',
      'password', '密码',
      'token', '令牌',
      'secret', '密钥',
      'location', 'gps', '位置',
    ];
    fields.forEach(f => this.sensitiveFields.add(f.toLowerCase()));
  }

  /**
   * 添加规则
   */
  addRule(rule: ConstraintRule): void {
    this.rules.push(rule);

    if (rule.offlineRestricted && rule.actionPattern) {
      // 提取动作关键词
      const pattern = rule.actionPattern.source.replace(/[()]/g, '');
      pattern.split('|').forEach(action => {
        this.offlineRestrictedActions.add(action.trim().toLowerCase());
      });
    }
  }

  /**
   * 移除规则
   */
  removeRule(name: string): void {
    this.rules = this.rules.filter(r => r.name !== name);
  }

  /**
   * 设置离线模式
   */
  setOfflineMode(offline: boolean): void {
    this.offlineMode = offline;
  }

  /**
   * 检查约束
   */
  check(action: string, data?: string): ConstraintResult {
    const result: ConstraintResult = {
      allowed: true,
      violated: ConstraintType.NONE,
      requiresAuth: false,
      suggestions: [],
    };

    // 1. 检查离线受限操作
    if (this.offlineMode) {
      const actionLower = action.toLowerCase();
      for (const restricted of this.offlineRestrictedActions) {
        if (actionLower.includes(restricted)) {
          result.allowed = false;
          result.violated = ConstraintType.FINANCIAL | ConstraintType.SECURITY;
          result.reason = `Action '${action}' requires online mode`;
          result.suggestions.push('Connect to network or use alternative offline action');
          return result;
        }
      }
    }

    // 2. 应用规则
    for (const rule of this.rules) {
      const ruleResult = this.applyRule(rule, action, data);
      if (!ruleResult.allowed) {
        return ruleResult;
      }
    }

    // 3. 检查敏感数据
    if (data) {
      const dataResult = this.checkSensitiveData(data);
      if (!dataResult.allowed) {
        return dataResult;
      }
    }

    return result;
  }

  private applyRule(rule: ConstraintRule, action: string, data?: string): ConstraintResult {
    const result: ConstraintResult = {
      allowed: true,
      violated: ConstraintType.NONE,
      requiresAuth: false,
      suggestions: [],
    };

    // 检查操作模式
    if (rule.actionPattern && !rule.actionPattern.test(action)) {
      return result;
    }

    // 检查数据模式
    if (rule.dataPattern && data && !rule.dataPattern.test(data)) {
      return result;
    }

    // 离线限制
    if (rule.offlineRestricted && this.offlineMode) {
      result.allowed = false;
      result.violated = rule.type;
      result.reason = rule.message;
      result.suggestions.push('Switch to online mode');
      return result;
    }

    // 授权要求
    if (rule.requiresAuth) {
      // TODO: 检查用户授权状态
      result.allowed = false;
      result.violated = ConstraintType.AUTH_REQUIRED;
      result.reason = rule.message;
      result.requiresAuth = true;
      result.suggestions.push('Request authorization from user');
    }

    return result;
  }

  private checkSensitiveData(data: string): ConstraintResult {
    const result: ConstraintResult = {
      allowed: true,
      violated: ConstraintType.NONE,
      requiresAuth: false,
      suggestions: [],
    };

    const dataLower = data.toLowerCase();

    for (const field of this.sensitiveFields) {
      if (dataLower.includes(field)) {
        result.allowed = false;

        if (field.includes('bank') || field.includes('card')) {
          result.violated = ConstraintType.FINANCIAL;
          result.reason = 'Data contains financial information';
        } else if (field.includes('location') || field.includes('gps')) {
          result.violated = ConstraintType.LOCATION;
          result.reason = 'Data contains location information';
        } else if (field.includes('password') || field.includes('token') || field.includes('secret')) {
          result.violated = ConstraintType.SECURITY;
          result.reason = 'Data contains security credentials';
        } else {
          result.violated = ConstraintType.PRIVACY;
          result.reason = 'Data contains sensitive personal information';
        }

        return result;
      }
    }

    return result;
  }

  /**
   * 添加敏感字段
   */
  addSensitiveField(field: string): void {
    this.sensitiveFields.add(field.toLowerCase());
  }

  /**
   * 移除敏感字段
   */
  removeSensitiveField(field: string): void {
    this.sensitiveFields.delete(field.toLowerCase());
  }

  /**
   * 获取离线受限操作
   */
  getOfflineRestrictedActions(): string[] {
    return Array.from(this.offlineRestrictedActions);
  }

  /**
   * 获取规则列表
   */
  getRules(): ConstraintRule[] {
    return [...this.rules];
  }

  /**
   * 是否允许
   */
  isAllowed(action: string, data?: string): boolean {
    return this.check(action, data).allowed;
  }
}

/**
 * 约束客户端
 */
export class ConstraintClient {
  private checker: ConstraintChecker;

  constructor(checker?: ConstraintChecker) {
    this.checker = checker || new ConstraintChecker();
  }

  /**
   * 设置离线模式
   */
  setOfflineMode(offline: boolean): void {
    this.checker.setOfflineMode(offline);
  }

  /**
   * 检查约束
   */
  check(action: string, data?: string): ConstraintResult {
    return this.checker.check(action, data);
  }

  /**
   * 是否允许
   */
  isAllowed(action: string, data?: string): boolean {
    return this.checker.isAllowed(action, data);
  }

  /**
   * 添加规则
   */
  addRule(rule: ConstraintRule): void {
    this.checker.addRule(rule);
  }

  /**
   * 获取离线受限操作
   */
  getOfflineRestrictedActions(): string[] {
    return this.checker.getOfflineRestrictedActions();
  }
}