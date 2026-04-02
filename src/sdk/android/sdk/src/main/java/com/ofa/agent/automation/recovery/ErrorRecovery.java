package com.ofa.agent.automation.recovery;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.automation.BySelector;
import com.ofa.agent.automation.Direction;

import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;

/**
 * Error Recovery - handles automation errors and provides recovery strategies.
 */
public class ErrorRecovery {

    private static final String TAG = "ErrorRecovery";

    private final AutomationEngine engine;
    private final List<RecoveryStrategy> strategies;
    private final int maxRecoveryAttempts;
    private RecoveryListener listener;

    /**
     * Recovery listener
     */
    public interface RecoveryListener {
        void onRecoveryAttempt(int attempt, @NonNull String error, @NonNull String strategy);
        void onRecoverySuccess(@NonNull String strategy);
        void onRecoveryFailure(@NonNull String lastError);
    }

    /**
     * Recovery strategy
     */
    public interface RecoveryStrategy {
        boolean canHandle(@NonNull String error);
        @NonNull
        AutomationResult recover(@NonNull AutomationEngine engine, @NonNull String error);
        String getName();
    }

    /**
     * Create error recovery
     */
    public ErrorRecovery(@NonNull AutomationEngine engine) {
        this(engine, 3);
    }

    /**
     * Create error recovery with custom max attempts
     */
    public ErrorRecovery(@NonNull AutomationEngine engine, int maxRecoveryAttempts) {
        this.engine = engine;
        this.maxRecoveryAttempts = maxRecoveryAttempts;
        this.strategies = new ArrayList<>();

        registerDefaultStrategies();
    }

    /**
     * Set recovery listener
     */
    public void setListener(@Nullable RecoveryListener listener) {
        this.listener = listener;
    }

    /**
     * Register a recovery strategy
     */
    public void registerStrategy(@NonNull RecoveryStrategy strategy) {
        strategies.add(strategy);
        Log.i(TAG, "Registered recovery strategy: " + strategy.getName());
    }

    /**
     * Attempt recovery from an error
     */
    @NonNull
    public AutomationResult recover(@NonNull AutomationResult failedResult) {
        String error = failedResult.getMessage();
        Log.i(TAG, "Attempting recovery from error: " + error);

        for (int attempt = 1; attempt <= maxRecoveryAttempts; attempt++) {
            Log.i(TAG, "Recovery attempt " + attempt + "/" + maxRecoveryAttempts);

            // Find applicable strategy
            for (RecoveryStrategy strategy : strategies) {
                if (strategy.canHandle(error)) {
                    if (listener != null) {
                        listener.onRecoveryAttempt(attempt, error, strategy.getName());
                    }

                    Log.i(TAG, "Trying strategy: " + strategy.getName());
                    AutomationResult result = strategy.recover(engine, error);

                    if (result.isSuccess()) {
                        Log.i(TAG, "Recovery successful with strategy: " + strategy.getName());
                        if (listener != null) {
                            listener.onRecoverySuccess(strategy.getName());
                        }
                        return result;
                    }
                }
            }

            // Wait before next attempt
            try {
                Thread.sleep(1000 * attempt);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                break;
            }
        }

        Log.w(TAG, "All recovery attempts failed");
        if (listener != null) {
            listener.onRecoveryFailure(error);
        }

        return new AutomationResult("recovery", "Recovery failed after " + maxRecoveryAttempts + " attempts: " + error);
    }

