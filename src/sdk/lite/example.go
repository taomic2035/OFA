// Package lite - Lite Agent示例代码
package lite

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// ExampleUsage 示例用法
func ExampleUsage() {
	// 1. 创建Lite Agent配置
	config := LiteConfig{
		ServerURL:      "localhost:9090",
		AgentID:        "watch-001",
		DeviceType:     "watch",
		HeartbeatInt:   60 * time.Second,
		BatterySaver:   true,
		CompressData:   true,
		BufferSize:     10,
		MaxMessageSize: 4096,
	}

	// 2. 创建Agent
	agent := NewLiteAgent(config)

	// 3. 设置连接(选择TCP/BLE/WebSocket)
	conn := NewTCPConnection(config.ServerURL)
	agent.SetConnection(conn)

	// 4. 注册技能
	agent.RegisterSkill(&HeartRateSkill{})
	agent.RegisterSkill(&StepCountSkill{})
	agent.RegisterSkill(&LocationSkill{})
	agent.RegisterSkill(&NotificationSkill{})

	// 5. 启动Agent
	ctx := context.Background()
	if err := agent.Start(ctx); err != nil {
		fmt.Printf("启动失败: %v\n", err)
		os.Exit(1)
	}

	// 6. 启用传感器(根据电池状态动态调整)
	agent.EnableSensor("step_counter", 1*time.Second)
	if battery := agent.GetBatteryInfo(); battery.Level > 30 {
		agent.EnableSensor("heart_rate", 5*time.Second)
	}

	// 7. 执行任务示例
	task := &LiteTask{
		ID:       "task-001",
		Skill:    "health.heart_rate",
		Action:   "measure",
		Timeout:  10 * time.Second,
		Priority: 1,
	}

	result, err := agent.ExecuteTask(ctx, task)
	if err != nil {
		fmt.Printf("任务执行失败: %v\n", err)
	} else {
		fmt.Printf("结果: %v\n", result)
	}

	// 8. 监听传感器数据
	go func() {
		for {
			if data, err := agent.GetSensorData("heart_rate"); err == nil {
				fmt.Printf("心率: %v\n", data)
			}
			time.Sleep(5 * time.Second)
		}
	}()

	// 9. 停止Agent
	defer agent.Stop()
}

// SmartWatchExample 智能手表示例
func SmartWatchExample() {
	config := LiteConfig{
		ServerURL:    "192.168.1.100:9090",
		AgentID:      "smartwatch-001",
		DeviceType:   "watch",
		HeartbeatInt: 60 * time.Second,
		BatterySaver: true, // 启用省电模式
		CompressData: true,
	}

	agent := NewLiteAgent(config)

	// 注册手表专用技能
	agent.RegisterSkill(&HeartRateSkill{})
	agent.RegisterSkill(&StepCountSkill{})
	agent.RegisterSkill(&LocationSkill{})
	agent.RegisterSkill(&NotificationSkill{})

	// 根据场景调整行为
	agent.UpdateBattery(85, false, 3.8, 25)

	// 正常模式：启用所有传感器
	agent.EnableSensor("step_counter", 1*time.Second)
	agent.EnableSensor("heart_rate", 10*time.Second)
	agent.EnableSensor("accelerometer", 100*time.Millisecond)

	// 模拟运动模式
	fmt.Println("=== 运动模式 ===")
	agent.EnableSensor("gps", 1*time.Second) // GPS高功耗
	agent.EnableSensor("heart_rate", 1*time.Second)

	// 模拟省电模式
	agent.UpdateBattery(15, false, 3.5, 20)
	fmt.Println("=== 省电模式 ===")
	agent.DisableSensor("gps")      // 关闭GPS
	agent.DisableSensor("heart_rate") // 减少心率检测
}

// FitnessBandExample 运动手环示例
func FitnessBandExample() {
	config := LiteConfig{
		ServerURL:    "192.168.1.100:9090",
		AgentID:      "band-001",
		DeviceType:   "band",
		HeartbeatInt: 120 * time.Second, // 更长心跳间隔
		BatterySaver: true,
		CompressData: true,
		BufferSize:   5, // 更小缓冲区
	}

	agent := NewLiteAgent(config)

	// 注册手环技能(精简版)
	agent.RegisterSkill(&HeartRateSkill{})
	agent.RegisterSkill(&StepCountSkill{})
	agent.RegisterSkill(&NotificationSkill{})

	// 手环功耗更低
	agent.EnableSensor("step_counter", 1*time.Second)
	// 心率仅每分钟检测一次
	agent.EnableSensor("heart_rate", 60*time.Second)
}

// SleepModeExample 睡眠模式示例
func SleepModeExample(agent *LiteAgent) {
	// 进入睡眠模式
	agent.DisableSensor("heart_rate")
	agent.DisableSensor("accelerometer")
	agent.DisableSensor("gps")
	agent.DisableSensor("light")

	// 只保留必要的计步
	agent.EnableSensor("step_counter", 10*time.Second)

	// 心跳间隔延长到5分钟
	agent.config.HeartbeatInt = 300 * time.Second
}

