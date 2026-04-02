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
import com.ofa.agent.core.TaskRequest;
import com.ofa.agent.core.TaskResult;

/**
 * Cross-Device Router - routes messages and notifications to appropriate devices.
 *
 * Routing decisions based on:
 * 1. Current scene context (running -> watch)
 * 2. Message type (urgent -> phone, casual -> watch)
 * 3. Device roles and capabilities
 * 4. Device availability and connectivity
 *
 * Example scenarios:
 * - Running scene + WeChat message -> route to watch
 * - Heart rate anomaly -> route to phone for alert
 * - Delivery notification -> route to watch for quick view
 * - Important work message -> route to phone
 */
public class CrossDeviceRouter {

    private static final String TAG = "CrossDeviceRouter";

    private final Context context;
    private final AgentProfile localProfile;
    private final PeerNetwork peerNetwork;
    private final SceneDetector sceneDetector;

    // Device registry
    private final Map<String, DeviceInfo> deviceRegistry = new ConcurrentHashMap<>();

    // Routing rules
    private final List<RoutingRule> routingRules = new ArrayList<>();

    // Listeners
    private final List<RoutingListener> listeners = new ArrayList<>();

    /**
     * Device information
     */
    public static class DeviceInfo {
        public final String deviceId;
        public final String name;
        public final AgentProfile.AgentType type;
        public final List<DeviceRole> roles;
        public final List<String> capabilities;
        public volatile boolean online;
        public volatile SceneContext currentScene;

        public DeviceInfo(String deviceId, String name, AgentProfile.AgentType type,
                          List<DeviceRole> roles, List<String> capabilities) {
            this.deviceId = deviceId;
            this.name = name;
            this.type = type;
            this.roles = new ArrayList<>(roles);
            this.capabilities = new ArrayList<>(capabilities);
            this.online = true;
            this.currentScene = SceneContext.unknown();
        }

        public boolean hasRole(int roleType) {
            for (DeviceRole role : roles) {
                if (role.isRole(roleType) && role.isActive()) {
                    return true;
                }
            }
            return false;
        }

        public boolean hasCapability(String capabilityId) {
            return capabilities.contains(capabilityId);
        }

        /**
         * Get display priority for this device
         */
        public int getDisplayPriority() {
            for (DeviceRole role : roles) {
                if (role.isRole(DeviceRole.DISPLAY)) {
                    return role.getPriority();
                }
            }
            return 0;
        }
    }

    /**
     * Routing rule
     */
    public static class RoutingRule {
        public final String name;
        public final RuleCondition condition;
        public final RoutingAction action;
        public final int priority;

        public RoutingRule(String name, RuleCondition condition, RoutingAction action, int priority) {
            this.name = name;
            this.condition = condition;
            this.action = action;
            this.priority = priority;
        }
    }

    /**
     * Rule condition interface
     */
    public interface RuleCondition {
        boolean matches(RoutingContext context);
    }

    /**
     * Routing action interface
     */
    public interface RoutingAction {
        String selectDevice(RoutingContext context);
    }

    /**
     * Routing context
     */
    public static class RoutingContext {
        public final SceneContext scene;
        public final String messageType;
        public final int urgency;
        public final Map<String, Object> metadata;
        public final List<DeviceInfo> availableDevices;

        public RoutingContext(SceneContext scene, String messageType, int urgency,
                              Map<String, Object> metadata, List<DeviceInfo> availableDevices) {
            this.scene = scene;
            this.messageType = messageType;
            this.urgency = urgency;
            this.metadata = new HashMap<>(metadata);
            this.availableDevices = new ArrayList<>(availableDevices);
        }
    }

    /**
     * Routing listener
     */
    public interface RoutingListener {
        void onDeviceSelected(String deviceId, String reason);
        void onRoutingFailed(String reason);
    }

