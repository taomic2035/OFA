package com.ofa.agent.automation.webview;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationResult;

import org.json.JSONArray;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * WebPageAdapter - adapts automation for specific web pages.
 * Provides page-specific operations and smart element detection.
 */
public class WebPageAdapter {

    private static final String TAG = "WebPageAdapter";

    private final WebViewAutomation webViewAutomation;
    private final JsExecutor jsExecutor;
    private final WebFormFiller formFiller;

    // Page patterns
    private final Map<String, PagePattern> pagePatterns = new HashMap<>();

    public WebPageAdapter(@NonNull WebViewAutomation webViewAutomation) {
        this.webViewAutomation = webViewAutomation;
        this.jsExecutor = webViewAutomation.getJsExecutor();
        this.formFiller = new WebFormFiller(webViewAutomation);

        // Register built-in page patterns
        registerBuiltInPatterns();
    }

    /**
     * Page pattern definition
     */
    public static class PagePattern {
        public final String name;
        public final String urlPattern; // Regex pattern
        public final Map<String, String> selectors; // Named selectors

        public PagePattern(String name, String urlPattern, Map<String, String> selectors) {
            this.name = name;
            this.urlPattern = urlPattern;
            this.selectors = selectors;
        }
    }

    /**
     * Register built-in page patterns for common sites
     */
    private void registerBuiltInPatterns() {
        // Login pages
        Map<String, String> loginSelectors = new HashMap<>();
        loginSelectors.put("username", "input[type='text'], input[type='email'], input[name*='user'], input[name*='email']");
        loginSelectors.put("password", "input[type='password']");
        loginSelectors.put("submit", "button[type='submit'], input[type='submit'], button:contains('登录'), button:contains('Login')");
        pagePatterns.put("login", new PagePattern("login", ".*login.*|.*signin.*|.*auth.*", loginSelectors));

        // Search pages
        Map<String, String> searchSelectors = new HashMap<>();
        searchSelectors.put("searchInput", "input[type='search'], input[name='q'], input[name='query'], input[placeholder*='搜索']");
        searchSelectors.put("searchButton", "button[type='submit'], button:contains('搜索'), button:contains('Search')");
        searchSelectors.put("results", ".search-results, .results, #results, [class*='result']");
        pagePatterns.put("search", new PagePattern("search", ".*search.*|.*query.*", searchSelectors));

        // Form pages
        Map<String, String> formSelectors = new HashMap<>();
        formSelectors.put("form", "form");
        formSelectors.put("submit", "button[type='submit'], input[type='submit']");
        pagePatterns.put("form", new PagePattern("form", ".*", formSelectors));

        // E-commerce checkout
        Map<String, String> checkoutSelectors = new HashMap<>();
        checkoutSelectors.put("address", "input[name*='address'], textarea[name*='address']");
        checkoutSelectors.put("phone", "input[type='tel'], input[name*='phone']");
        checkoutSelectors.put("submit", "button:contains('提交'), button:contains('支付'), button:contains('结算')");
        pagePatterns.put("checkout", new PagePattern("checkout", ".*checkout.*|.*cart.*|.*order.*", checkoutSelectors));

        Log.d(TAG, "Registered " + pagePatterns.size() + " built-in page patterns");
    }

    /**
     * Register custom page pattern
     */
    public void registerPattern(@NonNull String name, @NonNull String urlPattern,
                                 @NonNull Map<String, String> selectors) {
        pagePatterns.put(name, new PagePattern(name, urlPattern, selectors));
        Log.d(TAG, "Registered page pattern: " + name);
    }

    /**
     * Detect current page type
     */
    @Nullable
    public String detectPageType() {
        String url = webViewAutomation.getUrl();

        for (Map.Entry<String, PagePattern> entry : pagePatterns.entrySet()) {
            if (url.matches(entry.getValue().urlPattern)) {
                return entry.getKey();
            }
        }

        // Try to detect by content
        return detectByContent();
    }

