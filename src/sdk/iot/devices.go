// Package iot - 设备类型定义
package iot

import (
	"context"
	"fmt"
	"time"
)

// DeviceType 设备类型
type DeviceType string

const (
	TypeLight       DeviceType = "light"        // 智能灯
	TypeSwitch      DeviceType = "switch"       // 智能开关
	TypeSocket      DeviceType = "socket"       // 智能插座
	TypeSensor      DeviceType = "sensor"       // 传感器
	TypeThermostat  DeviceType = "thermostat"   // 温控器
	TypeLock        DeviceType = "lock"         // 智能门锁
	TypeCamera      DeviceType = "camera"       // 摄像头
	TypeSpeaker     DeviceType = "speaker"      // 智能音箱
	TypeAppliance   DeviceType = "appliance"    // 智能家电
	TypeGateway     DeviceType = "gateway"      // 网关
)

// DeviceCapability 设备能力
type DeviceCapability string

const (
	CapPower       DeviceCapability = "power"        // 电源控制
	CapBrightness  DeviceCapability = "brightness"   // 亮度调节
	CapColor       DeviceCapability = "color"        // 颜色控制
	CapTemperature DeviceCapability = "temperature"  // 温度控制
	CapHumidity    DeviceCapability = "humidity"     // 湿度监测
	CapMotion      DeviceCapability = "motion"       // 运动检测
	CapLock        DeviceCapability = "lock"         // 门锁控制
	CapCamera      DeviceCapability = "camera"       // 摄像头
	CapSpeaker     DeviceCapability = "speaker"      // 音频播放
	CapScreen      DeviceCapability = "screen"       // 屏幕显示
)

// === 智能灯 ===

// SmartLight 智能灯设备
type SmartLight struct {
	agent      *IoTAgent
	power      bool
	brightness int // 0-100
	color      *RGBColor
	colorTemp  int   // 色温 K
	mode       string // white, color, scene
}

// RGBColor RGB颜色
type RGBColor struct {
	R int `json:"r"` // 0-255
	G int `json:"g"` // 0-255
	B int `json:"b"` // 0-255
}

// NewSmartLight 创建智能灯
func NewSmartLight(deviceID, name string) *SmartLight {
	config := IoTConfig{
		DeviceID:   deviceID,
		DeviceType: string(TypeLight),
		DeviceName: name,
		ShadowEnabled: true,
	}

	return &SmartLight{
		agent:      NewIoTAgent(config),
		brightness: 100,
		color:      &RGBColor{R: 255, G: 255, B: 255},
		colorTemp:  4000,
		mode:       "white",
	}
}

// TurnOn 开灯
func (l *SmartLight) TurnOn() error {
	l.power = true
	return l.agent.SetProperty("power", true)
}

// TurnOff 关灯
func (l *SmartLight) TurnOff() error {
	l.power = false
	return l.agent.SetProperty("power", false)
}

// SetBrightness 设置亮度
func (l *SmartLight) SetBrightness(level int) error {
	if level < 0 || level > 100 {
		return fmt.Errorf("亮度范围0-100")
	}
	l.brightness = level
	return l.agent.SetProperty("brightness", level)
}

// SetColor 设置颜色
func (l *SmartLight) SetColor(r, g, b int) error {
	l.color = &RGBColor{R: r, G: g, B: b}
	l.mode = "color"
	return l.agent.SetProperty("color", map[string]int{"r": r, "g": g, "b": b})
}

// SetColorTemp 设置色温
func (l *SmartLight) SetColorTemp(kelvin int) error {
	l.colorTemp = kelvin
	l.mode = "white"
	return l.agent.SetProperty("color_temp", kelvin)
}

// GetState 获取状态
func (l *SmartLight) GetState() map[string]interface{} {
	return map[string]interface{}{
		"power":      l.power,
		"brightness": l.brightness,
		"color":      l.color,
		"color_temp": l.colorTemp,
		"mode":       l.mode,
	}
}

// === 智能插座 ===

// SmartSocket 智能插座
type SmartSocket struct {
	agent   *IoTAgent
	power   bool
	voltage float64
	current float64
	powerW  float64 // 功率(瓦)
}

