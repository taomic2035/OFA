package com.ofa.agent.social;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.concurrent.CopyOnWriteArrayList;

/**
 * 社会身份状态客户端 (v4.2.0)
 *
 * 端侧接收 Center 推送的社会身份状态，用于调整决策倾向。
 * 深层社会身份管理在 Center 端 SocialIdentityEngine 完成。
 */
public class SocialIdentityClient {

    private static volatile SocialIdentityClient instance;

    // 当前状态
    private SocialIdentityState currentState;

    // 监听器
    private final CopyOnWriteArrayList<SocialIdentityStateListener> listeners = new CopyOnWriteArrayList<>();

    // 配置
    private String centerAddress;
    private boolean syncEnabled = false;

    private SocialIdentityClient() {
        this.currentState = new SocialIdentityState();
    }

    /**
     * 获取单例实例
     */
    @NonNull
    public static SocialIdentityClient getInstance() {
        if (instance == null) {
            synchronized (SocialIdentityClient.class) {
                if (instance == null) {
                    instance = new SocialIdentityClient();
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
     * 接收 Center 推送的社会身份状态
     */
    public void receiveSocialIdentityState(@NonNull JSONObject stateJson) {
        try {
            SocialIdentityState newState = SocialIdentityState.fromJson(stateJson);
            updateState(newState);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 接收完整的决策上下文
     */
    public void receiveSocialDecisionContext(@NonNull JSONObject contextJson) {
        try {
            SocialIdentityState newState = SocialIdentityState.fromJson(contextJson);
            updateState(newState);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 更新状态
     */
    private void updateState(@NonNull SocialIdentityState newState) {
        this.currentState = newState;

        for (SocialIdentityStateListener listener : listeners) {
            listener.onSocialIdentityStateChanged(newState);
        }
    }

    // === 状态获取 ===

    /**
     * 获取当前社会身份状态
     */
    @NonNull
    public SocialIdentityState getCurrentState() {
        return currentState;
    }

    // === 教育相关 ===

    /**
     * 获取学历层次
     */
    @NonNull
    public String getEducationLevel() {
        return currentState.getEducationLevel();
    }

    /**
     * 获取学历名称
     */
    @NonNull
    public String getEducationLevelName() {
        return currentState.getEducationLevelName();
    }

    /**
     * 获取专业
     */
    @NonNull
    public String getMajor() {
        return currentState.getMajor();
    }

    /**
     * 是否高学历
     */
    public boolean isHighlyEducated() {
        return currentState.isHighlyEducated();
    }

    // === 职业相关 ===

    /**
     * 获取职业
     */
    @NonNull
    public String getOccupation() {
        return currentState.getOccupation();
    }

    /**
     * 获取职业阶段
     */
    @NonNull
    public String getCareerStage() {
        return currentState.getCareerStage();
    }

    /**
     * 获取职业阶段名称
     */
    @NonNull
    public String getCareerStageName() {
        return currentState.getCareerStageName();
    }

    /**
     * 获取工作满意度
     */
    public double getJobSatisfaction() {
        return currentState.getJobSatisfaction();
    }

    /**
     * 获取工作生活平衡
     */
    public double getWorkLifeBalance() {
        return currentState.getWorkLifeBalance();
    }

    /**
     * 是否资深职业人
     */
    public boolean isSeniorProfessional() {
        return currentState.isSeniorProfessional();
    }

    /**
     * 是否工作导向
     */
    public boolean isWorkOriented() {
        return currentState.isWorkOriented();
    }

    // === 社会阶层相关 ===

    /**
     * 获取收入层次
     */
    @NonNull
    public String getIncomeLevel() {
        return currentState.getIncomeLevel();
    }

    /**
     * 获取社会地位
     */
    @NonNull
    public String getSocialStatus() {
        return currentState.getSocialStatus();
    }

    /**
     * 获取社会地位名称
     */
    @NonNull
    public String getSocialStatusName() {
        return currentState.getSocialStatusName();
    }

    /**
     * 是否高收入
     */
    public boolean isHighIncome() {
        return currentState.isHighIncome();
    }

    /**
     * 是否有向上流动意愿
     */
    public boolean hasUpwardMobility() {
        return currentState.hasUpwardMobility();
    }

    /**
     * 获取综合资本
     */
    public double getOverallCapital() {
        return currentState.getOverallCapital();
    }

    // === 身份认同相关 ===

    /**
     * 获取主导角色
     */
    @NonNull
    public String getDominantRole() {
        return currentState.getDominantRole();
    }

    /**
     * 获取自我概念标签
     */
    @NonNull
    public java.util.List<String> getSelfConceptLabels() {
        return currentState.getSelfConceptLabels();
    }

    /**
     * 获取身份流动性
     */
    public double getIdentityFluidity() {
        return currentState.getIdentityFluidity();
    }

    // === 决策倾向 ===

    /**
     * 获取决策倾向
     */
    @NonNull
    public SocialIdentityState.SocialDecisionTendencies getTendencies() {
        return currentState.getTendencies();
    }

    /**
     * 获取认知复杂度
     */
    public double getCognitiveComplexity() {
        return currentState.getTendencies().getCognitiveComplexity();
    }

    /**
     * 获取适应性
     */
    public double getAdaptability() {
        return currentState.getTendencies().getAdaptability();
    }

    /**
     * 获取财务风险容忍度
     */
    public double getFinancialRiskTolerance() {
        return currentState.getTendencies().getFinancialRiskTolerance();
    }

    /**
     * 获取向上流动动力
     */
    public double getUpwardMobilityDrive() {
        return currentState.getTendencies().getUpwardMobilityDrive();
    }

    // === 决策建议 ===

    /**
     * 是否建议接受财务风险
     */
    public boolean shouldAcceptFinancialRisk(double riskLevel) {
        return riskLevel <= getFinancialRiskTolerance();
    }

    /**
     * 是否建议职业转型
     */
    public boolean shouldConsiderCareerChange() {
        // 工作满意度低且有向上流动意愿时建议转型
        return getJobSatisfaction() < 0.4 && hasUpwardMobility();
    }

    /**
     * 获取推荐的工作投入程度
     */
    public double getRecommendedWorkInvestment() {
        // 根据事业野心和工作生活平衡计算
        return (currentState.getCareerAmbition() + (1 - currentState.getWorkLifeBalance())) / 2;
    }

    /**
     * 是否建议追求更高社会地位
     */
    public boolean shouldPursueHigherStatus() {
        return hasUpwardMobility() && getUpwardMobilityDrive() > 0.5;
    }

    // === 行为上报 ===

    /**
     * 上报职业变化
     */
    @NonNull
    public JSONObject reportCareerChange(@NonNull String newOccupation, @NonNull String newIndustry) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "career_change");
            report.put("new_occupation", newOccupation);
            report.put("new_industry", newIndustry);
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报收入变化
     */
    @NonNull
    public JSONObject reportIncomeChange(@NonNull String newLevel, double newPercentile) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "income_change");
            report.put("new_level", newLevel);
            report.put("new_percentile", newPercentile);
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报角色变化
     */
    @NonNull
    public JSONObject reportRoleChange(@NonNull String roleName, double importance) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "role_change");
            report.put("role_name", roleName);
            report.put("importance", importance);
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报教育成就
     */
    @NonNull
    public JSONObject reportEducationAchievement(@NonNull String achievement) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "education_achievement");
            report.put("achievement", achievement);
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
    public void addListener(@NonNull SocialIdentityStateListener listener) {
        listeners.add(listener);
    }

    /**
     * 移除状态监听器
     */
    public void removeListener(@NonNull SocialIdentityStateListener listener) {
        listeners.remove(listener);
    }

    /**
     * 清除所有监听器
     */
    public void clearListeners() {
        listeners.clear();
    }

    /**
     * 社会身份状态监听器
     */
    public interface SocialIdentityStateListener {
        /**
         * 社会身份状态变化
         */
        void onSocialIdentityStateChanged(@NonNull SocialIdentityState state);
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
            currentState = SocialIdentityState.fromJson(snapshot);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 重置到默认状态
     */
    public void reset() {
        currentState = new SocialIdentityState();

        for (SocialIdentityStateListener listener : listeners) {
            listener.onSocialIdentityStateChanged(currentState);
        }
    }

    @NonNull
    @Override
    public String toString() {
        return "SocialIdentityClient{" +
                "state=" + currentState +
                ", listeners=" + listeners.size() +
                '}';
    }
}