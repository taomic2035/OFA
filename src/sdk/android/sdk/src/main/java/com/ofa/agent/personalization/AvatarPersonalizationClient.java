package com.ofa.agent.personalization;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.List;
import java.util.concurrent.CopyOnWriteArrayList;

/**
 * AvatarPersonalization状态客户端 (v5.4.0)
 *
 * 端侧接收 Center 推送的形象个性化状态，用于个性化展示和管理。
 * 深层个性化管理在 Center 端 PersonalizationEngine 完成。
 */
public class AvatarPersonalizationClient {

    private static volatile AvatarPersonalizationClient instance;

    // 当前状态
    private AvatarPersonalizationState currentState;

    // 监听器
    private final CopyOnWriteArrayList<PersonalizationStateListener> listeners = new CopyOnWriteArrayList<>();

    // 配置
    private String centerAddress;
    private boolean syncEnabled = false;

    private AvatarPersonalizationClient() {
        this.currentState = new AvatarPersonalizationState();
    }

    /**
     * 获取单例实例
     */
    @NonNull
    public static AvatarPersonalizationClient getInstance() {
        if (instance == null) {
            synchronized (AvatarPersonalizationClient.class) {
                if (instance == null) {
                    instance = new AvatarPersonalizationClient();
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
     * 接收 Center 推送的 AvatarPersonalization 状态
     */
    public void receivePersonalizationState(@NonNull JSONObject stateJson) {
        try {
            AvatarPersonalizationState newState = AvatarPersonalizationState.fromJson(stateJson);
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
            AvatarPersonalizationState newState = AvatarPersonalizationState.fromJson(contextJson);
            updateState(newState);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 更新状态
     */
    private void updateState(@NonNull AvatarPersonalizationState newState) {
        this.currentState = newState;

        for (PersonalizationStateListener listener : listeners) {
            listener.onPersonalizationStateChanged(newState);
        }
    }

    // === 状态获取 ===

    @NonNull
    public AvatarPersonalizationState getCurrentState() {
        return currentState;
    }

    // === Image Preferences ===

    @NonNull
    public AvatarPersonalizationState.ImagePreferences getImagePreferences() {
        return currentState.getImagePreferences();
    }

    public List<String> getPreferredColors() {
        return currentState.getImagePreferences().preferredColors;
    }

    public List<String> getPreferredStyles() {
        return currentState.getImagePreferences().preferredStyles;
    }

    public String getPresentationEffort() {
        return currentState.getImagePreferences().presentationEffort;
    }

    public double getStyleExperimentation() {
        return currentState.getImagePreferences().styleExperimentation;
    }

    public boolean isHighPresentationEffort() {
        return currentState.getImagePreferences().isHighPresentationEffort();
    }

    public boolean isExperimental() {
        return currentState.getImagePreferences().isExperimental();
    }

    public String getComfortPriority() {
        return currentState.getImagePreferences().comfortPriority;
    }

    public boolean isComfortOriented() {
        return currentState.getImagePreferences().isComfortOriented();
    }

    public String getGroomingRoutine() {
        return currentState.getImagePreferences().groomingRoutine;
    }

    public String getAccessoryFrequency() {
        return currentState.getImagePreferences().accessoryFrequency;
    }

    // === Image Evolution ===

    @NonNull
    public AvatarPersonalizationState.ImageEvolution getImageEvolution() {
        return currentState.getImageEvolution();
    }

    public String getEvolutionMode() {
        return currentState.getImageEvolution().evolutionMode;
    }

    public String getEvolutionTendency() {
        return currentState.getImageEvolution().evolutionTendency;
    }

    public double getTrendFollowingLevel() {
        return currentState.getImageEvolution().trendFollowingLevel;
    }

    public boolean isStableEvolution() {
        return currentState.getImageEvolution().isStable();
    }

    public boolean isExperimentalEvolution() {
        return currentState.getImageEvolution().isExperimental();
    }

    public String getSeasonalStyle() {
        java.util.Calendar cal = java.util.Calendar.getInstance();
        int month = cal.get(java.util.Calendar.MONTH) + 1;

        AvatarPersonalizationState.ImageEvolution evolution = currentState.getImageEvolution();
        if (month >= 3 && month <= 5) {
            return evolution.springStyle;
        } else if (month >= 6 && month <= 8) {
            return evolution.summerStyle;
        } else if (month >= 9 && month <= 11) {
            return evolution.autumnStyle;
        } else {
            return evolution.winterStyle;
        }
    }

    // === Scene Adaptation ===

    @NonNull
    public AvatarPersonalizationState.SceneAdaptationSettings getSceneAdaptationSettings() {
        return currentState.getSceneAdaptationSettings();
    }

    public String getAdaptationMode() {
        return currentState.getSceneAdaptationSettings().adaptationMode;
    }

    public boolean isAutoAdaptation() {
        return currentState.getSceneAdaptationSettings().isAutoAdaptation();
    }

    public String getDefaultWorkStyle() {
        return currentState.getSceneAdaptationSettings().defaultWorkStyle;
    }

    public String getDefaultHomeStyle() {
        return currentState.getSceneAdaptationSettings().defaultHomeStyle;
    }

    public String getDefaultSocialStyle() {
        return currentState.getSceneAdaptationSettings().defaultSocialStyle;
    }

    public String getDefaultActiveStyle() {
        return currentState.getSceneAdaptationSettings().defaultActiveStyle;
    }

    public boolean isLocationAwarenessEnabled() {
        return currentState.getSceneAdaptationSettings().locationAwareness;
    }

    public boolean isTimeAwarenessEnabled() {
        return currentState.getSceneAdaptationSettings().timeAwareness;
    }

    public boolean isWeatherAwarenessEnabled() {
        return currentState.getSceneAdaptationSettings().weatherAwareness;
    }

    // === Style Management ===

    @NonNull
    public AvatarPersonalizationState.StyleManagement getStyleManagement() {
        return currentState.getStyleManagement();
    }

    public boolean isRecommendationEnabled() {
        return currentState.getStyleManagement().recommendationEnabled;
    }

    public String getRecommendationSource() {
        return currentState.getStyleManagement().recommendationSource;
    }

    // === Context ===

    @NonNull
    public AvatarPersonalizationState.PersonalizationContext getContext() {
        return currentState.getContext();
    }

    // === Style Recommendations ===

    @NonNull
    public AvatarPersonalizationState.StyleRecommendation getRecommendedStyle() {
        return currentState.getContext().recommendedStyle;
    }

    public String getRecommendedStyleName() {
        return currentState.getContext().recommendedStyle.styleName;
    }

    public double getStyleConfidence() {
        return currentState.getContext().recommendedStyle.confidence;
    }

    public List<String> getRecommendedColorPalette() {
        return currentState.getContext().recommendedStyle.colorPalette;
    }

    // === Outfit Recommendations ===

    @NonNull
    public AvatarPersonalizationState.OutfitRecommendation getRecommendedOutfit() {
        return currentState.getContext().recommendedOutfit;
    }

    // === Scores ===

    public double getStyleScore() {
        return currentState.getContext().styleScore;
    }

    public double getConsistencyScore() {
        return currentState.getContext().consistencyScore;
    }

    public double getVersatilityScore() {
        return currentState.getContext().versatilityScore;
    }

    public double getAuthenticityScore() {
        return currentState.getContext().authenticityScore;
    }

    public double getEvolutionReadiness() {
        return currentState.getContext().evolutionReadiness;
    }

    // === Scene Info ===

    public String getCurrentScene() {
        return currentState.getContext().currentScene;
    }

    public double getSceneStyleMatch() {
        return currentState.getContext().sceneStyleMatch;
    }

    public boolean isSceneAdaptationNeeded() {
        return currentState.getContext().sceneAdaptationNeeded;
    }

    // === Evolution Info ===

    public String getEvolutionStage() {
        return currentState.getContext().evolutionStage;
    }

    public String getNextEvolutionPreview() {
        return currentState.getContext().nextEvolutionPreview;
    }

    // === Suggestions ===

    public List<AvatarPersonalizationState.StyleSuggestion> getStyleSuggestions() {
        return currentState.getContext().styleSuggestions;
    }

    public List<String> getImprovementTips() {
        return currentState.getContext().improvementTips;
    }

    public List<AvatarPersonalizationState.TrendAlert> getTrendAlerts() {
        return currentState.getContext().trendAlerts;
    }

    // === 决策辅助 ===

    /**
     * 获取适合当前场景的风格
     */
    @NonNull
    public String getStyleForScene(@NonNull String scene) {
        switch (scene) {
            case "work":
            case "meeting":
                return getDefaultWorkStyle();
            case "home":
                return getDefaultHomeStyle();
            case "social":
            case "party":
                return getDefaultSocialStyle();
            case "gym":
            case "sports":
                return getDefaultActiveStyle();
            default:
                return "casual";
        }
    }

    /**
     * 获取风格建议
     */
    @NonNull
    public String getStyleRecommendation() {
        AvatarPersonalizationState.StyleRecommendation rec = getRecommendedStyle();
        StringBuilder recommendation = new StringBuilder();

        if (rec.styleName != null && !rec.styleName.isEmpty()) {
            recommendation.append("推荐风格: ").append(rec.styleName);
        }

        if (rec.season != null && !rec.season.isEmpty()) {
            recommendation.append("\n季节: ").append(rec.season);
        }

        if (rec.occasion != null && !rec.occasion.isEmpty()) {
            recommendation.append("\n场合: ").append(rec.occasion);
        }

        if (rec.reason != null && !rec.reason.isEmpty()) {
            recommendation.append("\n原因: ").append(rec.reason);
        }

        return recommendation.toString();
    }

    /**
     * 获取穿搭建议
     */
    @NonNull
    public String getOutfitRecommendation() {
        AvatarPersonalizationState.OutfitRecommendation rec = getRecommendedOutfit();

        if (rec.outfitName == null || rec.outfitName.isEmpty()) {
            return "暂无穿搭建议";
        }

        StringBuilder recommendation = new StringBuilder();
        recommendation.append("推荐穿搭: ").append(rec.outfitName);

        if (rec.styleCategory != null && !rec.styleCategory.isEmpty()) {
            recommendation.append("\n风格: ").append(rec.styleCategory);
        }

        if (rec.occasion != null && !rec.occasion.isEmpty()) {
            recommendation.append("\n场合: ").append(rec.occasion);
        }

        if (rec.reason != null && !rec.reason.isEmpty()) {
            recommendation.append("\n原因: ").append(rec.reason);
        }

        return recommendation.toString();
    }

    /**
     * 获取颜色建议
     */
    @NonNull
    public String getColorRecommendation() {
        List<String> colors = getRecommendedColorPalette();
        if (colors.isEmpty()) {
            return "暂无颜色建议";
        }

        StringBuilder recommendation = new StringBuilder("推荐颜色: ");
        for (int i = 0; i < colors.size(); i++) {
            if (i > 0) recommendation.append(", ");
            recommendation.append(colors.get(i));
        }

        return recommendation.toString();
    }

    /**
     * 获取进化状态描述
     */
    @NonNull
    public String getEvolutionStatusDescription() {
        String stage = getEvolutionStage();
        String preview = getNextEvolutionPreview();
        double readiness = getEvolutionReadiness();

        StringBuilder status = new StringBuilder();
        status.append("当前进化模式: ").append(stage);

        if (preview != null && !preview.isEmpty()) {
            status.append("\n下一步: ").append(preview);
        }

        if (readiness > 0.7) {
            status.append("\n状态: 准备好尝试新风格");
        } else if (readiness > 0.4) {
            status.append("\n状态: 保持稳定，适度尝试");
        } else {
            status.append("\n状态: 保持当前风格");
        }

        return status.toString();
    }

    /**
     * 获取场景适应建议
     */
    @NonNull
    public String getSceneAdaptationRecommendation() {
        String scene = getCurrentScene();
        double match = getSceneStyleMatch();
        boolean needed = isSceneAdaptationNeeded();

        StringBuilder recommendation = new StringBuilder();
        recommendation.append("当前场景: ").append(scene);
        recommendation.append("\n风格匹配度: ").append(String.format("%.0f%%", match * 100));

        if (needed) {
            recommendation.append("\n建议调整风格以适应场景");
            recommendation.append("\n推荐风格: ").append(getStyleForScene(scene));
        } else {
            recommendation.append("\n当前风格与场景匹配良好");
        }

        return recommendation.toString();
    }

    /**
     * 获取综合评分描述
     */
    @NonNull
    public String getScoresDescription() {
        StringBuilder desc = new StringBuilder("风格评分:\n");
        desc.append(String.format("- 整体风格: %.0f/100\n", getStyleScore() * 100));
        desc.append(String.format("- 一致性: %.0f/100\n", getConsistencyScore() * 100));
        desc.append(String.format("- 多样性: %.0f/100\n", getVersatilityScore() * 100));
        desc.append(String.format("- 真实性: %.0f/100", getAuthenticityScore() * 100));
        return desc.toString();
    }

    /**
     * 获取改进建议汇总
     */
    @NonNull
    public String getImprovementSummary() {
        List<String> tips = getImprovementTips();
        if (tips.isEmpty()) {
            return "暂无改进建议";
        }

        StringBuilder summary = new StringBuilder("改进建议:\n");
        for (int i = 0; i < tips.size(); i++) {
            summary.append((i + 1)).append(". ").append(tips.get(i)).append("\n");
        }
        return summary.toString().trim();
    }

    /**
     * 获取综合个性化建议
     */
    @NonNull
    public String getComprehensiveAdvice() {
        StringBuilder advice = new StringBuilder();

        // 风格建议
        advice.append(getStyleRecommendation()).append("\n\n");

        // 颜色建议
        advice.append(getColorRecommendation()).append("\n\n");

        // 场景适应
        advice.append(getSceneAdaptationRecommendation()).append("\n\n");

        // 进化状态
        advice.append(getEvolutionStatusDescription());

        return advice.toString();
    }

    // === 行为上报 ===

    /**
     * 上报风格偏好更新
     */
    @NonNull
    public JSONObject reportStylePreferenceUpdate(@NonNull String styleName, double rating) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "style_preference_update");
            report.put("style_name", styleName);
            report.put("rating", rating);
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报场景适应
     */
    @NonNull
    public JSONObject reportSceneAdaptation(@NonNull String scene, @NonNull String adaptedStyle) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "scene_adaptation");
            report.put("scene", scene);
            report.put("adapted_style", adaptedStyle);
            report.put("previous_style", getRecommendedStyleName());
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报风格进化里程碑
     */
    @NonNull
    public JSONObject reportStyleMilestone(@NonNull String milestoneName, @NonNull String description) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "style_milestone");
            report.put("milestone_name", milestoneName);
            report.put("description", description);
            report.put("evolution_stage", getEvolutionStage());
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报穿搭选择
     */
    @NonNull
    public JSONObject reportOutfitSelection(@NonNull String outfitId, double satisfaction) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "outfit_selection");
            report.put("outfit_id", outfitId);
            report.put("satisfaction", satisfaction);
            report.put("style_score", getStyleScore());
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    // === 监听器管理 ===

    public void addListener(@NonNull PersonalizationStateListener listener) {
        listeners.add(listener);
    }

    public void removeListener(@NonNull PersonalizationStateListener listener) {
        listeners.remove(listener);
    }

    public void clearListeners() {
        listeners.clear();
    }

    /**
     * Personalization状态监听器
     */
    public interface PersonalizationStateListener {
        void onPersonalizationStateChanged(@NonNull AvatarPersonalizationState state);
    }

    // === 状态快照 ===

    @NonNull
    public JSONObject getStateSnapshot() {
        return currentState.toJson();
    }

    public void restoreStateSnapshot(@NonNull JSONObject snapshot) {
        try {
            currentState = AvatarPersonalizationState.fromJson(snapshot);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    public void reset() {
        currentState = new AvatarPersonalizationState();

        for (PersonalizationStateListener listener : listeners) {
            listener.onPersonalizationStateChanged(currentState);
        }
    }

    @NonNull
    @Override
    public String toString() {
        return "AvatarPersonalizationClient{" +
                "style='" + getRecommendedStyleName() + '\'' +
                ", scene='" + getCurrentScene() + '\'' +
                ", score=" + getStyleScore() +
                ", evolution='" + getEvolutionStage() + '\'' +
                '}';
    }
}