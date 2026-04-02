package com.ofa.agent.automation.template;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.automation.adapter.AppAdapter;
import com.ofa.agent.automation.adapter.AppAdapterManager;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * Template Registry - stores and manages operation templates.
 * Provides built-in templates for common operations.
 */
public class TemplateRegistry {

    private static final String TAG = "TemplateRegistry";

    private final Map<String, OperationTemplate> templates = new ConcurrentHashMap<>();
    private final AppAdapterManager adapterManager;

    /**
     * Create template registry
     */
    public TemplateRegistry(@Nullable AppAdapterManager adapterManager) {
        this.adapterManager = adapterManager;
        registerBuiltInTemplates();
    }

    /**
     * Register a template
     */
    public void register(@NonNull OperationTemplate template) {
        templates.put(template.getId(), template);
        Log.i(TAG, "Registered template: " + template.getName() + " (" + template.getId() + ")");
    }

    /**
     * Unregister a template
     */
    public void unregister(@NonNull String templateId) {
        OperationTemplate removed = templates.remove(templateId);
        if (removed != null) {
            Log.i(TAG, "Unregistered template: " + templateId);
        }
    }

    /**
     * Get template by ID
     */
    @Nullable
    public OperationTemplate getTemplate(@NonNull String templateId) {
        return templates.get(templateId);
    }

    /**
     * Get all templates
     */
    @NonNull
    public List<OperationTemplate> getAllTemplates() {
        return new ArrayList<>(templates.values());
    }

    /**
     * Get templates by category
     */
    @NonNull
    public List<OperationTemplate> getTemplatesByCategory(@NonNull String category) {
        List<OperationTemplate> result = new ArrayList<>();
        for (OperationTemplate template : templates.values()) {
            if (template.getCategory().equals(category)) {
                result.add(template);
            }
        }
        return result;
    }

    /**
     * Execute a template
     */
    @NonNull
    public AutomationResult execute(@NonNull AutomationEngine engine,
                                     @NonNull String templateId,
                                     @NonNull Map<String, String> params) {
        OperationTemplate template = getTemplate(templateId);
        if (template == null) {
            return new AutomationResult("template", "Template not found: " + templateId);
        }

        // Check required params
        if (!template.hasAllRequiredParams(params)) {
            List<String> missing = new ArrayList<>();
            for (String required : template.getRequiredParams()) {
                if (!params.containsKey(required) && !template.getDefaultParams().containsKey(required)) {
                    missing.add(required);
                }
            }
            return new AutomationResult("template",
                "Missing required parameters: " + missing.toString());
        }

        // Merge params
        Map<String, String> mergedParams = template.mergeParams(params);

        // Get adapter if available
        AppAdapter adapter = null;
        if (adapterManager != null) {
            adapter = adapterManager.detectAdapter(engine);
        }

        // Execute steps
        AutomationResult lastResult = null;
        for (OperationTemplate.TemplateStep step : template.getSteps()) {
            // Check condition
            if (step.getConditionParam() != null) {
                String conditionValue = mergedParams.get(step.getConditionParam());
                if (conditionValue == null || conditionValue.isEmpty()) {
                    if (step.isOptional()) {
                        continue;
                    }
                }
            }

            // Get step params, substituting from merged params
            Map<String, String> stepParams = new HashMap<>();
            for (Map.Entry<String, String> entry : step.getParams().entrySet()) {
                String value = entry.getValue();
                // Check if value is a param reference (starts with $)
                if (value.startsWith("$")) {
                    String paramName = value.substring(1);
                    String paramValue = mergedParams.get(paramName);
                    if (paramValue != null) {
                        stepParams.put(entry.getKey(), paramValue);
                    } else {
                        stepParams.put(entry.getKey(), value);
                    }
                } else {
                    stepParams.put(entry.getKey(), value);
                }
            }

            // Execute operation
            if (adapter != null && adapter.getSupportedOperations().contains(step.getOperation())) {
                // Use adapter
                lastResult = executeAdapterOperation(adapter, engine, step.getOperation(), stepParams);
            } else {
                // Use engine directly
                lastResult = executeEngineOperation(engine, step.getOperation(), stepParams);
            }

            // Check result
            if (!lastResult.isSuccess() && !step.isOptional()) {
                return lastResult;
            }

            // Wait after step
            if (step.getWaitAfterMs() > 0) {
                try {
                    Thread.sleep(step.getWaitAfterMs());
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                }
            }
        }

        return lastResult != null ? lastResult : new AutomationResult(templateId, 0);
    }

