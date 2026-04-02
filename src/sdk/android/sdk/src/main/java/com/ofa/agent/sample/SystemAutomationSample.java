package com.ofa.agent.sample;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;

import com.ofa.agent.automation.AutomationCapability;
import com.ofa.agent.automation.AutomationConfig;
import com.ofa.agent.automation.hybrid.HybridAutomationEngine;
import com.ofa.agent.automation.system.SystemAutomationEngine;
import com.ofa.agent.automation.system.SystemPermissionManager;
import com.ofa.agent.automation.system.KeepAliveManager;
import com.ofa.agent.automation.system.SilentInstaller;

/**
 * System Automation Sample - demonstrates ROM System Layer usage.
 *
 * This sample shows how to:
 * 1. Initialize the Hybrid Automation Engine
 * 2. Detect system capabilities
 * 3. Use system-level operations (if available)
 * 4. Handle graceful degradation
 */
public class SystemAutomationSample {

    private static final String TAG = "SystemAutomationSample";

    private final Context context;
    private HybridAutomationEngine engine;

    /**
     * Create sample
     */
    public SystemAutomationSample(@NonNull Context context) {
        this.context = context;
    }

    /**
     * Initialize the hybrid automation engine
     */
    public void initialize() {
        Log.i(TAG, "Initializing Hybrid Automation Engine...");

        // Create hybrid engine
        engine = new HybridAutomationEngine(context);

        // Configure with keep-alive enabled
        AutomationConfig config = AutomationConfig.builder()
            .enableKeepAlive(true)
            .enableLogging(true)
            .build();

        engine.initialize(config);

        // Log capability report
        Log.i(TAG, engine.getCapabilityReport());
    }

    /**
     * Demonstrate capability detection
     */
    public void demonstrateCapabilityDetection() {
        if (engine == null) {
            Log.w(TAG, "Engine not initialized");
            return;
        }

        Log.i(TAG, "=== Capability Detection ===");

        AutomationCapability capability = engine.getCapability();
        Log.i(TAG, "Current capability: " + capability.name());
        Log.i(TAG, "Has system level access: " + engine.hasSystemLevelAccess());
        Log.i(TAG, "Active engine: " + engine.getActiveEngineType());

        SystemAutomationEngine systemEngine = engine.getSystemEngine();
        SystemPermissionManager permManager = systemEngine.getPermissionManager();

        // Check specific permissions
        Log.i(TAG, "Can silent install: " + permManager.canSilentInstall());
        Log.i(TAG, "Can modify secure settings: " + permManager.canModifySecureSettings());
        Log.i(TAG, "Is system app: " + permManager.isSystemApp());
        Log.i(TAG, "Has root access: " + permManager.checkRootAccess());
    }

    /**
     * Demonstrate keep-alive strategies
     */
    public void demonstrateKeepAlive() {
        if (engine == null) {
            Log.w(TAG, "Engine not initialized");
            return;
        }

        Log.i(TAG, "=== Keep-Alive Strategies ===");

        SystemAutomationEngine systemEngine = engine.getSystemEngine();
        KeepAliveManager keepAliveManager = systemEngine.getKeepAliveManager();

        // Show available strategies
        Log.i(TAG, "Available strategies:");
        for (KeepAliveManager.KeepAliveStrategy strategy : keepAliveManager.getAvailableStrategies()) {
            Log.i(TAG, "  - " + strategy.name + ": " + strategy.description);
        }

        // Enable keep-alive
        engine.enableKeepAlive();
    }

