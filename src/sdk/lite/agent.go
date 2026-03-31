// Package lite - Lite Agent SDK for Wearables (Watch/Band)
// 轻量级Agent SDK，专为手表、手环等低功耗设备设计
package lite

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// LiteConfig Lite Agent配置
type LiteConfig struct {
	ServerURL      string        `json:"server_url"`      // Center服务器地址
	AgentID        string        `json:"agent_id"`        // Agent ID
	DeviceType     string        `json:"device_type"`     // 设备类型: watch, band, glass
	HeartbeatInt   time.Duration `json:"heartbeat_int"`   // 心跳间隔(默认60秒，低功耗)
	ReconnectInt   time.Duration `json:"reconnect_int"`   // 重连间隔
	MaxReconnect   int           `json:"max_reconnect"`   // 最大重连次数
	BatterySaver   bool          `json:"battery_saver"`   // 省电模式
	CompressData   bool          `json:"compress_data"`   // 数据压缩
	BufferSize     int           `json:"buffer_size"`     // 缓冲区大小
	MaxMessageSize int           `json:"max_message_size"` // 最大消息大小
}

// LiteAgent Lite Agent实例
type LiteAgent struct {
	config       LiteConfig
	conn         LiteConnection
	skills       map[string]LiteSkill
	status       *AgentStatus
	messageQueue chan *LiteMessage
	pendingTasks map[string]*LiteTask
	battery      *BatteryInfo
	sensors      *SensorManager
	running      bool
	mu           sync.RWMutex
}

// AgentStatus Agent状态
type AgentStatus struct {
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Status       string    `json:"status"` // online, offline, sleeping, busy
	BatteryLevel int       `json:"battery_level"` // 0-100
	Charging     bool      `json:"charging"`
	CPUUsage     float64   `json:"cpu_usage"`
	MemoryUsage  float64   `json:"memory_usage"`
	Temperature  float64   `json:"temperature"`
	LastActive   time.Time `json:"last_active"`
	Uptime       int64     `json:"uptime"` // seconds
}

// BatteryInfo 电池信息
type BatteryInfo struct {
	Level       int       `json:"level"`        // 0-100
	Charging    bool      `json:"charging"`
	Voltage     float64   `json:"voltage"`
	Temperature float64   `json:"temperature"`
	Health      string    `json:"health"` // good, fair, poor
	LastUpdate  time.Time `json:"last_update"`
}

// SensorManager 传感器管理器
type SensorManager struct {
	accelerometer *Sensor
	gyroscope     *Sensor
	heartRate     *Sensor
	stepCounter   *Sensor
	gps           *Sensor
	temperature   *Sensor
	light         *Sensor
	mu            sync.RWMutex
}

