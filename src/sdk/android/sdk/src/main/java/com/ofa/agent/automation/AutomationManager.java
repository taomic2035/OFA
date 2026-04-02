package com.ofa.agent.automation;

import android.content.Context;
import android.content.Intent;
import android.provider.Settings;
import android.util.Log;
import android.view.accessibility.AccessibilityManager;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.accessibility.AccessibilityEngine;
import com.ofa.agent.automation.accessibility.OFAAccessibilityService;

/**
 * Automation Manager - central manager for UI automation.
 * Manages accessibility service connection and provides unified API.
 */
public class AutomationManager implements AutomationListener {

    private static final String TAG = "AutomationManager";

    private static volatile AutomationManager instance;

    private final Context context;
    private final AccessibilityManager accessibilityManager;
    private final AccessibilityEngine engine;
    private AutomationConfig config;
    private AutomationListener externalListener;
    private volatile boolean initialized = false;

    /**
     * Get singleton instance
     */
    @Nullable
    public static AutomationManager getInstance() {
        return instance;
    }

    /**
     * Initialize automation manager
     */
    @NonNull
    public static AutomationManager init(@NonNull Context context) {
        if (instance == null) {
            synchronized (AutomationManager.class) {
                if (instance == null) {
                    instance = new AutomationManager(context.getApplicationContext());
                }
            }
        }
        return instance;
    }

    /**
     * Constructor (use init() instead)
     */
    private AutomationManager(@NonNull Context context) {
        this.context = context;
        this.accessibilityManager = (AccessibilityManager)
                context.getSystemService(Context.ACCESSIBILITY_SERVICE);
        this.engine = new AccessibilityEngine(context);
        this.engine.setListener(this);
    }

    /**
     * Start automation manager
     */
    public void start(@NonNull AutomationConfig config) {
        this.config = config;
        engine.initialize(config);

        // Check if accessibility service is enabled
        if (!isAccessibilityServiceEnabled()) {
            Log.w(TAG, "Accessibility service not enabled, requesting user to enable");
            if (externalListener != null) {
                externalListener.onEngineUnavailable("Accessibility service not enabled");
            }
        }

        // Try to connect to service
        OFAAccessibilityService service = OFAAccessibilityService.getInstance();
        if (service != null) {
            service.setEngine(engine);
            initialized = true;
            Log.i(TAG, "Connected to existing accessibility service");
        } else {
            Log.i(TAG, "Waiting for accessibility service to start");
        }
    }

    /**
     * Stop automation manager
     */
    public void stop() {
        initialized = false;
        engine.shutdown();
        Log.i(TAG, "Automation manager stopped");
    }

    /**
     * Get automation engine
     */
    @NonNull
    public AutomationEngine getEngine() {
        return engine;
    }

    /**
     * Check if automation is available
     */
    public boolean isAvailable() {
        return initialized && engine.isAvailable();
    }

    /**
     * Check if accessibility service is enabled
     */
    public boolean isAccessibilityServiceEnabled() {
        return accessibilityManager.isEnabled();
    }

    /**
     * Open accessibility settings
     */
    public void openAccessibilitySettings() {
        Intent intent = new Intent(Settings.ACTION_ACCESSIBILITY_SETTINGS);
        intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK);
        context.startActivity(intent);
    }

    /**
     * Set external listener
     */
    public void setListener(@Nullable AutomationListener listener) {
        this.externalListener = listener;
    }

    /**
     * Get configuration
     */
    @Nullable
    public AutomationConfig getConfig() {
        return config;
    }

    // ===== AutomationListener Implementation =====

    @Override
    public void onEngineAvailable(@NonNull AutomationCapability capability) {
        Log.i(TAG, "Engine available with capability: " + capability.getDescription());
        initialized = true;
        if (externalListener != null) {
            externalListener.onEngineAvailable(capability);
        }
    }

    @Override
    public void onEngineUnavailable(@NonNull String reason) {
        Log.w(TAG, "Engine unavailable: " + reason);
        initialized = false;
        if (externalListener != null) {
            externalListener.onEngineUnavailable(reason);
        }
    }

    @Override
    public void onOperationStart(@NonNull String operation, @Nullable String target) {
        if (externalListener != null) {
            externalListener.onOperationStart(operation, target);
        }
    }

    @Override
    public void onOperationComplete(@NonNull String operation, @NonNull AutomationResult result) {
        if (externalListener != null) {
            externalListener.onOperationComplete(operation, result);
        }
    }

    @Override
    public void onOperationError(@NonNull String operation, @NonNull String error, boolean willRetry) {
        Log.e(TAG, "Operation error: " + operation + " - " + error);
        if (externalListener != null) {
            externalListener.onOperationError(operation, error, willRetry);
        }
    }

    @Override
    public void onGesturePerformed(@NonNull String gestureType, int x, int y) {
        if (externalListener != null) {
            externalListener.onGesturePerformed(gestureType, x, y);
        }
    }

    @Override
    public void onElementFound(@NonNull BySelector selector, @NonNull AutomationNode node) {
        if (externalListener != null) {
            externalListener.onElementFound(selector, node);
        }
    }

    @Override
    public void onElementNotFound(@NonNull BySelector selector, boolean timedOut) {
        if (externalListener != null) {
            externalListener.onElementNotFound(selector, timedOut);
        }
    }

    @Override
    public void onPageChange(@Nullable String packageName, @Nullable String activityName) {
        if (externalListener != null) {
            externalListener.onPageChange(packageName, activityName);
        }
    }

    @Override
    public void onAccessibilityServiceStateChanged(boolean enabled) {
        Log.i(TAG, "Accessibility service state changed: " + enabled);
        if (externalListener != null) {
            externalListener.onAccessibilityServiceStateChanged(enabled);
        }
    }

    @Override
    public void onScreenshotCaptured(@Nullable String screenshotPath) {
        if (externalListener != null) {
            externalListener.onScreenshotCaptured(screenshotPath);
        }
    }

    // ===== Utility Methods =====

    /**
     * Get screen dimensions
     */
    @NonNull
    public ScreenDimension getScreenDimension() {
        return engine.getScreenDimension();
    }

    /**
     * Check if specific package is in foreground
     */
    public boolean isForeground(@NonNull String packageName) {
        return engine.isForeground(packageName);
    }

    /**
     * Get current foreground package
     */
    @Nullable
    public String getForegroundPackage() {
        return engine.getForegroundPackage();
    }

    /**
     * Get capability level
     */
    @NonNull
    public AutomationCapability getCapability() {
        return engine.getCapability();
    }
}