package com.ofa.agent.automation;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.advanced.ActionRecorder;
import com.ofa.agent.automation.advanced.ActionReplay;
import com.ofa.agent.automation.advanced.PageMonitor;
import com.ofa.agent.automation.advanced.ScrollHelper;
import com.ofa.agent.automation.advanced.ScreenCapture;
import com.ofa.agent.automation.vision.SimpleOcrHelper;
import com.ofa.agent.mcp.MCPProtocol;
import com.ofa.agent.mcp.ToolDefinition;
import com.ofa.agent.tool.ExecutionContext;
import com.ofa.agent.tool.ToolExecutor;
import com.ofa.agent.tool.ToolResult;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.File;
import java.util.List;
import java.util.Map;

/**
 * UI Tool - provides UI automation operations.
 * Phase 1: ui.click, ui.longClick, ui.swipe, ui.input, ui.find, ui.wait, ui.scrollFind
 * Phase 2: ui.pullToRefresh, ui.scrollToTop, ui.scrollToBottom, ui.capture,
 *          ui.waitForStable, ui.startRecord, ui.stopRecord, ui.replay, ui.findText
 */
public class UITool implements ToolExecutor {

    private static final String TAG = "UITool";

    private final Context context;
    private final AutomationManager automationManager;
    private final ScrollHelper scrollHelper;
    private final PageMonitor pageMonitor;
    private ScreenCapture screenCapture;
    private ActionRecorder actionRecorder;
    private ActionReplay actionReplay;
    private SimpleOcrHelper ocrHelper;

    public UITool(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.automationManager = AutomationManager.init(context);
        this.scrollHelper = new ScrollHelper(automationManager.getEngine());
        this.pageMonitor = new PageMonitor(automationManager.getEngine());
    }

    // Lazy initialization for optional components
    private ScreenCapture getScreenCapture() {
        if (screenCapture == null) {
            screenCapture = new ScreenCapture(context);
        }
        return screenCapture;
    }

    private ActionRecorder getActionRecorder() {
        if (actionRecorder == null) {
            actionRecorder = new ActionRecorder(automationManager.getEngine(), getScreenCapture());
        }
        return actionRecorder;
    }

    private ActionReplay getActionReplay() {
        if (actionReplay == null) {
            actionReplay = new ActionReplay(automationManager.getEngine(), getScreenCapture());
        }
        return actionReplay;
    }

    private SimpleOcrHelper getOcrHelper() {
        if (ocrHelper == null) {
            ocrHelper = new SimpleOcrHelper(context);
        }
        return ocrHelper;
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
        return "UI automation: click, swipe, input, find, scroll, capture, record";
    }

    @NonNull
    @Override
    public ToolResult execute(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String operation = getStringArg(args, "operation", "");

        switch (operation.toLowerCase()) {
            // Phase 1 operations
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

            // Phase 2 operations
            case "pulltorefresh":
                return executePullToRefresh();
            case "scrolltotop":
                return executeScrollToTop();
            case "scrolltobottom":
                return executeScrollToBottom();
            case "capture":
                return executeCapture(args);
            case "waitforstable":
                return executeWaitForStable(args);
            case "startrecord":
                return executeStartRecord(args);
            case "stoprecord":
                return executeStopRecord();
            case "replay":
                return executeReplay(args);
            case "findtext":
                return executeFindText(args);

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
                       args.containsKey("text");
            case "longclick":
                return args.containsKey("x") && args.containsKey("y") ||
                       args.containsKey("text");
            case "swipe":
                return args.containsKey("direction") ||
                       (args.containsKey("fromX") && args.containsKey("fromY") &&
                        args.containsKey("toX") && args.containsKey("toY"));
            case "input":
                return args.containsKey("text");
            case "find":
            case "wait":
            case "scrollfind":
            case "findtext":
                return args.containsKey("text");
            case "capture":
            case "pulltorefresh":
            case "scrolltotop":
            case "scrolltobottom":
            case "startrecord":
            case "stoprecord":
                return true;
            case "waitforstable":
                return true;
            case "replay":
                return args.containsKey("file");
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
            props.put("operation", createStringProp("Operation: 'click'", "click"));
            props.put("x", createIntProp("X coordinate (optional if text provided)"));
            props.put("y", createIntProp("Y coordinate (optional if text provided)"));
            props.put("text", createStringProp("Text to find and click"));
        } catch (Exception e) {}
        return new ToolDefinition("ui.click", "Click on UI element",
                MCPProtocol.buildObjectSchema(props, new String[]{"operation"}), true, null);
    }