    /**
     * Create cross-device router
     */
    public CrossDeviceRouter(@NonNull Context context, @NonNull AgentProfile localProfile,
                             @NonNull PeerNetwork peerNetwork, @NonNull SceneDetector sceneDetector) {
        this.context = context;
        this.localProfile = localProfile;
        this.peerNetwork = peerNetwork;
        this.sceneDetector = sceneDetector;

        // Initialize default routing rules
        initDefaultRules();

        // Register local device
        registerLocalDevice();

        // Listen for peer discovery
        peerNetwork.setPeerListener(new PeerNetwork.PeerListener() {
            @Override
            public void onPeerDiscovered(PeerNetwork.PeerInfo peer) {
                registerPeerDevice(peer);
            }

            @Override
            public void onPeerLost(String agentId) {
                unregisterDevice(agentId);
            }

            @Override
            public void onMessageReceived(String peerId, String message) {
                handlePeerMessage(peerId, message);
            }
        });

        // Listen for scene changes
        sceneDetector.addListener((oldScene, newScene) -> {
            updateDeviceScene(localProfile.getAgentId(), newScene);
        });
    }

    /**
     * Route a notification to the appropriate device
     *
     * @param messageType Message type (wechat, sms, delivery, urgent, etc.)
     * @param urgency     Urgency level (1-4)
     * @param content     Message content
     * @return Device ID selected for routing
     */
    @Nullable
    public String routeNotification(@NonNull String messageType, int urgency,
                                    @NonNull Map<String, Object> content) {
        // Get current scene
        SceneContext currentScene = sceneDetector.getCurrentScene();

        // Get available devices
        List<DeviceInfo> availableDevices = getAvailableDevices();

        // Create routing context
        RoutingContext routingContext = new RoutingContext(
            currentScene, messageType, urgency, content, availableDevices
        );

        // Apply routing rules
        String selectedDevice = applyRoutingRules(routingContext);

        if (selectedDevice != null) {
            notifyDeviceSelected(selectedDevice, "Routing rule matched");
            return selectedDevice;
        }

        // Fallback: select best display device
        selectedDevice = selectBestDisplayDevice(availableDevices, currentScene);
        if (selectedDevice != null) {
            notifyDeviceSelected(selectedDevice, "Best display device");
            return selectedDevice;
        }

        notifyRoutingFailed("No suitable device available");
        return null;
    }

    /**
     * Apply routing rules in priority order
     */
    @Nullable
    private String applyRoutingRules(@NonNull RoutingContext context) {
        // Sort rules by priority (higher first)
        List<RoutingRule> sortedRules = new ArrayList<>(routingRules);
        sortedRules.sort((a, b) -> Integer.compare(b.priority, a.priority));

        for (RoutingRule rule : sortedRules) {
            if (rule.condition.matches(context)) {
                return rule.action.selectDevice(context);
            }
        }

        return null;
    }

    /**
     * Select best display device based on scene and device capabilities
     */
    @Nullable
    private String selectBestDisplayDevice(@NonNull List<DeviceInfo> devices,
                                           @NonNull SceneContext scene) {
        // If scene prefers specific display type
        String preferredDisplay = scene.getPreferredDisplayDevice();

        if (preferredDisplay != null) {
            // Find device matching preferred type
            for (DeviceInfo device : devices) {
                if (matchesDisplayPreference(device, preferredDisplay)) {
                    return device.deviceId;
                }
            }
        }

        // Find device with highest display priority
        DeviceInfo bestDevice = null;
        int bestPriority = 0;

        for (DeviceInfo device : devices) {
            int priority = device.getDisplayPriority();
            if (priority > bestPriority) {
                bestPriority = priority;
                bestDevice = device;
            }
        }

        return bestDevice != null ? bestDevice.deviceId : null;
    }

    /**
     * Check if device matches display preference
     */
    private boolean matchesDisplayPreference(@NonNull DeviceInfo device, @NonNull String preference) {
        switch (preference) {
            case "watch":
                return device.type == AgentProfile.AgentType.LITE;
            case "phone":
                return device.type == AgentProfile.AgentType.MOBILE;
            case "tv":
                return device.type == AgentProfile.AgentType.IOT;
            case "none":
                return false;
            default:
                return true;
        }
    }

    /**
     * Forward notification to selected device
     */
    public boolean forwardToDevice(@NonNull String deviceId, @NonNull Map<String, Object> notification) {
        if (deviceId.equals(localProfile.getAgentId())) {
            // Local device, handle locally
            return true;
        }

        DeviceInfo device = deviceRegistry.get(deviceId);
        if (device == null || !device.online) {
            Log.w(TAG, "Device not available: " + deviceId);
            return false;
        }

        // Send via peer network
        try {
            org.json.JSONObject json = new org.json.JSONObject();
            json.put("type", "notification_forward");
            json.put("data", new org.json.JSONObject(notification));
            json.put("from", localProfile.getAgentId());

            return peerNetwork.send(deviceId, json.toString());
        } catch (Exception e) {
            Log.e(TAG, "Failed to forward notification: " + e.getMessage());
            return false;
        }
    }

