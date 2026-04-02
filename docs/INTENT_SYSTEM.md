# Intent System - 意图理解系统

## 概述

意图理解系统让应用能够读懂用户的自然语言指令，自动识别用户意图并提取参数，然后执行对应的操作。

## 核心组件

```
用户输入 → IntentEngine → UserIntent → IntentToolMapper → TaskExecutor → 执行
           (识别意图)    (解析结果)    (映射工具)        (执行任务)
```

### IntentDefinition - 意图定义

```java
IntentDefinition intent = new IntentDefinition.Builder()
    .id("order_food")
    .name("点餐")
    .description("订购外卖或餐厅点餐")
    .keywords("点", "订", "外卖", "餐", "美食")
    .patterns(
        "点(一份|一杯)?{item}",
        "订{item}外卖",
        "我要{item}",
        "帮我点{count}份{item}"
    )
    .slots(
        new SlotDefinition("item", "商品名称", SlotType.TEXT, true),
        new SlotDefinition("count", "数量", SlotType.NUMBER, false, "1"),
        new SlotDefinition("address", "地址", SlotType.LOCATION, false)
    )
    .confidenceThreshold(0.6f)
    .build();
```

### IntentEngine - 意图引擎

提供三种识别方式：

| 方式 | 说明 | 优先级 |
|------|------|--------|
| 模式匹配 | 正则表达式精确匹配 | 高 |
| 关键词检测 | 多关键词组合匹配 | 中 |
| 语义分析 | 基于上下文推断 | 低 |

```java
IntentEngine engine = new IntentEngine();
engine.register(intent);

// 识别意图
List<UserIntent> intents = engine.recognize("帮我点一杯珍珠奶茶");
// 返回: UserIntent(id="order_food", confidence=0.85, slots={item="珍珠奶茶", count="1"})
```

### UserIntent - 解析结果

```java
public class UserIntent {
    String intentId;          // 意图ID
    float confidence;         // 置信度 (0-1)
    Map<String, String> slots; // 提取的参数
    String originalText;      // 原始输入
    String matchedPattern;    // 匹配的模式
}
```

## Slot类型

| 类型 | 说明 | 示例 |
|------|------|------|
| TEXT | 文本 | 商品名称、描述 |
| NUMBER | 数字 | 数量、金额 |
| LOCATION | 位置 | 地址、城市 |
| TIME | 时间 | 日期、时刻 |
| PERSON | 人物 | 姓名、联系人 |
| ENTITY | 实体 | 品牌、类型 |

## 22个内置意图

### 查询类 (8个)
| 意图 | 关键词 | 示例 |
|------|--------|------|
| weather_query | 天气,温度,下雨 | "今天天气怎么样" |
| stock_query | 股票,股价,涨跌 | "苹果股价多少" |
| news_query | 新闻,头条,资讯 | "最近新闻" |
| traffic_query | 路况,堵车,导航 | "去机场的路况" |
| price_query | 价格,多少钱,费用 | "这个多少钱" |
| search_query | 搜索,查找,找 | "搜索附近的餐厅" |
| info_query | 信息,介绍,是什么 | "介绍一下这个景点" |
| location_query | 在哪,位置,地址 | "最近的便利店在哪" |

### 操作类 (8个)
| 意图 | 关键词 | 示例 |
|------|--------|------|
| app_launch | 打开,启动,运行 | "打开微信" |
| app_close | 关闭,退出,结束 | "关闭音乐" |
| call_contact | 打电话,致电,联系 | "给妈妈打电话" |
| send_message | 发消息,短信,发送 | "发给小明说晚安" |
| send_email | 发邮件,邮件 | "给老板发邮件" |
| play_media | 播放,听,看 | "播放周杰伦的歌" |
| take_photo | 拍照,照相,拍摄 | "帮我拍张照" |
| set_timer | 定时,倒计时,计时 | "定时10分钟" |

### 设置类 (4个)
| 意图 | 关键词 | 示例 |
|------|--------|------|
| setting_change | 设置,调整,修改 | "把音量调大" |
| alarm_set | 闹钟,提醒,叫醒 | "明天7点闹钟" |
| reminder_set | 提醒,别忘了,记着 | "提醒我吃药" |
| schedule_add | 日程,安排,计划 | "添加日程开会" |

### 其他类 (2个)
| 意图 | 关键词 | 示例 |
|------|--------|------|
| order_food | 点餐,外卖,订餐 | "点一杯奶茶" |
| control_device | 开关,控制,设备 | "打开空调" |

## 使用示例

### 注册自定义意图

```java
IntentRegistry registry = IntentRegistry.getInstance();
registry.register(new IntentDefinition.Builder()
    .id("bubble_tea_order")
    .name("点奶茶")
    .category("food")
    .keywords("奶茶", "气泡茶", "波霸", "喜茶", "奈雪")
    .patterns(
        "点(一杯)?{drink_name}",
        "我要{drink_name}",
        "来{count}杯{drink_name}",
        "帮我买{drink_name}"
    )
    .slots(
        SlotDefinition.text("drink_name", "饮品名称", true),
        SlotDefinition.number("count", "数量", false, "1"),
        SlotDefinition.text("sweetness", "甜度", false),
        SlotDefinition.text("ice", "冰度", false)
    )
    .build()
);
```

### 意图识别

```java
IntentEngine engine = new IntentEngine(registry);

// 简单输入
UserIntent intent = engine.recognizeOne("点一杯珍珠奶茶");
// → bubble_tea_order, drink_name=珍珠奶茶, count=1

// 复杂输入
UserIntent intent = engine.recognizeOne("帮我买3杯芝芝莓莓三分糖去冰");
// → bubble_tea_order, drink_name=芝芝莓莓, count=3, sweetness=三分糖, ice=去冰

// 多意图输入
List<UserIntent> intents = engine.recognize("打开美团点一杯奶茶");
// → [app_launch(app=美团), bubble_tea_order(drink_name=奶茶)]
```

### 意图到工具映射

```java
IntentToolMapper mapper = new IntentToolMapper();
mapper.map("bubble_tea_order", "food_delivery");

// 执行意图
TaskExecutor executor = new TaskExecutor(toolRegistry);
executor.execute(intent, context);
```

## 文件结构

```
sdk/src/main/java/com/ofa/agent/intent/
├── IntentDefinition.java   # 意图定义
├── IntentEngine.java       # 意识引擎
├── UserIntent.java         # 解析结果
├── IntentRegistry.java     # 意图注册表(22个内置意图)
├── IntentToolMapper.java   # 意图→工具映射
├── TaskExecutor.java       # 任务执行器
└── SlotDefinition.java     # Slot定义(嵌入IntentDefinition)
```

## 置信度计算

```
confidence = pattern_score * 0.5 + keyword_score * 0.3 + context_score * 0.2

pattern_score: 匹配模式的完整度 (0-1)
keyword_score: 关键词覆盖率 (匹配数/总数)
context_score: 上下文相关度 (0-1)
```

## 最佳实践

1. **设计patterns**: 覆盖用户常用表达方式
2. **设置keywords**: 使用领域核心词汇
3. **定义slots**: 明确必需和可选参数
4. **调整threshold**: 平衡准确率和召回率
5. **测试验证**: 使用真实用户输入测试