    /**
     * Execute operation via adapter
     */
    @NonNull
    private AutomationResult executeAdapterOperation(@NonNull AppAdapter adapter,
                                                      @NonNull AutomationEngine engine,
                                                      @NonNull String operation,
                                                      @NonNull Map<String, String> params) {
        switch (operation) {
            case "search":
                String query = params.get("query");
                if (query != null) {
                    return adapter.search(engine, query);
                }
                break;
            case "selectShop":
                String shopName = params.get("shopName");
                if (shopName != null) {
                    return adapter.selectShop(engine, shopName);
                }
                break;
            case "selectProduct":
                String productName = params.get("productName");
                if (productName != null) {
                    return adapter.selectProduct(engine, productName);
                }
                break;
            case "configureOptions":
                return adapter.configureOptions(engine, params);
            case "addToCart":
                int quantity = 1;
                String qtyStr = params.get("quantity");
                if (qtyStr != null) {
                    try {
                        quantity = Integer.parseInt(qtyStr);
                    } catch (NumberFormatException e) {
                        // Use default
                    }
                }
                return adapter.addToCart(engine, quantity);
            case "goToCart":
                return adapter.goToCart(engine);
            case "goToCheckout":
                return adapter.goToCheckout(engine);
            case "selectAddress":
                String address = params.get("address");
                if (address != null) {
                    return adapter.selectAddress(engine, address);
                }
                break;
            case "submitOrder":
                return adapter.submitOrder(engine);
            case "pay":
                String paymentMethod = params.get("paymentMethod");
                if (paymentMethod == null) {
                    paymentMethod = "alipay";
                }
                return adapter.pay(engine, paymentMethod);
            case "goBack":
                return adapter.goBack(engine);
            case "goToHome":
                return adapter.goToHome(engine);
        }
        return new AutomationResult(operation, "Unknown operation: " + operation);
    }

    /**
     * Execute operation directly via engine
     */
    @NonNull
    private AutomationResult executeEngineOperation(@NonNull AutomationEngine engine,
                                                     @NonNull String operation,
                                                     @NonNull Map<String, String> params) {
        // Simple engine operations
        switch (operation) {
            case "click":
                String clickTarget = params.get("target");
                if (clickTarget != null) {
                    return engine.click(clickTarget);
                }
                String clickX = params.get("x");
                String clickY = params.get("y");
                if (clickX != null && clickY != null) {
                    try {
                        int x = Integer.parseInt(clickX);
                        int y = Integer.parseInt(clickY);
                        return engine.click(x, y);
                    } catch (NumberFormatException e) {
                        return new AutomationResult("click", "Invalid coordinates");
                    }
                }
                break;
            case "swipe":
                // Handle swipe operation
                return engine.swipe(com.ofa.agent.automation.Direction.DOWN, 0);
            case "input":
                String inputText = params.get("text");
                if (inputText != null) {
                    return engine.inputText(inputText);
                }
                break;
            case "wait":
                String waitMs = params.get("duration");
                if (waitMs != null) {
                    try {
                        long duration = Long.parseLong(waitMs);
                        Thread.sleep(duration);
                        return new AutomationResult("wait", 0);
                    } catch (InterruptedException | NumberFormatException e) {
                        return new AutomationResult("wait", e.getMessage());
                    }
                }
                break;
        }
        return new AutomationResult(operation, "Unknown engine operation: " + operation);
    }