    // ===== Default Routing Rules =====

    private void initDefaultRules() {
        // Rule 1: Running scene -> route to watch (highest priority)
        routingRules.add(new RoutingRule(
            "running_to_watch",
            context -> context.scene.getSceneType().equals(SceneContext.RUNNING),
            context -> findDeviceByType(context, AgentProfile.AgentType.LITE),
            100
        ));

        // Rule 2: Urgent message -> route to phone (high priority)
        routingRules.add(new RoutingRule(
            "urgent_to_phone",
            context -> context.urgency >= 3, // High urgency
            context -> findDeviceByType(context, AgentProfile.AgentType.MOBILE),
            90
        ));

        // Rule 3: Delivery/logistics -> route to watch for quick view
        routingRules.add(new RoutingRule(
            "delivery_to_watch",
            context -> context.messageType.equals("delivery") ||
                       context.messageType.equals("logistics") ||
                       context.messageType.equals("taxi"),
            context -> findDeviceByType(context, AgentProfile.AgentType.LITE),
            80
        ));

        // Rule 4: Meeting scene -> silent notification
        routingRules.add(new RoutingRule(
            "meeting_silent",
            context -> context.scene.getSceneType().equals(SceneContext.MEETING),
            context -> {
                // Route to watch with silent mode
                String watchId = findDeviceByType(context, AgentProfile.AgentType.LITE);
                if (watchId != null) {
                    context.metadata.put("silent", true);
                    return watchId;
                }
                return findDeviceByType(context, AgentProfile.AgentType.MOBILE);
            },
            70
        ));

        // Rule 5: Driving -> voice notification to phone
        routingRules.add(new RoutingRule(
            "driving_voice",
            context -> context.scene.getSceneType().equals(SceneContext.DRIVING),
            context -> {
                context.metadata.put("voice", true);
                return findDeviceByType(context, AgentProfile.AgentType.MOBILE);
            },
            60
        ));

        // Rule 6: Casual message during physical activity -> watch
        routingRules.add(new RoutingRule(
            "casual_physical_to_watch",
            context -> context.scene.isPhysicalActivity() &&
                       context.urgency <= 2, // Low-medium urgency
            context -> findDeviceByType(context, AgentProfile.AgentType.LITE),
            50
        ));
    }

    /**
     * Find device by type
     */
    @Nullable
    private String findDeviceByType(@NonNull RoutingContext context,
                                    @NonNull AgentProfile.AgentType type) {
        for (DeviceInfo device : context.availableDevices) {
            if (device.type == type && device.online) {
                return device.deviceId;
            }
        }
        return null;
    }

    // ===== Device Registry =====

    private void registerLocalDevice() {
        List<DeviceRole> roles = new ArrayList<>();
        roles.add(DeviceRole.asDisplay(8));
        roles.add(DeviceRole.asExecutor(9));
        roles.add(DeviceRole.asCoordinator(7));

        List<String> capabilities = new ArrayList<>();
        for (AgentProfile.Capability cap : localProfile.getCapabilities()) {
            capabilities.add(cap.id);
        }

        DeviceInfo localDevice = new DeviceInfo(
            localProfile.getAgentId(),
            localProfile.getName(),
            localProfile.getType(),
            roles,
            capabilities
        );
        localDevice.currentScene = sceneDetector.getCurrentScene();

        deviceRegistry.put(localProfile.getAgentId(), localDevice);
    }

    private void registerPeerDevice(PeerNetwork.PeerInfo peer) {
        List<DeviceRole> roles = new ArrayList<>();
        // Infer roles from device type
        if (peer.type == AgentProfile.AgentType.LITE) {
            roles.add(DeviceRole.asSource(8));  // Watch as data source
            roles.add(DeviceRole.asDisplay(6)); // Watch can display
        } else {
            roles.add(DeviceRole.asDisplay(7));
            roles.add(DeviceRole.asExecutor(8));
        }

        DeviceInfo device = new DeviceInfo(
            peer.agentId,
            peer.name,
            peer.type,
            roles,
            peer.capabilities
        );

        deviceRegistry.put(peer.agentId, device);
        Log.i(TAG, "Device registered: " + peer.name + " (type=" + peer.type + ")");
    }

