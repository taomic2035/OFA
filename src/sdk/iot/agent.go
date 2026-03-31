// Package iot - IoT Agent SDK for Smart Home Devices
// 物联网Agent SDK，专为智能家居设备设计
package iot

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// IoTConfig IoT Agent配置
type IoTConfig struct {
	DeviceID       string        `json:"device_id"`        // 设备ID
	DeviceType     string        `json:"device_type"`      // 设备类型
	DeviceName     string        `json:"device_name"`      // 设备名称
	CenterURL      string        `json:"center_url"`       // Center服务地址
	MQTTBroker     string        `json:"mqtt_broker"`      // MQTT Broker地址
	MQTTClientID   string        `json:"mqtt_client_id"`   // MQTT客户端ID
	MQTTUser       string        `json:"mqtt_user"`        // MQTT用户名
	MQTTPassword   string        `json:"mqtt_password"`    // MQTT密码
	MQTTQoS        byte          `json:"mqtt_qos"`         // QoS级别(0,1,2)
	KeepAlive      time.Duration `json:"keep_alive"`       // 心跳间隔
	CleanSession   bool          `json:"clean_session"`    // 清除会话
	AutoReconnect  bool          `json:"auto_reconnect"`   // 自动重连
	ShadowEnabled  bool          `json:"shadow_enabled"`   // 启用设备影子
	ShadowSyncInt  time.Duration `json:"shadow_sync_int"`  // 影子同步间隔
}

// IoTAgent IoT Agent实例
type IoTAgent struct {
	config        IoTConfig
	mqttClient    MQTTClient
	shadow        *DeviceShadow
	properties    *DeviceProperties
	events        chan *DeviceEvent
	commands      chan *DeviceCommand
	subscriptions map[string]TopicHandler
	running       bool
	mu            sync.RWMutex
}

// DeviceShadow 设备影子
type DeviceShadow struct {
	DeviceID     string                 `json:"device_id"`
	Version      int64                  `json:"version"`
	Desired      map[string]interface{} `json:"desired"`       // 期望状态(云端)
	Reported     map[string]interface{} `json:"reported"`      // 报告状态(设备)
	Delta        map[string]interface{} `json:"delta"`         // 差异状态
	Metadata     *ShadowMetadata        `json:"metadata"`
	LastSyncTime time.Time              `json:"last_sync_time"`
	SyncStatus   string                 `json:"sync_status"`   // synced, pending, error
	mu           sync.RWMutex
}

// ShadowMetadata 影子元数据
type ShadowMetadata struct {
	DesiredVersion  int64                  `json:"desired_version"`
	ReportedVersion int64                  `json:"reported_version"`
	Timestamps      map[string]time.Time   `json:"timestamps"`
}

