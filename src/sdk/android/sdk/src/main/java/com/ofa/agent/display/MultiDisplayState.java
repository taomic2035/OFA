package com.ofa.agent.display;

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
 * MultiDisplay状态模型 (v5.5.0)
 *
 * 端侧接收 Center 推送的多端展示状态，用于渲染和显示管理。
 * 深层展示管理在 Center 端 DisplayEngine 完成。
 */
public class MultiDisplayState {

    // === Rendering Settings ===
    private RenderingSettings renderingSettings;

    // === Device Adaptation ===
    private DeviceAdaptation deviceAdaptation;

    // === Display Sync ===
    private DisplaySync displaySync;

    // === Scene Render Profiles ===
    private List<SceneRenderProfile> sceneRenderProfiles = new ArrayList<>();

    // === Display Context ===
    private DisplayContext context;

    public MultiDisplayState() {
        this.renderingSettings = new RenderingSettings();
        this.deviceAdaptation = new DeviceAdaptation();
        this.displaySync = new DisplaySync();
        this.context = new DisplayContext();
    }

    // === Getters ===

    @NonNull
    public RenderingSettings getRenderingSettings() {
        return renderingSettings;
    }

    @NonNull
    public DeviceAdaptation getDeviceAdaptation() {
        return deviceAdaptation;
    }

    @NonNull
    public DisplaySync getDisplaySync() {
        return displaySync;
    }

    @NonNull
    public List<SceneRenderProfile> getSceneRenderProfiles() {
        return sceneRenderProfiles;
    }

    @NonNull
    public DisplayContext getContext() {
        return context;
    }

    // === Rendering Settings ===

    public static class RenderingSettings {
        public String renderEngine = "webgl";
        public String engineVersion = "2.0";
        public String fallbackEngine = "canvas";

        public String qualityPreset = "medium";
        public String renderQuality = "medium";
        public String textureQuality = "medium";
        public String shadowQuality = "medium";
        public String antiAliasing = "smaa";
        public int anisotropicFiltering = 4;

        public String lightingMode = "mixed";
        public String lightingQuality = "medium";
        public boolean globalIllumination = false;
        public String ambientOcclusion = "ssao";
        public String reflections = "screen_space";

        public String animationQuality = "medium";
        public String animationBlendMode = "linear";
        public boolean inverseKinematics = true;
        public String facialRigQuality = "blendshapes";

        public boolean physicsEnabled = true;
        public String physicsQuality = "medium";
        public boolean clothSimulation = false;
        public boolean hairSimulation = false;

        public boolean postProcessingEnabled = true;
        public boolean bloom = true;
        public String depthOfField = "low";
        public String motionBlur = "none";
        public String colorGrading = "neutral";
        public boolean vignette = false;
        public double chromaticAberration = 0.0;

        public int targetFPS = 60;
        public boolean vSync = true;
        public boolean adaptiveQuality = true;
        public int maxDrawCalls = 1000;
        public int maxTextureMemory = 512;

        public String shaderComplexity = "standard";
        public List<String> customShaders = new ArrayList<>();

        public boolean isHighQuality() {
            return "high".equals(renderQuality) || "ultra".equals(renderQuality);
        }

        public boolean isPostProcessingEnabled() {
            return postProcessingEnabled;
        }

