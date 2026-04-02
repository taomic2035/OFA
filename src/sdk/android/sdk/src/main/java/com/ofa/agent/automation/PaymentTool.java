package com.ofa.agent.automation;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.mcp.MCPProtocol;
import com.ofa.agent.mcp.ToolDefinition;
import com.ofa.agent.tool.ExecutionContext;
import com.ofa.agent.tool.ToolExecutor;
import com.ofa.agent.tool.ToolResult;

import org.json.JSONObject;

import java.util.Map;

/**
 * Payment Tool - 提供支付相关操作
 * 支持微信支付、支付宝等主流支付方式
 */
public class PaymentTool implements ToolExecutor {

    private static final String TAG = "PaymentTool";

    private final Context context;
    private final AutomationManager automationManager;

    public PaymentTool(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.automationManager = AutomationManager.init(context);
    }

    @NonNull
    @Override
    public String getToolId() {
        return "payment";
    }

    @NonNull
    @Override
    public String getDescription() {
        return "Payment operations: pay, check status, cancel";
    }

    @NonNull
    @Override
    public ToolResult execute(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String operation = getStringArg(args, "operation", "pay");

        switch (operation.toLowerCase()) {
            case "pay":
                return executePay(args);
            case "status":
                return executeCheckStatus(args);
            case "cancel":
                return executeCancel(args);
            default:
                return new ToolResult(getToolId(), "Unknown operation: " + operation);
        }
    }

    @Override
    public boolean isAvailable() {
        return automationManager.isAvailable();
    }

    @Override
    public boolean requiresAuth() {
        return true;
    }

    @Override
    public boolean supportsOffline() {
        return false;
    }

    @Nullable
    @Override
    public String[] getRequiredPermissions() {
        return null;
    }

    @Override
    public boolean validateArgs(@NonNull Map<String, Object> args) {
        String operation = getStringArg(args, "operation", "pay");
        switch (operation.toLowerCase()) {
            case "pay":
                return args.containsKey("amount") || args.containsKey("method");
            case "status":
            case "cancel":
                return args.containsKey("orderId");
            default:
                return false;
        }
    }

    @Override
    public int getEstimatedTimeMs() {
        return 10000;
    }

    // ===== Tool Definitions =====

    @NonNull
    public static ToolDefinition getPayDefinition() {
        JSONObject props = new JSONObject();
        try {
            props.put("operation", createStringProp("Operation: 'pay'", "pay"));
            props.put("method", createStringProp("Payment method: wechat, alipay, bank"));
            props.put("amount", createNumberProp("Payment amount"));
            props.put("orderId", createStringProp("Order ID"));
            props.put("confirm", createBoolProp("Auto confirm payment"));
        } catch (Exception e) {}
        return new ToolDefinition("payment.pay", "Execute payment",
                MCPProtocol.buildObjectSchema(props, new String[]{"operation"}), true, null);
    }

    @NonNull
    public static ToolDefinition getStatusDefinition() {
        JSONObject props = new JSONObject();
        try {
            props.put("operation", createStringProp("Operation: 'status'", "status"));
            props.put("orderId", createStringProp("Order ID to check"));
        } catch (Exception e) {}
        return new ToolDefinition("payment.status", "Check payment status",
                MCPProtocol.buildObjectSchema(props, new String[]{"operation", "orderId"}), true, null);
    }

    @NonNull
    public static ToolDefinition getCancelDefinition() {
        JSONObject props = new JSONObject();
        try {
            props.put("operation", createStringProp("Operation: 'cancel'", "cancel"));
            props.put("orderId", createStringProp("Order ID to cancel"));
            props.put("reason", createStringProp("Cancellation reason"));
        } catch (Exception e) {}
        return new ToolDefinition("payment.cancel", "Cancel payment/order",
                MCPProtocol.buildObjectSchema(props, new String[]{"operation", "orderId"}), true, null);
    }

    // ===== Internal Operations =====

