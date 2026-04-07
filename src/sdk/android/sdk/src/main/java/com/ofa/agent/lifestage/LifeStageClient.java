package com.ofa.agent.lifestage;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.concurrent.CopyOnWriteArrayList;

/**
 * 人生阶段状态客户端 (v4.4.0)
 *
 * 端侧接收 Center 推送的人生阶段状态，用于调整决策倾向。
 * 深层人生阶段管理在 Center 端 LifeStageEngine 完成。
 */
public class LifeStageClient {

    private static volatile LifeStageClient instance;

    // 当前状态
    private LifeStageState currentState;

    // 监听器
    private final CopyOnWriteArrayList<LifeStageStateListener> listeners = new CopyOnWriteArrayList<>();

    // 配置
    private String centerAddress;
    private boolean syncEnabled = false;

    private LifeStageClient() {
        this.currentState = new LifeStageState();
    }

    /**
     * 获取单例实例
     */
    @NonNull
    public static LifeStageClient getInstance() {
        if (instance == null) {
            synchronized (LifeStageClient.class) {
                if (instance == null) {
                    instance = new LifeStageClient();
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
     * 接收 Center 推送的人生阶段状态
     */
    public void receiveLifeStageState(@NonNull JSONObject stateJson) {
        try {
            LifeStageState newState = LifeStageState.fromJson(stateJson);
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
            // 解析决策上下文中的阶段信息
            JSONObject stageJson = contextJson.optJSONObject("current_stage");
            if (stageJson == null) {
                stageJson = new JSONObject();
            }

            // 添加影响和轨迹信息
            JSONObject influenceJson = contextJson.optJSONObject("stage_influence");
            if (influenceJson != null) {
                stageJson.put("stage_influence", influenceJson);
            }

            JSONObject trajectoryJson = contextJson.optJSONObject("trajectory_summary");
            if (trajectoryJson != null) {
                stageJson.put("trajectory_summary", trajectoryJson);
            }

            JSONObject statusJson = contextJson.optJSONObject("development_status");
            if (statusJson != null) {
                JSONArray bottlenecks = statusJson.optJSONArray("bottlenecks");
                if (bottlenecks != null) {
                    stageJson.put("bottlenecks", bottlenecks);
                }
            }

            LifeStageState newState = LifeStageState.fromJson(stageJson);
            updateState(newState);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 更新状态
     */
    private void updateState(@NonNull LifeStageState newState) {
        this.currentState = newState;

        for (LifeStageStateListener listener : listeners) {
            listener.onLifeStageStateChanged(newState);
        }
    }

    // === 状态获取 ===

    /**
     * 获取当前人生阶段状态
     */
    @NonNull
    public LifeStageState getCurrentState() {
        return currentState;
    }

    // === 阶段基本信息 ===

    /**
     * 获取阶段名称
     */
    @NonNull
    public String getStageName() {
        return currentState.getStageName();
    }

    /**
     * 获取阶段名称（中文）
     */
    @NonNull
    public String getStageNameZh() {
        return currentState.getStageNameZh();
    }

    /**
     * 获取阶段描述
     */
    @NonNull
    public String getStageDescription() {
        return currentState.getStageDescription();
    }

    /**
     * 获取成长焦点
     */
    @NonNull
    public String getGrowthFocus() {
        return currentState.getGrowthFocus();
    }

    /**
     * 获取阶段年龄
     */
    public int getStageAge() {
        return currentState.getStageAge();
    }

    /**
     * 获取完成度
     */
    public double getCompleteness() {
        return currentState.getCompleteness();
    }

    /**
     * 获取满意度
     */
    public double getSatisfaction() {
        return currentState.getSatisfaction();
    }

    // === 阶段特征 ===

    /**
     * 是否青年或成年早期
     */
    public boolean isYoungAdult() {
        return currentState.isYoungAdult();
    }

    /**
     * 是否中年
     */
    public boolean isMidlife() {
        return currentState.isMidlife();
    }

    /**
     * 是否老年
     */
    public boolean isElderly() {
        return currentState.isElderly();
    }

    /**
     * 是否处于成长期
     */
    public boolean isGrowthPhase() {
        return currentState.isGrowthPhase();
    }

    /**
     * 是否面临挑战
     */
    public boolean hasChallenges() {
        return currentState.hasChallenges();
    }

    /**
     * 获取主要挑战
     */
    @Nullable
    public String getDominantChallenge() {
        return currentState.getDominantChallenge();
    }

    /**
     * 获取主要目标
     */
    @Nullable
    public String getDominantGoal() {
        return currentState.getDominantGoal();
    }

    // === 发展指标 ===

    /**
     * 获取发展指标
     */
    @NonNull
    public LifeStageState.DevelopmentMetrics getMetrics() {
        return currentState.getMetrics();
    }

    /**
     * 获取身体健康度
     */
    public double getPhysicalHealth() {
        return currentState.getMetrics().getPhysicalHealth();
    }

    /**
     * 获取心理健康度
     */
    public double getMentalHealth() {
        return currentState.getMetrics().getMentalHealth();
    }

    /**
     * 获取职业进展
     */
    public double getCareerProgress() {
        return currentState.getMetrics().getCareerProgress();
    }

    /**
     * 获取财务稳定度
     */
    public double getFinancialStability() {
        return currentState.getMetrics().getFinancialStability();
    }

    /**
     * 获取人生目标清晰度
     */
    public double getPurposeClarity() {
        return currentState.getMetrics().getPurposeClarity();
    }

    // === 阶段影响 ===

    /**
     * 获取阶段影响
     */
    @NonNull
    public LifeStageState.StageInfluence getInfluence() {
        return currentState.getInfluence();
    }

    /**
     * 是否偏向冒险
     */
    public boolean isRiskTaking() {
        return currentState.getInfluence().isRiskTaking();
    }

    /**
     * 是否偏向稳健
     */
    public boolean isConservative() {
        return currentState.getInfluence().isConservative();
    }

    /**
     * 是否重视事业
     */
    public boolean isCareerFocused() {
        return currentState.getInfluence().isCareerFocused();
    }

    /**
     * 是否重视家庭
     */
    public boolean isFamilyFocused() {
        return currentState.getInfluence().isFamilyFocused();
    }

    /**
     * 是否重视健康
     */
    public boolean isHealthFocused() {
        return currentState.getInfluence().isHealthFocused();
    }

    /**
     * 是否有紧迫感
     */
    public boolean hasUrgency() {
        return currentState.getInfluence().hasUrgency();
    }

    /**
     * 获取时间视角
     */
    @NonNull
    public String getTimePerspective() {
        return currentState.getInfluence().getTimePerspective();
    }

    // === 轨迹摘要 ===

    /**
     * 获取轨迹摘要
     */
    @NonNull
    public LifeStageState.TrajectorySummary getTrajectory() {
        return currentState.getTrajectory();
    }

    /**
     * 是否上升轨迹
     */
    public boolean isUpwardTrajectory() {
        return currentState.getTrajectory().isUpward();
    }

    /**
     * 是否下降轨迹
     */
    public boolean isDownwardTrajectory() {
        return currentState.getTrajectory().isDownward();
    }

    /**
     * 是否高韧性
     */
    public boolean hasHighResilience() {
        return currentState.getTrajectory().hasHighResilience();
    }

    /**
     * 是否有智慧积累
     */
    public boolean hasWisdom() {
        return currentState.getTrajectory().hasWisdom();
    }

    // === 发展瓶颈 ===

    /**
     * 获取发展瓶颈
     */
    @NonNull
    public java.util.List<String> getBottlenecks() {
        return currentState.getBottlenecks();
    }

    /**
     * 是否有发展瓶颈
     */
    public boolean hasBottlenecks() {
        return currentState.getBottlenecks().size() > 0;
    }

    // === 决策建议 ===

    /**
     * 获取推荐风险偏好
     */
    @NonNull
    public String getRecommendedRiskApproach() {
        if (isRiskTaking()) {
            return "可以考虑尝试新机会";
        } else if (isConservative()) {
            return "建议稳健谨慎";
        } else {
            return "根据具体情况权衡";
        }
    }

    /**
     * 获取推荐焦点
     */
    @NonNull
    public String getRecommendedFocus() {
        if (isCareerFocused() && isFamilyFocused()) {
            return "平衡事业与家庭";
        } else if (isCareerFocused()) {
            return "优先发展事业";
        } else if (isFamilyFocused()) {
            return "重视家庭关系";
        } else if (isHealthFocused()) {
            return "关注身心健康";
        } else {
            return "根据当前阶段需求";
        }
    }

    /**
     * 获取推荐时间管理
     */
    @NonNull
    public String getRecommendedTimeManagement() {
        String perspective = getTimePerspective();
        if ("future".equals(perspective)) {
            return "规划未来，着眼长远";
        } else if ("present".equals(perspective)) {
            return "把握当下，稳步前行";
        } else {
            return "回顾经验，传承智慧";
        }
    }

    /**
     * 是否建议学习新技能
     */
    public boolean shouldLearnNewSkills() {
        return isYoungAdult() || isGrowthPhase();
    }

    /**
     * 是否建议健康管理
     */
    public boolean shouldFocusOnHealth() {
        return isMidlife() || isElderly() || isHealthFocused() || getPhysicalHealth() < 0.5;
    }

    /**
     * 是否建议职业突破
     */
    public boolean shouldFocusOnCareer() {
        return isYoungAdult() && isCareerFocused() && getCareerProgress() < 0.7;
    }

    /**
     * 是否建议家庭投入
     */
    public boolean shouldFocusOnFamily() {
        return (isMidlife() || currentState.getStageName().equals("early_adult")) && isFamilyFocused();
    }

    // === 行为上报 ===

    /**
     * 上报里程碑达成
     */
    @NonNull
    public JSONObject reportMilestoneAchieved(@NonNull String milestoneName, double significance) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "milestone_achieved");
            report.put("milestone_name", milestoneName);
            report.put("significance", significance);
            report.put("stage_name", currentState.getStageName());
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报阶段目标进展
     */
    @NonNull
    public JSONObject reportGoalProgress(@NonNull String goal, double progress) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "goal_progress");
            report.put("goal", goal);
            report.put("progress", progress);
            report.put("stage_name", currentState.getStageName());
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报人生感悟
     */
    @NonNull
    public JSONObject reportLifeLesson(@NonNull String lesson, @NonNull String category) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "life_lesson");
            report.put("lesson", lesson);
            report.put("category", category);
            report.put("stage_name", currentState.getStageName());
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
    public void addListener(@NonNull LifeStageStateListener listener) {
        listeners.add(listener);
    }

    /**
     * 移除状态监听器
     */
    public void removeListener(@NonNull LifeStageStateListener listener) {
        listeners.remove(listener);
    }

    /**
     * 清除所有监听器
     */
    public void clearListeners() {
        listeners.clear();
    }

    /**
     * 人生阶段状态监听器
     */
    public interface LifeStageStateListener {
        /**
         * 人生阶段状态变化
         */
        void onLifeStageStateChanged(@NonNull LifeStageState state);
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
            currentState = LifeStageState.fromJson(snapshot);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 重置到默认状态
     */
    public void reset() {
        currentState = new LifeStageState();

        for (LifeStageStateListener listener : listeners) {
            listener.onLifeStageStateChanged(currentState);
        }
    }

    @NonNull
    @Override
    public String toString() {
        return "LifeStageClient{" +
                "state=" + currentState +
                ", listeners=" + listeners.size() +
                '}';
    }
}