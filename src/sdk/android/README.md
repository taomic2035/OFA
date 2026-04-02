# OFA Android Agent SDK

Android SDK for building intelligent agent applications with MCP support, dual LLM capabilities, intent understanding, cross-app UI automation, and smart social notifications.

## Features

### Core Capabilities
- **UI Automation**: Cross-app automation via AccessibilityService (Phase 1-5)
- **AI Agent Enhancement**: On-device ML inference and intelligent decision making
- **Social Notifications**: Smart message delivery across multiple channels
- **Skill Orchestration**: Create multi-step automation tasks
- **Intent Understanding**: Natural language to structured intents
- **Memory System**: Three-layer user memory (L1/L2/L3)
- **Dual LLM Support**: Cloud LLM + Local LLM with auto-failover
- **MCP Protocol**: Full Model Context Protocol support

### Automation Features
- **4 App Adapters**: 美团外卖, 饿了么, 淘宝, 京东
- **6 Error Recovery Strategies**: Auto-recovery from failures
- **5 Keep-Alive Strategies**: Background service protection
- **ROM System Layer**: Silent install, permission grant (requires root/system)

### AI Features
- **Local Intent Classification**: On-device intent recognition
- **Multi-Armed Bandit**: Thompson Sampling, UCB, Epsilon-Greedy
- **Operation Recommendation**: Context-aware suggestions
- **UI Element Recognition**: Screen understanding

### Social Notification Features
- **9 Communication Channels**: 微信, 电话, 短信, 支付宝, 抖音, 小红书, 钉钉, 企业微信, QQ
- **10 Message Types**: invitation, urgent, reminder, guide, payment, casual, business, greeting, location, unknown
- **4 Urgency Levels**: low, medium, high, critical
- **Smart Channel Selection**: Modern social habits based routing
- **Multi-Channel Fallback**: Automatic retry on alternative channels
- **Contact Integration**: Access device contacts with social handles

## Requirements

- Android SDK 24+ (Android 7.0)
- Java 17
- Gradle 8.2+

## Installation

```gradle
dependencies {
    implementation 'com.ofa:agent-sdk:1.0.9'
}
```

## Quick Start

### 1. Unified Agent Entry (Recommended)

```java
// Initialize agent with HYBRID mode (recommended)
OFAAndroidAgent agent = new OFAAndroidAgent.Builder(context)
    .runMode(AgentProfile.RunMode.HYBRID)
    .center("center.example.com", 9090)
    .enableAutomation(true)
    .enableSocial(true)
    .enablePeerNetwork(true)
    .build();

agent.initialize();

// Execute natural language input
CompletableFuture<TaskResult> result = agent.execute("帮我点一杯珍珠奶茶");

// Execute skill
result = agent.executeSkill("food_order.bubble_tea", Map.of("shop", "喜茶"));

// Send social notification
result = agent.sendNotification("约你明天吃饭", "张三", "13812345678");

// Check status
String report = agent.getStatusReport();
```

### 2. AI-Enhanced Automation

```java
// Initialize with memory support
UserMemoryManager memoryManager = new UserMemoryManager(context);
AIEnhancedOrchestrator orchestrator = new AIEnhancedOrchestrator(context, memoryManager);
orchestrator.initialize();

// Process natural language
AutomationResult result = orchestrator.processNaturalLanguage("帮我点一杯珍珠奶茶");

// Get recommendations
List<Recommendation> recommendations = orchestrator.getRecommendations();

// Execute smart operation
result = orchestrator.executeSmart("search", Map.of("query", "奶茶"));
```

### 3. Basic Automation

```java
AutomationOrchestrator orchestrator = new AutomationOrchestrator(context);
orchestrator.initialize();

// Execute operation
Map<String, String> params = new HashMap<>();
params.put("query", "奶茶");
params.put("shopName", "喜茶");
params.put("productName", "珍珠奶茶");

AutomationResult result = orchestrator.executeTemplate("food_delivery", params);
```

### 3. Intent Understanding

