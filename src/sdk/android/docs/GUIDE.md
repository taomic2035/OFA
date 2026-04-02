# OFA Android SDK 操作指导文档

## 1. 快速开始

### 1.1 环境要求

- Android SDK 24+ (Android 7.0)
- Java 17
- Gradle 8.2+

### 1.2 添加依赖

```gradle
dependencies {
    implementation 'com.ofa:agent-sdk:1.2.0'
}
```

### 1.3 配置权限

在 `AndroidManifest.xml` 中添加:

```xml
<!-- 必需权限 -->
<uses-permission android:name="android.permission.INTERNET" />
<uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
<uses-permission android:name="android.permission.SYSTEM_ALERT_WINDOW" />

<!-- 社交通知 -->
<uses-permission android:name="android.permission.READ_CONTACTS" />
<uses-permission android:name="android.permission.CALL_PHONE" />
<uses-permission android:name="android.permission.SEND_SMS" />

<!-- 无障碍服务 -->
<service
    android:name="com.ofa.agent.automation.accessibility.OFAAccessibilityService"
    android:permission="android.permission.BIND_ACCESSIBILITY_SERVICE"
    android:exported="true">
    <intent-filter>
        <action android:name="android.accessibilityservice.AccessibilityService" />
    </intent-filter>
    <meta-data
        android:name="android.accessibilityservice"
        android:resource="@xml/accessibility_config" />
</service>
```

### 1.4 初始化 SDK

```java
public class MyApplication extends Application {
    private OFAAndroidAgent agent;

    @Override
    public void onCreate() {
        super.onCreate();

        // 创建并初始化 Agent
        agent = new OFAAndroidAgent.Builder(this)
            .runMode(AgentProfile.RunMode.HYBRID)
            .center("center.ofa.com", 9090)
            .enableAutomation(true)
            .enableSocial(true)
            .enablePeerNetwork(true)
            .build();

        agent.initialize();
    }

    public OFAAndroidAgent getAgent() {
        return agent;
    }
}
```

---

## 2. 基本使用

### 2.1 执行自然语言任务

```java
OFAAndroidAgent agent = ((MyApplication) getApplication()).getAgent();

// 简单执行
CompletableFuture<TaskResult> future = agent.execute("帮我点一杯珍珠奶茶");

// 获取结果
future.thenAccept(result -> {
    if (result.success) {
        String intent = result.getString("intent");
        Log.i(TAG, "识别意图: " + intent);
    } else {
        Log.e(TAG, "执行失败: " + result.error);
    }
});

// 阻塞等待
TaskResult result = agent.execute("打开WiFi").join();
```

### 2.2 执行技能

```java
// 执行预定义技能
Map<String, String> inputs = new HashMap<>();
inputs.put("shop", "喜茶");
inputs.put("drink", "珍珠奶茶");

CompletableFuture<TaskResult> future = agent.executeSkill(
    "food_order.bubble_tea",
    inputs
);
```

### 2.3 发送社交通知

```java
// 智能发送 - 自动选择渠道
agent.sendNotification("约你明天吃饭", "张三", "13812345678")
    .thenAccept(result -> {
        Log.i(TAG, "发送渠道: " + result.getString("channel"));
    });

// 紧急通知 - 自动电话
agent.sendNotification(
    "[紧急] 服务器宕机了！",
    "运维",
    "13711112222"
);

// 攻略分享 - 自动小红书
agent.sendNotification(
    "分享一个旅游攻略：三亚三日游...",
    "好友",
    null
);
```

### 2.4 执行 UI 自动化

```java
// 获取自动化编排器
AutomationOrchestrator orchestrator = agent.getAutomationOrchestrator();

// 使用模板执行
Map<String, String> params = new HashMap<>();
params.put("query", "奶茶");
params.put("shopName", "喜茶");
params.put("productName", "珍珠奶茶");

AutomationResult result = orchestrator.executeTemplate("food_delivery", params);
```

### 2.5 记忆操作

```java
// 获取记忆管理器
UserMemoryManager memory = agent.getMemoryManager();

// 存储
agent.remember("preferred_tea_shop", "喜茶");
memory.set("preferred_drink", "珍珠奶茶");

// 获取
String shop = agent.recall("preferred_tea_shop");

// 获取建议
List<MemorySuggestion> suggestions = memory.getSuggestions("preferred", 5);
```

---

## 3. 运行模式

