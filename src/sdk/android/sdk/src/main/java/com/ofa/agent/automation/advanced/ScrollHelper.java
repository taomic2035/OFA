package com.ofa.agent.automation.advanced;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.automation.AutomationNode;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.automation.BySelector;
import com.ofa.agent.automation.Direction;

/**
 * Scroll Helper - advanced scrolling operations.
 * Provides smart scrolling with detection of scroll boundaries and content changes.
 */
public class ScrollHelper {

    private static final String TAG = "ScrollHelper";

    private final AutomationEngine engine;
    private int maxScrollAttempts = 20;
    private long scrollDelay = 500; // ms between scrolls
    private float scrollDistance = 0.7f; // percentage of screen height

    public ScrollHelper(@NonNull AutomationEngine engine) {
        this.engine = engine;
    }

    /**
     * Set maximum scroll attempts
     */
    public void setMaxScrollAttempts(int maxAttempts) {
        this.maxScrollAttempts = maxAttempts;
    }

    /**
     * Set delay between scrolls in milliseconds
     */
    public void setScrollDelay(long delayMs) {
        this.scrollDelay = delayMs;
    }

    /**
     * Set scroll distance as percentage of screen height (0.0 - 1.0)
     */
    public void setScrollDistance(float percentage) {
        this.scrollDistance = Math.max(0.1f, Math.min(1.0f, percentage));
    }

    /**
     * Scroll to find element by selector
     * @param selector Element selector
     * @param direction Scroll direction (UP to find elements below, DOWN to find above)
     * @return Result with found node or error
     */
    @NonNull
    public AutomationResult scrollFind(@NonNull BySelector selector, @NonNull Direction direction) {
        return scrollFind(selector, direction, maxScrollAttempts);
    }

    /**
     * Scroll to find element with custom max attempts
     */
    @NonNull
    public AutomationResult scrollFind(@NonNull BySelector selector,
                                        @NonNull Direction direction,
                                        int maxAttempts) {
        Log.i(TAG, "Starting scrollFind: " + selector.describe() + " direction=" + direction);

        // Check if already visible
        AutomationNode node = engine.findElement(selector);
        if (node != null) {
            Log.d(TAG, "Element already visible");
            return new AutomationResult("scrollFind", node, 0);
        }

        String previousSource = null;
        int noChangeCount = 0;
        int maxNoChange = 3; // Stop after 3 consecutive no-change scrolls

        for (int i = 0; i < maxAttempts; i++) {
            // Perform scroll
            AutomationResult scrollResult = engine.swipe(direction, 0);
            if (!scrollResult.isSuccess()) {
                return new AutomationResult("scrollFind",
                    "Scroll failed: " + scrollResult.getError());
            }

            // Wait for scroll to settle
            sleep(scrollDelay);

            // Check if element is now visible
            node = engine.findElement(selector);
            if (node != null) {
                Log.i(TAG, "Element found after " + (i + 1) + " scrolls");
                return new AutomationResult("scrollFind", node, 0);
            }

            // Detect if page changed (to know if we hit the end)
            String currentSource = engine.getPageSource();
            if (previousSource != null && currentSource.equals(previousSource)) {
                noChangeCount++;
                if (noChangeCount >= maxNoChange) {
                    Log.w(TAG, "No page change after " + maxNoChange + " scrolls, reached boundary");
                    return new AutomationResult("scrollFind",
                        "Element not found - reached scroll boundary");
                }
            } else {
                noChangeCount = 0;
            }
            previousSource = currentSource;
        }

        Log.w(TAG, "Element not found after " + maxAttempts + " scrolls");
        return new AutomationResult("scrollFind",
            "Element not found after " + maxAttempts + " scroll attempts");
    }

