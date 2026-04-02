package com.ofa.agent.automation.adapter.food;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.automation.AutomationNode;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.automation.BySelector;
import com.ofa.agent.automation.adapter.BaseAppAdapter;

import org.json.JSONObject;

import java.util.Arrays;
import java.util.List;
import java.util.Map;

/**
 * Meituan Waimai (美团外卖) App Adapter.
 * Package: com.sankuai.meituan
 */
public class MeituanAdapter extends BaseAppAdapter {

    // Package names
    private static final String PACKAGE_MAIN = "com.sankuai.meituan";
    private static final String PACKAGE_WAIMAI = "com.sankuai.meituan.takeoutnew";

    // Known resource IDs (these may change with app updates)
    private static final String ID_SEARCH_INPUT = "com.sankuai.meituan:id/search_input";
    private static final String ID_SEARCH_BUTTON = "com.sankuai.meituan:id/search_button";
    private static final String ID_CART_BUTTON = "com.sankuai.meituan:id/cart_button";
    private static final String ID_SUBMIT_ORDER = "com.sankuai.meituan:id/submit_order";

    // Page identifiers
    private static final String PAGE_HOME = "home";
    private static final String PAGE_SEARCH = "search";
    private static final String PAGE_SHOP = "shop";
    private static final String PAGE_PRODUCT = "product";
    private static final String PAGE_CART = "cart";
    private static final String PAGE_CHECKOUT = "checkout";
    private static final String PAGE_ORDER = "order";

    @Override
    @NonNull
    public String getPackageName() {
        return PACKAGE_MAIN;
    }

    @Override
    @NonNull
    public String getAppName() {
        return "美团外卖";
    }

    @Override
    public float canHandle(@NonNull AutomationEngine engine) {
        String pkg = engine.getForegroundPackage();
        if (pkg == null) return 0f;

        if (pkg.equals(PACKAGE_MAIN) || pkg.equals(PACKAGE_WAIMAI) ||
            pkg.startsWith("com.sankuai.meituan")) {
            // Check for Meituan-specific UI elements
            if (hasText(engine, "美团") || hasText(engine, "外卖") ||
                hasText(engine, "美食") || hasText(engine, "配送")) {
                return 0.9f;
            }
            return 0.7f;
        }

        return 0f;
    }

    @Override
    @NonNull
    public String detectPage(@NonNull AutomationEngine engine) {
        // Check for specific page indicators

        // Order page
        if (hasText(engine, "订单详情") || hasText(engine, "配送中") ||
            hasText(engine, "已送达")) {
            return PAGE_ORDER;
        }

        // Checkout page
        if (hasText(engine, "提交订单") || hasText(engine, "去支付")) {
            return PAGE_CHECKOUT;
        }

        // Cart page
        if (hasText(engine, "购物车") && hasText(engine, "去结算")) {
            return PAGE_CART;
        }

        // Product detail page
        if (hasText(engine, "加入购物车") || hasText(engine, "选规格") ||
            hasText(engine, "选口味")) {
            return PAGE_PRODUCT;
        }

        // Shop page
        if (hasText(engine, "搜索店内") || hasText(engine, "推荐") ||
            hasText(engine, "评价")) {
            return PAGE_SHOP;
        }

        // Search page
        if (hasText(engine, "搜索") && hasText(engine, "历史")) {
            return PAGE_SEARCH;
        }

        // Home page
        if (hasText(engine, "外卖") || hasText(engine, "美食") ||
            hasText(engine, "推荐")) {
            return PAGE_HOME;
        }

        return "unknown";
    }

    @Override
    @NonNull
    public AutomationResult search(@NonNull AutomationEngine engine, @NonNull String query) {
        Log.i(TAG, "Searching for: " + query);

        // Navigate to search
        if (!detectPage(engine).equals(PAGE_SEARCH)) {
            // Click search button
            BySelector searchSelector = BySelector.textContains("搜索")
                .orId("search_button").orId("et_search");

            AutomationNode searchNode = engine.findElement(searchSelector);
            if (searchNode != null) {
                engine.click(searchNode.getCenterX(), searchNode.getCenterY());
                waitForPage(engine, 1000);
            } else {
                return error("search", "Search button not found");
            }
        }

        // Input search text
        AutomationResult inputResult = inputSearch(engine, query, "search_input");
        if (!inputResult.isSuccess()) {
            return inputResult;
        }

        waitForPage(engine, 1500);

        // Click search on keyboard or button
        BySelector searchButtonSelector = BySelector.text("搜索")
            .orId("search_go").orId("tv_search");

        AutomationNode searchBtn = engine.findElement(searchButtonSelector);
        if (searchBtn != null) {
            return engine.click(searchBtn.getCenterX(), searchBtn.getCenterY());
        }

        return success("search");
    }

