package com.ofa.agent.automation.adapter.travel;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.automation.AutomationNode;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.automation.BySelector;
import com.ofa.agent.automation.adapter.BaseAppAdapter;

import java.util.Arrays;
import java.util.List;
import java.util.Map;

/**
 * Didi (滴滴出行) App Adapter
 * Supports ride-hailing, destination search, and order management
 */
public class DidiAdapter extends BaseAppAdapter {

    // Package name
    private static final String PACKAGE_NAME = "com.sdu.didi.psnger";

    // UI Element IDs
    private static final String ID_SEARCH_INPUT = "com.sdu.didi.psnger:id/search_input";
    private static final String ID_DEST_INPUT = "com.sdu.didi.psnger:id/destination_input";
    private static final String ID_CALL_CAR = "com.sdu.didi.psnger:id/call_car_btn";
    private static final String ID_CANCEL_ORDER = "com.sdu.didi.psnger:id/cancel_order";
    private static final String ID_CONTACT_DRIVER = "com.sdu.didi.psnger:id/contact_driver";

    // Didi-specific operations
    private static final List<String> DIDI_OPERATIONS = Arrays.asList(
        "search", "goBack", "goToHome",
        // Didi-specific
        "setPickup", "setDestination", "selectCarType", "callCar",
        "cancelOrder", "contactDriver", "getTripStatus",
        "estimatePrice", "scheduleRide", "payForRide", "rateDriver"
    );

    // Car types
    private static final String[] CAR_TYPES = {"快车", "专车", "出租车", "顺风车", "优享", "拼车"};

    @Override
    @NonNull
    public String getPackageName() {
        return PACKAGE_NAME;
    }

    @Override
    @NonNull
    public String getAppName() {
        return "Didi";
    }

    @Override
    @NonNull
    public List<String> getSupportedOperations() {
        return DIDI_OPERATIONS;
    }

    @Override
    public float canHandle(@NonNull AutomationEngine engine) {
        // Check for Didi-specific elements
        if (hasId(engine, "com.sdu.didi.psnger:id/")) {
            return 0.95f;
        }

        // Check for common Didi texts
        String[] didiTexts = {"呼叫", "目的地", "你要去哪里", "快车", "专车", "滴滴"};
        int matchCount = 0;
        for (String text : didiTexts) {
            if (hasText(engine, text)) {
                matchCount++;
            }
        }

        return matchCount >= 2 ? 0.8f : 0.0f;
    }

    @Override
    @NonNull
    public String detectPage(@NonNull AutomationEngine engine) {
        if (hasText(engine, "你要去哪里") || hasId(engine, ID_DEST_INPUT)) {
            return "destination_input";
        }

        if (hasText(engine, "呼叫") || hasText(engine, "立即呼叫")) {
            return "car_selection";
        }

        if (hasText(engine, "正在为您叫车") || hasText(engine, "司机接单")) {
            return "waiting_driver";
        }

        if (hasText(engine, "行程中") || hasText(engine, "预计到达")) {
            return "in_trip";
        }

        if (hasText(engine, "支付") || hasText(engine, "去支付")) {
            return "payment";
        }

        if (hasText(engine, "评价") || hasText(engine, "给司机评分")) {
            return "rating";
        }

        if (hasText(engine, "我的订单")) {
            return "orders";
        }

        return "home";
    }

    @Override
    @NonNull
    public AutomationResult search(@NonNull AutomationEngine engine, @NonNull String query) {
        Log.d(TAG, "Searching for: " + query);

        // Click "你要去哪里" input
        AutomationResult result = clickByText(engine, "你要去哪里", false);
        if (!result.isSuccess()) {
            result = clickByText(engine, "目的地", false);
        }

        if (!result.isSuccess()) {
            return error("search", "Failed to open destination input");
        }

        waitForPage(engine, 500);

        // Input destination
        result = engine.inputText(query);
        if (!result.isSuccess()) {
            AutomationNode inputNode = engine.findElement(
                BySelector.className("android.widget.EditText")
            );
            if (inputNode != null) {
                result = engine.inputText(query);
            }
        }

        if (!result.isSuccess()) {
            return error("search", "Failed to input destination");
        }

        waitForPage(engine, 1000);

        // Select first result
        return selectFirstResult(engine);
    }

