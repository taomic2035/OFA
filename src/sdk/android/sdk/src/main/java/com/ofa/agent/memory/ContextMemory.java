package com.ofa.agent.memory;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * 上下文记忆
 * 管理会话内的临时记忆和跨会话的持久记忆
 */
public class ContextMemory {

    private final String sessionId;
    private final UserMemoryManager persistentMemory;

    // 会话内临时记忆
    private final Map<String, Object> sessionVars;

    // 最近行为序列
    private final List<SessionAction> recentActions;

    // 当前上下文
    private String currentSkill;
    private String currentStep;
    private Map<String, Object> currentParams;

    /**
     * 会话行为记录
     */
    public static class SessionAction {
        public final long timestamp;
        public final String action;
        public final Map<String, Object> params;
        public final String result;

        public SessionAction(@NonNull String action, @Nullable Map<String, Object> params, @Nullable String result) {
            this.timestamp = System.currentTimeMillis();
            this.action = action;
            this.params = params != null ? new HashMap<>(params) : new HashMap<>();
            this.result = result;
        }
    }

    public ContextMemory(@NonNull String sessionId, @NonNull UserMemoryManager persistentMemory) {
        this.sessionId = sessionId;
        this.persistentMemory = persistentMemory;
        this.sessionVars = new HashMap<>();
        this.recentActions = new ArrayList<>();
        this.currentParams = new HashMap<>();
    }

    // ===== 会话变量 =====

    /**
     * 设置会话变量
     */
    public void setSessionVar(@NonNull String key, @Nullable Object value) {
        sessionVars.put(key, value);
    }

    /**
     * 获取会话变量
     */
    @Nullable
    public Object getSessionVar(@NonNull String key) {
        return sessionVars.get(key);
    }

    /**
     * 获取会话变量（带默认值）
     */
    @Nullable
    public Object getSessionVar(@NonNull String key, @Nullable Object defaultValue) {
        return sessionVars.getOrDefault(key, defaultValue);
    }

    // ===== 持久记忆集成 =====

    /**
     * 记住用户行为（持久化）
     */
    public void remember(@NonNull String key, @NonNull String value,
                         @Nullable String category, @Nullable Map<String, String> attributes) {
        persistentMemory.rememberPreference(key, value, category, attributes);

        // 同时记录到最近行为
        recentActions.add(new SessionAction("remember:" + key,
                Map.of("value", value), null));
    }

    /**
     * 获取推荐值
     */
    @Nullable
    public String getRecommendedValue(@NonNull String key) {
        return persistentMemory.getRecommendedValue(key);
    }

    /**
     * 获取智能默认值
     */
    @Nullable
    public UserMemoryManager.SmartDefault getSmartDefault(@NonNull String key) {
        return persistentMemory.getSmartDefault(key);
    }

    /**
     * 获取上次使用的值
     */
    @Nullable
    public String getLastValue(@NonNull String key) {
        return persistentMemory.getLastValue(key);
    }

    /**
     * 获取推荐列表
     */
    @NonNull
    public List<String> getRecommendations(@NonNull String key, int limit) {
        return persistentMemory.getRecommendedValues(key, limit);
    }

    /**
     * 自动补全
     */
    @NonNull
    public List<String> autocomplete(@NonNull String key, @NonNull String prefix, int limit) {
        return persistentMemory.autocomplete(key, prefix, limit);
    }

    // ===== 参数自动填充 =====

    /**
     * 智能填充参数
     * 对于未指定的参数，使用记忆中的推荐值
     */
    @NonNull
    public Map<String, Object> fillWithDefaults(@NonNull String skillId,
                                                  @NonNull Map<String, Object> params,
                                                  @NonNull List<String> fillableParams) {
        Map<String, Object> result = new HashMap<>(params);

        for (String param : fillableParams) {
            if (!result.containsKey(param) || result.get(param) == null) {
                String key = skillId + "." + param;
                String defaultValue = getRecommendedValue(key);
                if (defaultValue != null) {
                    result.put(param, defaultValue);
                    // 标记为自动填充
                    result.put("_auto_filled_" + param, true);
                }
            }
        }

        return result;
    }

