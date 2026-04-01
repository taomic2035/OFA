// Package messaging - Agent间P2P通信模块
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// MessageType 消息类型
type MessageType string

const (
	TypeDirect    MessageType = "direct"    // 直接消息
	TypeBroadcast MessageType = "broadcast" // 广播消息
	TypeMulticast MessageType = "multicast" // 组播消息
	TypeRequest   MessageType = "request"   // 请求消息
	TypeResponse  MessageType = "response"  // 响应消息
	TypeAck       MessageType = "ack"       // 确认消息
	TypePing      MessageType = "ping"      // 心跳检测
	TypePong      MessageType = "pong"      // 心跳响应
)

// MessagePriority 消息优先级
type MessagePriority int

const (
	PriorityLow      MessagePriority = 0
	PriorityNormal   MessagePriority = 1
	PriorityHigh     MessagePriority = 2
	PriorityCritical MessagePriority = 3
)

// MessageState 消息状态
type MessageState string

const (
	StatePending   MessageState = "pending"   // 待发送
	StateSent      MessageState = "sent"      // 已发送
	StateDelivered MessageState = "delivered" // 已送达
	StateAcked     MessageState = "acked"     // 已确认
	StateFailed    MessageState = "failed"    // 发送失败
	StateExpired   MessageState = "expired"   // 已过期
)

// Message 消息结构
type Message struct {
	ID         string                 `json:"id"`
	Type       MessageType            `json:"type"`
	Priority   MessagePriority        `json:"priority"`
	From       string                 `json:"from"`        // 发送者AgentID
	To         string                 `json:"to"`          // 接收者AgentID/GroupID/Broadcast
	Group      string                 `json:"group,omitempty"` // 组播组ID
	Action     string                 `json:"action"`      // 动作类型
	Payload    map[string]interface{} `json:"payload"`     // 消息内容
	Timestamp  time.Time              `json:"timestamp"`
	TTL        time.Duration          `json:"ttl"`         // 生存时间
	RetryCount int                    `json:"retry_count"` // 重试次数
	MaxRetry   int                    `json:"max_retry"`   // 最大重试
	RequireAck bool                   `json:"require_ack"` // 需要确认
	CorrelationID string              `json:"correlation_id,omitempty"` // 关联ID(用于请求-响应)
}

// MessageAck 确认消息
type MessageAck struct {
	MessageID   string        `json:"message_id"`
	From        string        `json:"from"`
	To          string        `json:"to"`
	Status      MessageState  `json:"status"`
	Error       string        `json:"error,omitempty"`
	Timestamp   time.Time     `json:"timestamp"`
	Latency     time.Duration `json:"latency"` // 往返延迟
}

// PeerConnection 对等连接
type PeerConnection struct {
	AgentID      string        `json:"agent_id"`
	Address      string        `json:"address"`      // IP:Port
	Conn         net.Conn      `json:"-"`            // 底层连接
	State        ConnState     `json:"state"`
	LastActive   time.Time     `json:"last_active"`
	LastPing     time.Time     `json:"last_ping"`
	RTT          time.Duration `json:"rtt"`          // 往返时间
	MessageCount int           `json:"message_count"`
	BytesSent    int64         `json:"bytes_sent"`
	BytesRecv    int64         `json:"bytes_recv"`
}

// ConnState 连接状态
type ConnState string

const (
	ConnStateConnecting ConnState = "connecting"
	ConnStateConnected  ConnState = "connected"
	ConnStateReady      ConnState = "ready"
	ConnStateClosing    ConnState = "closing"
	ConnStateClosed     ConnState = "closed"
	ConnStateError      ConnState = "error"
)

// PeerInfo 对等节点信息
type PeerInfo struct {
	AgentID     string            `json:"agent_id"`
	Address     string            `json:"address"`
	Capabilities []string         `json:"capabilities"`
	Metadata    map[string]string `json:"metadata"`
	LastSeen    time.Time         `json:"last_seen"`
	Online      bool              `json:"online"`
}

// P2PConfig P2P配置
type P2PConfig struct {
	LocalAddress     string        `json:"local_address"`     // 本地监听地址
	Port             int           `json:"port"`              // 监听端口
	MaxConnections   int           `json:"max_connections"`   // 最大连接数
	ConnectTimeout   time.Duration `json:"connect_timeout"`   // 连接超时
	WriteTimeout     time.Duration `json:"write_timeout"`     // 写超时
	ReadTimeout      time.Duration `json:"read_timeout"`      // 读超时
	PingInterval     time.Duration `json:"ping_interval"`     // 心跳间隔
	PongTimeout      time.Duration `json:"pong_timeout"`      // 心跳响应超时
	MaxMessageSize   int           `json:"max_message_size"`  // 最大消息大小
	BufferSize       int           `json:"buffer_size"`       // 缓冲区大小
	EnableNAT        bool          `json:"enable_nat"`        // 启用NAT穿透
	RelayServer      string        `json:"relay_server"`      // 中继服务器
}

