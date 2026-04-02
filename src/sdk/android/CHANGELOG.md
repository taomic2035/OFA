# OFA Android SDK v1.0.3 - Skill Orchestration System

## 发布日期: 2026-04-02

## 更新概述

本次更新为 OFA Android SDK 添加了完整的技能编排系统，用户可以创建复杂的多步骤自动化任务，如"点奶茶"、"早安问候"等场景。

## 新增功能

### 技能编排系统

#### 核心组件
- ✅ `SkillDefinition` - 技能定义，包含步骤、触发器、输入输出
- ✅ `SkillStep` - 步骤定义，支持12种步骤类型
- ✅ `SkillContext` - 执行上下文，管理状态和变量
- ✅ `SkillResult` - 执行结果
- ✅ `CompositeSkillExecutor` - 复合技能执行器
- ✅ `SkillRegistry` - 技能注册表（支持持久化）

#### 步骤类型

| 类型 | 描述 | 用途 |
|------|------|------|
| TOOL | 工具调用 | 执行MCP工具 |
| INTENT | 意图执行 | 解析自然语言 |
| DELAY | 延迟等待 | 等待APP加载 |
| WAIT_FOR | 等待条件 | 等待订单确认 |
| CONDITION | 条件判断 | 分支逻辑 |
| ASSIGN | 变量赋值 | 保存中间结果 |
| INPUT | 用户输入 | 请求用户选择 |
| CONFIRM | 用户确认 | 确认支付等 |
| NOTIFY | 通知 | 发送状态通知 |
| PARALLEL | 并行执行 | 同时执行多操作 |
| LOOP | 循环 | 重复执行 |
| SUB_SKILL | 子技能 | 调用其他技能 |

#### 特性
- **条件分支**：根据变量值选择不同执行路径
- **错误处理**：重试、错误跳转、可选步骤
- **用户交互**：输入、确认、选择
- **变量引用**：`${var}` 和 `${step.output.field}` 语法
- **触发器**：语音、定时、事件触发
- **持久化**：用户自定义技能自动保存

### 预置技能

#### 外卖订购技能
- `food_order.bubble_tea` - 点奶茶（完整流程：打开APP→搜索→选择→支付→跟踪）
- `food_order.track_delivery` - 配送跟踪（实时状态更新、到货提醒）

### 使用示例

```java
// 创建"点奶茶"技能执行
SkillDefinition skill = skillRegistry.getSkill("food_order.bubble_tea");

Map<String, Object> inputs = new HashMap<>();
inputs.put("drinkName", "珍珠奶茶");
inputs.put("sweetness", "五分糖");
inputs.put("temperature", "少冰");
inputs.put("size", "中杯");

executor.execute(skill, inputs, context);
```

### 项目结构更新

```
sdk/src/main/java/com/ofa/agent/skill/
├── SkillDefinition.java      # 技能定义
├── SkillStep.java            # 步骤定义
├── SkillContext.java         # 执行上下文
├── SkillResult.java          # 执行结果
├── SkillRegistry.java        # 技能注册表
├── CompositeSkillExecutor.java # 执行器
├── SkillExecutor.java        # 简单执行器接口
├── SkillExecutionException.java
└── builtin/
    ├── FoodDeliverySkills.java # 外卖技能
    ├── OfflineSkills.java
    └── ...
```

---

## v1.0.2 - Intent Understanding System

## 发布日期: 2026-04-02

## 更新概述

本次更新为 OFA Android SDK 添加了完整的意图理解系统，能够解析用户自然语言输入，识别意图并执行对应的工具操作。

## 新增功能

### 意图理解系统

#### 意图核心组件
- ✅ `IntentEngine` - 意图识别引擎，支持模式匹配和关键词检测
- ✅ `IntentDefinition` - 意图定义，包含关键词、正则模式和槽位
- ✅ `UserIntent` - 解析后的用户意图，含置信度和槽位值
- ✅ `IntentRegistry` - 预定义意图注册表（30+ 常用意图）
- ✅ `IntentToolMapper` - 意图到工具的映射器
- ✅ `TaskExecutor` - 任务执行器（组合识别+映射+执行）

#### 预定义意图类别
| 类别 | 意图数量 | 示例 |
|------|----------|------|
| system | 4 | 打开设置、调节音量 |
| communication | 3 | 打电话、发短信、发邮件 |
| media | 4 | 拍照、播放音乐、停止播放 |
| device | 5 | WiFi开关、蓝牙开关、亮度调节 |
| navigation | 3 | 导航、搜索位置、获取当前位置 |
| app | 3 | 打开应用、搜索、分享 |

#### 意图识别特性
- 模式匹配：正则表达式匹配用户输入
- 关键词检测：多关键词加权评分
- 槽位提取：从输入中自动提取参数（如联系人名、位置等）
- 置信度计算：返回排序的候选意图列表
- 高置信度自动确认：置信度 >= 0.85 时自动执行

### 任务执行流程

```
用户输入 → IntentEngine → UserIntent → IntentToolMapper → Tool执行
    ↓           ↓              ↓              ↓
 "打开WiFi"  模式匹配     device.wifi_on    wifi.status工具
            关键词检测    confidence=0.9    operation=enable
            槽位提取
```

### 使用示例

```java
// 初始化任务执行器
TaskExecutor executor = new TaskExecutor(context, toolRegistry);

// 处理用户输入
executor.process("打开WiFi", new TaskExecutor.Callback() {
    @Override
    public void onComplete(String taskId, TaskResult result) {
        if (result.isSuccess()) {
            Log.i(TAG, "执行成功: " + result.toolResult.getOutput());
        }
    }
});

// 注册自定义意图
IntentDefinition customIntent = new IntentDefinition.Builder()
    .id("custom.take_selfie")
    .category("media")
    .action("take_selfie")
    .keywords("自拍", "selfie")
    .pattern("自拍|拍.*自己")
    .defaultConfidence(0.9)
    .build();
executor.registerIntent(customIntent);
```

