package com.ofa.agent.intent;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.tool.ExecutionContext;
import com.ofa.agent.tool.ToolExecutor;
import com.ofa.agent.tool.ToolRegistry;
import com.ofa.agent.tool.ToolResult;

import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * 任务执行器
 * 组合意图识别、映射和工具执行
 */
public class TaskExecutor {

    private static final String TAG = "TaskExecutor";

    private final Context context;
    private final IntentEngine intentEngine;
    private final IntentToolMapper mapper;
    private final ToolRegistry toolRegistry;
    private final ExecutorService executor;
    private final IntentRegistry intentRegistry;

    private boolean autoConfirmHighConfidence = true;
    private double highConfidenceThreshold = 0.85;

    /**
     * 任务执行结果
     */
    public static class TaskResult {
        public final String taskId;
        public final UserIntent intent;
        public final IntentToolMapper.MappingResult mapping;
        public final ToolResult toolResult;
        public final ExecutionStatus status;
        public final String message;
        public final long executionTimeMs;

        public TaskResult(@NonNull String taskId, @Nullable UserIntent intent,
                          @Nullable IntentToolMapper.MappingResult mapping,
                          @Nullable ToolResult toolResult,
                          @NonNull ExecutionStatus status, @Nullable String message,
                          long executionTimeMs) {
            this.taskId = taskId;
            this.intent = intent;
            this.mapping = mapping;
            this.toolResult = toolResult;
            this.status = status;
            this.message = message;
            this.executionTimeMs = executionTimeMs;
        }

