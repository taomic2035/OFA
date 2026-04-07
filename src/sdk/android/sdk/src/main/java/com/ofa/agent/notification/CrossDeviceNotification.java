package com.ofa.agent.notification;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * 跨设备通知模型 (v3.4.0)
 */
public class CrossDeviceNotification {

    // 通知类型
    public static final String TYPE_MESSAGE = "message";
    public static final String TYPE_ALERT = "alert";
    public static final String TYPE_REMINDER = "reminder";
    public static final String TYPE_SYSTEM = "system";
    public static final String TYPE_SOCIAL = "social";
    public static final String TYPE_HEALTH = "health";
    public static final String TYPE_CALENDAR = "calendar";
    public static final String TYPE_CALL = "call";

    // 优先级
    public static final int PRIORITY_MIN = 0;
    public static final int PRIORITY_LOW = 1;
    public static final int PRIORITY_NORMAL = 2;
    public static final int PRIORITY_HIGH = 3;
    public static final int PRIORITY_MAX = 4;

    // 状态
    public static final String STATUS_PENDING = "pending";
    public static final String STATUS_DELIVERED = "delivered";
    public static final String STATUS_READ = "read";
    public static final String STATUS_DISMISSED = "dismissed";
    public static final String STATUS_FAILED = "failed";
    public static final String STATUS_EXPIRED = "expired";

    // 字段
    private String notificationId;
    private String identityId;
    private String type;
    private int priority;
    private String title;
    private String body;
    private String icon;
    private String image;
    private String sound;
    private boolean vibrate;

    // 动作
    private List<NotificationAction> actions;

    // 目标设备
    private List<String> targetDevices;
    private List<String> targetScenes;
    private List<String> excludeDevices;

    // 状态追踪
    private String status;
    private List<String> deliveredTo;
    private List<String> readBy;
    private List<String> dismissedBy;

    // 时间
    private long createdAt;
    private Long expiresAt;
    private Long deliveredAt;
    private Long readAt;

    // 元数据
    private String sourceApp;
    private String sourceAgent;
    private String category;
    private String groupId;
    private String threadId;

    // 扩展数据
    private Map<String, Object> data;
    private Map<String, Object> metadata;

    // === 构造函数 ===

    public CrossDeviceNotification() {
        this.priority = PRIORITY_NORMAL;
        this.status = STATUS_PENDING;
        this.vibrate = true;
        this.actions = new ArrayList<>();
        this.targetDevices = new ArrayList<>();
        this.targetScenes = new ArrayList<>();
        this.excludeDevices = new ArrayList<>();
        this.deliveredTo = new ArrayList<>();
        this.readBy = new ArrayList<>();
        this.dismissedBy = new ArrayList<>();
        this.data = new HashMap<>();
        this.metadata = new HashMap<>();
        this.createdAt = System.currentTimeMillis();
    }

    // === 工厂方法 ===

    /**
     * 创建简单通知
     */
    @NonNull
    public static CrossDeviceNotification create(@NonNull String title, @NonNull String body) {
        CrossDeviceNotification notification = new CrossDeviceNotification();
        notification.title = title;
        notification.body = body;
        return notification;
    }

    /**
     * 创建高优先级通知
     */
    @NonNull
    public static CrossDeviceNotification createAlert(@NonNull String title, @NonNull String body) {
        CrossDeviceNotification notification = create(title, body);
        notification.type = TYPE_ALERT;
        notification.priority = PRIORITY_HIGH;
        return notification;
    }