    @NonNull
    public static ToolDefinition getSwipeDefinition() {
        JSONObject props = new JSONObject();
        try {
            props.put("operation", createStringProp("Operation: 'swipe'", "swipe"));
            props.put("direction", createStringProp("Direction: UP, DOWN, LEFT, RIGHT"));
            props.put("fromX", createIntProp("Start X coordinate"));
            props.put("fromY", createIntProp("Start Y coordinate"));
            props.put("toX", createIntProp("End X coordinate"));
            props.put("toY", createIntProp("End Y coordinate"));
        } catch (Exception e) {}
        return new ToolDefinition("ui.swipe", "Perform swipe gesture",
                MCPProtocol.buildObjectSchema(props, new String[]{"operation"}), true, null);
    }

    @NonNull
    public static ToolDefinition getInputDefinition() {
        JSONObject props = new JSONObject();
        try {
            props.put("operation", createStringProp("Operation: 'input'", "input"));
            props.put("text", createStringProp("Text to input"));
            props.put("x", createIntProp("X coordinate to click first (optional)"));
            props.put("y", createIntProp("Y coordinate to click first (optional)"));
        } catch (Exception e) {}
        return new ToolDefinition("ui.input", "Input text",
                MCPProtocol.buildObjectSchema(props, new String[]{"operation", "text"}), true, null);
    }

    @NonNull
    public static ToolDefinition getFindDefinition() {
        JSONObject props = new JSONObject();
        try {
            props.put("operation", createStringProp("Operation: 'find'", "find"));
            props.put("text", createStringProp("Text to find"));
        } catch (Exception e) {}
        return new ToolDefinition("ui.find", "Find UI element",
                MCPProtocol.buildObjectSchema(props, new String[]{"operation"}), true, null);
    }

    @NonNull
    public static ToolDefinition getWaitDefinition() {
        JSONObject props = new JSONObject();
        try {
            props.put("operation", createStringProp("Operation: 'wait'", "wait"));
            props.put("text", createStringProp("Text to wait for"));
            props.put("timeout", createIntProp("Timeout in milliseconds"));
        } catch (Exception e) {}
        return new ToolDefinition("ui.wait", "Wait for UI element",
                MCPProtocol.buildObjectSchema(props, new String[]{"operation"}), true, null);
    }

    @NonNull
    public static ToolDefinition getScrollFindDefinition() {
        JSONObject props = new JSONObject();
        try {
            props.put("operation", createStringProp("Operation: 'scrollFind'", "scrollFind"));
            props.put("text", createStringProp("Text to find"));
            props.put("maxScrolls", createIntProp("Maximum number of scrolls"));
        } catch (Exception e) {}
        return new ToolDefinition("ui.scrollFind", "Scroll to find element",
                MCPProtocol.buildObjectSchema(props, new String[]{"operation"}), true, null);
    }

    @NonNull
    public static ToolDefinition getLongClickDefinition() {
        JSONObject props = new JSONObject();
        try {
            props.put("operation", createStringProp("Operation: 'longClick'", "longClick"));
            props.put("x", createIntProp("X coordinate"));
            props.put("y", createIntProp("Y coordinate"));
            props.put("text", createStringProp("Text to find and long click"));
        } catch (Exception e) {}
        return new ToolDefinition("ui.longClick", "Long click on element",
                MCPProtocol.buildObjectSchema(props, new String[]{"operation"}), true, null);
    }

    // Phase 2 Definitions

    @NonNull
    public static ToolDefinition getPullToRefreshDefinition() {
        JSONObject props = new JSONObject();
        try {
            props.put("operation", createStringProp("Operation: 'pullToRefresh'", "pullToRefresh"));
        } catch (Exception e) {}
        return new ToolDefinition("ui.pullToRefresh", "Pull to refresh (scroll down from top)",
                MCPProtocol.buildObjectSchema(props, new String[]{"operation"}), true, null);
    }

    @NonNull
    public static ToolDefinition getCaptureDefinition() {
        JSONObject props = new JSONObject();
        try {
            props.put("operation", createStringProp("Operation: 'capture'", "capture"));
            props.put("savePath", createStringProp("Path to save screenshot (optional)"));
        } catch (Exception e) {}
        return new ToolDefinition("ui.capture", "Capture screenshot",
                MCPProtocol.buildObjectSchema(props, new String[]{"operation"}), true, null);
    }

