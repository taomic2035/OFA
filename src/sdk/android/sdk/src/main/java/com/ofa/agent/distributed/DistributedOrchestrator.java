package com.ofa.agent.distributed;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

import com.ofa.agent.core.AgentProfile;
import com.ofa.agent.core.OFAAndroidAgent;
import com.ofa.agent.core.PeerNetwork;
import com.ofa.agent.core.TaskRequest;
import com.ofa.agent.core.TaskResult;

/**
 * Distributed Orchestrator - unified entry point for distributed agent operations.
 *
 * Coordinates:
 * - Scene detection and sharing across devices
 * - Cross-device message routing
 * - Event subscription and push notifications
 * - Health data bridging
 * - Device role management
 *
 * Usage Example:
 * ```java
 * OFAAndroidAgent agent = ...;
 * DistributedOrchestrator distributed = agent.getDistributedOrchestrator();
 *
 * // Get current scene (may be detected by watch)
 * SceneContext scene = distributed.getCurrentScene();
 *
 * // Route notification to best device
 * distributed.routeNotification("wechat", 2, messageData);
 *
 * // Subscribe to health alerts from watch
 * distributed.subscribeHealthAlerts(alert -> {
 *     showHealthWarning(alert);
 * });
 *
 * // Publish scene change (e.g., watch detected running)
 * distributed.publishSceneChange(scene);
 * ```
 */
public class DistributedOrchestrator {

    private static final String TAG = "DistributedOrchestrator";

    private final Context context;
    private final AgentProfile localProfile;
    private final OFAAndroidAgent agent;

    // Components
    private final SceneDetector sceneDetector;
    private final EventBus eventBus;
    private final CrossDeviceRouter router;
    private final HealthDataBridge healthBridge;

    // Executor
    private final ExecutorService executor;

    // State
    private volatile boolean initialized = false;

    // Listeners
    private final List<DistributedListener> listeners = new ArrayList<>();

    /**
     * Distributed listener interface
     */
    public interface DistributedListener {
        void onSceneChanged(@NonNull SceneContext oldScene, @NonNull SceneContext newScene);
        void onDeviceDiscovered(@NonNull CrossDeviceRouter.DeviceInfo device);
        void onDeviceLost(@NonNull String deviceId);
        void onHealthAlert(@NonNull HealthDataBridge.HealthAlert alert);
        void onNotificationForwarded(@NonNull String deviceId, @NonNull Map<String, Object> notification);
    }

    /**
     * Create distributed orchestrator
     */
    public DistributedOrchestrator(@NonNull Context context, @NonNull AgentProfile localProfile,
                                   @NonNull OFAAndroidAgent agent, @NonNull PeerNetwork peerNetwork) {
        this.context = context;
        this.localProfile = localProfile;
        this.agent = agent;

        // Initialize components
        this.sceneDetector = new SceneDetector(context);
        this.eventBus = new EventBus(context, localProfile, peerNetwork);
        this.router = new CrossDeviceRouter(context, localProfile, peerNetwork, sceneDetector);
        this.healthBridge = new HealthDataBridge(context, localProfile, eventBus);

        this.executor = Executors.newCachedThreadPool();

        // Setup listeners
        setupListeners();
    }

    /**
     * Initialize distributed system
     */
    public void initialize() {
        if (initialized) return;

        Log.i(TAG, "Initializing distributed orchestrator...");

        // Start scene detection
        sceneDetector.startDetection();

        // Subscribe to relevant events based on device type
        setupEventSubscriptions();

        initialized = true;
        Log.i(TAG, "Distributed orchestrator initialized");
    }

    /**
     * Shutdown distributed system
     */
    public void shutdown() {
        if (!initialized) return;

        Log.i(TAG, "Shutting down distributed orchestrator...");

        sceneDetector.stopDetection();
        healthBridge.stopMonitoring();
        executor.shutdown();

        initialized = false;
        Log.i(TAG, "Distributed orchestrator shutdown");
    }

    // ===== Scene Management =====

    /**
     * Get current scene context
     */
    @NonNull
    public SceneContext getCurrentScene() {
        return sceneDetector.getCurrentScene();
    }

    /**
     * Add scene change listener
     */
    public void addSceneListener(@NonNull SceneDetector.SceneChangeListener listener) {
        sceneDetector.addListener(listener);
    }

    /**
     * Publish scene change to other devices
     */
    public void publishSceneChange(@NonNull SceneContext scene) {
        eventBus.publishSceneChange(scene);
        Log.i(TAG, "Scene published: " + scene.getSceneType());
    }

    /**
     * Update scene from external device (e.g., watch)
     */
    public void updateSceneFromExternal(@NonNull SceneContext scene) {
        sceneDetector.updateSceneFromExternal(scene);
    }

    // ===== Routing =====