    /**
     * Detect page type by content analysis
     */
    @Nullable
    private String detectByContent() {
        // Check for login form
        if (hasPasswordInput() && hasTextInput()) {
            return "login";
        }

        // Check for search page
        if (hasSearchInput()) {
            return "search";
        }

        // Check for form page
        if (hasForm()) {
            return "form";
        }

        return "unknown";
    }

    /**
     * Check if page has password input
     */
    public boolean hasPasswordInput() {
        return jsExecutor.executeBoolean("document.querySelector('input[type=password]') !== null");
    }

    /**
     * Check if page has text input
     */
    public boolean hasTextInput() {
        return jsExecutor.executeBoolean("document.querySelector('input[type=text], input[type=email]') !== null");
    }

    /**
     * Check if page has search input
     */
    public boolean hasSearchInput() {
        return jsExecutor.executeBoolean(
            "document.querySelector('input[type=search], input[name=q], input[name=query]') !== null"
        );
    }

    /**
     * Check if page has form
     */
    public boolean hasForm() {
        return jsExecutor.executeBoolean("document.querySelector('form') !== null");
    }

    /**
     * Get selector for named element
     */
    @Nullable
    public String getSelector(@NonNull String elementName) {
        String pageType = detectPageType();
        if (pageType == null) return null;

        PagePattern pattern = pagePatterns.get(pageType);
        if (pattern != null) {
            return pattern.selectors.get(elementName);
        }

        return null;
    }

    // ===== Login Operations =====

    /**
     * Auto-fill login form
     */
    @NonNull
    public AutomationResult fillLoginForm(@NonNull String username, @NonNull String password) {
        String usernameSelector = getSelector("username");
        String passwordSelector = getSelector("password");

        if (usernameSelector == null || passwordSelector == null) {
            // Try to auto-detect
            usernameSelector = "input[type='text'], input[type='email'], input[name*='user'], input[name*='email']";
            passwordSelector = "input[type='password']";
        }

        // Fill username
        AutomationResult result = formFiller.fillText(usernameSelector, username);
        if (!result.isSuccess()) {
            return new AutomationResult("fillLoginForm", "Failed to fill username");
        }

        // Fill password
        result = formFiller.fillText(passwordSelector, password);
        if (!result.isSuccess()) {
            return new AutomationResult("fillLoginForm", "Failed to fill password");
        }

        return new AutomationResult("fillLoginForm", 0);
    }

    /**
     * Submit login
     */
    @NonNull
    public AutomationResult submitLogin() {
        String submitSelector = getSelector("submit");
        if (submitSelector == null) {
            submitSelector = "button[type='submit'], input[type='submit']";
        }

        return webViewAutomation.click(submitSelector);
    }

    /**
     * Complete login flow
     */
    @NonNull
    public AutomationResult login(@NonNull String username, @NonNull String password) {
        AutomationResult result = fillLoginForm(username, password);
        if (!result.isSuccess()) {
            return result;
        }

        return submitLogin();
    }

    // ===== Search Operations =====

    /**
     * Perform search
     */
    @NonNull
    public AutomationResult search(@NonNull String query) {
        String searchInputSelector = getSelector("searchInput");
        if (searchInputSelector == null) {
            searchInputSelector = "input[type='search'], input[name='q'], input[name='query']";
        }

        // Fill search input
        AutomationResult result = formFiller.fillText(searchInputSelector, query);
        if (!result.isSuccess()) {
            return new AutomationResult("search", "Failed to fill search input");
        }

        // Submit search
        String searchButtonSelector = getSelector("searchButton");
        if (searchButtonSelector != null) {
            return webViewAutomation.click(searchButtonSelector);
        } else {
            // Press Enter
            return webViewAutomation.executeJs(
                String.format(
                    "var el = document.querySelector('%s'); if (el) { " +
                    "  var e = new KeyboardEvent('keydown', { key: 'Enter', keyCode: 13, bubbles: true });" +
                    "  el.dispatchEvent(e);" +
                    "}",
                    searchInputSelector
                )
            );
        }
    }

