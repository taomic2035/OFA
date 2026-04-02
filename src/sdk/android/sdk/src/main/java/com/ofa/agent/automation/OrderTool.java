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
 * Order Tool - 提供订单相关操作
 * 支持查询订单状态、取消订单等
 */
public class OrderTool implements ToolExecutor {

    private static final String TAG = "OrderTool";

    private final Context context;
    private final AutomationManager automationManager;

    public OrderTool(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.automationManager = AutomationManager.init(context);
    }

    @NonNull
    @Override
    public String getToolId() {
        return "order";
    }

    @NonNull
    @Override
    public String getDescription() {
        return "Order operations: get status, list orders, cancel, track delivery";
    }

    @NonNull
    @Override
    public ToolResult execute(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String operation = getStringArg(args, "operation", "status");

        switch (operation.toLowerCase()) {
            case "status":
            case "getstatus":
                return executeGetStatus(args);
            case "list":
                return executeListOrders(args);
            case "cancel":
                return executeCancel(args);
            case "track":
                return executeTrack(args);
            case "reopen":
                return executeReopen(args);
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
        return false;
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
        String operation = getStringArg(args, "operation", "status");
        switch (operation.toLowerCase()) {
            case "status":
            case "cancel":
            case "track":
                return args.containsKey("orderId");
            case "list":
            case "reopen":
                return true;
            default:
                return false;
        }
    }

    @Override
    public int getEstimatedTimeMs() {
        return 5000;
    }

    // ===== Tool Definitions =====

    @NonNull
    public static ToolDefinition getStatusDefinition() {
        JSONObject props = new JSONObject();
        try {
            props.put("operation", createStringProp("Operation: 'status'", "status"));
            props.put("orderId", createStringProp("Order ID to check"));
        } catch (Exception e) {}
        return new ToolDefinition("order.getStatus", "Get order status",
                MCPProtocol.buildObjectSchema(props, new String[]{"operation"}), true, null);
    }

    @NonNull
    public static ToolDefinition getListDefinition() {
        JSONObject props = new JSONObject();
        try {
            props.put("operation", createStringProp("Operation: 'list'", "list"));
            props.put("status", createStringProp("Filter by status: pending, preparing, delivering, completed"));
            props.put("limit", createNumberProp("Max number of orders to return"));
        } catch (Exception e) {}
        return new ToolDefinition("order.list", "List orders",
                MCPProtocol.buildObjectSchema(props, new String[]{"operation"}), true, null);
    }

    @NonNull
    public static ToolDefinition getCancelDefinition() {
        JSONObject props = new JSONObject();
        try {
            props.put("operation", createStringProp("Operation: 'cancel'", "cancel"));
            props.put("orderId", createStringProp("Order ID to cancel"));
            props.put("reason", createStringProp("Cancellation reason"));
        } catch (Exception e) {}
        return new ToolDefinition("order.cancel", "Cancel order",
                MCPProtocol.buildObjectSchema(props, new String[]{"operation", "orderId"}), true, null);
    }

    @NonNull
    public static ToolDefinition getTrackDefinition() {
        JSONObject props = new JSONObject();
        try {
            props.put("operation", createStringProp("Operation: 'track'", "track"));
            props.put("orderId", createStringProp("Order ID to track"));
        } catch (Exception e) {}
        return new ToolDefinition("order.track", "Track delivery status",
                MCPProtocol.buildObjectSchema(props, new String[]{"operation", "orderId"}), true, null);
    }

    // ===== Internal Operations =====

    @NonNull
    private ToolResult executeGetStatus(@NonNull Map<String, Object> args) {
        AutomationEngine engine = automationManager.getEngine();
        if (!engine.isAvailable()) {
            return new ToolResult(getToolId(), "Automation engine not available");
        }

        String orderId = getStringArg(args, "orderId", null);

        Log.i(TAG, "Getting status for order: " + orderId);

        // Try to find order status on screen
        // Common status texts in Chinese food delivery apps
        String[] statusPatterns = {
            "待支付", "待商家接单", "商家已接单", "正在备餐", "骑手已取餐",
            "配送中", "即将送达", "已送达", "已完成", "已取消"
        };

        String foundStatus = null;
        for (String status : statusPatterns) {
            AutomationNode node = engine.findElement(BySelector.textContains(status));
            if (node != null) {
                foundStatus = status;
                break;
            }
        }

        // Parse status to standardized format
        String standardizedStatus = standardizeStatus(foundStatus);

        JSONObject output = new JSONObject();
        try {
            output.put("success", true);
            output.put("orderId", orderId);
            output.put("status", standardizedStatus);
            output.put("rawStatus", foundStatus);
            output.put("timestamp", System.currentTimeMillis());
        } catch (Exception e) {}

        return new ToolResult(getToolId(), output, 2000);
    }

    @NonNull
    private ToolResult executeListOrders(@NonNull Map<String, Object> args) {
        AutomationEngine engine = automationManager.getEngine();
        if (!engine.isAvailable()) {
            return new ToolResult(getToolId(), "Automation engine not available");
        }

        String statusFilter = getStringArg(args, "status", null);

        // Navigate to orders page if needed
        String[] orderButtons = {"订单", "我的订单", "查看订单", "全部订单"};
        boolean navigated = false;

        for (String button : orderButtons) {
            AutomationResult result = engine.click(button);
            if (result.isSuccess()) {
                navigated = true;
                break;
            }
        }

        if (!navigated) {
            // Try swipe down to find orders section
            engine.swipe(Direction.UP, 500);
        }

        JSONObject output = new JSONObject();
        try {
            output.put("success", true);
            output.put("message", "Orders page opened");
            output.put("filter", statusFilter);
        } catch (Exception e) {}

        return new ToolResult(getToolId(), output, 3000);
    }

    @NonNull
    private ToolResult executeCancel(@NonNull Map<String, Object> args) {
        AutomationEngine engine = automationManager.getEngine();
        if (!engine.isAvailable()) {
            return new ToolResult(getToolId(), "Automation engine not available");
        }

        String orderId = getStringArg(args, "orderId", null);
        String reason = getStringArg(args, "reason", "用户取消");

        Log.i(TAG, "Cancelling order: " + orderId);

        // Find and click cancel button
        String[] cancelButtons = {"取消订单", "申请退款", "退款"};
        boolean cancelled = false;

        for (String button : cancelButtons) {
            AutomationResult result = engine.click(button);
            if (result.isSuccess()) {
                cancelled = true;
                break;
            }
        }

        if (cancelled) {
            // Wait for confirmation dialog
            AutomationResult waitResult = engine.waitForElement(
                BySelector.text("确定").or(BySelector.text("确认取消")),
                3000);

            if (waitResult.isSuccess()) {
                // Click confirm
                engine.click("确定");
            }

            // Select reason if presented
            engine.waitForElement(BySelector.textContains("原因"), 2000);
            engine.click(reason);

            // Submit cancellation
            engine.click("提交");
        }

        JSONObject output = new JSONObject();
        try {
            output.put("success", cancelled);
            output.put("orderId", orderId);
            output.put("reason", reason);
        } catch (Exception e) {}

        return new ToolResult(getToolId(), output, 3000);
    }

    @NonNull
    private ToolResult executeTrack(@NonNull Map<String, Object> args) {
        AutomationEngine engine = automationManager.getEngine();
        if (!engine.isAvailable()) {
            return new ToolResult(getToolId(), "Automation engine not available");
        }

        String orderId = getStringArg(args, "orderId", null);

        // Try to find tracking info
        String[] trackButtons = {"查看配送", "配送详情", "查看骑手", "跟踪配送"};
        boolean tracked = false;

        for (String button : trackButtons) {
            AutomationResult result = engine.click(button);
            if (result.isSuccess()) {
                tracked = true;
                break;
            }
        }

        // Extract tracking info from screen
        String distance = null;
        String eta = null;

        // Look for distance info
        AutomationNode distanceNode = engine.findElement(
            BySelector.textContains("米").or(BySelector.textContains("公里")));
        if (distanceNode != null) {
            distance = distanceNode.getText();
        }

        // Look for ETA
        AutomationNode etaNode = engine.findElement(
            BySelector.textContains("分钟").or(BySelector.textContains("预计")));
        if (etaNode != null) {
            eta = etaNode.getText();
        }

        JSONObject output = new JSONObject();
        try {
            output.put("success", tracked);
            output.put("orderId", orderId);
            output.put("tracking", tracked);
            if (distance != null) output.put("distance", distance);
            if (eta != null) output.put("eta", eta);
        } catch (Exception e) {}

        return new ToolResult(getToolId(), output, 2000);
    }

    @NonNull
    private ToolResult executeReopen(@NonNull Map<String, Object> args) {
        AutomationEngine engine = automationManager.getEngine();
        if (!engine.isAvailable()) {
            return new ToolResult(getToolId(), "Automation engine not available");
        }

        String orderId = getStringArg(args, "orderId", null);

        // Find and click "再来一单" button
        String[] reopenButtons = {"再来一单", "重新购买", "再次购买"};
        boolean reopened = false;

        for (String button : reopenButtons) {
            AutomationResult result = engine.click(button);
            if (result.isSuccess()) {
                reopened = true;
                break;
            }
        }

        JSONObject output = new JSONObject();
        try {
            output.put("success", reopened);
            output.put("orderId", orderId);
            output.put("message", reopened ? "Order reopened" : "Could not reopen order");
        } catch (Exception e) {}

        return new ToolResult(getToolId(), output, 2000);
    }

    // ===== Helper Methods =====

    private String standardizeStatus(@Nullable String rawStatus) {
        if (rawStatus == null) return "unknown";

        if (rawStatus.contains("待支付")) return "pending_payment";
        if (rawStatus.contains("待商家接单")) return "pending_accept";
        if (rawStatus.contains("已接单") || rawStatus.contains("备餐")) return "preparing";
        if (rawStatus.contains("取餐")) return "picked_up";
        if (rawStatus.contains("配送中") || rawStatus.contains("送达")) return "delivering";
        if (rawStatus.contains("已完成")) return "completed";
        if (rawStatus.contains("取消")) return "cancelled";

        return rawStatus;
    }

    @Nullable
    private String getStringArg(@NonNull Map<String, Object> args, @NonNull String key, @Nullable String defaultVal) {
        Object value = args.get(key);
        if (value == null) return defaultVal;
        return value.toString();
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
}