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
 * UI Tool - provides UI automation operations.
 * Implements: ui.click, ui.longClick, ui.swipe, ui.input, ui.find, ui.wait
 */
public class UITool implements ToolExecutor {

    private static final String TAG = "UITool";

    private final AutomationManager automationManager;

    public UITool(@NonNull Context context) {
        this.automationManager = AutomationManager.init(context);
    }

    // ===== ToolExecutor Interface =====

    @NonNull
    @Override
    public String getToolId() {
        return "ui";
    }

    @NonNull
    @Override
    public String getDescription() {
        return "UI automation: click, swipe, input, find elements";
    }

    @NonNull
    @Override
    public ToolResult execute(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String operation = getStringArg(args, "operation", "");

        switch (operation.toLowerCase()) {
            case "click":
                return executeClick(args);
            case "longclick":
                return executeLongClick(args);
            case "swipe":
                return executeSwipe(args);
            case "input":
                return executeInput(args);
            case "find":
                return executeFind(args);
            case "wait":
                return executeWait(args);
            case "scrollfind":
                return executeScrollFind(args);
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
        return true;
    }

    @Nullable
    @Override
    public String[] getRequiredPermissions() {
        return null;
    }

    @Override
    public boolean validateArgs(@NonNull Map<String, Object> args) {
        String operation = getStringArg(args, "operation", null);
        if (operation == null) return false;

        switch (operation.toLowerCase()) {
            case "click":
                return args.containsKey("x") && args.containsKey("y") ||
                       args.containsKey("text") || args.containsKey("selector");
            case "longclick":
                return args.containsKey("x") && args.containsKey("y") ||
                       args.containsKey("text");
            case "swipe":
                return args.containsKey("fromX") && args.containsKey("fromY") &&
                       args.containsKey("toX") && args.containsKey("toY") ||
                       args.containsKey("direction");
            case "input":
                return args.containsKey("text");
            case "find":
                return args.containsKey("text") || args.containsKey("selector");
            case "wait":
                return args.containsKey("text") || args.containsKey("selector");
            case "scrollfind":
                return args.containsKey("text") || args.containsKey("selector");
            default:
                return false;
        }
    }

    @Override
    public int getEstimatedTimeMs() {
        return 1000;
    }

    // ===== Tool Definitions =====

    @NonNull
    public static ToolDefinition getClickDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'click'");
            operation.put("default", "click");
            props.put("operation", operation);

            JSONObject x = new JSONObject();
            x.put("type", "integer");
            x.put("description", "X coordinate (optional if text provided)");
            props.put("x", x);

            JSONObject y = new JSONObject();
            y.put("type", "integer");
            y.put("description", "Y coordinate (optional if text provided)");
            props.put("y", y);

            JSONObject text = new JSONObject();
            text.put("type", "string");
            text.put("description", "Text to find and click");
            props.put("text", text);
        } catch (Exception e) {}

        JSONObject schema = MCPProtocol.buildObjectSchema(props, new String[]{"operation"});
        return new ToolDefinition("ui.click", "Click on UI element by coordinate or text",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getSwipeDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'swipe'");
            operation.put("default", "swipe");
            props.put("operation", operation);

            JSONObject direction = new JSONObject();
            direction.put("type", "string");
            direction.put("description", "Direction: UP, DOWN, LEFT, RIGHT");
            props.put("direction", direction);

            JSONObject distance = new JSONObject();
            distance.put("type", "number");
            distance.put("description", "Distance in pixels (optional)");
            props.put("distance", distance);

            JSONObject fromX = new JSONObject();
            fromX.put("type", "integer");
            fromX.put("description", "Start X coordinate");
            props.put("fromX", fromX);

            JSONObject fromY = new JSONObject();
            fromY.put("type", "integer");
            fromY.put("description", "Start Y coordinate");
            props.put("fromY", fromY);

            JSONObject toX = new JSONObject();
            toX.put("type", "integer");
            toX.put("description", "End X coordinate");
            props.put("toX", toX);

            JSONObject toY = new JSONObject();
            toY.put("type", "integer");
            toY.put("description", "End Y coordinate");
            props.put("toY", toY);

            JSONObject duration = new JSONObject();
            duration.put("type", "integer");
            duration.put("description", "Duration in milliseconds");
            props.put("duration", duration);
        } catch (Exception e) {}

