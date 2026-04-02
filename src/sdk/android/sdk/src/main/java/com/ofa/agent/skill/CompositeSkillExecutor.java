package com.ofa.agent.skill;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.intent.TaskExecutor;
import com.ofa.agent.intent.UserIntent;
import com.ofa.agent.tool.ToolExecutor;
import com.ofa.agent.tool.ToolRegistry;
import com.ofa.agent.tool.ToolResult;

import org.json.JSONObject;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.TimeoutException;

/**
 * 复合技能执行器
 * 执行多步骤技能
 */
public class CompositeSkillExecutor {

    private static final String TAG = "CompositeSkill";

    private final Context context;
    private final ToolRegistry toolRegistry;
    private final TaskExecutor taskExecutor;
    private final ExecutorService executor;
    private final SkillRegistry skillRegistry;

    public CompositeSkillExecutor(@NonNull Context context, @NonNull ToolRegistry toolRegistry) {
        this.context = context.getApplicationContext();
        this.toolRegistry = toolRegistry;
        this.taskExecutor = new TaskExecutor(context, toolRegistry);
        this.executor = Executors.newCachedThreadPool();
        this.skillRegistry = SkillRegistry.getInstance();
    }

    /**
     * 执行技能
     */
    @NonNull
    public CompletableFuture<SkillResult> execute(@NonNull SkillDefinition skill,
                                                   @Nullable Map<String, Object> inputs) {
        SkillContext ctx = new SkillContext(skill.getId(), context);
        return execute(skill, inputs, ctx);
    }

    /**
     * 执行技能（带上下文）
     */
    @NonNull
    public CompletableFuture<SkillResult> execute(@NonNull SkillDefinition skill,
                                                   @Nullable Map<String, Object> inputs,
                                                   @NonNull SkillContext ctx) {
        return CompletableFuture.supplyAsync(() -> {
            ctx.setStartTime(System.currentTimeMillis());
            ctx.setStatus(SkillContext.ExecutionStatus.RUNNING);

            // 设置输入变量
            if (inputs != null) {
                for (Map.Entry<String, Object> entry : inputs.entrySet()) {
                    ctx.setVariable(entry.getKey(), entry.getValue());
                }
            }

            // 获取起始步骤
            String currentStepId = skill.getStartStep();
            int completedSteps = 0;
            int totalSteps = skill.getSteps().size();

            try {
                while (currentStepId != null && !ctx.isCancelled()) {
                    SkillStep step = skill.getStep(currentStepId);
                    if (step == null) {
                        Log.e(TAG, "Step not found: " + currentStepId);
                        return createFailureResult(ctx, skill, "Step not found: " + currentStepId,
                                currentStepId, completedSteps, totalSteps);
                    }

                    ctx.setCurrentStepId(currentStepId);

                    // 回调
                    if (ctx.getCallback() != null) {
                        ctx.getCallback().onStepStart(currentStepId, step);
                        ctx.getCallback().onProgress(
                                (int) ((completedSteps * 100.0) / totalSteps),
                                "执行: " + step.getName());
                    }

                    // 执行步骤
                    SkillContext.StepResult result = executeStep(step, ctx);

                    // 记录结果
                    ctx.setStepResult(currentStepId, result);

                    // 回调
                    if (ctx.getCallback() != null) {
                        ctx.getCallback().onStepComplete(currentStepId, result);
                    }

                    if (!result.success && !step.isOptional()) {
                        // 步骤失败且非可选
                        if (step.getOnErrorStep() != null) {
                            currentStepId = step.getOnErrorStep();
                            continue;
                        }
                        return createFailureResult(ctx, skill, result.error, currentStepId,
                                completedSteps, totalSteps);
                    }

                    completedSteps++;

                    // 确定下一步
                    currentStepId = determineNextStep(step, result, ctx);

                    Log.d(TAG, "Step " + currentStepId + " completed, next: " + currentStepId);
                }

                // 执行完成
                ctx.setEndTime(System.currentTimeMillis());
                ctx.setStatus(SkillContext.ExecutionStatus.COMPLETED);

                Map<String, Object> outputs = extractOutputs(skill, ctx);
                SkillResult finalResult = SkillResult.success(
                        ctx.getExecutionId(), skill.getId(), outputs,
                        ctx.getDuration(), completedSteps, totalSteps);

                if (ctx.getCallback() != null) {
                    ctx.getCallback().onComplete(finalResult);
                }

                return finalResult;

            } catch (Exception e) {
                Log.e(TAG, "Skill execution failed", e);
                ctx.setStatus(SkillContext.ExecutionStatus.FAILED);
                return createFailureResult(ctx, skill, e.getMessage(), currentStepId,
                        completedSteps, totalSteps);
            }
        }, executor);
    }

