package com.ofa.agent.automation.accessibility;

import android.accessibilityservice.AccessibilityServiceInfo;
import android.accessibilityservice.GestureDescription;
import android.content.Context;
import android.graphics.Bitmap;
import android.graphics.Rect;
import android.util.Log;
import android.view.accessibility.AccessibilityEvent;
import android.view.accessibility.AccessibilityManager;
import android.view.accessibility.AccessibilityNodeInfo;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationCapability;
import com.ofa.agent.automation.AutomationConfig;
import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.automation.AutomationListener;
import com.ofa.agent.automation.AutomationNode;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.automation.BySelector;
import com.ofa.agent.automation.Direction;
import com.ofa.agent.automation.ScreenDimension;

import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicReference;

/**
 * Accessibility-based automation engine.
 * Uses AccessibilityService for UI automation operations.
 */
public class AccessibilityEngine implements AutomationEngine {

    private static final String TAG = "AccessibilityEngine";

    private final Context context;
    private final AccessibilityManager accessibilityManager;
    private final NodeFinder nodeFinder;
    private final GesturePerformer gesturePerformer;

    private AutomationConfig config;
    private AutomationListener listener;
    private OFAAccessibilityService service;
    private ScreenDimension screenDimension;
    private volatile boolean initialized = false;

    // Current foreground package
    private volatile String foregroundPackage;

    /**
     * Constructor
     */
    public AccessibilityEngine(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.accessibilityManager = (AccessibilityManager)
                context.getSystemService(Context.ACCESSIBILITY_SERVICE);
        this.nodeFinder = new NodeFinder(this);
        this.gesturePerformer = new GesturePerformer(this);

        // Get screen dimensions
        android.util.DisplayMetrics metrics = context.getResources().getDisplayMetrics();
        this.screenDimension = new ScreenDimension(
                metrics.widthPixels,
                metrics.heightPixels,
                metrics.densityDpi,
                metrics.scaledDensity
        );
    }

    // ===== Service Connection =====

    /**
     * Set the connected accessibility service
     */
    public void setService(@Nullable OFAAccessibilityService service) {
        this.service = service;
        if (service != null && !initialized) {
            initialized = true;
            if (listener != null) {
                listener.onEngineAvailable(getCapability());
            }
            Log.i(TAG, "Accessibility service connected, engine ready");
        } else if (service == null && initialized) {
            initialized = false;
            if (listener != null) {
                listener.onEngineUnavailable("Accessibility service disconnected");
            }
            Log.w(TAG, "Accessibility service disconnected");
        }
    }

    /**
     * Get the accessibility service
     */
    @Nullable
    public OFAAccessibilityService getService() {
        return service;
    }

    /**
     * Get context
     */
    @NonNull
    public Context getContext() {
        return context;
    }

    /**
     * Get configuration
     */
    @Nullable
    public AutomationConfig getConfig() {
        return config;
    }

    /**
     * Update foreground package from event
     */
    public void updateForegroundPackage(@NonNull AccessibilityEvent event) {
        if (event.getPackageName() != null) {
            String newPackage = event.getPackageName().toString();
            if (!newPackage.equals(foregroundPackage)) {
                foregroundPackage = newPackage;
                if (listener != null) {
                    listener.onPageChange(newPackage, null);
                }
            }
        }
    }

    // ===== AutomationEngine Interface =====

    @Override
    public void initialize(@NonNull AutomationConfig config) {
        this.config = config;
        Log.i(TAG, "Engine initialized with config: autoRetry=" + config.isAutoRetryEnabled());
    }

    @Override
    public void shutdown() {
        initialized = false;
        service = null;
        Log.i(TAG, "Engine shutdown");
    }

    @Override
    public boolean isAvailable() {
        return initialized && service != null && service.getRootInActiveWindow() != null;
    }

