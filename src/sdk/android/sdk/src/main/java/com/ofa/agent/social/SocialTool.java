package com.ofa.agent.social;

import android.content.Context;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.tool.ToolDefinition;
import com.ofa.agent.tool.ToolExecutor;
import com.ofa.agent.tool.ToolRegistry;
import com.ofa.agent.tool.ToolResult;
import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.memory.UserMemoryManager;

import org.json.JSONObject;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Social Tools - MCP tool definitions for social notification system.
 *
 * Provides tools for intelligent social messaging:
 * - social.send: Smart message delivery with auto channel selection
 * - social.invite: Send invitation (约吃饭等)
 * - social.urgent: Send urgent message via phone
 * - social.guide: Share tips/guide via Xiaohongshu
 * - social.payment: Send payment reminder via Alipay
 * - social.classify: Classify message type
 * - social.contact.find: Find contact by name
 * - social.contact.search: Search contacts
 * - social.channel.list: List available channels
 * - social.stats: Get delivery statistics
 */
public class SocialTool {

    private final Context context;
    private final SocialOrchestrator orchestrator;
    private final MessageClassifier classifier;
    private final ChannelSelector channelSelector;

    /**
     * Create social tools
     */
    public SocialTool(@NonNull Context context,
                      @Nullable AutomationEngine automationEngine,
                      @Nullable UserMemoryManager memoryManager) {
        this.context = context;
        this.orchestrator = new SocialOrchestrator(context, automationEngine, memoryManager);
        this.classifier = new MessageClassifier(context, memoryManager);
        this.channelSelector = new ChannelSelector(context, automationEngine, memoryManager);
    }

    /**
     * Register all social tools
     */
    public void registerTools(@NonNull ToolRegistry registry) {
        // Main notification tool
        registry.register(ToolDefinition.create(
            "social.send",
            "智能发送消息 - 自动选择最佳渠道",
            "message", "string", true, "消息内容",
            "recipient", "string", false, "接收者姓名",
            "phone", "string", false, "接收者电话",
            "type", "string", false, "消息类型 (invitation/urgent/guide/payment/casual)",
            "channel", "string", false, "指定渠道 (wechat/phone/sms/alipay/xiaohongshu)"
        ), new SendExecutor());

        // Invitation tool
        registry.register(ToolDefinition.create(
            "social.invite",
            "发送邀请消息 (约吃饭/聚会等) - 自动使用微信",
            "activity", "string", true, "活动内容 (吃饭/看电影等)",
            "time", "string", false, "时间",
            "recipient", "string", true, "接收者姓名",
            "phone", "string", false, "接收者电话"
        ), new InviteExecutor());

        // Urgent message tool
        registry.register(ToolDefinition.create(
            "social.urgent",
            "发送紧急重要消息 - 自动使用电话",
            "message", "string", true, "紧急消息内容",
            "recipient", "string", true, "接收者姓名",
            "phone", "string", true, "接收者电话"
        ), new UrgentExecutor());

        // Guide sharing tool
        registry.register(ToolDefinition.create(
            "social.guide",
            "分享攻略/教程 - 自动使用小红书私信",
            "title", "string", true, "攻略标题",
            "content", "string", true, "攻略内容",
            "recipient", "string", true, "接收者"
        ), new GuideExecutor());

        // Payment reminder tool
        registry.register(ToolDefinition.create(
            "social.payment",
            "发送支付提醒 - 自动使用支付宝",
            "amount", "string", true, "金额",
            "recipient", "string", true, "接收者姓名",
            "phone", "string", false, "接收者电话"
        ), new PaymentExecutor());

        // Classification tool
        registry.register(ToolDefinition.create(
            "social.classify",
            "分析消息类型和推荐渠道",
            "message", "string", true, "要分析的消息"
        ), new ClassifyExecutor());

        // Contact find tool
        registry.register(ToolDefinition.create(
            "social.contact.find",
            "查找联系人",
            "name", "string", true, "联系人姓名"
        ), new ContactFindExecutor());

        // Contact search tool
        registry.register(ToolDefinition.create(
            "social.contact.search",
            "搜索联系人",
            "query", "string", true, "搜索关键词"
        ), new ContactSearchExecutor());

        // Channel list tool
        registry.register(ToolDefinition.createSimple(
            "social.channel.list",
            "获取可用的社交渠道列表"
        ), new ChannelListExecutor());

        // Statistics tool
        registry.register(ToolDefinition.createSimple(
            "social.stats",
            "获取社交通知统计信息"
        ), new StatsExecutor());
    }

