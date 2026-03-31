// Package decentralized - Peer 管理
package decentralized

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// PeerManager Peer管理器
type PeerManager struct {
	peers       map[string]*Peer
 connections map[string]net.Conn
	messageQueue chan *PeerMessage
	handlers    map[string]MessageHandler
	mu          sync.RWMutex
}

// Peer Peer信息
type Peer struct {
	ID        string    `json:"id"`
	Address   string    `json:"address"`
	Port      int       `json:"port"`
	Status    PeerStatus `json:"status"`
	LastPing  time.Time `json:"last_ping"`
	Latency   int64     `json:"latency_ms"`
}

// PeerStatus Peer状态
type PeerStatus string

const (
	PeerStatusConnected    PeerStatus = "connected"
	PeerStatusDisconnected PeerStatus = "disconnected"
	PeerStatusPending      PeerStatus = "pending"
)

// MessageHandler 消息处理器
type MessageHandler func(msg *PeerMessage)

// NewPeerManager 创建Peer管理器
func NewPeerManager() *PeerManager {
	pm := &PeerManager{
		peers:        make(map[string]*Peer),
		connections:  make(map[string]net.Conn),
		messageQueue: make(chan *PeerMessage, 1000),
		handlers:     make(map[string]MessageHandler),
	}

	// 启动消息处理
	go pm.processMessages()

	return pm
}

// AddPeer 添加Peer
func (pm *PeerManager) AddPeer(id, address string, port int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.peers[id] = &Peer{
		ID:       id,
		Address:  address,
		Port:     port,
		Status:   PeerStatusPending,
		LastPing: time.Now(),
	}

	// 异步连接
	go pm.connectPeer(id, address, port)
}

// connectPeer 连接Peer
func (pm *PeerManager) connectPeer(id, address string, port int) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", address, port), 5*time.Second)
	if err != nil {
		pm.mu.Lock()
		if peer, ok := pm.peers[id]; ok {
			peer.Status = PeerStatusDisconnected
		}
		pm.mu.Unlock()
		return
	}

	pm.mu.Lock()
	pm.connections[id] = conn
	if peer, ok := pm.peers[id]; ok {
		peer.Status = PeerStatusConnected
	}
	pm.mu.Unlock()

	// 启动接收循环
	go pm.receiveMessages(id, conn)
}

// RemovePeer 移除Peer
func (pm *PeerManager) RemovePeer(id string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 关闭连接
	if conn, ok := pm.connections[id]; ok {
		conn.Close()
		delete(pm.connections, id)
	}

	delete(pm.peers, id)
}

// SendMessage 发送消息
func (pm *PeerManager) SendMessage(ctx context.Context, peerID string, msg *PeerMessage) error {
	pm.mu.RLock()
	conn, ok := pm.connections[peerID]
	pm.mu.RUnlock()

	if !ok {
		return fmt.Errorf("Peer未连接: %s", peerID)
	}

	// 序列化消息
	data := serializeMessage(msg)

	// 发送
	_, err := conn.Write(data)
	return err
}

// BroadcastMessage 广播消息
func (pm *PeerManager) BroadcastMessage(msg *PeerMessage) {
	pm.mu.RLock()
	peers := make([]string, 0, len(pm.peers))
	for id := range pm.peers {
		peers = append(peers, id)
	}
	pm.mu.RUnlock()

	for _, peerID := range peers {
		pm.SendMessage(context.Background(), peerID, msg)
	}
}

// receiveMessages 接收消息
func (pm *PeerManager) receiveMessages(peerID string, conn net.Conn) {
	buffer := make([]byte, 4096)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			pm.RemovePeer(peerID)
			return
		}

		msg := deserializeMessage(buffer[:n])
		msg.FromNode = peerID
		pm.messageQueue <- msg
	}
}

// processMessages 处理消息队列
func (pm *PeerManager) processMessages() {
	for msg := range pm.messageQueue {
		pm.mu.RLock()
		handler, ok := pm.handlers[msg.Type]
		pm.mu.RUnlock()

		if ok {
			handler(msg)
		}
	}
}

// RegisterHandler 注册消息处理器
func (pm *PeerManager) RegisterHandler(msgType string, handler MessageHandler) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.handlers[msgType] = handler
}

// PingPeer Ping Peer
func (pm *PeerManager) PingPeer(peerID string) (int64, error) {
	pm.mu.RLock()
	conn, ok := pm.connections[peerID]
	pm.mu.RUnlock()

	if !ok {
		return 0, fmt.Errorf("Peer未连接: %s", peerID)
	}

	start := time.Now()

	// 发送Ping
	pingMsg := &PeerMessage{
		Type:      "ping",
		Timestamp: start,
	}

	err := pm.SendMessage(context.Background(), peerID, pingMsg)
	if err != nil {
		return 0, err
	}

	latency := time.Since(start).Milliseconds()

	pm.mu.Lock()
	if peer, ok := pm.peers[peerID]; ok {
		peer.LastPing = time.Now()
		peer.Latency = latency
	}
	pm.mu.Unlock()

	return latency, nil
}

// GetPeer 获取Peer
func (pm *PeerManager) GetPeer(peerID string) (*Peer, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	peer, ok := pm.peers[peerID]
	if !ok {
		return nil, fmt.Errorf("Peer不存在: %s", peerID)
	}
	return peer, nil
}

// ListPeers 列出Peers
func (pm *PeerManager) ListPeers(status PeerStatus) []*Peer {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	list := make([]*Peer, 0)
	for _, peer := range pm.peers {
		if status == "" || peer.Status == status {
			list = append(list, peer)
		}
	}
	return list
}

// serializeMessage 序列化消息
func serializeMessage(msg *PeerMessage) []byte {
	// 简化实现 - 实际会使用protobuf或JSON
	data := fmt.Sprintf("%s|%s|%d", msg.Type, msg.FromNode, msg.Timestamp.Unix())
	return []byte(data)
}

// deserializeMessage 反序列化消息
func deserializeMessage(data []byte) *PeerMessage {
	// 简化实现
	return &PeerMessage{
		Type:      "unknown",
		Timestamp: time.Now(),
	}
}

// StartHealthCheck 启动健康检查
func (pm *PeerManager) StartHealthCheck(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			pm.mu.RLock()
			peers := make([]string, 0, len(pm.peers))
			for id := range pm.peers {
				peers = append(peers, id)
			}
			pm.mu.RUnlock()

			for _, peerID := range peers {
				_, err := pm.PingPeer(peerID)
				if err != nil {
					pm.RemovePeer(peerID)
				}
			}
		}
	}()
}