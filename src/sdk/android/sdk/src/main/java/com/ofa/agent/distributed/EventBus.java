package com.ofa.agent.distributed;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

import com.ofa.agent.core.AgentProfile;
import com.ofa.agent.core.PeerNetwork;

/**
 * Event Bus - enables event subscription and push notification across devices.
 *
 * Event Types:
 * - scene_change: User scene/context changed
 * - health_data: Health sensor data (heart rate, temperature, steps)
 * - health_alert: Health anomaly detected (abnormal heart rate, temperature)
 * - notification: Message/notification received
 * - device_status: Device online/offline status
 * - location_update: Location changed
 * - activity_update: Activity recognized
 *
 * Subscription Model:
 * - Devices subscribe to event types they're interested in
 * - Events are pushed to all subscribers
 * - Local-first: events are also stored locally for offline sync
 */
public class EventBus {

    private static final String TAG = "EventBus";

    private final Context context;
    private final AgentProfile localProfile;
    private final PeerNetwork peerNetwork;

    // Subscriptions: eventType -> list of subscriber device IDs
    private final Map<String, List<String>> subscriptions = new ConcurrentHashMap<>();

    // Local subscribers: eventType -> list of local listeners
    private final Map<String, List<EventSubscriber>> localSubscribers = new ConcurrentHashMap<>();

    // Event history (for offline sync)
    private final List<Event> eventHistory = new ArrayList<>();
    private static final int MAX_HISTORY_SIZE = 100;

    /**
     * Event definition
     */
    public static class Event {
        public final String eventId;
        public final String eventType;
        public final String sourceDeviceId;
        public final long timestamp;
        public final Map<String, Object> data;
        public final int priority; // 1-4, higher is more important

        public Event(String eventId, String eventType, String sourceDeviceId,
                     long timestamp, Map<String, Object> data, int priority) {
            this.eventId = eventId;
            this.eventType = eventType;
            this.sourceDeviceId = sourceDeviceId;
            this.timestamp = timestamp;
            this.data = new HashMap<>(data);
            this.priority = priority;
        }

        /**
         * Convert to JSON for transmission
         */
        @NonNull
        public String toJson() {
            try {
                org.json.JSONObject json = new org.json.JSONObject();
                json.put("eventId", eventId);
                json.put("eventType", eventType);
                json.put("sourceDeviceId", sourceDeviceId);
                json.put("timestamp", timestamp);
                json.put("priority", priority);
                json.put("data", new org.json.JSONObject(data));
                return json.toString();
            } catch (Exception e) {
                return "{}";
            }
        }

        /**
         * Parse from JSON
         */
        @Nullable
        public static Event fromJson(@NonNull String jsonStr) {
            try {
                org.json.JSONObject json = new org.json.JSONObject(jsonStr);
                String eventId = json.optString("eventId");
                String eventType = json.optString("eventType");
                String sourceDeviceId = json.optString("sourceDeviceId");
                long timestamp = json.optLong("timestamp");
                int priority = json.optInt("priority", 2);

                Map<String, Object> data = new HashMap<>();
                org.json.JSONObject dataJson = json.optJSONObject("data");
                if (dataJson != null) {
                    for (java.util.Iterator<String> it = dataJson.keys(); it.hasNext(); ) {
                        String key = it.next();
                        data.put(key, dataJson.opt(key));
                    }
                }

                return new Event(eventId, eventType, sourceDeviceId, timestamp, data, priority);
            } catch (Exception e) {
                return null;
            }
        }
    }

    /**
     * Event subscriber interface
     */
    public interface EventSubscriber {
        void onEvent(@NonNull Event event);
    }

    /**
     * Create event bus
     */
    public EventBus(@NonNull Context context, @NonNull AgentProfile localProfile,
                    @NonNull PeerNetwork peerNetwork) {
        this.context = context;
        this.localProfile = localProfile;
        this.peerNetwork = peerNetwork;

        // Subscribe to peer messages
        peerNetwork.setPeerListener(new PeerNetwork.PeerListener() {
            @Override
            public void onPeerDiscovered(PeerNetwork.PeerInfo peer) {
                // Send subscription info to new peer
                sendSubscriptionInfo(peer.agentId);
            }

            @Override
            public void onPeerLost(String agentId) {
                // Remove subscriptions from lost peer
                removePeerSubscriptions(agentId);
            }

            @Override
            public void onMessageReceived(String peerId, String message) {
                handlePeerEvent(peerId, message);
            }
        });
    }

    // ===== Event Types =====

    public static final String SCENE_CHANGE = "scene_change";
    public static final String HEALTH_DATA = "health_data";
    public static final String HEALTH_ALERT = "health_alert";
    public static final String NOTIFICATION = "notification";
    public static final String DEVICE_STATUS = "device_status";
    public static final String LOCATION_UPDATE = "location_update";
    public static final String ACTIVITY_UPDATE = "activity_update";

