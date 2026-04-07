package com.ofa.agent.identity;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.List;

/**
 * DecisionContext - 决策上下文
 *
 * 提供身份相关的决策上下文，用于 AI 决策引擎。
 */
public class DecisionContext {

    private String identityId;
    private Personality personality;
    private ValueSystem valueSystem;
    private List<Interest> interests;
    private String speakingTone;
    private String responseLength;
    private String[] valuePriority;
    private PersonalIdentity identity;

    /**
     * 创建决策上下文
     */
    public DecisionContext(@NonNull PersonalIdentity identity) {
        this.identity = identity;
        this.identityId = identity.getId();

        this.personality = identity.getPersonality();
        this.valueSystem = identity.getValueSystem();
        this.interests = identity.getInterests();

        if (this.personality != null) {
            this.speakingTone = personality.getSpeakingTone();
            this.responseLength = personality.getResponseLength();
        } else {
            this.speakingTone = Personality.TONE_CASUAL;
            this.responseLength = Personality.LENGTH_MODERATE;
        }

        if (this.valueSystem != null) {
            this.valuePriority = valueSystem.getValuePriority();
        } else {
            this.valuePriority = new String[0];
        }
    }

    // === Getters ===

    public String getIdentityId() { return identityId; }
    public Personality getPersonality() { return personality; }
    public ValueSystem getValueSystem() { return valueSystem; }
    public List<Interest> getInterests() { return interests; }
    public String getSpeakingTone() { return speakingTone; }
    public String getResponseLength() { return responseLength; }
    public String[] getValuePriority() { return valuePriority; }

    /**
     * 获取是否偏好隐私
     */
    public boolean prefersPrivacy() {
        if (valueSystem == null) return true;
        return valueSystem.getPrivacy() > 0.6;
    }

    /**
     * 获取是否偏好效率
     */
    public boolean prefersEfficiency() {
        if (valueSystem == null) return false;
        return valueSystem.getEfficiency() > 0.6;
    }

    /**
     * 获取是否偏好详细回复
     */
    public boolean prefersDetailedResponse() {
        return responseLength.equals(Personality.LENGTH_DETAILED);
    }

    /**
     * 获取是否偏好简短回复
     */
    public boolean prefersBriefResponse() {
        return responseLength.equals(Personality.LENGTH_BRIEF);
    }

    /**
     * 获取是否允许表情
     */
    public boolean allowsEmoji() {
        if (personality == null) return true;
        return personality.getEmojiUsage() > 0.3;
    }

    /**
     * 获取风险承受度
     */
    public double getRiskTolerance() {
        if (valueSystem == null) return 0.4;
        return valueSystem.getRiskTolerance();
    }

    /**
     * 生成 AI Prompt 上下文
     */
    @NonNull
    public String generatePromptContext() {
        StringBuilder sb = new StringBuilder();

        sb.append("【用户身份上下文】\n");

        // 性格
        if (personality != null) {
            sb.append("性格特征: ").append(personality.getDescription()).append("\n");
            sb.append("沟通风格: ").append(speakingTone).append("\n");
            sb.append("回复长度偏好: ").append(responseLength).append("\n");
            if (allowsEmoji()) {
                sb.append("可以适当使用表情符号\n");
            }
        }

        // 价值观
        if (valueSystem != null && valuePriority.length > 0) {
            sb.append("价值观排序: ");
            for (int i = 0; i < Math.min(3, valuePriority.length); i++) {
                if (i > 0) sb.append(" > ");
                sb.append(valuePriority[i]);
            }
            sb.append("\n");

            if (prefersPrivacy()) {
                sb.append("注意: 用户高度重视隐私，避免泄露敏感信息\n");
            }
        }

        // 兴趣
        if (interests != null && !interests.isEmpty()) {
            List<Interest> topInterests = identity.getTopInterests(3);
            sb.append("主要兴趣: ");
            for (int i = 0; i < topInterests.size(); i++) {
                if (i > 0) sb.append(", ");
                sb.append(topInterests.get(i).getName());
            }
            sb.append("\n");
        }

        return sb.toString();
    }
}