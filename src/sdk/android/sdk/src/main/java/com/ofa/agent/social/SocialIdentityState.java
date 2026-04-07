package com.ofa.agent.social;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;

/**
 * 社会身份状态模型 (v4.2.0)
 *
 * 端侧只接收 Center 推送的社会身份状态，用于调整决策倾向。
 * 深层社会身份管理在 Center 端 SocialIdentityEngine 完成。
 */
public class SocialIdentityState {

    // === 教育背景 ===
    private String educationLevel;      // 学历层次
    private String major;               // 专业
    private String majorCategory;       // 专业类别
    private String schoolTier;          // 学校层次
    private double academicPerformance; // 学业表现
    private boolean continuingEducation; // 是否持续学习

    // === 职业画像 ===
    private String occupation;          // 职业名称
    private String industry;            // 行业
    private String careerStage;         // 职业阶段
    private int yearsOfExperience;      // 工作年限
    private String employmentStatus;    // 就业状态
    private double jobSatisfaction;     // 工作满意度
    private double workLifeBalance;     // 工作生活平衡
    private double careerAmbition;      // 事业野心
    private String workMode;            // 工作模式

    // === 社会阶层 ===
    private String incomeLevel;         // 收入层次
    private double incomePercentile;    // 收入百分位
    private String socialStatus;        // 社会地位
    private double economicCapital;     // 经济资本
    private double culturalCapital;     // 文化资本
    private double socialCapital;       // 社会资本
    private String mobilityAspiration;  // 流动意愿

    // === 身份认同 ===
    private List<String> selfConceptLabels; // 自我概念标签
    private List<SocialRole> socialRoles;   // 社会角色
    private String dominantRole;            // 主导角色
    private double identityFluidity;        // 身份流动性

    // === 决策倾向 ===
    private SocialDecisionTendencies tendencies;

    // === 时间属性 ===
    private long timestamp;

    public SocialIdentityState() {
        // 默认值
        this.educationLevel = "bachelor";
        this.major = "";
        this.majorCategory = "other";
        this.schoolTier = "regular";
        this.academicPerformance = 0.6;
        this.continuingEducation = true;

        this.occupation = "";
        this.industry = "";
        this.careerStage = "developing";
        this.yearsOfExperience = 3;
        this.employmentStatus = "employed";
        this.jobSatisfaction = 0.6;
        this.workLifeBalance = 0.5;
        this.careerAmbition = 0.6;
        this.workMode = "hybrid";

        this.incomeLevel = "middle";
        this.incomePercentile = 50;
        this.socialStatus = "middle";
        this.economicCapital = 0.5;
        this.culturalCapital = 0.5;
        this.socialCapital = 0.5;
        this.mobilityAspiration = "upward";

        this.selfConceptLabels = new ArrayList<>();
        this.socialRoles = new ArrayList<>();
        this.dominantRole = "";
        this.identityFluidity = 0.5;

        this.tendencies = new SocialDecisionTendencies();
        this.timestamp = System.currentTimeMillis();
    }

