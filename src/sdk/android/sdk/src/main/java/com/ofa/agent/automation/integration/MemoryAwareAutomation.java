package com.ofa.agent.automation.integration;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.automation.BySelector;
import com.ofa.agent.memory.UserMemoryManager;

import org.json.JSONObject;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Memory-Aware Automation - integrates with Memory system for smart automation.
 * Remembers user preferences and uses them to optimize automation decisions.
 */
public class MemoryAwareAutomation {

    private static final String TAG = "MemoryAwareAutomation";

    private final Context context;
    private final AutomationEngine engine;
    private final UserMemoryManager memoryManager;

    // Memory keys for automation preferences
    private static final String KEY_PREFERRED_SHOPS = "automation.preferred_shops";
    private static final String KEY_PREFERRED_PRODUCTS = "automation.preferred_products";
    private static final String KEY_PREFERRED_OPTIONS = "automation.preferred_options";
    private static final String KEY_PREFERRED_ADDRESS = "automation.preferred_address";
    private static final String KEY_PREFERRED_PAYMENT = "automation.preferred_payment";
    private static final String KEY_SEARCH_HISTORY = "automation.search_history";
    private static final String KEY_ORDER_HISTORY = "automation.order_history";

    /**
     * Create memory-aware automation
     */
    public MemoryAwareAutomation(@NonNull Context context,
                                  @NonNull AutomationEngine engine,
                                  @NonNull UserMemoryManager memoryManager) {
        this.context = context;
        this.engine = engine;
        this.memoryManager = memoryManager;
    }

    // ===== Preference Retrieval =====

    /**
     * Get preferred shop for a category
     */
    @Nullable
    public String getPreferredShop(@NonNull String category) {
        String key = KEY_PREFERRED_SHOPS + "." + category;
        String value = memoryManager.get(key);
        if (value != null) {
            Log.d(TAG, "Found preferred shop for " + category + ": " + value);
        }
        return value;
    }

    /**
     * Get preferred product options
     */
    @NonNull
    public Map<String, String> getPreferredOptions(@NonNull String productType) {
        Map<String, String> options = new HashMap<>();

        String key = KEY_PREFERRED_OPTIONS + "." + productType;
        String value = memoryManager.get(key);

        if (value != null) {
            try {
                JSONObject json = new JSONObject(value);
                for (String k : json.keySet()) {
                    options.put(k, json.getString(k));
                }
                Log.d(TAG, "Found preferred options for " + productType + ": " + options);
            } catch (Exception e) {
                Log.w(TAG, "Failed to parse preferred options: " + e.getMessage());
            }
        }

        return options;
    }

    /**
     * Get preferred address
     */
    @Nullable
    public String getPreferredAddress() {
        return memoryManager.get(KEY_PREFERRED_ADDRESS);
    }

    /**
     * Get preferred payment method
     */
    @Nullable
    public String getPreferredPaymentMethod() {
        return memoryManager.get(KEY_PREFERRED_PAYMENT);
    }

    /**
     * Get search history
     */
    @NonNull
    public List<String> getSearchHistory(int limit) {
        List<UserMemoryManager.MemorySuggestion> suggestions =
            memoryManager.getSuggestions(KEY_SEARCH_HISTORY, limit);
        return suggestions.stream()
            .map(s -> s.value)
            .collect(java.util.stream.Collectors.toList());
    }

    // ===== Preference Storage =====

    /**
     * Remember preferred shop
     */
    public void rememberPreferredShop(@NonNull String category, @NonNull String shopName) {
        String key = KEY_PREFERRED_SHOPS + "." + category;
        memoryManager.set(key, shopName, Map.of("category", category));
        Log.i(TAG, "Remembered preferred shop: " + category + " -> " + shopName);
    }

    /**
     * Remember product options
     */
    public void rememberOptions(@NonNull String productType, @NonNull Map<String, String> options) {
        String key = KEY_PREFERRED_OPTIONS + "." + productType;
        try {
            JSONObject json = new JSONObject(options);
            memoryManager.set(key, json.toString(), Map.of("productType", productType));
            Log.i(TAG, "Remembered options for " + productType + ": " + options);
        } catch (Exception e) {
            Log.w(TAG, "Failed to save options: " + e.getMessage());
        }
    }

    /**
     * Remember preferred address
     */
    public void rememberAddress(@NonNull String address) {
        memoryManager.set(KEY_PREFERRED_ADDRESS, address);
        Log.i(TAG, "Remembered address: " + address);
    }

    /**
     * Remember preferred payment method
     */
    public void rememberPaymentMethod(@NonNull String method) {
        memoryManager.set(KEY_PREFERRED_PAYMENT, method);
        Log.i(TAG, "Remembered payment method: " + method);
    }

    /**
     * Add to search history
     */
    public void addToSearchHistory(@NonNull String query) {
        memoryManager.set(KEY_SEARCH_HISTORY + "." + System.currentTimeMillis(), query);
        Log.d(TAG, "Added to search history: " + query);
    }

