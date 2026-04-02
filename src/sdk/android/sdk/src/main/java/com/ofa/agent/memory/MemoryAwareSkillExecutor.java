package com.ofa.agent.memory;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.skill.CompositeSkillExecutor;
import com.ofa.agent.skill.SkillContext;
import com.ofa.agent.skill.SkillDefinition;
import com.ofa.agent.skill.SkillResult;
import com.ofa.agent.skill.SkillStep;
import com.ofa.agent.tool.ToolRegistry;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.CompletableFuture;

/**
 * 记忆感知的技能执行器
 * 在执行技能时自动应用用户记忆和偏好
 */
public class MemoryAwareSkillExecutor {

    private static final String TAG = "MemoryAwareSkill";

    private final Context context;
    private final ToolRegistry toolRegistry;
    private final UserMemoryManager memoryManager;
    private final CompositeSkillExecutor baseExecutor;

    public MemoryAwareSkillExecutor(@NonNull Context context, @NonNull ToolRegistry toolRegistry) {
        this.context = context.getApplicationContext();
        this.toolRegistry = toolRegistry;
        this.memoryManager = UserMemoryManager.getInstance(context);
        this.baseExecutor = new CompositeSkillExecutor(context, toolRegistry);
    }

    /**
     * 执行技能（带记忆增强）
     */
    @NonNull
    public CompletableFuture<SkillResult> execute(@NonNull SkillDefinition skill,
                                                   @Nullable Map<String, Object> inputs) {
        return execute(skill, inputs, null);
    }

    /**
     * 执行技能（带记忆增强和上下文）
     */
    @NonNull
    public CompletableFuture<SkillResult> execute(@NonNull SkillDefinition skill,
                                                   @Nullable Map<String, Object> inputs,
                                                   @Nullable SkillContext ctx) {
        String sessionId = UUID.randomUUID().toString();
        ContextMemory contextMemory = new ContextMemory(sessionId, memoryManager);

        return executeWithMemory(skill, inputs, ctx, contextMemory);
    }

    /**
     * 带记忆执行
     */
    @NonNull
    private CompletableFuture<SkillResult> executeWithMemory(@NonNull SkillDefinition skill,
                                                              @Nullable Map<String, Object> inputs,
                                                              @Nullable SkillContext ctx,
                                                              @NonNull ContextMemory contextMemory) {
        // 设置当前技能上下文
        contextMemory.setCurrentSkill(skill.getId());

        // 智能填充参数
        Map<String, Object> enrichedInputs = enrichInputsWithMemory(skill, inputs, contextMemory);

        // 记录行为
        contextMemory.recordAction("skill:start:" + skill.getId(), enrichedInputs, null);

        // 创建带记忆回调的上下文
        SkillContext memoryCtx = ctx != null ? ctx : new SkillContext(skill.getId(), context);
        SkillContext.Callback originalCallback = memoryCtx.getCallback();

        memoryCtx.setCallback(new SkillContext.Callback() {
            @Override
            public void onStepStart(@NonNull String stepId, @NonNull SkillStep step) {
                contextMemory.setCurrentStep(stepId);
                if (originalCallback != null) {
                    originalCallback.onStepStart(stepId, step);
                }
            }

            @Override
            public void onStepComplete(@NonNull String stepId, @NonNull SkillContext.StepResult result) {
                // 记录步骤结果
                if (result.success && result.output != null) {
                    rememberStepOutput(skill.getId(), stepId, result.output, contextMemory);
                }

                if (originalCallback != null) {
                    originalCallback.onStepComplete(stepId, result);
                }
            }

            @Override
            public void onStatusChange(@NonNull SkillContext.ExecutionStatus oldStatus,
                                        @NonNull SkillContext.ExecutionStatus newStatus) {
                if (originalCallback != null) {
                    originalCallback.onStatusChange(oldStatus, newStatus);
                }
            }

            @Override
            public void onProgress(int progress, @Nullable String message) {
                if (originalCallback != null) {
                    originalCallback.onProgress(progress, message);
                }
            }

            @Override
            public void onComplete(@NonNull SkillResult result) {
                // 记录完成
                contextMemory.recordAction("skill:complete:" + skill.getId(),
                        enrichedInputs, result.isSuccess() ? "success" : result.getError());

                // 记住这次执行的参数（作为偏好）
                if (result.isSuccess()) {
                    rememberSuccessfulParams(skill.getId(), enrichedInputs);
                }

                if (originalCallback != null) {
                    originalCallback.onComplete(result);
                }
            }

            @Override
            public void onError(@NonNull String stepId, @NonNull String error) {
                contextMemory.recordAction("skill:error:" + stepId, null, error);

                if (originalCallback != null) {
                    originalCallback.onError(stepId, error);
                }
            }
        });

        // 设置用户交互处理器（带记忆）
        memoryCtx.setInteractionHandler(new MemoryAwareInteractionHandler(
                contextMemory, skill.getId(), memoryCtx.getInteractionHandler()));

        return baseExecutor.execute(skill, enrichedInputs, memoryCtx);
    }

