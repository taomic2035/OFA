package com.ofa.agent.offline;

import android.os.Handler;
import android.os.HandlerThread;
import android.os.Looper;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.skill.SkillExecutor;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.atomic.AtomicInteger;

/**
 * 本地任务调度器 - 支持离线执行
 */
public class LocalScheduler {
    private static final String TAG = "LocalScheduler";

    private final int workerCount;
    private final OfflineLevel offlineLevel;
    private final Map<String, SkillExecutor> skills = new ConcurrentHashMap<>();
    private final Map<String, LocalTask> tasks = new ConcurrentHashMap<>();
    private final ExecutorService executor;
    private final Handler mainHandler = new Handler(Looper.getMainLooper());
    private final HandlerThread workerThread;
    private final Handler workerHandler;

    private volatile boolean running = false;
    private final AtomicInteger pendingCount = new AtomicInteger(0);
    private final AtomicInteger completedCount = new AtomicInteger(0);

    private TaskListener taskListener;

    public LocalScheduler(int workerCount, OfflineLevel level) {
        this.workerCount = workerCount;
        this.offlineLevel = level;
        this.executor = Executors.newFixedThreadPool(workerCount);
        this.workerThread = new HandlerThread("LocalSchedulerWorker");
        this.workerThread.start();
        this.workerHandler = new Handler(workerThread.getLooper());
    }

    /**
     * 启动调度器
     */
    public void start() {
        if (running) return;
        running = true;
        Log.i(TAG, "Local scheduler started with " + workerCount + " workers, level: " + offlineLevel.getValue());
    }

    /**
     * 停止调度器
     */
    public void stop() {
        running = false;
        executor.shutdown();
        workerThread.quit();
        Log.i(TAG, "Local scheduler stopped");
    }

    /**
     * 注册技能
     */
    public void registerSkill(@NonNull String skillId, @NonNull SkillExecutor executor, boolean offlineCapable) {
        skills.put(skillId, executor);
        Log.i(TAG, "Registered local skill: " + skillId + " (offline: " + offlineCapable + ")");
    }

    /**
     * 注销技能
     */
    public void unregisterSkill(@NonNull String skillId) {
        skills.remove(skillId);
    }

    /**
     * 提交任务
     */
    @NonNull
    public String submitTask(@NonNull String skillId, @Nullable byte[] input) {
        LocalTask task = new LocalTask(skillId, input);
        tasks.put(task.getId(), task);
        pendingCount.incrementAndGet();

        executeTask(task);
        Log.i(TAG, "Task submitted: " + task.getId() + " -> " + skillId);

        return task.getId();
    }

    private void executeTask(@NonNull LocalTask task) {
        if (!running) {
            task.setStatus(TaskStatus.FAILED);
            task.setError("Scheduler not running");
            notifyTaskFailed(task);
            return;
        }

        executor.execute(() -> {
            task.setStatus(TaskStatus.RUNNING);

            try {
                SkillExecutor executor = skills.get(task.getSkillId());
                if (executor == null) {
                    throw new Exception("Skill not found: " + task.getSkillId());
                }

                byte[] output = executor.execute(task.getInput());
                task.setOutput(output);
                task.setStatus(TaskStatus.COMPLETED);
                task.setCompletedAt(System.currentTimeMillis());
                task.setSyncPending(offlineLevel != OfflineLevel.L4);

                completedCount.incrementAndGet();
                pendingCount.decrementAndGet();

                notifyTaskCompleted(task);
                Log.i(TAG, "Task completed: " + task.getId());

            } catch (Exception e) {
                Log.e(TAG, "Task failed: " + task.getId(), e);

                if (task.canRetry()) {
                    task.incrementRetry();
                    task.setStatus(TaskStatus.PENDING);
                    workerHandler.postDelayed(() -> executeTask(task), 1000);
                    Log.w(TAG, "Task retry " + task.getRetryCount() + ": " + task.getId());
                } else {
                    task.setStatus(TaskStatus.FAILED);
                    task.setError(e.getMessage());
                    pendingCount.decrementAndGet();
                    notifyTaskFailed(task);
                }
            }
        });
    }

    /**
     * 获取任务
     */
    @Nullable
    public LocalTask getTask(@NonNull String taskId) {
        return tasks.get(taskId);
    }

    /**
     * 取消任务
     */
    public boolean cancelTask(@NonNull String taskId) {
        LocalTask task = tasks.get(taskId);
        if (task != null && task.getStatus() == TaskStatus.PENDING) {
            task.setStatus(TaskStatus.CANCELLED);
            pendingCount.decrementAndGet();
            return true;
        }
        return false;
    }

    /**
     * 列出待处理任务
     */
    @NonNull
    public List<String> listPendingTasks() {
        List<String> result = new ArrayList<>();
        for (Map.Entry<String, LocalTask> entry : tasks.entrySet()) {
            if (entry.getValue().getStatus() == TaskStatus.PENDING) {
                result.add(entry.getKey());
            }
        }
        return result;
    }

    /**
     * 列出已注册技能
     */
    @NonNull
    public List<String> listSkills() {
        return new ArrayList<>(skills.keySet());
    }

    public int getPendingCount() {
        return pendingCount.get();
    }

    public int getCompletedCount() {
        return completedCount.get();
    }

    public OfflineLevel getOfflineLevel() {
        return offlineLevel;
    }

    public void setTaskListener(TaskListener listener) {
        this.taskListener = listener;
    }

    private void notifyTaskCompleted(LocalTask task) {
        if (taskListener != null) {
            mainHandler.post(() -> taskListener.onTaskCompleted(task.getId(), task.getOutput()));
        }
    }

    private void notifyTaskFailed(LocalTask task) {
        if (taskListener != null) {
            mainHandler.post(() -> taskListener.onTaskFailed(task.getId(), task.getError()));
        }
    }

    public interface TaskListener {
        void onTaskCompleted(String taskId, byte[] output);
        void onTaskFailed(String taskId, String error);
    }
}