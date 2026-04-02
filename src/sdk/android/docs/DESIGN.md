# OFA Android SDK 方案设计文档

## 1. 运行模式方案

### 1.1 STANDALONE 模式

**设计目标**: 完全离线运行，无任何网络依赖

**适用场景**:
- 隐私敏感应用
- 无网络环境
- 数据本地化要求

**能力范围**:
```
✅ 意图理解 (本地模式匹配)
✅ 技能执行
✅ UI 自动化 (AccessibilityService)
✅ 社交通知 (通过 UI 自动化)
✅ 记忆系统
✅ 本地 LLM (如已配置)
❌ 云端 LLM
❌ Center 任务
❌ Peer 通信
```

**初始化流程**:
```
1. 创建 AgentProfile (allowRemoteControl=false)
2. 初始化 LocalExecutionEngine
3. 不初始化 CenterConnection
4. 不初始化 PeerNetwork
5. 设置状态为 ONLINE
```

### 1.2 CONNECTED 模式

**设计目标**: 始终连接 Center，接收远程管理

**适用场景**:
- 企业设备管理
- 集群协作
- 远程监控

**能力范围**:
```
✅ 所有 STANDALONE 能力
✅ 云端 LLM
✅ Center 任务接收和执行
✅ Peer 通信
✅ 状态同步
✅ 远程配置更新
```

**连接流程**:
```
1. 创建 gRPC Channel
2. 发送注册请求 (AgentProfile)
3. 建立双向流
4. 启动心跳 (30秒间隔)
5. 监听任务分配
6. 处理配置更新
```

**重连机制**:
```
连接断开
    ↓
等待 5 秒
    ↓
尝试重连
    ↓
失败? → 指数退避 (5s → 10s → 30s → 60s)
成功? → 重置计时器，恢复正常
```

### 1.3 HYBRID 模式 (推荐)

**设计目标**: 本地优先，云端增强

**适用场景**:
- 消费者应用
- 混合网络环境
- 最佳用户体验

**智能路由决策**:

```java
CompletableFuture<TaskResult> executeHybrid(TaskRequest request) {
    // 决策因素
    boolean canBeOffline = localEngine.canExecute(request);
    boolean needsCloud = request.requiresCloudCapability();
    boolean centerAvailable = centerConnection.isConnected();
    boolean networkAvailable = isNetworkAvailable();

    // 决策树
    if (!networkAvailable && canBeOffline) {
        // 无网络，执行本地
        return executeLocally(request);
    }

    if (needsCloud && centerAvailable) {
        // 需要云端能力，且 Center 可用
        return executeViaCenter(request);
    }

    if (canBeOffline) {
        // 本地可执行，优先本地
        return executeLocally(request);
    }

    if (centerAvailable) {
        // 尝试 Center
        return executeViaCenter(request);
    }

    // 无法执行
    return error("No execution path available");
}
```

**降级策略**:

| 能力 | 首选 | 降级方案 |
|------|------|---------|
| LLM | 云端 | 本地 TFLite |
| 意图理解 | 云端 NLP | 本地模式匹配 |
| 社交通知 | 微信 (需登录) | 短信 |
| 技能执行 | Center 协调 | 本地执行 |

---

## 2. 社交通知方案

### 2.1 消息分类方案

**分类方法**: 关键词 + 模式匹配

**消息类型定义**:

```java
// 邀请类 - 社交邀请
TYPE_INVITATION = "invitation"
关键词: 约, 一起, 吃饭, 聚会, 活动, 看电影, 逛街, 旅游, 有空吗

// 紧急类 - 需要即时响应
TYPE_URGENT = "urgent"
关键词: 紧急, 急, 马上, 立即, 重要, 危险, 事故, 救命

// 提醒类 - 定时提醒
TYPE_REMINDER = "reminder"
关键词: 提醒, 记得, 别忘了, 明天, 会议, 预约, 到期

// 攻略类 - 内容分享
TYPE_GUIDE = "guide"
关键词: 攻略, 教程, 方法, 技巧, 推荐, 分享, 怎么, 如何

// 支付类 - 金融相关
TYPE_PAYMENT = "payment"
关键词: 转账, 付款, 支付, 还钱, 红包, 金额, 钱

// 日常类 - 轻松聊天
TYPE_CASUAL = "casual"
关键词: 哈, 无聊, 随便, 搞笑, 晚安, 早安

// 工作类 - 职场沟通
TYPE_BUSINESS = "business"
关键词: 工作, 任务, 项目, 会议, 审批, 领导
```

**分类算法**:

