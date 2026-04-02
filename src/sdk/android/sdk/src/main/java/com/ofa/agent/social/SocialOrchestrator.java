package com.ofa.agent.social;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationEngine;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.memory.UserMemoryManager;
import com.ofa.agent.social.adapter.ContactAdapter;
import com.ofa.agent.social.adapter.ContactInfo;

import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.CompletableFuture;
import java.util.stream.Collectors;

/**
 * Social Notification Orchestrator - intelligent social message delivery system.
 *
 * Key Features:
 * 1. Smart message classification (invitation, urgent, guide, payment, etc.)
 * 2. Intelligent channel selection based on modern social habits
 * 3. Multi-channel delivery with fallback
 * 4. User preference learning
 * 5. Contact integration
 * 6. Delivery status tracking
 *
 * Modern Social Habits:
 * - 约吃饭 → 微信 (social discussion, easy to confirm)
 * - 紧急重要 → 电话 (immediate response)
 * - 攻略分享 → 小红书私信 (content sharing platform)
 * - 支付提醒 → 支付宝 (financial platform)
 * - 工作通知 → 钉钉/企业微信 (business context)
 * - 日常聊天 → 微信/抖音 (social platforms)
 */
public class SocialOrchestrator {

    private static final String TAG = "SocialOrchestrator";

    private final Context context;
    private final UserMemoryManager memoryManager;

    // Core components
    private final MessageClassifier messageClassifier;
    private final ChannelSelector channelSelector;
    private final MessageSender messageSender;
    private final ContactAdapter contactAdapter;

    // Configuration
    private boolean enableMultiChannelFallback = true;
    private boolean enablePreferenceLearning = true;
    private boolean enableDeliveryTracking = true;
    private int maxFallbackChannels = 3;

    // Delivery history
    private final List<DeliveryRecord> deliveryHistory;
    private final Map<String, DeliveryStats> channelStats;

    // Listeners
    private NotificationListener listener;

    /**
     * Notification listener interface
     */
    public interface NotificationListener {
        void onNotificationStart(NotificationRequest request);
        void onChannelSelected(String channel, String reason);
        void onDeliveryStart(String channel);
        void onDeliverySuccess(DeliveryRecord record);
        void onDeliveryFailure(DeliveryRecord record);
        void onFallback(String fromChannel, String toChannel, String reason);
    }

    /**
     * Notification request
     */
    public static class NotificationRequest {
        public final String message;
        public final String recipientName;
        public final String recipientPhone;
        public final String messageTypeOverride;
        public final String channelOverride;
        public final Map<String, String> additionalInfo;
        public final long timestamp;

        public NotificationRequest(String message, String recipientName, String recipientPhone) {
            this(message, recipientName, recipientPhone, null, null, new HashMap<>());
        }

        public NotificationRequest(String message, String recipientName, String recipientPhone,
                                    String messageTypeOverride, String channelOverride,
                                    Map<String, String> additionalInfo) {
            this.message = message;
            this.recipientName = recipientName;
            this.recipientPhone = recipientPhone;
            this.messageTypeOverride = messageTypeOverride;
            this.channelOverride = channelOverride;
            this.additionalInfo = additionalInfo;
            this.timestamp = System.currentTimeMillis();
        }

        @NonNull
        @Override
        public String toString() {
            return String.format("NotificationRequest{to=%s, msg='%s'}",
                recipientName, message.substring(0, Math.min(50, message.length())));
        }
    }

    /**
     * Delivery record
     */
    public static class DeliveryRecord {
        public final String id;
        public final String message;
        public final String messageType;
        public final String primaryChannel;
        public final String recipientName;
        public final String recipientPhone;
        public final List<String> attemptedChannels;
        public final String successfulChannel;
        public final boolean success;
        public final String failureReason;
        public final long startTime;
        public final long endTime;
        public final long durationMs;
        public final int urgencyLevel;
        public final double confidence;

