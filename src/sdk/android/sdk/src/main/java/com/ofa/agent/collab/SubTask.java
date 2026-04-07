package com.ofa.agent.collab;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.HashMap;
import java.util.Map;

/**
 * 子任务模型 (v3.3.0)
 */
public class SubTask {

    // 状态常量
    public static final String STATUS_PENDING = "pending";
    public static final String STATUS_ASSIGNED = "assigned";
    public static final String STATUS_RUNNING = "running";
    public static final String STATUS_COMPLETED = "completed";
    public static final String STATUS_FAILED = "failed";
    public static final String STATUS_CANCELLED = "cancelled";

    private String subTaskId;
    private String parentTaskId;
    private String assignedTo;      // AgentID
    private String type;
    private Map<String, Object> payload;
    private String status;
    private int priority;
    private Map<String, Object> result;
    private String error;
    private long createdAt;
    private Long assignedAt;
    private Long completedAt;
    private long timeoutMs;
    private int retryCount;
    private int maxRetries;

    public SubTask() {
        this.status = STATUS_PENDING;
        this.priority = CollaborativeTask.PRIORITY_NORMAL;
        this.timeoutMs = 30 * 60 * 1000L;
        this.maxRetries = 3;
        this.retryCount = 0;
        this.createdAt = System.currentTimeMillis();
    }

    /**
     * 创建子任务
     */
    @NonNull
    public static SubTask create(@NonNull String subTaskId, @NonNull String parentTaskId,
                                  @NonNull String type, @NonNull Map<String, Object> payload) {
        SubTask subTask = new SubTask();
        subTask.subTaskId = subTaskId;
        subTask.parentTaskId = parentTaskId;
        subTask.type = type;
        subTask.payload = payload;
        return subTask;
    }

    /**
     * 从 JSON 解析
     */
    @NonNull
    public static SubTask fromJson(@NonNull JSONObject json) throws JSONException {
        SubTask subTask = new SubTask();

        subTask.subTaskId = json.optString("subtask_id");
        subTask.parentTaskId = json.optString("parent_task_id");
        subTask.assignedTo = json.optString("assigned_to");
        subTask.type = json.optString("type");
        subTask.status = json.optString("status", STATUS_PENDING);
        subTask.priority = json.optInt("priority", CollaborativeTask.PRIORITY_NORMAL);
        subTask.timeoutMs = json.optLong("timeout_ms", 30 * 60 * 1000L);
        subTask.retryCount = json.optInt("retry_count", 0);
        subTask.maxRetries = json.optInt("max_retries", 3);
        subTask.createdAt = json.optLong("created_at", System.currentTimeMillis());
        subTask.error = json.optString("error", null);

        // 解析 payload
        JSONObject payloadJson = json.optJSONObject("payload");
        if (payloadJson != null) {
            subTask.payload = new HashMap<>();
            for (java.util.Iterator<String> it = payloadJson.keys(); it.hasNext(); ) {
                String key = it.next();
                subTask.payload.put(key, payloadJson.get(key));
            }
        }

        // 解析 result
        JSONObject resultJson = json.optJSONObject("result");
        if (resultJson != null) {
            subTask.result = new HashMap<>();
            for (java.util.Iterator<String> it = resultJson.keys(); it.hasNext(); ) {
                String key = it.next();
                subTask.result.put(key, resultJson.get(key));
            }
        }

        // 解析时间戳
        if (json.has("assigned_at")) {
            subTask.assignedAt = json.getLong("assigned_at");
        }
        if (json.has("completed_at")) {
            subTask.completedAt = json.getLong("completed_at");
        }

        return subTask;
    }