// P2PManager P2P管理器
type P2PManager struct {
	config        P2PConfig
	localAgentID  string
	listener      net.Listener
	peers         map[string]*PeerConnection // AgentID -> Connection
	pendingAcks   map[string]*Message        // 等待确认的消息
	messageQueue  chan *Message              // 消息队列
	ackQueue      chan *MessageAck           // 确认队列
	handlers      map[string]MessageHandler  // 消息处理器
	peerRegistry  map[string]*PeerInfo       // 对等节点注册表
	groups        map[string][]string        // 组播组: GroupID -> []AgentID
	running       bool
	mu            sync.RWMutex
	wg            sync.WaitGroup
	notifyFunc    func(msg *Message)
	cancel        context.CancelFunc // 取消函数
}

// MessageHandler 消息处理函数
type MessageHandler func(ctx context.Context, msg *Message) (*Message, error)

// NewP2PManager 创建P2P管理器
func NewP2PManager(config P2PConfig, localAgentID string) *P2PManager {
	return &P2PManager{
		config:       config,
		localAgentID: localAgentID,
		peers:        make(map[string]*PeerConnection),
		pendingAcks:  make(map[string]*Message),
		messageQueue: make(chan *Message, config.BufferSize),
		ackQueue:     make(chan *MessageAck, config.BufferSize),
		handlers:     make(map[string]MessageHandler),
		peerRegistry: make(map[string]*PeerInfo),
		groups:       make(map[string][]string),
	}
}

// Start 启动P2P管理器
func (pm *P2PManager) Start(ctx context.Context) error {
	pm.mu.Lock()
	pm.running = true
	pm.mu.Unlock()

	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(ctx)
	pm.cancel = cancel

	// 启动监听
	if pm.config.LocalAddress != "" {
		addr := fmt.Sprintf("%s:%d", pm.config.LocalAddress, pm.config.Port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			cancel()
			return fmt.Errorf("监听失败: %w", err)
		}
		pm.listener = listener

		// 接受连接
		pm.wg.Add(1)
		go pm.acceptConnections(ctx)
	}

	// 启动消息处理
	pm.wg.Add(2)
	go pm.processMessages(ctx)
	go pm.processAcks(ctx)

	// 启动心跳
	pm.wg.Add(1)
	go pm.heartbeatLoop(ctx)

	// 启动超时检查
	pm.wg.Add(1)
	go pm.timeoutChecker(ctx)

	return nil
}

// Stop 停止P2P管理器
func (pm *P2PManager) Stop() error {
	pm.mu.Lock()
	pm.running = false
	pm.mu.Unlock()

	// 取消上下文以停止所有goroutine
	if pm.cancel != nil {
		pm.cancel()
	}

	// 关闭监听
	if pm.listener != nil {
		pm.listener.Close()
	}

	// 关闭所有连接
	for _, peer := range pm.peers {
		if peer.Conn != nil {
			peer.Conn.Close()
		}
	}

	// 等待所有goroutine完成
	pm.wg.Wait()

	// 关闭通道
	close(pm.messageQueue)
	close(pm.ackQueue)

	return nil
}

// acceptConnections 接受连接
func (pm *P2PManager) acceptConnections(ctx context.Context) {
	defer pm.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := pm.listener.Accept()
			if err != nil {
				if pm.running {
					// 记录错误
				}
				continue
			}

			// 处理新连接
			go pm.handleNewConnection(ctx, conn)
		}
	}
}

// handleNewConnection 处理新连接
func (pm *P2PManager) handleNewConnection(ctx context.Context, conn net.Conn) {
	// 等待对方发送AgentID
	// 简化实现：设置读超时
	conn.SetReadDeadline(time.Now().Add(pm.config.ConnectTimeout))

	// 接收握手消息
	buf := make([]byte, pm.config.MaxMessageSize)
	n, err := conn.Read(buf)
	if err != nil {
		conn.Close()
		return
	}

	var handshake struct {
		AgentID string `json:"agent_id"`
		Type    string `json:"type"`
	}
	if err := json.Unmarshal(buf[:n], &handshake); err != nil {
		conn.Close()
		return
	}

	// 创建对等连接
	peer := &PeerConnection{
		AgentID:    handshake.AgentID,
		Address:    conn.RemoteAddr().String(),
		Conn:       conn,
		State:      ConnStateConnected,
		LastActive: time.Now(),
	}

	pm.mu.Lock()
	pm.peers[handshake.AgentID] = peer
	pm.mu.Unlock()

	// 开始读取消息
	go pm.readLoop(ctx, peer)
}