        public boolean isPhysicsEnabled() {
            return physicsEnabled;
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("render_engine", renderEngine);
                json.put("quality_preset", qualityPreset);
                json.put("render_quality", renderQuality);
                json.put("texture_quality", textureQuality);
                json.put("shadow_quality", shadowQuality);
                json.put("anti_aliasing", antiAliasing);
                json.put("lighting_mode", lightingMode);
                json.put("animation_quality", animationQuality);
                json.put("physics_enabled", physicsEnabled);
                json.put("post_processing_enabled", postProcessingEnabled);
                json.put("target_fps", targetFPS);
                json.put("v_sync", vSync);
                json.put("adaptive_quality", adaptiveQuality);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static RenderingSettings fromJson(@NonNull JSONObject json) throws JSONException {
            RenderingSettings settings = new RenderingSettings();
            settings.renderEngine = json.optString("render_engine", "webgl");
            settings.qualityPreset = json.optString("quality_preset", "medium");
            settings.renderQuality = json.optString("render_quality", "medium");
            settings.textureQuality = json.optString("texture_quality", "medium");
            settings.shadowQuality = json.optString("shadow_quality", "medium");
            settings.antiAliasing = json.optString("anti_aliasing", "smaa");
            settings.lightingMode = json.optString("lighting_mode", "mixed");
            settings.animationQuality = json.optString("animation_quality", "medium");
            settings.physicsEnabled = json.optBoolean("physics_enabled", true);
            settings.postProcessingEnabled = json.optBoolean("post_processing_enabled", true);
            settings.targetFPS = json.optInt("target_fps", 60);
            settings.vSync = json.optBoolean("v_sync", true);
            settings.adaptiveQuality = json.optBoolean("adaptive_quality", true);
            return settings;
        }
    }

    // === Device Adaptation ===

    public static class DeviceAdaptation {
        public String adaptationMode = "auto";
        public boolean autoOptimize = true;

        public Map<String, DeviceRenderProfile> deviceProfiles = new HashMap<>();

        public MobileOptimizations mobileOptimizations = new MobileOptimizations();
        public DesktopSettings desktopSettings = new DesktopSettings();
        public VRSettings vrSettings = new VRSettings();
        public ARSettings arSettings = new ARSettings();

        public PerformanceScaling performanceScaling = new PerformanceScaling();
        public BatteryOptimization batteryOptimization = new BatteryOptimization();

        public boolean isAutoAdaptation() {
            return "auto".equals(adaptationMode);
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("adaptation_mode", adaptationMode);
                json.put("auto_optimize", autoOptimize);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static DeviceAdaptation fromJson(@NonNull JSONObject json) throws JSONException {
            DeviceAdaptation adaptation = new DeviceAdaptation();
            adaptation.adaptationMode = json.optString("adaptation_mode", "auto");
            adaptation.autoOptimize = json.optBoolean("auto_optimize", true);
            return adaptation;
        }
    }

    // === Device Render Profile ===

    public static class DeviceRenderProfile {
        public String deviceType;
        public String deviceTier;
        public String maxQuality;
        public String recommendedQuality;
        public double resolutionScale;
        public int textureLimit;
        public int shadowResolution;
        public int particleLimit;
        public double drawDistance;

        public boolean supportsRayTracing;
        public boolean supportsComputeShaders;

        public int minFPSForQualityUp;
        public int maxFPSForQualityDown;
    }

    // === Mobile Optimizations ===

    public static class MobileOptimizations {
        public String gpuPowerMode = "balanced";
        public String textureCompression = "astc";
        public boolean vertexCompression = true;
        public double meshSimplification = 0.3;

        public boolean textureStreaming = true;
        public double lodBias = 1.0;
        public boolean unloadUnusedAssets = true;
        public int memoryBudget = 256;

        public boolean thermalThrottling = true;
        public double thermalThreshold = 45.0;
        public boolean reduceOnThermal = true;

        public boolean batteryAwareMode = true;
        public String lowBatteryQuality = "low";
        public boolean powerSaveMode = false;

        public boolean networkStreaming = false;
        public String streamingQuality = "adaptive";
        public int cacheSize = 100;
    }

    // === Desktop Settings ===

    public static class DesktopSettings {
        public boolean multiMonitorEnabled = false;
        public int primaryMonitor = 0;
        public String monitorLayout = "single";

        public String windowMode = "windowed";
        public int resolutionWidth = 1920;
        public int resolutionHeight = 1080;
        public int refreshRate = 60;
        public boolean hdr = false;
        public boolean wideScreenSupport = true;

        public String gpuSelection = "auto";
        public String vendorPreference = "auto";
        public int vramLimit = 0;
    }

