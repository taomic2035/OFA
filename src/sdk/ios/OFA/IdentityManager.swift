// IdentityManager.swift
// OFA iOS SDK - Identity Management (v8.1.0)

import Foundation
import Combine

/// Personal identity model
public struct PersonalIdentity: Codable {
    public let id: String
    public var name: String
    public var nickname: String?
    public var avatar: String?

    // Personality (Big Five)
    public var personality: Personality

    // Values
    public var valueSystem: ValueSystem

    // Interests
    public var interests: [Interest]

    // Voice profile
    public var voiceProfile: VoiceProfile

    // Writing style
    public var writingStyle: WritingStyle

    // Version for sync
    public var version: Int64
    public var updatedAt: Date

    public init(id: String, name: String) {
        self.id = id
        self.name = name
        self.nickname = nil
        self.avatar = nil
        self.personality = Personality()
        self.valueSystem = ValueSystem()
        self.interests = []
        self.voiceProfile = VoiceProfile()
        self.writingStyle = WritingStyle()
        self.version = 1
        self.updatedAt = Date()
    }
}

/// Big Five personality model
public struct Personality: Codable {
    public var openness: Double      // 0-1
    public var conscientiousness: Double
    public var extraversion: Double
    public var agreeableness: Double
    public var neuroticism: Double

    public init(
        openness: Double = 0.5,
        conscientiousness: Double = 0.5,
        extraversion: Double = 0.5,
        agreeableness: Double = 0.5,
        neuroticism: Double = 0.5
    ) {
        self.openness = openness
        self.conscientiousness = conscientiousness
        self.extraversion = extraversion
        self.agreeableness = agreeableness
        self.neuroticism = neuroticism
    }
}

/// Value system
public struct ValueSystem: Codable {
    public var privacy: Double       // 0-1
    public var efficiency: Double
    public var health: Double
    public var family: Double
    public var career: Double

    public init(
        privacy: Double = 0.5,
        efficiency: Double = 0.5,
        health: Double = 0.5,
        family: Double = 0.5,
        career: Double = 0.5
    ) {
        self.privacy = privacy
        self.efficiency = efficiency
        self.health = health
        self.family = family
        self.career = career
    }
}

/// Interest model
public struct Interest: Codable {
    public let category: String
    public let name: String
    public let enthusiasm: Double    // 0-1
    public let keywords: [String]

    public init(category: String, name: String, enthusiasm: Double = 0.5, keywords: [String] = []) {
        self.category = category
        self.name = name
        self.enthusiasm = enthusiasm
        self.keywords = keywords
    }
}

/// Voice profile
public struct VoiceProfile: Codable {
    public var voiceType: String
    public var pitch: Double         // 0-1
    public var speed: Double
    public var tone: String

    public init(voiceType: String = "neutral", pitch: Double = 0.5, speed: Double = 0.5, tone: String = "neutral") {
        self.voiceType = voiceType
        self.pitch = pitch
        self.speed = speed
        self.tone = tone
    }
}

/// Writing style
public struct WritingStyle: Codable {
    public var formality: Double     // 0-1
    public var humor: Double
    public var emojiUsage: Double

    public init(formality: Double = 0.5, humor: Double = 0.5, emojiUsage: Double = 0.3) {
        self.formality = formality
        self.humor = humor
        self.emojiUsage = emojiUsage
    }
}

/// Behavior observation for personality inference
public struct BehaviorObservation: Codable {
    public let id: String
    public let type: String          // decision, interaction, preference, activity
    public let context: [String: AnyCodable]
    public let timestamp: Date

    public init(type: String, context: [String: Any]) {
        self.id = UUID().uuidString
        self.type = type
        self.context = context.mapValues { AnyCodable($0) }
        self.timestamp = Date()
    }
}

/// Identity manager - handles identity sync and behavior observation
public class IdentityManager: ObservableObject {

    // MARK: - Published Properties

    @Published public var currentIdentity: PersonalIdentity?
    @Published public var pendingBehaviors: [BehaviorObservation] = []

    // MARK: - Publishers

    public let identityPublisher: AnyPublisher<String?, Never>

    // MARK: - Private Properties

    private var identitySubject = CurrentValueSubject<String?, Never>(nil)
    private var localStore: LocalIdentityStore
    private var syncService: IdentitySyncService?
    private var inferenceEngine: PersonalityInferenceEngine

    private var cancellables = Set<AnyCancellable>()

    // MARK: - Initialization

    public init() {
        self.localStore = LocalIdentityStore()
        self.inferenceEngine = PersonalityInferenceEngine()

        identityPublisher = identitySubject.eraseToAnyPublisher()
    }

    // MARK: - Public Methods

    /// Initialize identity manager
    public func initialize() async throws {
        // Load from local store
        if let identity = await localStore.load() {
            currentIdentity = identity
            identitySubject.send(identity.id)
        }
    }

    /// Create new identity
    public func createIdentity(name: String) async throws -> PersonalIdentity {
        let identity = PersonalIdentity(id: UUID().uuidString, name: name)

        // Save locally
        await localStore.save(identity)

        currentIdentity = identity
        identitySubject.send(identity.id)

        // Sync to Center if connected
        if let syncService = syncService {
            try await syncService.sync(identity)
        }

        return identity
    }

