package com.ofa.agent.automation;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

/**
 * Automation configuration.
 */
public class AutomationConfig {

    // Default timeouts
    public static final long DEFAULT_CLICK_TIMEOUT = 5000;
    public static final long DEFAULT_SWIPE_TIMEOUT = 3000;
    public static final long DEFAULT_INPUT_TIMEOUT = 10000;
    public static final long DEFAULT_WAIT_TIMEOUT = 30000;
    public static final long DEFAULT_PAGE_STABLE_TIMEOUT = 5000;

    // Default gesture parameters
    public static final long DEFAULT_CLICK_DURATION = 100;
    public static final long DEFAULT_LONG_CLICK_DURATION = 500;
    public static final long DEFAULT_SWIPE_DURATION = 300;
    public static final int DEFAULT_SCROLL_DISTANCE = 500;

    // Timeout settings
    private long clickTimeout = DEFAULT_CLICK_TIMEOUT;
    private long swipeTimeout = DEFAULT_SWIPE_TIMEOUT;
    private long inputTimeout = DEFAULT_INPUT_TIMEOUT;
    private long waitTimeout = DEFAULT_WAIT_TIMEOUT;
    private long pageStableTimeout = DEFAULT_PAGE_STABLE_TIMEOUT;

    // Gesture settings
    private long clickDuration = DEFAULT_CLICK_DURATION;
    private long longClickDuration = DEFAULT_LONG_CLICK_DURATION;
    private long swipeDuration = DEFAULT_SWIPE_DURATION;
    private int scrollDistance = DEFAULT_SCROLL_DISTANCE;

    // Behavior settings
    private boolean autoRetry = true;
    private int maxRetries = 3;
    private long retryDelay = 1000;
    private boolean enableLogging = true;
    private boolean enableScreenshotOnError = true;

    // Package filter (null = all packages)
    @Nullable
    private String[] packageFilter;

    /**
     * Default constructor
     */
    public AutomationConfig() {}

    /**
     * Builder pattern
     */
    @NonNull
    public static Builder builder() {
        return new Builder();
    }

    // ===== Getters =====

    public long getClickTimeout() {
        return clickTimeout;
    }

    public long getSwipeTimeout() {
        return swipeTimeout;
    }

    public long getInputTimeout() {
        return inputTimeout;
    }

    public long getWaitTimeout() {
        return waitTimeout;
    }

    public long getPageStableTimeout() {
        return pageStableTimeout;
    }

    public long getClickDuration() {
        return clickDuration;
    }

    public long getLongClickDuration() {
        return longClickDuration;
    }

    public long getSwipeDuration() {
        return swipeDuration;
    }

    public int getScrollDistance() {
        return scrollDistance;
    }

    public boolean isAutoRetryEnabled() {
        return autoRetry;
    }

    public int getMaxRetries() {
        return maxRetries;
    }

    public long getRetryDelay() {
        return retryDelay;
    }

    public boolean isLoggingEnabled() {
        return enableLogging;
    }

    public boolean isScreenshotOnErrorEnabled() {
        return enableScreenshotOnError;
    }

    @Nullable
    public String[] getPackageFilter() {
        return packageFilter;
    }

    // ===== Builder =====

    public static class Builder {
        private final AutomationConfig config = new AutomationConfig();

        @NonNull
        public Builder clickTimeout(long timeout) {
            config.clickTimeout = timeout;
            return this;
        }

        @NonNull
        public Builder swipeTimeout(long timeout) {
            config.swipeTimeout = timeout;
            return this;
        }

        @NonNull
        public Builder inputTimeout(long timeout) {
            config.inputTimeout = timeout;
            return this;
        }

        @NonNull
        public Builder waitTimeout(long timeout) {
            config.waitTimeout = timeout;
            return this;
        }

        @NonNull
        public Builder pageStableTimeout(long timeout) {
            config.pageStableTimeout = timeout;
            return this;
        }

        @NonNull
        public Builder clickDuration(long duration) {
            config.clickDuration = duration;
            return this;
        }

        @NonNull
        public Builder longClickDuration(long duration) {
            config.longClickDuration = duration;
            return this;
        }

        @NonNull
        public Builder swipeDuration(long duration) {
            config.swipeDuration = duration;
            return this;
        }

        @NonNull
        public Builder scrollDistance(int distance) {
            config.scrollDistance = distance;
            return this;
        }

        @NonNull
        public Builder autoRetry(boolean enabled) {
            config.autoRetry = enabled;
            return this;
        }

        @NonNull
        public Builder maxRetries(int retries) {
            config.maxRetries = retries;
            return this;
        }

        @NonNull
        public Builder retryDelay(long delay) {
            config.retryDelay = delay;
            return this;
        }

        @NonNull
        public Builder enableLogging(boolean enabled) {
            config.enableLogging = enabled;
            return this;
        }

        @NonNull
        public Builder enableScreenshotOnError(boolean enabled) {
            config.enableScreenshotOnError = enabled;
            return this;
        }

        @NonNull
        public Builder packageFilter(@Nullable String[] packages) {
            config.packageFilter = packages;
            return this;
        }

        @NonNull
        public AutomationConfig build() {
            return config;
        }
    }
}