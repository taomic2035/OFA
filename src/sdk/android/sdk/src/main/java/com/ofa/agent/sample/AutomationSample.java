package com.ofa.agent.sample;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;

import com.ofa.agent.automation.AutomationCapability;
import com.ofa.agent.automation.AutomationConfig;
import com.ofa.agent.automation.AutomationListener;
import com.ofa.agent.automation.AutomationManager;
import com.ofa.agent.automation.AutomationNode;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.automation.BySelector;
import com.ofa.agent.automation.Direction;

/**
 * Sample demonstrating UI Automation usage.
 *
 * Prerequisites:
 * 1. User must enable accessibility service in Settings > Accessibility
 * 2. The accessibility service must be declared in AndroidManifest.xml
 *
 * Example usage:
 * <code>
 *     AutomationSample sample = new AutomationSample(context);
 *     sample.start();
 *     sample.clickOnText("Button");
 * </code>
 */
public class AutomationSample implements AutomationListener {

    private static final String TAG = "AutomationSample";

    private final Context context;
    private final AutomationManager automationManager;
    private volatile boolean ready = false;

    public AutomationSample(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.automationManager = AutomationManager.init(context);
    }

    /**
     * Start the automation manager
     */
    public void start() {
        Log.i(TAG, "Starting automation manager...");

        // Configure automation
        AutomationConfig config = AutomationConfig.builder()
                .clickTimeout(5000)
                .swipeTimeout(3000)
                .autoRetry(true)
                .maxRetries(3)
                .enableLogging(true)
                .build();

        automationManager.setListener(this);
        automationManager.start(config);
    }

    /**
     * Stop the automation manager
     */
    public void stop() {
        Log.i(TAG, "Stopping automation manager...");
        automationManager.stop();
        ready = false;
    }

    /**
     * Check if automation is ready
     */
    public boolean isReady() {
        return ready && automationManager.isAvailable();
    }

    /**
     * Open accessibility settings for user to enable service
     */
    public void openAccessibilitySettings() {
        automationManager.openAccessibilitySettings();
    }

    /**
     * Click on element by text
     */
    public void clickOnText(@NonNull String text) {
        if (!isReady()) {
            Log.w(TAG, "Automation not ready");
            return;
        }

        Log.i(TAG, "Clicking on text: " + text);
        AutomationResult result = automationManager.getEngine().click(text);

        if (result.isSuccess()) {
            Log.i(TAG, "Click successful: " + result.getExecutionTimeMs() + "ms");
        } else {
            Log.e(TAG, "Click failed: " + result.getError());
        }
    }

    /**
     * Click at coordinates
     */
    public void clickAt(int x, int y) {
        if (!isReady()) {
            Log.w(TAG, "Automation not ready");
            return;
        }

        Log.i(TAG, "Clicking at: " + x + ", " + y);
        AutomationResult result = automationManager.getEngine().click(x, y);

        if (result.isSuccess()) {
            Log.i(TAG, "Click successful: " + result.getExecutionTimeMs() + "ms");
        } else {
            Log.e(TAG, "Click failed: " + result.getError());
        }
    }

    /**
     * Swipe in a direction
     */
    public void swipe(@NonNull Direction direction) {
        if (!isReady()) {
            Log.w(TAG, "Automation not ready");
            return;
        }

        Log.i(TAG, "Swiping: " + direction);
        AutomationResult result = automationManager.getEngine().swipe(direction, 0);

        if (result.isSuccess()) {
            Log.i(TAG, "Swipe successful: " + result.getExecutionTimeMs() + "ms");
        } else {
            Log.e(TAG, "Swipe failed: " + result.getError());
        }
    }

    /**
     * Input text into focused element
     */
    public void inputText(@NonNull String text) {
        if (!isReady()) {
            Log.w(TAG, "Automation not ready");
            return;
        }

        Log.i(TAG, "Inputting text: " + text);
        AutomationResult result = automationManager.getEngine().inputText(text);

        if (result.isSuccess()) {
            Log.i(TAG, "Input successful: " + result.getExecutionTimeMs() + "ms");
        } else {
            Log.e(TAG, "Input failed: " + result.getError());
        }
    }

    /**
     * Find element by text
     */
    public AutomationNode findElement(@NonNull String text) {
        if (!isReady()) {
            Log.w(TAG, "Automation not ready");
            return null;
        }

        BySelector selector = BySelector.text(text);
        return automationManager.getEngine().findElement(selector);
    }

    /**
     * Scroll to find element
     */
    public void scrollFind(@NonNull String text) {
        if (!isReady()) {
            Log.w(TAG, "Automation not ready");
            return;
        }

        Log.i(TAG, "Scrolling to find: " + text);
        AutomationResult result = automationManager.getEngine().scrollFind(text, 10);

        if (result.isSuccess()) {
            Log.i(TAG, "Element found: " + result.getFoundNode());
        } else {
            Log.e(TAG, "Element not found: " + result.getError());
        }
    }

    /**
     * Wait for element to appear
     */
    public boolean waitForElement(@NonNull String text, long timeoutMs) {
        if (!isReady()) {
            Log.w(TAG, "Automation not ready");
            return false;
        }

        BySelector selector = BySelector.text(text);
        AutomationResult result = automationManager.getEngine().waitForElement(selector, timeoutMs);

        return result.isSuccess();
    }

    /**
     * Get current foreground package
     */
    public String getForegroundPackage() {
        return automationManager.getForegroundPackage();
    }

    // ===== AutomationListener Implementation =====

    @Override
    public void onEngineAvailable(@NonNull AutomationCapability capability) {
        Log.i(TAG, "Engine available with capability: " + capability.getDescription());
        ready = true;
    }

    @Override
    public void onEngineUnavailable(@NonNull String reason) {
        Log.w(TAG, "Engine unavailable: " + reason);
        ready = false;
    }

    @Override
    public void onOperationStart(@NonNull String operation, String target) {
        Log.d(TAG, "Operation started: " + operation + " -> " + target);
    }

    @Override
    public void onOperationComplete(@NonNull String operation, @NonNull AutomationResult result) {
        Log.d(TAG, "Operation completed: " + operation + " in " + result.getExecutionTimeMs() + "ms");
    }

    @Override
    public void onOperationError(@NonNull String operation, @NonNull String error, boolean willRetry) {
        Log.e(TAG, "Operation error: " + operation + " - " + error + ", retry: " + willRetry);
    }

    @Override
    public void onGesturePerformed(@NonNull String gestureType, int x, int y) {
        Log.d(TAG, "Gesture performed: " + gestureType + " at " + x + "," + y);
    }

    @Override
    public void onElementFound(@NonNull BySelector selector, @NonNull AutomationNode node) {
        Log.d(TAG, "Element found: " + selector.describe() + " -> " + node);
    }

    @Override
    public void onElementNotFound(@NonNull BySelector selector, boolean timedOut) {
        Log.w(TAG, "Element not found: " + selector.describe() + ", timeout: " + timedOut);
    }

    @Override
    public void onPageChange(String packageName, String activityName) {
        Log.d(TAG, "Page changed: " + packageName + " / " + activityName);
    }

    @Override
    public void onAccessibilityServiceStateChanged(boolean enabled) {
        Log.i(TAG, "Accessibility service state: " + (enabled ? "enabled" : "disabled"));
        if (!enabled) {
            ready = false;
        }
    }

    @Override
    public void onScreenshotCaptured(String screenshotPath) {
        Log.d(TAG, "Screenshot captured: " + screenshotPath);
    }
}