    @Override
    @NonNull
    public AutomationResult selectShop(@NonNull AutomationEngine engine, @NonNull String shopName) {
        Log.i(TAG, "Selecting shop: " + shopName);

        BySelector shopSelector = BySelector.textContains(shopName);
        AutomationResult result = scrollAndClick(engine, shopName, 5);

        if (result.isSuccess()) {
            waitForPage(engine, 2000);
        }

        return result;
    }

    @Override
    @NonNull
    public AutomationResult selectProduct(@NonNull AutomationEngine engine, @NonNull String productName) {
        Log.i(TAG, "Selecting product: " + productName);

        AutomationResult result = scrollAndClick(engine, productName, 5);
        if (result.isSuccess()) {
            waitForPage(engine, 1000);
        }

        return result;
    }

    @Override
    @NonNull
    public AutomationResult configureOptions(@NonNull AutomationEngine engine,
                                              @NonNull Map<String, String> options) {
        Log.i(TAG, "Configuring options: " + options);

        // Common option labels in Meituan
        String[][] optionMappings = {
            {"sweetness", "甜度"},
            {"temperature", "温度"},
            {"size", "杯型"},
            {"toppings", "小料"},
            {"spiciness", "辣度"}
        };

        for (Map.Entry<String, String> entry : options.entrySet()) {
            String optionKey = entry.getKey();
            String optionValue = entry.getValue();

            // Map to Chinese labels
            String label = optionKey;
            for (String[] mapping : optionMappings) {
                if (mapping[0].equals(optionKey)) {
                    label = mapping[1];
                    break;
                }
            }

            // Find and click the option
            BySelector optionSelector = BySelector.textContains(optionValue);
            AutomationNode optionNode = engine.findElement(optionSelector);

            if (optionNode != null) {
                engine.click(optionNode.getCenterX(), optionNode.getCenterY());
                waitForPage(engine, 300);
            } else {
                Log.w(TAG, "Option not found: " + optionKey + " = " + optionValue);
            }
        }

        return success("configureOptions");
    }

    @Override
    @NonNull
    public AutomationResult addToCart(@NonNull AutomationEngine engine, int quantity) {
        Log.i(TAG, "Adding to cart, quantity: " + quantity);

        // Adjust quantity if needed
        if (quantity > 1) {
            // Find quantity controls
            BySelector plusSelector = BySelector.textContains("+")
                .orId("btn_plus").orId("iv_add");

            for (int i = 1; i < quantity; i++) {
                AutomationNode plusNode = engine.findElement(plusSelector);
                if (plusNode != null) {
                    engine.click(plusNode.getCenterX(), plusNode.getCenterY());
                    waitForPage(engine, 200);
                }
            }
        }

        // Click add to cart button
        String[] addTexts = {"加入购物车", "加入购物车", "选好了", "确认"};

        for (String text : addTexts) {
            if (hasText(engine, text)) {
                return clickByText(engine, text, false);
            }
        }

        return error("addToCart", "Add to cart button not found");
    }

    @Override
    @NonNull
    public AutomationResult goToCart(@NonNull AutomationEngine engine) {
        Log.i(TAG, "Going to cart");

        String[] cartTexts = {"购物车", "去购物车"};
        for (String text : cartTexts) {
            if (hasText(engine, text)) {
                return clickByText(engine, text, false);
            }
        }

        // Try cart icon
        String[] cartIds = {"cart_icon", "iv_cart", "cart_button"};
        for (String id : cartIds) {
            if (hasId(engine, id)) {
                return clickById(engine, id);
            }
        }

        return error("goToCart", "Cart button not found");
    }

