package com.ofa.agent.automation.system;

import android.content.Context;
import android.os.Build;
import android.provider.Settings;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationCapability;
import com.ofa.agent.automation.AutomationConfig;
import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.automation.AutomationNode;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.automation.BySelector;
import com.ofa.agent.automation.Direction;
import com.ofa.agent.automation.accessibility.AccessibilityEngine;

import android.graphics.Bitmap;

import java.util.List;
import java.util.Map;

/**
 * System Automation Engine - provides system-level automation capabilities.
 * Falls back to accessibility-based engine when system permissions are not available.
 */
public class SystemAutomationEngine implements AutomationEngine {

    private static final String TAG = "SystemAutomationEngine";

    private final Context context;
    private final SystemPermissionManager permissionManager;
    private final SilentInstaller silentInstaller;
    private final KeepAliveManager keepAliveManager;

    // Primary engine (accessibility)
    private final AccessibilityEngine accessibilityEngine;

    // Configuration
    private AutomationConfig config;
    private boolean initialized = false;

    /**
     * Create system automation engine
     */
    public SystemAutomationEngine(@NonNull Context context,
                                   @NonNull AccessibilityEngine accessibilityEngine) {
        this.context = context;
        this.accessibilityEngine = accessibilityEngine;
        this.permissionManager = new SystemPermissionManager(context);
        this.silentInstaller = new SilentInstaller(context, permissionManager);
        this.keepAliveManager = new KeepAliveManager(context, permissionManager);
    }

    /**
     * Get permission manager
     */
    @NonNull
    public SystemPermissionManager getPermissionManager() {
        return permissionManager;
    }

    /**
     * Get silent installer
     */
    @NonNull
    public SilentInstaller getSilentInstaller() {
        return silentInstaller;
    }

    /**
     * Get keep-alive manager
     */
    @NonNull
    public KeepAliveManager getKeepAliveManager() {
        return keepAliveManager;
    }

    // ===== Lifecycle =====

    @Override
    public void initialize(@NonNull AutomationConfig config) {
        Log.i(TAG, "Initializing System Automation Engine...");

        this.config = config;
        permissionManager.detectCapabilities();

        // Initialize accessibility engine
        accessibilityEngine.initialize(config);

        // Enable keep-alive if requested
        if (config.enableKeepAlive()) {
            keepAliveManager.enableAll();
        }

        initialized = true;
        Log.i(TAG, "Initialized with capability: " + permissionManager.getCurrentCapability());
    }

    @Override
    public void shutdown() {
        Log.i(TAG, "Shutting down System Automation Engine...");

        accessibilityEngine.shutdown();
        keepAliveManager.disableAll();
        initialized = false;
    }

    // ===== Capability Detection =====

    @Override
    public boolean isAvailable() {
        return accessibilityEngine.isAvailable();
    }

    @Override
    @NonNull
    public AutomationCapability getCapability() {
        return permissionManager.getCurrentCapability();
    }

    /**
     * Check if system-level operations are available
     */
    public boolean hasSystemLevelAccess() {
        return permissionManager.getCurrentCapability() == AutomationCapability.SYSTEM_LEVEL;
    }

    // ===== Basic Operations (delegate to accessibility) =====

    @Override
    @NonNull
    public AutomationResult click(int x, int y) {
        return accessibilityEngine.click(x, y);
    }

    @Override
    @NonNull
    public AutomationResult click(@NonNull String text) {
        return accessibilityEngine.click(text);
    }

    @Override
    @NonNull
    public AutomationResult click(@NonNull BySelector selector) {
        return accessibilityEngine.click(selector);
    }

    @Override
    @NonNull
    public AutomationResult longClick(int x, int y) {
        return accessibilityEngine.longClick(x, y);
    }

    @Override
    @NonNull
    public AutomationResult longClick(@NonNull String text) {
        return accessibilityEngine.longClick(text);
    }

    @Override
    @NonNull
    public AutomationResult swipe(int fromX, int fromY, int toX, int toY, long duration) {
        return accessibilityEngine.swipe(fromX, fromY, toX, toY, duration);
    }

    @Override
    @NonNull
    public AutomationResult swipe(@NonNull Direction direction, float distance) {
        return accessibilityEngine.swipe(direction, distance);
    }

