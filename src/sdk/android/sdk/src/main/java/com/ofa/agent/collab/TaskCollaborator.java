package com.ofa.agent.collab;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.messaging.Message;
import com.ofa.agent.messaging.MessageBus;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CopyOnWriteArrayList;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;

/**
 * 任务协同执行器 (v3.3.0)
 *
 * 负责：
 * - 接收并执行分配给本设备的子任务
 * - 向 Center 报告执行结果
 * - 支持本地任务编排（可选）
 */
public class TaskCollaborator {

    private static final String TAG = "TaskCollaborator";

    // 消息类型
    private static final String MSG_SUBTASK_ASSIGN = "subtask_assign";
    private static final String MSG_SUBTASK_RESULT = "subtask_result";
    private static final String MSG_TASK_CANCEL = "task_cancel";

    private final ExecutorService executor;
    private final Map<String, TaskExecutor> taskExecutors;
    private final Map<String, Future<?>> runningTasks;
    private final List<TaskListener> listeners;

    // 本地任务缓存
    private final Map<String, SubTask> activeSubTasks;
    private final Map<String, CollaborativeTask> localTasks;

    // 消息总线
    private MessageBus messageBus;

    // 设备信息
    private String agentId;
    private String identityId;

    // 配置
    private CollaboratorConfig config;

    /**
     * 任务执行器接口
     */
    public interface TaskExecutor {
        /**
         * 执行子任务
         * @return 执行结果
         */
        @NonNull
        Map<String, Object> execute(@NonNull SubTask subTask) throws Exception;

        /**
         * 获取支持的任务类型
         */
        @NonNull
        String getTaskType();

        /**
         * 检查是否可以执行
         */
        boolean canExecute(@NonNull SubTask subTask);

        /**
         * 取消任务
         */
        void cancel(@NonNull String subTaskId);
    }

    /**
     * 任务监听器
     */
    public interface TaskListener {
        void onSubTaskReceived(@NonNull SubTask subTask);
        void onSubTaskStarted(@NonNull SubTask subTask);
        void onSubTaskCompleted(@NonNull SubTask subTask, @NonNull Map<String, Object> result);
        void onSubTaskFailed(@NonNull SubTask subTask, @NonNull String error);
        void onTaskCancelled(@NonNull String taskId);
    }

    /**
     * 配置
     */
    public static class CollaboratorConfig {
        public int maxConcurrentTasks = 5;
        public long defaultTimeoutMs = 30 * 60 * 1000L;
        public boolean reportProgress = true;
    }

    public TaskCollaborator() {
        this.executor = Executors.newCachedThreadPool();
        this.taskExecutors = new ConcurrentHashMap<>();
        this.runningTasks = new ConcurrentHashMap<>();
        this.listeners = new CopyOnWriteArrayList<>();
        this.activeSubTasks = new ConcurrentHashMap<>();
        this.localTasks = new ConcurrentHashMap<>();
        this.config = new CollaboratorConfig();
    }

    /**
     * 初始化
     */
    public void initialize(@NonNull String agentId, @NonNull String identityId,
                           @Nullable MessageBus messageBus) {
        this.agentId = agentId;
        this.identityId = identityId;
        this.messageBus = messageBus;

        if (messageBus != null) {
            messageBus.addListener(this::handleMessage);
        }

        Log.i(TAG, "TaskCollaborator initialized for " + agentId);
    }

    /**
     * 设置配置
     */
    public void setConfig(@NonNull CollaboratorConfig config) {
        this.config = config;
    }

    // === 任务执行器注册 ===

    /**
     * 注册任务执行器
     */
    public void registerExecutor(@NonNull TaskExecutor executor) {
        taskExecutors.put(executor.getTaskType(), executor);
        Log.d(TAG, "Task executor registered: " + executor.getTaskType());
    }

    /**
     * 注销任务执行器
     */
    public void unregisterExecutor(@NonNull String taskType) {
        taskExecutors.remove(taskType);
        Log.d(TAG, "Task executor unregistered: " + taskType);
    }

    /**
     * 获取支持的任务类型
     */
    @NonNull
    public List<String> getSupportedTaskTypes() {
        return new ArrayList<>(taskExecutors.keySet());
    }

    // === 任务执行 ===

    /**
     * 接收并执行子任务
     */
    public void receiveSubTask(@NonNull SubTask subTask) {
        if (activeSubTasks.size() >= config.maxConcurrentTasks) {
            reportSubTaskError(subTask, "Max concurrent tasks reached");
            return;
        }

        // 存储子任务
        activeSubTasks.put(subTask.getSubTaskId(), subTask);
        subTask.setStatus(SubTask.STATUS_ASSIGNED);
        subTask.setAssignedTo(agentId);
        subTask.setAssignedAt(System.currentTimeMillis());

        // 通知监听器
        notifySubTaskReceived(subTask);

        // 查找执行器
        TaskExecutor executor = findExecutor(subTask);
        if (executor == null) {
            reportSubTaskError(subTask, "No executor for task type: " + subTask.getType());
            return;
        }

        // 执行任务
        executeSubTask(subTask, executor);
    }

