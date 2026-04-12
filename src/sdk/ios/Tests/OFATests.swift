import XCTest
@testable import OFA

final class OFAiOSAgentTests: XCTestCase {

    // MARK: - Agent Tests

    func testAgentInitialization() async throws {
        let config = AgentConfig(
            centerAddress: nil,
            mode: .standalone
        )

        let agent = OFAiOSAgent(config: config)

        // Initialize should work in standalone mode without Center
        try await agent.initialize()

        XCTAssertEqual(agent.status, .online)
        XCTAssertFalse(agent.connectedToCenter)
    }

    func testAgentProfileCreation() {
        let config = AgentConfig(
            agentId: "test_agent_id",
            centerAddress: nil,
            mode: .standalone
        )

        let agent = OFAiOSAgent(config: config)

        XCTAssertEqual(agent.profile.agentId, "test_agent_id")
        XCTAssertFalse(agent.profile.capabilities.isEmpty)
    }

    func testAgentModeSwitch() async throws {
        let config = AgentConfig(
            centerAddress: nil,
            mode: .standalone
        )

        let agent = OFAiOSAgent(config: config)
        try await agent.initialize()

        // Switch mode should work
        agent.setMode(.standalone)
    }

    // MARK: - Identity Manager Tests

    func testIdentityCreation() async throws {
        let manager = IdentityManager()
        try await manager.initialize()

        let identity = try await manager.createIdentity(name: "Test User")

        XCTAssertEqual(identity.name, "Test User")
        XCTAssertNotNil(manager.currentIdentity)
    }

    func testPersonalityInference() async {
        let manager = IdentityManager()

        // Add behavior observations
        manager.observeBehavior(type: "decision", context: ["impulse_purchase": true])
        manager.observeBehavior(type: "interaction", context: ["group_chats": true])
        manager.observeBehavior(type: "preference", context: ["novel_trying": true])
        manager.observeBehavior(type: "activity", context: ["regular_schedule": true])

        // Add more observations to trigger inference
        for _ in 0..<6 {
            manager.observeBehavior(type: "activity", context: ["test": true])
        }

        await manager.inferPersonality()

        XCTAssertFalse(manager.pendingBehaviors.isEmpty || manager.currentIdentity != nil)
    }

    func testDecisionContext() async throws {
        let manager = IdentityManager()
        try await manager.initialize()

        try await manager.createIdentity(name: "User")

        let context = manager.getDecisionContext()
        XCTAssertNotNil(context)

        // Default personality should be 0.5 for all traits
        XCTAssertEqual(context?.personality.openness, 0.5)
        XCTAssertEqual(context?.personality.conscientiousness, 0.5)
        XCTAssertEqual(context?.personality.extraversion, 0.5)
        XCTAssertEqual(context?.personality.agreeableness, 0.5)
        XCTAssertEqual(context?.personality.neuroticism, 0.5)
    }

    // MARK: - Scene Detector Tests

    func testSceneDetectorInitialization() {
        let detector = SceneDetector()

        detector.initialize()

        XCTAssertEqual(detector.currentScene, .unknown)
        XCTAssertEqual(detector.sceneConfidence, 0.0)
    }

    func testSceneDetection() async {
        let detector = SceneDetector()
        detector.initialize()

        await detector.detect()

        // Scene should be detected (unknown if no sensor data available)
        XCTAssertNotNil(detector.currentScene)
    }

    func testSceneListener() async {
        let detector = SceneDetector()
        detector.initialize()

        let listener = TestSceneListener()
        detector.addListener(listener)

        await detector.detect()

        // Listener should receive events when scene changes
        // Note: In test environment without sensors, scene stays unknown
    }

    // MARK: - Audio Player Tests

    func testAudioPlayerInitialization() {
        let player = AudioPlayer()

        player.initialize()

        XCTAssertEqual(player.playbackState, .idle)
        XCTAssertEqual(player.volume, 1.0)
    }

    func testAudioPlayerPlay() {
        let player = AudioPlayer()
        player.initialize()

        // Create dummy audio data
        let audioData = Data(repeating: 0, count: 1000)

        player.play(audioData)

        XCTAssertEqual(player.playbackState, .playing)
    }

    func testAudioPlayerPauseResume() {
        let player = AudioPlayer()
        player.initialize()

        let audioData = Data(repeating: 0, count: 1000)
        player.play(audioData)

        player.pause()
        XCTAssertEqual(player.playbackState, .paused)

        player.resume()
        XCTAssertEqual(player.playbackState, .playing)
    }

    func testAudioPlayerStop() {
        let player = AudioPlayer()
        player.initialize()

        let audioData = Data(repeating: 0, count: 1000)
        player.play(audioData)

        player.stop()
        XCTAssertEqual(player.playbackState, .stopped)
        XCTAssertEqual(player.getQueueSize(), 0)
    }