    /// Update identity
    public func updateIdentity(_ identity: PersonalIdentity) async throws {
        var updated = identity
        updated.version += 1
        updated.updatedAt = Date()

        await localStore.save(updated)
        currentIdentity = updated

        // Sync to Center
        if let syncService = syncService {
            try await syncService.sync(updated)
        }
    }

    /// Observe behavior for personality inference
    public func observeBehavior(type: String, context: [String: Any]) {
        let observation = BehaviorObservation(type: type, context: context)

        // Add to pending behaviors
        pendingBehaviors.append(observation)

        // Trigger inference if enough observations
        if pendingBehaviors.count >= 10 {
            Task {
                await inferPersonality()
            }
        }
    }

    /// Infer personality from behaviors
    public func inferPersonality() async {
        guard let identity = currentIdentity else { return }

        // Run inference
        let inferredPersonality = inferenceEngine.infer(from: pendingBehaviors)

        // Update identity
        var updated = identity
        updated.personality = inferredPersonality
        updated.version += 1

        await localStore.save(updated)
        currentIdentity = updated

        // Clear processed behaviors
        let behaviorsToSync = pendingBehaviors
        pendingBehaviors = []

        // Sync to Center
        if let syncService = syncService {
            try? await syncService.sync(updated)
            try? await syncService.reportBehaviors(behaviorsToSync)
        }
    }

    /// Sync from Center
    public func syncFromCenter() async throws {
        guard let syncService = syncService else { return }

        let identity = try await syncService.fetch()
        await localStore.save(identity)
        currentIdentity = identity
        identitySubject.send(identity.id)
    }

    /// Restore from Center (full download)
    public func restoreFromCenter() async throws {
        guard let syncService = syncService else { return }

        let identity = try await syncService.fetch()
        await localStore.save(identity)
        currentIdentity = identity
        identitySubject.send(identity.id)
    }

    /// Set sync service
    public func setSyncService(_ service: IdentitySyncService) {
        self.syncService = service
    }

    /// Get pending behaviors for sync
    public func getPendingBehaviors() async -> [BehaviorObservation] {
        return pendingBehaviors
    }

    /// Get decision context
    public func getDecisionContext() -> DecisionContext? {
        guard let identity = currentIdentity else { return nil }

        return DecisionContext(
            personality: identity.personality,
            valueSystem: identity.valueSystem,
            voiceProfile: identity.voiceProfile,
            writingStyle: identity.writingStyle
        )
    }
}

// MARK: - Supporting Types

/// Decision context for agent decisions
public struct DecisionContext {
    public let personality: Personality
    public let valueSystem: ValueSystem
    public let voiceProfile: VoiceProfile
    public let writingStyle: WritingStyle
}

/// Local identity storage
public class LocalIdentityStore {

    private let fileName = "identity.json"

    public func load() async -> PersonalIdentity? {
        guard let data = try? Data(contentsOf: getFilePath()) else { return nil }
        return try? JSONDecoder().decode(PersonalIdentity.self, from: data)
    }

    public func save(_ identity: PersonalIdentity) async {
        let encoder = JSONEncoder()
        encoder.outputFormatting = .prettyPrinted

        if let data = try? encoder.encode(identity) {
            try? data.write(to: getFilePath())
        }
    }

    private func getFilePath() -> URL {
        let documents = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask).first!
        return documents.appendingPathComponent(fileName)
    }
}

/// Identity sync service (placeholder)
public class IdentitySyncService {
    public func sync(_ identity: PersonalIdentity) async throws {
        // Implementation would use WebSocket to sync
    }

    public func fetch() async throws -> PersonalIdentity {
        // Implementation would fetch from Center
        throw OFAError.syncError("Not implemented")
    }

    public func reportBehaviors(_ behaviors: [BehaviorObservation]) async throws {
        // Implementation would report to Center
    }
}

/// Personality inference engine
public class PersonalityInferenceEngine {

    public func infer(from behaviors: [BehaviorObservation]) -> Personality {
        var personality = Personality()

        // Simple inference rules
        for behavior in behaviors {
            switch behavior.type {
            case "decision":
                // Impulse purchase -> higher neuroticism
                if let impulse = behavior.context["impulse_purchase"]?.value as? Bool, impulse {
                    personality.neuroticism = min(1.0, personality.neuroticism + 0.05)
                }

            case "interaction":
                // Group chats -> higher extraversion
                if let group = behavior.context["group_chats"]?.value as? Bool, group {
                    personality.extraversion = min(1.0, personality.extraversion + 0.05)
                }

            case "preference":
                // Novel experiences -> higher openness
                if let novel = behavior.context["novel_trying"]?.value as? Bool, novel {
                    personality.openness = min(1.0, personality.openness + 0.05)
                }

            case "activity":
                // Regular schedule -> higher conscientiousness
                if let regular = behavior.context["regular_schedule"]?.value as? Bool, regular {
                    personality.conscientiousness = min(1.0, personality.conscientiousness + 0.05)
                }

            default:
                break
            }
        }

        return personality
    }
}