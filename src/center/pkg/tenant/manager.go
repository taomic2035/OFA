// Package tenant provides multi-tenant support for enterprise deployments
package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TenantStatus defines tenant status
type TenantStatus string

const (
	TenantActive   TenantStatus = "active"   // Active tenant
	TenantSuspended TenantStatus = "suspended" // Suspended tenant
	TenantDeleted   TenantStatus = "deleted"   // Deleted tenant
)

// TenantPlan defines tenant pricing plans
type TenantPlan string

const (
	PlanFree     TenantPlan = "free"     // Free tier
	PlanBasic    TenantPlan = "basic"    // Basic tier
	PlanPro      TenantPlan = "pro"      // Professional tier
	PlanEnterprise TenantPlan = "enterprise" // Enterprise tier
)

// Tenant represents a tenant in the system
type Tenant struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	DisplayName string            `json:"display_name"`
	Status      TenantStatus      `json:"status"`
	Plan        TenantPlan        `json:"plan"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	ExpiresAt   time.Time         `json:"expires_at,omitempty"`
	Metadata    map[string]string `json:"metadata"`
	Settings    TenantSettings    `json:"settings"`
	Quota       ResourceQuota     `json:"quota"`
	OwnerID     string            `json:"owner_id"`
	Admins      []string          `json:"admins"`
}

// TenantSettings holds tenant-specific settings
type TenantSettings struct {
	MaxAgents         int    `json:"max_agents"`
	MaxTasksPerDay    int    `json:"max_tasks_per_day"`
	MaxStorageMB      int    `json:"max_storage_mb"`
	RetentionDays     int    `json:"retention_days"`
	EnableCustomSkills bool   `json:"enable_custom_skills"`
	EnableAI          bool   `json:"enable_ai"`
	EnableFederated   bool   `json:"enable_federated"`
	AllowedRegions    []string `json:"allowed_regions"`
	LogLevel          string `json:"log_level"`
}

// ResourceQuota defines resource limits for a tenant
type ResourceQuota struct {
	// Compute resources
	MaxCPU          int     `json:"max_cpu"`           // CPU cores
	MaxMemoryMB     int     `json:"max_memory_mb"`     // Memory in MB
	MaxGPU          int     `json:"max_gpu"`           // GPU count
	MaxConcurrent   int     `json:"max_concurrent"`    // Concurrent tasks

	// Storage resources
	MaxStorageMB    int     `json:"max_storage_mb"`    // Storage quota
	MaxModels       int     `json:"max_models"`        // Max AI models
	MaxDatasets     int     `json:"max_datasets"`      // Max datasets

	// Network resources
	MaxBandwidthMBps int    `json:"max_bandwidth_mbps"` // Bandwidth limit
	MaxAPICallsDay  int     `json:"max_api_calls_day"` // Daily API call limit

	// Feature limits
	MaxAgents       int     `json:"max_agents"`
	MaxSkills       int     `json:"max_skills"`
	MaxWorkflows    int     `json:"max_workflows"`
	MaxABTests      int     `json:"max_ab_tests"`

	// Current usage
	UsedCPU         int     `json:"used_cpu"`
	UsedMemoryMB    int     `json:"used_memory_mb"`
	UsedGPU         int     `json:"used_gpu"`
	UsedStorageMB   int     `json:"used_storage_mb"`
	UsedAgents      int     `json:"used_agents"`
	UsedAPICalls    int     `json:"used_api_calls"`
}

// TenantUsage tracks resource usage
type TenantUsage struct {
	TenantID       string    `json:"tenant_id"`
	PeriodStart    time.Time `json:"period_start"`
	PeriodEnd      time.Time `json:"period_end"`
	CPUSeconds     int64     `json:"cpu_seconds"`
	MemoryMBHours  int64     `json:"memory_mb_hours"`
	GPUSeconds     int64     `json:"gpu_seconds"`
	StorageMB      int64     `json:"storage_mb"`
	APICalls       int64     `json:"api_calls"`
	TasksExecuted  int64     `json:"tasks_executed"`
	DataProcessedMB int64    `json:"data_processed_mb"`
	BillingAmount  float64   `json:"billing_amount"`
}

// TenantManager manages multi-tenant operations
type TenantManager struct {
	storagePath string

	// Tenant registry
	tenants    sync.Map // map[string]*Tenant
	tenantList []*Tenant

	// Usage tracking
	usage      sync.Map // map[string]*TenantUsage (tenantID -> usage)
	dailyUsage sync.Map // map[string]map[string]int64 (tenantID -> date -> usage)

	// Quota enforcement
	quotaEnforcers sync.Map // map[string]*QuotaEnforcer

	// Billing
	billings sync.Map // map[string]*BillingRecord

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// QuotaEnforcer enforces resource quotas
type QuotaEnforcer struct {
	tenantID string
	quota    *ResourceQuota
	manager  *TenantManager

	// Thresholds for alerts
	warningThreshold float64 // 80%
	criticalThreshold float64 // 95%

	mu sync.RWMutex
}

// BillingRecord represents a billing record
type BillingRecord struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenant_id"`
	PeriodStart   time.Time `json:"period_start"`
	PeriodEnd     time.Time `json:"period_end"`
	Plan          TenantPlan `json:"plan"`
	BaseAmount    float64   `json:"base_amount"`
	UsageAmount   float64   `json:"usage_amount"`
	TotalAmount   float64   `json:"total_amount"`
	Paid          bool      `json:"paid"`
	PaidAt        time.Time `json:"paid_at,omitempty"`
	PaymentMethod string    `json:"payment_method,omitempty"`
	InvoiceURL    string    `json:"invoice_url,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// NewTenantManager creates a new tenant manager
func NewTenantManager(storagePath string) (*TenantManager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	os.MkdirAll(storagePath, 0755)

	manager := &TenantManager{
		storagePath: storagePath,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Load existing tenants
	if err := manager.load(); err != nil {
		log.Printf("Warning: failed to load tenants: %v", err)
	}

	// Start usage collection
	go manager.usageCollector()

	return manager, nil
}

// CreateTenant creates a new tenant
func (m *TenantManager) CreateTenant(name, displayName, ownerID string, plan TenantPlan) (*Tenant, error) {
	if name == "" {
		return nil, errors.New("tenant name required")
	}

	// Check if tenant already exists
	if _, exists := m.tenants.Load(name); exists {
		return nil, fmt.Errorf("tenant already exists: %s", name)
	}

	// Generate tenant ID
	tenantID := generateTenantID(name)

	// Get default quota for plan
	quota := getDefaultQuota(plan)
	settings := getDefaultSettings(plan)

	tenant := &Tenant{
		ID:          tenantID,
		Name:        name,
		DisplayName: displayName,
		Status:      TenantActive,
		Plan:        plan,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    make(map[string]string),
		Settings:    settings,
		Quota:       quota,
		OwnerID:     ownerID,
		Admins:      []string{ownerID},
	}

	// Store tenant
	m.tenants.Store(tenantID, tenant)

	m.mu.Lock()
	m.tenantList = append(m.tenantList, tenant)
	m.mu.Unlock()

	// Create quota enforcer
	enforcer := &QuotaEnforcer{
		tenantID:        tenantID,
		quota:           &quota,
		manager:         m,
		warningThreshold: 0.8,
		criticalThreshold: 0.95,
	}
	m.quotaEnforcers.Store(tenantID, enforcer)

	// Initialize usage tracking
	usage := &TenantUsage{
		TenantID:    tenantID,
		PeriodStart: time.Now(),
	}
	m.usage.Store(tenantID, usage)

	// Save to disk
	m.save()

	log.Printf("Tenant created: %s (plan: %s)", tenantID, plan)

	return tenant, nil
}

// GetTenant retrieves a tenant by ID
func (m *TenantManager) GetTenant(tenantID string) (*Tenant, error) {
	if v, ok := m.tenants.Load(tenantID); ok {
		return v.(*Tenant), nil
	}
	return nil, fmt.Errorf("tenant not found: %s", tenantID)
}

// GetTenantByName retrieves a tenant by name
func (m *TenantManager) GetTenantByName(name string) (*Tenant, error) {
	m.tenants.Range(func(key, value interface{}) bool {
		t := value.(*Tenant)
		if t.Name == name {
			return false
		}
		return true
	})
	return nil, fmt.Errorf("tenant not found by name: %s", name)
}

// ListTenants lists all tenants
func (m *TenantManager) ListTenants() []*Tenant {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tenantList
}

// UpdateTenant updates tenant settings
func (m *TenantManager) UpdateTenant(tenantID string, updates map[string]interface{}) (*Tenant, error) {
	tenant, err := m.GetTenant(tenantID)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Apply updates
	for key, value := range updates {
		switch key {
		case "display_name":
			tenant.DisplayName = value.(string)
		case "status":
			tenant.Status = TenantStatus(value.(string))
		case "plan":
			tenant.Plan = TenantPlan(value.(string))
			// Update quota for new plan
			tenant.Quota = getDefaultQuota(tenant.Plan)
			tenant.Settings = getDefaultSettings(tenant.Plan)
		case "metadata":
			if meta, ok := value.(map[string]string); ok {
				tenant.Metadata = meta
			}
		}
	}

	tenant.UpdatedAt = time.Now()
	m.save()

	return tenant, nil
}

// DeleteTenant deletes a tenant (soft delete)
func (m *TenantManager) DeleteTenant(tenantID string) error {
	tenant, err := m.GetTenant(tenantID)
	if err != nil {
		return err
	}

	m.mu.Lock()
	tenant.Status = TenantDeleted
	tenant.UpdatedAt = time.Now()
	m.mu.Unlock()

	// Remove from list
	m.mu.Lock()
	for i, t := range m.tenantList {
		if t.ID == tenantID {
			m.tenantList = append(m.tenantList[:i], m.tenantList[i+1:]...)
			break
		}
	}
	m.mu.Unlock()

	m.save()

	log.Printf("Tenant deleted: %s", tenantID)

	return nil
}

// SuspendTenant suspends a tenant
func (m *TenantManager) SuspendTenant(tenantID string, reason string) error {
	tenant, err := m.GetTenant(tenantID)
	if err != nil {
		return err
	}

	m.mu.Lock()
	tenant.Status = TenantSuspended
	tenant.Metadata["suspension_reason"] = reason
	tenant.Metadata["suspended_at"] = time.Now().Format(time.RFC3339)
	tenant.UpdatedAt = time.Now()
	m.mu.Unlock()

	m.save()

	log.Printf("Tenant suspended: %s (reason: %s)", tenantID, reason)

	return nil
}

// ActivateTenant reactivates a suspended tenant
func (m *TenantManager) ActivateTenant(tenantID string) error {
	tenant, err := m.GetTenant(tenantID)
	if err != nil {
		return err
	}

	m.mu.Lock()
	tenant.Status = TenantActive
	delete(tenant.Metadata, "suspension_reason")
	delete(tenant.Metadata, "suspended_at")
	tenant.UpdatedAt = time.Now()
	m.mu.Unlock()

	m.save()

	log.Printf("Tenant activated: %s", tenantID)

	return nil
}

// CheckQuota checks if a resource request is within quota
func (m *TenantManager) CheckQuota(tenantID string, resource string, amount int) error {
	enforcer, ok := m.quotaEnforcers.Load(tenantID)
	if !ok {
		return fmt.Errorf("tenant quota enforcer not found: %s", tenantID)
	}

	return enforcer.(*QuotaEnforcer).Check(resource, amount)
}

// RecordUsage records resource usage for a tenant
func (m *TenantManager) RecordUsage(tenantID string, resource string, amount int64) error {
	tenant, err := m.GetTenant(tenantID)
	if err != nil {
		return err
	}

	// Update usage record
	usage, ok := m.usage.Load(tenantID)
	if !ok {
		usage = &TenantUsage{
			TenantID:    tenantID,
			PeriodStart: time.Now(),
		}
		m.usage.Store(tenantID, usage)
	}

	u := usage.(*TenantUsage)

	m.mu.Lock()
	switch resource {
	case "cpu_seconds":
		u.CPUSeconds += amount
		tenant.Quota.UsedCPU += int(amount / 60) // Convert to minutes
	case "memory_mb_hours":
		u.MemoryMBHours += amount
		tenant.Quota.UsedMemoryMB += int(amount)
	case "gpu_seconds":
		u.GPUSeconds += amount
		tenant.Quota.UsedGPU += int(amount / 60)
	case "storage_mb":
		u.StorageMB += amount
		tenant.Quota.UsedStorageMB += int(amount)
	case "api_calls":
		u.APICalls += amount
		tenant.Quota.UsedAPICalls += int(amount)
	case "tasks":
		u.TasksExecuted += amount
	}
	m.mu.Unlock()

	// Record daily usage
	dateKey := time.Now().Format("2006-01-02")
	daily, ok := m.dailyUsage.Load(tenantID)
	if !ok {
		daily = make(map[string]int64)
		m.dailyUsage.Store(tenantID, daily)
	}
	dailyMap := daily.(map[string]int64)
	dailyMap[dateKey+"_"+resource] += amount

	// Check quota thresholds
	if enforcer, ok := m.quotaEnforcers.Load(tenantID); ok {
		enforcer.(*QuotaEnforcer).CheckThresholds(resource)
	}

	return nil
}

// GetUsage retrieves usage statistics for a tenant
func (m *TenantManager) GetUsage(tenantID string, periodStart, periodEnd time.Time) (*TenantUsage, error) {
	usage, ok := m.usage.Load(tenantID)
	if !ok {
		return nil, fmt.Errorf("usage not found for tenant: %s", tenantID)
	}

	return usage.(*TenantUsage), nil
}

// GenerateBilling generates billing for a period
func (m *TenantManager) GenerateBilling(tenantID string, periodStart, periodEnd time.Time) (*BillingRecord, error) {
	tenant, err := m.GetTenant(tenantID)
	if err != nil {
		return nil, err
	}

	usage, err := m.GetUsage(tenantID, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}

	// Calculate base amount based on plan
	baseAmount := getPlanPrice(tenant.Plan)

	// Calculate usage amount
	usageAmount := calculateUsageAmount(tenant.Plan, usage)

	billing := &BillingRecord{
		ID:          generateBillingID(tenantID),
		TenantID:    tenantID,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Plan:        tenant.Plan,
		BaseAmount:  baseAmount,
		UsageAmount: usageAmount,
		TotalAmount: baseAmount + usageAmount,
		Paid:        false,
		CreatedAt:   time.Now(),
	}

	m.billings.Store(billing.ID, billing)
	m.save()

	return billing, nil
}

// QuotaEnforcer methods

// Check checks if a resource request is within quota
func (e *QuotaEnforcer) Check(resource string, amount int) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	switch resource {
	case "agents":
		if e.quota.UsedAgents+amount > e.quota.MaxAgents {
			return fmt.Errorf("agent quota exceeded: %d/%d", e.quota.UsedAgents+amount, e.quota.MaxAgents)
		}
	case "cpu":
		if e.quota.UsedCPU+amount > e.quota.MaxCPU {
			return fmt.Errorf("CPU quota exceeded: %d/%d", e.quota.UsedCPU+amount, e.quota.MaxCPU)
		}
	case "memory":
		if e.quota.UsedMemoryMB+amount > e.quota.MaxMemoryMB {
			return fmt.Errorf("memory quota exceeded: %d/%d MB", e.quota.UsedMemoryMB+amount, e.quota.MaxMemoryMB)
		}
	case "gpu":
		if e.quota.UsedGPU+amount > e.quota.MaxGPU {
			return fmt.Errorf("GPU quota exceeded: %d/%d", e.quota.UsedGPU+amount, e.quota.MaxGPU)
		}
	case "storage":
		if e.quota.UsedStorageMB+amount > e.quota.MaxStorageMB {
			return fmt.Errorf("storage quota exceeded: %d/%d MB", e.quota.UsedStorageMB+amount, e.quota.MaxStorageMB)
		}
	case "api_calls":
		if e.quota.UsedAPICalls+amount > e.quota.MaxAPICallsDay {
			return fmt.Errorf("API call quota exceeded: %d/%d", e.quota.UsedAPICalls+amount, e.quota.MaxAPICallsDay)
		}
	case "concurrent":
		if amount > e.quota.MaxConcurrent {
			return fmt.Errorf("concurrent task limit exceeded: %d/%d", amount, e.quota.MaxConcurrent)
		}
	}

	return nil
}

// CheckThresholds checks if usage exceeds warning/critical thresholds
func (e *QuotaEnforcer) CheckThresholds(resource string) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var used, max int
	switch resource {
	case "agents":
		used = e.quota.UsedAgents
		max = e.quota.MaxAgents
	case "cpu":
		used = e.quota.UsedCPU
		max = e.quota.MaxCPU
	case "memory":
		used = e.quota.UsedMemoryMB
		max = e.quota.MaxMemoryMB
	case "storage":
		used = e.quota.UsedStorageMB
		max = e.quota.MaxStorageMB
	case "api_calls":
		used = e.quota.UsedAPICalls
		max = e.quota.MaxAPICallsDay
	default:
		return
	}

	if max == 0 {
		return
	}

	usagePercent := float64(used) / float64(max)

	if usagePercent >= e.criticalThreshold {
		log.Printf("[CRITICAL] Tenant %s quota critical: %s at %.1f%%", e.tenantID, resource, usagePercent*100)
	} else if usagePercent >= e.warningThreshold {
		log.Printf("[WARNING] Tenant %s quota warning: %s at %.1f%%", e.tenantID, resource, usagePercent*100)
	}
}

// Helper functions

func generateTenantID(name string) string {
	return fmt.Sprintf("tenant-%s-%d", name, time.Now().UnixNano())
}

func generateBillingID(tenantID string) string {
	return fmt.Sprintf("bill-%s-%d", tenantID, time.Now().UnixNano())
}

func getDefaultQuota(plan TenantPlan) ResourceQuota {
	switch plan {
	case PlanFree:
		return ResourceQuota{
			MaxAgents:        2,
			MaxCPU:           1,
			MaxMemoryMB:      512,
			MaxGPU:           0,
			MaxStorageMB:     100,
			MaxConcurrent:    2,
			MaxAPICallsDay:   100,
			MaxSkills:        5,
			MaxWorkflows:     3,
			MaxABTests:       0,
		}
	case PlanBasic:
		return ResourceQuota{
			MaxAgents:        10,
			MaxCPU:           2,
			MaxMemoryMB:      2048,
			MaxGPU:           0,
			MaxStorageMB:     1024,
			MaxConcurrent:    10,
			MaxAPICallsDay:   1000,
			MaxSkills:        20,
			MaxWorkflows:     10,
			MaxABTests:       5,
		}
	case PlanPro:
		return ResourceQuota{
			MaxAgents:        50,
			MaxCPU:           8,
			MaxMemoryMB:      8192,
			MaxGPU:           1,
			MaxStorageMB:     10240,
			MaxConcurrent:    50,
			MaxAPICallsDay:   10000,
			MaxSkills:        100,
			MaxWorkflows:     50,
			MaxABTests:       20,
		}
	case PlanEnterprise:
		return ResourceQuota{
			MaxAgents:        math.MaxInt32,
			MaxCPU:           math.MaxInt32,
			MaxMemoryMB:      math.MaxInt32,
			MaxGPU:           math.MaxInt32,
			MaxStorageMB:     math.MaxInt32,
			MaxConcurrent:    math.MaxInt32,
			MaxAPICallsDay:   math.MaxInt32,
			MaxSkills:        math.MaxInt32,
			MaxWorkflows:     math.MaxInt32,
			MaxABTests:       math.MaxInt32,
		}
	default:
		return getDefaultQuota(PlanFree)
	}
}

func getDefaultSettings(plan TenantPlan) TenantSettings {
	switch plan {
	case PlanFree:
		return TenantSettings{
			MaxAgents:        2,
			MaxTasksPerDay:   100,
			MaxStorageMB:     100,
			RetentionDays:    7,
			EnableCustomSkills: false,
			EnableAI:         false,
			EnableFederated:  false,
			LogLevel:         "info",
		}
	case PlanBasic:
		return TenantSettings{
			MaxAgents:        10,
			MaxTasksPerDay:   1000,
			MaxStorageMB:     1024,
			RetentionDays:    30,
			EnableCustomSkills: true,
			EnableAI:         false,
			EnableFederated:  false,
			LogLevel:         "info",
		}
	case PlanPro:
		return TenantSettings{
			MaxAgents:        50,
			MaxTasksPerDay:   10000,
			MaxStorageMB:     10240,
			RetentionDays:    90,
			EnableCustomSkills: true,
			EnableAI:         true,
			EnableFederated:  false,
			LogLevel:         "debug",
		}
	case PlanEnterprise:
		return TenantSettings{
			MaxAgents:        math.MaxInt32,
			MaxTasksPerDay:   math.MaxInt32,
			MaxStorageMB:     math.MaxInt32,
			RetentionDays:    365,
			EnableCustomSkills: true,
			EnableAI:         true,
			EnableFederated:  true,
			LogLevel:         "debug",
		}
	default:
		return getDefaultSettings(PlanFree)
	}
}

func getPlanPrice(plan TenantPlan) float64 {
	switch plan {
	case PlanFree:
		return 0
	case PlanBasic:
		return 29.99
	case PlanPro:
		return 99.99
	case PlanEnterprise:
		return 499.99
	default:
		return 0
	}
}

func calculateUsageAmount(plan TenantPlan, usage *TenantUsage) float64 {
	// Simplified usage-based billing
	cpuRate := 0.001 // $ per CPU second
	memoryRate := 0.0001 // $ per MB hour
	gpuRate := 0.01 // $ per GPU second
	apiRate := 0.0001 // $ per API call

	cpuCost := float64(usage.CPUSeconds) * cpuRate
	memoryCost := float64(usage.MemoryMBHours) * memoryRate
	gpuCost := float64(usage.GPUSeconds) * gpuRate
	apiCost := float64(usage.APICalls) * apiRate

	return cpuCost + memoryCost + gpuCost + apiCost
}

// usageCollector collects usage metrics periodically
func (m *TenantManager) usageCollector() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.collectUsage()
		}
	}
}

// collectUsage collects usage for billing
func (m *TenantManager) collectUsage() {
	m.tenants.Range(func(key, value interface{}) bool {
		tenant := value.(*Tenant)
		if tenant.Status != TenantActive {
			return true
		}

		// Reset daily counters at midnight
		now := time.Now()
		if now.Hour() == 0 {
			m.resetDailyUsage(tenant.ID)
		}

		return true
	})
}

// resetDailyUsage resets daily usage counters
func (m *TenantManager) resetDailyUsage(tenantID string) {
	tenant, err := m.GetTenant(tenantID)
	if err != nil {
		return
	}

	m.mu.Lock()
	tenant.Quota.UsedAPICalls = 0
	m.mu.Unlock()
}

// load loads tenants from disk
func (m *TenantManager) load() error {
	if m.storagePath == "" {
		return nil
	}

	data, err := os.ReadFile(filepath.Join(m.storagePath, "tenants.json"))
	if err != nil {
		return err
	}

	var tenants []*Tenant
	if err := json.Unmarshal(data, &tenants); err != nil {
		return err
	}

	for _, t := range tenants {
		m.tenants.Store(t.ID, t)
		m.tenantList = append(m.tenantList, t)

		// Create quota enforcer
		enforcer := &QuotaEnforcer{
			tenantID:        t.ID,
			quota:           &t.Quota,
			manager:         m,
			warningThreshold: 0.8,
			criticalThreshold: 0.95,
		}
		m.quotaEnforcers.Store(t.ID, enforcer)
	}

	return nil
}

// save saves tenants to disk
func (m *TenantManager) save() error {
	if m.storagePath == "" {
		return nil
	}

	m.mu.RLock()
	tenants := make([]*Tenant, len(m.tenantList))
	copy(tenants, m.tenantList)
	m.mu.RUnlock()

	data, err := json.Marshal(tenants)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(m.storagePath, "tenants.json"), data, 0644)
}

// Close closes the tenant manager
func (m *TenantManager) Close() {
	m.cancel()
	m.save()
}