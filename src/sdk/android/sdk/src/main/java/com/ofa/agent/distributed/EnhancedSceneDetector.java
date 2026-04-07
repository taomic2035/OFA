package com.ofa.agent.distributed;

import android.content.Context;
import android.location.Location;
import android.os.Handler;
import android.os.Looper;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.state.DeviceLocation;
import com.ofa.agent.state.DeviceState;
import com.ofa.agent.state.StateChange;
import com.ofa.agent.state.StateSyncService;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.Calendar;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.CopyOnWriteArrayList;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * 增强版场景检测器 (v3.2.0)
 *
 * 支持更多场景类型：
 * - 运动场景：running, walking, cycling, swimming
 * - 工作场景：meeting, working, driving
 * - 生活场景：cooking, gaming, sleeping, resting
 *
 * 场景检测源：
 * - 传感器数据（加速度、心率、步数）
 * - 时间信息（时段、日程）
 * - 位置信息（家、公司、户外）
 * - 网络状态
 * - 活跃应用
 */
public class EnhancedSceneDetector {

    private static final String TAG = "EnhancedSceneDetector";

    // 场景类型常量
    public static final String SCENE_UNKNOWN = "unknown";
    public static final String SCENE_IDLE = "idle";
    public static final String SCENE_RUNNING = "running";
    public static final String SCENE_WALKING = "walking";
    public static final String SCENE_DRIVING = "driving";
    public static final String SCENE_CYCLING = "cycling";
    public static final String SCENE_MEETING = "meeting";
    public static final String SCENE_WORKING = "working";
    public static final String SCENE_RESTING = "resting";
    public static final String SCENE_SLEEPING = "sleeping";
    public static final String SCENE_GAMING = "gaming";
    public static final String SCENE_COOKING = "cooking";
    public static final String SCENE_SWIMMING = "swimming";

    // 消息类型常量（用于路由）
    public static final String MSG_TYPE_COMMAND = "command";
    public static final String MSG_TYPE_NOTIFICATION = "notification";
    public static final String MSG_TYPE_DATA = "data";
    public static final String MSG_TYPE_SYNC = "sync";
    public static final String MSG_TYPE_ALERT = "alert";
    public static final String MSG_TYPE_HEALTH = "health";
    public static final String MSG_TYPE_SOCIAL = "social";
    public static final String MSG_TYPE_SYSTEM = "system";

    private final Context context;
    private final ExecutorService executor;
    private final Handler mainHandler;

    // 基础场景检测器
    private SceneDetector baseDetector;

    // 状态同步服务
    private StateSyncService stateSyncService;

    // 当前场景
    private SceneContext currentScene;

    // 场景历史
    private final List<SceneRecord> sceneHistory;

    // 监听器
    private final List<SceneChangeListener> listeners;

    // 检测配置
    private SceneDetectionConfig config;

    // 位置信息
    private DeviceLocation homeLocation;
    private DeviceLocation workLocation;

    // 时段规则
    private final List<TimeRule> timeRules;

    // 活跃应用检测
    private final List<String> gamingApps;
    private final List<String> workApps;

    // 检测状态
    private boolean detecting = false;

    /**
     * 场景变更监听器
     */
    public interface SceneChangeListener {
        void onSceneChanged(@NonNull SceneContext oldScene, @NonNull SceneContext newScene);
        void onSceneConfidenceUpdated(@NonNull String scene, float confidence);
    }

    /**
     * 场景检测配置
     */
    public static class SceneDetectionConfig {
        public long detectionIntervalMs = 5000;
        public float runningHeartRateThreshold = 120f;
        public float walkingHeartRateThreshold = 80f;
        public int historySize = 100;
        public boolean enableLocationBasedDetection = true;
        public boolean enableTimeBasedDetection = true;
        public boolean enableAppBasedDetection = true;
    }

    /**
     * 场景记录
     */
    public static class SceneRecord {
        public String scene;
        public float confidence;
        public long timestamp;
        public Map<String, Object> metadata;

