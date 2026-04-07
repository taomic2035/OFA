package com.ofa.agent.philosophy;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.concurrent.CopyOnWriteArrayList;

/**
 * 三观状态客户端 (v4.1.0)
 *
 * 端侧接收 Center 推送的三观状态，用于调整决策倾向和行为。
 * 深层三观管理在 Center 端 PhilosophyEngine 完成。
 */
public class PhilosophyClient {

    private static volatile PhilosophyClient instance;

    // 当前状态
    private PhilosophyState currentState;

    // 监听器
    private final CopyOnWriteArrayList<PhilosophyStateListener> listeners = new CopyOnWriteArrayList<>();

    // 配置
    private String centerAddress;
    private boolean syncEnabled = false;

    private PhilosophyClient() {
        this.currentState = new PhilosophyState();
    }

    /**
     * 获取单例实例
     */
    @NonNull
    public static PhilosophyClient getInstance() {
        if (instance == null) {
            synchronized (PhilosophyClient.class) {
                if (instance == null) {
                    instance = new PhilosophyClient();
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
     * 接收 Center 推送的三观状态
     */
    public void receivePhilosophyState(@NonNull JSONObject stateJson) {
        try {
            PhilosophyState newState = PhilosophyState.fromJson(stateJson);
            updateState(newState);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 接收完整的三观决策上下文
     */
    public void receivePhilosophyContext(@NonNull JSONObject contextJson) {
        try {
            PhilosophyState newState = PhilosophyState.fromJson(contextJson);
            updateState(newState);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 更新状态
     */
    private void updateState(@NonNull PhilosophyState newState) {
        PhilosophyState oldState = this.currentState;
        this.currentState = newState;

        // 通知监听器
        for (PhilosophyStateListener listener : listeners) {
            listener.onPhilosophyStateChanged(newState);
        }
    }

    // === 状态获取 ===

    /**
     * 获取当前三观状态
     */
    @NonNull
    public PhilosophyState getCurrentState() {
        return currentState;
    }

    /**
     * 获取世界观类型
     */
    @NonNull
    public String getWorldviewType() {
        return currentState.getWorldviewType();
    }

    /**
     * 获取人生观类型
     */
    @NonNull
    public String getLifeViewType() {
        return currentState.getLifeViewType();
    }

    /**
     * 获取决策倾向
     */
    @NonNull
    public PhilosophyState.DecisionTendencies getDecisionTendencies() {
        return currentState.getTendencies();
    }

    /**
     * 获取人生目标
     */
    @NonNull
    public String getLifeGoal() {
        return currentState.getLifeGoal();
    }

    /**
     * 获取人生阶段
     */
    @NonNull
    public String getLifeStage() {
        return currentState.getLifeStage();
    }

    /**
     * 获取最重要的价值观
     */
    @NonNull
    public java.util.List<String> getTopValues() {
        return currentState.getTopValues();
    }

    // === 决策辅助 ===

    /**
     * 是否愿意冒险
     */
    public boolean isWillingToTakeRisk() {
        return currentState.isWillingToTakeRisk();
    }

    /**
     * 是否注重长远
     */
    public boolean isLongTermOriented() {
        return currentState.isLongTermOriented();
    }

    /**
     * 是否注重家庭
     */
    public boolean isFamilyOriented() {
        return currentState.isFamilyOriented();
    }

    /**
     * 是否信任他人
     */
    public boolean isTrusting() {
        return currentState.isTrusting();
    }

    /**
     * 是否道德敏感
     */
    public boolean isEthicallySensitive() {
        return currentState.isEthicallySensitive();
    }

    /**
     * 获取推荐决策风格
     */
    @NonNull
    public String getRecommendedDecisionStyle() {
        return currentState.getRecommendedDecisionStyle();
    }

    /**
     * 获取风险容忍度
     */
    public double getRiskTolerance() {
        return currentState.getRiskTolerance();
    }

    /**
     * 获取合作倾向
     */
    public double getCooperation() {
        return currentState.getTendencies().getCooperation();
    }

    /**
     * 获取自主性
     */
    public double getAutonomy() {
        return currentState.getTendencies().getAutonomy();
    }

    /**
     * 获取创新倾向
     */
    public double getInnovationTendency() {
        return currentState.getTendencies().getInnovationTendency();
    }

    /**
     * 获取工作生活平衡倾向
     */
    public double getWorkLifeBalance() {
        return currentState.getTendencies().getWorkLifeBalance();
    }

    // === 决策建议 ===

    /**
     * 获取风险决策建议
     *
     * @param riskLevel 风险级别 (0-1)
     * @return 是否建议接受该风险
     */
    public boolean shouldAcceptRisk(double riskLevel) {
        double tolerance = getRiskTolerance();
        // 风险级别高于容忍度时不建议接受
        return riskLevel <= tolerance;
    }

    /**
     * 获取时间决策建议
     *
     * @param immediateReward 即时回报
     * @param delayedReward 延迟回报
     * @param delayDays 延迟天数
     * @return 是否建议选择延迟回报
     */
    public boolean shouldChooseDelayed(double immediateReward, double delayedReward, int delayDays) {
        double gratification = currentState.getTendencies().getDelayedGratification();
        // 延迟满足倾向高时，更愿意等待
        double threshold = 1.0 - gratification;
        double ratio = immediateReward / delayedReward;
        return ratio < threshold;
    }

    /**
     * 获取合作决策建议
     *
     * @param trustLevel 对合作伙伴的信任级别
     * @return 是否建议合作
     */
    public boolean shouldCooperate(double trustLevel) {
        double baseCooperation = getCooperation();
        // 综合考虑基础合作倾向和具体信任度
        return (baseCooperation + trustLevel) / 2 > 0.5;
    }

    /**
     * 获取价值观冲突解决建议
     *
     * @param value1 价值观1
     * @param value2 价值观2
     * @return 应优先的价值观
     */
    @NonNull
    public String resolveValueConflict(@NonNull String value1, @NonNull String value2) {
        java.util.List<String> topValues = getTopValues();
        int index1 = topValues.indexOf(value1);
        int index2 = topValues.indexOf(value2);

        // 排名靠前的优先
        if (index1 == -1 && index2 == -1) {
            return value1; // 默认返回第一个
        }
        if (index1 == -1) {
            return value2;
        }
        if (index2 == -1) {
            return value1;
        }
        return index1 < index2 ? value1 : value2;
    }

    // === 行为上报 ===

    /**
     * 上报价值判断
     *
     * @param situation 情境描述
     * @param choice 选择
     * @param reasoning 推理过程
     * @return 上报结果
     */
    @NonNull
    public JSONObject reportValueJudgment(@NonNull String situation, @NonNull String choice, @NonNull String reasoning) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "value_judgment");
            report.put("situation", situation);
            report.put("choice", choice);
            report.put("reasoning", reasoning);
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报人生事件
     *
     * @param eventType 事件类型
     * @param description 事件描述
     * @param impact 影响程度
     * @return 上报结果
     */
    @NonNull
    public JSONObject reportLifeEvent(@NonNull String eventType, @NonNull String description, double impact) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "life_event");
            report.put("event_type", eventType);
            report.put("description", description);
            report.put("impact", impact);
            report.put("timestamp", System.currentTimeMillis());
        } catch (JSONException e) {
            // ignore
        }
        return report;
    }

    /**
     * 上报价值观变化
     *
     * @param valueName 价值观名称
     * @param oldValue 旧值
     * @param newValue 新值
     * @param reason 原因
     * @return 上报结果
     */
    @NonNull
    public JSONObject reportValueChange(@NonNull String valueName, double oldValue, double newValue, @NonNull String reason) {
        JSONObject report = new JSONObject();
        try {
            report.put("type", "value_change");
            report.put("value_name", valueName);
            report.put("old_value", oldValue);
            report.put("new_value", newValue);
            report.put("reason", reason);
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
    public void addListener(@NonNull PhilosophyStateListener listener) {
        listeners.add(listener);
    }

    /**
     * 移除状态监听器
     */
    public void removeListener(@NonNull PhilosophyStateListener listener) {
        listeners.remove(listener);
    }

    /**
     * 清除所有监听器
     */
    public void clearListeners() {
        listeners.clear();
    }

    /**
     * 三观状态监听器
     */
    public interface PhilosophyStateListener {
        /**
         * 三观状态变化
         */
        void onPhilosophyStateChanged(@NonNull PhilosophyState state);
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
            currentState = PhilosophyState.fromJson(snapshot);
        } catch (JSONException e) {
            // 解析失败
        }
    }

    /**
     * 重置到默认状态
     */
    public void reset() {
        currentState = new PhilosophyState();

        for (PhilosophyStateListener listener : listeners) {
            listener.onPhilosophyStateChanged(currentState);
        }
    }

    @NonNull
    @Override
    public String toString() {
        return "PhilosophyClient{" +
                "state=" + currentState +
                ", listeners=" + listeners.size() +
                '}';
    }
}