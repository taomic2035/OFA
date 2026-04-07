package com.ofa.agent.display;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.List;
import java.util.concurrent.CopyOnWriteArrayList;

/**
 * MultiDisplay状态客户端 (v5.5.0)
 *
 * 端侧接收 Center 推送的多端展示状态，用于渲染和显示管理。
 * 深层展示管理在 Center 端 DisplayEngine 完成。
 */
public class MultiDisplayClient {

    private static volatile MultiDisplayClient instance;

    // 当前状态
    private MultiDisplayState currentState;

    // 监听器
    private final CopyOnWriteArrayList<DisplayStateListener> listeners = new CopyOnWriteArrayList<>();

    // 配置
    private String centerAddress;
    private boolean syncEnabled = false;

    // 性能监控
    private FPSMonitor fpsMonitor;
    private MemoryMonitor memoryMonitor;

    private MultiDisplayClient() {
        this.currentState = new MultiDisplayState();
        this.fpsMonitor = new FPSMonitor();
        this.memoryMonitor = new MemoryMonitor();
    }

    /**
     * 获取单例实例
     */
    @NonNull
    public static MultiDisplayClient getInstance() {
        if (instance == null) {
            synchronized (MultiDisplayClient.class) {
                if (instance == null) {
                    instance = new MultiDisplayClient();
                }
            }
        }
        return instance;
    }

    /**
     * 初始化
     */
    public void initialize(@Nullable String centerAddress) {
        this.centerAddress = centerAddress;
        this.syncEnabled = centerAddress != null && !centerAddress.isEmpty();
    }

    // === 状态接收 ===

