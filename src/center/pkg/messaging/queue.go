// Package messaging - NATS消息队列集成
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// NATSConfig NATS配置
type NATSConfig struct {
	URLs           []string      `json:"urls"`            // NATS服务器地址
	Name           string        `json:"name"`            // 客户端名称
	ReconnectWait  time.Duration `json:"reconnect_wait"`  // 重连等待
	MaxReconnect   int           `json:"max_reconnect"`   // 最大重连次数
	PingInterval   time.Duration `json:"ping_interval"`   // 心跳间隔
	Timeout        time.Duration `json:"timeout"`         // 请求超时
	AsyncErrorCB   func(err error) `json:"-"`             // 异步错误回调
	DisconnectedCB func()         `json:"-"`             // 断开连接回调
	ReconnectedCB  func()         `json:"-"`             // 重连回调
}

// SubjectConfig 主题配置
type SubjectConfig struct {
	Subject      string `json:"subject"`       // 主题名称
	QueueGroup   string `json:"queue_group"`   // 队列组
	Durable      string `json:"durable"`       // 持久化名称
	MaxDeliver   int    `json:"max_deliver"`   // 最大投递次数
	AckWait      time.Duration `json:"ack_wait"` // 确认等待时间
}

// NATSMessage NATS消息包装
type NATSMessage struct {
	ID          string                 `json:"id"`
	Subject     string                 `json:"subject"`
	Reply       string                 `json:"reply,omitempty"`
	Data        []byte                 `json:"data"`
	Headers     map[string]string      `json:"headers,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Sequence    uint64                 `json:"sequence,omitempty"`
	Redelivered bool                   `json:"redelivered"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Subscription 订阅信息
type NATSSubscription struct {
	ID          string        `json:"id"`
	Subject     string        `json:"subject"`
	QueueGroup  string        `json:"queue_group"`
	Durable     string        `json:"durable"`
	Active      bool          `json:"active"`
	MessageCount int          `json:"message_count"`
	CreatedAt   time.Time     `json:"created_at"`
	LastActive  time.Time     `json:"last_active"`
	Unsubscribed bool         `json:"unsubscribed"`
}

// PublishResult 发布结果
type PublishResult struct {
	Subject   string        `json:"subject"`
	MessageID string        `json:"message_id"`
	Success   bool          `json:"success"`
	Error     string        `json:"error,omitempty"`
	Latency   time.Duration `json:"latency"`
	Timestamp time.Time     `json:"timestamp"`
}

// NATSManager NATS管理器
type NATSManager struct {
	config        NATSConfig
	conn          *nats.Conn
	js            nats.JetStreamContext // JetStream上下文
	subscriptions map[string]*NATSSubscription
	msgHandlers   map[string]MessageHandler // subject -> handler
	stats         *NATSStats
	running       bool
	mu            sync.RWMutex
}

// NATSStats NATS统计
type NATSStats struct {
	TotalPublished   int64 `json:"total_published"`
	TotalConsumed    int64 `json:"total_consumed"`
	TotalErrors      int64 `json:"total_errors"`
	TotalRetries     int64 `json:"total_retries"`
	AvgPublishLatency time.Duration `json:"avg_publish_latency"`
	AvgConsumeLatency time.Duration `json:"avg_consume_latency"`
}

// NewNATSManager 创建NATS管理器
func NewNATSManager(config NATSConfig) *NATSManager {
	return &NATSManager{
		config:        config,
		subscriptions: make(map[string]*NATSSubscription),
		msgHandlers:   make(map[string]MessageHandler),
		stats:         &NATSStats{},
	}
}

// Connect 连接到NATS服务器
func (nm *NATSManager) Connect(ctx context.Context) error {
	opts := []nats.Option{
		nats.Name(nm.config.Name),
		nats.ReconnectWait(nm.config.ReconnectWait),
		nats.MaxReconnects(nm.config.MaxReconnect),
		nats.PingInterval(nm.config.PingInterval),
		nats.Timeout(nm.config.Timeout),
	}

	// 设置回调
	if nm.config.AsyncErrorCB != nil {
		opts = append(opts, nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			nm.config.AsyncErrorCB(err)
			nm.mu.Lock()
			nm.stats.TotalErrors++
			nm.mu.Unlock()
		}))
	}

	if nm.config.DisconnectedCB != nil {
		opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			nm.config.DisconnectedCB()
		}))
	}

	if nm.config.ReconnectedCB != nil {
		opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
			nm.config.ReconnectedCB()
		}))
	}

	// 连接
	url := nats.DefaultURL
	if len(nm.config.URLs) > 0 {
		url = nm.config.URLs[0]
	}

	nc, err := nats.Connect(url, opts...)
	if err != nil {
		return fmt.Errorf("连接NATS失败: %w", err)
	}

	nm.conn = nc

	// 初始化JetStream
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return fmt.Errorf("初始化JetStream失败: %w", err)
	}
	nm.js = js

	nm.mu.Lock()
	nm.running = true
	nm.mu.Unlock()

	return nil
}

