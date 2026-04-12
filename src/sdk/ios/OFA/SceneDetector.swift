// SceneDetector.swift
// OFA iOS SDK - Scene Detection (v8.1.0)

import Foundation
import Combine
import CoreMotion
import CoreLocation

/// Scene type for user context
public enum SceneType: String, Codable {
    case running = "running"
    case walking = "walking"
    case driving = "driving"
    case meeting = "meeting"
    case sleeping = "sleeping"
    case exercise = "exercise"
    case work = "work"
    case home = "home"
    case travel = "travel"
    case healthAlert = "health_alert"
    case unknown = "unknown"
}

/// Scene state
public struct SceneState: Codable {
    public let id: String
    public let type: SceneType
    public let identityId: String
    public let agentId: String
    public let startTime: Date
    public var endTime: Date?
    public let confidence: Double
    public let context: [String: AnyCodable]
    public let actions: [SceneAction]
    public var active: Bool
}

/// Scene action
public struct SceneAction: Codable {
    public let type: String         // route, notify, block, alert
    public let targetAgent: String?
    public let payload: [String: AnyCodable]
    public let delay: TimeInterval?
}

/// Scene listener protocol
public protocol SceneListener {
    func onSceneStart(_ scene: SceneState)
    func onSceneEnd(_ scene: SceneState)
    func onSceneAction(_ scene: SceneState, action: SceneAction)
}

/// Scene detector - detects user context from device sensors
public class SceneDetector: ObservableObject {

    // MARK: - Published Properties

    @Published public var currentScene: SceneType = .unknown
    @Published public var sceneConfidence: Double = 0.0

    // MARK: - Private Properties

    private var motionManager: CMMotionActivityManager?
    private var locationManager: CLLocationManager?
    private var healthStore: HealthStore?

    private var listeners: [SceneListener] = []
    private var activeScene: SceneState?
    private var detectionInterval: TimeInterval = 5.0
    private var detectionTimer: Timer?

    // MARK: - Initialization

    public func initialize() {
        #if os(iOS)
        motionManager = CMMotionActivityManager()
        locationManager = CLLocationManager()
        locationManager?.requestWhenInUseAuthorization()

        #if !os(watchOS)
        healthStore = HealthStore()
        #endif
        #endif

        startDetection()
    }

    // MARK: - Public Methods

    /// Start scene detection
    public func startDetection() {
        detectionTimer = Timer.scheduledTimer(withTimeInterval: detectionInterval, repeats: true) { _ in
            Task {
                await self.detect()
            }
        }
    }

    /// Stop scene detection
    public func stopDetection() {
        detectionTimer?.invalidate()
        detectionTimer = nil
    }

    /// Detect current scene
    public func detect() async {
        let context = await gatherContext()
        let detectedScene = analyzeContext(context)

        currentScene = detectedScene.type
        sceneConfidence = detectedScene.confidence

        // Check if scene changed
        if activeScene?.type != detectedScene.type {
            // End previous scene
            if let previous = activeScene {
                endScene(previous)
            }

            // Start new scene if confidence > 0.7
            if detectedScene.confidence > 0.7 {
                startScene(detectedScene)
            }
        }
    }

    /// Add listener
    public func addListener(_ listener: SceneListener) {
        listeners.append(listener)
    }

    /// Remove listener
    public func removeListener(_ listener: SceneListener) {
        listeners.removeAll { $0 == listener }
    }

    /// Get active scene
    public func getActiveScene() -> SceneState? {
        return activeScene
    }

    // MARK: - Private Methods

    private func gatherContext() async -> [String: Any] {
        var context: [String: Any] = [:]

        #if os(iOS)
        // Activity type from motion
        if let activity = await getMotionActivity() {
            context["activity_type"] = activityToSceneType(activity)
        }

        // Location
        if let location = locationManager?.location {
            context["latitude"] = location.coordinate.latitude
            context["longitude"] = location.coordinate.longitude
            context["speed"] = location.speed
        }

        // Health data
        if let heartRate = await healthStore?.getHeartRate() {
            context["heart_rate"] = heartRate
        }
        #endif

        return context
    }