// Sensor 传感器
type Sensor struct {
	Name      string      `json:"name"`
	Type      string      `json:"type"`
	Active    bool        `json:"active"`
	Interval  time.Duration `json:"interval"`
	LastValue interface{} `json:"last_value"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// LiteMessage 轻量级消息
type LiteMessage struct {
	ID        uint16         `json:"id"`        // 2字节ID
	Type      MessageType    `json:"type"`      // 消息类型
	Timestamp uint32         `json:"timestamp"` // 4字节时间戳
	Payload   []byte         `json:"payload"`   // 载荷
	Priority  byte           `json:"priority"`  // 优先级(0-3)
	Compressed bool          `json:"compressed"`
}

// MessageType 消息类型(简化版)
type MessageType byte

const (
	TypeHeartbeat   MessageType = 0x01
	TypeTask        MessageType = 0x02
	TypeResult      MessageType = 0x03
	TypeSensor      MessageType = 0x04
	TypeNotification MessageType = 0x05
	TypeError       MessageType = 0x0F
)

// LiteTask 轻量级任务
type LiteTask struct {
	ID        string      `json:"id"`
	Skill     string      `json:"skill"`
	Action    string      `json:"action"`
	Params    interface{} `json:"params,omitempty"`
	Timeout   time.Duration `json:"timeout"`
	Priority  byte        `json:"priority"`
	CreatedAt time.Time   `json:"created_at"`
	StartedAt *time.Time  `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Status    string      `json:"status"` // pending, running, completed, failed
	Result    interface{} `json:"result,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// LiteSkill 轻量级技能接口
type LiteSkill interface {
	ID() string
	Actions() []string
	Execute(ctx context.Context, action string, params interface{}) (interface{}, error)
	PowerConsumption() int // 功耗等级 1-5
}

// LiteConnection 轻量级连接接口
type LiteConnection interface {
	Connect(ctx context.Context) error
	Disconnect() error
	Send(msg *LiteMessage) error
	Receive() (*LiteMessage, error)
	IsConnected() bool
}

// NewLiteAgent 创建Lite Agent
func NewLiteAgent(config LiteConfig) *LiteAgent {
	// 默认配置
	if config.HeartbeatInt == 0 {
		config.HeartbeatInt = 60 * time.Second // 低功耗，60秒心跳
	}
	if config.ReconnectInt == 0 {
		config.ReconnectInt = 30 * time.Second
	}
	if config.MaxReconnect == 0 {
		config.MaxReconnect = 5
	}
	if config.BufferSize == 0 {
		config.BufferSize = 10 // 小缓冲区，节省内存
	}
	if config.MaxMessageSize == 0 {
		config.MaxMessageSize = 4096 // 4KB最大消息
	}
	if config.DeviceType == "" {
		config.DeviceType = "watch"
	}

	return &LiteAgent{
		config:       config,
		skills:       make(map[string]LiteSkill),
		status:       &AgentStatus{Status: "offline"},
		messageQueue: make(chan *LiteMessage, config.BufferSize),
		pendingTasks: make(map[string]*LiteTask),
		battery:      &BatteryInfo{Health: "good"},
		sensors:      NewSensorManager(),
	}
}

// NewSensorManager 创建传感器管理器
func NewSensorManager() *SensorManager {
	return &SensorManager{
		accelerometer: &Sensor{Name: "accelerometer", Type: "motion"},
		gyroscope:     &Sensor{Name: "gyroscope", Type: "motion"},
		heartRate:     &Sensor{Name: "heart_rate", Type: "health"},
		stepCounter:   &Sensor{Name: "step_counter", Type: "activity"},
		gps:           &Sensor{Name: "gps", Type: "location"},
		temperature:   &Sensor{Name: "temperature", Type: "environment"},
		light:         &Sensor{Name: "light", Type: "environment"},
	}
}

// Start 启动Agent
func (la *LiteAgent) Start(ctx context.Context) error {
	la.mu.Lock()
	la.running = true
	la.status.Status = "online"
	la.status.LastActive = time.Now()
	la.mu.Unlock()

	// 连接服务器
	if la.conn != nil {
		if err := la.conn.Connect(ctx); err != nil {
			return fmt.Errorf("连接失败: %w", err)
		}
	}

	// 启动后台任务
	go la.heartbeatLoop(ctx)
	go la.messageLoop(ctx)
	go la.powerMonitor(ctx)

	return nil
}

// Stop 停止Agent
func (la *LiteAgent) Stop() {
	la.mu.Lock()
	la.running = false
	la.status.Status = "offline"
	la.mu.Unlock()

	if la.conn != nil {
		la.conn.Disconnect()
	}
}

// RegisterSkill 注册技能
func (la *LiteAgent) RegisterSkill(skill LiteSkill) {
	la.mu.Lock()
	defer la.mu.Unlock()
	la.skills[skill.ID()] = skill
}

// ExecuteTask 执行任务
func (la *LiteAgent) ExecuteTask(ctx context.Context, task *LiteTask) (interface{}, error) {
	// 检查电池
	if la.config.BatterySaver && la.battery.Level < 20 && !la.battery.Charging {
		return nil, fmt.Errorf("电量过低，暂停任务执行")
	}

	// 获取技能
	la.mu.RLock()
	skill, ok := la.skills[task.Skill]
	la.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("技能不存在: %s", task.Skill)
	}

	// 执行任务
	task.Status = "running"
	now := time.Now()
	task.StartedAt = &now

	result, err := skill.Execute(ctx, task.Action, task.Params)
	if err != nil {
		task.Status = "failed"
		task.Error = err.Error()
		return nil, err
	}

	task.Status = "completed"
	now = time.Now()
	task.CompletedAt = &now
	task.Result = result

	return result, nil
}

// heartbeatLoop 心跳循环
func (la *LiteAgent) heartbeatLoop(ctx context.Context) {
	// 根据省电模式调整心跳间隔
	interval := la.config.HeartbeatInt
	if la.config.BatterySaver && la.battery.Level < 30 {
		interval = interval * 2 // 省电模式，心跳加倍
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			la.sendHeartbeat(ctx)
		}
	}
}

// sendHeartbeat 发送心跳
func (la *LiteAgent) sendHeartbeat(ctx context.Context) {
	la.mu.RLock()
	status := AgentStatus{
		ID:           la.status.ID,
		Type:         la.status.Type,
		Status:       la.status.Status,
		BatteryLevel: la.battery.Level,
		Charging:     la.battery.Charging,
		LastActive:   time.Now(),
	}
	la.mu.RUnlock()

	// 压缩状态数据
	data, _ := json.Marshal(status)
	msg := &LiteMessage{
		Type:      TypeHeartbeat,
		Timestamp: uint32(time.Now().Unix()),
		Payload:   data,
		Priority:  0,
	}

	if la.config.CompressData && len(data) > 100 {
		msg.Compressed = true
		// 实际压缩逻辑
	}

	la.sendMessage(msg)
}

// messageLoop 消息循环
func (la *LiteAgent) messageLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-la.messageQueue:
			la.handleMessage(ctx, msg)
		}
	}
}

// handleMessage 处理消息
func (la *LiteAgent) handleMessage(ctx context.Context, msg *LiteMessage) {
	switch msg.Type {
	case TypeTask:
		la.handleTask(ctx, msg)
	case TypeNotification:
		la.handleNotification(msg)
	case TypeError:
		la.handleError(msg)
	}
}

// handleTask 处理任务
func (la *LiteAgent) handleTask(ctx context.Context, msg *LiteMessage) {
	var task LiteTask
	if err := json.Unmarshal(msg.Payload, &task); err != nil {
		return
	}

	la.mu.Lock()
	la.pendingTasks[task.ID] = &task
	la.mu.Unlock()

	// 执行任务
	result, err := la.ExecuteTask(ctx, &task)

	// 发送结果
	resultMsg := &LiteMessage{
		ID:        msg.ID,
		Type:      TypeResult,
		Timestamp: uint32(time.Now().Unix()),
		Priority:  msg.Priority,
	}

	if err != nil {
		resultMsg.Type = TypeError
		resultMsg.Payload = []byte(err.Error())
	} else {
		resultMsg.Payload, _ = json.Marshal(result)
	}

	la.sendMessage(resultMsg)

	// 清理
	la.mu.Lock()
	delete(la.pendingTasks, task.ID)
	la.mu.Unlock()
}

// handleNotification 处理通知
func (la *LiteAgent) handleNotification(msg *LiteMessage) {
	// 显示通知(震动/屏幕)
	// 由具体平台实现
}

// handleError 处理错误
func (la *LiteAgent) handleError(msg *LiteMessage) {
	// 记录错误
}

// sendMessage 发送消息
func (la *LiteAgent) sendMessage(msg *LiteMessage) {
	if la.conn != nil && la.conn.IsConnected() {
		la.conn.Send(msg)
	}
}

// powerMonitor 电源监控
func (la *LiteAgent) powerMonitor(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			la.updatePowerStatus()
		}
	}
}

// updatePowerStatus 更新电源状态
func (la *LiteAgent) updatePowerStatus() {
	la.mu.Lock()
	defer la.mu.Unlock()

	// 模拟电池消耗(实际由平台提供)
	// 这里只是示例
	la.battery.LastUpdate = time.Now()

	// 省电模式自动调整
	if la.config.BatterySaver && la.battery.Level < 20 {
		// 进入深度省电
		la.config.HeartbeatInt = 120 * time.Second
	}
}

// === 传感器管理 ===

// EnableSensor 启用传感器
func (la *LiteAgent) EnableSensor(sensorType string, interval time.Duration) error {
	la.sensors.mu.Lock()
	defer la.sensors.mu.Unlock()

	var sensor *Sensor
	switch sensorType {
	case "accelerometer":
		sensor = la.sensors.accelerometer
	case "gyroscope":
		sensor = la.sensors.gyroscope
	case "heart_rate":
		sensor = la.sensors.heartRate
	case "step_counter":
		sensor = la.sensors.stepCounter
	case "gps":
		sensor = la.sensors.gps
	case "temperature":
		sensor = la.sensors.temperature
	case "light":
		sensor = la.sensors.light
	default:
		return fmt.Errorf("未知传感器: %s", sensorType)
	}

	// 检查功耗
	if la.config.BatterySaver && la.battery.Level < 30 {
		if sensorType == "gps" || sensorType == "heart_rate" {
			return fmt.Errorf("省电模式，高功耗传感器不可用")
		}
	}

	sensor.Active = true
	sensor.Interval = interval

	return nil
}

// DisableSensor 禁用传感器
func (la *LiteAgent) DisableSensor(sensorType string) error {
	la.sensors.mu.Lock()
	defer la.sensors.mu.Unlock()

	switch sensorType {
	case "accelerometer":
		la.sensors.accelerometer.Active = false
	case "gyroscope":
		la.sensors.gyroscope.Active = false
	case "heart_rate":
		la.sensors.heartRate.Active = false
	case "step_counter":
		la.sensors.stepCounter.Active = false
	case "gps":
		la.sensors.gps.Active = false
	case "temperature":
		la.sensors.temperature.Active = false
	case "light":
		la.sensors.light.Active = false
	default:
		return fmt.Errorf("未知传感器: %s", sensorType)
	}

	return nil
}

// GetSensorData 获取传感器数据
func (la *LiteAgent) GetSensorData(sensorType string) (interface{}, error) {
	la.sensors.mu.RLock()
	defer la.sensors.mu.RUnlock()

	var sensor *Sensor
	switch sensorType {
	case "accelerometer":
		sensor = la.sensors.accelerometer
	case "gyroscope":
		sensor = la.sensors.gyroscope
	case "heart_rate":
		sensor = la.sensors.heartRate
	case "step_counter":
		sensor = la.sensors.stepCounter
	case "gps":
		sensor = la.sensors.gps
	case "temperature":
		sensor = la.sensors.temperature
	case "light":
		sensor = la.sensors.light
	default:
		return nil, fmt.Errorf("未知传感器: %s", sensorType)
	}

	if !sensor.Active {
		return nil, fmt.Errorf("传感器未启用: %s", sensorType)
	}

	return sensor.LastValue, nil
}

// UpdateSensorData 更新传感器数据(平台调用)
func (la *LiteAgent) UpdateSensorData(sensorType string, value interface{}) {
	la.sensors.mu.Lock()
	defer la.sensors.mu.Unlock()

	var sensor *Sensor
	switch sensorType {
	case "accelerometer":
		sensor = la.sensors.accelerometer
	case "gyroscope":
		sensor = la.sensors.gyroscope
	case "heart_rate":
		sensor = la.sensors.heartRate
	case "step_counter":
		sensor = la.sensors.stepCounter
	case "gps":
		sensor = la.sensors.gps
	case "temperature":
		sensor = la.sensors.temperature
	case "light":
		sensor = la.sensors.light
	default:
		return
	}

	sensor.LastValue = value
	sensor.UpdatedAt = time.Now()
}

// UpdateBattery 更新电池信息(平台调用)
func (la *LiteAgent) UpdateBattery(level int, charging bool, voltage, temperature float64) {
	la.mu.Lock()
	defer la.mu.Unlock()

	la.battery.Level = level
	la.battery.Charging = charging
	la.battery.Voltage = voltage
	la.battery.Temperature = temperature
	la.battery.LastUpdate = time.Now()

	// 评估电池健康
	if level < 20 {
		la.battery.Health = "poor"
	} else if level < 50 {
		la.battery.Health = "fair"
	} else {
		la.battery.Health = "good"
	}

	la.status.BatteryLevel = level
	la.status.Charging = charging
}

// SetConnection 设置连接
func (la *LiteAgent) SetConnection(conn LiteConnection) {
	la.conn = conn
}

// GetStatus 获取状态
func (la *LiteAgent) GetStatus() *AgentStatus {
	la.mu.RLock()
	defer la.mu.RUnlock()
	return la.status
}

// GetBatteryInfo 获取电池信息
func (la *LiteAgent) GetBatteryInfo() *BatteryInfo {
	la.mu.RLock()
	defer la.mu.RUnlock()
	return la.battery
}

// GetSkills 获取技能列表
func (la *LiteAgent) GetSkills() []string {
	la.mu.RLock()
	defer la.mu.RUnlock()

	skills := make([]string, 0, len(la.skills))
	for id := range la.skills {
		skills = append(skills, id)
	}
	return skills
}

// === 消息编解码 ===

// Encode 编码消息(二进制格式，节省带宽)
func (m *LiteMessage) Encode() []byte {
	size := 8 + len(m.Payload) // header + payload
	buf := make([]byte, size)

	// Header: ID(2) + Type(1) + Priority(1) + Compressed(1) + Timestamp(4)
	binary.BigEndian.PutUint16(buf[0:2], m.ID)
	buf[2] = byte(m.Type)
	buf[3] = m.Priority
	buf[4] = 0
	if m.Compressed {
		buf[4] = 1
	}
	binary.BigEndian.PutUint32(buf[5:9], m.Timestamp)

	// Payload
	copy(buf[9:], m.Payload)

	return buf
}

// Decode 解码消息
func DecodeMessage(data []byte) (*LiteMessage, error) {
	if len(data) < 9 {
		return nil, fmt.Errorf("消息长度不足")
	}

	msg := &LiteMessage{
		ID:        binary.BigEndian.Uint16(data[0:2]),
		Type:      MessageType(data[2]),
		Priority:  data[3],
		Compressed: data[4] == 1,
		Timestamp: binary.BigEndian.Uint32(data[5:9]),
		Payload:   data[9:],
	}

	return msg, nil
}

// === 内置技能 ===

// HeartRateSkill 心率技能
type HeartRateSkill struct{}

func (s *HeartRateSkill) ID() string           { return "health.heart_rate" }
func (s *HeartRateSkill) Actions() []string    { return []string{"measure", "history"} }
func (s *HeartRateSkill) PowerConsumption() int { return 3 }
func (s *HeartRateSkill) Execute(ctx context.Context, action string, params interface{}) (interface{}, error) {
	switch action {
	case "measure":
		// 触发心率测量
		return map[string]interface{}{"bpm": 72, "status": "normal"}, nil
	case "history":
		// 返回历史数据
		return []map[string]interface{}{}, nil
	default:
		return nil, fmt.Errorf("未知操作: %s", action)
	}
}

// StepCountSkill 计步技能
type StepCountSkill struct{}

func (s *StepCountSkill) ID() string           { return "activity.step_count" }
func (s *StepCountSkill) Actions() []string    { return []string{"get", "reset"} }
func (s *StepCountSkill) PowerConsumption() int { return 1 }
func (s *StepCountSkill) Execute(ctx context.Context, action string, params interface{}) (interface{}, error) {
	switch action {
	case "get":
		return map[string]interface{}{"steps": 8500, "distance": 6.2, "calories": 320}, nil
	case "reset":
		return map[string]interface{}{"success": true}, nil
	default:
		return nil, fmt.Errorf("未知操作: %s", action)
	}
}

// LocationSkill 位置技能
type LocationSkill struct{}

func (s *LocationSkill) ID() string           { return "location.gps" }
func (s *LocationSkill) Actions() []string    { return []string{"get", "track_start", "track_stop"} }
func (s *LocationSkill) PowerConsumption() int { return 5 } // GPS功耗最高
func (s *LocationSkill) Execute(ctx context.Context, action string, params interface{}) (interface{}, error) {
	switch action {
	case "get":
		return map[string]interface{}{
			"lat":       39.9042,
			"lon":       116.4074,
			"accuracy":  10.5,
			"timestamp": time.Now().Unix(),
		}, nil
	case "track_start":
		return map[string]interface{}{"tracking": true}, nil
	case "track_stop":
		return map[string]interface{}{"tracking": false}, nil
	default:
		return nil, fmt.Errorf("未知操作: %s", action)
	}
}

// NotificationSkill 通知技能
type NotificationSkill struct{}

func (s *NotificationSkill) ID() string           { return "notification.show" }
func (s *NotificationSkill) Actions() []string    { return []string{"alert", "vibrate", "dismiss"} }
func (s *NotificationSkill) PowerConsumption() int { return 1 }
func (s *NotificationSkill) Execute(ctx context.Context, action string, params interface{}) (interface{}, error) {
	switch action {
	case "alert":
		return map[string]interface{}{"shown": true}, nil
	case "vibrate":
		return map[string]interface{}{"vibrated": true}, nil
	case "dismiss":
		return map[string]interface{}{"dismissed": true}, nil
	default:
		return nil, fmt.Errorf("未知操作: %s", action)
	}
}

// IsRunning 检查运行状态
func (la *LiteAgent) IsRunning() bool {
	la.mu.RLock()
	defer la.mu.RUnlock()
	return la.running
}