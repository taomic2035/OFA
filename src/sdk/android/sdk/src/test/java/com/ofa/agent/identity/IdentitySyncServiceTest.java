package com.ofa.agent.identity;

import static org.junit.Assert.*;

import org.junit.Before;
import org.junit.Test;

import java.util.HashMap;
import java.util.Map;

/**
 * IdentitySyncService 单元测试
 */
public class IdentitySyncServiceTest {

    @Test
    public void testSyncResultSuccess() {
        PersonalIdentity identity = new PersonalIdentity();
        identity.setName("Test User");

        IdentitySyncService.SyncResult result = IdentitySyncService.SyncResult.success(identity);

        assertTrue(result.success);
        assertFalse(result.conflict);
        assertNotNull(result.identity);
        assertNull(result.error);
    }

    @Test
    public void testSyncResultConflict() {
        PersonalIdentity identity = new PersonalIdentity();
        identity.setName("Test User");

        IdentitySyncService.SyncResult result = IdentitySyncService.SyncResult.conflict(identity);

        assertTrue(result.success);
        assertTrue(result.conflict);
        assertNotNull(result.identity);
    }

    @Test
    public void testSyncResultFailure() {
        IdentitySyncService.SyncResult result = IdentitySyncService.SyncResult.failure("Network error");

        assertFalse(result.success);
        assertNull(result.identity);
        assertEquals("Network error", result.error);
    }
}