    /**
     * Set pickup location
     */
    @NonNull
    public AutomationResult setPickup(@NonNull AutomationEngine engine,
                                      @NonNull String pickup) {
        Log.d(TAG, "Setting pickup: " + pickup);

        // Click current location text
        AutomationResult result = clickByText(engine, "当前位置", false);
        if (!result.isSuccess()) {
            result = clickByText(engine, "出发地", false);
        }

        if (!result.isSuccess()) {
            return error("setPickup", "Failed to open pickup input");
        }

        waitForPage(engine, 500);

        // Input pickup location
        result = engine.inputText(pickup);
        if (!result.isSuccess()) {
            return error("setPickup", "Failed to input pickup");
        }

        waitForPage(engine, 1000);

        // Select first result
        return selectFirstResult(engine);
    }

    /**
     * Set destination
     */
    @NonNull
    public AutomationResult setDestination(@NonNull AutomationEngine engine,
                                           @NonNull String destination) {
        Log.d(TAG, "Setting destination: " + destination);
        return search(engine, destination);
    }

    /**
     * Select car type
     */
    @NonNull
    public AutomationResult selectCarType(@NonNull AutomationEngine engine,
                                          @NonNull String carType) {
        Log.d(TAG, "Selecting car type: " + carType);

        // Click on the car type tab
        return clickByText(engine, carType, true);
    }

    /**
     * Call a car (request ride)
     */
    @NonNull
    public AutomationResult callCar(@NonNull AutomationEngine engine,
                                    @Nullable String carType) {
        Log.d(TAG, "Calling car, type: " + carType);

        // Select car type if specified
        if (carType != null) {
            selectCarType(engine, carType);
            waitForPage(engine, 500);
        }

        // Click "呼叫" or "立即呼叫" button
        AutomationResult result = clickByText(engine, "呼叫", false);
        if (!result.isSuccess()) {
            result = clickByText(engine, "立即呼叫", false);
        }

        if (!result.isSuccess()) {
            result = clickByText(engine, "确认呼叫", false);
        }

        return result;
    }

    /**
     * Cancel current order
     */
    @NonNull
    public AutomationResult cancelOrder(@NonNull AutomationEngine engine,
                                        @Nullable String reason) {
        Log.d(TAG, "Cancelling order");

        // Click cancel button
        AutomationResult result = clickByText(engine, "取消订单", false);
        if (!result.isSuccess()) {
            result = clickById(engine, ID_CANCEL_ORDER);
        }

        if (!result.isSuccess()) {
            return error("cancelOrder", "Cancel button not found");
        }

        waitForPage(engine, 500);

        // Select cancellation reason if prompted
        if (reason != null) {
            scrollAndClick(engine, reason, 3);
        } else {
            // Select first reason
            AutomationNode reasonNode = engine.findElement(
                BySelector.className("android.widget.RadioButton")
            );
            if (reasonNode != null) {
                engine.click(reasonNode.getCenterX(), reasonNode.getCenterY());
            }
        }

        // Confirm cancellation
        return clickByText(engine, "确认取消", false);
    }

    /**
     * Contact driver
     */
    @NonNull
    public AutomationResult contactDriver(@NonNull AutomationEngine engine,
                                          @NonNull String method) {
        Log.d(TAG, "Contacting driver via: " + method);

        // Click contact button
        AutomationResult result = clickById(engine, ID_CONTACT_DRIVER);
        if (!result.isSuccess()) {
            result = clickByText(engine, "联系司机", false);
        }

        if (!result.isSuccess()) {
            return error("contactDriver", "Contact button not found");
        }

        waitForPage(engine, 300);

        // Select contact method
        switch (method.toLowerCase()) {
            case "call":
            case "电话":
                return clickByText(engine, "电话", false);
            case "message":
            case "消息":
            case "短信":
                return clickByText(engine, "短信", false);
            default:
                return clickByText(engine, method, false);
        }
    }

