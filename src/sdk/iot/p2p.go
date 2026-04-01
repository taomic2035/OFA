// Package iot - P2P通信
package iot

import (
	"encoding/json"
	"log"
	"net"
	"sync"
	"time"
)

// P2PMode P2P模式
type P2PMode int

const (
	P2PDisabled P2PMode = iota
	P2PDiscovery // 仅发现
	P2PMesh      // Mesh网络
)

// P2PConfig P2P配置
type P2PConfig struct {
	Mode          P2PMode `json:"mode"`
	ListenPort    int     `json:"listen_port"`
	DiscoveryPort int     `json:"discovery_port"`
	MaxPeers      int     `json:"max_peers"`
	HeartbeatInt  int     `json:"heartbeat_int"` // 心跳间隔(秒)
}

// DefaultP2PConfig 默认配置
func DefaultP2PConfig() P2PConfig {
	return P2PConfig{
		Mode:          P2PDiscovery,
		ListenPort:    7891,
		DiscoveryPort: 7890,
		MaxPeers:      10,
		HeartbeatInt:  30,
	}
}

// PeerDevice 对等设备
type PeerDevice struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Name       string    `json:"name"`
	Address    string    `json:"address"`
	Port       int       `json:"port"`
	LastSeen   time.Time `json:"last_seen"`
	IsOnline   bool      `json:"is_online"`
	Capabilities []string `json:"capabilities"`
}

