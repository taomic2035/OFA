package com.ofa.agent.social;

import android.content.Context;
import android.content.Intent;
import android.content.pm.PackageManager;
import android.content.pm.ResolveInfo;
import android.net.Uri;
import android.telephony.SmsManager;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.memory.UserMemoryManager;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Channel Selector - manages multiple social communication channels.
 *
 * Supported channels:
 * - WeChat (微信): com.tencent.mm - Social messaging
 * - Phone Call: tel: - Direct voice communication
 * - SMS: sms: - Text messaging
 * - Alipay (支付宝): com.eg.android.AlipayGphone - Financial messaging
 * - Douyin (抖音): com.ss.android.ugc.aweme - Video social platform
 * - Xiaohongshu (小红书): com.xingin.xhs - Content sharing platform
 * - DingTalk (钉钉): com.alibaba.android.rimet - Business communication
 * - WeCom (企业微信): com.tencent.wework - Enterprise messaging
 */
public class ChannelSelector {

    private static final String TAG = "ChannelSelector";

    private final Context context;
    private final AutomationEngine automationEngine;
    private final UserMemoryManager memoryManager;

    // Channel availability cache
    private final Map<String, Boolean> channelAvailability;

    // Channel priorities (user customizable)
    private final Map<String, Integer> channelPriorities;

    /**
     * Channel constants
     */
    public static final String CHANNEL_WECHAT = "wechat";
    public static final String CHANNEL_PHONE = "phone";
    public static final String CHANNEL_SMS = "sms";
    public static final String CHANNEL_ALIPAY = "alipay";
    public static final String CHANNEL_DOUYIN = "douyin";
    public static final String CHANNEL_XIAOHONGSHU = "xiaohongshu";
    public static final String CHANNEL_DINGTALK = "dingtalk";
    public static final String CHANNEL_WECOM = "wecom";
    public static final String CHANNEL_QQ = "qq";
    public static final String CHANNEL_TELEGRAM = "telegram";

    /**
     * App package names
     */
    public static final Map<String, String> CHANNEL_PACKAGES = Map.of(
        CHANNEL_WECHAT, "com.tencent.mm",
        CHANNEL_ALIPAY, "com.eg.android.AlipayGphone",
        CHANNEL_DOUYIN, "com.ss.android.ugc.aweme",
        CHANNEL_XIAOHONGSHU, "com.xingin.xhs",
        CHANNEL_DINGTALK, "com.alibaba.android.rimet",
        CHANNEL_WECOM, "com.tencent.wework",
        CHANNEL_QQ, "com.tencent.mobileqq"
    );

    /**
     * Channel capabilities
     */
    public static class ChannelCapability {
        public final String channel;
        public final boolean supportsText;
        public final boolean supportsVoice;
        public final boolean supportsImage;
        public final boolean supportsVideo;
        public final boolean supportsLocation;
        public final boolean supportsPayment;
        public final boolean supportsGroup;
        public final String packageName;

        public ChannelCapability(String channel, boolean supportsText, boolean supportsVoice,
                                  boolean supportsImage, boolean supportsVideo,
                                  boolean supportsLocation, boolean supportsPayment,
                                  boolean supportsGroup, String packageName) {
            this.channel = channel;
            this.supportsText = supportsText;
            this.supportsVoice = supportsVoice;
            this.supportsImage = supportsImage;
            this.supportsVideo = supportsVideo;
            this.supportsLocation = supportsLocation;
            this.supportsPayment = supportsPayment;
            this.supportsGroup = supportsGroup;
            this.packageName = packageName;
        }
    }

    // Predefined channel capabilities
    private static final Map<String, ChannelCapability> CAPABILITIES = new HashMap<>();

    static {
        CAPABILITIES.put(CHANNEL_WECHAT, new ChannelCapability(
            CHANNEL_WECHAT, true, true, true, true, true, true, true,
            CHANNEL_PACKAGES.get(CHANNEL_WECHAT)));

        CAPABILITIES.put(CHANNEL_PHONE, new ChannelCapability(
            CHANNEL_PHONE, false, true, false, false, false, false, false,
            null));

        CAPABILITIES.put(CHANNEL_SMS, new ChannelCapability(
            CHANNEL_SMS, true, false, false, false, false, false, true,
            null));

        CAPABILITIES.put(CHANNEL_ALIPAY, new ChannelCapability(
            CHANNEL_ALIPAY, true, false, true, false, false, true, false,
            CHANNEL_PACKAGES.get(CHANNEL_ALIPAY)));

        CAPABILITIES.put(CHANNEL_DOUYIN, new ChannelCapability(
            CHANNEL_DOUYIN, true, false, true, true, false, false, true,
            CHANNEL_PACKAGES.get(CHANNEL_DOUYIN)));

        CAPABILITIES.put(CHANNEL_XIAOHONGSHU, new ChannelCapability(
            CHANNEL_XIAOHONGSHU, true, false, true, true, false, false, true,
            CHANNEL_PACKAGES.get(CHANNEL_XIAOHONGSHU)));

        CAPABILITIES.put(CHANNEL_DINGTALK, new ChannelCapability(
            CHANNEL_DINGTALK, true, true, true, true, true, false, true,
            CHANNEL_PACKAGES.get(CHANNEL_DINGTALK)));

        CAPABILITIES.put(CHANNEL_WECOM, new ChannelCapability(
            CHANNEL_WECOM, true, true, true, true, true, false, true,
            CHANNEL_PACKAGES.get(CHANNEL_WECOM)));

        CAPABILITIES.put(CHANNEL_QQ, new ChannelCapability(
            CHANNEL_QQ, true, true, true, true, true, false, true,
            CHANNEL_PACKAGES.get(CHANNEL_QQ)));
    }