    /**
     * Get current trip status
     */
    @NonNull
    public TripStatus getTripStatus(@NonNull AutomationEngine engine) {
        Log.d(TAG, "Getting trip status");

        String page = detectPage(engine);

        switch (page) {
            case "car_selection":
                return new TripStatus("idle", "No active trip", null, null);
            case "waiting_driver":
                return new TripStatus("waiting", "Waiting for driver", null, null);
            case "in_trip":
                // Extract ETA if visible
                String eta = extractETA(engine);
                return new TripStatus("in_trip", "Trip in progress", eta, null);
            case "payment":
                String amount = extractAmount(engine);
                return new TripStatus("pending_payment", "Payment required", null, amount);
            case "rating":
                return new TripStatus("completed", "Trip completed, rating required", null, null);
            default:
                return new TripStatus("unknown", "Unknown status", null, null);
        }
    }

    /**
     * Estimate price for a trip
     */
    @NonNull
    public PriceEstimate estimatePrice(@NonNull AutomationEngine engine,
                                       @NonNull String destination) {
        Log.d(TAG, "Estimating price to: " + destination);

        // Set destination
        AutomationResult result = setDestination(engine, destination);
        if (!result.isSuccess()) {
            return new PriceEstimate(null, null, "Failed to set destination");
        }

        waitForPage(engine, 2000);

        // Extract prices for different car types
        Map<String, String> prices = new java.util.HashMap<>();

        for (String carType : CAR_TYPES) {
            // Click on car type
            selectCarType(engine, carType);
            waitForPage(engine, 500);

            // Extract price
            String price = extractPrice(engine);
            if (price != null) {
                prices.put(carType, price);
            }
        }

        return new PriceEstimate(prices, null, null);
    }

    /**
     * Schedule a ride for later
     */
    @NonNull
    public AutomationResult scheduleRide(@NonNull AutomationEngine engine,
                                         @NonNull String time) {
        Log.d(TAG, "Scheduling ride for: " + time);

        // Click schedule button
        AutomationResult result = clickByText(engine, "预约", false);
        if (!result.isSuccess()) {
            result = clickByText(engine, "现在出发", false);
        }

        if (!result.isSuccess()) {
            return error("scheduleRide", "Schedule button not found");
        }

        waitForPage(engine, 500);

        // Select time
        // TODO: Implement time picker interaction
        return success("scheduleRide");
    }

    /**
     * Pay for completed ride
     */
    @NonNull
    public AutomationResult payForRide(@NonNull AutomationEngine engine,
                                       @Nullable String paymentMethod) {
        Log.d(TAG, "Paying for ride");

        // Click pay button
        AutomationResult result = clickByText(engine, "去支付", false);
        if (!result.isSuccess()) {
            result = clickByText(engine, "支付", false);
        }

        if (!result.isSuccess()) {
            return error("payForRide", "Payment button not found");
        }

        waitForPage(engine, 500);

        // Select payment method if specified
        if (paymentMethod != null) {
            String[] methodTexts;
            switch (paymentMethod.toLowerCase()) {
                case "wechat":
                case "微信":
                    methodTexts = new String[]{"微信支付", "微信"};
                    break;
                case "alipay":
                case "支付宝":
                    methodTexts = new String[]{"支付宝", "支付宝支付"};
                    break;
                default:
                    methodTexts = new String[]{paymentMethod};
            }

            for (String method : methodTexts) {
                clickByText(engine, method, false);
            }
        }

        // Confirm payment
        return clickByText(engine, "确认支付", false);
    }

    /**
     * Rate driver after trip
     */
    @NonNull
    public AutomationResult rateDriver(@NonNull AutomationEngine engine, int rating) {
        Log.d(TAG, "Rating driver: " + rating);

        // Click on stars
        for (int i = 0; i < rating; i++) {
            // Click on star icons (usually from left to right)
            AutomationNode starNode = engine.findElement(
                BySelector.id("star_" + (i + 1))
            );
            if (starNode == null) {
                starNode = engine.findElement(
                    BySelector.className("android.widget.ImageView")
                        .textContains("star")
                );
            }

            if (starNode != null) {
                engine.click(starNode.getCenterX(), starNode.getCenterY());
            }
        }

        // Submit rating
        return clickByText(engine, "提交", false);
    }

    // ===== Helper Methods =====

    /**
     * Select first search result
     */
    private AutomationResult selectFirstResult(@NonNull AutomationEngine engine) {
        // Click first item in results list
        AutomationNode firstResult = engine.findElement(
            BySelector.className("android.widget.LinearLayout")
                .clickable(true)
        );

        if (firstResult != null) {
            return engine.click(firstResult.getCenterX(), firstResult.getCenterY());
        }

        // Fallback: click any visible location text
        String[] locationHints = {"米", "km", "公里"};
        for (String hint : locationHints) {
            AutomationNode node = engine.findElement(BySelector.textContains(hint));
            if (node != null) {
                return engine.click(node.getCenterX(), node.getCenterY());
            }
        }

        return error("selectFirstResult", "No results found");
    }

