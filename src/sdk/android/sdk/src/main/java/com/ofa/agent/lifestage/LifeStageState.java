package com.ofa.agent.lifestage;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * 人生阶段状态模型 (v4.4.0)
 *
 * 端侧只接收 Center 推送的人生阶段状态，用于调整决策倾向。
 * 深层人生阶段管理在 Center 端 LifeStageEngine 完成。
 */
public class LifeStageState {

    // === 当前阶段 ===
    private String stageId;
    private String stageName;          // childhood/adolescence/youth/early_adult/mid_adult/mature/elderly
    private String stageLabel;         // 阶段标签（中文）
    private String description;        // 阶段描述
    private String growthFocus;        // 成长焦点

    // === 阶段特征 ===
    private List<String> challenges;   // 阶段挑战
    private List<String> opportunities; // 阶段机遇
    private List<String> goals;        // 阶段目标
    private List<String> tasks;        // 发展任务

    // === 阶段属性 ===
    private int stageAge;              // 该阶段年龄
    private double completeness;       // 完成度 (0-1)
    private double satisfaction;       // 满意度 (0-1)

    // === 发展指标 ===
    private DevelopmentMetrics metrics;

    // === 阶段影响 ===
    private StageInfluence influence;

    // === 轨迹摘要 ===
    private TrajectorySummary trajectory;

    // === 发展瓶颈 ===
    private List<String> bottlenecks;

    // === 时间属性 ===
    private long timestamp;

    public LifeStageState() {
        this.stageName = "early_adult";
        this.stageLabel = "成年早期";
        this.description = "建设期，事业和家庭奠基";
        this.growthFocus = "事业家庭";
        this.challenges = new ArrayList<>();
        this.opportunities = new ArrayList<>();
        this.goals = new ArrayList<>();
        this.tasks = new ArrayList<>();
        this.metrics = new DevelopmentMetrics();
        this.influence = new StageInfluence();
        this.trajectory = new TrajectorySummary();
        this.bottlenecks = new ArrayList<>();
        this.timestamp = System.currentTimeMillis();
    }