```java
ClassificationResult classify(String message) {
    Map<String, Double> scores = new HashMap<>();

    // 1. 计算每种类型的得分
    for (String type : typePatterns.keySet()) {
        double score = 0;
        for (String keyword : typePatterns.get(type)) {
            if (message.contains(keyword)) {
                score += 1;
            }
        }
        scores.put(type, score);
    }

    // 2. 选择最高分类型
    String bestType = argmax(scores);

    // 3. 计算紧急度
    int urgency = calculateUrgency(message);

    // 4. 推荐渠道
    String channel = recommendChannel(bestType, urgency);

    return new ClassificationResult(bestType, urgency, channel);
}
```

### 2.2 渠道选择方案

**渠道优先级矩阵**:

| 消息类型 | 第1选择 | 第2选择 | 第3选择 |
|----------|---------|---------|---------|
| invitation | 微信 | 电话 | 短信 |
| urgent | 电话 | 微信 | 短信 |
| reminder | 短信 | 微信 | - |
| guide | 小红书 | 微信 | 抖音 |
| payment | 支付宝 | 微信 | - |
| casual | 微信 | 抖音 | QQ |
| business | 钉钉 | 企业微信 | 微信 |

**渠道能力检测**:

```java
boolean isChannelAvailable(String channel) {
    switch (channel) {
        case "wechat":
            return isAppInstalled("com.tencent.mm");
        case "alipay":
            return isAppInstalled("com.eg.android.AlipayGphone");
        case "phone":
        case "sms":
            return true; // 系统内置
        case "xiaohongshu":
            return isAppInstalled("com.xingin.xhs");
        // ...
    }
}
```

### 2.3 消息发送方案

**发送流程**:

```
1. 渠道选择
     ↓
2. 检查渠道可用性
     ↓
3. 检查联系人信息 (是否有该渠道账号)
     ↓
4. 执行发送
     ↓
5. 失败? → 尝试下一个渠道
     ↓
6. 记录发送结果
```

**微信发送实现** (通过 AccessibilityService):

```java
SendResult sendViaWeChat(String recipient, String message) {
    // 1. 启动微信
    launchApp("com.tencent.mm");
    waitFor(1500);

    // 2. 搜索联系人
    click(BySelector.text("搜索"));
    waitFor(500);
    inputText(recipient);
    waitFor(500);
    click(BySelector.textContains(recipient));

    // 3. 输入消息
    waitFor(500);
    inputText(message);

    // 4. 发送
    click(BySelector.text("发送"));

    return new SendResult(true, "wechat", recipient);
}
```

**降级策略**:

```
微信发送失败
    ↓
原因判断:
- 微信未安装 → 尝试短信
- 联系人无微信 → 尝试电话
- 网络问题 → 等待重试
- 权限问题 → 引导用户授权
```

---

## 3. 自动化方案

### 3.1 UI 自动化方案

**技术选型**: AccessibilityService

**优势**:
- 普通应用可用
- 用户授权即可
- 跨应用操作

**劣势**:
- 部分厂商限制
- 需要用户手动开启
- 性能受限

**服务配置**:

```xml
<accessibility-service
    android:accessibilityEventTypes="typeAllMask"
    android:accessibilityFeedbackType="feedbackGeneric"
    android:canPerformGestures="true"
    android:canRetrieveWindowContent="true"
    android:description="@string/accessibility_description"
    android:settingsActivity=".SettingsActivity" />
```

### 3.2 App 适配方案

**适配器接口设计**:

```java
public interface AppAdapter {
    // 基本信息
    String getPackageName();
    String getAppName();

    // 检测
    boolean isOnApp();
    String detectCurrentPage();

    // 操作
    boolean search(AutomationEngine engine, String query);
    boolean selectShop(AutomationEngine engine, String shopName);
    boolean selectProduct(AutomationEngine engine, String productName);
    boolean addToCart(AutomationEngine engine, int quantity);
    boolean goToCheckout(AutomationEngine engine);
    boolean submitOrder(AutomationEngine engine);

    // 状态
    OrderStatus getOrderStatus();
}
```

**美团适配器实现**:

```java
public class MeituanAdapter extends BaseAppAdapter {
    @Override
    public boolean search(AutomationEngine engine, String query) {
        // 1. 找到搜索框
        AutomationNode searchBox = engine.findElement(
            BySelector.className("EditText")
                .descContains("搜索"));

        if (searchBox == null) {
            // 尝试点击搜索图标
            engine.click(BySelector.desc("搜索"));
            waitFor(500);
            searchBox = engine.findElement(BySelector.className("EditText"));
        }

        // 2. 输入搜索词
        engine.inputText(searchBox, query);

        // 3. 等待结果
        return waitForElement(BySelector.textContains(query), 3000);
    }
}
```

