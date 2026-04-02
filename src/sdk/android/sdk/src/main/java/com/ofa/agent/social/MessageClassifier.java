package com.ofa.agent.social;

import android.content.Context;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.memory.UserMemoryManager;

import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Message Classifier - analyzes message content and determines:
 * 1. Message type (invitation, urgent, reminder, guide, payment, casual, business)
 * 2. Urgency level (low, medium, high, critical)
 * 3. Preferred delivery channel
 *
 * Uses pattern matching and user preferences for classification.
 */
public class MessageClassifier {

    private final Context context;
    private final UserMemoryManager memoryManager;

    // Message type patterns
    private final Map<String, List<String>> typePatterns;
    private final Map<String, Integer> urgencyKeywords;

    // User preferences for channels
    private final Map<String, String> userChannelPreferences;

    /**
     * Message type constants
     */
    public static final String TYPE_INVITATION = "invitation";
    public static final String TYPE_URGENT = "urgent";
    public static final String TYPE_REMINDER = "reminder";
    public static final String TYPE_GUIDE = "guide";
    public static final String TYPE_PAYMENT = "payment";
    public static final String TYPE_CASUAL = "casual";
    public static final String TYPE_BUSINESS = "business";
    public static final String TYPE_GREETING = "greeting";
    public static final String TYPE_LOCATION = "location";
    public static final String TYPE_UNKNOWN = "unknown";

    /**
     * Urgency level constants
     */
    public static final int URGENCY_LOW = 1;
    public static final int URGENCY_MEDIUM = 2;
    public static final int URGENCY_HIGH = 3;
    public static final int URGENCY_CRITICAL = 4;

    /**
     * Classification result
     */
    public static class ClassificationResult {
        public final String messageType;
        public final int urgencyLevel;
        public final String recommendedChannel;
        public final double confidence;
        public final Map<String, String> extractedInfo;
        public final String reason;

        public ClassificationResult(String messageType, int urgencyLevel,
                                     String recommendedChannel, double confidence,
                                     Map<String, String> extractedInfo, String reason) {
            this.messageType = messageType;
            this.urgencyLevel = urgencyLevel;
            this.recommendedChannel = recommendedChannel;
            this.confidence = confidence;
            this.extractedInfo = extractedInfo;
            this.reason = reason;
        }

        @NonNull
        @Override
        public String toString() {
            return String.format("Classification{type=%s, urgency=%d, channel=%s, conf=%.2f}",
                messageType, urgencyLevel, recommendedChannel, confidence);
        }
    }

    /**
     * Create message classifier
     */
    public MessageClassifier(@NonNull Context context,
                              @Nullable UserMemoryManager memoryManager) {
        this.context = context;
        this.memoryManager = memoryManager;
        this.typePatterns = new HashMap<>();
        this.urgencyKeywords = new HashMap<>();
        this.userChannelPreferences = new HashMap<>();

        initializePatterns();
        loadUserPreferences();
    }

    /**
     * Initialize message type patterns
     */
    private void initializePatterns() {
        // Invitation patterns - social invitations
        typePatterns.put(TYPE_INVITATION, List.of(
            "约", "一起", "吃饭", "聚会", "活动", "聚餐", "派对",
            "喝", "玩", "看电影", "逛街", "旅游", "周末",
            "有空吗", "方便吗", "要不要", "来不来",
            "邀请", "欢迎", "参加"
        ));

        // Urgent patterns - need immediate response
        typePatterns.put(TYPE_URGENT, List.of(
            "紧急", "急", "马上", "立即", "立刻", "快点",
            "重要", "关键", "危险", "事故", "急救",
            "报警", "求助", "救命", "通知",
            "截止", "限时", "最后", "错过", "取消"
        ));

        // Reminder patterns - scheduled reminders
        typePatterns.put(TYPE_REMINDER, List.of(
            "提醒", "记得", "别忘了", "记住",
            "明天", "后天", "下周", "几点",
            "会议", "预约", "挂号", "缴费",
            "到期", "续费", "还款", "账单"
        ));

        // Guide patterns - tips and tutorials
        typePatterns.put(TYPE_GUIDE, List.of(
            "攻略", "教程", "方法", "技巧", "经验",
            "推荐", "分享", "收藏", "建议", "指南",
            "怎么", "如何", "步骤", "流程",
            "秘籍", "心得", "总结"
        ));

        // Payment patterns - financial messages
        typePatterns.put(TYPE_PAYMENT, List.of(
            "转账", "付款", "支付", "还钱", "收款",
            "红包", "金额", "钱", "元", "块",
            "账单", "消费", "花费", "借", "还",
            "报销", "结算", "工资", "奖金"
        ));

        // Casual patterns - daily chatting
        typePatterns.put(TYPE_CASUAL, List.of(
            "哈", "哈哈", "呵呵", "哎", "嗯",
            "无聊", "闲", "随便", "聊聊",
            "搞笑", "有趣", "好玩", "可爱",
            "晚安", "早安", "祝福", "问候"
        ));

        // Business patterns - work related
        typePatterns.put(TYPE_BUSINESS, List.of(
            "工作", "任务", "项目", "报告", "文档",
            "会议", "审批", "流程", "进度",
            "老板", "领导", "同事", "客户",
            "邮件", "回复", "确认", "完成"
        ));

        // Greeting patterns
        typePatterns.put(TYPE_GREETING, List.of(
            "你好", "您好", "嗨", "在吗", "在不在",
            "早上好", "晚上好", "好久不见"
        ));

        // Location patterns
        typePatterns.put(TYPE_LOCATION, List.of(
            "位置", "地址", "在哪", "哪里", "地方",
            "定位", "地图", "导航", "过来", "到达"
        ));

        // Urgency keywords mapping
        urgencyKeywords.put("紧急", URGENCY_CRITICAL);
        urgencyKeywords.put("急", URGENCY_CRITICAL);
        urgencyKeywords.put("救命", URGENCY_CRITICAL);
        urgencyKeywords.put("事故", URGENCY_CRITICAL);
        urgencyKeywords.put("危险", URGENCY_CRITICAL);

        urgencyKeywords.put("马上", URGENCY_HIGH);
        urgencyKeywords.put("立即", URGENCY_HIGH);
        urgencyKeywords.put("立刻", URGENCY_HIGH);
        urgencyKeywords.put("重要", URGENCY_HIGH);
        urgencyKeywords.put("快点", URGENCY_HIGH);

        urgencyKeywords.put("明天", URGENCY_MEDIUM);
        urgencyKeywords.put("后天", URGENCY_MEDIUM);
        urgencyKeywords.put("下周", URGENCY_MEDIUM);
        urgencyKeywords.put("记得", URGENCY_MEDIUM);

        urgencyKeywords.put("有空", URGENCY_LOW);
        urgencyKeywords.put("方便", URGENCY_LOW);
        urgencyKeywords.put("一起", URGENCY_LOW);
        urgencyKeywords.put("周末", URGENCY_LOW);
    }