    @Override
    @NonNull
    public AutomationResult inputText(@NonNull String text) {
        return accessibilityEngine.inputText(text);
    }

    @Override
    @NonNull
    public AutomationResult inputText(@NonNull BySelector selector, @NonNull String text) {
        return accessibilityEngine.inputText(selector, text);
    }

    // ===== Advanced Operations =====

    @Override
    @NonNull
    public AutomationResult scrollFind(@NonNull BySelector selector, int maxScrolls) {
        return accessibilityEngine.scrollFind(selector, maxScrolls);
    }

    @Override
    @NonNull
    public AutomationResult waitForElement(@NonNull BySelector selector, long timeout) {
        return accessibilityEngine.waitForElement(selector, timeout);
    }

    @Override
    @NonNull
    public AutomationResult waitForPageStable(long timeout) {
        return accessibilityEngine.waitForPageStable(timeout);
    }

    // ===== Query Operations =====

    @Override
    @Nullable
    public AutomationNode findElement(@NonNull BySelector selector) {
        return accessibilityEngine.findElement(selector);
    }

    @Override
    @NonNull
    public List<AutomationNode> findElements(@NonNull BySelector selector) {
        return accessibilityEngine.findElements(selector);
    }

    @Override
    @NonNull
    public String getPageSource() {
        return accessibilityEngine.getPageSource();
    }

    @Override
    @Nullable
    public Bitmap takeScreenshot() {
        return accessibilityEngine.takeScreenshot();
    }

    @Override
    @Nullable
    public String getForegroundPackage() {
        return accessibilityEngine.getForegroundPackage();
    }

    // ===== Callbacks =====

    @Override
    public void setListener(@Nullable AutomationListener listener) {
        accessibilityEngine.setListener(listener);
    }

    // ===== System-Level Operations =====

    /**
     * Install APK (silent if possible)
     */
    @NonNull
    public AutomationResult installApp(@NonNull String apkPath) {
        Log.i(TAG, "Installing app: " + apkPath);

        SilentInstaller.InstallResult result = silentInstaller.install(apkPath);

        if (result.success) {
            return new AutomationResult("installApp", 0);
        }

        // Return result with appropriate message
        return new AutomationResult("installApp", result.message);
    }

    /**
     * Uninstall app (silent if possible)
     */
    @NonNull
    public AutomationResult uninstallApp(@NonNull String packageName) {
        Log.i(TAG, "Uninstalling app: " + packageName);

        SilentInstaller.InstallResult result = silentInstaller.uninstall(packageName);

        if (result.success) {
            return new AutomationResult("uninstallApp", 0);
        }

        return new AutomationResult("uninstallApp", result.message);
    }

    /**
     * Grant permission silently (requires system/root)
     */
    @NonNull
    public AutomationResult grantPermission(@NonNull String packageName,
                                             @NonNull String permission) {
        Log.i(TAG, "Granting permission: " + permission + " to " + packageName);

        if (!hasSystemLevelAccess()) {
            // Try via root
            if (permissionManager.checkRootAccess()) {
                return grantPermissionViaRoot(packageName, permission);
            }
            return new AutomationResult("grantPermission",
                "System-level access not available");
        }

        try {
            String command = "pm grant " + packageName + " " + permission;
            Process process = Runtime.getRuntime().exec(new String[]{"su", "-c", command});
            int exitCode = process.waitFor();

            if (exitCode == 0) {
                return new AutomationResult("grantPermission", 0);
            }
            return new AutomationResult("grantPermission", "Command failed");
        } catch (Exception e) {
            return new AutomationResult("grantPermission", e.getMessage());
        }
    }

    /**
     * Grant permission via root
     */
    @NonNull
    private AutomationResult grantPermissionViaRoot(@NonNull String packageName,
                                                     @NonNull String permission) {
        try {
            String command = "pm grant " + packageName + " " + permission;
            Process process = Runtime.getRuntime().exec(new String[]{"su", "-c", command});
            int exitCode = process.waitFor();

            if (exitCode == 0) {
                return new AutomationResult("grantPermission", 0);
            }
            return new AutomationResult("grantPermission", "Root command failed");
        } catch (Exception e) {
            return new AutomationResult("grantPermission", e.getMessage());
        }
    }