    /**
     * 转换为 JSON
     */
    @NonNull
    public JSONObject toJson() throws JSONException {
        JSONObject json = new JSONObject();

        json.put("subtask_id", subTaskId);
        json.put("parent_task_id", parentTaskId);
        json.put("assigned_to", assignedTo);
        json.put("type", type);
        json.put("status", status);
        json.put("priority", priority);
        json.put("timeout_ms", timeoutMs);
        json.put("retry_count", retryCount);
        json.put("max_retries", maxRetries);
        json.put("created_at", createdAt);

        if (payload != null) {
            JSONObject payloadJson = new JSONObject();
            for (Map.Entry<String, Object> e : payload.entrySet()) {
                payloadJson.put(e.getKey(), e.getValue());
            }
            json.put("payload", payloadJson);
        }

        if (result != null) {
            JSONObject resultJson = new JSONObject();
            for (Map.Entry<String, Object> e : result.entrySet()) {
                resultJson.put(e.getKey(), e.getValue());
            }
            json.put("result", resultJson);
        }

        if (error != null) {
            json.put("error", error);
        }

        if (assignedAt != null) {
            json.put("assigned_at", assignedAt);
        }
        if (completedAt != null) {
            json.put("completed_at", completedAt);
        }

        return json;
    }

    // === 状态检查 ===

    public boolean isPending() { return STATUS_PENDING.equals(status); }
    public boolean isAssigned() { return STATUS_ASSIGNED.equals(status); }
    public boolean isRunning() { return STATUS_RUNNING.equals(status); }
    public boolean isCompleted() { return STATUS_COMPLETED.equals(status); }
    public boolean isFailed() { return STATUS_FAILED.equals(status); }

    /**
     * 检查是否应该重试
     */
    public boolean shouldRetry() {
        return isFailed() && retryCount < maxRetries;
    }

    /**
     * 获取执行时长
     */
    public long getDuration() {
        if (assignedAt == null) {
            return 0;
        }
        long end = completedAt != null ? completedAt : System.currentTimeMillis();
        return end - assignedAt;
    }

    // === Getter/Setter ===

    public String getSubTaskId() { return subTaskId; }
    public void setSubTaskId(String subTaskId) { this.subTaskId = subTaskId; }

    public String getParentTaskId() { return parentTaskId; }
    public void setParentTaskId(String parentTaskId) { this.parentTaskId = parentTaskId; }

    public String getAssignedTo() { return assignedTo; }
    public void setAssignedTo(String assignedTo) { this.assignedTo = assignedTo; }

    public String getType() { return type; }
    public void setType(String type) { this.type = type; }

    public Map<String, Object> getPayload() { return payload; }
    public void setPayload(Map<String, Object> payload) { this.payload = payload; }

    public String getStatus() { return status; }
    public void setStatus(String status) { this.status = status; }

    public int getPriority() { return priority; }
    public void setPriority(int priority) { this.priority = priority; }

    public Map<String, Object> getResult() { return result; }
    public void setResult(Map<String, Object> result) { this.result = result; }

    public String getError() { return error; }
    public void setError(String error) { this.error = error; }

    public long getCreatedAt() { return createdAt; }
    public void setCreatedAt(long createdAt) { this.createdAt = createdAt; }

    public Long getAssignedAt() { return assignedAt; }
    public void setAssignedAt(Long assignedAt) { this.assignedAt = assignedAt; }

    public Long getCompletedAt() { return completedAt; }
    public void setCompletedAt(Long completedAt) { this.completedAt = completedAt; }

    public long getTimeoutMs() { return timeoutMs; }
    public void setTimeoutMs(long timeoutMs) { this.timeoutMs = timeoutMs; }

    public int getRetryCount() { return retryCount; }
    public void setRetryCount(int retryCount) { this.retryCount = retryCount; }

    public int getMaxRetries() { return maxRetries; }
    public void setMaxRetries(int maxRetries) { this.maxRetries = maxRetries; }

    /**
     * 增加重试计数
     */
    public void incrementRetry() {
        this.retryCount++;
    }

    @NonNull
    @Override
    public String toString() {
        return "SubTask{" +
                "subTaskId='" + subTaskId + '\'' +
                ", parentTaskId='" + parentTaskId + '\'' +
                ", type='" + type + '\'' +
                ", status='" + status + '\'' +
                ", assignedTo='" + assignedTo + '\'' +
                '}';
    }
}