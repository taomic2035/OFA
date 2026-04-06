package com.ofa.agent.automation.webview;

import android.os.Handler;
import android.os.Looper;
import android.util.Log;
import android.webkit.ValueCallback;
import android.webkit.WebView;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationResult;

import org.json.JSONObject;

import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicReference;

/**
 * JsExecutor - executes JavaScript code in WebView.
 * Provides synchronous and asynchronous JavaScript execution with result handling.
 */
public class JsExecutor {

    private static final String TAG = "JsExecutor";

    private final WebView webView;
    private final Handler mainHandler;

    // Configuration
    private long defaultTimeout = 10000;

    /**
     * Script execution result
     */
    public static class JsResult {
        public final boolean success;
        public final String result;
        public final String error;
        public final long executionTimeMs;

        public JsResult(boolean success, String result, String error, long executionTimeMs) {
            this.success = success;
            this.result = result;
            this.error = error;
            this.executionTimeMs = executionTimeMs;
        }
    }

    public JsExecutor(@NonNull WebView webView) {
        this.webView = webView;
        this.mainHandler = new Handler(Looper.getMainLooper());
    }

    /**
     * Execute JavaScript synchronously
     */
    @NonNull
    public AutomationResult execute(@NonNull String script) {
        return executeForResult(script, defaultTimeout);
    }

    /**
     * Execute JavaScript and wait for result
     */
    @NonNull
    public AutomationResult executeForResult(@NonNull String script, long timeoutMs) {
        Log.d(TAG, "Executing JS: " + script.substring(0, Math.min(100, script.length())) + "...");

        long startTime = System.currentTimeMillis();

        CountDownLatch latch = new CountDownLatch(1);
        AtomicReference<JsResult> resultRef = new AtomicReference<>();

        mainHandler.post(() -> {
            if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.KITKAT) {
                webView.evaluateJavascript(script, new ValueCallback<String>() {
                    @Override
                    public void onReceiveValue(String value) {
                        long execTime = System.currentTimeMillis() - startTime;
                        resultRef.set(new JsResult(true, value, null, execTime));
                        latch.countDown();
                    }
                });
            } else {
                // Fallback for older versions
                try {
                    webView.loadUrl("javascript:" + script);
                    long execTime = System.currentTimeMillis() - startTime;
                    resultRef.set(new JsResult(true, null, null, execTime));
                    latch.countDown();
                } catch (Exception e) {
                    resultRef.set(new JsResult(false, null, e.getMessage(), 0));
                    latch.countDown();
                }
            }
        });

