//! Skills Module
//! Sprint 29: Rust Agent SDK

use std::collections::HashMap;
use std::sync::Arc;
use async_trait::async_trait;
use serde::{Deserialize, Serialize};
use crate::error::{Error, Result};
use crate::message::Task;

/// 技能信息
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SkillInfo {
    pub id: String,
    pub name: String,
    pub description: String,
    pub operations: Vec<String>,
    pub version: String,
    pub author: String,
    pub tags: Vec<String>,
}

/// 技能处理器trait
#[async_trait]
pub trait SkillHandler: Send + Sync {
    /// 执行技能
    async fn execute(&self, operation: &str, input: serde_json::Value) -> Result<serde_json::Value>;

    /// 获取技能信息
    fn info(&self) -> SkillInfo;
}

/// 技能执行器
pub struct SkillExecutor {
    skills: Arc<tokio::sync::RwLock<HashMap<String, Arc<dyn SkillHandler>>>>,
    stats: Arc<tokio::sync::RwLock<HashMap<String, SkillStats>>>,
}

impl SkillExecutor {
    pub fn new() -> Self {
        Self {
            skills: Arc::new(tokio::sync::RwLock::new(HashMap::new())),
            stats: Arc::new(tokio::sync::RwLock::new(HashMap::new())),
        }
    }

    /// 注册技能
    pub async fn register(&self, skill_id: &str, handler: Arc<dyn SkillHandler>) {
        self.skills.write().await.insert(skill_id.to_string(), handler);
        self.stats.write().await.insert(skill_id.to_string(), SkillStats::default());
    }

    /// 注销技能
    pub async fn unregister(&self, skill_id: &str) {
        self.skills.write().await.remove(skill_id);
        self.stats.write().await.remove(skill_id);
    }

    /// 执行技能
    pub async fn execute(&self, task: &Task) -> Result<super::message::TaskResult> {
        let skills = self.skills.read().await;
        let handler = skills.get(&task.skill_id)
            .ok_or_else(|| Error::SkillNotFound(task.skill_id.clone()))?;

        let mut stats = self.stats.write().await;
        let skill_stats = stats.entry(task.skill_id.clone()).or_default();
        skill_stats.invocations += 1;

        drop(stats);
        drop(skills);

        let start = std::time::Instant::now();

        match handler.execute(&task.operation, task.input.clone()).await {
            Ok(result) => {
                let mut stats = self.stats.write().await;
                stats.get_mut(&task.skill_id).map(|s| {
                    s.successes += 1;
                    s.total_duration += start.elapsed();
                });

                Ok(super::message::TaskResult {
                    task_id: task.task_id.clone(),
                    success: true,
                    result: Some(result),
                    error: None,
                    agent_id: String::new(),
                    duration_ms: start.elapsed().as_millis() as u64,
                })
            }
            Err(e) => {
                let mut stats = self.stats.write().await;
                stats.get_mut(&task.skill_id).map(|s| {
                    s.failures += 1;
                });

                Ok(super::message::TaskResult {
                    task_id: task.task_id.clone(),
                    success: false,
                    result: None,
                    error: Some(e.to_string()),
                    agent_id: String::new(),
                    duration_ms: start.elapsed().as_millis() as u64,
                })
            }
        }
    }

    /// 列出技能
    pub async fn list_skills(&self) -> Vec<SkillInfo> {
        self.skills.read().await.values().map(|h| h.info()).collect()
    }

    /// 获取统计
    pub async fn get_stats(&self, skill_id: &str) -> Option<SkillStats> {
        self.stats.read().await.get(skill_id).cloned()
    }
}

/// 技能统计
#[derive(Debug, Clone, Default)]
pub struct SkillStats {
    pub invocations: u64,
    pub successes: u64,
    pub failures: u64,
    pub total_duration: std::time::Duration,
}

/// 技能注册表
pub struct SkillRegistry {
    skills: Arc<tokio::sync::RwLock<HashMap<String, SkillInfo>>>,
    categories: Arc<tokio::sync::RwLock<HashMap<String, Vec<String>>>>,
}

impl SkillRegistry {
    pub fn new() -> Self {
        Self {
            skills: Arc::new(tokio::sync::RwLock::new(HashMap::new())),
            categories: Arc::new(tokio::sync::RwLock::new(HashMap::new())),
        }
    }

    pub async fn register(&self, info: SkillInfo) {
        let id = info.id.clone();
        let tags = info.tags.clone();

        self.skills.write().await.insert(id.clone(), info);

        for tag in tags {
            self.categories.write().await
                .entry(tag)
                .or_default()
                .push(id.clone());
        }
    }

    pub async fn get(&self, skill_id: &str) -> Option<SkillInfo> {
        self.skills.read().await.get(skill_id).cloned()
    }

    pub async fn search(&self, query: &str) -> Vec<SkillInfo> {
        let query = query.to_lowercase();
        self.skills.read().await.values()
            .filter(|info| {
                info.name.to_lowercase().contains(&query) ||
                info.description.to_lowercase().contains(&query) ||
                info.id.to_lowercase().contains(&query)
            })
            .cloned()
            .collect()
    }

    pub async fn list_all(&self) -> Vec<SkillInfo> {
        self.skills.read().await.values().cloned().collect()
    }
}

// 内置技能实现

/// Echo技能
pub struct EchoSkill;

#[async_trait]
impl SkillHandler for EchoSkill {
    async fn execute(&self, _operation: &str, input: serde_json::Value) -> Result<serde_json::Value> {
        Ok(input)
    }

    fn info(&self) -> SkillInfo {
        SkillInfo {
            id: "echo".to_string(),
            name: "Echo".to_string(),
            description: "Echo input".to_string(),
            operations: vec!["echo".to_string()],
            version: "1.0.0".to_string(),
            author: "OFA".to_string(),
            tags: vec!["utility".to_string()],
        }
    }
}

/// 文本处理技能
pub struct TextProcessSkill;

#[async_trait]
impl SkillHandler for TextProcessSkill {
    async fn execute(&self, operation: &str, input: serde_json::Value) -> Result<serde_json::Value> {
        let text = input.get("text")
            .and_then(|v| v.as_str())
            .unwrap_or("");

        let result = match operation {
            "uppercase" => text.to_uppercase(),
            "lowercase" => text.to_lowercase(),
            "reverse" => text.chars().rev().collect(),
            "length" => return Ok(serde_json::json!(text.len())),
            _ => text.to_string(),
        };

        Ok(serde_json::json!(result))
    }

    fn info(&self) -> SkillInfo {
        SkillInfo {
            id: "text.process".to_string(),
            name: "Text Process".to_string(),
            description: "Text processing operations".to_string(),
            operations: vec!["uppercase".to_string(), "lowercase".to_string(), "reverse".to_string(), "length".to_string()],
            version: "1.0.0".to_string(),
            author: "OFA".to_string(),
            tags: vec!["text".to_string(), "utility".to_string()],
        }
    }
}