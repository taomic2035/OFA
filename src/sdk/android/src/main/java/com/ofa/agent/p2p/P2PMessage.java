package com.ofa.agent.p2p;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.nio.charset.StandardCharsets;
import java.util.UUID;

/**
 * P2P 消息
 */
public class P2PMessage {
    public static final String TYPE_DATA = "data";
    public static final String TYPE_BROADCAST = "broadcast";
    public static final String TYPE_REQUEST = "request";
    public static final String TYPE_RESPONSE = "response";

    public final String type;
    public final String fromId;
    public final String toId;
    public final byte[] data;
    public final long timestamp;
    public final String msgId;

    public P2PMessage(@NonNull String type, @NonNull String fromId, @Nullable String toId, @Nullable byte[] data) {
        this.type = type;
        this.fromId = fromId;
        this.toId = toId;
        this.data = data;
        this.timestamp = System.currentTimeMillis();
        this.msgId = UUID.randomUUID().toString().substring(0, 8);
    }

    private P2PMessage(String type, String fromId, String toId, byte[] data, long timestamp, String msgId) {
        this.type = type;
        this.fromId = fromId;
        this.toId = toId;
        this.data = data;
        this.timestamp = timestamp;
        this.msgId = msgId;
    }

    @NonNull
    public String toJson() {
        try {
            JSONObject json = new JSONObject();
            json.put("type", type);
            json.put("from", fromId);
            json.put("to", toId);
            json.put("timestamp", timestamp);
            json.put("msgId", msgId);

            if (data != null) {
                json.put("data", new String(data, StandardCharsets.UTF_8));
            }

            return json.toString();
        } catch (JSONException e) {
            return "{}";
        }
    }

    @NonNull
    public static P2PMessage fromJson(@NonNull String json) throws JSONException {
        JSONObject obj = new JSONObject(json);

        String type = obj.getString("type");
        String fromId = obj.getString("from");
        String toId = obj.optString("to", null);
        long timestamp = obj.optLong("timestamp", System.currentTimeMillis());
        String msgId = obj.optString("msgId", UUID.randomUUID().toString().substring(0, 8));

        byte[] data = null;
        if (obj.has("data")) {
            data = obj.getString("data").getBytes(StandardCharsets.UTF_8);
        }

        return new P2PMessage(type, fromId, toId, data, timestamp, msgId);
    }
}