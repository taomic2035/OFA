// AgentModeManager.swift
// OFA iOS SDK - Agent Mode Management (v8.1.0)

import Foundation
import Combine

/// Agent mode manager - handles mode transitions and sync scheduling
public class AgentModeManager: ObservableObject {

    // MARK: - Published Properties

    @Published public var mode: AgentMode
    @Published public var isSyncing: Bool = false
    @Published public var lastSyncTime: Date?

    // MARK: - Private Properties

    private var syncTimer: Timer?
    private var syncInterval: TimeInterval = 300.0  // 5 minutes default
    private var onSyncCallback: (() async throws -> Void)?

    private var cancellables = Set<AnyCancellable>()

    // MARK: - Initialization

    public init(mode: AgentMode = .sync) {
        self.mode = mode

        if mode == .sync {
            startSyncTimer()
        }
    }

    // MARK: - Public Methods

    /// Set agent mode
    public func setMode(_ newMode: AgentMode) {
        mode = newMode

        if newMode == .sync {
            startSyncTimer()
        } else {
            stopSyncTimer()
        }
    }

    /// Set sync interval
    public func setSyncInterval(_ interval: TimeInterval) {
        syncInterval = interval

        if mode == .sync {
            // Restart timer with new interval
            stopSyncTimer()
            startSyncTimer()
        }
    }

    /// Set sync callback
    public func setSyncCallback(_ callback: @escaping () async throws -> Void) {
        self.onSyncCallback = callback
    }

    /// Trigger manual sync
    public func triggerSync() async throws {
        guard mode == .sync else {
            throw OFAError.configurationError("Sync only available in SYNC mode")
        }

        isSyncing = true

        if let callback = onSyncCallback {
            try await callback()
        }

        lastSyncTime = Date()
        isSyncing = false
    }

    /// Get sync status
    public func getSyncStatus() -> SyncStatus {
        return SyncStatus(
            mode: mode,
            isSyncing: isSyncing,
            lastSyncTime: lastSyncTime,
            nextSyncTime: getNextSyncTime()
        )
    }

    // MARK: - Private Methods

    private func startSyncTimer() {
        syncTimer = Timer.scheduledTimer(withTimeInterval: syncInterval, repeats: true) { _ in
            Task {
                try? await self.performSync()
            }
        }
    }

    private func stopSyncTimer() {
        syncTimer?.invalidate()
        syncTimer = nil
    }

    private func performSync() async throws {
        isSyncing = true

        if let callback = onSyncCallback {
            try await callback()
        }

        lastSyncTime = Date()
        isSyncing = false
    }

    private func getNextSyncTime() -> Date? {
        guard mode == .sync, let timer = syncTimer else { return nil }
        return Date().addingTimeInterval(syncInterval)
    }
}

// MARK: - Supporting Types

/// Sync status
public struct SyncStatus {
    public let mode: AgentMode
    public let isSyncing: Bool
    public let lastSyncTime: Date?
    public let nextSyncTime: Date?

    public var needsSync: Bool {
        guard let lastSync = lastSyncTime else { return true }
        return Date().timeIntervalSince(lastSync) > 300  // 5 minutes
    }
}