    private func analyzeContext(_ context: [String: Any]) -> SceneState {
        var sceneType: SceneType = .unknown
        var confidence: Double = 0.0

        // Analyze activity type
        if let activity = context["activity_type"] as? String {
            switch activity {
            case "running":
                sceneType = .running
                confidence = 0.9
            case "walking":
                sceneType = .walking
                confidence = 0.85
            case "automotive":
                sceneType = .driving
                confidence = 0.9
            case "stationary":
                // Could be meeting, work, or home
                confidence = 0.5
            default:
                break
            }
        }

        // Check health alerts
        if let heartRate = context["heart_rate"] as? Double {
            if heartRate > 120 || heartRate < 50 {
                sceneType = .healthAlert
                confidence = 0.95
            }
        }

        // Create scene state
        return SceneState(
            id: UUID().uuidString,
            type: sceneType,
            identityId: "",
            agentId: "",
            startTime: Date(),
            endTime: nil,
            confidence: confidence,
            context: context.mapValues { AnyCodable($0) },
            actions: generateActions(for: sceneType),
            active: true
        )
    }

    private func generateActions(for sceneType: SceneType) -> [SceneAction] {
        var actions: [SceneAction] = []

        switch sceneType {
        case .running:
            actions.append(SceneAction(
                type: "route",
                targetAgent: "phone",
                payload: ["message_type": AnyCodable("running_status")],
                delay: nil
            ))
            actions.append(SceneAction(
                type: "filter",
                targetAgent: "watch",
                payload: ["filter_type": AnyCodable("urgent_only")],
                delay: nil
            ))

        case .meeting:
            actions.append(SceneAction(
                type: "block",
                targetAgent: nil,
                payload: ["block_calls": AnyCodable(true), "except_urgent": AnyCodable(true)],
                delay: nil
            ))

        case .healthAlert:
            actions.append(SceneAction(
                type: "alert",
                targetAgent: nil,
                payload: ["broadcast": AnyCodable(true)],
                delay: nil
            ))

        default:
            break
        }

        return actions
    }

    private func startScene(_ scene: SceneState) {
        activeScene = scene

        for listener in listeners {
            listener.onSceneStart(scene)
        }

        // Execute actions
        for action in scene.actions {
            executeAction(scene, action: action)
        }
    }

    private func endScene(_ scene: SceneState) {
        var endedScene = scene
        endedScene.endTime = Date()
        endedScene.active = false

        activeScene = nil

        for listener in listeners {
            listener.onSceneEnd(endedScene)
        }
    }

    private func executeAction(_ scene: SceneState, action: SceneAction) {
        for listener in listeners {
            listener.onSceneAction(scene, action: action)
        }
    }

    #if os(iOS)
    private func getMotionActivity() async -> CMMotionActivity? {
        guard let motionManager = motionManager else { return nil }

        return await withCheckedContinuation { continuation in
            motionManager.queryActivityStarting(from: Date.distantPast, to: Date(), to: .main) { activities, error in
                if let activities = activities, let last = activities.last {
                    continuation.resume(returning: last)
                } else {
                    continuation.resume(returning: nil)
                }
            }
        }
    }

    private func activityToSceneType(_ activity: CMMotionActivity) -> String {
        if activity.running {
            return "running"
        } else if activity.walking {
            return "walking"
        } else if activity.automotive {
            return "automotive"
        } else if activity.stationary {
            return "stationary"
        } else {
            return "unknown"
        }
    }
    #endif
}

// MARK: - Health Store (iOS)

#if os(iOS) && !os(watchOS)
import HealthKit

public class HealthStore {
    private var healthStore: HKHealthStore?

    public init() {
        if HKHealthStore.isHealthDataAvailable() {
            healthStore = HKHealthStore()
        }
    }

    public func getHeartRate() async -> Double? {
        guard let healthStore = healthStore else { return nil }

        let heartRateType = HKQuantityType.quantityTypeForIdentifier(.heartRate)!
        let query = HKSampleQuery(
            quantityType: heartRateType,
            predicate: nil,
            limit: 1,
            sortDescriptors: [NSSortDescriptor(key: HKSampleSortIdentifierStartDate, ascending: false)]
        ) { _, samples, _ in
            // Query completed
        }

        healthStore.execute(query)

        return nil // Placeholder - would need proper async implementation
    }
}
#else
public class HealthStore {
    public init() {}
    public func getHeartRate() async -> Double? { return nil }
}
#endif