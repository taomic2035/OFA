// Package desktop provides a cross-platform desktop agent implementation
// for Windows, macOS, and Linux.
package desktop

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"
)

// AgentConfig holds desktop agent configuration
type AgentConfig struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	CenterURL    string            `json:"center_url"`
	GRPCPort     int               `json:"grpc_port"`
	Capabilities []Capability      `json:"capabilities"`
	AutoStart    bool              `json:"auto_start"`
	MinimizeMode MinimizeMode      `json:"minimize_mode"`
	LogLevel     string            `json:"log_level"`
	LogFile      string            `json:"log_file"`
	DataDir      string            `json:"data_dir"`
	Metadata     map[string]string `json:"metadata"`
}

// Capability represents an agent capability
type Capability struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Operations  []string `json:"operations"`
}

// MinimizeMode defines minimize behavior
type MinimizeMode string

const (
	MinimizeHide      MinimizeMode = "hide"       // Hide to tray
	MinimizeMinimize  MinimizeMode = "minimize"   // Minimize to taskbar
	MinimizeNone      MinimizeMode = "none"       // No minimize support
)

// AgentState represents agent state
type AgentState string

const (
	StateDisconnected AgentState = "disconnected"
	StateConnecting   AgentState = "connecting"
	StateConnected    AgentState = "connected"
	StateBusy         AgentState = "busy"
	StateError        AgentState = "error"
)

// DefaultAgentConfig returns default configuration
func DefaultAgentConfig() *AgentConfig {
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, ".ofa", "agent")

	return &AgentConfig{
		Type:         "desktop",
		AutoStart:    true,
		MinimizeMode: MinimizeHide,
		LogLevel:     "info",
		DataDir:      dataDir,
		GRPCPort:     9090,
		Metadata:     make(map[string]string),
	}
}