// NewSmartSocket 创建智能插座
func NewSmartSocket(deviceID, name string) *SmartSocket {
	config := IoTConfig{
		DeviceID:   deviceID,
		DeviceType: string(TypeSocket),
		DeviceName: name,
		ShadowEnabled: true,
	}

	return &SmartSocket{
		agent: NewIoTAgent(config),
	}
}

// TurnOn 开启插座
func (s *SmartSocket) TurnOn() error {
	s.power = true
	return s.agent.SetProperty("power", true)
}

// TurnOff 关闭插座
func (s *SmartSocket) TurnOff() error {
	s.power = false
	return s.agent.SetProperty("power", false)
}

// UpdatePowerUsage 更新用电数据
func (s *SmartSocket) UpdatePowerUsage(voltage, current, powerW float64) error {
	s.voltage = voltage
	s.current = current
	s.powerW = powerW

	// 发布遥测数据
	return s.agent.PublishTelemetry(map[string]interface{}{
		"voltage": voltage,
		"current": current,
		"power":   powerW,
	})
}

// GetState 获取状态
func (s *SmartSocket) GetState() map[string]interface{} {
	return map[string]interface{}{
		"power":   s.power,
		"voltage": s.voltage,
		"current": s.current,
		"power_w": s.powerW,
	}
}

// === 温湿度传感器 ===

// TempHumiditySensor 温湿度传感器
type TempHumiditySensor struct {
	agent       *IoTAgent
	temperature float64
	humidity    float64
	battery     int
}

// NewTempHumiditySensor 创建温湿度传感器
func NewTempHumiditySensor(deviceID, name string) *TempHumiditySensor {
	config := IoTConfig{
		DeviceID:   deviceID,
		DeviceType: string(TypeSensor),
		DeviceName: name,
		ShadowEnabled: true,
	}

	return &TempHumiditySensor{
		agent:   NewIoTAgent(config),
		battery: 100,
	}
}

// UpdateReadings 更新读数
func (s *TempHumiditySensor) UpdateReadings(temp, humidity float64) error {
	s.temperature = temp
	s.humidity = humidity

	// 更新影子属性
	s.agent.SetProperty("temperature", temp)
	s.agent.SetProperty("humidity", humidity)

	// 发布遥测数据
	return s.agent.PublishTelemetry(map[string]interface{}{
		"temperature": temp,
		"humidity":    humidity,
	})
}

// UpdateBattery 更新电池电量
func (s *TempHumiditySensor) UpdateBattery(level int) error {
	s.battery = level
	return s.agent.SetProperty("battery", level)
}

// GetState 获取状态
func (s *TempHumiditySensor) GetState() map[string]interface{} {
	return map[string]interface{}{
		"temperature": s.temperature,
		"humidity":    s.humidity,
		"battery":     s.battery,
	}
}

// === 智能门锁 ===

// SmartLock 智能门锁
type SmartLock struct {
	agent     *IoTAgent
	locked    bool
	battery   int
	lastUser  string
	logs      []LockLog
}

// LockLog 门锁日志
type LockLog struct {
	Time    time.Time `json:"time"`
	Action  string    `json:"action"`  // lock, unlock
	User    string    `json:"user"`
	Method  string    `json:"method"`  // password, fingerprint, app, key
}

// NewSmartLock 创建智能门锁
func NewSmartLock(deviceID, name string) *SmartLock {
	config := IoTConfig{
		DeviceID:   deviceID,
		DeviceType: string(TypeLock),
		DeviceName: name,
		ShadowEnabled: true,
	}

	return &SmartLock{
		agent:   NewIoTAgent(config),
		battery: 100,
		logs:    make([]LockLog, 0),
	}
}

// Lock 上锁
func (l *SmartLock) Lock(user, method string) error {
	l.locked = true
	l.lastUser = user
	l.logs = append(l.logs, LockLog{
		Time:   time.Now(),
		Action: "lock",
		User:   user,
		Method: method,
	})

	l.agent.SetProperty("locked", true)
	return l.agent.PublishEvent("lock_change", map[string]interface{}{
		"action": "lock",
		"user":   user,
		"method": method,
	})
}

// Unlock 解锁
func (l *SmartLock) Unlock(user, method string) error {
	l.locked = false
	l.lastUser = user
	l.logs = append(l.logs, LockLog{
		Time:   time.Now(),
		Action: "unlock",
		User:   user,
		Method: method,
	})

	l.agent.SetProperty("locked", false)
	return l.agent.PublishEvent("lock_change", map[string]interface{}{
		"action": "unlock",
		"user":   user,
		"method": method,
	})
}

