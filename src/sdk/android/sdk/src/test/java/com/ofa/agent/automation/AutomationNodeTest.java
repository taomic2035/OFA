package com.ofa.agent.automation;

import org.json.JSONObject;
import org.junit.Test;
import static org.junit.Assert.*;

/**
 * Unit tests for AutomationNode
 */
public class AutomationNodeTest {

    @Test
    public void testBasicNode() {
        AutomationNode node = new AutomationNode(
                "android.widget.Button",
                "Click Me",
                null,
                "com.example:id/btn",
                0,
                true,
                true,
                false,
                false,
                false,
                false,
                true,
                100, 200, 300, 250,
                "com.example"
        );

        assertEquals("android.widget.Button", node.getClassName());
        assertEquals("Click Me", node.getText());
        assertEquals("com.example:id/btn", node.getResourceId());
        assertTrue(node.isClickable());
        assertTrue(node.isEnabled());
    }

    @Test
    public void testCenterCoordinates() {
        AutomationNode node = new AutomationNode(
                null, null, null, null, 0,
                false, true, false, false, false, false, true,
                100, 200, 300, 400,
                null
        );

        assertEquals(200, node.getCenterX());  // (100 + 300) / 2
        assertEquals(300, node.getCenterY());  // (200 + 400) / 2
    }

    @Test
    public void testDimensions() {
        AutomationNode node = new AutomationNode(
                null, null, null, null, 0,
                false, true, false, false, false, false, true,
                0, 0, 100, 200,
                null
        );

        assertEquals(100, node.getWidth());
        assertEquals(200, node.getHeight());
    }

    @Test
    public void testMatchesSelector() {
        AutomationNode node = new AutomationNode(
                "android.widget.Button",
                "Submit",
                null,
                "com.example:id/submit",
                0,
                true,
                true,
                false,
                false,
                false,
                false,
                true,
                0, 0, 100, 50,
                "com.example"
        );

        // Test text match
        BySelector textSelector = BySelector.text("Submit");
        assertTrue(node.matches(textSelector));

        // Test text contains
        BySelector containsSelector = BySelector.textContains("ub");
        assertTrue(node.matches(containsSelector));

        // Test ID match
        BySelector idSelector = BySelector.id("com.example:id/submit");
        assertTrue(node.matches(idSelector));

        // Test class match
        BySelector classSelector = BySelector.className("android.widget.Button");
        assertTrue(node.matches(classSelector));

        // Test clickable match
        BySelector clickableSelector = BySelector.clickable();
        assertTrue(node.matches(clickableSelector));

        // Test non-matching selector
        BySelector wrongTextSelector = BySelector.text("Cancel");
        assertFalse(node.matches(wrongTextSelector));
    }

    @Test
    public void testToJson() {
        AutomationNode node = new AutomationNode(
                "android.widget.Button",
                "Click",
                null,
                "id/btn",
                0,
                true,
                true,
                false,
                false,
                false,
                false,
                true,
                0, 0, 100, 50,
                "com.example"
        );

        JSONObject json = node.toJson();
        assertEquals("android.widget.Button", json.optString("className"));
        assertEquals("Click", json.optString("text"));
        assertEquals("id/btn", json.optString("resourceId"));
        assertTrue(json.optBoolean("clickable"));
        assertTrue(json.optBoolean("enabled"));
    }
}