    /**
     * 从 JSON 解析
     */
    @NonNull
    public static CrossDeviceNotification fromJson(@NonNull JSONObject json) throws JSONException {
        CrossDeviceNotification notification = new CrossDeviceNotification();

        notification.notificationId = json.optString("notification_id");
        notification.identityId = json.optString("identity_id");
        notification.type = json.optString("type", TYPE_MESSAGE);
        notification.priority = json.optInt("priority", PRIORITY_NORMAL);
        notification.title = json.optString("title");
        notification.body = json.optString("body");
        notification.icon = json.optString("icon");
        notification.image = json.optString("image");
        notification.sound = json.optString("sound");
        notification.vibrate = json.optBoolean("vibrate", true);
        notification.status = json.optString("status", STATUS_PENDING);
        notification.sourceApp = json.optString("source_app");
        notification.sourceAgent = json.optString("source_agent");
        notification.category = json.optString("category");
        notification.groupId = json.optString("group_id");
        notification.threadId = json.optString("thread_id");
        notification.createdAt = json.optLong("created_at", System.currentTimeMillis());

        // 解析过期时间
        if (json.has("expires_at")) {
            notification.expiresAt = json.getLong("expires_at");
        }
        if (json.has("delivered_at")) {
            notification.deliveredAt = json.getLong("delivered_at");
        }
        if (json.has("read_at")) {
            notification.readAt = json.getLong("read_at");
        }

        // 解析动作列表
        JSONArray actionsArray = json.optJSONArray("actions");
        if (actionsArray != null) {
            notification.actions = new ArrayList<>();
            for (int i = 0; i < actionsArray.length(); i++) {
                notification.actions.add(NotificationAction.fromJson(actionsArray.getJSONObject(i)));
            }
        }

        // 解析目标设备
        JSONArray targetDevicesArray = json.optJSONArray("target_devices");
        if (targetDevicesArray != null) {
            notification.targetDevices = new ArrayList<>();
            for (int i = 0; i < targetDevicesArray.length(); i++) {
                notification.targetDevices.add(targetDevicesArray.getString(i));
            }
        }

        // 解析目标场景
        JSONArray targetScenesArray = json.optJSONArray("target_scenes");
        if (targetScenesArray != null) {
            notification.targetScenes = new ArrayList<>();
            for (int i = 0; i < targetScenesArray.length(); i++) {
                notification.targetScenes.add(targetScenesArray.getString(i));
            }
        }

        // 解析送达列表
        JSONArray deliveredToArray = json.optJSONArray("delivered_to");
        if (deliveredToArray != null) {
            notification.deliveredTo = new ArrayList<>();
            for (int i = 0; i < deliveredToArray.length(); i++) {
                notification.deliveredTo.add(deliveredToArray.getString(i));
            }
        }

        // 解析已读列表
        JSONArray readByArray = json.optJSONArray("read_by");
        if (readByArray != null) {
            notification.readBy = new ArrayList<>();
            for (int i = 0; i < readByArray.length(); i++) {
                notification.readBy.add(readByArray.getString(i));
            }
        }

        // 解析 data
        JSONObject dataJson = json.optJSONObject("data");
        if (dataJson != null) {
            notification.data = new HashMap<>();
            for (java.util.Iterator<String> it = dataJson.keys(); it.hasNext(); ) {
                String key = it.next();
                notification.data.put(key, dataJson.get(key));
            }
        }

        // 解析 metadata
        JSONObject metadataJson = json.optJSONObject("metadata");
        if (metadataJson != null) {
            notification.metadata = new HashMap<>();
            for (java.util.Iterator<String> it = metadataJson.keys(); it.hasNext(); ) {
                String key = it.next();
                notification.metadata.put(key, metadataJson.get(key));
            }
        }

        return notification;
    }

    // === JSON 序列化 ===

    @NonNull
    public JSONObject toJson() throws JSONException {
        JSONObject json = new JSONObject();

        json.put("notification_id", notificationId);
        json.put("identity_id", identityId);
        json.put("type", type);
        json.put("priority", priority);
        json.put("title", title);
        json.put("body", body);
        json.put("vibrate", vibrate);
        json.put("status", status);
        json.put("created_at", createdAt);

        if (icon != null) json.put("icon", icon);
        if (image != null) json.put("image", image);
        if (sound != null) json.put("sound", sound);
        if (sourceApp != null) json.put("source_app", sourceApp);
        if (sourceAgent != null) json.put("source_agent", sourceAgent);
        if (category != null) json.put("category", category);
        if (groupId != null) json.put("group_id", groupId);
        if (threadId != null) json.put("thread_id", threadId);

        if (expiresAt != null) json.put("expires_at", expiresAt);
        if (deliveredAt != null) json.put("delivered_at", deliveredAt);
        if (readAt != null) json.put("read_at", readAt);

        if (!actions.isEmpty()) {
            JSONArray actionsArray = new JSONArray();
            for (NotificationAction action : actions) {
                actionsArray.put(action.toJson());
            }
            json.put("actions", actionsArray);
        }

        if (!targetDevices.isEmpty()) {
            json.put("target_devices", new JSONArray(targetDevices));
        }

        if (!deliveredTo.isEmpty()) {
            json.put("delivered_to", new JSONArray(deliveredTo));
        }

        if (!readBy.isEmpty()) {
            json.put("read_by", new JSONArray(readBy));
        }

        if (!data.isEmpty()) {
            JSONObject dataJson = new JSONObject();
            for (Map.Entry<String, Object> e : data.entrySet()) {
                dataJson.put(e.getKey(), e.getValue());
            }
            json.put("data", dataJson);
        }

        return json;
    }

    // === 状态检查 ===

    public boolean isPending() { return STATUS_PENDING.equals(status); }
    public boolean isDelivered() { return STATUS_DELIVERED.equals(status); }
    public boolean isRead() { return STATUS_READ.equals(status); }
    public boolean isDismissed() { return STATUS_DISMISSED.equals(status); }
    public boolean isExpired() {
        return STATUS_EXPIRED.equals(status) ||
                (expiresAt != null && System.currentTimeMillis() > expiresAt);
    }

    public boolean isHighPriority() { return priority >= PRIORITY_HIGH; }

