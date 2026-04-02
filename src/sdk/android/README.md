# OFA Android Agent SDK

Android SDK for building OFA Agent applications with MCP (Model Context Protocol) support, dual LLM capabilities, and intent understanding.

## Features

- **Skill Orchestration**: Create multi-step automation tasks (order food, morning routine, etc.)
- **Intent Understanding**: Parse natural language to structured intents and execute tools
- **Dual LLM Support**: Cloud LLM (OpenAI/Claude) + Local LLM (TensorFlow Lite)
- **Auto Failover**: Automatic switching between cloud and local LLM
- **Offline Capable**: Run LLM inference entirely on-device
- **MCP Protocol**: Full Model Context Protocol support
- **30+ Built-in Tools**: System, device, data, and AI tools
- **30+ Pre-defined Intents**: System, communication, media, device, navigation, app
- **P2P Communication**: Agent-to-agent messaging

## Requirements

- Android SDK 24+ (Android 7.0)
- Java 17
- Gradle 8.2+

## Installation

Add to your app's `build.gradle`:

```gradle
dependencies {
    implementation 'com.ofa:agent-sdk:1.0.0'
}
```

## Quick Start

### 1. Initialize Agent with Cloud LLM

```java
OFAAgent agent = new OFAAgent.Builder(context)
    .agentId("my-android-agent")
    .agentName("My Phone")
    .centerAddress("192.168.1.100")
    .centerPort(9090)
    .type(OFAAgent.AgentType.MOBILE)
    // Configure cloud LLM
    .cloudLLM("https://api.openai.com/v1", "sk-xxx", "gpt-4-turbo")
    .build();

agent.connect();
```

### 2. Initialize with Cloud + Local LLM (Recommended)

```java
OFAAgent agent = new OFAAgent.Builder(context)
    .agentId("my-android-agent")
    .centerAddress("192.168.1.100")
    .centerPort(9090)
    // Cloud LLM (primary)
    .cloudLLM("https://api.openai.com/v1", "sk-xxx", "gpt-4-turbo")
    // Local LLM (fallback/offline)
    .localLLM("/data/local/tmp/gemma-2b.tflite")
    .autoLLMFailover(true)  // Auto-switch on failure
    .offlineLevel(OfflineLevel.L4)
    .build();

agent.connect();
```

### 3. Use LLM Directly

```java
// Check LLM availability
if (agent.hasLLM()) {
    LLMProvider llm = agent.getLLMProvider();

    // Chat
    LLMResponse response = llm.chat("Hello, how are you?").join();
    if (response.isSuccess()) {
        Log.i("LLM", response.getContent());
    }

    // Stream chat
    llm.streamChat(messages, new StreamCallback() {
        @Override
        public void onToken(String token) {
            // Handle streaming token
        }
        @Override
        public void onComplete(LLMResponse response) {
            // Chat complete
        }
    });
}
```

### 4. Use LLM as MCP Tool

```java
// Call LLM through tool interface
Map<String, Object> args = new HashMap<>();
args.put("message", "Translate 'hello' to Chinese");
args.put("system", "You are a helpful translator");

ToolResult result = agent.callTool("llm.chat", args);
if (result.isSuccess()) {
    JSONObject output = result.getOutput();
    String content = output.getString("content");
}
```

### 5. Handle Offline Mode

```java
OfflineManager om = agent.getOfflineManager();

om.addOfflineModeListener(offline -> {
    if (offline) {
        // Switched to local LLM automatically
        Log.i("Agent", "Offline mode - using local LLM");
    } else {
        // Back online - syncing
        om.syncNow();
    }
});
```

### 6. Intent Understanding System

