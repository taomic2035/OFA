package com.ofa.agent.core;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.HashMap;
import java.util.Map;

/**
 * Task Result - represents the result of task execution.
 */
public class TaskResult {

    public final String taskId;
    public final boolean success;
    public final Map<String, Object> data;
    public final String error;
    public final long executionTimeMs;
    public final String executedBy; // "local", "center", "peer:<id>"

    private TaskResult(String taskId, boolean success, Map<String, Object> data,
                       String error, long executionTimeMs, String executedBy) {
        this.taskId = taskId;
        this.success = success;
        this.data = data != null ? data : new HashMap<>();
        this.error = error;
        this.executionTimeMs = executionTimeMs;
        this.executedBy = executedBy;
    }

    /**
     * Create success result
     */
    @NonNull
    public static TaskResult success(@NonNull String taskId, @NonNull Map<String, Object> data) {
        return new TaskResult(taskId, true, data, null, 0, "local");
    }

    /**
     * Create success result with execution time
     */
    @NonNull
    public static TaskResult success(@NonNull String taskId, @NonNull Map<String, Object> data,
                                      long executionTimeMs, @NonNull String executedBy) {
        return new TaskResult(taskId, true, data, null, executionTimeMs, executedBy);
    }

    /**
     * Create failure result
     */
    @NonNull
    public static TaskResult failure(@NonNull String taskId, @Nullable String error) {
        return new TaskResult(taskId, false, null, error, 0, "local");
    }

    /**
     * Create failure result with execution time
     */
    @NonNull
    public static TaskResult failure(@NonNull String taskId, @Nullable String error,
                                      long executionTimeMs, @NonNull String executedBy) {
        return new TaskResult(taskId, false, null, error, executionTimeMs, executedBy);
    }

    /**
     * Get data value
     */
    @Nullable
    public Object get(@NonNull String key) {
        return data.get(key);
    }

    /**
     * Get data as string
     */
    @Nullable
    public String getString(@NonNull String key) {
        Object value = data.get(key);
        return value != null ? value.toString() : null;
    }

    /**
     * Get data as int
     */
    public int getInt(@NonNull String key, int defaultValue) {
        Object value = data.get(key);
        if (value instanceof Number) {
            return ((Number) value).intValue();
        }
        return defaultValue;
    }

    /**
     * Get data as boolean
     */
    public boolean getBoolean(@NonNull String key, boolean defaultValue) {
        Object value = data.get(key);
        if (value instanceof Boolean) {
            return (Boolean) value;
        }
        return defaultValue;
    }

    @NonNull
    @Override
    public String toString() {
        if (success) {
            return String.format("TaskResult{id=%s, success=true, data=%s, time=%dms, by=%s}",
                taskId, data.keySet(), executionTimeMs, executedBy);
        } else {
            return String.format("TaskResult{id=%s, success=false, error=%s, time=%dms}",
                taskId, error, executionTimeMs);
        }
    }
}