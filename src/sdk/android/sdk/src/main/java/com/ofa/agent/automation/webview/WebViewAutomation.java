package com.ofa.agent.automation.webview;

import android.graphics.Bitmap;
import android.os.Build;
import android.os.Handler;
import android.os.Looper;
import android.util.Log;
import android.webkit.WebResourceError;
import android.webkit.WebResourceRequest;
import android.webkit.WebView;
import android.webkit.WebViewClient;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationResult;

import org.json.JSONArray;
import org.json.JSONObject;

import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicReference;

/**
 * WebView Automation - enables automation within WebView components.
 * Provides JavaScript execution, element manipulation, and event monitoring.
 */
public class WebViewAutomation {

    private static final String TAG = "WebViewAutomation";

    private final WebView webView;
    private final Handler mainHandler;
    private final JsExecutor jsExecutor;

    // State
    private boolean pageLoaded = false;
    private String currentUrl = "";
    private String pageTitle = "";
    private String lastError = null;

    // Configuration
    private long defaultTimeout = 10000; // 10 seconds
    private boolean javascriptEnabled = true;

    // Listeners
    private PageLoadListener pageLoadListener;
    private ErrorListener errorListener;

    /**
     * Page load listener
     */
    public interface PageLoadListener {
        void onPageStarted(String url);
        void onPageFinished(String url);
        void onPageError(String url, String error);
    }

    /**
     * Error listener
     */
    public interface ErrorListener {
        void onError(String error);
    }

    public WebViewAutomation(@NonNull WebView webView) {
        this.webView = webView;
        this.mainHandler = new Handler(Looper.getMainLooper());
        this.jsExecutor = new JsExecutor(webView);

        setupWebViewClient();
        enableJavaScript();
    }

    /**
     * Setup WebView client for monitoring
     */
    private void setupWebViewClient() {
        webView.setWebViewClient(new WebViewClient() {
            @Override
            public void onPageStarted(WebView view, String url, Bitmap favicon) {
                super.onPageStarted(view, url, favicon);
                pageLoaded = false;
                currentUrl = url;
                lastError = null;
                Log.d(TAG, "Page started: " + url);

                if (pageLoadListener != null) {
                    pageLoadListener.onPageStarted(url);
                }
            }

            @Override
            public void onPageFinished(WebView view, String url) {
                super.onPageFinished(view, url);
                pageLoaded = true;
                currentUrl = url;
                pageTitle = view.getTitle();
                Log.d(TAG, "Page finished: " + url);

                if (pageLoadListener != null) {
                    pageLoadListener.onPageFinished(url);
                }
            }

            @Override
            public void onReceivedError(WebView view, WebResourceRequest request, WebResourceError error) {
                super.onReceivedError(view, request, error);
                String errorDescription = error.getDescription() != null ? error.getDescription().toString() : "Unknown error";
                lastError = errorDescription;
                Log.e(TAG, "Page error: " + errorDescription);

                if (pageLoadListener != null) {
                    pageLoadListener.onPageError(request.getUrl().toString(), errorDescription);
                }

                if (errorListener != null) {
                    errorListener.onError(errorDescription);
                }
            }
        });
    }

    /**
     * Enable JavaScript
     */
    private void enableJavaScript() {
        if (webView.getSettings() != null) {
            webView.getSettings().setJavaScriptEnabled(true);
            webView.getSettings().setDomStorageEnabled(true);
            webView.getSettings().setJavaScriptCanOpenWindowsAutomatically(true);
            Log.d(TAG, "JavaScript enabled");
        }
    }

    // ===== Page Operations =====

    /**
     * Load URL
     */
    @NonNull
    public AutomationResult loadUrl(@NonNull String url) {
        return loadUrl(url, defaultTimeout);
    }

