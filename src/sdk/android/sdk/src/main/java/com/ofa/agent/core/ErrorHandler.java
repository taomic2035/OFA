package com.ofa.agent.core;

import android.os.Handler;
import android.os.Looper;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.TimeUnit;
import java.util.function.Function;

/**
 * Error Handling Framework for OFA Android SDK.
 *
 * Provides:
 * - Retry strategies with exponential backoff
 * - Circuit breaker pattern
 * - Error categorization and handling
 * - Connection recovery mechanisms
 * - Graceful degradation
 *
 * @version 1.4.0
 */
public class ErrorHandler {

    private static final String TAG = "OFAErrorHandler";

    // Error categories
    public enum ErrorCategory {
        NETWORK,         // Network connectivity issues
        CONNECTION,      // Center connection issues
        TIMEOUT,         // Operation timeout
        AUTHENTICATION,  // Authentication/authorization failures
        RESOURCE,        // Resource unavailable
        VALIDATION,      // Input validation errors
        INTERNAL,        // Internal SDK errors
        UNKNOWN          // Unknown errors
    }

    // Error severity
    public enum ErrorSeverity {
        LOW,      // Minor issue, auto-recovery possible
        MEDIUM,   // Moderate issue, needs retry
        HIGH,     // Serious issue, needs intervention
        CRITICAL  // System failure, needs restart
    }

    // Error recovery strategy
    public enum RecoveryStrategy {
        IMMEDIATE_RETRY,    // Retry immediately
        BACKOFF_RETRY,      // Retry with exponential backoff
        CIRCUIT_BREAK,      // Stop trying, circuit open
        GRACEFUL_DEGRADE,   // Use fallback/alternative
        MANUAL Intervention, // Requires manual intervention
        NONE                // No recovery possible
    }

    /**
     * OFAError represents a categorized error
     */
    public static class OFAError {
        private final String code;
        private final String message;
        private final ErrorCategory category;
        private final ErrorSeverity severity;
        private final RecoveryStrategy strategy;
        private final Throwable cause;
        private final long timestamp;
        private final String context;

        public OFAError(@NonNull String code, @NonNull String message,
                        @NonNull ErrorCategory category, @NonNull ErrorSeverity severity,
                        @NonNull RecoveryStrategy strategy, @Nullable Throwable cause,
                        @Nullable String context) {
            this.code = code;
            this.message = message;
            this.category = category;
            this.severity = severity;
            this.strategy = strategy;
            this.cause = cause;
            this.timestamp = System.currentTimeMillis();
            this.context = context;
        }

        @NonNull
        public String getCode() { return code; }
        @NonNull
        public String getMessage() { return message; }
        @NonNull
        public ErrorCategory getCategory() { return category; }
        @NonNull
        public ErrorSeverity getSeverity() { return severity; }
        @NonNull
        public RecoveryStrategy getStrategy() { return strategy; }
        @Nullable
        public Throwable getCause() { return cause; }
        public long getTimestamp() { return timestamp; }
        @Nullable
        public String getContext() { return context; }

        public boolean isRecoverable() {
            return strategy != RecoveryStrategy.NONE &&
                   strategy != RecoveryStrategy.MANUAL_INTERVENTION;
        }

        @Override
        public String toString() {
            return "OFAError{" +
                   "code='" + code + '\'' +
                   ", message='" + message + '\'' +
                   ", category=" + category +
                   ", severity=" + severity +
                   ", strategy=" + strategy +
                   '}';
        }
    }

    /**
     * Retry configuration
     */
    public static class RetryConfig {
        private final int maxAttempts;
        private final long initialDelayMs;
        private final long maxDelayMs;
        private final double backoffFactor;
        private final List<ErrorCategory> retryableCategories;

        public RetryConfig(int maxAttempts, long initialDelayMs, long maxDelayMs,
                           double backoffFactor, @Nullable List<ErrorCategory> retryableCategories) {
            this.maxAttempts = maxAttempts;
            this.initialDelayMs = initialDelayMs;
            this.maxDelayMs = maxDelayMs;
            this.backoffFactor = backoffFactor;
            this.retryableCategories = retryableCategories != null ? retryableCategories :
                defaultRetryableCategories();
        }