    /**
     * Extract ETA from current page
     */
    private String extractETA(@NonNull AutomationEngine engine) {
        // Look for time patterns like "预计15分钟到达"
        String pageSource = engine.getPageSource();
        // TODO: Parse ETA from page source
        return null;
    }

    /**
     * Extract amount from current page
     */
    private String extractAmount(@NonNull AutomationEngine engine) {
        // Look for price patterns like "¥12.5"
        AutomationNode priceNode = engine.findElement(
            BySelector.textContains("¥")
        );
        if (priceNode != null) {
            return priceNode.getText();
        }
        return null;
    }

    /**
     * Extract price from current page
     */
    private String extractPrice(@NonNull AutomationEngine engine) {
        // Look for estimated price
        AutomationNode priceNode = engine.findElement(
            BySelector.textContains("约")
                .or(BySelector.textContains("预估"))
        );
        if (priceNode != null) {
            String text = priceNode.getText();
            // Extract price from "约¥12.5" format
            int idx = text.indexOf("¥");
            if (idx >= 0) {
                return text.substring(idx);
            }
        }
        return null;
    }

    // ===== Placeholder implementations for BaseAppAdapter =====

    @Override
    @NonNull
    public AutomationResult selectShop(@NonNull AutomationEngine engine,
                                       @NonNull String shopName) {
        return error("selectShop", "Not applicable for Didi");
    }

    @Override
    @NonNull
    public AutomationResult selectProduct(@NonNull AutomationEngine engine,
                                          @NonNull String productName) {
        return error("selectProduct", "Not applicable for Didi");
    }

    @Override
    @NonNull
    public AutomationResult configureOptions(@NonNull AutomationEngine engine,
                                             @NonNull Map<String, String> options) {
        return error("configureOptions", "Not applicable for Didi");
    }

    @Override
    @NonNull
    public AutomationResult addToCart(@NonNull AutomationEngine engine, int quantity) {
        return error("addToCart", "Not applicable for Didi");
    }

    @Override
    @NonNull
    public AutomationResult goToCart(@NonNull AutomationEngine engine) {
        return error("goToCart", "Not applicable for Didi");
    }

    @Override
    @NonNull
    public AutomationResult goToCheckout(@NonNull AutomationEngine engine) {
        return error("goToCheckout", "Not applicable for Didi");
    }

    @Override
    @NonNull
    public AutomationResult selectAddress(@NonNull AutomationEngine engine,
                                          @NonNull String addressHint) {
        return error("selectAddress", "Not applicable for Didi");
    }

    @Override
    @NonNull
    public AutomationResult submitOrder(@NonNull AutomationEngine engine) {
        return error("submitOrder", "Not applicable for Didi");
    }

    @Override
    @NonNull
    public AutomationResult pay(@NonNull AutomationEngine engine,
                                @NonNull String paymentMethod) {
        return payForRide(engine, paymentMethod);
    }

    @Override
    @NonNull
    public OrderStatus getOrderStatus(@NonNull AutomationEngine engine,
                                      @Nullable String orderId) {
        TripStatus tripStatus = getTripStatus(engine);
        return new OrderStatus(
            tripStatus.status,
            tripStatus.statusText,
            0,
            null
        );
    }

    // ===== Data Classes =====

    /**
     * Trip status information
     */
    public static class TripStatus {
        public final String status;
        public final String statusText;
        public final String eta;
        public final String amount;

        public TripStatus(String status, String statusText, String eta, String amount) {
            this.status = status;
            this.statusText = statusText;
            this.eta = eta;
            this.amount = amount;
        }
    }

    /**
     * Price estimate information
     */
    public static class PriceEstimate {
        public final Map<String, String> pricesByCarType;
        public final String estimatedTime;
        public final String error;

        public PriceEstimate(Map<String, String> pricesByCarType,
                            String estimatedTime, String error) {
            this.pricesByCarType = pricesByCarType;
            this.estimatedTime = estimatedTime;
            this.error = error;
        }
    }
}