```java
IntentEngine intentEngine = new IntentEngine();
UserIntent intent = intentEngine.recognizeBest("打开WiFi");

if (intent != null) {
    String action = intent.getIntentName(); // "wifi_on"
    Map<String, String> slots = intent.getSlots();
}
```

### 4. Skill Execution

```java
SkillRegistry registry = SkillRegistry.getInstance(context);
CompositeSkillExecutor executor = new CompositeSkillExecutor(context, toolRegistry);

SkillDefinition skill = registry.getSkill("food_order.bubble_tea");
SkillResult result = executor.execute(skill, inputs).join();
```

### 5. Memory System

```java
UserMemoryManager memory = new UserMemoryManager(context);

// Store preference
memory.set("preferred_tea", "珍珠奶茶");

// Get suggestion
List<MemorySuggestion> suggestions = memory.getSuggestions("preferred_tea", 5);

// Remember interaction
memory.rememberInteraction("order", Map.of("shop", "喜茶", "drink", "珍珠奶茶"));
```

### 6. Social Notifications (NEW!)

```java
// Initialize social orchestrator
SocialOrchestrator social = new SocialOrchestrator(context, automationEngine, memoryManager);

// Smart notification - auto-select best channel
DeliveryRecord record = social.sendNotification(
    "约你明天吃饭",       // message
    "张三",              // recipient name
    "13812345678"        // phone (optional)
);
// → Automatically sends via WeChat (social invitation)

// Urgent message - auto-select phone
record = social.sendUrgent("服务器宕机了！", "技术主管", "13711112222");

// Share guide - auto-select Xiaohongshu
record = social.sendGuide("旅游攻略", "推荐几个不错的景点...", "好友");

// Payment reminder - auto-select Alipay
record = social.sendPaymentReminder("50", "借款人", "13555556666");
```

## Architecture

```
┌────────────────────────────────────────────────────────────────────┐
│                      OFAAndroidAgent                                │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │                    AgentModeManager                          │  │
│  │  STANDALONE │ CONNECTED │ HYBRID                             │  │
│  └─────────────────────────────────────────────────────────────┘  │
├────────────────────────────────────────────────────────────────────┤
│  ┌───────────────┐ ┌───────────────┐ ┌───────────────────────┐   │
│  │ Center        │ │ Peer          │ │ Local                 │   │
│  │ Connection    │ │ Network       │ │ Execution Engine      │   │
│  │ (gRPC)        │ │ (NSD/P2P)     │ │                       │   │
│  └───────────────┘ └───────────────┘ └───────────────────────┘   │
├────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │                   LocalExecutionEngine                       │  │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────────────┐   │  │
│  │  │ Intent  │ │ Skill   │ │ Auto-   │ │ Social          │   │  │
│  │  │ Engine  │ │ Executor│ │ mation  │ │ Orchestrator    │   │  │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────────────┘   │  │
│  └─────────────────────────────────────────────────────────────┘  │
├────────────────────────────────────────────────────────────────────┤
│  ┌───────────────┐ ┌───────────────┐ ┌───────────────────────┐   │
│  │ Memory        │ │ AI Enhanced   │ │ App Adapters          │   │
│  │ System        │ │ Orchestrator  │ │ (4 apps)              │   │
│  │ (L1/L2/L3)    │ │ (MAB/LLM)     │ │                       │   │
│  └───────────────┘ └───────────────┘ └───────────────────────┘   │
└────────────────────────────────────────────────────────────────────┘
```

## Running Modes

The Android Agent supports three running modes:

### 1. STANDALONE Mode
- Complete local execution
- No network dependency
- All capabilities work offline
- Best for: Privacy-focused apps, offline scenarios

### 2. CONNECTED Mode
- Always connected to Center
- Receives remote tasks
- Reports status to Center
- Best for: Managed device fleets, enterprise scenarios

### 3. HYBRID Mode (Recommended)
- Local-first execution
- Cloud enhancement when available
- Automatic fallback to local
- Best for: Consumer apps, mixed connectivity

