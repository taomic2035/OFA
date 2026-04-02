package com.ofa.agent.automation;

import org.junit.Test;
import static org.junit.Assert.*;

/**
 * Unit tests for BySelector
 */
public class BySelectorTest {

    @Test
    public void testTextSelector() {
        BySelector selector = BySelector.text("Hello");
        assertEquals("Hello", selector.getText());
        assertTrue(selector.hasCriteria());
    }

    @Test
    public void testTextContainsSelector() {
        BySelector selector = BySelector.textContains("ell");
        assertEquals("ell", selector.getTextContains());
        assertTrue(selector.hasCriteria());
    }

    @Test
    public void testIdSelector() {
        BySelector selector = BySelector.id("com.example:id/button");
        assertEquals("com.example:id/button", selector.getResourceId());
        assertTrue(selector.hasCriteria());
    }

    @Test
    public void testClassNameSelector() {
        BySelector selector = BySelector.className("android.widget.Button");
        assertEquals("android.widget.Button", selector.getClassName());
    }

    @Test
    public void testDescSelector() {
        BySelector selector = BySelector.desc("Submit");
        assertEquals("Submit", selector.getContentDescription());
    }

    @Test
    public void testClickableSelector() {
        BySelector selector = BySelector.clickable();
        assertTrue(selector.isClickable());
    }

    @Test
    public void testScrollableSelector() {
        BySelector selector = BySelector.scrollable();
        assertTrue(selector.isScrollable());
    }

    @Test
    public void testCombinedSelector() {
        BySelector selector = BySelector.text("Login")
                .andClassName("android.widget.Button")
                .clickable(true);

        assertEquals("Login", selector.getText());
        assertEquals("android.widget.Button", selector.getClassName());
        assertTrue(selector.isClickable());
    }

    @Test
    public void testIndexModifier() {
        BySelector selector = BySelector.text("Item").index(2);
        assertEquals(2, selector.getIndex());
    }

    @Test
    public void testEmptySelector() {
        BySelector selector = new BySelector();
        assertFalse(selector.hasCriteria());
    }

    @Test
    public void testDescribe() {
        BySelector selector = BySelector.text("Hello").clickable();
        String desc = selector.describe();
        assertTrue(desc.contains("text=Hello"));
        assertTrue(desc.contains("clickable"));
    }
}