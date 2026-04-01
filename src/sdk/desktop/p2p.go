// Package desktop provides P2P communication
package desktop

import (
	"encoding/json"
	"log"
	"net"
	"sync"
	"time"
)

// P2PConfig holds P2P configuration
type P2PConfig struct {
	ListenPort    int
	DiscoveryPort int
	EnableDiscovery bool
}

// DefaultP2PConfig returns default config
func DefaultP2PConfig() *P2PConfig {
	return &P2PConfig{
		ListenPort:    7891,
		DiscoveryPort: 7890,
		EnableDiscovery: true,
	}
}

// PeerInfo represents a peer
type PeerInfo struct {
	ID        string
	Name      string
	Address   string
	Port      int
	Skills    []string
	LastSeen  time.Time
	IsOnline  bool
}

// P2PClient manages P2P communication
type P2PClient struct {
	config  *P2PConfig
	agentID string
	peers   map[string]*PeerInfo

	tcpListener net.Listener
	udpConn     *net.UDPConn

	onPeerOnline  func(*PeerInfo)
	onPeerOffline func(*PeerInfo)
	onMessage     func(*P2PMessage)

	mu      sync.RWMutex
	running bool
	done    chan struct{}
}

// P2PMessage represents a P2P message
type P2PMessage struct {
	ID        string
	Type      string // discovery, task, data, custom
	From      string
	To        string
	Data      json.RawMessage
	Timestamp time.Time
}

// NewP2PClient creates a new P2P client
func NewP2PClient(agentID string, config *P2PConfig) *P2PClient {
	if config == nil {
		config = DefaultP2PConfig()
	}
	return &P2PClient{
		config:  config,
		agentID: agentID,
		peers:   make(map[string]*PeerInfo),
		done:    make(chan struct{}),
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
	listener, err := net.Listen("tcp", ":"+itoa(p.config.ListenPort))
	if err != nil {
		return err
	}
	p.tcpListener = listener

	// Start UDP for discovery
	udpAddr, _ := net.ResolveUDPAddr("udp", ":"+itoa(p.config.DiscoveryPort))
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		udpAddr, _ = net.ResolveUDPAddr("udp", ":0")
		udpConn, err = net.ListenUDP("udp", udpAddr)
		if err != nil {
			return err
		}
	}
	p.udpConn = udpConn

	// Start background loops
	go p.tcpAcceptLoop()
	go p.udpReceiveLoop()
	go p.peerTimeoutLoop()

	if p.config.EnableDiscovery {
		go p.discoveryLoop()
	}

	log.Printf("P2P client started on port %d", p.config.ListenPort)
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

// GetPeers returns all peers
func (p *P2PClient) GetPeers() []*PeerInfo {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var peers []*PeerInfo
	for _, peer := range p.peers {
		if peer.IsOnline {
			peers = append(peers, peer)
		}
	}
	return peers
}

// Send sends a message to a peer
func (p *P2PClient) Send(peerID string, msg *P2PMessage) error {
	p.mu.RLock()
	peer, ok := p.peers[peerID]
	if !ok || !peer.IsOnline {
		p.mu.RUnlock()
		return ErrPeerNotFound
	}
	p.mu.RUnlock()

	addr := peer.Address + ":" + itoa(peer.Port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	msg.Timestamp = time.Now()
	data, _ := json.Marshal(msg)
	conn.Write(data)

	return nil
}

// Broadcast broadcasts to all peers
func (p *P2PClient) Broadcast(msg *P2PMessage) {
	peers := p.GetPeers()
	for _, peer := range peers {
		if err := p.Send(peer.ID, msg); err != nil {
			log.Printf("Failed to send to %s: %v", peer.ID, err)
		}
	}
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

// OnMessage sets message callback
func (p *P2PClient) OnMessage(callback func(*P2PMessage)) {
	p.mu.Lock()
	p.onMessage = callback
	p.mu.Unlock()
}

// LocalPort returns local port
func (p *P2PClient) LocalPort() int {
	return p.config.ListenPort
}

func (p *P2PClient) tcpAcceptLoop() {
	for {
		select {
		case <-p.done:
			return
		default:
			conn, err := p.tcpListener.Accept()
			if err != nil {
				continue
			}
			go p.handleConnection(conn)
		}
	}
}

func (p *P2PClient) handleConnection(conn net.Conn) {
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	var msg P2PMessage
	if json.Unmarshal(buf[:n], &msg) == nil {
		p.mu.RLock()
		callback := p.onMessage
		p.mu.RUnlock()

		if callback != nil {
			callback(&msg)
		}
	}
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
				continue
			}

			var msg P2PMessage
			if json.Unmarshal(buf[:n], &msg) != nil {
				continue
			}

			if msg.Type == "discovery" {
				var peer PeerInfo
				if json.Unmarshal(msg.Data, &peer) == nil {
					peer.Address = remoteAddr.IP.String()
					peer.LastSeen = time.Now()
					peer.IsOnline = true

					p.mu.Lock()
					isNew := p.peers[peer.ID] == nil
					p.peers[peer.ID] = &peer

					callback := p.onPeerOnline
					p.mu.Unlock()

					if isNew && callback != nil {
						callback(&peer)
					}
				}
			}
		}
	}
}

func (p *P2PClient) discoveryLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.done:
			return
		case <-ticker.C:
			p.sendDiscovery()
		}
	}
}

func (p *P2PClient) sendDiscovery() {
	selfInfo := PeerInfo{
		ID:   p.agentID,
		Port: p.config.ListenPort,
	}

	msg := P2PMessage{
		ID:   generateP2PID(),
		Type: "discovery",
		From: p.agentID,
	}
	msg.Data, _ = json.Marshal(selfInfo)

	data, _ := json.Marshal(msg)

	broadcastAddr := &net.UDPAddr{
		IP:   net.IPv4(255, 255, 255, 255),
		Port: p.config.DiscoveryPort,
	}

	if p.udpConn != nil {
		p.udpConn.WriteToUDP(data, broadcastAddr)
	}
}

func (p *P2PClient) peerTimeoutLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.done:
			return
		case <-ticker.C:
			p.mu.Lock()
			now := time.Now()
			for id, peer := range p.peers {
				if now.Sub(peer.LastSeen) > 30*time.Second {
					peer.IsOnline = false
					callback := p.onPeerOffline
					if callback != nil {
						go callback(peer)
					}
					delete(p.peers, id)
				}
			}
			p.mu.Unlock()
		}
	}
}

func generateP2PID() string {
	return "p2p-" + time.Now().Format("20060102150405")
}

func itoa(i int) string {
	return string(rune(i))
}

// Errors
var ErrPeerNotFound = &P2PError{Message: "peer not found"}

// P2PError represents P2P error
type P2PError struct {
	Message string
}

func (e *P2PError) Error() string {
	return e.Message
}