### 3.3 错误恢复方案

**恢复策略**:

| 策略 | 触发条件 | 恢复动作 |
|------|----------|---------|
| ScrollToFind | 元素未找到 | 滚动页面继续查找 |
| WaitAndRetry | 超时 | 等待后重试 |
| DismissOverlay | 弹窗阻挡 | 按返回键关闭 |
| WaitForPage | 页面加载中 | 等待页面稳定 |
| HandlePermission | 权限被拒 | 引导用户授权 |
| HandleNetwork | 网络错误 | 等待网络恢复 |

**恢复流程**:

```java
AutomationResult executeWithRecovery(AutomationOperation operation) {
    int attempts = 0;
    AutomationResult result;

    do {
        result = operation.execute();

        if (result.isSuccess()) {
            return result;
        }

        // 分析错误类型
        ErrorType errorType = analyzeError(result);

        // 选择恢复策略
        RecoveryStrategy strategy = selectStrategy(errorType);

        // 执行恢复
        if (strategy != null) {
            strategy.recover();
        }

        attempts++;
    } while (attempts < maxRetries);

    return result;
}
```

---

## 4. 记忆系统方案

### 4.1 三层存储架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        L1: MemoryCache                          │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  内存 LRU 缓存                                           │   │
│  │  - 容量: 100 条                                          │   │
│  │  - 访问: 毫秒级                                          │   │
│  │  - 生命周期: 应用运行期间                                │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              ↓ 写入
                              ↑ 读取 (缓存未命中时)
┌─────────────────────────────────────────────────────────────────┐
│                        L2: MemoryDatabase                       │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  Room 数据库                                             │   │
│  │  - 持久化存储                                            │   │
│  │  - 访问: 毫秒级                                          │   │
│  │  - 生命周期: 永久                                        │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              ↓ 归档
                              ↑ 恢复
┌─────────────────────────────────────────────────────────────────┐
│                        L3: MemoryArchive                        │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  文件存储 (JSON)                                         │   │
│  │  - 冷数据备份                                            │   │
│  │  - 导入/导出                                             │   │
│  │  - 跨设备迁移                                            │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### 4.2 推荐算法

**推荐分数计算**:

```java
double calculateScore(MemoryEntry entry) {
    // 频率因子 (0-1)
    double frequencyScore = Math.min(1.0, entry.accessCount / 100.0);

    // 最近使用因子 (时间衰减)
    long daysSinceAccess = (now - entry.lastAccessTime) / DAY_MS;
    double recencyScore = Math.exp(-daysSinceAccess / 30.0); // 30天半衰期

    // 综合分数
    return frequencyScore * 0.4 + recencyScore * 0.6;
}
```

**智能补全**:

```java
List<MemorySuggestion> getSuggestions(String keyPrefix, int limit) {
    // 1. 从缓存获取
    List<MemorySuggestion> results = cache.getSuggestions(keyPrefix);

    // 2. 缓存不足，查询数据库
    if (results.size() < limit) {
        results.addAll(database.queryByPrefix(keyPrefix, limit - results.size()));
    }

    // 3. 按分数排序
    results.sort((a, b) -> Double.compare(b.score, a.score));

    // 4. 返回 Top N
    return results.subList(0, Math.min(limit, results.size()));
}
```

---

## 5. AI 能力方案

### 5.1 双 LLM 架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        LLMOrchestrator                          │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    请求路由                              │   │
│  │  1. 检查网络状态                                         │   │
│  │  2. 检查云端可用性                                       │   │
│  │  3. 选择最佳 Provider                                    │   │
│  │  4. 自动故障转移                                         │   │
│  └─────────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────┤
│  ┌───────────────────────┐    ┌───────────────────────┐       │
│  │   CloudLLMProvider    │    │   LocalLLMProvider    │       │
│  │   (OpenAI/Claude)     │    │   (TFLite)            │       │
│  │                       │    │                       │       │
│  │   优点:               │    │   优点:               │       │
│  │   - 能力强大          │    │   - 离线可用          │       │
│  │   - 知识丰富          │    │   - 隐私保护          │       │
│  │                       │    │   - 低延迟            │       │
│  │   缺点:               │    │                       │       │
│  │   - 需要网络          │    │   缺点:               │       │
│  │   - 有成本            │    │   - 能力有限          │       │
│  │   - 延迟较高          │    │   - 模型较大          │       │
│  └───────────────────────┘    └───────────────────────┘       │
└─────────────────────────────────────────────────────────────────┘
```

### 5.2 多臂老虎机 (MAB) 方案

**应用场景**: 选择最优选项 (店铺、支付方式、重试策略)

**算法选择**:

| 算法 | 特点 | 适用场景 |
|------|------|---------|
| Epsilon-Greedy | 简单，以 ε 概率探索 | 快速收敛 |
| UCB | 置信区间上界，平衡探索利用 | 稳定选择 |
| Thompson Sampling | 贝叶斯采样，概率匹配 | 不确定环境 |

**店铺选择示例**:

```java
// 1. 初始化
SmartDecisionEngine engine = new SmartDecisionEngine(context, memoryManager);
engine.registerOptions("shop_selection", List.of("喜茶", "奈雪", "蜜雪冰城"));

