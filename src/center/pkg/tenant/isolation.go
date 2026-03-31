// Package tenant provides tenant isolation mechanisms
package tenant

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// IsolationLevel defines data isolation levels
type IsolationLevel string

const (
	IsolationShared    IsolationLevel = "shared"    // Shared resources with namespace isolation
	IsolationDedicated IsolationLevel = "dedicated" // Dedicated resources per tenant
	IsolationHybrid    IsolationLevel = "hybrid"    // Hybrid isolation model
)

// IsolationConfig holds isolation configuration
type IsolationConfig struct {
	Level            IsolationLevel       `json:"level"`
	DataIsolation    DataIsolationConfig  `json:"data_isolation"`
	NetworkIsolation NetworkIsolationConfig `json:"network_isolation"`
	ResourceIsolation ResourceIsolationConfig `json:"resource_isolation"`
}

// DataIsolationConfig defines data isolation settings
type DataIsolationConfig struct {
	// Storage isolation
	SeparateDatabase bool   `json:"separate_database"`
	DatabasePrefix   string `json:"database_prefix"`
	SchemaPerTenant  bool   `json:"schema_per_tenant"`

	// Encryption
	EncryptAtRest    bool   `json:"encrypt_at_rest"`
	EncryptKeyPerTenant bool `json:"encrypt_key_per_tenant"`

	// Backup
	SeparateBackup   bool   `json:"separate_backup"`
	BackupRetention  int    `json:"backup_retention_days"`
}

// NetworkIsolationConfig defines network isolation settings
type NetworkIsolationConfig struct {
	// Network namespace
	SeparateNamespace bool   `json:"separate_namespace"`
	NamespacePrefix   string `json:"namespace_prefix"`

	// Network policies
	AllowInterTenant  bool   `json:"allow_inter_tenant"`
	AllowedNamespaces []string `json:"allowed_namespaces"`

	// Service mesh
	ServiceMeshEnabled bool  `json:"service_mesh_enabled"`
	MTLSEnabled        bool  `json:"mtls_enabled"`
}

// ResourceIsolationConfig defines resource isolation settings
type ResourceIsolationConfig struct {
	// Container isolation
	ContainerPerTenant bool   `json:"container_per_tenant"`
	PodPerTenant       bool   `json:"pod_per_tenant"`

	// Resource limits
	EnforceLimits      bool   `json:"enforce_limits"`
	CPULimit           string `json:"cpu_limit"`
	MemoryLimit        string `json:"memory_limit"`

	// Priority
	PriorityClass      string `json:"priority_class"`
}

// TenantIsolator provides tenant isolation
type TenantIsolator struct {
	config    IsolationConfig
	manager   *TenantManager

	// Isolation contexts
	contexts  sync.Map // map[string]*IsolationContext

	// Namespace registry
	namespaces sync.Map // map[string]string (tenantID -> namespace)

	mu sync.RWMutex
}

// IsolationContext represents tenant isolation context
type IsolationContext struct {
	TenantID         string            `json:"tenant_id"`
	Namespace        string            `json:"namespace"`
	DatabaseName     string            `json:"database_name"`
	EncryptionKey    []byte            `json:"-"`
	ResourceLimits   ResourceLimits    `json:"resource_limits"`
	NetworkPolicies  []NetworkPolicy   `json:"network_policies"`
	CreatedAt        time.Time         `json:"created_at"`
}

// ResourceLimits defines resource limits
type ResourceLimits struct {
	CPURequest    string `json:"cpu_request"`
	CPULimit      string `json:"cpu_limit"`
	MemoryRequest string `json:"memory_request"`
	MemoryLimit   string `json:"memory_limit"`
	EphemeralStorage string `json:"ephemeral_storage"`
}

// NetworkPolicy defines network access rules
type NetworkPolicy struct {
	Name         string   `json:"name"`
	IngressFrom  []string `json:"ingress_from"`
	EgressTo     []string `json:"egress_to"`
	Ports        []int    `json:"ports"`
}

// NewTenantIsolator creates a new tenant isolator
func NewTenantIsolator(config IsolationConfig, manager *TenantManager) *TenantIsolator {
	return &TenantIsolator{
		config:  config,
		manager: manager,
	}
}

// CreateIsolationContext creates isolation context for a tenant
func (i *TenantIsolator) CreateIsolationContext(tenantID string) (*IsolationContext, error) {
	tenant, err := i.manager.GetTenant(tenantID)
	if err != nil {
		return nil, err
	}

	// Generate namespace
	namespace := i.generateNamespace(tenantID)

	// Generate database name
	databaseName := i.generateDatabaseName(tenantID)

	// Generate encryption key
	encryptionKey := generateEncryptionKey()

	// Create resource limits
	limits := i.generateResourceLimits(&tenant.Quota)

	// Create network policies
	policies := i.generateNetworkPolicies(tenantID, tenant.Settings.AllowedRegions)

	ctx := &IsolationContext{
		TenantID:        tenantID,
		Namespace:       namespace,
		DatabaseName:    databaseName,
		EncryptionKey:   encryptionKey,
		ResourceLimits:  limits,
		NetworkPolicies: policies,
		CreatedAt:       time.Now(),
	}

	i.contexts.Store(tenantID, ctx)
	i.namespaces.Store(tenantID, namespace)

	return ctx, nil
}

// GetIsolationContext retrieves isolation context for a tenant
func (i *TenantIsolator) GetIsolationContext(tenantID string) (*IsolationContext, error) {
	if v, ok := i.contexts.Load(tenantID); ok {
		return v.(*IsolationContext), nil
	}

	// Create if not exists
	return i.CreateIsolationContext(tenantID)
}

