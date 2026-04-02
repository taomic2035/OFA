package com.ofa.agent.intent;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.ArrayList;
import java.util.Collections;
import java.util.Comparator;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * 意图识别引擎
 * 核心引擎，负责解析用户输入并识别意图
 */
public class IntentEngine {

    private static final double CONFIDENCE_THRESHOLD = 0.3;
    private static final int MAX_RESULTS = 5;

    private final Map<String, IntentDefinition> definitions;
    private final Map<String, List<IntentDefinition>> categoryIndex;

    public IntentEngine() {
        this.definitions = new ConcurrentHashMap<>();
        this.categoryIndex = new ConcurrentHashMap<>();
    }

    /**
     * 注册意图定义
     */
    public void register(@NonNull IntentDefinition definition) {
        definitions.put(definition.getId(), definition);
        categoryIndex.computeIfAbsent(definition.getCategory(), k -> new ArrayList<>())
                .add(definition);
    }

    /**
     * 批量注册意图定义
     */
    public void registerAll(@NonNull List<IntentDefinition> definitions) {
        for (IntentDefinition def : definitions) {
            register(def);
        }
    }

    /**
     * 注销意图定义
     */
    public void unregister(@NonNull String intentId) {
        IntentDefinition removed = definitions.remove(intentId);
        if (removed != null) {
            List<IntentDefinition> categoryDefs = categoryIndex.get(removed.getCategory());
            if (categoryDefs != null) {
                categoryDefs.remove(removed);
            }
        }
    }

    /**
     * 获取已注册的意图定义
     */
    @Nullable
    public IntentDefinition getDefinition(@NonNull String intentId) {
        return definitions.get(intentId);
    }

    /**
     * 获取某分类下的所有意图
     */
    @NonNull
    public List<IntentDefinition> getDefinitionsByCategory(@NonNull String category) {
        List<IntentDefinition> result = categoryIndex.get(category);
        return result != null ? new ArrayList<>(result) : Collections.emptyList();
    }

    /**
     * 获取所有已注册的意图定义
     */
    @NonNull
    public List<IntentDefinition> getAllDefinitions() {
        return new ArrayList<>(definitions.values());
    }

    /**
     * 识别用户输入的意图
     * 返回按置信度排序的意图列表
     */
    @NonNull
    public List<UserIntent> recognize(@NonNull String input) {
        return recognize(input, CONFIDENCE_THRESHOLD);
    }

    /**
     * 识别用户输入的意图
     * @param input 用户输入
     * @param minConfidence 最小置信度阈值
     * @return 按置信度排序的意图列表
     */
    @NonNull
    public List<UserIntent> recognize(@NonNull String input, double minConfidence) {
        List<IntentMatch> matches = new ArrayList<>();

        // 遍历所有定义，计算匹配分数
        for (IntentDefinition def : definitions.values()) {
            double score = def.matchScore(input);
            if (score >= minConfidence) {
                // 提取槽位
                Map<String, Object> slots = extractSlots(def, input);
                matches.add(new IntentMatch(def, score, slots));
            }
        }

        // 按置信度降序排序
        Collections.sort(matches, (a, b) -> Double.compare(b.score, a.score));

        // 限制结果数量
        List<UserIntent> results = new ArrayList<>();
        int count = Math.min(matches.size(), MAX_RESULTS);
        for (int i = 0; i < count; i++) {
            IntentMatch match = matches.get(i);
            results.add(new UserIntent.Builder()
                    .id(match.definition.getId())
                    .category(match.definition.getCategory())
                    .action(match.definition.getAction())
                    .confidence(match.score)
                    .slots(match.slots)
                    .rawInput(input)
                    .build());
        }

        return results;
    }

    /**
     * 识别单个最佳意图
     */
    @Nullable
    public UserIntent recognizeBest(@NonNull String input) {
        List<UserIntent> results = recognize(input);
        return results.isEmpty() ? null : results.get(0);
    }

    /**
     * 识别特定分类的意图
     */
    @Nullable
    public UserIntent recognizeInCategory(@NonNull String input, @NonNull String category) {
        List<IntentDefinition> categoryDefs = categoryIndex.get(category);
        if (categoryDefs == null || categoryDefs.isEmpty()) {
            return null;
        }

        IntentMatch bestMatch = null;
        for (IntentDefinition def : categoryDefs) {
            double score = def.matchScore(input);
            if (score >= CONFIDENCE_THRESHOLD) {
                if (bestMatch == null || score > bestMatch.score) {
                    Map<String, Object> slots = extractSlots(def, input);
                    bestMatch = new IntentMatch(def, score, slots);
                }
            }
        }

        if (bestMatch != null) {
            return new UserIntent.Builder()
                    .id(bestMatch.definition.getId())
                    .category(bestMatch.definition.getCategory())
                    .action(bestMatch.definition.getAction())
                    .confidence(bestMatch.score)
                    .slots(bestMatch.slots)
                    .rawInput(input)
                    .build();
        }

        return null;
    }

    /**
     * 提取槽位值
     */
    @NonNull
    private Map<String, Object> extractSlots(@NonNull IntentDefinition definition, @NonNull String input) {
        Map<String, Object> slots = new ConcurrentHashMap<>();

        for (IntentDefinition.SlotDefinition slotDef : definition.getSlots()) {
            Object value = slotDef.extract(input);
            if (value != null) {
                slots.put(slotDef.name, value);
            }
        }

        return slots;
    }

    /**
     * 清空所有注册
     */
    public void clear() {
        definitions.clear();
        categoryIndex.clear();
    }

    /**
     * 获取已注册的意图数量
     */
    public int size() {
        return definitions.size();
    }

    /**
     * 内部匹配结果类
     */
    private static class IntentMatch {
        final IntentDefinition definition;
        final double score;
        final Map<String, Object> slots;

        IntentMatch(IntentDefinition definition, double score, Map<String, Object> slots) {
            this.definition = definition;
            this.score = score;
            this.slots = slots;
        }
    }
}