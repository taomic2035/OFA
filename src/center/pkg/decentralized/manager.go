// Package decentralized - 去中心化管理
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

// NetworkType 网络类型
type NetworkType string

const (
	NetworkFullP2P   NetworkType = "full_p2p"   // 全P2P网络
	NetworkHybrid    NetworkType = "hybrid"     // 混合网络(Center + P2P)
	NetworkFederated NetworkType = "federated"  // 联邦网络
	NetworkMesh      NetworkType = "mesh"       // 网状网络
)

// DecentralizedManager 去中心化管理器
type DecentralizedManager struct {
	networkType  NetworkType
	nodes        map[string]*NodeInfo
	peers        *PeerManager
	consensus    *ConsensusEngine
	replication  *DataReplicator
	discovery    *PeerDiscovery
	syncManager  *SyncManager
	trustManager *TrustManager
	mu           sync.RWMutex
}

// NodeInfo 节点信息
type NodeInfo struct {
	ID           string            `json:"id"`
	Address      string            `json:"address"`
	Port         int               `json:"port"`
	Type         NodeType          `json:"type"`
	Status       NodeStatus        `json:"status"`
	Capabilities []string          `json:"capabilities"`
	Load         int               `json:"load"`
	MaxLoad      int               `json:"max_load"`
	TrustScore   float64           `json:"trust_score"`
	LastSeen     time.Time         `json:"last_seen"`
	JoinedAt     time.Time         `json:"joined_at"`
	Metadata     map[string]string `json:"metadata"`
	PublicKey    string            `json:"public_key"`
}

// NodeType 节点类型
type NodeType string

const (
	NodeTypeFull    NodeType = "full"    // 全节点(存储完整数据)
	NodeTypeLight   NodeType = "light"   // 轻节点(仅验证)
	NodeTypeEdge    NodeType = "edge"    // 边缘节点(处理本地数据)
	NodeTypeArchive NodeType = "archive" // 归档节点(历史数据)
)

// NodeStatus 节点状态
type NodeStatus string

const (
	NodeStatusOnline    NodeStatus = "online"
	NodeStatusOffline   NodeStatus = "offline"
	NodeStatusBusy      NodeStatus = "busy"
	NodeStatusSyncing   NodeStatus = "syncing"
	NodeStatusDegraded  NodeStatus = "degraded"
)

// NewDecentralizedManager 创建去中心化管理器
func NewDecentralizedManager(networkType NetworkType) *DecentralizedManager {
	return &DecentralizedManager{
		networkType:  networkType,
		nodes:        make(map[string]*NodeInfo),
		peers:        NewPeerManager(),
		consensus:    NewConsensusEngine(),
		replication:  NewDataReplicator(),
		discovery:    NewPeerDiscovery(),
		syncManager:  NewSyncManager(),
		trustManager: NewTrustManager(),
	}
}

// JoinNetwork 加入网络
func (m *DecentralizedManager) JoinNetwork(ctx context.Context, nodeInfo *NodeInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 验证节点
	if err := m.validateNode(nodeInfo); err != nil {
		return err
	}

	// 生成节点ID
	if nodeInfo.ID == "" {
		nodeInfo.ID = generateNodeID(nodeInfo.Address, nodeInfo.Port)
	}

	nodeInfo.JoinedAt = time.Now()
	nodeInfo.LastSeen = time.Now()
	nodeInfo.Status = NodeStatusOnline

	m.nodes[nodeInfo.ID] = nodeInfo

	// 添加到Peer管理
	m.peers.AddPeer(nodeInfo.ID, nodeInfo.Address, nodeInfo.Port)

	// 发现其他节点
	m.discovery.DiscoverPeers(ctx, nodeInfo.ID)

	// 同步数据
	m.syncManager.StartSync(ctx, nodeInfo.ID)

	return nil
}

