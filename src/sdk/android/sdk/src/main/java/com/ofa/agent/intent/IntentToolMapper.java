package com.ofa.agent.intent;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * 意图到工具映射器
 * 将识别的意图映射到具体的工具执行
 */
public class IntentToolMapper {

    private final Map<String, MappingRule> mappings;

    public IntentToolMapper() {
        this.mappings = new ConcurrentHashMap<>();
        initDefaultMappings();
    }

    /**
     * 映射规则
     */
    public static class MappingRule {
        public final String intentId;
        public final String toolName;
        public final Map<String, String> slotToParam;  // 槽位名 -> 工具参数名
        public final Map<String, Object> fixedParams;   // 固定参数
        public final boolean requiresConfirmation;
        public final String confirmationMessage;

        public MappingRule(@NonNull String intentId, @NonNull String toolName,
                           @Nullable Map<String, String> slotToParam,
                           @Nullable Map<String, Object> fixedParams,
                           boolean requiresConfirmation, @Nullable String confirmationMessage) {
            this.intentId = intentId;
            this.toolName = toolName;
            this.slotToParam = slotToParam != null ? new HashMap<>(slotToParam) : new HashMap<>();
            this.fixedParams = fixedParams != null ? new HashMap<>(fixedParams) : new HashMap<>();
            this.requiresConfirmation = requiresConfirmation;
            this.confirmationMessage = confirmationMessage;
        }
    }

    /**
     * 映射结果
     */
    public static class MappingResult {
        public final String toolName;
        public final Map<String, Object> params;
        public final boolean requiresConfirmation;
        public final String confirmationMessage;
        public final List<String> missingRequiredSlots;

        public MappingResult(@NonNull String toolName, @NonNull Map<String, Object> params,
                             boolean requiresConfirmation, @Nullable String confirmationMessage,
                             @Nullable List<String> missingRequiredSlots) {
            this.toolName = toolName;
            this.params = params;
            this.requiresConfirmation = requiresConfirmation;
            this.confirmationMessage = confirmationMessage;
            this.missingRequiredSlots = missingRequiredSlots != null
                    ? new java.util.ArrayList<>(missingRequiredSlots)
                    : new java.util.ArrayList<>();
        }

        public boolean isReady() {
            return missingRequiredSlots.isEmpty();
        }
    }

    /**
     * 初始化默认映射
     * 注意：工具名需与 BuiltInTools 中的 ToolDefinition 名称匹配
     */
    private void initDefaultMappings() {
        // 系统意图 -> 使用 app.* 和 settings.* 工具
        addMapping("system.open_settings", "settings.open",
                null, null, false, null);

        addMapping("system.close_app", "app.launch",
                Map.of("app_name", "packageName"),
                Map.of("operation", "close"), true, "确认关闭应用?");

        addMapping("system.volume", "audio.control",
                Map.of("action", "direction", "level", "level"),
                null, false, null);

        // 通讯意图 -> 使用 contacts.* 和 phone.* 工具
        addMapping("communication.call", "phone.call",
                Map.of("contact", "contactName", "phone_number", "phoneNumber"),
                null, true, "确认拨打 {contact} 的电话?");

        addMapping("communication.sms", "phone.sms",
                Map.of("contact", "recipient", "content", "message"),
                null, true, "确认发送短信给 {recipient}?");

        addMapping("communication.email", "mail.send",
                Map.of("recipient", "to", "subject", "subject", "content", "body"),
                null, false, null);

        // 媒体意图 -> 使用 camera.* 和 media.* 工具
        addMapping("media.capture", "camera.capture",
                Map.of("type", "mode"),
                Map.of("mode", "photo"), false, null);

        addMapping("media.play_music", "media.play",
                Map.of("song", "query", "artist", "artist"),
                Map.of("type", "music"), false, null);

        addMapping("media.stop", "media.control",
                null, Map.of("action", "stop"), false, null);

        addMapping("media.view_image", "media.images",
                Map.of("album", "albumName"),
                null, false, null);

        // 设备意图 -> 使用 wifi.*, bluetooth.*, battery.* 工具
        addMapping("device.wifi_on", "wifi.status",
                null, Map.of("operation", "enable"), false, null);

        addMapping("device.wifi_off", "wifi.status",
                null, Map.of("operation", "disable"), false, null);

        addMapping("device.bluetooth_on", "bluetooth.status",
                null, Map.of("operation", "enable"), false, null);

        addMapping("device.brightness", "display.brightness",
                Map.of("level", "level", "direction", "direction"),
                null, false, null);

        addMapping("device.battery", "battery.status",
                null, null, false, null);

        // 导航意图 -> 使用 location.* 和 maps.* 工具
        addMapping("navigation.navigate", "maps.navigate",
                Map.of("destination", "destination", "origin", "origin"),
                null, true, "确认导航到 {destination}?");

        addMapping("navigation.search_location", "maps.search",
                Map.of("query", "query"),
                null, false, null);

        addMapping("navigation.current_location", "location.get",
                null, null, false, null);

        // 应用意图 -> 使用 app.* 工具
        addMapping("app.open", "app.launch",
                Map.of("app_name", "packageName"),
                null, false, null);

        addMapping("app.search", "search.query",
                Map.of("query", "query", "scope", "scope"),
                null, false, null);

        addMapping("app.share", "content.share",
                Map.of("content", "content", "target", "platform"),
                null, false, null);
    }