    /**
     * Record order completion
     */
    public void recordOrder(@NonNull String shopName, @NonNull String productName) {
        String key = KEY_ORDER_HISTORY + "." + System.currentTimeMillis();
        try {
            JSONObject order = new JSONObject();
            order.put("shop", shopName);
            order.put("product", productName);
            order.put("timestamp", System.currentTimeMillis());

            memoryManager.set(key, order.toString(), Map.of("type", "order"));
            Log.i(TAG, "Recorded order: " + shopName + " - " + productName);
        } catch (Exception e) {
            Log.w(TAG, "Failed to record order: " + e.getMessage());
        }
    }

    // ===== Smart Operations =====

    /**
     * Smart search with history suggestions
     */
    @NonNull
    public List<String> getSearchSuggestions(@NonNull String partialQuery) {
        List<UserMemoryManager.MemorySuggestion> suggestions =
            memoryManager.autocomplete(KEY_SEARCH_HISTORY, partialQuery, 5);
        return suggestions.stream()
            .map(s -> s.value)
            .collect(java.util.stream.Collectors.toList());
    }

    /**
     * Get recommended shops based on history
     */
    @NonNull
    public List<String> getRecommendedShops(@NonNull String category) {
        List<UserMemoryManager.MemorySuggestion> suggestions =
            memoryManager.getTopValues(KEY_PREFERRED_SHOPS + "." + category, 3);
        return suggestions.stream()
            .map(s -> s.value)
            .collect(java.util.stream.Collectors.toList());
    }

    /**
     * Apply preferred options to product configuration
     */
    @NonNull
    public Map<String, String> applyPreferredOptions(@NonNull String productType,
                                                      @Nullable Map<String, String> userOptions) {
        Map<String, String> preferred = getPreferredOptions(productType);
        Map<String, String> result = new HashMap<>(preferred);

        // User options override preferred
        if (userOptions != null) {
            result.putAll(userOptions);
        }

        return result;
    }

    /**
     * Build smart search query with context
     */
    @NonNull
    public String buildSmartQuery(@NonNull String baseQuery, @Nullable String category) {
        StringBuilder query = new StringBuilder(baseQuery);

        if (category != null) {
            String preferredShop = getPreferredShop(category);
            if (preferredShop != null) {
                query.append(" ").append(preferredShop);
            }
        }

        return query.toString();
    }

    // ===== Memory Analysis =====

    /**
     * Get memory statistics for automation
     */
    @NonNull
    public AutomationMemoryStats getStats() {
        AutomationMemoryStats stats = new AutomationMemoryStats();

        stats.preferredShopsCount = memoryManager.countKeys(KEY_PREFERRED_SHOPS);
        stats.searchHistoryCount = memoryManager.countKeys(KEY_SEARCH_HISTORY);
        stats.orderHistoryCount = memoryManager.countKeys(KEY_ORDER_HISTORY);
        stats.hasPreferredAddress = getPreferredAddress() != null;
        stats.hasPreferredPayment = getPreferredPaymentMethod() != null;

        return stats;
    }

    /**
     * Clear automation memory
     */
    public void clearMemory() {
        memoryManager.deleteByKeyPrefix(KEY_PREFERRED_SHOPS);
        memoryManager.deleteByKeyPrefix(KEY_PREFERRED_PRODUCTS);
        memoryManager.deleteByKeyPrefix(KEY_PREFERRED_OPTIONS);
        memoryManager.deleteByKeyPrefix(KEY_SEARCH_HISTORY);
        memoryManager.deleteByKeyPrefix(KEY_ORDER_HISTORY);
        memoryManager.delete(KEY_PREFERRED_ADDRESS);
        memoryManager.delete(KEY_PREFERRED_PAYMENT);

        Log.i(TAG, "Cleared all automation memory");
    }

    /**
     * Export automation preferences
     */
    @NonNull
    public JSONObject exportPreferences() {
        JSONObject export = new JSONObject();
        try {
            export.put("preferredAddress", getPreferredAddress());
            export.put("preferredPayment", getPreferredPaymentMethod());
            export.put("searchHistory", getSearchHistory(10));
        } catch (Exception e) {
            Log.w(TAG, "Failed to export preferences: " + e.getMessage());
        }
        return export;
    }

    /**
     * Import automation preferences
     */
    public void importPreferences(@NonNull JSONObject preferences) {
        try {
            if (preferences.has("preferredAddress")) {
                rememberAddress(preferences.getString("preferredAddress"));
            }
            if (preferences.has("preferredPayment")) {
                rememberPaymentMethod(preferences.getString("preferredPayment"));
            }
            if (preferences.has("searchHistory")) {
                // Handle search history import
            }
        } catch (Exception e) {
            Log.w(TAG, "Failed to import preferences: " + e.getMessage());
        }
    }

    /**
     * Memory statistics
     */
    public static class AutomationMemoryStats {
        public int preferredShopsCount;
        public int searchHistoryCount;
        public int orderHistoryCount;
        public boolean hasPreferredAddress;
        public boolean hasPreferredPayment;

        @NonNull
        @Override
        public String toString() {
            return "AutomationMemoryStats{" +
                "shops=" + preferredShopsCount +
                ", searches=" + searchHistoryCount +
                ", orders=" + orderHistoryCount +
                ", address=" + hasPreferredAddress +
                ", payment=" + hasPreferredPayment +
                '}';
        }
    }
}