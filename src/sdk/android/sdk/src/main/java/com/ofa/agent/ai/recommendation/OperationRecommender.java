package com.ofa.agent.ai.recommendation;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.automation.adapter.AppAdapter;
import com.ofa.agent.automation.adapter.AppAdapterManager;
import com.ofa.agent.memory.UserMemoryManager;

import org.json.JSONArray;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Operation Recommender - recommends optimal operations based on context and history.
 * Learns from user behavior and automation outcomes.
 */
public class OperationRecommender {

    private static final String TAG = "OperationRecommender";

    private final Context context;
    private final UserMemoryManager memoryManager;
    private final AppAdapterManager adapterManager;

    private final Map<String, OperationStats> operationStats;
    private final Map<String, List<String>> operationSequences;

    private RecommendationListener listener;
    private int maxRecommendations = 5;

    /**
     * Recommendation listener
     */
    public interface RecommendationListener {
        void onRecommendation(@NonNull List<Recommendation> recommendations);
    }

    /**
     * Operation statistics
     */
    private static class OperationStats {
        int count = 0;
        int successCount = 0;
        long totalTimeMs = 0;
        long lastUsed = 0;

        void record(boolean success, long timeMs) {
            count++;
            if (success) successCount++;
            totalTimeMs += timeMs;
            lastUsed = System.currentTimeMillis();
        }

        double getSuccessRate() {
            return count > 0 ? (double) successCount / count : 0;
        }

        double getAverageTime() {
            return count > 0 ? (double) totalTimeMs / count : 0;
        }
    }

    /**
     * Recommendation result
     */
    public static class Recommendation {
        public final String operation;
        public final Map<String, String> params;
        public final double score;
        public final String reason;

        public Recommendation(String operation, Map<String, String> params, double score, String reason) {
            this.operation = operation;
            this.params = params;
            this.score = score;
            this.reason = reason;
        }

        @NonNull
        @Override
        public String toString() {
            return String.format("Recommendation{op=%s, score=%.2f, reason=%s}",
                operation, score, reason);
        }
    }

    /**
     * Create operation recommender
     */
    public OperationRecommender(@NonNull Context context,
                                 @Nullable UserMemoryManager memoryManager,
                                 @Nullable AppAdapterManager adapterManager) {
        this.context = context;
        this.memoryManager = memoryManager;
        this.adapterManager = adapterManager;
        this.operationStats = new HashMap<>();
        this.operationSequences = new HashMap<>();

        loadFromMemory();
    }

    /**
     * Set recommendation listener
     */
    public void setListener(@Nullable RecommendationListener listener) {
        this.listener = listener;
    }

    /**
     * Set max recommendations
     */
    public void setMaxRecommendations(int max) {
        this.maxRecommendations = Math.max(1, max);
    }

    /**
     * Get recommendations for current context
     */
    @NonNull
    public List<Recommendation> getRecommendations() {
        return getRecommendations(new HashMap<>());
    }

    /**
     * Get recommendations with context
     */
    @NonNull
    public List<Recommendation> getRecommendations(@NonNull Map<String, String> context) {
        List<Recommendation> recommendations = new ArrayList<>();

        // Get current app info
        String appName = context.getOrDefault("app_name", "unknown");
        String page = context.getOrDefault("page", "unknown");
        String lastOperation = context.get("last_operation");

        // Score each possible operation
        Map<String, Double> scores = new HashMap<>();

        // Score based on page context
        scores.putAll(getPageBasedRecommendations(page));

        // Score based on operation sequence
        if (lastOperation != null) {
            scores.putAll(getSequenceBasedRecommendations(lastOperation));
        }

        // Score based on success history
        scores.putAll(getHistoryBasedRecommendations());

        // Sort by score and take top N
        List<Map.Entry<String, Double>> sorted = new ArrayList<>(scores.entrySet());
        sorted.sort((a, b) -> Double.compare(b.getValue(), a.getValue()));

        for (int i = 0; i < Math.min(maxRecommendations, sorted.size()); i++) {
            Map.Entry<String, Double> entry = sorted.get(i);
            String operation = entry.getKey();
            double score = entry.getValue();

            String reason = generateReason(operation, score, context);
            Map<String, String> params = getDefaultParams(operation, context);

            recommendations.add(new Recommendation(operation, params, score, reason));
        }

        // Notify listener
        if (listener != null) {
            listener.onRecommendation(recommendations);
        }

        Log.d(TAG, "Generated " + recommendations.size() + " recommendations");
        return recommendations;
    }

    /**
     * Get page-based recommendations
     */
    @NonNull
    private Map<String, Double> getPageBasedRecommendations(@NonNull String page) {
        Map<String, Double> scores = new HashMap<>();

        switch (page) {
            case "home":
                scores.put("search", 0.9);
                scores.put("goToCart", 0.3);
                break;

            case "search":
                scores.put("selectShop", 0.9);
                scores.put("selectProduct", 0.8);
                break;

            case "shop":
                scores.put("selectProduct", 0.9);
                scores.put("goToCart", 0.5);
                break;

            case "product":
                scores.put("configureOptions", 0.9);
                scores.put("addToCart", 0.85);
                break;

            case "cart":
                scores.put("goToCheckout", 0.95);
                scores.put("selectProduct", 0.3);
                break;

            case "checkout":
                scores.put("selectAddress", 0.9);
                scores.put("submitOrder", 0.85);
                break;

            default:
                scores.put("goBack", 0.5);
                scores.put("goToHome", 0.4);
        }

        return scores;
    }

