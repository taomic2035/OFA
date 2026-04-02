package com.ofa.agent.skill;

import android.content.Context;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.tool.ExecutionContext;

import org.json.JSONObject;

import java.util.HashMap;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;

/**
 * 技能执行上下文
 * 管理执行状态、变量和回调
 */
public class SkillContext {

    private final String executionId;
    private final String skillId;
    private final Context androidContext;
    private final ExecutionContext toolContext;

    // 变量存储
    private final Map<String, Object> variables;

    // 步骤结果
    private final Map<String, StepResult> stepResults;

    // 当前状态
    private String currentStepId;
    private ExecutionStatus status;
    private long startTime;
    private long endTime;

    // 回调
    private Callback callback;

    // 用户交互
    private UserInteractionHandler interactionHandler;

    /**
     * 步骤执行结果
     */
    public static class StepResult {
        public final String stepId;
        public final boolean success;
        public final Object output;
        public final String error;
        public final long executionTimeMs;

        public StepResult(@NonNull String stepId, boolean success,
                          @Nullable Object output, @Nullable String error, long executionTimeMs) {
            this.stepId = stepId;
            this.success = success;
            this.output = output;
            this.error = error;
            this.executionTimeMs = executionTimeMs;
        }
    }

    /**
     * 执行状态
     */
    public enum ExecutionStatus {
        PENDING,
        RUNNING,
        WAITING_INPUT,
        WAITING_CONFIRM,
        PAUSED,
        COMPLETED,
        FAILED,
        CANCELLED
    }

    /**
     * 回调接口
     */
    public interface Callback {
        void onStepStart(@NonNull String stepId, @NonNull SkillStep step);
        void onStepComplete(@NonNull String stepId, @NonNull StepResult result);
        void onStatusChange(@NonNull ExecutionStatus oldStatus, @NonNull ExecutionStatus newStatus);
        void onProgress(int progress, @Nullable String message);
        void onComplete(@NonNull SkillResult result);
        void onError(@NonNull String stepId, @NonNull String error);
    }

    /**
     * 用户交互处理器
     */
    public interface UserInteractionHandler {
        void requestInput(@NonNull String prompt, @NonNull InputCallback callback);
        void requestConfirm(@NonNull String message, @NonNull ConfirmCallback callback);
        void requestChoice(@NonNull String prompt, @NonNull String[] options,
                           @NonNull ChoiceCallback callback);
    }

    public interface InputCallback {
        void onInput(@Nullable String input);
    }

    public interface ConfirmCallback {
        void onConfirm(boolean confirmed);
    }

    public interface ChoiceCallback {
        void onChoice(int selectedIndex, @Nullable String selectedValue);
    }

    public SkillContext(@NonNull String skillId, @NonNull Context context) {
        this.executionId = UUID.randomUUID().toString();
        this.skillId = skillId;
        this.androidContext = context.getApplicationContext();
        this.variables = new ConcurrentHashMap<>();
        this.stepResults = new ConcurrentHashMap<>();
        this.status = ExecutionStatus.PENDING;
        this.toolContext = ExecutionContext.builder(androidContext)
                .executionId(executionId)
                .autoPopulateDeviceInfo()
                .build();
    }

    // ===== 变量操作 =====

    /**
     * 设置变量
     */
    public void setVariable(@NonNull String name, @Nullable Object value) {
        variables.put(name, value);
    }

    /**
     * 获取变量
     */
    @Nullable
    public Object getVariable(@NonNull String name) {
        return variables.get(name);
    }

    /**
     * 获取变量（带默认值）
     */
    @Nullable
    public Object getVariable(@NonNull String name, @Nullable Object defaultValue) {
        return variables.getOrDefault(name, defaultValue);
    }

    /**
     * 获取变量作为字符串
     */
    @Nullable
    public String getVariableAsString(@NonNull String name) {
        Object value = variables.get(name);
        return value != null ? value.toString() : null;
    }

    /**
     * 获取变量作为整数
     */
    public int getVariableAsInt(@NonNull String name, int defaultValue) {
        Object value = variables.get(name);
        if (value instanceof Number) {
            return ((Number) value).intValue();
        }
        return defaultValue;
    }

    /**
     * 获取变量作为布尔值
     */
    public boolean getVariableAsBoolean(@NonNull String name, boolean defaultValue) {
        Object value = variables.get(name);
        if (value instanceof Boolean) {
            return (Boolean) value;
        }
        return defaultValue;
    }

    /**
     * 解析变量引用
     * 支持 ${var} 和 ${step.output.field} 格式
     */
    @Nullable
    public Object resolveValue(@NonNull String expression) {
        if (expression.startsWith("${") && expression.endsWith("}")) {
            String varPath = expression.substring(2, expression.length() - 1);
            return resolveVariablePath(varPath);
        }
        return expression;
    }

