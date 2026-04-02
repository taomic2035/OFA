# Intent Understanding System

## 概述

OFA Android SDK 意图理解系统使用户能够通过自然语言与 Agent 交互。系统自动解析用户输入，识别意图，提取参数，并执行对应的工具操作。

## 架构

```
┌─────────────────────────────────────────────────────────────┐
│                    Intent Understanding                       │
│                                                              │
│  用户输入                                                     │
│     │                                                        │
│     ▼                                                        │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐      │
│  │ IntentEngine │───▶│ IntentTool  │───▶│    Tool     │      │
│  │             │    │   Mapper    │    │  Executor   │      │
│  └─────────────┘    └─────────────┘    └─────────────┘      │
│         │                  │                 │               │
│         ▼                  ▼                 ▼               │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐      │
│  │ UserIntent  │    │MappingResult│    │ ToolResult  │      │
│  │ - category  │    │ - toolName  │    │ - output    │      │
│  │ - action    │    │ - params    │    │ - error     │      │
│  │ - confidence│    │ - confirm   │    │ - time      │      │
│  │ - slots     │    │ - missing   │    │             │      │
│  └─────────────┘    └─────────────┘    └─────────────┘      │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## 核心组件

### 1. IntentDefinition (意图定义)

定义一个意图的匹配规则和槽位。

```java
IntentDefinition intent = new IntentDefinition.Builder()
    .id("device.wifi_on")                    // 意图ID
    .category("device")                       // 分类
    .action("wifi_on")                        // 动作
    .description("打开WiFi")                  // 描述
    .keywords("wifi", "无线", "网络", "打开")  // 关键词
    .pattern("打开.*wifi|开启.*无线")         // 正则模式
    .defaultConfidence(0.9)                   // 默认置信度
    .build();
```

### 2. IntentEngine (意图引擎)

核心识别引擎，负责解析用户输入。

```java
IntentEngine engine = new IntentEngine();

// 注册意图
engine.register(definition);

// 识别最佳匹配
UserIntent best = engine.recognizeBest("打开WiFi");
// Result: device.wifi_on, confidence=0.9

// 获取所有候选
List<UserIntent> candidates = engine.recognize("打开WiFi");
// Returns sorted list by confidence
```

### 3. UserIntent (用户意图)

表示解析后的意图。

```java
UserIntent intent = engine.recognizeBest("打电话给张三");

intent.getId();           // "communication.call"
intent.getCategory();     // "communication"
intent.getAction();       // "call"
intent.getConfidence();   // 0.85
intent.getSlots();        // {contact: "张三"}
intent.isHighConfidence(); // true (>= 0.7)
```

### 4. IntentToolMapper (映射器)

将意图映射到工具调用。

```java
IntentToolMapper mapper = new IntentToolMapper();

// 意图自动映射到工具
MappingResult mapping = mapper.map(intent);
// toolName: "phone.call"
// params: {contactName: "张三"}
// requiresConfirmation: true
```

### 5. TaskExecutor (任务执行器)

组合所有组件，提供完整的处理流程。

```java
TaskExecutor executor = new TaskExecutor(context, toolRegistry);

executor.process("打开WiFi", new TaskExecutor.Callback() {
    @Override
    public void onComplete(String taskId, TaskResult result) {
        // 处理结果
    }
});
```

## 预定义意图

### 系统控制 (system)

| 意图ID | 描述 | 示例输入 |
|--------|------|----------|
| `system.open_settings` | 打开设置 | "打开设置", "进入设置" |
| `system.close_app` | 关闭应用 | "关闭微信", "退出应用" |
| `system.volume` | 调节音量 | "把音量调大", "静音" |
| `system.restart` | 重启 | "重启设备" |

### 通讯 (communication)

| 意图ID | 描述 | 示例输入 |
|--------|------|----------|
| `communication.call` | 打电话 | "打电话给张三", "拨打110" |
| `communication.sms` | 发短信 | "发短信给李四", "给王五发消息" |
| `communication.email` | 发邮件 | "发邮件给test@example.com" |

### 媒体 (media)

| 意图ID | 描述 | 示例输入 |
|--------|------|----------|
| `media.capture` | 拍照/录像 | "帮我拍照", "打开相机录像" |
| `media.play_music` | 播放音乐 | "播放周杰伦的歌", "听音乐" |
| `media.stop` | 停止播放 | "停止播放", "暂停音乐" |
| `media.view_image` | 查看图片 | "查看相册", "看照片" |

### 设备控制 (device)

| 意图ID | 描述 | 示例输入 |
|--------|------|----------|
| `device.wifi_on` | 打开WiFi | "打开WiFi", "开启无线网络" |
| `device.wifi_off` | 关闭WiFi | "关闭WiFi" |
| `device.bluetooth_on` | 打开蓝牙 | "打开蓝牙" |
| `device.brightness` | 调节亮度 | "把屏幕调亮", "亮度调到50" |
| `device.battery` | 检查电池 | "电池还剩多少", "电量查询" |

### 导航 (navigation)

| 意图ID | 描述 | 示例输入 |
|--------|------|----------|
| `navigation.navigate` | 导航 | "导航到北京天安门", "带我去机场" |
| `navigation.search_location` | 搜索位置 | "北京在哪", "查找附近的餐厅" |
| `navigation.current_location` | 当前位置 | "我在哪", "我的位置" |

### 应用 (app)

| 意图ID | 描述 | 示例输入 |
|--------|------|----------|
| `app.open` | 打开应用 | "打开微信", "运行支付宝" |
| `app.search` | 搜索 | "搜索天气", "在淘宝搜手机" |
| `app.share` | 分享 | "分享这张照片" |

## 槽位提取

意图可以定义槽位，从用户输入中提取参数。

```java
// 定义带槽位的意图
IntentDefinition callIntent = new IntentDefinition.Builder()
    .id("communication.call")
    .category("communication")
    .action("call")
    .keywords("电话", "拨打", "call")
    .pattern("打电话给.*|拨打.*")
    // 提取联系人名
    .slotWithPattern("contact", "string", "(给|打给)\\s*([\\w\\s]+)", true)
    // 提取电话号码
    .slotWithPattern("phone_number", "string", "(\\d{11})", false)
    .build();

