package com.ofa.agent.automation;

import org.junit.Test;
import static org.junit.Assert.*;

/**
 * Unit tests for AutomationResult
 */
public class AutomationResultTest {

    @Test
    public void testSuccessResult() {
        JSONObject data = new JSONObject();
        try {
            data.put("x", 100);
            data.put("y", 200);
        } catch (Exception e) {}

        AutomationResult result = new AutomationResult("click", data, 50);

        assertTrue(result.isSuccess());
        assertEquals("click", result.getOperation());
        assertEquals(50, result.getExecutionTimeMs());
        assertNotNull(result.getData());
        assertNull(result.getError());
    }

    @Test
    public void testErrorResult() {
        AutomationResult result = new AutomationResult("click", "Element not found");

        assertFalse(result.isSuccess());
        assertEquals("click", result.getOperation());
        assertEquals("Element not found", result.getError());
        assertNull(result.getData());
    }

    @Test
    public void testGetDataString() {
        JSONObject data = new JSONObject();
        try {
            data.put("text", "hello");
        } catch (Exception e) {}

        AutomationResult result = new AutomationResult("input", data, 100);

        assertEquals("hello", result.getDataString("text"));
        assertNull(result.getDataString("missing"));
    }

    @Test
    public void testToJson() {
        AutomationResult result = new AutomationResult("click", "Failed", 0);
        JSONObject json = result.toJson();

        assertFalse(json.optBoolean("success"));
        assertEquals("click", json.optString("operation"));
        assertEquals("Failed", json.optString("error"));
    }
}