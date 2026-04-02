package com.ofa.agent.sample;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;

import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.memory.UserMemoryManager;
import com.ofa.agent.social.ChannelSelector;
import com.ofa.agent.social.MessageClassifier;
import com.ofa.agent.social.MessageSender;
import com.ofa.agent.social.SocialOrchestrator;
import com.ofa.agent.social.adapter.ContactInfo;

import java.util.List;
import java.util.Map;

/**
 * Social Notification Sample - demonstrates intelligent social messaging.
 *
 * Scenarios:
 * 1. 约吃饭 → 微信发送邀请
 * 2. 紧急重要 → 电话通知
 * 3. 攻略分享 → 小红书私信
 * 4. 支付提醒 → 支付宝
 * 5. 工作通知 → 钉钉/企业微信
 * 6. 日常聊天 → 微信/抖音
 */
public class SocialNotificationSample {

    private static final String TAG = "SocialSample";

    private final Context context;
    private final SocialOrchestrator orchestrator;

    public SocialNotificationSample(@NonNull Context context,
                                     @NonNull AutomationEngine automationEngine,
                                     @NonNull UserMemoryManager memoryManager) {
        this.context = context;
        this.orchestrator = new SocialOrchestrator(context, automationEngine, memoryManager);

        // Set up listener
        orchestrator.setListener(new SocialOrchestrator.NotificationListener() {
            @Override
            public void onNotificationStart(SocialOrchestrator.NotificationRequest request) {
                Log.i(TAG, "Notification start: " + request);
            }

            @Override
            public void onChannelSelected(String channel, String reason) {
                Log.i(TAG, "Channel selected: " + channel + " (" + reason + ")");
            }

            @Override
            public void onDeliveryStart(String channel) {
                Log.d(TAG, "Delivery via " + channel);
            }

            @Override
            public void onDeliverySuccess(SocialOrchestrator.DeliveryRecord record) {
                Log.i(TAG, "✓ Delivery success: " + record);
            }

            @Override
            public void onDeliveryFailure(SocialOrchestrator.DeliveryRecord record) {
                Log.w(TAG, "✗ Delivery failed: " + record);
            }

            @Override
            public void onFallback(String fromChannel, String toChannel, String reason) {
                Log.w(TAG, "Fallback: " + fromChannel + " → " + toChannel);
            }
        });
    }

    /**
     * Run all sample scenarios
     */
    public void runAllScenarios() {
        Log.i(TAG, "=== Running Social Notification Scenarios ===\n");

        scenario1_Invitation();       // 约吃饭
        scenario2_UrgentMessage();    // 紧急重要
        scenario3_ShareGuide();       // 攻略分享
        scenario4_PaymentReminder();  // 支付提醒
        scenario5_WorkNotification(); // 工作通知
        scenario6_CasualChat();       // 日常聊天

        printStatistics();
    }

    /**
     * Scenario 1: 约吃饭 (Invitation)
     *
     * Modern habit: Use WeChat for social invitations
     * - Easy to discuss details
     * - Can share location/restaurant
     * - Non-urgent, casual communication
     */
    public void scenario1_Invitation() {
        Log.i(TAG, "\n--- Scenario 1: 约吃饭 ---");

        // Example 1: Simple invitation
        SocialOrchestrator.DeliveryRecord record1 = orchestrator.sendNotification(
            "周末有空吗？约你一起吃饭",
            "张三",
            "13812345678"
        );

        Log.i(TAG, "Result: channel=" + record1.primaryChannel +
            ", success=" + record1.success);

        // Example 2: Using convenience method
        SocialOrchestrator.DeliveryRecord record2 = orchestrator.sendInvitation(
            "吃火锅",
            "这周六下午",
            "李四",
            "13987654321"
        );

        Log.i(TAG, "Result: channel=" + record2.primaryChannel +
            ", success=" + record2.success);

        // Example 3: Group invitation
        SocialOrchestrator.DeliveryRecord record3 = orchestrator.sendNotification(
            "周五晚上部门聚餐，地点在公司楼下的川菜馆，大家有空吗？",
            "王五", // Group chat contact
            null
        );

        Log.i(TAG, "Result: channel=" + record3.primaryChannel +
            ", success=" + record3.success);
    }

