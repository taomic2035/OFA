# Memory System - 用户记忆系统

## 概述

用户记忆系统用于存储和管理用户偏好、习惯和历史行为，让系统能够"记住"用户的选择，随着使用越来越懂用户。

## 三层架构

系统采用三层存储架构，兼顾性能和可靠性：

```
┌─────────────────────────────────────────────────────┐
│                     L1 Cache                         │
│  MemoryCache - 内存缓存 (LRU策略, 快速访问)           │
│  容量: 100条热数据                                   │
│  特点: 毫秒级访问, 自动淘汰冷数据                      │
└─────────────────────────────────────────────────────┘
                         ↓ 未命中
┌─────────────────────────────────────────────────────┐
│                    L2 Database                       │
│  Room Database - 持久化存储                          │
│  特点: 可靠存储, 支持复杂查询, 离线可用                │
│  索引: key, category, timestamp, score              │
└─────────────────────────────────────────────────────┘
                         ↓ 归档
┌─────────────────────────────────────────────────────┐
│                    L3 Archive                        │
│  MemoryArchive - 文件归档                            │
│  特点: 冷数据备份, 支持导入导出                       │
│  格式: JSON文件                                      │
└─────────────────────────────────────────────────────┘
```

### 数据流向

**读取流程:**
```
查询 → L1缓存 → (命中) → 返回
              → (未命中) → L2数据库 → 回填L1 → 返回
```

**写入流程:**
```
写入 → L1缓存 → L2数据库(异步) → 完成
```

**归档流程:**
```
定期清理 → 查询60天以上数据 → 归档到L3 → 删除L2数据
```

## 核心组件

### MemoryEntry - 记忆条目

```java
MemoryEntry entry = new MemoryEntry.Builder()
    .key("bubble_tea.drink_name")    // 记忆键
    .category("food")                 // 分类
    .value("珍珠奶茶")                // 值
    .attribute("sweetness", "五分糖") // 附加属性
    .attribute("ice", "少冰")
    .score(2.5f)                      // 重要性评分
    .context("点奶茶技能")            // 上下文
    .build();
```

### MemoryEntity - Room实体

```java
@Entity(tableName = "memories")
public class MemoryEntity {
    long id;           // 主键
    String key;        // 记忆键
    String category;   // 分类
    String value;      // 值
    String attributes; // JSON属性
    long timestamp;    // 时间戳
    int count;         // 使用次数
    float score;       // 重要性评分
    String context;    // 上下文
    long lastAccessed; // 最后访问时间
}
```

### MemoryDao - 数据访问对象

提供丰富的查询方法：

```java
@Query("SELECT * FROM memories WHERE key = :key ORDER BY score DESC LIMIT 1")
MemoryEntity getTopRecommendation(String key);

@Query("SELECT * FROM memories WHERE key = :key ORDER BY timestamp DESC LIMIT 1")
MemoryEntity getLastUsed(String key);

@Query("SELECT * FROM memories WHERE key = :key ORDER BY count DESC LIMIT 1")
MemoryEntity getMostUsed(String key);

@Query("SELECT * FROM memories WHERE key = :key AND value LIKE :prefix || '%'")
List<MemoryEntity> autocomplete(String key, String prefix, int limit);
```

## 使用示例

### 记录偏好

```java
UserMemoryManager memory = UserMemoryManager.getInstance(context);

// 记录奶茶偏好
memory.rememberPreference(
    "bubble_tea.drink_name",
    "芝芝莓莓",
    "food",
    Map.of("sweetness", "三分糖", "ice", "去冰")
);

// 记录技能参数
memory.rememberSkillParams("order_bubble_tea", params);
```

### 获取推荐

```java
// 获取推荐值（分数最高的）
String recommended = memory.getRecommendedValue("bubble_tea.drink_name");

// 获取推荐列表
List<String> top3 = memory.getRecommendedValues("bubble_tea.drink_name", 3);

// 获取智能默认值
SmartDefault defaults = memory.getSmartDefault("bubble_tea.drink_name");
// defaults.recommendedValue - 推荐值
// defaults.lastUsedValue - 最近使用
// defaults.mostUsedValue - 最常用
// defaults.confidence - 置信度
```

### 自动补全

