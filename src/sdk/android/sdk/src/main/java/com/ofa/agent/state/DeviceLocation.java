package com.ofa.agent.state;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

/**
 * 设备位置模型 (v3.1.0)
 */
public class DeviceLocation {

    private double latitude;
    private double longitude;
    private double accuracy;
    private String locationType;  // home, work, outdoor, unknown

    // 位置类型常量
    public static final String TYPE_HOME = "home";
    public static final String TYPE_WORK = "work";
    public static final String TYPE_OUTDOOR = "outdoor";
    public static final String TYPE_UNKNOWN = "unknown";

    public DeviceLocation() {
        this.locationType = TYPE_UNKNOWN;
    }

    public DeviceLocation(double latitude, double longitude) {
        this.latitude = latitude;
        this.longitude = longitude;
        this.accuracy = 0;
        this.locationType = TYPE_UNKNOWN;
    }

    public DeviceLocation(double latitude, double longitude, double accuracy, String locationType) {
        this.latitude = latitude;
        this.longitude = longitude;
        this.accuracy = accuracy;
        this.locationType = locationType;
    }

    /**
     * 从 JSON 解析
     */
    @NonNull
    public static DeviceLocation fromJson(@NonNull JSONObject json) throws JSONException {
        DeviceLocation location = new DeviceLocation();
        location.latitude = json.optDouble("latitude", 0);
        location.longitude = json.optDouble("longitude", 0);
        location.accuracy = json.optDouble("accuracy", 0);
        location.locationType = json.optString("location_type", TYPE_UNKNOWN);
        return location;
    }

    /**
     * 转换为 JSON
     */
    @NonNull
    public JSONObject toJson() throws JSONException {
        JSONObject json = new JSONObject();
        json.put("latitude", latitude);
        json.put("longitude", longitude);
        json.put("accuracy", accuracy);
        json.put("location_type", locationType);
        return json;
    }

    /**
     * 计算与另一位置的距离（米）
     * 使用 Haversine 公式
     */
    public double distanceTo(@Nullable DeviceLocation other) {
        if (other == null) {
            return -1;
        }

        double lat1 = this.latitude;
        double lon1 = this.longitude;
        double lat2 = other.latitude;
        double lon2 = other.longitude;

        double earthRadius = 6371000; // 地球半径（米）

        double dLat = Math.toRadians(lat2 - lat1);
        double dLon = Math.toRadians(lon2 - lon1);

        double a = Math.sin(dLat / 2) * Math.sin(dLat / 2) +
                Math.cos(Math.toRadians(lat1)) * Math.cos(Math.toRadians(lat2)) *
                        Math.sin(dLon / 2) * Math.sin(dLon / 2);

        double c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));

        return earthRadius * c;
    }

    /**
     * 检查是否在家
     */
    public boolean isAtHome() {
        return TYPE_HOME.equals(locationType);
    }

    /**
     * 检查是否在工作
     */
    public boolean isAtWork() {
        return TYPE_WORK.equals(locationType);
    }

    /**
     * 检查是否在户外
     */
    public boolean isOutdoor() {
        return TYPE_OUTDOOR.equals(locationType);
    }

    // Getter/Setter

    public double getLatitude() { return latitude; }
    public void setLatitude(double latitude) { this.latitude = latitude; }

    public double getLongitude() { return longitude; }
    public void setLongitude(double longitude) { this.longitude = longitude; }

    public double getAccuracy() { return accuracy; }
    public void setAccuracy(double accuracy) { this.accuracy = accuracy; }

    public String getLocationType() { return locationType; }
    public void setLocationType(String locationType) { this.locationType = locationType; }

    @NonNull
    @Override
    public String toString() {
        return "DeviceLocation{" +
                "lat=" + latitude +
                ", lon=" + longitude +
                ", accuracy=" + accuracy +
                ", type='" + locationType + '\'' +
                '}';
    }
}