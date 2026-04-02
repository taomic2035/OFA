package com.ofa.agent.automation.hybrid;

import android.content.Context;
import android.graphics.Bitmap;
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
import com.ofa.agent.automation.system.SystemAutomationEngine;

import java.util.List;

/**
 * Hybrid Automation Engine - combines accessibility and system-level capabilities.
 * Automatically chooses the best approach based on current permissions and requirements.
 */
public class HybridAutomationEngine implements AutomationEngine {

    private static final String TAG = "HybridAutomationEngine";

    private final Context context;

    // Dual engines
    private final AccessibilityEngine accessibilityEngine;
    private final SystemAutomationEngine systemEngine;

    // Current active engine
    private AutomationEngine activeEngine;

    // Configuration
    private AutomationConfig config;
    private boolean initialized = false;
    private AutomationListener listener;

    /**
     * Create hybrid automation engine
     */
    public HybridAutomationEngine(@NonNull Context context) {
        this.context = context;

        // Create accessibility engine
        this.accessibilityEngine = new AccessibilityEngine(context);

        // Create system engine with accessibility as fallback
        this.systemEngine = new SystemAutomationEngine(context, accessibilityEngine);

        // Default to accessibility
        this.activeEngine = accessibilityEngine;
    }

    /**
     * Get system engine for system-level operations
     */
    @NonNull
    public SystemAutomationEngine getSystemEngine() {
        return systemEngine;
    }

    /**
     * Get accessibility engine
     */
    @NonNull
    public AccessibilityEngine getAccessibilityEngine() {
        return accessibilityEngine;
    }

    /**
     * Select best engine for current context
     */
    private void selectBestEngine() {
        AutomationCapability capability = systemEngine.getPermissionManager().getCurrentCapability();

        if (capability == AutomationCapability.SYSTEM_LEVEL) {
            activeEngine = systemEngine;
            Log.d(TAG, "Using System Engine (SYSTEM_LEVEL capability)");
        } else if (capability == AutomationCapability.ENHANCED) {
            activeEngine = systemEngine;
            Log.d(TAG, "Using System Engine (ENHANCED capability)");
        } else {
            activeEngine = accessibilityEngine;
            Log.d(TAG, "Using Accessibility Engine (BASIC capability)");
        }
    }

    // ===== Lifecycle =====

    @Override
    public void initialize(@NonNull AutomationConfig config) {
        Log.i(TAG, "Initializing Hybrid Automation Engine...");

        this.config = config;

        // Initialize accessibility engine first (always needed as fallback)
        accessibilityEngine.initialize(config);

        // Initialize system engine
        systemEngine.initialize(config);

        // Select best engine based on capabilities
        selectBestEngine();

        // Enable keep-alive if requested
        if (config.enableKeepAlive()) {
            systemEngine.enableKeepAlive();
        }

        initialized = true;

        Log.i(TAG, "Initialized - Active engine: " + activeEngine.getClass().getSimpleName());
        Log.i(TAG, "Capability: " + systemEngine.getPermissionManager().getCurrentCapability());
    }

    @Override
    public void shutdown() {
        Log.i(TAG, "Shutting down Hybrid Automation Engine...");

        accessibilityEngine.shutdown();
        systemEngine.shutdown();

        initialized = false;
    }

    // ===== Capability Detection =====

    @Override
    public boolean isAvailable() {
        return accessibilityEngine.isAvailable() || systemEngine.isAvailable();
    }

    @Override
    @NonNull
    public AutomationCapability getCapability() {
        return systemEngine.getCapability();
    }

    /**
     * Get detailed capability information
     */
    @NonNull
    public String getCapabilityReport() {
        return systemEngine.getCapabilityReport();
    }

    /**
     * Check if system-level operations are available
     */
    public boolean hasSystemLevelAccess() {
        return systemEngine.hasSystemLevelAccess();
    }

    /**
     * Refresh capabilities (e.g., after permission change)
     */
    public void refreshCapabilities() {
        systemEngine.getPermissionManager().detectCapabilities();
        selectBestEngine();
    }

    // ===== Basic Operations =====

    @Override
    @NonNull
    public AutomationResult click(int x, int y) {
        return activeEngine.click(x, y);
    }

    @Override
    @NonNull
    public AutomationResult click(@NonNull String text) {
        return activeEngine.click(text);
    }

    @Override
    @NonNull
    public AutomationResult click(@NonNull BySelector selector) {
        return activeEngine.click(selector);
    }

    @Override
    @NonNull
    public AutomationResult longClick(int x, int y) {
        return activeEngine.longClick(x, y);
    }

    @Override
    @NonNull
    public AutomationResult longClick(@NonNull String text) {
        return activeEngine.longClick(text);
    }

    @Override
    @NonNull
    public AutomationResult swipe(int fromX, int fromY, int toX, int toY, long duration) {
        return activeEngine.swipe(fromX, fromY, toX, toY, duration);
    }