    /**
     * Create channel selector
     */
    public ChannelSelector(@NonNull Context context,
                           @Nullable AutomationEngine automationEngine,
                           @Nullable UserMemoryManager memoryManager) {
        this.context = context;
        this.automationEngine = automationEngine;
        this.memoryManager = memoryManager;
        this.channelAvailability = new HashMap<>();
        this.channelPriorities = new HashMap<>();

        // Default priorities
        channelPriorities.put(CHANNEL_WECHAT, 10);
        channelPriorities.put(CHANNEL_PHONE, 9);
        channelPriorities.put(CHANNEL_SMS, 8);
        channelPriorities.put(CHANNEL_ALIPAY, 7);
        channelPriorities.put(CHANNEL_XIAOHONGSHU, 6);
        channelPriorities.put(CHANNEL_DOUYIN, 5);
        channelPriorities.put(CHANNEL_DINGTALK, 4);
        channelPriorities.put(CHANNEL_WECOM, 3);
        channelPriorities.put(CHANNEL_QQ, 2);

        // Check channel availability
        checkAllChannels();
    }

    /**
     * Check availability of all channels
     */
    private void checkAllChannels() {
        PackageManager pm = context.getPackageManager();

        for (Map.Entry<String, String> entry : CHANNEL_PACKAGES.entrySet()) {
            String channel = entry.getKey();
            String packageName = entry.getValue();

            try {
                pm.getPackageInfo(packageName, PackageManager.GET_ACTIVITIES);
                channelAvailability.put(channel, true);
                Log.d(TAG, "Channel " + channel + " is available");
            } catch (PackageManager.NameNotFoundException e) {
                channelAvailability.put(channel, false);
                Log.d(TAG, "Channel " + channel + " is NOT available");
            }
        }

        // Phone and SMS are always available on phones
        channelAvailability.put(CHANNEL_PHONE, true);
        channelAvailability.put(CHANNEL_SMS, true);
    }

    /**
     * Check if a channel is available
     */
    public boolean isChannelAvailable(@NonNull String channel) {
        Boolean available = channelAvailability.get(channel);
        return available != null && available;
    }

    /**
     * Get all available channels
     */
    @NonNull
    public List<String> getAvailableChannels() {
        List<String> available = new ArrayList<>();
        for (Map.Entry<String, Boolean> entry : channelAvailability.entrySet()) {
            if (entry.getValue()) {
                available.add(entry.getKey());
            }
        }
        return available;
    }

    /**
     * Get channel capability
     */
    @Nullable
    public ChannelCapability getCapability(@NonNull String channel) {
        return CAPABILITIES.get(channel);
    }

    /**
     * Select best channel based on requirements
     */
    @NonNull
    public String selectBestChannel(@NonNull String messageType,
                                     int urgencyLevel,
                                     @NonNull Map<String, String> requirements) {
        // Get recommended channels for message type
        List<String> recommendedChannels = getRecommendedChannels(messageType, urgencyLevel);

        // Filter by availability
        List<String> availableRecommended = new ArrayList<>();
        for (String channel : recommendedChannels) {
            if (isChannelAvailable(channel)) {
                availableRecommended.add(channel);
            }
        }

        // Filter by capabilities if requirements specified
        if (!requirements.isEmpty()) {
            availableRecommended = filterByCapabilities(availableRecommended, requirements);
        }

        // Return best available channel
        if (!availableRecommended.isEmpty()) {
            return availableRecommended.get(0);
        }

        // Fallback to WeChat or SMS
        if (isChannelAvailable(CHANNEL_WECHAT)) {
            return CHANNEL_WECHAT;
        }
        return CHANNEL_SMS; // Always available
    }

