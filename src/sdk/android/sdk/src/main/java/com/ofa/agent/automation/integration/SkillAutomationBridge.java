package com.ofa.agent.automation.integration;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.automation.adapter.AppAdapter;
import com.ofa.agent.automation.adapter.AppAdapterManager;
import com.ofa.agent.automation.template.OperationTemplate;
import com.ofa.agent.automation.template.TemplateRegistry;
import com.ofa.agent.skill.SkillContext;
import com.ofa.agent.skill.SkillDefinition;
import com.ofa.agent.skill.SkillStep;
import com.ofa.agent.skill.CompositeSkillExecutor;
import com.ofa.agent.memory.UserMemoryManager;

import org.json.JSONObject;

import java.util.HashMap;
import java.util.Map;

/**
 * Skill-Automation Bridge - bridges skill system with automation capabilities.
 * Enables skills to execute UI automation through adapters and templates.
 */
public class SkillAutomationBridge {

    private static final String TAG = "SkillAutomationBridge";

    private final Context context;
    private final AutomationEngine engine;
    private final AppAdapterManager adapterManager;
    private final TemplateRegistry templateRegistry;
    private final UserMemoryManager memoryManager;

    private MemoryAwareAutomation memoryAutomation;
    private SkillExecutionListener listener;

    /**
     * Skill execution listener
     */
    public interface SkillExecutionListener {
        void onStepStarted(@NonNull String stepName);
        void onStepCompleted(@NonNull String stepName, @NonNull AutomationResult result);
        void onSkillCompleted(@NonNull String skillId, boolean success);
    }

    /**
     * Create skill-automation bridge
     */
    public SkillAutomationBridge(@NonNull Context context,
                                  @NonNull AutomationEngine engine,
                                  @NonNull AppAdapterManager adapterManager,
                                  @NonNull TemplateRegistry templateRegistry,
                                  @Nullable UserMemoryManager memoryManager) {
        this.context = context;
        this.engine = engine;
        this.adapterManager = adapterManager;
        this.templateRegistry = templateRegistry;
        this.memoryManager = memoryManager;

        if (memoryManager != null) {
            this.memoryAutomation = new MemoryAwareAutomation(context, engine, memoryManager);
        }
    }

    /**
     * Set execution listener
     */
    public void setListener(@Nullable SkillExecutionListener listener) {
        this.listener = listener;
    }

    /**
     * Execute a skill with automation
     */
    @NonNull
    public AutomationResult executeSkill(@NonNull SkillDefinition skill,
                                          @NonNull Map<String, String> params) {
        Log.i(TAG, "Executing skill: " + skill.getId());

        // Detect current app and get adapter
        AppAdapter adapter = adapterManager.detectAdapter(engine);
        String currentPage = adapter != null ? adapter.detectPage(engine) : "unknown";

        Log.i(TAG, "Current app: " + (adapter != null ? adapter.getAppName() : "unknown"));
        Log.i(TAG, "Current page: " + currentPage);

        // Create skill context
        SkillContext skillContext = new SkillContext(params);
        skillContext.setVariable("current_page", currentPage);
        if (adapter != null) {
            skillContext.setVariable("app_name", adapter.getAppName());
        }

        // Execute each step
        AutomationResult lastResult = null;
        for (SkillStep step : skill.getSteps()) {
            if (listener != null) {
                listener.onStepStarted(step.getName());
            }

            lastResult = executeStep(step, adapter, skillContext);

            if (listener != null) {
                listener.onStepCompleted(step.getName(), lastResult);
            }

            if (!lastResult.isSuccess() && !step.isOptional()) {
                Log.w(TAG, "Step failed: " + step.getName() + " - " + lastResult.getMessage());
                if (listener != null) {
                    listener.onSkillCompleted(skill.getId(), false);
                }
                return lastResult;
            }

            // Update context with step result
            if (lastResult.getData() != null) {
                skillContext.setVariable("last_result", lastResult.getData().toString());
            }
        }

        // Record successful execution
        if (memoryAutomation != null && lastResult != null && lastResult.isSuccess()) {
            recordSkillExecution(skill, params);
        }

        if (listener != null) {
            listener.onSkillCompleted(skill.getId(), true);
        }

        return lastResult != null ? lastResult : new AutomationResult(skill.getId(), 0);
    }

