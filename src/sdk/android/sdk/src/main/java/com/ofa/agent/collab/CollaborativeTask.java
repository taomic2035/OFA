package com.ofa.agent.collab;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.CopyOnWriteArrayList;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * 协同任务模型 (v3.3.0)
 *
 * 表示需要在多个设备间协同完成的任务
 */
public class CollaborativeTask {

    // 任务状态
    public static final String STATUS_PENDING = "pending";
    public static final String STATUS_ASSIGNED = "assigned";
    public static final String STATUS_RUNNING = "running";
    public static final String STATUS_COMPLETED = "completed";
    public static final String STATUS_FAILED = "failed";
    public static final String STATUS_CANCELLED = "cancelled";
    public static final String STATUS_TIMEOUT = "timeout";

    // 任务优先级
    public static final int PRIORITY_LOW = 0;
    public static final int PRIORITY_NORMAL = 1;
    public static final int PRIORITY_HIGH = 2;
    public static final int PRIORITY_URGENT = 3;

    // 拆分策略
    public static final String SPLIT_NONE = "none";
    public static final String SPLIT_PARALLEL = "parallel";
    public static final String SPLIT_SEQUENCE = "sequence";
    public static final String SPLIT_MAP_REDUCE = "map_reduce";
    public static final String SPLIT_BY_DEVICE = "by_device";

    // 合并策略
    public static final String MERGE_NONE = "none";
    public static final String MERGE_ALL = "all";
    public static final String MERGE_FIRST = "first";
    public static final String MERGE_CONSENSUS = "consensus";
    public static final String MERGE_AGGREGATE = "aggregate";
    public static final String MERGE_BEST = "best";

    // 任务字段
    private String taskId;
    private String identityId;
    private String type;
    private String description;
    private int priority;
    private String status;
    private Map<String, Object> payload;

    // 拆分与合并
    private String splitStrategy;
    private String mergeStrategy;
    private List<SubTask> subTasks;

    // 执行约束
    private List<String> requiredCapabilities;
    private List<String> preferredDevices;
    private int minDevices;
    private int maxDevices;
    private long timeoutMs;
    private int maxRetries;

    // 结果
    private Map<String, Object> result;
    private String error;

    // 时间信息
    private long createdAt;
    private Long startedAt;
    private Long completedAt;

    // 元数据
    private Map<String, Object> metadata;

    // === 构造函数 ===

    public CollaborativeTask() {
        this.priority = PRIORITY_NORMAL;
        this.status = STATUS_PENDING;
        this.splitStrategy = SPLIT_NONE;
        this.mergeStrategy = MERGE_ALL;
        this.subTasks = new ArrayList<>();
        this.requiredCapabilities = new ArrayList<>();
        this.preferredDevices = new ArrayList<>();
        this.minDevices = 1;
        this.maxDevices = Integer.MAX_VALUE;
        this.timeoutMs = 30 * 60 * 1000L; // 30分钟
        this.maxRetries = 3;
        this.metadata = new HashMap<>();
        this.createdAt = System.currentTimeMillis();
    }

    // === 工厂方法 ===

    /**
     * 创建简单任务
     */
    @NonNull
    public static CollaborativeTask create(@NonNull String taskId, @NonNull String type,
                                            @NonNull Map<String, Object> payload) {
        CollaborativeTask task = new CollaborativeTask();
        task.taskId = taskId;
        task.type = type;
        task.payload = payload;
        return task;
    }

    /**
     * 创建并行任务
     */
    @NonNull
    public static CollaborativeTask createParallel(@NonNull String taskId, @NonNull String type,
                                                    @NonNull Map<String, Object> payload,
                                                    int minDevices) {
        CollaborativeTask task = create(taskId, type, payload);
        task.splitStrategy = SPLIT_PARALLEL;
        task.mergeStrategy = MERGE_ALL;
        task.minDevices = minDevices;
        return task;
    }

