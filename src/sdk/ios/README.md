# OFA iOS SDK

Version: 8.1.0

OFA iOS SDK provides multi-platform support for iPhone, iPad, Mac, and Apple Watch.

## Features

- **Multi-platform Support**: iPhone, iPad, Mac, Apple Watch
- **WebSocket Connection**: Connect to OFA Center server
- **Identity Management**: Personal identity sync and behavior observation
- **Scene Detection**: Detect user context from device sensors
- **Audio Playback**: TTS and voice streaming playback
- **Swift Concurrency**: Full async/await support

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

### Basic Setup

```swift
import OFA

// Create configuration
let config = AgentConfig(
    centerAddress: "ws://your-center:8080/ws",
    mode: .sync
)

// Create agent
let agent = OFAiOSAgent(config: config)

// Initialize
try await agent.initialize()

// Connect to Center
try await agent.connectCenter()
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
            Button("Connect") {
                Task { try? await agent.connectCenter() }
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

### Platform Detection

Device types supported:
- iPhone
- iPad  
- Mac
- Watch

## License

MIT License