    /**
     * Execute a single skill step
     */
    @NonNull
    private AutomationResult executeStep(@NonNull SkillStep step,
                                          @Nullable AppAdapter adapter,
                                          @NonNull SkillContext context) {
        Log.d(TAG, "Executing step: " + step.getName() + " (type: " + step.getType() + ")");

        switch (step.getType()) {
            case TOOL:
                return executeToolStep(step, adapter, context);

            case DELAY:
                return executeDelayStep(step);

            case CONDITION:
                return executeConditionStep(step, context);

            case ASSIGN:
                return executeAssignStep(step, context);

            default:
                return new AutomationResult(step.getName(), "Unsupported step type: " + step.getType());
        }
    }

    /**
     * Execute TOOL type step
     */
    @NonNull
    private AutomationResult executeToolStep(@NonNull SkillStep step,
                                              @Nullable AppAdapter adapter,
                                              @NonNull SkillContext context) {
        String toolName = step.getToolName();
        Map<String, String> params = resolveParams(step.getParams(), context);

        Log.d(TAG, "Tool step: " + toolName + " with params: " + params);

        // Try adapter first
        if (adapter != null && adapter.getSupportedOperations().contains(toolName)) {
            return executeAdapterOperation(adapter, toolName, params);
        }

        // Try template
        OperationTemplate template = templateRegistry.getTemplate(toolName);
        if (template != null) {
            return templateRegistry.execute(engine, toolName, params);
        }

        // Fallback to engine operation
        return executeEngineOperation(toolName, params);
    }

    /**
     * Execute via adapter
     */
    @NonNull
    private AutomationResult executeAdapterOperation(@NonNull AppAdapter adapter,
                                                      @NonNull String operation,
                                                      @NonNull Map<String, String> params) {
        switch (operation) {
            case "search":
                String query = params.get("query");
                return query != null ? adapter.search(engine, query) :
                    new AutomationResult(operation, "Missing query");

            case "selectShop":
                String shopName = params.get("shopName");
                return shopName != null ? adapter.selectShop(engine, shopName) :
                    new AutomationResult(operation, "Missing shopName");

            case "selectProduct":
                String productName = params.get("productName");
                return productName != null ? adapter.selectProduct(engine, productName) :
                    new AutomationResult(operation, "Missing productName");

            case "configureOptions":
                return adapter.configureOptions(engine, params);

            case "addToCart":
                int quantity = Integer.parseInt(params.getOrDefault("quantity", "1"));
                return adapter.addToCart(engine, quantity);

            case "goToCart":
                return adapter.goToCart(engine);

            case "goToCheckout":
                return adapter.goToCheckout(engine);

            case "selectAddress":
                String address = params.get("address");
                return address != null ? adapter.selectAddress(engine, address) :
                    new AutomationResult(operation, "Missing address");

            case "submitOrder":
                return adapter.submitOrder(engine);

            case "pay":
                String paymentMethod = params.getOrDefault("paymentMethod", "alipay");
                return adapter.pay(engine, paymentMethod);

            case "goBack":
                return adapter.goBack(engine);

            case "goToHome":
                return adapter.goToHome(engine);

            default:
                return new AutomationResult(operation, "Unknown operation: " + operation);
        }
    }

    /**
     * Execute via engine
     */
    @NonNull
    private AutomationResult executeEngineOperation(@NonNull String operation,
                                                     @NonNull Map<String, String> params) {
        switch (operation) {
            case "click":
                String target = params.get("target");
                if (target != null) {
                    return engine.click(target);
                }
                break;

            case "input":
                String text = params.get("text");
                if (text != null) {
                    return engine.inputText(text);
                }
                break;

            case "swipe":
                String direction = params.get("direction");
                if (direction != null) {
                    return engine.swipe(com.ofa.agent.automation.Direction.valueOf(direction), 0);
                }
                break;

            case "wait":
                long duration = Long.parseLong(params.getOrDefault("duration", "1000"));
                try {
                    Thread.sleep(duration);
                    return new AutomationResult("wait", 0);
                } catch (InterruptedException e) {
                    return new AutomationResult("wait", e.getMessage());
                }
        }

        return new AutomationResult(operation, "Unknown engine operation: " + operation);
    }

    /**
     * Execute DELAY step
     */
    @NonNull
    private AutomationResult executeDelayStep(@NonNull SkillStep step) {
        long delayMs = step.getDelayMs();
        try {
            Thread.sleep(delayMs);
            return new AutomationResult("delay", 0);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            return new AutomationResult("delay", e.getMessage());
        }
    }