// Disconnect 断开连接
func (nm *NATSManager) Disconnect() {
	nm.mu.Lock()
	nm.running = false
	nm.mu.Unlock()

	// 取消所有订阅
	for _, sub := range nm.subscriptions {
		if sub.Active && !sub.Unsubscribed {
			nm.unsubscribeInternal(sub.ID)
		}
	}

	if nm.conn != nil {
		nm.conn.Close()
	}
}

// Publish 发布消息
func (nm *NATSManager) Publish(ctx context.Context, subject string, data []byte) (*PublishResult, error) {
	if nm.conn == nil {
		return nil, fmt.Errorf("未连接到NATS")
	}

	start := time.Now()
	result := &PublishResult{
		Subject:   subject,
		MessageID: fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Timestamp: start,
	}

	msg := &nats.Msg{
		Subject: subject,
		Data:    data,
		Header:  nats.Header{},
	}
	msg.Header.Set("Message-ID", result.MessageID)

	err := nm.conn.PublishMsg(msg)
	result.Latency = time.Since(start)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		nm.mu.Lock()
		nm.stats.TotalErrors++
		nm.mu.Unlock()
		return result, err
	}

	result.Success = true
	nm.mu.Lock()
	nm.stats.TotalPublished++
	nm.mu.Unlock()

	return result, nil
}