// Connect 连接到对等节点
func (pm *P2PManager) Connect(ctx context.Context, peerInfo *PeerInfo) error {
	pm.mu.RLock()
	_, exists := pm.peers[peerInfo.AgentID]
	pm.mu.RUnlock()

	if exists {
		return nil // 已连接
	}

	// 检查连接数限制
	pm.mu.RLock()
	if len(pm.peers) >= pm.config.MaxConnections {
		pm.mu.RUnlock()
		return fmt.Errorf("达到最大连接数限制")
	}
	pm.mu.RUnlock()

	// 建立连接
	conn, err := net.DialTimeout("tcp", peerInfo.Address, pm.config.ConnectTimeout)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}

	// 发送握手
	handshake := map[string]string{
		"agent_id": pm.localAgentID,
		"type":     "handshake",
	}
	data, _ := json.Marshal(handshake)
	conn.Write(data)

	peer := &PeerConnection{
		AgentID:    peerInfo.AgentID,
		Address:    peerInfo.Address,
		Conn:       conn,
		State:      ConnStateConnected,
		LastActive: time.Now(),
	}

	pm.mu.Lock()
	pm.peers[peerInfo.AgentID] = peer
	pm.peerRegistry[peerInfo.AgentID] = peerInfo
	pm.mu.Unlock()

	// 开始读取消息
	go pm.readLoop(ctx, peer)

	return nil
}

// Disconnect 断开与对等节点的连接
func (pm *P2PManager) Disconnect(agentID string) error {
	pm.mu.Lock()
	peer, ok := pm.peers[agentID]
	if ok {
		if peer.Conn != nil {
			peer.Conn.Close()
		}
		delete(pm.peers, agentID)
	}
	pm.mu.Unlock()

	if !ok {
		return fmt.Errorf("未找到连接: %s", agentID)
	}
	return nil
}

// readLoop 读取消息循环
func (pm *P2PManager) readLoop(ctx context.Context, peer *PeerConnection) {
	buf := make([]byte, pm.config.MaxMessageSize)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if peer.Conn == nil {
				return
			}

			peer.Conn.SetReadDeadline(time.Now().Add(pm.config.ReadTimeout))
			n, err := peer.Conn.Read(buf)
			if err != nil {
				pm.handleDisconnect(peer.AgentID)
				return
			}

			peer.BytesRecv += int64(n)
			peer.LastActive = time.Now()

			// 解析消息
			var msg Message
			if err := json.Unmarshal(buf[:n], &msg); err != nil {
				continue
			}

			peer.MessageCount++

			// 处理消息
			if msg.Type == TypeAck {
				var ack MessageAck
				if err := json.Unmarshal(buf[:n], &ack); err == nil {
					pm.ackQueue <- &ack
				}
			} else if msg.Type == TypePing {
				// 响应Pong
				pm.sendPong(peer, msg.ID)
			} else if msg.Type == TypePong {
				// 更新RTT
				pm.mu.Lock()
				peer.RTT = time.Since(peer.LastPing)
				pm.mu.Unlock()
			} else {
				// 普通消息
				pm.messageQueue <- &msg

				// 发送ACK
				if msg.RequireAck {
					pm.sendAck(peer, msg.ID, StateDelivered)
				}
			}
		}
	}
}

// handleDisconnect 处理断开连接
func (pm *P2PManager) handleDisconnect(agentID string) {
	pm.mu.Lock()
	if peer, ok := pm.peers[agentID]; ok {
		if peer.Conn != nil {
			peer.Conn.Close()
		}
		peer.State = ConnStateClosed
		delete(pm.peers, agentID)
	}
	if info, ok := pm.peerRegistry[agentID]; ok {
		info.Online = false
	}
	pm.mu.Unlock()
}

// processMessages 处理消息队列
func (pm *P2PManager) processMessages(ctx context.Context) {
	defer pm.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-pm.messageQueue:
			if !ok {
				return
			}

			// 查找处理器
			pm.mu.RLock()
			handler, ok := pm.handlers[msg.Action]
			pm.mu.RUnlock()

			if ok {
				resp, err := handler(ctx, msg)
				if err == nil && resp != nil && msg.CorrelationID != "" {
					// 发送响应
					pm.Send(msg.From, resp)
				}
			}

			// 通知上层
			if pm.notifyFunc != nil {
				pm.notifyFunc(msg)
			}
		}
	}
}

