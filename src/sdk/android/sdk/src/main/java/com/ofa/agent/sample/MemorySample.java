package com.ofa.agent.sample;

import android.content.Context;

import androidx.annotation.NonNull;

import com.ofa.agent.memory.MemoryArchive;
import com.ofa.agent.memory.MemoryEntry;
import com.ofa.agent.memory.MemoryStats;
import com.ofa.agent.memory.SmartDefault;
import com.ofa.agent.memory.UserMemoryManager;

import java.io.File;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Memory System Sample - 三层记忆系统使用示例
 *
 * 三层架构:
 * L1: MemoryCache - 内存缓存 (LRU策略, 快速访问)
 * L2: Room Database - 持久化存储 (可靠存储)
 * L3: MemoryArchive - 文件归档 (冷数据备份)
 *
 * 使用场景:
 * 1. 记住用户偏好 - 奶茶口味、甜度、地址等
 * 2. 智能推荐 - 基于使用频率、最近使用、上下文
 * 3. 自动补全 - 输入部分内容自动推荐完整值
 * 4. 导入导出 - 备份和恢复用户记忆
 */
public class MemorySample {

    private final Context context;
    private final UserMemoryManager memoryManager;

    public MemorySample(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.memoryManager = UserMemoryManager.getInstance(context);
    }

    // ===== 基础记忆操作 =====

    /**
     * 示例: 记录点奶茶的偏好
     */
    public void rememberBubbleTeaPreferences() {
        // 记录奶茶名称偏好
        memoryManager.rememberPreference(
                "bubble_tea.drink_name",
                "珍珠奶茶",
                "food",
                createAttributes("sweetness", "五分糖", "ice", "少冰", "size", "大杯")
        );

        // 记录奶茶品牌偏好
        memoryManager.rememberPreference(
                "bubble_tea.brand",
                "喜茶",
                "food",
                null
        );

        // 记录奶茶甜度偏好
        memoryManager.rememberPreference(
                "bubble_tea.sweetness",
                "五分糖",
                "food",
                null
        );

        // 记录地址
        memoryManager.rememberPreference(
                "delivery.address",
                "北京市朝阳区xxx街道xxx号",
                "location",
                createAttributes("tag", "家", "lat", "39.9", "lng", "116.4")
        );

        System.out.println("已记录奶茶偏好");
    }

    /**
     * 示例: 记录技能执行参数
     */
    public void rememberSkillExecution() {
        Map<String, Object> params = new HashMap<>();
        params.put("drink_name", "芝芝莓莓");
        params.put("sweetness", "三分糖");
        params.put("ice", "去冰");
        params.put("size", "中杯");

        memoryManager.rememberSkillParams("order_bubble_tea", params);
        System.out.println("已记录技能执行参数");
    }

    /**
     * 示例: 记录复杂记忆条目
     */
    public void rememberComplexEntry() {
        MemoryEntry entry = new MemoryEntry.Builder()
                .key("bubble_tea.order_history")
                .category("food")
                .value("喜茶-芝芝莓莓-三分糖-去冰-中杯")
                .attribute("brand", "喜茶")
                .attribute("drink", "芝芝莓莓")
                .attribute("sweetness", "三分糖")
                .attribute("ice", "去冰")
                .attribute("size", "中杯")
                .attribute("price", "32")
                .score(2.5f)  // 设置较高的初始分数
                .context("点奶茶技能")
                .build();

        memoryManager.remember(entry);
        System.out.println("已记录复杂记忆条目: " + entry);
    }

    // ===== 查询与推荐 =====

    /**
     * 示例: 获取推荐值
     */
    public void getRecommendations() {
        // 获取推荐的奶茶名称
        String recommendedDrink = memoryManager.getRecommendedValue("bubble_tea.drink_name");
        System.out.println("推荐奶茶: " + recommendedDrink);

        // 获取推荐列表（前3个）
        List<String> topDrinks = memoryManager.getRecommendedValues("bubble_tea.drink_name", 3);
        System.out.println("推荐列表: " + topDrinks);

        // 获取智能默认值
        SmartDefault smartDefault = memoryManager.getSmartDefault("bubble_tea.drink_name");
        if (smartDefault != null) {
            System.out.println("智能默认值:");
            System.out.println("  推荐: " + smartDefault.recommendedValue);
            System.out.println("  最近使用: " + smartDefault.lastUsedValue);
            System.out.println("  最常用: " + smartDefault.mostUsedValue);
            System.out.println("  置信度: " + smartDefault.confidence);
        }

        // 获取最后使用的值
        String lastUsed = memoryManager.getLastValue("bubble_tea.drink_name");
        System.out.println("最后使用: " + lastUsed);
    }

