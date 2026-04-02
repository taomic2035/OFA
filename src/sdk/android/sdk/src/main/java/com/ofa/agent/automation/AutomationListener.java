package com.ofa.agent.automation;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

/**
 * Automation event listener interface.
 */
public interface AutomationListener {

    /**
     * Called when automation engine becomes available
     * @param capability Available capability level
     */
    void onEngineAvailable(@NonNull AutomationCapability capability);

    /**
     * Called when automation engine becomes unavailable
     * @param reason Reason for unavailability
     */
    void onEngineUnavailable(@NonNull String reason);

    /**
     * Called before an operation starts
     * @param operation Operation name
     * @param target Target description (may be null for coordinate-based operations)
     */
    void onOperationStart(@NonNull String operation, @Nullable String target);

    /**
     * Called when an operation completes
     * @param operation Operation name
     * @param result Operation result
     */
    void onOperationComplete(@NonNull String operation, @NonNull AutomationResult result);

    /**
     * Called when an operation fails
     * @param operation Operation name
     * @param error Error message
     * @param willRetry Whether operation will be retried
     */
    void onOperationError(@NonNull String operation, @NonNull String error, boolean willRetry);

    /**
     * Called when a gesture is performed
     * @param gestureType Gesture type (click, swipe, etc.)
     * @param x X coordinate or start X
     * @param y Y coordinate or start Y
     */
    void onGesturePerformed(@NonNull String gestureType, int x, int y);

    /**
     * Called when element is found
     * @param selector Selector used
     * @param node Found node
     */
    void onElementFound(@NonNull BySelector selector, @NonNull AutomationNode node);

    /**
     * Called when element is not found
     * @param selector Selector used
     * @param timedOut Whether search timed out
     */
    void onElementNotFound(@NonNull BySelector selector, boolean timedOut);

    /**
     * Called when page changes
     * @param packageName Package name of new foreground app
     * @param activityName Activity name (if available)
     */
    void onPageChange(@Nullable String packageName, @Nullable String activityName);

    /**
     * Called when accessibility service state changes
     * @param enabled Whether service is enabled
     */
    void onAccessibilityServiceStateChanged(boolean enabled);

    /**
     * Called when a screenshot is captured (if enabled)
     * @param screenshotPath Path to saved screenshot or null if in-memory only
     */
    void onScreenshotCaptured(@Nullable String screenshotPath);
}