```java
// Initialize task executor
ToolRegistry registry = agent.getToolRegistry();
TaskExecutor executor = new TaskExecutor(context, registry);

// Process user input
executor.process("打开WiFi", new TaskExecutor.Callback() {
    @Override
    public void onStatusUpdate(String taskId, ExecutionStatus status, String message) {
        Log.d(TAG, "Status: " + status);
    }

    @Override
    public void onConfirmationRequired(String taskId, UserIntent intent, String message) {
        // Show confirmation dialog
        // User confirms: executor.confirmAndExecute(taskId, intent, mapping, null);
    }

    @Override
    public void onSlotMissing(String taskId, List<String> missingSlots) {
        // Ask user for missing information
    }

    @Override
    public void onComplete(String taskId, TaskResult result) {
        if (result.isSuccess()) {
            Log.i(TAG, "Success: " + result.toolResult.getOutput());
        } else {
            Log.e(TAG, "Failed: " + result.message);
        }
    }
});
```

### 7. Custom Intent Registration

```java
// Define custom intent
IntentDefinition selfieIntent = new IntentDefinition.Builder()
    .id("custom.take_selfie")
    .category("media")
    .action("take_selfie")
    .description("使用前置摄像头自拍")
    .keywords("自拍", "selfie", "拍自己")
    .pattern("自拍|拍.*自己|selfie")
    .defaultConfidence(0.9)
    .build();

// Register intent
executor.registerIntent(selfieIntent);

// Map intent to tool
executor.registerMapping("custom.take_selfie", "camera.capture",
    Map.of("mode", "mode"),
    Map.of("mode", "photo", "camera", "front"),
    false, null);
```

## Pre-defined Intents

| Category | Intent | Example Input | Tool |
|----------|--------|---------------|------|
| system | `open_settings` | "打开设置" | `settings.open` |
| system | `volume` | "把音量调大" | `audio.control` |
| communication | `call` | "打电话给张三" | `phone.call` |
| communication | `sms` | "发短信给李四" | `phone.sms` |
| media | `capture` | "帮我拍照" | `camera.capture` |
| media | `play_music` | "播放周杰伦的歌" | `media.play` |
| device | `wifi_on` | "打开WiFi" | `wifi.status` |
| device | `battery` | "电池还剩多少" | `battery.status` |
| navigation | `navigate` | "导航到北京" | `maps.navigate` |
| navigation | `current_location` | "我在哪" | `location.get` |
| app | `open` | "打开微信" | `app.launch` |

## MCP Tools

### System Tools

| Tool | Description | Offline |
|------|-------------|---------|
| `app.launch` | Launch application | ✅ |
| `app.list` | List installed apps | ✅ |
| `app.info` | Get app info | ✅ |
| `settings.get` | Get device setting | ✅ |
| `settings.set` | Set device setting | ✅ |
| `clipboard.read` | Read clipboard | ✅ |
| `clipboard.write` | Write to clipboard | ✅ |
| `file.read` | Read file | ✅ |
| `file.write` | Write file | ✅ |
| `file.list` | List files | ✅ |
| `notification.send` | Send notification | ✅ |

### Device Tools

| Tool | Description | Permissions |
|------|-------------|-------------|
| `camera.capture` | Capture photo | CAMERA |
| `camera.scan` | Scan QR/barcode | CAMERA |
| `camera.list` | List cameras | - |
| `bluetooth.scan` | Scan Bluetooth | BLUETOOTH |
| `bluetooth.list` | List paired devices | BLUETOOTH |
| `wifi.scan` | Scan WiFi | LOCATION |
| `wifi.status` | WiFi status | - |
| `nfc.status` | NFC status | - |
| `sensor.list` | List sensors | - |
| `sensor.read` | Read sensor | - |
| `battery.status` | Battery status | - |

### Data Tools

| Tool | Description | Permissions |
|------|-------------|-------------|
| `contacts.query` | Query contacts | READ_CONTACTS |
| `contacts.search` | Search contacts | READ_CONTACTS |
| `calendar.query` | Query events | READ_CALENDAR |
| `calendar.today` | Today's events | READ_CALENDAR |
| `media.images` | Query images | STORAGE |
| `media.videos` | Query videos | STORAGE |
| `media.audio` | Query audio | STORAGE |

### AI Tools