        public SceneRecord(String scene, float confidence, Map<String, Object> metadata) {
            this.scene = scene;
            this.confidence = confidence;
            this.timestamp = System.currentTimeMillis();
            this.metadata = metadata != null ? metadata : new HashMap<>();
        }
    }

    /**
     * 时段规则
     */
    public static class TimeRule {
        public int startHour;
        public int startMinute;
        public int endHour;
        public int endMinute;
        public String scene;
        public float confidence;
        public int[] daysOfWeek; // 0=Sunday, 6=Saturday

        public TimeRule(int startHour, int startMinute, int endHour, int endMinute,
                        String scene, float confidence, int[] daysOfWeek) {
            this.startHour = startHour;
            this.startMinute = startMinute;
            this.endHour = endHour;
            this.endMinute = endMinute;
            this.scene = scene;
            this.confidence = confidence;
            this.daysOfWeek = daysOfWeek;
        }
    }

    public EnhancedSceneDetector(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.executor = Executors.newSingleThreadExecutor();
        this.mainHandler = new Handler(Looper.getMainLooper());
        this.sceneHistory = new ArrayList<>();
        this.listeners = new CopyOnWriteArrayList<>();
        this.timeRules = new ArrayList<>();
        this.config = new SceneDetectionConfig();

        // 初始化基础检测器
        this.baseDetector = new SceneDetector(context);

        // 初始化应用列表
        this.gamingApps = initGamingApps();
        this.workApps = initWorkApps();

        // 初始化时段规则
        initTimeRules();

        // 初始化当前场景
        this.currentScene = new SceneContext(SCENE_UNKNOWN, 0f, System.currentTimeMillis(), null, null);
    }

    /**
     * 设置状态同步服务
     */
    public void setStateSyncService(@Nullable StateSyncService service) {
        this.stateSyncService = service;

        if (service != null) {
            // 监听状态变更以更新场景
            service.addListener(new StateSyncService.StateListener() {
                @Override
                public void onLocalStateChanged(StateChange change) {
                    if (change.isSceneChange()) {
                        handleSceneStateChange(change);
                    }
                }

                @Override
                public void onRemoteStateChanged(StateChange change) {
                    // 其他设备场景变更可能影响路由决策
                }

                @Override
                public void onDeviceOnline(DeviceState state) {}

                @Override
                public void onDeviceOffline(String agentId) {}
            });
        }
    }

    /**
     * 设置配置
     */
    public void setConfig(@NonNull SceneDetectionConfig config) {
        this.config = config;
    }

    /**
     * 设置家位置
     */
    public void setHomeLocation(@Nullable DeviceLocation location) {
        this.homeLocation = location;
    }

    /**
     * 设置工作位置
     */
    public void setWorkLocation(@Nullable DeviceLocation location) {
        this.workLocation = location;
    }

    /**
     * 开始检测
     */
    public void startDetection() {
        if (detecting) {
            return;
        }

        detecting = true;

        // 启动基础检测器
        baseDetector.startDetection();

        // 添加监听器
        baseDetector.addListener(new SceneDetector.SceneChangeListener() {
            @Override
            public void onSceneChanged(SceneContext oldScene, SceneContext newScene) {
                handleBaseSceneChange(oldScene, newScene);
            }
        });

        // 启动增强检测
        startEnhancedDetection();

        Log.i(TAG, "Enhanced scene detection started");
    }

    /**
     * 停止检测
     */
    public void stopDetection() {
        if (!detecting) {
            return;
        }

        detecting = false;

        baseDetector.stopDetection();

        Log.i(TAG, "Enhanced scene detection stopped");
    }

    /**
     * 获取当前场景
     */
    @NonNull
    public SceneContext getCurrentScene() {
        return currentScene;
    }

    /**
     * 获取场景历史
     */
    @NonNull
    public List<SceneRecord> getSceneHistory(int limit) {
        if (limit <= 0 || limit >= sceneHistory.size()) {
            return new ArrayList<>(sceneHistory);
        }
        return new ArrayList<>(sceneHistory.subList(
                Math.max(0, sceneHistory.size() - limit), sceneHistory.size()));
    }

    /**
     * 添加场景监听器
     */
    public void addListener(@NonNull SceneChangeListener listener) {
        listeners.add(listener);
    }

