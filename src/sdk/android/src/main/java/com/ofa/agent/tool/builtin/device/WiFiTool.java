package com.ofa.agent.tool.builtin.device;

import android.Manifest;
import android.content.Context;
import android.net.wifi.ScanResult;
import android.net.wifi.WifiConfiguration;
import android.net.wifi.WifiInfo;
import android.net.wifi.WifiManager;
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

import java.util.List;
import java.util.Map;

/**
 * WiFi Tool - scan and manage WiFi connections.
 */
public class WiFiTool implements ToolExecutor {

    private static final String TAG = "WiFiTool";

    private final Context context;
    private final WifiManager wifiManager;

    public WiFiTool(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.wifiManager = (WifiManager) context.getSystemService(Context.WIFI_SERVICE);
    }

    @NonNull
    @Override
    public String getToolId() {
        return "wifi";
    }

    @NonNull
    @Override
    public String getDescription() {
        return "Scan and manage WiFi connections";
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
        return wifiManager != null;
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
        return PermissionManager.getLocationPermissions(); // WiFi scan needs location permission
    }

    @Override
    public boolean validateArgs(@NonNull Map<String, Object> args) {
        return true;
    }

    @Override
    public int getEstimatedTimeMs() {
        return 3000;
    }

    // ===== Tool Definitions =====

    @NonNull
    public static ToolDefinition getScanDefinition() {
        JSONObject schema = MCPProtocol.buildSimpleTextSchema("operation",
                "Operation: 'scan'");
        return new ToolDefinition("wifi.scan", "Scan for nearby WiFi networks",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getListDefinition() {
        JSONObject schema = MCPProtocol.buildSimpleTextSchema("operation",
                "Operation: 'list'");
        return new ToolDefinition("wifi.list", "List configured WiFi networks",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getInfoDefinition() {
        JSONObject schema = MCPProtocol.buildSimpleTextSchema("operation",
                "Operation: 'info'");
        return new ToolDefinition("wifi.status", "Get current WiFi connection status",
                schema, true, null);
    }

    // ===== Internal Operations =====

    @NonNull
    private ToolResult executeScan(@NonNull ExecutionContext ctx) {
        try {
            boolean scanStarted = wifiManager.startScan();

            JSONObject output = new JSONObject();
            output.put("success", scanStarted);

            if (!scanStarted) {
                output.put("message", "Scan failed to start");
                return new ToolResult(getToolId(), output, 50);
            }

            // Get scan results (may be from previous scan)
            List<ScanResult> scanResults = wifiManager.getScanResults();

            JSONArray networksArray = new JSONArray();
            for (ScanResult result : scanResults) {
                JSONObject netJson = new JSONObject();
                netJson.put("ssid", result.SSID);
                netJson.put("bssid", result.BSSID);
                netJson.put("frequency", result.frequency);
                netJson.put("level", result.level);
                netJson.put("capabilities", result.capabilities);
                networksArray.put(netJson);
            }

            output.put("count", scanResults.size());
            output.put("networks", networksArray);

            return new ToolResult(getToolId(), output, 1000);

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
            List<WifiConfiguration> configurations = wifiManager.getConfiguredNetworks();

            JSONArray networksArray = new JSONArray();
            if (configurations != null) {
                for (WifiConfiguration config : configurations) {
                    JSONObject netJson = new JSONObject();
                    netJson.put("networkId", config.networkId);
                    netJson.put("ssid", config.SSID);
                    netJson.put("bssid", config.BSSID);
                    netJson.put("hiddenSSID", config.hiddenSSID);
                    netJson.put("priority", config.priority);
                    netJson.put("status", config.status);
                    networksArray.put(netJson);
                }
            }

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("count", configurations != null ? configurations.size() : 0);
            output.put("networks", networksArray);

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
            WifiInfo connectionInfo = wifiManager.getConnectionInfo();

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("enabled", wifiManager.isWifiEnabled());
            output.put("connected", connectionInfo != null && connectionInfo.getNetworkId() != -1);

            if (connectionInfo != null) {
                output.put("ssid", connectionInfo.getSSID());
                output.put("bssid", connectionInfo.getBSSID());
                output.put("networkId", connectionInfo.getNetworkId());
                output.put("linkSpeed", connectionInfo.getLinkSpeed());
                output.put("signalStrength", connectionInfo.getRssi());
                output.put("ipAddress", formatIpAddress(connectionInfo.getIpAddress()));
                output.put("macAddress", connectionInfo.getMacAddress());
            }

            output.put("scanResultsAvailable", wifiManager.getScanResults().size());

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
        boolean success = wifiManager.setWifiEnabled(true);

        JSONObject output = new JSONObject();
        output.put("success", success);
        output.put("enabled", wifiManager.isWifiEnabled());
        output.put("action", "enable");

        return new ToolResult(getToolId(), output, 100);
    }

    @NonNull
    private ToolResult executeDisable(@NonNull ExecutionContext ctx) {
        boolean success = wifiManager.setWifiEnabled(false);

        JSONObject output = new JSONObject();
        output.put("success", success);
        output.put("enabled", wifiManager.isWifiEnabled());
        output.put("action", "disable");

        return new ToolResult(getToolId(), output, 100);
    }

    // ===== Helper Methods =====

    @NonNull
    private String formatIpAddress(int ipAddress) {
        return String.format("%d.%d.%d.%d",
                (ipAddress & 0xff),
                (ipAddress >> 8 & 0xff),
                (ipAddress >> 16 & 0xff),
                (ipAddress >> 24 & 0xff));
    }

    @Nullable
    private String getStringArg(@NonNull Map<String, Object> args, @NonNull String key, @Nullable String defaultVal) {
        Object value = args.get(key);
        if (value == null) return defaultVal;
        return value.toString();
    }
}