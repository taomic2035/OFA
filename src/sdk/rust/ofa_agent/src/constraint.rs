//! OFA Rust SDK - Constraint Module
//! Agent 交互约束检查

use std::collections::{HashMap, HashSet};
use std::sync::Arc;
use tokio::sync::RwLock;
use regex::Regex;

/// 约束类型
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum ConstraintType {
    None = 0,
    Privacy = 1,
    Financial = 2,
    Security = 4,
    AuthRequired = 8,
    Location = 16,
    Personal = 32,
    Device = 64,
}

/// 约束检查结果
#[derive(Debug, Clone)]
pub struct ConstraintResult {
    pub allowed: bool,
    pub violated: ConstraintType,
    pub reason: Option<String>,
    pub requires_auth: bool,
    pub suggestions: Vec<String>,
}

impl Default for ConstraintResult {
    fn default() -> Self {
        Self {
            allowed: true,
            violated: ConstraintType::None,
            reason: None,
            requires_auth: false,
            suggestions: Vec::new(),
        }
    }
}

/// 约束规则
#[derive(Debug, Clone)]
pub struct ConstraintRule {
    pub name: String,
    pub constraint_type: ConstraintType,
    pub action_pattern: Option<Regex>,
    pub data_pattern: Option<Regex>,
    pub offline_restricted: bool,
    pub requires_auth: bool,
    pub message: String,
}

/// 约束检查器
pub struct ConstraintChecker {
    rules: Arc<RwLock<Vec<ConstraintRule>>>,
    offline_restricted_actions: Arc<RwLock<HashSet<String>>>,
    sensitive_fields: Arc<RwLock<HashSet<String>>>,
    offline_mode: Arc<std::sync::atomic::AtomicBool>,
}

impl ConstraintChecker {
    pub fn new() -> Self {
        let checker = Self {
            rules: Arc::new(RwLock::new(Vec::new())),
            offline_restricted_actions: Arc::new(RwLock::new(HashSet::new())),
            sensitive_fields: Arc::new(RwLock::new(HashSet::new())),
            offline_mode: Arc::new(std::sync::atomic::AtomicBool::new(false)),
        };

        // 加载默认规则
        checker.load_default_rules();
        checker.load_sensitive_fields();

        checker
    }

    fn load_default_rules(&self) {
        // 财务操作
        self.add_rule(ConstraintRule {
            name: "financial_operations".to_string(),
            constraint_type: ConstraintType::Financial,
            action_pattern: Regex::new(r"(?i)(payment|transfer|withdraw|pay)").ok(),
            data_pattern: None,
            offline_restricted: true,
            requires_auth: true,
            message: "Financial operations require online mode and authorization".to_string(),
        });

        // 隐私数据
        self.add_rule(ConstraintRule {
            name: "privacy_data".to_string(),
            constraint_type: ConstraintType::Privacy,
            action_pattern: None,
            data_pattern: Regex::new(r"(?i)(idcard|id_card|身份证|passport|护照)").ok(),
            offline_restricted: false,
            requires_auth: false,
            message: "Data contains sensitive personal information".to_string(),
        });

        // 位置信息
        self.add_rule(ConstraintRule {
            name: "location_data".to_string(),
            constraint_type: ConstraintType::Location,
            action_pattern: None,
            data_pattern: Regex::new(r"(?i)(location|gps|latitude|longitude|经纬度)").ok(),
            offline_restricted: false,
            requires_auth: true,
            message: "Location data sharing requires authorization".to_string(),
        });

        // 安全操作
        self.add_rule(ConstraintRule {
            name: "security_operations".to_string(),
            constraint_type: ConstraintType::Security,
            action_pattern: Regex::new(r"(?i)(delete|password|auth|login|logout)").ok(),
            data_pattern: None,
            offline_restricted: true,
            requires_auth: true,
            message: "Security operations require online mode and authorization".to_string(),
        });
    }

    fn load_sensitive_fields(&self) {
        let fields = [
            "idcard", "id_card", "身份证", "passport", "护照",
            "phone", "mobile", "电话", "手机",
            "email", "邮箱",
            "address", "地址",
            "bank_account", "银行卡",
            "password", "密码",
            "token", "令牌",
            "secret", "密钥",
            "location", "gps", "位置",
        ];

        let mut sensitive = self.sensitive_fields.write().blocking_read();
        for field in fields {
            sensitive.insert(field.to_lowercase());
        }
    }

