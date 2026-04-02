package com.ofa.agent.automation;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;

import com.ofa.agent.automation.hybrid.HybridAutomationEngine;
import com.ofa.agent.automation.system.SystemAutomationEngine;
import com.ofa.agent.mcp.ToolDefinition;
import com.ofa.agent.tool.ToolExecutor;
import com.ofa.agent.tool.ToolResult;

import org.json.JSONException;
import org.json.JSONObject;

/**
 * System Tool - provides system-level automation operations.
 * Requires system permissions (INSTALL_PACKAGES, WRITE_SECURE_SETTINGS) or root access.
 */
public class SystemTool implements ToolExecutor {

    private static final String TAG = "SystemTool";

    private final Context context;
    private HybridAutomationEngine hybridEngine;

    /**
     * Create system tool
     */
    public SystemTool(@NonNull Context context) {
        this.context = context;
    }

    /**
     * Set hybrid engine (initialized externally)
     */
    public void setHybridEngine(@NonNull HybridAutomationEngine engine) {
        this.hybridEngine = engine;
    }

    // ===== ToolExecutor Interface =====

    @Override
    @NonNull
    public String getCategory() {
        return "system";
    }

    @Override
    @NonNull
    public String getDescription() {
        return "System-level automation operations (requires system/root permissions)";
    }

    @Override
    @NonNull
    public ToolDefinition getDefinition() {
        return getCapabilityDefinition();
    }

    @Override
    @NonNull
    public ToolResult execute(@NonNull String operation,
                               @NonNull JSONObject params,
                               @NonNull Context context) {
        Log.i(TAG, "Executing system operation: " + operation);

        if (hybridEngine == null) {
            return new ToolResult("system", "Hybrid engine not initialized", false);
        }

        SystemAutomationEngine systemEngine = hybridEngine.getSystemEngine();

        try {
            switch (operation) {
                case "system.install":
                    return handleInstall(params, systemEngine);

                case "system.uninstall":
                    return handleUninstall(params, systemEngine);

                case "system.grantPermission":
                    return handleGrantPermission(params, systemEngine);

                case "system.setSecureSetting":
                    return handleSetSetting(params, systemEngine);

                case "system.enableAccessibility":
                    return handleEnableAccessibility(params, systemEngine);

                case "system.keepAlive":
                    return handleKeepAlive(params, hybridEngine);

                case "system.getCapability":
                    return handleGetCapability(hybridEngine);

                default:
                    return new ToolResult("system", "Unknown operation: " + operation, false);
            }
        } catch (JSONException e) {
            return new ToolResult("system", "Parameter error: " + e.getMessage(), false);
        }
    }

    @Override
    public int getEstimatedTimeMs() {
        return 5000; // System operations can take time
    }

    // ===== Tool Definitions =====

    @NonNull
    public static ToolDefinition getInstallDefinition() {
        return new ToolDefinition(
            "system.install",
            "Install APK silently (requires system/root permission)",
            createParamSchema(new String[]{"apkPath"}, new String[]{"string"})
        );
    }

    @NonNull
    public static ToolDefinition getUninstallDefinition() {
        return new ToolDefinition(
            "system.uninstall",
            "Uninstall app silently (requires system/root permission)",
            createParamSchema(new String[]{"packageName"}, new String[]{"string"})
        );
    }

    @NonNull
    public static ToolDefinition getGrantPermissionDefinition() {
        return new ToolDefinition(
            "system.grantPermission",
            "Grant permission silently (requires system/root permission)",
            createParamSchema(new String[]{"packageName", "permission"}, new String[]{"string", "string"})
        );
    }

    @NonNull
    public static ToolDefinition getSetSecureSettingDefinition() {
        return new ToolDefinition(
            "system.setSecureSetting",
            "Set secure setting (requires system/root permission)",
            createParamSchema(new String[]{"key", "value"}, new String[]{"string", "string"})
        );
    }

    @NonNull
    public static ToolDefinition getEnableAccessibilityDefinition() {
        return new ToolDefinition(
            "system.enableAccessibility",
            "Enable accessibility service (requires system/root permission)",
            createParamSchema(new String[]{"serviceName"}, new String[]{"string"})
        );
    }

    @NonNull
    public static ToolDefinition getKeepAliveDefinition() {
        return new ToolDefinition(
            "system.keepAlive",
            "Enable/disable background keep-alive",
            createParamSchema(new String[]{"enable"}, new String[]{"boolean"})
        );
    }