### 3.1 选择运行模式

```java
// STANDALONE - 完全离线
OFAAndroidAgent standaloneAgent = new OFAAndroidAgent.Builder(context)
    .runMode(AgentProfile.RunMode.STANDALONE)
    .enableAutomation(true)
    .enableSocial(true)
    .build();

// CONNECTED - 连接 Center
OFAAndroidAgent connectedAgent = new OFAAndroidAgent.Builder(context)
    .runMode(AgentProfile.RunMode.CONNECTED)
    .center("center.ofa.com", 9090)
    .build();

// HYBRID - 混合模式 (推荐)
OFAAndroidAgent hybridAgent = new OFAAndroidAgent.Builder(context)
    .runMode(AgentProfile.RunMode.HYBRID)
    .center("center.ofa.com", 9090)
    .enableAutomation(true)
    .enableSocial(true)
    .enablePeerNetwork(true)
    .build();
```

### 3.2 运行时切换模式

```java
// 切换到独立模式
agent.switchMode(AgentProfile.RunMode.STANDALONE);

// 检查当前模式
AgentProfile.RunMode mode = agent.getRunMode();

// 检查连接状态
boolean connected = agent.isCenterConnected();
boolean network = agent.isNetworkAvailable();
```

### 3.3 监听状态变化

```java
agent.addModeChangeListener((oldMode, newMode) -> {
    Log.i(TAG, "模式切换: " + oldMode + " → " + newMode);
});

agent.addStatusChangeListener((oldStatus, newStatus) -> {
    Log.i(TAG, "状态变化: " + oldStatus + " → " + newStatus);
});
```

---

## 4. 社交通知详细指南

### 4.1 消息类型

SDK 自动识别以下消息类型:

| 类型 | 示例 | 自动选择渠道 |
|------|------|-------------|
| invitation | 约你吃饭、周末聚会 | 微信 |
| urgent | 紧急！服务器宕机 | 电话 |
| reminder | 记得明天开会 | 短信 |
| guide | 分享一个旅游攻略 | 小红书私信 |
| payment | 还我50块钱 | 支付宝 |
| business | 明天有个项目会议 | 钉钉 |
| casual | 好久不见，最近怎么样 | 微信 |

### 4.2 自定义渠道选择

```java
// 强制指定渠道
TaskRequest request = new TaskRequest.Builder()
    .type(TaskRequest.TYPE_SOCIAL)
    .param("message", "约你吃饭")
    .param("recipient", "张三")
    .param("channel", "sms") // 强制使用短信
    .build();

agent.executeTask(request);
```

### 4.3 联系人集成

```java
// 从通讯录查找联系人
SocialOrchestrator social = agent.getSocialOrchestrator();
ContactInfo contact = social.findContact("张三");

if (contact != null) {
    String phone = contact.getPrimaryPhone();
    String wechat = contact.getWeChatId(); // 从备注解析

    // 使用联系人信息发送
    social.sendNotification("约你吃饭", contact.getDisplayName(), phone);
}

// 搜索联系人
List<ContactInfo> contacts = social.searchContacts("李");
```

### 4.4 发送监听

```java
SocialOrchestrator social = agent.getSocialOrchestrator();
social.setListener(new SocialOrchestrator.NotificationListener() {
    @Override
    public void onNotificationStart(NotificationRequest request) {
        Log.d(TAG, "开始发送通知");
    }

    @Override
    public void onChannelSelected(String channel, String reason) {
        Log.d(TAG, "选择渠道: " + channel + ", 原因: " + reason);
    }

    @Override
    public void onDeliveryStart(String channel) {
        Log.d(TAG, "正在通过 " + channel + " 发送");
    }

    @Override
    public void onDeliverySuccess(DeliveryRecord record) {
        Log.i(TAG, "发送成功: " + record.successfulChannel);
    }

    @Override
    public void onDeliveryFailure(DeliveryRecord record) {
        Log.e(TAG, "发送失败: " + record.failureReason);
    }

    @Override
    public void onFallback(String from, String to, String reason) {
        Log.w(TAG, "降级: " + from + " → " + to);
    }
});
```

---

## 5. UI 自动化详细指南

### 5.1 基本操作

