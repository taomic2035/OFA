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
 * Xiaohongshu (小红书) App Adapter
 * Supports content browsing, searching, liking, collecting, and shopping
 */
public class XiaohongshuAdapter extends BaseAppAdapter {

    // Package name
    private static final String PACKAGE_NAME = "com.xingin.xhs";

    // UI Element IDs
    private static final String ID_SEARCH_INPUT = "com.xingin.xhs:id/search_edit_text";
    private static final String ID_TAB_HOME = "com.xingin.xhs:id/tab_home";
    private static final String ID_TAB_SHOPPING = "com.xingin.xhs:id/tab_shopping";
    private static final String ID_TAB_ME = "com.xingin.xhs:id/tab_me";
    private static final String ID_LIKE_BUTTON = "com.xingin.xhs:id/like_btn";
    private static final String ID_COLLECT_BUTTON = "com.xingin.xhs:id/collect_btn";
    private static final String ID_COMMENT_BUTTON = "com.xingin.xhs:id/comment_btn";

    // Xiaohongshu-specific operations
    private static final List<String> XHS_OPERATIONS = Arrays.asList(
        "search", "selectShop", "selectProduct", "configureOptions",
        "addToCart", "goToCart", "goToCheckout", "selectAddress",
        "submitOrder", "pay", "goBack", "goToHome",
        // Xiaohongshu-specific
        "browseNote", "likeNote", "collectNote", "commentNote",
        "followUser", "shareNote", "searchNote", "viewProfile",
        "publishNote", "openShoppingTab"
    );

    @Override
    @NonNull
    public String getPackageName() {
        return PACKAGE_NAME;
    }

    @Override
    @NonNull
    public String getAppName() {
        return "Xiaohongshu";
    }

    @Override
    @NonNull
    public List<String> getSupportedOperations() {
        return XHS_OPERATIONS;
    }

    @Override
    public float canHandle(@NonNull AutomationEngine engine) {
        // Check for Xiaohongshu-specific elements
        if (hasId(engine, "com.xingin.xhs:id/")) {
            return 0.95f;
        }

        // Check for common Xiaohongshu texts
        String[] xhsTexts = {"首页", "购物", "消息", "我", "发现", "关注"};
        int matchCount = 0;
        for (String text : xhsTexts) {
            if (hasText(engine, text)) {
                matchCount++;
            }
        }

        return matchCount >= 2 ? 0.8f : 0.0f;
    }

    @Override
    @NonNull
    public String detectPage(@NonNull AutomationEngine engine) {
        if (hasText(engine, "搜索") && hasId(engine, ID_SEARCH_INPUT)) {
            return "search";
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

        if (hasText(engine, "笔记详情") || hasId(engine, ID_LIKE_BUTTON)) {
            return "note_detail";
        }

        if (hasText(engine, "粉丝") || hasText(engine, "关注")) {
            if (!hasText(engine, "首页")) {
                return "user_profile";
            }
        }

        if (hasText(engine, "发布笔记")) {
            return "publish";
        }

        return "home";
    }

    @Override
    @NonNull
    public AutomationResult search(@NonNull AutomationEngine engine, @NonNull String query) {
        Log.d(TAG, "Searching for: " + query);

        // Click search bar at top
        AutomationResult result = clickByText(engine, "搜索", false);
        if (!result.isSuccess()) {
            result = clickById(engine, ID_SEARCH_INPUT);
        }

        if (!result.isSuccess()) {
            return error("search", "Failed to open search");
        }

        waitForPage(engine, 500);

        // Input search text
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
            return error("search", "Failed to input search text");
        }

        engine.pressEnter();
        waitForPage(engine, 1000);

        return success("search");
    }

    @Override
    @NonNull
    public AutomationResult selectShop(@NonNull AutomationEngine engine,
                                       @NonNull String shopName) {
        Log.d(TAG, "Selecting shop: " + shopName);

        return scrollAndClick(engine, shopName, 5);
    }

    @Override
    @NonNull
    public AutomationResult selectProduct(@NonNull AutomationEngine engine,
                                          @NonNull String productName) {
        Log.d(TAG, "Selecting product: " + productName);

        return scrollAndClick(engine, productName, 5);
    }

    @Override
    @NonNull
    public AutomationResult configureOptions(@NonNull AutomationEngine engine,
                                             @NonNull Map<String, String> options) {
        Log.d(TAG, "Configuring options: " + options);

        // Click "选择" button
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
            scrollAndClick(engine, entry.getValue(), 3);
        }

        // Confirm
        clickByText(engine, "确定", true);
        clickByText(engine, "完成", true);

