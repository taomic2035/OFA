// Package lite - P2P通信(轻量级，适合手表/手环)
package lite

import (
	"encoding/json"
	"log"
	"net"
	"sync"
	"time"
)

// P2PConfig P2P配置
type P2PConfig struct {
	ListenPort    int  // 监听端口
	DiscoveryPort int  // 发现端口
	EnableBLE     bool // 启用BLE
	EnableWiFi    bool // 启用WiFi
	PowerSave     bool // 省电模式
}

// DefaultP2PConfig 默认配置
func DefaultP2PConfig() P2PConfig {
	return P2PConfig{
		ListenPort:    7891,
		DiscoveryPort: 7890,
		EnableBLE:     true,
		EnableWiFi:    false,
		PowerSave:     true,
	}
}

// Peer 对等节点(简化版)
type Peer struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Port      int       `json:"port"`
	Type      string    `json:"type"` // watch, phone, tablet
	LastSeen  time.Time `json:"last_seen"`
	IsOnline  bool      `json:"is_online"`
}

// P2PMessage P2P消息(简化版)
type P2PMessage struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"` // discovery, task, data, ack
	From      string      `json:"from"`
	To        string      `json:"to"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// P2PClient P2P客户端(轻量级)
type P2PClient struct {
	config   P2PConfig
	agentID  string

	peers    map[string]*Peer
	handlers map[string]func(*P2PMessage)

	udpConn  *net.UDPConn
	tcpLn    net.Listener

	onPeerOnline  func(*Peer)
	onPeerOffline func(*Peer)

	mu      sync.RWMutex
	running bool
}

// NewP2PClient 创建P2P客户端
func NewP2PClient(agentID string, config P2PConfig) *P2PClient {
	return &P2PClient{
		config:   config,
		agentID:  agentID,
		peers:    make(map[string]*Peer),
		handlers: make(map[string]func(*P2PMessage)),
	}
}

// Start 启动P2P
func (p *P2PClient) Start() error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return nil
	}
	p.running = true
	p.mu.Unlock()

	// 启动UDP发现
	if p.config.EnableWiFi {
		go p.udpDiscoveryLoop()
	}

	// 启动TCP监听
	if p.config.EnableWiFi {
		go p.tcpListenLoop()
	}

	// 启动超时检测
	go p.peerTimeoutLoop()

	log.Printf("P2P client started (BLE: %v, WiFi: %v)", p.config.EnableBLE, p.config.EnableWiFi)
	return nil
}

// Stop 停止P2P
func (p *P2PClient) Stop() {
	p.mu.Lock()
	p.running = false
	p.mu.Unlock()

	if p.udpConn != nil {
		p.udpConn.Close()
	}
	if p.tcpLn != nil {
		p.tcpLn.Close()
	}
}

// GetPeers 获取对等节点
func (p *P2PClient) GetPeers() []*Peer {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var peers []*Peer
	for _, peer := range p.peers {
		if peer.IsOnline {
			peers = append(peers, peer)
		}
	}
	return peers
}

// GetPeer 获取特定节点
func (p *P2PClient) GetPeer(id string) *Peer {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.peers[id]
}

// Send 发送消息
func (p *P2PClient) Send(peerID string, msg *P2PMessage) error {
	p.mu.RLock()
	peer, ok := p.peers[peerID]
	if !ok || !peer.IsOnline {
		p.mu.RUnlock()
		return ErrPeerOffline
	}
	p.mu.RUnlock()

	// 如果是BLE模式，使用BLE发送
	if peer.Type == "phone" && p.config.EnableBLE {
		return p.sendViaBLE(peer, msg)
	}

	// TCP发送
	return p.sendViaTCP(peer, msg)
}

// Broadcast 广播消息
func (p *P2PClient) Broadcast(msg *P2PMessage) error {
	peers := p.GetPeers()
	for _, peer := range peers {
		p.Send(peer.ID, msg)
	}
	return nil
}

// OnMessage 注册消息处理器
func (p *P2PClient) OnMessage(msgType string, handler func(*P2PMessage)) {
	p.mu.Lock()
	p.handlers[msgType] = handler
	p.mu.Unlock()
}

// OnPeerOnline 设置节点上线回调
func (p *P2PClient) OnPeerOnline(handler func(*Peer)) {
	p.mu.Lock()
	p.onPeerOnline = handler
	p.mu.Unlock()
}

// OnPeerOffline 设置节点下线回调
func (p *P2PClient) OnPeerOffline(handler func(*Peer)) {
	p.mu.Lock()
	p.onPeerOffline = handler
	p.mu.Unlock()
}

// Discover 手动发现
func (p *P2PClient) Discover() {
	p.sendDiscoveryBroadcast()
}

func (p *P2PClient) sendViaBLE(peer *Peer, msg *P2PMessage) error {
	// BLE发送由平台实现
	// 这里是模拟
	log.Printf("BLE send to %s: %s", peer.ID, msg.Type)
	return nil
}

func (p *P2PClient) sendViaTCP(peer *Peer, msg *P2PMessage) error {
	if p.config.PowerSave {
		// 省电模式：批量发送，减少连接次数
		// 实际实现需要缓冲队列
	}

	addr := peer.Address + ":" + itoa(peer.Port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	msg.Timestamp = time.Now()
	msg.From = p.agentID

	data, _ := json.Marshal(msg)
	conn.Write(data)

	return nil
}

func (p *P2PClient) udpDiscoveryLoop() {
	// 绑定UDP
	addr := &net.UDPAddr{Port: p.config.DiscoveryPort}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Printf("UDP listen failed: %v", err)
		return
	}
	p.udpConn = conn

	buf := make([]byte, 1024)
	for {
		p.mu.RLock()
		running := p.running
		p.mu.RUnlock()

		if !running {
			return
		}

		n, remoteAddr, err := p.udpConn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		var msg P2PMessage
		if json.Unmarshal(buf[:n], &msg) != nil {
			continue
		}

		if msg.Type == "discovery" {
			p.handleDiscovery(&msg, remoteAddr)
		}
	}
}

func (p *P2PClient) tcpListenLoop() {
	ln, err := net.Listen("tcp", ":"+itoa(p.config.ListenPort))
	if err != nil {
		log.Printf("TCP listen failed: %v", err)
		return
	}
	p.tcpLn = ln

	for {
		p.mu.RLock()
		running := p.running
		p.mu.RUnlock()

		if !running {
			return
		}

		conn, err := p.tcpLn.Accept()
		if err != nil {
			continue
		}

		go p.handleTCPConn(conn)
	}
}

func (p *P2PClient) handleTCPConn(conn net.Conn) {
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	var msg P2PMessage
	if json.Unmarshal(buf[:n], &msg) == nil {
		p.handleMessage(&msg)
	}
}

func (p *P2PClient) handleDiscovery(msg *P2PMessage, addr *net.UDPAddr) {
	// 解析节点信息
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		return
	}

	peer := &Peer{
		ID:       getString(data, "id"),
		Name:     getString(data, "name"),
		Address:  addr.IP.String(),
		Port:     getInt(data, "port"),
		Type:     getString(data, "type"),
		LastSeen: time.Now(),
		IsOnline: true,
	}

	if peer.ID == p.agentID {
		return // 忽略自己
	}

	p.mu.Lock()
	isNew := p.peers[peer.ID] == nil
	p.peers[peer.ID] = peer
	onOnline := p.onPeerOnline
	p.mu.Unlock()

	if isNew && onOnline != nil {
		go onOnline(peer)
	}

	// 发送响应
	p.sendDiscoveryAck(addr)
}

func (p *P2PClient) sendDiscoveryBroadcast() {
	self := Peer{
		ID:   p.agentID,
		Port: p.config.ListenPort,
		Type: "watch",
	}

	msg := &P2PMessage{
		ID:   generateLiteID(),
		Type: "discovery",
		From: p.agentID,
		Data: self,
	}

	data, _ := json.Marshal(msg)

	broadcastAddr := &net.UDPAddr{
		IP:   net.IPv4(255, 255, 255, 255),
		Port: p.config.DiscoveryPort,
	}

	if p.udpConn != nil {
		p.udpConn.WriteToUDP(data, broadcastAddr)
	}
}

func (p *P2PClient) sendDiscoveryAck(addr *net.UDPAddr) {
	msg := &P2PMessage{
		ID:   generateLiteID(),
		Type: "discovery_ack",
		From: p.agentID,
	}

	data, _ := json.Marshal(msg)
	if p.udpConn != nil {
		p.udpConn.WriteToUDP(data, addr)
	}
}

func (p *P2PClient) handleMessage(msg *P2PMessage) {
	p.mu.RLock()
	handler := p.handlers[msg.Type]
	p.mu.RUnlock()

	if handler != nil {
		handler(msg)
	}
}

func (p *P2PClient) peerTimeoutLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		p.mu.RLock()
		running := p.running
		p.mu.RUnlock()

		if !running {
			return
		}

		<-ticker.C

		p.mu.Lock()
		now := time.Now()
		for id, peer := range p.peers {
			if now.Sub(peer.LastSeen) > 60*time.Second {
				peer.IsOnline = false
				if p.onPeerOffline != nil {
					go p.onPeerOffline(peer)
				}
				delete(p.peers, id)
			}
		}
		p.mu.Unlock()
	}
}

// RequestTask 向对等节点请求执行任务
func (p *P2PClient) RequestTask(peerID, skill, action string, params interface{}) (interface{}, error) {
	msg := &P2PMessage{
		ID:   generateLiteID(),
		Type: "task",
		From: p.agentID,
		To:   peerID,
		Data: map[string]interface{}{
			"skill":  skill,
			"action": action,
			"params": params,
		},
	}

	if err := p.Send(peerID, msg); err != nil {
		return nil, err
	}

	// 等待响应(简化)
	return map[string]string{"status": "sent"}, nil
}

// Errors
var ErrPeerOffline = &P2PError{Message: "peer offline"}

type P2PError struct {
	Message string
}

func (e *P2PError) Error() string { return e.Message }

// Helpers
func itoa(i int) string {
	if i <= 0 {
		return "0"
	}
	var buf [10]byte
	n := len(buf)
	for i > 0 {
		n--
		buf[n] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[n:])
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case int:
			return n
		case float64:
			return int(n)
		}
	}
	return 0
}