    /**
     * Get search results
     */
    @Nullable
    public List<SearchResult> getSearchResults() {
        String resultsSelector = getSelector("results");
        if (resultsSelector == null) {
            resultsSelector = "[class*='result'], .search-result, .item";
        }

        String script = String.format(
            "(function() {" +
            "  var items = document.querySelectorAll('%s');" +
            "  var results = [];" +
            "  for (var i = 0; i < items.length && i < 20; i++) {" +
            "    var item = items[i];" +
            "    var link = item.querySelector('a');" +
            "    var title = item.querySelector('h1, h2, h3, h4, .title');" +
            "    results.push({" +
            "      title: title ? title.textContent.trim() : ''," +
            "      url: link ? link.href : ''," +
            "      text: item.textContent.trim().substring(0, 200)" +
            "    });" +
            "  }" +
            "  return JSON.stringify(results);" +
            "})()",
            resultsSelector
        );

        JSONArray results = jsExecutor.executeJsonArray(script);
        if (results == null) return null;

        List<SearchResult> searchResults = new ArrayList<>();
        for (int i = 0; i < results.length(); i++) {
            try {
                JSONObject item = results.getJSONObject(i);
                searchResults.add(new SearchResult(
                    item.optString("title"),
                    item.optString("url"),
                    item.optString("text")
                ));
            } catch (Exception e) {
                // Skip invalid item
            }
        }

        return searchResults;
    }

    /**
     * Search result
     */
    public static class SearchResult {
        public final String title;
        public final String url;
        public final String snippet;

        public SearchResult(String title, String url, String snippet) {
            this.title = title;
            this.url = url;
            this.snippet = snippet;
        }
    }

    // ===== Form Operations =====

    /**
     * Auto-fill detected form
     */
    @NonNull
    public AutomationResult autoFillForm(@NonNull Map<String, String> data) {
        // Find form
        String formSelector = getSelector("form");
        if (formSelector == null) {
            formSelector = "form";
        }

        // Get form fields
        JSONArray fields = jsExecutor.executeJsonArray(
            String.format(
                "(function() {" +
                "  var form = document.querySelector('%s');" +
                "  if (!form) return [];" +
                "  var inputs = form.querySelectorAll('input, textarea, select');" +
                "  var result = [];" +
                "  for (var i = 0; i < inputs.length; i++) {" +
                "    var el = inputs[i];" +
                "    if (el.type === 'submit' || el.type === 'button') continue;" +
                "    result.push({" +
                "      name: el.name || el.id," +
                "      type: el.type," +
                "      selector: el.id ? '#' + el.id : el.name ? '[name=\"' + el.name + '\"]' : ''" +
                "    });" +
                "  }" +
                "  return JSON.stringify(result);" +
                "})()",
                formSelector
            )
        );

        if (fields == null || fields.length() == 0) {
            return new AutomationResult("autoFillForm", "No form fields found");
        }

        int filled = 0;
        for (int i = 0; i < fields.length(); i++) {
            try {
                JSONObject field = fields.getJSONObject(i);
                String name = field.optString("name");
                String selector = field.optString("selector");
                String type = field.optString("type");

                if (name == null || name.isEmpty() || selector == null || selector.isEmpty()) {
                    continue;
                }

                // Find matching value from data
                String value = findMatchingValue(name, data);
                if (value == null) continue;

                // Fill based on type
                AutomationResult result;
                switch (type) {
                    case "select-one":
                        result = formFiller.fillSelect(selector, value);
                        break;
                    case "checkbox":
                        result = formFiller.setCheckbox(selector, Boolean.parseBoolean(value));
                        break;
                    default:
                        result = formFiller.fillText(selector, value);
                }

                if (result.isSuccess()) {
                    filled++;
                }
            } catch (Exception e) {
                Log.w(TAG, "Error filling field", e);
            }
        }

        return new AutomationResult("autoFillForm", 0);
    }