```java
// Initialize with HYBRID mode (default)
OFAAndroidAgent agent = new OFAAndroidAgent.Builder(context)
    .runMode(AgentProfile.RunMode.HYBRID)
    .center("center.example.com", 9090)
    .enableAutomation(true)
    .enableSocial(true)
    .enablePeerNetwork(true)
    .build();

agent.initialize();

// Switch mode at runtime
agent.switchMode(AgentProfile.RunMode.STANDALONE);

// Check status
boolean connected = agent.isCenterConnected();
List<AgentProfile> peers = agent.getPeers();
```

## Task Execution Flow

```
User Input → TaskRequest → AgentModeManager
                                ↓
            ┌───────────────────┼───────────────────┐
            ↓                   ↓                   ↓
        STANDALONE          CONNECTED            HYBRID
            ↓                   ↓                   ↓
     LocalExecution      CenterConnection    Intelligent
        Engine                │               Routing
            ↓                  ↓                   ↓
     ┌──────┴──────┐     ┌─────┴─────┐     ┌──────┴──────┐
     ↓             ↓     ↓           ↓     ↓             ↓
   Intent     Skill   Task      Result  Local      Center
   Engine    Executor Assigned  Returned  First      Fallback
```

## App Adapters

### Supported Apps

| App | Package | Operations |
|-----|---------|------------|
| 美团外卖 | com.sankuai.meituan | search, selectShop, selectProduct, addToCart, checkout, pay |
| 饿了么 | me.ele | search, selectShop, selectProduct, addToCart, checkout, pay |
| 淘宝 | com.taobao.taobao | search, selectProduct, configureOptions, addToCart, checkout, pay |
| 京东 | com.jingdong.app.mall | search, selectProduct, configureOptions, addToCart, checkout, pay |

### Using Adapters

```java
// Adapter is auto-detected based on current app
AppAdapterManager adapterManager = new AppAdapterManager();
AppAdapter adapter = adapterManager.detectAdapter(engine);

if (adapter != null) {
    adapter.search(engine, "奶茶");
    adapter.selectShop(engine, "喜茶");
    adapter.selectProduct(engine, "珍珠奶茶");
    adapter.addToCart(engine, 1);
    adapter.goToCheckout(engine);
    adapter.submitOrder(engine);
}
```

## Operation Templates

### Built-in Templates

| Template | Description | Required Params |
|----------|-------------|-----------------|
| `food_delivery` | Complete food order flow | query, shopName, productName |
| `shopping` | Shopping order flow | query, productName |
| `search_and_add` | Search and add to cart | query, productName |

### Custom Templates

```java
OperationTemplate template = new OperationTemplate.Builder(
    "custom_order",
    "Custom Order",
    "Custom order flow",
    "food"
)
    .requiredParam("shopName")
    .requiredParam("productName")
    .addStep(OperationTemplate.TemplateStep.builder("search")
        .param("query", "$shopName")
        .waitAfter(1500)
        .build())
    .addStep(OperationTemplate.TemplateStep.builder("selectProduct")
        .param("productName", "$productName")
        .build())
    .build();

templateRegistry.register(template);
```

## Error Recovery

### Recovery Strategies

| Strategy | Trigger | Action |
|----------|---------|--------|
| ScrollToFind | Element not found | Scroll to find element |
| WaitAndRetry | Timeout | Wait and retry |
| DismissOverlay | Dialog blocking | Press back |
| WaitForPage | Page loading | Wait for stability |
| HandlePermission | Permission denied | Guide to settings |
| HandleNetwork | Network error | Wait and refresh |

### Using Recovery

```java
ErrorRecovery recovery = new ErrorRecovery(engine);
recovery.setListener(new ErrorRecovery.RecoveryListener() {
    @Override
    public void onRecoveryAttempt(int attempt, String error, String strategy) {
        Log.d(TAG, "Recovery attempt " + attempt + ": " + strategy);
    }

    @Override
    public void onRecoverySuccess(String strategy) {
        Log.i(TAG, "Recovery succeeded: " + strategy);
    }
});

// Automatic recovery on failure
if (!result.isSuccess()) {
    result = recovery.recover(result);
}
```

## Retry Policy

### Preset Policies