    /**
     * Get recommended channels for message type
     *
     * Modern social habits mapping:
     * - Invitation (约吃饭) → WeChat (social discussion)
     * - Urgent (紧急) → Phone → WeChat → SMS
     * - Reminder (提醒) → SMS → WeChat (reliable delivery)
     * - Guide (攻略) → Xiaohongshu → WeChat (content sharing)
     * - Payment (支付) → Alipay → WeChat (financial)
     * - Casual (日常) → WeChat → Douyin → QQ
     * - Business (工作) → DingTalk → WeCom → WeChat
     * - Location (位置) → WeChat (has location sharing)
     */
    @NonNull
    private List<String> getRecommendedChannels(@NonNull String messageType, int urgencyLevel) {
        List<String> channels = new ArrayList<>();

        // Urgency override
        if (urgencyLevel >= MessageClassifier.URGENCY_CRITICAL) {
            channels.add(CHANNEL_PHONE);
            channels.add(CHANNEL_SMS);
            channels.add(CHANNEL_WECHAT);
            return channels;
        }

        // Type-based recommendation
        switch (messageType) {
            case MessageClassifier.TYPE_INVITATION:
                // 约吃饭、聚会 → 微信首选 (方便讨论、确认)
                channels.add(CHANNEL_WECHAT);
                channels.add(CHANNEL_PHONE);
                channels.add(CHANNEL_SMS);
                break;

            case MessageClassifier.TYPE_URGENT:
                // 紧急重要 → 电话 → 微信 → 短信
                channels.add(CHANNEL_PHONE);
                channels.add(CHANNEL_WECHAT);
                channels.add(CHANNEL_SMS);
                break;

            case MessageClassifier.TYPE_REMINDER:
                // 提醒 → 短信 (确保送达) → 微信
                channels.add(CHANNEL_SMS);
                channels.add(CHANNEL_WECHAT);
                channels.add(CHANNEL_PHONE);
                break;

            case MessageClassifier.TYPE_GUIDE:
                // 攻略、教程 → 小红书私信 → 微信
                channels.add(CHANNEL_XIAOHONGSHU);
                channels.add(CHANNEL_WECHAT);
                channels.add(CHANNEL_DOUYIN);
                break;

            case MessageClassifier.TYPE_PAYMENT:
                // 支付、转账 → 支付宝 → 微信
                channels.add(CHANNEL_ALIPAY);
                channels.add(CHANNEL_WECHAT);
                break;

            case MessageClassifier.TYPE_CASUAL:
                // 日常聊天 → 微信 → 抖音 → QQ
                channels.add(CHANNEL_WECHAT);
                channels.add(CHANNEL_DOUYIN);
                channels.add(CHANNEL_QQ);
                break;

            case MessageClassifier.TYPE_BUSINESS:
                // 工作通知 → 钉钉 → 企业微信 → 微信
                channels.add(CHANNEL_DINGTALK);
                channels.add(CHANNEL_WECOM);
                channels.add(CHANNEL_WECHAT);
                break;

            case MessageClassifier.TYPE_GREETING:
                // 问候 → 微信
                channels.add(CHANNEL_WECHAT);
                channels.add(CHANNEL_SMS);
                break;

            case MessageClassifier.TYPE_LOCATION:
                // 位置分享 → 微信 (有定位功能)
                channels.add(CHANNEL_WECHAT);
                channels.add(CHANNEL_SMS);
                break;

            default:
                // 默认 → 微信
                channels.add(CHANNEL_WECHAT);
                channels.add(CHANNEL_SMS);
        }

        return channels;
    }

    /**
     * Filter channels by capability requirements
     */
    @NonNull
    private List<String> filterByCapabilities(@NonNull List<String> channels,
                                                @NonNull Map<String, String> requirements) {
        List<String> filtered = new ArrayList<>();

        for (String channel : channels) {
            ChannelCapability cap = getCapability(channel);
            if (cap == null) continue;

            boolean matches = true;

            for (Map.Entry<String, String> req : requirements.entrySet()) {
                String key = req.getKey();
                boolean required = Boolean.parseBoolean(req.getValue());

                switch (key) {
                    case "text":
                        if (required && !cap.supportsText) matches = false;
                        break;
                    case "voice":
                        if (required && !cap.supportsVoice) matches = false;
                        break;
                    case "image":
                        if (required && !cap.supportsImage) matches = false;
                        break;
                    case "video":
                        if (required && !cap.supportsVideo) matches = false;
                        break;
                    case "location":
                        if (required && !cap.supportsLocation) matches = false;
                        break;
                    case "payment":
                        if (required && !cap.supportsPayment) matches = false;
                        break;
                    case "group":
                        if (required && !cap.supportsGroup) matches = false;
                        break;
                }
            }

            if (matches) {
                filtered.add(channel);
            }
        }

        return filtered;
    }

    /**
     * Set channel priority
     */
    public void setChannelPriority(@NonNull String channel, int priority) {
        channelPriorities.put(channel, priority);
    }

    /**
     * Get channel priority
     */
    public int getChannelPriority(@NonNull String channel) {
        Integer priority = channelPriorities.get(channel);
        return priority != null ? priority : 0;
    }

    /**
     * Get channel availability report
     */
    @NonNull
    public String getAvailabilityReport() {
        StringBuilder sb = new StringBuilder();
        sb.append("Channel Availability Report:\n");

        for (Map.Entry<String, Boolean> entry : channelAvailability.entrySet()) {
            sb.append("  ").append(entry.getKey())
              .append(": ").append(entry.getValue() ? "Available" : "Not Available")
              .append("\n");
        }

        return sb.toString();
    }
}