    @NonNull
    public static ToolDefinition getWaitForStableDefinition() {
        JSONObject props = new JSONObject();
        try {
            props.put("operation", createStringProp("Operation: 'waitForStable'", "waitForStable"));
            props.put("timeout", createIntProp("Timeout in milliseconds"));
        } catch (Exception e) {}
        return new ToolDefinition("ui.waitForStable", "Wait for page to stabilize",
                MCPProtocol.buildObjectSchema(props, new String[]{"operation"}), true, null);
    }

    @NonNull
    public static ToolDefinition getStartRecordDefinition() {
        JSONObject props = new JSONObject();
        try {
            props.put("operation", createStringProp("Operation: 'startRecord'", "startRecord"));
            props.put("name", createStringProp("Recording name (optional)"));
        } catch (Exception e) {}
        return new ToolDefinition("ui.startRecord", "Start recording UI actions",
                MCPProtocol.buildObjectSchema(props, new String[]{"operation"}), true, null);
    }

    @NonNull
    public static ToolDefinition getStopRecordDefinition() {
        JSONObject props = new JSONObject();
        try {
            props.put("operation", createStringProp("Operation: 'stopRecord'", "stopRecord"));
        } catch (Exception e) {}
        return new ToolDefinition("ui.stopRecord", "Stop recording UI actions",
                MCPProtocol.buildObjectSchema(props, new String[]{"operation"}), true, null);
    }

    @NonNull
    public static ToolDefinition getReplayDefinition() {
        JSONObject props = new JSONObject();
        try {
            props.put("operation", createStringProp("Operation: 'replay'", "replay"));
            props.put("file", createStringProp("Recording file path"));
            props.put("respectTiming", createBoolProp("Respect original timing"));
        } catch (Exception e) {}
        return new ToolDefinition("ui.replay", "Replay recorded actions",
                MCPProtocol.buildObjectSchema(props, new String[]{"operation", "file"}), true, null);
    }

    @NonNull
    public static ToolDefinition getFindTextDefinition() {
        JSONObject props = new JSONObject();
        try {
            props.put("operation", createStringProp("Operation: 'findText'", "findText"));
            props.put("text", createStringProp("Text to find in screen"));
        } catch (Exception e) {}
        return new ToolDefinition("ui.findText", "Find text on screen using OCR",
                MCPProtocol.buildObjectSchema(props, new String[]{"operation", "text"}), true, null);
    }

    // Helper for creating property definitions
    private static JSONObject createStringProp(String desc, String def) throws Exception {
        JSONObject prop = new JSONObject();
        prop.put("type", "string");
        prop.put("description", desc);
        if (def != null) prop.put("default", def);
        return prop;
    }

    private static JSONObject createIntProp(String desc) throws Exception {
        JSONObject prop = new JSONObject();
        prop.put("type", "integer");
        prop.put("description", desc);
        return prop;
    }

    private static JSONObject createBoolProp(String desc) throws Exception {
        JSONObject prop = new JSONObject();
        prop.put("type", "boolean");
        prop.put("description", desc);
        return prop;
    }

    // ===== Internal Operations =====

