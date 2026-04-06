package com.ofa.agent.automation.webview;

import android.os.Handler;
import android.os.Looper;
import android.util.Log;
import android.webkit.WebView;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationResult;

import org.json.JSONArray;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;

/**
 * WebEventListener - monitors and captures web page events.
 * Supports click, input, form submit, navigation, and custom events.
 */
public class WebEventListener {

    private static final String TAG = "WebEventListener";

    private final WebView webView;
    private final JsExecutor jsExecutor;
    private final Handler mainHandler;

    // Event storage
    private final Map<String, EventCallback> eventCallbacks = new ConcurrentHashMap<>();
    private final List<CapturedEvent> capturedEvents = new ArrayList<>();
    private final AtomicLong eventIdCounter = new AtomicLong(0);

    // Configuration
    private boolean capturing = false;
    private int maxCapturedEvents = 100;

    /**
     * Captured event
     */
    public static class CapturedEvent {
        public final long id;
        public final String type;
        public final String selector;
        public final JSONObject data;
        public final long timestamp;

        public CapturedEvent(long id, String type, String selector, JSONObject data) {
            this.id = id;
            this.type = type;
            this.selector = selector;
            this.data = data;
            this.timestamp = System.currentTimeMillis();
        }
    }

    /**
     * Event callback interface
     */
    public interface EventCallback {
        void onEvent(CapturedEvent event);
    }

    public WebEventListener(@NonNull WebView webView, @NonNull JsExecutor jsExecutor) {
        this.webView = webView;
        this.jsExecutor = jsExecutor;
        this.mainHandler = new Handler(Looper.getMainLooper());
    }

    /**
     * Start capturing events
     */
    public void startCapturing() {
        if (capturing) return;

        capturing = true;
        capturedEvents.clear();

        // Inject event capture script
        injectEventCaptureScript();

        Log.i(TAG, "Event capturing started");
    }

    /**
     * Stop capturing events
     */
    public void stopCapturing() {
        if (!capturing) return;

        capturing = false;

        // Remove event listeners
        jsExecutor.execute(
            "document.removeEventListener('click', window._ofaCaptureClick, true);" +
            "document.removeEventListener('input', window._ofaCaptureInput, true);" +
            "document.removeEventListener('change', window._ofaCaptureChange, true);" +
            "document.removeEventListener('submit', window._ofaCaptureSubmit, true);"
        );

        Log.i(TAG, "Event capturing stopped, captured " + capturedEvents.size() + " events");
    }

    /**
     * Inject event capture script
     */
    private void injectEventCaptureScript() {
        String script =
            // Click handler
            "window._ofaCaptureClick = function(e) {" +
            "  var target = e.target;" +
            "  var selector = _ofaGetSelector(target);" +
            "  var data = {" +
            "    tagName: target.tagName," +
            "    text: target.textContent ? target.textContent.substring(0, 100) : ''," +
            "    className: target.className," +
            "    id: target.id," +
            "    href: target.href || ''" +
            "  };" +
            "  window._ofaEventQueue = window._ofaEventQueue || [];" +
            "  window._ofaEventQueue.push({ type: 'click', selector: selector, data: data });" +
            "};" +

            // Input handler
            "window._ofaCaptureInput = function(e) {" +
            "  var target = e.target;" +
            "  var selector = _ofaGetSelector(target);" +
            "  var data = {" +
            "    tagName: target.tagName," +
            "    name: target.name," +
            "    value: target.value," +
            "    type: target.type" +
            "  };" +
            "  window._ofaEventQueue = window._ofaEventQueue || [];" +
            "  window._ofaEventQueue.push({ type: 'input', selector: selector, data: data });" +
            "};" +

            // Change handler
            "window._ofaCaptureChange = function(e) {" +
            "  var target = e.target;" +
            "  var selector = _ofaGetSelector(target);" +
            "  var data = {" +
            "    tagName: target.tagName," +
            "    name: target.name," +
            "    value: target.value," +
            "    checked: target.checked" +
            "  };" +
            "  window._ofaEventQueue = window._ofaEventQueue || [];" +
            "  window._ofaEventQueue.push({ type: 'change', selector: selector, data: data });" +
            "};" +

            // Submit handler
            "window._ofaCaptureSubmit = function(e) {" +
            "  var target = e.target;" +
            "  var selector = _ofaGetSelector(target);" +
            "  var data = {" +
            "    action: target.action," +
            "    method: target.method," +
            "    name: target.name," +
            "    id: target.id" +
            "  };" +
            "  window._ofaEventQueue = window._ofaEventQueue || [];" +
            "  window._ofaEventQueue.push({ type: 'submit', selector: selector, data: data });" +
            "};" +

            // Helper to get selector
            "function _ofaGetSelector(el) {" +
            "  if (el.id) return '#' + el.id;" +
            "  var path = [];" +
            "  while (el && el.nodeType === Node.ELEMENT_NODE) {" +
            "    var selector = el.tagName.toLowerCase();" +
            "    if (el.className) {" +
            "      selector += '.' + el.className.trim().split(/\\s+/).join('.');" +
            "    }" +
            "    path.unshift(selector);" +
            "    el = el.parentElement;" +
            "  }" +
            "  return path.join(' > ');" +
            "}" +

            // Register listeners
            "document.addEventListener('click', window._ofaCaptureClick, true);" +
            "document.addEventListener('input', window._ofaCaptureInput, true);" +
            "document.addEventListener('change', window._ofaCaptureChange, true);" +
            "document.addEventListener('submit', window._ofaCaptureSubmit, true);" +

            // Initialize queue
            "window._ofaEventQueue = [];";

        jsExecutor.execute(script);
    }

