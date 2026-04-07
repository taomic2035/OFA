package com.ofa.agent.culture;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.concurrent.CopyOnWriteArrayList;

/**
 * 地域文化状态客户端 (v4.3.0)
 *
 * 端侧接收 Center 推送的地域文化状态，用于调整决策倾向。
 * 深层地域文化管理在 Center 端 RegionalCultureEngine 完成。
 */
public class RegionalCultureClient {

    private static volatile RegionalCultureClient instance;

    // 当前状态
    private RegionalCultureState currentState;

    // 监听器
    private final CopyOnWriteArrayList<RegionalCultureStateListener> listeners = new CopyOnWriteArrayList<>();

    // 配置
    private String centerAddress;
    private boolean syncEnabled = false;

    private RegionalCultureClient() {
        this.currentState = new RegionalCultureState();
    }

    /**
     * 获取单例实例
     */
    @NonNull
    public static RegionalCultureClient getInstance() {
        if (instance == null) {
            synchronized (RegionalCultureClient.class) {
                if (instance == null) {
                    instance = new RegionalCultureClient();
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
     * 接收 Center 推送的地域文化状态
     */
    public void receiveRegionalCultureState(@NonNull JSONObject stateJson) {
        try {
            RegionalCultureState newState = RegionalCultureState.fromJson(stateJson);
            updateState(newState);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 接收完整的决策上下文
     */
    public void receiveCulturalDecisionContext(@NonNull JSONObject contextJson) {
        try {
            RegionalCultureState newState = RegionalCultureState.fromJson(contextJson);
            updateState(newState);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 更新状态
     */
    private void updateState(@NonNull RegionalCultureState newState) {
        this.currentState = newState;

        for (RegionalCultureStateListener listener : listeners) {
            listener.onRegionalCultureStateChanged(newState);
        }
    }

    // === 状态获取 ===

    /**
     * 获取当前地域文化状态
     */
    @NonNull
    public RegionalCultureState getCurrentState() {
        return currentState;
    }

    // === 基本信息获取 ===

    /**
     * 获取省份
     */
    @NonNull
    public String getProvince() {
        return currentState.getProvince();
    }

    /**
     * 获取城市
     */
    @NonNull
    public String getCity() {
        return currentState.getCity();
    }

    /**
     * 获取大区域
     */
    @NonNull
    public String getRegion() {
        return currentState.getRegion();
    }

    /**
     * 获取大区域名称
     */
    @NonNull
    public String getRegionName() {
        return currentState.getRegionName();
    }

    /**
     * 获取城市等级名称
     */
    @NonNull
    public String getCityTierName() {
        return currentState.getCityTierName();
    }

    /**
     * 是否大都市
     */
    public boolean isMetropolitan() {
        return currentState.isMetropolitan();
    }

    // === 文化维度获取 ===

    /**
     * 获取集体主义倾向
     */
    public double getCollectivism() {
        return currentState.getCollectivism();
    }

    /**
     * 是否集体主义倾向
     */
    public boolean isCollectivist() {
        return currentState.isCollectivist();
    }

    /**
     * 获取传统导向
     */
    public double getTraditionOriented() {
        return currentState.getTraditionOriented();
    }

    /**
     * 获取创新导向
     */
    public double getInnovationOriented() {
        return currentState.getInnovationOriented();
    }

    /**
     * 是否传统导向
     */
    public boolean isTraditional() {
        return currentState.isTraditional();
    }

    // === 沟通风格获取 ===

    /**
     * 获取沟通风格
     */
    @NonNull
    public String getCommunicationStyle() {
        return currentState.getCommunicationStyle();
    }

    /**
     * 获取沟通风格名称
     */
    @NonNull
    public String getCommunicationStyleName() {
        return currentState.getCommunicationStyleName();
    }

    /**
     * 是否直接沟通
     */
    public boolean isDirectCommunication() {
        return currentState.isDirectCommunication();
    }

    /**
     * 获取表达开放度
     */
    public double getExpressionLevel() {
        return currentState.getExpressionLevel();
    }

    // === 社交风格获取 ===

    /**
     * 获取社交风格名称
     */
    @NonNull
    public String getSocialStyleName() {
        return currentState.getSocialStyleName();
    }

    /**
     * 获取好客程度
     */
    public double getHospitality() {
        return currentState.getHospitality();
    }

    /**
     * 获取面子意识
     */
    public double getFaceConscious() {
        return currentState.getFaceConscious();
    }

    /**
     * 是否高面子意识
     */
    public boolean hasHighFaceConsciousness() {
        return currentState.hasHighFaceConsciousness();
    }

    // === 时间相关 ===

    /**
     * 获取生活节奏
     */
    @NonNull
    public String getLifePace() {
        return currentState.getLifePace();
    }

    /**
     * 获取时间紧迫感
     */
    public double getTimeUrgency() {
        return currentState.getTimeUrgency();
    }

    /**
     * 是否快节奏
     */
    public boolean isFastPaced() {
        return currentState.isFastPaced();
    }

    // === 迁移相关 ===

    /**
     * 是否本地人
     */
    public boolean isNativeBorn() {
        return currentState.isNativeBorn();
    }

    /**
     * 是否有迁移经历
     */
    public boolean hasMigrationExperience() {
        return currentState.hasMigrationExperience();
    }

    /**
     * 获取文化适应能力
     */
    public double getCulturalAdaptation() {
        return currentState.getCulturalAdaptation();
    }

    // === 决策倾向 ===

    /**
     * 获取决策倾向
     */
    @NonNull
    public RegionalCultureState.CulturalTendencies getTendencies() {
        return currentState.getTendencies();
    }

    /**
     * 获取群体决策偏好
     */
    public double getGroupDecisionPreference() {
        return currentState.getTendencies().getGroupDecisionPreference();
    }

    /**
     * 获取创新开放度
     */
    public double getInnovationOpenness() {
        return currentState.getTendencies().getInnovationOpenness();
    }

    /**
     * 获取文化适应灵活性
     */
    public double getCulturalFlexibility() {
        return currentState.getTendencies().getCulturalFlexibility();
    }

    // === 决策建议 ===

    /**
     * 是否建议直接表达
     */
    public boolean shouldCommunicateDirectly() {
        return isDirectCommunication();
    }

    /**
     * 是否建议群体决策
     */
    public boolean shouldMakeGroupDecision() {
        return getGroupDecisionPreference() > 0.6;
    }

    /**
     * 是否建议尝试新事物
     */
    public boolean shouldTryNewThings() {
        return getInnovationOpenness() > 0.6;
    }

    /**
     * 是否需要维护面子
     */
    public boolean shouldPreserveFace() {
        return hasHighFaceConsciousness();
    }

    /**
     * 是否建议快速行动
     */
    public boolean shouldActQuickly() {
        return isFastPaced();
    }

    /**
     * 获取推荐沟通方式
     */
    @NonNull
    public String getRecommendedCommunicationApproach() {
        if (isDirectCommunication()) {
            return "直接坦诚表达";
        } else if (hasHighFaceConsciousness()) {
            return "委婉含蓄，注意维护对方面子";
        } else {
            return "根据情境灵活选择";
        }
    }

    /**
     * 获取推荐决策方式
     */
    @NonNull
    public String getRecommendedDecisionApproach() {
        if (shouldMakeGroupDecision()) {
            return "咨询他人意见，群体决策";
        } else if (isTraditional()) {
            return "参考传统做法，谨慎决策";
        } else {
            return "独立思考，自主决策";
        }
    }

    // === 行为上报 ===

    /**
     * 上报位置变化
     */
    @NonNull
    public JSONObject reportLocationChange(@NonNull String newProvince, @NonNull String newCity) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "location_change");
            report.put("new_province", newProvince);
            report.put("new_city", newCity);
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报文化适应
     */
    @NonNull
    public JSONObject reportCulturalAdaptation(@NonNull String aspect, double success) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "cultural_adaptation");
            report.put("aspect", aspect);
            report.put("success", success);
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
    public void addListener(@NonNull RegionalCultureStateListener listener) {
        listeners.add(listener);
    }

    /**
     * 移除状态监听器
     */
    public void removeListener(@NonNull RegionalCultureStateListener listener) {
        listeners.remove(listener);
    }

    /**
     * 清除所有监听器
     */
    public void clearListeners() {
        listeners.clear();
    }

    /**
     * 地域文化状态监听器
     */
    public interface RegionalCultureStateListener {
        /**
         * 地域文化状态变化
         */
        void onRegionalCultureStateChanged(@NonNull RegionalCultureState state);
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
            currentState = RegionalCultureState.fromJson(snapshot);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 重置到默认状态
     */
    public void reset() {
        currentState = new RegionalCultureState();

        for (RegionalCultureStateListener listener : listeners) {
            listener.onRegionalCultureStateChanged(currentState);
        }
    }

    @NonNull
    @Override
    public String toString() {
        return "RegionalCultureClient{" +
                "state=" + currentState +
                ", listeners=" + listeners.size() +
                '}';
    }
}