// PublishWithReply 发布消息并等待回复
func (nm *NATSManager) PublishWithReply(ctx context.Context, subject string, data []byte, timeout time.Duration) ([]byte, error) {
	if nm.conn == nil {
		return nil, fmt.Errorf("未连接到NATS")
	}

	if timeout == 0 {
		timeout = nm.config.Timeout
	}

	msg := &nats.Msg{
		Subject: subject,
		Data:    data,
	}

	resp, err := nm.conn.RequestMsg(msg, timeout)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// PublishAsync 异步发布
func (nm *NATSManager) PublishAsync(ctx context.Context, subject string, data []byte, ackHandler func(guid string, err error)) (string, error) {
	if nm.js == nil {
		return "", fmt.Errorf("JetStream未初始化")
	}

	_, err := nm.js.PublishAsync(subject, data, nats.AckWait(5*time.Second))
	if err != nil {
		return "", err
	}

	nm.mu.Lock()
	nm.stats.TotalPublished++
	nm.mu.Unlock()

	return fmt.Sprintf("msg-%d", time.Now().UnixNano()), nil
}

// Subscribe 订阅主题
func (nm *NATSManager) Subscribe(ctx context.Context, subject string, handler MessageHandler) (*NATSSubscription, error) {
	return nm.SubscribeWithQueue(ctx, subject, "", handler)
}

// SubscribeWithQueue 带队列组订阅
func (nm *NATSManager) SubscribeWithQueue(ctx context.Context, subject string, queueGroup string, handler MessageHandler) (*NATSSubscription, error) {
	if nm.conn == nil {
		return nil, fmt.Errorf("未连接到NATS")
	}

	subID := fmt.Sprintf("sub-%d", time.Now().UnixNano())
	natsSub := &NATSSubscription{
		ID:         subID,
		Subject:    subject,
		QueueGroup: queueGroup,
		Active:     true,
		CreatedAt:  time.Now(),
	}

	var sub *nats.Subscription
	var err error

	// 包装处理器
	natsHandler := func(msg *nats.Msg) {

		// 转换消息
		natsMsg := &NATSMessage{
			ID:        msg.Header.Get("Message-ID"),
			Subject:   msg.Subject,
			Reply:     msg.Reply,
			Data:      msg.Data,
			Timestamp: time.Now(),
		}

		// 转换为内部消息格式
		internalMsg := &Message{
			ID:        natsMsg.ID,
			Type:      TypeDirect,
			From:      "nats",
			To:        subject,
			Timestamp: natsMsg.Timestamp,
		}
		json.Unmarshal(msg.Data, internalMsg)

		// 调用处理器
		_, hErr := handler(ctx, internalMsg)

		nm.mu.Lock()
		nm.stats.TotalConsumed++
		natsSub.MessageCount++
		natsSub.LastActive = time.Now()
		nm.mu.Unlock()

		if hErr != nil {
			nm.mu.Lock()
			nm.stats.TotalErrors++
			nm.mu.Unlock()
			// 不确认消息，允许重试
			return
		}

		// 确认消息
		if msg.Reply != "" {
			msg.Respond([]byte("ACK"))
		}
	}

	// 根据是否有队列组选择订阅方式
	if queueGroup != "" {
		sub, err = nm.conn.QueueSubscribe(subject, queueGroup, natsHandler)
	} else {
		sub, err = nm.conn.Subscribe(subject, natsHandler)
	}

	if err != nil {
		return nil, fmt.Errorf("订阅失败: %w", err)
	}

	// 确保订阅有效
	_ = sub

	natsSub.Active = true

	nm.mu.Lock()
	nm.subscriptions[subID] = natsSub
	nm.msgHandlers[subject] = handler
	nm.mu.Unlock()

	return natsSub, nil
}

// SubscribeDurable 持久化订阅
func (nm *NATSManager) SubscribeDurable(ctx context.Context, subject string, durable string, handler MessageHandler) (*NATSSubscription, error) {
	if nm.js == nil {
		return nil, fmt.Errorf("JetStream未初始化")
	}

	subID := fmt.Sprintf("sub-%d", time.Now().UnixNano())
	natsSub := &NATSSubscription{
		ID:        subID,
		Subject:   subject,
		Durable:   durable,
		Active:    true,
		CreatedAt: time.Now(),
	}

	// 包装处理器
	natsHandler := func(msg *nats.Msg) {
		start := time.Now()

		internalMsg := &Message{
			ID:        msg.Header.Get("Message-ID"),
			Type:      TypeDirect,
			From:      "nats",
			To:        subject,
			Timestamp: time.Now(),
		}
		json.Unmarshal(msg.Data, internalMsg)

		_, err := handler(ctx, internalMsg)

		nm.mu.Lock()
		nm.stats.TotalConsumed++
		nm.stats.AvgConsumeLatency = time.Since(start)
		natsSub.MessageCount++
		natsSub.LastActive = time.Now()
		nm.mu.Unlock()

		if err != nil {
			nm.mu.Lock()
			nm.stats.TotalErrors++
			nm.stats.TotalRetries++
			nm.mu.Unlock()
			msg.Nak() // 负面确认，重试
			return
		}

		msg.Ack() // 确认
	}

	sub, err := nm.js.Subscribe(subject, natsHandler,
		nats.Durable(durable),
		nats.ManualAck(),
	)
	if err != nil {
		return nil, fmt.Errorf("持久化订阅失败: %w", err)
	}
	_ = sub // 订阅对象由NATS管理

	nm.mu.Lock()
	nm.subscriptions[subID] = natsSub
	nm.mu.Unlock()

	return natsSub, nil
}

// Unsubscribe 取消订阅
func (nm *NATSManager) Unsubscribe(subID string) error {
	return nm.unsubscribeInternal(subID)
}

func (nm *NATSManager) unsubscribeInternal(subID string) error {
	nm.mu.RLock()
	sub, ok := nm.subscriptions[subID]
	nm.mu.RUnlock()

	if !ok {
		return fmt.Errorf("订阅不存在: %s", subID)
	}

	sub.Active = false
	sub.Unsubscribed = true
	sub.LastActive = time.Now()

	nm.mu.Lock()
	delete(nm.subscriptions, subID)
	nm.mu.Unlock()

	return nil
}

// CreateStream 创建流
func (nm *NATSManager) CreateStream(ctx context.Context, name string, subjects []string) error {
	if nm.js == nil {
		return fmt.Errorf("JetStream未初始化")
	}

	streamConfig := &nats.StreamConfig{
		Name:     name,
		Subjects: subjects,
		Storage:  nats.FileStorage,
		Replicas: 1,
		Retention: nats.LimitsPolicy,
		MaxAge:   7 * 24 * time.Hour, // 保留7天
	}

	_, err := nm.js.AddStream(streamConfig)
	if err != nil {
		return fmt.Errorf("创建流失败: %w", err)
	}

	return nil
}

// DeleteStream 删除流
func (nm *NATSManager) DeleteStream(name string) error {
	if nm.js == nil {
		return fmt.Errorf("JetStream未初始化")
	}

	return nm.js.DeleteStream(name)
}

// ListStreams 列出所有流
func (nm *NATSManager) ListStreams() ([]string, error) {
	if nm.js == nil {
		return nil, fmt.Errorf("JetStream未初始化")
	}

	streams := make([]string, 0)
	info := nm.js.StreamsInfo()
	for streamInfo := range info {
		streams = append(streams, streamInfo.Config.Name)
	}

	return streams, nil
}

// CreateConsumer 创建消费者
func (nm *NATSManager) CreateConsumer(stream string, config *nats.ConsumerConfig) error {
	if nm.js == nil {
		return fmt.Errorf("JetStream未初始化")
	}

	_, err := nm.js.AddConsumer(stream, config)
	return err
}

// GetSubscription 获取订阅信息
func (nm *NATSManager) GetSubscription(subID string) (*NATSSubscription, error) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	sub, ok := nm.subscriptions[subID]
	if !ok {
		return nil, fmt.Errorf("订阅不存在: %s", subID)
	}
	return sub, nil
}