    /**
     * 移除场景监听器
     */
    public void removeListener(@NonNull SceneChangeListener listener) {
        listeners.remove(listener);
    }

    // === 增强检测逻辑 ===

    private void startEnhancedDetection() {
        executor.execute(() -> {
            while (detecting) {
                try {
                    analyzeScene();
                    Thread.sleep(config.detectionIntervalMs);
                } catch (InterruptedException e) {
                    break;
                }
            }
        });
    }

    private void analyzeScene() {
        SceneContext baseScene = baseDetector.getCurrentScene();

        // 收集上下文信息
        Map<String, Object> metadata = new HashMap<>();
        metadata.put("base_scene", baseScene.getSceneType());
        metadata.put("base_confidence", baseScene.getConfidence());

        // 时段检测
        String timeScene = detectTimeBasedScene();
        float timeConfidence = 0.5f;
        if (!timeScene.equals(SCENE_UNKNOWN)) {
            metadata.put("time_scene", timeScene);
            metadata.put("time_confidence", timeConfidence);
        }

        // 位置检测
        String locationScene = SCENE_UNKNOWN;
        if (config.enableLocationBasedDetection && stateSyncService != null) {
            DeviceState state = stateSyncService.getCurrentState();
            if (state != null && state.getLocation() != null) {
                locationScene = detectLocationBasedScene(state.getLocation());
                metadata.put("location_scene", locationScene);
            }
        }

        // 应用检测
        String appScene = SCENE_UNKNOWN;
        if (config.enableAppBasedDetection) {
            // 检测活跃应用（需要特殊权限）
            // appScene = detectAppBasedScene();
            metadata.put("app_scene", appScene);
        }

        // 综合场景决策
        SceneContext newScene = decideFinalScene(
                baseScene.getSceneType(), baseScene.getConfidence(),
                timeScene, timeConfidence,
                locationScene, appScene,
                metadata);

        // 检查场景变更
        if (!newScene.getSceneType().equals(currentScene.getSceneType())) {
            notifySceneChange(currentScene, newScene);
        }

        currentScene = newScene;

        // 记录历史
        recordScene(newScene);

        // 同步到状态服务
        if (stateSyncService != null) {
            stateSyncService.updateScene(newScene.getSceneType(), metadata);
        }
    }

    // === 时段检测 ===

    private String detectTimeBasedScene() {
        if (!config.enableTimeBasedDetection) {
            return SCENE_UNKNOWN;
        }

        Calendar cal = Calendar.getInstance();
        int hour = cal.get(Calendar.HOUR_OF_DAY);
        int minute = cal.get(Calendar.MINUTE);
        int dayOfWeek = cal.get(Calendar.DAY_OF_WEEK) - 1; // 0=Sunday

        int currentTime = hour * 60 + minute;

        for (TimeRule rule : timeRules) {
            // 检查星期
            if (rule.daysOfWeek != null && rule.daysOfWeek.length > 0) {
                boolean dayMatch = false;
                for (int d : rule.daysOfWeek) {
                    if (d == dayOfWeek) {
                        dayMatch = true;
                        break;
                    }
                }
                if (!dayMatch) {
                    continue;
                }
            }

            // 检查时间
            int startTime = rule.startHour * 60 + rule.startMinute;
            int endTime = rule.endHour * 60 + rule.endMinute;

            if (currentTime >= startTime && currentTime <= endTime) {
                return rule.scene;
            }
        }

        return SCENE_UNKNOWN;
    }

    // === 位置检测 ===

    private String detectLocationBasedScene(@NonNull DeviceLocation location) {
        if (homeLocation != null) {
            double distance = location.distanceTo(homeLocation);
            if (distance < 100) { // 100米内
                return SCENE_RESTING; // 在家
            }
        }

        if (workLocation != null) {
            double distance = location.distanceTo(workLocation);
            if (distance < 100) {
                return SCENE_WORKING; // 在公司
            }
        }

        return SCENE_UNKNOWN;
    }

    // === 场景决策 ===