        return success("configureOptions");
    }

    @Override
    @NonNull
    public AutomationResult addToCart(@NonNull AutomationEngine engine, int quantity) {
        Log.d(TAG, "Adding to cart, quantity: " + quantity);

        AutomationResult result = clickByText(engine, "加入购物车", false);
        if (!result.isSuccess()) {
            result = clickByText(engine, "加购", false);
        }

        return result;
    }

    @Override
    @NonNull
    public AutomationResult goToCart(@NonNull AutomationEngine engine) {
        Log.d(TAG, "Going to cart");

        // Go to shopping tab first
        AutomationResult result = clickByText(engine, "购物", false);
        if (result.isSuccess()) {
            waitForPage(engine, 500);
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
            clickByText(engine, method, false);
        }

        return clickByText(engine, "立即支付", false);
    }

    // ===== Xiaohongshu-Specific Operations =====

    /**
     * Browse a note (scroll through images)
     */
    @NonNull
    public AutomationResult browseNote(@NonNull AutomationEngine engine, int swipeCount) {
        Log.d(TAG, "Browsing note with " + swipeCount + " swipes");

        for (int i = 0; i < swipeCount; i++) {
            engine.swipe(com.ofa.agent.automation.Direction.LEFT, 0);
            try {
                Thread.sleep(500);
            } catch (InterruptedException e) {
                break;
            }
        }

        return success("browseNote");
    }

    /**
     * Like current note
     */
    @NonNull
    public AutomationResult likeNote(@NonNull AutomationEngine engine) {
        Log.d(TAG, "Liking note");

        return clickById(engine, ID_LIKE_BUTTON);
    }

    /**
     * Collect (save) current note
     */
    @NonNull
    public AutomationResult collectNote(@NonNull AutomationEngine engine,
                                        @Nullable String collectionName) {
        Log.d(TAG, "Collecting note");

        AutomationResult result = clickById(engine, ID_COLLECT_BUTTON);
        if (!result.isSuccess()) {
            result = clickByText(engine, "收藏", false);
        }

        if (result.isSuccess() && collectionName != null) {
            // Select collection if specified
            waitForPage(engine, 300);
            scrollAndClick(engine, collectionName, 3);
        }

        return result;
    }

    /**
     * Comment on current note
     */
    @NonNull
    public AutomationResult commentNote(@NonNull AutomationEngine engine,
                                        @NonNull String comment) {
        Log.d(TAG, "Commenting: " + comment);

        AutomationResult result = clickById(engine, ID_COMMENT_BUTTON);
        if (!result.isSuccess()) {
            result = clickByText(engine, "评论", false);
        }

        if (!result.isSuccess()) {
            return error("commentNote", "Failed to open comments");
        }

        waitForPage(engine, 500);

        result = engine.inputText(comment);
        if (!result.isSuccess()) {
            return error("commentNote", "Failed to input comment");
        }

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
     * Share current note
     */
    @NonNull
    public AutomationResult shareNote(@NonNull AutomationEngine engine) {
        Log.d(TAG, "Sharing note");

        // Click share button
        AutomationResult result = clickByText(engine, "分享", false);
        if (!result.isSuccess()) {
            // Try share icon
            result = clickByText(engine, "...", false);
            if (result.isSuccess()) {
                waitForPage(engine, 300);
                result = clickByText(engine, "分享", false);
            }
        }

        return result;
    }

    /**
     * Search specifically for notes
     */
    @NonNull
    public AutomationResult searchNote(@NonNull AutomationEngine engine,
                                       @NonNull String query) {
        Log.d(TAG, "Searching for notes: " + query);

        AutomationResult result = search(engine, query);

        if (result.isSuccess()) {
            // Click on "笔记" tab
            clickByText(engine, "笔记", true);
        }

        return result;
    }

    /**
     * View user profile
     */
    @NonNull
    public AutomationResult viewProfile(@NonNull AutomationEngine engine) {
        Log.d(TAG, "Viewing user profile");

        // Click on user avatar/name
        return clickByText(engine, "头像", false);
    }

    /**
     * Open publish note page
     */
    @NonNull
    public AutomationResult publishNote(@NonNull AutomationEngine engine) {
        Log.d(TAG, "Opening publish page");

        // Click the + button at bottom
        AutomationResult result = clickByText(engine, "+", true);
        if (!result.isSuccess()) {
            // Try clicking "发布" text
            result = clickByText(engine, "发布", false);
        }

        return result;
    }

    /**
     * Open shopping tab
     */
    @NonNull
    public AutomationResult openShoppingTab(@NonNull AutomationEngine engine) {
        Log.d(TAG, "Opening shopping tab");

        return clickByText(engine, "购物", false);
    }

    @Override
    @NonNull
    public OrderStatus getOrderStatus(@NonNull AutomationEngine engine,
                                      @Nullable String orderId) {
        Log.d(TAG, "Getting order status");

        // Navigate to orders
        AutomationResult result = clickByText(engine, "我", false);
        if (result.isSuccess()) {
            waitForPage(engine, 500);
            result = clickByText(engine, "订单", false);
        }

        if (!result.isSuccess()) {
            return new OrderStatus("unknown", "Could not find orders", 0, null);
        }

        waitForPage(engine, 1000);

        // Parse status
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