    /**
     * Modify secure setting (requires system/root)
     */
    @NonNull
    public AutomationResult setSecureSetting(@NonNull String key, @NonNull String value) {
        Log.i(TAG, "Setting secure setting: " + key + " = " + value);

        if (!permissionManager.canModifySecureSettings()) {
            // Try via root
            if (permissionManager.checkRootAccess()) {
                return setSecureSettingViaRoot(key, value);
            }
            return new AutomationResult("setSecureSetting",
                "Secure settings permission not available");
        }

        try {
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
                // Try direct modification
                Settings.Secure.putString(context.getContentResolver(), key, value);
                return new AutomationResult("setSecureSetting", 0);
            }
            return new AutomationResult("setSecureSetting", "Not supported on this API level");
        } catch (Exception e) {
            Log.w(TAG, "Direct setting failed, trying root: " + e.getMessage());
            return setSecureSettingViaRoot(key, value);
        }
    }

    /**
     * Set secure setting via root
     */
    @NonNull
    private AutomationResult setSecureSettingViaRoot(@NonNull String key, @NonNull String value) {
        try {
            String command = "settings put secure " + key + " " + value;
            Process process = Runtime.getRuntime().exec(new String[]{"su", "-c", command});
            int exitCode = process.waitFor();

            if (exitCode == 0) {
                return new AutomationResult("setSecureSetting", 0);
            }
            return new AutomationResult("setSecureSetting", "Root command failed");
        } catch (Exception e) {
            return new AutomationResult("setSecureSetting", e.getMessage());
        }
    }

    /**
     * Enable accessibility service (requires system/root)
     */
    @NonNull
    public AutomationResult enableAccessibilityService(@NonNull String serviceName) {
        Log.i(TAG, "Enabling accessibility service: " + serviceName);

        if (!hasSystemLevelAccess() && !permissionManager.checkRootAccess()) {
            return new AutomationResult("enableAccessibility",
                "Cannot enable accessibility service - requires system access");
        }

        try {
            // Build the enabled services string
            String enabledServices = Settings.Secure.getString(
                context.getContentResolver(), Settings.Secure.ENABLED_ACCESSIBILITY_SERVICES);

            if (enabledServices != null && enabledServices.contains(serviceName)) {
                return new AutomationResult("enableAccessibility", 0);
            }

            String newEnabledServices = (enabledServices != null && !enabledServices.isEmpty()) ?
                enabledServices + ":" + serviceName : serviceName;

            // Set via root
            String command = "settings put secure enabled_accessibility_services " + newEnabledServices;
            Process process = Runtime.getRuntime().exec(new String[]{"su", "-c", command});
            int exitCode = process.waitFor();

            if (exitCode == 0) {
                // Also enable accessibility globally
                command = "settings put secure accessibility_enabled 1";
                Runtime.getRuntime().exec(new String[]{"su", "-c", command});

                return new AutomationResult("enableAccessibility", 0);
            }
            return new AutomationResult("enableAccessibility", "Root command failed");
        } catch (Exception e) {
            return new AutomationResult("enableAccessibility", e.getMessage());
        }
    }

    /**
     * Enable keep-alive strategies
     */
    @NonNull
    public AutomationResult enableKeepAlive() {
        return keepAliveManager.enableAll();
    }

    /**
     * Disable keep-alive strategies
     */
    @NonNull
    public AutomationResult disableKeepAlive() {
        return keepAliveManager.disableAll();
    }

    /**
     * Get capability report
     */
    @NonNull
    public String getCapabilityReport() {
        StringBuilder sb = new StringBuilder();

        sb.append("=== System Automation Engine Report ===\n\n");
        sb.append(permissionManager.getCapabilityReport());
        sb.append("\n");
        sb.append(keepAliveManager.getStatusReport());
        sb.append("\n");
        sb.append("Silent Install: ").append(silentInstaller.canInstallSilently() ? "Available" : "Not Available").append("\n");

        return sb.toString();
    }

    // ===== Config =====

    @Override
    @NonNull
    public AutomationConfig getConfig() {
        return config != null ? config : AutomationConfig.getDefault();
    }
}