    /**
     * Demonstrate silent installation (if capable)
     */
    public void demonstrateSilentInstall(@NonNull String apkPath) {
        if (engine == null) {
            Log.w(TAG, "Engine not initialized");
            return;
        }

        Log.i(TAG, "=== Silent Install Demo ===");

        if (!engine.hasSystemLevelAccess()) {
            Log.w(TAG, "System-level access not available - will use user-guided installation");
        }

        SystemAutomationEngine systemEngine = engine.getSystemEngine();
        SilentInstaller installer = systemEngine.getSilentInstaller();

        // Check capability
        if (installer.canInstallSilently()) {
            Log.i(TAG, "Silent installation is available");
        } else {
            Log.i(TAG, "Silent installation not available - will prompt user");
        }

        // Perform installation
        SilentInstaller.InstallResult result = installer.install(apkPath);
        Log.i(TAG, "Install result: success=" + result.success + ", method=" + result.methodUsed);
    }

    /**
     * Demonstrate permission granting (if capable)
     */
    public void demonstrateGrantPermission(@NonNull String packageName, @NonNull String permission) {
        if (engine == null) {
            Log.w(TAG, "Engine not initialized");
            return;
        }

        Log.i(TAG, "=== Grant Permission Demo ===");
        Log.i(TAG, "Granting " + permission + " to " + packageName);

        if (!engine.hasSystemLevelAccess()) {
            Log.w(TAG, "System-level access required for silent permission grant");
            return;
        }

        SystemAutomationEngine systemEngine = engine.getSystemEngine();
        var result = systemEngine.grantPermission(packageName, permission);
        Log.i(TAG, "Grant result: " + (result.isSuccess() ? "success" : result.getMessage()));
    }

    /**
     * Demonstrate secure settings modification (if capable)
     */
    public void demonstrateSecureSetting(@NonNull String key, @NonNull String value) {
        if (engine == null) {
            Log.w(TAG, "Engine not initialized");
            return;
        }

        Log.i(TAG, "=== Secure Setting Demo ===");
        Log.i(TAG, "Setting " + key + " = " + value);

        SystemAutomationEngine systemEngine = engine.getSystemEngine();

        if (!systemEngine.getPermissionManager().canModifySecureSettings()) {
            Log.w(TAG, "Secure settings modification not available");
            return;
        }

        var result = systemEngine.setSecureSetting(key, value);
        Log.i(TAG, "Set result: " + (result.isSuccess() ? "success" : result.getMessage()));
    }

    /**
     * Demonstrate accessibility service enablement (if capable)
     */
    public void demonstrateEnableAccessibility(@NonNull String serviceName) {
        if (engine == null) {
            Log.w(TAG, "Engine not initialized");
            return;
        }

        Log.i(TAG, "=== Enable Accessibility Demo ===");

        SystemAutomationEngine systemEngine = engine.getSystemEngine();

        // Check if already enabled
        if (systemEngine.getPermissionManager().isSpecificAccessibilityEnabled(serviceName)) {
            Log.i(TAG, "Accessibility service already enabled: " + serviceName);
            return;
        }

        // Try to enable
        var result = systemEngine.enableAccessibilityService(serviceName);
        Log.i(TAG, "Enable result: " + (result.isSuccess() ? "success" : result.getMessage()));
    }

    /**
     * Shutdown engine
     */
    public void shutdown() {
        if (engine != null) {
            Log.i(TAG, "Shutting down engine...");
            engine.shutdown();
        }
    }

    /**
     * Get full status report
     */
    @NonNull
    public String getStatusReport() {
        if (engine == null) {
            return "Engine not initialized";
        }
        return engine.getStatusSummary();
    }

    /**
     * Run all demonstrations
     */
    public void runAllDemos() {
        Log.i(TAG, "\n========== System Automation Demo ==========\n");

        initialize();
        demonstrateCapabilityDetection();
        demonstrateKeepAlive();

        // These require system-level access
        // demonstrateSilentInstall("/path/to/app.apk");
        // demonstrateGrantPermission("com.example.app", "android.permission.READ_CONTACTS");
        // demonstrateSecureSetting("screen_brightness", "128");
        // demonstrateEnableAccessibility("com.ofa.agent/.automation.accessibility.OFAAccessibilityService");

        Log.i(TAG, "\n==========================================\n");
        Log.i(TAG, getStatusReport());
    }
}