        public DeliveryRecord(String id, String message, String messageType,
                              String primaryChannel, String recipientName, String recipientPhone,
                              List<String> attemptedChannels, String successfulChannel,
                              boolean success, String failureReason,
                              long startTime, long endTime, int urgencyLevel, double confidence) {
            this.id = id;
            this.message = message;
            this.messageType = messageType;
            this.primaryChannel = primaryChannel;
            this.recipientName = recipientName;
            this.recipientPhone = recipientPhone;
            this.attemptedChannels = attemptedChannels;
            this.successfulChannel = successfulChannel;
            this.success = success;
            this.failureReason = failureReason;
            this.startTime = startTime;
            this.endTime = endTime;
            this.durationMs = endTime - startTime;
            this.urgencyLevel = urgencyLevel;
            this.confidence = confidence;
        }

        @NonNull
        @Override
        public String toString() {
            return String.format("DeliveryRecord{id=%s, type=%s, channel=%s, success=%s}",
                id, messageType, successfulChannel != null ? successfulChannel : primaryChannel, success);
        }
    }

    /**
     * Delivery statistics per channel
     */
    private static class DeliveryStats {
        int totalAttempts = 0;
        int successCount = 0;
        int failureCount = 0;
        long totalDurationMs = 0;

        void recordSuccess(long durationMs) {
            totalAttempts++;
            successCount++;
            totalDurationMs += durationMs;
        }

        void recordFailure() {
            totalAttempts++;
            failureCount++;
        }

        double getSuccessRate() {
            return totalAttempts > 0 ? (double) successCount / totalAttempts : 0;
        }

        double getAverageDuration() {
            return successCount > 0 ? (double) totalDurationMs / successCount : 0;
        }
    }

    /**
     * Create social orchestrator
     */
    public SocialOrchestrator(@NonNull Context context,
                               @Nullable AutomationEngine automationEngine,
                               @Nullable UserMemoryManager memoryManager) {
        this.context = context;
        this.memoryManager = memoryManager;

        this.messageClassifier = new MessageClassifier(context, memoryManager);
        this.channelSelector = new ChannelSelector(context, automationEngine, memoryManager);
        this.messageSender = new MessageSender(context, automationEngine, channelSelector);
        this.contactAdapter = new ContactAdapter(context);

        this.deliveryHistory = new ArrayList<>();
        this.channelStats = new HashMap<>();

        // Initialize stats for each channel
        for (String channel : channelSelector.getAvailableChannels()) {
            channelStats.put(channel, new DeliveryStats());
        }

        Log.i(TAG, "Social Orchestrator initialized with " +
            channelSelector.getAvailableChannels().size() + " channels");
    }

    /**
     * Set notification listener
     */
    public void setListener(@Nullable NotificationListener listener) {
        this.listener = listener;
    }

    /**
     * Configure fallback behavior
     */
    public void setMultiChannelFallback(boolean enabled, int maxChannels) {
        this.enableMultiChannelFallback = enabled;
        this.maxFallbackChannels = maxChannels;
    }

    /**
     * Send smart notification - main entry point
     *
     * Automatically:
     * 1. Classifies message type
     * 2. Determines urgency
     * 3. Selects optimal channel based on social habits
     * 4. Delivers message
     * 5. Falls back to alternative channels if needed
     * 6. Records delivery for learning
     */
    @NonNull
    public DeliveryRecord sendNotification(@NonNull String message,
                                            @Nullable String recipientName,
                                            @Nullable String recipientPhone) {
        return sendNotification(new NotificationRequest(message, recipientName, recipientPhone));
    }