    /**
     * Execute CONDITION step
     */
    @NonNull
    private AutomationResult executeConditionStep(@NonNull SkillStep step,
                                                   @NonNull SkillContext context) {
        String condition = step.getCondition();
        boolean result = evaluateCondition(condition, context);

        if (result) {
            // Execute true branch steps
            for (SkillStep subStep : step.getTrueSteps()) {
                AutomationResult r = executeStep(subStep, adapterManager.getCurrentAdapter(), context);
                if (!r.isSuccess() && !subStep.isOptional()) {
                    return r;
                }
            }
        } else if (step.getFalseSteps() != null) {
            // Execute false branch steps
            for (SkillStep subStep : step.getFalseSteps()) {
                AutomationResult r = executeStep(subStep, adapterManager.getCurrentAdapter(), context);
                if (!r.isSuccess() && !subStep.isOptional()) {
                    return r;
                }
            }
        }

        return new AutomationResult("condition", 0);
    }

    /**
     * Execute ASSIGN step
     */
    @NonNull
    private AutomationResult executeAssignStep(@NonNull SkillStep step,
                                                @NonNull SkillContext context) {
        String variable = step.getVariable();
        String value = step.getValue();

        // Resolve value (could be from context)
        String resolvedValue = resolveValue(value, context);
        context.setVariable(variable, resolvedValue);

        return new AutomationResult("assign", 0);
    }

    /**
     * Resolve parameters with context substitution
     */
    @NonNull
    private Map<String, String> resolveParams(@NonNull Map<String, String> params,
                                               @NonNull SkillContext context) {
        Map<String, String> resolved = new HashMap<>();

        for (Map.Entry<String, String> entry : params.entrySet()) {
            String value = resolveValue(entry.getValue(), context);
            resolved.put(entry.getKey(), value);
        }

        return resolved;
    }

    /**
     * Resolve a value with context substitution
     */
    @NonNull
    private String resolveValue(@NonNull String value, @NonNull SkillContext context) {
        if (value.startsWith("${") && value.endsWith("}")) {
            String varName = value.substring(2, value.length() - 1);
            Object varValue = context.getVariable(varName);
            return varValue != null ? varValue.toString() : value;
        }
        return value;
    }

    /**
     * Evaluate a condition
     */
    private boolean evaluateCondition(@NonNull String condition, @NonNull SkillContext context) {
        // Simple condition evaluation
        // Supports: ${var} == "value", ${var} != "value"
        try {
            if (condition.contains("==")) {
                String[] parts = condition.split("==");
                String left = resolveValue(parts[0].trim(), context);
                String right = parts[1].trim().replace("\"", "");
                return left.equals(right);
            } else if (condition.contains("!=")) {
                String[] parts = condition.split("!=");
                String left = resolveValue(parts[0].trim(), context);
                String right = parts[1].trim().replace("\"", "");
                return !left.equals(right);
            }
        } catch (Exception e) {
            Log.w(TAG, "Condition evaluation error: " + e.getMessage());
        }
        return false;
    }

    /**
     * Record skill execution in memory
     */
    private void recordSkillExecution(@NonNull SkillDefinition skill,
                                       @NonNull Map<String, String> params) {
        if (memoryManager == null) return;

        try {
            String key = "skill.execution." + skill.getId() + "." + System.currentTimeMillis();
            JSONObject record = new JSONObject();
            record.put("skillId", skill.getId());
            record.put("params", new JSONObject(params));
            record.put("timestamp", System.currentTimeMillis());

            memoryManager.set(key, record.toString());
            Log.d(TAG, "Recorded skill execution: " + skill.getId());
        } catch (Exception e) {
            Log.w(TAG, "Failed to record skill execution: " + e.getMessage());
        }
    }

    /**
     * Execute template by name
     */
    @NonNull
    public AutomationResult executeTemplate(@NonNull String templateId,
                                             @NonNull Map<String, String> params) {
        return templateRegistry.execute(engine, templateId, params);
    }

    /**
     * Get available templates
     */
    @NonNull
    public java.util.List<OperationTemplate> getAvailableTemplates() {
        return templateRegistry.getAllTemplates();
    }

    /**
     * Get current adapter
     */
    @Nullable
    public AppAdapter getCurrentAdapter() {
        return adapterManager.getCurrentAdapter();
    }
}