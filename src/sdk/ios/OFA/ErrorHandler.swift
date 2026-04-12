// ErrorHandler.swift
// OFA iOS SDK - Error Handling Framework (v9.8.0)
// Aligned with Android SDK ErrorHandler

import Foundation
import Combine

// MARK: - Error Categories

/// Error category for classification
public enum ErrorCategory: String, Codable {
    case network = "network"         // Network connectivity issues
    case connection = "connection"   // Center connection issues
    case timeout = "timeout"         // Operation timeout
    case authentication = "authentication"  // Authentication failures
    case resource = "resource"       // Resource unavailable
    case validation = "validation"   // Input validation errors
    case internal = "internal"       // Internal SDK errors
    case unknown = "unknown"         // Unknown errors
}

/// Error severity level
public enum ErrorSeverity: String, Codable {
    case low = "low"         // Minor issue, auto-recovery possible
    case medium = "medium"   // Moderate issue, needs retry
    case high = "high"       // Serious issue, needs intervention
    case critical = "critical"  // System failure, needs restart
}

/// Recovery strategy
public enum RecoveryStrategy: String, Codable {
    case immediateRetry = "immediate_retry"    // Retry immediately
    case backoffRetry = "backoff_retry"        // Retry with exponential backoff
    case circuitBreak = "circuit_break"        // Stop trying, circuit open
    case gracefulDegrade = "graceful_degrade"  // Use fallback
    case manualIntervention = "manual_intervention"  // Requires manual intervention
    case none = "none"         // No recovery possible
}

// MARK: - OFAError

/// OFAError represents a categorized error with recovery information
public struct OFAError: Error, Codable {
    public let code: String
    public let message: String
    public let category: ErrorCategory
    public let severity: ErrorSeverity
    public let strategy: RecoveryStrategy
    public let cause: String?
    public let timestamp: Date
    public let context: String?

    public init(
        code: String,
        message: String,
        category: ErrorCategory,
        severity: ErrorSeverity,
        strategy: RecoveryStrategy,
        cause: Error? = nil,
        context: String? = nil
    ) {
        self.code = code
        self.message = message
        self.category = category
        self.severity = severity
        self.strategy = strategy
        self.cause = cause?.localizedDescription
        self.timestamp = Date()
        self.context = context
    }

    public var isRecoverable: Bool {
        return strategy != .none && strategy != .manualIntervention
    }

    public var localizedDescription: String {
        return "[\(code)] \(message) (\(category.rawValue))"
    }
}

// MARK: - Retry Configuration

/// Retry configuration for RetryExecutor
public struct RetryConfig {
    public let maxAttempts: Int
    public let initialDelayMs: TimeInterval
    public let maxDelayMs: TimeInterval
    public let backoffFactor: Double
    public let retryableCategories: [ErrorCategory]

    public init(
        maxAttempts: Int = 3,
        initialDelayMs: TimeInterval = 1000,
        maxDelayMs: TimeInterval = 10000,
        backoffFactor: Double = 2.0,
        retryableCategories: [ErrorCategory] = [.network, .connection, .timeout]
    ) {
        self.maxAttempts = maxAttempts
        self.initialDelayMs = initialDelayMs
        self.maxDelayMs = maxDelayMs
        self.backoffFactor = backoffFactor
        self.retryableCategories = retryableCategories
    }

    public func shouldRetry(category: ErrorCategory) -> Bool {
        return retryableCategories.contains(category)
    }

    public static var defaultConfig: RetryConfig {
        return RetryConfig(maxAttempts: 3, initialDelayMs: 1000, maxDelayMs: 10000, backoffFactor: 2.0)
    }

    public static var aggressiveConfig: RetryConfig {
        return RetryConfig(maxAttempts: 5, initialDelayMs: 500, maxDelayMs: 30000, backoffFactor: 1.5)
    }

    public static var conservativeConfig: RetryConfig {
        return RetryConfig(maxAttempts: 2, initialDelayMs: 2000, maxDelayMs: 5000, backoffFactor: 2.0)
    }
}

// MARK: - Circuit Breaker

/// Circuit breaker state
public enum CircuitState: String, Codable {
    case closed = "closed"      // Normal operation
    case open = "open"          // Circuit tripped, reject all
    case halfOpen = "half_open" // Testing if recovered
}

/// Circuit breaker for preventing cascading failures
public class CircuitBreaker: ObservableObject {

    // MARK: - Published Properties

    @Published public var state: CircuitState = .closed
    @Published public var failureCount: Int = 0

    // MARK: - Configuration

    public let name: String
    public let failureThreshold: Int
    public let recoveryTimeMs: TimeInterval
    public let halfOpenDurationMs: TimeInterval