    /**
     * Load URL with timeout
     */
    @NonNull
    public AutomationResult loadUrl(@NonNull String url, long timeoutMs) {
        Log.i(TAG, "Loading URL: " + url);

        CountDownLatch latch = new CountDownLatch(1);
        AtomicReference<Boolean> success = new AtomicReference<>(false);

        PageLoadListener listener = new PageLoadListener() {
            @Override
            public void onPageStarted(String u) {}

            @Override
            public void onPageFinished(String u) {
                if (u.equals(url)) {
                    success.set(true);
                    latch.countDown();
                }
            }

            @Override
            public void onPageError(String u, String error) {
                if (u.equals(url)) {
                    latch.countDown();
                }
            }
        };

        pageLoadListener = listener;

        mainHandler.post(() -> webView.loadUrl(url));

        try {
            if (latch.await(timeoutMs, TimeUnit.MILLISECONDS)) {
                if (success.get()) {
                    return new AutomationResult("loadUrl", 0);
                } else {
                    return new AutomationResult("loadUrl", "Failed to load: " + lastError);
                }
            } else {
                return new AutomationResult("loadUrl", "Timeout loading URL");
            }
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            return new AutomationResult("loadUrl", "Interrupted");
        } finally {
            pageLoadListener = null;
        }
    }

    /**
     * Wait for page load
     */
    @NonNull
    public AutomationResult waitForPageLoad(long timeoutMs) {
        if (pageLoaded) {
            return new AutomationResult("waitForPageLoad", 0);
        }

        CountDownLatch latch = new CountDownLatch(1);

        pageLoadListener = new PageLoadListener() {
            @Override
            public void onPageStarted(String url) {}

            @Override
            public void onPageFinished(String url) {
                latch.countDown();
            }

            @Override
            public void onPageError(String url, String error) {
                latch.countDown();
            }
        };

        try {
            if (latch.await(timeoutMs, TimeUnit.MILLISECONDS)) {
                return new AutomationResult("waitForPageLoad", 0);
            } else {
                return new AutomationResult("waitForPageLoad", "Timeout waiting for page load");
            }
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            return new AutomationResult("waitForPageLoad", "Interrupted");
        } finally {
            pageLoadListener = null;
        }
    }

    /**
     * Go back
     */
    @NonNull
    public AutomationResult goBack() {
        if (webView.canGoBack()) {
            mainHandler.post(() -> webView.goBack());
            return new AutomationResult("goBack", 0);
        }
        return new AutomationResult("goBack", "Cannot go back");
    }

    /**
     * Go forward
     */
    @NonNull
    public AutomationResult goForward() {
        if (webView.canGoForward()) {
            mainHandler.post(() -> webView.goForward());
            return new AutomationResult("goForward", 0);
        }
        return new AutomationResult("goForward", "Cannot go forward");
    }

    /**
     * Reload page
     */
    @NonNull
    public AutomationResult reload() {
        mainHandler.post(() -> webView.reload());
        return new AutomationResult("reload", 0);
    }

    // ===== JavaScript Operations =====

    /**
     * Execute JavaScript
     */
    @NonNull
    public AutomationResult executeJs(@NonNull String script) {
        return jsExecutor.execute(script);
    }

    /**
     * Execute JavaScript with result callback
     */
    @NonNull
    public AutomationResult executeJsForResult(@NonNull String script, long timeoutMs) {
        return jsExecutor.executeForResult(script, timeoutMs);
    }

    /**
     * Execute JavaScript that returns a value
     */
    @Nullable
    public String evalJs(@NonNull String expression) {
        return jsExecutor.eval(expression);
    }

    // ===== Element Operations =====

    /**
     * Click element by selector
     */
    @NonNull
    public AutomationResult click(@NonNull String selector) {
        String script = String.format(
            "(function() {" +
            "  var el = document.querySelector('%s');" +
            "  if (el) { el.click(); return true; }" +
            "  return false;" +
            "})()",
            escapeJsString(selector)
        );

        AutomationResult result = jsExecutor.executeForResult(script, defaultTimeout);
        if (result.isSuccess() && "true".equals(result.getDataString("result"))) {
            return new AutomationResult("click", 0);
        }
        return new AutomationResult("click", "Element not found: " + selector);
    }

    /**
     * Click element by XPath
     */
    @NonNull
    public AutomationResult clickByXPath(@NonNull String xpath) {
        String script = String.format(
            "(function() {" +
            "  var result = document.evaluate('%s', document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null);" +
            "  var el = result.singleNodeValue;" +
            "  if (el) { el.click(); return true; }" +
            "  return false;" +
            "})()",
            escapeJsString(xpath)
        );

        AutomationResult result = jsExecutor.executeForResult(script, defaultTimeout);
        if (result.isSuccess() && "true".equals(result.getDataString("result"))) {
            return new AutomationResult("clickByXPath", 0);
        }
        return new AutomationResult("clickByXPath", "Element not found: " + xpath);
    }