    // ===== Publishing =====

    /**
     * Publish an event to all subscribers
     */
    public void publish(@NonNull String eventType, @NonNull Map<String, Object> data, int priority) {
        Event event = new Event(
            generateEventId(),
            eventType,
            localProfile.getAgentId(),
            System.currentTimeMillis(),
            data,
            priority
        );

        // Store in history
        addToHistory(event);

        // Notify local subscribers
        notifyLocalSubscribers(event);

        // Push to remote subscribers
        pushToRemoteSubscribers(event);

        Log.d(TAG, "Event published: " + eventType + " (priority=" + priority + ")");
    }

    /**
     * Publish health data event
     */
    public void publishHealthData(@NonNull String dataType, float value, @Nullable Map<String, Object> extra) {
        Map<String, Object> data = new HashMap<>();
        data.put("data_type", dataType);
        data.put("value", value);
        data.put("unit", getUnitForDataType(dataType));
        if (extra != null) {
            data.putAll(extra);
        }

        publish(HEALTH_DATA, data, 2);
    }

    /**
     * Publish health alert event
     */
    public void publishHealthAlert(@NonNull String alertType, float value, float threshold,
                                   @Nullable String recommendation) {
        Map<String, Object> data = new HashMap<>();
        data.put("alert_type", alertType);
        data.put("value", value);
        data.put("threshold", threshold);
        data.put("severity", calculateSeverity(value, threshold));
        if (recommendation != null) {
            data.put("recommendation", recommendation);
        }

        publish(HEALTH_ALERT, data, 4); // Highest priority
    }

    /**
     * Publish scene change event
     */
    public void publishSceneChange(@NonNull SceneContext scene) {
        Map<String, Object> data = new HashMap<>();
        data.put("scene_type", scene.getSceneType());
        data.put("confidence", scene.getConfidence());
        data.put("category", scene.getSceneCategory());
        data.put("preferred_display", scene.getPreferredDisplayDevice());
        data.putAll(scene.getMetadata());

        publish(SCENE_CHANGE, data, 3);
    }

    /**
     * Publish device status event
     */
    public void publishDeviceStatus(boolean online, @Nullable String reason) {
        Map<String, Object> data = new HashMap<>();
        data.put("online", online);
        data.put("agent_id", localProfile.getAgentId());
        data.put("agent_type", localProfile.getType().getValue());
        if (reason != null) {
            data.put("reason", reason);
        }

        publish(DEVICE_STATUS, data, 2);
    }

    // ===== Subscribing =====

    /**
     * Subscribe to an event type locally
     */
    public void subscribe(@NonNull String eventType, @NonNull EventSubscriber subscriber) {
        List<EventSubscriber> subscribers = localSubscribers.computeIfAbsent(
            eventType, k -> new ArrayList<>()
        );
        subscribers.add(subscriber);
        Log.i(TAG, "Local subscribed to: " + eventType);
    }

    /**
     * Unsubscribe from an event type
     */
    public void unsubscribe(@NonNull String eventType, @NonNull EventSubscriber subscriber) {
        List<EventSubscriber> subscribers = localSubscribers.get(eventType);
        if (subscribers != null) {
            subscribers.remove(subscriber);
        }
    }

    /**
     * Request remote device to subscribe to our events
     */
    public void requestRemoteSubscription(@NonNull String deviceId, @NonNull String eventType) {
        try {
            org.json.JSONObject json = new org.json.JSONObject();
            json.put("type", "subscription_request");
            json.put("event_type", eventType);
            json.put("subscriber_id", localProfile.getAgentId());

            peerNetwork.send(deviceId, json.toString());
        } catch (Exception e) {
            Log.w(TAG, "Failed to request subscription: " + e.getMessage());
        }
    }

    /**
     * Accept subscription request from remote device
     */
    public void acceptRemoteSubscription(@NonNull String subscriberId, @NonNull String eventType) {
        List<String> subscribers = subscriptions.computeIfAbsent(
            eventType, k -> new ArrayList<>()
        );
        if (!subscribers.contains(subscriberId)) {
            subscribers.add(subscriberId);
            Log.i(TAG, "Remote subscription accepted: " + subscriberId + " for " + eventType);
        }
    }

    // ===== Event Handling =====

    private void notifyLocalSubscribers(@NonNull Event event) {
        List<EventSubscriber> subscribers = localSubscribers.get(event.eventType);
        if (subscribers != null) {
            for (EventSubscriber subscriber : subscribers) {
                try {
                    subscriber.onEvent(event);
                } catch (Exception e) {
                    Log.w(TAG, "Subscriber error: " + e.getMessage());
                }
            }
        }
    }