| Tool | Description |
|------|-------------|
| `speech.synthesize` | Text-to-speech |
| `speech.stop` | Stop speech |

## Custom Tools

Implement `ToolExecutor` interface:

```java
public class MyTool implements ToolExecutor {

    @NonNull
    @Override
    public String getToolId() {
        return "my.tool";
    }

    @NonNull
    @Override
    public String getDescription() {
        return "My custom tool";
    }

    @NonNull
    @Override
    public ToolResult execute(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        // Get arguments
        String param = (String) args.get("param");

        // Process
        JSONObject output = new JSONObject();
        output.put("result", process(param));

        return new ToolResult(getToolId(), output, 100);
    }

    @Override
    public boolean isAvailable() {
        return true;
    }

    @Override
    public boolean supportsOffline() {
        return true;
    }
}
```

Register custom tool:

```java
ToolRegistry registry = agent.getToolRegistry();
registry.register(definition, new MyTool());
```

## Built-in Offline Skills

The SDK includes several offline-capable skills:

| Skill | ID | Description |
|-------|-----|-------------|
| Echo | `echo` | Echo input for testing |
| Text Process | `text.process` | Text operations (uppercase, lowercase, reverse, length) |
| Calculator | `calculator` | Math operations (add, sub, mul, div, sqrt, sin, cos, etc.) |
| Timestamp | `timestamp` | Time formatting and conversion |
| JSON Format | `json.format` | JSON beautification and validation |
| Hash | `hash` | Hash calculation (MD5, SHA-1, SHA-256, SHA-512) |

Register offline skills:

```java
OfflineSkills.registerAll(offlineManager);
```

## Offline Support

| Level | Description | Tool Availability |
|-------|-------------|-------------------|
| L1 | Complete offline | Offline-capable tools only |
| L2 | LAN collaboration | Tools + P2P device access |
| L3 | Weak network | Cached requests |
| L4 | Online | Full access |

```java
// Set offline mode
agent.setOfflineMode(true);

// Check offline status
if (agent.isOfflineMode()) {
    // Use offline-capable tools only
}
```

## Permissions

Add to `AndroidManifest.xml`:

```xml
<!-- Required -->
<uses-permission android:name="android.permission.INTERNET" />
<uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />

<!-- Optional - for specific tools -->
<uses-permission android:name="android.permission.CAMERA" />
<uses-permission android:name="android.permission.BLUETOOTH" />
<uses-permission android:name="android.permission.BLUETOOTH_ADMIN" />
<uses-permission android:name="android.permission.ACCESS_FINE_LOCATION" />
<uses-permission android:name="android.permission.READ_CONTACTS" />
<uses-permission android:name="android.permission.READ_CALENDAR" />
<uses-permission android:name="android.permission.READ_EXTERNAL_STORAGE" />
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      Android Device                              │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                  OFA Agent SDK v1.0.2                        ││
│  │  ┌─────────────────────────────────────────────────────────┐││
│  │  │              Intent Understanding Layer                  │││
│  │  │  ┌───────────┐ ┌───────────┐ ┌───────────┐ ┌───────────┐ │││
│  │  │  │  Intent   │ │  Intent   │ │   Tool    │ │   Task    │ │││
│  │  │  │  Engine   │ │ Registry  │ │  Mapper   │ │ Executor  │ │││
│  │  │  └─────┬─────┘ └─────┬─────┘ └─────┬─────┘ └─────┬─────┘ │││
│  │  └────────│─────────────│─────────────│─────────────│───────┘││
│  │           │             │             │             │        ││
│  │  ┌────────┴─────────────┴─────────────┴─────────────┴───────┐││
│  │  │                    LLM Provider Layer                     │││
│  │  │  ┌───────────────┐           ┌───────────────────────┐   │││
│  │  │  │ CloudLLM      │           │     LocalLLM          │   │││
│  │  │  │ (OpenAI API)  │◄─failover─┤  (TensorFlow Lite)    │   │││
│  │  │  └───────────────┘           └───────────────────────┘   │││
│  │  └──────────────────────────────────────────────────────────┘││
│  │  ┌──────────────────────────────────────────────────────────┐││
│  │  │               MCP Server + Tool Registry                 │││
│  │  │  System │ Device │ Data │ AI Tools (30+ built-in)       │││
│  │  └──────────────────────────────────────────────────────────┘││
│  │  ┌──────────────────────────────────────────────────────────┐││
│  │  │              Offline Execution Layer                     │││
│  │  └──────────────────────────────────────────────────────────┘││
│  └──────────────────────────────────────────────────────────────┘│
└──────────────────────────────────────────────────────────────────┘
                         │
                         │ gRPC
                         ▼
                ┌───────────────┐
                │  OFA Center   │
                │ (gRPC:9090)   │
                └───────────────┘
```

