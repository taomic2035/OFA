package com.ofa.agent.sample;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;

import com.ofa.agent.intent.IntentDefinition;
import com.ofa.agent.intent.IntentEngine;
import com.ofa.agent.intent.IntentRegistry;
import com.ofa.agent.intent.IntentToolMapper;
import com.ofa.agent.intent.TaskExecutor;
import com.ofa.agent.intent.UserIntent;
import com.ofa.agent.tool.BuiltInTools;
import com.ofa.agent.tool.ToolRegistry;

import org.json.JSONObject;

import java.util.List;
import java.util.Map;
import java.util.concurrent.CompletableFuture;

/**
 * 意图理解系统示例
 * 展示如何使用意图识别、映射和执行
 */
public class IntentSample {

    private static final String TAG = "IntentSample";

    private final Context context;
    private final TaskExecutor taskExecutor;
    private final ToolRegistry toolRegistry;

    public IntentSample(@NonNull Context context) {
        this.context = context.getApplicationContext();

        // 初始化工具注册表
        this.toolRegistry = new ToolRegistry(context);
        BuiltInTools.registerAll(context, toolRegistry);

        // 初始化任务执行器
        this.taskExecutor = new TaskExecutor(context, toolRegistry);

        // 配置自动确认高置信度意图
        taskExecutor.setAutoConfirmHighConfidence(true);
        taskExecutor.setHighConfidenceThreshold(0.85);

        Log.i(TAG, "Intent system initialized with " + toolRegistry.getCount() + " tools");
    }

    /**
     * 示例：处理用户语音输入
     */
    public void processUserInput(@NonNull String input) {
        Log.i(TAG, "Processing user input: " + input);

        taskExecutor.process(input, new TaskExecutor.Callback() {
            @Override
            public void onStatusUpdate(@NonNull String taskId,
                                       @NonNull TaskExecutor.ExecutionStatus status,
                                       String message) {
                Log.d(TAG, "[" + taskId + "] Status: " + status + " - " + message);
            }

            @Override
            public void onConfirmationRequired(@NonNull String taskId,
                                               @NonNull UserIntent intent,
                                               @NonNull String message) {
                Log.i(TAG, "[" + taskId + "] Confirmation required: " + message);
                // 在实际应用中，这里应该弹出确认对话框
                // 用户确认后调用 taskExecutor.confirmAndExecute(...)
            }

            @Override
            public void onSlotMissing(@NonNull String taskId,
                                      @NonNull List<String> missingSlots) {
                Log.w(TAG, "[" + taskId + "] Missing slots: " + missingSlots);
                // 在实际应用中，这里应该提示用户提供缺失信息
            }

            @Override
            public void onComplete(@NonNull String taskId,
                                   @NonNull TaskExecutor.TaskResult result) {
                Log.i(TAG, "[" + taskId + "] Complete: " + result.status);

                if (result.isSuccess()) {
                    JSONObject output = result.toolResult.getOutput();
                    Log.i(TAG, "Result: " + output);
                } else {
                    Log.e(TAG, "Failed: " + result.message);
                }
            }
        });
    }

    /**
     * 示例：直接意图识别（不执行）
     */
    public void recognizeIntent(@NonNull String input) {
        IntentEngine engine = taskExecutor.getIntentEngine();

        // 获取最佳匹配
        UserIntent best = engine.recognizeBest(input);
        if (best != null) {
            Log.i(TAG, "Best match: " + best.getFullName() +
                    " (confidence: " + best.getConfidence() + ")");
            Log.d(TAG, "Slots: " + best.getSlots());
        }

        // 获取所有候选
        List<UserIntent> candidates = engine.recognize(input);
        Log.d(TAG, "Found " + candidates.size() + " candidates");
        for (UserIntent intent : candidates) {
            Log.d(TAG, "  - " + intent.getFullName() + " (" + intent.getConfidence() + ")");
        }
    }

