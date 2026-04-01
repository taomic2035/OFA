# OFA Android SDK - MCP/Tool Integration

## Overview

The OFA Android SDK now supports MCP (Model Context Protocol) for AI agent tool interactions. This enables:

- **Local Tool Execution**: Call device tools directly without network
- **Offline Operation**: Tools work in L1-L4 offline levels
- **AI Agent Integration**: OpenAI-compatible function calling interface

## Quick Start

### 1. Initialize OFA Agent

```java
OFAAgent agent = new OFAAgent.Builder(context)
    .agentId("my-agent-001")
    .agentName("My Android Agent")
    .type(OFAAgent.AgentType.MOBILE)
    .centerAddress("192.168.1.100")
    .centerPort(9090)
    .offlineLevel(OfflineLevel.L4)
    .enableTools(true)  // Enable MCP/Tool support
    .build();

agent.connect();
```

### 2. Call Tools Directly

```java
// Call camera tool
Map<String, Object> args = new HashMap<>();
args.put("operation", "capture");
args.put("cameraId", "0");

ToolResult result = agent.callTool("camera.capture", args);

if (result.isSuccess()) {
    JSONObject output = result.getOutput();
    String savedPath = output.getString("savedPath");
    Log.i("Camera", "Photo saved to: " + savedPath);
}
```

### 3. Use with AI Agent

```java
AIAgentInterface aiInterface = agent.getAIAgentInterface();

// Get tools as OpenAI functions
JSONArray functions = aiInterface.getToolsAsFunctions();

// After AI returns function call:
ToolCallingAdapter.FunctionCallInfo callInfo = adapter.parseFunctionCall(aiResponse);
if (callInfo != null) {
    ToolResult result = aiInterface.callTool(callInfo.toolName, callInfo.arguments);
}
```

## Available Tools

### System Tools

| Tool | Description | Offline |
|------|-------------|---------|
| `app.launch` | Launch application by package name | ✅ |
| `app.list` | List installed applications | ✅ |
| `app.info` | Get application information | ✅ |
| `settings.get` | Get device setting value | ✅ |
| `settings.set` | Set device setting value | ✅ |
| `clipboard.read` | Read clipboard content | ✅ |
| `clipboard.write` | Write text to clipboard | ✅ |
| `file.read` | Read file content | ✅ |
| `file.write` | Write content to file | ✅ |
| `file.list` | List files in directory | ✅ |
| `notification.send` | Send system notification | ✅ |

### Device Tools

| Tool | Description | Offline | Permissions |
|------|-------------|---------|-------------|
| `camera.capture` | Capture photo | ✅ | CAMERA |
| `camera.scan` | Scan QR/barcode | ✅ | CAMERA |
| `camera.list` | List available cameras | ✅ | - |
| `bluetooth.scan` | Scan Bluetooth devices | ✅ | BLUETOOTH |
| `bluetooth.list` | List paired devices | ✅ | BLUETOOTH |
| `bluetooth.status` | Get Bluetooth status | ✅ | - |
| `wifi.scan` | Scan WiFi networks | ✅ | LOCATION |
| `wifi.list` | List configured networks | ✅ | LOCATION |
| `wifi.status` | Get WiFi connection info | ✅ | - |
| `nfc.status` | Get NFC adapter status | ✅ | - |
| `sensor.list` | List available sensors | ✅ | - |
| `sensor.read` | Read sensor values | ✅ | - |
| `battery.status` | Get battery status | ✅ | - |

### Data Tools

| Tool | Description | Offline | Permissions |
|------|-------------|---------|-------------|
| `contacts.query` | Query contacts list | ✅ | READ_CONTACTS |
| `contacts.search` | Search contacts | ✅ | READ_CONTACTS |
| `contacts.count` | Get contacts count | ✅ | READ_CONTACTS |
| `calendar.query` | Query calendar events | ✅ | READ_CALENDAR |
| `calendar.calendars` | List calendars | ✅ | READ_CALENDAR |
| `calendar.today` | Get today's events | ✅ | READ_CALENDAR |
| `media.images` | Query image files | ✅ | STORAGE |
| `media.videos` | Query video files | ✅ | STORAGE |
| `media.audio` | Query audio files | ✅ | STORAGE |

### AI Tools

| Tool | Description | Offline |
|------|-------------|---------|
| `speech.synthesize` | Text-to-speech synthesis | ✅ |
| `speech.stop` | Stop current speech | ✅ |
| `speech.status` | Get TTS status | ✅ |

## Offline Levels

| Level | Description | Tool Availability |
|-------|-------------|-------------------|
| L1 | Complete offline | Only offline-capable tools |
| L2 | LAN collaboration | Tools + P2P device access |
| L3 | Weak network | Cached requests, sync later |
| L4 | Online | Full tool access |

## Constraint Checking

Tools automatically check constraints before execution:

```java
// Financial operations require online mode
if (offlineMode && tool.hasConstraint(ConstraintType.FINANCIAL)) {
    // Tool blocked
}

// Privacy-sensitive data requires auth
if (tool.hasConstraint(ConstraintType.PRIVACY)) {
    // Request user authorization
}
```

## Permission Handling

```java
// Check permissions before calling tool
PermissionManager pm = new PermissionManager(activity);

if (!pm.checkPermissions(new String[]{Manifest.permission.CAMERA})) {
    pm.requestPermissions(new String[]{Manifest.permission.CAMERA},
        REQUEST_CODE, new PermissionManager.PermissionCallback() {
            @Override
            public void onGranted(String[] permissions) {
                // Call camera tool
            }

            @Override
            public void onDenied(String[] permissions, boolean[] grantResults) {
                // Handle denial
            }
        });
}
```

## AI Agent Integration Example

```java
// Initialize
OFAAgent agent = new OFAAgent.Builder(context)
    .enableTools(true)
    .build();

AIAgentInterface ai = agent.getAIAgentInterface();

// Get tool suggestions based on context
List<ToolSuggestion> suggestions = ai.suggestTools("Take a photo of the document");
// Returns: camera.capture with high confidence

// Convert to OpenAI function format
JSONArray functions = ai.getToolsAsFunctions();

// Send to AI model with function calling support
// ... your AI API call here ...

// Parse and execute function call
JSONObject aiResponse = getAIResponse();
FunctionCallInfo callInfo = ((ToolCallingAdapter) ai).parseFunctionCall(aiResponse);

if (callInfo != null) {
    ToolResult result = ai.callTool(callInfo.toolName, callInfo.arguments);
    // Return result to AI for continued conversation
}
```

## Best Practices

1. **Permission Requests**: Request permissions before using tools that require them
2. **Offline Handling**: Check `isOfflineMode()` and tool's `offlineCapable` flag
3. **Error Handling**: Always check `result.isSuccess()` before using output
4. **Resource Cleanup**: Call `agent.shutdown()` when done

## File Structure

```
src/main/java/com/ofa/agent/
├── OFAAgent.java              # Main agent class
├── mcp/                        # MCP protocol layer
│   ├── MCPServer.java
│   ├── MCPServerImpl.java
│   ├── MCPClient.java
│   ├── ToolDefinition.java
│   └── MCPProtocol.java
├── tool/                       # Tool layer
│   ├── ToolRegistry.java
│   ├── ToolExecutor.java
│   ├── ToolResult.java
│   ├── ExecutionContext.java
│   ├── PermissionManager.java
│   └── builtin/               # Built-in tools
│       ├── system/
│       ├── device/
│       ├── data/
│       └── ai/
└── ai/                         # AI interface
    ├── AIAgentInterface.java
    └── ToolCallingAdapter.java
```