        JSONObject schema = MCPProtocol.buildObjectSchema(props, new String[]{"operation"});
        return new ToolDefinition("ui.swipe", "Perform swipe gesture",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getInputDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'input'");
            operation.put("default", "input");
            props.put("operation", operation);

            JSONObject text = new JSONObject();
            text.put("type", "string");
            text.put("description", "Text to input");
            props.put("text", text);

            JSONObject x = new JSONObject();
            x.put("type", "integer");
            x.put("description", "X coordinate to click first (optional)");
            props.put("x", x);

            JSONObject y = new JSONObject();
            y.put("type", "integer");
            y.put("description", "Y coordinate to click first (optional)");
            props.put("y", y);
        } catch (Exception e) {}

        JSONObject schema = MCPProtocol.buildObjectSchema(props, new String[]{"operation", "text"});
        return new ToolDefinition("ui.input", "Input text into focused element",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getFindDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'find'");
            operation.put("default", "find");
            props.put("operation", operation);

            JSONObject text = new JSONObject();
            text.put("type", "string");
            text.put("description", "Text to find");
            props.put("text", text);
        } catch (Exception e) {}

        JSONObject schema = MCPProtocol.buildObjectSchema(props, new String[]{"operation"});
        return new ToolDefinition("ui.find", "Find UI element by text",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getWaitDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'wait'");
            operation.put("default", "wait");
            props.put("operation", operation);

            JSONObject text = new JSONObject();
            text.put("type", "string");
            text.put("description", "Text to wait for");
            props.put("text", text);

            JSONObject timeout = new JSONObject();
            timeout.put("type", "integer");
            timeout.put("description", "Timeout in milliseconds");
            timeout.put("default", 30000);
            props.put("timeout", timeout);
        } catch (Exception e) {}

        JSONObject schema = MCPProtocol.buildObjectSchema(props, new String[]{"operation"});
        return new ToolDefinition("ui.wait", "Wait for UI element to appear",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getScrollFindDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'scrollFind'");
            operation.put("default", "scrollFind");
            props.put("operation", operation);

            JSONObject text = new JSONObject();
            text.put("type", "string");
            text.put("description", "Text to find");
            props.put("text", text);

            JSONObject maxScrolls = new JSONObject();
            maxScrolls.put("type", "integer");
            maxScrolls.put("description", "Maximum number of scrolls");
            maxScrolls.put("default", 10);
            props.put("maxScrolls", maxScrolls);
        } catch (Exception e) {}

        JSONObject schema = MCPProtocol.buildObjectSchema(props, new String[]{"operation"});
        return new ToolDefinition("ui.scrollFind", "Scroll to find UI element",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getLongClickDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'longClick'");
            operation.put("default", "longClick");
            props.put("operation", operation);

            JSONObject x = new JSONObject();
            x.put("type", "integer");
            x.put("description", "X coordinate");
            props.put("x", x);

            JSONObject y = new JSONObject();
            y.put("type", "integer");
            y.put("description", "Y coordinate");
            props.put("y", y);

            JSONObject text = new JSONObject();
            text.put("type", "string");
            text.put("description", "Text to find and long click");
            props.put("text", text);
        } catch (Exception e) {}

