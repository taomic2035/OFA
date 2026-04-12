// AudioPlayer.swift
// OFA iOS SDK - Audio Playback (v8.1.0)

import Foundation
import AVFoundation
import Combine

/// Audio playback state
public enum AudioPlaybackState: String, Codable {
    case idle = "idle"
    case playing = "playing"
    case paused = "paused"
    case stopped = "stopped"
    case error = "error"
}

/// Audio format configuration
public struct AudioFormatConfig {
    public let sampleRate: Double
    public let channels: Int
    public let bitDepth: Int

    public init(sampleRate: Double = 24000, channels: Int = 1, bitDepth: Int = 16) {
        self.sampleRate = sampleRate
        self.channels = channels
        self.bitDepth = bitDepth
    }
}

/// Audio player - handles TTS and voice streaming playback
public class AudioPlayer: ObservableObject {

    // MARK: - Published Properties

    @Published public var playbackState: AudioPlaybackState = .idle
    @Published public var volume: Float = 1.0
    @Published public var currentTime: TimeInterval = 0
    @Published public var duration: TimeInterval = 0

    // MARK: - Private Properties

    private var audioEngine: AVAudioEngine?
    private var playerNode: AVAudioPlayerNode?
    private var audioQueue: [Data] = []
    private var isStreaming: Bool = false

    private var formatConfig: AudioFormatConfig
    private var audioFormat: AVAudioFormat?

    // MARK: - Initialization

    public init(formatConfig: AudioFormatConfig = AudioFormatConfig()) {
        self.formatConfig = formatConfig
    }

    public func initialize() {
        // Create audio engine
        audioEngine = AVAudioEngine()
        playerNode = AVAudioPlayerNode()

        // Create audio format
        audioFormat = AVAudioFormat(
            commonFormat: .pcmFormatInt16,
            sampleRate: formatConfig.sampleRate,
            channels: formatConfig.channels,
            interleaved: true
        )

        // Attach player node
        audioEngine?.attach(playerNode!)

        // Connect to output
        if let format = audioFormat {
            audioEngine?.connect(playerNode!, to: audioEngine!.outputNode, format: format)
        }

        // Prepare engine
        audioEngine?.prepare()
    }

    // MARK: - Public Methods

    /// Play audio data (non-streaming)
    public func play(_ audioData: Data) {
        stop()

        guard let playerNode = playerNode,
              let audioEngine = audioEngine,
              let format = audioFormat else {
            playbackState = .error
            return
        }

        // Convert data to buffer
        let buffer = createBuffer(from: audioData, format: format)

        // Start engine
        try? audioEngine.start()

        // Play buffer
        playerNode.scheduleBuffer(buffer, completionCallbackType: .dataPlayedBack) { _ in
            DispatchQueue.main.async {
                self.playbackState = .idle
                self.stopEngine()
            }
        }

        playerNode.play()
        playbackState = .playing
    }

    /// Play streaming audio
    public func playStream() {
        guard let playerNode = playerNode,
              let audioEngine = audioEngine,
              let format = audioFormat else {
            playbackState = .error
            return
        }

        stop()
        isStreaming = true

        // Start engine
        try? audioEngine.start()

        // Play queued buffers
        playbackState = .playing
        playQueuedBuffers()
    }

    /// Queue audio chunk for streaming
    public func queueAudio(_ audioData: Data) {
        audioQueue.append(audioData)

        if isStreaming && playbackState == .playing {
            scheduleNextBuffer()
        }
    }

    /// Pause playback
    public func pause() {
        playerNode?.pause()
        playbackState = .paused
    }

    /// Resume playback
    public func resume() {
        playerNode?.play()
        playbackState = .playing
    }

    /// Stop playback
    public func stop() {
        playerNode?.stop()
        audioQueue.removeAll()
        isStreaming = false

        stopEngine()
        playbackState = .stopped
    }

    /// Set volume (0-1)
    public func setVolume(_ volume: Float) {
        self.volume = max(0, min(1, volume))
        playerNode?.volume = volume
        audioEngine?.outputNode.volume = volume
    }