    /**
     * 从 JSON 解析
     */
    @NonNull
    public static LifeStageState fromJson(@NonNull JSONObject json) throws JSONException {
        LifeStageState state = new LifeStageState();

        // 当前阶段
        state.stageId = json.optString("stage_id", "");
        state.stageName = json.optString("stage_name", "early_adult");
        state.stageLabel = json.optString("stage_label", "成年早期");
        state.description = json.optString("description", "");
        state.growthFocus = json.optString("growth_focus", "");

        // 阶段特征
        state.challenges = parseStringList(json, "challenges");
        state.opportunities = parseStringList(json, "opportunities");
        state.goals = parseStringList(json, "goals");
        state.tasks = parseStringList(json, "tasks");

        // 阶段属性
        state.stageAge = json.optInt("stage_age", 28);
        state.completeness = json.optDouble("completeness", 0.3);
        state.satisfaction = json.optDouble("satisfaction", 0.5);

        // 发展指标
        JSONObject metricsJson = json.optJSONObject("development_metrics");
        if (metricsJson != null) {
            state.metrics = DevelopmentMetrics.fromJson(metricsJson);
        }

        // 阶段影响
        JSONObject influenceJson = json.optJSONObject("stage_influence");
        if (influenceJson != null) {
            state.influence = StageInfluence.fromJson(influenceJson);
        }

        // 轨迹摘要
        JSONObject trajectoryJson = json.optJSONObject("trajectory_summary");
        if (trajectoryJson != null) {
            state.trajectory = TrajectorySummary.fromJson(trajectoryJson);
        }

        // 发展瓶颈
        state.bottlenecks = parseStringList(json, "bottlenecks");

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
            json.put("stage_id", stageId);
            json.put("stage_name", stageName);
            json.put("stage_label", stageLabel);
            json.put("description", description);
            json.put("growth_focus", growthFocus);
            json.put("challenges", listToJson(challenges));
            json.put("opportunities", listToJson(opportunities));
            json.put("goals", listToJson(goals));
            json.put("tasks", listToJson(tasks));
            json.put("stage_age", stageAge);
            json.put("completeness", completeness);
            json.put("satisfaction", satisfaction);
            if (metrics != null) {
                json.put("development_metrics", metrics.toJson());
            }
            if (influence != null) {
                json.put("stage_influence", influence.toJson());
            }
            if (trajectory != null) {
                json.put("trajectory_summary", trajectory.toJson());
            }
            json.put("bottlenecks", listToJson(bottlenecks));
            json.put("timestamp", timestamp);
        } catch (JSONException e) {
            // ignore
        }
        return json;
    }

    // === 描述方法 ===

    /**
     * 获取阶段名称（中文）
     */
    @NonNull
    public String getStageNameZh() {
        switch (stageName) {
            case "childhood": return "童年期";
            case "adolescence": return "青春期";
            case "youth": return "青年期";
            case "early_adult": return "成年早期";
            case "mid_adult": return "成年中期";
            case "mature": return "成熟期";
            case "elderly": return "老年期";
            default: return stageName;
        }
    }

    /**
     * 是否青年或成年早期
     */
    public boolean isYoungAdult() {
        return "youth".equals(stageName) || "early_adult".equals(stageName);
    }

    /**
     * 是否中年
     */
    public boolean isMidlife() {
        return "mid_adult".equals(stageName) || "mature".equals(stageName);
    }

    /**
     * 是否老年
     */
    public boolean isElderly() {
        return "elderly".equals(stageName);
    }

    /**
     * 是否处于成长期
     */
    public boolean isGrowthPhase() {
        return completeness < 0.7;
    }

    /**
     * 是否面临挑战
     */
    public boolean hasChallenges() {
        return challenges.size() > 0;
    }

    /**
     * 获取主要挑战
     */
    @Nullable
    public String getDominantChallenge() {
        if (challenges.size() > 0) {
            return challenges.get(0);
        }
        return null;
    }

    /**
     * 获取主要目标
     */
    @Nullable
    public String getDominantGoal() {
        if (goals.size() > 0) {
            return goals.get(0);
        }
        return null;
    }

    /**
     * 获取阶段描述
     */
    @NonNull
    public String getStageDescription() {
        StringBuilder desc = new StringBuilder();

        desc.append("处于").append(getStageNameZh()).append("，");
        desc.append(description).append("。");

        if (challenges.size() > 0) {
            desc.append("面临挑战：").append(getDominantChallenge()).append("等。");
        }

        if (goals.size() > 0) {
            desc.append("目标：").append(getDominantGoal()).append("。");
        }

        if (completeness < 0.3) {
            desc.append("刚刚开始这一阶段。");
        } else if (completeness < 0.7) {
            desc.append("正在努力成长中。");
        } else {
            desc.append("接近阶段目标。");
        }

        return desc.toString();
    }

    // === Getter/Setter ===

    public String getStageId() { return stageId; }
    public void setStageId(String stageId) { this.stageId = stageId; }

    public String getStageName() { return stageName; }
    public void setStageName(String stageName) { this.stageName = stageName; }

    public String getStageLabel() { return stageLabel; }
    public void setStageLabel(String stageLabel) { this.stageLabel = stageLabel; }

    public String getDescription() { return description; }
    public void setDescription(String description) { this.description = description; }

    public String getGrowthFocus() { return growthFocus; }
    public void setGrowthFocus(String growthFocus) { this.growthFocus = growthFocus; }

    public List<String> getChallenges() { return challenges; }
    public void setChallenges(List<String> challenges) { this.challenges = challenges; }

    public List<String> getOpportunities() { return opportunities; }
    public void setOpportunities(List<String> opportunities) { this.opportunities = opportunities; }

    public List<String> getGoals() { return goals; }
    public void setGoals(List<String> goals) { this.goals = goals; }

    public List<String> getTasks() { return tasks; }
    public void setTasks(List<String> tasks) { this.tasks = tasks; }

    public int getStageAge() { return stageAge; }
    public void setStageAge(int stageAge) { this.stageAge = stageAge; }

    public double getCompleteness() { return completeness; }
    public void setCompleteness(double completeness) { this.completeness = clamp01(completeness); }

    public double getSatisfaction() { return satisfaction; }
    public void setSatisfaction(double satisfaction) { this.satisfaction = clamp01(satisfaction); }

    public DevelopmentMetrics getMetrics() { return metrics; }
    public void setMetrics(DevelopmentMetrics metrics) { this.metrics = metrics; }

    public StageInfluence getInfluence() { return influence; }
    public void setInfluence(StageInfluence influence) { this.influence = influence; }

    public TrajectorySummary getTrajectory() { return trajectory; }
    public void setTrajectory(TrajectorySummary trajectory) { this.trajectory = trajectory; }

    public List<String> getBottlenecks() { return bottlenecks; }
    public void setBottlenecks(List<String> bottlenecks) { this.bottlenecks = bottlenecks; }

    public long getTimestamp() { return timestamp; }
    public void setTimestamp(long timestamp) { this.timestamp = timestamp; }

    @NonNull
    @Override
    public String toString() {
        return "LifeStageState{" +
                "stage='" + getStageNameZh() + '\'' +
                ", age=" + stageAge +
                ", completeness=" + completeness +
                ", growthFocus='" + growthFocus + '\'' +
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
        JSONArray array = json.optJSONArray(key);
        if (array != null) {
            for (int i = 0; i < array.length(); i++) {
                result.add(array.optString(i));
            }
        }
        return result;
    }

    private JSONArray listToJson(List<String> list) {
        JSONArray array = new JSONArray();
        for (String s : list) {
            array.put(s);
        }
        return array;
    }

    /**
     * 发展指标
     */
    public static class DevelopmentMetrics {
        private double physicalHealth;
        private double mentalHealth;
        private double cognitiveGrowth;
        private double emotionalMaturity;
        private double socialDevelopment;
        private double relationshipQuality;
        private double careerProgress;
        private double financialStability;
        private double skillDevelopment;
        private double selfAwareness;
        private double purposeClarity;
        private double lifeSatisfaction;

        public DevelopmentMetrics() {
            this.physicalHealth = 0.8;
            this.mentalHealth = 0.7;
            this.cognitiveGrowth = 0.6;
            this.emotionalMaturity = 0.6;
            this.socialDevelopment = 0.6;
            this.relationshipQuality = 0.6;
            this.careerProgress = 0.5;
            this.financialStability = 0.5;
            this.skillDevelopment = 0.6;
            this.selfAwareness = 0.6;
            this.purposeClarity = 0.5;
            this.lifeSatisfaction = 0.6;
        }

        public static DevelopmentMetrics fromJson(JSONObject json) {
            DevelopmentMetrics m = new DevelopmentMetrics();
            m.physicalHealth = json.optDouble("physical_health", 0.8);
            m.mentalHealth = json.optDouble("mental_health", 0.7);
            m.cognitiveGrowth = json.optDouble("cognitive_growth", 0.6);
            m.emotionalMaturity = json.optDouble("emotional_maturity", 0.6);
            m.socialDevelopment = json.optDouble("social_development", 0.6);
            m.relationshipQuality = json.optDouble("relationship_quality", 0.6);
            m.careerProgress = json.optDouble("career_progress", 0.5);
            m.financialStability = json.optDouble("financial_stability", 0.5);
            m.skillDevelopment = json.optDouble("skill_development", 0.6);
            m.selfAwareness = json.optDouble("self_awareness", 0.6);
            m.purposeClarity = json.optDouble("purpose_clarity", 0.5);
            m.lifeSatisfaction = json.optDouble("life_satisfaction", 0.6);
            return m;
        }

        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("physical_health", physicalHealth);
                json.put("mental_health", mentalHealth);
                json.put("cognitive_growth", cognitiveGrowth);
                json.put("emotional_maturity", emotionalMaturity);
                json.put("social_development", socialDevelopment);
                json.put("relationship_quality", relationshipQuality);
                json.put("career_progress", careerProgress);
                json.put("financial_stability", financialStability);
                json.put("skill_development", skillDevelopment);
                json.put("self_awareness", selfAwareness);
                json.put("purpose_clarity", purposeClarity);
                json.put("life_satisfaction", lifeSatisfaction);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        // Getters
        public double getPhysicalHealth() { return physicalHealth; }
        public double getMentalHealth() { return mentalHealth; }
        public double getCognitiveGrowth() { return cognitiveGrowth; }
        public double getEmotionalMaturity() { return emotionalMaturity; }
        public double getSocialDevelopment() { return socialDevelopment; }
        public double getRelationshipQuality() { return relationshipQuality; }
        public double getCareerProgress() { return careerProgress; }
        public double getFinancialStability() { return financialStability; }
        public double getSkillDevelopment() { return skillDevelopment; }
        public double getSelfAwareness() { return selfAwareness; }
        public double getPurposeClarity() { return purposeClarity; }
        public double getLifeSatisfaction() { return lifeSatisfaction; }
    }

    /**
     * 阶段影响
     */
    public static class StageInfluence {
        private double challengeLevel;
        private double opportunityLevel;
        private String growthFocus;
        private String taskPriority;
        private double riskTaking;
        private double socialFocus;
        private double careerFocus;
        private double familyFocus;
        private double healthFocus;
        private String timePerspective;  // future/present/past
        private double urgencyLevel;
        private double patienceLevel;

        public StageInfluence() {
            this.riskTaking = 0.5;
            this.socialFocus = 0.5;
            this.careerFocus = 0.5;
            this.familyFocus = 0.5;
            this.healthFocus = 0.5;
            this.urgencyLevel = 0.5;
            this.patienceLevel = 0.5;
            this.timePerspective = "future";
        }

        public static StageInfluence fromJson(JSONObject json) {
            StageInfluence i = new StageInfluence();
            i.challengeLevel = json.optDouble("challenge_level", 0.5);
            i.opportunityLevel = json.optDouble("opportunity_level", 0.5);
            i.growthFocus = json.optString("growth_focus", "");
            i.taskPriority = json.optString("task_priority", "");
            i.riskTaking = json.optDouble("risk_taking", 0.5);
            i.socialFocus = json.optDouble("social_focus", 0.5);
            i.careerFocus = json.optDouble("career_focus", 0.5);
            i.familyFocus = json.optDouble("family_focus", 0.5);
            i.healthFocus = json.optDouble("health_focus", 0.5);
            i.timePerspective = json.optString("time_perspective", "future");
            i.urgencyLevel = json.optDouble("urgency_level", 0.5);
            i.patienceLevel = json.optDouble("patience_level", 0.5);
            return i;
        }

        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("challenge_level", challengeLevel);
                json.put("opportunity_level", opportunityLevel);
                json.put("growth_focus", growthFocus);
                json.put("task_priority", taskPriority);
                json.put("risk_taking", riskTaking);
                json.put("social_focus", socialFocus);
                json.put("career_focus", careerFocus);
                json.put("family_focus", familyFocus);
                json.put("health_focus", healthFocus);
                json.put("time_perspective", timePerspective);
                json.put("urgency_level", urgencyLevel);
                json.put("patience_level", patienceLevel);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        /**
         * 是否偏向冒险
         */
        public boolean isRiskTaking() {
            return riskTaking > 0.6;
        }

        /**
         * 是否偏向稳健
         */
        public boolean isConservative() {
            return riskTaking < 0.4;
        }

        /**
         * 是否重视事业
         */
        public boolean isCareerFocused() {
            return careerFocus > 0.6;
        }

        /**
         * 是否重视家庭
         */
        public boolean isFamilyFocused() {
            return familyFocus > 0.6;
        }

        /**
         * 是否重视健康
         */
        public boolean isHealthFocused() {
            return healthFocus > 0.6;
        }

        /**
         * 是否有紧迫感
         */
        public boolean hasUrgency() {
            return urgencyLevel > 0.6;
        }

        // Getters
        public double getChallengeLevel() { return challengeLevel; }
        public double getOpportunityLevel() { return opportunityLevel; }
        public String getGrowthFocus() { return growthFocus; }
        public String getTaskPriority() { return taskPriority; }
        public double getRiskTaking() { return riskTaking; }
        public double getSocialFocus() { return socialFocus; }
        public double getCareerFocus() { return careerFocus; }
        public double getFamilyFocus() { return familyFocus; }
        public double getHealthFocus() { return healthFocus; }
        public String getTimePerspective() { return timePerspective; }
        public double getUrgencyLevel() { return urgencyLevel; }
        public double getPatienceLevel() { return patienceLevel; }
    }

    /**
     * 轨迹摘要
     */
    public static class TrajectorySummary {
        private String direction;       // upward/stable/downward/fluctuating
        private double resilience;
        private double adaptability;
        private double wisdomLevel;
        private int milestoneCount;
        private int turningPointCount;
        private int lessonCount;

        public TrajectorySummary() {
            this.direction = "stable";
            this.resilience = 0.5;
            this.adaptability = 0.5;
            this.wisdomLevel = 0.3;
        }

        public static TrajectorySummary fromJson(JSONObject json) {
            TrajectorySummary t = new TrajectorySummary();
            t.direction = json.optString("direction", "stable");
            t.resilience = json.optDouble("resilience", 0.5);
            t.adaptability = json.optDouble("adaptability", 0.5);
            t.wisdomLevel = json.optDouble("wisdom_level", 0.3);
            t.milestoneCount = json.optInt("milestone_count", 0);
            t.turningPointCount = json.optInt("turning_point_count", 0);
            t.lessonCount = json.optInt("lesson_count", 0);
            return t;
        }

        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("direction", direction);
                json.put("resilience", resilience);
                json.put("adaptability", adaptability);
                json.put("wisdom_level", wisdomLevel);
                json.put("milestone_count", milestoneCount);
                json.put("turning_point_count", turningPointCount);
                json.put("lesson_count", lessonCount);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        /**
         * 是否上升轨迹
         */
        public boolean isUpward() {
            return "upward".equals(direction);
        }

        /**
         * 是否下降轨迹
         */
        public boolean isDownward() {
            return "downward".equals(direction);
        }

        /**
         * 是否高韧性
         */
        public boolean hasHighResilience() {
            return resilience > 0.7;
        }

        /**
         * 是否有智慧积累
         */
        public boolean hasWisdom() {
            return wisdomLevel > 0.5;
        }

        // Getters
        public String getDirection() { return direction; }
        public double getResilience() { return resilience; }
        public double getAdaptability() { return adaptability; }
        public double getWisdomLevel() { return wisdomLevel; }
        public int getMilestoneCount() { return milestoneCount; }
        public int getTurningPointCount() { return turningPointCount; }
        public int getLessonCount() { return lessonCount; }
    }
}