    private void pushToRemoteSubscribers(@NonNull Event event) {
        List<String> subscribers = subscriptions.get(event.eventType);
        if (subscribers != null && !subscribers.isEmpty()) {
            String eventJson = event.toJson();

            for (String subscriberId : subscribers) {
                try {
                    org.json.JSONObject message = new org.json.JSONObject();
                    message.put("type", "event_push");
                    message.put("event", eventJson);

                    peerNetwork.send(subscriberId, message.toString());
                } catch (Exception e) {
                    Log.w(TAG, "Failed to push to " + subscriberId + ": " + e.getMessage());
                }
            }
        }
    }

    private void handlePeerEvent(String peerId, String message) {
        try {
            org.json.JSONObject json = new org.json.JSONObject(message);
            String type = json.optString("type");

            switch (type) {
                case "event_push":
                    Event event = Event.fromJson(json.optString("event"));
                    if (event != null) {
                        addToHistory(event);
                        notifyLocalSubscribers(event);
                    }
                    break;

                case "subscription_request":
                    String eventType = json.optString("event_type");
                    String subscriberId = json.optString("subscriber_id");
                    acceptRemoteSubscription(subscriberId, eventType);

                    // Send confirmation
                    org.json.JSONObject confirm = new org.json.JSONObject();
                    confirm.put("type", "subscription_confirmed");
                    confirm.put("event_type", eventType);
                    peerNetwork.send(peerId, confirm.toString());
                    break;

                case "subscription_confirmed":
                    Log.i(TAG, "Subscription confirmed by " + peerId);
                    break;

                case "subscription_info":
                    // Received subscription info from peer
                    handleSubscriptionInfo(json);
                    break;
            }
        } catch (Exception e) {
            Log.w(TAG, "Event handling error: " + e.getMessage());
        }
    }

    private void handleSubscriptionInfo(org.json.JSONObject json) {
        try {
            org.json.JSONArray subscribedTypes = json.optJSONArray("subscribed_types");
            if (subscribedTypes != null) {
                for (int i = 0; i < subscribedTypes.length(); i++) {
                    String eventType = subscribedTypes.getString(i);
                    acceptRemoteSubscription(json.optString("device_id"), eventType);
                }
            }
        } catch (Exception e) {
            Log.w(TAG, "Subscription info error: " + e.getMessage());
        }
    }

    private void sendSubscriptionInfo(String peerId) {
        try {
            org.json.JSONObject json = new org.json.JSONObject();
            json.put("type", "subscription_info");
            json.put("device_id", localProfile.getAgentId());

            org.json.JSONArray subscribedTypes = new org.json.JSONArray();
            for (String eventType : localSubscribers.keySet()) {
                subscribedTypes.put(eventType);
            }
            json.put("subscribed_types", subscribedTypes);

            peerNetwork.send(peerId, json.toString());
        } catch (Exception e) {
            Log.w(TAG, "Failed to send subscription info: " + e.getMessage());
        }
    }

    private void removePeerSubscriptions(String agentId) {
        for (List<String> subscribers : subscriptions.values()) {
            subscribers.remove(agentId);
        }
        Log.i(TAG, "Removed subscriptions for: " + agentId);
    }

    // ===== History =====

    private void addToHistory(@NonNull Event event) {
        eventHistory.add(event);
        if (eventHistory.size() > MAX_HISTORY_SIZE) {
            eventHistory.remove(0);
        }
    }

    /**
     * Get recent events of a specific type
     */
    @NonNull
    public List<Event> getRecentEvents(@NonNull String eventType, int limit) {
        List<Event> result = new ArrayList<>();
        for (int i = eventHistory.size() - 1; i >= 0 && result.size() < limit; i--) {
            Event event = eventHistory.get(i);
            if (event.eventType.equals(eventType)) {
                result.add(event);
            }
        }
        return result;
    }

    /**
     * Get all recent events
     */
    @NonNull
    public List<Event> getAllRecentEvents(int limit) {
        List<Event> result = new ArrayList<>();
        for (int i = eventHistory.size() - 1; i >= 0 && result.size() < limit; i--) {
            result.add(eventHistory.get(i));
        }
        return result;
    }

    // ===== Utility =====

    private String generateEventId() {
        return "evt_" + System.currentTimeMillis() + "_" +
               java.util.UUID.randomUUID().toString().substring(0, 8);
    }

    private String getUnitForDataType(String dataType) {
        switch (dataType) {
            case "heart_rate": return "bpm";
            case "temperature": return "°C";
            case "steps": return "steps";
            case "distance": return "m";
            case "calories": return "cal";
            default: return "";
        }
    }

    private String calculateSeverity(float value, float threshold) {
        float ratio = value / threshold;
        if (ratio > 1.5) return "critical";
        if (ratio > 1.2) return "high";
        if (ratio > 1.0) return "moderate";
        return "low";
    }
}