## API Reference

### OFAAgent

| Method | Description |
|--------|-------------|
| `connect()` | Connect to Center |
| `disconnect()` | Disconnect from Center |
| `shutdown()` | Shutdown agent completely |
| `isConnected()` | Check connection status |
| `getMCPServer()` | Get MCP Server instance |
| `getToolRegistry()` | Get Tool Registry |
| `getAIAgentInterface()` | Get AI Agent interface |
| `callTool(name, args)` | Call a tool |
| `getAvailableTools()` | List available tools |
| `setOfflineMode(offline)` | Set offline mode |
| `isOfflineMode()` | Check offline status |

### TaskExecutor

| Method | Description |
|--------|-------------|
| `process(input)` | Process user input string |
| `process(input, callback, ctx)` | Process with callback |
| `confirmAndExecute(taskId, intent, mapping, ctx)` | Execute after confirmation |
| `fillSlotsAndExecute(taskId, intent, mapping, slots, ctx)` | Execute with filled slots |
| `registerIntent(definition)` | Register custom intent |
| `registerMapping(intentId, toolName, ...)` | Register intent-tool mapping |
| `getIntentEngine()` | Get IntentEngine instance |
| `getIntentRegistry()` | Get IntentRegistry instance |

### IntentEngine

| Method | Description |
|--------|-------------|
| `recognize(input)` | Recognize intents from input |
| `recognizeBest(input)` | Get best matching intent |
| `recognizeInCategory(input, category)` | Recognize in specific category |
| `register(definition)` | Register intent definition |
| `getAllDefinitions()` | Get all registered definitions |

### IntentDefinition.Builder

| Method | Description |
|--------|-------------|
| `id(id)` | Set intent ID |
| `category(category)` | Set category (system, media, etc.) |
| `action(action)` | Set action name |
| `keywords(...)` | Add matching keywords |
| `pattern(regex)` | Add regex pattern |
| `slot(name, type, desc, required)` | Add slot definition |
| `slotWithPattern(name, type, pattern, required)` | Add slot with extraction pattern |
| `defaultConfidence(confidence)` | Set default confidence |
| `build()` | Build definition |

### ToolExecutor

| Method | Description |
|--------|-------------|
| `getToolId()` | Return tool ID |
| `getDescription()` | Return description |
| `execute(args, ctx)` | Execute tool |
| `isAvailable()` | Check availability |
| `supportsOffline()` | Check offline support |

### ToolResult

| Field | Description |
|-------|-------------|
| `success` | Execution success |
| `toolName` | Tool name |
| `output` | JSON output |
| `error` | Error message |
| `executionTimeMs` | Execution time |

## Testing

Run build test:

```bash
# Windows
test_build.bat

# Linux/Mac
./test_build.sh
```

## Documentation

- [Skill System Guide](docs/SKILL_SYSTEM.md) - 创建自定义技能
- [Intent Understanding System](docs/INTENT_SYSTEM.md) - 意图识别
- [MCP Tools Guide](docs/MCP_TOOLS_GUIDE.md) - 工具使用
- [LLM Integration Plan](docs/LLM_INTEGRATION_PLAN.md) - LLM集成
- [Changelog](CHANGELOG.md)

## License

MIT License