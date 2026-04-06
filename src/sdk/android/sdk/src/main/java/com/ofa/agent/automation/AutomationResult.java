package com.ofa.agent.automation;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONObject;

/**
 * Automation operation result.
 */
public class AutomationResult {

    private final boolean success;
    private final String operation;
    private final JSONObject data;
    private final String error;
    private final long executionTimeMs;
    private final AutomationNode foundNode;

    /**
     * Success result with data
     */
    public AutomationResult(@NonNull String operation, @NonNull JSONObject data,
                            long executionTimeMs) {
        this.success = true;
        this.operation = operation;
        this.data = data;
        this.error = null;
        this.executionTimeMs = executionTimeMs;
        this.foundNode = null;
    }

    /**
     * Success result with found node
     */
    public AutomationResult(@NonNull String operation, @NonNull AutomationNode foundNode,
                            long executionTimeMs) {
        this.success = true;
        this.operation = operation;
        this.data = null;
        this.error = null;
        this.executionTimeMs = executionTimeMs;
        this.foundNode = foundNode;
    }

    /**
     * Success result (simple)
     */
    public AutomationResult(@NonNull String operation, long executionTimeMs) {
        this.success = true;
        this.operation = operation;
        this.data = null;
        this.error = null;
        this.executionTimeMs = executionTimeMs;
        this.foundNode = null;
    }

    /**
     * Error result
     */
    public AutomationResult(@NonNull String operation, @NonNull String error,
                            long executionTimeMs) {
        this.success = false;
        this.operation = operation;
        this.data = null;
        this.error = error;
        this.executionTimeMs = executionTimeMs;
        this.foundNode = null;
    }

    /**
     * Simple error result
     */
    public AutomationResult(@NonNull String operation, @NonNull String error) {
        this(operation, error, 0);
    }

    // ===== Getters =====

    public boolean isSuccess() {
        return success;
    }

    @NonNull
    public String getOperation() {
        return operation;
    }

    @Nullable
    public JSONObject getData() {
        return data;
    }

    @Nullable
    public String getError() {
        return error;
    }

    public long getExecutionTimeMs() {
        return executionTimeMs;
    }

    @Nullable
    public AutomationNode getFoundNode() {
        return foundNode;
    }

    // ===== Utility Methods =====

    /**
     * Get data field as string
     */
    @Nullable
    public String getDataString(@NonNull String key) {
        if (data == null) return null;
        return data.optString(key, null);
    }

    /**
     * Get data field as int
     */
    public int getDataInt(@NonNull String key, int defaultValue) {
        if (data == null) return defaultValue;
        return data.optInt(key, defaultValue);
    }

    /**
     * Convert to JSON
     */
    @NonNull
    public JSONObject toJson() {
        JSONObject json = new JSONObject();
        try {
            json.put("success", success);
            json.put("operation", operation);
            if (success && data != null) {
                json.put("data", data);
            }
            if (success && foundNode != null) {
                json.put("node", foundNode.toJson());
            }
            if (!success && error != null) {
                json.put("error", error);
            }
            json.put("executionTimeMs", executionTimeMs);
        } catch (Exception e) {
            // Should not fail
        }
        return json;
    }

    /**
     * Create a copy with additional data
     */
    @NonNull
    public AutomationResult setData(@NonNull String key, @NonNull Object value) {
        try {
            JSONObject newData = data != null ? new JSONObject(data.toString()) : new JSONObject();
            newData.put(key, value);
            return new AutomationResult(operation, newData, executionTimeMs);
        } catch (Exception e) {
            return this;
        }
    }

    /**
     * Get message (error or success message)
     */
    @NonNull
    public String getMessage() {
        return success ? "Success" : (error != null ? error : "Unknown error");
    }

    @NonNull
    @Override
    public String toString() {
        if (success) {
            return "AutomationResult{" + operation + ": success, " + executionTimeMs + "ms}";
        } else {
            return "AutomationResult{" + operation + ": error - " + error + "}";
        }
    }
}