```java
// 用户输入"珍"，系统补全
List<String> suggestions = memory.autocomplete("bubble_tea.drink_name", "珍", 5);
// 返回: ["珍珠奶茶", "芝芝莓莓"...]
```

### 导入导出

```java
// 导出记忆
memory.exportMemories(new MemoryArchive.ExportCallback() {
    void onSuccess(File file) { ... }
    void onError(String error) { ... }
});

// 导入记忆
memory.importMemories(importFile, new MemoryArchive.ImportCallback() {
    void onSuccess(List<MemoryEntry> entries) { ... }
    void onError(String error) { ... }
});
```

## 推荐算法

系统综合考虑多个因素计算推荐分数：

```java
float calculateRecommendationScore() {
    // 使用次数权重 (最大1.0)
    float countWeight = Math.min(count * 0.1f, 1.0f);

    // 时间衰减
    float timeDecay = exp(-ageHours * 0.01);

    // 最近访问加成
    float accessDecay = exp(-accessAgeHours * 0.02);

    // 综合分数
    return score * 0.3f + countWeight * 0.4f + timeDecay * 0.2f + accessDecay * 0.1f;
}
```

**分数组成:**
- `score * 0.3` - 基础评分权重
- `countWeight * 0.4` - 使用频率权重（点5次以上得满分）
- `timeDecay * 0.2` - 时间衰减（最近使用的权重更高）
- `accessDecay * 0.1` - 最近访问加成

## 智能默认值策略

当用户未指定参数时，系统按以下优先级提供默认值：

1. **推荐值** - 综合分数最高的选项
2. **最近使用** - 上次使用的值
3. **最常用** - 使用次数最多的值

示例场景：

```
用户点奶茶历史:
- 芝芝莓莓: 5次 (最近1小时前)
- 多肉葡萄: 3次 (最近1天前)
- 珍珠奶茶: 1次 (最近1周前)

推荐结果:
- 推荐值: 芝芝莓莓 (分数最高)
- 最近使用: 芝芝莓莓
- 最常用: 芝芝莓莓
- 置信度: 0.85 (高度推荐)
```

## 与技能系统集成

`MemoryAwareSkillExecutor` 将记忆系统与技能执行结合：

```java
MemoryAwareSkillExecutor executor = new MemoryAwareSkillExecutor(context, toolRegistry);

// 执行技能时自动应用记忆
CompletableFuture<SkillResult> result = executor.execute(skill, inputs);

// 获取预填充建议
Map<String, Object> suggestions = executor.getSuggestedInputs("order_bubble_tea",
    List.of("drink_name", "sweetness", "ice"));

// 根据偏好排序选项
List<String> sorted = executor.sortOptionsByPreference("bubble_tea.drink_name", options);
```

**自动填充流程:**

```
技能执行开始
  ↓
检查参数是否提供
  ↓ (未提供)
从记忆中查找推荐值
  ↓
填充参数 (标记为_memory_xxx)
  ↓
执行技能
  ↓ (成功)
记录参数作为偏好
```

## 文件结构

```
sdk/src/main/java/com/ofa/agent/memory/
├── MemoryEntry.java          # 记忆条目定义
├── MemoryEntity.java         # Room数据库实体
├── MemoryDao.java            # Room数据访问对象
├── MemoryDatabase.java       # Room数据库
├── MemoryCache.java          # L1内存缓存
├── MemoryArchive.java        # L3文件归档
├── UserMemoryManager.java    # 主管理器(三层集成)
├── ContextMemory.java        # 会话级记忆
└── MemoryAwareSkillExecutor.java # 记忆感知执行器
```

## 性能特性

| 层级 | 存储位置 | 访问延迟 | 容量 | 用途 |
|------|---------|---------|------|------|
| L1 | 内存 | < 1ms | 100条 | 热数据缓存 |
| L2 | Room DB | < 10ms | 无限 | 持久化存储 |
| L3 | JSON文件 | < 100ms | 无限 | 冷数据归档 |

## 最佳实践

1. **合理命名key**: 使用点分隔命名，如 `skill_name.param_name`
2. **设置category**: 便于按类别查询和管理
3. **添加attributes**: 存储相关属性，如甜度、冰度等
4. **定期清理**: 调用 `cleanupExpired()` 归档旧数据
5. **导入导出**: 定期备份用户记忆