// Package decentralized - Peer发现
// 0.9.0 Beta: 去中心化增强
package decentralized

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PeerDiscovery Peer发现器
type PeerDiscovery struct {
	method       DiscoveryMethod
	knownPeers   map[string]*PeerEndpoint
	bootstraps   []string // 启动节点
	localNodeID  string
	discoveryLog []*DiscoveryEvent
	mu           sync.RWMutex
}

// DiscoveryMethod 发现方法
type DiscoveryMethod string

const (
	DiscoveryDHT       DiscoveryMethod = "dht"       // DHT发现
	DiscoveryBootstrap DiscoveryMethod = "bootstrap" // 启动节点发现
	DiscoveryBroadcast DiscoveryMethod = "broadcast" // 广播发现
	DiscoveryGossip    DiscoveryMethod = "gossip"    // Gossip协议
	DiscoveryRegistry  DiscoveryMethod = "registry"  // 注册中心
)

// PeerEndpoint Peer端点
type PeerEndpoint struct {
	ID        string    `json:"id"`
	Address   string    `json:"address"`
	Port      int       `json:"port"`
	Source    string    `json:"source"`    // 发现来源
	LastSeen  time.Time `json:"last_seen"`
	Status    string    `json:"status"`
}

// DiscoveryEvent 发现事件
type DiscoveryEvent struct {
	Type      string    `json:"type"`
	PeerID    string    `json:"peer_id"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

// NewPeerDiscovery 创建Peer发现器
func NewPeerDiscovery() *PeerDiscovery {
	return &PeerDiscovery{
		method:       DiscoveryBootstrap,
		knownPeers:   make(map[string]*PeerEndpoint),
		bootstraps:   make([]string, 0),
		discoveryLog: make([]*DiscoveryEvent, 0),
	}
}

// SetMethod 设置发现方法
func (d *PeerDiscovery) SetMethod(method DiscoveryMethod) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.method = method
}

// SetLocalNodeID 设置本地节点ID
func (d *PeerDiscovery) SetLocalNodeID(nodeID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.localNodeID = nodeID
}

// AddBootstrap 添加启动节点
func (d *PeerDiscovery) AddBootstrap(address string, port int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	endpoint := fmt.Sprintf("%s:%d", address, port)
	d.bootstraps = append(d.bootstraps, endpoint)
}

// DiscoverPeers 发现Peers
func (d *PeerDiscovery) DiscoverPeers(ctx context.Context, localNodeID string) ([]*PeerEndpoint, error) {
	d.mu.Lock()
	d.localNodeID = localNodeID
	method := d.method
	bootstraps := d.bootstraps
	d.mu.Unlock()

	var peers []*PeerEndpoint

	switch method {
	case DiscoveryBootstrap:
		peers = d.discoverFromBootstrap(ctx, bootstraps)

	case DiscoveryDHT:
		peers = d.discoverFromDHT(ctx)

	case DiscoveryBroadcast:
		peers = d.discoverFromBroadcast(ctx)

	case DiscoveryGossip:
		peers = d.discoverFromGossip(ctx)

	case DiscoveryRegistry:
		peers = d.discoverFromRegistry(ctx)

	default:
		peers = d.discoverFromBootstrap(ctx, bootstraps)
	}

	// 添加发现的Peers
	d.mu.Lock()
	for _, peer := range peers {
		if peer.ID != d.localNodeID {
			d.knownPeers[peer.ID] = peer
			d.logDiscovery("found", peer.ID, peer.Source)
		}
	}
	d.mu.Unlock()

	return peers, nil
}

// discoverFromBootstrap 从启动节点发现
func (d *PeerDiscovery) discoverFromBootstrap(ctx context.Context, bootstraps []string) []*PeerEndpoint {
	peers := make([]*PeerEndpoint, 0)

	for _, endpoint := range bootstraps {
		// 模拟连接启动节点并获取Peer列表
		peer := &PeerEndpoint{
			ID:       generatePeerID(endpoint),
			Address:  parseAddress(endpoint),
			Port:     parsePort(endpoint),
			Source:   "bootstrap",
			LastSeen: time.Now(),
			Status:   "online",
		}
		peers = append(peers, peer)
	}

	return peers
}

// discoverFromDHT DHT发现
func (d *PeerDiscovery) discoverFromDHT(ctx context.Context) []*PeerEndpoint {
	// 简化实现 - DHT查询
	peers := make([]*PeerEndpoint, 0)

	// 模拟DHT查询
	for i := 0; i < 5; i++ {
		peer := &PeerEndpoint{
			ID:       generatePeerID(fmt.Sprintf("dht-node-%d", i)),
			Address:  fmt.Sprintf("10.0.0.%d", i+1),
			Port:     9090,
			Source:   "dht",
			LastSeen: time.Now(),
			Status:   "online",
		}
		peers = append(peers, peer)
	}

	return peers
}

// discoverFromBroadcast 广播发现
func (d *PeerDiscovery) discoverFromBroadcast(ctx context.Context) []*PeerEndpoint {
	// 简化实现 - 广播发现
	peers := make([]*PeerEndpoint, 0)

	// 模拟广播响应
	peer := &PeerEndpoint{
		ID:       generatePeerID("broadcast-peer"),
		Address:  "10.0.0.100",
		Port:     9090,
		Source:   "broadcast",
		LastSeen: time.Now(),
		Status:   "online",
	}
	peers = append(peers, peer)

	return peers
}

// discoverFromGossip Gossip发现
func (d *PeerDiscovery) discoverFromGossip(ctx context.Context) []*PeerEndpoint {
	// 简化实现 - Gossip协议
	peers := make([]*PeerEndpoint, 0)

	d.mu.RLock()
	for _, peer := range d.knownPeers {
		// 从已知Peers获取更多Peers
		newPeer := &PeerEndpoint{
			ID:       generatePeerID(peer.Address + "-gossip"),
			Address:  peer.Address,
			Port:     peer.Port + 1,
			Source:   "gossip",
			LastSeen: time.Now(),
			Status:   "online",
		}
		peers = append(peers, newPeer)
	}
	d.mu.RUnlock()

	return peers
}

// discoverFromRegistry 注册中心发现
func (d *PeerDiscovery) discoverFromRegistry(ctx context.Context) []*PeerEndpoint {
	// 简化实现 - 从注册中心获取
	peers := make([]*PeerEndpoint, 0)

	for i := 0; i < 3; i++ {
		peer := &PeerEndpoint{
			ID:       generatePeerID(fmt.Sprintf("registry-node-%d", i)),
			Address:  fmt.Sprintf("registry.%d.example.com", i),
			Port:     9090,
			Source:   "registry",
			LastSeen: time.Now(),
			Status:   "online",
		}
		peers = append(peers, peer)
	}

	return peers
}

// logDiscovery 记录发现日志
func (d *PeerDiscovery) logDiscovery(eventType, peerID, source string) {
	event := &DiscoveryEvent{
		Type:      eventType,
		PeerID:    peerID,
		Timestamp: time.Now(),
		Source:    source,
	}
	d.discoveryLog = append(d.discoveryLog, event)
}

// AddPeer 手动添加Peer
func (d *PeerDiscovery) AddPeer(peerID, address string, port int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.knownPeers[peerID] = &PeerEndpoint{
		ID:       peerID,
		Address:  address,
		Port:     port,
		Source:   "manual",
		LastSeen: time.Now(),
		Status:   "online",
	}

	d.logDiscovery("added", peerID, "manual")
}

// RemovePeer 移除Peer
func (d *PeerDiscovery) RemovePeer(peerID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.knownPeers, peerID)
	d.logDiscovery("removed", peerID, "manual")
}

// GetPeer 获取Peer
func (d *PeerDiscovery) GetPeer(peerID string) (*PeerEndpoint, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	peer, ok := d.knownPeers[peerID]
	if !ok {
		return nil, fmt.Errorf("Peer不存在: %s", peerID)
	}
	return peer, nil
}

// ListPeers 列出Peers
func (d *PeerDiscovery) ListPeers() []*PeerEndpoint {
	d.mu.RLock()
	defer d.mu.RUnlock()

	list := make([]*PeerEndpoint, 0, len(d.knownPeers))
	for _, peer := range d.knownPeers {
		list = append(list, peer)
	}
	return list
}

// RefreshPeer 刷新Peer
func (d *PeerDiscovery) RefreshPeer(ctx context.Context, peerID string) error {
	d.mu.RLock()
	peer, ok := d.knownPeers[peerID]
	d.mu.RUnlock()

	if !ok {
		return fmt.Errorf("Peer不存在: %s", peerID)
	}

	// 模拟刷新
	peer.LastSeen = time.Now()

	d.mu.Lock()
	d.knownPeers[peerID] = peer
	d.logDiscovery("refreshed", peerID, "refresh")
	d.mu.Unlock()

	return nil
}

// StartContinuousDiscovery 启动持续发现
func (d *PeerDiscovery) StartContinuousDiscovery(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				d.mu.RLock()
				localID := d.localNodeID
				d.mu.RUnlock()

				if localID != "" {
					d.DiscoverPeers(ctx, localID)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// DiscoveryStats 发现统计
type DiscoveryStats struct {
	TotalPeers       int       `json:"total_peers"`
	OnlinePeers      int       `json:"online_peers"`
	DiscoveryEvents  int       `json:"discovery_events"`
	LastDiscovery    time.Time `json:"last_discovery"`
	BootstrapNodes   int       `json:"bootstrap_nodes"`
	Method           string    `json:"method"`
}

// GetStats 获取统计
func (d *PeerDiscovery) GetStats() *DiscoveryStats {
	d.mu.RLock()
	defer d.mu.RUnlock()

	stats := &DiscoveryStats{
		TotalPeers:      len(d.knownPeers),
		DiscoveryEvents: len(d.discoveryLog),
		BootstrapNodes:  len(d.bootstraps),
		Method:          string(d.method),
	}

	for _, peer := range d.knownPeers {
		if peer.Status == "online" {
			stats.OnlinePeers++
		}
	}

	if len(d.discoveryLog) > 0 {
		stats.LastDiscovery = d.discoveryLog[len(d.discoveryLog)-1].Timestamp
	}

	return stats
}

// generatePeerID 生成Peer ID
func generatePeerID(source string) string {
	return fmt.Sprintf("peer-%s-%d", source[:8], time.Now().UnixNano()%10000)
}

// parseAddress 解析地址
func parseAddress(endpoint string) string {
	// 简化实现
	return "10.0.0.1"
}

// parsePort 解析端口
func parsePort(endpoint string) int {
	// 简化实现
	return 9090
}