        private static List<ErrorCategory> defaultRetryableCategories() {
            List<ErrorCategory> categories = new ArrayList<>();
            categories.add(ErrorCategory.NETWORK);
            categories.add(ErrorCategory.CONNECTION);
            categories.add(ErrorCategory.TIMEOUT);
            return categories;
        }

        public int getMaxAttempts() { return maxAttempts; }
        public long getInitialDelayMs() { return initialDelayMs; }
        public long getMaxDelayMs() { return maxDelayMs; }
        public double getBackoffFactor() { return backoffFactor; }
        public boolean shouldRetry(@NonNull ErrorCategory category) {
            return retryableCategories.contains(category);
        }

        public static RetryConfig defaultConfig() {
            return new RetryConfig(3, 1000, 10000, 2.0, null);
        }

        public static RetryConfig aggressiveConfig() {
            return new RetryConfig(5, 500, 30000, 1.5, null);
        }

        public static RetryConfig conservativeConfig() {
            return new RetryConfig(2, 2000, 5000, 2.0, null);
        }
    }

    /**
     * Circuit breaker for preventing cascading failures
     */
    public static class CircuitBreaker {
        private final String name;
        private final int failureThreshold;
        private final long recoveryTimeMs;
        private final long halfOpenDurationMs;

        private volatile State state = State.CLOSED;
        private volatile int failureCount = 0;
        private volatile long lastFailureTime = 0;
        private volatile long lastStateChangeTime = 0;
        private volatile int successCountInHalfOpen = 0;

        private final Handler handler = new Handler(Looper.getMainLooper());

        public enum State {
            CLOSED,      // Normal operation
            OPEN,        // Circuit tripped, reject all
            HALF_OPEN    // Testing if recovered
        }

        public CircuitBreaker(@NonNull String name, int failureThreshold,
                              long recoveryTimeMs, long halfOpenDurationMs) {
            this.name = name;
            this.failureThreshold = failureThreshold;
            this.recoveryTimeMs = recoveryTimeMs;
            this.halfOpenDurationMs = halfOpenDurationMs;
            this.lastStateChangeTime = System.currentTimeMillis();
        }

        public static CircuitBreaker defaultBreaker(@NonNull String name) {
            return new CircuitBreaker(name, 5, 30000, 5000);
        }

        public boolean allowExecution() {
            long now = System.currentTimeMillis();

            switch (state) {
                case CLOSED:
                    return true;

                case OPEN:
                    // Check if recovery time elapsed
                    if (now - lastStateChangeTime >= recoveryTimeMs) {
                        transitionToHalfOpen();
                        return true;
                    }
                    return false;

                case HALF_OPEN:
                    return true;
            }
            return false;
        }

        public void recordSuccess() {
            failureCount = 0;

            if (state == State.HALF_OPEN) {
                successCountInHalfOpen++;
                if (successCountInHalfOpen >= 2) {
                    transitionToClosed();
                }
            }
        }

        public void recordFailure() {
            failureCount++;
            lastFailureTime = System.currentTimeMillis();

            if (state == State.HALF_OPEN) {
                transitionToOpen();
            } else if (state == State.CLOSED && failureCount >= failureThreshold) {
                transitionToOpen();
            }
        }

        @NonNull
        public State getState() { return state; }
        public int getFailureCount() { return failureCount; }

        private void transitionToOpen() {
            state = State.OPEN;
            lastStateChangeTime = System.currentTimeMillis();
            successCountInHalfOpen = 0;
            Log.w(TAG, "CircuitBreaker[" + name + "] OPEN - failures: " + failureCount);
        }

        private void transitionToHalfOpen() {
            state = State.HALF_OPEN;
            lastStateChangeTime = System.currentTimeMillis();
            successCountInHalfOpen = 0;
            Log.i(TAG, "CircuitBreaker[" + name + "] HALF_OPEN - testing");

            // Auto-close if half-open duration passes without issues
            handler.postDelayed(() -> {
                if (state == State.HALF_OPEN) {
                    transitionToClosed();
                }
            }, halfOpenDurationMs);
        }