// DeleteIsolationContext removes isolation context
func (i *TenantIsolator) DeleteIsolationContext(tenantID string) error {
	ctx, err := i.GetIsolationContext(tenantID)
	if err != nil {
		return err
	}

	// Cleanup resources
	if err := i.cleanupResources(ctx); err != nil {
		return err
	}

	i.contexts.Delete(tenantID)
	i.namespaces.Delete(tenantID)

	return nil
}

// ValidateAccess validates tenant access to resources
func (i *TenantIsolator) ValidateAccess(tenantID, resource, action string) error {
	tenant, err := i.manager.GetTenant(tenantID)
	if err != nil {
		return err
	}

	// Check tenant status
	if tenant.Status != TenantActive {
		return fmt.Errorf("tenant %s is not active", tenantID)
	}

	// Check quota
	switch resource {
	case "agent":
		if action == "create" && tenant.Quota.UsedAgents >= tenant.Quota.MaxAgents {
			return errors.New("agent quota exceeded")
		}
	case "skill":
		if action == "create" {
			// Check skill quota
		}
	case "workflow":
		if action == "create" {
			// Check workflow quota
		}
	}

	return nil
}

// GetNamespace returns namespace for a tenant
func (i *TenantIsolator) GetNamespace(tenantID string) string {
	if v, ok := i.namespaces.Load(tenantID); ok {
		return v.(string)
	}
	return i.generateNamespace(tenantID)
}

// IsolateRequest isolates a request to tenant context
func (i *TenantIsolator) IsolateRequest(ctx context.Context, tenantID string) (context.Context, error) {
	isolationCtx, err := i.GetIsolationContext(tenantID)
	if err != nil {
		return nil, err
	}

	// Add isolation context to request context
	ctx = context.WithValue(ctx, "tenant_id", tenantID)
	ctx = context.WithValue(ctx, "namespace", isolationCtx.Namespace)
	ctx = context.WithValue(ctx, "database", isolationCtx.DatabaseName)

	return ctx, nil
}

// FilterResources filters resources by tenant
func (i *TenantIsolator) FilterResources(tenantID string, resources []map[string]interface{}) []map[string]interface{} {
	var filtered []map[string]interface{}

	for _, res := range resources {
		if resTenantID, ok := res["tenant_id"].(string); ok {
			if resTenantID == tenantID {
				filtered = append(filtered, res)
			}
		}
	}

	return filtered
}

// EncryptData encrypts data with tenant key
func (i *TenantIsolator) EncryptData(tenantID string, data []byte) ([]byte, error) {
	if !i.config.DataIsolation.EncryptAtRest {
		return data, nil
	}

	ctx, err := i.GetIsolationContext(tenantID)
	if err != nil {
		return nil, err
	}

	// Simple XOR encryption for demo
	// In production, use AES-GCM
	encrypted := make([]byte, len(data))
	for j := range data {
		encrypted[j] = data[j] ^ ctx.EncryptionKey[j%len(ctx.EncryptionKey)]
	}

	return encrypted, nil
}

// DecryptData decrypts data with tenant key
func (i *TenantIsolator) DecryptData(tenantID string, encrypted []byte) ([]byte, error) {
	if !i.config.DataIsolation.EncryptAtRest {
		return encrypted, nil
	}

	// XOR is symmetric
	return i.EncryptData(tenantID, encrypted)
}

// Helper functions

func (i *TenantIsolator) generateNamespace(tenantID string) string {
	prefix := i.config.NetworkIsolation.NamespacePrefix
	if prefix == "" {
		prefix = "ofa"
	}
	return fmt.Sprintf("%s-%s", prefix, strings.TrimPrefix(tenantID, "tenant-"))
}

func (i *TenantIsolator) generateDatabaseName(tenantID string) string {
	prefix := i.config.DataIsolation.DatabasePrefix
	if prefix == "" {
		prefix = "ofa"
	}
	return fmt.Sprintf("%s_%s", prefix, strings.ReplaceAll(tenantID, "-", "_"))
}

func (i *TenantIsolator) generateResourceLimits(quota *ResourceQuota) ResourceLimits {
	return ResourceLimits{
		CPURequest:    fmt.Sprintf("%dm", quota.MaxCPU*100),
		CPULimit:      fmt.Sprintf("%dm", quota.MaxCPU*100),
		MemoryRequest: fmt.Sprintf("%dMi", quota.MaxMemoryMB),
		MemoryLimit:   fmt.Sprintf("%dMi", quota.MaxMemoryMB),
	}
}

func (i *TenantIsolator) generateNetworkPolicies(tenantID string, allowedRegions []string) []NetworkPolicy {
	policies := []NetworkPolicy{
		{
			Name:        fmt.Sprintf("%s-default-deny", tenantID),
			IngressFrom: []string{},
			EgressTo:    []string{},
			Ports:       []int{8080, 9090},
		},
		{
			Name:        fmt.Sprintf("%s-allow-center", tenantID),
			IngressFrom: []string{"center"},
			EgressTo:    []string{"center"},
			Ports:       []int{8080, 9090},
		},
	}

	if len(allowedRegions) > 0 {
		for _, region := range allowedRegions {
			policies = append(policies, NetworkPolicy{
				Name:        fmt.Sprintf("%s-allow-region-%s", tenantID, region),
				IngressFrom: []string{region},
				EgressTo:    []string{region},
				Ports:       []int{8080, 9090},
			})
		}
	}

	return policies
}

func (i *TenantIsolator) cleanupResources(ctx *IsolationContext) error {
	// Placeholder for resource cleanup
	// In production, would delete namespaces, databases, etc.
	return nil
}

func generateEncryptionKey() []byte {
	key := make([]byte, 32)
	// In production, use proper key generation
	for j := range key {
		key[j] = byte(j % 256)
	}
	return key
}