        try {
            if (latch.await(timeoutMs, TimeUnit.MILLISECONDS)) {
                JsResult jsResult = resultRef.get();
                if (jsResult != null && jsResult.success) {
                    JSONObject data = new JSONObject();
                    try {
                        data.put("result", jsResult.result);
                        data.put("executionTimeMs", jsResult.executionTimeMs);
                    } catch (Exception e) {
                        // Ignore
                    }
                    return new AutomationResult("executeJs", data, jsResult.executionTimeMs);
                } else {
                    return new AutomationResult("executeJs", jsResult != null ? jsResult.error : "Unknown error");
                }
            } else {
                return new AutomationResult("executeJs", "Timeout executing JavaScript");
            }
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            return new AutomationResult("executeJs", "Interrupted");
        }
    }

    /**
     * Evaluate JavaScript expression and return result
     */
    @Nullable
    public String eval(@NonNull String expression) {
        AutomationResult result = executeForResult(expression, defaultTimeout);
        if (result.isSuccess()) {
            String value = result.getDataString("result");
            // Remove quotes around string results
            if (value != null && value.startsWith("\"") && value.endsWith("\"")) {
                value = value.substring(1, value.length() - 1);
                // Unescape
                value = value.replace("\\\"", "\"")
                             .replace("\\n", "\n")
                             .replace("\\t", "\t")
                             .replace("\\\\", "\\");
            }
            return "null".equals(value) ? null : value;
        }
        return null;
    }

    /**
     * Execute JavaScript asynchronously
     */
    public void executeAsync(@NonNull String script, @Nullable JsCallback callback) {
        Log.d(TAG, "Executing JS async: " + script.substring(0, Math.min(100, script.length())) + "...");

        mainHandler.post(() -> {
            if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.KITKAT) {
                webView.evaluateJavascript(script, new ValueCallback<String>() {
                    @Override
                    public void onReceiveValue(String value) {
                        if (callback != null) {
                            callback.onResult(value);
                        }
                    }
                });
            } else {
                webView.loadUrl("javascript:" + script);
                if (callback != null) {
                    callback.onResult(null);
                }
            }
        });
    }

    /**
     * Execute script that returns JSON
     */
    @Nullable
    public JSONObject executeJson(@NonNull String script) {
        String result = eval(script);
        if (result != null && !result.equals("undefined") && !result.equals("null")) {
            try {
                return new JSONObject(result);
            } catch (Exception e) {
                Log.e(TAG, "Failed to parse JSON result: " + e.getMessage());
            }
        }
        return null;
    }

    /**
     * Execute script that returns JSON array
     */
    @Nullable
    public org.json.JSONArray executeJsonArray(@NonNull String script) {
        String result = eval(script);
        if (result != null && !result.equals("undefined") && !result.equals("null")) {
            try {
                return new org.json.JSONArray(result);
            } catch (Exception e) {
                Log.e(TAG, "Failed to parse JSON array result: " + e.getMessage());
            }
        }
        return null;
    }

    /**
     * Execute script and return boolean
     */
    public boolean executeBoolean(@NonNull String script) {
        String result = eval(script);
        return "true".equals(result);
    }

    /**
     * Execute script and return integer
     */
    public int executeInt(@NonNull String script, int defaultValue) {
        String result = eval(script);
        if (result != null) {
            try {
                // Remove decimal if present
                if (result.contains(".")) {
                    result = result.substring(0, result.indexOf("."));
                }
                return Integer.parseInt(result);
            } catch (NumberFormatException e) {
                Log.w(TAG, "Failed to parse int: " + result);
            }
        }
        return defaultValue;
    }

    /**
     * Execute script and return double
     */
    public double executeDouble(@NonNull String script, double defaultValue) {
        String result = eval(script);
        if (result != null) {
            try {
                return Double.parseDouble(result);
            } catch (NumberFormatException e) {
                Log.w(TAG, "Failed to parse double: " + result);
            }
        }
        return defaultValue;
    }

    // ===== Helper Scripts =====

    /**
     * Get document ready state
     */
    @NonNull
    public String getReadyState() {
        String state = eval("document.readyState");
        return state != null ? state : "unknown";
    }

    /**
     * Check if page is fully loaded
     */
    public boolean isPageComplete() {
        return "complete".equals(getReadyState());
    }

    /**
     * Get page scroll position
     */
    @NonNull
    public int[] getScrollPosition() {
        int x = executeInt("window.scrollX", 0);
        int y = executeInt("window.scrollY", 0);
        return new int[] { x, y };
    }

    /**
     * Get page dimensions
     */
    @NonNull
    public int[] getPageDimensions() {
        int width = executeInt("document.body.scrollWidth", 0);
        int height = executeInt("document.body.scrollHeight", 0);
        return new int[] { width, height };
    }

    /**
     * Get viewport dimensions
     */
    @NonNull
    public int[] getViewportDimensions() {
        int width = executeInt("window.innerWidth", 0);
        int height = executeInt("window.innerHeight", 0);
        return new int[] { width, height };
    }

    /**
     * Get element count by selector
     */
    public int getElementCount(@NonNull String selector) {
        String script = String.format("document.querySelectorAll('%s').length", selector);
        return executeInt(script, 0);
    }

    /**
     * Check if element exists
     */
    public boolean hasElement(@NonNull String selector) {
        return getElementCount(selector) > 0;
    }

    /**
     * Get element bounds
     */
    @Nullable
    public int[] getElementBounds(@NonNull String selector) {
        String script = String.format(
            "(function() {" +
            "  var el = document.querySelector('%s');" +
            "  if (el) {" +
            "    var rect = el.getBoundingClientRect();" +
            "    return rect.left + ',' + rect.top + ',' + rect.width + ',' + rect.height;" +
            "  }" +
            "  return null;" +
            "})()",
            escapeString(selector)
        );

        String result = eval(script);
        if (result != null && !result.equals("null")) {
            try {
                String[] parts = result.split(",");
                return new int[] {
                    Integer.parseInt(parts[0].trim()),
                    Integer.parseInt(parts[1].trim()),
                    Integer.parseInt(parts[2].trim()),
                    Integer.parseInt(parts[3].trim())
                };
            } catch (Exception e) {
                Log.w(TAG, "Failed to parse bounds: " + result);
            }
        }
        return null;
    }

    /**
     * Inject helper functions into the page
     */
    public void injectHelpers() {
        String helpers =
            // Wait for element helper
            "window._waitForElement = function(selector, timeout) {" +
            "  return new Promise(function(resolve, reject) {" +
            "    var start = Date.now();" +
            "    function check() {" +
            "      var el = document.querySelector(selector);" +
            "      if (el) { resolve(el); }" +
            "      else if (Date.now() - start > timeout) { reject(new Error('Timeout')); }" +
            "      else { setTimeout(check, 100); }" +
            "    }" +
            "    check();" +
            "  });" +
            "};" +

            // Get all forms helper
            "window._getAllForms = function() {" +
            "  var forms = document.querySelectorAll('form');" +
            "  var result = [];" +
            "  for (var i = 0; i < forms.length; i++) {" +
            "    var form = forms[i];" +
            "    result.push({" +
            "      id: form.id," +
            "      name: form.name," +
            "      action: form.action," +
            "      method: form.method" +
            "    });" +
            "  }" +
            "  return result;" +
            "};" +

            // Get visible text helper
            "window._getVisibleText = function() {" +
            "  return document.body.innerText;" +
            "};" +

            // Check if element is visible
            "window._isVisible = function(selector) {" +
            "  var el = document.querySelector(selector);" +
            "  if (!el) return false;" +
            "  var rect = el.getBoundingClientRect();" +
            "  return rect.width > 0 && rect.height > 0;" +
            "};";

        execute(helpers);
        Log.d(TAG, "Helper functions injected");
    }

    /**
     * Escape string for JavaScript
     */
    @NonNull
    private String escapeString(@NonNull String s) {
        return s.replace("\\", "\\\\")
                .replace("'", "\\'")
                .replace("\"", "\\\"")
                .replace("\n", "\\n")
                .replace("\r", "\\r")
                .replace("\t", "\\t");
    }

    // ===== Configuration =====

    public void setDefaultTimeout(long timeoutMs) {
        this.defaultTimeout = timeoutMs;
    }

    /**
     * Callback interface for async execution
     */
    public interface JsCallback {
        void onResult(@Nullable String result);
    }
}