    /// 添加规则
    pub fn add_rule(&self, rule: ConstraintRule) {
        if rule.offline_restricted {
            if let Some(pattern) = &rule.action_pattern {
                let source = pattern.as_str();
                let clean = source.replace("(?i)", "").replace('(', "").replace(')', "");
                for action in clean.split('|') {
                    self.offline_restricted_actions
                        .write()
                        .blocking_read()
                        .insert(action.trim().to_lowercase());
                }
            }
        }

        self.rules.write().blocking_read().push(rule);
    }

    /// 移除规则
    pub async fn remove_rule(&self, name: &str) {
        let mut rules = self.rules.write().await;
        rules.retain(|r| r.name != name);
    }

    /// 设置离线模式
    pub fn set_offline_mode(&self, offline: bool) {
        self.offline_mode.store(offline, std::sync::atomic::Ordering::SeqCst);
    }

    /// 检查约束
    pub async fn check(&self, action: &str, data: Option<&str>) -> ConstraintResult {
        let mut result = ConstraintResult::default();

        // 1. 检查离线受限操作
        if self.offline_mode.load(std::sync::atomic::Ordering::SeqCst) {
            let action_lower = action.to_lowercase();
            let restricted = self.offline_restricted_actions.read().await;

            for r in restricted.iter() {
                if action_lower.contains(r) {
                    result.allowed = false;
                    result.violated = ConstraintType::Financial;
                    result.reason = Some(format!("Action '{}' requires online mode", action));
                    result.suggestions.push("Connect to network or use alternative offline action".to_string());
                    return result;
                }
            }
        }

        // 2. 应用规则
        let rules = self.rules.read().await;
        for rule in rules.iter() {
            let rule_result = self.apply_rule(rule, action, data).await;
            if !rule_result.allowed {
                return rule_result;
            }
        }

        // 3. 检查敏感数据
        if let Some(data_str) = data {
            let data_result = self.check_sensitive_data(data_str).await;
            if !data_result.allowed {
                return data_result;
            }
        }

        result
    }

    async fn apply_rule(&self, rule: &ConstraintRule, action: &str, data: Option<&str>) -> ConstraintResult {
        let mut result = ConstraintResult::default();

        // 检查操作模式
        if let Some(pattern) = &rule.action_pattern {
            if !pattern.is_match(action) {
                return result;
            }
        }

        // 检查数据模式
        if let (Some(pattern), Some(data_str)) = (&rule.data_pattern, data) {
            if !pattern.is_match(data_str) {
                return result;
            }
        }

        // 离线限制
        if rule.offline_restricted && self.offline_mode.load(std::sync::atomic::Ordering::SeqCst) {
            return ConstraintResult {
                allowed: false,
                violated: rule.constraint_type,
                reason: Some(rule.message.clone()),
                requires_auth: rule.requires_auth,
                suggestions: vec!["Switch to online mode".to_string()],
            };
        }

        // 授权要求
        if rule.requires_auth {
            // TODO: 检查用户授权状态
            return ConstraintResult {
                allowed: false,
                violated: ConstraintType::AuthRequired,
                reason: Some(rule.message.clone()),
                requires_auth: true,
                suggestions: vec!["Request authorization from user".to_string()],
            };
        }

        result
    }

    async fn check_sensitive_data(&self, data: &str) -> ConstraintResult {
        let data_lower = data.to_lowercase();
        let sensitive = self.sensitive_fields.read().await;

        for field in sensitive.iter() {
            if data_lower.contains(field) {
                let (constraint_type, reason) = if field.contains("bank") || field.contains("card") {
                    (ConstraintType::Financial, "Data contains financial information")
                } else if field.contains("location") || field.contains("gps") {
                    (ConstraintType::Location, "Data contains location information")
                } else if field.contains("password") || field.contains("token") || field.contains("secret") {
                    (ConstraintType::Security, "Data contains security credentials")
                } else {
                    (ConstraintType::Privacy, "Data contains sensitive personal information")
                };

                return ConstraintResult {
                    allowed: false,
                    violated: constraint_type,
                    reason: Some(reason.to_string()),
                    requires_auth: false,
                    suggestions: Vec::new(),
                };
            }
        }

        ConstraintResult::default()
    }

    /// 添加敏感字段
    pub async fn add_sensitive_field(&self, field: &str) {
        self.sensitive_fields.write().await.insert(field.to_lowercase());
    }

    /// 移除敏感字段
    pub async fn remove_sensitive_field(&self, field: &str) {
        self.sensitive_fields.write().await.remove(&field.to_lowercase());
    }

    /// 是否允许
    pub async fn is_allowed(&self, action: &str, data: Option<&str>) -> bool {
        self.check(action, data).await.allowed
    }
}

impl Default for ConstraintChecker {
    fn default() -> Self {
        Self::new()
    }
}