    /**
     * 示例：注册自定义意图
     */
    public void registerCustomIntent() {
        IntentDefinition customIntent = new IntentDefinition.Builder()
                .id("custom.take_selfie")
                .category("media")
                .action("take_selfie")
                .description("使用前置摄像头自拍")
                .keywords("自拍", "selfie", "自拍照片", "拍自己")
                .pattern("自拍|拍.*自己|自拍.*照片|selfie")
                .slot("mode", "string", "拍摄模式", false)
                .defaultConfidence(0.9)
                .build();

        taskExecutor.registerIntent(customIntent);

        // 映射到工具
        taskExecutor.registerMapping("custom.take_selfie", "camera.capture",
                Map.of("mode", "mode"),
                Map.of("mode", "photo", "camera", "front"),
                false, null);

        Log.i(TAG, "Registered custom intent: custom.take_selfie");
    }

    /**
     * 示例：处理确认流程
     */
    public CompletableFuture<TaskExecutor.TaskResult> handleConfirmation(
            @NonNull String taskId,
            @NonNull UserIntent intent,
            @NonNull IntentToolMapper.MappingResult mapping,
            boolean userConfirmed) {

        if (userConfirmed) {
            Log.i(TAG, "User confirmed, executing task: " + taskId);
            return taskExecutor.confirmAndExecute(taskId, intent, mapping, null);
        } else {
            Log.i(TAG, "User rejected task: " + taskId);
            return CompletableFuture.completedFuture(
                    new TaskExecutor.TaskResult(taskId, intent, mapping, null,
                            TaskExecutor.ExecutionStatus.FAILED, "User rejected",
                            0));
        }
    }

    /**
     * 示例：补充缺失槽位
     */
    public CompletableFuture<TaskExecutor.TaskResult> provideMissingInfo(
            @NonNull String taskId,
            @NonNull UserIntent intent,
            @NonNull IntentToolMapper.MappingResult mapping,
            @NonNull Map<String, Object> additionalInfo) {

        Log.i(TAG, "Providing missing info for task: " + taskId);
        return taskExecutor.fillSlotsAndExecute(taskId, intent, mapping, additionalInfo, null);
    }

    /**
     * 示例：查看已注册的意图
     */
    public void listRegisteredIntents() {
        IntentEngine engine = taskExecutor.getIntentEngine();
        List<IntentDefinition> definitions = engine.getAllDefinitions();

        Log.i(TAG, "Registered intents (" + definitions.size() + "):");
        for (IntentDefinition def : definitions) {
            Log.i(TAG, "  - " + def.getId() + ": " + def.getDescription());
            Log.d(TAG, "    Keywords: " + def.getKeywords());
        }
    }

    /**
     * 示例：查看意图-工具映射
     */
    public void listMappings() {
        IntentToolMapper mapper = taskExecutor.getMapper();
        List<String> mappedIntents = mapper.getAllMappedIntents();

        Log.i(TAG, "Intent-tool mappings (" + mapper.size() + "):");
        for (String intentId : mappedIntents) {
            IntentToolMapper.MappingRule rule = mapper.getRule(intentId);
            if (rule != null) {
                Log.i(TAG, "  - " + intentId + " -> " + rule.toolName);
            }
        }
    }

    /**
     * 演示各种用户输入的处理
     */
    public void runDemo() {
        Log.i(TAG, "=== Intent System Demo ===");

        // 系统控制
        processUserInput("打开设置");
        processUserInput("把音量调大一点");

        // 设备控制
        processUserInput("打开WiFi");
        processUserInput("关闭蓝牙");
        processUserInput("电池还剩多少");

        // 媒体操作
        processUserInput("帮我拍张照片");
        processUserInput("播放周杰伦的歌");

        // 应用操作
        processUserInput("打开微信");

        // 导航
        processUserInput("导航到北京天安门");
        processUserInput("我在哪");

        // 查看已注册内容
        listRegisteredIntents();
        listMappings();
    }

    /**
     * 关闭资源
     */
    public void shutdown() {
        taskExecutor.shutdown();
    }
}