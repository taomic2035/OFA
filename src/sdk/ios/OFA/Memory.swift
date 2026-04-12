// Memory.swift
// OFA iOS SDK - Memory System (v9.8.0)
// Aligned with Android SDK Memory Package

import Foundation
import Combine

// MARK: - Memory Levels

/// Memory storage level
public enum MemoryLevel: Int, Codable {
    case l1 = 1  // Immediate context (seconds, ~10 entries)
    case l2 = 2  // Recent context (minutes, ~100 entries)
    case l3 = 3  // Session context (hours, ~500 entries)
    case archive = 4  // Long-term storage (unlimited)
}

// MARK: - Memory Entry

/// Memory entry
public struct MemoryEntry: Codable, Identifiable {
    public let id: String
    public let key: String
    public let value: AnyCodable
    public let level: MemoryLevel
    public let category: MemoryCategory
    public let importance: Double     // 0-1
    public let createdAt: Date
    public var lastAccessed: Date
    public var accessCount: Int
    public var expiresAt: Date?

    public init(
        key: String,
        value: Any,
        level: MemoryLevel = .l2,
        category: MemoryCategory = .general,
        importance: Double = 0.5
    ) {
        self.id = UUID().uuidString
        self.key = key
        self.value = AnyCodable(value)
        self.level = level
        self.category = category
        self.importance = importance
        self.createdAt = Date()
        self.lastAccessed = Date()
        self.accessCount = 0
        self.expiresAt = nil
    }

    /// Update access
    public func access() -> MemoryEntry {
        var updated = self
        updated.lastAccessed = Date()
        updated.accessCount += 1
        return updated
    }

    /// Check if expired
    public var isExpired: Bool {
        guard let expiresAt = expiresAt else { return false }
        return Date() > expiresAt
    }
}

// MARK: - Memory Category

/// Memory category
public enum MemoryCategory: String, Codable {
    case general = "general"
    case identity = "identity"
    case behavior = "behavior"
    case scene = "scene"
    case preference = "preference"
    case task = "task"
    case conversation = "conversation"
    case health = "health"
    case location = "location"
    case device = "device"
}

// MARK: - Memory Cache

/// Memory cache for L1/L2 storage
public class MemoryCache: ObservableObject {

    // MARK: - Published Properties

    @Published public var entryCount: Int = 0
    @Published public var hitRate: Double = 0.0

    // MARK: - Private Properties

    private var cache: [String: MemoryEntry] = [:]
    private let maxSize: Int
    private let level: MemoryLevel
    private var hits: Int = 0
    private var misses: Int = 0

    // MARK: - Initialization

    public init(level: MemoryLevel, maxSize: Int) {
        self.level = level
        self.maxSize = maxSize
    }

    // MARK: - Public Methods

    /// Put entry in cache
    public func put(entry: MemoryEntry) {
        if cache.count >= maxSize {
            evictOldest()
        }

        cache[entry.key] = entry
        entryCount = cache.count
    }

    /// Get entry from cache
    public func get(key: String) -> MemoryEntry? {
        if let entry = cache[key] {
            if entry.isExpired {
                cache.removeValue(forKey: key)
                misses += 1
                updateHitRate()
                return nil
            }

            // Update access
            let updated = entry.access()
            cache[key] = updated
            hits += 1
            updateHitRate()
            return updated
        }

        misses += 1
        updateHitRate()
        return nil
    }

    /// Remove entry
    public func remove(key: String) {
        cache.removeValue(forKey: key)
        entryCount = cache.count
    }

    /// Clear cache
    public func clear() {
        cache.removeAll()
        entryCount = 0
        hits = 0
        misses = 0
        hitRate = 0.0
    }

    /// Get all entries
    public func getAll() -> [MemoryEntry] {
        return Array(cache.values)
    }

    /// Get entries by category
    public func getByCategory(category: MemoryCategory) -> [MemoryEntry] {
        return cache.values.filter { $0.category == category }
    }

    // MARK: - Private Methods

    private func evictOldest() {
        // Find oldest accessed entry
        let sorted = cache.values.sorted { $0.lastAccessed < $1.lastAccessed }
        if let oldest = sorted.first {
            cache.removeValue(forKey: oldest.key)
        }
    }

