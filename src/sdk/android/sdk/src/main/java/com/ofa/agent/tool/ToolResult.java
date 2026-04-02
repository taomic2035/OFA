package com.ofa.agent.tool;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONObject;

/**
 * Tool execution result - output from a tool execution.
 */
public class ToolResult {

    private final boolean success;
    private final String toolName;
    private final JSONObject output;
    private final String error;
    private final long executionTimeMs;
    private final boolean cached;

    /**
     * Success constructor
     */
    public ToolResult(@NonNull String toolName, @NonNull JSONObject output,
                      long executionTimeMs, boolean cached) {
        this.success = true;
        this.toolName = toolName;
        this.output = output;
        this.error = null;
        this.executionTimeMs = executionTimeMs;
        this.cached = cached;
    }

    /**
     * Success constructor without cache flag
     */
    public ToolResult(@NonNull String toolName, @NonNull JSONObject output,
                      long executionTimeMs) {
        this(toolName, output, executionTimeMs, false);
    }

    /**
     * Error constructor
     */
    public ToolResult(@NonNull String toolName, @NonNull String error,
                      long executionTimeMs) {
        this.success = false;
        this.toolName = toolName;
        this.output = null;
        this.error = error;
        this.executionTimeMs = executionTimeMs;
        this.cached = false;
    }

    /**
     * Simple error constructor
     */
    public ToolResult(@NonNull String toolName, @NonNull String error) {
        this(toolName, error, 0);
    }

    @NonNull
    public String getToolName() {
        return toolName;
    }

    public boolean isSuccess() {
        return success;
    }

    @Nullable
    public JSONObject getOutput() {
        return output;
    }

    @Nullable
    public String getError() {
        return error;
    }

    public long getExecutionTimeMs() {
        return executionTimeMs;
    }

    public boolean isCached() {
        return cached;
    }

    /**
     * Get output as string
     */
    @Nullable
    public String getOutputAsString() {
        return output != null ? output.toString() : null;
    }

    /**
     * Get specific output field
     */
    @Nullable
    public Object getOutputField(@NonNull String key) {
        if (output == null) return null;
        return output.opt(key);
    }

    /**
     * Get output field as string
     */
    @Nullable
    public String getOutputFieldString(@NonNull String key) {
        if (output == null) return null;
        return output.optString(key, null);
    }

    /**
     * Convert to JSON for MCP response
     */
    @NonNull
    public JSONObject toJson() {
        JSONObject json = new JSONObject();
        try {
            json.put("success", success);
            json.put("tool", toolName);
            if (success && output != null) {
                json.put("output", output);
            } else if (error != null) {
                json.put("error", error);
            }
            json.put("executionTimeMs", executionTimeMs);
            json.put("cached", cached);
        } catch (Exception e) {
            // Should not fail
        }
        return json;
    }

    /**
     * Create from JSON
     */
    @Nullable
    public static ToolResult fromJson(@NonNull JSONObject json) {
        try {
            String toolName = json.getString("tool");
            boolean success = json.getBoolean("success");
            long executionTimeMs = json.optLong("executionTimeMs", 0);
            boolean cached = json.optBoolean("cached", false);

            if (success) {
                JSONObject output = json.getJSONObject("output");
                return new ToolResult(toolName, output, executionTimeMs, cached);
            } else {
                String error = json.getString("error");
                return new ToolResult(toolName, error, executionTimeMs);
            }
        } catch (Exception e) {
            return null;
        }
    }

    @NonNull
    @Override
    public String toString() {
        if (success) {
            return "ToolResult{" + toolName + ": success, " + executionTimeMs + "ms}";
        } else {
            return "ToolResult{" + toolName + ": error - " + error + "}";
        }
    }
}