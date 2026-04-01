//! OFA Rust SDK - Offline Module
//! 支持离线能力等级 L1-L4

use std::collections::{HashMap, HashSet, VecDeque};
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant, SystemTime, UNIX_EPOCH};
use tokio::sync::RwLock;
use uuid::Uuid;

/// 离线能力等级
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum OfflineLevel {
    /// 不支持离线
    None = 0,
    /// 完全离线 (本地执行)
    L1 = 1,
    /// 局域网协作
    L2 = 2,
    /// 弱网同步
    L3 = 3,
    /// 在线模式
    L4 = 4,
}

impl Default for OfflineLevel {
    fn default() -> Self {
        OfflineLevel::L1
    }
}

/// 任务状态
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum TaskStatus {
    Pending,
    Running,
    Completed,
    Failed,
    Cancelled,
}

/// 离线任务
#[derive(Debug, Clone)]
pub struct LocalTask {
    pub id: String,
    pub skill_id: String,
    pub input: Vec<u8>,
    pub output: Option<Vec<u8>>,
    pub status: TaskStatus,
    pub error: Option<String>,
    pub created_at: u64,
    pub completed_at: Option<u64>,
    pub retry_count: u32,
    pub max_retries: u32,
    pub sync_pending: bool,
}

impl LocalTask {
    pub fn new(skill_id: String, input: Vec<u8>) -> Self {
        Self {
            id: format!("local-{}", &Uuid::new_v4().to_string()[..8]),
            skill_id,
            input,
            output: None,
            status: TaskStatus::Pending,
            error: None,
            created_at: SystemTime::now()
                .duration_since(UNIX_EPOCH)
                .unwrap()
                .as_millis() as u64,
            completed_at: None,
            retry_count: 0,
            max_retries: 3,
            sync_pending: true,
        }
    }

    pub fn can_retry(&self) -> bool {
        self.retry_count < self.max_retries
    }
}

/// 技能处理器类型
pub type SkillHandler = Box<dyn Fn(&[u8]) -> Result<Vec<u8>, String> + Send + Sync>;

/// 技能信息
struct SkillInfo {
    handler: SkillHandler,
    offline_capable: bool,
}

/// 本地调度器
pub struct LocalScheduler {
    worker_count: usize,
    offline_level: OfflineLevel,
    skills: Arc<RwLock<HashMap<String, SkillInfo>>>,
    tasks: Arc<RwLock<HashMap<String, LocalTask>>>,
    task_queue: Arc<Mutex<VecDeque<String>>>,
    running: Arc<std::sync::atomic::AtomicBool>,
    pending_count: Arc<std::sync::atomic::AtomicUsize>,
    completed_count: Arc<std::sync::atomic::AtomicUsize>,
}

impl LocalScheduler {
    pub fn new(worker_count: usize, offline_level: OfflineLevel) -> Self {
        Self {
            worker_count,
            offline_level,
            skills: Arc::new(RwLock::new(HashMap::new())),
            tasks: Arc::new(RwLock::new(HashMap::new())),
            task_queue: Arc::new(Mutex::new(VecDeque::new())),
            running: Arc::new(std::sync::atomic::AtomicBool::new(false)),
            pending_count: Arc::new(std::sync::atomic::AtomicUsize::new(0)),
            completed_count: Arc::new(std::sync::atomic::AtomicUsize::new(0)),
        }
    }

    /// 启动调度器
    pub async fn start(&self) {
        if self.running.load(std::sync::atomic::Ordering::SeqCst) {
            return;
        }
        self.running.store(true, std::sync::atomic::Ordering::SeqCst);
        log::info!(
            "Local scheduler started with {} workers, level: {}",
            self.worker_count,
            self.offline_level as u8
        );

        // 启动工作任务
        for _ in 0..self.worker_count {
            let skills = self.skills.clone();
            let tasks = self.tasks.clone();
            let task_queue = self.task_queue.clone();
            let running = self.running.clone();
            let pending_count = self.pending_count.clone();
            let completed_count = self.completed_count.clone();
            let offline_level = self.offline_level;

            tokio::spawn(async move {
                while running.load(std::sync::atomic::Ordering::SeqCst) {
                    let task_id = {
                        let mut queue = task_queue.lock().unwrap();
                        queue.pop_front()
                    };

                    if let Some(id) = task_id {
                        Self::execute_task(
                            &id,
                            &skills,
                            &tasks,
                            pending_count.clone(),
                            completed_count.clone(),
                            offline_level,
                        ).await;
                    } else {
                        tokio::time::sleep(Duration::from_millis(100)).await;
                    }
                }
            });
        }
    }