    @Override
    @NonNull
    public AutomationResult goToCheckout(@NonNull AutomationEngine engine) {
        Log.i(TAG, "Going to checkout");

        String[] checkoutTexts = {"去结算", "去支付", "立即购买"};
        for (String text : checkoutTexts) {
            if (hasText(engine, text)) {
                AutomationResult result = clickByText(engine, text, false);
                if (result.isSuccess()) {
                    waitForPage(engine, 1500);
                }
                return result;
            }
        }

        return error("goToCheckout", "Checkout button not found");
    }

    @Override
    @NonNull
    public AutomationResult selectAddress(@NonNull AutomationEngine engine,
                                           @NonNull String addressHint) {
        Log.i(TAG, "Selecting address: " + addressHint);

        // Check if address selection is needed
        if (!hasText(engine, "选择地址") && !hasText(engine, "收货地址")) {
            return success("selectAddress"); // Address already selected
        }

        // Click on address area
        BySelector addressSelector = BySelector.textContains("地址")
            .orTextContains("收货");

        AutomationNode addressNode = engine.findElement(addressSelector);
        if (addressNode != null) {
            engine.click(addressNode.getCenterX(), addressNode.getCenterY());
            waitForPage(engine, 1000);

            // Select from address list
            BySelector hintSelector = BySelector.textContains(addressHint);
            AutomationResult result = scrollAndClick(engine, addressHint, 3);
            return result;
        }

        return success("selectAddress");
    }

    @Override
    @NonNull
    public AutomationResult submitOrder(@NonNull AutomationEngine engine) {
        Log.i(TAG, "Submitting order");

        String[] submitTexts = {"提交订单", "确认订单", "立即下单"};
        for (String text : submitTexts) {
            if (hasText(engine, text)) {
                return clickByText(engine, text, false);
            }
        }

        return error("submitOrder", "Submit button not found");
    }

    @Override
    @NonNull
    public AutomationResult pay(@NonNull AutomationEngine engine, @NonNull String paymentMethod) {
        Log.i(TAG, "Paying with: " + paymentMethod);

        // Wait for payment page
        waitForPage(engine, 2000);

        // Select payment method
        String[] methodTexts;
        switch (paymentMethod.toLowerCase()) {
            case "alipay":
            case "支付宝":
                methodTexts = new String[]{"支付宝", "支付宝支付"};
                break;
            case "wechat":
            case "微信":
                methodTexts = new String[]{"微信支付", "微信"};
                break;
            case "card":
            case "银行卡":
                methodTexts = new String[]{"银行卡", "银行卡支付"};
                break;
            default:
                methodTexts = new String[]{paymentMethod};
        }

        for (String method : methodTexts) {
            if (hasText(engine, method)) {
                clickByText(engine, method, false);
                break;
            }
        }

        waitForPage(engine, 500);

        // Click pay button
        String[] payTexts = {"确认支付", "立即支付", "支付"};
        for (String text : payTexts) {
            if (hasText(engine, text)) {
                return clickByText(engine, text, false);
            }
        }

        return error("pay", "Pay button not found");
    }

    @Override
    @NonNull
    public OrderStatus getOrderStatus(@NonNull AutomationEngine engine, @Nullable String orderId) {
        // Look for order status elements
        String[] deliveredTexts = {"已送达", "已完成", "订单完成"};
        String[] deliveringTexts = {"配送中", "骑手正在送达"};
        String[] preparingTexts = {"商家已接单", "正在备餐", "制作中"};
        String[] pendingTexts = {"待商家接单", "等待商家接单"};

        for (String text : deliveredTexts) {
            if (hasText(engine, text)) {
                return new OrderStatus("delivered", text, 0, null);
            }
        }

        for (String text : deliveringTexts) {
            if (hasText(engine, text)) {
                // Try to extract ETA
                long eta = extractETA(engine);
                return new OrderStatus("delivering", text, eta, null);
            }
        }

        for (String text : preparingTexts) {
            if (hasText(engine, text)) {
                return new OrderStatus("preparing", text, 0, null);
            }
        }

        for (String text : pendingTexts) {
            if (hasText(engine, text)) {
                return new OrderStatus("pending", text, 0, null);
            }
        }

        return new OrderStatus("unknown", "Unknown status", 0, null);
    }

    /**
     * Extract estimated time of arrival from page
     */
    private long extractETA(@NonNull AutomationEngine engine) {
        // Try to find ETA text like "预计15分钟送达"
        String source = engine.getPageSource();
        // This is a simplified extraction - real implementation would use regex
        return 0;
    }
}