    // === VR Settings ===

    public static class VRSettings {
        public boolean vrEnabled = false;
        public String vrPlatform = "openxr";
        public String vrSdk = "openxr";

        public boolean singlePassRendering = true;
        public boolean foveatedRendering = false;
        public double foveationLevel = 0.5;

        public int vrTargetFPS = 72;
        public boolean reprojection = true;
        public boolean spaceWarp = false;

        public String comfortMode = "none";
        public boolean vignetteOnMotion = false;
        public boolean snapTurn = true;
        public boolean teleportOnly = false;

        public boolean handTrackingEnabled = true;
        public String handTrackingQuality = "medium";
        public boolean gestureRecognition = true;
    }

    // === AR Settings ===

    public static class ARSettings {
        public boolean arEnabled = false;
        public String arPlatform = "arcore";
        public String arSessionType = "world";

        public boolean worldTracking = true;
        public boolean planeDetection = true;
        public boolean objectDetection = false;
        public boolean imageTracking = false;
        public boolean faceTracking = false;
        public boolean bodyTracking = false;

        public boolean occlusionEnabled = true;
        public String occlusionQuality = "medium";
        public boolean lightEstimation = true;
        public boolean shadowCasting = true;

        public boolean arPhysicsEnabled = false;
        public String collisionDetection = "basic";

        public boolean environmentProbe = false;
        public boolean reflectionProbe = false;
    }

    // === Performance Scaling ===

    public static class PerformanceScaling {
        public String scalingMode = "dynamic";
        public double scalingAggression = 0.5;

        public List<QualityTier> qualityTiers = new ArrayList<>();
        public int currentTier = 1;

        public int fpsSampleWindow = 60;
        public int fpsTargetLow = 30;
        public int fpsTargetHigh = 60;

        public int qualityUpDelay = 300;
        public int qualityDownDelay = 60;
    }

    // === Quality Tier ===

    public static class QualityTier {
        public String tierName;
        public double resolution;
        public String textureQuality;
        public String shadowQuality;
        public String effectQuality;
        public double drawDistance;
        public int particleCount;
    }

    // === Battery Optimization ===

    public static class BatteryOptimization {
        public String optimizationLevel = "medium";

        public double highBatteryThreshold = 0.7;
        public double mediumBatteryThreshold = 0.4;
        public double lowBatteryThreshold = 0.2;

        public String highBatteryQuality = "high";
        public String mediumBatteryQuality = "medium";
        public String lowBatteryQuality = "low";
        public String criticalBatteryQuality = "minimal";

        public boolean reduceAnimations = true;
        public boolean reduceParticles = true;
        public boolean reducePhysics = false;
        public boolean reduceShadows = true;
        public boolean reducePostProcess = true;

        public boolean boostOnCharging = true;
        public String chargingQuality = "ultra";
    }

    // === Display Sync ===

    public static class DisplaySync {
        public String syncMode = "state_sync";
        public int syncFrequency = 100;
        public String syncPriority = "high";

        public boolean stateSyncEnabled = true;
        public List<String> syncedStates = new ArrayList<>();
        public String interpolationMode = "linear";
        public boolean predictionEnabled = true;
        public int predictionSteps = 2;

        public String conflictResolution = "timestamp";
        public String masterDevice = "";

        public String networkProtocol = "websocket";
        public boolean compressionEnabled = true;
        public int compressionLevel = 6;
        public boolean deltaEncoding = true;

        public boolean latencyCompensation = true;
        public int maxLatency = 200;
        public int bufferSize = 3;
        public int interpolationBuffer = 2;

        public String offlineMode = "hybrid";
        public int reconnectAttempts = 5;
        public int reconnectDelay = 1000;
        public boolean stateRecovery = true;

        public boolean isFullSync() {
            return "full_sync".equals(syncMode);
        }

