package com.ofa.agent.avatar;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.List;
import java.util.concurrent.CopyOnWriteArrayList;

/**
 * Avatar状态客户端 (v5.0.0)
 *
 * 端侧接收 Center 推送的 Avatar 状态，用于调整外在呈现。
 * 深层 Avatar 管理在 Center 端 AvatarEngine 完成。
 */
public class AvatarClient {

    private static volatile AvatarClient instance;

    // 当前状态
    private AvatarState currentState;

    // 监听器
    private final CopyOnWriteArrayList<AvatarStateListener> listeners = new CopyOnWriteArrayList<>();

    // 配置
    private String centerAddress;
    private boolean syncEnabled = false;

    private AvatarClient() {
        this.currentState = new AvatarState();
    }

    /**
     * 获取单例实例
     */
    @NonNull
    public static AvatarClient getInstance() {
        if (instance == null) {
            synchronized (AvatarClient.class) {
                if (instance == null) {
                    instance = new AvatarClient();
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
     * 接收 Center 推送的 Avatar 状态
     */
    public void receiveAvatarState(@NonNull JSONObject stateJson) {
        try {
            AvatarState newState = AvatarState.fromJson(stateJson);
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
            AvatarState newState = AvatarState.fromJson(contextJson);
            updateState(newState);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 更新状态
     */
    private void updateState(@NonNull AvatarState newState) {
        this.currentState = newState;

        for (AvatarStateListener listener : listeners) {
            listener.onAvatarStateChanged(newState);
        }
    }

    // === 状态获取 ===

    /**
     * 获取当前 Avatar 状态
     */
    @NonNull
    public AvatarState getCurrentState() {
        return currentState;
    }

    // === 面部特征获取 ===

    /**
     * 获取面部特征
     */
    @NonNull
    public AvatarState.FacialFeatures getFacialFeatures() {
        return currentState.getFacialFeatures();
    }

    /**
     * 获取脸型
     */
    @NonNull
    public String getFaceShape() {
        return currentState.getFacialFeatures().faceShape;
    }

    /**
     * 获取眸型
     */
    @NonNull
    public String getEyeShape() {
        return currentState.getFacialFeatures().eyeShape;
    }

    /**
     * 获取肤色
     */
    @NonNull
    public String getSkinTone() {
        return currentState.getFacialFeatures().skinTone;
    }

    /**
     * 获取表现力
     */
    public double getExpressiveness() {
        return currentState.getFacialFeatures().expressiveness;
    }

    /**
     * 是否高表现力
     */
    public boolean isHighExpressiveness() {
        return currentState.getFacialFeatures().isHighExpressiveness();
    }

    // === 体型特征获取 ===

    /**
     * 获取体型特征
     */
    @NonNull
    public AvatarState.BodyFeatures getBodyFeatures() {
        return currentState.getBodyFeatures();
    }

    /**
     * 获取身高
     */
    public double getHeight() {
        return currentState.getBodyFeatures().height;
    }

    /**
     * 获取体重
     */
    public double getWeight() {
        return currentState.getBodyFeatures().weight;
    }

    /**
     * 获取体型
     */
    @NonNull
    public String getBodyType() {
        return currentState.getBodyFeatures().bodyType;
    }

    /**
     * 获取姿态
     */
    @NonNull
    public String getPosture() {
        return currentState.getBodyFeatures().posture;
    }

    /**
     * 获取姿态得分
     */
    public double getPostureScore() {
        return currentState.getBodyFeatures().postureScore;
    }

    /**
     * 获取动作风格
     */
    @NonNull
    public String getMovementStyle() {
        return currentState.getBodyFeatures().movementStyle;
    }

    /**
     * 获取手势频率
     */
    public double getGestureFrequency() {
        return currentState.getBodyFeatures().gestureFrequency;
    }

    /**
     * 是否自信姿态
     */
    public boolean isConfidentPosture() {
        return currentState.getBodyFeatures().isConfidentPosture();
    }

    /**
     * 是否活跃动作
     */
    public boolean isActiveMovement() {
        return currentState.getBodyFeatures().isActiveMovement();
    }

    /**
     * 获取 BMI
     */
    public double getBMI() {
        return currentState.getBodyFeatures().getBMI();
    }

    // === 年龄外观获取 ===

    /**
     * 获取年龄外观
     */
    @NonNull
    public AvatarState.AgeAppearance getAgeAppearance() {
        return currentState.getAgeAppearance();
    }

    /**
     * 获取外观年龄
     */
    public int getApparentAge() {
        return currentState.getAgeAppearance().apparentAge;
    }

    /**
     * 获取年龄范围
     */
    @NonNull
    public String getAgeRange() {
        return currentState.getAgeAppearance().ageRange;
    }

    /**
     * 获取衰老阶段
     */
    @NonNull
    public String getAgingStage() {
        return currentState.getAgeAppearance().agingStage;
    }

    /**
     * 是否年轻外观
     */
    public boolean isYouthful() {
        return currentState.getAgeAppearance().isYouthful();
    }

    /**
     * 是否老年外观
     */
    public boolean isSenior() {
        return currentState.getAgeAppearance().isSenior();
    }

    // === 风格偏好获取 ===

    /**
     * 获取风格偏好
     */
    @NonNull
    public AvatarState.StylePreferences getStylePreferences() {
        return currentState.getStylePreferences();
    }

    /**
     * 获取服装风格
     */
    @NonNull
    public String getClothingStyle() {
        return currentState.getStylePreferences().clothingStyle;
    }

    /**
     * 获取整体气质
     */
    @NonNull
    public String getOverallVibe() {
        return currentState.getStylePreferences().overallVibe;
    }

    /**
     * 获取品牌意识
     */
    public double getBrandAwareness() {
        return currentState.getStylePreferences().brandAwareness;
    }

    /**
     * 是否正式风格
     */
    public boolean isFormalStyle() {
        return currentState.getStylePreferences().isFormalStyle();
    }

    /**
     * 是否高品牌意识
     */
    public boolean isBrandConscious() {
        return currentState.getStylePreferences().isBrandConscious();
    }

    // === 3D模型获取 ===

    /**
     * 获取3D模型引用
     */
    @NonNull
    public AvatarState.Model3DReference getModel3D() {
        return currentState.getModel3D();
    }

    /**
     * 获取模型ID
     */
    @NonNull
    public String getModelId() {
        return currentState.getModel3D().modelId;
    }

    /**
     * 获取渲染质量
     */
    @NonNull
    public String getRenderQuality() {
        return currentState.getModel3D().renderQuality;
    }

    /**
     * 是否支持动画
     */
    public boolean supportsAnimation() {
        return currentState.getModel3D().supportsAnimation();
    }

    // === 展示设置获取 ===

    /**
     * 获取展示设置
     */
    @NonNull
    public AvatarState.DisplaySettings getDisplaySettings() {
        return currentState.getDisplaySettings();
    }

    /**
     * 获取渲染模式
     */
    @NonNull
    public String getRenderMode() {
        return currentState.getDisplaySettings().renderMode;
    }

    /**
     * 获取动画状态
     */
    @NonNull
    public String getAnimationState() {
        return currentState.getDisplaySettings().animationState;
    }

    /**
     * 获取表情
     */
    @NonNull
    public String getExpression() {
        return currentState.getDisplaySettings().expression;
    }

    /**
     * 是否 3D 模式
     */
    public boolean is3DMode() {
        return currentState.getDisplaySettings().is3DMode();
    }

    /**
     * 是否 VR 模式
     */
    public boolean isVRMode() {
        return currentState.getDisplaySettings().isVRMode();
    }

    // === 决策上下文获取 ===

    /**
     * 获取决策上下文
     */
    @NonNull
    public AvatarState.AvatarDecisionContext getDecisionContext() {
        return currentState.getDecisionContext();
    }

    /**
     * 获取推荐风格
     */
    @NonNull
    public String getRecommendedStyle() {
        return currentState.getDecisionContext().recommendedStyle;
    }

    /**
     * 获取推荐姿态
     */
    @NonNull
    public String getRecommendedPosture() {
        return currentState.getDecisionContext().recommendedPosture;
    }

    /**
     * 获取推荐表情
     */
    @NonNull
    public String getRecommendedExpression() {
        return currentState.getDecisionContext().recommendedExpression;
    }

    /**
     * 获取场景适应
     */
    @NonNull
    public AvatarState.SceneAdaptation getSceneAdaptation() {
        return currentState.getDecisionContext().sceneAdaptation;
    }

    /**
     * 获取社交适应
     */
    @NonNull
    public AvatarState.SocialAdaptation getSocialAdaptation() {
        return currentState.getDecisionContext().socialAdaptation;
    }

    /**
     * 获取文化适应
     */
    @NonNull
    public AvatarState.CulturalAdaptation getCulturalAdaptation() {
        return currentState.getDecisionContext().culturalAdaptation;
    }

    // === 场景相关 ===

    /**
     * 获取当前场景
     */
    @NonNull
    public String getCurrentScene() {
        return currentState.getDecisionContext().sceneAdaptation.currentScene;
    }

    /**
     * 是否正式场景
     */
    public boolean isFormalScene() {
        return currentState.getDecisionContext().sceneAdaptation.isFormalScene();
    }

    /**
     * 获取动画集
     */
    @NonNull
    public String getAnimationSet() {
        return currentState.getDecisionContext().sceneAdaptation.animationSet;
    }

    // === 社交相关 ===

    /**
     * 获取社交上下文
     */
    @NonNull
    public String getSocialContext() {
        return currentState.getDecisionContext().socialAdaptation.socialContext;
    }

    /**
     * 获取距离级别
     */
    @NonNull
    public String getDistanceLevel() {
        return currentState.getDecisionContext().socialAdaptation.distanceLevel;
    }

    /**
     * 获取眼神接触级别
     */
    public double getEyeContactLevel() {
        return currentState.getDecisionContext().socialAdaptation.eyeContactLevel;
    }

    /**
     * 获取手势级别
     */
    public double getGestureLevel() {
        return currentState.getDecisionContext().socialAdaptation.gestureLevel;
    }

    /**
     * 是否亲密社交
     */
    public boolean isIntimateContext() {
        return currentState.getDecisionContext().socialAdaptation.isIntimateContext();
    }

    // === 文化相关 ===

    /**
     * 获取文化上下文
     */
    @NonNull
    public String getCulturalContext() {
        return currentState.getDecisionContext().culturalAdaptation.culturalContext;
    }

    /**
     * 获取正式度
     */
    public double getFormalityLevel() {
        return currentState.getDecisionContext().culturalAdaptation.formalityLevel;
    }

    /**
     * 获取问候方式
     */
    @NonNull
    public String getGreetingStyle() {
        return currentState.getDecisionContext().culturalAdaptation.greetingStyle;
    }

    /**
     * 是否高正式度
     */
    public boolean isHighFormality() {
        return currentState.getDecisionContext().culturalAdaptation.isHighFormality();
    }

    // === 决策辅助 ===

    /**
     * 获取适合当前场景的风格建议
     */
    @NonNull
    public String getStyleRecommendation() {
        String style = getRecommendedStyle();
        String vibe = getOverallVibe();

        if (isFormalScene()) {
            return "建议穿着正式风格：" + style;
        } else if ("intimate".equals(getSocialContext())) {
            return "轻松自然的穿着风格";
        } else if ("professional".equals(vibe)) {
            return "专业得体的穿着风格";
        } else {
            return "当前适合：" + style + " 风格";
        }
    }

    /**
     * 获取适合当前场景的姿态建议
     */
    @NonNull
    public String getPostureRecommendation() {
        String posture = getRecommendedPosture();

        if (isFormalScene()) {
            return "保持端正自信的姿态";
        } else if (isIntimateContext()) {
            return "放松自然的姿态";
        } else {
            return posture + " 姿态";
        }
    }

    /**
     * 获取适合当前场景的表情建议
     */
    @NonNull
    public String getExpressionRecommendation() {
        String expression = getRecommendedExpression();
        double expressiveness = getExpressiveness();

        if (isHighExpressiveness()) {
            return "可以展现丰富的面部表情";
        } else if (isFormalScene()) {
            return "保持专业自然的表情";
        } else {
            return expression + " 表情";
        }
    }

    /**
     * 获取动画建议
     */
    @NonNull
    public String getAnimationRecommendation() {
        if (!supportsAnimation()) {
            return "当前模型不支持动画";
        }

        String animationSet = getAnimationSet();
        String scene = getCurrentScene();

        switch (scene) {
            case "meeting":
                return "使用会议场景动画：" + animationSet;
            case "casual":
                return "使用轻松场景动画：" + animationSet;
            case "formal":
                return "使用正式场景动画：" + animationSet;
            default:
                return "使用默认动画：" + animationSet;
        }
    }

    /**
     * 获取渲染配置建议
     */
    @NonNull
    public String getRenderRecommendation() {
        String quality = getRenderQuality();
        String mode = getRenderMode();

        return "渲染模式：" + mode + "，质量：" + quality;
    }

    // === 行为上报 ===

    /**
     * 上报风格变化
     */
    @NonNull
    public JSONObject reportStyleChange(@NonNull String newStyle, @NonNull String reason) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "style_change");
            report.put("new_style", newStyle);
            report.put("reason", reason);
            report.put("previous_style", getClothingStyle());
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报姿态变化
     */
    @NonNull
    public JSONObject reportPostureChange(@NonNull String newPosture, @NonNull String context) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "posture_change");
            report.put("new_posture", newPosture);
            report.put("context", context);
            report.put("previous_posture", getPosture());
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报表情变化
     */
    @NonNull
    public JSONObject reportExpressionChange(@NonNull String newExpression, double intensity) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "expression_change");
            report.put("new_expression", newExpression);
            report.put("intensity", intensity);
            report.put("previous_expression", getExpression());
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
    public JSONObject reportSceneChange(@NonNull String newScene, @NonNull String previousScene) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "scene_change");
            report.put("new_scene", newScene);
            report.put("previous_scene", previousScene);
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    // === 监听器管理 ===

    /**
     * 添加状态监听器
     */
    public void addListener(@NonNull AvatarStateListener listener) {
        listeners.add(listener);
    }

    /**
     * 移除状态监听器
     */
    public void removeListener(@NonNull AvatarStateListener listener) {
        listeners.remove(listener);
    }

    /**
     * 清除所有监听器
     */
    public void clearListeners() {
        listeners.clear();
    }

    /**
     * Avatar状态监听器
     */
    public interface AvatarStateListener {
        /**
         * Avatar状态变化
         */
        void onAvatarStateChanged(@NonNull AvatarState state);
    }

    // === 状态快照 ===

    /**
     * 获取状态快照
     */
    @NonNull
    public JSONObject getStateSnapshot() {
        return currentState.toJson();
    }

    /**
     * 恢复状态快照
     */
    public void restoreStateSnapshot(@NonNull JSONObject snapshot) {
        try {
            currentState = AvatarState.fromJson(snapshot);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 重置到默认状态
     */
    public void reset() {
        currentState = new AvatarState();

        for (AvatarStateListener listener : listeners) {
            listener.onAvatarStateChanged(currentState);
        }
    }

    @NonNull
    @Override
    public String toString() {
        return "AvatarClient{" +
                "age=" + getApparentAge() +
                ", bodyType='" + getBodyType() + '\'' +
                ", style='" + getClothingStyle() + '\'' +
                ", posture='" + getPosture() + '\'' +
                ", scene='" + getCurrentScene() + '\'' +
                '}';
    }
}