package ofa

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// PeerInfo represents a peer node
type PeerInfo struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Address   string            `json:"address"`
	Port      int               `json:"port"`
	Skills    []string          `json:"skills"`
	Metadata  map[string]string `json:"metadata"`
	LastSeen  int64             `json:"lastSeen"`
	IsOnline  bool              `json:"isOnline"`
}

// P2PMessageType defines P2P message type
type P2PMessageType int

const (
	P2PMsgDiscovery P2PMessageType = iota
	P2PMsgDiscoveryAck
	P2PMsgTaskRequest
	P2PMsgTaskResponse
	P2PMsgDataSync
	P2PMsgHeartbeat
	P2PMsgCustom
)

// P2PMessage represents a P2P message
type P2PMessage struct {
	ID        string          `json:"id"`
	Type      P2PMessageType  `json:"type"`
	From      string          `json:"from"`
	To        string          `json:"to"`
	Data      json.RawMessage `json:"data"`
	Timestamp int64           `json:"timestamp"`
}

// P2PMessageHandler handles P2P messages
type P2PMessageHandler func(*P2PMessage)

// P2PClient manages P2P communication
type P2PClient struct {
	agentID       string
	listenPort    int
	discoveryPort int

	peers    map[string]*PeerInfo
	handlers map[P2PMessageType][]P2PMessageHandler
	onPeerOnline  func(*PeerInfo)
	onPeerOffline func(*PeerInfo)

	mu       sync.Mutex
	running  bool
	done     chan struct{}

	tcpListener net.Listener
	udpConn     *net.UDPConn
}

// NewP2PClient creates a new P2P client
func NewP2PClient(agentID string, listenPort, discoveryPort int) *P2PClient {
	return &P2PClient{
		agentID:       agentID,
		listenPort:    listenPort,
		discoveryPort: discoveryPort,
		peers:         make(map[string]*PeerInfo),
		handlers:      make(map[P2PMessageType][]P2PMessageHandler),
		done:          make(chan struct{}),
	}
}

// Start starts the P2P client
func (p *P2PClient) Start() error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return nil
	}
	p.running = true
	p.mu.Unlock()

	// Start TCP listener
	if p.listenPort == 0 {
		p.listenPort = 7891
	}

	addr := fmt.Sprintf(":%d", p.listenPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start TCP listener: %w", err)
	}
	p.tcpListener = listener

	// Start UDP for discovery
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", p.discoveryPort))
	if err != nil {
		udpAddr, err = net.ResolveUDPAddr("udp", ":0")
		if err != nil {
			return fmt.Errorf("failed to resolve UDP address: %w", err)
		}
	}

	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to start UDP listener: %w", err)
	}
	p.udpConn = udpConn

	// Start background loops
	go p.tcpAcceptLoop()
	go p.udpReceiveLoop()
	go p.peerTimeoutLoop()

	log.Printf("P2P client started on port %d", p.listenPort)
	return nil
}

// Stop stops the P2P client
func (p *P2PClient) Stop() {
	p.mu.Lock()
	p.running = false
	p.mu.Unlock()

	close(p.done)

	if p.tcpListener != nil {
		p.tcpListener.Close()
	}
	if p.udpConn != nil {
		p.udpConn.Close()
	}
}

// StartDiscovery starts peer discovery
func (p *P2PClient) StartDiscovery() {
	go p.discoveryLoop()
}

// GetPeers returns discovered peers
func (p *P2PClient) GetPeers() []*PeerInfo {
	p.mu.Lock()
	defer p.mu.Unlock()

	var peers []*PeerInfo
	for _, peer := range p.peers {
		if peer.IsOnline {
			peers = append(peers, peer)
		}
	}
	return peers
}

// GetPeer returns a specific peer
func (p *P2PClient) GetPeer(peerID string) (*PeerInfo, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	peer, ok := p.peers[peerID]
	if !ok {
		return nil, fmt.Errorf("peer not found: %s", peerID)
	}
	return peer, nil
}

// Send sends a message to a peer
func (p *P2PClient) Send(peerID string, msg *P2PMessage) error {
	p.mu.Lock()
	peer, ok := p.peers[peerID]
	if !ok || !peer.IsOnline {
		p.mu.Unlock()
		return fmt.Errorf("peer not online: %s", peerID)
	}
	p.mu.Unlock()

	// Connect to peer
	addr := fmt.Sprintf("%s:%d", peer.Address, peer.Port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to peer: %w", err)
	}
	defer conn.Close()

	// Send message
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	_, err = conn.Write(data)
	return err
}

// Broadcast broadcasts to all peers
func (p *P2PClient) Broadcast(msg *P2PMessage) error {
	peers := p.GetPeers()
	for _, peer := range peers {
		if err := p.Send(peer.ID, msg); err != nil {
			log.Printf("Failed to send to peer %s: %v", peer.ID, err)
		}
	}
	return nil
}