        private void transitionToClosed() {
            state = State.CLOSED;
            lastStateChangeTime = System.currentTimeMillis();
            failureCount = 0;
            successCountInHalfOpen = 0;
            Log.i(TAG, "CircuitBreaker[" + name + "] CLOSED - recovered");
        }

        public void reset() {
            transitionToClosed();
        }
    }

    /**
     * Retry executor with exponential backoff
     */
    public static class RetryExecutor {
        private final RetryConfig config;
        private final Handler handler;
        private int attemptCount = 0;
        private volatile boolean cancelled = false;

        public RetryExecutor(@NonNull RetryConfig config) {
            this.config = config;
            this.handler = new Handler(Looper.getMainLooper());
        }

        public <T> CompletableFuture<T> execute(
                @NonNull Function<Integer, CompletableFuture<T>> operation,
                @Nullable CircuitBreaker circuitBreaker) {
            CompletableFuture<T> result = new CompletableFuture<>();
            attemptCount = 0;
            cancelled = false;
            doExecute(operation, result, circuitBreaker);
            return result;
        }

        private <T> void doExecute(
                Function<Integer, CompletableFuture<T>> operation,
                CompletableFuture<T> result,
                @Nullable CircuitBreaker circuitBreaker) {

            if (cancelled) {
                result.completeExceptionally(new OFAError(
                    "RETRY_CANCELLED", "Retry cancelled",
                    ErrorCategory.INTERNAL, ErrorSeverity.MEDIUM,
                    RecoveryStrategy.NONE, null, null
                ));
                return;
            }

            // Check circuit breaker
            if (circuitBreaker != null && !circuitBreaker.allowExecution()) {
                result.completeExceptionally(new OFAError(
                    "CIRCUIT_OPEN", "Circuit breaker is open",
                    ErrorCategory.INTERNAL, ErrorSeverity.HIGH,
                    RecoveryStrategy.CIRCUIT_BREAK, null, null
                ));
                return;
            }

            attemptCount++;

            try {
                CompletableFuture<T> opResult = operation.apply(attemptCount);
                opResult.whenComplete((value, error) -> {
                    if (error == null) {
                        if (circuitBreaker != null) {
                            circuitBreaker.recordSuccess();
                        }
                        result.complete(value);
                    } else {
                        OFAError ofaError = categorizeError(error);
                        handleFailure(ofaError, operation, result, circuitBreaker);
                    }
                });
            } catch (Exception e) {
                OFAError ofaError = categorizeError(e);
                handleFailure(ofaError, operation, result, circuitBreaker);
            }
        }

        private <T> void handleFailure(
                @NonNull OFAError error,
                @NonNull Function<Integer, CompletableFuture<T>> operation,
                @NonNull CompletableFuture<T> result,
                @Nullable CircuitBreaker circuitBreaker) {

            if (circuitBreaker != null) {
                circuitBreaker.recordFailure();
            }

            // Check if we should retry
            if (attemptCount < config.getMaxAttempts() &&
                config.shouldRetry(error.getCategory()) &&
                error.isRecoverable()) {

                long delay = calculateDelay();
                Log.w(TAG, "Retrying after " + delay + "ms (attempt " + attemptCount + "/" +
                       config.getMaxAttempts() + "): " + error.getMessage());

                handler.postDelayed(() ->
                    doExecute(operation, result, circuitBreaker), delay);
            } else {
                Log.e(TAG, "Retry exhausted: " + error.getMessage());
                result.completeExceptionally(error);
            }
        }

        private long calculateDelay() {
            long delay = (long) (config.getInitialDelayMs() *
                        Math.pow(config.getBackoffFactor(), attemptCount - 1));
            return Math.min(delay, config.getMaxDelayMs());
        }

        public void cancel() {
            cancelled = true;
        }

        public int getAttemptCount() {
            return attemptCount;
        }
    }

    /**
     * Connection recovery manager
     */
    public static class ConnectionRecoveryManager {
        private final CenterConnection connection;
        private final RetryExecutor retryExecutor;
        private final CircuitBreaker circuitBreaker;
        private final Handler handler;
        private volatile boolean recovering = false;