    /**
     * Register default recovery strategies
     */
    private void registerDefaultStrategies() {
        // Element not found - try scrolling
        registerStrategy(new RecoveryStrategy() {
            @Override
            public boolean canHandle(@NonNull String error) {
                return error.toLowerCase().contains("not found") ||
                       error.toLowerCase().contains("element not");
            }

            @NonNull
            @Override
            public AutomationResult recover(@NonNull AutomationEngine engine, @NonNull String error) {
                Log.d(TAG, "Recovery: Scrolling to find element");
                // Try scrolling down
                AutomationResult swipeResult = engine.swipe(Direction.DOWN, 0);
                if (swipeResult.isSuccess()) {
                    try {
                        Thread.sleep(500);
                    } catch (InterruptedException e) {
                        Thread.currentThread().interrupt();
                    }
                }
                return new AutomationResult("scroll_recovery", 0);
            }

            @Override
            public String getName() {
                return "ScrollToFind";
            }
        });

        // Timeout - retry operation
        registerStrategy(new RecoveryStrategy() {
            @Override
            public boolean canHandle(@NonNull String error) {
                return error.toLowerCase().contains("timeout") ||
                       error.toLowerCase().contains("timed out");
            }

            @NonNull
            @Override
            public AutomationResult recover(@NonNull AutomationEngine engine, @NonNull String error) {
                Log.d(TAG, "Recovery: Waiting and retrying");
                try {
                    Thread.sleep(2000);
                    return new AutomationResult("timeout_recovery", 0);
                } catch (InterruptedException e) {
                    return new AutomationResult("timeout_recovery", e.getMessage());
                }
            }

            @Override
            public String getName() {
                return "WaitAndRetry";
            }
        });

        // Dialog/popup blocking - dismiss
        registerStrategy(new RecoveryStrategy() {
            @Override
            public boolean canHandle(@NonNull String error) {
                return error.toLowerCase().contains("blocked") ||
                       error.toLowerCase().contains("overlay") ||
                       error.toLowerCase().contains("dialog");
            }

            @NonNull
            @Override
            public AutomationResult recover(@NonNull AutomationEngine engine, @NonNull String error) {
                Log.d(TAG, "Recovery: Trying to dismiss dialogs");
                // Try pressing back
                engine.swipe(Direction.LEFT, 0);
                try {
                    Thread.sleep(300);
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                }
                // Try clicking outside (center of screen might have overlay)
                return new AutomationResult("dismiss_recovery", 0);
            }

            @Override
            public String getName() {
                return "DismissOverlay";
            }
        });

        // Page loading - wait for stability
        registerStrategy(new RecoveryStrategy() {
            @Override
            public boolean canHandle(@NonNull String error) {
                return error.toLowerCase().contains("loading") ||
                       error.toLowerCase().contains("not ready") ||
                       error.toLowerCase().contains("page not");
            }

            @NonNull
            @Override
            public AutomationResult recover(@NonNull AutomationEngine engine, @NonNull String error) {
                Log.d(TAG, "Recovery: Waiting for page to stabilize");
                AutomationResult result = engine.waitForPageStable(5000);
                return result;
            }

            @Override
            public String getName() {
                return "WaitForPage";
            }
        });

        // Permission denied - go back and retry
        registerStrategy(new RecoveryStrategy() {
            @Override
            public boolean canHandle(@NonNull String error) {
                return error.toLowerCase().contains("permission") ||
                       error.toLowerCase().contains("denied") ||
                       error.toLowerCase().contains("unauthorized");
            }

            @NonNull
            @Override
            public AutomationResult recover(@NonNull AutomationEngine engine, @NonNull String error) {
                Log.d(TAG, "Recovery: Handling permission issue");
                // Navigate back and return error - user needs to grant permission
                return new AutomationResult("permission_recovery",
                    "Permission required - please grant and retry");
            }

            @Override
            public String getName() {
                return "HandlePermission";
            }
        });

        // Network error - wait and retry
        registerStrategy(new RecoveryStrategy() {
            @Override
            public boolean canHandle(@NonNull String error) {
                return error.toLowerCase().contains("network") ||
                       error.toLowerCase().contains("connection") ||
                       error.toLowerCase().contains("offline");
            }

            @NonNull
            @Override
            public AutomationResult recover(@NonNull AutomationEngine engine, @NonNull String error) {
                Log.d(TAG, "Recovery: Waiting for network");
                try {
                    Thread.sleep(3000);
                    // Try refreshing
                    engine.swipe(Direction.DOWN, 0);
                    Thread.sleep(2000);
                    return new AutomationResult("network_recovery", 0);
                } catch (InterruptedException e) {
                    return new AutomationResult("network_recovery", e.getMessage());
                }
            }

            @Override
            public String getName() {
                return "HandleNetwork";
            }
        });

        Log.i(TAG, "Registered " + strategies.size() + " default recovery strategies");
    }

    /**
     * Get all registered strategies
     */
    @NonNull
    public List<RecoveryStrategy> getStrategies() {
        return new ArrayList<>(strategies);
    }

    /**
     * Clear all strategies
     */
    public void clearStrategies() {
        strategies.clear();
    }

    /**
     * Check if error is recoverable
     */
    public boolean isRecoverable(@NonNull String error) {
        for (RecoveryStrategy strategy : strategies) {
            if (strategy.canHandle(error)) {
                return true;
            }
        }
        return false;
    }

    /**
     * Get applicable strategies for an error
     */
    @NonNull
    public List<RecoveryStrategy> getApplicableStrategies(@NonNull String error) {
        List<RecoveryStrategy> applicable = new ArrayList<>();
        for (RecoveryStrategy strategy : strategies) {
            if (strategy.canHandle(error)) {
                applicable.add(strategy);
            }
        }
        return applicable;
    }
}