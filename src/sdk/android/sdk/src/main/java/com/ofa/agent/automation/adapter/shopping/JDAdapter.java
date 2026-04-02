package com.ofa.agent.automation.adapter.shopping;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.automation.AutomationNode;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.automation.BySelector;
import com.ofa.agent.automation.adapter.BaseAppAdapter;

import java.util.Map;

/**
 * JD (京东) App Adapter.
 * Package: com.jingdong.app.mall
 */
public class JDAdapter extends BaseAppAdapter {

    private static final String TAG = "JDAdapter";

    // Package names
    private static final String PACKAGE_MAIN = "com.jingdong.app.mall";

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
        return "京东";
    }

    @Override
    public float canHandle(@NonNull AutomationEngine engine) {
        String pkg = engine.getForegroundPackage();
        if (pkg == null) return 0f;

        if (pkg.equals(PACKAGE_MAIN) || pkg.startsWith("com.jingdong")) {
            // Check for JD-specific UI elements
            if (hasText(engine, "京东") || hasText(engine, "JD") ||
                hasText(engine, "京东物流") || hasText(engine, "自营")) {
                return 0.9f;
            }
            return 0.7f;
        }

        return 0f;
    }

    @Override
    @NonNull
    public String detectPage(@NonNull AutomationEngine engine) {
        // Order page
        if (hasText(engine, "订单详情") || hasText(engine, "物流详情") ||
            hasText(engine, "已签收") || hasText(engine, "配送中")) {
            return PAGE_ORDER;
        }

        // Checkout page
        if (hasText(engine, "提交订单") || hasText(engine, "去付款") ||
            hasText(engine, "京东支付")) {
            return PAGE_CHECKOUT;
        }

        // Cart page
        if (hasText(engine, "购物车") && (hasText(engine, "去结算") ||
            hasText(engine, "全选"))) {
            return PAGE_CART;
        }

        // Product detail page
        if (hasText(engine, "加入购物车") || hasText(engine, "立即购买") ||
            hasText(engine, "商品详情") || hasText(engine, "规格")) {
            return PAGE_PRODUCT;
        }

        // Shop page
        if (hasText(engine, "店铺") || hasText(engine, "全部商品") ||
            hasText(engine, "进店")) {
            return PAGE_SHOP;
        }

        // Search page
        if (hasText(engine, "搜索") && (hasText(engine, "商品") ||
            hasText(engine, "店铺") || hasText(engine, "历史"))) {
            return PAGE_SEARCH;
        }

        // Home page
        if (hasText(engine, "首页") || hasText(engine, "推荐") ||
            hasText(engine, "分类") || hasText(engine, "我的")) {
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
            BySelector searchSelector = BySelector.textContains("搜索")
                .orId("search_btn").orId("home_search_view");

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

        // Click search
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
            waitForPage(engine, 1500);
        }

        return result;
    }

    @Override
    @NonNull
    public AutomationResult configureOptions(@NonNull AutomationEngine engine,
                                              @NonNull Map<String, String> options) {
        Log.i(TAG, "Configuring options: " + options);

        // JD common option labels
        String[][] optionMappings = {
            {"color", "颜色"},
            {"size", "尺码"},
            {"style", "款式"},
            {"version", "版本"},
            {"spec", "规格"},
            {"service", "服务"}
        };

        for (Map.Entry<String, String> entry : options.entrySet()) {
            String optionKey = entry.getKey();
            String optionValue = entry.getValue();

            String label = optionKey;
            for (String[] mapping : optionMappings) {
                if (mapping[0].equals(optionKey)) {
                    label = mapping[1];
                    break;
                }
            }

            BySelector optionSelector = BySelector.textContains(optionValue)
                .orDescContains(optionValue);

            AutomationNode optionNode = engine.findElement(optionSelector);
            if (optionNode != null) {
                engine.click(optionNode.getCenterX(), optionNode.getCenterY());
                waitForPage(engine, 300);
            }
        }

        return success("configureOptions");
    }

    @Override
    @NonNull
    public AutomationResult addToCart(@NonNull AutomationEngine engine, int quantity) {
        Log.i(TAG, "Adding to cart, quantity: " + quantity);

        // Adjust quantity
        if (quantity > 1) {
            BySelector plusSelector = BySelector.textContains("+")
                .orId("btn_plus").orId("iv_add").orId("add_count");

            for (int i = 1; i < quantity; i++) {
                AutomationNode plusNode = engine.findElement(plusSelector);
                if (plusNode != null) {
                    engine.click(plusNode.getCenterX(), plusNode.getCenterY());
                    waitForPage(engine, 200);
                }
            }
        }

        // Click add to cart button
        String[] addTexts = {"加入购物车", "加入购物车", "添加购物车"};

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

        String[] cartTexts = {"购物车", "我的购物车"};
        for (String text : cartTexts) {
            if (hasText(engine, text)) {
                return clickByText(engine, text, false);
            }
        }

        // Try cart icon
        String[] cartIds = {"cart_icon", "shopping_cart_icon", "tab_cart"};
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

        String[] checkoutTexts = {"去结算", "结算", "立即购买"};
        for (String text : checkoutTexts) {
            if (hasText(engine, text)) {
                AutomationResult result = clickByText(engine, text, false);
                if (result.isSuccess()) {
                    waitForPage(engine, 2000);
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

        if (!hasText(engine, "选择收货地址") && !hasText(engine, "收货地址")) {
            return success("selectAddress");
        }

        BySelector addressSelector = BySelector.textContains("地址")
            .orTextContains("收货").orId("address_view");

        AutomationNode addressNode = engine.findElement(addressSelector);
        if (addressNode != null) {
            engine.click(addressNode.getCenterX(), addressNode.getCenterY());
            waitForPage(engine, 1000);

            AutomationResult result = scrollAndClick(engine, addressHint, 3);
            return result;
        }

        return success("selectAddress");
    }

    @Override
    @NonNull
    public AutomationResult submitOrder(@NonNull AutomationEngine engine) {
        Log.i(TAG, "Submitting order");

        String[] submitTexts = {"提交订单", "确认订单", "去付款"};
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

        waitForPage(engine, 2000);

        String[] methodTexts;
        switch (paymentMethod.toLowerCase()) {
            case "jd":
            case "京东支付":
                methodTexts = new String[]{"京东支付", "JD支付"};
                break;
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

        String[] payTexts = {"确认支付", "立即支付", "付款"};
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
        String[] deliveredTexts = {"已签收", "已收货", "已完成", "订单完成"};
        String[] deliveringTexts = {"配送中", "正在配送", "派送中"};
        String[] shippedTexts = {"已发货", "仓库已发货"};
        String[] pendingTexts = {"待发货", "等待发货"};

        for (String text : deliveredTexts) {
            if (hasText(engine, text)) {
                return new OrderStatus("delivered", text, 0, null);
            }
        }

        for (String text : deliveringTexts) {
            if (hasText(engine, text)) {
                return new OrderStatus("delivering", text, 0, null);
            }
        }

        for (String text : shippedTexts) {
            if (hasText(engine, text)) {
                return new OrderStatus("shipped", text, 0, null);
            }
        }

        for (String text : pendingTexts) {
            if (hasText(engine, text)) {
                return new OrderStatus("pending", text, 0, null);
            }
        }

        return new OrderStatus("unknown", "Unknown status", 0, null);
    }
}