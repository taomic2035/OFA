package com.ofa.agent.automation.adapter.social;

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
 * Douyin (抖音) App Adapter
 * Supports video browsing, searching, liking, and shopping operations
 */
public class DouyinAdapter extends BaseAppAdapter {

    // Package name
    private static final String PACKAGE_NAME = "com.ss.android.ugc.aweme";

    // UI Element IDs (may vary by version)
    private static final String ID_SEARCH_INPUT = "com.ss.android.ugc.aweme:id/et_search";
    private static final String ID_SEARCH_BUTTON = "com.ss.android.ugc.aweme:id/tv_search";
    private static final String ID_TAB_HOME = "com.ss.android.ugc.aweme:id/tab_home";
    private static final String ID_TAB_MALL = "com.ss.android.ugc.aweme:id/tab_mall";
    private static final String ID_LIKE_BUTTON = "com.ss.android.ugc.aweme:id/like_button";
    private static final String ID_COMMENT_BUTTON = "com.ss.android.ugc.aweme:id/comment_button";
    private static final String ID_SHARE_BUTTON = "com.ss.android.ugc.aweme:id/share_button";
    private static final String ID_FOLLOW_BUTTON = "com.ss.android.ugc.aweme:id/follow_button";

    // Douyin-specific operations
    private static final List<String> DOUYIN_OPERATIONS = Arrays.asList(
        "search", "selectShop", "selectProduct", "configureOptions",
        "addToCart", "goToCart", "goToCheckout", "selectAddress",
        "submitOrder", "pay", "goBack", "goToHome",
        // Douyin-specific
        "watchVideo", "likeVideo", "commentVideo", "followUser",
        "shareVideo", "browseFeed", "openMall", "searchUser"
    );

    @Override
    @NonNull
    public String getPackageName() {
        return PACKAGE_NAME;
    }

    @Override
    @NonNull
    public String getAppName() {
        return "Douyin";
    }

    @Override
    @NonNull
    public List<String> getSupportedOperations() {
        return DOUYIN_OPERATIONS;
    }

    @Override
    public float canHandle(@NonNull AutomationEngine engine) {
        // Check for Douyin-specific elements
        if (hasId(engine, "com.ss.android.ugc.aweme:id/")) {
            return 0.95f;
        }

        // Check for common Douyin texts
        String[] douyinTexts = {"推荐", "关注", "朋友", "我", "商城"};
        int matchCount = 0;
        for (String text : douyinTexts) {
            if (hasText(engine, text)) {
                matchCount++;
            }
        }

        return matchCount >= 2 ? 0.8f : 0.0f;
    }

    @Override
    @NonNull
    public String detectPage(@NonNull AutomationEngine engine) {
        // Detect current page type
        if (hasText(engine, "搜索")) {
            if (hasId(engine, ID_SEARCH_INPUT)) {
                return "search";
            }
        }

        if (hasText(engine, "商城") || hasId(engine, ID_TAB_MALL)) {
            return "mall";
        }

        if (hasText(engine, "购物车")) {
            return "cart";
        }

        if (hasText(engine, "确认订单") || hasText(engine, "提交订单")) {
            return "checkout";
        }

        if (hasText(engine, "商品详情") || hasText(engine, "立即购买")) {
            return "product_detail";
        }

        if (hasId(engine, ID_LIKE_BUTTON) || hasText(engine, "评论")) {
            return "video_feed";
        }

        if (hasText(engine, "作品") || hasText(engine, "喜欢")) {
            return "user_profile";
        }

        return "home";
    }

    @Override
    @NonNull
    public AutomationResult search(@NonNull AutomationEngine engine, @NonNull String query) {
        Log.d(TAG, "Searching for: " + query);

        // Click search button (usually at top-right)
        AutomationResult result = clickByText(engine, "搜索", false);
        if (!result.isSuccess()) {
            // Try clicking search icon
            result = clickById(engine, ID_SEARCH_BUTTON);
        }

        if (!result.isSuccess()) {
            return error("search", "Failed to open search");
        }

        waitForPage(engine, 500);

        // Input search text
        result = engine.inputText(BySelector.id(ID_SEARCH_INPUT), query);
        if (!result.isSuccess()) {
            // Fallback: try finding any input field
            AutomationNode inputNode = engine.findElement(
                BySelector.className("android.widget.EditText")
            );
            if (inputNode != null) {
                result = engine.inputText(query);
            }
        }

        if (!result.isSuccess()) {
            return error("search", "Failed to input search text");
        }

        // Press enter/search on keyboard
        engine.pressEnter();
        waitForPage(engine, 1000);

        return success("search");
    }

