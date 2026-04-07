package com.ofa.agent.state;

import static org.junit.Assert.*;

import org.json.JSONException;
import org.json.JSONObject;
import org.junit.Test;

import java.util.Map;

public class StateChangeTest {

    @Test
    public void testCreateStateChange() {
        DeviceState newState = DeviceState.create("agent-001", "identity-001", "mobile", "Test");
        newState.setBatteryLevel(75);

        StateChange change = StateChange.create(
                "agent-001",
                "identity-001",
                DeviceState.CHANGE_BATTERY,
                null,
                newState
        );

        assertEquals("agent-001", change.getAgentId());
        assertEquals("identity-001", change.getIdentityId());
        assertEquals(DeviceState.CHANGE_BATTERY, change.getChangeType());
        assertNull(change.getOldState());
        assertNotNull(change.getNewState());
        assertTrue(change.getTimestamp() > 0);
    }

    @Test
    public void testJsonSerialization() throws JSONException {
        DeviceState oldState = DeviceState.create("agent-001", "identity-001", "mobile", "Test");
        oldState.setBatteryLevel(85);

        DeviceState newState = DeviceState.create("agent-001", "identity-001", "mobile", "Test");
        newState.setBatteryLevel(75);

        StateChange change = StateChange.create(
                "agent-001",
                "identity-001",
                DeviceState.CHANGE_BATTERY,
                oldState,
                newState
        );
        change.setVersion(12345L);

        JSONObject json = change.toJson();

        assertEquals("agent-001", json.getString("agent_id"));
        assertEquals("identity-001", json.getString("identity_id"));
        assertEquals(DeviceState.CHANGE_BATTERY, json.getString("change_type"));
        assertEquals(12345L, json.getLong("version"));
        assertTrue(json.has("old_state"));
        assertTrue(json.has("new_state"));
    }

    @Test
    public void testJsonDeserialization() throws JSONException {
        JSONObject json = new JSONObject();
        json.put("agent_id", "agent-002");
        json.put("identity_id", "identity-002");
        json.put("change_type", DeviceState.CHANGE_ONLINE);
        json.put("timestamp", System.currentTimeMillis());
        json.put("version", 100L);

        JSONObject newStateJson = new JSONObject();
        newStateJson.put("agent_id", "agent-002");
        newStateJson.put("identity_id", "identity-002");
        newStateJson.put("device_type", "watch");
        newStateJson.put("device_name", "Test Watch");
        newStateJson.put("online", true);
        json.put("new_state", newStateJson);

        StateChange change = StateChange.fromJson(json);

        assertEquals("agent-002", change.getAgentId());
        assertEquals("identity-002", change.getIdentityId());
        assertEquals(DeviceState.CHANGE_ONLINE, change.getChangeType());
        assertEquals(100L, change.getVersion());
        assertNotNull(change.getNewState());
        assertNull(change.getOldState());
    }

    @Test
    public void testToMap() throws JSONException {
        DeviceState newState = DeviceState.create("agent-001", "identity-001", "mobile", "Test");
        newState.setBatteryLevel(75);

        StateChange change = StateChange.create(
                "agent-001",
                "identity-001",
                DeviceState.CHANGE_BATTERY,
                null,
                newState
        );

        Map<String, Object> map = change.toMap();

        assertEquals("agent-001", map.get("agent_id"));
        assertEquals("identity-001", map.get("identity_id"));
        assertEquals(DeviceState.CHANGE_BATTERY, map.get("change_type"));
        assertNotNull(map.get("new_state"));
    }

    @Test
    public void testIsOnlineChange() {
        StateChange change = new StateChange();
        change.setChangeType(DeviceState.CHANGE_ONLINE);

        assertTrue(change.isOnlineChange());
        assertFalse(change.isOfflineChange());
        assertFalse(change.isBatteryChange());
    }

    @Test
    public void testIsOfflineChange() {
        StateChange change = new StateChange();
        change.setChangeType(DeviceState.CHANGE_OFFLINE);

        assertFalse(change.isOnlineChange());
        assertTrue(change.isOfflineChange());
        assertFalse(change.isBatteryChange());
    }

    @Test
    public void testIsBatteryChange() {
        StateChange change = new StateChange();
        change.setChangeType(DeviceState.CHANGE_BATTERY);

        assertFalse(change.isOnlineChange());
        assertFalse(change.isOfflineChange());
        assertTrue(change.isBatteryChange());
    }

    @Test
    public void testIsNetworkChange() {
        StateChange change = new StateChange();
        change.setChangeType(DeviceState.CHANGE_NETWORK);

        assertTrue(change.isNetworkChange());
        assertFalse(change.isSceneChange());
    }

    @Test
    public void testIsSceneChange() {
        StateChange change = new StateChange();
        change.setChangeType(DeviceState.CHANGE_SCENE);

        assertFalse(change.isNetworkChange());
        assertTrue(change.isSceneChange());
    }

    @Test
    public void testSettersAndGetters() {
        StateChange change = new StateChange();

        change.setAgentId("test-agent");
        change.setIdentityId("test-identity");
        change.setChangeType(DeviceState.CHANGE_LOCATION);
        change.setTimestamp(123456789L);
        change.setVersion(999L);

        DeviceState state = new DeviceState();
        change.setNewState(state);

        assertEquals("test-agent", change.getAgentId());
        assertEquals("test-identity", change.getIdentityId());
        assertEquals(DeviceState.CHANGE_LOCATION, change.getChangeType());
        assertEquals(123456789L, change.getTimestamp());
        assertEquals(999L, change.getVersion());
        assertNotNull(change.getNewState());
    }
}