    /**
     * 从 JSON 解析
     */
    @NonNull
    public static CollaborativeTask fromJson(@NonNull JSONObject json) throws JSONException {
        CollaborativeTask task = new CollaborativeTask();

        task.taskId = json.optString("task_id");
        task.identityId = json.optString("identity_id");
        task.type = json.optString("type");
        task.description = json.optString("description");
        task.priority = json.optInt("priority", PRIORITY_NORMAL);
        task.status = json.optString("status", STATUS_PENDING);
        task.splitStrategy = json.optString("split_strategy", SPLIT_NONE);
        task.mergeStrategy = json.optString("merge_strategy", MERGE_ALL);
        task.minDevices = json.optInt("min_devices", 1);
        task.maxDevices = json.optInt("max_devices", Integer.MAX_VALUE);
        task.timeoutMs = json.optLong("timeout", 30 * 60 * 1000L);
        task.maxRetries = json.optInt("max_retries", 3);
        task.createdAt = json.optLong("created_at", System.currentTimeMillis());
        task.error = json.optString("error", null);

        // 解析 payload
        JSONObject payloadJson = json.optJSONObject("payload");
        if (payloadJson != null) {
            task.payload = new HashMap<>();
            for (java.util.Iterator<String> it = payloadJson.keys(); it.hasNext(); ) {
                String key = it.next();
                task.payload.put(key, payloadJson.get(key));
            }
        }

        // 解析 result
        JSONObject resultJson = json.optJSONObject("result");
        if (resultJson != null) {
            task.result = new HashMap<>();
            for (java.util.Iterator<String> it = resultJson.keys(); it.hasNext(); ) {
                String key = it.next();
                task.result.put(key, resultJson.get(key));
            }
        }

        // 解析 requiredCapabilities
        JSONArray capsArray = json.optJSONArray("required_capabilities");
        if (capsArray != null) {
            task.requiredCapabilities = new ArrayList<>();
            for (int i = 0; i < capsArray.length(); i++) {
                task.requiredCapabilities.add(capsArray.getString(i));
            }
        }

        // 解析 preferredDevices
        JSONArray devicesArray = json.optJSONArray("preferred_devices");
        if (devicesArray != null) {
            task.preferredDevices = new ArrayList<>();
            for (int i = 0; i < devicesArray.length(); i++) {
                task.preferredDevices.add(devicesArray.getString(i));
            }
        }

        // 解析 subTasks
        JSONArray subTasksArray = json.optJSONArray("subtasks");
        if (subTasksArray != null) {
            task.subTasks = new ArrayList<>();
            for (int i = 0; i < subTasksArray.length(); i++) {
                task.subTasks.add(SubTask.fromJson(subTasksArray.getJSONObject(i)));
            }
        }

        // 解析时间戳
        if (json.has("started_at")) {
            task.startedAt = json.getLong("started_at");
        }
        if (json.has("completed_at")) {
            task.completedAt = json.getLong("completed_at");
        }

        // 解析 metadata
        JSONObject metadataJson = json.optJSONObject("metadata");
        if (metadataJson != null) {
            task.metadata = new HashMap<>();
            for (java.util.Iterator<String> it = metadataJson.keys(); it.hasNext(); ) {
                String key = it.next();
                task.metadata.put(key, metadataJson.get(key));
            }
        }

        return task;
    }

    // === JSON 序列化 ===

    @NonNull
    public JSONObject toJson() throws JSONException {
        JSONObject json = new JSONObject();

        json.put("task_id", taskId);
        json.put("identity_id", identityId);
        json.put("type", type);
        json.put("description", description);
        json.put("priority", priority);
        json.put("status", status);
        json.put("split_strategy", splitStrategy);
        json.put("merge_strategy", mergeStrategy);
        json.put("min_devices", minDevices);
        json.put("max_devices", maxDevices);
        json.put("timeout", timeoutMs);
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

        if (!requiredCapabilities.isEmpty()) {
            json.put("required_capabilities", new JSONArray(requiredCapabilities));
        }

        if (!preferredDevices.isEmpty()) {
            json.put("preferred_devices", new JSONArray(preferredDevices));
        }

        if (!subTasks.isEmpty()) {
            JSONArray subTasksArray = new JSONArray();
            for (SubTask st : subTasks) {
                subTasksArray.put(st.toJson());
            }
            json.put("subtasks", subTasksArray);
        }

        if (startedAt != null) {
            json.put("started_at", startedAt);
        }
        if (completedAt != null) {
            json.put("completed_at", completedAt);
        }

        if (!metadata.isEmpty()) {
            JSONObject metadataJson = new JSONObject();
            for (Map.Entry<String, Object> e : metadata.entrySet()) {
                metadataJson.put(e.getKey(), e.getValue());
            }
            json.put("metadata", metadataJson);
        }

        return json;
    }

    // === 状态检查 ===