    /**
     * 示例: 自动补全
     */
    public void demonstrateAutocomplete() {
        // 用户输入"珍"，系统自动补全
        List<String> suggestions = memoryManager.autocomplete("bubble_tea.drink_name", "珍", 5);
        System.out.println("补全建议（输入'珍'): " + suggestions);

        // 用户输入"芝"，系统自动补全
        suggestions = memoryManager.autocomplete("bubble_tea.drink_name", "芝", 5);
        System.out.println("补全建议（输入'芝'): " + suggestions);
    }

    /**
     * 示例: 搜索记忆
     */
    public void searchMemories() {
        // 搜索包含"奶茶"的记忆
        List<MemoryEntry> results = memoryManager.search("奶茶");
        System.out.println("搜索'奶茶'结果: " + results.size() + " 条");

        for (MemoryEntry entry : results) {
            System.out.println("  - " + entry.getKey() + ": " + entry.getValue());
        }
    }

    /**
     * 示例: 获取分类记忆
     */
    public void getCategoryMemories() {
        // 获取food分类下的所有记忆
        List<MemoryEntry> foodMemories = memoryManager.getEntriesByCategory("food");
        System.out.println("food分类记忆: " + foodMemories.size() + " 条");

        for (MemoryEntry entry : foodMemories) {
            System.out.println("  - " + entry.getKey() + ": " + entry.getValue() +
                    " (count: " + entry.getCount() + ")");
        }
    }

    // ===== 导入导出 =====

    /**
     * 示例: 导出记忆
     */
    public void exportMemories() {
        memoryManager.exportMemories(new MemoryArchive.ExportCallback() {
            @Override
            public void onSuccess(File file) {
                System.out.println("导出成功: " + file.getAbsolutePath());
            }

            @Override
            public void onError(String error) {
                System.out.println("导出失败: " + error);
            }
        });
    }

    /**
     * 示例: 导入记忆
     */
    public void importMemories(File importFile) {
        memoryManager.importMemories(importFile, new MemoryArchive.ImportCallback() {
            @Override
            public void onSuccess(List<MemoryEntry> entries) {
                System.out.println("导入成功: " + entries.size() + " 条记忆");
            }

            @Override
            public void onError(String error) {
                System.out.println("导入失败: " + error);
            }
        });
    }

    /**
     * 示例: 列出归档
     */
    public void listArchives() {
        List<MemoryArchive.ArchiveInfo> archives = memoryManager.listArchives();
        System.out.println("归档列表: " + archives.size() + " 个");

        for (MemoryArchive.ArchiveInfo info : archives) {
            System.out.println("  - " + info.name + " (size: " + info.size +
                    ", time: " + info.timestamp + ")");
        }
    }

    // ===== 管理操作 =====

    /**
     * 示例: 清理过期记忆
     */
    public void cleanupExpired() {
        memoryManager.cleanupExpired();
        System.out.println("已清理过期记忆并归档旧数据");
    }

    /**
     * 示例: 删除特定记忆
     */
    public void forgetSpecific() {
        memoryManager.forget("bubble_tea.drink_name", "某款不喜欢的奶茶");
        System.out.println("已删除特定记忆");
    }

    /**
     * 示例: 清除某key的所有记忆
     */
    public void forgetAllForKey() {
        memoryManager.forgetAll("bubble_tea.temp_key");
        System.out.println("已清除该key的所有记忆");
    }

    /**
     * 示例: 获取统计信息
     */
    public void showStats() {
        MemoryStats stats = memoryManager.getStats();
        System.out.println("记忆统计:");
        System.out.println("  总Keys: " + stats.totalKeys);
        System.out.println("  总Entries: " + stats.totalEntries);
        System.out.println("  总Categories: " + stats.totalCategories);
        System.out.println("  L1缓存大小: " + stats.cacheSize);
        System.out.println("  热点Keys: " + stats.hotKeyCount);
    }

