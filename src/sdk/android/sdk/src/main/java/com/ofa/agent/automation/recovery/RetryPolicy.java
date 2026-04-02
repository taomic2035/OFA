package com.ofa.agent.automation.recovery;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationResult;

import java.util.Random;

/**
 * Retry Policy - defines how operations should be retried on failure.
 * Supports exponential backoff, jitter, and custom conditions.
 */
public class RetryPolicy {

    private static final String TAG = "RetryPolicy";

    private final int maxRetries;
    private final long initialDelayMs;
    private final long maxDelayMs;
    private final double backoffMultiplier;
    private final double jitterFactor;
    private final RetryCondition condition;

    private int retryCount = 0;
    private long totalDelayMs = 0;

    /**
     * Retry condition interface
     */
    public interface RetryCondition {
        boolean shouldRetry(@NonNull AutomationResult result);
    }

    /**
     * Create retry policy with defaults
     */
    public RetryPolicy() {
        this(3, 1000, 30000, 2.0, 0.1, null);
    }

    /**
     * Create retry policy with custom settings
     */
    public RetryPolicy(int maxRetries, long initialDelayMs, long maxDelayMs,
                        double backoffMultiplier, double jitterFactor,
                        @Nullable RetryCondition condition) {
        this.maxRetries = maxRetries;
        this.initialDelayMs = initialDelayMs;
        this.maxDelayMs = maxDelayMs;
        this.backoffMultiplier = backoffMultiplier;
        this.jitterFactor = jitterFactor;
        this.condition = condition;
    }

    /**
     * Check if should retry
     */
    public boolean shouldRetry(@NonNull AutomationResult result) {
        if (retryCount >= maxRetries) {
            Log.d(TAG, "Max retries reached: " + retryCount);
            return false;
        }

        if (result.isSuccess()) {
            return false;
        }

        if (condition != null && !condition.shouldRetry(result)) {
            Log.d(TAG, "Retry condition returned false");
            return false;
        }

        return true;
    }

    /**
     * Get delay before next retry
     */
    public long getNextDelay() {
        // Calculate exponential backoff
        long delay = (long) (initialDelayMs * Math.pow(backoffMultiplier, retryCount));

        // Cap at max delay
        delay = Math.min(delay, maxDelayMs);

        // Add jitter
        if (jitterFactor > 0) {
            Random random = new Random();
            long jitter = (long) (delay * jitterFactor * (random.nextDouble() * 2 - 1));
            delay += jitter;
        }

        return Math.max(0, delay);
    }

    /**
     * Record a retry attempt
     */
    public void recordRetry() {
        retryCount++;
        totalDelayMs += getNextDelay();
        Log.d(TAG, "Retry #" + retryCount + ", total delay: " + totalDelayMs + "ms");
    }

    /**
     * Reset retry count
     */
    public void reset() {
        retryCount = 0;
        totalDelayMs = 0;
    }

    /**
     * Get current retry count
     */
    public int getRetryCount() {
        return retryCount;
    }

    /**
     * Get total delay so far
     */
    public long getTotalDelayMs() {
        return totalDelayMs;
    }

    /**
     * Wait for next retry
     */
    public void waitBeforeRetry() {
        long delay = getNextDelay();
        Log.d(TAG, "Waiting " + delay + "ms before retry");

        try {
            Thread.sleep(delay);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            Log.w(TAG, "Retry wait interrupted");
        }
    }

    /**
     * Get retry statistics
     */
    @NonNull
    public RetryStats getStats() {
        return new RetryStats(retryCount, maxRetries, totalDelayMs);
    }

    /**
     * Retry statistics
     */
    public static class RetryStats {
        public final int attempts;
        public final int maxAttempts;
        public final long totalDelayMs;

        public RetryStats(int attempts, int maxAttempts, long totalDelayMs) {
            this.attempts = attempts;
            this.maxAttempts = maxAttempts;
            this.totalDelayMs = totalDelayMs;
        }

        @NonNull
        @Override
        public String toString() {
            return "RetryStats{attempts=" + attempts + "/" + maxAttempts +
                ", totalDelay=" + totalDelayMs + "ms}";
        }
    }

    // ===== Builder =====

    /**
     * Create builder
     */
    @NonNull
    public static Builder builder() {
        return new Builder();
    }

    /**
     * Builder for RetryPolicy
     */
    public static class Builder {
        private int maxRetries = 3;
        private long initialDelayMs = 1000;
        private long maxDelayMs = 30000;
        private double backoffMultiplier = 2.0;
        private double jitterFactor = 0.1;
        private RetryCondition condition = null;

        @NonNull
        public Builder maxRetries(int max) {
            this.maxRetries = max;
            return this;
        }

        @NonNull
        public Builder initialDelay(long delayMs) {
            this.initialDelayMs = delayMs;
            return this;
        }

        @NonNull
        public Builder maxDelay(long delayMs) {
            this.maxDelayMs = delayMs;
            return this;
        }

        @NonNull
        public Builder backoffMultiplier(double multiplier) {
            this.backoffMultiplier = multiplier;
            return this;
        }

        @NonNull
        public Builder jitterFactor(double factor) {
            this.jitterFactor = factor;
            return this;
        }

        @NonNull
        public Builder condition(@NonNull RetryCondition condition) {
            this.condition = condition;
            return this;
        }

        @NonNull
        public RetryPolicy build() {
            return new RetryPolicy(maxRetries, initialDelayMs, maxDelayMs,
                backoffMultiplier, jitterFactor, condition);
        }
    }

    // ===== Preset Policies =====

    /**
     * No retry policy
     */
    @NonNull
    public static RetryPolicy noRetry() {
        return new RetryPolicy(0, 0, 0, 1.0, 0, null);
    }

    /**
     * Quick retry policy (3 attempts, short delays)
     */
    @NonNull
    public static RetryPolicy quick() {
        return new RetryPolicy(3, 500, 5000, 1.5, 0.1, null);
    }

    /**
     * Standard retry policy (3 attempts, exponential backoff)
     */
    @NonNull
    public static RetryPolicy standard() {
        return new RetryPolicy(3, 1000, 30000, 2.0, 0.1, null);
    }

    /**
     * Aggressive retry policy (5 attempts, longer delays)
     */
    @NonNull
    public static RetryPolicy aggressive() {
        return new RetryPolicy(5, 2000, 60000, 2.0, 0.2, null);
    }

    /**
     * Network retry policy (optimized for network errors)
     */
    @NonNull
    public static RetryPolicy network() {
        return builder()
            .maxRetries(5)
            .initialDelay(1000)
            .maxDelay(30000)
            .backoffMultiplier(2.0)
            .jitterFactor(0.3)
            .condition(result -> {
                String msg = result.getMessage();
                if (msg == null) return false;
                String lower = msg.toLowerCase();
                return lower.contains("network") || lower.contains("timeout") ||
                       lower.contains("connection") || lower.contains("server");
            })
            .build();
    }

    /**
     * UI retry policy (optimized for UI operations)
     */
    @NonNull
    public static RetryPolicy ui() {
        return builder()
            .maxRetries(3)
            .initialDelay(500)
            .maxDelay(10000)
            .backoffMultiplier(1.5)
            .jitterFactor(0.1)
            .condition(result -> {
                String msg = result.getMessage();
                if (msg == null) return false;
                String lower = msg.toLowerCase();
                // Retry on not found, loading, timeout
                return lower.contains("not found") || lower.contains("loading") ||
                       lower.contains("timeout") || lower.contains("element");
            })
            .build();
    }
}