    /**
     * Get captured events
     */
    @NonNull
    public List<CapturedEvent> getCapturedEvents() {
        return new ArrayList<>(capturedEvents);
    }

    /**
     * Poll for new events from JavaScript queue
     */
    @NonNull
    public List<CapturedEvent> pollEvents() {
        List<CapturedEvent> newEvents = new ArrayList<>();

        // Get events from JS queue
        String result = jsExecutor.eval(
            "(function() {" +
            "  var events = window._ofaEventQueue || [];" +
            "  window._ofaEventQueue = [];" +
            "  return JSON.stringify(events);" +
            "})()"
        );

        if (result != null && !result.equals("[]") && !result.equals("null")) {
            try {
                JSONArray events = new JSONArray(result);
                for (int i = 0; i < events.length(); i++) {
                    JSONObject event = events.getJSONObject(i);
                    long id = eventIdCounter.incrementAndGet();
                    CapturedEvent ce = new CapturedEvent(
                        id,
                        event.optString("type"),
                        event.optString("selector"),
                        event.optJSONObject("data")
                    );
                    newEvents.add(ce);

                    // Add to captured events
                    synchronized (capturedEvents) {
                        capturedEvents.add(ce);
                        // Trim if over limit
                        while (capturedEvents.size() > maxCapturedEvents) {
                            capturedEvents.remove(0);
                        }
                    }

                    // Notify callback
                    EventCallback callback = eventCallbacks.get(ce.type);
                    if (callback != null) {
                        callback.onEvent(ce);
                    }
                }
            } catch (Exception e) {
                Log.e(TAG, "Error parsing events", e);
            }
        }

        return newEvents;
    }

    /**
     * Clear captured events
     */
    public void clearCapturedEvents() {
        synchronized (capturedEvents) {
            capturedEvents.clear();
        }
        jsExecutor.execute("window._ofaEventQueue = [];");
    }

    /**
     * Register callback for specific event type
     */
    public void registerCallback(@NonNull String eventType, @NonNull EventCallback callback) {
        eventCallbacks.put(eventType, callback);
    }

    /**
     * Unregister callback
     */
    public void unregisterCallback(@NonNull String eventType) {
        eventCallbacks.remove(eventType);
    }

    /**
     * Wait for specific event type
     */
    @Nullable
    public CapturedEvent waitForEvent(@NonNull String eventType, long timeoutMs) {
        long startTime = System.currentTimeMillis();

        while (System.currentTimeMillis() - startTime < timeoutMs) {
            List<CapturedEvent> events = pollEvents();
            for (CapturedEvent event : events) {
                if (event.type.equals(eventType)) {
                    return event;
                }
            }

            try {
                Thread.sleep(100);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                break;
            }
        }

        return null;
    }

    /**
     * Wait for click on specific element
     */
    @Nullable
    public CapturedEvent waitForClick(@NonNull String selector, long timeoutMs) {
        long startTime = System.currentTimeMillis();

        while (System.currentTimeMillis() - startTime < timeoutMs) {
            List<CapturedEvent> events = pollEvents();
            for (CapturedEvent event : events) {
                if ("click".equals(event.type) &&
                    (event.selector.contains(selector) || selector.equals(event.selector))) {
                    return event;
                }
            }

            try {
                Thread.sleep(100);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                break;
            }
        }

        return null;
    }

    // ===== Custom Event Monitoring =====