// LeaveNetwork 离开网络
func (m *DecentralizedManager) LeaveNetwork(nodeID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	node, ok := m.nodes[nodeID]
	if !ok {
		return fmt.Errorf("节点不存在: %s", nodeID)
	}

	node.Status = NodeStatusOffline
	m.peers.RemovePeer(nodeID)

	// 通知其他节点
	m.peers.BroadcastMessage(&PeerMessage{
		Type:      "node_leave",
		FromNode:  nodeID,
		Timestamp: time.Now(),
	})

	return nil
}

// GetNode 获取节点信息
func (m *DecentralizedManager) GetNode(nodeID string) (*NodeInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	node, ok := m.nodes[nodeID]
	if !ok {
		return nil, fmt.Errorf("节点不存在: %s", nodeID)
	}
	return node, nil
}

// ListNodes 列出所有节点
func (m *DecentralizedManager) ListNodes(status NodeStatus) []*NodeInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	list := make([]*NodeInfo, 0)
	for _, node := range m.nodes {
		if status == "" || node.Status == status {
			list = append(list, node)
		}
	}
	return list
}

// UpdateNodeStatus 更新节点状态
func (m *DecentralizedManager) UpdateNodeStatus(nodeID string, status NodeStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	node, ok := m.nodes[nodeID]
	if !ok {
		return fmt.Errorf("节点不存在: %s", nodeID)
	}

	node.Status = status
	node.LastSeen = time.Now()

	return nil
}

// validateNode 验证节点
func (m *DecentralizedManager) validateNode(node *NodeInfo) error {
	if node.Address == "" {
		return fmt.Errorf("节点地址不能为空")
	}
	if node.Port <= 0 || node.Port > 65535 {
		return fmt.Errorf("端口无效: %d", node.Port)
	}
	return nil
}

// generateNodeID 生成节点ID
func generateNodeID(address string, port int) string {
	data := fmt.Sprintf("%s:%d:%d", address, port, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

// NetworkStats 网络统计
type NetworkStats struct {
	TotalNodes     int       `json:"total_nodes"`
	OnlineNodes    int       `json:"online_nodes"`
	OfflineNodes   int       `json:"offline_nodes"`
	FullNodes      int       `json:"full_nodes"`
	LightNodes     int       `json:"light_nodes"`
	EdgeNodes      int       `json:"edge_nodes"`
	AvgTrustScore  float64   `json:"avg_trust_score"`
	NetworkHealth  float64   `json:"network_health"`
	LastUpdate     time.Time `json:"last_update"`
}

// GetStats 获取网络统计
func (m *DecentralizedManager) GetStats() *NetworkStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := &NetworkStats{
		TotalNodes: len(m.nodes),
		LastUpdate: time.Now(),
	}

	if len(m.nodes) == 0 {
		return stats
	}

	totalTrust := 0.0
	for _, node := range m.nodes {
		if node.Status == NodeStatusOnline {
			stats.OnlineNodes++
		} else {
			stats.OfflineNodes++
		}

		switch node.Type {
		case NodeTypeFull:
			stats.FullNodes++
		case NodeTypeLight:
			stats.LightNodes++
		case NodeTypeEdge:
			stats.EdgeNodes++
		}

		totalTrust += node.TrustScore
	}

	stats.AvgTrustScore = totalTrust / float64(len(m.nodes))

	// 计算网络健康度
	if stats.TotalNodes > 0 {
		stats.NetworkHealth = float64(stats.OnlineNodes) / float64(stats.TotalNodes) * stats.AvgTrustScore
	}

	return stats
}

// DistributeTask 分布式任务
func (m *DecentralizedManager) DistributeTask(ctx context.Context, task *DistributedTask) (*TaskResult, error) {
	// 选择执行节点
	nodes := m.selectExecutionNodes(task)

	if len(nodes) == 0 {
		return nil, fmt.Errorf("没有可用节点执行任务")
	}

	// 使用共识机制确定执行方案
	proposal := m.consensus.CreateProposal("task_execution", task, nodes[0].ID)

	// 获取节点投票
	for _, node := range nodes {
		vote := m.consensus.EvaluateProposal(proposal, node)
		proposal.Votes = append(proposal.Votes, vote)
	}

	// 执行共识决策
	decision, err := m.consensus.MakeDecision(proposal)
	if err != nil {
		return nil, err
	}

	// 分配任务到选定节点
	executorID := decision.SelectedNode

	// 复制任务数据
	m.replication.ReplicateData(ctx, task.Data, []string{executorID})

	// 发送任务到执行节点
	m.peers.SendMessage(ctx, executorID, &PeerMessage{
		Type:      "task_execute",
		FromNode:  "manager",
		ToNode:    executorID,
		Data:      task,
		Timestamp: time.Now(),
	})

	// 等待结果
	result, err := m.waitForResult(ctx, task.ID, 30*time.Second)
	if err != nil {
		return nil, err
	}

	// 更新信任评分
	m.trustManager.UpdateTrustScore(executorID, result.Success)

	return result, nil
}

// selectExecutionNodes 选择执行节点
func (m *DecentralizedManager) selectExecutionNodes(task *DistributedTask) []*NodeInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	candidates := make([]*NodeInfo, 0)

	for _, node := range m.nodes {
		// 检查状态
		if node.Status != NodeStatusOnline {
			continue
		}

		// 检查能力
		for _, cap := range node.Capabilities {
			if cap == task.RequiredCapability {
				candidates = append(candidates, node)
				break
			}
		}
	}

	// 按信任评分和负载排序
	return sortByTrustAndLoad(candidates)
}