    // ===== Tool Executors =====

    /**
     * Smart send executor
     */
    private class SendExecutor implements ToolExecutor {
        @NonNull
        @Override
        public ToolResult execute(@NonNull Map<String, String> params) {
            String message = params.get("message");
            String recipient = params.get("recipient");
            String phone = params.get("phone");
            String type = params.get("type");
            String channel = params.get("channel");

            if (message == null) {
                return ToolResult.error("Missing message parameter");
            }

            SocialOrchestrator.NotificationRequest request = new SocialOrchestrator.NotificationRequest(
                message, recipient, phone, type, channel, new HashMap<>());

            SocialOrchestrator.DeliveryRecord record = orchestrator.sendNotification(request);

            if (record.success) {
                return ToolResult.success(Map.of(
                    "channel", record.successfulChannel,
                    "type", record.messageType,
                    "duration", String.valueOf(record.durationMs),
                    "recipient", record.recipientName != null ? record.recipientName : ""
                ));
            } else {
                return ToolResult.error(record.failureReason);
            }
        }

        @Override
        public long getEstimatedTimeMs() {
            return 3000;
        }
    }

    /**
     * Invitation executor
     */
    private class InviteExecutor implements ToolExecutor {
        @NonNull
        @Override
        public ToolResult execute(@NonNull Map<String, String> params) {
            String activity = params.get("activity");
            String time = params.get("time");
            String recipient = params.get("recipient");
            String phone = params.get("phone");

            if (activity == null || recipient == null) {
                return ToolResult.error("Missing activity or recipient");
            }

            SocialOrchestrator.DeliveryRecord record = orchestrator.sendInvitation(
                activity, time, recipient, phone);

            return record.success ?
                ToolResult.success(Map.of("channel", record.successfulChannel)) :
                ToolResult.error(record.failureReason);
        }

        @Override
        public long getEstimatedTimeMs() {
            return 3000;
        }
    }

    /**
     * Urgent executor
     */
    private class UrgentExecutor implements ToolExecutor {
        @NonNull
        @Override
        public ToolResult execute(@NonNull Map<String, String> params) {
            String message = params.get("message");
            String recipient = params.get("recipient");
            String phone = params.get("phone");

            if (message == null || recipient == null || phone == null) {
                return ToolResult.error("Missing message, recipient, or phone for urgent notification");
            }

            SocialOrchestrator.DeliveryRecord record = orchestrator.sendUrgent(
                message, recipient, phone);

            return record.success ?
                ToolResult.success(Map.of("channel", record.successfulChannel)) :
                ToolResult.error(record.failureReason);
        }

        @Override
        public long getEstimatedTimeMs() {
            return 1000;
        }
    }

    /**
     * Guide executor
     */
    private class GuideExecutor implements ToolExecutor {
        @NonNull
        @Override
        public ToolResult execute(@NonNull Map<String, String> params) {
            String title = params.get("title");
            String content = params.get("content");
            String recipient = params.get("recipient");

            if (title == null || content == null || recipient == null) {
                return ToolResult.error("Missing title, content, or recipient");
            }

            SocialOrchestrator.DeliveryRecord record = orchestrator.sendGuide(
                title, content, recipient);

            return record.success ?
                ToolResult.success(Map.of("channel", record.successfulChannel)) :
                ToolResult.error(record.failureReason);
        }

        @Override
        public long getEstimatedTimeMs() {
            return 5000;
        }
    }

    /**
     * Payment executor
     */
    private class PaymentExecutor implements ToolExecutor {
        @NonNull
        @Override
        public ToolResult execute(@NonNull Map<String, String> params) {
            String amount = params.get("amount");
            String recipient = params.get("recipient");
            String phone = params.get("phone");

            if (amount == null || recipient == null) {
                return ToolResult.error("Missing amount or recipient");
            }

            SocialOrchestrator.DeliveryRecord record = orchestrator.sendPaymentReminder(
                amount, recipient, phone);

            return record.success ?
                ToolResult.success(Map.of("channel", record.successfulChannel)) :
                ToolResult.error(record.failureReason);
        }

        @Override
        public long getEstimatedTimeMs() {
            return 3000;
        }
    }

