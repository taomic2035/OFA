package com.ofa.agent.distributed;

import android.content.Context;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * Scene Context - represents the current user context/scene.
 *
 * Scenes include:
 * - Running, Walking, Cycling (physical activities)
 * - Driving, Commuting (transportation)
 * - Working, Meeting, Focus (work context)
 * - Sleeping, Resting (rest context)
 * - Home, Office, Outdoor (location context)
 *
 * Detected from:
 * - Wearable sensors (heart rate, motion)
 * - Phone sensors (GPS, activity recognition)
 * - Time of day
 * - User calendar/events
 */
public class SceneContext {

    // ===== Scene Types =====

    public static final String RUNNING = "running";           // 跑步
    public static final String WALKING = "walking";           // 步行
    public static final String CYCLING = "cycling";           // 骑行
    public static final String DRIVING = "driving";           // 驾驶
    public static final String COMMUTING = "commuting";       // 通勤中
    public static final String WORKING = "working";           // 工作
    public static final String MEETING = "meeting";           // 会议中
    public static final String FOCUS = "focus";               // 专注模式
    public static final String SLEEPING = "sleeping";         // 睡眠
    public static final String RESTING = "resting";           // 休息
    public static final String HOME = "home";                 // 在家
    public static final String OFFICE = "office";             // 在办公室
    public static final String OUTDOOR = "outdoor";           // 户外
    public static final String UNKNOWN = "unknown";           // 未知

    // ===== Scene Properties =====

    private final String sceneType;
    private final float confidence;       // 0-1, detection confidence
    private final long timestamp;         // detection time
    private final String sourceDeviceId;  // which device detected this
    private final Map<String, Object> metadata; // additional context data

    // Derived properties
    private final boolean isPhysicalActivity;  // 是否体力活动
    private final boolean isQuietContext;      // 是否安静场景
    private final boolean requiresMinimalUI;   // 是否需要最小UI干扰
    private final String preferredDisplayDevice; // 推荐显示设备

    public SceneContext(@NonNull String sceneType, float confidence, long timestamp,
                        @Nullable String sourceDeviceId, @NonNull Map<String, Object> metadata) {
        this.sceneType = sceneType;
        this.confidence = confidence;
        this.timestamp = timestamp;
        this.sourceDeviceId = sourceDeviceId;
        this.metadata = new ConcurrentHashMap<>(metadata);

        // Derive properties based on scene type
        this.isPhysicalActivity = derivePhysicalActivity(sceneType);
        this.isQuietContext = deriveQuietContext(sceneType);
        this.requiresMinimalUI = deriveMinimalUI(sceneType);
        this.preferredDisplayDevice = derivePreferredDisplay(sceneType);
    }

    // ===== Getters =====

    @NonNull
    public String getSceneType() { return sceneType; }
    public float getConfidence() { return confidence; }
    public long getTimestamp() { return timestamp; }
    @Nullable
    public String getSourceDeviceId() { return sourceDeviceId; }
    @NonNull
    public Map<String, Object> getMetadata() { return metadata; }

    public boolean isPhysicalActivity() { return isPhysicalActivity; }
    public boolean isQuietContext() { return isQuietContext; }
    public boolean requiresMinimalUI() { return requiresMinimalUI; }
    @Nullable
    public String getPreferredDisplayDevice() { return preferredDisplayDevice; }

    // ===== Scene Derivation =====

    private boolean derivePhysicalActivity(@NonNull String type) {
        return type.equals(RUNNING) || type.equals(WALKING) ||
               type.equals(CYCLING) || type.equals(COMMUTING);
    }

    private boolean deriveQuietContext(@NonNull String type) {
        return type.equals(MEETING) || type.equals(FOCUS) ||
               type.equals(SLEEPING) || type.equals(RESTING);
    }

    private boolean deriveMinimalUI(@NonNull String type) {
        // Physical activities and meetings need minimal UI interruption
        return type.equals(RUNNING) || type.equals(CYCLING) ||
               type.equals(MEETING) || type.equals(FOCUS) ||
               type.equals(SLEEPING);
    }

    @Nullable
    private String derivePreferredDisplay(@NonNull String type) {
        // Physical activities prefer wearable display
        if (type.equals(RUNNING) || type.equals(CYCLING) || type.equals(WALKING)) {
            return "watch"; // Prefer watch for quick glances
        }
        // Sleep prefers no display or gentle notification
        if (type.equals(SLEEPING)) {
            return "none"; // Or gentle vibration
        }
        // Default to phone
        return "phone";
    }

    // ===== Scene Categories =====

    /**
     * Get scene category for routing decisions
     */
    @NonNull
    public String getSceneCategory() {
        if (isPhysicalActivity) return "physical_activity";
        if (isQuietContext) return "quiet_context";
        if (sceneType.equals(HOME) || sceneType.equals(OFFICE)) return "location";
        return "general";
    }

    /**
     * Get recommended notification style for this scene
     */
    @NonNull
    public NotificationStyle getNotificationStyle() {
        if (sceneType.equals(RUNNING) || sceneType.equals(CYCLING)) {
            return NotificationStyle.MINIMAL; // Quick glance only
        }
        if (sceneType.equals(MEETING) || sceneType.equals(FOCUS)) {
            return NotificationStyle.SILENT; // No sound, just vibration
        }
        if (sceneType.equals(SLEEPING)) {
            return NotificationStyle.NONE; // No notification
        }
        if (sceneType.equals(DRIVING)) {
            return NotificationStyle.VOICE; // Voice announcement
        }
        return NotificationStyle.NORMAL;
    }

    /**
     * Notification style enum
     */
    public enum NotificationStyle {
        NONE,       // 不通知
        MINIMAL,    // 最小化（手表闪烁）
        SILENT,     // 静音振动
        VOICE,      // 语音播报
        NORMAL      // 正常通知
    }

    /**
     * Check if scene is valid (recent enough)
     */
    public boolean isValid(long maxAgeMs) {
        return System.currentTimeMillis() - timestamp < maxAgeMs;
    }

    /**
     * Create unknown scene
     */
    @NonNull
    public static SceneContext unknown() {
        return new SceneContext(UNKNOWN, 0f, System.currentTimeMillis(),
            null, new ConcurrentHashMap<>());
    }

    @NonNull
    @Override
    public String toString() {
        return String.format("SceneContext{type=%s, confidence=%.2f, category=%s, preferred=%s}",
            sceneType, confidence, getSceneCategory(), preferredDisplayDevice);
    }
}