    /**
     * Send notification with request object
     */
    @NonNull
    public DeliveryRecord sendNotification(@NonNull NotificationRequest request) {
        String id = generateId();
        long startTime = System.currentTimeMillis();

        Log.i(TAG, "Processing notification: " + request);

        // Notify listener
        if (listener != null) {
            listener.onNotificationStart(request);
        }

        // Step 1: Classify message
        MessageClassifier.ClassificationResult classification;
        if (request.messageTypeOverride != null) {
            // Use override type
            classification = new MessageClassifier.ClassificationResult(
                request.messageTypeOverride,
                MessageClassifier.URGENCY_MEDIUM,
                request.channelOverride != null ? request.channelOverride : ChannelSelector.CHANNEL_WECHAT,
                1.0,
                request.additionalInfo,
                "Manual override"
            );
        } else {
            // Auto-classify
            classification = messageClassifier.classify(
                request.message, request.recipientName, request.recipientPhone);
        }

        Log.i(TAG, "Classification: " + classification);

        // Step 2: Select channel
        String primaryChannel = classification.recommendedChannel;

        // Check if user has preference for this contact
        if (enablePreferenceLearning && request.recipientName != null) {
            String preferred = messageClassifier.getChannelPreference(request.recipientName);
            if (preferred != null && channelSelector.isChannelAvailable(preferred)) {
                primaryChannel = preferred;
                Log.d(TAG, "Using user preference: " + preferred);
            }
        }

        // Override if specified
        if (request.channelOverride != null &&
            channelSelector.isChannelAvailable(request.channelOverride)) {
            primaryChannel = request.channelOverride;
        }

        if (listener != null) {
            listener.onChannelSelected(primaryChannel, classification.reason);
        }

        // Step 3: Build fallback chain
        List<String> channelChain = buildChannelChain(primaryChannel, classification);

        // Step 4: Try delivery
        List<String> attemptedChannels = new ArrayList<>();
        String successfulChannel = null;
        String failureReason = null;

        for (int i = 0; i < channelChain.size(); i++) {
            String channel = channelChain.get(i);

            if (!channelSelector.isChannelAvailable(channel)) {
                continue;
            }

            attemptedChannels.add(channel);

            if (listener != null) {
                listener.onDeliveryStart(channel);
            }

            MessageSender.SendResult result = messageSender.send(
                channel, request.recipientName, request.recipientPhone, request.message);

            if (result.success) {
                successfulChannel = channel;
                // Update stats
                channelStats.computeIfAbsent(channel, k -> new DeliveryStats())
                    .recordSuccess(result.durationMs);

                if (listener != null) {
                    listener.onDeliverySuccess(createRecord(id, request, classification,
                        primaryChannel, attemptedChannels, successfulChannel, true, null,
                        startTime, System.currentTimeMillis()));
                }

                break;
            } else {
                failureReason = result.error;
                channelStats.computeIfAbsent(channel, k -> new DeliveryStats()).recordFailure();

                // Fallback to next channel
                if (enableMultiChannelFallback && i < channelChain.size() - 1) {
                    String nextChannel = channelChain.get(i + 1);
                    if (listener != null) {
                        listener.onFallback(channel, nextChannel, result.error);
                    }
                    Log.w(TAG, "Fallback from " + channel + " to " + nextChannel);
                }
            }
        }

        long endTime = System.currentTimeMillis();

        // Step 5: Create and record delivery
        DeliveryRecord record = createRecord(id, request, classification,
            primaryChannel, attemptedChannels, successfulChannel,
            successfulChannel != null, successfulChannel == null ? failureReason : null,
            startTime, endTime);

        // Record in history
        if (enableDeliveryTracking) {
            deliveryHistory.add(record);
            if (deliveryHistory.size() > 1000) {
                // Keep last 1000 records
                deliveryHistory.remove(0);
            }
            saveDeliveryToMemory(record);
        }

        // Step 6: Learn from success/failure
        if (enablePreferenceLearning && successfulChannel != null && request.recipientName != null) {
            learnFromDelivery(request.recipientName, successfulChannel, classification);
        }

        if (!record.success && listener != null) {
            listener.onDeliveryFailure(record);
        }

        Log.i(TAG, "Delivery complete: " + record);
        return record;
    }

    /**
     * Send notification asynchronously
     */
    @NonNull
    public CompletableFuture<DeliveryRecord> sendNotificationAsync(@NonNull String message,
                                                                    @Nullable String recipientName,
                                                                    @Nullable String recipientPhone) {
        return CompletableFuture.supplyAsync(() ->
            sendNotification(message, recipientName, recipientPhone));
    }

    /**
     * Batch send notifications
     */
    @NonNull
    public List<DeliveryRecord> sendBatchNotifications(@NonNull List<NotificationRequest> requests) {
        return requests.stream()
            .map(this::sendNotification)
            .collect(Collectors.toList());
    }