    /**
     * Register built-in templates
     */
    private void registerBuiltInTemplates() {
        // Food delivery template
        OperationTemplate foodDelivery = new OperationTemplate.Builder(
            "food_delivery",
            "点外卖",
            "Complete food delivery order flow",
            "food"
        )
            .requiredParam("query")
            .requiredParam("shopName")
            .requiredParam("productName")
            .defaultParam("quantity", "1")
            .defaultParam("paymentMethod", "alipay")
            .addStep(OperationTemplate.TemplateStep.builder("search")
                .param("query", "$query")
                .waitAfter(1500)
                .build())
            .addStep(OperationTemplate.TemplateStep.builder("selectShop")
                .param("shopName", "$shopName")
                .waitAfter(2000)
                .build())
            .addStep(OperationTemplate.TemplateStep.builder("selectProduct")
                .param("productName", "$productName")
                .waitAfter(1000)
                .build())
            .addStep(OperationTemplate.TemplateStep.builder("configureOptions")
                .condition("options")
                .optional()
                .waitAfter(500)
                .build())
            .addStep(OperationTemplate.TemplateStep.builder("addToCart")
                .param("quantity", "$quantity")
                .waitAfter(500)
                .build())
            .addStep(OperationTemplate.TemplateStep.builder("goToCheckout")
                .waitAfter(1500)
                .build())
            .addStep(OperationTemplate.TemplateStep.builder("selectAddress")
                .param("address", "$address")
                .condition("address")
                .optional()
                .waitAfter(1000)
                .build())
            .addStep(OperationTemplate.TemplateStep.builder("submitOrder")
                .waitAfter(2000)
                .build())
            .addStep(OperationTemplate.TemplateStep.builder("pay")
                .param("paymentMethod", "$paymentMethod")
                .build())
            .build();

        register(foodDelivery);

        // Shopping template
        OperationTemplate shopping = new OperationTemplate.Builder(
            "shopping",
            "购物下单",
            "Complete shopping order flow",
            "shopping"
        )
            .requiredParam("query")
            .requiredParam("productName")
            .defaultParam("quantity", "1")
            .defaultParam("paymentMethod", "alipay")
            .addStep(OperationTemplate.TemplateStep.builder("search")
                .param("query", "$query")
                .waitAfter(1500)
                .build())
            .addStep(OperationTemplate.TemplateStep.builder("selectProduct")
                .param("productName", "$productName")
                .waitAfter(1000)
                .build())
            .addStep(OperationTemplate.TemplateStep.builder("configureOptions")
                .condition("options")
                .optional()
                .waitAfter(500)
                .build())
            .addStep(OperationTemplate.TemplateStep.builder("addToCart")
                .param("quantity", "$quantity")
                .waitAfter(500)
                .build())
            .addStep(OperationTemplate.TemplateStep.builder("goToCheckout")
                .waitAfter(1500)
                .build())
            .addStep(OperationTemplate.TemplateStep.builder("selectAddress")
                .param("address", "$address")
                .condition("address")
                .optional()
                .waitAfter(1000)
                .build())
            .addStep(OperationTemplate.TemplateStep.builder("submitOrder")
                .waitAfter(2000)
                .build())
            .addStep(OperationTemplate.TemplateStep.builder("pay")
                .param("paymentMethod", "$paymentMethod")
                .build())
            .build();

        register(shopping);

        // Simple search and add template
        OperationTemplate searchAdd = new OperationTemplate.Builder(
            "search_and_add",
            "搜索并加入购物车",
            "Search for product and add to cart",
            "common"
        )
            .requiredParam("query")
            .requiredParam("productName")
            .defaultParam("quantity", "1")
            .addStep(OperationTemplate.TemplateStep.builder("search")
                .param("query", "$query")
                .waitAfter(1500)
                .build())
            .addStep(OperationTemplate.TemplateStep.builder("selectProduct")
                .param("productName", "$productName")
                .waitAfter(1000)
                .build())
            .addStep(OperationTemplate.TemplateStep.builder("addToCart")
                .param("quantity", "$quantity")
                .build())
            .build();

        register(searchAdd);

        Log.i(TAG, "Registered " + templates.size() + " built-in templates");
    }

    /**
     * Get template count
     */
    public int getTemplateCount() {
        return templates.size();
    }

    /**
     * Clear all templates
     */
    public void clear() {
        templates.clear();
        Log.i(TAG, "All templates cleared");
    }
}