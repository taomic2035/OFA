# OFA Android Agent SDK

Android SDK for building OFA Agent applications.

## Requirements

- Android SDK 24+ (Android 7.0)
- Java 17
- Gradle 8.2+

## Installation

Add to your app's `build.gradle`:

```gradle
dependencies {
    implementation 'com.ofa:agent-sdk:0.5.0'
}
```

## Quick Start

### 1. Initialize Agent

```java
OFAAgent agent = new OFAAgent.Builder(context)
    .agentId("my-android-agent")
    .agentName("My Phone")
    .centerAddress("192.168.1.100")
    .centerPort(9090)
    .type(OFAAgent.AgentType.MOBILE)
    .build();
```

### 2. Register Skills

```java
// Register built-in skills
agent.registerSkill("echo", new EchoSkill());
agent.registerSkill("text.process", new TextProcessSkill());

// Register custom skill
agent.registerSkill("my.skill", new MyCustomSkill());
```

### 3. Connect to Center

```java
agent.setConnectionListener(new OFAAgent.ConnectionListener() {
    @Override
    public void onConnected() {
        Log.i("Agent", "Connected to Center");
    }

    @Override
    public void onDisconnected() {
        Log.i("Agent", "Disconnected from Center");
    }

    @Override
    public void onError(String message) {
        Log.e("Agent", "Error: " + message);
    }
});

agent.connect();
```

### 4. Disconnect

```java
agent.disconnect();
```

## Custom Skills

Implement `SkillExecutor` interface:

```java
public class MyCustomSkill implements SkillExecutor {

    @Override
    public String getSkillId() {
        return "my.custom.skill";
    }

    @Override
    public String getSkillName() {
        return "My Custom Skill";
    }

    @Override
    public String getCategory() {
        return "utility";
    }

    @Override
    public byte[] execute(byte[] input) throws SkillExecutionException {
        // Parse input (JSON format)
        String jsonInput = new String(input, StandardCharsets.UTF_8);

        // Process
        String result = processInput(jsonInput);

        // Return output (JSON format)
        JsonObject output = new JsonObject();
        output.addProperty("result", result);
        return output.toString().getBytes(StandardCharsets.UTF_8);
    }
}
```

## Built-in Skills

| Skill ID | Description | Operations |
|----------|-------------|------------|
| echo | Echo test | Returns input with length |
| text.process | Text processing | uppercase, lowercase, reverse, length |

## Permissions

Add to `AndroidManifest.xml`:

```xml
<uses-permission android:name="android.permission.INTERNET" />
<uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
<uses-permission android:name="android.permission.FOREGROUND_SERVICE" />
```

## Background Service

For background operation, use Android WorkManager:

```java
// Start periodic heartbeat
Constraints constraints = new Constraints.Builder()
    .setRequiredNetworkType(NetworkType.CONNECTED)
    .build();

PeriodicWorkRequest heartbeatWork = new PeriodicWorkRequest.Builder(
    AgentHeartbeatWorker.class,
    15, // minutes
    TimeUnit.MINUTES
)
    .setConstraints(constraints)
    .build();

WorkManager.getInstance(context).enqueueUniquePeriodicWork(
    "agent_heartbeat",
    ExistingPeriodicWorkPolicy.KEEP,
    heartbeatWork
);
```

## Architecture

```
┌─────────────────────────────────────┐
│           Android Device             │
│  ┌─────────────────────────────────┐│
│  │        OFA Agent SDK            ││
│  │  ┌───────────┐ ┌──────────────┐ ││
│  │  │  Agent    │ │   Skill      │ ││
│  │  │  Core     │ │   Executors  │ ││
│  │  └─────┬─────┘ └──────────────┘ ││
│  │        │                        ││
│  │  ┌─────┴─────┐                  ││
│  │  │  gRPC     │                  ││
│  │  │  Client   │                  ││
│  │  └─────┬─────┘                  ││
│  └────────┼────────────────────────┘│
└───────────┼──────────────────────────┘
            │
            │ gRPC
            ▼
    ┌───────────────┐
    │ OFA Center    │
    │ (gRPC:9090)   │
    └───────────────┘
```

## API Reference

### OFAAgent

| Method | Description |
|--------|-------------|
| `connect()` | Connect to Center |
| `disconnect()` | Disconnect from Center |
| `isConnected()` | Check connection status |
| `registerSkill(id, executor)` | Register a skill |
| `unregisterSkill(id)` | Unregister a skill |
| `getRegisteredSkills()` | Get registered skill IDs |
| `getAgentId()` | Get agent ID |

### SkillExecutor

| Method | Description |
|--------|-------------|
| `getSkillId()` | Return skill ID |
| `getSkillName()` | Return skill name |
| `getCategory()` | Return skill category |
| `execute(input)` | Execute skill with input |

## License

MIT License