        private static final long HEALTH_CHECK_INTERVAL = 5000; // 5 seconds
        private static final int MAX_RECOVERY_ATTEMPTS = 10;

        public ConnectionRecoveryManager(@NonNull CenterConnection connection) {
            this.connection = connection;
            this.retryExecutor = new RetryExecutor(RetryConfig.aggressiveConfig());
            this.circuitBreaker = CircuitBreaker.defaultBreaker("center_connection");
            this.handler = new Handler(Looper.getMainLooper());
        }

        public void startRecovery() {
            if (recovering) return;

            recovering = true;
            Log.i(TAG, "Starting connection recovery...");

            doRecovery();
        }

        private void doRecovery() {
            if (!recovering) return;

            if (!circuitBreaker.allowExecution()) {
                Log.w(TAG, "Circuit breaker open, waiting for recovery window");
                handler.postDelayed(this::doRecovery, circuitBreaker.recoveryTimeMs);
                return;
            }

            // Check current connection state
            if (connection.isConnected()) {
                Log.i(TAG, "Connection recovered!");
                recovering = false;
                circuitBreaker.recordSuccess();
                return;
            }

            // Attempt reconnect
            Log.i(TAG, "Attempting reconnect...");

            CompletableFuture<Boolean> connectResult = retryExecutor.execute(
                attempt -> {
                    CompletableFuture<Boolean> future = new CompletableFuture<>();
                    try {
                        connection.connect();
                        handler.postDelayed(() -> {
                            future.complete(connection.isConnected());
                        }, 2000);
                    } catch (Exception e) {
                        future.completeExceptionally(e);
                    }
                    return future;
                },
                circuitBreaker
            );

            connectResult.whenComplete((connected, error) -> {
                if (error == null && connected) {
                    Log.i(TAG, "Recovery successful!");
                    recovering = false;
                } else {
                    Log.w(TAG, "Recovery attempt failed");
                    handler.postDelayed(this::doRecovery, HEALTH_CHECK_INTERVAL);
                }
            });
        }

        public void stopRecovery() {
            recovering = false;
            retryExecutor.cancel();
        }

        @NonNull
        public CircuitBreaker.State getCircuitState() {
            return circuitBreaker.getState();
        }

        public boolean isRecovering() {
            return recovering;
        }
    }

    /**
     * Fallback provider for graceful degradation
     */
    public static class FallbackProvider {
        private final List<FallbackHandler> handlers = new ArrayList<>();

        public interface FallbackHandler {
            @Nullable
            <T> T provideFallback(@NonNull OFAError error, @Nullable String context);
            boolean canHandle(@NonNull ErrorCategory category);
        }

        public void registerHandler(@NonNull FallbackHandler handler) {
            handlers.add(handler);
        }

        @Nullable
        public <T> T getFallback(@NonNull OFAError error, @Nullable String context) {
            for (FallbackHandler handler : handlers) {
                if (handler.canHandle(error.getCategory())) {
                    T result = handler.provideFallback(error, context);
                    if (result != null) {
                        Log.i(TAG, "Using fallback for " + error.getCategory());
                        return result;
                    }
                }
            }
            return null;
        }

        public static FallbackProvider createDefault() {
            FallbackProvider provider = new FallbackProvider();

            // Network fallback - use cached data
            provider.registerHandler(new FallbackHandler() {
                @Override
                public <T> T provideFallback(@NonNull OFAError error, @Nullable String context) {
                    // Return cached response if available
                    return null;
                }

                @Override
                public boolean canHandle(@NonNull ErrorCategory category) {
                    return category == ErrorCategory.NETWORK ||
                           category == ErrorCategory.CONNECTION;
                }
            });

            // Timeout fallback - use quick response
            provider.registerHandler(new FallbackHandler() {
                @Override
                public <T> T provideFallback(@NonNull OFAError error, @Nullable String context) {
                    // Return simplified response
                    return null;
                }

                @Override
                public boolean canHandle(@NonNull ErrorCategory category) {
                    return category == ErrorCategory.TIMEOUT;
                }
            });

            return provider;
        }
    }

    // Error categorization

