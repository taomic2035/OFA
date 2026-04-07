package com.ofa.agent.philosophy;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * 三观状态模型 (v4.1.0)
 *
 * 端侧只接收 Center 推送的三观状态，用于调整决策倾向和行为。
 * 深层三观管理在 Center 端 PhilosophyEngine 完成。
 */
public class PhilosophyState {

    // === 世界观 ===
    private double materialism;       // 唯物主义倾向
    private double agency;            // 自主意识信念
    private double trustInPeople;     // 对人的信任度
    private double optimism;          // 乐观程度
    private double individualism;     // 个人主义倾向

    // === 人生观 ===
    private String lifeGoal;          // 人生目标
    private String lifeStage;         // 人生阶段
    private double presentFocus;      // 当下关注
    private double futureFocus;       // 未来关注
    private double hedonism;          // 享乐主义
    private double balance;           // 平衡主义

    // === 价值观 ===
    private List<String> topValues;   // 最重要的价值观
    private double riskTolerance;     // 风险容忍度
    private double honesty;           // 诚实
    private double compassion;        // 同情心
    private double family;            // 家庭

    // === 决策倾向 ===
    private DecisionTendencies tendencies;

    // === 道德指南 ===
    private List<String> primaryMoralValues;
    private double ethicalSensitivity;

    // === 时间属性 ===
    private long timestamp;

    public PhilosophyState() {
        // 默认值
        this.materialism = 0.6;
        this.agency = 0.7;
        this.trustInPeople = 0.5;
        this.optimism = 0.6;
        this.individualism = 0.5;
        this.lifeGoal = "实现自我价值";
        this.lifeStage = "early_career";
        this.presentFocus = 0.4;
        this.futureFocus = 0.5;
        this.hedonism = 0.3;
        this.balance = 0.5;
        this.topValues = new ArrayList<>();
        this.topValues.add("family");
        this.topValues.add("health");
        this.topValues.add("honesty");
        this.riskTolerance = 0.4;
        this.honesty = 0.8;
        this.compassion = 0.7;
        this.family = 0.8;
        this.tendencies = new DecisionTendencies();
        this.primaryMoralValues = new ArrayList<>();
        this.ethicalSensitivity = 0.6;
        this.timestamp = System.currentTimeMillis();
    }

    /**
     * 从 JSON 解析 (Center 推送的状态)
     */
    @NonNull
    public static PhilosophyState fromJson(@NonNull JSONObject json) throws JSONException {
        PhilosophyState state = new PhilosophyState();

        // 世界观
        JSONObject worldview = json.optJSONObject("worldview");
        if (worldview != null) {
            state.materialism = worldview.optDouble("materialism", 0.6);
            state.agency = worldview.optDouble("agency", 0.7);
            state.trustInPeople = worldview.optDouble("trust_in_people", 0.5);
            state.optimism = worldview.optDouble("optimism", 0.6);
            state.individualism = worldview.optDouble("individualism", 0.5);
        }

        // 人生观
        JSONObject lifeView = json.optJSONObject("life_view");
        if (lifeView != null) {
            state.lifeGoal = lifeView.optString("life_goal", "实现自我价值");
            state.lifeStage = lifeView.optString("life_stage", "early_career");
            state.presentFocus = lifeView.optDouble("present_focus", 0.4);
            state.futureFocus = lifeView.optDouble("future_focus", 0.5);
            state.hedonism = lifeView.optDouble("hedonism", 0.3);
            state.balance = lifeView.optDouble("balance", 0.5);
        }

        // 价值观
        JSONObject valueSystem = json.optJSONObject("value_system");
        if (valueSystem != null) {
            state.riskTolerance = valueSystem.optDouble("risk_tolerance", 0.4);
            state.honesty = valueSystem.optDouble("honesty", 0.8);
            state.compassion = valueSystem.optDouble("compassion", 0.7);
            state.family = valueSystem.optDouble("family", 0.8);
        }

        // 决策倾向
        JSONObject tendencies = json.optJSONObject("decision_tendencies");
        if (tendencies != null) {
            state.tendencies = DecisionTendencies.fromJson(tendencies);
        }

        // 道德指南
        JSONObject moral = json.optJSONObject("moral_guidance");
        if (moral != null) {
            state.primaryMoralValues = parseStringList(moral, "primary_moral_values");
            state.ethicalSensitivity = moral.optDouble("ethical_sensitivity", 0.6);
        }

        state.timestamp = json.optLong("timestamp", System.currentTimeMillis());

        return state;
    }

