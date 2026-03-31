// Package wasm - WebAssembly技能运行时
// 0.9.0 Beta: WASM技能支持
package wasm

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// WASMConfig WASM运行时配置
type WASMConfig struct {
	MaxMemoryMB     int           `json:"max_memory_mb"`      // 最大内存(MB)
	Timeout         time.Duration `json:"timeout"`            // 执行超时
	EnableFileSystem bool          `json:"enable_file_system"` // 启用文件系统
	EnableNetwork   bool          `json:"enable_network"`     // 启用网络
	MaxTableSize    int           `json:"max_table_size"`     // 最大表大小
	FuelLimit       uint64        `json:"fuel_limit"`         // 燃料限制(指令数)
}

// DefaultWASMConfig 默认配置
func DefaultWASMConfig() WASMConfig {
	return WASMConfig{
		MaxMemoryMB:     128,
		Timeout:         30 * time.Second,
		EnableFileSystem: false,
		EnableNetwork:   false,
		MaxTableSize:    65536,
		FuelLimit:       100000000, // 1亿指令
	}
}

// WASMSkill WASM技能
type WASMSkill struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Operations  []string          `json:"operations"`
	Binary      []byte            `json:"binary"`
	Exports     map[string]string `json:"exports"`
	Config      SkillConfig       `json:"config"`
	LoadedAt    time.Time         `json:"loaded_at"`
}

// SkillConfig 技能配置
type SkillConfig struct {
	MemoryPages int    `json:"memory_pages"`
	Timeout     int    `json:"timeout_ms"`
	Permissions []string `json:"permissions"`
}

// SkillResult 技能执行结果
type SkillResult struct {
	Success   bool            `json:"success"`
	Output    interface{}     `json:"output"`
	Error     string          `json:"error,omitempty"`
	Duration  time.Duration   `json:"duration"`
	MemoryUsed int64          `json:"memory_used"`
}

// WASMRuntime WASM运行时
type WASMRuntime struct {
	config     WASMConfig
	skills     map[string]*WASMSkill
	modules    map[string]api.Module
	runtime    wazero.Runtime
	mu         sync.RWMutex
	stats      *RuntimeStats
}

// RuntimeStats 运行时统计
type RuntimeStats struct {
	TotalLoads      int64         `json:"total_loads"`
	TotalExecutions int64         `json:"total_executions"`
	TotalErrors     int64         `json:"total_errors"`
	TotalDuration   time.Duration `json:"total_duration"`
	AvgDuration     time.Duration `json:"avg_duration"`
	MemoryUsed      int64         `json:"memory_used"`
}

// NewWASMRuntime 创建WASM运行时
func NewWASMRuntime(config WASMConfig) *WASMRuntime {
	ctx := context.Background()

	// 创建运行时
	rt := wazero.NewRuntime(ctx)

	return &WASMRuntime{
		config:  config,
		skills:  make(map[string]*WASMSkill),
		modules: make(map[string]api.Module),
		runtime: rt,
		stats:   &RuntimeStats{},
	}
}

// LoadSkill 加载WASM技能
func (w *WASMRuntime) LoadSkill(ctx context.Context, skill *WASMSkill) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.skills[skill.ID]; exists {
		return fmt.Errorf("技能已存在: %s", skill.ID)
	}

	// 编译模块
	compiled, err := w.runtime.CompileModule(ctx, skill.Binary)
	if err != nil {
		return fmt.Errorf("编译失败: %w", err)
	}

	// 实例化模块
	module, err := w.runtime.InstantiateModule(ctx, compiled, wazero.NewModuleConfig().
		WithName(skill.ID).
		WithMemoryLimitPages(uint32(w.config.MaxMemoryMB / 64)))

	if err != nil {
		return fmt.Errorf("实例化失败: %w", err)
	}

	skill.LoadedAt = time.Now()
	w.skills[skill.ID] = skill
	w.modules[skill.ID] = module

	w.stats.TotalLoads++

	return nil
}

