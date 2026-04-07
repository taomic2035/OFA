package com.ofa.agent.identity;

import androidx.annotation.NonNull;

import java.util.HashMap;
import java.util.Map;

/**
 * ValueSystem - 价值观系统
 *
 * 描述用户的核心价值观和决策倾向，用于个性化决策支持。
 */
public class ValueSystem {

    // 核心价值观权重 (0-1)
    private double privacy;         // 隐私重视程度
    private double efficiency;      // 效率优先程度
    private double health;          // 健康重视程度
    private double family;          // 家庭重视程度
    private double career;          // 事业重视程度
    private double entertainment;   // 娱乐重视程度
    private double learning;        // 学习重视程度
    private double social;          // 社交重视程度
    private double finance;         // 财务重视程度
    private double environment;     // 环保重视程度

    // 决策倾向
    private double riskTolerance;   // 风险承受度 (0-1)
    private double impulsiveness;   // 冲动程度 (0-1)
    private double patience;        // 耐心程度 (0-1)

    // 自定义价值观
    private Map<String, Double> customValues;

    // 价值观描述（AI 生成）
    private String summary;

    /**
     * 创建默认价值观
     */
    public ValueSystem() {
        // 默认价值观
        this.privacy = 0.7;
        this.efficiency = 0.6;
        this.health = 0.7;
        this.family = 0.8;
        this.career = 0.6;
        this.entertainment = 0.5;
        this.learning = 0.6;
        this.social = 0.5;
        this.finance = 0.6;
        this.environment = 0.5;

        // 决策倾向
        this.riskTolerance = 0.4;
        this.impulsiveness = 0.3;
        this.patience = 0.6;

        // 其他
        this.customValues = new HashMap<>();
        this.summary = "";
    }

    // === 更新方法 ===

    /**
     * 更新价值观
     */
    public void updateValue(@NonNull String key, double value) {
        value = clamp01(value);

        switch (key) {
            case "privacy":
                privacy = value;
                break;
            case "efficiency":
                efficiency = value;
                break;
            case "health":
                health = value;
                break;
            case "family":
                family = value;
                break;
            case "career":
                career = value;
                break;
            case "entertainment":
                entertainment = value;
                break;
            case "learning":
                learning = value;
                break;
            case "social":
                social = value;
                break;
            case "finance":
                finance = value;
                break;
            case "environment":
                environment = value;
                break;
            case "risk_tolerance":
                riskTolerance = value;
                break;
            case "impulsiveness":
                impulsiveness = value;
                break;
            case "patience":
                patience = value;
                break;
            default:
                customValues.put(key, value);
        }
    }

    /**
     * 批量更新价值观
     */
    public void updateValues(@NonNull Map<String, Double> updates) {
        for (Map.Entry<String, Double> entry : updates.entrySet()) {
            updateValue(entry.getKey(), entry.getValue());
        }
    }

    /**
     * 获取价值观优先级排序
     */
    @NonNull
    public String[] getValuePriority() {
        Map<String, Double> values = new HashMap<>();
        values.put("privacy", privacy);
        values.put("efficiency", efficiency);
        values.put("health", health);
        values.put("family", family);
        values.put("career", career);
        values.put("entertainment", entertainment);
        values.put("learning", learning);
        values.put("social", social);
        values.put("finance", finance);
        values.put("environment", environment);

        // 简单排序（降序）
        return values.entrySet().stream()
            .sorted((a, b) -> Double.compare(b.getValue(), a.getValue()))
            .map(Map.Entry::getKey)
            .toArray(String[]::new);
    }

    /**
     * 获取最重要的价值观（前 3 个）
     */
    @NonNull
    public String[] getTopValues(int limit) {
        String[] priority = getValuePriority();
        if (priority.length <= limit) {
            return priority;
        }
        String[] top = new String[limit];
        System.arraycopy(priority, 0, top, 0, limit);
        return top;
    }

