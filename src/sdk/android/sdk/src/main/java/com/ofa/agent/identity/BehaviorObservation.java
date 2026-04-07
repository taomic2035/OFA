package com.ofa.agent.identity;

import androidx.annotation.NonNull;

import java.util.HashMap;
import java.util.Map;

/**
 * BehaviorObservation - 行为观察记录
 *
 * 用于从用户行为推断性格特质。
 */
public class BehaviorObservation {

    private String id;
    private String type;           // decision/interaction/preference/activity
    private Map<String, Object> context;
    private String outcome;
    private Map<String, Double> inferences;  // 推断的性格特质
    private long timestamp;

    // 行为类型常量
    public static final String TYPE_DECISION = "decision";
    public static final String TYPE_INTERACTION = "interaction";
    public static final String TYPE_PREFERENCE = "preference";
    public static final String TYPE_ACTIVITY = "activity";

    /**
     * 创建行为观察
     */
    public BehaviorObservation(@NonNull String type, @NonNull Map<String, Object> context) {
        this.id = generateId();
        this.type = type;
        this.context = new HashMap<>(context);
        this.outcome = "";
        this.inferences = new HashMap<>();
        this.timestamp = System.currentTimeMillis();
    }

    /**
     * 创建行为观察（完整参数）
     */
    public BehaviorObservation(@NonNull String id, @NonNull String type,
                               @NonNull Map<String, Object> context,
                               @NonNull String outcome,
                               @NonNull Map<String, Double> inferences,
                               long timestamp) {
        this.id = id;
        this.type = type;
        this.context = new HashMap<>(context);
        this.outcome = outcome;
        this.inferences = new HashMap<>(inferences);
        this.timestamp = timestamp;
    }

    // === 推断规则 ===

    /**
     * 根据行为类型自动推断性格变化
     */
    public void autoInfer() {
        inferences.clear();

        switch (type) {
            case TYPE_DECISION:
                inferFromDecision();
                break;
            case TYPE_INTERACTION:
                inferFromInteraction();
                break;
            case TYPE_PREFERENCE:
                inferFromPreference();
                break;
            case TYPE_ACTIVITY:
                inferFromActivity();
                break;
        }
    }

    /**
     * 从决策行为推断
     */
    private void inferFromDecision() {
        String decisionType = (String) context.get("decision_type");

        if (decisionType == null) return;

        // 冲动购买 → 神经质+, 尽责性-
        if (decisionType.equals("impulse_purchase")) {
            inferences.put("neuroticism", 0.05);
            inferences.put("conscientiousness", -0.03);
        }

        // 仔细规划 → 尽责性+
        if (decisionType.equals("careful_planning")) {
            inferences.put("conscientiousness", 0.05);
        }

        // 风险投资 → 风险承受度+
        if (decisionType.equals("risky_investment")) {
            inferences.put("risk_tolerance", 0.05);
        }

        // 保守储蓄 → 风险承受度-
        if (decisionType.equals("conservative_saving")) {
            inferences.put("risk_tolerance", -0.03);
        }
    }

    /**
     * 从交互行为推断
     */
    private void inferFromInteraction() {
        String interactionType = (String) context.get("interaction_type");

        if (interactionType == null) return;

        // 群聊活跃 → 外向性+
        if (interactionType.equals("group_chats")) {
            inferences.put("extraversion", 0.05);
        }

        // 私聊偏好 → 外向性-
        if (interactionType.equals("private_chats")) {
            inferences.put("extraversion", -0.03);
        }

        // 表情丰富 → 开放性+
        if (interactionType.equals("emoji_heavy")) {
            inferences.put("openness", 0.03);
        }

        // 正式语言 → 尽责性+
        if (interactionType.equals("formal_language")) {
            inferences.put("conscientiousness", 0.03);
            inferences.put("formality", 0.05);
        }
    }

    /**
     * 从偏好行为推断
     */
    private void inferFromPreference() {
        String preferenceType = (String) context.get("preference_type");

        if (preferenceType == null) return;

        // 尝试新事物 → 开放性+
        if (preferenceType.equals("novel_trying")) {
            inferences.put("openness", 0.05);
        }

        // 偏好熟悉 → 开放性-, 尽责性+
        if (preferenceType.equals("routine_following")) {
            inferences.put("openness", -0.03);
            inferences.put("conscientiousness", 0.03);
        }

        // 健康饮食 → 健康重视度+
        if (preferenceType.equals("healthy_eating")) {
            inferences.put("health", 0.05);
        }

        // 追求隐私 → 隐私重视度+
        if (preferenceType.equals("privacy_focused")) {
            inferences.put("privacy", 0.05);
        }
    }

    /**
     * 从活动行为推断
     */
    private void inferFromActivity() {
        String activityType = (String) context.get("activity_type");

        if (activityType == null) return;

        // 探索新地方 → 开放性+
        if (activityType.equals("exploring_new")) {
            inferences.put("openness", 0.05);
        }

        // 规律作息 → 尽责性+
        if (activityType.equals("regular_schedule")) {
            inferences.put("conscientiousness", 0.05);
        }

        // 学习新技能 → 学习重视度+
        if (activityType.equals("learning_skills")) {
            inferences.put("learning", 0.05);
        }

        // 社交活动 → 社交重视度+
        if (activityType.equals("social_events")) {
            inferences.put("social", 0.05);
            inferences.put("extraversion", 0.03);
        }
    }

    // === 辅助方法 ===

    private String generateId() {
        return System.currentTimeMillis() + "_" + Integer.toHexString((int)(Math.random() * 10000));
    }

    // === Getters ===

    public String getId() { return id; }
    public String getType() { return type; }
    public Map<String, Object> getContext() { return new HashMap<>(context); }
    public String getOutcome() { return outcome; }
    public Map<String, Double> getInferences() { return new HashMap<>(inferences); }
    public long getTimestamp() { return timestamp; }

    /**
     * 转换为 JSON
     */
    @NonNull
    public String toJson() {
        StringBuilder sb = new StringBuilder();
        sb.append("{");
        sb.append("\"id\":\"").append(id).append("\",");
        sb.append("\"type\":\"").append(type).append("\",");
        sb.append("\"timestamp\":").append(timestamp);
        sb.append("}");
        return sb.toString();
    }
}