    @Override
    @NonNull
    public AutomationResult selectShop(@NonNull AutomationEngine engine, @NonNull String shopName) {
        Log.d(TAG, "Selecting shop: " + shopName);

        // In search results, look for shop
        return scrollAndClick(engine, shopName, 5);
    }

    @Override
    @NonNull
    public AutomationResult selectProduct(@NonNull AutomationEngine engine, @NonNull String productName) {
        Log.d(TAG, "Selecting product: " + productName);

        // Look for product in search results or shop
        return scrollAndClick(engine, productName, 5);
    }

    @Override
    @NonNull
    public AutomationResult configureOptions(@NonNull AutomationEngine engine,
                                             @NonNull Map<String, String> options) {
        Log.d(TAG, "Configuring options: " + options);

        // Click "选择规格" or similar
        AutomationResult result = clickByText(engine, "选择", false);
        if (!result.isSuccess()) {
            result = clickByText(engine, "规格", false);
        }

        if (!result.isSuccess()) {
            return error("configureOptions", "No options to configure");
        }

        waitForPage(engine, 500);

        // Select each option
        for (Map.Entry<String, String> entry : options.entrySet()) {
            result = scrollAndClick(engine, entry.getValue(), 3);
            if (!result.isSuccess()) {
                Log.w(TAG, "Failed to select option: " + entry.getKey() + " = " + entry.getValue());
            }
        }

        // Confirm selection
        clickByText(engine, "确定", true);
        clickByText(engine, "完成", true);

        return success("configureOptions");
    }

    @Override
    @NonNull
    public AutomationResult addToCart(@NonNull AutomationEngine engine, int quantity) {
        Log.d(TAG, "Adding to cart, quantity: " + quantity);

        // Click "加入购物车"
        AutomationResult result = clickByText(engine, "加入购物车", false);
        if (!result.isSuccess()) {
            result = clickByText(engine, "加购", false);
        }

        if (result.isSuccess()) {
            // Adjust quantity if needed
            if (quantity > 1) {
                // Find + button and click (quantity - 1) times
                for (int i = 1; i < quantity; i++) {
                    clickByText(engine, "+", true);
                }
            }

            // Confirm
            clickByText(engine, "确定", true);
        }

        return result;
    }

    @Override
    @NonNull
    public AutomationResult goToCart(@NonNull AutomationEngine engine) {
        Log.d(TAG, "Going to cart");

        // Go to mall tab first
        AutomationResult result = clickByText(engine, "商城", false);
        if (result.isSuccess()) {
            waitForPage(engine, 500);
            // Then click cart icon
            result = clickByText(engine, "购物车", false);
        }

        return result;
    }

    @Override
    @NonNull
    public AutomationResult goToCheckout(@NonNull AutomationEngine engine) {
        Log.d(TAG, "Going to checkout");

        return clickByText(engine, "去结算", false);
    }

    @Override
    @NonNull
    public AutomationResult selectAddress(@NonNull AutomationEngine engine,
                                          @NonNull String addressHint) {
        Log.d(TAG, "Selecting address: " + addressHint);

        // If address selection page, click on matching address
        return scrollAndClick(engine, addressHint, 3);
    }

    @Override
    @NonNull
    public AutomationResult submitOrder(@NonNull AutomationEngine engine) {
        Log.d(TAG, "Submitting order");

        return clickByText(engine, "提交订单", false);
    }

    @Override
    @NonNull
    public AutomationResult pay(@NonNull AutomationEngine engine,
                                @NonNull String paymentMethod) {
        Log.d(TAG, "Paying with: " + paymentMethod);

        // Select payment method
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
            AutomationResult result = clickByText(engine, method, false);
            if (result.isSuccess()) {
                break;
            }
        }