    /**
     * Classify executor
     */
    private class ClassifyExecutor implements ToolExecutor {
        @NonNull
        @Override
        public ToolResult execute(@NonNull Map<String, String> params) {
            String message = params.get("message");

            if (message == null) {
                return ToolResult.error("Missing message parameter");
            }

            MessageClassifier.ClassificationResult result = classifier.classify(message);

            return ToolResult.success(Map.of(
                "type", result.messageType,
                "channel", result.recommendedChannel,
                "urgency", String.valueOf(result.urgencyLevel),
                "confidence", String.valueOf(result.confidence),
                "reason", result.reason
            ));
        }

        @Override
        public long getEstimatedTimeMs() {
            return 100;
        }
    }

    /**
     * Contact find executor
     */
    private class ContactFindExecutor implements ToolExecutor {
        @NonNull
        @Override
        public ToolResult execute(@NonNull Map<String, String> params) {
            String name = params.get("name");

            if (name == null) {
                return ToolResult.error("Missing name parameter");
            }

            SocialOrchestrator.ContactInfo contact = orchestrator.findContact(name);

            if (contact == null) {
                return ToolResult.error("Contact not found: " + name);
            }

            return ToolResult.success(Map.of(
                "id", contact.id,
                "name", contact.displayName,
                "phone", contact.getPrimaryPhone() != null ? contact.getPrimaryPhone() : "",
                "email", contact.getPrimaryEmail() != null ? contact.getPrimaryEmail() : "",
                "wechat", contact.getWeChatId() != null ? contact.getWeChatId() : ""
            ));
        }

        @Override
        public long getEstimatedTimeMs() {
            return 500;
        }
    }

    /**
     * Contact search executor
     */
    private class ContactSearchExecutor implements ToolExecutor {
        @NonNull
        @Override
        public ToolResult execute(@NonNull Map<String, String> params) {
            String query = params.get("query");

            if (query == null) {
                return ToolResult.error("Missing query parameter");
            }

            List<SocialOrchestrator.ContactInfo> contacts = orchestrator.searchContacts(query);

            Map<String, String> result = new HashMap<>();
            result.put("count", String.valueOf(contacts.size()));

            for (int i = 0; i < Math.min(10, contacts.size()); i++) {
                SocialOrchestrator.ContactInfo c = contacts.get(i);
                result.put("contact_" + i, c.displayName + "|" + c.getPrimaryPhone());
            }

            return ToolResult.success(result);
        }

        @Override
        public long getEstimatedTimeMs() {
            return 500;
        }
    }

    /**
     * Channel list executor
     */
    private class ChannelListExecutor implements ToolExecutor {
        @NonNull
        @Override
        public ToolResult execute(@NonNull Map<String, String> params) {
            List<String> channels = channelSelector.getAvailableChannels();

            Map<String, String> result = new HashMap<>();
            result.put("count", String.valueOf(channels.size()));

            for (int i = 0; i < channels.size(); i++) {
                String channel = channels.get(i);
                ChannelSelector.ChannelCapability cap = channelSelector.getCapability(channel);
                if (cap != null) {
                    result.put("channel_" + i, channel + "|" + cap.packageName +
                        "|text:" + cap.supportsText + "|voice:" + cap.supportsVoice);
                } else {
                    result.put("channel_" + i, channel + "|built-in");
                }
            }

            return ToolResult.success(result);
        }

        @Override
        public long getEstimatedTimeMs() {
            return 100;
        }
    }

    /**
     * Stats executor
     */
    private class StatsExecutor implements ToolExecutor {
        @NonNull
        @Override
        public ToolResult execute(@NonNull Map<String, String> params) {
            Map<String, Map<String, Double>> stats = orchestrator.getChannelStatistics();

            Map<String, String> result = new HashMap<>();

            for (Map.Entry<String, Map<String, Double>> entry : stats.entrySet()) {
                String channel = entry.getKey();
                Map<String, Double> data = entry.getValue();
                result.put(channel, String.format("success=%.1f%%|avg=%.0fms|attempts=%d",
                    data.get("successRate") * 100,
                    data.get("avgDuration"),
                    data.get("totalAttempts").intValue()));
            }

            return ToolResult.success(result);
        }

        @Override
        public long getEstimatedTimeMs() {
            return 100;
        }
    }

    /**
     * Get orchestrator
     */
    @NonNull
    public SocialOrchestrator getOrchestrator() {
        return orchestrator;
    }

    /**
     * Shutdown
     */
    public void shutdown() {
        orchestrator.shutdown();
    }
}