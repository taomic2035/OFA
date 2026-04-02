package com.ofa.agent.social;

import android.content.Context;
import android.content.Intent;
import android.net.Uri;
import android.telephony.SmsManager;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.automation.BySelector;
import com.ofa.agent.automation.Direction;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;

/**
 * Message Sender - sends messages through various social channels.
 *
 * Supports:
 * - Phone calls (tel:)
 * - SMS (sms:)
 * - WeChat messages (via AccessibilityService)
 * - Alipay messages (via AccessibilityService)
 * - Douyin messages (via AccessibilityService)
 * - Xiaohongshu private messages (via AccessibilityService)
 * - DingTalk messages (via AccessibilityService)
 * - WeCom messages (via AccessibilityService)
 */
public class MessageSender {

    private static final String TAG = "MessageSender";

    private final Context context;
    private final AutomationEngine automationEngine;
    private final ChannelSelector channelSelector;

    private SendListener listener;
    private int retryCount = 3;
    private long retryDelayMs = 1000;

    /**
     * Send result
     */
    public static class SendResult {
        public final boolean success;
        public final String channel;
        public final String recipient;
        public final String message;
        public final String error;
        public final long durationMs;

        public SendResult(boolean success, String channel, String recipient,
                          String message, String error, long durationMs) {
            this.success = success;
            this.channel = channel;
            this.recipient = recipient;
            this.message = message;
            this.error = error;
            this.durationMs = durationMs;
        }

        @NonNull
        @Override
        public String toString() {
            return String.format("SendResult{channel=%s, success=%s, recipient=%s}",
                channel, success, recipient);
        }
    }

    /**
     * Send listener interface
     */
    public interface SendListener {
        void onSendStart(String channel, String recipient, String message);
        void onSendSuccess(SendResult result);
        void onSendFailure(SendResult result);
        void onRetry(int attempt, String channel, String reason);
    }

    /**
     * Create message sender
     */
    public MessageSender(@NonNull Context context,
                         @Nullable AutomationEngine automationEngine,
                         @NonNull ChannelSelector channelSelector) {
        this.context = context;
        this.automationEngine = automationEngine;
        this.channelSelector = channelSelector;
    }

    /**
     * Set send listener
     */
    public void setListener(@Nullable SendListener listener) {
        this.listener = listener;
    }

    /**
     * Set retry configuration
     */
    public void setRetryConfig(int count, long delayMs) {
        this.retryCount = count;
        this.retryDelayMs = delayMs;
    }

    /**
     * Send message through specified channel
     */
    @NonNull
    public SendResult send(@NonNull String channel,
                            @Nullable String recipient,
                            @Nullable String phoneNumber,
                            @NonNull String message) {
        long startTime = System.currentTimeMillis();

        // Notify listener
        if (listener != null) {
            listener.onSendStart(channel, recipient, message);
        }

        Log.i(TAG, "Sending via " + channel + " to " + recipient);

        SendResult result;
        int attempts = 0;

        do {
            attempts++;
            result = sendOnce(channel, recipient, phoneNumber, message, startTime);

            if (result.success) {
                if (listener != null) {
                    listener.onSendSuccess(result);
                }
                return result;
            }

            // Retry logic
            if (attempts < retryCount && listener != null) {
                listener.onRetry(attempts, channel, result.error);
                try {
                    Thread.sleep(retryDelayMs);
                } catch (InterruptedException e) {
                    break;
                }
            }
        } while (attempts < retryCount);

        // Final failure
        if (listener != null) {
            listener.onSendFailure(result);
        }

        return result;
    }