```java
// No retry
RetryPolicy.noRetry();

// Quick retry (3 attempts, 500ms start)
RetryPolicy.quick();

// Standard retry (3 attempts, 1s start, 2x backoff)
RetryPolicy.standard();

// Aggressive retry (5 attempts, 2s start)
RetryPolicy.aggressive();

// Network optimized (5 attempts, 30% jitter)
RetryPolicy.network();

// UI optimized (3 attempts, 500ms start)
RetryPolicy.ui();
```

### Custom Policy

```java
RetryPolicy policy = RetryPolicy.builder()
    .maxRetries(5)
    .initialDelay(1000)
    .maxDelay(30000)
    .backoffMultiplier(2.0)
    .jitterFactor(0.1)
    .condition(result -> result.getMessage().contains("timeout"))
    .build();
```

## AI Decision Engine

### Multi-Armed Bandit

```java
SmartDecisionEngine decisionEngine = new SmartDecisionEngine(context, memoryManager);

// Register options
decisionEngine.registerOptions("shop_selection", Arrays.asList("喜茶", "奈雪", "蜜雪冰城"));

// Select optimal option
String selected = decisionEngine.selectOption("shop_selection");

// Report outcome
decisionEngine.reportReward("shop_selection", selected, 1.0); // success = 1.0
```

### Decision Types

| Type | Description |
|------|-------------|
| `shop_selection` | Select optimal shop |
| `payment_method` | Select payment method |
| `retry_strategy` | Select retry approach |
| `timing` | Select execution timing |

## System-Level Operations

> Requires system permissions (INSTALL_PACKAGES, WRITE_SECURE_SETTINGS) or root access.

### Silent Install

```java
HybridAutomationEngine engine = new HybridAutomationEngine(context);

if (engine.hasSystemLevelAccess()) {
    // Silent install
    engine.installApp("/path/to/app.apk");

    // Silent uninstall
    engine.uninstallApp("com.example.app");

    // Grant permission silently
    engine.grantPermission("com.example.app", "android.permission.READ_CONTACTS");

    // Enable accessibility service
    engine.enableAccessibilityService("com.ofa.agent/.automation.accessibility.OFAAccessibilityService");
}
```

### Keep-Alive Strategies

```java
engine.enableKeepAlive();  // Enable all strategies
engine.disableKeepAlive(); // Disable
```

| Strategy | Description |
|----------|-------------|
| ForegroundService | Persistent notification |
| WakeLock | CPU wake lock |
| BatteryOptimization | Whitelist exemption |
| SystemApp | System app privilege |
| RootKeepAlive | Process priority via root |

## Performance Monitoring

```java
PerformanceMonitor monitor = new PerformanceMonitor();

// Start timing
PerformanceMonitor.OperationTimer timer = monitor.startOperation("search");

// ... perform operation ...

// Complete
timer.success(); // or timer.failure("error message")

// Get stats
String report = monitor.generateReport();
```

## Automation Logger

```java
AutomationLogger logger = new AutomationLogger(context);

// Log operations
logger.info("search", "Searching for: 奶茶");
logger.warn("checkout", "Payment method not selected");
logger.error("pay", "Payment failed", "Network error");

// Export logs
JSONArray logs = logger.exportLogs();
```

## MCP Tools (40+ Built-in)

### System Tools
- `app.launch`, `app.list`, `app.info`
- `settings.get`, `settings.set`
- `clipboard.read`, `clipboard.write`
- `file.read`, `file.write`, `file.list`
- `notification.send`, `notification.cancel`

### Device Tools
- `camera.capture`, `camera.scan`
- `bluetooth.scan`, `bluetooth.list`
- `wifi.scan`, `wifi.list`, `wifi.info`
- `nfc.status`, `nfc.read`
- `sensor.list`, `sensor.read`
- `battery.status`

### Data Tools
- `contacts.query`, `contacts.search`
- `calendar.query`, `calendar.today`
- `media.images`, `media.videos`, `media.audio`

### AI Tools
- `speech.synthesize`, `speech.stop`