    /**
     * Categorize an exception into OFAError
     */
    @NonNull
    public static OFAError categorizeError(@NonNull Throwable error) {
        String message = error.getMessage() != null ? error.getMessage() : "Unknown error";

        // Network errors
        if (isNetworkError(error)) {
            return new OFAError(
                "NETWORK_ERROR", message,
                ErrorCategory.NETWORK, ErrorSeverity.MEDIUM,
                RecoveryStrategy.BACKOFF_RETRY, error, null
            );
        }

        // Connection errors
        if (isConnectionError(error)) {
            return new OFAError(
                "CONNECTION_ERROR", message,
                ErrorCategory.CONNECTION, ErrorSeverity.MEDIUM,
                RecoveryStrategy.BACKOFF_RETRY, error, null
            );
        }

        // Timeout errors
        if (isTimeoutError(error)) {
            return new OFAError(
                "TIMEOUT_ERROR", message,
                ErrorCategory.TIMEOUT, ErrorSeverity.LOW,
                RecoveryStrategy.IMMEDIATE_RETRY, error, null
            );
        }

        // Authentication errors
        if (isAuthenticationError(error)) {
            return new OFAError(
                "AUTH_ERROR", message,
                ErrorCategory.AUTHENTICATION, ErrorSeverity.HIGH,
                RecoveryStrategy.MANUAL_INTERVENTION, error, null
            );
        }

        // Internal errors
        if (isInternalError(error)) {
            return new OFAError(
                "INTERNAL_ERROR", message,
                ErrorCategory.INTERNAL, ErrorSeverity.HIGH,
                RecoveryStrategy.GRACEFUL_DEGRADE, error, null
            );
        }

        // Unknown errors
        return new OFAError(
            "UNKNOWN_ERROR", message,
            ErrorCategory.UNKNOWN, ErrorSeverity.MEDIUM,
            RecoveryStrategy.NONE, error, null
        );
    }

    private static boolean isNetworkError(Throwable error) {
        String className = error.getClass().getName();
        return className.contains("Socket") ||
               className.contains("Network") ||
               className.contains("ConnectException") ||
               messageContains(error, "network", "offline", "unreachable", "no internet");
    }

    private static boolean isConnectionError(Throwable error) {
        String className = error.getClass().getName();
        return className.contains("Connection") ||
               className.contains("ClosedChannelException") ||
               messageContains(error, "connection", "closed", "disconnected", "broken pipe");
    }

    private static boolean isTimeoutError(Throwable error) {
        String className = error.getClass().getName();
        return className.contains("Timeout") ||
               className.contains("TimeoutException") ||
               messageContains(error, "timeout", "timed out", "deadline");
    }

    private static boolean isAuthenticationError(Throwable error) {
        String className = error.getClass().getName();
        return className.contains("Auth") ||
               className.contains("Security") ||
               className.contains("Permission") ||
               messageContains(error, "auth", "token", "permission", "access denied");
    }

    private static boolean isInternalError(Throwable error) {
        String className = error.getClass().getName();
        return className.contains("Illegal") ||
               className.contains("NullPointer") ||
               className.contains("OutOfMemory") ||
               className.contains("RuntimeException");
    }

    private static boolean messageContains(Throwable error, String... keywords) {
        if (error.getMessage() == null) return false;
        String message = error.getMessage().toLowerCase();
        for (String keyword : keywords) {
            if (message.contains(keyword.toLowerCase())) {
                return true;
            }
        }
        return false;
    }

    /**
     * Error listener for monitoring errors
     */
    public interface ErrorListener {
        void onErrorOccurred(@NonNull OFAError error);
        void onErrorRecovered(@NonNull OFAError error);
    }

    private static final List<ErrorListener> listeners = new ArrayList<>();

    public static void addListener(@NonNull ErrorListener listener) {
        listeners.add(listener);
    }

    public static void removeListener(@NonNull ErrorListener listener) {
        listeners.remove(listener);
    }

    public static void notifyError(@NonNull OFAError error) {
        for (ErrorListener listener : listeners) {
            listener.onErrorOccurred(error);
        }
    }

    public static void notifyRecovery(@NonNull OFAError error) {
        for (ErrorListener listener : listeners) {
            listener.onErrorRecovered(error);
        }
    }
}