    @Override
    @NonNull
    public AutomationCapability getCapability() {
        if (!isAccessibilityEnabled()) {
            return AutomationCapability.NONE;
        }

        // Check if we have enhanced capabilities
        if (service != null) {
            AccessibilityServiceInfo info = service.getServiceInfo();
            if (info != null) {
                int capabilities = info.getCapabilities();
                if ((capabilities & AccessibilityServiceInfo.CAPABILITY_CAN_PERFORM_GESTURES) != 0) {
                    if ((capabilities & AccessibilityServiceInfo.CAPABILITY_CAN_REQUEST_TOUCH_EXPLORATION) != 0) {
                        return AutomationCapability.FULL_ACCESSIBILITY;
                    }
                    return AutomationCapability.ENHANCED;
                }
            }
        }
        return AutomationCapability.BASIC;
    }

    // ===== Click Operations =====

    @Override
    @NonNull
    public AutomationResult click(int x, int y) {
        String operation = "click_coordinate";
        if (!isAvailable()) {
            return errorResult(operation, "Engine not available");
        }

        if (listener != null) {
            listener.onOperationStart(operation, x + "," + y);
        }

        long startTime = System.currentTimeMillis();
        AutomationResult result = gesturePerformer.performClick(x, y);
        long elapsed = System.currentTimeMillis() - startTime;

        if (listener != null) {
            listener.onOperationComplete(operation, result);
            listener.onGesturePerformed("click", x, y);
        }

        return result;
    }

    @Override
    @NonNull
    public AutomationResult click(@NonNull String text) {
        return click(BySelector.text(text));
    }

    @Override
    @NonNull
    public AutomationResult click(@NonNull BySelector selector) {
        String operation = "click_element";
        if (!isAvailable()) {
            return errorResult(operation, "Engine not available");
        }

        if (listener != null) {
            listener.onOperationStart(operation, selector.describe());
        }

        long startTime = System.currentTimeMillis();

        // Find element first
        AutomationNode node = findElement(selector);
        if (node == null) {
            if (listener != null) {
                listener.onElementNotFound(selector, false);
            }
            return errorResult(operation, "Element not found: " + selector.describe());
        }

        // Click at center of element
        int x = node.getCenterX();
        int y = node.getCenterY();
        AutomationResult result = gesturePerformer.performClick(x, y);

        long elapsed = System.currentTimeMillis() - startTime;

        if (listener != null) {
            listener.onElementFound(selector, node);
            listener.onOperationComplete(operation, result);
            listener.onGesturePerformed("click", x, y);
        }

        return result;
    }

    // ===== Long Click Operations =====

    @Override
    @NonNull
    public AutomationResult longClick(int x, int y) {
        String operation = "longClick_coordinate";
        if (!isAvailable()) {
            return errorResult(operation, "Engine not available");
        }

        if (listener != null) {
            listener.onOperationStart(operation, x + "," + y);
        }

        long startTime = System.currentTimeMillis();
        long duration = config != null ? config.getLongClickDuration() : AutomationConfig.DEFAULT_LONG_CLICK_DURATION;
        AutomationResult result = gesturePerformer.performLongClick(x, y, duration);
        long elapsed = System.currentTimeMillis() - startTime;

        if (listener != null) {
            listener.onOperationComplete(operation, result);
            listener.onGesturePerformed("longClick", x, y);
        }

        return result;
    }

    @Override
    @NonNull
    public AutomationResult longClick(@NonNull String text) {
        return longClick(BySelector.text(text));
    }

    @Override
    @NonNull
    public AutomationResult longClick(@NonNull BySelector selector) {
        String operation = "longClick_element";
        if (!isAvailable()) {
            return errorResult(operation, "Engine not available");
        }

        AutomationNode node = findElement(selector);
        if (node == null) {
            return errorResult(operation, "Element not found: " + selector.describe());
        }

        return longClick(node.getCenterX(), node.getCenterY());
    }

    // ===== Swipe Operations =====