    /**
     * 检查是否需要用户关注
     */
    public boolean requiresAttention() {
        return !isRead() && !isDismissed() && !isExpired();
    }

    // === Getter/Setter ===

    public String getNotificationId() { return notificationId; }
    public void setNotificationId(String notificationId) { this.notificationId = notificationId; }

    public String getIdentityId() { return identityId; }
    public void setIdentityId(String identityId) { this.identityId = identityId; }

    public String getType() { return type; }
    public void setType(String type) { this.type = type; }

    public int getPriority() { return priority; }
    public void setPriority(int priority) { this.priority = priority; }

    public String getTitle() { return title; }
    public void setTitle(String title) { this.title = title; }

    public String getBody() { return body; }
    public void setBody(String body) { this.body = body; }

    public String getIcon() { return icon; }
    public void setIcon(String icon) { this.icon = icon; }

    public String getImage() { return image; }
    public void setImage(String image) { this.image = image; }

    public String getSound() { return sound; }
    public void setSound(String sound) { this.sound = sound; }

    public boolean isVibrate() { return vibrate; }
    public void setVibrate(boolean vibrate) { this.vibrate = vibrate; }

    public List<NotificationAction> getActions() { return actions; }
    public void setActions(List<NotificationAction> actions) { this.actions = actions; }

    public String getStatus() { return status; }
    public void setStatus(String status) { this.status = status; }

    public List<String> getDeliveredTo() { return deliveredTo; }
    public List<String> getReadBy() { return readBy; }
    public List<String> getDismissedBy() { return dismissedBy; }

    public long getCreatedAt() { return createdAt; }
    public void setCreatedAt(long createdAt) { this.createdAt = createdAt; }

    public Long getExpiresAt() { return expiresAt; }
    public void setExpiresAt(Long expiresAt) { this.expiresAt = expiresAt; }

    public Long getDeliveredAt() { return deliveredAt; }
    public void setDeliveredAt(Long deliveredAt) { this.deliveredAt = deliveredAt; }

    public Long getReadAt() { return readAt; }
    public void setReadAt(Long readAt) { this.readAt = readAt; }

    public String getSourceApp() { return sourceApp; }
    public void setSourceApp(String sourceApp) { this.sourceApp = sourceApp; }

    public String getSourceAgent() { return sourceAgent; }
    public void setSourceAgent(String sourceAgent) { this.sourceAgent = sourceAgent; }

    public String getCategory() { return category; }
    public void setCategory(String category) { this.category = category; }

    public String getGroupId() { return groupId; }
    public void setGroupId(String groupId) { this.groupId = groupId; }

    public String getThreadId() { return threadId; }
    public void setThreadId(String threadId) { this.threadId = threadId; }

    public Map<String, Object> getData() { return data; }
    public void setData(Map<String, Object> data) { this.data = data; }

    public Map<String, Object> getMetadata() { return metadata; }
    public void setMetadata(Map<String, Object> metadata) { this.metadata = metadata; }

    // === 辅助方法 ===

    /**
     * 添加动作
     */
    public void addAction(@NonNull NotificationAction action) {
        if (actions == null) {
            actions = new ArrayList<>();
        }
        actions.add(action);
    }

    /**
     * 添加数据
     */
    public void addData(@NonNull String key, Object value) {
        if (data == null) {
            data = new HashMap<>();
        }
        data.put(key, value);
    }

    /**
     * 设置过期时间
     */
    public void setTTL(long ttlMs) {
        this.expiresAt = createdAt + ttlMs;
    }

    /**
     * 标记已送达
     */
    public void markDelivered(@NonNull String agentId) {
        if (deliveredTo == null) {
            deliveredTo = new ArrayList<>();
        }
        if (!deliveredTo.contains(agentId)) {
            deliveredTo.add(agentId);
        }
        if (STATUS_PENDING.equals(status)) {
            status = STATUS_DELIVERED;
            deliveredAt = System.currentTimeMillis();
        }
    }

    /**
     * 标记已读
     */
    public void markRead(@NonNull String agentId) {
        if (readBy == null) {
            readBy = new ArrayList<>();
        }
        if (!readBy.contains(agentId)) {
            readBy.add(agentId);
        }
        status = STATUS_READ;
        readAt = System.currentTimeMillis();
    }

    /**
     * 标记已忽略
     */
    public void markDismissed(@NonNull String agentId) {
        if (dismissedBy == null) {
            dismissedBy = new ArrayList<>();
        }
        if (!dismissedBy.contains(agentId)) {
            dismissedBy.add(agentId);
        }
        status = STATUS_DISMISSED;
    }

    @NonNull
    @Override
    public String toString() {
        return "CrossDeviceNotification{" +
                "id='" + notificationId + '\'' +
                ", type='" + type + '\'' +
                ", title='" + title + '\'' +
                ", status='" + status + '\'' +
                '}';
    }
}