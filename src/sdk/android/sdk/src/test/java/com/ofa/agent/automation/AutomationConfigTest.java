package com.ofa.agent.automation;

import org.junit.Test;
import static org.junit.Assert.*;

/**
 * Unit tests for AutomationConfig
 */
public class AutomationConfigTest {

    @Test
    public void testDefaultConfig() {
        AutomationConfig config = new AutomationConfig();

        assertEquals(AutomationConfig.DEFAULT_CLICK_TIMEOUT, config.getClickTimeout());
        assertEquals(AutomationConfig.DEFAULT_SWIPE_TIMEOUT, config.getSwipeTimeout());
        assertEquals(AutomationConfig.DEFAULT_WAIT_TIMEOUT, config.getWaitTimeout());
    }

    @Test
    public void testBuilderPattern() {
        AutomationConfig config = AutomationConfig.builder()
                .clickTimeout(10000)
                .swipeTimeout(5000)
                .autoRetry(true)
                .maxRetries(5)
                .enableLogging(false)
                .build();

        assertEquals(10000, config.getClickTimeout());
        assertEquals(5000, config.getSwipeTimeout());
        assertTrue(config.isAutoRetryEnabled());
        assertEquals(5, config.getMaxRetries());
        assertFalse(config.isLoggingEnabled());
    }

    @Test
    public void testGestureSettings() {
        AutomationConfig config = AutomationConfig.builder()
                .clickDuration(200)
                .longClickDuration(1000)
                .swipeDuration(500)
                .scrollDistance(800)
                .build();

        assertEquals(200, config.getClickDuration());
        assertEquals(1000, config.getLongClickDuration());
        assertEquals(500, config.getSwipeDuration());
        assertEquals(800, config.getScrollDistance());
    }
}