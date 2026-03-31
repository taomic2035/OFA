// Package iot - MQTT协议实现
package iot

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// MQTTConfig MQTT配置
type MQTTConfig struct {
	Broker       string        `json:"broker"`        // broker地址
	ClientID     string        `json:"client_id"`     // 客户端ID
	Username     string        `json:"username"`      // 用户名
	Password     string        `json:"password"`      // 密码
	QoS          byte          `json:"qos"`           // QoS级别
	KeepAlive    time.Duration `json:"keep_alive"`    // 心跳间隔
	CleanSession bool          `json:"clean_session"` // 清除会话
	AutoReconnect bool         `json:"auto_reconnect"` // 自动重连
	MaxReconnect  int          `json:"max_reconnect"`  // 最大重连次数
	ConnectTimeout time.Duration `json:"connect_timeout"` // 连接超时
	WriteTimeout   time.Duration `json:"write_timeout"`   // 写超时

	// TLS配置
	UseTLS    bool   `json:"use_tls"`
	CertFile  string `json:"cert_file"`
	KeyFile   string `json:"key_file"`
	CAFile    string `json:"ca_file"`
	SkipVerify bool  `json:"skip_verify"`
}

// MQTTClientImpl MQTT客户端实现
type MQTTClientImpl struct {
	config      MQTTConfig
	client      MQTT.Client
	opts        *MQTT.ClientOptions
	handlers    map[string]TopicHandler
	connected   bool
	reconnects  int
	mu          sync.RWMutex
	onConnect   func()
	onDisconnect func(error)
}

// NewMQTTClient 创建MQTT客户端
func NewMQTTClient(config MQTTConfig) *MQTTClientImpl {
	// 默认配置
	if config.KeepAlive == 0 {
		config.KeepAlive = 60 * time.Second
	}
	if config.ConnectTimeout == 0 {
		config.ConnectTimeout = 10 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 5 * time.Second
	}
	if config.MaxReconnect == 0 {
		config.MaxReconnect = 10
	}

	return &MQTTClientImpl{
		config:   config,
		handlers: make(map[string]TopicHandler),
	}
}

// Connect 连接MQTT Broker
func (mc *MQTTClientImpl) Connect(ctx context.Context) error {
	mc.opts = MQTT.NewClientOptions()
	mc.opts.AddBroker(mc.config.Broker)
	mc.opts.SetClientID(mc.config.ClientID)
	mc.opts.SetUsername(mc.config.Username)
	mc.opts.SetPassword(mc.config.Password)
	mc.opts.SetKeepAlive(mc.config.KeepAlive)
	mc.opts.SetCleanSession(mc.config.CleanSession)
	mc.opts.SetAutoReconnect(mc.config.AutoReconnect)
	mc.opts.SetConnectTimeout(mc.config.ConnectTimeout)
	mc.opts.SetWriteTimeout(mc.config.WriteTimeout)

	// TLS配置
	if mc.config.UseTLS {
		tlsConfig, err := mc.createTLSConfig()
		if err != nil {
			return fmt.Errorf("TLS配置失败: %w", err)
		}
		mc.opts.SetTLSConfig(tlsConfig)
	}

	// 回调设置
	mc.opts.OnConnect = mc.onConnectHandler
	mc.opts.OnConnectionLost = mc.onConnectionLostHandler
	mc.opts.OnReconnecting = mc.onReconnectingHandler

	mc.client = MQTT.NewClient(mc.opts)

	token := mc.client.Connect()
	if token.Error() != nil {
		return fmt.Errorf("MQTT连接失败: %w", token.Error())
	}

	mc.mu.Lock()
	mc.connected = true
	mc.mu.Unlock()

	// 重新订阅
	mc.resubscribe()

	return nil
}

// createTLSConfig 创建TLS配置
func (mc *MQTTClientImpl) createTLSConfig() (*tls.Config, error) {
	config := &tls.Config{
		InsecureSkipVerify: mc.config.SkipVerify,
	}

	// 加载客户端证书
	if mc.config.CertFile != "" && mc.config.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(mc.config.CertFile, mc.config.KeyFile)
		if err != nil {
			return nil, err
		}
		config.Certificates = []tls.Certificate{cert}
	}

	// 加载CA证书
	if mc.config.CAFile != "" {
		caCert, err := ioutil.ReadFile(mc.config.CAFile)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		config.RootCAs = caCertPool
	}

	return config, nil
}

// Disconnect 断开连接
func (mc *MQTTClientImpl) Disconnect() {
	mc.mu.Lock()
	mc.connected = false
	mc.mu.Unlock()

	if mc.client != nil && mc.client.IsConnected() {
		mc.client.Disconnect(250) // 等待250ms
	}
}

// Publish 发布消息
func (mc *MQTTClientImpl) Publish(topic string, qos byte, payload []byte) error {
	mc.mu.RLock()
	connected := mc.connected
	mc.mu.RUnlock()

	if !connected {
		return fmt.Errorf("MQTT未连接")
	}

	token := mc.client.Publish(topic, qos, false, payload)
	token.WaitTimeout(mc.config.WriteTimeout)
	return token.Error()
}

