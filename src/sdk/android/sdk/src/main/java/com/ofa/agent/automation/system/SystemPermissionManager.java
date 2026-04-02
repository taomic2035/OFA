package com.ofa.agent.automation.system;

import android.content.Context;
import android.content.pm.PackageManager;
import android.os.Build;
import android.util.Log;

import androidx.annotation.NonNull;

import com.ofa.agent.automation.AutomationCapability;
import com.ofa.agent.automation.AutomationConfig;
import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.automation.AutomationResult;

import java.util.ArrayList;
import java.util.List;

/**
 * System Permission Manager - detects and manages system-level permissions.
 * Provides capability detection and graceful degradation for ROM-integrated scenarios.
 */
public class SystemPermissionManager {

    private static final String TAG = "SystemPermissionManager";

    private final Context context;
    private AutomationCapability currentCapability;
    private final List<PermissionStatus> permissionStatuses = new ArrayList<>();

    // Permission constants
    public static final String PERMISSION_INSTALL_PACKAGES = "android.permission.INSTALL_PACKAGES";
    public static final String PERMISSION_DELETE_PACKAGES = "android.permission.DELETE_PACKAGES";
    public static final String PERMISSION_WRITE_SECURE_SETTINGS = "android.permission.WRITE_SECURE_SETTINGS";
    public static final String PERMISSION_INTERACT_ACROSS_USERS = "android.permission.INTERACT_ACROSS_USERS";
    public static final String PERMISSION_MANAGE_USERS = "android.permission.MANAGE_USERS";
    public static final String PERMISSION_SYSTEM_ALERT_WINDOW = "android.permission.SYSTEM_ALERT_WINDOW";

    /**
     * Permission status result
     */
    public static class PermissionStatus {
        public final String permission;
        public final boolean granted;
        public final boolean canRequest;
        public final String reason;

        public PermissionStatus(String permission, boolean granted, boolean canRequest, String reason) {
            this.permission = permission;
            this.granted = granted;
            this.canRequest = canRequest;
            this.reason = reason;
        }
    }

    /**
     * Create permission manager
     */
    public SystemPermissionManager(@NonNull Context context) {
        this.context = context;
        detectCapabilities();
    }

    /**
     * Detect current system capabilities
     */
    @NonNull
    public AutomationCapability detectCapabilities() {
        Log.i(TAG, "Detecting system capabilities...");

        permissionStatuses.clear();

        // Check system-level permissions
        boolean hasInstallPermission = checkPermission(PERMISSION_INSTALL_PACKAGES);
        boolean hasDeletePermission = checkPermission(PERMISSION_DELETE_PACKAGES);
        boolean hasSecureSettings = checkPermission(PERMISSION_WRITE_SECURE_SETTINGS);
        boolean hasSystemAlert = checkSystemAlertWindow();

        // Add permission statuses
        permissionStatuses.add(new PermissionStatus(
            PERMISSION_INSTALL_PACKAGES, hasInstallPermission,
            canRequestRootOrSystem(), "Required for silent app installation"));

        permissionStatuses.add(new PermissionStatus(
            PERMISSION_DELETE_PACKAGES, hasDeletePermission,
            canRequestRootOrSystem(), "Required for silent app uninstallation"));

        permissionStatuses.add(new PermissionStatus(
            PERMISSION_WRITE_SECURE_SETTINGS, hasSecureSettings,
            canRequestRootOrSystem(), "Required for modifying secure settings"));

        permissionStatuses.add(new PermissionStatus(
            PERMISSION_SYSTEM_ALERT_WINDOW, hasSystemAlert,
            true, "Required for overlay windows and keep-alive"));

        // Determine capability level
        if (hasInstallPermission && hasDeletePermission && hasSecureSettings) {
            currentCapability = AutomationCapability.SYSTEM_LEVEL;
            Log.i(TAG, "Capability: SYSTEM_LEVEL - Full system access");
        } else if (hasSystemAlert) {
            currentCapability = AutomationCapability.ENHANCED;
            Log.i(TAG, "Capability: ENHANCED - Limited system access");
        } else {
            currentCapability = AutomationCapability.BASIC;
            Log.i(TAG, "Capability: BASIC - Standard accessibility");
        }

        return currentCapability;
    }

    /**
     * Get current capability level
     */
    @NonNull
    public AutomationCapability getCurrentCapability() {
        return currentCapability;
    }

