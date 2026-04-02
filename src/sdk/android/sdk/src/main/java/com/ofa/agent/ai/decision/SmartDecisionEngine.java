package com.ofa.agent.ai.decision;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.memory.UserMemoryManager;

import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Smart Decision Engine - makes intelligent decisions for automation operations.
 * Uses multi-armed bandit algorithms to learn optimal choices.
 */
public class SmartDecisionEngine {

    private static final String TAG = "SmartDecisionEngine";

    private final Context context;
    private final UserMemoryManager memoryManager;

    private final Map<String, MultiArmedBandit> bandits;
    private final Map<String, ContextualInfo> contextHistory;

    private DecisionListener listener;
    private boolean useMemoryForRewards = true;

    /**
     * Decision listener
     */
    public interface DecisionListener {
        void onDecisionMade(@NonNull String decisionType, @NonNull String selectedOption,
                            @NonNull Map<String, Double> scores);
        void onRewardUpdate(@NonNull String decisionType, @NonNull String option, double reward);
    }

    /**
     * Contextual information for decisions
     */
    private static class ContextualInfo {
        final long timestamp;
        final Map<String, String> features;

        ContextualInfo(Map<String, String> features) {
            this.timestamp = System.currentTimeMillis();
            this.features = new HashMap<>(features);
        }
    }

    /**
     * Decision type constants
     */
    public static final String DECISION_SHOP_SELECTION = "shop_selection";
    public static final String DECISION_PRODUCT_SELECTION = "product_selection";
    public static final String DECISION_TIMING = "timing";
    public static final String DECISION_RETRY_STRATEGY = "retry_strategy";
    public static final String DECISION_PAYMENT_METHOD = "payment_method";

    /**
     * Create smart decision engine
     */
    public SmartDecisionEngine(@NonNull Context context,
                                @Nullable UserMemoryManager memoryManager) {
        this.context = context;
        this.memoryManager = memoryManager;
        this.bandits = new HashMap<>();
        this.contextHistory = new HashMap<>();

        initializeBandits();
    }

    /**
     * Set decision listener
     */
    public void setListener(@Nullable DecisionListener listener) {
        this.listener = listener;
    }

    /**
     * Initialize bandits for different decision types
     */
    private void initializeBandits() {
        // Shop selection bandit
        bandits.put(DECISION_SHOP_SELECTION,
            new MultiArmedBandit(DECISION_SHOP_SELECTION, MultiArmedBandit.Strategy.THOMPSON_SAMPLING));

        // Payment method bandit
        bandits.put(DECISION_PAYMENT_METHOD,
            new MultiArmedBandit(DECISION_PAYMENT_METHOD, MultiArmedBandit.Strategy.UCB));

        // Retry strategy bandit
        bandits.put(DECISION_RETRY_STRATEGY,
            new MultiArmedBandit(DECISION_RETRY_STRATEGY, MultiArmedBandit.Strategy.EPSILON_GREEDY));

        Log.i(TAG, "Initialized " + bandits.size() + " decision bandits");
    }

    /**
     * Register options for a decision type
     */
    public void registerOptions(@NonNull String decisionType, @NonNull List<String> options) {
        MultiArmedBandit bandit = bandits.get(decisionType);
        if (bandit == null) {
            bandit = new MultiArmedBandit(decisionType, MultiArmedBandit.Strategy.THOMPSON_SAMPLING);
            bandits.put(decisionType, bandit);
        }

        for (String option : options) {
            bandit.addArm(option);
        }

        Log.i(TAG, "Registered " + options.size() + " options for " + decisionType);
    }

    /**
     * Select best option for decision type
     */
    @Nullable
    public String selectOption(@NonNull String decisionType) {
        return selectOption(decisionType, new HashMap<>());
    }