### UI Automation Tools
- `ui.click`, `ui.longClick`, `ui.swipe`, `ui.input`
- `ui.find`, `ui.wait`, `ui.scrollFind`
- `ui.capture`, `ui.waitForStable`
- `ui.startRecord`, `ui.stopRecord`, `ui.replay`

### System Tools (ROM)
- `system.install`, `system.uninstall`
- `system.grantPermission`, `system.setSecureSetting`
- `system.enableAccessibility`, `system.keepAlive`
- `system.getCapability`

### Social Tools (NEW!)
- `social.send` - Smart message with auto channel selection
- `social.invite` - Send invitation (约吃饭) via WeChat
- `social.urgent` - Send urgent message via phone
- `social.guide` - Share tips via Xiaohongshu
- `social.payment` - Payment reminder via Alipay
- `social.classify` - Analyze message type and recommended channel
- `social.contact.find` - Find contact by name
- `social.contact.search` - Search contacts
- `social.channel.list` - List available channels
- `social.stats` - Get delivery statistics

## Social Notifications

### Smart Channel Selection

The system automatically selects the best communication channel based on modern social habits:

| Message Type | Example | Auto-Selected Channel | Reason |
|--------------|---------|----------------------|--------|
| Invitation (邀请) | 约你吃饭 | 微信 (WeChat) | Easy discussion, social context |
| Urgent (紧急) | 服务器宕机！ | 电话 (Phone) | Immediate response required |
| Guide (攻略) | 旅游攻略分享 | 小红书私信 | Content sharing platform |
| Payment (支付) | 还我50块钱 | 支付宝 (Alipay) | Financial context |
| Business (工作) | 明天开会 | 钉钉/企业微信 | Work context |
| Reminder (提醒) | 记得缴费 | 短信 (SMS) | Reliable delivery |
| Casual (日常) | 好久不见 | 微信 (WeChat) | Social platform |

### Message Types

```java
// Supported message types
MessageClassifier.TYPE_INVITATION  // 邀请: 约吃饭, 聚会
MessageClassifier.TYPE_URGENT      // 紧急: 服务器宕机, 急事
MessageClassifier.TYPE_REMINDER    // 提醒: 记得缴费, 明天会议
MessageClassifier.TYPE_GUIDE       // 攻略: 旅游攻略, 使用教程
MessageClassifier.TYPE_PAYMENT     // 支付: 转账, 还款
MessageClassifier.TYPE_CASUAL      // 日常: 聊天, 问候
MessageClassifier.TYPE_BUSINESS    // 工作: 任务, 审批
MessageClassifier.TYPE_GREETING    // 问候: 你好, 在吗
MessageClassifier.TYPE_LOCATION    // 位置: 我在咖啡厅
```

### Urgency Levels

```java
MessageClassifier.URGENCY_LOW       // Low: 可以等等
MessageClassifier.URGENCY_MEDIUM    // Medium: 明天/后天
MessageClassifier.URGENCY_HIGH      // High: 今天/现在
MessageClassifier.URGENCY_CRITICAL  // Critical: 紧急/急
```

### Supported Channels

| Channel | Package | Capabilities |
|---------|---------|--------------|
| 微信 (WeChat) | com.tencent.mm | Text, Voice, Image, Video, Location, Payment, Group |
| 电话 (Phone) | built-in | Voice |
| 短信 (SMS) | built-in | Text, Group |
| 支付宝 (Alipay) | com.eg.android.AlipayGphone | Text, Image, Payment |
| 抖音 (Douyin) | com.ss.android.ugc.aweme | Text, Image, Video, Group |
| 小红书 (Xiaohongshu) | com.xingin.xhs | Text, Image, Video, Group |
| 钉钉 (DingTalk) | com.alibaba.android.rimet | Text, Voice, Image, Video, Location, Group |
| 企业微信 (WeCom) | com.tencent.wework | Text, Voice, Image, Video, Location, Group |
| QQ | com.tencent.mobileqq | Text, Voice, Image, Video, Location, Group |

### Using Social Tools

