package com.ofa.agent.skill;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONObject;

import java.util.Map;

/**
 * 技能执行结果
 */
public class SkillResult {

    private final String executionId;
    private final String skillId;
    private final boolean success;
    private final Map<String, Object> outputs;
    private final String error;
    private final String failedStepId;
    private final long executionTimeMs;
    private final int completedSteps;
    private final int totalSteps;

    public SkillResult(@NonNull String executionId, @NonNull String skillId,
                       boolean success, @Nullable Map<String, Object> outputs,
                       @Nullable String error, @Nullable String failedStepId,
                       long executionTimeMs, int completedSteps, int totalSteps) {
        this.executionId = executionId;
        this.skillId = skillId;
        this.success = success;
        this.outputs = outputs != null ? new java.util.HashMap<>(outputs) : new java.util.HashMap<>();
        this.error = error;
        this.failedStepId = failedStepId;
        this.executionTimeMs = executionTimeMs;
        this.completedSteps = completedSteps;
        this.totalSteps = totalSteps;
    }

    @NonNull
    public String getExecutionId() { return executionId; }

    @NonNull
    public String getSkillId() { return skillId; }

    public boolean isSuccess() { return success; }

    @NonNull
    public Map<String, Object> getOutputs() { return outputs; }

    @Nullable
    public Object getOutput(@NonNull String name) { return outputs.get(name); }

    @Nullable
    public String getError() { return error; }

    @Nullable
    public String getFailedStepId() { return failedStepId; }

    public long getExecutionTimeMs() { return executionTimeMs; }

    public int getCompletedSteps() { return completedSteps; }

    public int getTotalSteps() { return totalSteps; }

    public double getProgress() {
        if (totalSteps == 0) return 0;
        return (double) completedSteps / totalSteps;
    }

    @NonNull
    public JSONObject toJson() {
        JSONObject json = new JSONObject();
        try {
            json.put("executionId", executionId);
            json.put("skillId", skillId);
            json.put("success", success);
            json.put("executionTimeMs", executionTimeMs);
            json.put("completedSteps", completedSteps);
            json.put("totalSteps", totalSteps);
            json.put("progress", getProgress());
            if (error != null) {
                json.put("error", error);
                json.put("failedStepId", failedStepId);
            }
            if (!outputs.isEmpty()) {
                json.put("outputs", new JSONObject(outputs));
            }
        } catch (Exception e) {
            // Ignore
        }
        return json;
    }

    @NonNull
    @Override
    public String toString() {
        return "SkillResult{" + skillId + ", success=" + success +
                ", steps=" + completedSteps + "/" + totalSteps +
                ", time=" + executionTimeMs + "ms}";
    }

    /**
     * 创建成功结果
     */
    @NonNull
    public static SkillResult success(@NonNull String executionId, @NonNull String skillId,
                                       @Nullable Map<String, Object> outputs,
                                       long executionTimeMs, int completedSteps, int totalSteps) {
        return new SkillResult(executionId, skillId, true, outputs, null, null,
                executionTimeMs, completedSteps, totalSteps);
    }

    /**
     * 创建失败结果
     */
    @NonNull
    public static SkillResult failure(@NonNull String executionId, @NonNull String skillId,
                                       @Nullable String error, @Nullable String failedStepId,
                                       long executionTimeMs, int completedSteps, int totalSteps) {
        return new SkillResult(executionId, skillId, false, null, error, failedStepId,
                executionTimeMs, completedSteps, totalSteps);
    }
}