```java
AutomationEngine engine = orchestrator.getAutomationEngine();

// 点击
engine.click(100, 200);                           // 坐标点击
engine.click("确定");                              // 文本点击
engine.click(BySelector.id("btn_submit"));        // 选择器点击

// 长按
engine.longClick("标题");

// 滑动
engine.swipe(Direction.DOWN, 0.5f);               // 方向滑动
engine.swipe(100, 500, 100, 100, 500);            // 自定义滑动

// 输入
engine.inputText("Hello World");
engine.inputText(BySelector.className("EditText"), "输入内容");

// 查找
AutomationNode node = engine.findElement(BySelector.text("确定"));
List<AutomationNode> nodes = engine.findElements(BySelector.className("Button"));
```

### 5.2 高级操作

```java
// 等待元素出现
engine.waitForElement(BySelector.text("加载完成"), 5000);

// 滚动查找
engine.scrollFind("查看更多", 10);  // 最多滚动10次

// 等待页面稳定
engine.waitForPageStable(3000);

// 截图
Bitmap screenshot = engine.takeScreenshot();

// 获取页面源码
String source = engine.getPageSource();
```

### 5.3 使用模板

```java
// 外卖模板
Map<String, String> params = new HashMap<>();
params.put("query", "奶茶");
params.put("shopName", "喜茶");
params.put("productName", "珍珠奶茶");

AutomationResult result = orchestrator.executeTemplate("food_delivery", params);

// 购物模板
params = new HashMap<>();
params.put("query", "手机壳");
params.put("productName", "透明手机壳");

result = orchestrator.executeTemplate("shopping", params);
```

### 5.4 错误恢复

```java
// 配置恢复策略
ErrorRecovery recovery = new ErrorRecovery(engine);
recovery.setListener(new ErrorRecovery.RecoveryListener() {
    @Override
    public void onRecoveryAttempt(int attempt, String error, String strategy) {
        Log.d(TAG, "恢复尝试 #" + attempt + ": " + strategy);
    }

    @Override
    public void onRecoverySuccess(String strategy) {
        Log.i(TAG, "恢复成功: " + strategy);
    }
});

// 执行带恢复的操作
if (!result.isSuccess()) {
    result = recovery.recover(result);
}
```

---

## 6. 记忆系统详细指南

### 6.1 三层存储

```java
UserMemoryManager memory = agent.getMemoryManager();

// 存储到三层
memory.set("user_name", "张三");          // L1 Cache → L2 Database → L3 Archive

// 读取 (自动从 L1 → L2 → L3)
String name = memory.get("user_name");

// 删除
memory.delete("user_name");

// 按前缀删除
memory.deleteByKeyPrefix("temp_");
```

### 6.2 智能推荐

```java
// 记录用户偏好
memory.set("preferred_tea_shop", "喜茶");
memory.set("preferred_tea_shop", "奈雪的茶"); // 会记住两个，按频率推荐

// 获取推荐
List<MemorySuggestion> suggestions = memory.getSuggestions("preferred_tea", 5);

for (MemorySuggestion s : suggestions) {
    Log.d(TAG, s.key + " = " + s.value + " (score: " + s.score + ")");
}
```

### 6.3 记录交互

```java
// 记录用户操作
Map<String, String> interaction = new HashMap<>();
interaction.put("shop", "喜茶");
interaction.put("drink", "珍珠奶茶");
interaction.put("sweetness", "半糖");

memory.rememberInteraction("order", interaction);

// 后续可以通过 getInteractions("order") 获取历史
```

### 6.4 导入导出

```java
// 导出所有记忆
String json = memory.exportAll();

// 导入记忆
memory.importAll(json);
```

---

## 7. Peer 网络详细指南

### 7.1 发现 Peer

```java
// 获取发现的 Peer
List<AgentProfile> peers = agent.getPeers();

for (AgentProfile peer : peers) {
    Log.d(TAG, "Peer: " + peer.getName() + " (" + peer.getAgentId() + ")");
}
```

### 7.2 发送消息

```java
// 发送消息给 Peer
boolean sent = agent.sendToPeer("peer-123", "Hello from OFA!");
```

### 7.3 任务委托

```java
// 创建任务请求
TaskRequest request = new TaskRequest.Builder()
    .type(TaskRequest.TYPE_SKILL)
    .param("skillId", "data_processing")
    .param("input", "some data")
    .build();

// 委托给 Peer 执行
TaskResult result = agent.requestFromPeer("peer-123", request);

if (result != null && result.success) {
    Log.i(TAG, "Peer 执行成功: " + result.getString("result"));
}
```

### 7.4 能力匹配