    /// 停止调度器
    pub async fn stop(&self) {
        self.running.store(false, std::sync::atomic::Ordering::SeqCst);
        log::info!("Local scheduler stopped");
    }

    async fn execute_task(
        task_id: &str,
        skills: &Arc<RwLock<HashMap<String, SkillInfo>>>,
        tasks: &Arc<RwLock<HashMap<String, LocalTask>>>,
        pending_count: Arc<std::sync::atomic::AtomicUsize>,
        completed_count: Arc<std::sync::atomic::AtomicUsize>,
        offline_level: OfflineLevel,
    ) {
        // 获取任务
        let skill_id = {
            let tasks_read = tasks.read().await;
            let task = tasks_read.get(task_id);
            task.map(|t| t.skill_id.clone())
        };

        let Some(skill_id) = skill_id else { return };

        // 执行
        let result = {
            let skills_read = skills.read().await;
            let skill_info = skills_read.get(&skill_id);

            if let Some(info) = skill_info {
                if !info.offline_capable {
                    Err("Skill does not support offline execution".to_string())
                } else {
                    let tasks_read = tasks.read().await;
                    let task = tasks_read.get(task_id);
                    let input = task.map(|t| t.input.clone()).unwrap_or_default();
                    (info.handler)(&input)
                }
            } else {
                Err(format!("Skill not found: {}", skill_id))
            }
        };

        // 更新任务状态
        let mut tasks_write = tasks.write().await;
        if let Some(task) = tasks_write.get_mut(task_id) {
            match result {
                Ok(output) => {
                    task.output = Some(output);
                    task.status = TaskStatus::Completed;
                    task.completed_at = Some(
                        SystemTime::now()
                            .duration_since(UNIX_EPOCH)
                            .unwrap()
                            .as_millis() as u64,
                    );
                    task.sync_pending = offline_level != OfflineLevel::L4;

                    completed_count.fetch_add(1, std::sync::atomic::Ordering::SeqCst);
                    pending_count.fetch_sub(1, std::sync::atomic::Ordering::SeqCst);

                    log::info!("Task {} completed: {}", task_id, task.skill_id);
                }
                Err(e) => {
                    if task.can_retry() {
                        task.retry_count += 1;
                        task.status = TaskStatus::Pending;
                        log::warn!("Task {} retry {}", task_id, task.retry_count);
                    } else {
                        task.status = TaskStatus::Failed;
                        task.error = Some(e);
                        pending_count.fetch_sub(1, std::sync::atomic::Ordering::SeqCst);
                        log::error!("Task {} failed", task_id);
                    }
                }
            }
        }
    }

    /// 注册技能
    pub async fn register_skill<F>(&self, skill_id: &str, handler: F, offline_capable: bool)
    where
        F: Fn(&[u8]) -> Result<Vec<u8>, String> + Send + Sync + 'static,
    {
        let mut skills = self.skills.write().await;
        skills.insert(
            skill_id.to_string(),
            SkillInfo {
                handler: Box::new(handler),
                offline_capable,
            },
        );
        log::info!("Registered local skill: {} (offline: {})", skill_id, offline_capable);
    }

    /// 提交任务
    pub async fn submit_task(&self, skill_id: &str, input: Vec<u8>) -> String {
        let task = LocalTask::new(skill_id.to_string(), input);
        let task_id = task.id.clone();

        let mut tasks = self.tasks.write().await;
        tasks.insert(task_id.clone(), task);

        let mut queue = self.task_queue.lock().unwrap();
        queue.push_back(task_id.clone());

        self.pending_count.fetch_add(1, std::sync::atomic::Ordering::SeqCst);

        log::info!("Task submitted: {} -> {}", task_id, skill_id);
        task_id
    }