    private func updateHitRate() {
        let total = hits + misses
        if total > 0 {
            hitRate = Double(hits) / Double(total)
        }
    }
}

// MARK: - Memory Archive

/// Memory archive for L3 long-term storage
public class MemoryArchive {

    private let fileName = "memory_archive.json"
    private var archive: [MemoryEntry] = []

    // MARK: - Public Methods

    /// Save entry to archive
    public func save(entry: MemoryEntry) async throws {
        archive.append(entry)
        await persist()
    }

    /// Load entry from archive
    public func load(key: String) async -> MemoryEntry? {
        return archive.first { $0.key == key }
    }

    /// Load all entries
    public func loadAll() async -> [MemoryEntry] {
        return archive
    }

    /// Load entries by category
    public func loadByCategory(category: MemoryCategory) async -> [MemoryEntry] {
        return archive.filter { $0.category == category }
    }

    /// Search entries
    public func search(query: String) async -> [MemoryEntry] {
        return archive.filter { entry in
            entry.key.contains(query) ||
            (entry.value.value as? String)?.contains(query) ?? false
        }
    }

    /// Delete entry
    public func delete(key: String) async throws {
        archive.removeAll { $0.key == key }
        await persist()
    }

    /// Clear archive
    public func clear() async throws {
        archive.removeAll()
        await persist()
    }

    /// Initialize - load from file
    public func initialize() async throws {
        guard let data = try? Data(contentsOf: getFilePath()) else { return }
        archive = try JSONDecoder().decode([MemoryEntry].self, from: data)
    }

    // MARK: - Private Methods

    private func persist() async {
        let encoder = JSONEncoder()
        encoder.outputFormatting = .prettyPrinted

        if let data = try? encoder.encode(archive) {
            try? data.write(to: getFilePath())
        }
    }

    private func getFilePath() -> URL {
        let documents = FileManager.default.urls(for: .documentDirectory, in: .userDomainMask).first!
        return documents.appendingPathComponent(fileName)
    }
}

// MARK: - Context Memory

/// Context memory - manages current conversation/operation context
public class ContextMemory: ObservableObject {

    @Published public var contextId: String = UUID().uuidString
    @Published public var entries: [MemoryEntry] = []

    private let l1Cache: MemoryCache
    private let maxContextSize: Int = 10

    public init(l1Cache: MemoryCache) {
        self.l1Cache = l1Cache
    }

    /// Add context entry
    public func add(key: String, value: Any, category: MemoryCategory = .general) {
        let entry = MemoryEntry(
            key: key,
            value: value,
            level: .l1,
            category: category,
            importance: 1.0
        )

        entries.append(entry)
        l1Cache.put(entry: entry)

        // Keep only recent entries
        if entries.count > maxContextSize {
            entries.removeFirst()
        }
    }

    /// Get context value
    public func get(key: String) -> Any? {
        if let entry = l1Cache.get(key: key) {
            return entry.value.value
        }

        if let entry = entries.first { $0.key == key } {
            return entry.value.value
        }

        return nil
    }

    /// Clear context
    public func clear() {
        entries.removeAll()
        contextId = UUID().uuidString
    }

    /// Get context summary
    public func getSummary() -> [String: Any] {
        var summary: [String: Any] = [:]
        for entry in entries {
            summary[entry.key] = entry.value.value
        }
        return summary
    }
}

// MARK: - User Memory Manager

/// User memory manager - coordinates all memory levels
public class UserMemoryManager: ObservableObject {

    // MARK: - Published Properties

    @Published public var totalEntries: Int = 0
    @Published public var isInitialized: Bool = false

    // MARK: - Components

    public let l1Cache: MemoryCache  // Immediate context (10 entries)
    public let l2Cache: MemoryCache  // Recent context (100 entries)
    public let l3Cache: MemoryCache  // Session context (500 entries)
    public let archive: MemoryArchive

    public let contextMemory: ContextMemory

    // MARK: - Private Properties

    private var cancellables = Set<AnyCancellable>()