    /**
     * Input text into element
     */
    @NonNull
    public AutomationResult input(@NonNull String selector, @NonNull String text) {
        String script = String.format(
            "(function() {" +
            "  var el = document.querySelector('%s');" +
            "  if (el) { el.value = '%s'; el.dispatchEvent(new Event('input', { bubbles: true })); return true; }" +
            "  return false;" +
            "})()",
            escapeJsString(selector),
            escapeJsString(text)
        );

        AutomationResult result = jsExecutor.executeForResult(script, defaultTimeout);
        if (result.isSuccess() && "true".equals(result.getDataString("result"))) {
            return new AutomationResult("input", 0);
        }
        return new AutomationResult("input", "Element not found: " + selector);
    }

    /**
     * Get element text
     */
    @Nullable
    public String getText(@NonNull String selector) {
        String script = String.format(
            "(function() {" +
            "  var el = document.querySelector('%s');" +
            "  return el ? el.textContent || el.innerText : null;" +
            "})()",
            escapeJsString(selector)
        );

        return evalJs(script);
    }

    /**
     * Get element value
     */
    @Nullable
    public String getValue(@NonNull String selector) {
        String script = String.format(
            "(function() {" +
            "  var el = document.querySelector('%s');" +
            "  return el ? el.value : null;" +
            "})()",
            escapeJsString(selector)
        );

        return evalJs(script);
    }

    /**
     * Get element attribute
     */
    @Nullable
    public String getAttribute(@NonNull String selector, @NonNull String attribute) {
        String script = String.format(
            "(function() {" +
            "  var el = document.querySelector('%s');" +
            "  return el ? el.getAttribute('%s') : null;" +
            "})()",
            escapeJsString(selector),
            escapeJsString(attribute)
        );

        return evalJs(script);
    }

    /**
     * Check if element exists
     */
    public boolean elementExists(@NonNull String selector) {
        String script = String.format(
            "(function() {" +
            "  return document.querySelector('%s') !== null;" +
            "})()",
            escapeJsString(selector)
        );

        String result = evalJs(script);
        return "true".equals(result);
    }

    /**
     * Wait for element
     */
    @NonNull
    public AutomationResult waitForElement(@NonNull String selector, long timeoutMs) {
        long startTime = System.currentTimeMillis();
        long interval = 200;

        while (System.currentTimeMillis() - startTime < timeoutMs) {
            if (elementExists(selector)) {
                return new AutomationResult("waitForElement", 0);
            }

            try {
                Thread.sleep(interval);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                break;
            }
        }

        return new AutomationResult("waitForElement", "Element not found within timeout: " + selector);
    }

    // ===== Page Information =====

    /**
     * Get page URL
     */
    @NonNull
    public String getUrl() {
        return currentUrl;
    }

    /**
     * Get page title
     */
    @NonNull
    public String getTitle() {
        return pageTitle;
    }

    /**
     * Get page HTML
     */
    @Nullable
    public String getHtml() {
        return evalJs("document.documentElement.outerHTML");
    }