```java
// Initialize
SocialOrchestrator social = new SocialOrchestrator(context, automationEngine, memoryManager);

// Set listener for tracking
social.setListener(new NotificationListener() {
    void onChannelSelected(String channel, String reason) {
        Log.i(TAG, "Selected: " + channel + " because " + reason);
    }
    void onDeliverySuccess(DeliveryRecord record) {
        Log.i(TAG, "Sent via " + record.successfulChannel);
    }
    void onFallback(String from, String to, String reason) {
        Log.w(TAG, "Fallback: " + from + " → " + to);
    }
});

// Send with auto-classification
DeliveryRecord record = social.sendNotification("约你明天吃饭", "张三", "13812345678");
// → Automatically uses WeChat

// Send urgent (auto phone call)
record = social.sendUrgent("紧急！服务器挂了！", "运维", "13711112222");

// Share guide (auto Xiaohongshu)
record = social.sendGuide("三亚攻略", "推荐酒店...", "好友");

// Payment reminder (auto Alipay)
record = social.sendPaymentReminder("100", "借款人", "13555556666");

// Get statistics
Map<String, Map<String, Double>> stats = social.getChannelStatistics();
// Returns: successRate, avgDuration, totalAttempts per channel
```

### Multi-Channel Fallback

```java
// Enable fallback (default: true)
social.setMultiChannelFallback(true, 3); // max 3 channels

// Example: If WeChat fails, automatically try SMS, then Phone
DeliveryRecord record = social.sendNotification("重要消息", "联系人", "13812345678");
// Attempted: [wechat, sms, phone]
```

### Contact Integration

```java
// Find contact
ContactInfo contact = social.findContact("张三");
if (contact != null) {
    String phone = contact.getPrimaryPhone();
    String wechat = contact.getWeChatId();
}

// Search contacts
List<ContactInfo> contacts = social.searchContacts("李");

// Contact notes can store social handles:
// "微信: abc123\n抖音: @user\n小红书: @creator"
```

## Permissions

```xml
<!-- Required -->
<uses-permission android:name="android.permission.INTERNET" />
<uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />

<!-- Automation -->
<uses-permission android:name="android.permission.SYSTEM_ALERT_WINDOW" />

<!-- Optional tools -->
<uses-permission android:name="android.permission.CAMERA" />
<uses-permission android:name="android.permission.BLUETOOTH" />
<uses-permission android:name="android.permission.ACCESS_FINE_LOCATION" />
<uses-permission android:name="android.permission.READ_CONTACTS" />
<uses-permission android:name="android.permission.READ_CALENDAR" />
```

## Project Statistics

| Metric | Count |
|--------|-------|
| Java Classes | 120+ |
| Built-in Intents | 22 |
| Step Types | 12 |
| Built-in Tools | 50+ |
| App Adapters | 4 |
| Operation Templates | 3 |
| Recovery Strategies | 6 |
| Retry Presets | 6 |
| AI Components | 9 |
| Decision Strategies | 3 |
| Social Channels | 9 |
| Message Types | 10 |

## Version History

| Version | Feature |
|---------|---------|
| v1.1.0 | Social Notification System (Smart messaging across 9 channels) |
| v1.0.9 | AI Agent Enhancement (LocalAI, MAB, Recommendations) |
| v1.0.8 | Integration & Optimization (Memory, Intent, Skill bridges) |
| v1.0.7 | ROM System Layer (Silent install, Keep-alive) |
| v1.0.6 | App Adapter Layer (美团, 饿了么, 淘宝, 京东) |
| v1.0.5 | Enhanced Automation (Scroll, Capture, Record/Replay) |
| v1.0.4 | Basic Automation (AccessibilityService) |
| v1.0.3 | Memory System (L1/L2/L3) |
| v1.0.2 | Skill Orchestration |
| v1.0.1 | Intent Understanding |

## Documentation

- [Skill System Guide](docs/SKILL_SYSTEM.md)
- [Intent Understanding System](docs/INTENT_SYSTEM.md)
- [Automation System](docs/AUTOMATION_SYSTEM.md)
- [AI Agent Plan](docs/AI_AGENT_PLAN.md)
- [Changelog](../../CHANGELOG.md)

## License

MIT License