// processAcks 处理确认队列
func (pm *P2PManager) processAcks(ctx context.Context) {
	defer pm.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case ack, ok := <-pm.ackQueue:
			if !ok {
				return
			}

			pm.mu.Lock()
			if _, exists := pm.pendingAcks[ack.MessageID]; exists {
				delete(pm.pendingAcks, ack.MessageID)
				// 可以触发回调
			}
			pm.mu.Unlock()
		}
	}
}

// Send 发送消息
func (pm *P2PManager) Send(to string, msg *Message) error {
	msg.From = pm.localAgentID
	msg.To = to
	msg.Timestamp = time.Now()

	if msg.ID == "" {
		msg.ID = fmt.Sprintf("msg-%d", time.Now().UnixNano())
	}

	pm.mu.RLock()
	peer, ok := pm.peers[to]
	pm.mu.RUnlock()

	if !ok {
		return fmt.Errorf("未连接到节点: %s", to)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	peer.Conn.SetWriteDeadline(time.Now().Add(pm.config.WriteTimeout))
	n, err := peer.Conn.Write(data)
	if err != nil {
		pm.handleDisconnect(to)
		return err
	}

	peer.BytesSent += int64(n)
	peer.MessageCount++
	peer.LastActive = time.Now()

	// 等待确认
	if msg.RequireAck {
		pm.mu.Lock()
		pm.pendingAcks[msg.ID] = msg
		pm.mu.Unlock()
	}

	return nil
}

// SendAndWait 发送消息并等待响应
func (pm *P2PManager) SendAndWait(ctx context.Context, to string, msg *Message, timeout time.Duration) (*Message, error) {
	correlationID := fmt.Sprintf("corr-%d", time.Now().UnixNano())
	msg.CorrelationID = correlationID
	msg.RequireAck = true

	// 创建响应通道
	respChan := make(chan *Message, 1)

	// 注册临时处理器
	pm.RegisterHandler(correlationID, func(ctx context.Context, m *Message) (*Message, error) {
		respChan <- m
		return nil, nil
	})

	if err := pm.Send(to, msg); err != nil {
		return nil, err
	}

	select {
	case resp := <-respChan:
		return resp, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("等待响应超时")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Broadcast 广播消息
func (pm *P2PManager) Broadcast(msg *Message) error {
	msg.Type = TypeBroadcast
	msg.From = pm.localAgentID
	msg.To = "broadcast"
	msg.Timestamp = time.Now()

	pm.mu.RLock()
	peers := make([]*PeerConnection, 0, len(pm.peers))
	for _, peer := range pm.peers {
		peers = append(peers, peer)
	}
	pm.mu.RUnlock()

	var lastErr error
	for _, peer := range peers {
		if err := pm.Send(peer.AgentID, msg); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// Multicast 组播消息
func (pm *P2PManager) Multicast(groupID string, msg *Message) error {
	msg.Type = TypeMulticast
	msg.From = pm.localAgentID
	msg.Group = groupID
	msg.Timestamp = time.Now()

	pm.mu.RLock()
	members, ok := pm.groups[groupID]
	pm.mu.RUnlock()

	if !ok {
		return fmt.Errorf("组不存在: %s", groupID)
	}

	var lastErr error
	for _, agentID := range members {
		if err := pm.Send(agentID, msg); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// sendAck 发送确认
func (pm *P2PManager) sendAck(peer *PeerConnection, messageID string, status MessageState) {
	ack := &MessageAck{
		MessageID: messageID,
		From:      pm.localAgentID,
		To:        peer.AgentID,
		Status:    status,
		Timestamp: time.Now(),
	}

	data, _ := json.Marshal(ack)
	peer.Conn.Write(data)
}

// sendPong 发送Pong响应
func (pm *P2PManager) sendPong(peer *PeerConnection, pingID string) {
	pong := &Message{
		ID:        pingID,
		Type:      TypePong,
		From:      pm.localAgentID,
		To:        peer.AgentID,
		Timestamp: time.Now(),
	}
	data, _ := json.Marshal(pong)
	peer.Conn.Write(data)
}

// heartbeatLoop 心跳循环
func (pm *P2PManager) heartbeatLoop(ctx context.Context) {
	defer pm.wg.Done()

	ticker := time.NewTicker(pm.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pm.mu.RLock()
			peers := make([]*PeerConnection, 0, len(pm.peers))
			for _, peer := range pm.peers {
				peers = append(peers, peer)
			}
			pm.mu.RUnlock()

			for _, peer := range peers {
				ping := &Message{
					ID:        fmt.Sprintf("ping-%d", time.Now().UnixNano()),
					Type:      TypePing,
					From:      pm.localAgentID,
					To:        peer.AgentID,
					Timestamp: time.Now(),
				}

				pm.mu.Lock()
				peer.LastPing = time.Now()
				pm.mu.Unlock()

				data, _ := json.Marshal(ping)
				peer.Conn.Write(data)
			}
		}
	}
}

// timeoutChecker 超时检查器
func (pm *P2PManager) timeoutChecker(ctx context.Context) {
	defer pm.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()

			pm.mu.Lock()
			// 检查消息超时
			for id, msg := range pm.pendingAcks {
				if msg.TTL > 0 && now.Sub(msg.Timestamp) > msg.TTL {
					delete(pm.pendingAcks, id)
					// 标记为超时
				}
			}

			// 检查连接超时
			for _, peer := range pm.peers {
				if now.Sub(peer.LastActive) > pm.config.PongTimeout*2 {
					if peer.Conn != nil {
						peer.Conn.Close()
					}
					peer.State = ConnStateError
				}
			}
			pm.mu.Unlock()
		}
	}
}

// RegisterHandler 注册消息处理器
func (pm *P2PManager) RegisterHandler(action string, handler MessageHandler) {
	pm.mu.Lock()
	pm.handlers[action] = handler
	pm.mu.Unlock()
}

// SetNotifyFunc 设置通知函数
func (pm *P2PManager) SetNotifyFunc(fn func(msg *Message)) {
	pm.mu.Lock()
	pm.notifyFunc = fn
	pm.mu.Unlock()
}

// CreateGroup 创建组播组
func (pm *P2PManager) CreateGroup(groupID string, members []string) {
	pm.mu.Lock()
	pm.groups[groupID] = members
	pm.mu.Unlock()
}

// JoinGroup 加入组播组
func (pm *P2PManager) JoinGroup(groupID string, agentID string) {
	pm.mu.Lock()
	if pm.groups[groupID] == nil {
		pm.groups[groupID] = make([]string, 0)
	}
	pm.groups[groupID] = append(pm.groups[groupID], agentID)
	pm.mu.Unlock()
}

// LeaveGroup 离开组播组
func (pm *P2PManager) LeaveGroup(groupID string, agentID string) {
	pm.mu.Lock()
	if members, ok := pm.groups[groupID]; ok {
		newMembers := make([]string, 0)
		for _, m := range members {
			if m != agentID {
				newMembers = append(newMembers, m)
			}
		}
		pm.groups[groupID] = newMembers
	}
	pm.mu.Unlock()
}

// GetPeers 获取所有对等节点
func (pm *P2PManager) GetPeers() []*PeerInfo {
	pm.mu.RLock()
	peers := make([]*PeerInfo, 0, len(pm.peerRegistry))
	for _, info := range pm.peerRegistry {
		peers = append(peers, info)
	}
	pm.mu.RUnlock()
	return peers
}

// GetConnectedPeers 获取已连接的对等节点
func (pm *P2PManager) GetConnectedPeers() []string {
	pm.mu.RLock()
	peers := make([]string, 0, len(pm.peers))
	for id := range pm.peers {
		peers = append(peers, id)
	}
	pm.mu.RUnlock()
	return peers
}

// GetPeer 获取对等节点信息
func (pm *P2PManager) GetPeer(agentID string) (*PeerConnection, error) {
	pm.mu.RLock()
	peer, ok := pm.peers[agentID]
	pm.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("未找到对等节点: %s", agentID)
	}
	return peer, nil
}

// IsConnected 检查是否已连接
func (pm *P2PManager) IsConnected(agentID string) bool {
	pm.mu.RLock()
	_, ok := pm.peers[agentID]
	pm.mu.RUnlock()
	return ok
}

// GetStatistics 获取统计信息
func (pm *P2PManager) GetStatistics() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	totalSent := int64(0)
	totalRecv := int64(0)
	for _, peer := range pm.peers {
		totalSent += peer.BytesSent
		totalRecv += peer.BytesRecv
	}

	return map[string]interface{}{
		"connected_peers":  len(pm.peers),
		"registered_peers": len(pm.peerRegistry),
		"pending_acks":     len(pm.pendingAcks),
		"groups":           len(pm.groups),
		"total_bytes_sent": totalSent,
		"total_bytes_recv": totalRecv,
		"running":          pm.running,
	}
}