    private SceneContext decideFinalScene(
            String baseScene, float baseConfidence,
            String timeScene, float timeConfidence,
            String locationScene, String appScene,
            Map<String, Object> metadata) {

        // 优先级：运动场景 > 工作场景 > 时段场景 > 位置场景

        // 运动场景优先
        if (isMotionScene(baseScene) && baseConfidence > 0.6f) {
            return new SceneContext(baseScene, baseConfidence, System.currentTimeMillis(), null, metadata);
        }

        // 应用场景
        if (!appScene.equals(SCENE_UNKNOWN)) {
            return new SceneContext(appScene, 0.8f, System.currentTimeMillis(), null, metadata);
        }

        // 位置 + 时段组合
        if (locationScene.equals(SCENE_WORKING) && timeScene.equals(SCENE_MEETING)) {
            return new SceneContext(SCENE_MEETING, 0.7f, System.currentTimeMillis(), null, metadata);
        }

        if (locationScene.equals(SCENE_WORKING) && timeScene.equals(SCENE_WORKING)) {
            return new SceneContext(SCENE_WORKING, 0.8f, System.currentTimeMillis(), null, metadata);
        }

        // 睡眠检测
        if (timeScene.equals(SCENE_SLEEPING)) {
            return new SceneContext(SCENE_SLEEPING, 0.7f, System.currentTimeMillis(), null, metadata);
        }

        // 位置场景
        if (!locationScene.equals(SCENE_UNKNOWN)) {
            return new SceneContext(locationScene, 0.6f, System.currentTimeMillis(), null, metadata);
        }

        // 时段场景
        if (!timeScene.equals(SCENE_UNKNOWN)) {
            return new SceneContext(timeScene, timeConfidence, System.currentTimeMillis(), null, metadata);
        }

        // 回退到基础场景
        if (!baseScene.equals(SCENE_UNKNOWN)) {
            return new SceneContext(baseScene, baseConfidence, System.currentTimeMillis(), null, metadata);
        }

        // 默认空闲
        return new SceneContext(SCENE_IDLE, 0.5f, System.currentTimeMillis(), null, metadata);
    }

    private boolean isMotionScene(String scene) {
        return scene.equals(SCENE_RUNNING) ||
                scene.equals(SCENE_WALKING) ||
                scene.equals(SCENE_CYCLING) ||
                scene.equals(SCENE_SWIMMING) ||
                scene.equals(SCENE_DRIVING);
    }

    // === 初始化 ===

    private void initTimeRules() {
        // 工作时间 (周一到周五 9:00-18:00)
        timeRules.add(new TimeRule(9, 0, 18, 0, SCENE_WORKING, 0.6f, new int[]{1, 2, 3, 4, 5}));

        // 午休时间 (周一到周五 12:00-14:00)
        timeRules.add(new TimeRule(12, 0, 14, 0, SCENE_RESTING, 0.5f, new int[]{1, 2, 3, 4, 5}));

        // 会议时间示例 (周一到周五 10:00-11:00)
        timeRules.add(new TimeRule(10, 0, 11, 0, SCENE_MEETING, 0.4f, new int[]{1, 2, 3, 4, 5}));

        // 睡眠时间 (每天 23:00-7:00)
        timeRules.add(new TimeRule(23, 0, 23, 59, SCENE_SLEEPING, 0.7f, null));
        timeRules.add(new TimeRule(0, 0, 7, 0, SCENE_SLEEPING, 0.7f, null));

        // 早晨 (每天 7:00-9:00)
        timeRules.add(new TimeRule(7, 0, 9, 0, SCENE_RESTING, 0.5f, null));
    }

    private List<String> initGamingApps() {
        List<String> apps = new ArrayList<>();
        apps.add("com.tencent.tmgp.sgame");      // 王者荣耀
        apps.add("com.tencent.tmgp.pubgmhd");    // 和平精英
        apps.add("com.tencent.tmgp.cod");        // 使命召唤
        apps.add("com.miHoYo.Yuanshen");         // 原神
        apps.add("com.netease.dwrg");            // 阴阳师
        return apps;
    }