```java
// 查找有特定能力的 Peer
SocialOrchestrator social = agent.getSocialOrchestrator();
PeerInfo peer = social.getOrchestrator()... // 需要通过 PeerNetwork 访问

// 或者直接遍历
for (AgentProfile peer : agent.getPeers()) {
    if (peer.hasCapability("local_llm")) {
        // 这个 Peer 有本地 LLM 能力
    }
}
```

---

## 8. 高级主题

### 8.1 自定义工具

```java
// 1. 实现 ToolExecutor
public class MyCustomTool implements ToolExecutor {
    @Override
    public ToolResult execute(Map<String, String> params) {
        String input = params.get("input");

        // 处理逻辑
        String output = process(input);

        Map<String, Object> result = new HashMap<>();
        result.put("output", output);
        return ToolResult.success(result);
    }

    @Override
    public long getEstimatedTimeMs() {
        return 1000;
    }
}

// 2. 注册工具
ToolRegistry registry = agent.getToolRegistry();
registry.register(
    ToolDefinition.create(
        "my.custom.tool",
        "自定义工具描述",
        "input", "string", true, "输入参数"
    ),
    new MyCustomTool()
);

// 3. 使用工具
Map<String, String> params = new HashMap<>();
params.put("input", "test");
ToolResult result = registry.execute("my.custom.tool", params);
```

### 8.2 自定义技能

```java
// 1. 创建技能定义
SkillDefinition skill = new SkillDefinition.Builder("my_custom_skill")
    .name("我的自定义技能")
    .description("技能描述")
    .input("query", "string", "搜索关键词")
    .output("result", "string", "结果")
    .addStep(SkillStep.builder("TOOL")
        .toolId("my.custom.tool")
        .param("input", "$query")
        .outputKey("step1_result")
        .build())
    .addStep(SkillStep.builder("DELAY")
        .delayMs(1000)
        .build())
    .addStep(SkillStep.builder("TOOL")
        .toolId("another.tool")
        .param("data", "$step1_result")
        .build())
    .build();

// 2. 注册技能
SkillRegistry registry = SkillRegistry.getInstance(context);
registry.register(skill);

// 3. 执行技能
Map<String, String> inputs = new HashMap<>();
inputs.put("query", "test");

SkillResult result = agent.executeSkill("my_custom_skill", inputs).join();
```

### 8.3 自定义 App 适配器

```java
// 1. 继承 BaseAppAdapter
public class MyCustomAppAdapter extends BaseAppAdapter {

    public MyCustomAppAdapter(Context context) {
        super(context, "com.example.myapp");
    }

    @Override
    public String getAppName() {
        return "我的应用";
    }

    @Override
    public boolean isOnApp(AutomationEngine engine) {
        // 检测是否在应用内
        return engine.findElements(BySelector.packageName(getPackageName())).size() > 0;
    }

    @Override
    public String detectCurrentPage(AutomationEngine engine) {
        // 检测当前页面
        if (engine.findElement(BySelector.id("home_container")) != null) {
            return "home";
        }
        if (engine.findElement(BySelector.id("detail_container")) != null) {
            return "detail";
        }
        return "unknown";
    }

    @Override
    public boolean search(AutomationEngine engine, String query) {
        // 实现搜索逻辑
        AutomationNode searchBox = engine.findElement(
            BySelector.className("EditText").id("search_input"));
        if (searchBox == null) return false;

        engine.inputText(searchBox, query);
        engine.pressEnter();
        return true;
    }

    // 实现其他方法...
}

// 2. 注册适配器
AutomationOrchestrator orchestrator = agent.getAutomationOrchestrator();
orchestrator.getAdapterManager().registerAdapter(new MyCustomAppAdapter(context));
```

### 8.4 自定义消息类型

```java
// 修改 MessageClassifier
MessageClassifier classifier = new MessageClassifier(context, memoryManager) {
    @Override
    public ClassificationResult classify(String message) {
        // 添加自定义分类逻辑
        if (message.contains("自定义关键词")) {
            return new ClassificationResult(
                "custom_type",
                URGENCY_MEDIUM,
                "custom_channel",
                0.9,
                new HashMap<>(),
                "自定义分类"
            );
        }
        return super.classify(message);
    }
};
```

---

## 9. 最佳实践

### 9.1 权限请求