// ListSubscriptions 列出所有订阅
func (nm *NATSManager) ListSubscriptions() []*NATSSubscription {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	subs := make([]*NATSSubscription, 0, len(nm.subscriptions))
	for _, sub := range nm.subscriptions {
		subs = append(subs, sub)
	}
	return subs
}

// GetStatistics 获取统计信息
func (nm *NATSManager) GetStatistics() map[string]interface{} {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	status := "disconnected"
	if nm.conn != nil && nm.conn.IsConnected() {
		status = "connected"
	}

	return map[string]interface{}{
		"status":              status,
		"total_published":     nm.stats.TotalPublished,
		"total_consumed":      nm.stats.TotalConsumed,
		"total_errors":        nm.stats.TotalErrors,
		"total_retries":       nm.stats.TotalRetries,
		"avg_publish_latency": nm.stats.AvgPublishLatency,
		"avg_consume_latency": nm.stats.AvgConsumeLatency,
		"subscriptions":       len(nm.subscriptions),
	}
}

// Request 请求-响应模式
func (nm *NATSManager) Request(ctx context.Context, subject string, data []byte, timeout time.Duration) ([]byte, error) {
	if nm.conn == nil {
		return nil, fmt.Errorf("未连接到NATS")
	}

	if timeout == 0 {
		timeout = nm.config.Timeout
	}

	msg := &nats.Msg{
		Subject: subject,
		Data:    data,
	}

	resp, err := nm.conn.RequestMsg(msg, timeout)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// Respond 响应消息
func (nm *NATSManager) Respond(msg *NATSMessage, response []byte) error {
	if nm.conn == nil {
		return fmt.Errorf("未连接到NATS")
	}

	if msg.Reply == "" {
		return fmt.Errorf("无回复地址")
	}

	return nm.conn.Publish(msg.Reply, response)
}

// IsConnected 检查是否已连接
func (nm *NATSManager) IsConnected() bool {
	if nm.conn == nil {
		return false
	}
	return nm.conn.IsConnected()
}

// GetConnectionStatus 获取连接状态
func (nm *NATSManager) GetConnectionStatus() map[string]interface{} {
	if nm.conn == nil {
		return map[string]interface{}{
			"connected": false,
			"status":    "not_initialized",
		}
	}

	return map[string]interface{}{
		"connected":         nm.conn.IsConnected(),
		"status":            nm.conn.Status().String(),
		"server_id":         nm.conn.ConnectedServerId(),
		"server_url":        nm.conn.ConnectedUrl(),
		"subscriptions":     len(nm.subscriptions),
		"messages_sent":     nm.conn.Stats().OutMsgs,
		"messages_received": nm.conn.Stats().InMsgs,
		"bytes_sent":        nm.conn.Stats().OutBytes,
		"bytes_received":    nm.conn.Stats().InBytes,
	}
}