    private List<String> initWorkApps() {
        List<String> apps = new ArrayList<>();
        apps.add("com.tencent.wework");          // 企业微信
        apps.add("com.alibaba.android.rimet");   // 钉钉
        apps.add("com.feeyo.vms.pro");           // 飞书
        apps.add("com.microsoft.teams");         // Teams
        apps.add("com.zoom.us");                 // Zoom
        return apps;
    }

    // === 事件处理 ===

    private void handleBaseSceneChange(SceneContext oldScene, SceneContext newScene) {
        // 基础检测器检测到场景变更，触发增强分析
        executor.execute(this::analyzeScene);
    }

    private void handleSceneStateChange(StateChange change) {
        // 状态服务报告的场景变更
        if (change.getNewState() != null) {
            String scene = change.getNewState().getScene();
            if (!scene.equals(currentScene.getSceneType())) {
                SceneContext newScene = new SceneContext(
                        scene, 0.8f, System.currentTimeMillis(), null, null);
                notifySceneChange(currentScene, newScene);
                currentScene = newScene;
            }
        }
    }

    // === 通知 ===

    private void notifySceneChange(SceneContext oldScene, SceneContext newScene) {
        Log.i(TAG, "Scene changed: " + oldScene.getSceneType() + " -> " + newScene.getSceneType());

        mainHandler.post(() -> {
            for (SceneChangeListener l : listeners) {
                l.onSceneChanged(oldScene, newScene);
            }
        });

        recordScene(newScene);
    }

    private void recordScene(SceneContext scene) {
        sceneHistory.add(new SceneRecord(
                scene.getSceneType(), scene.getConfidence(), scene.getMetadata()));

        // 限制历史大小
        while (sceneHistory.size() > config.historySize) {
            sceneHistory.remove(0);
        }
    }

    // === 场景查询 ===

    /**
     * 检查是否在运动场景
     */
    public boolean isInMotionScene() {
        return isMotionScene(currentScene.getSceneType());
    }

    /**
     * 检查是否在睡眠场景
     */
    public boolean isSleeping() {
        return currentScene.getSceneType().equals(SCENE_SLEEPING);
    }

    /**
     * 检查是否在工作场景
     */
    public boolean isWorking() {
        String scene = currentScene.getSceneType();
        return scene.equals(SCENE_WORKING) || scene.equals(SCENE_MEETING);
    }

    /**
     * 检查是否在驾驶场景
     */
    public boolean isDriving() {
        return currentScene.getSceneType().equals(SCENE_DRIVING);
    }

    /**
     * 检查是否在游戏场景
     */
    public boolean isGaming() {
        return currentScene.getSceneType().equals(SCENE_GAMING);
    }

    /**
     * 获取场景置信度
     */
    public float getConfidence() {
        return currentScene.getConfidence();
    }

    /**
     * 获取场景持续时间（毫秒）
     */
    public long getSceneDuration() {
        return System.currentTimeMillis() - currentScene.getTimestamp();
    }

    /**
     * 获取场景统计
     */
    @NonNull
    public SceneStats getStats() {
        SceneStats stats = new SceneStats();
        stats.currentScene = currentScene.getSceneType();
        stats.confidence = currentScene.getConfidence();
        stats.duration = getSceneDuration();
        stats.historyCount = sceneHistory.size();

        // 统计场景分布
        for (SceneRecord record : sceneHistory) {
            Integer count = stats.sceneDistribution.get(record.scene);
            stats.sceneDistribution.put(record.scene, count == null ? 1 : count + 1);
        }

        return stats;
    }

    /**
     * 场景统计信息
     */
    public static class SceneStats {
        public String currentScene;
        public float confidence;
        public long duration;
        public int historyCount;
        public Map<String, Integer> sceneDistribution = new HashMap<>();

        @NonNull
        @Override
        public String toString() {
            return "SceneStats{" +
                    "current='" + currentScene + '\'' +
                    ", confidence=" + confidence +
                    ", duration=" + duration +
                    '}';
        }
    }

    /**
     * 清理资源
     */
    public void cleanup() {
        stopDetection();
        listeners.clear();
        sceneHistory.clear();
    }
}