    // ===== 完整场景演示 =====

    /**
     * 场景: 第一次点奶茶
     */
    public void scenarioFirstOrder() {
        System.out.println("\n=== 场景: 第一次点奶茶 ===");

        // 用户首次使用，没有记忆
        String recommended = memoryManager.getRecommendedValue("bubble_tea.drink_name");
        System.out.println("推荐值: " + (recommended != null ? recommended : "无（首次使用）"));

        // 用户选择了芝芝莓莓
        memoryManager.rememberPreference("bubble_tea.drink_name", "芝芝莓莓", "food", null);
        memoryManager.rememberPreference("bubble_tea.sweetness", "三分糖", "food", null);

        System.out.println("已记录用户选择");
    }

    /**
     * 场景: 第二次点奶茶（系统记住上次选择）
     */
    public void scenarioSecondOrder() {
        System.out.println("\n=== 场景: 第二次点奶茶 ===");

        // 系统推荐上次的选择
        SmartDefault defaults = memoryManager.getSmartDefault("bubble_tea.drink_name");
        if (defaults != null) {
            System.out.println("系统推荐: " + defaults.recommendedValue);
            System.out.println("上次选择: " + defaults.lastUsedValue);
        }

        // 用户这次选了不同的
        memoryManager.rememberPreference("bubble_tea.drink_name", "多肉葡萄", "food", null);

        // 查看使用统计
        int count1 = memoryManager.getUsageCount("bubble_tea.drink_name", "芝芝莓莓");
        int count2 = memoryManager.getUsageCount("bubble_tea.drink_name", "多肉葡萄");
        System.out.println("芝芝莓莓 使用次数: " + count1);
        System.out.println("多肉葡萄 使用次数: " + count2);
    }

    /**
     * 场景: 多次使用后（系统越来越懂用户）
     */
    public void scenarioAfterMultipleOrders() {
        System.out.println("\n=== 场景: 多次使用后 ===");

        // 模拟用户点了5次芝芝莓莓，3次多肉葡萄，1次珍珠奶茶
        for (int i = 0; i < 5; i++) {
            memoryManager.rememberPreference("bubble_tea.drink_name", "芝芝莓莓", "food", null);
        }
        for (int i = 0; i < 3; i++) {
            memoryManager.rememberPreference("bubble_tea.drink_name", "多肉葡萄", "food", null);
        }
        memoryManager.rememberPreference("bubble_tea.drink_name", "珍珠奶茶", "food", null);

        // 查看推荐排序
        List<String> recommendations = memoryManager.getRecommendedValues("bubble_tea.drink_name", 5);
        System.out.println("推荐排序（按分数）: " + recommendations);

        // 查看智能默认值
        SmartDefault defaults = memoryManager.getSmartDefault("bubble_tea.drink_name");
        if (defaults != null) {
            System.out.println("最推荐: " + defaults.recommendedValue +
                    "（置信度: " + defaults.confidence + ")");
            System.out.println("最常用: " + defaults.mostUsedValue);
        }
    }

    // ===== 辅助方法 =====

    private Map<String, String> createAttributes(String... keyValues) {
        Map<String, String> attrs = new HashMap<>();
        for (int i = 0; i < keyValues.length - 1; i += 2) {
            attrs.put(keyValues[i], keyValues[i + 1]);
        }
        return attrs;
    }

    /**
     * 运行所有示例
     */
    public void runAllExamples() {
        System.out.println("\n========================================");
        System.out.println("Memory System Sample - 三层记忆系统演示");
        System.out.println("========================================\n");

        // 基础操作
        rememberBubbleTeaPreferences();
        rememberComplexEntry();

        // 查询推荐
        getRecommendations();
        demonstrateAutocomplete();

        // 搜索
        searchMemories();
        getCategoryMemories();

        // 统计
        showStats();

        // 场景演示
        scenarioFirstOrder();
        scenarioSecondOrder();
        scenarioAfterMultipleOrders();

        System.out.println("\n=== 演示完成 ===");
    }
}