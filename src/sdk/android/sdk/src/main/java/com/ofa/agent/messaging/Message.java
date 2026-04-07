package com.ofa.agent.messaging;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.HashMap;
import java.util.Map;

/**
 * 消息模型 (v3.0.0)
 *
 * 表示设备间传递的消息。
 */
public class Message {

    // 消息类型
    public static final String TYPE_COMMAND = "command";
    public static final String TYPE_NOTIFICATION = "notification";
    public static final String TYPE_DATA = "data";
    public static final String TYPE_SYNC = "sync";
    public static final String TYPE_ACK = "ack";
    public static final String TYPE_HEARTBEAT = "heartbeat";

    // 消息优先级
    public static final int PRIORITY_LOW = 0;
    public static final int PRIORITY_NORMAL = 1;
    public static final int PRIORITY_HIGH = 2;
    public static final int PRIORITY_URGENT = 3;

    // 消息状态
    public static final String STATUS_PENDING = "pending";
    public static final String STATUS_SENT = "sent";
    public static final String STATUS_DELIVERED = "delivered";
    public static final String STATUS_ACKED = "acked";
    public static final String STATUS_FAILED = "failed";
    public static final String STATUS_EXPIRED = "expired";

    // 字段
    public String id;
    public String fromAgent;
    public String toAgent;
    public String identityId;
    public String type;
    public int priority;
    public Map<String, Object> payload;
    public Map<String, String> metadata;
    public long createdAt;
    public Long expiresAt;
    public Long deliveredAt;
    public Long ackedAt;
    public String status;
    public int retryCount;
    public int maxRetries;

    public Message() {
        this.payload = new HashMap<>();
        this.metadata = new HashMap<>();
        this.priority = PRIORITY_NORMAL;
        this.status = STATUS_PENDING;
        this.maxRetries = 3;
        this.retryCount = 0;
    }

    /**
     * 创建命令消息
     */
    public static Message createCommand(String from, String to, Map<String, Object> payload) {
        Message msg = new Message();
        msg.fromAgent = from;
        msg.toAgent = to;
        msg.type = TYPE_COMMAND;
        msg.payload = payload;
        msg.createdAt = System.currentTimeMillis();
        return msg;
    }

    /**
     * 创建通知消息
     */
    public static Message createNotification(String from, String to, Map<String, Object> payload) {
        Message msg = new Message();
        msg.fromAgent = from;
        msg.toAgent = to;
        msg.type = TYPE_NOTIFICATION;
        msg.payload = payload;
        msg.createdAt = System.currentTimeMillis();
        return msg;
    }

    /**
     * 创建数据消息
     */
    public static Message createData(String from, String to, Map<String, Object> payload) {
        Message msg = new Message();
        msg.fromAgent = from;
        msg.toAgent = to;
        msg.type = TYPE_DATA;
        msg.payload = payload;
        msg.createdAt = System.currentTimeMillis();
        return msg;
    }

    /**
     * 创建同步消息
     */
    public static Message createSync(String from, String to, Map<String, Object> payload) {
        Message msg = new Message();
        msg.fromAgent = from;
        msg.toAgent = to;
        msg.type = TYPE_SYNC;
        msg.payload = payload;
        msg.priority = PRIORITY_HIGH;
        msg.createdAt = System.currentTimeMillis();
        return msg;
    }

    /**
     * 创建确认消息
     */
    public static Message createAck(String from, String to, String originalMessageId) {
        Message msg = new Message();
        msg.fromAgent = from;
        msg.toAgent = to;
        msg.type = TYPE_ACK;
        msg.priority = PRIORITY_HIGH;
        msg.createdAt = System.currentTimeMillis();
        msg.metadata.put("original_message_id", originalMessageId);
        return msg;
    }

    /**
     * 检查是否过期
     */
    public boolean isExpired() {
        if (expiresAt == null) {
            return false;
        }
        return System.currentTimeMillis() > expiresAt;
    }