// 2. 选择
String selected = engine.selectOption("shop_selection");

// 3. 执行并反馈
boolean success = executeOrder(selected);
double reward = success ? 1.0 : 0.0;
engine.reportReward("shop_selection", selected, reward);

// 4. 学习
// 随着使用，系统会学习用户的偏好
// 偏好喜茶 → 喜茶被选中的概率会逐渐提高
```

---

## 6. Peer 网络方案

### 6.1 服务发现

**NSD (Network Service Discovery)**:

```
Agent 注册服务:
1. 监听 TCP 端口 (自动分配)
2. 创建 NsdServiceInfo
   - serviceName: "OFA_<AgentName>"
   - serviceType: "_ofa_agent._tcp"
   - port: <localPort>
3. 调用 nsdManager.registerService()

Agent 发现服务:
1. 调用 nsdManager.discoverServices()
2. 收到 onServiceFound 回调
3. 调用 nsdManager.resolveService() 获取详细信息
4. 存储到 discoveredPeers
```

### 6.2 P2P 通信

**协议设计**:

```json
// 心跳
{"type": "ping"}
{"type": "pong"}

// 任务请求
{
    "type": "task_request",
    "taskId": "task-123",
    "taskType": "skill",
    "params": {...}
}

// 任务响应
{
    "type": "task_response",
    "taskId": "task-123",
    "success": true,
    "result": {...}
}

// 消息
{
    "type": "message",
    "from": "agent-456",
    "content": "Hello!"
}
```

### 6.3 任务委托

**委托流程**:

```
Agent A (发起方)
     ↓
1. 发现 Agent B 有某项能力
     ↓
2. 发送任务请求
     ↓
3. Agent B 执行任务
     ↓
4. Agent B 返回结果
     ↓
5. Agent A 处理结果
```

**能力匹配**:

```java
PeerInfo findPeerWithCapability(String capabilityId) {
    for (PeerInfo peer : discoveredPeers.values()) {
        if (peer.capabilities.contains(capabilityId)) {
            return peer;
        }
    }
    return null;
}
```

---

## 7. 离线优先方案

### 7.1 离线能力分级

| 级别 | 能力 | 说明 |
|------|------|------|
| L4 | 完全离线 | 所有核心功能可用 |
| L3 | 基本离线 | 大部分功能可用，云端增强不可用 |
| L2 | 部分离线 | 仅本地缓存可用 |
| L1 | 依赖网络 | 大部分功能需要网络 |

### 7.2 数据缓存策略

```java
class OfflineCache {
    // 缓存策略
    void put(String key, Object value, long ttlMs);

    // 缓存获取
    Object get(String key);

    // 缓存失效
    void invalidate(String key);

    // 同步策略
    void syncWithServer();
}
```

### 7.3 任务队列

```java
class LocalTaskQueue {
    // 添加任务
    void enqueue(TaskRequest request);

    // 执行任务 (在线时同步)
    void processQueue();

    // 重试失败任务
    void retryFailed();
}
```

---

## 8. 扩展开发指南

### 8.1 添加新的消息类型

```java
// 1. 定义关键词
typePatterns.put("new_type", List.of("关键词1", "关键词2"));

// 2. 定义渠道映射
private String recommendChannel(String type, int urgency) {
    switch (type) {
        case "new_type":
            return "wechat"; // 或其他渠道
    }
}
```

### 8.2 添加新的社交渠道

```java
// 1. 在 ChannelSelector 中添加常量
public static final String CHANNEL_NEW = "new_channel";
public static final Map<String, String> CHANNEL_PACKAGES = Map.of(
    CHANNEL_NEW, "com.example.newapp"
);

// 2. 在 MessageSender 中实现发送方法
private SendResult sendViaNewChannel(String recipient, String message) {
    // 实现 UI 自动化发送
}
```

### 8.3 添加新的 App 适配器

```java
// 1. 创建适配器类
public class NewAppAdapter extends BaseAppAdapter {
    // 实现所有方法
}

// 2. 注册到管理器
adapterManager.registerAdapter(new NewAppAdapter(context));
```