    /**
     * Select best option with context
     */
    @Nullable
    public String selectOption(@NonNull String decisionType,
                                @NonNull Map<String, String> context) {
        MultiArmedBandit bandit = bandits.get(decisionType);
        if (bandit == null) {
            Log.w(TAG, "No bandit for decision type: " + decisionType);
            return null;
        }

        String selected = bandit.selectArm();

        // Store context for later reward calculation
        if (selected != null) {
            contextHistory.put(decisionType + ":" + selected, new ContextualInfo(context));
        }

        // Notify listener
        if (listener != null && selected != null) {
            listener.onDecisionMade(decisionType, selected, bandit.getArmScores());
        }

        Log.d(TAG, "Selected " + selected + " for " + decisionType);
        return selected;
    }

    /**
     * Report reward for a decision
     */
    public void reportReward(@NonNull String decisionType,
                              @NonNull String option,
                              double reward) {
        MultiArmedBandit bandit = bandits.get(decisionType);
        if (bandit == null) {
            Log.w(TAG, "No bandit for decision type: " + decisionType);
            return;
        }

        bandit.update(option, reward);

        // Store in memory for long-term learning
        if (memoryManager != null && useMemoryForRewards) {
            String key = "decision.reward." + decisionType + "." + option;
            try {
                JSONObject record = new JSONObject();
                record.put("reward", reward);
                record.put("timestamp", System.currentTimeMillis());
                memoryManager.set(key, record.toString());
            } catch (Exception e) {
                Log.w(TAG, "Failed to store reward: " + e.getMessage());
            }
        }

        // Notify listener
        if (listener != null) {
            listener.onRewardUpdate(decisionType, option, reward);
        }

        Log.d(TAG, "Reported reward " + reward + " for " + option + " in " + decisionType);
    }

    /**
     * Calculate reward from automation result
     */
    public double calculateReward(@NonNull AutomationResult result) {
        if (result.isSuccess()) {
            return 1.0;
        }

        // Partial reward based on error type
        String error = result.getMessage();
        if (error == null) return 0;

        error = error.toLowerCase();

        if (error.contains("timeout")) {
            return 0.3; // Timeout might be recoverable
        }
        if (error.contains("not found")) {
            return 0.2; // Element not found
        }
        if (error.contains("permission")) {
            return 0.1; // Permission issue
        }

        return 0;
    }

    /**
     * Get best known option for decision type
     */
    @Nullable
    public String getBestOption(@NonNull String decisionType) {
        MultiArmedBandit bandit = bandits.get(decisionType);
        return bandit != null ? bandit.getBestArm() : null;
    }

    /**
     * Get scores for all options in a decision type
     */
    @NonNull
    public Map<String, Double> getOptionScores(@NonNull String decisionType) {
        MultiArmedBandit bandit = bandits.get(decisionType);
        return bandit != null ? bandit.getArmScores() : new HashMap<>();
    }

    /**
     * Get statistics for a decision type
     */
    @NonNull
    public String getDecisionStats(@NonNull String decisionType) {
        MultiArmedBandit bandit = bandits.get(decisionType);
        return bandit != null ? bandit.getStats() : "No bandit for " + decisionType;
    }

    /**
     * Get all decision types
     */
    @NonNull
    public String[] getDecisionTypes() {
        return bandits.keySet().toArray(new String[0]);
    }

    /**
     * Reset learning for a decision type
     */
    public void resetDecision(@NonNull String decisionType) {
        MultiArmedBandit bandit = bandits.get(decisionType);
        if (bandit != null) {
            bandit.reset();
        }
    }

    /**
     * Reset all learning
     */
    public void resetAll() {
        for (MultiArmedBandit bandit : bandits.values()) {
            bandit.reset();
        }
        contextHistory.clear();
        Log.i(TAG, "Reset all decision learning");
    }

    /**
     * Export decision data
     */
    @NonNull
    public JSONObject exportData() {
        JSONObject data = new JSONObject();
        try {
            for (Map.Entry<String, MultiArmedBandit> entry : bandits.entrySet()) {
                data.put(entry.getKey(), entry.getValue().getStats());
            }
        } catch (Exception e) {
            Log.w(TAG, "Failed to export data: " + e.getMessage());
        }
        return data;
    }

    /**
     * Enable/disable memory-based rewards
     */
    public void setUseMemoryForRewards(boolean enabled) {
        this.useMemoryForRewards = enabled;
    }
}