    /**
     * 执行单个步骤
     */
    @NonNull
    private SkillContext.StepResult executeStep(@NonNull SkillStep step, @NonNull SkillContext ctx) {
        long startTime = System.currentTimeMillis();
        int retryCount = 0;
        int maxRetries = step.getRetryCount();

        while (retryCount <= maxRetries) {
            try {
                Object output = null;

                switch (step.getType()) {
                    case TOOL:
                        output = executeToolStep(step, ctx);
                        break;
                    case INTENT:
                        output = executeIntentStep(step, ctx);
                        break;
                    case DELAY:
                        output = executeDelayStep(step, ctx);
                        break;
                    case WAIT_FOR:
                        output = executeWaitForStep(step, ctx);
                        break;
                    case CONDITION:
                        output = executeConditionStep(step, ctx);
                        break;
                    case ASSIGN:
                        output = executeAssignStep(step, ctx);
                        break;
                    case INPUT:
                        output = executeInputStep(step, ctx);
                        break;
                    case CONFIRM:
                        output = executeConfirmStep(step, ctx);
                        break;
                    case NOTIFY:
                        output = executeNotifyStep(step, ctx);
                        break;
                    case SUB_SKILL:
                        output = executeSubSkillStep(step, ctx);
                        break;
                    case LOOP:
                        // 循环由 determineNextStep 处理
                        output = true;
                        break;
                    case PARALLEL:
                        output = executeParallelStep(step, ctx);
                        break;
                }

                return new SkillContext.StepResult(step.getId(), true, output, null,
                        System.currentTimeMillis() - startTime);

            } catch (Exception e) {
                retryCount++;
                if (retryCount <= maxRetries) {
                    Log.w(TAG, "Step " + step.getId() + " failed, retrying (" + retryCount + "/" + maxRetries + ")");
                    try {
                        Thread.sleep(step.getRetryDelay());
                    } catch (InterruptedException ie) {
                        Thread.currentThread().interrupt();
                        break;
                    }
                } else {
                    return new SkillContext.StepResult(step.getId(), false, null, e.getMessage(),
                            System.currentTimeMillis() - startTime);
                }
            }
        }

        return new SkillContext.StepResult(step.getId(), false, null, "Max retries exceeded",
                System.currentTimeMillis() - startTime);
    }

    // ===== 步骤执行方法 =====

    @Nullable
    private Object executeToolStep(@NonNull SkillStep step, @NonNull SkillContext ctx) throws Exception {
        String toolName = step.getAction();
        if (toolName == null) throw new Exception("Tool name is required");

        ToolExecutor tool = toolRegistry.getExecutor(toolName);
        if (tool == null) throw new Exception("Tool not found: " + toolName);

        Map<String, Object> params = resolveParams(step.getParams(), ctx);

        ToolResult result = tool.execute(params, ctx.getToolContext());
        if (!result.isSuccess()) {
            throw new Exception(result.getError());
        }

        return result.getOutput();
    }

    @Nullable
    private Object executeIntentStep(@NonNull SkillStep step, @NonNull SkillContext ctx) throws Exception {
        String intentInput = resolveTemplate(step.getAction(), ctx);
        if (intentInput == null) throw new Exception("Intent input is required");

        UserIntent intent = taskExecutor.getIntentEngine().recognizeBest(intentInput);
        if (intent == null) throw new Exception("Intent not recognized: " + intentInput);

        return taskExecutor.getMapper().map(intent);
    }

    @Nullable
    private Object executeDelayStep(@NonNull SkillStep step, @NonNull SkillContext ctx) throws Exception {
        int delayMs = step.getTimeout() > 0 ? step.getTimeout() : 1000;
        Object delayParam = step.getParams().get("duration");
        if (delayParam instanceof Number) {
            delayMs = ((Number) delayParam).intValue();
        }

        Thread.sleep(delayMs);
        return delayMs;
    }

    @Nullable
    private Object executeWaitForStep(@NonNull SkillStep step, @NonNull SkillContext ctx) throws Exception {
        String condition = step.getAction();
        int timeout = step.getTimeout() > 0 ? step.getTimeout() : 60000;
        int checkInterval = 1000;

        long startTime = System.currentTimeMillis();
        while (System.currentTimeMillis() - startTime < timeout) {
            if (evaluateCondition(condition, ctx)) {
                return true;
            }
            Thread.sleep(checkInterval);
            if (ctx.isCancelled()) return false;
        }

        throw new Exception("Wait condition timeout: " + condition);
    }