    // MARK: - Private Properties

    private var lastStateChangeTime: Date = Date()
    private var successCountInHalfOpen: Int = 0
    private var recoveryTimer: Timer?

    // MARK: - Initialization

    public init(
        name: String,
        failureThreshold: Int = 5,
        recoveryTimeMs: TimeInterval = 30000,
        halfOpenDurationMs: TimeInterval = 5000
    ) {
        self.name = name
        self.failureThreshold = failureThreshold
        self.recoveryTimeMs = recoveryTimeMs
        self.halfOpenDurationMs = halfOpenDurationMs
    }

    public static func defaultBreaker(name: String) -> CircuitBreaker {
        return CircuitBreaker(name: name, failureThreshold: 5, recoveryTimeMs: 30000, halfOpenDurationMs: 5000)
    }

    // MARK: - Public Methods

    /// Check if execution is allowed
    public func allowExecution() -> Bool {
        let now = Date()

        switch state {
        case .closed:
            return true

        case .open:
            // Check if recovery time elapsed
            let elapsed = now.timeIntervalSince(lastStateChangeTime) * 1000
            if elapsed >= recoveryTimeMs {
                transitionToHalfOpen()
                return true
            }
            return false

        case .halfOpen:
            return true
        }
    }

    /// Record successful execution
    public func recordSuccess() {
        failureCount = 0

        if state == .halfOpen {
            successCountInHalfOpen += 1
            if successCountInHalfOpen >= 2 {
                transitionToClosed()
            }
        }
    }

    /// Record failed execution
    public func recordFailure() {
        failureCount += 1

        if state == .halfOpen {
            transitionToOpen()
        } else if state == .closed && failureCount >= failureThreshold {
            transitionToOpen()
        }
    }

    /// Reset circuit breaker
    public func reset() {
        transitionToClosed()
    }

    // MARK: - State Transitions

    private func transitionToOpen() {
        state = .open
        lastStateChangeTime = Date()
        successCountInHalfOpen = 0
        print("CircuitBreaker[\(name)] OPEN - failures: \(failureCount)")
    }

    private func transitionToHalfOpen() {
        state = .halfOpen
        lastStateChangeTime = Date()
        successCountInHalfOpen = 0
        print("CircuitBreaker[\(name)] HALF_OPEN - testing")

        // Auto-close if half-open duration passes without issues
        recoveryTimer = Timer.scheduledTimer(withTimeInterval: halfOpenDurationMs / 1000, repeats: false) { _ in
            if self.state == .halfOpen {
                self.transitionToClosed()
            }
        }
    }

    private func transitionToClosed() {
        state = .closed
        lastStateChangeTime = Date()
        failureCount = 0
        successCountInHalfOpen = 0
        recoveryTimer?.invalidate()
        recoveryTimer = nil
        print("CircuitBreaker[\(name)] CLOSED - recovered")
    }
}

// MARK: - Retry Executor

/// Retry executor with exponential backoff
public class RetryExecutor {

    private let config: RetryConfig
    private var attemptCount: Int = 0
    private var cancelled: Bool = false

    public init(config: RetryConfig = .defaultConfig) {
        self.config = config
    }

    /// Execute operation with retry
    public func execute<T>(
        operation: @escaping (Int) async throws -> T,
        circuitBreaker: CircuitBreaker? = nil
    ) async throws -> T {
        attemptCount = 0
        cancelled = false

        return try await doExecute(operation: operation, circuitBreaker: circuitBreaker)
    }

    private func doExecute<T>(
        operation: @escaping (Int) async throws -> T,
        circuitBreaker: CircuitBreaker?
    ) async throws -> T {

        if cancelled {
            throw OFAError(
                code: "RETRY_CANCELLED",
                message: "Retry cancelled",
                category: .internal,
                severity: .medium,
                strategy: .none
            )
        }

        // Check circuit breaker
        if let breaker = circuitBreaker, !breaker.allowExecution() {
            throw OFAError(
                code: "CIRCUIT_OPEN",
                message: "Circuit breaker is open",
                category: .internal,
                severity: .high,
                strategy: .circuitBreak
            )
        }

        attemptCount += 1

        do {
            let result = try await operation(attemptCount)
            circuitBreaker?.recordSuccess()
            return result
        } catch {
            let ofaError = ErrorHandler.categorizeError(error)
            circuitBreaker?.recordFailure()
            return try await handleFailure(error: ofaError, operation: operation, circuitBreaker: circuitBreaker)
        }
    }

