# Skill System Guide

## 概述

OFA Android SDK 的技能系统允许用户创建复杂的多步骤自动化任务。每个技能由多个步骤组成，支持条件分支、循环、用户交互等高级功能。

## 核心概念

### SkillDefinition (技能定义)

定义一个完整的技能，包含：
- 基本信息（ID、名称、描述、分类）
- 输入/输出参数
- 步骤列表
- 触发条件
- 元数据

### SkillStep (步骤)

单个执行步骤，支持多种类型：

| 类型 | 描述 | 示例 |
|------|------|------|
| TOOL | 调用工具 | `app.launch`, `camera.capture` |
| INTENT | 执行意图 | 自然语言解析 |
| DELAY | 延迟等待 | 等待APP加载 |
| WAIT_FOR | 等待条件 | 等待订单确认 |
| CONDITION | 条件判断 | 是否下雨？ |
| ASSIGN | 变量赋值 | 保存订单ID |
| INPUT | 请求用户输入 | 选择餐厅 |
| CONFIRM | 请求确认 | 确认支付 |
| NOTIFY | 发送通知 | 订单状态通知 |
| PARALLEL | 并行执行 | 同时执行多个操作 |
| SUB_SKILL | 调用子技能 | 跟踪配送状态 |

### SkillContext (执行上下文)

管理执行状态：
- 变量存储
- 步骤结果
- 执行状态
- 回调通知
- 用户交互处理

### CompositeSkillExecutor (执行器)

执行技能的核心引擎，负责：
- 步骤调度
- 条件求值
- 错误处理和重试
- 回调通知

## 创建技能

### 方式1：代码创建

```java
SkillDefinition skill = new SkillDefinition.Builder()
    .id("custom.my_skill")
    .name("我的技能")
    .description("技能描述")
    .category("custom")
    .tag("标签1")
    .tag("标签2")

    // 添加步骤
    .step(new SkillStep.Builder()
        .id("step1")
        .name("步骤1")
        .type(SkillStep.StepType.TOOL)
        .action("app.launch")
        .param("packageName", "com.example.app")
        .timeout(10000)
        .nextStep("step2")
        .build())

    .step(new SkillStep.Builder()
        .id("step2")
        .name("步骤2")
        .type(SkillStep.StepType.CONFIRM)
        .param("message", "确认继续？")
        .build())

    // 触发条件
    .trigger("voice", "打开我的应用")
    .trigger("schedule", "08:00")

    .estimatedTimeMs(60000)
    .build();

// 注册技能
SkillRegistry.getInstance(context).saveSkill(skill);
```

### 方式2：JSON配置

```json
{
  "id": "custom.my_skill",
  "name": "我的技能",
  "description": "技能描述",
  "category": "custom",
  "tags": ["标签1", "标签2"],
  "triggers": [
    {"type": "voice", "pattern": "打开我的应用"}
  ],
  "steps": [
    {
      "id": "step1",
      "name": "步骤1",
      "type": "TOOL",
      "action": "app.launch",
      "params": {"packageName": "com.example.app"},
      "timeout": 10000,
      "nextStep": "step2"
    },
    {
      "id": "step2",
      "name": "步骤2",
      "type": "CONFIRM",
      "params": {"message": "确认继续？"}
    }
  ]
}
```

## 执行技能

```java
// 获取执行器
CompositeSkillExecutor executor = new CompositeSkillExecutor(context, toolRegistry);

// 准备输入
Map<String, Object> inputs = new HashMap<>();
inputs.put("drinkName", "珍珠奶茶");
inputs.put("sweetness", "五分糖");

// 创建上下文
SkillContext ctx = new SkillContext(skill.getId(), context);
ctx.setCallback(new SkillContext.Callback() {
    @Override
    public void onStepStart(String stepId, SkillStep step) {
        Log.d(TAG, "开始执行: " + step.getName());
    }

    @Override
    public void onComplete(SkillResult result) {
        if (result.isSuccess()) {
            Log.i(TAG, "执行成功");
        } else {
            Log.e(TAG, "执行失败: " + result.getError());
        }
    }
});

// 设置用户交互处理器
ctx.setInteractionHandler(new SkillContext.UserInteractionHandler() {
    @Override
    public void requestInput(String prompt, SkillContext.InputCallback callback) {
        // 显示输入对话框
        String input = showInputDialog(prompt);
        callback.onInput(input);
    }

    @Override
    public void requestConfirm(String message, SkillContext.ConfirmCallback callback) {
        // 显示确认对话框
        boolean confirmed = showConfirmDialog(message);
        callback.onConfirm(confirmed);
    }
});

// 执行
executor.execute(skill, inputs, ctx);
```

## 变量和表达式

### 变量引用

使用 `${variableName}` 语法引用变量：

```java
.param("message", "您的${drinkName}已下单")
.param("packageName", "${app == 'meituan' ? 'com.sankuai.meituan' : 'me.ele'}")
```

### 步骤结果引用

使用 `stepId.output.field` 语法引用步骤输出：

```java
.param("orderId", "${save_order_info.orderId}")
```

### 条件表达式

```java
.branch("${weather == 'rain'}", "remind_umbrella")
.branch("${status != 'delivered'}", "continue_waiting")
```

## 错误处理

### 重试机制

```java
.step(new SkillStep.Builder()
    .id("open_app")
    .type(SkillStep.StepType.TOOL)
    .action("app.launch")
    .retry(3, 2000)  // 最多重试3次，间隔2秒
    .build())
```

### 错误跳转

```java
.step(new SkillStep.Builder()
    .id("search")
    .type(SkillStep.StepType.TOOL)
    .action("ui.search")
    .onError("manual_search")  // 失败时跳转到 manual_search 步骤
    .build())

.step(new SkillStep.Builder()
    .id("manual_search")
    .type(SkillStep.StepType.NOTIFY)
    .param("message", "请手动搜索")
    .build())
```

### 可选步骤

```java
.step(new SkillStep.Builder()
    .id("optional_step")
    .optional(true)  // 失败不影响整体执行
    .build())
```

## 内置技能示例

### 点奶茶技能

```java
// 触发: "点一杯珍珠奶茶"
// 流程:
// 1. 确认订单信息
// 2. 打开美团APP
// 3. 搜索奶茶店
// 4. 选择商品和规格
// 5. 确认地址
// 6. 提交订单
// 7. 支付
// 8. 跟踪配送
```

### 早安问候技能

```java
// 触发: "早安" 或 定时 07:00
// 流程:
// 1. 获取天气
// 2. 播报今日日程
// 3. 播放新闻
```

## 技能注册表

```java
SkillRegistry registry = SkillRegistry.getInstance(context);

// 注册技能
registry.saveSkill(skill);

// 获取技能
SkillDefinition skill = registry.getSkill("skill_id");

// 搜索技能
List<SkillDefinition> results = registry.searchSkills("奶茶");

// 根据语音匹配
SkillDefinition matched = registry.matchTrigger("点一杯奶茶");

// 删除技能
registry.deleteSkill("skill_id");
```

## 最佳实践

1. **步骤粒度**：每个步骤完成一个明确的任务
2. **错误处理**：为可能失败的步骤添加重试和错误跳转
3. **用户反馈**：在关键步骤添加通知和确认
4. **超时设置**：为涉及网络/用户交互的步骤设置合理超时
5. **变量命名**：使用有意义的变量名，便于调试
6. **技能文档**：添加清晰的描述和标签

---

**版本**: v1.0.2
**日期**: 2026-04-02