    /// 获取任务
    pub async fn get_task(&self, task_id: &str) -> Option<LocalTask> {
        let tasks = self.tasks.read().await;
        tasks.get(task_id).cloned()
    }

    /// 获取待处理任务数
    pub fn pending_count(&self) -> usize {
        self.pending_count.load(std::sync::atomic::Ordering::SeqCst)
    }

    /// 获取已完成任务数
    pub fn completed_count(&self) -> usize {
        self.completed_count.load(std::sync::atomic::Ordering::SeqCst)
    }
}

/// 离线缓存
pub struct OfflineCache {
    cache: Arc<RwLock<HashMap<String, CacheEntry>>>,
    pending_sync: Arc<RwLock<HashSet<String>>>,
    max_size: usize,
    current_size: Arc<std::sync::atomic::AtomicUsize>,
}

#[derive(Clone)]
struct CacheEntry {
    data: Vec<u8>,
    timestamp: u64,
    expiry: Option<u64>,
    synced: bool,
}

impl OfflineCache {
    pub fn new(max_size: usize) -> Self {
        Self {
            cache: Arc::new(RwLock::new(HashMap::new())),
            pending_sync: Arc::new(RwLock::new(HashSet::new())),
            max_size,
            current_size: Arc::new(std::sync::atomic::AtomicUsize::new(0)),
        }
    }

    /// 存储数据
    pub async fn put(&self, key: &str, data: Vec<u8>, expiry_ms: Option<u64>) {
        let timestamp = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_millis() as u64;
        let expiry = expiry_ms.map(|ms| timestamp + ms);

        let entry = CacheEntry {
            data: data.clone(),
            timestamp,
            expiry,
            synced: false,
        };

        let mut cache = self.cache.write().await;
        cache.insert(key.to_string(), entry);

        self.current_size.fetch_add(data.len(), std::sync::atomic::Ordering::SeqCst);

        let mut pending = self.pending_sync.write().await;
        pending.insert(key.to_string());
    }

    /// 获取数据
    pub async fn get(&self, key: &str) -> Option<Vec<u8>> {
        let cache = self.cache.read().await;
        let entry = cache.get(key)?;

        // 检查过期
        if let Some(expiry) = entry.expiry {
            let now = SystemTime::now()
                .duration_since(UNIX_EPOCH)
                .unwrap()
                .as_millis() as u64;
            if now > expiry {
                return None;
            }
        }

        Some(entry.data.clone())
    }

    /// 获取待同步键列表
    pub async fn get_pending_keys(&self) -> Vec<String> {
        let pending = self.pending_sync.read().await;
        pending.iter().cloned().collect()
    }

    /// 标记已同步
    pub async fn mark_synced(&self, key: &str) {
        let mut pending = self.pending_sync.write().await;
        pending.remove(key);

        let mut cache = self.cache.write().await;
        if let Some(entry) = cache.get_mut(key) {
            entry.synced = true;
        }
    }
}

/// 离线管理器
pub struct OfflineManager {
    level: OfflineLevel,
    scheduler: LocalScheduler,
    cache: OfflineCache,
    offline_mode: Arc<std::sync::atomic::AtomicBool>,
}

impl OfflineManager {
    pub fn new(level: OfflineLevel) -> Self {
        Self {
            level,
            scheduler: LocalScheduler::new(4, level),
            cache: OfflineCache::new(10 * 1024 * 1024),
            offline_mode: Arc::new(std::sync::atomic::AtomicBool::new(level == OfflineLevel::L1)),
        }
    }

    pub async fn start(&self) {
        self.scheduler.start().await;
        log::info!("Offline manager started at level {}", self.level as u8);
    }

    pub async fn stop(&self) {
        self.scheduler.stop().await;
        log::info!("Offline manager stopped");
    }

    pub fn set_offline_mode(&self, offline: bool) {
        self.offline_mode.store(offline, std::sync::atomic::Ordering::SeqCst);
    }

    pub fn is_offline(&self) -> bool {
        self.offline_mode.load(std::sync::atomic::Ordering::SeqCst)
    }

    pub fn scheduler(&self) -> &LocalScheduler {
        &self.scheduler
    }

    pub fn cache(&self) -> &OfflineCache {
        &self.cache
    }
}