        public boolean isStateSyncEnabled() {
            return stateSyncEnabled;
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("sync_mode", syncMode);
                json.put("sync_frequency", syncFrequency);
                json.put("sync_priority", syncPriority);
                json.put("state_sync_enabled", stateSyncEnabled);
                json.put("interpolation_mode", interpolationMode);
                json.put("prediction_enabled", predictionEnabled);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static DisplaySync fromJson(@NonNull JSONObject json) throws JSONException {
            DisplaySync sync = new DisplaySync();
            sync.syncMode = json.optString("sync_mode", "state_sync");
            sync.syncFrequency = json.optInt("sync_frequency", 100);
            sync.syncPriority = json.optString("sync_priority", "high");
            sync.stateSyncEnabled = json.optBoolean("state_sync_enabled", true);
            sync.interpolationMode = json.optString("interpolation_mode", "linear");
            sync.predictionEnabled = json.optBoolean("prediction_enabled", true);
            return sync;
        }
    }

    // === Scene Render Profile ===

    public static class SceneRenderProfile {
        public String sceneName;
        public String sceneType;

        public String cameraMode = "orbit";
        public double cameraFOV = 60.0;
        public double cameraNearClip = 0.1;
        public double cameraFarClip = 1000.0;
        public Position3D cameraPosition = new Position3D();
        public Position3D cameraTarget = new Position3D();

        public String environmentLight = "neutral";
        public double sunIntensity = 1.0;
        public String sunColor = "#ffffff";
        public String ambientColor = "#404040";
        public boolean fogEnabled = false;
        public String fogColor = "#cccccc";
        public double fogDensity = 0.01;

        public String backgroundType = "solid";
        public String backgroundColor = "#1a1a2e";
        public String skyboxPath;
        public String backgroundImage;

        public Position3D avatarPosition = new Position3D();
        public Rotation3D avatarRotation = new Rotation3D();
        public double avatarScale = 1.0;
        public boolean avatarVisible = true;
        public boolean avatarShadow = true;

        public String defaultAnimation = "idle";
        public List<String> idleAnimations = new ArrayList<>();
        public double transitionTime = 0.3;

        public String qualityOverride;
        public boolean enabled = true;

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("scene_name", sceneName);
                json.put("scene_type", sceneType);
                json.put("camera_mode", cameraMode);
                json.put("camera_fov", cameraFOV);
                json.put("background_type", backgroundType);
                json.put("enabled", enabled);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static SceneRenderProfile fromJson(@NonNull JSONObject json) throws JSONException {
            SceneRenderProfile profile = new SceneRenderProfile();
            profile.sceneName = json.optString("scene_name", "");
            profile.sceneType = json.optString("scene_type", "");
            profile.cameraMode = json.optString("camera_mode", "orbit");
            profile.cameraFOV = json.optDouble("camera_fov", 60.0);
            profile.backgroundType = json.optString("background_type", "solid");
            profile.enabled = json.optBoolean("enabled", true);
            return profile;
        }
    }

    // === Position3D ===

    public static class Position3D {
        public double x = 0.0;
        public double y = 0.0;
        public double z = 0.0;

        public Position3D() {}

        public Position3D(double x, double y, double z) {
            this.x = x;
            this.y = y;
            this.z = z;
        }
    }

    // === Rotation3D ===

    public static class Rotation3D {
        public double x = 0.0;
        public double y = 0.0;
        public double z = 0.0;

        public Rotation3D() {}

        public Rotation3D(double x, double y, double z) {
            this.x = x;
            this.y = y;
            this.z = z;
        }
    }

    // === Display Context ===

    public static class DisplayContext {
        public String identityId;

        public String currentDevice;
        public String deviceType;
        public Resolution screenResolution = new Resolution();
        public String displayMode;

        public String currentQuality = "medium";
        public double currentFPS = 60.0;
        public double frameTime = 16.67;
        public int drawCalls;
        public long memoryUsage;
        public long textureMemory;
        public double gpuUtilization;