// ChargingModeExample 充电模式示例
func ChargingModeExample(agent *LiteAgent) {
	// 更新电池状态
	agent.UpdateBattery(50, true, 4.0, 30)

	// 充电时可以执行高功耗任务
	agent.EnableSensor("gps", 1*time.Second)
	agent.EnableSensor("heart_rate", 1*time.Second)

	// 关闭省电模式
	agent.config.BatterySaver = false
	agent.config.HeartbeatInt = 30 * time.Second
}

// DataSyncExample 数据同步示例
func DataSyncExample(agent *LiteAgent) {
	// 收集健康数据
	heartRateData, _ := agent.GetSensorData("heart_rate")
	stepData, _ := agent.GetSensorData("step_counter")
	locationData, _ := agent.GetSensorData("gps")

	// 打包数据
	syncData := map[string]interface{}{
		"heart_rate": heartRateData,
		"steps":      stepData,
		"location":   locationData,
		"timestamp":  time.Now().Unix(),
	}

	// 发送到Center
	data, _ := json.Marshal(syncData)
	msg := &LiteMessage{
		Type:      TypeSensor,
		Timestamp: uint32(time.Now().Unix()),
		Payload:   data,
		Priority:  1,
	}

	if agent.conn != nil && agent.conn.IsConnected() {
		agent.conn.Send(msg)
	}
}

// SensorDataCollector 传感器数据收集器
type SensorDataCollector struct {
	agent    *LiteAgent
	interval time.Duration
	data     map[string][]interface{}
}

// NewSensorDataCollector 创建数据收集器
func NewSensorDataCollector(agent *LiteAgent, interval time.Duration) *SensorDataCollector {
	return &SensorDataCollector{
		agent:    agent,
		interval: interval,
		data:     make(map[string][]interface{}),
	}
}

// Start 开始收集
func (sdc *SensorDataCollector) Start(ctx context.Context) {
	ticker := time.NewTicker(sdc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sdc.collect()
		}
	}
}

// collect 收集数据
func (sdc *SensorDataCollector) collect() {
	// 收集心率
	if hr, err := sdc.agent.GetSensorData("heart_rate"); err == nil {
		sdc.data["heart_rate"] = append(sdc.data["heart_rate"], hr)
		if len(sdc.data["heart_rate"]) > 100 {
			sdc.data["heart_rate"] = sdc.data["heart_rate"][1:]
		}
	}

	// 收集步数
	if steps, err := sdc.agent.GetSensorData("step_counter"); err == nil {
		sdc.data["steps"] = append(sdc.data["steps"], steps)
		if len(sdc.data["steps"]) > 100 {
			sdc.data["steps"] = sdc.data["steps"][1:]
		}
	}
}

// GetAverageHeartRate 获取平均心率
func (sdc *SensorDataCollector) GetAverageHeartRate() float64 {
	hrData := sdc.data["heart_rate"]
	if len(hrData) == 0 {
		return 0
	}

	var total float64
	for _, d := range hrData {
		if m, ok := d.(map[string]interface{}); ok {
			if bpm, ok := m["bpm"].(float64); ok {
				total += bpm
			}
		}
	}

	return total / float64(len(hrData))
}

// PowerOptimizer 功耗优化器
type PowerOptimizer struct {
	agent        *LiteAgent
	lowThreshold  int // 低电量阈值
	criticalLevel int // 临界电量
}

// NewPowerOptimizer 创建功耗优化器
func NewPowerOptimizer(agent *LiteAgent) *PowerOptimizer {
	return &PowerOptimizer{
		agent:        agent,
		lowThreshold:  30,
		criticalLevel: 15,
	}
}

// Optimize 根据电量优化
func (po *PowerOptimizer) Optimize() {
	battery := po.agent.GetBatteryInfo()

	switch {
	case battery.Level < po.criticalLevel:
		// 临界模式
		po.applyCriticalMode()
	case battery.Level < po.lowThreshold:
		// 省电模式
		po.applyPowerSaveMode()
	case battery.Charging:
		// 充电模式
		po.applyChargingMode()
	default:
		// 正常模式
		po.applyNormalMode()
	}
}

func (po *PowerOptimizer) applyCriticalMode() {
	// 关闭所有非必要功能
	po.agent.DisableSensor("gps")
	po.agent.DisableSensor("heart_rate")
	po.agent.DisableSensor("accelerometer")
	po.agent.config.HeartbeatInt = 300 * time.Second
	po.agent.config.BatterySaver = true
}

func (po *PowerOptimizer) applyPowerSaveMode() {
	// 减少检测频率
	po.agent.DisableSensor("gps")
	po.agent.EnableSensor("heart_rate", 60*time.Second)
	po.agent.config.HeartbeatInt = 120 * time.Second
	po.agent.config.BatterySaver = true
}

func (po *PowerOptimizer) applyChargingMode() {
	// 充电时全功能运行
	po.agent.EnableSensor("gps", 5*time.Second)
	po.agent.EnableSensor("heart_rate", 5*time.Second)
	po.agent.config.HeartbeatInt = 30 * time.Second
	po.agent.config.BatterySaver = false
}

func (po *PowerOptimizer) applyNormalMode() {
	// 正常使用
	po.agent.DisableSensor("gps")
	po.agent.EnableSensor("heart_rate", 10*time.Second)
	po.agent.config.HeartbeatInt = 60 * time.Second
}