    @Override
    @NonNull
    public AutomationResult swipe(@NonNull Direction direction, float distance) {
        return activeEngine.swipe(direction, distance);
    }

    @Override
    @NonNull
    public AutomationResult inputText(@NonNull String text) {
        return activeEngine.inputText(text);
    }

    @Override
    @NonNull
    public AutomationResult inputText(@NonNull BySelector selector, @NonNull String text) {
        return activeEngine.inputText(selector, text);
    }

    // ===== Advanced Operations =====

    @Override
    @NonNull
    public AutomationResult scrollFind(@NonNull BySelector selector, int maxScrolls) {
        return activeEngine.scrollFind(selector, maxScrolls);
    }

    @Override
    @NonNull
    public AutomationResult waitForElement(@NonNull BySelector selector, long timeout) {
        return activeEngine.waitForElement(selector, timeout);
    }

    @Override
    @NonNull
    public AutomationResult waitForPageStable(long timeout) {
        return activeEngine.waitForPageStable(timeout);
    }

    // ===== Query Operations =====

    @Override
    @Nullable
    public AutomationNode findElement(@NonNull BySelector selector) {
        return activeEngine.findElement(selector);
    }

    @Override
    @NonNull
    public List<AutomationNode> findElements(@NonNull BySelector selector) {
        return activeEngine.findElements(selector);
    }

    @Override
    @NonNull
    public String getPageSource() {
        return activeEngine.getPageSource();
    }

    @Override
    @Nullable
    public Bitmap takeScreenshot() {
        return activeEngine.takeScreenshot();
    }

    @Override
    @Nullable
    public String getForegroundPackage() {
        return activeEngine.getForegroundPackage();
    }

    // ===== System-Level Operations (via system engine) =====

    /**
     * Install APK
     */
    @NonNull
    public AutomationResult installApp(@NonNull String apkPath) {
        return systemEngine.installApp(apkPath);
    }

    /**
     * Uninstall app
     */
    @NonNull
    public AutomationResult uninstallApp(@NonNull String packageName) {
        return systemEngine.uninstallApp(packageName);
    }

    /**
     * Grant permission
     */
    @NonNull
    public AutomationResult grantPermission(@NonNull String packageName,
                                             @NonNull String permission) {
        return systemEngine.grantPermission(packageName, permission);
    }

    /**
     * Set secure setting
     */
    @NonNull
    public AutomationResult setSecureSetting(@NonNull String key, @NonNull String value) {
        return systemEngine.setSecureSetting(key, value);
    }

    /**
     * Enable accessibility service
     */
    @NonNull
    public AutomationResult enableAccessibilityService(@NonNull String serviceName) {
        return systemEngine.enableAccessibilityService(serviceName);
    }

    /**
     * Enable keep-alive
     */
    @NonNull
    public AutomationResult enableKeepAlive() {
        return systemEngine.enableKeepAlive();
    }

    /**
     * Disable keep-alive
     */
    @NonNull
    public AutomationResult disableKeepAlive() {
        return systemEngine.disableKeepAlive();
    }

    // ===== Callbacks =====

    @Override
    public void setListener(@Nullable AutomationListener listener) {
        this.listener = listener;
        accessibilityEngine.setListener(listener);
        systemEngine.setListener(listener);
    }

    // ===== Config =====

    @Override
    @NonNull
    public AutomationConfig getConfig() {
        return config != null ? config : AutomationConfig.getDefault();
    }

    // ===== Utility =====

    /**
     * Get active engine type
     */
    @NonNull
    public String getActiveEngineType() {
        return activeEngine.getClass().getSimpleName();
    }

    /**
     * Force use of accessibility engine
     */
    public void forceAccessibilityEngine() {
        activeEngine = accessibilityEngine;
        Log.i(TAG, "Forced to use Accessibility Engine");
    }

    /**
     * Force use of system engine
     */
    public void forceSystemEngine() {
        if (systemEngine.getPermissionManager().getCurrentCapability() != AutomationCapability.NONE) {
            activeEngine = systemEngine;
            Log.i(TAG, "Forced to use System Engine");
        } else {
            Log.w(TAG, "Cannot force System Engine - no capability available");
        }
    }

    /**
     * Auto-select best engine
     */
    public void autoSelectEngine() {
        selectBestEngine();
    }

    /**
     * Get status summary
     */
    @NonNull
    public String getStatusSummary() {
        StringBuilder sb = new StringBuilder();
        sb.append("Hybrid Automation Engine Status:\n");
        sb.append("  Initialized: ").append(initialized).append("\n");
        sb.append("  Available: ").append(isAvailable()).append("\n");
        sb.append("  Capability: ").append(getCapability()).append("\n");
        sb.append("  Active Engine: ").append(getActiveEngineType()).append("\n");
        sb.append("  System Level Access: ").append(hasSystemLevelAccess()).append("\n");
        return sb.toString();
    }
}