    /**
     * 接收 Center 推送的 MultiDisplay 状态
     */
    public void receiveDisplayState(@NonNull JSONObject stateJson) {
        try {
            MultiDisplayState newState = MultiDisplayState.fromJson(stateJson);
            updateState(newState);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 接收决策上下文
     */
    public void receiveDecisionContext(@NonNull JSONObject contextJson) {
        try {
            MultiDisplayState newState = MultiDisplayState.fromJson(contextJson);
            updateState(newState);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 更新状态
     */
    private void updateState(@NonNull MultiDisplayState newState) {
        this.currentState = newState;

        for (DisplayStateListener listener : listeners) {
            listener.onDisplayStateChanged(newState);
        }
    }

    // === 状态获取 ===

    @NonNull
    public MultiDisplayState getCurrentState() {
        return currentState;
    }

    // === Rendering Settings ===

    @NonNull
    public MultiDisplayState.RenderingSettings getRenderingSettings() {
        return currentState.getRenderingSettings();
    }

    public String getRenderEngine() {
        return currentState.getRenderingSettings().renderEngine;
    }

    public String getQualityPreset() {
        return currentState.getRenderingSettings().qualityPreset;
    }

    public String getRenderQuality() {
        return currentState.getRenderingSettings().renderQuality;
    }

    public int getTargetFPS() {
        return currentState.getRenderingSettings().targetFPS;
    }

    public boolean isAdaptiveQuality() {
        return currentState.getRenderingSettings().adaptiveQuality;
    }

    public boolean isPostProcessingEnabled() {
        return currentState.getRenderingSettings().isPostProcessingEnabled();
    }

    public boolean isPhysicsEnabled() {
        return currentState.getRenderingSettings().isPhysicsEnabled();
    }

    public boolean isHighQuality() {
        return currentState.getRenderingSettings().isHighQuality();
    }

    // === Device Adaptation ===

    @NonNull
    public MultiDisplayState.DeviceAdaptation getDeviceAdaptation() {
        return currentState.getDeviceAdaptation();
    }

    public boolean isAutoAdaptation() {
        return currentState.getDeviceAdaptation().isAutoAdaptation();
    }

    @NonNull
    public MultiDisplayState.MobileOptimizations getMobileOptimizations() {
        return currentState.getDeviceAdaptation().mobileOptimizations;
    }

    public boolean isBatteryAwareMode() {
        return currentState.getDeviceAdaptation().mobileOptimizations.batteryAwareMode;
    }

    public boolean isThermalThrottling() {
        return currentState.getDeviceAdaptation().mobileOptimizations.thermalThrottling;
    }

    // === Display Sync ===

    @NonNull
    public MultiDisplayState.DisplaySync getDisplaySync() {
        return currentState.getDisplaySync();
    }

    public boolean isStateSyncEnabled() {
        return currentState.getDisplaySync().isStateSyncEnabled();
    }

    public String getSyncMode() {
        return currentState.getDisplaySync().syncMode;
    }

    public int getSyncFrequency() {
        return currentState.getDisplaySync().syncFrequency;
    }

    public boolean isFullSync() {
        return currentState.getDisplaySync().isFullSync();
    }

    // === Context ===

    @NonNull
    public MultiDisplayState.DisplayContext getContext() {
        return currentState.getContext();
    }

    public String getCurrentQuality() {
        return currentState.getContext().currentQuality;
    }

    public double getCurrentFPS() {
        return currentState.getContext().currentFPS;
    }

    public String getSyncStatus() {
        return currentState.getContext().syncStatus;
    }

    public boolean isSynced() {
        return currentState.getContext().isSynced();
    }

    public boolean isOffline() {
        return currentState.getContext().isOffline();
    }

    public double getBatteryLevel() {
        return currentState.getContext().batteryLevel;
    }

    public boolean isLowBattery() {
        return currentState.getContext().isLowBattery();
    }

    public boolean isCharging() {
        return currentState.getContext().isCharging();
    }

    public String getThermalState() {
        return currentState.getContext().thermalState;
    }

    public boolean isThermalHot() {
        return currentState.getContext().isThermalHot();
    }

    public String getRecommendedQuality() {
        return currentState.getContext().recommendedQuality;
    }

    public boolean isQualityAdjustmentNeeded() {
        return currentState.getContext().qualityAdjustmentNeeded;
    }

    public List<String> getConnectedDevices() {
        return currentState.getContext().connectedDevices;
    }

    // === Scene Management ===

    public String getCurrentScene() {
        return currentState.getContext().currentScene;
    }

    @Nullable
    public MultiDisplayState.SceneRenderProfile getSceneProfile(String sceneName) {
        for (MultiDisplayState.SceneRenderProfile profile : currentState.getSceneRenderProfiles()) {
            if (profile.sceneName.equals(sceneName)) {
                return profile;
            }
        }
        return null;
    }

    // === 决策辅助 ===

    /**
     * 获取推荐的质量设置
     */
    @NonNull
    public String getRecommendedQualitySetting() {
        // 考虑电池状态
        if (isLowBattery() && !isCharging()) {
            return "low";
        }

        // 考虑温度状态
        if (isThermalHot()) {
            return "low";
        }

        // 返回推荐的 Quality
        String recommended = getRecommendedQuality();
        if (recommended != null && !recommended.isEmpty()) {
            return recommended;
        }

        // 默认中等质量
        return "medium";
    }

    /**
     * 获取渲染建议
     */
    @NonNull
    public String getRenderingRecommendation() {
        StringBuilder recommendation = new StringBuilder();

        String quality = getRecommendedQualitySetting();
        recommendation.append("推荐渲染质量: ").append(quality);

        if (isLowBattery()) {
            recommendation.append("\n原因: 低电量模式");
        } else if (isThermalHot()) {
            recommendation.append("\n原因: 设备过热");
        } else if (isQualityAdjustmentNeeded()) {
            recommendation.append("\n原因: 性能优化");
        }

        return recommendation.toString();
    }

    /**
     * 获取同步建议
     */
    @NonNull
    public String getSyncRecommendation() {
        StringBuilder recommendation = new StringBuilder();

        if (isSynced()) {
            recommendation.append("同步状态: 已同步");
            recommendation.append("\n延迟: ").append(currentState.getContext().syncLatency).append("ms");
        } else if (isOffline()) {
            recommendation.append("同步状态: 离线");
            recommendation.append("\n使用本地缓存数据");
        } else {
            recommendation.append("同步状态: 同步中...");
        }

        // 显示连接的设备
        List<String> devices = getConnectedDevices();
        if (!devices.isEmpty()) {
            recommendation.append("\n已连接设备: ").append(devices.size());
        }

        return recommendation.toString();
    }

    /**
     * 获取性能建议
     */
    @NonNull
    public String getPerformanceRecommendation() {
        StringBuilder recommendation = new StringBuilder();

        double fps = getCurrentFPS();
        int targetFPS = getTargetFPS();

        recommendation.append("当前FPS: ").append(String.format("%.1f", fps));
        recommendation.append(" (目标: ").append(targetFPS).append(")");

        if (fps < targetFPS * 0.8) {
            recommendation.append("\n建议: 降低渲染质量以提升帧率");
        } else if (fps >= targetFPS * 0.95) {
            recommendation.append("\n状态: 性能良好");
        } else {
            recommendation.append("\n状态: 性能正常");
        }

        // 内存使用
        long memory = currentState.getContext().memoryUsage;
        if (memory > 0) {
            recommendation.append("\n内存使用: ").append(memory / 1024 / 1024).append("MB");
        }

        return recommendation.toString();
    }

    /**
     * 获取场景渲染配置
     */
    @NonNull
    public String getSceneRenderConfig() {
        String scene = getCurrentScene();
        if (scene == null || scene.isEmpty()) {
            return "当前场景: 默认";
        }

        MultiDisplayState.SceneRenderProfile profile = getSceneProfile(scene);
        if (profile == null) {
            return "当前场景: " + scene;
        }

        StringBuilder config = new StringBuilder();
        config.append("场景: ").append(scene);
        config.append("\n相机模式: ").append(profile.cameraMode);
        config.append("\nFOV: ").append(profile.cameraFOV);
        config.append("\n背景: ").append(profile.backgroundType);

        return config.toString();
    }

    /**
     * 获取综合展示建议
     */
    @NonNull
    public String getComprehensiveAdvice() {
        StringBuilder advice = new StringBuilder();

        // 渲染建议
        advice.append(getRenderingRecommendation()).append("\n\n");

        // 性能建议
        advice.append(getPerformanceRecommendation()).append("\n\n");

        // 同步建议
        advice.append(getSyncRecommendation());

        return advice.toString();
    }

    // === 性能监控 ===

    /**
     * 开始 FPS 监控
     */
    public void startFPSMonitor() {
        fpsMonitor.start();
    }

    /**
     * 停止 FPS 监控
     */
    public void stopFPSMonitor() {
        fpsMonitor.stop();
    }

    /**
     * 记录帧
     */
    public void recordFrame() {
        fpsMonitor.recordFrame();
    }

    /**
     * 获取平均 FPS
     */
    public double getAverageFPS() {
        return fpsMonitor.getAverageFPS();
    }

    // === 行为上报 ===

    /**
     * 上报性能指标
     */
    @NonNull
    public JSONObject reportPerformanceMetrics(double fps, long memory, double gpuUtil) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "performance_metrics");
            report.put("fps", fps);
            report.put("memory_usage", memory);
            report.put("gpu_utilization", gpuUtil);
            report.put("quality", getCurrentQuality());
            report.put("battery_level", getBatteryLevel());
            report.put("thermal_state", getThermalState());
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报质量变化
     */
    @NonNull
    public JSONObject reportQualityChange(@NonNull String newQuality, @NonNull String reason) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "quality_change");
            report.put("new_quality", newQuality);
            report.put("previous_quality", getCurrentQuality());
            report.put("reason", reason);
            report.put("fps", getCurrentFPS());
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报设备连接
     */
    @NonNull
    public JSONObject reportDeviceConnection(@NonNull String deviceId, @NonNull String deviceType, boolean connected) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "device_connection");
            report.put("device_id", deviceId);
            report.put("device_type", deviceType);
            report.put("connected", connected);
            report.put("sync_status", getSyncStatus());
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报场景变化
     */
    @NonNull
    public JSONObject reportSceneChange(@NonNull String newScene) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "scene_change");
            report.put("new_scene", newScene);
            report.put("previous_scene", getCurrentScene());
            report.put("quality", getCurrentQuality());
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报同步状态
     */
    @NonNull
    public JSONObject reportSyncStatus(@NonNull String status, int latency) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "sync_status");
            report.put("status", status);
            report.put("latency", latency);
            report.put("connected_devices", getConnectedDevices().size());
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    // === 监听器管理 ===

    public void addListener(@NonNull DisplayStateListener listener) {
        listeners.add(listener);
    }

    public void removeListener(@NonNull DisplayStateListener listener) {
        listeners.remove(listener);
    }

    public void clearListeners() {
        listeners.clear();
    }

    /**
     * Display状态监听器
     */
    public interface DisplayStateListener {
        void onDisplayStateChanged(@NonNull MultiDisplayState state);
    }

    // === 状态快照 ===

    @NonNull
    public JSONObject getStateSnapshot() {
        return currentState.toJson();
    }

    public void restoreStateSnapshot(@NonNull JSONObject snapshot) {
        try {
            currentState = MultiDisplayState.fromJson(snapshot);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    public void reset() {
        currentState = new MultiDisplayState();

        for (DisplayStateListener listener : listeners) {
            listener.onDisplayStateChanged(currentState);
        }
    }

    @NonNull
    @Override
    public String toString() {
        return "MultiDisplayClient{" +
                "quality='" + getCurrentQuality() + '\'' +
                ", fps=" + getCurrentFPS() +
                ", sync='" + getSyncStatus() + '\'' +
                ", battery=" + getBatteryLevel() +
                ", thermal='" + getThermalState() + '\'' +
                '}';
    }

    // === FPS Monitor ===

    private static class FPSMonitor {
        private boolean running = false;
        private long frameCount = 0;
        private long startTime = 0;
        private double averageFPS = 60.0;

        void start() {
            running = true;
            frameCount = 0;
            startTime = System.currentTimeMillis();
        }

        void stop() {
            running = false;
        }

        void recordFrame() {
            if (running) {
                frameCount++;
            }
        }

        double getAverageFPS() {
            if (!running || frameCount == 0) {
                return averageFPS;
            }
            long elapsed = System.currentTimeMillis() - startTime;
            if (elapsed > 0) {
                averageFPS = (frameCount * 1000.0) / elapsed;
            }
            return averageFPS;
        }
    }

    // === Memory Monitor ===

    private static class MemoryMonitor {
        long getUsedMemory() {
            Runtime runtime = Runtime.getRuntime();
            return runtime.totalMemory() - runtime.freeMemory();
        }
    }
}