// P2PMessage P2P消息
type P2PMessage struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"` // discovery, event, command, data
	From      string      `json:"from"`
	To        string      `json:"to"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// P2PClient P2P客户端
type P2PClient struct {
	config  P2PConfig
	deviceID string

	peers    map[string]*PeerDevice
	handlers map[string]func(*P2PMessage)

	udpConn net.UDPConn
	tcpLn   net.Listener

	onPeerOnline  func(*PeerDevice)
	onPeerOffline func(*PeerDevice)

	mu      sync.RWMutex
	running bool
}

// NewP2PClient 创建P2P客户端
func NewP2PClient(deviceID string, config P2PConfig) *P2PClient {
	return &P2PClient{
		config:   config,
		deviceID: deviceID,
		peers:    make(map[string]*PeerDevice),
		handlers: make(map[string]func(*P2PMessage)),
	}
}

// Start 启动P2P
func (pc *P2PClient) Start() error {
	pc.mu.Lock()
	if pc.running {
		pc.mu.Unlock()
		return nil
	}
	pc.running = true
	pc.mu.Unlock()

	// 启动UDP发现
	go pc.discoveryLoop()

	// 启动TCP监听
	go pc.listenLoop()

	// 启动心跳检测
	go pc.heartbeatLoop()

	log.Printf("P2P client started on port %d", pc.config.ListenPort)
	return nil
}

// Stop 停止P2P
func (pc *P2PClient) Stop() {
	pc.mu.Lock()
	pc.running = false
	pc.mu.Unlock()

	if pc.udpConn.LocalAddr() != nil {
		pc.udpConn.Close()
	}
	if pc.tcpLn != nil {
		pc.tcpLn.Close()
	}
}

// GetPeers 获取对等设备
func (pc *P2PClient) GetPeers() []*PeerDevice {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	var peers []*PeerDevice
	for _, p := range pc.peers {
		if p.IsOnline {
			peers = append(peers, p)
		}
	}
	return peers
}

// GetPeer 获取特定设备
func (pc *P2PClient) GetPeer(id string) *PeerDevice {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.peers[id]
}

// Send 发送消息
func (pc *P2PClient) Send(peerID string, msg *P2PMessage) error {
	pc.mu.RLock()
	peer, ok := pc.peers[peerID]
	if !ok || !peer.IsOnline {
		pc.mu.RUnlock()
		return ErrPeerNotOnline
	}
	pc.mu.RUnlock()

	// TCP连接发送
	addr := peer.Address + ":" + itoa(peer.Port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	msg.From = pc.deviceID
	msg.Timestamp = time.Now()

	data, _ := json.Marshal(msg)
	_, err = conn.Write(data)
	return err
}

// Broadcast 广播消息
func (pc *P2PClient) Broadcast(msg *P2PMessage) {
	peers := pc.GetPeers()
	for _, peer := range peers {
		pc.Send(peer.ID, msg)
	}
}

// OnMessage 注册消息处理器
func (pc *P2PClient) OnMessage(msgType string, handler func(*P2PMessage)) {
	pc.mu.Lock()
	pc.handlers[msgType] = handler
	pc.mu.Unlock()
}

// OnPeerOnline 设置上线回调
func (pc *P2PClient) OnPeerOnline(handler func(*PeerDevice)) {
	pc.mu.Lock()
	pc.onPeerOnline = handler
	pc.mu.Unlock()
}

// OnPeerOffline 设置下线回调
func (pc *P2PClient) OnPeerOffline(handler func(*PeerDevice)) {
	pc.mu.Lock()
	pc.onPeerOffline = handler
	pc.mu.Unlock()
}

func (pc *P2PClient) discoveryLoop() {
	// 绑定UDP
	addr := &net.UDPAddr{Port: pc.config.DiscoveryPort}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Printf("UDP listen failed: %v", err)
		return
	}
	pc.udpConn = *conn

	buf := make([]byte, 1024)

	for {
		pc.mu.RLock()
		running := pc.running
		pc.mu.RUnlock()

		if !running {
			return
		}

		n, remoteAddr, err := pc.udpConn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		var msg P2PMessage
		if json.Unmarshal(buf[:n], &msg) != nil {
			continue
		}

		if msg.Type == "discovery" {
			pc.handleDiscovery(&msg, remoteAddr)
		}
	}
}

func (pc *P2PClient) listenLoop() {
	ln, err := net.Listen("tcp", ":"+itoa(pc.config.ListenPort))
	if err != nil {
		log.Printf("TCP listen failed: %v", err)
		return
	}
	pc.tcpLn = ln

	for {
		pc.mu.RLock()
		running := pc.running
		pc.mu.RUnlock()

		if !running {
			return
		}

		conn, err := pc.tcpLn.Accept()
		if err != nil {
			continue
		}

		go pc.handleConnection(conn)
	}
}

func (pc *P2PClient) heartbeatLoop() {
	ticker := time.NewTicker(time.Duration(pc.config.HeartbeatInt) * time.Second)
	defer ticker.Stop()

	for {
		pc.mu.RLock()
		running := pc.running
		pc.mu.RUnlock()

		if !running {
			return
		}

		<-ticker.C

		// 发送发现广播
		pc.sendDiscovery()

		// 检查超时
		pc.checkPeerTimeout()
	}
}

func (pc *P2PClient) handleDiscovery(msg *P2PMessage, addr *net.UDPAddr) {
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		return
	}

	peer := &PeerDevice{
		ID:         getStringVal(data, "id"),
		Type:       getStringVal(data, "type"),
		Name:       getStringVal(data, "name"),
		Address:    addr.IP.String(),
		Port:       getIntVal(data, "port"),
		LastSeen:   time.Now(),
		IsOnline:   true,
	}

	if peer.ID == pc.deviceID {
		return
	}

	pc.mu.Lock()
	isNew := pc.peers[peer.ID] == nil
	pc.peers[peer.ID] = peer

	onOnline := pc.onPeerOnline
	pc.mu.Unlock()

	if isNew && onOnline != nil {
		go onOnline(peer)
	}
}

func (pc *P2PClient) handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	var msg P2PMessage
	if json.Unmarshal(buf[:n], &msg) == nil {
		pc.mu.RLock()
		handler := pc.handlers[msg.Type]
		pc.mu.RUnlock()

		if handler != nil {
			handler(&msg)
		}
	}
}

func (pc *P2PClient) sendDiscovery() {
	self := map[string]interface{}{
		"id":   pc.deviceID,
		"type": "iot_device",
		"port": pc.config.ListenPort,
	}

	msg := &P2PMessage{
		ID:        generateP2PID(),
		Type:      "discovery",
		From:      pc.deviceID,
		Data:      self,
		Timestamp: time.Now(),
	}

	data, _ := json.Marshal(msg)

	broadcastAddr := &net.UDPAddr{
		IP:   net.IPv4(255, 255, 255, 255),
		Port: pc.config.DiscoveryPort,
	}

	pc.udpConn.WriteToUDP(data, broadcastAddr)
}

func (pc *P2PClient) checkPeerTimeout() {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	now := time.Now()
	timeout := time.Duration(pc.config.HeartbeatInt*3) * time.Second

	for id, peer := range pc.peers {
		if now.Sub(peer.LastSeen) > timeout {
			peer.IsOnline = false
			if pc.onPeerOffline != nil {
				go pc.onPeerOffline(peer)
			}
			delete(pc.peers, id)
		}
	}
}

// Errors
var ErrPeerNotOnline = &P2PError{Message: "peer not online"}

type P2PError struct {
	Message string
}

func (e *P2PError) Error() string { return e.Message }

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

func getStringVal(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getIntVal(m map[string]interface{}, key string) int {
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

func generateP2PID() string {
	return "p2p-" + time.Now().Format("150405")
}