    @Override
    @NonNull
    public AutomationResult swipe(int fromX, int fromY, int toX, int toY, long duration) {
        String operation = "swipe";
        if (!isAvailable()) {
            return errorResult(operation, "Engine not available");
        }

        if (listener != null) {
            listener.onOperationStart(operation, fromX + "," + fromY + " -> " + toX + "," + toY);
        }

        long startTime = System.currentTimeMillis();
        AutomationResult result = gesturePerformer.performSwipe(fromX, fromY, toX, toY, duration);
        long elapsed = System.currentTimeMillis() - startTime;

        if (listener != null) {
            listener.onOperationComplete(operation, result);
            listener.onGesturePerformed("swipe", fromX, fromY);
        }

        return result;
    }

    @Override
    @NonNull
    public AutomationResult swipe(@NonNull Direction direction, float distance) {
        int centerX = screenDimension.getCenterX();
        int centerY = screenDimension.getCenterY();

        int scrollDistance = config != null ?
                config.getScrollDistance() : AutomationConfig.DEFAULT_SCROLL_DISTANCE;
        int pixels = distance > 0 ? (int) distance : scrollDistance;

        int fromX, fromY, toX, toY;

        switch (direction) {
            case UP:
                fromX = centerX;
                fromY = centerY + pixels / 2;
                toX = centerX;
                toY = centerY - pixels / 2;
                break;
            case DOWN:
                fromX = centerX;
                fromY = centerY - pixels / 2;
                toX = centerX;
                toY = centerY + pixels / 2;
                break;
            case LEFT:
                fromX = centerX + pixels / 2;
                fromY = centerY;
                toX = centerX - pixels / 2;
                toY = centerY;
                break;
            case RIGHT:
                fromX = centerX - pixels / 2;
                fromY = centerY;
                toX = centerX + pixels / 2;
                toY = centerY;
                break;
            default:
                fromX = centerX;
                fromY = centerY;
                toX = centerX;
                toY = centerY;
        }

        long duration = config != null ? config.getSwipeDuration() : AutomationConfig.DEFAULT_SWIPE_DURATION;
        return swipe(fromX, fromY, toX, toY, duration);
    }

    // ===== Input Operations =====

    @Override
    @NonNull
    public AutomationResult inputText(@NonNull String text) {
        String operation = "input_text";
        if (!isAvailable()) {
            return errorResult(operation, "Engine not available");
        }

        if (listener != null) {
            listener.onOperationStart(operation, text);
        }

        long startTime = System.currentTimeMillis();
        AutomationResult result = gesturePerformer.performInput(text);
        long elapsed = System.currentTimeMillis() - startTime;

        if (listener != null) {
            listener.onOperationComplete(operation, result);
        }

        return result;
    }

    @Override
    @NonNull
    public AutomationResult inputText(int x, int y, @NonNull String text) {
        // Click first to focus
        AutomationResult clickResult = click(x, y);
        if (!clickResult.isSuccess()) {
            return clickResult;
        }

        // Small delay for focus
        try { Thread.sleep(200); } catch (Exception e) {}

        return inputText(text);
    }

    @Override
    @NonNull
    public AutomationResult inputText(@NonNull BySelector selector, @NonNull String text) {
        AutomationNode node = findElement(selector);
        if (node == null) {
            return errorResult("input_text", "Element not found: " + selector.describe());
        }

        return inputText(node.getCenterX(), node.getCenterY(), text);
    }

    // ===== Scroll Operations =====

    @Override
    @NonNull
    public AutomationResult scrollFind(@NonNull String text, int maxScrolls) {
        return scrollFind(BySelector.text(text), maxScrolls);
    }

    @Override
    @NonNull
    public AutomationResult scrollFind(@NonNull BySelector selector, int maxScrolls) {
        String operation = "scroll_find";
        if (!isAvailable()) {
            return errorResult(operation, "Engine not available");
        }

        if (listener != null) {
            listener.onOperationStart(operation, selector.describe());
        }

        long startTime = System.currentTimeMillis();

        for (int i = 0; i < maxScrolls; i++) {
            // Check if element exists
            AutomationNode node = findElement(selector);
            if (node != null) {
                long elapsed = System.currentTimeMillis() - startTime;
                if (listener != null) {
                    listener.onElementFound(selector, node);
                    listener.onOperationComplete(operation,
                            new AutomationResult(operation, node, elapsed));
                }
                return new AutomationResult(operation, node, elapsed);
            }

            // Scroll down
            AutomationResult scrollResult = swipe(Direction.DOWN, 0);
            if (!scrollResult.isSuccess()) {
                return scrollResult;
            }

            // Wait for scroll to complete
            try { Thread.sleep(500); } catch (Exception e) {}
        }

        long elapsed = System.currentTimeMillis() - startTime;
        if (listener != null) {
            listener.onElementNotFound(selector, true);
        }
        return errorResult(operation, "Element not found after " + maxScrolls + " scrolls");
    }