    /// Clear audio queue
    public func clearQueue() {
        audioQueue.removeAll()
    }

    /// Get queue size
    public func getQueueSize() -> Int {
        return audioQueue.count
    }

    /// Check if playing
    public func isPlaying() -> Bool {
        return playbackState == .playing
    }

    // MARK: - Private Methods

    private func createBuffer(from data: Data, format: AVAudioFormat) -> AVAudioPCMBuffer {
        let frameCount = UInt32(data.count) / UInt32(format.channelCount) / 2

        let buffer = AVAudioPCMBuffer(pcmFormat: format, frameCapacity: frameCount)!
        buffer.frameLength = frameCount

        // Copy data to buffer
        data.withUnsafeBytes { bytes in
            if let baseAddress = bytes.baseAddress {
                memcpy(buffer.int16ChannelData![0], baseAddress, data.count)
            }
        }

        return buffer
    }

    private func playQueuedBuffers() {
        guard let playerNode = playerNode,
              let format = audioFormat else { return }

        while !audioQueue.isEmpty {
            let data = audioQueue.removeFirst()
            let buffer = createBuffer(from: data, format: format)

            playerNode.scheduleBuffer(buffer, completionCallbackType: .dataPlayedBack) { _ in
                DispatchQueue.main.async {
                    self.onBufferCompleted()
                }
            }
        }
    }

    private func scheduleNextBuffer() {
        guard let playerNode = playerNode,
              let format = audioFormat,
              !audioQueue.isEmpty else { return }

        let data = audioQueue.removeFirst()
        let buffer = createBuffer(from: data, format: format)

        playerNode.scheduleBuffer(buffer, completionCallbackType: .dataPlayedBack) { _ in
            DispatchQueue.main.async {
                self.onBufferCompleted()
            }
        }
    }

    private func onBufferCompleted() {
        if audioQueue.isEmpty && !isStreaming {
            playbackState = .idle
            stopEngine()
        } else if !audioQueue.isEmpty {
            scheduleNextBuffer()
        }
    }

    private func stopEngine() {
        audioEngine?.pause()
        audioEngine?.stop()
    }

    /// Release resources
    public func release() {
        stop()
        audioEngine?.disconnect(playerNode!)
        audioEngine?.detach(playerNode!)
        audioEngine = nil
        playerNode = nil
    }
}

// MARK: - Audio Stream Receiver

/// Audio stream receiver - handles receiving audio from Center
public class AudioStreamReceiver: ObservableObject {

    // MARK: - Published Properties

    @Published public var isReceiving: Bool = false
    @Published public var currentStreamId: String?

    // MARK: - Private Properties

    private var audioPlayer: AudioPlayer
    private var streamData: Data = Data()
    private var autoPlay: Bool = true

    // MARK: - Initialization

    public init(audioPlayer: AudioPlayer) {
        self.audioPlayer = audioPlayer
    }

    // MARK: - Public Methods

    /// Handle audio stream start
    public func handleStreamStart(streamId: String, format: String, sampleRate: Int) {
        currentStreamId = streamId
        isReceiving = true
        streamData = Data()

        // Initialize player with correct format
        audioPlayer.initialize()
    }

    /// Handle audio chunk
    public func handleStreamChunk(streamId: String, audioData: Data) {
        guard currentStreamId == streamId else { return }

        streamData.append(audioData)

        if autoPlay {
            audioPlayer.queueAudio(audioData)
        }
    }

    /// Handle audio stream end
    public func handleStreamEnd(streamId: String) {
        guard currentStreamId == streamId else { return }

        isReceiving = false

        // Play complete stream if not auto-playing
        if !autoPlay || !audioPlayer.isPlaying() {
            audioPlayer.play(streamData)
        }
    }

    /// Set auto-play mode
    public func setAutoPlay(_ autoPlay: Bool) {
        self.autoPlay = autoPlay
    }

    /// Get stream data
    public func getStreamData() -> Data {
        return streamData
    }

    /// Stop receiving
    public func stop() {
        isReceiving = false
        audioPlayer.stop()
        streamData = Data()
        currentStreamId = nil
    }
}