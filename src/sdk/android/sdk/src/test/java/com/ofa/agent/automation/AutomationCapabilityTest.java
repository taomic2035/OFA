package com.ofa.agent.automation;

import org.junit.Test;
import static org.junit.Assert.*;

/**
 * Unit tests for AutomationCapability
 */
public class AutomationCapabilityTest {

    @Test
    public void testCapabilityLevels() {
        assertTrue(AutomationCapability.SYSTEM_LEVEL.getLevel() > AutomationCapability.FULL_ACCESSIBILITY.getLevel());
        assertTrue(AutomationCapability.FULL_ACCESSIBILITY.getLevel() > AutomationCapability.ENHANCED.getLevel());
        assertTrue(AutomationCapability.ENHANCED.getLevel() > AutomationCapability.BASIC.getLevel());
        assertTrue(AutomationCapability.BASIC.getLevel() > AutomationCapability.NONE.getLevel());
    }

    @Test
    public void testSupports() {
        assertTrue(AutomationCapability.SYSTEM_LEVEL.supports(AutomationCapability.BASIC));
        assertTrue(AutomationCapability.ENHANCED.supports(AutomationCapability.BASIC));
        assertFalse(AutomationCapability.BASIC.supports(AutomationCapability.ENHANCED));
        assertFalse(AutomationCapability.NONE.supports(AutomationCapability.BASIC));
    }

    @Test
    public void testSupportsBasicOperations() {
        assertTrue(AutomationCapability.BASIC.supportsBasicOperations());
        assertTrue(AutomationCapability.ENHANCED.supportsBasicOperations());
        assertFalse(AutomationCapability.NONE.supportsBasicOperations());
    }

    @Test
    public void testSupportsScrollOperations() {
        assertTrue(AutomationCapability.ENHANCED.supportsScrollOperations());
        assertTrue(AutomationCapability.SYSTEM_LEVEL.supportsScrollOperations());
        assertFalse(AutomationCapability.BASIC.supportsScrollOperations());
    }

    @Test
    public void testSupportsSystemOperations() {
        assertTrue(AutomationCapability.SYSTEM_LEVEL.supportsSystemOperations());
        assertFalse(AutomationCapability.FULL_ACCESSIBILITY.supportsSystemOperations());
    }

    @Test
    public void testGetDescription() {
        assertNotNull(AutomationCapability.BASIC.getDescription());
        assertEquals("Basic accessibility service", AutomationCapability.BASIC.getDescription());
    }
}