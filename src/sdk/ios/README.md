# OFA iOS SDK

Version: 9.8.0

OFA iOS SDK provides multi-platform support for iPhone, iPad, Mac, and Apple Watch.
Aligned with Android SDK architecture for consistent cross-platform behavior.

## Features

### Core Features (v8.1.0)
- **Multi-platform Support**: iPhone, iPad, Mac, Apple Watch
- **WebSocket Connection**: Connect to OFA Center server
- **Identity Management**: Personal identity sync and behavior observation
- **Scene Detection**: Detect user context from device sensors
- **Audio Playback**: TTS and voice streaming playback
- **Swift Concurrency**: Full async/await support

### New Features (v9.8.0)
- **Error Handling Framework**: CircuitBreaker, RetryExecutor, FallbackProvider
- **Distributed Components**: EventBus, CrossDeviceRouter, DistributedOrchestrator
- **Memory System**: L1/L2/L3/Archive multi-level memory storage
- **Connection Recovery**: Automatic reconnection with exponential backoff
- **Graceful Degradation**: Fallback mechanisms for offline scenarios

## Requirements

- iOS 15.0+ / macOS 12.0+ / watchOS 8.0+
- Swift 5.7+
- Xcode 14.0+

## Installation

### Swift Package Manager

Add the following to your Package.swift dependencies:

```swift
dependencies: [
    .package(path: "path/to/OFA")
]
```

Or in Xcode:
1. File > Add Packages...
2. Add local package path

### Manual Integration

1. Clone or copy the OFA folder to your project
2. Add the Swift files to your Xcode project target

## Usage

### Basic Setup (v9.8.0)

```swift
import OFA

// Create configuration
let config = AgentConfig(
    centerAddress: "ws://your-center:8080/ws",
    mode: .sync,
    enableCache: true
)

// Create agent
let agent = OFAiOSAgent(config: config)

// Initialize (includes memory and distributed components)
try await agent.initialize()

// Connect to Center with automatic retry
try await agent.connectCenter()
```

### Error Handling (v9.8.0)

```swift
// The agent now includes CircuitBreaker and RetryExecutor
// Connections automatically retry with exponential backoff

// Check connection recovery status
let (isRecovering, circuitState) = agent.getConnectionRecoveryStatus()
print("Recovering: \(isRecovering), Circuit: \(circuitState.rawValue)")

// Handle errors manually
try await agent.syncWithCenter()
// If fails, agent.handleError() will attempt recovery

// Use RetryExecutor for custom operations
let executor = RetryExecutor(config: .aggressiveConfig)
let breaker = CircuitBreaker.defaultBreaker(name: "custom")

let result = try await executor.execute(
    operation: { attempt in
        // Your operation here
        return "success"
    },
    circuitBreaker: breaker
)
```

### Memory System (v9.8.0)

```swift
// Get memory statistics
let stats = agent.getMemoryStats()
print("L1: \(stats.l1Count), L2: \(stats.l2Count), L3: \(stats.l3Count)")
print("Hit Rate: \(stats.averageHitRate)")

// Store and retrieve memories
try await agent.storeMemory("last_task", value: "search_query", category: .task)
let lastTask = await agent.retrieveMemory("last_task")

// Use memory manager directly
let memoryManager = UserMemoryManager()
try await memoryManager.initialize()

let entry = MemoryEntry(
    key: "preference_theme",
    value: "dark",
    level: .l2,
    category: .preference,
    importance: 0.8
)
try await memoryManager.store(entry: entry)
```

### Distributed Components (v9.8.0)

```swift
// Access distributed orchestrator
let orchestrator = agent.getDistributedOrchestrator()

// Publish events
let event = Event(
    type: .sceneChange,
    sourceAgentId: agent.profile.agentId,
    payload: ["scene": "running"]
)
orchestrator.eventBus.publish(event: event)

// Route messages
let message = Message(
    type: "notification",
    sourceAgentId: agent.profile.agentId,
    content: "Hello from iOS"
)
let result = orchestrator.routeMessage(message: message)
print("Routed to: \(result.target.rawValue)")
```

### SwiftUI Integration

```swift
import SwiftUI
import OFA

struct ContentView: View {
    @StateObject private var agent = OFAiOSAgent(
        config: AgentConfig(centerAddress: "ws://localhost:8080/ws")
    )

    var body: some View {
        VStack {
            Text("Status: \(agent.status.rawValue)")
            Text("Scene: \(agent.currentScene.rawValue)")
            Text("Errors: \(agent.errorCount)")

            if agent.connectedToCenter {
                Text("Connected to Center")
            }

            Button("Connect") {
                Task { try? await agent.connectCenter() }
            }

            Button("Sync") {
                Task { try? await agent.syncWithCenter() }
            }
        }
        .onAppear {
            Task { try? await agent.initialize() }
        }
    }
}
```

### Identity Management

```swift
let identityManager = IdentityManager()
try await identityManager.initialize()

// Create identity
let identity = try await identityManager.createIdentity(name: "User")

// Observe behavior
identityManager.observeBehavior(
    type: "decision",
    context: ["impulse_purchase": true]
)

// Trigger personality inference
await identityManager.inferPersonality()
```

### Scene Detection

```swift
let detector = SceneDetector()
detector.initialize()

// Add listener
let listener = MySceneListener()
detector.addListener(listener)

// Manual detection
await detector.detect()
```

### Audio Playback

```swift
let player = AudioPlayer()
player.initialize()

// Play audio data
player.play(audioData)

// Streaming playback
player.playStream()
player.queueAudio(chunk1)
player.queueAudio(chunk2)
```

## Architecture

### Core Components

| Component | Description |
|-----------|-------------|
| OFAiOSAgent | Main entry point for iPhone/iPad/Mac |
| CenterConnection | WebSocket connection to Center |
| IdentityManager | Identity sync and behavior observation |
| SceneDetector | Scene detection from sensors |
| AudioPlayer | Audio playback and streaming |
| AgentModeManager | Mode transitions and sync scheduling |

### New Components (v9.8.0)

| Component | Description |
|-----------|-------------|
| ErrorHandler | Error categorization and notification |
| CircuitBreaker | Prevent cascading failures |
| RetryExecutor | Retry with exponential backoff |
| ConnectionRecoveryManager | Automatic reconnection |
| FallbackProvider | Graceful degradation |
| EventBus | Publish/subscribe for events |
| CrossDeviceRouter | Message routing between devices |
| DistributedOrchestrator | Coordinates distributed components |
| MemoryManager | L1/L2/L3/Archive memory storage |
| ContextMemory | Current operation context |

### Platform Detection

Device types supported:
- iPhone
- iPad  
- Mac
- Watch

### Memory Levels

| Level | Duration | Capacity |
|-------|----------|----------|
| L1 | Seconds | ~10 entries |
| L2 | Minutes | ~100 entries |
| L3 | Hours | ~500 entries |
| Archive | Unlimited | Persistent |

## Alignment with Android SDK

The iOS SDK (v9.8.0) is aligned with Android SDK architecture:
- Same error handling patterns (CircuitBreaker, RetryExecutor)
- Same memory levels (L1/L2/L3/Archive)
- Same distributed components (EventBus, CrossDeviceRouter)
- Same scene detection logic
- Same identity sync protocol

## License

MIT License