// waitForResult 等待结果
func (m *DecentralizedManager) waitForResult(ctx context.Context, taskID string, timeout time.Duration) (*TaskResult, error) {
	resultChan := make(chan *TaskResult, 1)

	go func() {
		// 模等待结果 - 实际实现会监听消息
		time.Sleep(100 * time.Millisecond)
		resultChan <- &TaskResult{
			TaskID:  taskID,
			Success: true,
			Output:  map[string]interface{}{"status": "completed"},
		}
	}()

	select {
	case result := <-resultChan:
		return result, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("等待结果超时")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// sortByTrustAndLoad 按信任和负载排序
func sortByTrustAndLoad(nodes []*NodeInfo) []*NodeInfo {
	if len(nodes) <= 1 {
		return nodes
	}

	// 简单排序
	for i := 0; i < len(nodes)-1; i++ {
		for j := i + 1; j < len(nodes); j++ {
			scoreI := nodes[i].TrustScore * (1.0 - float64(nodes[i].Load)/float64(nodes[i].MaxLoad+1))
			scoreJ := nodes[j].TrustScore * (1.0 - float64(nodes[j].Load)/float64(nodes[j].MaxLoad+1))
			if scoreJ > scoreI {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}

	return nodes
}

// DistributedTask 分布式任务
type DistributedTask struct {
	ID               string                 `json:"id"`
	Type             string                 `json:"type"`
	RequiredCapability string               `json:"required_capability"`
	Priority         int                    `json:"priority"`
	Data             map[string]interface{} `json:"data"`
	ReplicationLevel int                    `json:"replication_level"`
	Timeout          time.Duration          `json:"timeout"`
	CreatedAt        time.Time              `json:"created_at"`
}

// TaskResult 任务结果
type TaskResult struct {
	TaskID    string                 `json:"task_id"`
	Success   bool                   `json:"success"`
	Output    map[string]interface{} `json:"output"`
	Error     string                 `json:"error,omitempty"`
	Executor  string                 `json:"executor"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
}

// PeerMessage Peer消息
type PeerMessage struct {
	Type      string                 `json:"type"`
	FromNode  string                 `json:"from_node"`
	ToNode    string                 `json:"to_node,omitempty"`
	Data      interface{}            `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}