    /**
     * 执行子任务
     */
    private void executeSubTask(@NonNull SubTask subTask, @NonNull TaskExecutor executor) {
        subTask.setStatus(SubTask.STATUS_RUNNING);
        notifySubTaskStarted(subTask);

        Future<?> future = this.executor.submit(() -> {
            try {
                // 执行
                Map<String, Object> result = executor.execute(subTask);

                // 成功
                subTask.setResult(result);
                subTask.setStatus(SubTask.STATUS_COMPLETED);
                subTask.setCompletedAt(System.currentTimeMillis());

                // 报告结果
                reportSubTaskResult(subTask);

                // 通知监听器
                notifySubTaskCompleted(subTask, result);

                Log.d(TAG, "SubTask completed: " + subTask.getSubTaskId());

            } catch (Exception e) {
                // 失败
                String error = e.getMessage() != null ? e.getMessage() : "Unknown error";
                subTask.setError(error);
                subTask.setStatus(SubTask.STATUS_FAILED);

                // 报告错误
                reportSubTaskError(subTask, error);

                // 通知监听器
                notifySubTaskFailed(subTask, error);

                Log.e(TAG, "SubTask failed: " + subTask.getSubTaskId(), e);

            } finally {
                // 清理
                runningTasks.remove(subTask.getSubTaskId());
                activeSubTasks.remove(subTask.getSubTaskId());
            }
        });

        runningTasks.put(subTask.getSubTaskId(), future);
    }

    /**
     * 取消子任务
     */
    public void cancelSubTask(@NonNull String subTaskId) {
        SubTask subTask = activeSubTasks.get(subTaskId);
        if (subTask == null) {
            return;
        }

        // 取消执行
        Future<?> future = runningTasks.remove(subTaskId);
        if (future != null) {
            future.cancel(true);
        }

        // 通知执行器取消
        TaskExecutor executor = taskExecutors.get(subTask.getType());
        if (executor != null) {
            executor.cancel(subTaskId);
        }

        subTask.setStatus(SubTask.STATUS_CANCELLED);
        activeSubTasks.remove(subTaskId);

        Log.d(TAG, "SubTask cancelled: " + subTaskId);
    }

    /**
     * 取消任务的所有子任务
     */
    public void cancelTask(@NonNull String parentTaskId) {
        List<String> toCancel = new ArrayList<>();
        for (SubTask st : activeSubTasks.values()) {
            if (parentTaskId.equals(st.getParentTaskId())) {
                toCancel.add(st.getSubTaskId());
            }
        }

        for (String subTaskId : toCancel) {
            cancelSubTask(subTaskId);
        }

        // 通知监听器
        notifyTaskCancelled(parentTaskId);
    }

    // === 结果报告 ===

    private void reportSubTaskResult(@NonNull SubTask subTask) {
        if (messageBus == null) {
            return;
        }

        try {
            Message msg = new Message();
            msg.id = generateMessageId();
            msg.fromAgent = agentId;
            msg.toAgent = "center";
            msg.identityId = identityId;
            msg.type = Message.TYPE_DATA;
            msg.priority = Message.PRIORITY_HIGH;

            Map<String, Object> payload = new HashMap<>();
            payload.put("action", MSG_SUBTASK_RESULT);
            payload.put("subtask_id", subTask.getSubTaskId());
            payload.put("parent_task_id", subTask.getParentTaskId());
            payload.put("status", subTask.getStatus());
            payload.put("result", subTask.getResult());
            msg.payload = payload;

            messageBus.send(msg);

        } catch (Exception e) {
            Log.e(TAG, "Failed to report subtask result", e);
        }
    }

    private void reportSubTaskError(@NonNull SubTask subTask, @NonNull String error) {
        if (messageBus == null) {
            return;
        }

        try {
            Message msg = new Message();
            msg.id = generateMessageId();
            msg.fromAgent = agentId;
            msg.toAgent = "center";
            msg.identityId = identityId;
            msg.type = Message.TYPE_DATA;
            msg.priority = Message.PRIORITY_HIGH;

            Map<String, Object> payload = new HashMap<>();
            payload.put("action", MSG_SUBTASK_RESULT);
            payload.put("subtask_id", subTask.getSubTaskId());
            payload.put("parent_task_id", subTask.getParentTaskId());
            payload.put("status", SubTask.STATUS_FAILED);
            payload.put("error", error);
            msg.payload = payload;

            messageBus.send(msg);

        } catch (Exception e) {
            Log.e(TAG, "Failed to report subtask error", e);
        }
    }

    // === 消息处理 ===

    private void handleMessage(@NonNull Message message) {
        if (message.payload == null) {
            return;
        }

        Object actionObj = message.payload.get("action");
        if (actionObj == null) {
            return;
        }

        String action = actionObj.toString();

        switch (action) {
            case MSG_SUBTASK_ASSIGN:
                handleSubTaskAssign(message);
                break;

            case MSG_TASK_CANCEL:
                handleTaskCancel(message);
                break;
        }
    }