    @Nullable
    private Object executeConditionStep(@NonNull SkillStep step, @NonNull SkillContext ctx) {
        String condition = step.getAction();
        return evaluateCondition(condition, ctx);
    }

    @Nullable
    private Object executeAssignStep(@NonNull SkillStep step, @NonNull SkillContext ctx) {
        for (Map.Entry<String, Object> entry : step.getParams().entrySet()) {
            Object value = resolveValue(entry.getValue(), ctx);
            ctx.setVariable(entry.getKey(), value);
        }
        return true;
    }

    @Nullable
    private Object executeInputStep(@NonNull SkillStep step, @NonNull SkillContext ctx) throws Exception {
        Object promptObj = step.getParams().get("prompt");
        String prompt = promptObj != null ? resolveTemplate(promptObj.toString(), ctx) : null;
        if (prompt == null) prompt = step.getName();

        final CompletableFuture<String> future = new CompletableFuture<>();

        ctx.requestInput(prompt, new SkillContext.InputCallback() {
            @Override
            public void onInput(@Nullable String input) {
                future.complete(input);
            }
        });

        try {
            String input = future.get(step.getTimeout(), TimeUnit.MILLISECONDS);
            Object varNameObj = step.getParams().get("variable");
            if (varNameObj != null) {
                ctx.setVariable(varNameObj.toString(), input);
            }
            ctx.setStatus(SkillContext.ExecutionStatus.RUNNING);
            return input;
        } catch (TimeoutException e) {
            throw new Exception("Input timeout");
        }
    }

    @Nullable
    private Object executeConfirmStep(@NonNull SkillStep step, @NonNull SkillContext ctx) throws Exception {
        Object msgObj = step.getParams().get("message");
        String message = msgObj != null ? resolveTemplate(msgObj.toString(), ctx) : null;
        if (message == null) message = "确认继续？";

        final CompletableFuture<Boolean> future = new CompletableFuture<>();

        ctx.requestConfirm(message, new SkillContext.ConfirmCallback() {
            @Override
            public void onConfirm(boolean confirmed) {
                future.complete(confirmed);
            }
        });

        try {
            boolean confirmed = future.get(step.getTimeout(), TimeUnit.MILLISECONDS);
            ctx.setStatus(SkillContext.ExecutionStatus.RUNNING);
            if (!confirmed) {
                throw new Exception("User cancelled");
            }
            return confirmed;
        } catch (TimeoutException e) {
            throw new Exception("Confirm timeout");
        }
    }

    @Nullable
    private Object executeNotifyStep(@NonNull SkillStep step, @NonNull SkillContext ctx) throws Exception {
        Object titleObj = step.getParams().get("title");
        Object msgObj = step.getParams().get("message");
        String title = titleObj != null ? resolveTemplate(titleObj.toString(), ctx) : null;
        String message = msgObj != null ? resolveTemplate(msgObj.toString(), ctx) : null;

        // 使用通知工具
        Map<String, Object> params = new HashMap<>();
        params.put("operation", "send");
        if (title != null) params.put("title", title);
        if (message != null) params.put("message", message);

        ToolExecutor notifyTool = toolRegistry.getExecutor("notification.send");
        if (notifyTool != null) {
            ToolResult result = notifyTool.execute(params, ctx.getToolContext());
            return result.isSuccess();
        }

        return true;
    }

    @Nullable
    private Object executeSubSkillStep(@NonNull SkillStep step, @NonNull SkillContext ctx) throws Exception {
        String subSkillId = step.getAction();
        if (subSkillId == null) throw new Exception("Sub-skill ID is required");

        SkillDefinition subSkill = skillRegistry.getSkill(subSkillId);
        if (subSkill == null) throw new Exception("Sub-skill not found: " + subSkillId);

        Map<String, Object> subInputs = resolveParams(step.getParams(), ctx);

        CompletableFuture<SkillResult> future = execute(subSkill, subInputs, ctx);
        SkillResult result = future.get(step.getTimeout(), TimeUnit.MILLISECONDS);

        if (!result.isSuccess()) {
            throw new Exception("Sub-skill failed: " + result.getError());
        }

        return result.getOutputs();
    }

    @Nullable
    private Object executeParallelStep(@NonNull SkillStep step, @NonNull SkillContext ctx) throws Exception {
        // 并行执行多个动作
        List<String> actions = (List<String>) step.getParams().get("actions");
        if (actions == null || actions.isEmpty()) return true;

        CompletableFuture<?>[] futures = actions.stream()
                .map(action -> CompletableFuture.runAsync(() -> {
                    try {
                        SkillStep subStep = new SkillStep.Builder()
                                .id(step.getId() + "_parallel")
                                .type(SkillStep.StepType.TOOL)
                                .action(action)
                                .build();
                        executeStep(subStep, ctx);
                    } catch (Exception e) {
                        Log.e(TAG, "Parallel step failed: " + action, e);
                    }
                }, executor))
                .toArray(CompletableFuture[]::new);

        CompletableFuture.allOf(futures).get(step.getTimeout(), TimeUnit.MILLISECONDS);
        return true;
    }