    /**
     * Route notification to appropriate device
     *
     * @param messageType Message type (wechat, sms, delivery, urgent, etc.)
     * @param urgency     Urgency level (1-4)
     * @param content     Message content
     * @return Selected device ID, or null if local
     */
    @Nullable
    public String routeNotification(@NonNull String messageType, int urgency,
                                    @NonNull Map<String, Object> content) {
        String deviceId = router.routeNotification(messageType, urgency, content);

        if (deviceId != null && !deviceId.equals(localProfile.getAgentId())) {
            // Forward to remote device
            router.forwardToDevice(deviceId, content);
            notifyNotificationForwarded(deviceId, content);
        }

        return deviceId;
    }

    /**
     * Route message to best display device based on current scene
     */
    @Nullable
    public String routeToBestDisplay() {
        SceneContext scene = getCurrentScene();
        String preferredDisplay = scene.getPreferredDisplayDevice();

        if (preferredDisplay == null || preferredDisplay.equals("phone")) {
            return localProfile.getAgentId();
        }

        // Find matching device
        Map<String, CrossDeviceRouter.DeviceInfo> devices = router.getDeviceRegistry();
        for (CrossDeviceRouter.DeviceInfo device : devices.values()) {
            if (matchesDisplayPreference(device, preferredDisplay)) {
                return device.deviceId;
            }
        }

        return localProfile.getAgentId();
    }

    private boolean matchesDisplayPreference(CrossDeviceRouter.DeviceInfo device, String preference) {
        switch (preference) {
            case "watch":
                return device.type == AgentProfile.AgentType.LITE;
            case "phone":
                return device.type == AgentProfile.AgentType.MOBILE;
            default:
                return true;
        }
    }

    /**
     * Add routing listener
     */
    public void addRoutingListener(@NonNull CrossDeviceRouter.RoutingListener listener) {
        router.addListener(listener);
    }

    /**
     * Add custom routing rule
     */
    public void addRoutingRule(@NonNull CrossDeviceRouter.RoutingRule rule) {
        router.addRoutingRule(rule);
    }

    // ===== Health Data =====

    /**
     * Start health monitoring
     */
    public void startHealthMonitoring() {
        healthBridge.startMonitoring();
    }

    /**
     * Stop health monitoring
     */
    public void stopHealthMonitoring() {
        healthBridge.stopMonitoring();
    }

    /**
     * Get current health readings
     */
    @NonNull
    public Map<String, Float> getCurrentHealthReadings() {
        return healthBridge.getCurrentReadings();
    }

    /**
     * Subscribe to health alerts
     */
    public void subscribeHealthAlerts(@NonNull HealthDataBridge.HealthAlertListener listener) {
        healthBridge.addAlertListener(listener);
    }

    /**
     * Get health history
     */
    @NonNull
    public List<HealthDataBridge.HealthRecord> getHealthHistory(@Nullable String dataType, int limit) {
        return healthBridge.getHealthHistory(dataType, limit);
    }

    /**
     * Publish health data to other devices
     */
    public void publishHealthData(@NonNull String dataType, float value) {
        eventBus.publishHealthData(dataType, value, null);
    }

    // ===== Event Management =====

    /**
     * Subscribe to events
     */
    public void subscribe(@NonNull String eventType, @NonNull EventBus.EventSubscriber subscriber) {
        eventBus.subscribe(eventType, subscriber);
    }

    /**
     * Unsubscribe from events
     */
    public void unsubscribe(@NonNull String eventType, @NonNull EventBus.EventSubscriber subscriber) {
        eventBus.unsubscribe(eventType, subscriber);
    }

    /**
     * Publish event
     */
    public void publish(@NonNull String eventType, @NonNull Map<String, Object> data, int priority) {
        eventBus.publish(eventType, data, priority);
    }

    /**
     * Get recent events
     */
    @NonNull
    public List<EventBus.Event> getRecentEvents(@NonNull String eventType, int limit) {
        return eventBus.getRecentEvents(eventType, limit);
    }

    // ===== Device Registry =====

    /**
     * Get all registered devices
     */
    @NonNull
    public Map<String, CrossDeviceRouter.DeviceInfo> getDeviceRegistry() {
        return router.getDeviceRegistry();
    }

    /**
     * Get devices by type
     */
    @NonNull
    public List<CrossDeviceRouter.DeviceInfo> getDevicesByType(@NonNull AgentProfile.AgentType type) {
        List<CrossDeviceRouter.DeviceInfo> result = new ArrayList<>();
        for (CrossDeviceRouter.DeviceInfo device : router.getDeviceRegistry().values()) {
            if (device.type == type) {
                result.add(device);
            }
        }
        return result;
    }

    /**
     * Get wearable devices
     */
    @NonNull
    public List<CrossDeviceRouter.DeviceInfo> getWearableDevices() {
        return getDevicesByType(AgentProfile.AgentType.LITE);
    }