    /**
     * Load user preferences from memory
     */
    private void loadUserPreferences() {
        if (memoryManager == null) return;

        // Load channel preferences for each contact
        List<UserMemoryManager.MemorySuggestion> prefs =
            memoryManager.getTopValues("social.channel", 100);

        for (UserMemoryManager.MemorySuggestion p : prefs) {
            String key = p.key.replace("social.channel.", "");
            userChannelPreferences.put(key, p.value);
        }
    }

    /**
     * Classify message content
     */
    @NonNull
    public ClassificationResult classify(@NonNull String message) {
        return classify(message, null, null);
    }

    /**
     * Classify message with contact and time context
     */
    @NonNull
    public ClassificationResult classify(@NonNull String message,
                                          @Nullable String contactName,
                                          @Nullable String contactPhone) {
        String messageType = TYPE_UNKNOWN;
        int urgencyLevel = URGENCY_LOW;
        double maxScore = 0;
        Map<String, String> extractedInfo = new HashMap<>();
        String reason = "";

        // Score each message type
        for (Map.Entry<String, List<String>> entry : typePatterns.entrySet()) {
            String type = entry.getKey();
            List<String> patterns = entry.getValue();

            double score = calculateTypeScore(message, patterns);
            if (score > maxScore) {
                maxScore = score;
                messageType = type;
                reason = "Matched patterns for " + type;
            }
        }

        // Calculate urgency level
        urgencyLevel = calculateUrgency(message);

        // Extract information from message
        extractedInfo.putAll(extractInfo(message, messageType));

        // Determine recommended channel
        String recommendedChannel = recommendChannel(messageType, urgencyLevel,
            contactName, contactPhone, extractedInfo);

        // Calculate confidence
        double confidence = Math.min(1.0, maxScore / 3.0);

        // Override with user preferences if available
        if (contactName != null && userChannelPreferences.containsKey(contactName)) {
            String preferred = userChannelPreferences.get(contactName);
            if (preferred != null) {
                recommendedChannel = preferred;
                reason = "User preference for " + contactName;
                confidence = 1.0;
            }
        }

        return new ClassificationResult(messageType, urgencyLevel,
            recommendedChannel, confidence, extractedInfo, reason);
    }

    /**
     * Calculate type score based on pattern matching
     */
    private double calculateTypeScore(@NonNull String message, @NonNull List<String> patterns) {
        double score = 0;
        String lowerMessage = message.toLowerCase();

        for (String pattern : patterns) {
            if (lowerMessage.contains(pattern.toLowerCase())) {
                score += 1;
                // Bonus for multiple matches
                if (score > 1) {
                    score += 0.5;
                }
            }
        }

        return score;
    }

    /**
     * Calculate urgency level from message
     */
    private int calculateUrgency(@NonNull String message) {
        int maxUrgency = URGENCY_LOW;
        String lowerMessage = message.toLowerCase();

        for (Map.Entry<String, Integer> entry : urgencyKeywords.entrySet()) {
            if (lowerMessage.contains(entry.getKey().toLowerCase())) {
                maxUrgency = Math.max(maxUrgency, entry.getValue());
            }
        }

        // Time-based urgency indicators
        if (lowerMessage.contains("现在") || lowerMessage.contains("今天")) {
            maxUrgency = Math.max(maxUrgency, URGENCY_HIGH);
        }

        return maxUrgency;
    }