    /**
     * Check if a permission is granted
     */
    public boolean checkPermission(@NonNull String permission) {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            return context.checkSelfPermission(permission) == PackageManager.PERMISSION_GRANTED;
        }
        return true; // Pre-M devices have all permissions at install
    }

    /**
     * Check system alert window permission
     */
    public boolean checkSystemAlertWindow() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            return android.provider.Settings.canDrawOverlays(context);
        }
        return true;
    }

    /**
     * Check if we can request root or system-level permissions
     */
    public boolean canRequestRootOrSystem() {
        // Check if running as system app
        boolean isSystemApp = isSystemApp();

        // Check if we have root access (try a simple command)
        boolean hasRoot = checkRootAccess();

        return isSystemApp || hasRoot;
    }

    /**
     * Check if this app is installed as system app
     */
    public boolean isSystemApp() {
        try {
            String appPath = context.getApplicationInfo().sourceDir;
            // System apps are typically in /system/app or /system/priv-app
            return appPath != null && (appPath.startsWith("/system/") ||
                                        appPath.startsWith("/vendor/") ||
                                        appPath.startsWith("/oem/"));
        } catch (Exception e) {
            Log.w(TAG, "Failed to check system app status: " + e.getMessage());
            return false;
        }
    }

    /**
     * Check if root access is available
     */
    public boolean checkRootAccess() {
        try {
            Process process = Runtime.getRuntime().exec("su");
            process.getOutputStream().write("exit\n".getBytes());
            process.getOutputStream().flush();
            int exitCode = process.waitFor();
            return exitCode == 0;
        } catch (Exception e) {
            Log.w(TAG, "Root check failed: " + e.getMessage());
            return false;
        }
    }

    /**
     * Get all permission statuses
     */
    @NonNull
    public List<PermissionStatus> getPermissionStatuses() {
        return new ArrayList<>(permissionStatuses);
    }

    /**
     * Check if accessibility service is enabled
     */
    public boolean isAccessibilityEnabled() {
        int accessibilityEnabled = 0;
        try {
            accessibilityEnabled = android.provider.Settings.Secure.getInt(
                context.getContentResolver(),
                android.provider.Settings.Secure.ACCESSIBILITY_ENABLED);
        } catch (android.provider.Settings.SettingNotFoundException e) {
            Log.w(TAG, "Accessibility setting not found");
        }
        return accessibilityEnabled == 1;
    }

    /**
     * Check if specific accessibility service is enabled
     */
    public boolean isSpecificAccessibilityEnabled(@NonNull String serviceName) {
        String enabledServices = android.provider.Settings.Secure.getString(
            context.getContentResolver(),
            android.provider.Settings.Secure.ENABLED_ACCESSIBILITY_SERVICES);

        if (enabledServices != null) {
            return enabledServices.contains(serviceName);
        }
        return false;
    }

    /**
     * Request accessibility service enable (user-guided)
     */
    @NonNull
    public AutomationResult requestAccessibilityEnable() {
        if (isAccessibilityEnabled()) {
            return new AutomationResult("accessibility", 0);
        }

        // Cannot enable directly, need user action
        Log.w(TAG, "Accessibility service needs user authorization");
        return new AutomationResult("accessibility",
            "Accessibility service not enabled. Please enable it in Settings.");
    }

    /**
     * Try to grant permission via root (if available)
     */
    @NonNull
    public AutomationResult grantPermissionViaRoot(@NonNull String permission) {
        if (!checkRootAccess()) {
            return new AutomationResult("grantPermission", "Root access not available");
        }

        try {
            String packageName = context.getPackageName();
            String command = "pm grant " + packageName + " " + permission;
            Process process = Runtime.getRuntime().exec(new String[]{"su", "-c", command});
            int exitCode = process.waitFor();

            if (exitCode == 0) {
                Log.i(TAG, "Permission granted via root: " + permission);
                detectCapabilities(); // Refresh capabilities
                return new AutomationResult("grantPermission", 0);
            } else {
                return new AutomationResult("grantPermission", "Root command failed");
            }
        } catch (Exception e) {
            Log.e(TAG, "Failed to grant permission via root: " + e.getMessage());
            return new AutomationResult("grantPermission", e.getMessage());
        }
    }

    /**
     * Get capability report as string
     */
    @NonNull
    public String getCapabilityReport() {
        StringBuilder sb = new StringBuilder();
        sb.append("System Capability Report:\n");
        sb.append("  Level: ").append(currentCapability.name()).append("\n");
        sb.append("  System App: ").append(isSystemApp()).append("\n");
        sb.append("  Root Access: ").append(checkRootAccess()).append("\n");
        sb.append("  Accessibility: ").append(isAccessibilityEnabled()).append("\n");
        sb.append("  Overlay Window: ").append(checkSystemAlertWindow()).append("\n");
        sb.append("\nPermissions:\n");

        for (PermissionStatus status : permissionStatuses) {
            sb.append("  ").append(status.permission).append(": ");
            sb.append(status.granted ? "GRANTED" : "NOT GRANTED");
            sb.append(" (").append(status.reason).append(")\n");
        }

        return sb.toString();
    }

    /**
     * Check if silent installation is possible
     */
    public boolean canSilentInstall() {
        return currentCapability == AutomationCapability.SYSTEM_LEVEL ||
               checkPermission(PERMISSION_INSTALL_PACKAGES);
    }

    /**
     * Check if secure settings modification is possible
     */
    public boolean canModifySecureSettings() {
        return currentCapability == AutomationCapability.SYSTEM_LEVEL ||
               checkPermission(PERMISSION_WRITE_SECURE_SETTINGS);
    }
}