// DesktopAgent represents a desktop agent
type DesktopAgent struct {
	config       *AgentConfig
	state        AgentState
	skills       map[string]Skill
	skillManager *SkillManager

	// Connection
	connector Connector

	// Task handling
	taskQueue   chan Task
	taskResults chan TaskResult

	// State management
	pendingTasks sync.Map
	lastActivity time.Time

	// System integration
	trayIcon   TrayIcon
	fileWatcher FileWatcher

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// Skill represents an executable skill
type Skill interface {
	ID() string
	Name() string
	Execute(operation string, params map[string]interface{}) (interface{}, error)
	Operations() []string
}

// Connector handles connection to Center
type Connector interface {
	Connect(ctx context.Context, url string) error
	Disconnect() error
	Register(capabilities []Capability) error
	SendHeartbeat() error
	ReceiveTasks() <-chan Task
	SendResult(result TaskResult) error
	IsConnected() bool
}

// TrayIcon handles system tray integration
type TrayIcon interface {
	Initialize() error
	SetTooltip(text string)
	SetIcon(iconPath string)
	ShowNotification(title, message string)
	SetMenu(items []MenuItem)
	OnClick(handler func())
	OnQuit(handler func())
	Run() error
	Stop()
}

// MenuItem represents a tray menu item
type MenuItem struct {
	Label    string
	Enabled  bool
	Handler  func()
	Children []MenuItem
}

// FileWatcher watches file system changes
type FileWatcher interface {
	AddPath(path string) error
	RemovePath(path string) error
	Events() <-chan FileEvent
	Start() error
	Stop()
}

// FileEvent represents a file system event
type FileEvent struct {
	Path      string
	Operation string // create, modify, delete
	Timestamp time.Time
}

// Task represents a task to execute
type Task struct {
	ID          string                 `json:"id"`
	SkillID     string                 `json:"skill_id"`
	Operation   string                 `json:"operation"`
	Params      map[string]interface{} `json:"params"`
	Priority    int                    `json:"priority"`
	Timeout     int                    `json:"timeout"`
	SubmittedAt time.Time              `json:"submitted_at"`
}

// TaskResult represents a task execution result
type TaskResult struct {
	TaskID    string      `json:"task_id"`
	Success   bool        `json:"success"`
	Data      interface{} `json:"data"`
	Error     string      `json:"error,omitempty"`
	Duration  int64       `json:"duration_ms"`
	Completed time.Time   `json:"completed"`
}

// NewDesktopAgent creates a new desktop agent
func NewDesktopAgent(config *AgentConfig) (*DesktopAgent, error) {
	ctx, cancel := context.WithCancel(context.Background())

	agent := &DesktopAgent{
		config:      config,
		state:       StateDisconnected,
		skills:      make(map[string]Skill),
		taskQueue:   make(chan Task, 100),
		taskResults: make(chan TaskResult, 100),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Initialize skill manager
	agent.skillManager = NewSkillManager(config.DataDir)

	// Load built-in skills
	agent.loadBuiltInSkills()

	// Load user skills
	agent.loadUserSkills()

	return agent, nil
}

// Start starts the desktop agent
func (a *DesktopAgent) Start() error {
	log.Println("Starting OFA Desktop Agent...")

	// Ensure data directory exists
	if err := os.MkdirAll(a.config.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	// Initialize system tray
	if a.config.MinimizeMode == MinimizeHide {
		if err := a.initTrayIcon(); err != nil {
			log.Printf("Failed to initialize tray icon: %v", err)
		}
	}

	// Start task processor
	go a.taskProcessor()

	// Start heartbeat
	go a.heartbeatLoop()

	// Start auto-reconnect
	go a.reconnectLoop()

	// Connect to Center
	if err := a.connect(); err != nil {
		log.Printf("Initial connection failed: %v", err)
	}

	// Handle shutdown signals
	go a.handleSignals()

	return nil
}

// Stop stops the desktop agent
func (a *DesktopAgent) Stop() error {
	log.Println("Stopping OFA Desktop Agent...")

	a.cancel()

	if a.connector != nil {
		a.connector.Disconnect()
	}

	if a.trayIcon != nil {
		a.trayIcon.Stop()
	}

	if a.fileWatcher != nil {
		a.fileWatcher.Stop()
	}

	return nil
}

// connect establishes connection to Center
func (a *DesktopAgent) connect() error {
	a.setState(StateConnecting)

	if a.connector == nil {
		a.connector = NewGRPCConnector(a.config.GRPCPort)
	}

	err := a.connector.Connect(a.ctx, a.config.CenterURL)
	if err != nil {
		a.setState(StateError)
		return err
	}

	// Register capabilities
	capabilities := a.getCapabilities()
	if err := a.connector.Register(capabilities); err != nil {
		a.setState(StateError)
		return err
	}

	a.setState(StateConnected)
	log.Printf("Connected to Center: %s", a.config.CenterURL)

	// Start receiving tasks
	go a.receiveTasks()

	return nil
}

// receiveTasks receives tasks from Center
func (a *DesktopAgent) receiveTasks() {
	if a.connector == nil {
		return
	}

	taskChan := a.connector.ReceiveTasks()
	for {
		select {
		case <-a.ctx.Done():
			return
		case task, ok := <-taskChan:
			if !ok {
				return
			}
			a.taskQueue <- task
		}
	}
}

// taskProcessor processes incoming tasks
func (a *DesktopAgent) taskProcessor() {
	for {
		select {
		case <-a.ctx.Done():
			return
		case task := <-a.taskQueue:
			go a.executeTask(task)
		case result := <-a.taskResults:
			go a.sendResult(result)
		}
	}
}

// executeTask executes a single task
func (a *DesktopAgent) executeTask(task Task) {
	a.setState(StateBusy)
	a.lastActivity = time.Now()

	startTime := time.Now()
	result := TaskResult{
		TaskID:    task.ID,
		Completed: time.Now(),
	}

	// Find skill
	skill, ok := a.skills[task.SkillID]
	if !ok {
		result.Success = false
		result.Error = fmt.Sprintf("Skill not found: %s", task.SkillID)
		result.Duration = time.Since(startTime).Milliseconds()
		a.taskResults <- result
		a.setState(StateConnected)
		return
	}

	// Execute skill
	data, err := skill.Execute(task.Operation, task.Params)
	result.Duration = time.Since(startTime).Milliseconds()

	if err != nil {
		result.Success = false
		result.Error = err.Error()
	} else {
		result.Success = true
		result.Data = data
	}

	a.taskResults <- result
	a.setState(StateConnected)
}

// sendResult sends task result to Center
func (a *DesktopAgent) sendResult(result TaskResult) {
	if a.connector != nil && a.connector.IsConnected() {
		if err := a.connector.SendResult(result); err != nil {
			log.Printf("Failed to send result: %v", err)
		}
	}
}

// heartbeatLoop sends periodic heartbeats
func (a *DesktopAgent) heartbeatLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			if a.connector != nil && a.connector.IsConnected() {
				a.connector.SendHeartbeat()
			}
		}
	}
}

// reconnectLoop handles automatic reconnection
func (a *DesktopAgent) reconnectLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			if a.connector == nil || !a.connector.IsConnected() {
				if err := a.connect(); err != nil {
					log.Printf("Reconnection failed: %v", err)
				}
			}
		}
	}
}