    private func handleFailure<T>(
        error: OFAError,
        operation: @escaping (Int) async throws -> T,
        circuitBreaker: CircuitBreaker?
    ) async throws -> T {

        // Check if we should retry
        if attemptCount < config.maxAttempts &&
           config.shouldRetry(category: error.category) &&
           error.isRecoverable {

            let delay = calculateDelay()
            print("Retrying after \(delay)ms (attempt \(attemptCount)/\(config.maxAttempts)): \(error.message)")

            try await Task.sleep(nanoseconds: UInt64(delay * 1_000_000))
            return try await doExecute(operation: operation, circuitBreaker: circuitBreaker)
        }

        print("Retry exhausted: \(error.message)")
        throw error
    }

    private func calculateDelay() -> TimeInterval {
        let delay = config.initialDelayMs * pow(config.backoffFactor, Double(attemptCount - 1))
        return min(delay, config.maxDelayMs)
    }

    public func cancel() {
        cancelled = true
    }

    public var currentAttempt: Int {
        return attemptCount
    }
}

// MARK: - Connection Recovery Manager

/// Connection recovery manager for automatic reconnection
public class ConnectionRecoveryManager: ObservableObject {

    @Published public var isRecovering: Bool = false
    @Published public var recoveryAttempts: Int = 0

    private let retryExecutor: RetryExecutor
    private let circuitBreaker: CircuitBreaker
    private var recoveryTask: Task<Void, Never>?

    private static let healthCheckInterval: TimeInterval = 5.0
    private static let maxRecoveryAttempts: Int = 10

    private var centerConnection: CenterConnection?

    public init(centerConnection: CenterConnection? = nil) {
        self.centerConnection = centerConnection
        self.retryExecutor = RetryExecutor(config: .aggressiveConfig)
        self.circuitBreaker = CircuitBreaker.defaultBreaker(name: "center_connection")
    }

    public func setConnection(_ connection: CenterConnection) {
        self.centerConnection = connection
    }

    /// Start recovery process
    public func startRecovery() async {
        if isRecovering { return }

        isRecovering = true
        recoveryAttempts = 0
        print("Starting connection recovery...")

        recoveryTask = Task {
            await doRecovery()
        }
    }

    private func doRecovery() async {
        while isRecovering && recoveryAttempts < Self.maxRecoveryAttempts {
            if !circuitBreaker.allowExecution() {
                print("Circuit breaker open, waiting for recovery window")
                try? await Task.sleep(nanoseconds: UInt64(circuitBreaker.recoveryTimeMs * 1_000_000))
                continue
            }

            // Check current connection state
            if centerConnection?.connectionState == .connected {
                print("Connection recovered!")
                isRecovering = false
                circuitBreaker.recordSuccess()
                return
            }

            // Attempt reconnect
            recoveryAttempts += 1
            print("Attempting reconnect (attempt \(recoveryAttempts))...")

            do {
                if let address = centerConnection?.centerAddress {
                    try await centerConnection?.connect(address: address)
                    if centerConnection?.connectionState == .connected {
                        print("Recovery successful!")
                        isRecovering = false
                        circuitBreaker.recordSuccess()
                        return
                    }
                }
            } catch {
                circuitBreaker.recordFailure()
                print("Recovery attempt failed: \(error.localizedDescription)")
            }

            try? await Task.sleep(nanoseconds: UInt64(Self.healthCheckInterval * 1_000_000))
        }

        isRecovering = false
        print("Recovery exhausted after \(recoveryAttempts) attempts")
    }

    /// Stop recovery process
    public func stopRecovery() {
        isRecovering = false
        recoveryTask?.cancel()
        retryExecutor.cancel()
    }

    public var circuitState: CircuitState {
        return circuitBreaker.state
    }
}

// MARK: - Fallback Provider

/// Fallback provider for graceful degradation
public class FallbackProvider {

    private var handlers: [FallbackHandler] = []

    public protocol FallbackHandler {
        func canHandle(category: ErrorCategory) -> Bool
        func provideFallback<T>(error: OFAError, context: String?) -> T?
    }

    public func registerHandler(handler: FallbackHandler) {
        handlers.append(handler)
    }

    public func getFallback<T>(error: OFAError, context: String?) -> T? {
        for handler in handlers {
            if handler.canHandle(category: error.category) {
                if let result: T = handler.provideFallback(error: error, context: context) {
                    print("Using fallback for \(error.category.rawValue)")
                    return result
                }
            }
        }
        return nil
    }

    public static func createDefault() -> FallbackProvider {
        let provider = FallbackProvider()

        // Network fallback - use cached data
        provider.registerHandler(NetworkFallbackHandler())

        // Timeout fallback - use quick response
        provider.registerHandler(TimeoutFallbackHandler())

        return provider
    }
}

