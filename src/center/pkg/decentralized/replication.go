// Package decentralized - 数据复制
// Sprint 33: v9.0 去中心化增强
package decentralized

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// DataReplicator 数据复制器
type DataReplicator struct {
	strategy     ReplicationStrategy
	replicas     map[string]*ReplicaInfo
	dataStore    map[string][]byte
	hashIndex    map[string]string // dataHash -> dataID
	mu           sync.RWMutex
	replicaLevel int // 默认复制级别
}

// ReplicationStrategy 复制策略
type ReplicationStrategy string

const (
	StrategyFull      ReplicationStrategy = "full"      // 全复制
	StrategyPartial   ReplicationStrategy = "partial"   // 部分复制
	StrategyGeo       ReplicationStrategy = "geo"       // 地理复制
	StrategyOnDemand  ReplicationStrategy = "on_demand" // 按需复制
	StrategyShard     ReplicationStrategy = "shard"     // 分片复制
)

// ReplicaInfo 复制信息
type ReplicaInfo struct {
	DataID       string    `json:"data_id"`
	DataHash     string    `json:"data_hash"`
	Nodes        []string  `json:"nodes"`        // 存储节点
	PrimaryNode  string    `json:"primary_node"` // 主节点
	Status       ReplicaStatus `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Size         int64     `json:"size"`
	Version      int       `json:"version"`
}

// ReplicaStatus 复制状态
type ReplicaStatus string

const (
	ReplicaStatusComplete ReplicaStatus = "complete"
	ReplicaStatusPartial  ReplicaStatus = "partial"
	ReplicaStatusPending  ReplicaStatus = "pending"
	ReplicaStatusFailed   ReplicaStatus = "failed"
)

// NewDataReplicator 创建数据复制器
func NewDataReplicator() *DataReplicator {
	return &DataReplicator{
		strategy:     StrategyPartial,
		replicas:     make(map[string]*ReplicaInfo),
		dataStore:    make(map[string][]byte),
		hashIndex:    make(map[string]string),
		replicaLevel: 3,
	}
}

// SetStrategy 设置策略
func (r *DataReplicator) SetStrategy(strategy ReplicationStrategy) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.strategy = strategy
}

// SetReplicaLevel 设置复制级别
func (r *DataReplicator) SetReplicaLevel(level int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.replicaLevel = level
}

// StoreData 存储数据
func (r *DataReplicator) StoreData(dataID string, data []byte) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 计算哈希
	hash := sha256.Sum256(data)
	dataHash := hex.EncodeToString(hash[:])

	// 存储数据
	r.dataStore[dataID] = data
	r.hashIndex[dataHash] = dataID

	return dataHash
}

// GetData 获取数据
func (r *DataReplicator) GetData(dataID string) ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	data, ok := r.dataStore[dataID]
	if !ok {
		return nil, fmt.Errorf("数据不存在: %s", dataID)
	}
	return data, nil
}

// GetDataByHash 通过哈希获取数据
func (r *DataReplicator) GetDataByHash(dataHash string) ([]byte, error) {
	r.mu.RLock()
	dataID, ok := r.hashIndex[dataHash]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("数据哈希不存在: %s", dataHash)
	}

	return r.GetData(dataID)
}

// ReplicateData 复制数据
func (r *DataReplicator) ReplicateData(ctx context.Context, data map[string]interface{}, targetNodes []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 生成数据ID
	dataID := generateDataID()

	// 序列化数据
	dataBytes := serializeData(data)
	dataHash := r.StoreData(dataID, dataBytes)

	// 选择目标节点
	nodes := r.selectReplicaNodes(targetNodes, r.replicaLevel)

	// 创建复制记录
	replica := &ReplicaInfo{
		DataID:      dataID,
		DataHash:    dataHash,
		Nodes:       nodes,
		PrimaryNode: nodes[0],
		Status:      ReplicaStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Size:        int64(len(dataBytes)),
		Version:     1,
	}

	r.replicas[dataID] = replica

	// 异步复制到各节点
	for _, nodeID := range nodes {
		go r.replicateToNode(ctx, dataID, dataBytes, nodeID)
	}

	return nil
}

// selectReplicaNodes 选择复制节点
func (r *DataReplicator) selectReplicaNodes(candidates []string, level int) []string {
	if len(candidates) <= level {
		return candidates
	}

	// 选择前level个节点
	return candidates[:level]
}

// replicateToNode 复制到节点
func (r *DataReplicator) replicateToNode(ctx context.Context, dataID string, data []byte, nodeID string) {
	// 模拟复制 - 实际会通过网络发送
	time.Sleep(50 * time.Millisecond)

	r.mu.Lock()
	if replica, ok := r.replicas[dataID]; ok {
		// 检查是否所有节点都已完成
		complete := true
		for _, node := range replica.Nodes {
			// 简化检查
			if node != nodeID {
				complete = false
				break
			}
		}

		if len(replica.Nodes) > 0 {
			replica.Status = ReplicaStatusComplete
		}
		replica.UpdatedAt = time.Now()
	}
	r.mu.Unlock()
}

// VerifyReplica 验证复制
func (r *DataReplicator) VerifyReplica(dataID, nodeID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	replica, ok := r.replicas[dataID]
	if !ok {
		return false, fmt.Errorf("复制不存在: %s", dataID)
	}

	// 检查节点是否在复制列表中
	for _, node := range replica.Nodes {
		if node == nodeID {
			return true, nil
		}
	}

	return false, nil
}

// GetReplicaInfo 获取复制信息
func (r *DataReplicator) GetReplicaInfo(dataID string) (*ReplicaInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	replica, ok := r.replicas[dataID]
	if !ok {
		return nil, fmt.Errorf("复制不存在: %s", dataID)
	}
	return replica, nil
}

// ListReplicas 列出复制
func (r *DataReplicator) ListReplicas(status ReplicaStatus) []*ReplicaInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*ReplicaInfo, 0)
	for _, replica := range r.replicas {
		if status == "" || replica.Status == status {
			list = append(list, replica)
		}
	}
	return list
}

// RepairReplica 修复复制
func (r *DataReplicator) RepairReplica(ctx context.Context, dataID string) error {
	r.mu.Lock()
	replica, ok := r.replicas[dataID]
	if !ok {
		r.mu.Unlock()
		return fmt.Errorf("复制不存在: %s", dataID)
	}

	// 获取数据
	data, ok := r.dataStore[dataID]
	if !ok {
		r.mu.Unlock()
		return fmt.Errorf("数据不存在: %s", dataID)
	}

	// 重新复制到缺失节点
	for _, nodeID := range replica.Nodes {
		go r.replicateToNode(ctx, dataID, data, nodeID)
	}

	replica.Status = ReplicaStatusPending
	replica.UpdatedAt = time.Now()
	r.mu.Unlock()

	return nil
}

// DeleteReplica 删除复制
func (r *DataReplicator) DeleteReplica(dataID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 删除复制记录
	delete(r.replicas, dataID)

	// 删除数据
	delete(r.dataStore, dataID)

	return nil
}

// generateDataID 生成数据ID
func generateDataID() string {
	return fmt.Sprintf("data-%d", time.Now().UnixNano())
}

// serializeData 序列化数据
func serializeData(data map[string]interface{}) []byte {
	// 简化实现
	result := make([]byte, 0)
	for k, v := range data {
		result = append(result, []byte(k+":"+fmt.Sprintf("%v", v))...)
	}
	return result
}

// ReplicationStats 复制统计
type ReplicationStats struct {
	TotalReplicas    int     `json:"total_replicas"`
	CompleteReplicas int     `json:"complete_replicas"`
	PartialReplicas  int     `json:"partial_replicas"`
	PendingReplicas  int     `json:"pending_replicas"`
	FailedReplicas   int     `json:"failed_replicas"`
	TotalDataSize    int64   `json:"total_data_size"`
	AvgReplicaNodes  float64 `json:"avg_replica_nodes"`
}

// GetStats 获取统计
func (r *DataReplicator) GetStats() *ReplicationStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := &ReplicationStats{
		TotalReplicas: len(r.replicas),
	}

	if len(r.replicas) == 0 {
		return stats
	}

	totalNodes := 0
	for _, replica := range r.replicas {
		stats.TotalDataSize += replica.Size
		totalNodes += len(replica.Nodes)

		switch replica.Status {
		case ReplicaStatusComplete:
			stats.CompleteReplicas++
		case ReplicaStatusPartial:
			stats.PartialReplicas++
		case ReplicaStatusPending:
			stats.PendingReplicas++
		case ReplicaStatusFailed:
			stats.FailedReplicas++
		}
	}

	stats.AvgReplicaNodes = float64(totalNodes) / float64(len(r.replicas))

	return stats
}