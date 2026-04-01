package com.ofa.agent.tool.builtin.system;

import android.content.Context;
import android.content.Intent;
import android.content.pm.ApplicationInfo;
import android.content.pm.PackageManager;
import android.graphics.drawable.Drawable;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.mcp.MCPProtocol;
import com.ofa.agent.mcp.ToolDefinition;
import com.ofa.agent.tool.ExecutionContext;
import com.ofa.agent.tool.ToolExecutor;
import com.ofa.agent.tool.ToolResult;

import org.json.JSONArray;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * App Tool - manage applications on the device.
 * Supports: launch, list, info operations.
 */
public class AppTool implements ToolExecutor {

    private static final String TAG = "AppTool";

    private final Context context;
    private final PackageManager packageManager;

    public AppTool(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.packageManager = context.getPackageManager();
    }

    // ===== Tool Executor Interface =====

    @NonNull
    @Override
    public String getToolId() {
        return "app";
    }

    @NonNull
    @Override
    public String getDescription() {
        return "Manage applications: launch, list, get info";
    }

    @NonNull
    @Override
    public ToolResult execute(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String operation = getStringArg(args, "operation", "list");

        switch (operation.toLowerCase()) {
            case "launch":
                return executeLaunch(args, ctx);
            case "list":
                return executeList(args, ctx);
            case "info":
                return executeInfo(args, ctx);
            default:
                return new ToolResult(getToolId(), "Unknown operation: " + operation);
        }
    }

