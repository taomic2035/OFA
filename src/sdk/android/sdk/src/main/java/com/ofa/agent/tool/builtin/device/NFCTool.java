package com.ofa.agent.tool.builtin.device;

import android.content.Context;
import android.nfc.NfcAdapter;
import android.nfc.NfcManager;
import android.os.Build;
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
 * NFC Tool - check NFC status and read/write tags.
 * Note: Actual NFC tag operations require Activity context and callbacks.
 */
public class NFCTool implements ToolExecutor {

    private static final String TAG = "NFCTool";

    private final Context context;
    private final NfcAdapter nfcAdapter;

    public NFCTool(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.nfcAdapter = NfcAdapter.getDefaultAdapter(context);
    }

    @NonNull
    @Override
    public String getToolId() {
        return "nfc";
    }

    @NonNull
    @Override
    public String getDescription() {
        return "Check NFC status and read/write tags";
    }

    @NonNull
    @Override
    public ToolResult execute(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String operation = getStringArg(args, "operation", "status");

        switch (operation.toLowerCase()) {
            case "status":
                return executeStatus(ctx);
            case "read":
                return executeRead(ctx);
            case "write":
                return executeWrite(args, ctx);
            case "enable":
                return executeEnable(ctx);
            case "disable":
                return executeDisable(ctx);
            default:
                return new ToolResult(getToolId(), "Unknown operation: " + operation);
        }
    }

    @Override
    public boolean isAvailable() {
        return nfcAdapter != null;
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
        return null; // NFC doesn't require normal permissions
    }

    @Override
    public boolean validateArgs(@NonNull Map<String, Object> args) {
        return true;
    }

    @Override
    public int getEstimatedTimeMs() {
        return 50;
    }

    // ===== Tool Definitions =====

    @NonNull
    public static ToolDefinition getStatusDefinition() {
        JSONObject schema = MCPProtocol.buildSimpleTextSchema("operation",
                "Operation: 'status'");
        return new ToolDefinition("nfc.status", "Check NFC adapter status",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getReadDefinition() {
        JSONObject schema = MCPProtocol.buildSimpleTextSchema("operation",
                "Operation: 'read'");
        return new ToolDefinition("nfc.read", "Read NFC tag content",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getWriteDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'write'");
            props.put("operation", operation);

            JSONObject data = new JSONObject();
            data.put("type", "string");
            data.put("description", "Data to write to NFC tag");
            props.put("data", data);
        } catch (Exception e) {
            // Should not fail
        }

        JSONObject schema = MCPProtocol.buildObjectSchema(props, new String[]{"data"});
        return new ToolDefinition("nfc.write", "Write data to NFC tag",
                schema, true, null);
    }

    // ===== Internal Operations =====

    @NonNull
    private ToolResult executeStatus(@NonNull ExecutionContext ctx) {
        JSONObject output = new JSONObject();

        try {
            output.put("success", true);
            output.put("available", nfcAdapter != null);

            if (nfcAdapter != null) {
                output.put("enabled", nfcAdapter.isEnabled());
                // NFC HCE status check removed - method deprecated/removed
            } else {
                output.put("enabled", false);
                output.put("message", "NFC not available on this device");
            }

            return new ToolResult(getToolId(), output, 20);

        } catch (Exception e) {
            Log.e(TAG, "Get status failed", e);
            return new ToolResult(getToolId(), "Get status failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeRead(@NonNull ExecutionContext ctx) {
        // Note: NFC reading requires Activity setup and callback handling
        // This is a placeholder that returns status

        try {
            JSONObject output = new JSONObject();
            output.put("success", false);
            output.put("message", "NFC reading requires Activity context and callback setup");
            output.put("hint", "Use NfcAdapter.enableReaderMode() in your Activity");
            return new ToolResult(getToolId(), output, 50);
        } catch (Exception e) {
            return new ToolResult(getToolId(), "Error: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeWrite(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String data = getStringArg(args, "data", null);

        if (data == null) {
            return new ToolResult(getToolId(), "Missing data parameter");
        }

        // Note: NFC writing requires Activity setup and callback handling

        try {
            JSONObject output = new JSONObject();
            output.put("success", false);
            output.put("message", "NFC writing requires Activity context and callback setup");
            output.put("hint", "Use NfcAdapter.enableReaderMode() in your Activity");
            output.put("dataProvided", data.length());
            return new ToolResult(getToolId(), output, 50);
        } catch (Exception e) {
            return new ToolResult(getToolId(), "Error: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeEnable(@NonNull ExecutionContext ctx) {
        // Note: Enabling NFC programmatically is not allowed on modern Android

        try {
            JSONObject output = new JSONObject();
            output.put("success", false);
            output.put("message", "NFC cannot be enabled programmatically on modern Android");
            output.put("hint", "User must enable NFC in Settings");
            return new ToolResult(getToolId(), output, 20);
        } catch (Exception e) {
            return new ToolResult(getToolId(), "Error: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeDisable(@NonNull ExecutionContext ctx) {
        // Note: Disabling NFC programmatically is not allowed on modern Android

        try {
            JSONObject output = new JSONObject();
            output.put("success", false);
            output.put("message", "NFC cannot be disabled programmatically on modern Android");
            output.put("hint", "User must disable NFC in Settings");
            return new ToolResult(getToolId(), output, 20);
        } catch (Exception e) {
            return new ToolResult(getToolId(), "Error: " + e.getMessage());
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