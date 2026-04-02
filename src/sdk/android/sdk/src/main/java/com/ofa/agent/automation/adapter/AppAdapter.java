package com.ofa.agent.automation.adapter;

import android.graphics.Bitmap;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.automation.AutomationNode;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.automation.BySelector;

import java.util.List;
import java.util.Map;

/**
 * App Adapter interface - defines operations for specific apps.
 * Each app (Meituan, Eleme, Taobao, etc.) can have its own adapter
 * that understands the app's UI structure and provides common operations.
 */
public interface AppAdapter {

    /**
     * Get the package name this adapter handles
     */
    @NonNull
    String getPackageName();

    /**
     * Get the friendly app name
     */
    @NonNull
    String getAppName();

    /**
     * Get supported operations
     */
    @NonNull
    List<String> getSupportedOperations();

    /**
     * Check if this adapter can handle the current page
     * @param engine Automation engine
     * @return Confidence score (0.0 - 1.0), higher means more confident
     */
    float canHandle(@NonNull AutomationEngine engine);

    /**
     * Detect current page type
     * @param engine Automation engine
     * @return Page type identifier (e.g., "home", "search", "product_detail", "cart")
     */
    @NonNull
    String detectPage(@NonNull AutomationEngine engine);

    /**
     * Search for shop or product
     * @param engine Automation engine
     * @param query Search query
     * @return Result of search operation
     */
    @NonNull
    AutomationResult search(@NonNull AutomationEngine engine, @NonNull String query);

    /**
     * Select a shop from search results
     * @param engine Automation engine
     * @param shopName Shop name or partial name
     * @return Result of selection
     */
    @NonNull
    AutomationResult selectShop(@NonNull AutomationEngine engine, @NonNull String shopName);

    /**
     * Select a product from shop
     * @param engine Automation engine
     * @param productName Product name or partial name
     * @return Result of selection
     */
    @NonNull
    AutomationResult selectProduct(@NonNull AutomationEngine engine, @NonNull String productName);

    /**
     * Configure product options (size, color, specs, etc.)
     * @param engine Automation engine
     * @param options Map of option name to value
     * @return Result of configuration
     */
    @NonNull
    AutomationResult configureOptions(@NonNull AutomationEngine engine,
                                       @NonNull Map<String, String> options);

    /**
     * Add current product to cart
     * @param engine Automation engine
     * @param quantity Quantity to add
     * @return Result of add to cart
     */
    @NonNull
    AutomationResult addToCart(@NonNull AutomationEngine engine, int quantity);

    /**
     * Go to cart page
     * @param engine Automation engine
     * @return Result of navigation
     */
    @NonNull
    AutomationResult goToCart(@NonNull AutomationEngine engine);

    /**
     * Go to checkout page
     * @param engine Automation engine
     * @return Result of navigation
     */
    @NonNull
    AutomationResult goToCheckout(@NonNull AutomationEngine engine);

    /**
     * Select delivery address
     * @param engine Automation engine
     * @param addressHint Address hint (name, partial address, or index)
     * @return Result of selection
     */
    @NonNull
    AutomationResult selectAddress(@NonNull AutomationEngine engine, @NonNull String addressHint);

    /**
     * Submit order (without payment)
     * @param engine Automation engine
     * @return Result of submission
     */
    @NonNull
    AutomationResult submitOrder(@NonNull AutomationEngine engine);

    /**
     * Pay for order
     * @param engine Automation engine
     * @param paymentMethod Payment method ("alipay", "wechat", "card", etc.)
     * @return Result of payment
     */
    @NonNull
    AutomationResult pay(@NonNull AutomationEngine engine, @NonNull String paymentMethod);

    /**
     * Get order status
     * @param engine Automation engine
     * @param orderId Order ID (optional, use last order if null)
     * @return Order status info
     */
    @NonNull
    OrderStatus getOrderStatus(@NonNull AutomationEngine engine, @Nullable String orderId);

    /**
     * Navigate back
     * @param engine Automation engine
     * @return Result of navigation
     */
    @NonNull
    AutomationResult goBack(@NonNull AutomationEngine engine);

    /**
     * Go to home page of the app
     * @param engine Automation engine
     * @return Result of navigation
     */
    @NonNull
    AutomationResult goToHome(@NonNull AutomationEngine engine);

    /**
     * Order status information
     */
    class OrderStatus {
        public final String status; // "pending", "confirmed", "preparing", "delivering", "delivered", "cancelled"
        public final String statusText;
        public final long estimatedDeliveryTime;
        public final String orderId;

        public OrderStatus(String status, String statusText, long estimatedDeliveryTime, String orderId) {
            this.status = status;
            this.statusText = statusText;
            this.estimatedDeliveryTime = estimatedDeliveryTime;
            this.orderId = orderId;
        }
    }
}