// MARK: - Default Fallback Handlers

/// Network fallback handler
public class NetworkFallbackHandler: FallbackProvider.FallbackHandler {
    public func canHandle(category: ErrorCategory) -> Bool {
        return category == .network || category == .connection
    }

    public func provideFallback<T>(error: OFAError, context: String?) -> T? {
        // Return cached response if available
        // Implementation would check local cache
        return nil
    }
}

/// Timeout fallback handler
public class TimeoutFallbackHandler: FallbackProvider.FallbackHandler {
    public func canHandle(category: ErrorCategory) -> Bool {
        return category == .timeout
    }

    public func provideFallback<T>(error: OFAError, context: String?) -> T? {
        // Return simplified response
        return nil
    }
}

// MARK: - Error Handler

/// Error handler main class
public class ErrorHandler {

    // MARK: - Error Listeners

    private static var listeners: [ErrorListener] = []

    public protocol ErrorListener {
        func onErrorOccurred(error: OFAError)
        func onErrorRecovered(error: OFAError)
    }

    public static func addListener(listener: ErrorListener) {
        listeners.append(listener)
    }

    public static func removeListener(listener: ErrorListener) {
        listeners.removeAll { $0 == listener }
    }

    public static func notifyError(error: OFAError) {
        for listener in listeners {
            listener.onErrorOccurred(error: error)
        }
    }

    public static func notifyRecovery(error: OFAError) {
        for listener in listeners {
            listener.onErrorRecovered(error: error)
        }
    }

    // MARK: - Error Categorization

    /// Categorize an error into OFAError
    public static func categorizeError(_ error: Error) -> OFAError {
        let message = error.localizedDescription

        // Network errors
        if isNetworkError(error) {
            return OFAError(
                code: "NETWORK_ERROR",
                message: message,
                category: .network,
                severity: .medium,
                strategy: .backoffRetry,
                cause: error
            )
        }

        // Connection errors
        if isConnectionError(error) {
            return OFAError(
                code: "CONNECTION_ERROR",
                message: message,
                category: .connection,
                severity: .medium,
                strategy: .backoffRetry,
                cause: error
            )
        }

        // Timeout errors
        if isTimeoutError(error) {
            return OFAError(
                code: "TIMEOUT_ERROR",
                message: message,
                category: .timeout,
                severity: .low,
                strategy: .immediateRetry,
                cause: error
            )
        }

        // Authentication errors
        if isAuthenticationError(error) {
            return OFAError(
                code: "AUTH_ERROR",
                message: message,
                category: .authentication,
                severity: .high,
                strategy: .manualIntervention,
                cause: error
            )
        }

        // Internal errors
        if isInternalError(error) {
            return OFAError(
                code: "INTERNAL_ERROR",
                message: message,
                category: .internal,
                severity: .high,
                strategy: .gracefulDegrade,
                cause: error
            )
        }

        // Unknown errors
        return OFAError(
            code: "UNKNOWN_ERROR",
            message: message,
            category: .unknown,
            severity: .medium,
            strategy: .none,
            cause: error
        )
    }

    // MARK: - Error Detection Helpers

    private static func isNetworkError(_ error: Error) -> Bool {
        let errorType = type(of: error)
        let typeName = String(describing: errorType)

        return typeName.contains("URLError") ||
               typeName.contains("Network") ||
               messageContains(error, keywords: ["network", "offline", "unreachable", "no internet"])
    }

    private static func isConnectionError(_ error: Error) -> Bool {
        let typeName = String(describing: type(of: error))

        return typeName.contains("Connection") ||
               typeName.contains("WebSocket") ||
               messageContains(error, keywords: ["connection", "closed", "disconnected", "broken pipe"])
    }

    private static func isTimeoutError(_ error: Error) -> Bool {
        let typeName = String(describing: type(of: error))

        return typeName.contains("Timeout") ||
               messageContains(error, keywords: ["timeout", "timed out", "deadline"])
    }

    private static func isAuthenticationError(_ error: Error) -> Bool {
        let typeName = String(describing: type(of: error))

        return typeName.contains("Auth") ||
               typeName.contains("Security") ||
               messageContains(error, keywords: ["auth", "token", "permission", "access denied"])
    }

    private static func isInternalError(_ error: Error) -> Bool {
        let typeName = String(describing: type(of: error))

        return typeName.contains("Fatal") ||
               typeName.contains("Runtime") ||
               typeName.contains("Illegal")
    }

    private static func messageContains(_ error: Error, keywords: [String]) -> Bool {
        let message = error.localizedDescription.lowercased()
        return keywords.contains { keyword in message.contains(keyword.lowercased()) }
    }
}