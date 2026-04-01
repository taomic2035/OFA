import Foundation

/// 离线缓存
public actor OfflineCache {
    private var cache: [String: CacheEntry] = [:]
    private var pendingSync: Set<String> = []
    private var hits = 0
    private var misses = 0
    private let maxSize: Int64
    private var currentSize: Int64 = 0
    private let fileManager = FileManager.default
    private let cacheDirectory: URL

    public init(maxSize: Int64 = 10 * 1024 * 1024) {
        self.maxSize = maxSize
        self.cacheDirectory = FileManager.default.temporaryDirectory.appendingPathComponent("ofa_cache")

        // 创建缓存目录
        try? fileManager.createDirectory(at: cacheDirectory, withIntermediateDirectories: true)

        // 加载持久化数据
        Task {
            await loadFromDisk()
        }
    }

    /// 存储数据
    public func put(_ key: String, data: Data, expiryMs: Int = 0) {
        let timestamp = Date()
        let expiry = expiryMs > 0 ? timestamp.addingTimeInterval(Double(expiryMs) / 1000) : nil

        // 检查容量
        if currentSize + Int64(data.count) > maxSize {
            evictIfNeeded(needed: Int64(data.count))
        }

        let entry = CacheEntry(
            data: data,
            timestamp: timestamp,
            expiry: expiry,
            synced: false
        )

        cache[key] = entry
        currentSize += Int64(data.count)
        pendingSync.insert(key)

        // 持久化
        saveToDisk(key: key, entry: entry)
    }

    /// 获取数据
    public func get(_ key: String) -> Data? {
        guard let entry = cache[key] else {
            misses += 1
            return nil
        }

        // 检查过期
        if let expiry = entry.expiry, Date() > expiry {
            remove(key)
            misses += 1
            return nil
        }

        hits += 1
        return entry.data
    }

    /// 删除数据
    public func remove(_ key: String) {
        guard let entry = cache.removeValue(forKey: key) else { return }
        currentSize -= Int64(entry.data.count)
        pendingSync.remove(key)
        deleteFromDisk(key: key)
    }

    /// 清空缓存
    public func clear() {
        cache.removeAll()
        pendingSync.removeAll()
        currentSize = 0

        // 清理磁盘
        try? fileManager.removeItem(at: cacheDirectory)
        try? fileManager.createDirectory(at: cacheDirectory, withIntermediateDirectories: true)
    }

    /// 获取待同步键列表
    public func getPendingKeys() -> [String] {
        Array(pendingSync)
    }

    /// 标记已同步
    public func markSynced(_ key: String) {
        pendingSync.remove(key)
        if var entry = cache[key] {
            entry.synced = true
            cache[key] = entry
        }
        deleteFromDisk(key: key, isPending: true)
    }

    /// 获取待同步数量
    public func getPendingCount() -> Int {
        pendingSync.count
    }

    /// 获取命中率
    public func hitRate() -> Double {
        let total = hits + misses
        return total > 0 ? Double(hits) / Double(total) : 0.0
    }

    /// 获取当前大小
    public func getCurrentSize() -> Int64 {
        currentSize
    }

    // MARK: - Private

    private func evictIfNeeded(needed: Int64) {
        let now = Date()

        // 清理过期项
        for (key, entry) in cache {
            if let expiry = entry.expiry, now > expiry {
                remove(key)
            }
        }

        // 清理最旧的已同步项
        while currentSize + needed > maxSize && !cache.isEmpty {
            var oldestKey: String?
            var oldestTime = Date.distantFuture

            for (key, entry) in cache where entry.synced {
                if entry.timestamp < oldestTime {
                    oldestTime = entry.timestamp
                    oldestKey = key
                }
            }

            if let key = oldestKey {
                remove(key)
            } else {
                break
            }
        }
    }

    private func saveToDisk(key: String, entry: CacheEntry) {
        let fileURL = cacheDirectory.appendingPathComponent(key.addingPercentEncoding(withAllowedCharacters: .alphanumerics) ?? key)
        try? entry.data.write(to: fileURL)

        // 保存元数据
        let metaURL = fileURL.appendingPathExtension("meta")
        let meta: [String: Any] = [
            "timestamp": entry.timestamp,
            "expiry": entry.expiry ?? nil,
            "synced": entry.synced
        ]
        try? (meta as NSDictionary).write(to: metaURL)
    }

    private func deleteFromDisk(key: String, isPending: Bool = false) {
        let safeKey = key.addingPercentEncoding(withAllowedCharacters: .alphanumerics) ?? key

        if !isPending {
            let fileURL = cacheDirectory.appendingPathComponent(safeKey)
            try? fileManager.removeItem(at: fileURL)
            try? fileManager.removeItem(at: fileURL.appendingPathExtension("meta"))
        }

        let pendingFile = cacheDirectory.appendingPathComponent("pending_\(safeKey)")
        try? fileManager.removeItem(at: pendingFile)
    }

    private func loadFromDisk() {
        guard let files = try? fileManager.contentsOfDirectory(at: cacheDirectory, includingPropertiesForKeys: nil) else {
            return
        }

        for fileURL in files where !fileURL.pathExtension.isEmpty {
            guard let data = try? Data(contentsOf: fileURL) else { continue }

            let key = fileURL.deletingPathExtension().lastPathComponent
            let metaURL = fileURL.appendingPathExtension("meta")

            var timestamp = Date()
            var expiry: Date?
            var synced = false

            if let meta = NSDictionary(contentsOf: metaURL) {
                if let ts = meta["timestamp"] as? Date { timestamp = ts }
                if let exp = meta["expiry"] as? Date { expiry = exp }
                if let syn = meta["synced"] as? Bool { synced = syn }
            }

            let entry = CacheEntry(
                data: data,
                timestamp: timestamp,
                expiry: expiry,
                synced: synced
            )

            cache[key] = entry
            currentSize += Int64(data.count)

            if !synced {
                pendingSync.insert(key)
            }
        }
    }
}

/// 缓存条目
private struct CacheEntry {
    let data: Data
    let timestamp: Date
    let expiry: Date?
    var synced: Bool
}