    /**
     * 转换为 JSON
     */
    @NonNull
    public JSONObject toJson() {
        JSONObject json = new JSONObject();
        try {
            // 世界观
            JSONObject worldview = new JSONObject();
            worldview.put("materialism", materialism);
            worldview.put("agency", agency);
            worldview.put("trust_in_people", trustInPeople);
            worldview.put("optimism", optimism);
            worldview.put("individualism", individualism);
            json.put("worldview", worldview);

            // 人生观
            JSONObject lifeView = new JSONObject();
            lifeView.put("life_goal", lifeGoal);
            lifeView.put("life_stage", lifeStage);
            lifeView.put("present_focus", presentFocus);
            lifeView.put("future_focus", futureFocus);
            lifeView.put("hedonism", hedonism);
            lifeView.put("balance", balance);
            json.put("life_view", lifeView);

            // 价值观
            JSONObject valueSystem = new JSONObject();
            valueSystem.put("risk_tolerance", riskTolerance);
            valueSystem.put("honesty", honesty);
            valueSystem.put("compassion", compassion);
            valueSystem.put("family", family);
            json.put("value_system", valueSystem);

            // 决策倾向
            if (tendencies != null) {
                json.put("decision_tendencies", tendencies.toJson());
            }

            // 道德指南
            JSONObject moral = new JSONObject();
            moral.put("primary_moral_values", arrayToJson(primaryMoralValues));
            moral.put("ethical_sensitivity", ethicalSensitivity);
            json.put("moral_guidance", moral);

            json.put("timestamp", timestamp);

        } catch (JSONException e) {
            // ignore
        }
        return json;
    }

    /**
     * 获取世界观类型
     */
    @NonNull
    public String getWorldviewType() {
        if (materialism > 0.7 && agency > 0.7) {
            return "rationalist"; // 理性主义者
        }
        if (individualism > 0.7) {
            return "individualist"; // 个人主义者
        }
        if (individualism < 0.3) {
            return "collectivist"; // 集体主义者
        }
        if (optimism > 0.7) {
            return "optimist"; // 乐观主义者
        }
        if (optimism < 0.3) {
            return "pessimist"; // 悲观主义者
        }
        return "pragmatist"; // 实用主义者
    }

    /**
     * 获取人生观类型
     */
    @NonNull
    public String getLifeViewType() {
        if (hedonism > 0.7 && presentFocus > 0.7) {
            return "hedonist"; // 享乐主义者
        }
        if (futureFocus > 0.7) {
            return "ambitious"; // 进取者
        }
        if (balance > 0.7) {
            return "balanced"; // 平衡主义者
        }
        return "pragmatic"; // 务实主义者
    }

    /**
     * 是否愿意冒险
     */
    public boolean isWillingToTakeRisk() {
        return riskTolerance > 0.5 || (optimism > 0.6 && agency > 0.6);
    }

    /**
     * 是否注重长远
     */
    public boolean isLongTermOriented() {
        return futureFocus > 0.6;
    }

    /**
     * 是否注重家庭
     */
    public boolean isFamilyOriented() {
        return family > 0.7;
    }

    /**
     * 是否信任他人
     */
    public boolean isTrusting() {
        return trustInPeople > 0.6;
    }

    /**
     * 是否道德敏感
     */
    public boolean isEthicallySensitive() {
        return ethicalSensitivity > 0.6;
    }

    /**
     * 获取推荐决策风格
     */
    @NonNull
    public String getRecommendedDecisionStyle() {
        if (isLongTermOriented() && tendencies.getLongTermPlanning() > 0.6) {
            return "deliberate"; // 深思熟虑
        }
        if (isWillingToTakeRisk() && tendencies.getInnovationTendency() > 0.6) {
            return "innovative"; // 创新型
        }
        if (isTrusting() && tendencies.getCooperation() > 0.6) {
            return "collaborative"; // 协作型
        }
        if (balance > 0.6) {
            return "balanced"; // 平衡型
        }
        return "pragmatic"; // 务实型
    }

    /**
     * 获取人生阶段名称
     */
    @NonNull
    public String getLifeStageName() {
        switch (lifeStage) {
            case "youth":
                return "青年期";
            case "early_career":
                return "事业起步期";
            case "mid_career":
                return "事业成熟期";
            case "mature":
                return "成熟期";
            case "late":
                return "晚年期";
            default:
                return lifeStage;
        }
    }

    // === Getter/Setter ===

    public double getMaterialism() { return materialism; }
    public void setMaterialism(double materialism) { this.materialism = clamp01(materialism); }

    public double getAgency() { return agency; }
    public void setAgency(double agency) { this.agency = clamp01(agency); }

    public double getTrustInPeople() { return trustInPeople; }
    public void setTrustInPeople(double trustInPeople) { this.trustInPeople = clamp01(trustInPeople); }

    public double getOptimism() { return optimism; }
    public void setOptimism(double optimism) { this.optimism = clamp01(optimism); }

    public double getIndividualism() { return individualism; }
    public void setIndividualism(double individualism) { this.individualism = clamp01(individualism); }

    public String getLifeGoal() { return lifeGoal; }
    public void setLifeGoal(String lifeGoal) { this.lifeGoal = lifeGoal; }

    public String getLifeStage() { return lifeStage; }
    public void setLifeStage(String lifeStage) { this.lifeStage = lifeStage; }

    public double getPresentFocus() { return presentFocus; }
    public void setPresentFocus(double presentFocus) { this.presentFocus = clamp01(presentFocus); }

    public double getFutureFocus() { return futureFocus; }
    public void setFutureFocus(double futureFocus) { this.futureFocus = clamp01(futureFocus); }

    public double getHedonism() { return hedonism; }
    public void setHedonism(double hedonism) { this.hedonism = clamp01(hedonism); }