    /**
     * 添加映射规则
     */
    public void addMapping(@NonNull String intentId, @NonNull String toolName,
                           @Nullable Map<String, String> slotToParam,
                           @Nullable Map<String, Object> fixedParams,
                           boolean requiresConfirmation, @Nullable String confirmationMessage) {
        mappings.put(intentId, new MappingRule(intentId, toolName, slotToParam,
                fixedParams, requiresConfirmation, confirmationMessage));
    }

    /**
     * 简单映射（无槽位转换）
     */
    public void addSimpleMapping(@NonNull String intentId, @NonNull String toolName) {
        addMapping(intentId, toolName, null, null, false, null);
    }

    /**
     * 移除映射
     */
    public void removeMapping(@NonNull String intentId) {
        mappings.remove(intentId);
    }

    /**
     * 获取映射规则
     */
    @Nullable
    public MappingRule getRule(@NonNull String intentId) {
        return mappings.get(intentId);
    }

    /**
     * 映射意图到工具调用
     */
    @Nullable
    public MappingResult map(@NonNull UserIntent intent) {
        MappingRule rule = mappings.get(intent.getId());
        if (rule == null) {
            // 尝试按 category.action 格式查找
            String fullName = intent.getFullName();
            rule = mappings.get(fullName);
            if (rule == null) {
                return null;
            }
        }

        // 构建参数
        Map<String, Object> params = new HashMap<>(rule.fixedParams);

        // 添加槽位转换
        List<String> missingSlots = new java.util.ArrayList<>();
        for (Map.Entry<String, String> entry : rule.slotToParam.entrySet()) {
            String slotName = entry.getKey();
            String paramName = entry.getValue();

            Object value = intent.getSlot(slotName);
            if (value != null) {
                params.put(paramName, value);
            } else {
                // 检查是否是必需槽位
                IntentDefinition definition = getIntentDefinition(intent.getId());
                if (definition != null && definition.getRequiredSlots().contains(slotName)) {
                    missingSlots.add(slotName);
                }
            }
        }

        // 构建确认消息
        String confirmationMsg = rule.confirmationMessage;
        if (confirmationMsg != null) {
            confirmationMsg = fillTemplate(confirmationMsg, intent.getSlots(), params);
        }

        return new MappingResult(rule.toolName, params, rule.requiresConfirmation,
                confirmationMsg, missingSlots);
    }

    /**
     * 检查意图是否有映射
     */
    public boolean hasMapping(@NonNull String intentId) {
        return mappings.containsKey(intentId) || mappings.containsKey(intentId.replace("_", "."));
    }

    /**
     * 获取所有映射的意图ID
     */
    @NonNull
    public List<String> getAllMappedIntents() {
        return new java.util.ArrayList<>(mappings.keySet());
    }

    /**
     * 填充模板字符串
     */
    private String fillTemplate(@NonNull String template,
                                @NonNull Map<String, Object> slots,
                                @NonNull Map<String, Object> params) {
        String result = template;
        for (Map.Entry<String, Object> entry : slots.entrySet()) {
            result = result.replace("{" + entry.getKey() + "}",
                    String.valueOf(entry.getValue()));
        }
        for (Map.Entry<String, Object> entry : params.entrySet()) {
            result = result.replace("{" + entry.getKey() + "}",
                    String.valueOf(entry.getValue()));
        }
        return result;
    }

    /**
     * 获取意图定义（需要外部提供）
     */
    @Nullable
    private IntentDefinition getIntentDefinition(@NonNull String intentId) {
        // 这个方法需要IntentEngine支持，暂时返回null
        // 在实际使用中应该通过IntentEngine获取
        return null;
    }

    /**
     * 清空所有映射
     */
    public void clear() {
        mappings.clear();
    }

    /**
     * 获取映射数量
     */
    public int size() {
        return mappings.size();
    }
}