// SendTaskRequest sends a task request to a peer
func (p *P2PClient) SendTaskRequest(ctx context.Context, peerID, skillID, operation string, input json.RawMessage, timeout time.Duration) (json.RawMessage, error) {
	msg := &P2PMessage{
		ID:        generateID(),
		Type:      P2PMsgTaskRequest,
		From:      p.agentID,
		To:        peerID,
		Timestamp: time.Now().UnixMilli(),
	}

	taskData := map[string]interface{}{
		"skillId":   skillID,
		"operation": operation,
		"input":     input,
	}
	msg.Data, _ = json.Marshal(taskData)

	if err := p.Send(peerID, msg); err != nil {
		return nil, err
	}

	// In production: wait for response
	// For now, return placeholder
	return json.RawMessage(`{"status":"sent"}`), nil
}

// OnMessage registers a message handler
func (p *P2PClient) OnMessage(msgType P2PMessageType, handler P2PMessageHandler) {
	p.mu.Lock()
	p.handlers[msgType] = append(p.handlers[msgType], handler)
	p.mu.Unlock()
}

// OnPeerOnline sets peer online callback
func (p *P2PClient) OnPeerOnline(callback func(*PeerInfo)) {
	p.mu.Lock()
	p.onPeerOnline = callback
	p.mu.Unlock()
}

// OnPeerOffline sets peer offline callback
func (p *P2PClient) OnPeerOffline(callback func(*PeerInfo)) {
	p.mu.Lock()
	p.onPeerOffline = callback
	p.mu.Unlock()
}

// LocalPort returns the local port
func (p *P2PClient) LocalPort() int {
	return p.listenPort
}

func (p *P2PClient) tcpAcceptLoop() {
	for {
		select {
		case <-p.done:
			return
		default:
			conn, err := p.tcpListener.Accept()
			if err != nil {
				if p.running {
					log.Printf("TCP accept error: %v", err)
				}
				continue
			}

			go p.handleTCPConnection(conn)
		}
	}
}

func (p *P2PClient) handleTCPConnection(conn net.Conn) {
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	var msg P2PMessage
	if err := json.Unmarshal(buf[:n], &msg); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	p.processMessage(&msg)
}

func (p *P2PClient) udpReceiveLoop() {
	buf := make([]byte, 4096)

	for {
		select {
		case <-p.done:
			return
		default:
			n, remoteAddr, err := p.udpConn.ReadFromUDP(buf)
			if err != nil {
				if p.running {
					log.Printf("UDP read error: %v", err)
				}
				continue
			}

			var msg P2PMessage
			if err := json.Unmarshal(buf[:n], &msg); err != nil {
				continue
			}

			// Handle discovery
			if msg.Type == P2PMsgDiscovery {
				var peer PeerInfo
				if err := json.Unmarshal(msg.Data, &peer); err != nil {
					continue
				}

				peer.Address = remoteAddr.IP.String()
				peer.LastSeen = time.Now().UnixMilli()
				peer.IsOnline = true

				p.mu.Lock()
				isNew := p.peers[peer.ID] == nil
				p.peers[peer.ID] = &peer

				if isNew && p.onPeerOnline != nil {
					go p.onPeerOnline(&peer)
				}
				p.mu.Unlock()

				// Send ack
				ack := &P2PMessage{
					ID:        generateID(),
					Type:      P2PMsgDiscoveryAck,
					From:      p.agentID,
					Timestamp: time.Now().UnixMilli(),
				}
				p.sendUDPAck(remoteAddr, ack)
			}
		}
	}
}

func (p *P2PClient) sendUDPAck(addr *net.UDPAddr, msg *P2PMessage) {
	data, _ := json.Marshal(msg)
	p.udpConn.WriteToUDP(data, addr)
}

func (p *P2PClient) discoveryLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.done:
			return
		case <-ticker.C:
			p.sendDiscoveryBroadcast()
		}
	}
}

func (p *P2PClient) sendDiscoveryBroadcast() {
	selfInfo := &PeerInfo{
		ID:      p.agentID,
		Address: "0.0.0.0",
		Port:    p.listenPort,
	}

	msg := &P2PMessage{
		ID:        generateID(),
		Type:      P2PMsgDiscovery,
		From:      p.agentID,
		Timestamp: time.Now().UnixMilli(),
	}
	msg.Data, _ = json.Marshal(selfInfo)

	data, _ := json.Marshal(msg)

	// Broadcast to discovery port
	broadcastAddr := &net.UDPAddr{
		IP:   net.IPv4(255, 255, 255, 255),
		Port: p.discoveryPort,
	}

	if p.udpConn != nil {
		p.udpConn.WriteToUDP(data, broadcastAddr)
	}
}

func (p *P2PClient) peerTimeoutLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.done:
			return
		case <-ticker.C:
			p.mu.Lock()
			now := time.Now().UnixMilli()
			for id, peer := range p.peers {
				if now - peer.LastSeen > 30000 { // 30s timeout
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
}

func (p *P2PClient) processMessage(msg *P2PMessage) {
	p.mu.Lock()
	handlers := p.handlers[msg.Type]
	p.mu.Unlock()

	for _, handler := range handlers {
		go handler(msg)
	}
}