```java
// 在使用前检查和请求权限
public class PermissionHelper {
    public static boolean checkAndRequest(Activity activity) {
        String[] permissions = {
            Manifest.permission.READ_CONTACTS,
            Manifest.permission.CALL_PHONE,
            Manifest.permission.SEND_SMS
        };

        List<String> needed = new ArrayList<>();
        for (String perm : permissions) {
            if (ContextCompat.checkSelfPermission(activity, perm)
                != PackageManager.PERMISSION_GRANTED) {
                needed.add(perm);
            }
        }

        if (!needed.isEmpty()) {
            ActivityCompat.requestPermissions(activity,
                needed.toArray(new String[0]), 100);
            return false;
        }

        return true;
    }
}
```

### 9.2 无障碍服务引导

```java
public class AccessibilityHelper {
    public static boolean isAccessibilityEnabled(Context context) {
        int enabled = Settings.Secure.getInt(
            context.getContentResolver(),
            Settings.Secure.ACCESSIBILITY_ENABLED,
            0
        );
        if (enabled != 1) return false;

        String services = Settings.Secure.getString(
            context.getContentResolver(),
            Settings.Secure.ENABLED_ACCESSIBILITY_SERVICES
        );
        return services != null &&
            services.contains(context.getPackageName());
    }

    public static void openAccessibilitySettings(Context context) {
        Intent intent = new Intent(Settings.ACTION_ACCESSIBILITY_SETTINGS);
        context.startActivity(intent);
    }
}
```

### 9.3 错误处理

```java
agent.execute("帮我点奶茶")
    .exceptionally(e -> {
        Log.e(TAG, "执行失败", e);
        showError("操作失败: " + e.getMessage());
        return TaskResult.failure("error", e.getMessage());
    })
    .thenAccept(result -> {
        if (result.success) {
            showSuccess("操作成功");
        } else {
            showError(result.error);
        }
    });
```

### 9.4 资源释放

```java
@Override
protected void onDestroy() {
    super.onDestroy();

    // 释放 Agent 资源
    if (agent != null) {
        agent.shutdown();
    }
}
```

---

## 10. 故障排除

### 10.1 无障碍服务不工作

**症状**: UI 自动化操作失败

**解决方案**:
1. 检查无障碍服务是否启用
2. 检查应用是否有 SYSTEM_ALERT_WINDOW 权限
3. 某些 ROM 需要在开发者选项中开启权限

### 10.2 社交通知发送失败

**症状**: 消息发送返回失败

**解决方案**:
1. 检查目标应用是否已安装
2. 检查是否已登录目标应用
3. 检查联系人是否有对应的社交账号
4. 尝试使用备用渠道

### 10.3 Center 连接失败

**症状**: 无法连接到 OFA Center

**解决方案**:
1. 检查网络连接
2. 检查 Center 地址和端口是否正确
3. 切换到 STANDALONE 模式继续使用

### 10.4 Peer 发现不到

**症状**: getPeers() 返回空列表

**解决方案**:
1. 确保在同一局域网
2. 检查 NSD 是否正常工作
3. 确保两个设备都开启了 PeerNetwork

---

## 11. API 参考

### 11.1 OFAAndroidAgent

| 方法 | 说明 |
|------|------|
| `initialize()` | 初始化 Agent |
| `execute(String)` | 执行自然语言 |
| `executeSkill(String, Map)` | 执行技能 |
| `executeAutomation(String, Map)` | 执行自动化 |
| `sendNotification(String, String, String)` | 发送社交通知 |
| `switchMode(RunMode)` | 切换运行模式 |
| `getRunMode()` | 获取当前模式 |
| `getPeers()` | 获取发现的 Peer |
| `remember(String, String)` | 存储记忆 |
| `recall(String)` | 读取记忆 |
| `shutdown()` | 关闭 Agent |

### 11.2 TaskRequest

| 方法 | 说明 |
|------|------|
| `intent(String)` | 创建意图请求 |
| `skill(String, Map)` | 创建技能请求 |
| `automation(String, Map)` | 创建自动化请求 |
| `social(String, String, String)` | 创建社交通知请求 |
| `naturalLanguage(String)` | 创建自然语言请求 |

### 11.3 TaskResult

| 字段 | 说明 |
|------|------|
| `success` | 是否成功 |
| `data` | 返回数据 |
| `error` | 错误信息 |
| `executionTimeMs` | 执行时间 |
| `executedBy` | 执行者 (local/center/peer:xxx) |