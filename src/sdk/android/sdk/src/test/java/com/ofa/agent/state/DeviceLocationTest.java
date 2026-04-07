package com.ofa.agent.state;

import static org.junit.Assert.*;

import org.json.JSONException;
import org.json.JSONObject;
import org.junit.Test;

public class DeviceLocationTest {

    @Test
    public void testCreateLocation() {
        DeviceLocation location = new DeviceLocation(39.9042, 116.4074);

        assertEquals(39.9042, location.getLatitude(), 0.0001);
        assertEquals(116.4074, location.getLongitude(), 0.0001);
        assertEquals(DeviceLocation.TYPE_UNKNOWN, location.getLocationType());
    }

    @Test
    public void testCreateLocationWithType() {
        DeviceLocation location = new DeviceLocation(39.9042, 116.4074, 10.0, "home");

        assertEquals(39.9042, location.getLatitude(), 0.0001);
        assertEquals(116.4074, location.getLongitude(), 0.0001);
        assertEquals(10.0, location.getAccuracy(), 0.0001);
        assertEquals("home", location.getLocationType());
    }

    @Test
    public void testJsonSerialization() throws JSONException {
        DeviceLocation location = new DeviceLocation(31.2304, 121.4737, 15.0, "work");

        JSONObject json = location.toJson();

        assertEquals(31.2304, json.getDouble("latitude"), 0.0001);
        assertEquals(121.4737, json.getDouble("longitude"), 0.0001);
        assertEquals(15.0, json.getDouble("accuracy"), 0.0001);
        assertEquals("work", json.getString("location_type"));
    }

    @Test
    public void testJsonDeserialization() throws JSONException {
        JSONObject json = new JSONObject();
        json.put("latitude", 22.5431);
        json.put("longitude", 114.0579);
        json.put("accuracy", 20.0);
        json.put("location_type", "outdoor");

        DeviceLocation location = DeviceLocation.fromJson(json);

        assertEquals(22.5431, location.getLatitude(), 0.0001);
        assertEquals(114.0579, location.getLongitude(), 0.0001);
        assertEquals(20.0, location.getAccuracy(), 0.0001);
        assertEquals("outdoor", location.getLocationType());
    }

    @Test
    public void testDistanceCalculation() {
        // 北京到上海的距离约 1000+ 公里
        DeviceLocation beijing = new DeviceLocation(39.9042, 116.4074);
        DeviceLocation shanghai = new DeviceLocation(31.2304, 121.4737);

        double distance = beijing.distanceTo(shanghai);

        // 实际距离约 1068 公里
        assertTrue("Distance should be around 1068km", distance > 1000000 && distance < 1100000);
    }

    @Test
    public void testDistanceToNull() {
        DeviceLocation location = new DeviceLocation(39.9042, 116.4074);

        assertEquals(-1, location.distanceTo(null), 0.001);
    }

    @Test
    public void testDistanceToSameLocation() {
        DeviceLocation loc1 = new DeviceLocation(39.9042, 116.4074);
        DeviceLocation loc2 = new DeviceLocation(39.9042, 116.4074);

        assertEquals(0, loc1.distanceTo(loc2), 1.0); // 允许 1 米误差
    }

    @Test
    public void testIsAtHome() {
        DeviceLocation location = new DeviceLocation();

        location.setLocationType("home");
        assertTrue(location.isAtHome());
        assertFalse(location.isAtWork());
        assertFalse(location.isOutdoor());
    }

    @Test
    public void testIsAtWork() {
        DeviceLocation location = new DeviceLocation();

        location.setLocationType("work");
        assertFalse(location.isAtHome());
        assertTrue(location.isAtWork());
        assertFalse(location.isOutdoor());
    }

    @Test
    public void testIsOutdoor() {
        DeviceLocation location = new DeviceLocation();

        location.setLocationType("outdoor");
        assertFalse(location.isAtHome());
        assertFalse(location.isAtWork());
        assertTrue(location.isOutdoor());
    }

    @Test
    public void testLocationTypeConstants() {
        assertEquals("home", DeviceLocation.TYPE_HOME);
        assertEquals("work", DeviceLocation.TYPE_WORK);
        assertEquals("outdoor", DeviceLocation.TYPE_OUTDOOR);
        assertEquals("unknown", DeviceLocation.TYPE_UNKNOWN);
    }
}