    public double getBalance() { return balance; }
    public void setBalance(double balance) { this.balance = clamp01(balance); }

    public List<String> getTopValues() { return topValues; }
    public void setTopValues(List<String> topValues) { this.topValues = topValues; }

    public double getRiskTolerance() { return riskTolerance; }
    public void setRiskTolerance(double riskTolerance) { this.riskTolerance = clamp01(riskTolerance); }

    public double getHonesty() { return honesty; }
    public void setHonesty(double honesty) { this.honesty = clamp01(honesty); }

    public double getCompassion() { return compassion; }
    public void setCompassion(double compassion) { this.compassion = clamp01(compassion); }

    public double getFamily() { return family; }
    public void setFamily(double family) { this.family = clamp01(family); }

    public DecisionTendencies getTendencies() { return tendencies; }
    public void setTendencies(DecisionTendencies tendencies) { this.tendencies = tendencies; }

    public List<String> getPrimaryMoralValues() { return primaryMoralValues; }
    public void setPrimaryMoralValues(List<String> primaryMoralValues) { this.primaryMoralValues = primaryMoralValues; }

    public double getEthicalSensitivity() { return ethicalSensitivity; }
    public void setEthicalSensitivity(double ethicalSensitivity) { this.ethicalSensitivity = clamp01(ethicalSensitivity); }

    public long getTimestamp() { return timestamp; }
    public void setTimestamp(long timestamp) { this.timestamp = timestamp; }

    @NonNull
    @Override
    public String toString() {
        return "PhilosophyState{" +
                "worldview='" + getWorldviewType() + '\'' +
                ", lifeView='" + getLifeViewType() + '\'' +
                ", lifeStage='" + lifeStage + '\'' +
                ", riskTolerance=" + riskTolerance +
                '}';
    }

    // === 辅助方法 ===

    private double clamp01(double value) {
        if (value < 0) return 0;
        if (value > 1) return 1;
        return value;
    }

    private static List<String> parseStringList(JSONObject json, String key) {
        List<String> result = new ArrayList<>();
        org.json.JSONArray array = json.optJSONArray(key);
        if (array != null) {
            for (int i = 0; i < array.length(); i++) {
                result.add(array.optString(i));
            }
        }
        return result;
    }

    private org.json.JSONArray arrayToJson(List<String> list) {
        org.json.JSONArray array = new org.json.JSONArray();
        for (String s : list) {
            array.put(s);
        }
        return array;
    }

    /**
     * 决策倾向
     */
    public static class DecisionTendencies {
        private double riskTolerance;
        private double delayedGratification;
        private double longTermPlanning;
        private double cooperation;
        private double autonomy;
        private double innovationTendency;
        private double workLifeBalance;

        public DecisionTendencies() {
            this.riskTolerance = 0.4;
            this.delayedGratification = 0.5;
            this.longTermPlanning = 0.5;
            this.cooperation = 0.5;
            this.autonomy = 0.5;
            this.innovationTendency = 0.5;
            this.workLifeBalance = 0.5;
        }

        @NonNull
        public static DecisionTendencies fromJson(@NonNull JSONObject json) {
            DecisionTendencies t = new DecisionTendencies();
            t.riskTolerance = json.optDouble("risk_tolerance", 0.4);
            t.delayedGratification = json.optDouble("delayed_gratification", 0.5);
            t.longTermPlanning = json.optDouble("long_term_planning", 0.5);
            t.cooperation = json.optDouble("cooperation", 0.5);
            t.autonomy = json.optDouble("autonomy", 0.5);
            t.innovationTendency = json.optDouble("innovation_tendency", 0.5);
            t.workLifeBalance = json.optDouble("work_life_balance", 0.5);
            return t;
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("risk_tolerance", riskTolerance);
                json.put("delayed_gratification", delayedGratification);
                json.put("long_term_planning", longTermPlanning);
                json.put("cooperation", cooperation);
                json.put("autonomy", autonomy);
                json.put("innovation_tendency", innovationTendency);
                json.put("work_life_balance", workLifeBalance);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public double getRiskTolerance() { return riskTolerance; }
        public void setRiskTolerance(double riskTolerance) { this.riskTolerance = riskTolerance; }

        public double getDelayedGratification() { return delayedGratification; }
        public void setDelayedGratification(double delayedGratification) { this.delayedGratification = delayedGratification; }

        public double getLongTermPlanning() { return longTermPlanning; }
        public void setLongTermPlanning(double longTermPlanning) { this.longTermPlanning = longTermPlanning; }

        public double getCooperation() { return cooperation; }
        public void setCooperation(double cooperation) { this.cooperation = cooperation; }

        public double getAutonomy() { return autonomy; }
        public void setAutonomy(double autonomy) { this.autonomy = autonomy; }

        public double getInnovationTendency() { return innovationTendency; }
        public void setInnovationTendency(double innovationTendency) { this.innovationTendency = innovationTendency; }

        public double getWorkLifeBalance() { return workLifeBalance; }
        public void setWorkLifeBalance(double workLifeBalance) { this.workLifeBalance = workLifeBalance; }
    }
}