    /**
     * Single send attempt
     */
    @NonNull
    private SendResult sendOnce(@NonNull String channel,
                                 @Nullable String recipient,
                                 @Nullable String phoneNumber,
                                 @NonNull String message,
                                 long startTime) {
        try {
            switch (channel) {
                case ChannelSelector.CHANNEL_PHONE:
                    return sendViaPhone(phoneNumber, message, startTime);

                case ChannelSelector.CHANNEL_SMS:
                    return sendViaSMS(phoneNumber, message, startTime);

                case ChannelSelector.CHANNEL_WECHAT:
                    return sendViaWeChat(recipient, message, startTime);

                case ChannelSelector.CHANNEL_ALIPAY:
                    return sendViaAlipay(recipient, message, startTime);

                case ChannelSelector.CHANNEL_DOUYIN:
                    return sendViaDouyin(recipient, message, startTime);

                case ChannelSelector.CHANNEL_XIAOHONGSHU:
                    return sendViaXiaohongshu(recipient, message, startTime);

                case ChannelSelector.CHANNEL_DINGTALK:
                    return sendViaDingTalk(recipient, message, startTime);

                case ChannelSelector.CHANNEL_WECOM:
                    return sendViaWeCom(recipient, message, startTime);

                case ChannelSelector.CHANNEL_QQ:
                    return sendViaQQ(recipient, message, startTime);

                default:
                    return new SendResult(false, channel, recipient, message,
                        "Unknown channel", System.currentTimeMillis() - startTime);
            }
        } catch (Exception e) {
            return new SendResult(false, channel, recipient, message,
                e.getMessage(), System.currentTimeMillis() - startTime);
        }
    }