    public boolean isPending() {
        return STATUS_PENDING.equals(status);
    }

    public boolean isRunning() {
        return STATUS_RUNNING.equals(status);
    }

    public boolean isCompleted() {
        return STATUS_COMPLETED.equals(status);
    }

    public boolean isFailed() {
        return STATUS_FAILED.equals(status);
    }

    public boolean isCancelled() {
        return STATUS_CANCELLED.equals(status);
    }

    // === Getter/Setter ===

    public String getTaskId() { return taskId; }
    public void setTaskId(String taskId) { this.taskId = taskId; }

    public String getIdentityId() { return identityId; }
    public void setIdentityId(String identityId) { this.identityId = identityId; }

    public String getType() { return type; }
    public void setType(String type) { this.type = type; }

    public String getDescription() { return description; }
    public void setDescription(String description) { this.description = description; }

    public int getPriority() { return priority; }
    public void setPriority(int priority) { this.priority = priority; }

    public String getStatus() { return status; }
    public void setStatus(String status) { this.status = status; }

    public Map<String, Object> getPayload() { return payload; }
    public void setPayload(Map<String, Object> payload) { this.payload = payload; }

    public String getSplitStrategy() { return splitStrategy; }
    public void setSplitStrategy(String splitStrategy) { this.splitStrategy = splitStrategy; }

    public String getMergeStrategy() { return mergeStrategy; }
    public void setMergeStrategy(String mergeStrategy) { this.mergeStrategy = mergeStrategy; }

    public List<SubTask> getSubTasks() { return subTasks; }
    public void setSubTasks(List<SubTask> subTasks) { this.subTasks = subTasks; }

    public List<String> getRequiredCapabilities() { return requiredCapabilities; }
    public void setRequiredCapabilities(List<String> requiredCapabilities) { this.requiredCapabilities = requiredCapabilities; }

    public List<String> getPreferredDevices() { return preferredDevices; }
    public void setPreferredDevices(List<String> preferredDevices) { this.preferredDevices = preferredDevices; }

    public int getMinDevices() { return minDevices; }
    public void setMinDevices(int minDevices) { this.minDevices = minDevices; }

    public int getMaxDevices() { return maxDevices; }
    public void setMaxDevices(int maxDevices) { this.maxDevices = maxDevices; }

    public long getTimeoutMs() { return timeoutMs; }
    public void setTimeoutMs(long timeoutMs) { this.timeoutMs = timeoutMs; }

    public int getMaxRetries() { return maxRetries; }
    public void setMaxRetries(int maxRetries) { this.maxRetries = maxRetries; }

    public Map<String, Object> getResult() { return result; }
    public void setResult(Map<String, Object> result) { this.result = result; }

    public String getError() { return error; }
    public void setError(String error) { this.error = error; }

    public long getCreatedAt() { return createdAt; }
    public void setCreatedAt(long createdAt) { this.createdAt = createdAt; }

    public Long getStartedAt() { return startedAt; }
    public void setStartedAt(Long startedAt) { this.startedAt = startedAt; }

    public Long getCompletedAt() { return completedAt; }
    public void setCompletedAt(Long completedAt) { this.completedAt = completedAt; }

    public Map<String, Object> getMetadata() { return metadata; }
    public void setMetadata(Map<String, Object> metadata) { this.metadata = metadata; }

    /**
     * 添加子任务
     */
    public void addSubTask(@NonNull SubTask subTask) {
        if (subTasks == null) {
            subTasks = new ArrayList<>();
        }
        subTasks.add(subTask);
    }

    /**
     * 添加所需能力
     */
    public void addRequiredCapability(@NonNull String capability) {
        if (requiredCapabilities == null) {
            requiredCapabilities = new ArrayList<>();
        }
        if (!requiredCapabilities.contains(capability)) {
            requiredCapabilities.add(capability);
        }
    }

    /**
     * 获取执行时长（毫秒）
     */
    public long getDuration() {
        if (startedAt == null) {
            return 0;
        }
        long end = completedAt != null ? completedAt : System.currentTimeMillis();
        return end - startedAt;
    }

    @NonNull
    @Override
    public String toString() {
        return "CollaborativeTask{" +
                "taskId='" + taskId + '\'' +
                ", type='" + type + '\'' +
                ", status='" + status + '\'' +
                ", subTasks=" + (subTasks != null ? subTasks.size() : 0) +
                '}';
    }
}