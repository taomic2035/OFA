package com.ofa.agent.notification;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.HashMap;
import java.util.Map;

/**
 * 通知动作 (v3.4.0)
 */
public class NotificationAction {

    // 动作类型
    public static final String TYPE_OPEN = "open";
    public static final String TYPE_REPLY = "reply";
    public static final String TYPE_DISMISS = "dismiss";
    public static final String TYPE_CUSTOM = "custom";

    private String actionId;
    private String title;
    private String type;
    private Map<String, Object> payload;

    public NotificationAction() {
        this.type = TYPE_OPEN;
        this.payload = new HashMap<>();
    }

    /**
     * 创建动作
     */
    @NonNull
    public static NotificationAction create(@NonNull String actionId, @NonNull String title, @NonNull String type) {
        NotificationAction action = new NotificationAction();
        action.actionId = actionId;
        action.title = title;
        action.type = type;
        return action;
    }

    /**
     * 创建打开动作
     */
    @NonNull
    public static NotificationAction open(@NonNull String actionId, @NonNull String title) {
        return create(actionId, title, TYPE_OPEN);
    }

    /**
     * 创建回复动作
     */
    @NonNull
    public static NotificationAction reply(@NonNull String actionId, @NonNull String title) {
        return create(actionId, title, TYPE_REPLY);
    }

    /**
     * 创建忽略动作
     */
    @NonNull
    public static NotificationAction dismiss(@NonNull String actionId) {
        return create(actionId, "Dismiss", TYPE_DISMISS);
    }

    /**
     * 从 JSON 解析
     */
    @NonNull
    public static NotificationAction fromJson(@NonNull JSONObject json) throws JSONException {
        NotificationAction action = new NotificationAction();
        action.actionId = json.optString("action_id");
        action.title = json.optString("title");
        action.type = json.optString("type", TYPE_OPEN);

        JSONObject payloadJson = json.optJSONObject("payload");
        if (payloadJson != null) {
            action.payload = new HashMap<>();
            for (java.util.Iterator<String> it = payloadJson.keys(); it.hasNext(); ) {
                String key = it.next();
                action.payload.put(key, payloadJson.get(key));
            }
        }

        return action;
    }

    /**
     * 转换为 JSON
     */
    @NonNull
    public JSONObject toJson() throws JSONException {
        JSONObject json = new JSONObject();
        json.put("action_id", actionId);
        json.put("title", title);
        json.put("type", type);

        if (payload != null && !payload.isEmpty()) {
            JSONObject payloadJson = new JSONObject();
            for (Map.Entry<String, Object> e : payload.entrySet()) {
                payloadJson.put(e.getKey(), e.getValue());
            }
            json.put("payload", payloadJson);
        }

        return json;
    }

    // Getter/Setter

    public String getActionId() { return actionId; }
    public void setActionId(String actionId) { this.actionId = actionId; }

    public String getTitle() { return title; }
    public void setTitle(String title) { this.title = title; }

    public String getType() { return type; }
    public void setType(String type) { this.type = type; }

    public Map<String, Object> getPayload() { return payload; }
    public void setPayload(Map<String, Object> payload) { this.payload = payload; }

    public void addPayload(@NonNull String key, Object value) {
        if (payload == null) {
            payload = new HashMap<>();
        }
        payload.put(key, value);
    }

    @NonNull
    @Override
    public String toString() {
        return "NotificationAction{" +
                "id='" + actionId + '\'' +
                ", title='" + title + '\'' +
                ", type='" + type + '\'' +
                '}';
    }
}