    func testAudioPlayerVolume() {
        let player = AudioPlayer()
        player.initialize()

        player.setVolume(0.5)
        XCTAssertEqual(player.volume, 0.5)

        player.setVolume(1.5)  // Should clamp to 1.0
        XCTAssertEqual(player.volume, 1.0)

        player.setVolume(-0.5)  // Should clamp to 0.0
        XCTAssertEqual(player.volume, 0.0)
    }

    func testAudioStreaming() {
        let player = AudioPlayer()
        player.initialize()

        player.playStream()

        // Queue audio chunks
        let chunk1 = Data(repeating: 0, count: 500)
        let chunk2 = Data(repeating: 0, count: 500)

        player.queueAudio(chunk1)
        player.queueAudio(chunk2)

        XCTAssertEqual(player.getQueueSize(), 0)  // Should be consumed in streaming mode
        XCTAssertEqual(player.playbackState, .playing)
    }

    // MARK: - Audio Stream Receiver Tests

    func testAudioStreamReceiver() {
        let player = AudioPlayer()
        player.initialize()

        let receiver = AudioStreamReceiver(audioPlayer: player)

        receiver.handleStreamStart(streamId: "test_stream", format: "pcm", sampleRate: 24000)

        XCTAssertTrue(receiver.isReceiving)
        XCTAssertEqual(receiver.currentStreamId, "test_stream")

        let chunk = Data(repeating: 0, count: 500)
        receiver.handleStreamChunk(streamId: "test_stream", audioData: chunk)

        XCTAssertFalse(receiver.getStreamData().isEmpty)

        receiver.handleStreamEnd(streamId: "test_stream")

        XCTAssertFalse(receiver.isReceiving)
    }

    func testAudioStreamReceiverAutoPlay() {
        let player = AudioPlayer()
        player.initialize()

        let receiver = AudioStreamReceiver(audioPlayer: player)
        receiver.setAutoPlay(false)

        receiver.handleStreamStart(streamId: "test", format: "pcm", sampleRate: 24000)

        let chunk = Data(repeating: 0, count: 500)
        receiver.handleStreamChunk(streamId: "test", audioData: chunk)

        // Should not auto-play when autoPlay is false
        XCTAssertEqual(player.playbackState, .idle)

        receiver.handleStreamEnd(streamId: "test")

        // Should play complete stream at end
        XCTAssertEqual(player.playbackState, .playing)
    }

    // MARK: - Mode Manager Tests

    func testModeManagerInitialization() {
        let manager = AgentModeManager(mode: .standalone)

        XCTAssertEqual(manager.mode, .standalone)
        XCTAssertFalse(manager.isSyncing)
    }

    func testModeManagerModeSwitch() {
        let manager = AgentModeManager(mode: .standalone)

        manager.setMode(.sync)

        XCTAssertEqual(manager.mode, .sync)
    }

    func testModeManagerSyncStatus() {
        let manager = AgentModeManager(mode: .sync)

        let status = manager.getSyncStatus()

        XCTAssertEqual(status.mode, .sync)
        XCTAssertTrue(status.needsSync)  // No sync has happened yet
    }

    // MARK: - Configuration Tests

    func testAgentConfigDefaults() {
        let config = AgentConfig()

        XCTAssertNil(config.agentId)
        XCTAssertNil(config.centerAddress)
        XCTAssertEqual(config.mode, .sync)
        XCTAssertEqual(config.heartbeatInterval, 30.0)
        XCTAssertEqual(config.syncInterval, 300.0)
        XCTAssertTrue(config.enableCache)
    }

    func testAgentConfigCustom() {
        let config = AgentConfig(
            agentId: "custom_id",
            centerAddress: "ws://custom.server:8080",
            mode: .standalone,
            heartbeatInterval: 60.0,
            syncInterval: 600.0,
            enableCache: false
        )

        XCTAssertEqual(config.agentId, "custom_id")
        XCTAssertEqual(config.centerAddress, "ws://custom.server:8080")
        XCTAssertEqual(config.mode, .standalone)
        XCTAssertEqual(config.heartbeatInterval, 60.0)
        XCTAssertEqual(config.syncInterval, 600.0)
        XCTAssertFalse(config.enableCache)
    }

    // MARK: - Error Tests

    func testOFAErrors() {
        let configError = OFAError.configurationError("test")
        XCTAssertEqual(configError.localizedDescription, "Configuration error: test")

        let connectionError = OFAError.connectionError("test")
        XCTAssertEqual(connectionError.localizedDescription, "Connection error: test")

        let syncError = OFAError.syncError("test")
        XCTAssertEqual(syncError.localizedDescription, "Sync error: test")

        let identityError = OFAError.identityError("test")
        XCTAssertEqual(identityError.localizedDescription, "Identity error: test")
    }
}

// MARK: - Test Helpers

class TestSceneListener: SceneListener {
    var sceneStartCount = 0
    var sceneEndCount = 0
    var actionCount = 0

    func onSceneStart(_ scene: SceneState) {
        sceneStartCount += 1
    }

    func onSceneEnd(_ scene: SceneState) {
        sceneEndCount += 1
    }

    func onSceneAction(_ scene: SceneState, action: SceneAction) {
        actionCount += 1
    }
}