    @Override
    public boolean isAvailable() {
        return true;
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
            case "launch":
                return args.containsKey("packageName");
            case "list":
            case "info":
                return true;
            default:
                return false;
        }
    }

    @Override
    public int getEstimatedTimeMs() {
        return 500;
    }

    // ===== Tool Definitions =====

    /**
     * Get tool definition for app.launch
     */
    @NonNull
    public static ToolDefinition getLaunchDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'launch'");
            operation.put("default", "launch");
            props.put("operation", operation);

            JSONObject packageName = new JSONObject();
            packageName.put("type", "string");
            packageName.put("description", "Package name to launch");
            props.put("packageName", packageName);
        } catch (Exception e) {
            // Should not fail
        }

        JSONObject schema = MCPProtocol.buildObjectSchema(props, new String[]{"packageName"});
        return new ToolDefinition("app.launch", "Launch an application by package name",
                schema, true, null);
    }

    /**
     * Get tool definition for app.list
     */
    @NonNull
    public static ToolDefinition getListDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'list'");
            operation.put("default", "list");
            props.put("operation", operation);

            JSONObject filter = new JSONObject();
            filter.put("type", "string");
            filter.put("description", "Filter type: 'all', 'system', 'user', 'thirdParty'");
            filter.put("default", "user");
            props.put("filter", filter);
        } catch (Exception e) {
            // Should not fail
        }

        JSONObject schema = MCPProtocol.buildObjectSchema(props, null);
        return new ToolDefinition("app.list", "List installed applications",
                schema, true, null);
    }

    /**
     * Get tool definition for app.info
     */
    @NonNull
    public static ToolDefinition getInfoDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'info'");
            operation.put("default", "info");
            props.put("operation", operation);

            JSONObject packageName = new JSONObject();
            packageName.put("type", "string");
            packageName.put("description", "Package name to get info");
            props.put("packageName", packageName);
        } catch (Exception e) {
            // Should not fail
        }

        JSONObject schema = MCPProtocol.buildObjectSchema(props, new String[]{"packageName"});
        return new ToolDefinition("app.info", "Get application information",
                schema, true, null);
    }

    // ===== Internal Operations =====

    @NonNull
    private ToolResult executeLaunch(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String packageName = getStringArg(args, "packageName", null);
        if (packageName == null) {
            return new ToolResult(getToolId(), "Missing packageName");
        }

        try {
            Intent intent = packageManager.getLaunchIntentForPackage(packageName);
            if (intent == null) {
                return new ToolResult(getToolId(), "Cannot launch package: " + packageName);
            }

            intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK);
            context.startActivity(intent);

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("packageName", packageName);
            output.put("launched", true);

            return new ToolResult(getToolId(), output, 100);

        } catch (Exception e) {
            Log.e(TAG, "Launch failed", e);
            return new ToolResult(getToolId(), "Launch failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeList(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String filter = getStringArg(args, "filter", "user");

        try {
            List<ApplicationInfo> apps;

            switch (filter.toLowerCase()) {
                case "all":
                    apps = packageManager.getInstalledApplications(PackageManager.GET_META_DATA);
                    break;
                case "system":
                    apps = getSystemApps();
                    break;
                case "thirdparty":
                case "user":
                    apps = getUserApps();
                    break;
                default:
                    apps = getUserApps();
            }

            JSONArray appsArray = new JSONArray();
            for (ApplicationInfo app : apps) {
                JSONObject appJson = new JSONObject();
                appJson.put("packageName", app.packageName);
                appJson.put("name", getAppName(app));
                appJson.put("system", (app.flags & ApplicationInfo.FLAG_SYSTEM) != 0);
                appsArray.put(appJson);
            }

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("count", apps.size());
            output.put("apps", appsArray);
            output.put("filter", filter);

            return new ToolResult(getToolId(), output, 200);

        } catch (Exception e) {
            Log.e(TAG, "List apps failed", e);
            return new ToolResult(getToolId(), "List failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeInfo(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String packageName = getStringArg(args, "packageName", null);
        if (packageName == null) {
            return new ToolResult(getToolId(), "Missing packageName");
        }

        try {
            ApplicationInfo appInfo = packageManager.getApplicationInfo(packageName,
                    PackageManager.GET_META_DATA);

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("packageName", packageName);
            output.put("name", getAppName(appInfo));
            output.put("system", (appInfo.flags & ApplicationInfo.FLAG_SYSTEM) != 0);
            output.put("enabled", appInfo.enabled);
            output.put("targetSdkVersion", appInfo.targetSdkVersion);
            output.put("sourceDir", appInfo.sourceDir);
            output.put("dataDir", appInfo.dataDir);

            // Version info
            try {
                android.content.pm.PackageInfo pkgInfo = packageManager.getPackageInfo(packageName, 0);
                output.put("versionName", pkgInfo.versionName);
                output.put("versionCode", pkgInfo.versionCode);
            } catch (Exception e) {
                output.put("versionName", "unknown");
            }

            return new ToolResult(getToolId(), output, 100);

        } catch (PackageManager.NameNotFoundException e) {
            return new ToolResult(getToolId(), "Package not found: " + packageName);
        } catch (Exception e) {
            Log.e(TAG, "Get info failed", e);
            return new ToolResult(getToolId(), "Get info failed: " + e.getMessage());
        }
    }

    // ===== Helper Methods =====

    @NonNull
    private List<ApplicationInfo> getUserApps() {
        List<ApplicationInfo> allApps = packageManager.getInstalledApplications(PackageManager.GET_META_DATA);
        List<ApplicationInfo> userApps = new ArrayList<>();

        for (ApplicationInfo app : allApps) {
            if ((app.flags & ApplicationInfo.FLAG_SYSTEM) == 0) {
                userApps.add(app);
            }
        }

        return userApps;
    }

    @NonNull
    private List<ApplicationInfo> getSystemApps() {
        List<ApplicationInfo> allApps = packageManager.getInstalledApplications(PackageManager.GET_META_DATA);
        List<ApplicationInfo> systemApps = new ArrayList<>();

        for (ApplicationInfo app : allApps) {
            if ((app.flags & ApplicationInfo.FLAG_SYSTEM) != 0) {
                systemApps.add(app);
            }
        }

        return systemApps;
    }

    @NonNull
    private String getAppName(@NonNull ApplicationInfo app) {
        try {
            return packageManager.getApplicationLabel(app).toString();
        } catch (Exception e) {
            return app.packageName;
        }
    }

    @Nullable
    private String getStringArg(@NonNull Map<String, Object> args, @NonNull String key, @Nullable String defaultVal) {
        Object value = args.get(key);
        if (value == null) return defaultVal;
        return value.toString();
    }
}