    /**
     * Send via phone call
     * Note: This initiates a call, the message is spoken after connection
     */
    @NonNull
    private SendResult sendViaPhone(@Nullable String phoneNumber,
                                     @NonNull String message,
                                     long startTime) {
        if (phoneNumber == null || phoneNumber.isEmpty()) {
            return new SendResult(false, ChannelSelector.CHANNEL_PHONE, null, message,
                "No phone number provided", System.currentTimeMillis() - startTime);
        }

        try {
            Intent intent = new Intent(Intent.ACTION_CALL);
            intent.setData(Uri.parse("tel:" + phoneNumber));
            intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK);
            context.startActivity(intent);

            // Note: After call connects, voice message would be spoken
            // This requires additional integration with telephony service

            Log.i(TAG, "Phone call initiated to " + phoneNumber);
            return new SendResult(true, ChannelSelector.CHANNEL_PHONE, phoneNumber, message,
                null, System.currentTimeMillis() - startTime);

        } catch (SecurityException e) {
            return new SendResult(false, ChannelSelector.CHANNEL_PHONE, phoneNumber, message,
                "CALL_PHONE permission not granted", System.currentTimeMillis() - startTime);
        } catch (Exception e) {
            return new SendResult(false, ChannelSelector.CHANNEL_PHONE, phoneNumber, message,
                e.getMessage(), System.currentTimeMillis() - startTime);
        }
    }

    /**
     * Send via SMS
     */
    @NonNull
    private SendResult sendViaSMS(@Nullable String phoneNumber,
                                   @NonNull String message,
                                   long startTime) {
        if (phoneNumber == null || phoneNumber.isEmpty()) {
            return new SendResult(false, ChannelSelector.CHANNEL_SMS, null, message,
                "No phone number provided", System.currentTimeMillis() - startTime);
        }

        try {
            SmsManager smsManager = SmsManager.getDefault();

            // Split message if too long
            if (message.length() > 160) {
                ArrayList<String> parts = smsManager.divideMessage(message);
                smsManager.sendMultipartTextMessage(phoneNumber, null, parts, null, null);
            } else {
                smsManager.sendTextMessage(phoneNumber, null, message, null, null);
            }

            Log.i(TAG, "SMS sent to " + phoneNumber);
            return new SendResult(true, ChannelSelector.CHANNEL_SMS, phoneNumber, message,
                null, System.currentTimeMillis() - startTime);

        } catch (Exception e) {
            return new SendResult(false, ChannelSelector.CHANNEL_SMS, phoneNumber, message,
                e.getMessage(), System.currentTimeMillis() - startTime);
        }
    }

    /**
     * Send via WeChat (微信)
     * Uses AccessibilityService to automate WeChat messaging
     */
    @NonNull
    private SendResult sendViaWeChat(@Nullable String recipient,
                                      @NonNull String message,
                                      long startTime) {
        if (automationEngine == null) {
            return new SendResult(false, ChannelSelector.CHANNEL_WECHAT, recipient, message,
                "AutomationEngine not available", System.currentTimeMillis() - startTime);
        }

        if (!channelSelector.isChannelAvailable(ChannelSelector.CHANNEL_WECHAT)) {
            return new SendResult(false, ChannelSelector.CHANNEL_WECHAT, recipient, message,
                "WeChat not installed", System.currentTimeMillis() - startTime);
        }

        try {
            // Step 1: Launch WeChat
            AutomationResult launchResult = launchApp(ChannelSelector.CHANNEL_PACKAGES.get(ChannelSelector.CHANNEL_WECHAT));
            if (!launchResult.isSuccess()) {
                return new SendResult(false, ChannelSelector.CHANNEL_WECHAT, recipient, message,
                    "Failed to launch WeChat", System.currentTimeMillis() - startTime);
            }

            // Wait for app to load
            Thread.sleep(1500);

            // Step 2: Navigate to chat if recipient specified
            if (recipient != null) {
                // Search for contact
                AutomationResult searchResult = automationEngine.click(BySelector.text("搜索"));
                if (searchResult.isSuccess()) {
                    Thread.sleep(500);
                    automationEngine.inputText(recipient);
                    Thread.sleep(500);
                    automationEngine.click(BySelector.textContains(recipient));
                }
            }

            // Step 3: Input message
            Thread.sleep(500);
            AutomationResult inputResult = automationEngine.inputText(message);
            if (!inputResult.isSuccess()) {
                // Fallback: try to find input field
                automationEngine.click(BySelector.className("EditText"));
                Thread.sleep(300);
                automationEngine.inputText(message);
            }

            // Step 4: Send message
            Thread.sleep(300);
            AutomationResult sendResult = automationEngine.click(BySelector.text("发送"));
            if (!sendResult.isSuccess()) {
                // Alternative: click send button by description
                sendResult = automationEngine.click(BySelector.desc("发送"));
            }

            long duration = System.currentTimeMillis() - startTime;
            boolean success = sendResult.isSuccess();

            if (success) {
                Log.i(TAG, "WeChat message sent to " + recipient);
            } else {
                Log.w(TAG, "WeChat send may have failed");
            }

            return new SendResult(success, ChannelSelector.CHANNEL_WECHAT, recipient, message,
                success ? null : "Send button click failed", duration);

        } catch (InterruptedException e) {
            return new SendResult(false, ChannelSelector.CHANNEL_WECHAT, recipient, message,
                "Interrupted", System.currentTimeMillis() - startTime);
        } catch (Exception e) {
            return new SendResult(false, ChannelSelector.CHANNEL_WECHAT, recipient, message,
                e.getMessage(), System.currentTimeMillis() - startTime);
        }
    }

    /**
     * Send via Alipay (支付宝)
     */
    @NonNull
    private SendResult sendViaAlipay(@Nullable String recipient,
                                      @NonNull String message,
                                      long startTime) {
        if (automationEngine == null || !channelSelector.isChannelAvailable(ChannelSelector.CHANNEL_ALIPAY)) {
            return new SendResult(false, ChannelSelector.CHANNEL_ALIPAY, recipient, message,
                "Alipay not available", System.currentTimeMillis() - startTime);
        }

        try {
            // Launch Alipay
            launchApp(ChannelSelector.CHANNEL_PACKAGES.get(ChannelSelector.CHANNEL_ALIPAY));
            Thread.sleep(1500);

            // Navigate to messaging (朋友 tab)
            automationEngine.click(BySelector.text("朋友"));
            Thread.sleep(500);

            // Find and send to recipient
            if (recipient != null) {
                automationEngine.click(BySelector.textContains(recipient));
                Thread.sleep(500);
            }

            // Input and send message
            automationEngine.inputText(message);
            Thread.sleep(300);
            automationEngine.click(BySelector.text("发送"));

            long duration = System.currentTimeMillis() - startTime;
            return new SendResult(true, ChannelSelector.CHANNEL_ALIPAY, recipient, message,
                null, duration);

        } catch (Exception e) {
            return new SendResult(false, ChannelSelector.CHANNEL_ALIPAY, recipient, message,
                e.getMessage(), System.currentTimeMillis() - startTime);
        }
    }

    /**
     * Send via Douyin (抖音)
     */
    @NonNull
    private SendResult sendViaDouyin(@Nullable String recipient,
                                      @NonNull String message,
                                      long startTime) {
        if (automationEngine == null || !channelSelector.isChannelAvailable(ChannelSelector.CHANNEL_DOUYIN)) {
            return new SendResult(false, ChannelSelector.CHANNEL_DOUYIN, recipient, message,
                "Douyin not available", System.currentTimeMillis() - startTime);
        }

        try {
            // Launch Douyin
            launchApp(ChannelSelector.CHANNEL_PACKAGES.get(ChannelSelector.CHANNEL_DOUYIN));
            Thread.sleep(2000);

            // Navigate to messages (私信)
            automationEngine.click(BySelector.desc("消息"));
            Thread.sleep(500);

            if (recipient != null) {
                // Find recipient
                automationEngine.click(BySelector.textContains(recipient));
                Thread.sleep(500);
            }

            // Send message
            automationEngine.inputText(message);
            Thread.sleep(300);
            automationEngine.click(BySelector.text("发送"));

            long duration = System.currentTimeMillis() - startTime;
            return new SendResult(true, ChannelSelector.CHANNEL_DOUYIN, recipient, message,
                null, duration);

        } catch (Exception e) {
            return new SendResult(false, ChannelSelector.CHANNEL_DOUYIN, recipient, message,
                e.getMessage(), System.currentTimeMillis() - startTime);
        }
    }

    /**
     * Send via Xiaohongshu (小红书) private message
     */
    @NonNull
    private SendResult sendViaXiaohongshu(@Nullable String recipient,
                                           @NonNull String message,
                                           long startTime) {
        if (automationEngine == null || !channelSelector.isChannelAvailable(ChannelSelector.CHANNEL_XIAOHONGSHU)) {
            return new SendResult(false, ChannelSelector.CHANNEL_XIAOHONGSHU, recipient, message,
                "Xiaohongshu not available", System.currentTimeMillis() - startTime);
        }

        try {
            // Launch Xiaohongshu
            launchApp(ChannelSelector.CHANNEL_PACKAGES.get(ChannelSelector.CHANNEL_XIAOHONGSHU));
            Thread.sleep(1500);

            // Navigate to messages
            automationEngine.click(BySelector.text("消息"));
            Thread.sleep(500);

            if (recipient != null) {
                automationEngine.click(BySelector.textContains(recipient));
                Thread.sleep(500);
            }

            // Send message
            automationEngine.inputText(message);
            Thread.sleep(300);
            automationEngine.click(BySelector.text("发送"));

            long duration = System.currentTimeMillis() - startTime;
            return new SendResult(true, ChannelSelector.CHANNEL_XIAOHONGSHU, recipient, message,
                null, duration);

        } catch (Exception e) {
            return new SendResult(false, ChannelSelector.CHANNEL_XIAOHONGSHU, recipient, message,
                e.getMessage(), System.currentTimeMillis() - startTime);
        }
    }

    /**
     * Send via DingTalk (钉钉)
     */
    @NonNull
    private SendResult sendViaDingTalk(@Nullable String recipient,
                                        @NonNull String message,
                                        long startTime) {
        if (automationEngine == null || !channelSelector.isChannelAvailable(ChannelSelector.CHANNEL_DINGTALK)) {
            return new SendResult(false, ChannelSelector.CHANNEL_DINGTALK, recipient, message,
                "DingTalk not available", System.currentTimeMillis() - startTime);
        }

        try {
            // Launch DingTalk
            launchApp(ChannelSelector.CHANNEL_PACKAGES.get(ChannelSelector.CHANNEL_DINGTALK));
            Thread.sleep(1500);

            // Navigate to messages
            automationEngine.click(BySelector.text("消息"));
            Thread.sleep(500);

            if (recipient != null) {
                automationEngine.click(BySelector.textContains(recipient));
                Thread.sleep(500);
            }

            // Send message
            automationEngine.inputText(message);
            Thread.sleep(300);
            automationEngine.click(BySelector.text("发送"));

            long duration = System.currentTimeMillis() - startTime;
            return new SendResult(true, ChannelSelector.CHANNEL_DINGTALK, recipient, message,
                null, duration);

        } catch (Exception e) {
            return new SendResult(false, ChannelSelector.CHANNEL_DINGTALK, recipient, message,
                e.getMessage(), System.currentTimeMillis() - startTime);
        }
    }

    /**
     * Send via WeCom (企业微信)
     */
    @NonNull
    private SendResult sendViaWeCom(@Nullable String recipient,
                                     @NonNull String message,
                                     long startTime) {
        if (automationEngine == null || !channelSelector.isChannelAvailable(ChannelSelector.CHANNEL_WECOM)) {
            return new SendResult(false, ChannelSelector.CHANNEL_WECOM, recipient, message,
                "WeCom not available", System.currentTimeMillis() - startTime);
        }

        try {
            // Launch WeCom
            launchApp(ChannelSelector.CHANNEL_PACKAGES.get(ChannelSelector.CHANNEL_WECOM));
            Thread.sleep(1500);

            // Navigate to messages
            automationEngine.click(BySelector.text("消息"));
            Thread.sleep(500);

            if (recipient != null) {
                automationEngine.click(BySelector.textContains(recipient));
                Thread.sleep(500);
            }

            // Send message
            automationEngine.inputText(message);
            Thread.sleep(300);
            automationEngine.click(BySelector.text("发送"));

            long duration = System.currentTimeMillis() - startTime;
            return new SendResult(true, ChannelSelector.CHANNEL_WECOM, recipient, message,
                null, duration);

        } catch (Exception e) {
            return new SendResult(false, ChannelSelector.CHANNEL_WECOM, recipient, message,
                e.getMessage(), System.currentTimeMillis() - startTime);
        }
    }

    /**
     * Send via QQ
     */
    @NonNull
    private SendResult sendViaQQ(@Nullable String recipient,
                                  @NonNull String message,
                                  long startTime) {
        if (automationEngine == null || !channelSelector.isChannelAvailable(ChannelSelector.CHANNEL_QQ)) {
            return new SendResult(false, ChannelSelector.CHANNEL_QQ, recipient, message,
                "QQ not available", System.currentTimeMillis() - startTime);
        }

        try {
            // Launch QQ
            launchApp(ChannelSelector.CHANNEL_PACKAGES.get(ChannelSelector.CHANNEL_QQ));
            Thread.sleep(1500);

            // Navigate to contacts/messages
            automationEngine.click(BySelector.text("消息"));
            Thread.sleep(500);

            if (recipient != null) {
                automationEngine.click(BySelector.textContains(recipient));
                Thread.sleep(500);
            }

            // Send message
            automationEngine.inputText(message);
            Thread.sleep(300);
            automationEngine.click(BySelector.text("发送"));

            long duration = System.currentTimeMillis() - startTime;
            return new SendResult(true, ChannelSelector.CHANNEL_QQ, recipient, message,
                null, duration);

        } catch (Exception e) {
            return new SendResult(false, ChannelSelector.CHANNEL_QQ, recipient, message,
                e.getMessage(), System.currentTimeMillis() - startTime);
        }
    }

    /**
     * Launch app helper
     */
    @NonNull
    private AutomationResult launchApp(@Nullable String packageName) {
        if (packageName == null) {
            return new AutomationResult("launch", "No package name");
        }

        try {
            Intent intent = context.getPackageManager()
                .getLaunchIntentForPackage(packageName);
            if (intent == null) {
                return new AutomationResult("launch", "App not found: " + packageName);
            }
            intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK);
            context.startActivity(intent);
            return new AutomationResult("launch", true, "App launched", packageName);
        } catch (Exception e) {
            return new AutomationResult("launch", e.getMessage());
        }
    }

    /**
     * Get supported channels
     */
    @NonNull
    public List<String> getSupportedChannels() {
        return channelSelector.getAvailableChannels();
    }
}