        public boolean isSuccess() {
            return status == ExecutionStatus.COMPLETED && toolResult != null && toolResult.isSuccess();
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("taskId", taskId);
                json.put("status", status.name());
                json.put("message", message != null ? message : "");
                json.put("executionTimeMs", executionTimeMs);
                if (intent != null) {
                    json.put("intent", intent.getFullName());
                    json.put("confidence", intent.getConfidence());
                }
                if (toolResult != null) {
                    json.put("toolResult", toolResult.toJson());
                }
            } catch (Exception e) {
                Log.e(TAG, "Failed to serialize TaskResult", e);
            }
            return json;
        }
    }

    /**
     * 执行状态
     */
    public enum ExecutionStatus {
        RECOGNIZED,      // 意图已识别
        NEEDS_CONFIRM,   // 需要确认
        NEEDS_INFO,      // 需要更多信息
        EXECUTING,       // 正在执行
        COMPLETED,       // 已完成
        FAILED,          // 失败
        NO_MATCH,        // 无匹配意图
        NO_TOOL          // 无对应工具
    }

    /**
     * 执行回调
     */
    public interface Callback {
        void onStatusUpdate(@NonNull String taskId, @NonNull ExecutionStatus status, @Nullable String message);
        void onConfirmationRequired(@NonNull String taskId, @NonNull UserIntent intent, @NonNull String message);
        void onSlotMissing(@NonNull String taskId, @NonNull List<String> missingSlots);
        void onComplete(@NonNull String taskId, @NonNull TaskResult result);
    }

    public TaskExecutor(@NonNull Context context, @NonNull ToolRegistry toolRegistry) {
        this.context = context.getApplicationContext();
        this.toolRegistry = toolRegistry;
        this.intentEngine = new IntentEngine();
        this.mapper = new IntentToolMapper();
        this.intentRegistry = new IntentRegistry(intentEngine);
        this.executor = Executors.newCachedThreadPool();

        // 注册默认意图
        intentRegistry.registerDefaultIntents();
    }

    /**
     * 设置自动确认高置信度意图
     */
    public void setAutoConfirmHighConfidence(boolean enabled) {
        this.autoConfirmHighConfidence = enabled;
    }

    /**
     * 设置高置信度阈值
     */
    public void setHighConfidenceThreshold(double threshold) {
        this.highConfidenceThreshold = Math.max(0.5, Math.min(1.0, threshold));
    }

    /**
     * 处理用户输入
     */
    @NonNull
    public CompletableFuture<TaskResult> process(@NonNull String input) {
        return process(input, null, null);
    }

    /**
     * 处理用户输入（带回调）
     */
    @NonNull
    public CompletableFuture<TaskResult> process(@NonNull String input,
                                                  @Nullable Callback callback,
                                                  @Nullable ExecutionContext ctx) {
        String taskId = UUID.randomUUID().toString();
        long startTime = System.currentTimeMillis();

        return CompletableFuture.supplyAsync(() -> {
            try {
                // 1. 识别意图
                if (callback != null) {
                    callback.onStatusUpdate(taskId, ExecutionStatus.EXECUTING, "正在识别意图...");
                }

                UserIntent intent = intentEngine.recognizeBest(input);
                if (intent == null) {
                    return new TaskResult(taskId, null, null, null,
                            ExecutionStatus.NO_MATCH, "无法识别意图", System.currentTimeMillis() - startTime);
                }

                Log.i(TAG, "Recognized intent: " + intent.getFullName() + " (confidence: " + intent.getConfidence() + ")");

                // 2. 映射到工具
                IntentToolMapper.MappingResult mapping = mapper.map(intent);
                if (mapping == null) {
                    return new TaskResult(taskId, intent, null, null,
                            ExecutionStatus.NO_TOOL, "意图无对应工具: " + intent.getFullName(),
                            System.currentTimeMillis() - startTime);
                }

                // 3. 检查缺失槽位
                if (!mapping.isReady()) {
                    if (callback != null) {
                        callback.onSlotMissing(taskId, mapping.missingRequiredSlots);
                    }
                    return new TaskResult(taskId, intent, mapping, null,
                            ExecutionStatus.NEEDS_INFO, "缺少必要信息: " + mapping.missingRequiredSlots,
                            System.currentTimeMillis() - startTime);
                }

                // 4. 检查是否需要确认
                boolean needsConfirm = mapping.requiresConfirmation;
                if (needsConfirm && autoConfirmHighConfidence && intent.getConfidence() >= highConfidenceThreshold) {
                    needsConfirm = false;
                    Log.i(TAG, "Auto-confirming high confidence intent");
                }

                if (needsConfirm) {
                    if (callback != null) {
                        callback.onConfirmationRequired(taskId, intent, mapping.confirmationMessage);
                    }
                    return new TaskResult(taskId, intent, mapping, null,
                            ExecutionStatus.NEEDS_CONFIRM, mapping.confirmationMessage,
                            System.currentTimeMillis() - startTime);
                }

                // 5. 执行工具
                if (callback != null) {
                    callback.onStatusUpdate(taskId, ExecutionStatus.EXECUTING, "正在执行: " + mapping.toolName);
                }

                ToolExecutor tool = toolRegistry.getExecutor(mapping.toolName);
                if (tool == null) {
                    return new TaskResult(taskId, intent, mapping, null,
                            ExecutionStatus.NO_TOOL, "工具未注册: " + mapping.toolName,
                            System.currentTimeMillis() - startTime);
                }

                // 构建执行上下文
                ExecutionContext execCtx = ctx != null ? ctx
                        : ExecutionContext.builder(context)
                        .autoPopulateDeviceInfo()
                        .executionId(taskId)
                        .build();

                // 执行
                ToolResult toolResult = tool.execute(mapping.params, execCtx);

                TaskResult result = new TaskResult(taskId, intent, mapping, toolResult,
                        ExecutionStatus.COMPLETED,
                        toolResult.isSuccess() ? "执行成功" : toolResult.getError(),
                        System.currentTimeMillis() - startTime);

                if (callback != null) {
                    callback.onComplete(taskId, result);
                }

                return result;

            } catch (Exception e) {
                Log.e(TAG, "Task execution failed", e);
                return new TaskResult(taskId, null, null, null,
                        ExecutionStatus.FAILED, "执行错误: " + e.getMessage(),
                        System.currentTimeMillis() - startTime);
            }
        }, executor);
    }

    /**
     * 确认并执行任务（用于需要确认的意图）
     */
    @NonNull
    public CompletableFuture<TaskResult> confirmAndExecute(@NonNull String taskId,
                                                            @NonNull UserIntent intent,
                                                            @NonNull IntentToolMapper.MappingResult mapping,
                                                            @Nullable ExecutionContext ctx) {
        long startTime = System.currentTimeMillis();

        return CompletableFuture.supplyAsync(() -> {
            try {
                ToolExecutor tool = toolRegistry.getExecutor(mapping.toolName);
                if (tool == null) {
                    return new TaskResult(taskId, intent, mapping, null,
                            ExecutionStatus.NO_TOOL, "工具未注册: " + mapping.toolName,
                            System.currentTimeMillis() - startTime);
                }

                ExecutionContext execCtx = ctx != null ? ctx
                        : ExecutionContext.builder(context)
                        .autoPopulateDeviceInfo()
                        .executionId(taskId)
                        .build();

                ToolResult toolResult = tool.execute(mapping.params, execCtx);

                return new TaskResult(taskId, intent, mapping, toolResult,
                        ExecutionStatus.COMPLETED,
                        toolResult.isSuccess() ? "执行成功" : toolResult.getError(),
                        System.currentTimeMillis() - startTime);

            } catch (Exception e) {
                Log.e(TAG, "Confirmed task execution failed", e);
                return new TaskResult(taskId, intent, mapping, null,
                        ExecutionStatus.FAILED, "执行错误: " + e.getMessage(),
                        System.currentTimeMillis() - startTime);
            }
        }, executor);
    }

    /**
     * 补充槽位信息并执行
     */
    @NonNull
    public CompletableFuture<TaskResult> fillSlotsAndExecute(@NonNull String taskId,
                                                              @NonNull UserIntent intent,
                                                              @NonNull IntentToolMapper.MappingResult mapping,
                                                              @NonNull Map<String, Object> additionalSlots,
                                                              @Nullable ExecutionContext ctx) {
        // 合并槽位
        Map<String, Object> mergedParams = new HashMap<>(intent.getSlots());
        mergedParams.putAll(additionalSlots);

        // 重新构建意图
        UserIntent newIntent = new UserIntent.Builder()
                .id(intent.getId())
                .category(intent.getCategory())
                .action(intent.getAction())
                .confidence(intent.getConfidence())
                .slots(mergedParams)
                .rawInput(intent.getRawInput())
                .build();

        // 重新映射
        IntentToolMapper.MappingResult newMapping = mapper.map(newIntent);
        if (newMapping == null || !newMapping.isReady()) {
            return CompletableFuture.completedFuture(
                    new TaskResult(taskId, newIntent, newMapping, null,
                            ExecutionStatus.NEEDS_INFO,
                            "仍缺少必要信息",
                            0));
        }

        return confirmAndExecute(taskId, newIntent, newMapping, ctx);
    }

    /**
     * 获取意图引擎
     */
    @NonNull
    public IntentEngine getIntentEngine() {
        return intentEngine;
    }

    /**
     * 获取意图注册表
     */
    @NonNull
    public IntentRegistry getIntentRegistry() {
        return intentRegistry;
    }

    /**
     * 获取映射器
     */
    @NonNull
    public IntentToolMapper getMapper() {
        return mapper;
    }

    /**
     * 注册自定义意图
     */
    public void registerIntent(@NonNull IntentDefinition definition) {
        intentRegistry.registerCustom(definition);
    }

    /**
     * 注册意图-工具映射
     */
    public void registerMapping(@NonNull String intentId, @NonNull String toolName,
                                @Nullable Map<String, String> slotMapping,
                                @Nullable Map<String, Object> fixedParams,
                                boolean requiresConfirm, @Nullable String confirmMsg) {
        mapper.addMapping(intentId, toolName, slotMapping, fixedParams, requiresConfirm, confirmMsg);
    }

    /**
     * 关闭执行器
     */
    public void shutdown() {
        executor.shutdown();
    }
}