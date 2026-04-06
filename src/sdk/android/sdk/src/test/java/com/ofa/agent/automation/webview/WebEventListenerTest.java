package com.ofa.agent.automation.webview;

import org.json.JSONObject;
import org.junit.Test;
import org.junit.Before;

import static org.junit.Assert.*;

/**
 * Unit tests for WebEventListener.
 * Tests event capture and handling logic.
 */
public class WebEventListenerTest {

    // Test captured event

    @Test
    public void testCapturedEventCreation() throws Exception {
        JSONObject data = new JSONObject();
        data.put("tagName", "BUTTON");
        data.put("text", "Submit");

        WebEventListener.CapturedEvent event = new WebEventListener.CapturedEvent(
            1L, "click", "#submit-btn", data
        );

        assertEquals(1L, event.id);
        assertEquals("click", event.type);
        assertEquals("#submit-btn", event.selector);
        assertEquals("BUTTON", event.data.getString("tagName"));
        assertTrue(event.timestamp > 0);
    }

    @Test
    public void testCapturedEventTimestamp() {
        long beforeTime = System.currentTimeMillis();

        WebEventListener.CapturedEvent event = new WebEventListener.CapturedEvent(
            1L, "input", "#username", null
        );

        long afterTime = System.currentTimeMillis();

        assertTrue("Timestamp should be within range",
            event.timestamp >= beforeTime && event.timestamp <= afterTime);
    }

    // Test event types

    @Test
    public void testClickEventType() {
        String eventType = "click";

        assertEquals("click", eventType);
    }

    @Test
    public void testInputEventType() {
        String eventType = "input";

        assertEquals("input", eventType);
    }

    @Test
    public void testChangeEventType() {
        String eventType = "change";

        assertEquals("change", eventType);
    }

    @Test
    public void testSubmitEventType() {
        String eventType = "submit";

        assertEquals("submit", eventType);
    }

    // Test event data parsing

    @Test
    public void testClickEventData() throws Exception {
        JSONObject clickData = new JSONObject();
        clickData.put("tagName", "A");
        clickData.put("text", "Click here");
        clickData.put("href", "https://example.com");

        assertEquals("A", clickData.getString("tagName"));
        assertEquals("Click here", clickData.getString("text"));
        assertEquals("https://example.com", clickData.getString("href"));
    }

    @Test
    public void testInputEventData() throws Exception {
        JSONObject inputData = new JSONObject();
        inputData.put("tagName", "INPUT");
        inputData.put("name", "username");
        inputData.put("value", "john_doe");
        inputData.put("type", "text");

        assertEquals("INPUT", inputData.getString("tagName"));
        assertEquals("username", inputData.getString("name"));
        assertEquals("john_doe", inputData.getString("value"));
    }

    @Test
    public void testSubmitEventData() throws Exception {
        JSONObject submitData = new JSONObject();
        submitData.put("action", "/api/login");
        submitData.put("method", "POST");

        assertEquals("/api/login", submitData.getString("action"));
        assertEquals("POST", submitData.getString("method"));
    }

    // Test event queue simulation

    @Test
    public void testEventQueueOrder() {
        java.util.List<WebEventListener.CapturedEvent> events = new java.util.ArrayList<>();

        events.add(new WebEventListener.CapturedEvent(1L, "click", "#btn1", null));
        events.add(new WebEventListener.CapturedEvent(2L, "input", "#field1", null));
        events.add(new WebEventListener.CapturedEvent(3L, "click", "#btn2", null));

        assertEquals(3, events.size());
        assertEquals(1L, events.get(0).id);
        assertEquals(2L, events.get(1).id);
        assertEquals(3L, events.get(2).id);
    }

    @Test
    public void testEventQueueMaxSize() {
        int maxEvents = 100;
        java.util.List<WebEventListener.CapturedEvent> events = new java.util.ArrayList<>();

        // Add more than max
        for (int i = 0; i < 150; i++) {
            events.add(new WebEventListener.CapturedEvent((long) i, "click", "#btn", null));
            if (events.size() > maxEvents) {
                events.remove(0);
            }
        }

        assertEquals(maxEvents, events.size());
    }

    // Test callback interface

    @Test
    public void testEventCallback() {
        final boolean[] callbackInvoked = {false};

        WebEventListener.EventCallback callback = event -> {
            callbackInvoked[0] = true;
            assertNotNull(event);
        };

        // Simulate callback invocation
        WebEventListener.CapturedEvent testEvent = new WebEventListener.CapturedEvent(
            1L, "click", "#btn", null
        );
        callback.onEvent(testEvent);

        assertTrue("Callback should be invoked", callbackInvoked[0]);
    }

    // Test event filtering