    /**
     * Build channel fallback chain
     */
    @NonNull
    private List<String> buildChannelChain(@NonNull String primaryChannel,
                                            @NonNull MessageClassifier.ClassificationResult classification) {
        List<String> chain = new ArrayList<>();
        chain.add(primaryChannel);

        if (!enableMultiChannelFallback) {
            return chain;
        }

        // Get recommended channels for type
        List<String> recommended = channelSelector.getAvailableChannels();

        // Add alternatives based on urgency
        if (classification.urgencyLevel >= MessageClassifier.URGENCY_HIGH) {
            // High urgency: add phone/sms as backup
            if (!chain.contains(ChannelSelector.CHANNEL_PHONE) &&
                channelSelector.isChannelAvailable(ChannelSelector.CHANNEL_PHONE)) {
                chain.add(ChannelSelector.CHANNEL_PHONE);
            }
            if (!chain.contains(ChannelSelector.CHANNEL_SMS)) {
                chain.add(ChannelSelector.CHANNEL_SMS);
            }
        }

        // Add remaining available channels
        for (String channel : recommended) {
            if (!chain.contains(channel) && chain.size() < maxFallbackChannels) {
                chain.add(channel);
            }
        }

        return chain;
    }

    /**
     * Create delivery record
     */
    @NonNull
    private DeliveryRecord createRecord(@NonNull String id,
                                         @NonNull NotificationRequest request,
                                         @NonNull MessageClassifier.ClassificationResult classification,
                                         @NonNull String primaryChannel,
                                         @NonNull List<String> attemptedChannels,
                                         @Nullable String successfulChannel,
                                         boolean success,
                                         @Nullable String failureReason,
                                         long startTime, long endTime) {
        return new DeliveryRecord(
            id, request.message, classification.messageType,
            primaryChannel, request.recipientName, request.recipientPhone,
            new ArrayList<>(attemptedChannels), successfulChannel,
            success, failureReason, startTime, endTime,
            classification.urgencyLevel, classification.confidence
        );
    }

    /**
     * Learn from successful delivery
     */
    private void learnFromDelivery(@NonNull String contactName,
                                    @NonNull String successfulChannel,
                                    @NonNull MessageClassifier.ClassificationResult classification) {
        // Only save preference for high confidence classifications
        if (classification.confidence >= 0.7) {
            messageClassifier.saveChannelPreference(contactName, successfulChannel);
            Log.d(TAG, "Saved preference: " + contactName + " → " + successfulChannel);
        }
    }

    /**
     * Save delivery to memory
     */
    private void saveDeliveryToMemory(@NonNull DeliveryRecord record) {
        if (memoryManager == null) return;

        try {
            JSONObject json = new JSONObject();
            json.put("id", record.id);
            json.put("type", record.messageType);
            json.put("channel", record.successfulChannel != null ?
                record.successfulChannel : record.primaryChannel);
            json.put("success", record.success);
            json.put("duration", record.durationMs);
            json.put("urgency", record.urgencyLevel);
            json.put("recipient", record.recipientName);

            memoryManager.set("social.delivery." + record.id, json.toString());
        } catch (Exception e) {
            Log.w(TAG, "Failed to save delivery: " + e.getMessage());
        }
    }

    /**
     * Generate unique ID
     */
    @NonNull
    private String generateId() {
        return "notif_" + System.currentTimeMillis() + "_" +
            (int)(Math.random() * 10000);
    }

    // ===== Contact Management =====

    /**
     * Find contact by name
     */
    @Nullable
    public ContactInfo findContact(@NonNull String name) {
        return contactAdapter.findContact(name);
    }

    /**
     * Get all contacts
     */
    @NonNull
    public List<ContactInfo> getAllContacts() {
        return contactAdapter.getAllContacts();
    }

    /**
     * Search contacts
     */
    @NonNull
    public List<ContactInfo> searchContacts(@NonNull String query) {
        return contactAdapter.searchContacts(query);
    }

    // ===== Statistics & Reporting =====

    /**
     * Get delivery history
     */
    @NonNull
    public List<DeliveryRecord> getDeliveryHistory(int limit) {
        if (limit <= 0 || limit >= deliveryHistory.size()) {
            return new ArrayList<>(deliveryHistory);
        }
        return deliveryHistory.subList(
            Math.max(0, deliveryHistory.size() - limit), deliveryHistory.size());
    }