    /**
     * Find device by capability
     */
    @Nullable
    public CrossDeviceRouter.DeviceInfo findDeviceByCapability(@NonNull String capabilityId) {
        for (CrossDeviceRouter.DeviceInfo device : router.getDeviceRegistry().values()) {
            if (device.hasCapability(capabilityId) && device.online) {
                return device;
            }
        }
        return null;
    }

    // ===== Task Delegation =====

    /**
     * Delegate task to another device
     */
    @Nullable
    public TaskResult delegateTask(@NonNull String deviceId, @NonNull TaskRequest request) {
        if (deviceId.equals(localProfile.getAgentId())) {
            // Execute locally
            return agent.execute(request).join();
        }

        // Find peer network
        PeerNetwork peerNetwork = agent.getPeerNetwork();
        if (peerNetwork == null) {
            Log.w(TAG, "Peer network not available");
            return TaskResult.failure(request.taskId, "Peer network not available");
        }

        return peerNetwork.requestTask(deviceId, request);
    }

    /**
     * Delegate task to best capable device
     */
    @Nullable
    public TaskResult delegateToBestDevice(@NonNull String requiredCapability, @NonNull TaskRequest request) {
        CrossDeviceRouter.DeviceInfo device = findDeviceByCapability(requiredCapability);
        if (device == null) {
            // Execute locally if possible
            if (localProfile.hasCapability(requiredCapability)) {
                return agent.execute(request).join();
            }
            return TaskResult.failure(request.taskId, "No device with capability: " + requiredCapability);
        }

        return delegateTask(device.deviceId, request);
    }

    // ===== Listeners Setup =====

    private void setupListeners() {
        // Scene changes
        sceneDetector.addListener((oldScene, newScene) -> {
            notifySceneChanged(oldScene, newScene);
            // Auto-publish scene changes to other devices
            publishSceneChange(newScene);
        });

        // Device discovery
        router.addListener(new CrossDeviceRouter.RoutingListener() {
            @Override
            public void onDeviceSelected(String deviceId, String reason) {
                Log.d(TAG, "Device selected: " + deviceId + " (" + reason + ")");
            }

            @Override
            public void onRoutingFailed(String reason) {
                Log.w(TAG, "Routing failed: " + reason);
            }
        });

        // Health alerts
        healthBridge.addAlertListener(alert -> {
            notifyHealthAlert(alert);
        });
    }

    private void setupEventSubscriptions() {
        // Subscribe to events based on device role
        if (localProfile.getType() == AgentProfile.AgentType.MOBILE) {
            // Phone subscribes to health alerts from wearables
            subscribe(EventBus.HEALTH_ALERT, event -> {
                Log.i(TAG, "Received health alert from wearable");
            });

            // Phone subscribes to scene changes from wearables
            subscribe(EventBus.SCENE_CHANGE, event -> {
                String sceneType = (String) event.data.get("scene_type");
                Float confidence = (Float) event.data.get("confidence");
                if (sceneType != null && confidence != null) {
                    updateSceneFromExternal(new SceneContext(
                        sceneType, confidence, event.timestamp,
                        event.sourceDeviceId, event.data
                    ));
                }
            });
        }

        if (localProfile.getType() == AgentProfile.AgentType.LITE) {
            // Watch publishes health data
            subscribe(EventBus.HEALTH_DATA, event -> {
                // Echo health data to other devices
            });
        }
    }

    // ===== Listener Notifications =====

    private void notifySceneChanged(SceneContext oldScene, SceneContext newScene) {
        for (DistributedListener listener : listeners) {
            listener.onSceneChanged(oldScene, newScene);
        }
    }

    private void notifyDeviceDiscovered(CrossDeviceRouter.DeviceInfo device) {
        for (DistributedListener listener : listeners) {
            listener.onDeviceDiscovered(device);
        }
    }

    private void notifyDeviceLost(String deviceId) {
        for (DistributedListener listener : listeners) {
            listener.onDeviceLost(deviceId);
        }
    }

    private void notifyHealthAlert(HealthDataBridge.HealthAlert alert) {
        for (DistributedListener listener : listeners) {
            listener.onHealthAlert(alert);
        }
    }

    private void notifyNotificationForwarded(String deviceId, Map<String, Object> notification) {
        for (DistributedListener listener : listeners) {
            listener.onNotificationForwarded(deviceId, notification);
        }
    }

    /**
     * Add distributed listener
     */
    public void addListener(@NonNull DistributedListener listener) {
        listeners.add(listener);
    }

    /**
     * Remove distributed listener
     */
    public void removeListener(@NonNull DistributedListener listener) {
        listeners.remove(listener);
    }

    // ===== Component Access =====

    @NonNull
    public SceneDetector getSceneDetector() {
        return sceneDetector;
    }

    @NonNull
    public EventBus getEventBus() {
        return eventBus;
    }

    @NonNull
    public CrossDeviceRouter getRouter() {
        return router;
    }

    @NonNull
    public HealthDataBridge getHealthBridge() {
        return healthBridge;
    }
}