// UpdateBattery 更新电池电量
func (l *SmartLock) UpdateBattery(level int) error {
	l.battery = level
	return l.agent.SetProperty("battery", level)
}

// GetState 获取状态
func (l *SmartLock) GetState() map[string]interface{} {
	return map[string]interface{}{
		"locked":    l.locked,
		"battery":   l.battery,
		"last_user": l.lastUser,
	}
}

// === 温控器 ===

// Thermostat 温控器
type Thermostat struct {
	agent        *IoTAgent
	mode         string // off, heat, cool, auto
	targetTemp   float64
	currentTemp  float64
	fanSpeed     string // auto, low, medium, high
	humidity     float64
}

// NewThermostat 创建温控器
func NewThermostat(deviceID, name string) *Thermostat {
	config := IoTConfig{
		DeviceID:   deviceID,
		DeviceType: string(TypeThermostat),
		DeviceName: name,
		ShadowEnabled: true,
	}

	return &Thermostat{
		agent:      NewIoTAgent(config),
		mode:       "off",
		targetTemp: 22.0,
		fanSpeed:   "auto",
	}
}

// SetMode 设置模式
func (t *Thermostat) SetMode(mode string) error {
	validModes := map[string]bool{"off": true, "heat": true, "cool": true, "auto": true}
	if !validModes[mode] {
		return fmt.Errorf("无效模式: %s", mode)
	}
	t.mode = mode
	return t.agent.SetProperty("mode", mode)
}

// SetTargetTemp 设置目标温度
func (t *Thermostat) SetTargetTemp(temp float64) error {
	if temp < 10 || temp > 35 {
		return fmt.Errorf("温度范围10-35°C")
	}
	t.targetTemp = temp
	return t.agent.SetProperty("target_temp", temp)
}

// SetFanSpeed 设置风速
func (t *Thermostat) SetFanSpeed(speed string) error {
	validSpeeds := map[string]bool{"auto": true, "low": true, "medium": true, "high": true}
	if !validSpeeds[speed] {
		return fmt.Errorf("无效风速: %s", speed)
	}
	t.fanSpeed = speed
	return t.agent.SetProperty("fan_speed", speed)
}

// UpdateCurrentTemp 更新当前温度
func (t *Thermostat) UpdateCurrentTemp(temp, humidity float64) error {
	t.currentTemp = temp
	t.humidity = humidity

	t.agent.SetProperty("current_temp", temp)
	t.agent.SetProperty("humidity", humidity)

	return t.agent.PublishTelemetry(map[string]interface{}{
		"current_temp": temp,
		"humidity":     humidity,
	})
}

// GetState 获取状态
func (t *Thermostat) GetState() map[string]interface{} {
	return map[string]interface{}{
		"mode":         t.mode,
		"target_temp":  t.targetTemp,
		"current_temp": t.currentTemp,
		"fan_speed":    t.fanSpeed,
		"humidity":     t.humidity,
	}
}

// === 设备工厂 ===

// DeviceFactory 设备工厂
type DeviceFactory struct{}

// CreateDevice 创建设备
func (f *DeviceFactory) CreateDevice(deviceType DeviceType, deviceID, name string) (interface{}, error) {
	switch deviceType {
	case TypeLight:
		return NewSmartLight(deviceID, name), nil
	case TypeSocket:
		return NewSmartSocket(deviceID, name), nil
	case TypeSensor:
		return NewTempHumiditySensor(deviceID, name), nil
	case TypeLock:
		return NewSmartLock(deviceID, name), nil
	case TypeThermostat:
		return NewThermostat(deviceID, name), nil
	default:
		return nil, fmt.Errorf("不支持设备类型: %s", deviceType)
	}
}

// StartDevice 启动设备(通用方法)
func StartDevice(ctx context.Context, device interface{}) error {
	switch d := device.(type) {
	case *SmartLight:
		return d.agent.Start(ctx)
	case *SmartSocket:
		return d.agent.Start(ctx)
	case *TempHumiditySensor:
		return d.agent.Start(ctx)
	case *SmartLock:
		return d.agent.Start(ctx)
	case *Thermostat:
		return d.agent.Start(ctx)
	default:
		return fmt.Errorf("未知设备类型")
	}
}