    // MARK: - Initialization

    public init() {
        l1Cache = MemoryCache(level: .l1, maxSize: 10)
        l2Cache = MemoryCache(level: .l2, maxSize: 100)
        l3Cache = MemoryCache(level: .l3, maxSize: 500)
        archive = MemoryArchive()
        contextMemory = ContextMemory(l1Cache: l1Cache)
    }

    /// Initialize memory manager
    public func initialize() async throws {
        try await archive.initialize()
        isInitialized = true
        updateTotalEntries()
    }

    // MARK: - Public Methods

    /// Store memory entry
    public func store(entry: MemoryEntry) async throws {
        // Store in appropriate cache level
        switch entry.level {
        case .l1:
            l1Cache.put(entry: entry)
        case .l2:
            l2Cache.put(entry: entry)
        case .l3:
            l3Cache.put(entry: entry)
        case .archive:
            try await archive.save(entry: entry)
        }

        updateTotalEntries()
    }

    /// Retrieve memory entry
    public func retrieve(key: String) async -> MemoryEntry? {
        // Try L1 first
        if let entry = l1Cache.get(key: key) {
            return entry
        }

        // Try L2
        if let entry = l2Cache.get(key: key) {
            // Promote to L1
            l1Cache.put(entry: entry)
            return entry
        }

        // Try L3
        if let entry = l3Cache.get(key: key) {
            // Promote to L2
            l2Cache.put(entry: entry)
            return entry
        }

        // Try archive
        if let entry = await archive.load(key: key) {
            // Promote to L3
            l3Cache.put(entry: entry)
            return entry
        }

        return nil
    }

    /// Update memory entry
    public func update(key: String, value: Any) async throws {
        if let existing = await retrieve(key: key) {
            let updated = MemoryEntry(
                key: key,
                value: value,
                level: existing.level,
                category: existing.category,
                importance: existing.importance
            )
            try await store(entry: updated)
        }
    }

    /// Delete memory entry
    public func delete(key: String) async throws {
        l1Cache.remove(key: key)
        l2Cache.remove(key: key)
        l3Cache.remove(key: key)
        try await archive.delete(key: key)
        updateTotalEntries()
    }

    /// Search memories
    public func search(query: String) async -> [MemoryEntry] {
        var results: [MemoryEntry] = []

        // Search all caches
        results.append(contentsOf: l1Cache.getAll().filter { $0.key.contains(query) })
        results.append(contentsOf: l2Cache.getAll().filter { $0.key.contains(query) })
        results.append(contentsOf: l3Cache.getAll().filter { $0.key.contains(query) })

        // Search archive
        results.append(contentsOf: await archive.search(query: query))

        return results
    }

    /// Get memories by category
    public func getByCategory(category: MemoryCategory) async -> [MemoryEntry] {
        var results: [MemoryEntry] = []

        results.append(contentsOf: l1Cache.getByCategory(category: category))
        results.append(contentsOf: l2Cache.getByCategory(category: category))
        results.append(contentsOf: l3Cache.getByCategory(category: category))
        results.append(contentsOf: await archive.loadByCategory(category: category))

        return results
    }

    /// Clear all memories
    public func clearAll() async throws {
        l1Cache.clear()
        l2Cache.clear()
        l3Cache.clear()
        try await archive.clear()
        contextMemory.clear()
        updateTotalEntries()
    }

    /// Promote important memories
    public func promoteImportant() async {
        // Find high-importance entries in lower levels and promote
        let l3Important = l3Cache.getAll().filter { $0.importance > 0.8 }
        for entry in l3Important {
            l2Cache.put(entry: entry)
        }

        let l2Important = l2Cache.getAll().filter { $0.importance > 0.8 }
        for entry in l2Important {
            l1Cache.put(entry: entry)
        }
    }

    /// Archive old memories
    public func archiveOld() async throws {
        let threshold = Date().addingTimeInterval(-3600)  // 1 hour ago

        let oldL3 = l3Cache.getAll().filter { $0.lastAccessed < threshold }
        for entry in oldL3 {
            try await archive.save(entry: entry)
            l3Cache.remove(key: entry.key)
        }
    }