        // Click pay button
        return clickByText(engine, "立即支付", false);
    }

    // ===== Douyin-Specific Operations =====

    /**
     * Watch video for specified duration
     */
    @NonNull
    public AutomationResult watchVideo(@NonNull AutomationEngine engine, long durationMs) {
        Log.d(TAG, "Watching video for " + durationMs + "ms");

        try {
            Thread.sleep(durationMs);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }

        return success("watchVideo");
    }

    /**
     * Like current video
     */
    @NonNull
    public AutomationResult likeVideo(@NonNull AutomationEngine engine) {
        Log.d(TAG, "Liking video");

        // Double tap to like
        AutomationNode likeButton = engine.findElement(BySelector.id(ID_LIKE_BUTTON));
        if (likeButton != null) {
            return engine.doubleClick(likeButton.getCenterX(), likeButton.getCenterY());
        }

        // Fallback: click heart icon
        return clickByText(engine, "喜欢", false);
    }

    /**
     * Comment on current video
     */
    @NonNull
    public AutomationResult commentVideo(@NonNull AutomationEngine engine,
                                         @NonNull String comment) {
        Log.d(TAG, "Commenting: " + comment);

        // Click comment button
        AutomationResult result = clickById(engine, ID_COMMENT_BUTTON);
        if (!result.isSuccess()) {
            result = clickByText(engine, "评论", false);
        }

        if (!result.isSuccess()) {
            return error("commentVideo", "Failed to open comments");
        }

        waitForPage(engine, 500);

        // Input comment
        result = engine.inputText(comment);
        if (!result.isSuccess()) {
            return error("commentVideo", "Failed to input comment");
        }

        // Submit comment
        return clickByText(engine, "发送", false);
    }

    /**
     * Follow current user
     */
    @NonNull
    public AutomationResult followUser(@NonNull AutomationEngine engine) {
        Log.d(TAG, "Following user");

        return clickByText(engine, "关注", true);
    }

    /**
     * Share current video
     */
    @NonNull
    public AutomationResult shareVideo(@NonNull AutomationEngine engine) {
        Log.d(TAG, "Sharing video");

        return clickById(engine, ID_SHARE_BUTTON);
    }

    /**
     * Swipe to next video
     */
    @NonNull
    public AutomationResult browseFeed(@NonNull AutomationEngine engine) {
        Log.d(TAG, "Browsing to next video");

        return engine.swipe(com.ofa.agent.automation.Direction.UP, 0);
    }

    /**
     * Open Douyin Mall
     */
    @NonNull
    public AutomationResult openMall(@NonNull AutomationEngine engine) {
        Log.d(TAG, "Opening mall");

        return clickByText(engine, "商城", false);
    }

    /**
     * Search for a user
     */
    @NonNull
    public AutomationResult searchUser(@NonNull AutomationEngine engine,
                                       @NonNull String username) {
        Log.d(TAG, "Searching for user: " + username);

        // Open search
        AutomationResult result = search(engine, username);

        if (result.isSuccess()) {
            // Click on "用户" tab
            clickByText(engine, "用户", true);
        }

        return result;
    }

    @Override
    @NonNull
    public OrderStatus getOrderStatus(@NonNull AutomationEngine engine,
                                      @Nullable String orderId) {
        Log.d(TAG, "Getting order status");

        // Navigate to orders page
        AutomationResult result = clickByText(engine, "我的订单", false);
        if (!result.isSuccess()) {
            result = clickByText(engine, "订单", false);
        }

        if (!result.isSuccess()) {
            return new OrderStatus("unknown", "Could not find orders", 0, null);
        }

        waitForPage(engine, 1000);

        // Parse order status from page
        String[] statusTexts = {"待付款", "待发货", "待收货", "已完成", "已取消"};
        String[] statusCodes = {"pending_payment", "pending_shipment",
                                "pending_delivery", "completed", "cancelled"};

        for (int i = 0; i < statusTexts.length; i++) {
            if (hasText(engine, statusTexts[i])) {
                return new OrderStatus(statusCodes[i], statusTexts[i], 0, null);
            }
        }

        return new OrderStatus("unknown", "Status not detected", 0, null);
    }
}