    // ===== Wait Operations =====

    @Override
    @NonNull
    public AutomationResult waitForElement(@NonNull BySelector selector, long timeout) {
        String operation = "wait_for_element";
        if (!isAvailable()) {
            return errorResult(operation, "Engine not available");
        }

        long startTime = System.currentTimeMillis();
        long deadline = startTime + timeout;

        while (System.currentTimeMillis() < deadline) {
            AutomationNode node = findElement(selector);
            if (node != null) {
                long elapsed = System.currentTimeMillis() - startTime;
                return new AutomationResult(operation, node, elapsed);
            }

            try { Thread.sleep(200); } catch (Exception e) {}
        }

        return errorResult(operation, "Element not found within timeout");
    }

    @Override
    @NonNull
    public AutomationResult waitForPageStable(long timeout) {
        String operation = "wait_for_page_stable";
        if (!isAvailable()) {
            return errorResult(operation, "Engine not available");
        }

        // Simple implementation: wait for no layout changes
        long startTime = System.currentTimeMillis();
        String lastSource = getPageSource();

        while (System.currentTimeMillis() - startTime < timeout) {
            try { Thread.sleep(500); } catch (Exception e) {}
            String currentSource = getPageSource();
            if (currentSource.equals(lastSource)) {
                long elapsed = System.currentTimeMillis() - startTime;
                JSONObject data = new JSONObject();
                try { data.put("stable", true); } catch (Exception e) {}
                return new AutomationResult(operation, data, elapsed);
            }
            lastSource = currentSource;
        }

        return errorResult(operation, "Page not stable within timeout");
    }

    // ===== Query Operations =====

    @Override
    @Nullable
    public AutomationNode findElement(@NonNull BySelector selector) {
        if (!isAvailable()) {
            return null;
        }
        return nodeFinder.findElement(selector);
    }

    @Override
    @NonNull
    public List<AutomationNode> findElements(@NonNull BySelector selector) {
        if (!isAvailable()) {
            return new ArrayList<>();
        }
        return nodeFinder.findElements(selector);
    }

    @Override
    @NonNull
    public String getPageSource() {
        if (!isAvailable()) {
            return "";
        }
        return nodeFinder.getPageSource();
    }

    @Override
    @Nullable
    public Bitmap takeScreenshot() {
        // Accessibility service cannot take screenshots directly
        // This would require additional implementation or MediaProjection API
        Log.w(TAG, "Screenshot not available via AccessibilityService");
        return null;
    }

    // ===== Callbacks =====

    @Override
    public void setListener(@Nullable AutomationListener listener) {
        this.listener = listener;
    }

    // ===== Utility =====

    @Override
    @NonNull
    public ScreenDimension getScreenDimension() {
        return screenDimension;
    }

    @Override
    public boolean isForeground(@NonNull String packageName) {
        return packageName.equals(foregroundPackage);
    }

    @Override
    @Nullable
    public String getForegroundPackage() {
        return foregroundPackage;
    }

    // ===== Helper Methods =====

    /**
     * Check if accessibility service is enabled
     */
    public boolean isAccessibilityEnabled() {
        return accessibilityManager.isEnabled();
    }

    /**
     * Create error result
     */
    @NonNull
    private AutomationResult errorResult(@NonNull String operation, @NonNull String error) {
        if (listener != null) {
            listener.onOperationError(operation, error,
                    config != null && config.isAutoRetryEnabled());
        }
        return new AutomationResult(operation, error);
    }
}