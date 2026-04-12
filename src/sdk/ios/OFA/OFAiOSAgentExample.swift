// OFAiOSAgentExample.swift
// OFA iOS SDK - Example Usage (v8.1.0)
// Demonstrates how to use OFA iOS SDK in your app

import Foundation
import SwiftUI
import OFA

// MARK: - Example App

@main
struct OFAExampleApp: App {
    @StateObject private var agent = OFAiOSAgent(
        config: AgentConfig(
            centerAddress: "ws://localhost:8080/ws",
            mode: .sync
        )
    )

    var body: some Scene {
        WindowGroup {
            OFAExampleView(agent: agent)
        }
    }
}

// MARK: - Example View

struct OFAExampleView: View {
    @ObservedObject var agent: OFAiOSAgent

    var body: some View {
        VStack(spacing: 20) {
            // Status Section
            VStack {
                Text("OFA iOS SDK Example")
                    .font(.largeTitle)
                    .fontWeight(.bold)

                HStack {
                    Circle()
                        .fill(statusColor)
                        .frame(width: 20, height: 20)
                    Text("Status: \(agent.status.rawValue)")
                }

                if agent.connectedToCenter {
                    Text("Connected to Center")
                        .foregroundColor(.green)
                } else {
                    Text("Disconnected")
                        .foregroundColor(.red)
                }
            }
            .padding()

            Divider()

            // Device Info Section
            VStack(alignment: .leading, spacing: 8) {
                Text("Device Information")
                    .font(.headline)

                Text("Agent ID: \(agent.profile.agentId)")
                Text("Device Type: \(agent.profile.deviceType.rawValue)")
                Text("Device Name: \(agent.profile.deviceName)")
                Text("OS Version: \(agent.profile.osVersion)")
                Text("Capabilities: \(agent.profile.capabilities.joined(separator: ", "))")
            }
            .padding()

            Divider()

            // Actions Section
            VStack(spacing: 12) {
                Text("Actions")
                    .font(.headline)

                Button("Initialize Agent") {
                    Task {
                        try? await agent.initialize()
                    }
                }

                Button("Connect to Center") {
                    Task {
                        try? await agent.connectCenter()
                    }
                }
                .disabled(agent.connectedToCenter)

                Button("Disconnect") {
                    Task {
                        await agent.disconnect()
                    }
                }
                .disabled(!agent.connectedToCenter)

                Button("Sync with Center") {
                    Task {
                        try? await agent.syncWithCenter()
                    }
                }
                .disabled(!agent.connectedToCenter)
            }
            .padding()

            Divider()

            // Identity Section
            if let identityId = agent.currentIdentityId {
                VStack {
                    Text("Current Identity: \(identityId)")
                        .font(.subheadline)
                }
                .padding()
            }
        }
        .padding()
        .onAppear {
            Task {
                try? await agent.initialize()
            }
        }
    }

    private var statusColor: Color {
        switch agent.status {
        case .online: return .green
        case .offline: return .gray
        case .busy: return .orange
        case .error: return .red
        }
    }
}

// MARK: - Scene Listener Example

class ExampleSceneListener: SceneListener {
    func onSceneStart(_ scene: SceneState) {
        print("Scene started: \(scene.type.rawValue)")
        print("Confidence: \(scene.confidence)")

        // Handle scene-specific actions
        for action in scene.actions {
            print("Action: \(action.type)")
            handleSceneAction(action)
        }
    }

    func onSceneEnd(_ scene: SceneState) {
        print("Scene ended: \(scene.type.rawValue)")
    }

    func onSceneAction(_ scene: SceneState, action: SceneAction) {
        print("Scene action triggered: \(action.type)")
        handleSceneAction(action)
    }

    private func handleSceneAction(_ action: SceneAction) {
        switch action.type {
        case "route":
            print("Routing message to: \(action.targetAgent ?? "unknown")")
        case "filter":
            print("Applying filter")
        case "block":
            print("Blocking notifications")
        case "alert":
            print("Sending alert")
        default:
            break
        }
    }
}

// MARK: - Usage Examples

/// Example: Initialize and connect
func exampleBasicUsage() async throws {
    // Create configuration
    let config = AgentConfig(
        agentId: "my_custom_agent_id",
        centerAddress: "ws://your-center-server:8080/ws",
        mode: .sync,
        heartbeatInterval: 30.0,
        syncInterval: 300.0
    )

    // Create agent
    let agent = OFAiOSAgent(config: config)

    // Initialize
    try await agent.initialize()

    // Connect to Center
    try await agent.connectCenter()

    // Get status
    let status = agent.getStatus()
    print("Agent status: \(status.rawValue)")
}

/// Example: Scene detection
func exampleSceneDetection() async {
    let sceneDetector = SceneDetector()

    // Initialize
    sceneDetector.initialize()

    // Add listener
    let listener = ExampleSceneListener()
    sceneDetector.addListener(listener)

    // Manual detection
    await sceneDetector.detect()

    // Get active scene
    if let scene = sceneDetector.getActiveScene() {
        print("Active scene: \(scene.type.rawValue)")
    }

    // Stop detection when done
    sceneDetector.stopDetection()
}

/// Example: Audio playback
func exampleAudioPlayback() {
    let audioPlayer = AudioPlayer()

    // Initialize
    audioPlayer.initialize()

    // Play audio data
    let audioData = Data()  // Your PCM audio data
    audioPlayer.play(audioData)

    // Or use streaming
    audioPlayer.playStream()
    audioPlayer.queueAudio(audioData)

    // Control playback
    audioPlayer.pause()
    audioPlayer.resume()
    audioPlayer.stop()

    // Set volume
    audioPlayer.setVolume(0.5)

    // Clean up
    audioPlayer.release()
}

/// Example: Identity management
func exampleIdentityManagement() async throws {
    let identityManager = IdentityManager()

    // Initialize
    try await identityManager.initialize()

    // Create new identity
    let identity = try await identityManager.createIdentity(name: "User")

    // Update identity
    var updated = identity
    updated.nickname = "My Nickname"
    try await identityManager.updateIdentity(updated)

    // Observe behavior
    identityManager.observeBehavior(
        type: "decision",
        context: ["impulse_purchase": true]
    )

    // Trigger personality inference
    await identityManager.inferPersonality()

    // Get decision context
    if let context = identityManager.getDecisionContext() {
        print("Personality openness: \(context.personality.openness)")
    }
}

/// Example: Audio stream receiver
func exampleAudioStreamReceiver() {
    let audioPlayer = AudioPlayer()
    audioPlayer.initialize()

    let receiver = AudioStreamReceiver(audioPlayer: audioPlayer)

    // Handle stream start
    receiver.handleStreamStart(
        streamId: "stream_123",
        format: "pcm",
        sampleRate: 24000
    )

    // Handle audio chunks (typically received via WebSocket)
    let chunkData = Data()  // Your audio chunk
    receiver.handleStreamChunk(streamId: "stream_123", audioData: chunkData)

    // Handle stream end
    receiver.handleStreamEnd(streamId: "stream_123")

    // Get full stream data
    let fullData = receiver.getStreamData()
}

// MARK: - Preview

#if DEBUG
struct OFAExampleView_Previews: PreviewProvider {
    static var previews: some View {
        OFAExampleView(
            agent: OFAiOSAgent(
                config: AgentConfig(
                    centerAddress: nil,
                    mode: .standalone
                )
            )
        )
    }
}
#endif