    /**
     * Get channel statistics
     */
    @NonNull
    public Map<String, Map<String, Double>> getChannelStatistics() {
        Map<String, Map<String, Double>> stats = new HashMap<>();

        for (Map.Entry<String, DeliveryStats> entry : channelStats.entrySet()) {
            Map<String, Double> channelData = new HashMap<>();
            channelData.put("successRate", entry.getValue().getSuccessRate());
            channelData.put("avgDuration", entry.getValue().getAverageDuration());
            channelData.put("totalAttempts", (double) entry.getValue().totalAttempts);
            stats.put(entry.getKey(), channelData);
        }

        return stats;
    }

    /**
     * Get status report
     */
    @NonNull
    public String getStatusReport() {
        StringBuilder sb = new StringBuilder();
        sb.append("=== Social Notification Orchestrator ===\n\n");

        sb.append("Available Channels:\n");
        sb.append(channelSelector.getAvailabilityReport()).append("\n");

        sb.append("Channel Statistics:\n");
        Map<String, Map<String, Double>> stats = getChannelStatistics();
        for (Map.Entry<String, Map<String, Double>> entry : stats.entrySet()) {
            Map<String, Double> data = entry.getValue();
            sb.append(String.format("  %s: success=%.1f%%, avg=%.0fms, attempts=%d\n",
                entry.getKey(),
                data.get("successRate") * 100,
                data.get("avgDuration"),
                data.get("totalAttempts").intValue()));
        }

        sb.append("\nDelivery History: ").append(deliveryHistory.size()).append(" records\n");

        sb.append("\nMessage Types:\n");
        for (String type : messageClassifier.getMessageTypes()) {
            sb.append("  - ").append(type).append("\n");
        }

        return sb.toString();
    }

    // ===== Convenience Methods =====

    /**
     * Send invitation (约吃饭等)
     */
    @NonNull
    public DeliveryRecord sendInvitation(@NonNull String activity,
                                          @Nullable String time,
                                          @NonNull String recipientName,
                                          @Nullable String recipientPhone) {
        String message = "约你" + activity;
        if (time != null) {
            message += "，" + time;
        }

        NotificationRequest request = new NotificationRequest(
            message, recipientName, recipientPhone,
            MessageClassifier.TYPE_INVITATION, null,
            Map.of("activity", activity, "time", time != null ? time : "")
        );

        return sendNotification(request);
    }

    /**
     * Send urgent message (紧急重要)
     */
    @NonNull
    public DeliveryRecord sendUrgent(@NonNull String message,
                                      @NonNull String recipientName,
                                      @Nullable String recipientPhone) {
        NotificationRequest request = new NotificationRequest(
            "[紧急] " + message, recipientName, recipientPhone,
            MessageClassifier.TYPE_URGENT, ChannelSelector.CHANNEL_PHONE,
            new HashMap<>()
        );

        return sendNotification(request);
    }

    /**
     * Send guide/tips (攻略分享)
     */
    @NonNull
    public DeliveryRecord sendGuide(@NonNull String title,
                                     @NonNull String content,
                                     @NonNull String recipientName) {
        String message = "分享一个攻略：" + title + "\n" + content;

        NotificationRequest request = new NotificationRequest(
            message, recipientName, null,
            MessageClassifier.TYPE_GUIDE, ChannelSelector.CHANNEL_XIAOHONGSHU,
            Map.of("title", title)
        );

        return sendNotification(request);
    }

    /**
     * Send payment reminder (支付提醒)
     */
    @NonNull
    public DeliveryRecord sendPaymentReminder(@NonNull String amount,
                                               @NonNull String recipientName,
                                               @Nullable String recipientPhone) {
        String message = "请支付：" + amount + "元";

        NotificationRequest request = new NotificationRequest(
            message, recipientName, recipientPhone,
            MessageClassifier.TYPE_PAYMENT, ChannelSelector.CHANNEL_ALIPAY,
            Map.of("amount", amount)
        );

        return sendNotification(request);
    }

    /**
     * Shutdown
     */
    public void shutdown() {
        Log.i(TAG, "Shutting down Social Orchestrator...");
        deliveryHistory.clear();
        channelStats.clear();
    }
}