// Execute 执行WASM技能
func (w *WASMRuntime) Execute(ctx context.Context, skillID, operation string, input interface{}) *SkillResult {
	start := time.Now()
	result := &SkillResult{}

	w.mu.RLock()
	module, exists := w.modules[skillID]
	w.mu.RUnlock()

	if !exists {
		result.Error = fmt.Sprintf("技能未加载: %s", skillID)
		return result
	}

	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, w.config.Timeout)
	defer cancel()

	// 获取导出函数
	fn := module.ExportedFunction(operation)
	if fn == nil {
		result.Error = fmt.Sprintf("操作不存在: %s", operation)
		return result
	}

	// 准备输入参数
	inputJSON, err := json.Marshal(input)
	if err != nil {
		result.Error = fmt.Sprintf("序列化输入失败: %v", err)
		return result
	}

	// 写入内存
	mem := module.Memory()
	if mem == nil {
		result.Error = "模块没有内存"
		return result
	}

	// 分配内存存储输入
	allocFn := module.ExportedFunction("alloc")
	if allocFn == nil {
		result.Error = "模块没有alloc函数"
		return result
	}

	// 调用alloc分配内存
	ptr, err := allocFn.Call(ctx, uint64(len(inputJSON)))
	if err != nil {
		result.Error = fmt.Sprintf("分配内存失败: %v", err)
		return result
	}

	// 写入数据到内存
	ok := mem.Write(uint32(ptr[0]), inputJSON)
	if !ok {
		result.Error = "写入内存失败"
		return result
	}

	// 执行技能函数
	output, err := fn.Call(ctx, ptr[0], uint64(len(inputJSON)))
	if err != nil {
		result.Error = fmt.Sprintf("执行失败: %v", err)
		w.stats.TotalErrors++
		return result
	}

	// 读取输出
	if len(output) >= 2 {
		outputPtr := uint32(output[0])
		outputLen := uint32(output[1])
		outputBytes, ok := mem.Read(outputPtr, outputLen)
		if ok {
			var outputData interface{}
			if err := json.Unmarshal(outputBytes, &outputData); err == nil {
				result.Output = outputData
			} else {
				result.Output = string(outputBytes)
			}
		}
	}

	result.Success = true
	result.Duration = time.Since(start)
	w.stats.TotalExecutions++
	w.stats.TotalDuration += result.Duration

	return result
}

// UnloadSkill 卸载技能
func (w *WASMRuntime) UnloadSkill(skillID string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	module, exists := w.modules[skillID]
	if !exists {
		return fmt.Errorf("技能不存在: %s", skillID)
	}

	module.Close(w.runtime.Context())
	delete(w.modules, skillID)
	delete(w.skills, skillID)

	return nil
}

// ListSkills 列出已加载技能
func (w *WASMRuntime) ListSkills() []*WASMSkill {
	w.mu.RLock()
	defer w.mu.RUnlock()

	skills := make([]*WASMSkill, 0, len(w.skills))
	for _, skill := range w.skills {
		skills = append(skills, skill)
	}
	return skills
}

// GetSkill 获取技能
func (w *WASMRuntime) GetSkill(skillID string) (*WASMSkill, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	skill, exists := w.skills[skillID]
	return skill, exists
}

// GetStats 获取统计信息
func (w *WASMRuntime) GetStats() *RuntimeStats {
	return w.stats
}

// Close 关闭运行时
func (w *WASMRuntime) Close(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 关闭所有模块
	for _, module := range w.modules {
		module.Close(ctx)
	}

	w.modules = make(map[string]api.Module)
	w.skills = make(map[string]*WASMSkill)

	return w.runtime.Close(ctx)
}

// ValidateSkill 验证技能
func (w *WASMRuntime) ValidateSkill(skill *WASMSkill) error {
	if skill.ID == "" {
		return fmt.Errorf("技能ID不能为空")
	}
	if skill.Name == "" {
		return fmt.Errorf("技能名称不能为空")
	}
	if len(skill.Binary) == 0 {
		return fmt.Errorf("技能二进制不能为空")
	}
	if len(skill.Operations) == 0 {
		return fmt.Errorf("技能必须至少有一个操作")
	}
	return nil
}

// CreateSkillFromBinary 从二进制创建技能
func CreateSkillFromBinary(id, name string, binary []byte, operations []string) *WASMSkill {
	return &WASMSkill{
		ID:          id,
		Name:        name,
		Version:     "1.0.0",
		Description: fmt.Sprintf("WASM skill: %s", name),
		Binary:      binary,
		Operations:  operations,
		Exports:     make(map[string]string),
		Config: SkillConfig{
			MemoryPages: 128,
			Timeout:     30000,
			Permissions: []string{},
		},
	}
}