---

## v1.0.1 - Dual LLM Support

### 发布日期: 2026-04-02

### 更新概述

本次更新为 OFA Android SDK 添加了双 LLM 架构支持，实现云端 LLM 和本地 LLM 的无缝切换，支持离线 AI 推理。

## 新增功能

### 双 LLM 架构

#### LLM 提供者接口
- ✅ `LLMProvider` - 统一 LLM 接口
- ✅ `LLMConfig` - 配置构建器
- ✅ `LLMResponse` - 响应结构（含 ToolCall 支持）
- ✅ `Message` - 聊天消息（支持多角色）
- ✅ `StreamCallback` - 流式回调接口
- ✅ `LLMStats` - 统计追踪

#### 云端 LLM (CloudLLMProvider)
- ✅ OpenAI 兼容 HTTP API
- ✅ 支持自定义端点
- ✅ API Key 认证
- ✅ 流式响应
- ✅ Tool Calling 支持

#### 本地 LLM (LocalLLMProvider)
- ✅ TensorFlow Lite 推理引擎
- ✅ GPU 加速支持（可选）
- ✅ 温度采样
- ✅ 流式生成
- ✅ 嵌入向量生成

#### LLM 编排器 (LLMOrchestrator)
- ✅ 多提供者管理
- ✅ 自动故障转移
- ✅ 负载均衡策略
- ✅ 健康检查

### OFAAgent Builder 新增方法

```java
// LLM 配置
Builder llmProvider(LLMProvider provider)
Builder fallbackLLMProvider(LLMProvider provider)
Builder cloudLLM(String endpoint, String apiKey, String model)
Builder localLLM(String modelPath)
Builder autoLLMFailover(boolean enable)
```

### OFAAgent 新增方法

```java
// LLM 访问
boolean hasLLM()
LLMProvider getLLMProvider()
LLMOrchestrator getLLMOrchestrator()
```

## 项目结构更新

```
sdk/src/main/java/com/ofa/agent/
├── intent/                      # 意图理解系统 (NEW)
│   ├── IntentEngine.java        # 意图识别引擎
│   ├── IntentDefinition.java    # 意图定义
│   ├── UserIntent.java          # 解析后的意图
│   ├── IntentRegistry.java      # 预定义意图注册表
│   ├── IntentToolMapper.java    # 意图-工具映射
│   └── TaskExecutor.java        # 任务执行器
├── llm/
│   ├── LLMProvider.java         # LLM 接口
│   ├── LLMConfig.java           # 配置
│   ├── LLMResponse.java         # 响应
│   ├── Message.java             # 消息
│   ├── StreamCallback.java      # 流式回调
│   ├── LLMStats.java            # 统计
│   ├── cloud/
│   │   └── CloudLLMProvider.java
│   ├── local/
│   │   ├── LocalLLMProvider.java
│   │   ├── TFLiteEngine.java
│   │   └── Tokenizer.java
│   ├── orchestrator/
│   │   └── LLMOrchestrator.java
│   └── tool/
│       └── LLMChatTool.java     # MCP 工具包装
├── grpc/
│   ├── AgentGrpc.java           # gRPC Stub
│   └── AgentOuterClass.java     # Protobuf 消息
└── constraint/
    ├── ConstraintType.java      # 约束类型
    ├── ConstraintResult.java    # 检查结果
    ├── ConstraintRule.java      # 约束规则
    └── ConstraintChecker.java   # 检查器
```

## 编译系统更新

### Gradle 配置
- ✅ 多模块结构 (`sdk/` 子模块)
- ✅ Gradle 8.2 + AGP 8.2.0
- ✅ Java 17 兼容性
- ✅ Maven Publishing 配置

### 编译命令
```bash
cd src/sdk/android
export JAVA_HOME="D:/Java/jdk-17"
export ANDROID_HOME="D:/Android/Sdk"
./gradlew.bat assembleRelease --no-daemon
```

### 输出位置
- AAR: `sdk/build/outputs/aar/sdk-release.aar`

## 依赖更新

```groovy
// TensorFlow Lite (本地 LLM)
implementation 'org.tensorflow:tensorflow-lite:2.14.0'
implementation 'org.tensorflow:tensorflow-lite-support:0.4.4'

// OkHttp (云端 LLM HTTP 客户端)
implementation 'com.squareup.okhttp3:okhttp:4.12.0'
```

## Bug 修复

- 修复 `ConstraintChecker` 多类定义问题（拆分为独立文件）
- 修复 `LLMChatTool` 缺少接口方法实现
- 修复 `LLMOrchestrator.embed()` 返回类型不匹配
- 修复 `ToolDefinition` 构造函数兼容性
- 修复多个工具类 JSONException 未捕获问题
- 修复 `CameraTool` 字节数组类型错误
- 修复 `SensorTool` 重复 case 标签
- 修复 `NFCTool` API 兼容性问题

## 兼容性

- Android SDK: 24+ (Android 7.0)
- Target SDK: 34 (Android 14)
- Java: 17
- Gradle: 8.2
- Android Gradle Plugin: 8.2.0

## 下一步计划

- [ ] TensorFlow Lite 模型加载优化
- [ ] 更多本地模型支持 (Gemma, Phi, etc.)
- [ ] 量化模型支持
- [ ] 单元测试完善
- [ ] 性能基准测试

---

**版本**: v1.0.1
**作者**: OFA Team
**日期**: 2026-04-02