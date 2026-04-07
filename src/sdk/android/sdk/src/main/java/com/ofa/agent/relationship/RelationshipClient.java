package com.ofa.agent.relationship;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.List;
import java.util.concurrent.CopyOnWriteArrayList;

/**
 * 人际关系状态客户端 (v4.6.0)
 *
 * 端侧接收 Center 推送的人际关系状态，用于调整社交决策。
 * 深层人际关系管理在 Center 端 RelationshipEngine 完成。
 */
public class RelationshipClient {

    private static volatile RelationshipClient instance;

    // 当前状态
    private RelationshipState currentState;

    // 监听器
    private final CopyOnWriteArrayList<RelationshipStateListener> listeners = new CopyOnWriteArrayList<>();

    // 配置
    private String centerAddress;
    private boolean syncEnabled = false;

    private RelationshipClient() {
        this.currentState = new RelationshipState();
    }

    /**
     * 获取单例实例
     */
    @NonNull
    public static RelationshipClient getInstance() {
        if (instance == null) {
            synchronized (RelationshipClient.class) {
                if (instance == null) {
                    instance = new RelationshipClient();
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
     * 接收 Center 推送的人际关系状态
     */
    public void receiveRelationshipState(@NonNull JSONObject stateJson) {
        try {
            RelationshipState newState = RelationshipState.fromJson(stateJson);
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
            RelationshipState newState = RelationshipState.fromJson(contextJson);
            updateState(newState);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 更新状态
     */
    private void updateState(@NonNull RelationshipState newState) {
        this.currentState = newState;

        for (RelationshipStateListener listener : listeners) {
            listener.onRelationshipStateChanged(newState);
        }
    }

    // === 状态获取 ===

    /**
     * 获取当前人际关系状态
     */
    @NonNull
    public RelationshipState getCurrentState() {
        return currentState;
    }

    // === 网络摘要获取 ===

    /**
     * 获取网络摘要
     */
    @NonNull
    public RelationshipState.NetworkSummary getNetworkSummary() {
        return currentState.getNetworkSummary();
    }

    /**
     * 获取总联系人数
     */
    public int getTotalContacts() {
        return currentState.getNetworkSummary().totalContacts;
    }

    /**
     * 获取亲密联系人数
     */
    public int getCloseContacts() {
        return currentState.getNetworkSummary().closeContacts;
    }

    /**
     * 获取支持联系人数
     */
    public int getSupportContacts() {
        return currentState.getNetworkSummary().supportContacts;
    }

    /**
     * 获取社交资本
     */
    public double getSocialCapital() {
        return currentState.getNetworkSummary().socialCapital;
    }

    /**
     * 获取网络健康度
     */
    public double getNetworkHealth() {
        return currentState.getNetworkSummary().networkHealth;
    }

    /**
     * 是否有支持网络
     */
    public boolean hasSupportNetwork() {
        return currentState.getNetworkSummary().hasSupport();
    }

    // === 关系倾向获取 ===

    /**
     * 获取关系倾向
     */
    @NonNull
    public RelationshipState.RelationshipOrientation getOrientation() {
        return currentState.getOrientation();
    }

    /**
     * 是否外向
     */
    public boolean isExtraverted() {
        return currentState.getOrientation().isExtraverted();
    }

    /**
     * 是否内向
     */
    public boolean isIntroverted() {
        return currentState.getOrientation().isIntroverted();
    }

    /**
     * 获取关系寻求倾向
     */
    public double getRelationshipSeeking() {
        return currentState.getOrientation().relationshipSeeking;
    }

    /**
     * 获取亲密舒适度
     */
    public double getIntimacyComfort() {
        return currentState.getOrientation().intimacyComfort;
    }

    // === 社交风格获取 ===

    /**
     * 获取社交风格
     */
    @NonNull
    public RelationshipState.SocialStyle getSocialStyle() {
        return currentState.getSocialStyle();
    }

    /**
     * 获取社交能量
     */
    public double getSocialEnergy() {
        return currentState.getSocialStyle().socialEnergy;
    }

    /**
     * 是否高社交能量
     */
    public boolean hasHighSocialEnergy() {
        return currentState.getSocialStyle().hasHighSocialEnergy();
    }

    /**
     * 是否低社交能量
     */
    public boolean hasLowSocialEnergy() {
        return currentState.getSocialStyle().hasLowSocialEnergy();
    }

    /**
     * 是否偏好一对一
     */
    public boolean prefersOneOnOne() {
        return currentState.getSocialStyle().prefersOneOnOne();
    }

    /**
     * 是否偏好群体
     */
    public boolean prefersGroup() {
        return currentState.getSocialStyle().prefersGroup();
    }

    // === 依恋风格获取 ===

    /**
     * 获取依恋风格
     */
    @NonNull
    public RelationshipState.AttachmentStyle getAttachmentStyle() {
        return currentState.getAttachmentStyle();
    }

    /**
     * 获取依恋风格类型
     */
    @NonNull
    public String getAttachmentStyleType() {
        return currentState.getAttachmentStyle().primaryStyle;
    }

    /**
     * 是否安全依恋
     */
    public boolean isSecurelyAttached() {
        return currentState.isSecurelyAttached();
    }

    /**
     * 是否焦虑依恋
     */
    public boolean isAnxiouslyAttached() {
        return currentState.isAnxiouslyAttached();
    }

    /**
     * 是否回避依恋
     */
    public boolean isAvoidantlyAttached() {
        return currentState.isAvoidantlyAttached();
    }

    /**
     * 获取焦虑水平
     */
    public double getAnxietyLevel() {
        return currentState.getAttachmentStyle().anxietyLevel;
    }

    /**
     * 获取回避水平
     */
    public double getAvoidanceLevel() {
        return currentState.getAttachmentStyle().avoidanceLevel;
    }

    // === 决策影响获取 ===

    /**
     * 获取决策影响
     */
    @NonNull
    public RelationshipState.RelationshipDecisionInfluence getDecisionInfluence() {
        return currentState.getDecisionInfluence();
    }

    /**
     * 获取社交趋近度
     */
    public double getSocialApproach() {
        return currentState.getDecisionInfluence().socialApproach;
    }

    /**
     * 获取对他人的信任
     */
    public double getTrustInOthers() {
        return currentState.getDecisionInfluence().trustInOthers;
    }

    /**
     * 获取合作准备度
     */
    public double getCooperationReadiness() {
        return currentState.getDecisionInfluence().cooperationReadiness;
    }

    /**
     * 获取安全感水平
     */
    public double getSecurityLevel() {
        return currentState.getDecisionInfluence().securityLevel;
    }

    /**
     * 获取冲突风格
     */
    @NonNull
    public String getConflictStyle() {
        return currentState.getDecisionInfluence().conflictStyle;
    }

    // === 社交建议获取 ===

    /**
     * 获取社交建议
     */
    @NonNull
    public RelationshipState.SocialGuidance getSocialGuidance() {
        return currentState.getSocialGuidance();
    }

    /**
     * 是否应社交
     */
    public boolean shouldSocialize() {
        return currentState.shouldSocialize();
    }

    /**
     * 获取社交能量水平
     */
    public double getSocialEnergyLevel() {
        return currentState.getSocialEnergyLevel();
    }

    /**
     * 获取偏好群体规模
     */
    @NonNull
    public String getPreferredGroupSize() {
        return currentState.getSocialGuidance().preferredGroupSize;
    }

    /**
     * 获取偏好深度
     */
    @NonNull
    public String getPreferredDepth() {
        return currentState.getSocialGuidance().preferredDepth;
    }

    /**
     * 获取冲突风险
     */
    public double getConflictRisk() {
        return currentState.getConflictRisk();
    }

    /**
     * 是否需要自我关怀
     */
    public boolean needsSelfCare() {
        return currentState.needsSelfCare();
    }

    /**
     * 获取需要维护的关系
     */
    @NonNull
    public List<String> getMaintenanceNeeded() {
        return currentState.getSocialGuidance().maintenanceNeeded;
    }

    /**
     * 获取紧张关系
     */
    @NonNull
    public List<String> getTensionRelationships() {
        return currentState.getSocialGuidance().tensionRelationships;
    }

    /**
     * 获取边界提醒
     */
    @NonNull
    public List<String> getBoundaryReminders() {
        return currentState.getSocialGuidance().boundaryReminders;
    }

    /**
     * 获取成长机会
     */
    @NonNull
    public List<String> getGrowthOpportunities() {
        return currentState.getSocialGuidance().growthOpportunities;
    }

    // === 决策辅助 ===

    /**
     * 是否建议社交互动
     */
    public boolean shouldEngageSocially() {
        return shouldSocialize() && hasHighSocialEnergy();
    }

    /**
     * 是否建议建立新关系
     */
    public boolean shouldBuildNewRelationships() {
        RelationshipState.NetworkSummary network = currentState.getNetworkSummary();
        return network.weakTies < 3 || network.totalContacts < 5;
    }

    /**
     * 是否建议维护现有关系
     */
    public boolean shouldMaintainRelationships() {
        return currentState.getSocialGuidance().needsMaintenance();
    }

    /**
     * 是否有冲突风险提示
     */
    public boolean hasConflictRiskWarning() {
        return currentState.getSocialGuidance().hasConflictRisk();
    }

    /**
     * 获取推荐社交方式
     */
    @NonNull
    public String getRecommendedSocialApproach() {
        if (needsSelfCare()) {
            return "优先照顾自己，适度独处";
        }

        if (!shouldSocialize()) {
            return "社交能量较低，建议安静休息";
        }

        if (prefersOneOnOne()) {
            return "一对一深度交流";
        } else if (prefersGroup()) {
            return "参与群体活动";
        } else {
            return "根据情境灵活选择";
        }
    }

    /**
     * 获取推荐沟通方式
     */
    @NonNull
    public String getRecommendedCommunicationApproach() {
        RelationshipState.SocialStyle style = currentState.getSocialStyle();

        if (style.deepTalkPreference > 0.6) {
            return "深入探讨，分享真实想法";
        } else if (style.smallTalkComfort > 0.6) {
            return "轻松闲聊，建立连接";
        } else {
            return "根据关系深浅调整";
        }
    }

    /**
     * 获取关系建议
     */
    @NonNull
    public String getRelationshipAdvice() {
        StringBuilder advice = new StringBuilder();

        // 社交能量提示
        if (hasLowSocialEnergy()) {
            advice.append("社交能量较低，建议节省精力。");
        }

        // 依恋风格提示
        if (isAnxiouslyAttached()) {
            advice.append("注意不要过度寻求确认，保持独立。");
        } else if (isAvoidantlyAttached()) {
            advice.append("尝试适度开放，建立更深连接。");
        }

        // 维护建议
        List<String> maintenance = getMaintenanceNeeded();
        if (maintenance.size() > 0) {
            advice.append("需要联系：").append(String.join("、", maintenance.subList(0, Math.min(3, maintenance.size())))).append("。");
        }

        // 冲突风险
        if (hasConflictRiskWarning()) {
            advice.append("存在关系紧张，注意沟通方式。");
        }

        // 边界提醒
        List<String> boundaries = getBoundaryReminders();
        if (boundaries.size() > 0) {
            advice.append(String.join("；", boundaries)).append("。");
        }

        return advice.length() > 0 ? advice.toString() : "人际关系状态良好";
    }

    /**
     * 获取社交活动建议
     */
    @NonNull
    public String getSocialActivitySuggestion() {
        if (!shouldSocialize()) {
            return "建议休息充电，待能量恢复后再社交";
        }

        String groupSize = getPreferredGroupSize();
        String depth = getPreferredDepth();

        if ("one_on_one".equals(groupSize) && "deep".equals(depth)) {
            return "适合与亲密朋友一对一深度交流";
        } else if ("one_on_one".equals(groupSize)) {
            return "适合与朋友轻松聊天";
        } else if ("deep".equals(depth)) {
            return "适合参与有深度讨论的小组活动";
        } else {
            return "适合参与轻松的社交活动";
        }
    }

    // === 行为上报 ===

    /**
     * 上报社交互动
     */
    @NonNull
    public JSONObject reportSocialInteraction(@NonNull String personId, @NonNull String type, double satisfaction) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "social_interaction");
            report.put("person_id", personId);
            report.put("interaction_type", type);
            report.put("satisfaction", satisfaction);
            report.put("social_energy_before", getSocialEnergy());
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报关系变化
     */
    @NonNull
    public JSONObject reportRelationshipChange(@NonNull String personId, @NonNull String changeType, double impact) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "relationship_change");
            report.put("person_id", personId);
            report.put("change_type", changeType); // new/deepened/distant/ended
            report.put("impact", impact);
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报冲突事件
     */
    @NonNull
    public JSONObject reportConflict(@NonNull String personId, @NonNull String description, double intensity) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "conflict_report");
            report.put("person_id", personId);
            report.put("description", description);
            report.put("intensity", intensity);
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报能量变化
     */
    @NonNull
    public JSONObject reportEnergyChange(double newEnergyLevel, @NonNull String activity) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "energy_change");
            report.put("new_energy_level", newEnergyLevel);
            report.put("activity", activity);
            report.put("previous_energy", getSocialEnergy());
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
    public void addListener(@NonNull RelationshipStateListener listener) {
        listeners.add(listener);
    }

    /**
     * 移除状态监听器
     */
    public void removeListener(@NonNull RelationshipStateListener listener) {
        listeners.remove(listener);
    }

    /**
     * 清除所有监听器
     */
    public void clearListeners() {
        listeners.clear();
    }

    /**
     * 人际关系状态监听器
     */
    public interface RelationshipStateListener {
        /**
         * 人际关系状态变化
         */
        void onRelationshipStateChanged(@NonNull RelationshipState state);
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
            currentState = RelationshipState.fromJson(snapshot);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 重置到默认状态
     */
    public void reset() {
        currentState = new RelationshipState();

        for (RelationshipStateListener listener : listeners) {
            listener.onRelationshipStateChanged(currentState);
        }
    }

    @NonNull
    @Override
    public String toString() {
        return "RelationshipClient{" +
                "contacts=" + getTotalContacts() +
                ", closeContacts=" + getCloseContacts() +
                ", attachment=" + getAttachmentStyleType() +
                ", energy=" + getSocialEnergy() +
                '}';
    }
}