// handleSignals handles OS signals
func (a *DesktopAgent) handleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		a.Stop()
	case <-a.ctx.Done():
	}
}

// loadBuiltInSkills loads built-in skills
func (a *DesktopAgent) loadBuiltInSkills() {
	// Register built-in skills
	a.RegisterSkill(&SystemInfoSkill{})
	a.RegisterSkill(&FileOperationSkill{})
	a.RegisterSkill(&CommandSkill{dataDir: a.config.DataDir})
	a.RegisterSkill(&EchoSkill{})
}

// loadUserSkills loads user-defined skills
func (a *DesktopAgent) loadUserSkills() {
	skills, err := a.skillManager.LoadSkills()
	if err != nil {
		log.Printf("Failed to load user skills: %v", err)
		return
	}

	for _, skill := range skills {
		a.skills[skill.ID()] = skill
	}
}

// RegisterSkill registers a skill
func (a *DesktopAgent) RegisterSkill(skill Skill) {
	a.skills[skill.ID()] = skill
}

// UnregisterSkill unregisters a skill
func (a *DesktopAgent) UnregisterSkill(skillID string) {
	delete(a.skills, skillID)
}

// getCapabilities returns agent capabilities
func (a *DesktopAgent) getCapabilities() []Capability {
	capabilities := make([]Capability, 0, len(a.skills))

	for _, skill := range a.skills {
		capabilities = append(capabilities, Capability{
			ID:          skill.ID(),
			Name:        skill.Name(),
			Version:     "1.0.0",
			Operations:  skill.Operations(),
		})
	}

	return capabilities
}

// initTrayIcon initializes system tray
func (a *DesktopAgent) initTrayIcon() error {
	tray, err := NewPlatformTrayIcon()
	if err != nil {
		return err
	}

	a.trayIcon = tray

	tray.SetTooltip("OFA Desktop Agent")
	tray.SetMenu([]MenuItem{
		{Label: "Status: Disconnected", Enabled: false},
		{Label: "Open Dashboard", Handler: func() {
			a.openDashboard()
		}},
		{Label: "Settings", Handler: func() {
			a.openSettings()
		}},
		{Label: "Quit", Handler: func() {
			a.Stop()
		}},
	})

	go func() {
		if err := tray.Run(); err != nil {
			log.Printf("Tray icon error: %v", err)
		}
	}()

	return nil
}

// openDashboard opens the web dashboard
func (a *DesktopAgent) openDashboard() {
	// Open browser with dashboard URL
	dashboardURL := fmt.Sprintf("http://localhost:%d/dashboard", a.config.GRPCPort)
	OpenURL(dashboardURL)
}

// openSettings opens settings UI
func (a *DesktopAgent) openSettings() {
	// Open settings window or file
	settingsPath := filepath.Join(a.config.DataDir, "config.json")
	OpenFile(settingsPath)
}

// setState sets agent state
func (a *DesktopAgent) setState(state AgentState) {
	a.mu.Lock()
	a.state = state
	a.mu.Unlock()

	// Update tray icon
	if a.trayIcon != nil {
		a.updateTrayState()
	}
}

// GetState returns current agent state
func (a *DesktopAgent) GetState() AgentState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.state
}

// updateTrayState updates tray icon state
func (a *DesktopAgent) updateTrayState() {
	if a.trayIcon == nil {
		return
	}

	stateText := string(a.state)
	a.trayIcon.SetTooltip(fmt.Sprintf("OFA Desktop Agent - %s", stateText))

	menuItems := []MenuItem{
		{Label: fmt.Sprintf("Status: %s", stateText), Enabled: false},
		{Label: "Open Dashboard", Handler: func() { a.openDashboard() }},
		{Label: "Settings", Handler: func() { a.openSettings() }},
	}

	a.trayIcon.SetMenu(menuItems)
}

// GetStats returns agent statistics
func (a *DesktopAgent) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"id":            a.config.ID,
		"name":          a.config.Name,
		"state":         a.GetState(),
		"connected":     a.connector != nil && a.connector.IsConnected(),
		"skills":        len(a.skills),
		"last_activity": a.lastActivity,
		"platform":      runtime.GOOS,
		"arch":          runtime.GOARCH,
	}
}

// SaveConfig saves configuration to file
func (a *DesktopAgent) SaveConfig() error {
	configPath := filepath.Join(a.config.DataDir, "config.json")
	data, err := json.MarshalIndent(a.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

// LoadConfig loads configuration from file
func LoadConfig(path string) (*AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := DefaultAgentConfig()
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}