        public double averageFPS = 60.0;
        public double minFPS = 60.0;
        public double maxFPS = 60.0;
        public int droppedFrames;
        public String thermalState = "normal";
        public double batteryLevel = 1.0;
        public String batteryState = "unknown";

        public String currentScene;
        public double sceneLoadTime;
        public int assetsLoaded;
        public int assetsTotal;

        public String syncStatus = "synced";
        public long lastSyncTime;
        public int syncLatency;
        public List<String> connectedDevices = new ArrayList<>();

        public String recommendedQuality = "medium";
        public String recommendedDevice;
        public boolean qualityAdjustmentNeeded;

        public long timestamp;

        public boolean isSynced() {
            return "synced".equals(syncStatus);
        }

        public boolean isOffline() {
            return "offline".equals(syncStatus);
        }

        public boolean isThermalHot() {
            return "hot".equals(thermalState);
        }

        public boolean isLowBattery() {
            return batteryLevel < 0.2;
        }

        public boolean isCharging() {
            return "charging".equals(batteryState) || "full".equals(batteryState);
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("identity_id", identityId);
                json.put("current_device", currentDevice);
                json.put("device_type", deviceType);
                json.put("current_quality", currentQuality);
                json.put("current_fps", currentFPS);
                json.put("sync_status", syncStatus);
                json.put("battery_level", batteryLevel);
                json.put("thermal_state", thermalState);
                json.put("recommended_quality", recommendedQuality);
                json.put("quality_adjustment_needed", qualityAdjustmentNeeded);
                json.put("timestamp", timestamp);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static DisplayContext fromJson(@NonNull JSONObject json) throws JSONException {
            DisplayContext context = new DisplayContext();
            context.identityId = json.optString("identity_id", "");
            context.currentDevice = json.optString("current_device", "");
            context.deviceType = json.optString("device_type", "phone");
            context.currentQuality = json.optString("current_quality", "medium");
            context.currentFPS = json.optDouble("current_fps", 60.0);
            context.syncStatus = json.optString("sync_status", "synced");
            context.batteryLevel = json.optDouble("battery_level", 1.0);
            context.thermalState = json.optString("thermal_state", "normal");
            context.recommendedQuality = json.optString("recommended_quality", "medium");
            context.qualityAdjustmentNeeded = json.optBoolean("quality_adjustment_needed", false);
            context.timestamp = json.optLong("timestamp", System.currentTimeMillis());
            return context;
        }
    }

    // === Resolution ===

    public static class Resolution {
        public int width;
        public int height;
        public double scale = 1.0;

        public int getPixelWidth() {
            return (int) (width * scale);
        }

        public int getPixelHeight() {
            return (int) (height * scale);
        }
    }

    // === JSON Serialization ===

    @NonNull
    public JSONObject toJson() {
        JSONObject json = new JSONObject();
        try {
            json.put("rendering_settings", renderingSettings.toJson());
            json.put("device_adaptation", deviceAdaptation.toJson());
            json.put("display_sync", displaySync.toJson());
            json.put("context", context.toJson());
        } catch (JSONException e) {
            // ignore
        }
        return json;
    }

    @NonNull
    public static MultiDisplayState fromJson(@NonNull JSONObject json) throws JSONException {
        MultiDisplayState state = new MultiDisplayState();
        JSONObject renderJson = json.optJSONObject("rendering_settings");
        if (renderJson != null) {
            state.renderingSettings = RenderingSettings.fromJson(renderJson);
        }
        JSONObject adaptJson = json.optJSONObject("device_adaptation");
        if (adaptJson != null) {
            state.deviceAdaptation = DeviceAdaptation.fromJson(adaptJson);
        }
        JSONObject syncJson = json.optJSONObject("display_sync");
        if (syncJson != null) {
            state.displaySync = DisplaySync.fromJson(syncJson);
        }
        JSONObject contextJson = json.optJSONObject("context");
        if (contextJson != null) {
            state.context = DisplayContext.fromJson(contextJson);
        }
        return state;
    }
}