// DeviceProperties 设备属性
type DeviceProperties struct {
	Online       bool                   `json:"online"`
	Status       string                 `json:"status"`       // online, offline, maintenance
	FirmwareVer  string                 `json:"firmware_version"`
	IPAddress    string                 `json:"ip_address"`
	MACAddress   string                 `json:"mac_address"`
	RSSI         int                    `json:"rssi"`         // 信号强度
	Battery      int                    `json:"battery"`      // 电池电量(-1表示无电池)
	Capabilities []string               `json:"capabilities"`
	Custom       map[string]interface{} `json:"custom"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// DeviceEvent 设备事件
type DeviceEvent struct {
	ID         string                 `json:"id"`
	DeviceID   string                 `json:"device_id"`
	Type       string                 `json:"type"`       // property_change, alert, status
	Timestamp  time.Time              `json:"timestamp"`
	Properties map[string]interface{} `json:"properties"`
	Payload    []byte                 `json:"payload,omitempty"`
}

// DeviceCommand 设备命令
type DeviceCommand struct {
	ID         string                 `json:"id"`
	DeviceID   string                 `json:"device_id"`
	Type       string                 `json:"type"`       // set_property, action, ota
	Params     map[string]interface{} `json:"params"`
	Timestamp  time.Time              `json:"timestamp"`
	Timeout    time.Duration          `json:"timeout"`
	CorrelationID string              `json:"correlation_id,omitempty"`
}

// CommandResult 命令执行结果
type CommandResult struct {
	CommandID   string      `json:"command_id"`
	Success     bool        `json:"success"`
	Result      interface{} `json:"result,omitempty"`
	Error       string      `json:"error,omitempty"`
	CompletedAt time.Time   `json:"completed_at"`
}

// TopicHandler 主题处理器
type TopicHandler func(topic string, payload []byte)

// MQTTClient MQTT客户端接口
type MQTTClient interface {
	Connect(ctx context.Context) error
	Disconnect()
	Publish(topic string, qos byte, payload []byte) error
	Subscribe(topic string, qos byte, handler TopicHandler) error
	Unsubscribe(topic string) error
	IsConnected() bool
}

// NewIoTAgent 创建IoT Agent
func NewIoTAgent(config IoTConfig) *IoTAgent {
	// 默认配置
	if config.KeepAlive == 0 {
		config.KeepAlive = 60 * time.Second
	}
	if config.MQTTQoS == 0 {
		config.MQTTQoS = 1
	}
	if config.ShadowSyncInt == 0 {
		config.ShadowSyncInt = 30 * time.Second
	}

	return &IoTAgent{
		config:        config,
		shadow:        NewDeviceShadow(config.DeviceID),
		properties:    &DeviceProperties{Status: "offline"},
		events:        make(chan *DeviceEvent, 100),
		commands:      make(chan *DeviceCommand, 50),
		subscriptions: make(map[string]TopicHandler),
	}
}

// NewDeviceShadow 创建设备影子
func NewDeviceShadow(deviceID string) *DeviceShadow {
	return &DeviceShadow{
		DeviceID: deviceID,
		Desired:  make(map[string]interface{}),
		Reported: make(map[string]interface{}),
		Delta:    make(map[string]interface{}),
		Metadata: &ShadowMetadata{
			Timestamps: make(map[string]time.Time),
		},
		SyncStatus: "pending",
	}
}

// Start 启动Agent
func (ia *IoTAgent) Start(ctx context.Context) error {
	ia.mu.Lock()
	ia.running = true
	ia.properties.Online = true
	ia.properties.Status = "online"
	ia.mu.Unlock()

	// 连接MQTT
	if ia.mqttClient != nil {
		if err := ia.mqttClient.Connect(ctx); err != nil {
			return fmt.Errorf("MQTT连接失败: %w", err)
		}

		// 订阅主题
		ia.subscribeTopics(ctx)
	}

	// 启动影子同步
	if ia.config.ShadowEnabled {
		go ia.shadowSyncLoop(ctx)
	}

	// 启动事件处理
	go ia.eventLoop(ctx)
	go ia.commandLoop(ctx)

	// 发布上线事件
	ia.PublishEvent("status", map[string]interface{}{
		"status": "online",
		"time":   time.Now().Unix(),
	})

	return nil
}

// Stop 停止Agent
func (ia *IoTAgent) Stop() {
	ia.mu.Lock()
	ia.running = false
	ia.properties.Online = false
	ia.properties.Status = "offline"
	ia.mu.Unlock()

	// 发布离线事件
	ia.PublishEvent("status", map[string]interface{}{
		"status": "offline",
		"time":   time.Now().Unix(),
	})

	if ia.mqttClient != nil {
		ia.mqttClient.Disconnect()
	}
}

// subscribeTopics 订阅主题
func (ia *IoTAgent) subscribeTopics(ctx context.Context) {
	// 命令主题
	cmdTopic := fmt.Sprintf("devices/%s/commands", ia.config.DeviceID)
	ia.mqttClient.Subscribe(cmdTopic, ia.config.MQTTQoS, ia.handleCommand)

	// 影子主题
	if ia.config.ShadowEnabled {
		shadowTopic := fmt.Sprintf("$shadow/devices/%s/delta", ia.config.DeviceID)
		ia.mqttClient.Subscribe(shadowTopic, ia.config.MQTTQoS, ia.handleShadowDelta)
	}
}

// === 属性管理 ===

// SetProperty 设置属性
func (ia *IoTAgent) SetProperty(key string, value interface{}) error {
	ia.mu.Lock()
	defer ia.mu.Unlock()

	// 更新报告状态
	ia.shadow.mu.Lock()
	ia.shadow.Reported[key] = value
	ia.shadow.Metadata.ReportedVersion++
	ia.shadow.Metadata.Timestamps[key] = time.Now()
	ia.shadow.mu.Unlock()

	// 发布属性变更事件
	ia.PublishEvent("property_change", map[string]interface{}{
		"key":   key,
		"value": value,
	})

	// 同步到影子
	if ia.config.ShadowEnabled {
		ia.syncReportedShadow()
	}

	return nil
}

// GetProperty 获取属性
func (ia *IoTAgent) GetProperty(key string) (interface{}, error) {
	ia.shadow.mu.RLock()
	defer ia.shadow.mu.RUnlock()

	val, ok := ia.shadow.Reported[key]
	if !ok {
		return nil, fmt.Errorf("属性不存在: %s", key)
	}
	return val, nil
}

// GetDesiredProperty 获取期望属性
func (ia *IoTAgent) GetDesiredProperty(key string) (interface{}, error) {
	ia.shadow.mu.RLock()
	defer ia.shadow.mu.RUnlock()

	val, ok := ia.shadow.Desired[key]
	if !ok {
		return nil, fmt.Errorf("期望属性不存在: %s", key)
	}
	return val, nil
}

// GetAllProperties 获取所有属性
func (ia *IoTAgent) GetAllProperties() map[string]interface{} {
	ia.shadow.mu.RLock()
	defer ia.shadow.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range ia.shadow.Reported {
		result[k] = v
	}
	return result
}

// === 设备影子 ===

// UpdateDesired 更新期望状态(云端调用)
func (ia *IoTAgent) UpdateDesired(desired map[string]interface{}) error {
	ia.shadow.mu.Lock()
	defer ia.shadow.mu.Unlock()

	// 合并期望状态
	for k, v := range desired {
		ia.shadow.Desired[k] = v
	}
	ia.shadow.Metadata.DesiredVersion++
	ia.shadow.Version++

	// 计算Delta
	ia.calculateDelta()

	return nil
}

// calculateDelta 计算差异
func (ia *IoTAgent) calculateDelta() {
	ia.shadow.Delta = make(map[string]interface{})

	for k, desiredVal := range ia.shadow.Desired {
		reportedVal, ok := ia.shadow.Reported[k]
		if !ok || !isEqual(reportedVal, desiredVal) {
			ia.shadow.Delta[k] = desiredVal
		}
	}
}

// syncReportedShadow 同步报告状态到影子
func (ia *IoTAgent) syncReportedShadow() {
	if ia.mqttClient == nil || !ia.mqttClient.IsConnected() {
		return
	}

	topic := fmt.Sprintf("$shadow/devices/%s/reported", ia.config.DeviceID)

	ia.shadow.mu.RLock()
	reported := map[string]interface{}{
		"reported": ia.shadow.Reported,
		"version":  ia.shadow.Metadata.ReportedVersion,
	}
	ia.shadow.mu.RUnlock()

	payload, _ := json.Marshal(reported)
	ia.mqttClient.Publish(topic, ia.config.MQTTQoS, payload)
}

// shadowSyncLoop 影子同步循环
func (ia *IoTAgent) shadowSyncLoop(ctx context.Context) {
	ticker := time.NewTicker(ia.config.ShadowSyncInt)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ia.syncReportedShadow()
		}
	}
}

// handleShadowDelta 处理影子差异
func (ia *IoTAgent) handleShadowDelta(topic string, payload []byte) {
	var delta struct {
		State   map[string]interface{} `json:"state"`
		Version int64                  `json:"version"`
	}

	if err := json.Unmarshal(payload, &delta); err != nil {
		return
	}

	// 应用差异到设备
	for key, value := range delta.State {
		// 触发属性设置命令
		cmd := &DeviceCommand{
			ID:        fmt.Sprintf("cmd-%d", time.Now().UnixNano()),
			DeviceID:  ia.config.DeviceID,
			Type:      "set_property",
			Params:    map[string]interface{}{"key": key, "value": value},
			Timestamp: time.Now(),
		}

		select {
		case ia.commands <- cmd:
		default:
		}
	}
}

// === 事件发布 ===

// PublishEvent 发布事件
func (ia *IoTAgent) PublishEvent(eventType string, properties map[string]interface{}) error {
	if ia.mqttClient == nil || !ia.mqttClient.IsConnected() {
		return nil
	}

	event := &DeviceEvent{
		ID:         fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		DeviceID:   ia.config.DeviceID,
		Type:       eventType,
		Timestamp:  time.Now(),
		Properties: properties,
	}

	topic := fmt.Sprintf("devices/%s/events", ia.config.DeviceID)
	payload, _ := json.Marshal(event)

	return ia.mqttClient.Publish(topic, ia.config.MQTTQoS, payload)
}

// PublishTelemetry 发布遥测数据
func (ia *IoTAgent) PublishTelemetry(data map[string]interface{}) error {
	if ia.mqttClient == nil || !ia.mqttClient.IsConnected() {
		return nil
	}

	telemetry := map[string]interface{}{
		"device_id": ia.config.DeviceID,
		"timestamp": time.Now().Unix(),
		"data":      data,
	}

	topic := fmt.Sprintf("devices/%s/telemetry", ia.config.DeviceID)
	payload, _ := json.Marshal(telemetry)

	return ia.mqttClient.Publish(topic, ia.config.MQTTQoS, payload)
}

// === 命令处理 ===

// handleCommand 处理命令
func (ia *IoTAgent) handleCommand(topic string, payload []byte) {
	var cmd DeviceCommand
	if err := json.Unmarshal(payload, &cmd); err != nil {
		return
	}

	cmd.Timestamp = time.Now()

	select {
	case ia.commands <- &cmd:
	default:
	}
}

// commandLoop 命令处理循环
func (ia *IoTAgent) commandLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case cmd := <-ia.commands:
			result := ia.executeCommand(ctx, cmd)
			ia.publishCommandResult(result)
		}
	}
}

// executeCommand 执行命令
func (ia *IoTAgent) executeCommand(ctx context.Context, cmd *DeviceCommand) *CommandResult {
	result := &CommandResult{
		CommandID:   cmd.ID,
		CompletedAt: time.Now(),
	}

	switch cmd.Type {
	case "set_property":
		if key, ok := cmd.Params["key"].(string); ok {
			if value, exists := cmd.Params["value"]; exists {
				if err := ia.SetProperty(key, value); err != nil {
					result.Error = err.Error()
					return result
				}
				result.Success = true
				result.Result = map[string]interface{}{key: value}
			}
		}

	case "get_property":
		if key, ok := cmd.Params["key"].(string); ok {
			val, err := ia.GetProperty(key)
			if err != nil {
				result.Error = err.Error()
				return result
			}
			result.Success = true
			result.Result = map[string]interface{}{key: val}
		}

	case "action":
		// 执行自定义动作
		action, ok := cmd.Params["action"].(string)
		if !ok {
			result.Error = "缺少action参数"
			return result
		}
		result.Success = true
		result.Result = map[string]interface{}{"action": action, "executed": true}

	case "ota":
		// OTA升级
		url, ok := cmd.Params["url"].(string)
		if !ok {
			result.Error = "缺少url参数"
			return result
		}
		result.Success = true
		result.Result = map[string]interface{}{"ota_url": url, "status": "downloading"}

	default:
		result.Error = fmt.Sprintf("未知命令类型: %s", cmd.Type)
	}

	return result
}

// publishCommandResult 发布命令结果
func (ia *IoTAgent) publishCommandResult(result *CommandResult) {
	if ia.mqttClient == nil || !ia.mqttClient.IsConnected() {
		return
	}

	topic := fmt.Sprintf("devices/%s/commands/response", ia.config.DeviceID)
	payload, _ := json.Marshal(result)
	ia.mqttClient.Publish(topic, ia.config.MQTTQoS, payload)
}

// eventLoop 事件循环
func (ia *IoTAgent) eventLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event := <-ia.events:
			ia.processEvent(event)
		}
	}
}

// processEvent 处理事件
func (ia *IoTAgent) processEvent(event *DeviceEvent) {
	// 事件处理逻辑(可扩展)
}

// === 工具方法 ===

// SetMQTTClient 设置MQTT客户端
func (ia *IoTAgent) SetMQTTClient(client MQTTClient) {
	ia.mqttClient = client
}

// GetShadow 获取设备影子
func (ia *IoTAgent) GetShadow() *DeviceShadow {
	return ia.shadow
}

// GetProperties 获取设备属性
func (ia *IoTAgent) GetProperties() *DeviceProperties {
	ia.mu.RLock()
	defer ia.mu.RUnlock()
	return ia.properties
}

// UpdateProperties 更新设备属性
func (ia *IoTAgent) UpdateProperties(props map[string]interface{}) {
	ia.mu.Lock()
	defer ia.mu.Unlock()

	for k, v := range props {
		switch k {
		case "firmware_version":
			ia.properties.FirmwareVer = v.(string)
		case "ip_address":
			ia.properties.IPAddress = v.(string)
		case "rssi":
			ia.properties.RSSI = v.(int)
		case "battery":
			ia.properties.Battery = v.(int)
		}
	}
	ia.properties.UpdatedAt = time.Now()
}

// IsRunning 检查运行状态
func (ia *IoTAgent) IsRunning() bool {
	ia.mu.RLock()
	defer ia.mu.RUnlock()
	return ia.running
}

// IsConnected 检查连接状态
func (ia *IoTAgent) IsConnected() bool {
	return ia.mqttClient != nil && ia.mqttClient.IsConnected()
}

// isEqual 比较两个值是否相等
func isEqual(a, b interface{}) bool {
	// 简化比较
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}