    @NonNull
    private ToolResult executeClick(@NonNull Map<String, Object> args) {
        AutomationEngine engine = automationManager.getEngine();
        if (!engine.isAvailable()) {
            return new ToolResult(getToolId(), "Automation engine not available");
        }

        Integer x = getIntArg(args, "x");
        Integer y = getIntArg(args, "y");

        if (x != null && y != null) {
            AutomationResult result = engine.click(x, y);
            return convertResult(result);
        }

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

        String directionStr = getStringArg(args, "direction", null);
        if (directionStr != null) {
            Direction direction = Direction.fromString(directionStr);
            Float distance = getFloatArg(args, "distance");
            AutomationResult result = engine.swipe(direction, distance != null ? distance : 0);
            return convertResult(result);
        }

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

    // Phase 2 Operations

    @NonNull
    private ToolResult executePullToRefresh() {
        AutomationResult result = scrollHelper.pullToRefresh();
        return convertResult(result);
    }

    @NonNull
    private ToolResult executeScrollToTop() {
        AutomationResult result = scrollHelper.scrollToTop();
        return convertResult(result);
    }

    @NonNull
    private ToolResult executeScrollToBottom() {
        AutomationResult result = scrollHelper.scrollToBottom();
        return convertResult(result);
    }

    @NonNull
    private ToolResult executeCapture(@NonNull Map<String, Object> args) {
        String savePath = getStringArg(args, "savePath", null);
        File saveDir = savePath != null ? new File(savePath) : null;

        String path = getScreenCapture().captureToFile(saveDir);
        if (path != null) {
            JSONObject output = new JSONObject();
            try {
                output.put("success", true);
                output.put("path", path);
            } catch (Exception e) {}
            return new ToolResult(getToolId(), output, 500);
        } else {
            return new ToolResult(getToolId(), "Screenshot capture failed");
        }
    }

    @NonNull
    private ToolResult executeWaitForStable(@NonNull Map<String, Object> args) {
        Long timeout = getLongArg(args, "timeout");
        long timeoutMs = timeout != null ? timeout : 5000;

        boolean stable = pageMonitor.waitForStable(timeoutMs);

        JSONObject output = new JSONObject();
        try {
            output.put("success", stable);
            output.put("stable", stable);
        } catch (Exception e) {}
        return new ToolResult(getToolId(), output, timeoutMs);
    }

    @NonNull
    private ToolResult executeStartRecord(@NonNull Map<String, Object> args) {
        String name = getStringArg(args, "name", null);
        getActionRecorder().startRecording(name);

        JSONObject output = new JSONObject();
        try {
            output.put("success", true);
            output.put("recording", true);
        } catch (Exception e) {}
        return new ToolResult(getToolId(), output, 100);
    }

    @NonNull
    private ToolResult executeStopRecord() {
        List<ActionRecorder.RecordedAction> actions = getActionRecorder().stopRecording();
        String savedPath = getActionRecorder().saveRecording();

        JSONObject output = new JSONObject();
        try {
            output.put("success", true);
            output.put("actionCount", actions.size());
            if (savedPath != null) {
                output.put("savedPath", savedPath);
            }
        } catch (Exception e) {}
        return new ToolResult(getToolId(), output, 500);
    }

    @NonNull
    private ToolResult executeReplay(@NonNull Map<String, Object> args) {
        String filePath = getStringArg(args, "file", null);
        if (filePath == null) {
            return new ToolResult(getToolId(), "Missing file path");
        }

        Boolean respectTiming = getBoolArg(args, "respectTiming");

        File file = new File(filePath);
        ActionReplay.PlaybackResult result = getActionReplay().playFromFile(file, respectTiming != null && respectTiming);

        JSONObject output = new JSONObject();
        try {
            output.put("success", result.isSuccess());
            output.put("successCount", result.successCount);
            output.put("failCount", result.failCount);
            if (result.error != null) {
                output.put("error", result.error);
            }
        } catch (Exception e) {}
        return new ToolResult(getToolId(), output, 0);
    }

    @NonNull
    private ToolResult executeFindText(@NonNull Map<String, Object> args) {
        String text = getStringArg(args, "text", null);
        if (text == null) {
            return new ToolResult(getToolId(), "Missing text");
        }

        // Note: This requires screenshot capability
        android.graphics.Bitmap bitmap = getScreenCapture().captureBitmap();
        if (bitmap == null) {
            return new ToolResult(getToolId(), "Screenshot capture failed");
        }

        try {
            List<SimpleOcrHelper.TextRegion> regions = getOcrHelper().findText(bitmap, text);

            JSONObject output = new JSONObject();
            try {
                output.put("success", !regions.isEmpty());
                output.put("count", regions.size());

                JSONArray regionsArray = new JSONArray();
                for (SimpleOcrHelper.TextRegion region : regions) {
                    JSONObject regionJson = new JSONObject();
                    regionJson.put("text", region.text);
                    regionJson.put("x", region.x);
                    regionJson.put("y", region.y);
                    regionJson.put("width", region.width);
                    regionJson.put("height", region.height);
                    regionJson.put("centerX", region.getCenterX());
                    regionJson.put("centerY", region.getCenterY());
                    regionsArray.put(regionJson);
                }
                output.put("regions", regionsArray);
            } catch (Exception e) {}

            return new ToolResult(getToolId(), output, 1000);
        } finally {
            bitmap.recycle();
        }
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
    private String getStringArg(@NonNull Map<String, Object> args, @NonNull String key, @Nullable String defaultVal) {
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

    @Nullable
    private Boolean getBoolArg(@NonNull Map<String, Object> args, @NonNull String key) {
        Object value = args.get(key);
        if (value == null) return null;
        if (value instanceof Boolean) return (Boolean) value;
        return Boolean.parseBoolean(value.toString());
    }
}