    /**
     * 检查是否应该重试
     */
    public boolean shouldRetry() {
        return STATUS_FAILED.equals(status) &&
                retryCount < maxRetries &&
                !isExpired();
    }

    /**
     * 设置过期时间
     */
    public Message setTTL(long ttlMs) {
        this.expiresAt = createdAt + ttlMs;
        return this;
    }

    /**
     * 设置优先级
     */
    public Message setPriority(int priority) {
        this.priority = priority;
        return this;
    }

    /**
     * 添加元数据
     */
    public Message addMetadata(String key, String value) {
        if (metadata == null) {
            metadata = new HashMap<>();
        }
        metadata.put(key, value);
        return this;
    }

    // JSON 序列化
    public JSONObject toJson() throws JSONException {
        JSONObject json = new JSONObject();
        json.put("id", id);
        json.put("from_agent", fromAgent);
        json.put("to_agent", toAgent);
        json.put("identity_id", identityId);
        json.put("type", type);
        json.put("priority", priority);

        if (payload != null) {
            JSONObject payloadJson = new JSONObject();
            for (Map.Entry<String, Object> e : payload.entrySet()) {
                payloadJson.put(e.getKey(), e.getValue());
            }
            json.put("payload", payloadJson);
        }

        if (metadata != null) {
            JSONObject metadataJson = new JSONObject();
            for (Map.Entry<String, String> e : metadata.entrySet()) {
                metadataJson.put(e.getKey(), e.getValue());
            }
            json.put("metadata", metadataJson);
        }

        json.put("created_at", createdAt);
        if (expiresAt != null) {
            json.put("expires_at", expiresAt);
        }
        if (deliveredAt != null) {
            json.put("delivered_at", deliveredAt);
        }
        if (ackedAt != null) {
            json.put("acked_at", ackedAt);
        }

        json.put("status", status);
        json.put("retry_count", retryCount);
        json.put("max_retries", maxRetries);

        return json;
    }

    // JSON 反序列化
    public static Message fromJson(JSONObject json) throws JSONException {
        Message msg = new Message();
        msg.id = json.optString("id");
        msg.fromAgent = json.optString("from_agent");
        msg.toAgent = json.optString("to_agent");
        msg.identityId = json.optString("identity_id");
        msg.type = json.optString("type", TYPE_DATA);
        msg.priority = json.optInt("priority", PRIORITY_NORMAL);
        msg.createdAt = json.optLong("created_at", System.currentTimeMillis());
        msg.status = json.optString("status", STATUS_PENDING);
        msg.retryCount = json.optInt("retry_count", 0);
        msg.maxRetries = json.optInt("max_retries", 3);

        if (json.has("expires_at")) {
            msg.expiresAt = json.getLong("expires_at");
        }
        if (json.has("delivered_at")) {
            msg.deliveredAt = json.getLong("delivered_at");
        }
        if (json.has("acked_at")) {
            msg.ackedAt = json.getLong("acked_at");
        }

        // 解析 payload
        JSONObject payloadJson = json.optJSONObject("payload");
        if (payloadJson != null) {
            msg.payload = new HashMap<>();
            for (java.util.Iterator<String> it = payloadJson.keys(); it.hasNext(); ) {
                String key = it.next();
                msg.payload.put(key, payloadJson.get(key));
            }
        }

        // 解析 metadata
        JSONObject metadataJson = json.optJSONObject("metadata");
        if (metadataJson != null) {
            msg.metadata = new HashMap<>();
            for (java.util.Iterator<String> it = metadataJson.keys(); it.hasNext(); ) {
                String key = it.next();
                msg.metadata.put(key, metadataJson.getString(key));
            }
        }

        return msg;
    }

    @Override
    public String toString() {
        return "Message{" +
                "id='" + id + '\'' +
                ", from='" + fromAgent + '\'' +
                ", to='" + toAgent + '\'' +
                ", type='" + type + '\'' +
                ", priority=" + priority +
                ", status='" + status + '\'' +
                '}';
    }
}