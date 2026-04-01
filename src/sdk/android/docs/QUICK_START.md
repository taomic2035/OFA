# OFA Android SDK 快速入门

本指南帮助你快速集成 OFA Android SDK 到你的应用中。

## 1. 添加依赖

在 `build.gradle` 中添加：

```gradle
dependencies {
    implementation 'com.ofa:agent-sdk:1.0.0'
}
```

## 2. 配置权限

在 `AndroidManifest.xml` 中添加必需权限：

```xml
<!-- 必需 -->
<uses-permission android:name="android.permission.INTERNET" />
<uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />

<!-- 可选 - 根据使用的工具添加 -->
<uses-permission android:name="android.permission.CAMERA" />
<uses-permission android:name="android.permission.BLUETOOTH" />
<uses-permission android:name="android.permission.READ_CONTACTS" />
```

## 3. 初始化 Agent

### 方式一：在 Application 中初始化

```java
public class MyApp extends Application {
    private OFAAgent agent;

    @Override
    public void onCreate() {
        super.onCreate();

        agent = new OFAAgent.Builder(this)
            .agentId("my-device")
            .agentName("My Phone")
            .centerAddress("192.168.1.100")
            .centerPort(9090)
            .offlineLevel(OfflineLevel.L4)
            .enableTools(true)
            .build();

        // 注册离线技能
        OfflineSkills.registerAll(agent.getOfflineManager());

        agent.connect();
    }
}
```

### 方式二：使用后台服务

```java
// 启动服务
Intent intent = new Intent(this, OFAAgentService.class);
intent.setAction("CONNECT");
ContextCompat.startForegroundService(this, intent);
```

## 4. 调用工具

```java
// 查询电池状态
ToolResult result = agent.callTool("battery.status", new HashMap<>());
if (result.isSuccess()) {
    int level = result.getOutput().optInt("level");
    Log.i("Battery", "Level: " + level + "%");
}

// 调用计算器
OfflineManager om = agent.getOfflineManager();
String taskId = om.executeLocal("calculator", "10 + 20".getBytes());
```

## 5. 处理离线模式

```java
OfflineManager om = agent.getOfflineManager();

// 监听离线状态
om.addOfflineModeListener(offline -> {
    if (offline) {
        // 进入离线模式
    } else {
        // 恢复在线，同步数据
        om.syncNow();
    }
});
```

## 6. 集成 AI Agent

```java
AIAgentInterface ai = agent.getAIAgentInterface();

// 获取工具列表 (OpenAI 函数格式)
JSONArray functions = ai.getToolsAsFunctions();

// 调用工具
Map<String, Object> args = new HashMap<>();
args.put("operation", "capture");
ToolResult result = ai.callTool("camera.capture", args);
```

## 下一步

- 查看 [README.md](README.md) 了解完整 API
- 查看 [MCP_TOOLS_GUIDE.md](docs/MCP_TOOLS_GUIDE.md) 了解工具开发
- 查看 [CHANGELOG.md](CHANGELOG.md) 了解版本更新