    /**
     * Find matching value for field name
     */
    @Nullable
    private String findMatchingValue(@NonNull String fieldName, @NonNull Map<String, String> data) {
        String lowerName = fieldName.toLowerCase();

        // Direct match
        if (data.containsKey(fieldName)) {
            return data.get(fieldName);
        }

        // Fuzzy match
        for (Map.Entry<String, String> entry : data.entrySet()) {
            String key = entry.getKey().toLowerCase();
            if (lowerName.contains(key) || key.contains(lowerName)) {
                return entry.getValue();
            }
        }

        // Common field mappings
        if (lowerName.contains("email") || lowerName.contains("mail")) {
            return data.get("email");
        }
        if (lowerName.contains("phone") || lowerName.contains("tel") || lowerName.contains("mobile")) {
            return data.get("phone");
        }
        if (lowerName.contains("name") && !lowerName.contains("user")) {
            return data.get("name");
        }
        if (lowerName.contains("address")) {
            return data.get("address");
        }

        return null;
    }

    /**
     * Submit form
     */
    @NonNull
    public AutomationResult submitForm() {
        String formSelector = getSelector("form");
        if (formSelector == null) {
            formSelector = "form";
        }
        return formFiller.submitForm(formSelector);
    }

    // ===== Page Analysis =====

    /**
     * Analyze page content
     */
    @NonNull
    public PageAnalysis analyzePage() {
        PageAnalysis analysis = new PageAnalysis();

        // Get page type
        analysis.pageType = detectPageType();

        // Get title
        analysis.title = webViewAutomation.getTitle();

        // Get URL
        analysis.url = webViewAutomation.getUrl();

        // Count elements
        analysis.inputCount = jsExecutor.executeInt("document.querySelectorAll('input, textarea, select').length", 0);
        analysis.linkCount = jsExecutor.executeInt("document.querySelectorAll('a').length", 0);
        analysis.imageCount = jsExecutor.executeInt("document.querySelectorAll('img').length", 0);
        analysis.buttonCount = jsExecutor.executeInt("document.querySelectorAll('button, input[type=submit], input[type=button]').length", 0);
        analysis.formCount = jsExecutor.executeInt("document.querySelectorAll('form').length", 0);

        // Check for key elements
        analysis.hasLoginForm = hasPasswordInput() && hasTextInput();
        analysis.hasSearchInput = hasSearchInput();

        return analysis;
    }

    /**
     * Page analysis result
     */
    public static class PageAnalysis {
        public String pageType;
        public String title;
        public String url;
        public int inputCount;
        public int linkCount;
        public int imageCount;
        public int buttonCount;
        public int formCount;
        public boolean hasLoginForm;
        public boolean hasSearchInput;

        @NonNull
        @Override
        public String toString() {
            return String.format("PageAnalysis{type=%s, inputs=%d, forms=%d, links=%d}",
                pageType, inputCount, formCount, linkCount);
        }
    }

    // ===== Utility Methods =====

    /**
     * Take snapshot of current page state
     */
    @NonNull
    public JSONObject snapshot() {
        JSONObject snapshot = webViewAutomation.getPageContent();

        try {
            snapshot.put("pageType", detectPageType());
            snapshot.put("timestamp", System.currentTimeMillis());
        } catch (Exception e) {
            Log.e(TAG, "Error creating snapshot", e);
        }

        return snapshot;
    }

    /**
     * Check if page matches pattern
     */
    public boolean matchesPattern(@NonNull String patternName) {
        String url = webViewAutomation.getUrl();
        PagePattern pattern = pagePatterns.get(patternName);

        if (pattern != null) {
            return url.matches(pattern.urlPattern);
        }

        return false;
    }

    // ===== Getters =====

    public WebFormFiller getFormFiller() {
        return formFiller;
    }

    public WebViewAutomation getWebViewAutomation() {
        return webViewAutomation;
    }
}