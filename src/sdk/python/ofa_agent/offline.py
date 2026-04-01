"""
OFA Offline Module
支持离线能力等级 L1-L4
"""

import asyncio
import json
import logging
import os
import sqlite3
import threading
import time
import uuid
from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Dict, List, Optional, Callable, Tuple
from queue import Queue, Empty
from concurrent.futures import ThreadPoolExecutor, Future

logger = logging.getLogger(__name__)


class OfflineLevel(Enum):
    """离线能力等级"""
    NONE = 0      # 不支持离线
    L1 = 1        # 完全离线 (本地执行)
    L2 = 2        # 局域网协作
    L3 = 3        # 弱网同步
    L4 = 4        # 在线模式


class TaskStatus(Enum):
    """任务状态"""
    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    CANCELLED = "cancelled"


@dataclass
class OfflineTask:
    """离线任务"""
    id: str
    skill_id: str
    input_data: Any
    output_data: Any = None
    status: TaskStatus = TaskStatus.PENDING
    error: Optional[str] = None
    created_at: float = field(default_factory=time.time)
    completed_at: Optional[float] = None
    retry_count: int = 0
    max_retries: int = 3
    sync_pending: bool = True

    def to_dict(self) -> Dict[str, Any]:
        return {
            "id": self.id,
            "skill_id": self.skill_id,
            "input": self.input_data,
            "output": self.output_data,
            "status": self.status.value,
            "error": self.error,
            "created_at": self.created_at,
            "completed_at": self.completed_at,
            "retry_count": self.retry_count,
            "sync_pending": self.sync_pending,
        }


class OfflineCache:
    """离线数据缓存"""

    def __init__(self, max_size: int = 10 * 1024 * 1024, db_path: Optional[str] = None):
        self.max_size = max_size
        self.current_size = 0
        self._cache: Dict[str, Tuple[Any, float, float, bool]] = {}  # key -> (data, timestamp, expiry, synced)
        self._pending_sync: List[str] = []
        self._hits = 0
        self._misses = 0
        self._lock = threading.RLock()

        # 持久化存储
        self._db_path = db_path or os.path.join(os.getcwd(), "ofa_offline_cache.db")
        self._init_db()

    def _init_db(self):
        """初始化数据库"""
        conn = sqlite3.connect(self._db_path)
        conn.execute("""
            CREATE TABLE IF NOT EXISTS cache (
                key TEXT PRIMARY KEY,
                data TEXT,
                timestamp REAL,
                expiry REAL,
                synced INTEGER
            )
        """)
        conn.execute("""
            CREATE TABLE IF NOT EXISTS pending_sync (
                key TEXT PRIMARY KEY
            )
        """)
        conn.commit()
        conn.close()

    def put(self, key: str, data: Any, expiry_ms: int = 0) -> bool:
        """存储数据"""
        with self._lock:
            data_str = json.dumps(data) if not isinstance(data, str) else data
            data_size = len(data_str.encode('utf-8'))

            # 检查容量
            if self.current_size + data_size > self.max_size:
                self._evictIfNeeded(data_size)

            timestamp = time.time()
            expiry = expiry_ms / 1000 if expiry_ms > 0 else 0

            # 内存缓存
            self._cache[key] = (data, timestamp, expiry, False)
            self.current_size += data_size

            # 待同步队列
            if key not in self._pending_sync:
                self._pending_sync.append(key)

            # 持久化
            conn = sqlite3.connect(self._db_path)
            conn.execute(
                "INSERT OR REPLACE INTO cache VALUES (?, ?, ?, ?, 0)",
                (key, data_str, timestamp, expiry)
            )
            conn.execute(
                "INSERT OR IGNORE INTO pending_sync VALUES (?)",
                (key,)
            )
            conn.commit()
            conn.close()

            return True

    def get(self, key: str) -> Optional[Any]:
        """获取数据"""
        with self._lock:
            # 先检查内存
            if key in self._cache:
                data, timestamp, expiry, synced = self._cache[key]
                if expiry > 0 and time.time() > timestamp + expiry:
                    self._remove(key)
                    self._misses += 1
                    return None
                self._hits += 1
                return data

            # 检查数据库
            conn = sqlite3.connect(self._db_path)
            cursor = conn.execute(
                "SELECT data, timestamp, expiry FROM cache WHERE key = ?",
                (key,)
            )
            row = cursor.fetchone()
            conn.close()

            if row:
                data_str, timestamp, expiry = row
                if expiry > 0 and time.time() > timestamp + expiry:
                    self._remove(key)
                    self._misses += 1
                    return None

                data = json.loads(data_str)
                self._cache[key] = (data, timestamp, expiry, False)
                self._hits += 1
                return data

            self._misses += 1
            return None

    def _remove(self, key: str):
        """删除缓存项"""
        with self._lock:
            if key in self._cache:
                _, _, _, _ = self._cache[key]
                # Note: size tracking needs improvement
                del self._cache[key]

            conn = sqlite3.connect(self._db_path)
            conn.execute("DELETE FROM cache WHERE key = ?", (key,))
            conn.execute("DELETE FROM pending_sync WHERE key = ?", (key,))
            conn.commit()
            conn.close()

            if key in self._pending_sync:
                self._pending_sync.remove(key)

    def _evictIfNeeded(self, needed: int):
        """清理过期/旧数据"""
        now = time.time()
        with self._lock:
            # 清理过期项
            expired = [
                k for k, (_, ts, exp, _) in self._cache.items()
                if exp > 0 and now > ts + exp
            ]
            for k in expired:
                self._remove(k)

    def get_pending_keys(self) -> List[str]:
        """获取待同步键列表"""
        with self._lock:
            return self._pending_sync.copy()

    def mark_synced(self, key: str):
        """标记已同步"""
        with self._lock:
            if key in self._cache:
                data, ts, exp, _ = self._cache[key]
                self._cache[key] = (data, ts, exp, True)

            conn = sqlite3.connect(self._db_path)
            conn.execute("UPDATE cache SET synced = 1 WHERE key = ?", (key,))
            conn.execute("DELETE FROM pending_sync WHERE key = ?", (key,))
            conn.commit()
            conn.close()

            if key in self._pending_sync:
                self._pending_sync.remove(key)

    def hit_rate(self) -> float:
        """命中率"""
        total = self._hits + self._misses
        return self._hits / total if total > 0 else 0.0

    def clear(self):
        """清空缓存"""
        with self._lock:
            self._cache.clear()
            self._pending_sync.clear()
            self.current_size = 0

            conn = sqlite3.connect(self._db_path)
            conn.execute("DELETE FROM cache")
            conn.execute("DELETE FROM pending_sync")
            conn.commit()
            conn.close()