    /**
     * Get sequence-based recommendations
     */
    @NonNull
    private Map<String, Double> getSequenceBasedRecommendations(@NonNull String lastOperation) {
        Map<String, Double> scores = new HashMap<>();

        String key = "sequence." + lastOperation;
        List<String> nextOps = operationSequences.get(key);

        if (nextOps != null) {
            // Count occurrences
            Map<String, Integer> counts = new HashMap<>();
            for (String op : nextOps) {
                counts.merge(op, 1, Integer::sum);
            }

            // Convert to scores
            int maxCount = counts.values().stream().max(Integer::compare).orElse(1);
            for (Map.Entry<String, Integer> entry : counts.entrySet()) {
                scores.put(entry.getKey(), (double) entry.getValue() / maxCount);
            }
        }

        return scores;
    }

    /**
     * Get history-based recommendations
     */
    @NonNull
    private Map<String, Double> getHistoryBasedRecommendations() {
        Map<String, Double> scores = new HashMap<>();

        for (Map.Entry<String, OperationStats> entry : operationStats.entrySet()) {
            OperationStats stats = entry.getValue();
            if (stats.count > 0) {
                // Score based on success rate and recency
                double recencyScore = Math.max(0, 1 - (System.currentTimeMillis() - stats.lastUsed) / (7 * 24 * 60 * 60 * 1000.0));
                double score = stats.getSuccessRate() * 0.7 + recencyScore * 0.3;
                scores.put(entry.getKey(), score);
            }
        }

        return scores;
    }

    /**
     * Generate reason for recommendation
     */
    @NonNull
    private String generateReason(@NonNull String operation, double score,
                                   @NonNull Map<String, String> context) {
        if (score > 0.8) {
            return "Highly recommended based on current context";
        } else if (score > 0.6) {
            return "Recommended based on history";
        } else if (score > 0.4) {
            return "Suggested as next step";
        } else {
            return "Available operation";
        }
    }

    /**
     * Get default params for operation
     */
    @NonNull
    private Map<String, String> getDefaultParams(@NonNull String operation,
                                                  @NonNull Map<String, String> context) {
        Map<String, String> params = new HashMap<>();

        // Try to get from memory
        if (memoryManager != null) {
            String key = "operation.default." + operation;
            String value = memoryManager.get(key);
            if (value != null) {
                try {
                    JSONObject json = new JSONObject(value);
                    for (String k : json.keySet()) {
                        params.put(k, json.getString(k));
                    }
                } catch (Exception e) {
                    // Ignore
                }
            }
        }

        return params;
    }

    /**
     * Record operation execution
     */
    public void recordOperation(@NonNull String operation, boolean success, long timeMs) {
        OperationStats stats = operationStats.computeIfAbsent(operation, k -> new OperationStats());
        stats.record(success, timeMs);

        // Save to memory
        if (memoryManager != null) {
            try {
                JSONObject record = new JSONObject();
                record.put("count", stats.count);
                record.put("successCount", stats.successCount);
                record.put("successRate", stats.getSuccessRate());
                record.put("avgTime", stats.getAverageTime());
                memoryManager.set("operation.stats." + operation, record.toString());
            } catch (Exception e) {
                Log.w(TAG, "Failed to save operation stats: " + e.getMessage());
            }
        }
    }

    /**
     * Record operation sequence
     */
    public void recordSequence(@NonNull String fromOperation, @NonNull String toOperation) {
        String key = "sequence." + fromOperation;
        List<String> sequence = operationSequences.computeIfAbsent(key, k -> new ArrayList<>());
        sequence.add(toOperation);

        // Keep only recent N entries
        if (sequence.size() > 100) {
            sequence.remove(0);
        }
    }

    /**
     * Load stats from memory
     */
    private void loadFromMemory() {
        if (memoryManager == null) return;

        // Load operation stats
        List<UserMemoryManager.MemorySuggestion> stats =
            memoryManager.getTopValues("operation.stats", 100);

        for (UserMemoryManager.MemorySuggestion s : stats) {
            try {
                JSONObject json = new JSONObject(s.value);
                OperationStats os = new OperationStats();
                os.count = json.optInt("count", 0);
                os.successCount = json.optInt("successCount", 0);
                // Restore other fields...
                operationStats.put(s.key.replace("operation.stats.", ""), os);
            } catch (Exception e) {
                // Ignore
            }
        }

        Log.d(TAG, "Loaded " + operationStats.size() + " operation stats from memory");
    }

    /**
     * Get operation statistics
     */
    @NonNull
    public Map<String, OperationStats> getOperationStats() {
        return new HashMap<>(operationStats);
    }

    /**
     * Clear all learning
     */
    public void clearLearning() {
        operationStats.clear();
        operationSequences.clear();

        if (memoryManager != null) {
            memoryManager.deleteByKeyPrefix("operation.stats.");
            memoryManager.deleteByKeyPrefix("sequence.");
        }

        Log.i(TAG, "Cleared all learning data");
    }
}