// Subscribe 订阅主题
func (mc *MQTTClientImpl) Subscribe(topic string, qos byte, handler TopicHandler) error {
	mc.mu.Lock()
	mc.handlers[topic] = handler
	connected := mc.connected
	mc.mu.Unlock()

	if !connected {
		return nil
	}

	token := mc.client.Subscribe(topic, qos, mc.messageHandler)
	token.WaitTimeout(mc.config.WriteTimeout)
	return token.Error()
}

// Unsubscribe 取消订阅
func (mc *MQTTClientImpl) Unsubscribe(topic string) error {
	mc.mu.Lock()
	delete(mc.handlers, topic)
	connected := mc.connected
	mc.mu.Unlock()

	if !connected {
		return nil
	}

	token := mc.client.Unsubscribe(topic)
	token.WaitTimeout(mc.config.WriteTimeout)
	return token.Error()
}

// IsConnected 检查连接状态
func (mc *MQTTClientImpl) IsConnected() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.connected && mc.client != nil && mc.client.IsConnected()
}

// messageHandler 消息处理器
func (mc *MQTTClientImpl) messageHandler(client MQTT.Client, msg MQTT.Message) {
	mc.mu.RLock()
	handler, ok := mc.handlers[msg.Topic()]
	mc.mu.RUnlock()

	if ok && handler != nil {
		handler(msg.Topic(), msg.Payload())
	}
}

// onConnectHandler 连接回调
func (mc *MQTTClientImpl) onConnectHandler(client MQTT.Client) {
	mc.mu.Lock()
	mc.connected = true
	mc.reconnects = 0
	mc.mu.Unlock()

	// 重新订阅
	mc.resubscribe()

	if mc.onConnect != nil {
		mc.onConnect()
	}
}

// onConnectionLostHandler 断开连接回调
func (mc *MQTTClientImpl) onConnectionLostHandler(client MQTT.Client, err error) {
	mc.mu.Lock()
	mc.connected = false
	mc.mu.Unlock()

	if mc.onDisconnect != nil {
		mc.onDisconnect(err)
	}
}

// onReconnectingHandler 重连回调
func (mc *MQTTClientImpl) onReconnectingHandler(client MQTT.Client, options *MQTT.ClientOptions) {
	mc.mu.Lock()
	mc.reconnects++
	mc.mu.Unlock()
}

// resubscribe 重新订阅
func (mc *MQTTClientImpl) resubscribe() {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	for topic, handler := range mc.handlers {
		mc.client.Subscribe(topic, mc.config.QoS, mc.messageHandler)
		_ = handler // 避免未使用警告
	}
}

// SetOnConnect 设置连接回调
func (mc *MQTTClientImpl) SetOnConnect(handler func()) {
	mc.onConnect = handler
}

// SetOnDisconnect 设置断开回调
func (mc *MQTTClientImpl) SetOnDisconnect(handler func(error)) {
	mc.onDisconnect = handler
}

// GetReconnectCount 获取重连次数
func (mc *MQTTClientImpl) GetReconnectCount() int {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.reconnects
}

// === MQTT消息类型 ===

// MQTTMessage MQTT消息
type MQTTMessage struct {
	Topic     string                 `json:"topic"`
	QoS       byte                   `json:"qos"`
	Retained  bool                   `json:"retained"`
	Payload   []byte                 `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
	Headers   map[string]interface{} `json:"headers,omitempty"`
}

// MQTTTopic MQTT主题管理
type MQTTTopic struct {
	DeviceID string
	Prefix   string // 前缀，如 "devices" 或 "$shadow"
}

// 主题生成方法
func (t *MQTTTopic) Command() string {
	return fmt.Sprintf("%s/%s/commands", t.Prefix, t.DeviceID)
}

func (t *MQTTTopic) CommandResponse() string {
	return fmt.Sprintf("%s/%s/commands/response", t.Prefix, t.DeviceID)
}

func (t *MQTTTopic) Events() string {
	return fmt.Sprintf("%s/%s/events", t.Prefix, t.DeviceID)
}

func (t *MQTTTopic) Telemetry() string {
	return fmt.Sprintf("%s/%s/telemetry", t.Prefix, t.DeviceID)
}

func (t *MQTTTopic) ShadowDesired() string {
	return fmt.Sprintf("$shadow/%s/%s/desired", t.Prefix, t.DeviceID)
}

func (t *MQTTTopic) ShadowReported() string {
	return fmt.Sprintf("$shadow/%s/%s/reported", t.Prefix, t.DeviceID)
}

func (t *MQTTTopic) ShadowDelta() string {
	return fmt.Sprintf("$shadow/%s/%s/delta", t.Prefix, t.DeviceID)
}

// === QoS级别说明 ===
// QoS 0: At most once - 最多一次，可能丢失
// QoS 1: At least once - 至少一次，可能重复
// QoS 2: Exactly once - 恰好一次，保证不丢失不重复

// QoSLevel QoS级别
type QoSLevel byte

const (
	QoSAtMostOnce  QoSLevel = 0
	QoSAtLeastOnce QoSLevel = 1
	QoSExactlyOnce QoSLevel = 2
)

// String 返回QoS描述
func (q QoSLevel) String() string {
	switch q {
	case QoSAtMostOnce:
		return "At most once (0)"
	case QoSAtLeastOnce:
		return "At least once (1)"
	case QoSExactlyOnce:
		return "Exactly once (2)"
	default:
		return "Unknown QoS"
	}
}