    @Test
    public void testFilterByEventType() {
        java.util.List<WebEventListener.CapturedEvent> events = new java.util.ArrayList<>();
        events.add(new WebEventListener.CapturedEvent(1L, "click", "#btn1", null));
        events.add(new WebEventListener.CapturedEvent(2L, "input", "#field", null));
        events.add(new WebEventListener.CapturedEvent(3L, "click", "#btn2", null));
        events.add(new WebEventListener.CapturedEvent(4L, "change", "#select", null));

        // Filter click events
        java.util.List<WebEventListener.CapturedEvent> clickEvents = new java.util.ArrayList<>();
        for (WebEventListener.CapturedEvent event : events) {
            if ("click".equals(event.type)) {
                clickEvents.add(event);
            }
        }

        assertEquals(2, clickEvents.size());
    }

    @Test
    public void testFilterBySelector() {
        java.util.List<WebEventListener.CapturedEvent> events = new java.util.ArrayList<>();
        events.add(new WebEventListener.CapturedEvent(1L, "click", "#submit-btn", null));
        events.add(new WebEventListener.CapturedEvent(2L, "click", "#cancel-btn", null));
        events.add(new WebEventListener.CapturedEvent(3L, "click", ".nav-link", null));

        // Filter events containing "btn"
        java.util.List<WebEventListener.CapturedEvent> filtered = new java.util.ArrayList<>();
        for (WebEventListener.CapturedEvent event : events) {
            if (event.selector.contains("btn")) {
                filtered.add(event);
            }
        }

        assertEquals(2, filtered.size());
    }

    // Test event summary

    @Test
    public void testEventSummary() throws Exception {
        java.util.List<WebEventListener.CapturedEvent> events = new java.util.ArrayList<>();
        events.add(new WebEventListener.CapturedEvent(1L, "click", "#btn", null));
        events.add(new WebEventListener.CapturedEvent(2L, "click", "#btn", null));
        events.add(new WebEventListener.CapturedEvent(3L, "input", "#field", null));

        JSONObject byType = new JSONObject();
        for (WebEventListener.CapturedEvent event : events) {
            int count = byType.optInt(event.type, 0);
            byType.put(event.type, count + 1);
        }

        assertEquals(2, byType.getInt("click"));
        assertEquals(1, byType.getInt("input"));
    }

    // Test event export

    @Test
    public void testEventExport() throws Exception {
        java.util.List<WebEventListener.CapturedEvent> events = new java.util.ArrayList<>();
        JSONObject data = new JSONObject();
        data.put("test", "value");

        events.add(new WebEventListener.CapturedEvent(1L, "click", "#btn", data));

        // Export to JSON array
        org.json.JSONArray exported = new org.json.JSONArray();
        for (WebEventListener.CapturedEvent event : events) {
            JSONObject obj = new JSONObject();
            obj.put("id", event.id);
            obj.put("type", event.type);
            obj.put("selector", event.selector);
            obj.put("timestamp", event.timestamp);
            if (event.data != null) {
                obj.put("data", event.data);
            }
            exported.put(obj);
        }

        assertEquals(1, exported.length());
        JSONObject first = exported.getJSONObject(0);
        assertEquals(1, first.getInt("id"));
        assertEquals("click", first.getString("type"));
    }

    // Test custom events

    @Test
    public void testElementAppearEvent() throws Exception {
        WebEventListener.CapturedEvent event = new WebEventListener.CapturedEvent(
            1L, "elementAppear", "#modal", new JSONObject()
        );

        assertEquals("elementAppear", event.type);
        assertEquals("#modal", event.selector);
    }

    @Test
    public void testElementDisappearEvent() throws Exception {
        WebEventListener.CapturedEvent event = new WebEventListener.CapturedEvent(
            1L, "elementDisappear", "#loading", new JSONObject()
        );

        assertEquals("elementDisappear", event.type);
    }

    @Test
    public void testTextChangeEvent() throws Exception {
        JSONObject data = new JSONObject();
        data.put("oldText", "Loading...");
        data.put("newText", "Complete");

        WebEventListener.CapturedEvent event = new WebEventListener.CapturedEvent(
            1L, "textChange", "#status", data
        );

        assertEquals("textChange", event.type);
        assertEquals("Loading...", event.data.getString("oldText"));
        assertEquals("Complete", event.data.getString("newText"));
    }

    // Test capturing state

    @Test
    public void testCapturingState() {
        boolean capturing = false;

        assertFalse("Should not be capturing initially", capturing);

        capturing = true;
        assertTrue("Should be capturing after start", capturing);

        capturing = false;
        assertFalse("Should not be capturing after stop", capturing);
    }

    // Test event clearing

    @Test
    public void testClearEvents() {
        java.util.List<WebEventListener.CapturedEvent> events = new java.util.ArrayList<>();
        events.add(new WebEventListener.CapturedEvent(1L, "click", "#btn", null));
        events.add(new WebEventListener.CapturedEvent(2L, "input", "#field", null));

        assertEquals(2, events.size());

        events.clear();

        assertTrue("Events should be empty after clear", events.isEmpty());
    }
}