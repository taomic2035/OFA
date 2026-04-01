package com.ofa.agent.tool.builtin.device;

import android.bluetooth.BluetoothAdapter;
import android.bluetooth.BluetoothDevice;
import android.bluetooth.BluetoothManager;
import android.content.Context;
import android.os.Build;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;
import androidx.core.content.ContextCompat;

import com.ofa.agent.mcp.MCPProtocol;
import com.ofa.agent.mcp.ToolDefinition;
import com.ofa.agent.tool.PermissionManager;
import com.ofa.agent.tool.ExecutionContext;
import com.ofa.agent.tool.ToolExecutor;
import com.ofa.agent.tool.ToolResult;

import org.json.JSONArray;
import org.json.JSONObject;

import java.util.Map;
import java.util.Set;

/**
 * Bluetooth Tool - scan and manage Bluetooth devices.
 */
public class BluetoothTool implements ToolExecutor {

    private static final String TAG = "BluetoothTool";

    private final Context context;
    private final BluetoothManager bluetoothManager;
    private final BluetoothAdapter bluetoothAdapter;

    public BluetoothTool(@NonNull Context context) {
        this.context = context.getApplicationContext();

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            this.bluetoothManager = (BluetoothManager) context.getSystemService(Context.BLUETOOTH_SERVICE);
            this.bluetoothAdapter = bluetoothManager != null ? bluetoothManager.getAdapter() : null;
        } else {
            this.bluetoothManager = null;
            this.bluetoothAdapter = BluetoothAdapter.getDefaultAdapter();
        }
    }

    @NonNull
    @Override
    public String getToolId() {
        return "bluetooth";
    }

    @NonNull
    @Override
    public String getDescription() {
        return "Scan and manage Bluetooth devices";
    }

    @NonNull
    @Override
    public ToolResult execute(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String operation = getStringArg(args, "operation", "scan");

        switch (operation.toLowerCase()) {
            case "scan":
                return executeScan(ctx);
            case "list":
                return executeList(ctx);
            case "info":
                return executeInfo(ctx);
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
        return bluetoothAdapter != null;
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
        return PermissionManager.getBluetoothPermissions();
    }

    @Override
    public boolean validateArgs(@NonNull Map<String, Object> args) {
        return true;
    }

    @Override
    public int getEstimatedTimeMs() {
        return 5000;
    }

    // ===== Tool Definitions =====

    @NonNull
    public static ToolDefinition getScanDefinition() {
        JSONObject schema = MCPProtocol.buildSimpleTextSchema("operation",
                "Operation: 'scan'");
        return new ToolDefinition("bluetooth.scan", "Scan for nearby Bluetooth devices",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getListDefinition() {
        JSONObject schema = MCPProtocol.buildSimpleTextSchema("operation",
                "Operation: 'list'");
        return new ToolDefinition("bluetooth.list", "List paired Bluetooth devices",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getInfoDefinition() {
        JSONObject schema = MCPProtocol.buildSimpleTextSchema("operation",
                "Operation: 'info'");
        return new ToolDefinition("bluetooth.status", "Get Bluetooth adapter status",
                schema, true, null);
    }

    // ===== Internal Operations =====

    @NonNull
    private ToolResult executeScan(@NonNull ExecutionContext ctx) {
        // Note: Bluetooth scanning requires additional setup and callbacks
        // This is a simplified implementation

        try {
            JSONObject output = new JSONObject();

            if (!bluetoothAdapter.isEnabled()) {
                output.put("success", false);
                output.put("message", "Bluetooth is disabled");
                return new ToolResult(getToolId(), output, 50);
            }

            // Get bonded devices instead of actual scan (scan requires callbacks)
            Set<BluetoothDevice> bondedDevices = bluetoothAdapter.getBondedDevices();

            JSONArray devicesArray = new JSONArray();
            for (BluetoothDevice device : bondedDevices) {
                JSONObject devJson = new JSONObject();
                devJson.put("name", device.getName());
                devJson.put("address", device.getAddress());
                devJson.put("bonded", true);
                devicesArray.put(devJson);
            }

            output.put("success", true);
            output.put("message", "Listing bonded devices (actual scan requires callback setup)");
            output.put("count", bondedDevices.size());
            output.put("devices", devicesArray);

            return new ToolResult(getToolId(), output, 500);

        } catch (SecurityException e) {
            return new ToolResult(getToolId(), "Permission denied: " + e.getMessage());
        } catch (Exception e) {
            Log.e(TAG, "Scan failed", e);
            return new ToolResult(getToolId(), "Scan failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeList(@NonNull ExecutionContext ctx) {
        try {
            Set<BluetoothDevice> bondedDevices = bluetoothAdapter.getBondedDevices();

            JSONArray devicesArray = new JSONArray();
            for (BluetoothDevice device : bondedDevices) {
                JSONObject devJson = new JSONObject();
                devJson.put("name", device.getName());
                devJson.put("address", device.getAddress());
                devJson.put("bondState", device.getBondState());
                devicesArray.put(devJson);
            }

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("count", bondedDevices.size());
            output.put("devices", devicesArray);

            return new ToolResult(getToolId(), output, 100);

        } catch (SecurityException e) {
            return new ToolResult(getToolId(), "Permission denied: " + e.getMessage());
        } catch (Exception e) {
            Log.e(TAG, "List failed", e);
            return new ToolResult(getToolId(), "List failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeInfo(@NonNull ExecutionContext ctx) {
        try {
            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("enabled", bluetoothAdapter.isEnabled());
            output.put("name", bluetoothAdapter.getName());
            output.put("address", bluetoothAdapter.getAddress());
            output.put("scanMode", bluetoothAdapter.getScanMode());
            output.put("discoverable", bluetoothAdapter.getScanMode() == BluetoothAdapter.SCAN_MODE_CONNECTABLE_DISCOVERABLE);

            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M && bluetoothManager != null) {
                output.put("connectedDevices", bluetoothManager.getConnectedDevices(BluetoothDevice.DEVICE_TYPE_LE).size());
            }

            return new ToolResult(getToolId(), output, 50);

        } catch (SecurityException e) {
            return new ToolResult(getToolId(), "Permission denied: " + e.getMessage());
        } catch (Exception e) {
            Log.e(TAG, "Get info failed", e);
            return new ToolResult(getToolId(), "Get info failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeEnable(@NonNull ExecutionContext ctx) {
        try {
            boolean success = bluetoothAdapter.enable();

            JSONObject output = new JSONObject();
            output.put("success", success);
            output.put("enabled", bluetoothAdapter.isEnabled());
            output.put("action", "enable");

            return new ToolResult(getToolId(), output, 100);

        } catch (SecurityException e) {
            return new ToolResult(getToolId(), "Permission denied: " + e.getMessage());
        } catch (Exception e) {
            Log.e(TAG, "Enable failed", e);
            return new ToolResult(getToolId(), "Enable failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeDisable(@NonNull ExecutionContext ctx) {
        try {
            boolean success = bluetoothAdapter.disable();

            JSONObject output = new JSONObject();
            output.put("success", success);
            output.put("enabled", bluetoothAdapter.isEnabled());
            output.put("action", "disable");

            return new ToolResult(getToolId(), output, 100);

        } catch (SecurityException e) {
            return new ToolResult(getToolId(), "Permission denied: " + e.getMessage());
        } catch (Exception e) {
            Log.e(TAG, "Disable failed", e);
            return new ToolResult(getToolId(), "Disable failed: " + e.getMessage());
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