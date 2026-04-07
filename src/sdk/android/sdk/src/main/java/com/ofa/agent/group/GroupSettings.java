package com.ofa.agent.group;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

/**
 * 群组设置模型 (v3.5.0)
 */
public class GroupSettings {

    private int maxMembers;
    private boolean allowGuests;
    private boolean broadcastEnabled;
    private boolean syncEnabled;
    private int quietHoursStart;
    private int quietHoursEnd;
    private int defaultPriority;
    private String autoActivateScene;

    public GroupSettings() {
        // 默认值
        this.maxMembers = 10;
        this.allowGuests = true;
        this.broadcastEnabled = true;
        this.syncEnabled = true;
        this.quietHoursStart = 22;
        this.quietHoursEnd = 7;
        this.defaultPriority = 50;
    }

    /**
     * 从 JSON 解析
     */
    @NonNull
    public static GroupSettings fromJson(@NonNull JSONObject json) throws JSONException {
        GroupSettings settings = new GroupSettings();
        settings.maxMembers = json.optInt("max_members", 10);
        settings.allowGuests = json.optBoolean("allow_guests", true);
        settings.broadcastEnabled = json.optBoolean("broadcast_enabled", true);
        settings.syncEnabled = json.optBoolean("sync_enabled", true);
        settings.quietHoursStart = json.optInt("quiet_hours_start", 22);
        settings.quietHoursEnd = json.optInt("quiet_hours_end", 7);
        settings.defaultPriority = json.optInt("default_priority", 50);
        settings.autoActivateScene = json.optString("auto_activate_scene", null);
        return settings;
    }

    /**
     * 转换为 JSON
     */
    @NonNull
    public JSONObject toJson() throws JSONException {
        JSONObject json = new JSONObject();
        json.put("max_members", maxMembers);
        json.put("allow_guests", allowGuests);
        json.put("broadcast_enabled", broadcastEnabled);
        json.put("sync_enabled", syncEnabled);
        json.put("quiet_hours_start", quietHoursStart);
        json.put("quiet_hours_end", quietHoursEnd);
        json.put("default_priority", defaultPriority);

        if (autoActivateScene != null) {
            json.put("auto_activate_scene", autoActivateScene);
        }

        return json;
    }

    /**
     * 检查当前是否在勿扰时段
     */
    public boolean isQuietHours() {
        java.util.Calendar cal = java.util.Calendar.getInstance();
        int hour = cal.get(java.util.Calendar.HOUR_OF_DAY);

        if (quietHoursStart > quietHoursEnd) {
            // 跨午夜: 22:00 - 07:00
            return hour >= quietHoursStart || hour < quietHoursEnd;
        }

        return hour >= quietHoursStart && hour < quietHoursEnd;
    }

    /**
     * 创建 Builder
     */
    @NonNull
    public static Builder builder() {
        return new Builder();
    }

    // === Getter/Setter ===

    public int getMaxMembers() { return maxMembers; }
    public void setMaxMembers(int maxMembers) { this.maxMembers = maxMembers; }

    public boolean isAllowGuests() { return allowGuests; }
    public void setAllowGuests(boolean allowGuests) { this.allowGuests = allowGuests; }

    public boolean isBroadcastEnabled() { return broadcastEnabled; }
    public void setBroadcastEnabled(boolean broadcastEnabled) { this.broadcastEnabled = broadcastEnabled; }

    public boolean isSyncEnabled() { return syncEnabled; }
    public void setSyncEnabled(boolean syncEnabled) { this.syncEnabled = syncEnabled; }

    public int getQuietHoursStart() { return quietHoursStart; }
    public void setQuietHoursStart(int quietHoursStart) { this.quietHoursStart = quietHoursStart; }

    public int getQuietHoursEnd() { return quietHoursEnd; }
    public void setQuietHoursEnd(int quietHoursEnd) { this.quietHoursEnd = quietHoursEnd; }

    public int getDefaultPriority() { return defaultPriority; }
    public void setDefaultPriority(int defaultPriority) { this.defaultPriority = defaultPriority; }

    @Nullable
    public String getAutoActivateScene() { return autoActivateScene; }
    public void setAutoActivateScene(@Nullable String autoActivateScene) { this.autoActivateScene = autoActivateScene; }

    @NonNull
    @Override
    public String toString() {
        return "GroupSettings{" +
                "maxMembers=" + maxMembers +
                ", broadcastEnabled=" + broadcastEnabled +
                ", quietHours=" + quietHoursStart + "-" + quietHoursEnd +
                '}';
    }

    /**
     * Builder 模式
     */
    public static class Builder {
        private final GroupSettings settings = new GroupSettings();

        public Builder maxMembers(int maxMembers) {
            settings.maxMembers = maxMembers;
            return this;
        }

        public Builder allowGuests(boolean allow) {
            settings.allowGuests = allow;
            return this;
        }

        public Builder broadcastEnabled(boolean enabled) {
            settings.broadcastEnabled = enabled;
            return this;
        }

        public Builder syncEnabled(boolean enabled) {
            settings.syncEnabled = enabled;
            return this;
        }

        public Builder quietHours(int start, int end) {
            settings.quietHoursStart = start;
            settings.quietHoursEnd = end;
            return this;
        }

        public Builder defaultPriority(int priority) {
            settings.defaultPriority = priority;
            return this;
        }

        public Builder autoActivateScene(@Nullable String scene) {
            settings.autoActivateScene = scene;
            return this;
        }

        public GroupSettings build() {
            return settings;
        }
    }
}