    private void handleSubTaskAssign(@NonNull Message message) {
        try {
            Object subTaskObj = message.payload.get("subtask");
            SubTask subTask = null;

            if (subTaskObj instanceof JSONObject) {
                subTask = SubTask.fromJson((JSONObject) subTaskObj);
            } else if (subTaskObj instanceof Map) {
                JSONObject json = new JSONObject((Map) subTaskObj);
                subTask = SubTask.fromJson(json);
            }

            if (subTask != null) {
                receiveSubTask(subTask);
            }

        } catch (JSONException e) {
            Log.e(TAG, "Failed to parse subtask assignment", e);
        }
    }

    private void handleTaskCancel(@NonNull Message message) {
        Object taskIdObj = message.payload.get("task_id");
        if (taskIdObj != null) {
            cancelTask(taskIdObj.toString());
        }
    }

    // === 本地任务管理 ===

    /**
     * 创建本地任务
     */
    @NonNull
    public CollaborativeTask createLocalTask(@NonNull String type,
                                              @NonNull Map<String, Object> payload) {
        String taskId = "local_" + System.currentTimeMillis();
        CollaborativeTask task = CollaborativeTask.create(taskId, type, payload);
        task.setIdentityId(identityId);

        localTasks.put(taskId, task);

        return task;
    }

    /**
     * 获取本地任务
     */
    @Nullable
    public CollaborativeTask getLocalTask(@NonNull String taskId) {
        return localTasks.get(taskId);
    }

    /**
     * 移除本地任务
     */
    public void removeLocalTask(@NonNull String taskId) {
        localTasks.remove(taskId);
        cancelTask(taskId);
    }

    // === 监听器管理 ===

    public void addListener(@NonNull TaskListener listener) {
        listeners.add(listener);
    }

    public void removeListener(@NonNull TaskListener listener) {
        listeners.remove(listener);
    }

    private void notifySubTaskReceived(@NonNull SubTask subTask) {
        for (TaskListener l : listeners) {
            l.onSubTaskReceived(subTask);
        }
    }

    private void notifySubTaskStarted(@NonNull SubTask subTask) {
        for (TaskListener l : listeners) {
            l.onSubTaskStarted(subTask);
        }
    }

    private void notifySubTaskCompleted(@NonNull SubTask subTask, @NonNull Map<String, Object> result) {
        for (TaskListener l : listeners) {
            l.onSubTaskCompleted(subTask, result);
        }
    }

    private void notifySubTaskFailed(@NonNull SubTask subTask, @NonNull String error) {
        for (TaskListener l : listeners) {
            l.onSubTaskFailed(subTask, error);
        }
    }

    private void notifyTaskCancelled(@NonNull String taskId) {
        for (TaskListener l : listeners) {
            l.onTaskCancelled(taskId);
        }
    }

    // === 辅助方法 ===

    private TaskExecutor findExecutor(@NonNull SubTask subTask) {
        // 先按类型精确匹配
        TaskExecutor executor = taskExecutors.get(subTask.getType());
        if (executor != null && executor.canExecute(subTask)) {
            return executor;
        }

        // 再检查所有执行器
        for (TaskExecutor e : taskExecutors.values()) {
            if (e.canExecute(subTask)) {
                return e;
            }
        }

        return null;
    }

    private String generateMessageId() {
        return "msg_" + System.currentTimeMillis() + "_" + agentId;
    }

    // === 状态查询 ===

    /**
     * 获取活跃子任务数
     */
    public int getActiveSubTaskCount() {
        return activeSubTasks.size();
    }

    /**
     * 获取活跃子任务列表
     */
    @NonNull
    public List<SubTask> getActiveSubTasks() {
        return new ArrayList<>(activeSubTasks.values());
    }

    /**
     * 获取统计信息
     */
    @NonNull
    public CollaboratorStats getStats() {
        CollaboratorStats stats = new CollaboratorStats();
        stats.activeSubTasks = activeSubTasks.size();
        stats.localTasks = localTasks.size();
        stats.registeredExecutors = taskExecutors.size();
        stats.runningTasks = runningTasks.size();
        return stats;
    }

    /**
     * 统计信息
     */
    public static class CollaboratorStats {
        public int activeSubTasks;
        public int localTasks;
        public int registeredExecutors;
        public int runningTasks;

        @NonNull
        @Override
        public String toString() {
            return "CollaboratorStats{" +
                    "active=" + activeSubTasks +
                    ", local=" + localTasks +
                    ", executors=" + registeredExecutors +
                    ", running=" + runningTasks +
                    '}';
        }
    }

    /**
     * 清理资源
     */
    public void cleanup() {
        // 取消所有任务
        for (String subTaskId : new ArrayList<>(activeSubTasks.keySet())) {
            cancelSubTask(subTaskId);
        }

        executor.shutdown();
        listeners.clear();
        activeSubTasks.clear();
        localTasks.clear();

        Log.i(TAG, "TaskCollaborator cleaned up");
    }
}