        JSONObject schema = MCPProtocol.buildObjectSchema(props, new String[]{"operation"});
        return new ToolDefinition("ui.longClick", "Long click on UI element",
                schema, true, null);
    }

    // ===== Internal Operations =====

    @NonNull
    private ToolResult executeClick(@NonNull Map<String, Object> args) {
        AutomationEngine engine = automationManager.getEngine();

        if (!engine.isAvailable()) {
            return new ToolResult(getToolId(), "Automation engine not available");
        }

        // Check for coordinate-based click
        Integer x = getIntArg(args, "x");
        Integer y = getIntArg(args, "y");

        if (x != null && y != null) {
            AutomationResult result = engine.click(x, y);
            return convertResult(result);
        }

        // Check for text-based click
        String text = getStringArg(args, "text", null);
        if (text != null) {
            AutomationResult result = engine.click(text);
            return convertResult(result);
        }

        return new ToolResult(getToolId(), "Missing x,y coordinates or text");
    }

    @NonNull
    private ToolResult executeLongClick(@NonNull Map<String, Object> args) {
        AutomationEngine engine = automationManager.getEngine();

        if (!engine.isAvailable()) {
            return new ToolResult(getToolId(), "Automation engine not available");
        }

        Integer x = getIntArg(args, "x");
        Integer y = getIntArg(args, "y");

        if (x != null && y != null) {
            AutomationResult result = engine.longClick(x, y);
            return convertResult(result);
        }

        String text = getStringArg(args, "text", null);
        if (text != null) {
            AutomationResult result = engine.longClick(text);
            return convertResult(result);
        }

        return new ToolResult(getToolId(), "Missing x,y coordinates or text");
    }

    @NonNull
    private ToolResult executeSwipe(@NonNull Map<String, Object> args) {
        AutomationEngine engine = automationManager.getEngine();

        if (!engine.isAvailable()) {
            return new ToolResult(getToolId(), "Automation engine not available");
        }

        // Check for direction-based swipe
        String directionStr = getStringArg(args, "direction", null);
        if (directionStr != null) {
            Direction direction = Direction.fromString(directionStr);
            Float distance = getFloatArg(args, "distance");
            AutomationResult result = engine.swipe(direction, distance != null ? distance : 0);
            return convertResult(result);
        }

        // Check for coordinate-based swipe
        Integer fromX = getIntArg(args, "fromX");
        Integer fromY = getIntArg(args, "fromY");
        Integer toX = getIntArg(args, "toX");
        Integer toY = getIntArg(args, "toY");
        Long duration = getLongArg(args, "duration");

        if (fromX != null && fromY != null && toX != null && toY != null) {
            long dur = duration != null ? duration : AutomationConfig.DEFAULT_SWIPE_DURATION;
            AutomationResult result = engine.swipe(fromX, fromY, toX, toY, dur);
            return convertResult(result);
        }

        return new ToolResult(getToolId(), "Missing direction or coordinates");
    }

    @NonNull
    private ToolResult executeInput(@NonNull Map<String, Object> args) {
        AutomationEngine engine = automationManager.getEngine();

        if (!engine.isAvailable()) {
            return new ToolResult(getToolId(), "Automation engine not available");
        }

        String text = getStringArg(args, "text", null);
        if (text == null) {
            return new ToolResult(getToolId(), "Missing text");
        }

        Integer x = getIntArg(args, "x");
        Integer y = getIntArg(args, "y");

        AutomationResult result;
        if (x != null && y != null) {
            result = engine.inputText(x, y, text);
        } else {
            result = engine.inputText(text);
        }

        return convertResult(result);
    }

    @NonNull
    private ToolResult executeFind(@NonNull Map<String, Object> args) {
        AutomationEngine engine = automationManager.getEngine();

        if (!engine.isAvailable()) {
            return new ToolResult(getToolId(), "Automation engine not available");
        }

        String text = getStringArg(args, "text", null);
        if (text == null) {
            return new ToolResult(getToolId(), "Missing text");
        }

        BySelector selector = BySelector.text(text);
        AutomationNode node = engine.findElement(selector);

        if (node != null) {
            JSONObject output = new JSONObject();
            try {
                output.put("success", true);
                output.put("found", true);
                output.put("node", node.toJson());
            } catch (Exception e) {}
            return new ToolResult(getToolId(), output, 100);
        } else {
            return new ToolResult(getToolId(), "Element not found: " + text);
        }
    }

    @NonNull
    private ToolResult executeWait(@NonNull Map<String, Object> args) {
        AutomationEngine engine = automationManager.getEngine();

        if (!engine.isAvailable()) {
            return new ToolResult(getToolId(), "Automation engine not available");
        }

        String text = getStringArg(args, "text", null);
        if (text == null) {
            return new ToolResult(getToolId(), "Missing text");
        }

        Long timeout = getLongArg(args, "timeout");
        long timeoutMs = timeout != null ? timeout : AutomationConfig.DEFAULT_WAIT_TIMEOUT;

        BySelector selector = BySelector.text(text);
        AutomationResult result = engine.waitForElement(selector, timeoutMs);

        return convertResult(result);
    }

    @NonNull
    private ToolResult executeScrollFind(@NonNull Map<String, Object> args) {
        AutomationEngine engine = automationManager.getEngine();

        if (!engine.isAvailable()) {
            return new ToolResult(getToolId(), "Automation engine not available");
        }

        String text = getStringArg(args, "text", null);
        if (text == null) {
            return new ToolResult(getToolId(), "Missing text");
        }

        Integer maxScrolls = getIntArg(args, "maxScrolls");
        int scrolls = maxScrolls != null ? maxScrolls : 10;

        BySelector selector = BySelector.text(text);
        AutomationResult result = engine.scrollFind(selector, scrolls);

        return convertResult(result);
    }

    // ===== Helper Methods =====

    @NonNull
    private ToolResult convertResult(@NonNull AutomationResult autoResult) {
        if (autoResult.isSuccess()) {
            JSONObject output = new JSONObject();
            try {
                output.put("success", true);
                output.put("operation", autoResult.getOperation());
                output.put("executionTimeMs", autoResult.getExecutionTimeMs());

                if (autoResult.getData() != null) {
                    output.put("data", autoResult.getData());
                }
                if (autoResult.getFoundNode() != null) {
                    output.put("node", autoResult.getFoundNode().toJson());
                }
            } catch (Exception e) {}

            return new ToolResult(getToolId(), output, autoResult.getExecutionTimeMs());
        } else {
            return new ToolResult(getToolId(), autoResult.getError(), autoResult.getExecutionTimeMs());
        }
    }

    @Nullable
    private String getStringArg(@NonNull Map<String, Object> args, @NonNull String key,
                                @Nullable String defaultVal) {
        Object value = args.get(key);
        if (value == null) return defaultVal;
        return value.toString();
    }

    @Nullable
    private Integer getIntArg(@NonNull Map<String, Object> args, @NonNull String key) {
        Object value = args.get(key);
        if (value == null) return null;
        if (value instanceof Integer) return (Integer) value;
        if (value instanceof Number) return ((Number) value).intValue();
        try {
            return Integer.parseInt(value.toString());
        } catch (Exception e) {
            return null;
        }
    }

    @Nullable
    private Long getLongArg(@NonNull Map<String, Object> args, @NonNull String key) {
        Object value = args.get(key);
        if (value == null) return null;
        if (value instanceof Long) return (Long) value;
        if (value instanceof Number) return ((Number) value).longValue();
        try {
            return Long.parseLong(value.toString());
        } catch (Exception e) {
            return null;
        }
    }

    @Nullable
    private Float getFloatArg(@NonNull Map<String, Object> args, @NonNull String key) {
        Object value = args.get(key);
        if (value == null) return null;
        if (value instanceof Float) return (Float) value;
        if (value instanceof Number) return ((Number) value).floatValue();
        try {
            return Float.parseFloat(value.toString());
        } catch (Exception e) {
            return null;
        }
    }
}