    /**
     * Get page content as JSON
     */
    @NonNull
    public JSONObject getPageContent() {
        JSONObject content = new JSONObject();

        try {
            content.put("url", currentUrl);
            content.put("title", pageTitle);
            content.put("loaded", pageLoaded);

            // Get all input elements
            String inputsScript = "(function() {" +
                "  var inputs = document.querySelectorAll('input, textarea, select');" +
                "  var result = [];" +
                "  for (var i = 0; i < inputs.length; i++) {" +
                "    var el = inputs[i];" +
                "    result.push({" +
                "      type: el.type || el.tagName.toLowerCase()," +
                "      name: el.name," +
                "      id: el.id," +
                "      value: el.value," +
                "      placeholder: el.placeholder" +
                "    });" +
                "  }" +
                "  return JSON.stringify(result);" +
                "})()";

            String inputsJson = evalJs(inputsScript);
            if (inputsJson != null) {
                content.put("inputs", new JSONArray(inputsJson));
            }

            // Get all links
            String linksScript = "(function() {" +
                "  var links = document.querySelectorAll('a');" +
                "  var result = [];" +
                "  for (var i = 0; i < links.length && i < 20; i++) {" +
                "    var el = links[i];" +
                "    result.push({" +
                "      href: el.href," +
                "      text: el.textContent.trim().substring(0, 50)" +
                "    });" +
                "  }" +
                "  return JSON.stringify(result);" +
                "})()";

            String linksJson = evalJs(linksScript);
            if (linksJson != null) {
                content.put("links", new JSONArray(linksJson));
            }

            // Get all buttons
            String buttonsScript = "(function() {" +
                "  var buttons = document.querySelectorAll('button, input[type=submit], input[type=button]');" +
                "  var result = [];" +
                "  for (var i = 0; i < buttons.length; i++) {" +
                "    var el = buttons[i];" +
                "    result.push({" +
                "      text: (el.textContent || el.value || '').trim().substring(0, 30)," +
                "      id: el.id," +
                "      className: el.className" +
                "    });" +
                "  }" +
                "  return JSON.stringify(result);" +
                "})()";

            String buttonsJson = evalJs(buttonsScript);
            if (buttonsJson != null) {
                content.put("buttons", new JSONArray(buttonsJson));
            }

        } catch (Exception e) {
            Log.e(TAG, "Error getting page content", e);
        }

        return content;
    }

    // ===== Utility Methods =====

    /**
     * Escape JavaScript string
     */
    @NonNull
    private String escapeJsString(@NonNull String s) {
        return s.replace("\\", "\\\\")
                .replace("'", "\\'")
                .replace("\"", "\\\"")
                .replace("\n", "\\n")
                .replace("\r", "\\r")
                .replace("\t", "\\t");
    }

    /**
     * Scroll to element
     */
    @NonNull
    public AutomationResult scrollToElement(@NonNull String selector) {
        String script = String.format(
            "(function() {" +
            "  var el = document.querySelector('%s');" +
            "  if (el) { el.scrollIntoView({ behavior: 'smooth', block: 'center' }); return true; }" +
            "  return false;" +
            "})()",
            escapeJsString(selector)
        );

        AutomationResult result = jsExecutor.executeForResult(script, defaultTimeout);
        if (result.isSuccess() && "true".equals(result.getDataString("result"))) {
            return new AutomationResult("scrollToElement", 0);
        }
        return new AutomationResult("scrollToElement", "Element not found: " + selector);
    }

    /**
     * Scroll page
     */
    @NonNull
    public AutomationResult scroll(int x, int y) {
        String script = String.format("window.scrollTo(%d, %d);", x, y);
        return jsExecutor.execute(script);
    }

    /**
     * Scroll to top
     */
    @NonNull
    public AutomationResult scrollToTop() {
        return jsExecutor.execute("window.scrollTo(0, 0);");
    }

    /**
     * Scroll to bottom
     */
    @NonNull
    public AutomationResult scrollToBottom() {
        return jsExecutor.execute("window.scrollTo(0, document.body.scrollHeight);");
    }

    /**
     * Inject custom script
     */
    @NonNull
    public AutomationResult injectScript(@NonNull String scriptContent) {
        String script = String.format(
            "(function() {" +
            "  var script = document.createElement('script');" +
            "  script.textContent = '%s';" +
            "  document.head.appendChild(script);" +
            "  return true;" +
            "})()",
            escapeJsString(scriptContent)
        );

        return jsExecutor.execute(script);
    }

    // ===== Configuration =====

    public void setDefaultTimeout(long timeoutMs) {
        this.defaultTimeout = timeoutMs;
    }

    public void setPageLoadListener(@Nullable PageLoadListener listener) {
        this.pageLoadListener = listener;
    }

    public void setErrorListener(@Nullable ErrorListener listener) {
        this.errorListener = listener;
    }

    // ===== Getters =====

    public WebView getWebView() {
        return webView;
    }

    public JsExecutor getJsExecutor() {
        return jsExecutor;
    }

    public boolean isPageLoaded() {
        return pageLoaded;
    }
}