    /// Get memory statistics
    public func getStats() -> MemoryStats {
        return MemoryStats(
            l1Count: l1Cache.entryCount,
            l2Count: l2Cache.entryCount,
            l3Count: l3Cache.entryCount,
            archiveCount: archive.loadAll().count,
            l1HitRate: l1Cache.hitRate,
            l2HitRate: l2Cache.hitRate,
            l3HitRate: l3Cache.hitRate
        )
    }

    // MARK: - Private Methods

    private func updateTotalEntries() {
        totalEntries = l1Cache.entryCount + l2Cache.entryCount + l3Cache.entryCount
    }
}

// MARK: - Memory Statistics

/// Memory statistics
public struct MemoryStats: Codable {
    public let l1Count: Int
    public let l2Count: Int
    public let l3Count: Int
    public let archiveCount: Int
    public let l1HitRate: Double
    public let l2HitRate: Double
    public let l3HitRate: Double

    public var totalCount: Int {
        return l1Count + l2Count + l3Count + archiveCount
    }

    public var averageHitRate: Double {
        return (l1HitRate + l2HitRate + l3HitRate) / 3.0
    }
}

// MARK: - Memory-Aware Skill Executor

/// Memory-aware skill executor - provides memory context to skills
public class MemoryAwareSkillExecutor {

    private let memoryManager: UserMemoryManager
    private let skillRegistry: SkillRegistry

    public init(memoryManager: UserMemoryManager, skillRegistry: SkillRegistry) {
        self.memoryManager = memoryManager
        self.skillRegistry = skillRegistry
    }

    /// Execute skill with memory context
    public func execute(skillId: String, params: [String: Any]) async throws -> SkillResult {
        // Get skill definition
        guard let skill = skillRegistry.get(skillId: skillId) else {
            throw OFAError(
                code: "SKILL_NOT_FOUND",
                message: "Skill not found: \(skillId)",
                category: .internal,
                severity: .medium,
                strategy: .none
            )
        }

        // Load relevant memories
        let relevantMemories = await loadRelevantMemories(skill: skill)

        // Create execution context
        let context = SkillContext(
            params: params,
            memories: relevantMemories,
            identityId: nil  // Would get from identity manager
        )

        // Execute skill
        let result = try await skill.execute(context: context)

        // Store result as memory
        let entry = MemoryEntry(
            key: "skill_result_\(skillId)_\(UUID().uuidString)",
            value: result.output ?? "",
            level: .l2,
            category: .task,
            importance: 0.7
        )
        try await memoryManager.store(entry: entry)

        return result
    }

    private func loadRelevantMemories(skill: Skill) async -> [MemoryEntry] {
        // Load memories relevant to skill category
        if let category = skill.memoryCategory {
            return await memoryManager.getByCategory(category: category)
        }

        // Search by skill keywords
        var results: [MemoryEntry] = []
        for keyword in skill.keywords {
            results.append(contentsOf: await memoryManager.search(query: keyword))
        }

        return results
    }
}

// MARK: - Skill Types (Simplified)

/// Skill context
public struct SkillContext {
    public let params: [String: Any]
    public let memories: [MemoryEntry]
    public let identityId: String?
}

/// Skill result
public struct SkillResult: Codable {
    public let success: Bool
    public let output: String?
    public let data: [String: AnyCodable]?
    public let error: String?

    public init(success: Bool, output: String? = nil, data: [String: Any]? = nil, error: String? = nil) {
        self.success = success
        self.output = output
        self.data = data?.mapValues { AnyCodable($0) }
        self.error = error
    }
}

/// Skill definition (simplified)
public protocol Skill {
    var id: String { get }
    var name: String { get }
    var keywords: [String] { get }
    var memoryCategory: MemoryCategory? { get }

    func execute(context: SkillContext) async throws -> SkillResult
}

/// Skill registry (placeholder)
public class SkillRegistry {
    private var skills: [String: Skill] = [:]

    public func register(skill: Skill) {
        skills[skill.id] = skill
    }

    public func get(skillId: String) -> Skill? {
        return skills[skillId]
    }
}