class LocalScheduler:
    """本地任务调度器"""

    def __init__(self, worker_count: int = 4, offline_level: OfflineLevel = OfflineLevel.L1):
        self.worker_count = worker_count
        self.offline_level = offline_level
        self._skills: Dict[str, Callable] = {}
        self._skill_metadata: Dict[str, Dict[str, Any]] = {}
        self._tasks: Dict[str, OfflineTask] = {}
        self._task_queue: Queue = Queue()
        self._executor: Optional[ThreadPoolExecutor] = None
        self._running = False
        self._lock = threading.RLock()
        self._pending_count = 0
        self._completed_count = 0

    def start(self):
        """启动调度器"""
        if self._running:
            return

        self._running = True
        self._executor = ThreadPoolExecutor(max_workers=self.worker_count)

        for i in range(self.worker_count):
            self._executor.submit(self._worker_loop, i)

        logger.info(f"Local scheduler started with {self.worker_count} workers, level: {self.offline_level.value}")

    def stop(self):
        """停止调度器"""
        self._running = False
        if self._executor:
            self._executor.shutdown(wait=True)
        logger.info("Local scheduler stopped")

    def _worker_loop(self, worker_id: int):
        """工作线程循环"""
        logger.debug(f"Local worker {worker_id} started")

        while self._running:
            try:
                task = self._task_queue.get(timeout=0.1)
                if task:
                    self._execute_task(task)
            except Empty:
                continue
            except Exception as e:
                logger.error(f"Worker {worker_id} error: {e}")

    def _execute_task(self, task: OfflineTask):
        """执行任务"""
        with self._lock:
            task.status = TaskStatus.RUNNING

        try:
            handler = self._skills.get(task.skill_id)
            if not handler:
                task.status = TaskStatus.FAILED
                task.error = f"Skill not found: {task.skill_id}"
                return

            # 检查技能是否支持离线
            skill_meta = self._skill_metadata.get(task.skill_id, {})
            if not skill_meta.get("offline_capable", True):
                task.status = TaskStatus.FAILED
                task.error = "Skill does not support offline execution"
                return

            # 执行技能
            result = handler(task.input_data)
            task.output_data = result
            task.status = TaskStatus.COMPLETED
            task.completed_at = time.time()
            task.sync_pending = self.offline_level != OfflineLevel.L4

            with self._lock:
                self._completed_count += 1
                self._pending_count -= 1

            logger.info(f"Task {task.id} completed: {task.skill_id}")

        except Exception as e:
            task.status = TaskStatus.FAILED
            task.error = str(e)

            if task.retry_count < task.max_retries:
                task.retry_count += 1
                task.status = TaskStatus.PENDING
                self._task_queue.put(task)
                logger.warning(f"Task {task.id} retry {task.retry_count}")
            else:
                with self._lock:
                    self._pending_count -= 1
                logger.error(f"Task {task.id} failed: {e}")

    def register_skill(
        self,
        skill_id: str,
        handler: Callable,
        offline_capable: bool = True,
        category: str = "general"
    ):
        """注册技能"""
        with self._lock:
            self._skills[skill_id] = handler
            self._skill_metadata[skill_id] = {
                "offline_capable": offline_capable,
                "category": category,
            }
        logger.info(f"Registered local skill: {skill_id} (offline: {offline_capable})")

    def submit_task(self, skill_id: str, input_data: Any) -> str:
        """提交任务"""
        task_id = f"local-{uuid.uuid4().hex[:8]}"
        task = OfflineTask(
            id=task_id,
            skill_id=skill_id,
            input_data=input_data,
        )

        with self._lock:
            self._tasks[task_id] = task
            self._pending_count += 1

        self._task_queue.put(task)
        logger.info(f"Task submitted: {task_id} -> {skill_id}")

        return task_id

    def submit_async(self, skill_id: str, input_data: Any) -> Future:
        """异步提交任务"""
        task_id = self.submit_task(skill_id, input_data)
        future: Future = Future()

        # 监控任务完成
        def monitor():
            while self._running:
                with self._lock:
                    task = self._tasks.get(task_id)
                    if task and task.status in [TaskStatus.COMPLETED, TaskStatus.FAILED, TaskStatus.CANCELLED]:
                        if task.status == TaskStatus.COMPLETED:
                            future.set_result(task.output_data)
                        else:
                            future.set_exception(Exception(task.error or "Task failed"))
                        return
                time.sleep(0.05)

        threading.Thread(target=monitor, daemon=True).start()
        return future

    def get_task(self, task_id: str) -> Optional[OfflineTask]:
        """获取任务"""
        with self._lock:
            return self._tasks.get(task_id)

    def cancel_task(self, task_id: str) -> bool:
        """取消任务"""
        with self._lock:
            task = self._tasks.get(task_id)
            if task and task.status == TaskStatus.PENDING:
                task.status = TaskStatus.CANCELLED
                return True
        return False

    def list_skills(self) -> List[str]:
        """列出已注册技能"""
        with self._lock:
            return list(self._skills.keys())

    def list_pending_tasks(self) -> List[str]:
        """列出待处理任务"""
        with self._lock:
            return [
                t.id for t in self._tasks.values()
                if t.status == TaskStatus.PENDING
            ]

    def pending_count(self) -> int:
        """待处理任务数"""
        return self._pending_count

    def completed_count(self) -> int:
        """已完成任务数"""
        return self._completed_count


