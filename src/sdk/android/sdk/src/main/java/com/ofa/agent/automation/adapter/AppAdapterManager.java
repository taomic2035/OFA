package com.ofa.agent.automation.adapter;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationEngine;

import java.util.ArrayList;
import java.util.Collections;
import java.util.Comparator;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * App Adapter Manager - manages app-specific adapters.
 * Automatically selects the appropriate adapter based on current app.
 */
public class AppAdapterManager {

    private static final String TAG = "AppAdapterManager";

    private final Map<String, AppAdapter> adapters = new ConcurrentHashMap<>();
    private final List<AppAdapter> adapterList = new ArrayList<>();

    private AppAdapter currentAdapter;
    private String currentPackage;
    private float currentConfidence;

    /**
     * Register an adapter
     */
    public void register(@NonNull AppAdapter adapter) {
        String packageName = adapter.getPackageName();
        adapters.put(packageName, adapter);
        adapterList.add(adapter);
        Log.i(TAG, "Registered adapter: " + adapter.getAppName() + " (" + packageName + ")");
    }

    /**
     * Unregister an adapter
     */
    public void unregister(@NonNull String packageName) {
        AppAdapter removed = adapters.remove(packageName);
        if (removed != null) {
            adapterList.remove(removed);
            Log.i(TAG, "Unregistered adapter for: " + packageName);
        }
    }

    /**
     * Get adapter by package name
     */
    @Nullable
    public AppAdapter getAdapter(@NonNull String packageName) {
        return adapters.get(packageName);
    }

    /**
     * Get all registered adapters
     */
    @NonNull
    public List<AppAdapter> getAllAdapters() {
        return new ArrayList<>(adapterList);
    }

    /**
     * Detect and return the best adapter for current state
     * @param engine Automation engine
     * @return Best matching adapter or null
     */
    @Nullable
    public AppAdapter detectAdapter(@NonNull AutomationEngine engine) {
        String foregroundPackage = engine.getForegroundPackage();
        if (foregroundPackage == null) {
            return null;
        }

        // First, try exact package match
        AppAdapter exactMatch = adapters.get(foregroundPackage);
        if (exactMatch != null) {
            float confidence = exactMatch.canHandle(engine);
            if (confidence > 0.5f) {
                currentAdapter = exactMatch;
                currentPackage = foregroundPackage;
                currentConfidence = confidence;
                return exactMatch;
            }
        }

        // Try all adapters and pick the one with highest confidence
        AppAdapter bestAdapter = null;
        float bestConfidence = 0;

        for (AppAdapter adapter : adapterList) {
            float confidence = adapter.canHandle(engine);
            if (confidence > bestConfidence) {
                bestConfidence = confidence;
                bestAdapter = adapter;
            }
        }

        if (bestAdapter != null && bestConfidence > 0.3f) {
            currentAdapter = bestAdapter;
            currentPackage = foregroundPackage;
            currentConfidence = bestConfidence;
            return bestAdapter;
        }

        return null;
    }

    /**
     * Get current active adapter
     */
    @Nullable
    public AppAdapter getCurrentAdapter() {
        return currentAdapter;
    }

    /**
     * Get confidence for current adapter
     */
    public float getCurrentConfidence() {
        return currentConfidence;
    }

    /**
     * Check if adapter is available for package
     */
    public boolean hasAdapter(@NonNull String packageName) {
        return adapters.containsKey(packageName);
    }

    /**
     * Check if current app has adapter
     */
    public boolean hasAdapterForCurrentApp(@NonNull AutomationEngine engine) {
        String pkg = engine.getForegroundPackage();
        return pkg != null && hasAdapter(pkg);
    }

    /**
     * Execute operation on current adapter
     * @param engine Automation engine
     * @param operation Operation name
     * @param params Operation parameters
     * @return Result of operation
     */
    @NonNull
    public AutomationResult execute(@NonNull AutomationEngine engine,
                                     @NonNull String operation,
                                     @NonNull Map<String, String> params) {
        AppAdapter adapter = detectAdapter(engine);
        if (adapter == null) {
            return new com.ofa.agent.automation.AutomationResult(
                operation, "No adapter found for current app");
        }

        return executeOperation(adapter, engine, operation, params);
    }

    /**
     * Execute operation on specific adapter
     */
    @NonNull
    private AutomationResult executeOperation(@NonNull AppAdapter adapter,
                                               @NonNull AutomationEngine engine,
                                               @NonNull String operation,
                                               @NonNull Map<String, String> params) {
        if (!adapter.getSupportedOperations().contains(operation)) {
            return new com.ofa.agent.automation.AutomationResult(
                operation, "Operation not supported: " + operation);
        }

        switch (operation) {
            case "search":
                String query = params.get("query");
                if (query != null) {
                    return adapter.search(engine, query);
                }
                break;

            case "selectShop":
                String shopName = params.get("shopName");
                if (shopName != null) {
                    return adapter.selectShop(engine, shopName);
                }
                break;

            case "selectProduct":
                String productName = params.get("productName");
                if (productName != null) {
                    return adapter.selectProduct(engine, productName);
                }
                break;

            case "configureOptions":
                return adapter.configureOptions(engine, params);

            case "addToCart":
                int quantity = 1;
                String qtyStr = params.get("quantity");
                if (qtyStr != null) {
                    try {
                        quantity = Integer.parseInt(qtyStr);
                    } catch (NumberFormatException e) {
                        // Use default
                    }
                }
                return adapter.addToCart(engine, quantity);

            case "goToCart":
                return adapter.goToCart(engine);

            case "goToCheckout":
                return adapter.goToCheckout(engine);

            case "selectAddress":
                String addressHint = params.get("address");
                if (addressHint != null) {
                    return adapter.selectAddress(engine, addressHint);
                }
                break;

            case "submitOrder":
                return adapter.submitOrder(engine);

            case "pay":
                String paymentMethod = params.get("paymentMethod");
                if (paymentMethod == null) {
                    paymentMethod = "alipay"; // Default
                }
                return adapter.pay(engine, paymentMethod);

            case "goBack":
                return adapter.goBack(engine);

            case "goToHome":
                return adapter.goToHome(engine);
        }

        return new com.ofa.agent.automation.AutomationResult(
            operation, "Unknown operation or missing parameters: " + operation);
    }

    /**
     * Get supported operations for current app
     */
    @NonNull
    public List<String> getSupportedOperations(@NonNull AutomationEngine engine) {
        AppAdapter adapter = detectAdapter(engine);
        if (adapter == null) {
            return Collections.emptyList();
        }
        return adapter.getSupportedOperations();
    }

    /**
     * Get current page type
     */
    @NonNull
    public String getCurrentPage(@NonNull AutomationEngine engine) {
        AppAdapter adapter = detectAdapter(engine);
        if (adapter == null) {
            return "unknown";
        }
        return adapter.detectPage(engine);
    }

    /**
     * Clear all adapters
     */
    public void clear() {
        adapters.clear();
        adapterList.clear();
        currentAdapter = null;
        Log.i(TAG, "All adapters cleared");
    }

    /**
     * Get adapter count
     */
    public int getAdapterCount() {
        return adapters.size();
    }
}