// 使用
UserIntent intent = engine.recognizeBest("打电话给张三");
intent.getSlot("contact");  // "张三"
```

## 意图-工具映射

将意图映射到具体的工具执行。

```java
// 添加映射规则
mapper.addMapping(
    "communication.call",     // 意图ID
    "phone.call",             // 工具名
    Map.of("contact", "contactName", "phone_number", "phoneNumber"), // 槽位映射
    null,                     // 固定参数
    true,                     // 需要确认
    "确认拨打 {contact} 的电话?"  // 确认消息模板
);
```

## 确认流程

某些意图需要用户确认才能执行。

```java
executor.process("打电话给张三", new TaskExecutor.Callback() {
    @Override
    public void onConfirmationRequired(String taskId, UserIntent intent, String message) {
        // 显示确认对话框
        showConfirmDialog(message, (confirmed) -> {
            if (confirmed) {
                // 用户确认后执行
                executor.confirmAndExecute(taskId, intent, mapping, null);
            }
        });
    }
});
```

## 缺失槽位处理

当必需槽位缺失时，提示用户提供信息。

```java
@Override
public void onSlotMissing(String taskId, List<String> missingSlots) {
    // 提示用户补充信息
    String prompt = "请提供: " + String.join(", ", missingSlots);
    askUser(prompt, (input) -> {
        // 补充槽位并执行
        Map<String, Object> additional = new HashMap<>();
        additional.put(missingSlots.get(0), input);
        executor.fillSlotsAndExecute(taskId, intent, mapping, additional, null);
    });
}
```

## 自定义意图

### 注册自定义意图

```java
IntentDefinition myIntent = new IntentDefinition.Builder()
    .id("custom.my_action")
    .category("custom")
    .action("my_action")
    .description("我的自定义意图")
    .keywords("关键词1", "关键词2")
    .pattern("匹配模式.*")
    .slot("param1", "string", "参数描述", true)
    .defaultConfidence(0.8)
    .build();

// 注册到执行器
executor.registerIntent(myIntent);

// 映射到工具
executor.registerMapping(
    "custom.my_action",
    "my.custom.tool",
    Map.of("param1", "toolParam"),
    null,
    false,
    null
);
```

### 动态意图注册

```java
// 运行时添加意图
IntentDefinition dynamicIntent = new IntentDefinition.Builder()
    .id("dynamic.temp")
    .category("temp")
    .action("temp_action")
    .keywords("临时", "动态")
    .build();

executor.getIntentEngine().register(dynamicIntent);
```

## 置信度机制

### 置信度计算

1. **模式匹配**: 匹配正则表达式时，置信度 = `defaultConfidence + 0.2`
2. **关键词匹配**: 按匹配关键词比例计算，最高 +0.3 加成
3. **阈值检查**: 低于 0.3 的结果不返回

### 置信度等级

| 等级 | 范围 | 行为 |
|------|------|------|
| 高 | >= 0.7 | 自动执行（可配置） |
| 中 | 0.4 - 0.7 | 需要确认 |
| 低 | < 0.4 | 可能不匹配 |

### 配置自动确认

```java
// 开启高置信度自动确认
executor.setAutoConfirmHighConfidence(true);

// 设置阈值（默认 0.85）
executor.setHighConfidenceThreshold(0.9);
```

## 完整示例

```java
// 1. 创建工具注册表
ToolRegistry registry = new ToolRegistry(context);
BuiltInTools.registerAll(context, registry);

// 2. 创建任务执行器
TaskExecutor executor = new TaskExecutor(context, registry);

// 3. 配置
executor.setAutoConfirmHighConfidence(true);
executor.setHighConfidenceThreshold(0.85);

// 4. 处理用户输入
public void handleUserInput(String input) {
    executor.process(input, new TaskExecutor.Callback() {
        @Override
        public void onStatusUpdate(String taskId, ExecutionStatus status, String message) {
            Log.d(TAG, "状态更新: " + status + " - " + message);
        }

        @Override
        public void onConfirmationRequired(String taskId, UserIntent intent, String message) {
            showConfirmation(taskId, intent, message);
        }

        @Override
        public void onSlotMissing(String taskId, List<String> missingSlots) {
            requestMissingInfo(taskId, missingSlots);
        }

        @Override
        public void onComplete(String taskId, TaskResult result) {
            if (result.isSuccess()) {
                handleSuccess(result.toolResult.getOutput());
            } else {
                handleError(result.message);
            }
        }
    });
}

// 5. 资源清理
@Override
protected void onDestroy() {
    executor.shutdown();
    super.onDestroy();
}
```

## 性能优化

### 意图数量优化

- 默认注册 22 个常用意图
- 按需注册自定义意图
- 及时注销不需要的意图

### 匹配优化

- 按分类索引意图
- 短路返回高置信度结果
- 结果数量限制（默认最多 5 个）

---

**版本**: v1.0.2
**日期**: 2026-04-02