    /**
     * Scroll to top of list
     */
    @NonNull
    public AutomationResult scrollToTop() {
        Log.i(TAG, "Scrolling to top");

        String previousSource = null;
        int noChangeCount = 0;

        for (int i = 0; i < maxScrollAttempts; i++) {
            AutomationResult result = engine.swipe(Direction.DOWN, 0);
            if (!result.isSuccess()) {
                return result;
            }

            sleep(scrollDelay);

            String currentSource = engine.getPageSource();
            if (previousSource != null && currentSource.equals(previousSource)) {
                noChangeCount++;
                if (noChangeCount >= 2) {
                    Log.i(TAG, "Reached top after " + (i + 1) + " scrolls");
                    return new AutomationResult("scrollToTop", 0);
                }
            } else {
                noChangeCount = 0;
            }
            previousSource = currentSource;
        }

        return new AutomationResult("scrollToTop", 0);
    }

    /**
     * Scroll to bottom of list
     */
    @NonNull
    public AutomationResult scrollToBottom() {
        Log.i(TAG, "Scrolling to bottom");

        String previousSource = null;
        int noChangeCount = 0;

        for (int i = 0; i < maxScrollAttempts; i++) {
            AutomationResult result = engine.swipe(Direction.UP, 0);
            if (!result.isSuccess()) {
                return result;
            }

            sleep(scrollDelay);

            String currentSource = engine.getPageSource();
            if (previousSource != null && currentSource.equals(previousSource)) {
                noChangeCount++;
                if (noChangeCount >= 2) {
                    Log.i(TAG, "Reached bottom after " + (i + 1) + " scrolls");
                    return new AutomationResult("scrollToBottom", 0);
                }
            } else {
                noChangeCount = 0;
            }
            previousSource = currentSource;
        }

        return new AutomationResult("scrollToBottom", 0);
    }

    /**
     * Pull to refresh (scroll down from top)
     */
    @NonNull
    public AutomationResult pullToRefresh() {
        Log.i(TAG, "Performing pull to refresh");

        // First ensure we're at the top
        scrollToTop();

        // Then perform a larger down swipe
        AutomationResult result = engine.swipe(Direction.DOWN, 0);
        if (!result.isSuccess()) {
            return result;
        }

        // Wait for refresh to complete
        sleep(1500);

        return new AutomationResult("pullToRefresh", 0);
    }

    /**
     * Scroll by specific distance
     * @param direction Scroll direction
     * @param distancePixels Distance in pixels
     */
    @NonNull
    public AutomationResult scrollBy(@NonNull Direction direction, int distancePixels) {
        return engine.swipe(direction, distancePixels);
    }

    /**
     * Scroll until page doesn't change
     * @return Result indicating if scroll completed
     */
    @NonNull
    public AutomationResult scrollToEnd(@NonNull Direction direction) {
        Log.i(TAG, "Scrolling to end in direction: " + direction);

        String previousSource = null;
        int noChangeCount = 0;

        for (int i = 0; i < maxScrollAttempts; i++) {
            AutomationResult result = engine.swipe(direction, 0);
            if (!result.isSuccess()) {
                return result;
            }

            sleep(scrollDelay);

            String currentSource = engine.getPageSource();
            if (previousSource != null && currentSource.equals(previousSource)) {
                noChangeCount++;
                if (noChangeCount >= 2) {
                    Log.i(TAG, "Reached end after " + (i + 1) + " scrolls");
                    return new AutomationResult("scrollToEnd", 0);
                }
            } else {
                noChangeCount = 0;
            }
            previousSource = currentSource;
        }

        return new AutomationResult("scrollToEnd", 0);
    }

    /**
     * Fling scroll (fast scroll)
     */
    @NonNull
    public AutomationResult fling(@NonNull Direction direction) {
        Log.i(TAG, "Flinging in direction: " + direction);

        // Short duration for fling effect
        AutomationResult result = engine.swipe(direction, 0);
        sleep(300); // Short wait for fling to complete

        return result;
    }

    /**
     * Check if can scroll in direction
     */
    public boolean canScroll(@NonNull Direction direction) {
        String before = engine.getPageSource();
        engine.swipe(direction, 0);
        sleep(200);
        String after = engine.getPageSource();

        // If page changed, we can scroll
        return !before.equals(after);
    }

    private void sleep(long ms) {
        try {
            Thread.sleep(ms);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }
}