package com.ofa.agent.identity;

import androidx.annotation.NonNull;

import java.util.ArrayList;
import java.util.List;

/**
 * WritingStyle - 写作风格配置
 *
 * 描述用户的写作风格，用于文本生成个性化。
 */
public class WritingStyle {

    // 风格参数
    private double formality;      // 正式程度 (0-1)
    private double verbosity;      // 冗长程度 (0-1)
    private double humor;          // 幽默程度 (0-1)
    private double technicality;   // 专业程度 (0-1)

    // 文风特征
    private boolean useEmoji;
    private boolean useGIFs;
    private boolean useMarkdown;

    // 标志性用语
    private String signaturePhrase;

    // 常用词汇
    private List<String> frequentWords;
    private List<String> avoidWords;

    // 语言习惯
    private String preferredGreeting;  // 偏好的问候语
    private String preferredClosing;   // 偏好的结束语

    /**
     * 创建默认写作风格
     */
    public WritingStyle() {
        this.formality = 0.4;
        this.verbosity = 0.5;
        this.humor = 0.3;
        this.technicality = 0.5;

        this.useEmoji = true;
        this.useGIFs = false;
        this.useMarkdown = true;

        this.signaturePhrase = "";
        this.frequentWords = new ArrayList<>();
        this.avoidWords = new ArrayList<>();

        this.preferredGreeting = "你好";
        this.preferredClosing = "祝好";
    }

    // === 更新方法 ===

    /**
     * 设置风格参数
     */
    public void setStyleParams(double formality, double verbosity, double humor, double technicality) {
        this.formality = clamp01(formality);
        this.verbosity = clamp01(verbosity);
        this.humor = clamp01(humor);
        this.technicality = clamp01(technicality);
    }

    /**
     * 添加常用词
     */
    public void addFrequentWord(@NonNull String word) {
        if (!frequentWords.contains(word)) {
            frequentWords.add(word);
        }
    }

    /**
     * 添加避免词
     */
    public void addAvoidWord(@NonNull String word) {
        if (!avoidWords.contains(word)) {
            avoidWords.add(word);
        }
    }

    // === 辅助方法 ===

    private double clamp01(double value) {
        if (value < 0) return 0;
        if (value > 1) return 1;
        return value;
    }

    // === Getters ===

    public double getFormality() { return formality; }
    public double getVerbosity() { return verbosity; }
    public double getHumor() { return humor; }
    public double getTechnicality() { return technicality; }

    public boolean isUseEmoji() { return useEmoji; }
    public boolean isUseGIFs() { return useGIFs; }
    public boolean isUseMarkdown() { return useMarkdown; }

    public String getSignaturePhrase() { return signaturePhrase; }
    public List<String> getFrequentWords() { return new ArrayList<>(frequentWords); }
    public List<String> getAvoidWords() { return new ArrayList<>(avoidWords); }

    public String getPreferredGreeting() { return preferredGreeting; }
    public String getPreferredClosing() { return preferredClosing; }

    // === Setters ===

    public void setFormality(double formality) { this.formality = clamp01(formality); }
    public void setVerbosity(double verbosity) { this.verbosity = clamp01(verbosity); }
    public void setHumor(double humor) { this.humor = clamp01(humor); }
    public void setTechnicality(double technicality) { this.technicality = clamp01(technicality); }

    public void setUseEmoji(boolean useEmoji) { this.useEmoji = useEmoji; }
    public void setUseGIFs(boolean useGIFs) { this.useGIFs = useGIFs; }
    public void setUseMarkdown(boolean useMarkdown) { this.useMarkdown = useMarkdown; }

    public void setSignaturePhrase(String signaturePhrase) { this.signaturePhrase = signaturePhrase; }
    public void setFrequentWords(List<String> frequentWords) { this.frequentWords = new ArrayList<>(frequentWords); }
    public void setAvoidWords(List<String> avoidWords) { this.avoidWords = new ArrayList<>(avoidWords); }

    public void setPreferredGreeting(String greeting) { this.preferredGreeting = greeting; }
    public void setPreferredClosing(String closing) { this.preferredClosing = closing; }

    /**
     * 转换为 JSON 字符串
     */
    @NonNull
    public String toJson() {
        StringBuilder sb = new StringBuilder();
        sb.append("{");
        sb.append("\"formality\":").append(formality).append(",");
        sb.append("\"verbosity\":").append(verbosity).append(",");
        sb.append("\"humor\":").append(humor).append(",");
        sb.append("\"technicality\":").append(technicality).append(",");
        sb.append("\"use_emoji\":").append(useEmoji).append(",");
        sb.append("\"use_markdown\":").append(useMarkdown).append(",");
        sb.append("\"preferred_greeting\":\"").append(preferredGreeting).append("\",");
        sb.append("\"preferred_closing\":\"").append(preferredClosing).append("\"");
        sb.append("}");
        return sb.toString();
    }
}