    /**
     * 智能排序选项
     * 根据用户历史偏好排序选项列表
     */
    @NonNull
    public <T> List<T> sortOptions(@NonNull String key, @NonNull List<T> options) {
        if (options.size() <= 1) return new ArrayList<>(options);

        List<MemoryEntry> memories = persistentMemory.getEntries(key);
        if (memories.isEmpty()) return new ArrayList<>(options);

        // 构建排序权重
        Map<String, Float> weights = new HashMap<>();
        for (MemoryEntry entry : memories) {
            weights.put(entry.getValue(), entry.calculateRecommendationScore());
        }

        // 排序
        List<T> sorted = new ArrayList<>(options);
        sorted.sort((a, b) -> {
            float wa = weights.getOrDefault(String.valueOf(a), 0f);
            float wb = weights.getOrDefault(String.valueOf(b), 0f);
            return Float.compare(wb, wa); // 降序
        });

        return sorted;
    }

    // ===== 当前上下文 =====

    public void setCurrentSkill(@Nullable String skillId) {
        this.currentSkill = skillId;
    }

    @Nullable
    public String getCurrentSkill() {
        return currentSkill;
    }

    public void setCurrentStep(@Nullable String stepId) {
        this.currentStep = stepId;
    }

    @Nullable
    public String getCurrentStep() {
        return currentStep;
    }

    public void setCurrentParams(@Nullable Map<String, Object> params) {
        this.currentParams = params != null ? new HashMap<>(params) : new HashMap<>();
    }

    @NonNull
    public Map<String, Object> getCurrentParams() {
        return new HashMap<>(currentParams);
    }

    // ===== 行为记录 =====

    /**
     * 记录行为
     */
    public void recordAction(@NonNull String action, @Nullable Map<String, Object> params, @Nullable String result) {
        recentActions.add(new SessionAction(action, params, result));

        // 保持最近100条
        if (recentActions.size() > 100) {
            recentActions.remove(0);
        }
    }

    /**
     * 获取最近行为
     */
    @NonNull
    public List<SessionAction> getRecentActions(int limit) {
        int start = Math.max(0, recentActions.size() - limit);
        return new ArrayList<>(recentActions.subList(start, recentActions.size()));
    }

    /**
     * 获取最后一次行为
     */
    @Nullable
    public SessionAction getLastAction() {
        return recentActions.isEmpty() ? null : recentActions.get(recentActions.size() - 1);
    }

    // ===== 用户画像 =====

    /**
     * 获取用户偏好摘要
     */
    @NonNull
    public UserPreferenceSummary getPreferenceSummary() {
        UserMemoryManager.MemoryStats stats = persistentMemory.getStats();

        Map<String, String> topPreferences = new HashMap<>();
        // 获取一些关键偏好

        return new UserPreferenceSummary(
                sessionId,
                stats.totalKeys,
                stats.totalEntries,
                recentActions.size(),
                topPreferences
        );
    }

    public static class UserPreferenceSummary {
        public final String sessionId;
        public final int memoryKeys;
        public final int memoryEntries;
        public final int sessionActions;
        public final Map<String, String> topPreferences;

        public UserPreferenceSummary(String sessionId, int memoryKeys, int memoryEntries,
                                      int sessionActions, Map<String, String> topPreferences) {
            this.sessionId = sessionId;
            this.memoryKeys = memoryKeys;
            this.memoryEntries = memoryEntries;
            this.sessionActions = sessionActions;
            this.topPreferences = topPreferences;
        }
    }

    // ===== 清理 =====

    /**
     * 清除会话记忆
     */
    public void clearSession() {
        sessionVars.clear();
        recentActions.clear();
        currentSkill = null;
        currentStep = null;
        currentParams.clear();
    }

    /**
     * 获取会话ID
     */
    @NonNull
    public String getSessionId() {
        return sessionId;
    }
}