    /**
     * Monitor element for specific event
     */
    @NonNull
    public AutomationResult monitorElement(@NonNull String selector, @NonNull String eventType,
                                            @NonNull EventCallback callback) {
        String escapedSelector = escapeJsString(selector);
        String escapedEventType = escapeJsString(eventType);
        String callbackId = selector + "_" + eventType;

        String script = String.format(
            "(function() {" +
            "  var el = document.querySelector('%s');" +
            "  if (!el) return false;" +
            "  el.addEventListener('%s', function(e) {" +
            "    var selector = '%s';" +
            "    window._ofaEventQueue = window._ofaEventQueue || [];" +
            "    window._ofaEventQueue.push({ type: '%s', selector: selector, data: {} });" +
            "  });" +
            "  return true;" +
            "})()",
            escapedSelector, escapedEventType, escapedSelector, escapedEventType
        );

        AutomationResult result = jsExecutor.execute(script);
        if (result.isSuccess()) {
            registerCallback(eventType, callback);
        }

        return result;
    }

    /**
     * Monitor for element appearance
     */
    @Nullable
    public CapturedEvent waitForElementAppear(@NonNull String selector, long timeoutMs) {
        long startTime = System.currentTimeMillis();

        while (System.currentTimeMillis() - startTime < timeoutMs) {
            Boolean exists = jsExecutor.executeBoolean(
                String.format("document.querySelector('%s') !== null", escapeJsString(selector))
            );

            if (exists) {
                return new CapturedEvent(
                    eventIdCounter.incrementAndGet(),
                    "elementAppear",
                    selector,
                    new JSONObject()
                );
            }

            try {
                Thread.sleep(200);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                break;
            }
        }

        return null;
    }

    /**
     * Monitor for element disappearance
     */
    @Nullable
    public CapturedEvent waitForElementDisappear(@NonNull String selector, long timeoutMs) {
        long startTime = System.currentTimeMillis();

        while (System.currentTimeMillis() - startTime < timeoutMs) {
            Boolean exists = jsExecutor.executeBoolean(
                String.format("document.querySelector('%s') !== null", escapeJsString(selector))
            );

            if (!exists) {
                return new CapturedEvent(
                    eventIdCounter.incrementAndGet(),
                    "elementDisappear",
                    selector,
                    new JSONObject()
                );
            }

            try {
                Thread.sleep(200);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                break;
            }
        }

        return null;
    }

    /**
     * Monitor for text change
     */
    @Nullable
    public CapturedEvent waitForTextChange(@NonNull String selector, long timeoutMs) {
        String initialText = jsExecutor.eval(
            String.format(
                "(function() {" +
                "  var el = document.querySelector('%s');" +
                "  return el ? el.textContent : null;" +
                "})()",
                escapeJsString(selector)
            )
        );

        long startTime = System.currentTimeMillis();

        while (System.currentTimeMillis() - startTime < timeoutMs) {
            String currentText = jsExecutor.eval(
                String.format(
                    "(function() {" +
                    "  var el = document.querySelector('%s');" +
                    "  return el ? el.textContent : null;" +
                    "})()",
                    escapeJsString(selector)
                )
            );

            if (currentText != null && !currentText.equals(initialText)) {
                try {
                    JSONObject data = new JSONObject();
                    data.put("oldText", initialText);
                    data.put("newText", currentText);
                    return new CapturedEvent(
                        eventIdCounter.incrementAndGet(),
                        "textChange",
                        selector,
                        data
                    );
                } catch (Exception e) {
                    // Ignore
                }
            }

            try {
                Thread.sleep(200);
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                break;
            }
        }

        return null;
    }

    /**
     * Get event summary as JSON
     */
    @NonNull
    public JSONObject getEventSummary() {
        JSONObject summary = new JSONObject();

        try {
            summary.put("capturing", capturing);
            summary.put("totalCaptured", capturedEvents.size());

            // Count by type
            JSONObject byType = new JSONObject();
            for (CapturedEvent event : capturedEvents) {
                int count = byType.optInt(event.type, 0);
                byType.put(event.type, count + 1);
            }
            summary.put("byType", byType);

        } catch (Exception e) {
            Log.e(TAG, "Error creating event summary", e);
        }

        return summary;
    }

    /**
     * Export captured events as JSON array
     */
    @NonNull
    public JSONArray exportEvents() {
        JSONArray events = new JSONArray();

        synchronized (capturedEvents) {
            for (CapturedEvent event : capturedEvents) {
                try {
                    JSONObject obj = new JSONObject();
                    obj.put("id", event.id);
                    obj.put("type", event.type);
                    obj.put("selector", event.selector);
                    obj.put("timestamp", event.timestamp);
                    if (event.data != null) {
                        obj.put("data", event.data);
                    }
                    events.put(obj);
                } catch (Exception e) {
                    // Skip invalid event
                }
            }
        }

        return events;
    }

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

    // ===== Configuration =====

    public void setMaxCapturedEvents(int max) {
        this.maxCapturedEvents = max;
    }

    public boolean isCapturing() {
        return capturing;
    }
}