    /**
     * 使用记忆增强输入参数
     */
    @NonNull
    private Map<String, Object> enrichInputsWithMemory(@NonNull SkillDefinition skill,
                                                        @Nullable Map<String, Object> inputs,
                                                        @NonNull ContextMemory contextMemory) {
        Map<String, Object> enriched = inputs != null ? new HashMap<>(inputs) : new HashMap<>();
        String skillId = skill.getId();

        // 获取技能定义的输入参数
        Map<String, Object> inputDefs = skill.getInputs();
        for (String paramName : inputDefs.keySet()) {
            String memoryKey = skillId + "." + paramName;

            // 如果用户未提供，尝试从记忆中获取
            if (!enriched.containsKey(paramName) || enriched.get(paramName) == null) {
                String recommendedValue = contextMemory.getRecommendedValue(memoryKey);
                if (recommendedValue != null) {
                    enriched.put(paramName, recommendedValue);
                    enriched.put("_memory_" + paramName, true); // 标记为从记忆填充
                    Log.d(TAG, "Filled " + paramName + " from memory: " + recommendedValue);
                }
            }
        }

        return enriched;
    }

    /**
     * 记住步骤输出
     */
    private void rememberStepOutput(@NonNull String skillId, @NonNull String stepId,
                                     @NonNull Object output, @NonNull ContextMemory contextMemory) {
        // 根据步骤类型决定记忆内容
        if (output instanceof Map) {
            @SuppressWarnings("unchecked")
            Map<String, Object> outputMap = (Map<String, Object>) output;
            for (Map.Entry<String, Object> entry : outputMap.entrySet()) {
                String key = skillId + "." + stepId + "." + entry.getKey();
                String value = String.valueOf(entry.getValue());
                contextMemory.remember(key, value, "skill_output", null);
            }
        }
    }

    /**
     * 记住成功的参数（作为用户偏好）
     */
    private void rememberSuccessfulParams(@NonNull String skillId, @NonNull Map<String, Object> params) {
        for (Map.Entry<String, Object> entry : params.entrySet()) {
            String key = entry.getKey();

            // 跳过内部标记
            if (key.startsWith("_")) continue;

            Object value = entry.getValue();
            if (value != null) {
                String memoryKey = skillId + "." + key;
                memoryManager.rememberPreference(memoryKey, String.valueOf(value),
                        "skill_params", null);
            }
        }

        Log.d(TAG, "Remembered successful params for " + skillId);
    }

    /**
     * 获取预填充的建议
     */
    @NonNull
    public Map<String, Object> getSuggestedInputs(@NonNull String skillId,
                                                   @NonNull List<String> paramNames) {
        Map<String, Object> suggestions = new HashMap<>();

        for (String param : paramNames) {
            String memoryKey = skillId + "." + param;
            UserMemoryManager.SmartDefault smartDefault = memoryManager.getSmartDefault(memoryKey);

            if (smartDefault != null) {
                Map<String, Object> suggestion = new HashMap<>();
                suggestion.put("recommended", smartDefault.recommendedValue);
                suggestion.put("lastUsed", smartDefault.lastUsedValue);
                suggestion.put("mostUsed", smartDefault.mostUsedValue);
                suggestion.put("confidence", smartDefault.confidence);
                suggestions.put(param, suggestion);
            }
        }

        return suggestions;
    }

    /**
     * 排序选项（根据用户偏好）
     */
    @NonNull
    public <T> List<T> sortOptionsByPreference(@NonNull String key, @NonNull List<T> options) {
        return new ContextMemory("temp", memoryManager).sortOptions(key, options);
    }

    /**
     * 获取基础执行器
     */
    @NonNull
    public CompositeSkillExecutor getBaseExecutor() {
        return baseExecutor;
    }

    /**
     * 获取记忆管理器
     */
    @NonNull
    public UserMemoryManager getMemoryManager() {
        return memoryManager;
    }

    /**
     * 关闭
     */
    public void shutdown() {
        baseExecutor.shutdown();
    }

    /**
     * 记忆感知的用户交互处理器
     */
    private static class MemoryAwareInteractionHandler implements SkillContext.UserInteractionHandler {

        private final ContextMemory contextMemory;
        private final String skillId;
        private final SkillContext.UserInteractionHandler delegate;

        MemoryAwareInteractionHandler(@NonNull ContextMemory contextMemory,
                                        @NonNull String skillId,
                                        @Nullable SkillContext.UserInteractionHandler delegate) {
            this.contextMemory = contextMemory;
            this.skillId = skillId;
            this.delegate = delegate;
        }

        @Override
        public void requestInput(@NonNull String prompt, @NonNull SkillContext.InputCallback callback) {
            // 记录交互
            contextMemory.recordAction("interaction:input", Map.of("prompt", prompt), null);

            if (delegate != null) {
                delegate.requestInput(prompt, callback);
            } else {
                callback.onInput(null);
            }
        }

        @Override
        public void requestConfirm(@NonNull String message, @NonNull SkillContext.ConfirmCallback callback) {
            contextMemory.recordAction("interaction:confirm", Map.of("message", message), null);

            if (delegate != null) {
                delegate.requestConfirm(message, callback);
            } else {
                callback.onConfirm(true);
            }
        }

        @Override
        public void requestChoice(@NonNull String prompt, @NonNull String[] options,
                                   @NonNull SkillContext.ChoiceCallback callback) {
            contextMemory.recordAction("interaction:choice",
                    Map.of("prompt", prompt, "options", options), null);

            // 尝试从记忆中找到用户偏好的选项
            String memoryKey = skillId + ".choice";
            List<String> recommended = contextMemory.getRecommendations(memoryKey, 3);

            // 如果有推荐，可以将推荐选项放在前面
            if (!recommended.isEmpty() && options.length > 1) {
                Log.d(TAG, "User has preferences for choice: " + recommended);
                // 这里可以提示用户"上次选择了X"
            }

            if (delegate != null) {
                delegate.requestChoice(prompt, options, callback);
            } else {
                callback.onChoice(0, options.length > 0 ? options[0] : null);
            }
        }
    }
}