    @Nullable
    private Object resolveVariablePath(@NonNull String path) {
        String[] parts = path.split("\\.");
        Object current = null;

        for (int i = 0; i < parts.length; i++) {
            String part = parts[i];

            if (i == 0) {
                // 第一部分可能是变量或步骤结果
                current = variables.get(part);
                if (current == null) {
                    StepResult result = stepResults.get(part);
                    if (result != null) {
                        current = result.output;
                    }
                }
            } else {
                // 后续部分是字段访问
                if (current == null) return null;

                if (current instanceof Map) {
                    current = ((Map<?, ?>) current).get(part);
                } else if (current instanceof JSONObject) {
                    current = ((JSONObject) current).opt(part);
                } else {
                    // 尝试反射
                    try {
                        java.lang.reflect.Field field = current.getClass().getDeclaredField(part);
                        field.setAccessible(true);
                        current = field.get(current);
                    } catch (Exception e) {
                        return null;
                    }
                }
            }
        }

        return current;
    }

    // ===== 步骤结果 =====

    /**
     * 记录步骤结果
     */
    public void setStepResult(@NonNull String stepId, @NonNull StepResult result) {
        stepResults.put(stepId, result);
    }

    /**
     * 获取步骤结果
     */
    @Nullable
    public StepResult getStepResult(@NonNull String stepId) {
        return stepResults.get(stepId);
    }

    /**
     * 获取步骤输出
     */
    @Nullable
    public Object getStepOutput(@NonNull String stepId) {
        StepResult result = stepResults.get(stepId);
        return result != null ? result.output : null;
    }

    // ===== 状态管理 =====

    @NonNull
    public String getExecutionId() { return executionId; }

    @NonNull
    public String getSkillId() { return skillId; }

    @NonNull
    public Context getAndroidContext() { return androidContext; }

    @NonNull
    public ExecutionContext getToolContext() { return toolContext; }

    @Nullable
    public String getCurrentStepId() { return currentStepId; }

    public void setCurrentStepId(@Nullable String stepId) { this.currentStepId = stepId; }

    @NonNull
    public ExecutionStatus getStatus() { return status; }

    public void setStatus(@NonNull ExecutionStatus status) {
        ExecutionStatus oldStatus = this.status;
        this.status = status;
        if (callback != null) {
            callback.onStatusChange(oldStatus, status);
        }
    }

    public long getStartTime() { return startTime; }

    public void setStartTime(long startTime) { this.startTime = startTime; }

    public long getEndTime() { return endTime; }

    public void setEndTime(long endTime) { this.endTime = endTime; }

    public long getDuration() {
        if (startTime == 0) return 0;
        return (endTime > 0 ? endTime : System.currentTimeMillis()) - startTime;
    }

    // ===== 回调 =====

    @Nullable
    public Callback getCallback() { return callback; }

    public void setCallback(@Nullable Callback callback) { this.callback = callback; }

    @Nullable
    public UserInteractionHandler getInteractionHandler() { return interactionHandler; }

    public void setInteractionHandler(@Nullable UserInteractionHandler handler) {
        this.interactionHandler = handler;
    }

    // ===== 用户交互 =====

    /**
     * 请求用户输入
     */
    public void requestInput(@NonNull String prompt, @NonNull InputCallback callback) {
        setStatus(ExecutionStatus.WAITING_INPUT);
        if (interactionHandler != null) {
            interactionHandler.requestInput(prompt, callback);
        } else {
            callback.onInput(null);
        }
    }

    /**
     * 请求用户确认
     */
    public void requestConfirm(@NonNull String message, @NonNull ConfirmCallback callback) {
        setStatus(ExecutionStatus.WAITING_CONFIRM);
        if (interactionHandler != null) {
            interactionHandler.requestConfirm(message, callback);
        } else {
            callback.onConfirm(true);
        }
    }

    /**
     * 请求用户选择
     */
    public void requestChoice(@NonNull String prompt, @NonNull String[] options,
                              @NonNull ChoiceCallback callback) {
        setStatus(ExecutionStatus.WAITING_INPUT);
        if (interactionHandler != null) {
            interactionHandler.requestChoice(prompt, options, callback);
        } else {
            callback.onChoice(0, options.length > 0 ? options[0] : null);
        }
    }

    /**
     * 恢复执行
     */
    public void resume() {
        if (status == ExecutionStatus.PAUSED || status == ExecutionStatus.WAITING_INPUT
                || status == ExecutionStatus.WAITING_CONFIRM) {
            setStatus(ExecutionStatus.RUNNING);
        }
    }

    /**
     * 暂停执行
     */
    public void pause() {
        if (status == ExecutionStatus.RUNNING) {
            setStatus(ExecutionStatus.PAUSED);
        }
    }

    /**
     * 取消执行
     */
    public void cancel() {
        setStatus(ExecutionStatus.CANCELLED);
    }

    /**
     * 是否已取消
     */
    public boolean isCancelled() {
        return status == ExecutionStatus.CANCELLED;
    }

    /**
     * 是否已完成
     */
    public boolean isCompleted() {
        return status == ExecutionStatus.COMPLETED || status == ExecutionStatus.FAILED
                || status == ExecutionStatus.CANCELLED;
    }
}