    @NonNull
    public static ToolDefinition getCapabilityDefinition() {
        return new ToolDefinition(
            "system.getCapability",
            "Get system capability report",
            "{}"
        );
    }

    // ===== Helper Methods =====

    @NonNull
    private static String createParamSchema(String[] names, String[] types) {
        StringBuilder sb = new StringBuilder();
        sb.append("{\"type\":\"object\",\"properties\":{");

        for (int i = 0; i < names.length; i++) {
            if (i > 0) sb.append(",");
            sb.append("\"").append(names[i]).append("\":{\"type\":\"").append(types[i]).append("\"}");
        }

        sb.append("},\"required\":[");
        for (int i = 0; i < names.length; i++) {
            if (i > 0) sb.append(",");
            sb.append("\"").append(names[i]).append("\"");
        }
        sb.append("]}");

        return sb.toString();
    }

    // ===== Operation Handlers =====

    @NonNull
    private ToolResult handleInstall(@NonNull JSONObject params,
                                      @NonNull SystemAutomationEngine engine) throws JSONException {
        String apkPath = params.getString("apkPath");

        AutomationResult result = engine.installApp(apkPath);

        JSONObject data = new JSONObject();
        data.put("success", result.isSuccess());
        data.put("message", result.getMessage());

        return new ToolResult("system.install", data, result.isSuccess());
    }

    @NonNull
    private ToolResult handleUninstall(@NonNull JSONObject params,
                                        @NonNull SystemAutomationEngine engine) throws JSONException {
        String packageName = params.getString("packageName");

        AutomationResult result = engine.uninstallApp(packageName);

        JSONObject data = new JSONObject();
        data.put("success", result.isSuccess());
        data.put("message", result.getMessage());

        return new ToolResult("system.uninstall", data, result.isSuccess());
    }

    @NonNull
    private ToolResult handleGrantPermission(@NonNull JSONObject params,
                                              @NonNull SystemAutomationEngine engine) throws JSONException {
        String packageName = params.getString("packageName");
        String permission = params.getString("permission");

        AutomationResult result = engine.grantPermission(packageName, permission);

        JSONObject data = new JSONObject();
        data.put("success", result.isSuccess());
        data.put("message", result.getMessage());

        return new ToolResult("system.grantPermission", data, result.isSuccess());
    }

    @NonNull
    private ToolResult handleSetSetting(@NonNull JSONObject params,
                                         @NonNull SystemAutomationEngine engine) throws JSONException {
        String key = params.getString("key");
        String value = params.getString("value");

        AutomationResult result = engine.setSecureSetting(key, value);

        JSONObject data = new JSONObject();
        data.put("success", result.isSuccess());
        data.put("message", result.getMessage());

        return new ToolResult("system.setSecureSetting", data, result.isSuccess());
    }

    @NonNull
    private ToolResult handleEnableAccessibility(@NonNull JSONObject params,
                                                  @NonNull SystemAutomationEngine engine) throws JSONException {
        String serviceName = params.getString("serviceName");

        AutomationResult result = engine.enableAccessibilityService(serviceName);

        JSONObject data = new JSONObject();
        data.put("success", result.isSuccess());
        data.put("message", result.getMessage());

        return new ToolResult("system.enableAccessibility", data, result.isSuccess());
    }

    @NonNull
    private ToolResult handleKeepAlive(@NonNull JSONObject params,
                                        @NonNull HybridAutomationEngine engine) throws JSONException {
        boolean enable = params.optBoolean("enable", true);

        AutomationResult result = enable ?
            engine.enableKeepAlive() : engine.disableKeepAlive();

        JSONObject data = new JSONObject();
        data.put("success", result.isSuccess());
        data.put("enabled", enable);

        return new ToolResult("system.keepAlive", data, result.isSuccess());
    }

    @NonNull
    private ToolResult handleGetCapability(@NonNull HybridAutomationEngine engine) {
        String report = engine.getCapabilityReport();

        JSONObject data = new JSONObject();
        try {
            data.put("report", report);
            data.put("capability", engine.getCapability().name());
            data.put("systemLevelAccess", engine.hasSystemLevelAccess());
            data.put("activeEngine", engine.getActiveEngineType());
        } catch (JSONException e) {
            // Ignore
        }

        return new ToolResult("system.getCapability", data, true);
    }
}