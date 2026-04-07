package com.ofa.agent.state;

import static org.junit.Assert.*;

import org.json.JSONException;
import org.json.JSONObject;
import org.junit.Test;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class DeviceStateTest {

    @Test
    public void testCreateDeviceState() {
        DeviceState state = DeviceState.create("agent-001", "identity-001", "mobile", "Test Phone");

        assertEquals("agent-001", state.getAgentId());
        assertEquals("identity-001", state.getIdentityId());
        assertEquals("mobile", state.getDeviceType());
        assertEquals("Test Phone", state.getDeviceName());
        assertTrue(state.isOnline());
    }

    @Test
    public void testBatteryLevelBounds() {
        DeviceState state = new DeviceState();

        state.setBatteryLevel(150);
        assertEquals(100, state.getBatteryLevel());

        state.setBatteryLevel(-10);
        assertEquals(0, state.getBatteryLevel());

        state.setBatteryLevel(50);
        assertEquals(50, state.getBatteryLevel());
    }

    @Test
    public void testNetworkStrengthBounds() {
        DeviceState state = new DeviceState();

        state.setNetworkStrength(150);
        assertEquals(100, state.getNetworkStrength());

        state.setNetworkStrength(-10);
        assertEquals(0, state.getNetworkStrength());

        state.setNetworkStrength(75);
        assertEquals(75, state.getNetworkStrength());
    }

    @Test
    public void testPriorityBounds() {
        DeviceState state = new DeviceState();

        state.setPriority(150);
        assertEquals(100, state.getPriority());

        state.setPriority(-10);
        assertEquals(0, state.getPriority());

        state.setPriority(75);
        assertEquals(75, state.getPriority());
    }

    @Test
    public void testJsonSerialization() throws JSONException {
        DeviceState state = DeviceState.create("agent-001", "identity-001", "mobile", "Test Phone");
        state.setBatteryLevel(85);
        state.setCharging(true);
        state.setNetworkType("wifi");
        state.setNetworkStrength(90);
        state.setScene("running");
        state.setPriority(100);
        state.setTrustLevel("primary");

        // 添加能力
        state.addCapability("ui_automation");
        state.addCapability("speech");

        // 添加场景上下文
        state.addSceneContext("speed", 5.5);

        JSONObject json = state.toJson();

        assertEquals("agent-001", json.getString("agent_id"));
        assertEquals("identity-001", json.getString("identity_id"));
        assertEquals(85, json.getInt("battery_level"));
        assertTrue(json.getBoolean("charging"));
        assertEquals("wifi", json.getString("network_type"));
        assertEquals("running", json.getString("scene"));
    }

    @Test
    public void testJsonDeserialization() throws JSONException {
        JSONObject json = new JSONObject();
        json.put("agent_id", "agent-002");
        json.put("identity_id", "identity-002");
        json.put("device_type", "watch");
        json.put("device_name", "Test Watch");
        json.put("online", true);
        json.put("battery_level", 75);
        json.put("charging", false);
        json.put("network_type", "cellular");
        json.put("scene", "idle");

        DeviceState state = DeviceState.fromJson(json);

        assertEquals("agent-002", state.getAgentId());
        assertEquals("identity-002", state.getIdentityId());
        assertEquals("watch", state.getDeviceType());
        assertEquals("Test Watch", state.getDeviceName());
        assertTrue(state.isOnline());
        assertEquals(75, state.getBatteryLevel());
        assertFalse(state.isCharging());
        assertEquals("cellular", state.getNetworkType());
        assertEquals("idle", state.getScene());
    }

    @Test
    public void testDetectChangeType() {
        DeviceState oldState = DeviceState.create("agent-001", "identity-001", "mobile", "Test");
        oldState.setBatteryLevel(85);
        oldState.setOnline(true);
        oldState.setScene("idle");
        oldState.setNetworkType("wifi");

        DeviceState newState = DeviceState.create("agent-001", "identity-001", "mobile", "Test");
        newState.setBatteryLevel(75);
        newState.setOnline(true);
        newState.setScene("idle");
        newState.setNetworkType("wifi");

        // 电池变更
        assertEquals(DeviceState.CHANGE_BATTERY, newState.detectChangeType(oldState));

        // 离线变更
        newState.setOnline(false);
        assertEquals(DeviceState.CHANGE_OFFLINE, newState.detectChangeType(oldState));

        // 场景变更
        newState.setOnline(true);
        newState.setScene("running");
        assertEquals(DeviceState.CHANGE_SCENE, newState.detectChangeType(oldState));

        // 网络变更
        newState.setScene("idle");
        newState.setNetworkType("cellular");
        assertEquals(DeviceState.CHANGE_NETWORK, newState.detectChangeType(oldState));

        // 新状态（无旧状态）
        assertEquals(DeviceState.CHANGE_FULL, newState.detectChangeType(null));
    }

    @Test
    public void testIsLowBattery() {
        DeviceState state = new DeviceState();

        state.setBatteryLevel(15);
        state.setCharging(false);
        assertTrue(state.isLowBattery());

        state.setBatteryLevel(15);
        state.setCharging(true);
        assertFalse(state.isLowBattery()); // 充电中不算低电量

        state.setBatteryLevel(25);
        state.setCharging(false);
        assertFalse(state.isLowBattery());
    }

    @Test
    public void testHasNetwork() {
        DeviceState state = new DeviceState();

        state.setNetworkType("wifi");
        state.setNetworkStrength(80);
        assertTrue(state.hasNetwork());

        state.setNetworkType("cellular");
        state.setNetworkStrength(50);
        assertTrue(state.hasNetwork());

        state.setNetworkType("offline");
        state.setNetworkStrength(0);
        assertFalse(state.hasNetwork());

        state.setNetworkType("wifi");
        state.setNetworkStrength(0);
        assertFalse(state.hasNetwork());
    }

    @Test
    public void testHasWiFiAndCellular() {
        DeviceState state = new DeviceState();

        state.setNetworkType("wifi");
        assertTrue(state.hasWiFi());
        assertFalse(state.hasCellular());

        state.setNetworkType("cellular");
        assertFalse(state.hasWiFi());
        assertTrue(state.hasCellular());

        state.setNetworkType("offline");
        assertFalse(state.hasWiFi());
        assertFalse(state.hasCellular());
    }

    @Test
    public void testIsInScene() {
        DeviceState state = new DeviceState();
        state.setScene("running");

        assertTrue(state.isInScene("running"));
        assertFalse(state.isInScene("idle"));
    }

    @Test
    public void testHasCapability() {
        DeviceState state = new DeviceState();
        state.addCapability("ui_automation");
        state.addCapability("speech");

        assertTrue(state.hasCapability("ui_automation"));
        assertTrue(state.hasCapability("speech"));
        assertFalse(state.hasCapability("camera"));
    }

    @Test
    public void testAddRemoveCapability() {
        DeviceState state = new DeviceState();

        state.addCapability("feature1");
        state.addCapability("feature2");
        assertEquals(2, state.getCapabilities().size());

        // 重复添加不增加
        state.addCapability("feature1");
        assertEquals(2, state.getCapabilities().size());

        state.removeCapability("feature1");
        assertEquals(1, state.getCapabilities().size());
        assertFalse(state.hasCapability("feature1"));
    }

    @Test
    public void testLocationWithJson() throws JSONException {
        DeviceState state = DeviceState.create("agent-001", "identity-001", "mobile", "Test");
        DeviceLocation location = new DeviceLocation(39.9042, 116.4074, 10.0, "home");
        state.setLocation(location);

        JSONObject json = state.toJson();
        assertTrue(json.has("location"));

        JSONObject locJson = json.getJSONObject("location");
        assertEquals(39.9042, locJson.getDouble("latitude"), 0.0001);
        assertEquals(116.4074, locJson.getDouble("longitude"), 0.0001);
        assertEquals("home", locJson.getString("location_type"));

        // 反序列化
        DeviceState parsed = DeviceState.fromJson(json);
        assertNotNull(parsed.getLocation());
        assertEquals(39.9042, parsed.getLocation().getLatitude(), 0.0001);
        assertEquals("home", parsed.getLocation().getLocationType());
    }

    @Test
    public void testCapabilitiesWithJson() throws JSONException {
        DeviceState state = DeviceState.create("agent-001", "identity-001", "mobile", "Test");
        state.addCapability("ui_automation");
        state.addCapability("speech");
        state.addCapability("sensor");

        JSONObject json = state.toJson();
        assertTrue(json.has("capabilities"));

        // 反序列化
        DeviceState parsed = DeviceState.fromJson(json);
        assertEquals(3, parsed.getCapabilities().size());
        assertTrue(parsed.hasCapability("ui_automation"));
        assertTrue(parsed.hasCapability("speech"));
        assertTrue(parsed.hasCapability("sensor"));
    }
}