    @NonNull
    private ToolResult executePay(@NonNull Map<String, Object> args) {
        AutomationEngine engine = automationManager.getEngine();
        if (!engine.isAvailable()) {
            return new ToolResult(getToolId(), "Automation engine not available");
        }

        String method = getStringArg(args, "method", "wechat");
        String orderId = getStringArg(args, "orderId", null);
        Boolean autoConfirm = getBoolArg(args, "confirm");

        Log.i(TAG, "Initiating payment via " + method + " for order: " + orderId);

        // Step 1: Find and click payment button
        AutomationResult result;

        // Common payment button texts
        String[] payButtons = {"支付", "立即支付", "确认支付", "去支付", "提交订单"};
        boolean clicked = false;

        for (String buttonText : payButtons) {
            result = engine.click(buttonText);
            if (result.isSuccess()) {
                clicked = true;
                Log.d(TAG, "Clicked payment button: " + buttonText);
                break;
            }
        }

        if (!clicked) {
            // Try using OCR to find payment button
            return new ToolResult(getToolId(), "Could not find payment button");
        }

        // Step 2: Select payment method if needed
        if (!method.equals("wechat")) {
            // Wait for payment method selection
            result = engine.waitForElement(BySelector.textContains("支付方式"), 3000);
            if (result.isSuccess()) {
                String methodText = method.equals("alipay") ? "支付宝" : "微信支付";
                result = engine.click(methodText);
                Log.d(TAG, "Selected payment method: " + methodText);
            }
        }

        // Step 3: Confirm payment if auto-confirm enabled
        if (autoConfirm != null && autoConfirm) {
            result = engine.waitForElement(BySelector.textContains("确认支付"), 5000);
            if (result.isSuccess()) {
                result = engine.click("确认支付");
                Log.d(TAG, "Auto-confirmed payment");
            }
        }

        JSONObject output = new JSONObject();
        try {
            output.put("success", true);
            output.put("method", method);
            output.put("orderId", orderId);
            output.put("status", "processing");
            output.put("message", "Payment initiated, please complete on your device");
        } catch (Exception e) {}

        return new ToolResult(getToolId(), output, 5000);
    }

    @NonNull
    private ToolResult executeCheckStatus(@NonNull Map<String, Object> args) {
        String orderId = getStringArg(args, "orderId", null);
        if (orderId == null) {
            return new ToolResult(getToolId(), "Missing orderId");
        }

        // In real implementation, this would query the order system
        // For now, return a placeholder response
        JSONObject output = new JSONObject();
        try {
            output.put("success", true);
            output.put("orderId", orderId);
            output.put("status", "pending");
            output.put("message", "Order status check - requires backend integration");
        } catch (Exception e) {}

        return new ToolResult(getToolId(), output, 1000);
    }

    @NonNull
    private ToolResult executeCancel(@NonNull Map<String, Object> args) {
        AutomationEngine engine = automationManager.getEngine();
        if (!engine.isAvailable()) {
            return new ToolResult(getToolId(), "Automation engine not available");
        }

        String orderId = getStringArg(args, "orderId", null);
        String reason = getStringArg(args, "reason", "User requested");

        // Try to find cancel button
        String[] cancelButtons = {"取消订单", "取消支付", "取消"};
        boolean cancelled = false;

        for (String buttonText : cancelButtons) {
            AutomationResult result = engine.click(buttonText);
            if (result.isSuccess()) {
                cancelled = true;
                Log.d(TAG, "Clicked cancel button: " + buttonText);
                break;
            }
        }

        // Confirm cancellation if dialog appears
        if (cancelled) {
            AutomationResult result = engine.waitForElement(
                BySelector.text("确定").or(BySelector.text("确认")), 2000);
            if (result.isSuccess()) {
                engine.click("确定");
            }
        }

        JSONObject output = new JSONObject();
        try {
            output.put("success", cancelled);
            output.put("orderId", orderId);
            output.put("reason", reason);
        } catch (Exception e) {}

        return new ToolResult(getToolId(), output, 2000);
    }

    // ===== Helper Methods =====

    @Nullable
    private String getStringArg(@NonNull Map<String, Object> args, @NonNull String key, @Nullable String defaultVal) {
        Object value = args.get(key);
        if (value == null) return defaultVal;
        return value.toString();
    }

    @Nullable
    private Boolean getBoolArg(@NonNull Map<String, Object> args, @NonNull String key) {
        Object value = args.get(key);
        if (value == null) return null;
        if (value instanceof Boolean) return (Boolean) value;
        return Boolean.parseBoolean(value.toString());
    }

    private static JSONObject createStringProp(String desc, String def) throws Exception {
        JSONObject prop = new JSONObject();
        prop.put("type", "string");
        prop.put("description", desc);
        if (def != null) prop.put("default", def);
        return prop;
    }

    private static JSONObject createNumberProp(String desc) throws Exception {
        JSONObject prop = new JSONObject();
        prop.put("type", "number");
        prop.put("description", desc);
        return prop;
    }

    private static JSONObject createBoolProp(String desc) throws Exception {
        JSONObject prop = new JSONObject();
        prop.put("type", "boolean");
        prop.put("description", desc);
        return prop;
    }
}