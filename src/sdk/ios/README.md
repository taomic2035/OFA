# OFA iOS Agent SDK

iOS SDK for building OFA Agent applications.

## Requirements

- iOS 13.0+ / macOS 12.0+
- Swift 5.9+
- Xcode 15.0+

## Installation

### Swift Package Manager

Add to your `Package.swift`:

```swift
dependencies: [
    .package(url: "https://github.com/ofa/ios-agent-sdk.git", from: "1.0.0")
]
```

Or in Xcode:
1. File → Add Packages
2. Enter package URL: `https://github.com/ofa/ios-agent-sdk`

## Quick Start

### 1. Initialize Agent

```swift
import OFAAgent

let agent = OFAAgent(
    agentId: "my-ios-agent",
    name: "My iPhone",
    type: .mobile,
    centerAddress: "192.168.1.100",
    centerPort: 9090
)
```

### 2. Register Skills

```swift
// Register built-in skills
agent.registerSkill(EchoSkill())
agent.registerSkill(TextProcessSkill())
agent.registerSkill(CalculatorSkill())

// Register custom skill
agent.registerSkill(MyCustomSkill())
```

### 3. Set Delegates

```swift
class MyViewController: UIViewController, OFAAgentDelegate, OFAAgentTaskDelegate {
    func agent(_ agent: OFAAgent, didChangeConnectionState state: OFAAgent.ConnectionState) {
        switch state {
        case .connected:
            print("Connected to Center")
        case .disconnected:
            print("Disconnected from Center")
        case .error(let error):
            print("Connection error: \(error)")
        default:
            break
        }
    }

    func agent(_ agent: OFAAgent, didReceiveTask taskId: String, skillId: String) {
        print("Received task: \(taskId)")
    }

    func agent(_ agent: OFAAgent, didCompleteTask taskId: String) {
        print("Task completed: \(taskId)")
    }

    func agent(_ agent: OFAAgent, didFailTask taskId: String, error: Error) {
        print("Task failed: \(error)")
    }
}
```

### 4. Connect to Center

```swift
Task {
    do {
        try await agent.connect()
    } catch {
        print("Connection failed: \(error)")
    }
}
```

### 5. Disconnect

```swift
agent.disconnect()
```

## Custom Skills

Implement `SkillExecutor` protocol:

```swift
import OFAAgent

class MyCustomSkill: SkillExecutor, @unchecked Sendable {
    var skillId: String { "my.custom.skill" }
    var skillName: String { "My Custom Skill" }
    var category: String { "utility" }

    func execute(_ input: Data) async throws -> Data {
        // Parse input
        guard let json = try JSONSerialization.jsonObject(with: input) as? [String: Any] else {
            throw OFAError.executionFailed("Invalid input")
        }

        // Process
        let result = processInput(json)

        // Return output
        let output = ["result": result]
        return try JSONSerialization.data(withJSONObject: output)
    }

    private func processInput(_ input: [String: Any]) -> String {
        // Your processing logic
        return "processed"
    }
}
```

## Built-in Skills

| Skill ID | Description | Operations |
|----------|-------------|------------|
| echo | Echo test | Returns input with length |
| text.process | Text processing | uppercase, lowercase, reverse, length |
| calculator | Calculator | add, sub, mul, div, pow, sqrt |

## Background Mode

For background operation, enable background modes:

1. Project → Target → Signing & Capabilities
2. Add Capability → Background Modes
3. Check "Background fetch" and "Remote notifications"

```swift
// In AppDelegate
func application(_ application: UIApplication,
                 didRegisterForRemoteNotificationsWithDeviceToken deviceToken: Data) {
    // Handle push notifications for task alerts
}
```

## Architecture

```
┌─────────────────────────────────────┐
│           iOS Device                 │
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

| Property | Type | Description |
|----------|------|-------------|
| `agentId` | String | Agent identifier |
| `name` | String | Agent name |
| `type` | AgentType | Agent type |
| `connectionState` | ConnectionState | Current connection state |
| `heartbeatInterval` | TimeInterval | Heartbeat interval (seconds) |

| Method | Description |
|--------|-------------|
| `connect()` | Connect to Center |
| `disconnect()` | Disconnect from Center |
| `registerSkill(_:)` | Register a skill |
| `unregisterSkill(_:)` | Unregister a skill |
| `getRegisteredSkills()` | Get registered skill IDs |

### SkillExecutor

| Property | Description |
|----------|-------------|
| `skillId` | Unique skill identifier |
| `skillName` | Human-readable name |
| `category` | Skill category |

| Method | Description |
|--------|-------------|
| `execute(_:)` | Execute skill with input data |

## License

MIT License