    private void unregisterDevice(String deviceId) {
        deviceRegistry.remove(deviceId);
        Log.i(TAG, "Device unregistered: " + deviceId);
    }

    private void updateDeviceScene(String deviceId, SceneContext scene) {
        DeviceInfo device = deviceRegistry.get(deviceId);
        if (device != null) {
            device.currentScene = scene;
            Log.d(TAG, "Device scene updated: " + deviceId + " -> " + scene.getSceneType());
        }
    }

    @NonNull
    private List<DeviceInfo> getAvailableDevices() {
        List<DeviceInfo> available = new ArrayList<>();
        for (DeviceInfo device : deviceRegistry.values()) {
            if (device.online) {
                available.add(device);
            }
        }
        return available;
    }

    // ===== Message Handling =====

    private void handlePeerMessage(String peerId, String message) {
        try {
            org.json.JSONObject json = new org.json.JSONObject(message);
            String type = json.optString("type");

            switch (type) {
                case "scene_update":
                    handleSceneUpdate(peerId, json);
                    break;

                case "notification_forward":
                    handleForwardedNotification(json);
                    break;

                case "health_alert":
                    handleHealthAlert(json);
                    break;

                case "event_subscription":
                    handleSubscription(peerId, json);
                    break;
            }
        } catch (Exception e) {
            Log.w(TAG, "Message parse error: " + e.getMessage());
        }
    }

    private void handleSceneUpdate(String peerId, org.json.JSONObject json) {
        try {
            String sceneType = json.optString("scene_type");
            float confidence = (float) json.optDouble("confidence", 0);

            SceneContext scene = new SceneContext(
                sceneType, confidence, System.currentTimeMillis(),
                peerId, new HashMap<>()
            );

            updateDeviceScene(peerId, scene);

            // Also update local scene detector for cross-device context
            sceneDetector.updateSceneFromExternal(scene);

        } catch (Exception e) {
            Log.w(TAG, "Scene update error: " + e.getMessage());
        }
    }

    private void handleForwardedNotification(org.json.JSONObject json) {
        // Handle notification forwarded from another device
        Log.i(TAG, "Received forwarded notification");
        // Would delegate to SocialOrchestrator
    }

    private void handleHealthAlert(org.json.JSONObject json) {
        try {
            String alertType = json.optString("alert_type");
            float value = (float) json.optDouble("value", 0);
            String sourceDevice = json.optString("source");

            Log.i(TAG, "Health alert from " + sourceDevice + ": " + alertType + "=" + value);

            // Trigger local alert
            Map<String, Object> alert = new HashMap<>();
            alert.put("type", "health");
            alert.put("alert_type", alertType);
            alert.put("value", value);
            alert.put("source", sourceDevice);
            alert.put("urgency", 3); // High urgency

            // Route to local display
            // Would delegate to notification system

        } catch (Exception e) {
            Log.w(TAG, "Health alert error: " + e.getMessage());
        }
    }

    private void handleSubscription(String peerId, org.json.JSONObject json) {
        // Handle event subscription request
        Log.i(TAG, "Subscription request from " + peerId);
    }

    // ===== Listener Notifications =====

    private void notifyDeviceSelected(String deviceId, String reason) {
        for (RoutingListener listener : listeners) {
            listener.onDeviceSelected(deviceId, reason);
        }
    }

    private void notifyRoutingFailed(String reason) {
        for (RoutingListener listener : listeners) {
            listener.onRoutingFailed(reason);
        }
    }

    public void addListener(@NonNull RoutingListener listener) {
        listeners.add(listener);
    }

    public void removeListener(@NonNull RoutingListener listener) {
        listeners.remove(listener);
    }

    /**
     * Add custom routing rule
     */
    public void addRoutingRule(@NonNull RoutingRule rule) {
        routingRules.add(rule);
    }

    /**
     * Get all registered devices
     */
    @NonNull
    public Map<String, DeviceInfo> getDeviceRegistry() {
        return new HashMap<>(deviceRegistry);
    }
}