    /**
     * Scenario 2: 紧急重要消息 (Urgent/Critical)
     *
     * Modern habit: Phone call for urgent matters
     * - Immediate response required
     * - Voice communication is direct
     * - Can convey urgency through tone
     */
    public void scenario2_UrgentMessage() {
        Log.i(TAG, "\n--- Scenario 2: 紧急重要 ---");

        // Example 1: Critical notification
        SocialOrchestrator.DeliveryRecord record1 = orchestrator.sendUrgent(
            "服务器宕机了！请立即处理！",
            "技术主管",
            "13711112222"
        );

        Log.i(TAG, "Result: channel=" + record1.primaryChannel +
            ", urgency=" + record1.urgencyLevel +
            ", success=" + record1.success);

        // Example 2: Auto-classified urgent
        SocialOrchestrator.DeliveryRecord record2 = orchestrator.sendNotification(
            "紧急！明天早上的会议取消了，请马上通知相关人员",
            "同事A",
            "13666667777"
        );

        Log.i(TAG, "Result: channel=" + record2.primaryChannel +
            ", urgency=" + record2.urgencyLevel +
            ", success=" + record2.success);
    }

    /**
     * Scenario 3: 攻略分享 (Guide/Tips Sharing)
     *
     * Modern habit: Xiaohongshu (小红书) for content sharing
     * - Content sharing platform
     * - Good for tips and recommendations
     * - Can include images and detailed content
     */
    public void scenario3_ShareGuide() {
        Log.i(TAG, "\n--- Scenario 3: 攻略分享 ---");

        // Example 1: Travel guide
        SocialOrchestrator.DeliveryRecord record1 = orchestrator.sendGuide(
            "三亚旅游攻略",
            "推荐几个不错的酒店和景点...",
            "好友A"
        );

        Log.i(TAG, "Result: channel=" + record1.primaryChannel +
            ", type=" + record1.messageType +
            ", success=" + record1.success);

        // Example 2: Food recommendation
        SocialOrchestrator.DeliveryRecord record2 = orchestrator.sendNotification(
            "发现一家特别好吃的奶茶店！推荐他们的招牌珍珠奶茶，攻略如下...",
            "吃货朋友",
            null
        );

        Log.i(TAG, "Result: channel=" + record2.primaryChannel +
            ", type=" + record2.messageType +
            ", success=" + record2.success);

        // Example 3: Tech tips
        SocialOrchestrator.DeliveryRecord record3 = orchestrator.sendGuide(
            "提高效率的10个APP",
            "这些APP能帮你更好地管理时间...",
            "技术群"
        );

        Log.i(TAG, "Result: channel=" + record3.primaryChannel +
            ", success=" + record3.success);
    }

    /**
     * Scenario 4: 支付提醒 (Payment Reminder)
     *
     * Modern habit: Alipay for financial messages
     * - Financial platform context
     * - Can include payment links
     * - Secure for money-related topics
     */
    public void scenario4_PaymentReminder() {
        Log.i(TAG, "\n--- Scenario 4: 支付提醒 ---");

        // Example 1: Payment request
        SocialOrchestrator.DeliveryRecord record1 = orchestrator.sendPaymentReminder(
            "50",
            "借款人",
            "13555556666"
        );

        Log.i(TAG, "Result: channel=" + record1.primaryChannel +
            ", type=" + record1.messageType +
            ", success=" + record1.success);

        // Example 2: Bill reminder
        SocialOrchestrator.DeliveryRecord record2 = orchestrator.sendNotification(
            "水电费账单到期了，记得缴费，一共200元",
            "家人",
            null
        );

        Log.i(TAG, "Result: channel=" + record2.primaryChannel +
            ", type=" + record2.messageType +
            ", success=" + record2.success);

        // Example 3: Loan repayment
        SocialOrchestrator.DeliveryRecord record3 = orchestrator.sendNotification(
            "上个月借的500元还一下呗",
            "朋友",
            null
        );

        Log.i(TAG, "Result: channel=" + record3.primaryChannel +
            ", success=" + record3.success);
    }

    /**
     * Scenario 5: 工作通知 (Business/Work)
     *
     * Modern habit: DingTalk/WeCom for work messages
     * - Professional context
     * - Enterprise communication
     * - Better for work-related topics
     */
    public void scenario5_WorkNotification() {
        Log.i(TAG, "\n--- Scenario 5: 工作通知 ---");

        // Example 1: Task notification
        SocialOrchestrator.DeliveryRecord record1 = orchestrator.sendNotification(
            "本周的项目进度报告需要在周五前提交",
            "项目组成员",
            null
        );

        Log.i(TAG, "Result: channel=" + record1.primaryChannel +
            ", type=" + record1.messageType +
            ", success=" + record1.success);

        // Example 2: Meeting reminder
        SocialOrchestrator.DeliveryRecord record2 = orchestrator.sendNotification(
            "明天下午2点有个重要的客户会议，请准时参加",
            "团队成员",
            null
        );

        Log.i(TAG, "Result: channel=" + record2.primaryChannel +
            ", type=" + record2.messageType +
            ", success=" + record2.success);

        // Example 3: Approval request
        SocialOrchestrator.DeliveryRecord record3 = orchestrator.sendNotification(
            "请假申请需要你审批一下",
            "领导",
            null
        );

        Log.i(TAG, "Result: channel=" + record3.primaryChannel +
            ", success=" + record3.success);
    }

