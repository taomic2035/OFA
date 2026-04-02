package com.ofa.agent.automation;

import android.graphics.Bitmap;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.List;

/**
 * Automation Engine interface - the core interface for UI automation.
 * Implementations provide different levels of automation capability.
 */
public interface AutomationEngine {

    // ===== Lifecycle =====

    /**
     * Initialize the engine with configuration
     * @param config Automation configuration
     */
    void initialize(@NonNull AutomationConfig config);

    /**
     * Shutdown the engine and release resources
     */
    void shutdown();

    // ===== Capability Detection =====

    /**
     * Check if engine is available and ready
     * @return true if engine can perform operations
     */
    boolean isAvailable();

    /**
     * Get engine capability level
     * @return Current capability
     */
    @NonNull
    AutomationCapability getCapability();

    // ===== Basic Click Operations =====

    /**
     * Click at screen coordinates
     * @param x X coordinate
     * @param y Y coordinate
     * @return Operation result
     */
    @NonNull
    AutomationResult click(int x, int y);

    /**
     * Click element by text
     * @param text Text to find and click
     * @return Operation result
     */
    @NonNull
    AutomationResult click(@NonNull String text);

    /**
     * Click element by selector
     * @param selector Element selector
     * @return Operation result
     */
    @NonNull
    AutomationResult click(@NonNull BySelector selector);

    // ===== Long Click Operations =====

    /**
     * Long click at screen coordinates
     * @param x X coordinate
     * @param y Y coordinate
     * @return Operation result
     */
    @NonNull
    AutomationResult longClick(int x, int y);

    /**
     * Long click element by text
     * @param text Text to find and long click
     * @return Operation result
     */
    @NonNull
    AutomationResult longClick(@NonNull String text);

    /**
     * Long click element by selector
     * @param selector Element selector
     * @return Operation result
     */
    @NonNull
    AutomationResult longClick(@NonNull BySelector selector);

    // ===== Swipe Operations =====

    /**
     * Swipe from one point to another
     * @param fromX Start X coordinate
     * @param fromY Start Y coordinate
     * @param toX End X coordinate
     * @param toY End Y coordinate
     * @param duration Duration in milliseconds
     * @return Operation result
     */
    @NonNull
    AutomationResult swipe(int fromX, int fromY, int toX, int toY, long duration);

    /**
     * Swipe in a direction
     * @param direction Direction (UP, DOWN, LEFT, RIGHT)
     * @param distance Distance in pixels (or percentage if supported)
     * @return Operation result
     */
    @NonNull
    AutomationResult swipe(@NonNull Direction direction, float distance);

    // ===== Input Operations =====

    /**
     * Input text into focused element
     * @param text Text to input
     * @return Operation result
     */
    @NonNull
    AutomationResult inputText(@NonNull String text);

    /**
     * Click at position and input text
     * @param x X coordinate to click first
     * @param y Y coordinate to click first
     * @param text Text to input
     * @return Operation result
     */
    @NonNull
    AutomationResult inputText(int x, int y, @NonNull String text);

    /**
     * Find element by selector and input text
     * @param selector Element selector
     * @param text Text to input
     * @return Operation result
     */
    @NonNull
    AutomationResult inputText(@NonNull BySelector selector, @NonNull String text);

    // ===== Scroll Operations =====

    /**
     * Scroll to find element by text
     * @param text Text to find
     * @param maxScrolls Maximum number of scrolls
     * @return Operation result
     */
    @NonNull
    AutomationResult scrollFind(@NonNull String text, int maxScrolls);

    /**
     * Scroll to find element by selector
     * @param selector Element selector
     * @param maxScrolls Maximum number of scrolls
     * @return Operation result
     */
    @NonNull
    AutomationResult scrollFind(@NonNull BySelector selector, int maxScrolls);

    // ===== Wait Operations =====

    /**
     * Wait for element to appear
     * @param selector Element selector
     * @param timeout Timeout in milliseconds
     * @return Operation result
     */
    @NonNull
    AutomationResult waitForElement(@NonNull BySelector selector, long timeout);

    /**
     * Wait for page to stabilize (no changes for a period)
     * @param timeout Timeout in milliseconds
     * @return Operation result
     */
    @NonNull
    AutomationResult waitForPageStable(long timeout);

    // ===== Query Operations =====

    /**
     * Find single element
     * @param selector Element selector
     * @return Found node or null
     */
    @Nullable
    AutomationNode findElement(@NonNull BySelector selector);

    /**
     * Find all matching elements
     * @param selector Element selector
     * @return List of found nodes
     */
    @NonNull
    List<AutomationNode> findElements(@NonNull BySelector selector);

    /**
     * Get current page source (XML/JSON representation)
     * @return Page source string
     */
    @NonNull
    String getPageSource();

    /**
     * Take screenshot
     * @return Screenshot bitmap or null
     */
    @Nullable
    Bitmap takeScreenshot();

    // ===== Callbacks =====

    /**
     * Set automation listener for events
     * @param listener Event listener
     */
    void setListener(@Nullable AutomationListener listener);

    // ===== Utility =====

    /**
     * Get screen dimensions
     * @return Screen dimensions
     */
    @NonNull
    ScreenDimension getScreenDimension();

    /**
     * Check if specific package is in foreground
     * @param packageName Package name to check
     * @return true if package is foreground
     */
    boolean isForeground(@NonNull String packageName);

    /**
     * Get current foreground package
     * @return Package name or null
     */
    @Nullable
    String getForegroundPackage();
}