class OfflineManager:
    """离线管理器 - 综合管理离线模式"""

    def __init__(self, offline_level: OfflineLevel = OfflineLevel.L1):
        self.level = offline_level
        self.scheduler = LocalScheduler(worker_count=4, offline_level=offline_level)
        self.cache = OfflineCache()
        self._offline_mode = offline_level == OfflineLevel.L1
        self._sync_callback: Optional[Callable] = None

    def start(self):
        """启动离线管理器"""
        self.scheduler.start()
        logger.info(f"Offline manager started at level {self.level.value}")

    def stop(self):
        """停止离线管理器"""
        self.scheduler.stop()
        self.cache.clear()
        logger.info("Offline manager stopped")

    def set_offline_mode(self, offline: bool):
        """设置离线模式"""
        self._offline_mode = offline
        logger.info(f"Offline mode: {offline}")

    def is_offline(self) -> bool:
        """是否处于离线模式"""
        return self._offline_mode

    def register_skill(self, skill_id: str, handler: Callable, offline_capable: bool = True):
        """注册技能"""
        self.scheduler.register_skill(skill_id, handler, offline_capable)

    def execute_local(self, skill_id: str, input_data: Any) -> str:
        """本地执行"""
        return self.scheduler.submit_task(skill_id, input_data)

    def execute_sync(self, skill_id: str, input_data: Any, timeout: float = 30.0) -> Any:
        """同步执行"""
        future = self.scheduler.submit_async(skill_id, input_data)
        return future.result(timeout=timeout)

    def cache_data(self, key: str, data: Any, expiry_ms: int = 0):
        """缓存数据"""
        self.cache.put(key, data, expiry_ms)

    def get_cached(self, key: str) -> Optional[Any]:
        """获取缓存"""
        return self.cache.get(key)

    def get_pending_sync(self) -> List[str]:
        """获取待同步数据"""
        return self.cache.get_pending_keys()

    def sync_now(self) -> bool:
        """立即同步"""
        pending = self.get_pending_sync()
        if not pending:
            return True

        if self._sync_callback:
            for key in pending:
                data = self.cache.get(key)
                if data:
                    try:
                        self._sync_callback(key, data)
                        self.cache.mark_synced(key)
                    except Exception as e:
                        logger.error(f"Sync failed for {key}: {e}")
                        return False
            return True

        logger.warning("No sync callback configured")
        return False

    def set_sync_callback(self, callback: Callable):
        """设置同步回调"""
        self._sync_callback = callback

    def get_stats(self) -> Dict[str, Any]:
        """获取统计信息"""
        return {
            "offline_mode": self._offline_mode,
            "level": self.level.value,
            "pending_tasks": self.scheduler.pending_count(),
            "completed_tasks": self.scheduler.completed_count(),
            "pending_sync": len(self.get_pending_sync()),
            "cache_hit_rate": self.cache.hit_rate(),
        }