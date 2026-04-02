package com.ofa.agent.automation.adapter;

import android.util.Log;

import androidx.annotation.NonNull;

import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.automation.AutomationNode;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.automation.BySelector;
import com.ofa.agent.automation.Direction;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Map;

/**
 * Base app adapter with common utilities.
 * Subclasses can extend this to implement app-specific behavior.
 */
public abstract class BaseAppAdapter implements AppAdapter {

    protected final String TAG = getClass().getSimpleName();

    // Common operation list
    protected static final List<String> COMMON_OPERATIONS = Arrays.asList(
        "search", "selectShop", "selectProduct", "configureOptions",
        "addToCart", "goToCart", "goToCheckout", "selectAddress",
        "submitOrder", "pay", "goBack", "goToHome"
    );

    // ===== Utility Methods =====

    /**
     * Wait for page to load
     */
    protected void waitForPage(@NonNull AutomationEngine engine, long timeoutMs) {
        try {
            Thread.sleep(Math.min(timeoutMs, 3000));
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }

    /**
     * Click element by text with fallback
     */
    @NonNull
    protected AutomationResult clickByText(@NonNull AutomationEngine engine,
                                            @NonNull String text,
                                            boolean exact) {
        BySelector selector = exact ?
            BySelector.text(text) :
            BySelector.textContains(text);

        AutomationNode node = engine.findElement(selector);
        if (node != null) {
            return engine.click(node.getCenterX(), node.getCenterY());
        }

        return new AutomationResult("click", "Element not found: " + text);
    }

    /**
     * Click element by ID
     */
    @NonNull
    protected AutomationResult clickById(@NonNull AutomationEngine engine,
                                          @NonNull String id) {
        BySelector selector = BySelector.id(id);
        AutomationNode node = engine.findElement(selector);
        if (node != null) {
            return engine.click(node.getCenterX(), node.getCenterY());
        }

        return new AutomationResult("click", "Element not found with id: " + id);
    }

    /**
     * Scroll and find element
     */
    @NonNull
    protected AutomationResult scrollAndClick(@NonNull AutomationEngine engine,
                                               @NonNull String text,
                                               int maxScrolls) {
        BySelector selector = BySelector.textContains(text);

        // First check if already visible
        AutomationNode node = engine.findElement(selector);
        if (node != null) {
            return engine.click(node.getCenterX(), node.getCenterY());
        }

        // Scroll to find
        AutomationResult result = engine.scrollFind(selector, maxScrolls);
        if (result.isSuccess() && result.getFoundNode() != null) {
            AutomationNode foundNode = result.getFoundNode();
            return engine.click(foundNode.getCenterX(), foundNode.getCenterY());
        }

        return new AutomationResult("scrollAndClick", "Element not found after scrolling: " + text);
    }

    /**
     * Input text into search box
     */
    @NonNull
    protected AutomationResult inputSearch(@NonNull AutomationEngine engine,
                                            @NonNull String text,
                                            @NonNull String searchInputId) {
        // Try to find search input by ID
        BySelector selector = BySelector.id(searchInputId);
        AutomationNode node = engine.findElement(selector);

        if (node != null) {
            return engine.inputText(selector, text);
        }

        // Fallback: try common search input patterns
        String[] searchIds = {
            searchInputId,
            "search_input",
            "et_search",
            "search_edit",
            "input_search"
        };

        for (String id : searchIds) {
            selector = BySelector.id(id);
            node = engine.findElement(selector);
            if (node != null) {
                return engine.inputText(selector, text);
            }
        }

        // Try by hint text
        String[] hints = {"搜索", "搜一搜", "输入搜索内容"};
        for (String hint : hints) {
            node = engine.findElement(BySelector.textContains(hint));
            if (node != null) {
                engine.click(node.getCenterX(), node.getCenterY());
                waitForPage(engine, 500);
                return engine.inputText(text);
            }
        }

        return new AutomationResult("inputSearch", "Search input not found");
    }

    /**
     * Check if text exists on page
     */
    protected boolean hasText(@NonNull AutomationEngine engine, @NonNull String text) {
        return engine.findElement(BySelector.textContains(text)) != null;
    }

    /**
     * Check if ID exists on page
     */
    protected boolean hasId(@NonNull AutomationEngine engine, @NonNull String id) {
        return engine.findElement(BySelector.id(id)) != null;
    }

    /**
     * Press back button
     */
    @NonNull
    protected AutomationResult pressBack(@NonNull AutomationEngine engine) {
        // Use coordinate-based back press (top-left corner typically)
        // Or rely on system back if available
        return engine.swipe(Direction.LEFT, 0);
    }

    /**
     * Swipe down to refresh
     */
    @NonNull
    protected AutomationResult swipeRefresh(@NonNull AutomationEngine engine) {
        return engine.swipe(Direction.DOWN, 0);
    }

    /**
     * Get all visible texts on page
     */
    @NonNull
    protected List<String> getVisibleTexts(@NonNull AutomationEngine engine) {
        List<String> texts = new ArrayList<>();
        List<AutomationNode> nodes = engine.findElements(BySelector.clickable());

        for (AutomationNode node : nodes) {
            if (node.getText() != null && !node.getText().isEmpty()) {
                texts.add(node.getText());
            }
        }

        return texts;
    }

    // ===== Default Implementations =====

    @Override
    @NonNull
    public List<String> getSupportedOperations() {
        return COMMON_OPERATIONS;
    }

    @Override
    @NonNull
    public AutomationResult goBack(@NonNull AutomationEngine engine) {
        // Try to find back button first
        BySelector backSelector = BySelector.textContains("返回")
            .orClassName("android.widget.ImageButton");

        AutomationNode backButton = engine.findElement(backSelector);
        if (backButton != null) {
            return engine.click(backButton.getCenterX(), backButton.getCenterY());
        }

        // Use swipe gesture
        return engine.swipe(Direction.LEFT, 0);
    }

    @Override
    @NonNull
    public AutomationResult goToHome(@NonNull AutomationEngine engine) {
        // Try to find home button
        String[] homeTexts = {"首页", "首页推荐", "Home"};

        for (String text : homeTexts) {
            if (hasText(engine, text)) {
                return clickByText(engine, text, false);
            }
        }

        // Fallback: look for bottom nav home icon
        String[] homeIds = {"tab_home", "home_tab", "nav_home", "bottom_home"};
        for (String id : homeIds) {
            if (hasId(engine, id)) {
                return clickById(engine, id);
            }
        }

        return new AutomationResult("goToHome", "Home button not found");
    }

    @Override
    @NonNull
    public OrderStatus getOrderStatus(@NonNull AutomationEngine engine, String orderId) {
        // Default implementation - subclasses should override
        return new OrderStatus("unknown", "Status detection not implemented", 0, null);
    }

    /**
     * Helper to create success result
     */
    @NonNull
    protected AutomationResult success(@NonNull String operation) {
        return new AutomationResult(operation, 0);
    }

    /**
     * Helper to create error result
     */
    @NonNull
    protected AutomationResult error(@NonNull String operation, @NonNull String message) {
        return new AutomationResult(operation, message);
    }
}