    // === 辅助方法 ===

    private double clamp01(double value) {
        if (value < 0) return 0;
        if (value > 1) return 1;
        return value;
    }

    // === Getters ===

    public double getPrivacy() { return privacy; }
    public double getEfficiency() { return efficiency; }
    public double getHealth() { return health; }
    public double getFamily() { return family; }
    public double getCareer() { return career; }
    public double getEntertainment() { return entertainment; }
    public double getLearning() { return learning; }
    public double getSocial() { return social; }
    public double getFinance() { return finance; }
    public double getEnvironment() { return environment; }

    public double getRiskTolerance() { return riskTolerance; }
    public double getImpulsiveness() { return impulsiveness; }
    public double getPatience() { return patience; }

    public Map<String, Double> getCustomValues() { return new HashMap<>(customValues); }
    public String getSummary() { return summary; }
    public void setSummary(String summary) { this.summary = summary; }

    /**
     * 转换为 JSON 字符串
     */
    @NonNull
    public String toJson() {
        StringBuilder sb = new StringBuilder();
        sb.append("{");
        sb.append("\"privacy\":").append(privacy).append(",");
        sb.append("\"efficiency\":").append(efficiency).append(",");
        sb.append("\"health\":").append(health).append(",");
        sb.append("\"family\":").append(family).append(",");
        sb.append("\"career\":").append(career).append(",");
        sb.append("\"entertainment\":").append(entertainment).append(",");
        sb.append("\"learning\":").append(learning).append(",");
        sb.append("\"social\":").append(social).append(",");
        sb.append("\"finance\":").append(finance).append(",");
        sb.append("\"environment\":").append(environment).append(",");
        sb.append("\"risk_tolerance\":").append(riskTolerance).append(",");
        sb.append("\"impulsiveness\":").append(impulsiveness).append(",");
        sb.append("\"patience\":").append(patience);
        sb.append("}");
        return sb.toString();
    }

    /**
     * 从 JSON 解析
     */
    @Nullable
    public static ValueSystem fromJson(@NonNull String json) {
        try {
            org.json.JSONObject obj = new org.json.JSONObject(json);
            return fromJsonObject(obj);
        } catch (Exception e) {
            return null;
        }
    }

    /**
     * 从 JSONObject 解析
     */
    @Nullable
    public static ValueSystem fromJsonObject(org.json.JSONObject obj) {
        if (obj == null) return null;
        try {
            ValueSystem valueSystem = new ValueSystem();

            if (obj.has("privacy")) {
                valueSystem.privacy = obj.getDouble("privacy");
            }
            if (obj.has("efficiency")) {
                valueSystem.efficiency = obj.getDouble("efficiency");
            }
            if (obj.has("health")) {
                valueSystem.health = obj.getDouble("health");
            }
            if (obj.has("family")) {
                valueSystem.family = obj.getDouble("family");
            }
            if (obj.has("career")) {
                valueSystem.career = obj.getDouble("career");
            }
            if (obj.has("entertainment")) {
                valueSystem.entertainment = obj.getDouble("entertainment");
            }
            if (obj.has("learning")) {
                valueSystem.learning = obj.getDouble("learning");
            }
            if (obj.has("social")) {
                valueSystem.social = obj.getDouble("social");
            }
            if (obj.has("finance")) {
                valueSystem.finance = obj.getDouble("finance");
            }
            if (obj.has("environment")) {
                valueSystem.environment = obj.getDouble("environment");
            }
            if (obj.has("risk_tolerance")) {
                valueSystem.riskTolerance = obj.getDouble("risk_tolerance");
            }
            if (obj.has("impulsiveness")) {
                valueSystem.impulsiveness = obj.getDouble("impulsiveness");
            }
            if (obj.has("patience")) {
                valueSystem.patience = obj.getDouble("patience");
            }

            return valueSystem;
        } catch (Exception e) {
            return null;
        }
    }
}