    /**
     * 从 JSON 解析
     */
    @NonNull
    public static SocialIdentityState fromJson(@NonNull JSONObject json) throws JSONException {
        SocialIdentityState state = new SocialIdentityState();

        // 教育背景
        JSONObject education = json.optJSONObject("education");
        if (education != null) {
            state.educationLevel = education.optString("education_level", "bachelor");
            state.major = education.optString("major", "");
            state.majorCategory = education.optString("major_category", "other");
            state.schoolTier = education.optString("school_tier", "regular");
            state.academicPerformance = education.optDouble("academic_performance", 0.6);
            state.continuingEducation = education.optBoolean("continuing_education", true);
        }

        // 职业画像
        JSONObject career = json.optJSONObject("career");
        if (career != null) {
            state.occupation = career.optString("occupation", "");
            state.industry = career.optString("industry", "");
            state.careerStage = career.optString("career_stage", "developing");
            state.yearsOfExperience = career.optInt("years_of_experience", 3);
            state.employmentStatus = career.optString("employment_status", "employed");
            state.jobSatisfaction = career.optDouble("job_satisfaction", 0.6);
            state.workLifeBalance = career.optDouble("work_life_balance", 0.5);
            state.careerAmbition = career.optDouble("career_ambition", 0.6);
            state.workMode = career.optString("work_mode", "hybrid");
        }

        // 社会阶层
        JSONObject socialClass = json.optJSONObject("social_class");
        if (socialClass != null) {
            state.incomeLevel = socialClass.optString("income_level", "middle");
            state.incomePercentile = socialClass.optDouble("income_percentile", 50);
            state.socialStatus = socialClass.optString("social_status", "middle");
            state.economicCapital = socialClass.optDouble("economic_capital", 0.5);
            state.culturalCapital = socialClass.optDouble("cultural_capital", 0.5);
            state.socialCapital = socialClass.optDouble("social_capital", 0.5);
            state.mobilityAspiration = socialClass.optString("mobility_aspiration", "upward");
        }

        // 身份认同
        JSONObject identity = json.optJSONObject("identity");
        if (identity != null) {
            state.selfConceptLabels = parseStringList(identity, "self_concept_labels");
            state.socialRoles = parseSocialRoles(identity);
            state.identityFluidity = identity.optDouble("identity_fluidity", 0.5);
        }

        // 决策倾向
        JSONObject tendencies = json.optJSONObject("decision_tendencies");
        if (tendencies != null) {
            state.tendencies = SocialDecisionTendencies.fromJson(tendencies);
        }

        // 主导角色
        state.dominantRole = json.optString("dominant_role", "");

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
            // 教育背景
            JSONObject education = new JSONObject();
            education.put("education_level", educationLevel);
            education.put("major", major);
            education.put("major_category", majorCategory);
            education.put("school_tier", schoolTier);
            education.put("academic_performance", academicPerformance);
            education.put("continuing_education", continuingEducation);
            json.put("education", education);

            // 职业画像
            JSONObject career = new JSONObject();
            career.put("occupation", occupation);
            career.put("industry", industry);
            career.put("career_stage", careerStage);
            career.put("years_of_experience", yearsOfExperience);
            career.put("employment_status", employmentStatus);
            career.put("job_satisfaction", jobSatisfaction);
            career.put("work_life_balance", workLifeBalance);
            career.put("career_ambition", careerAmbition);
            career.put("work_mode", workMode);
            json.put("career", career);

            // 社会阶层
            JSONObject socialClass = new JSONObject();
            socialClass.put("income_level", incomeLevel);
            socialClass.put("income_percentile", incomePercentile);
            socialClass.put("social_status", socialStatus);
            socialClass.put("economic_capital", economicCapital);
            socialClass.put("cultural_capital", culturalCapital);
            socialClass.put("social_capital", socialCapital);
            socialClass.put("mobility_aspiration", mobilityAspiration);
            json.put("social_class", socialClass);

            // 身份认同
            JSONObject identity = new JSONObject();
            identity.put("self_concept_labels", listToJson(selfConceptLabels));
            identity.put("identity_fluidity", identityFluidity);
            json.put("identity", identity);

            // 决策倾向
            if (tendencies != null) {
                json.put("decision_tendencies", tendencies.toJson());
            }

            json.put("dominant_role", dominantRole);
            json.put("timestamp", timestamp);

        } catch (JSONException e) {
            // ignore
        }
        return json;
    }

    // === 描述方法 ===

    /**
     * 获取学历名称
     */
    @NonNull
    public String getEducationLevelName() {
        switch (educationLevel) {
            case "high_school": return "高中";
            case "associate": return "大专";
            case "bachelor": return "本科";
            case "master": return "硕士";
            case "doctorate": return "博士";
            default: return educationLevel;
        }
    }

    /**
     * 获取职业阶段名称
     */
    @NonNull
    public String getCareerStageName() {
        switch (careerStage) {
            case "entry": return "入门期";
            case "developing": return "发展期";
            case "established": return "成熟期";
            case "senior": return "资深期";
            case "executive": return "高管期";
            default: return careerStage;
        }
    }

    /**
     * 获取社会地位名称
     */
    @NonNull
    public String getSocialStatusName() {
        switch (socialStatus) {
            case "marginal": return "边缘阶层";
            case "working": return "工薪阶层";
            case "middle": return "中产阶层";
            case "professional": return "专业阶层";
            case "elite": return "精英阶层";
            default: return socialStatus;
        }
    }

    /**
     * 是否高学历
     */
    public boolean isHighlyEducated() {
        return "master".equals(educationLevel) || "doctorate".equals(educationLevel);
    }

    /**
     * 是否资深职业人
     */
    public boolean isSeniorProfessional() {
        return "senior".equals(careerStage) || "executive".equals(careerStage);
    }

    /**
     * 是否高收入
     */
    public boolean isHighIncome() {
        return "upper_middle".equals(incomeLevel) || "upper".equals(incomeLevel);
    }

    /**
     * 是否有向上流动意愿
     */
    public boolean hasUpwardMobility() {
        return "upward".equals(mobilityAspiration);
    }

    /**
     * 是否工作导向
     */
    public boolean isWorkOriented() {
        return workLifeBalance < 0.5;
    }

    /**
     * 获取综合资本
     */
    public double getOverallCapital() {
        return (economicCapital + culturalCapital + socialCapital) / 3;
    }

    // === Getter/Setter ===

    public String getEducationLevel() { return educationLevel; }
    public void setEducationLevel(String educationLevel) { this.educationLevel = educationLevel; }

    public String getMajor() { return major; }
    public void setMajor(String major) { this.major = major; }

    public String getMajorCategory() { return majorCategory; }
    public void setMajorCategory(String majorCategory) { this.majorCategory = majorCategory; }

    public String getSchoolTier() { return schoolTier; }
    public void setSchoolTier(String schoolTier) { this.schoolTier = schoolTier; }

    public double getAcademicPerformance() { return academicPerformance; }
    public void setAcademicPerformance(double academicPerformance) { this.academicPerformance = clamp01(academicPerformance); }

    public boolean isContinuingEducation() { return continuingEducation; }
    public void setContinuingEducation(boolean continuingEducation) { this.continuingEducation = continuingEducation; }

    public String getOccupation() { return occupation; }
    public void setOccupation(String occupation) { this.occupation = occupation; }

    public String getIndustry() { return industry; }
    public void setIndustry(String industry) { this.industry = industry; }

    public String getCareerStage() { return careerStage; }
    public void setCareerStage(String careerStage) { this.careerStage = careerStage; }

    public int getYearsOfExperience() { return yearsOfExperience; }
    public void setYearsOfExperience(int yearsOfExperience) { this.yearsOfExperience = yearsOfExperience; }

    public String getEmploymentStatus() { return employmentStatus; }
    public void setEmploymentStatus(String employmentStatus) { this.employmentStatus = employmentStatus; }

    public double getJobSatisfaction() { return jobSatisfaction; }
    public void setJobSatisfaction(double jobSatisfaction) { this.jobSatisfaction = clamp01(jobSatisfaction); }

    public double getWorkLifeBalance() { return workLifeBalance; }
    public void setWorkLifeBalance(double workLifeBalance) { this.workLifeBalance = clamp01(workLifeBalance); }

    public double getCareerAmbition() { return careerAmbition; }
    public void setCareerAmbition(double careerAmbition) { this.careerAmbition = clamp01(careerAmbition); }

    public String getWorkMode() { return workMode; }
    public void setWorkMode(String workMode) { this.workMode = workMode; }

    public String getIncomeLevel() { return incomeLevel; }
    public void setIncomeLevel(String incomeLevel) { this.incomeLevel = incomeLevel; }

    public double getIncomePercentile() { return incomePercentile; }
    public void setIncomePercentile(double incomePercentile) { this.incomePercentile = clamp(incomePercentile, 0, 100); }

    public String getSocialStatus() { return socialStatus; }
    public void setSocialStatus(String socialStatus) { this.socialStatus = socialStatus; }

    public double getEconomicCapital() { return economicCapital; }
    public void setEconomicCapital(double economicCapital) { this.economicCapital = clamp01(economicCapital); }

    public double getCulturalCapital() { return culturalCapital; }
    public void setCulturalCapital(double culturalCapital) { this.culturalCapital = clamp01(culturalCapital); }

    public double getSocialCapital() { return socialCapital; }
    public void setSocialCapital(double socialCapital) { this.socialCapital = clamp01(socialCapital); }

    public String getMobilityAspiration() { return mobilityAspiration; }
    public void setMobilityAspiration(String mobilityAspiration) { this.mobilityAspiration = mobilityAspiration; }

    public List<String> getSelfConceptLabels() { return selfConceptLabels; }
    public void setSelfConceptLabels(List<String> selfConceptLabels) { this.selfConceptLabels = selfConceptLabels; }

    public List<SocialRole> getSocialRoles() { return socialRoles; }
    public void setSocialRoles(List<SocialRole> socialRoles) { this.socialRoles = socialRoles; }

    public String getDominantRole() { return dominantRole; }
    public void setDominantRole(String dominantRole) { this.dominantRole = dominantRole; }

    public double getIdentityFluidity() { return identityFluidity; }
    public void setIdentityFluidity(double identityFluidity) { this.identityFluidity = clamp01(identityFluidity); }

    public SocialDecisionTendencies getTendencies() { return tendencies; }
    public void setTendencies(SocialDecisionTendencies tendencies) { this.tendencies = tendencies; }

    public long getTimestamp() { return timestamp; }
    public void setTimestamp(long timestamp) { this.timestamp = timestamp; }

    @NonNull
    @Override
    public String toString() {
        return "SocialIdentityState{" +
                "education='" + getEducationLevelName() + '\'' +
                ", occupation='" + occupation + '\'' +
                ", stage='" + getCareerStageName() + '\'' +
                ", status='" + getSocialStatusName() + '\'' +
                '}';
    }

    // === 辅助方法 ===

    private double clamp01(double value) {
        return clamp(value, 0, 1);
    }

    private double clamp(double value, double min, double max) {
        if (value < min) return min;
        if (value > max) return max;
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

    private static List<SocialRole> parseSocialRoles(JSONObject json) {
        List<SocialRole> result = new ArrayList<>();
        JSONArray array = json.optJSONArray("social_roles");
        if (array != null) {
            for (int i = 0; i < array.length(); i++) {
                JSONObject roleJson = array.optJSONObject(i);
                if (roleJson != null) {
                    result.add(SocialRole.fromJson(roleJson));
                }
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
     * 社会角色
     */
    public static class SocialRole {
        private String roleId;
        private String roleName;
        private String roleCategory;
        private double importance;
        private double satisfaction;
        private double timeSpent;

        public static SocialRole fromJson(JSONObject json) {
            SocialRole role = new SocialRole();
            role.roleId = json.optString("role_id", "");
            role.roleName = json.optString("role_name", "");
            role.roleCategory = json.optString("role_category", "");
            role.importance = json.optDouble("importance", 0.5);
            role.satisfaction = json.optDouble("satisfaction", 0.5);
            role.timeSpent = json.optDouble("time_spent", 0.5);
            return role;
        }

        public String getRoleId() { return roleId; }
        public String getRoleName() { return roleName; }
        public String getRoleCategory() { return roleCategory; }
        public double getImportance() { return importance; }
        public double getSatisfaction() { return satisfaction; }
        public double getTimeSpent() { return timeSpent; }
    }

    /**
     * 社会决策倾向
     */
    public static class SocialDecisionTendencies {
        private double cognitiveComplexity;
        private double adaptability;
        private double workPriority;
        private double careerRiskTaking;
        private double financialRiskTolerance;
        private double upwardMobilityDrive;
        private double socialLeverage;
        private double decisionConsistency;

        public SocialDecisionTendencies() {
            this.cognitiveComplexity = 0.5;
            this.adaptability = 0.5;
            this.workPriority = 0.5;
            this.careerRiskTaking = 0.5;
            this.financialRiskTolerance = 0.4;
            this.upwardMobilityDrive = 0.5;
            this.socialLeverage = 0.5;
            this.decisionConsistency = 0.5;
        }

        public static SocialDecisionTendencies fromJson(JSONObject json) {
            SocialDecisionTendencies t = new SocialDecisionTendencies();
            t.cognitiveComplexity = json.optDouble("cognitive_complexity", 0.5);
            t.adaptability = json.optDouble("adaptability", 0.5);
            t.workPriority = json.optDouble("work_priority", 0.5);
            t.careerRiskTaking = json.optDouble("career_risk_taking", 0.5);
            t.financialRiskTolerance = json.optDouble("financial_risk_tolerance", 0.4);
            t.upwardMobilityDrive = json.optDouble("upward_mobility_drive", 0.5);
            t.socialLeverage = json.optDouble("social_leverage", 0.5);
            t.decisionConsistency = json.optDouble("decision_consistency", 0.5);
            return t;
        }

        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("cognitive_complexity", cognitiveComplexity);
                json.put("adaptability", adaptability);
                json.put("work_priority", workPriority);
                json.put("career_risk_taking", careerRiskTaking);
                json.put("financial_risk_tolerance", financialRiskTolerance);
                json.put("upward_mobility_drive", upwardMobilityDrive);
                json.put("social_leverage", socialLeverage);
                json.put("decision_consistency", decisionConsistency);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public double getCognitiveComplexity() { return cognitiveComplexity; }
        public double getAdaptability() { return adaptability; }
        public double getWorkPriority() { return workPriority; }
        public double getCareerRiskTaking() { return careerRiskTaking; }
        public double getFinancialRiskTolerance() { return financialRiskTolerance; }
        public double getUpwardMobilityDrive() { return upwardMobilityDrive; }
        public double getSocialLeverage() { return socialLeverage; }
        public double getDecisionConsistency() { return decisionConsistency; }
    }
}