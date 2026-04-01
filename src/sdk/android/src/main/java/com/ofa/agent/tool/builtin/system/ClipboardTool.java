package com.ofa.agent.tool.builtin.system;

import android.content.ClipData;
import android.content.ClipboardManager;
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
 * Clipboard Tool - read and write clipboard content.
 */
public class ClipboardTool implements ToolExecutor {

    private static final String TAG = "ClipboardTool";

    private final Context context;
    private final ClipboardManager clipboardManager;

    public ClipboardTool(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.clipboardManager = (ClipboardManager) context.getSystemService(Context.CLIPBOARD_SERVICE);
    }

    @NonNull
    @Override
    public String getToolId() {
        return "clipboard";
    }

    @NonNull
    @Override
    public String getDescription() {
        return "Read and write clipboard content";
    }

    @NonNull
    @Override
    public ToolResult execute(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String operation = getStringArg(args, "operation", "read");

        switch (operation.toLowerCase()) {
            case "read":
                return executeRead(ctx);
            case "write":
                return executeWrite(args, ctx);
            case "clear":
                return executeClear(ctx);
            default:
                return new ToolResult(getToolId(), "Unknown operation: " + operation);
        }
    }

    @Override
    public boolean isAvailable() {
        return clipboardManager != null;
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

        if ("write".equalsIgnoreCase(operation)) {
            return args.containsKey("text");
        }

        return true;
    }

    @Override
    public int getEstimatedTimeMs() {
        return 10;
    }

    // ===== Tool Definitions =====

    @NonNull
    public static ToolDefinition getReadDefinition() {
        JSONObject schema = MCPProtocol.buildSimpleTextSchema("operation",
                "Operation: 'read' (default)");
        return new ToolDefinition("clipboard.read", "Read clipboard content",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getWriteDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'write'");
            operation.put("default", "write");
            props.put("operation", operation);

            JSONObject text = new JSONObject();
            text.put("type", "string");
            text.put("description", "Text to write to clipboard");
            props.put("text", text);

            JSONObject label = new JSONObject();
            label.put("type", "string");
            label.put("description", "Optional label for clipboard content");
            props.put("label", label);
        } catch (Exception e) {
            // Should not fail
        }

        JSONObject schema = MCPProtocol.buildObjectSchema(props, new String[]{"text"});
        return new ToolDefinition("clipboard.write", "Write text to clipboard",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getClearDefinition() {
        JSONObject schema = MCPProtocol.buildSimpleTextSchema("operation",
                "Operation: 'clear'");
        return new ToolDefinition("clipboard.clear", "Clear clipboard content",
                schema, true, null);
    }

    // ===== Internal Operations =====

    @NonNull
    private ToolResult executeRead(@NonNull ExecutionContext ctx) {
        if (clipboardManager == null) {
            return new ToolResult(getToolId(), "Clipboard manager not available");
        }

        try {
            ClipData clip = clipboardManager.getPrimaryClip();

            JSONObject output = new JSONObject();
            output.put("success", true);

            if (clip != null && clip.getItemCount() > 0) {
                ClipData.Item item = clip.getItemAt(0);
                String text = item.getText() != null ? item.getText().toString() : "";
                String label = clip.getDescription().getLabel() != null
                        ? clip.getDescription().getLabel().toString() : "";

                output.put("text", text);
                output.put("label", label);
                output.put("hasContent", true);
                output.put("itemCount", clip.getItemCount());
            } else {
                output.put("text", "");
                output.put("hasContent", false);
            }

            return new ToolResult(getToolId(), output, 5);

        } catch (Exception e) {
            Log.e(TAG, "Read clipboard failed", e);
            return new ToolResult(getToolId(), "Read failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeWrite(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        if (clipboardManager == null) {
            return new ToolResult(getToolId(), "Clipboard manager not available");
        }

        String text = getStringArg(args, "text", "");
        String label = getStringArg(args, "label", "OFA Clipboard");

        try {
            ClipData clip = ClipData.newPlainText(label, text);
            clipboardManager.setPrimaryClip(clip);

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("text", text);
            output.put("label", label);
            output.put("written", true);

            return new ToolResult(getToolId(), output, 5);

        } catch (Exception e) {
            Log.e(TAG, "Write clipboard failed", e);
            return new ToolResult(getToolId(), "Write failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeClear(@NonNull ExecutionContext ctx) {
        if (clipboardManager == null) {
            return new ToolResult(getToolId(), "Clipboard manager not available");
        }

        try {
            // Set empty clip to clear
            ClipData clip = ClipData.newPlainText("", "");
            clipboardManager.setPrimaryClip(clip);

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("cleared", true);

            return new ToolResult(getToolId(), output, 5);

        } catch (Exception e) {
            Log.e(TAG, "Clear clipboard failed", e);
            return new ToolResult(getToolId(), "Clear failed: " + e.getMessage());
        }
    }

    // ===== Helper Methods =====

    @Nullable
    private String getStringArg(@NonNull Map<String, Object> args, @NonNull String key, @Nullable String defaultVal) {
        Object value = args.get(key);
        if (value == null) return defaultVal;
        return value.toString();
    }
}