    // ===== 辅助方法 =====

    @Nullable
    private String resolveTemplate(@Nullable String template, @NonNull SkillContext ctx) {
        if (template == null) return null;

        // 替换 ${var} 变量
        StringBuilder result = new StringBuilder();
        int i = 0;
        while (i < template.length()) {
            int start = template.indexOf("${", i);
            if (start == -1) {
                result.append(template.substring(i));
                break;
            }
            result.append(template.substring(i, start));
            int end = template.indexOf("}", start);
            if (end == -1) {
                result.append(template.substring(start));
                break;
            }
            String varPath = template.substring(start + 2, end);
            Object value = ctx.resolveValue("${" + varPath + "}");
            result.append(value != null ? value.toString() : "");
            i = end + 1;
        }
        return result.toString();
    }

    @NonNull
    private Map<String, Object> resolveParams(@NonNull Map<String, Object> params, @NonNull SkillContext ctx) {
        Map<String, Object> resolved = new HashMap<>();
        for (Map.Entry<String, Object> entry : params.entrySet()) {
            resolved.put(entry.getKey(), resolveValue(entry.getValue(), ctx));
        }
        return resolved;
    }

    @Nullable
    private Object resolveValue(@Nullable Object value, @NonNull SkillContext ctx) {
        if (value instanceof String) {
            return ctx.resolveValue((String) value);
        }
        return value;
    }

    private boolean evaluateCondition(@Nullable String condition, @NonNull SkillContext ctx) {
        if (condition == null) return false;

        // 简单条件求值
        Object value = ctx.resolveValue(condition);
        if (value instanceof Boolean) return (Boolean) value;
        if (value == null) return false;

        // 支持简单比较: ${var} == value
        String condStr = condition.toString();
        if (condStr.contains("==")) {
            String[] parts = condStr.split("==");
            if (parts.length == 2) {
                Object left = ctx.resolveValue(parts[0].trim());
                String right = parts[1].trim().replace("\"", "").replace("'", "");
                return left != null && left.toString().equals(right);
            }
        }
        if (condStr.contains("!=")) {
            String[] parts = condStr.split("!=");
            if (parts.length == 2) {
                Object left = ctx.resolveValue(parts[0].trim());
                String right = parts[1].trim().replace("\"", "").replace("'", "");
                return left == null || !left.toString().equals(right);
            }
        }

        return true;
    }

    @Nullable
    private String determineNextStep(@NonNull SkillStep step, @NonNull SkillContext.StepResult result,
                                      @NonNull SkillContext ctx) {
        // 检查条件分支
        for (SkillStep.Branch branch : step.getBranches()) {
            if (evaluateCondition(branch.condition, ctx)) {
                return branch.targetStep;
            }
        }

        // 如果是条件步骤，根据结果决定
        if (step.getType() == SkillStep.StepType.CONDITION && result.output instanceof Boolean) {
            // 已由分支处理
        }

        // 返回默认下一步
        return step.getNextStep();
    }

    @NonNull
    private Map<String, Object> extractOutputs(@NonNull SkillDefinition skill, @NonNull SkillContext ctx) {
        Map<String, Object> outputs = new HashMap<>();
        for (String outputName : skill.getOutputs().keySet()) {
            Object value = ctx.getVariable(outputName);
            if (value != null) {
                outputs.put(outputName, value);
            }
        }
        return outputs;
    }

    @NonNull
    private SkillResult createFailureResult(@NonNull SkillContext ctx, @NonNull SkillDefinition skill,
                                             @Nullable String error, @Nullable String failedStepId,
                                             int completedSteps, int totalSteps) {
        ctx.setEndTime(System.currentTimeMillis());
        ctx.setStatus(SkillContext.ExecutionStatus.FAILED);

        SkillResult result = SkillResult.failure(ctx.getExecutionId(), skill.getId(),
                error, failedStepId, ctx.getDuration(), completedSteps, totalSteps);

        if (ctx.getCallback() != null && failedStepId != null) {
            ctx.getCallback().onError(failedStepId, error != null ? error : "Unknown error");
        }
        if (ctx.getCallback() != null) {
            ctx.getCallback().onComplete(result);
        }

        return result;
    }

    /**
     * 关闭执行器
     */
    public void shutdown() {
        executor.shutdown();
        taskExecutor.shutdown();
    }
}