    /**
     * Scenario 6: 日常聊天 (Casual)
     *
     * Modern habit: WeChat/Douyin for casual chat
     * - Social platforms
     * - Relaxed communication
     * - Can include fun content
     */
    public void scenario6_CasualChat() {
        Log.i(TAG, "\n--- Scenario 6: 日常聊天 ---");

        // Example 1: Greeting
        SocialOrchestrator.DeliveryRecord record1 = orchestrator.sendNotification(
            "好久不见！最近怎么样？",
            "老同学",
            null
        );

        Log.i(TAG, "Result: channel=" + record1.primaryChannel +
            ", type=" + record1.messageType +
            ", success=" + record1.success);

        // Example 2: Fun message
        SocialOrchestrator.DeliveryRecord record2 = orchestrator.sendNotification(
            "看到这个搞笑视频了，哈哈哈",
            "好友B",
            null
        );

        Log.i(TAG, "Result: channel=" + record2.primaryChannel +
            ", type=" + record2.messageType +
            ", success=" + record2.success);

        // Example 3: Location sharing
        SocialOrchestrator.DeliveryRecord record3 = orchestrator.sendNotification(
            "我在咖啡厅等你，位置发你了",
            "约会对象",
            "13888889999"
        );

        Log.i(TAG, "Result: channel=" + record3.primaryChannel +
            ", type=" + record3.messageType +
            ", success=" + record3.success);
    }

    /**
     * Demonstrate contact integration
     */
    public void demonstrateContactSearch() {
        Log.i(TAG, "\n--- Contact Search Demo ---");

        // Find contact by name
        ContactInfo contact = orchestrator.findContact("张三");
        if (contact != null) {
            Log.i(TAG, "Found contact: " + contact);
            Log.i(TAG, "  Phone: " + contact.getPrimaryPhone());
            Log.i(TAG, "  WeChat: " + contact.getWeChatId());
        }

        // Search contacts
        List<ContactInfo> contacts = orchestrator.searchContacts("李");
        Log.i(TAG, "Found " + contacts.size() + " contacts matching '李'");
        for (ContactInfo c : contacts) {
            Log.i(TAG, "  - " + c.displayName + " (" + c.getPrimaryPhone() + ")");
        }
    }

    /**
     * Demonstrate message classification
     */
    public void demonstrateClassification() {
        Log.i(TAG, "\n--- Message Classification Demo ---");

        MessageClassifier classifier = new MessageClassifier(context, null);

        String[] messages = {
            "约你明天吃饭",
            "紧急！服务器挂了！",
            "分享一个旅游攻略",
            "还我50块钱",
            "明天开会，记得准备材料",
            "好久不见"
        };

        for (String msg : messages) {
            MessageClassifier.ClassificationResult result = classifier.classify(msg);
            Log.i(TAG, String.format("'%s' → type=%s, channel=%s, urgency=%d",
                msg.substring(0, Math.min(20, msg.length())),
                result.messageType,
                result.recommendedChannel,
                result.urgencyLevel));
        }
    }

    /**
     * Print statistics
     */
    public void printStatistics() {
        Log.i(TAG, "\n=== Statistics ===");

        String report = orchestrator.getStatusReport();
        Log.i(TAG, report);
    }

    /**
     * Batch notifications demo
     */
    public void demonstrateBatchNotifications() {
        Log.i(TAG, "\n--- Batch Notifications Demo ---");

        List<SocialOrchestrator.NotificationRequest> requests = List.of(
            new SocialOrchestrator.NotificationRequest("约你吃饭", "张三", null),
            new SocialOrchestrator.NotificationRequest("紧急通知", "李四", "13611112222"),
            new SocialOrchestrator.NotificationRequest("分享攻略", "王五", null)
        );

        List<SocialOrchestrator.DeliveryRecord> results = orchestrator.sendBatchNotifications(requests);

        Log.i(TAG, "Batch sent: " + results.size() + " notifications");
        for (SocialOrchestrator.DeliveryRecord r : results) {
            Log.i(TAG, "  " + r);
        }
    }

    /**
     * Async notification demo
     */
    public void demonstrateAsyncNotification() {
        Log.i(TAG, "\n--- Async Notification Demo ---");

        orchestrator.sendNotificationAsync("异步发送测试", "测试用户", null)
            .thenAccept(record -> {
                Log.i(TAG, "Async result: " + record);
            });
    }
}