    /**
     * Extract relevant information from message
     */
    @NonNull
    private Map<String, String> extractInfo(@NonNull String message, @NonNull String type) {
        Map<String, String> info = new HashMap<>();
        String lowerMessage = message.toLowerCase();

        // Extract time information
        if (lowerMessage.contains("明天")) {
            info.put("time", "明天");
        } else if (lowerMessage.contains("后天")) {
            info.put("time", "后天");
        } else if (lowerMessage.contains("周末")) {
            info.put("time", "周末");
        } else if (lowerMessage.contains("今天")) {
            info.put("time", "今天");
        }

        // Extract activity for invitations
        if (type.equals(TYPE_INVITATION)) {
            if (lowerMessage.contains("吃饭") || lowerMessage.contains("聚餐")) {
                info.put("activity", "吃饭");
            } else if (lowerMessage.contains("喝") || lowerMessage.contains("咖啡")) {
                info.put("activity", "喝咖啡");
            } else if (lowerMessage.contains("电影") || lowerMessage.contains("看电影")) {
                info.put("activity", "看电影");
            } else if (lowerMessage.contains("玩") || lowerMessage.contains("逛街")) {
                info.put("activity", "游玩");
            }
        }

        // Extract amount for payment
        if (type.equals(TYPE_PAYMENT)) {
            // Simple amount extraction - look for numbers
            java.util.regex.Pattern pattern = java.util.regex.Pattern.compile("(\\d+)[元块]");
            java.util.regex.Matcher matcher = pattern.matcher(message);
            if (matcher.find()) {
                info.put("amount", matcher.group(1));
            }
        }

        return info;
    }

    /**
     * Recommend best channel for message delivery
     *
     * Channel selection logic based on modern social habits:
     * - Urgent/Critical → Phone call (immediate response needed)
     * - Invitation → WeChat (social, casual, easy to discuss)
     * - Guide/Tips → Red (Xiaohongshu)私信 (content sharing platform)
     * - Payment → Alipay (financial platform)
     * - Business → DingTalk/WeCom (work platforms)
     * - Reminder → SMS (reliable delivery) or WeChat
     * - Casual → WeChat/Douyin (social platforms)
     */
    @NonNull
    private String recommendChannel(@NonNull String messageType, int urgencyLevel,
                                     @Nullable String contactName,
                                     @Nullable String contactPhone,
                                     @NonNull Map<String, String> extractedInfo) {
        // Priority 1: Urgency-based selection
        if (urgencyLevel >= URGENCY_CRITICAL) {
            return ChannelSelector.CHANNEL_PHONE; // Phone call for critical
        }

        if (urgencyLevel >= URGENCY_HIGH) {
            // High urgency: phone if available, otherwise WeChat
            if (contactPhone != null && !contactPhone.isEmpty()) {
                return ChannelSelector.CHANNEL_PHONE;
            }
            return ChannelSelector.CHANNEL_WECHAT;
        }

        // Priority 2: Message type-based selection
        switch (messageType) {
            case TYPE_INVITATION:
                // Social invitation - WeChat for discussion
                return ChannelSelector.CHANNEL_WECHAT;

            case TYPE_GUIDE:
                // Tips/guides - Red (Xiaohongshu) for content sharing
                return ChannelSelector.CHANNEL_XIAOHONGSHU;

            case TYPE_PAYMENT:
                // Financial - Alipay
                return ChannelSelector.CHANNEL_ALIPAY;

            case TYPE_BUSINESS:
                // Work-related - DingTalk or WeCom
                return ChannelSelector.CHANNEL_DINGTALK;

            case TYPE_REMINDER:
                // Reminder - SMS for reliability
                return ChannelSelector.CHANNEL_SMS;

            case TYPE_CASUAL:
            case TYPE_GREETING:
                // Casual chat - WeChat
                return ChannelSelector.CHANNEL_WECHAT;

            case TYPE_LOCATION:
                // Location sharing - WeChat (has location feature)
                return ChannelSelector.CHANNEL_WECHAT;

            case TYPE_URGENT:
                return ChannelSelector.CHANNEL_PHONE;

            default:
                // Default to WeChat
                return ChannelSelector.CHANNEL_WECHAT;
        }
    }

    /**
     * Save user channel preference
     */
    public void saveChannelPreference(@NonNull String contactName, @NonNull String channel) {
        userChannelPreferences.put(contactName, channel);

        if (memoryManager != null) {
            memoryManager.set("social.channel." + contactName, channel);
        }
    }

    /**
     * Get user channel preference
     */
    @Nullable
    public String getChannelPreference(@NonNull String contactName) {
        return userChannelPreferences.get(contactName);
    }

    /**
     * Get all supported message types
     */
    @NonNull
    public List<String> getMessageTypes() {
        return new ArrayList<>(typePatterns.keySet());
    }
}