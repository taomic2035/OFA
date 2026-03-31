package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// SyncItem represents an item to be synchronized
type SyncItem struct {
	Type      string    `json:"type"`       // agent, task, message
	ID        string    `json:"id"`
	Data      []byte    `json:"data"`
	Version   int64     `json:"version"`    // Version number for conflict resolution
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`     // Source node ID
}

// SyncConfig holds sync configuration
type SyncConfig struct {
	NodeID        string
	SyncInterval  time.Duration
	RetryInterval time.Duration
	MaxRetries    int
	BatchSize     int
}

// DefaultSyncConfig returns default sync configuration
func DefaultSyncConfig() *SyncConfig {
	return &SyncConfig{
		SyncInterval:  30 * time.Second,
		RetryInterval: 5 * time.Second,
		MaxRetries:    3,
		BatchSize:     100,
	}
}

// DataSync manages data synchronization between cluster nodes
type DataSync struct {
	config    *SyncConfig
	discovery *ServiceDiscovery
	client    *http.Client

	// Pending sync items
	pending sync.Map // map[string]*SyncItem

	// Last sync versions per type
	versions sync.Map // map[string]int64

	ctx    context.Context
	cancel context.CancelFunc

	mu sync.RWMutex
}

// NewDataSync creates a new data sync manager
func NewDataSync(config *SyncConfig, discovery *ServiceDiscovery) (*DataSync, error) {
	ctx, cancel := context.WithCancel(context.Background())

	ds := &DataSync{
		config:    config,
		discovery: discovery,
		client:    &http.Client{Timeout: 10 * time.Second},
		ctx:       ctx,
		cancel:    cancel,
	}

	return ds, nil
}

// Start begins the sync process
func (ds *DataSync) Start() {
	go ds.syncLoop()
	go ds.retryLoop()
}

// Stop stops the sync process
func (ds *DataSync) Stop() {
	ds.cancel()
}

// AddSyncItem adds an item to be synchronized
func (ds *DataSync) AddSyncItem(item *SyncItem) {
	item.Source = ds.config.NodeID
	item.Timestamp = time.Now()

	// Get current version
	currentVersion, _ := ds.versions.Load(item.Type)
	item.Version = currentVersion.(int64) + 1

	// Store pending item
	ds.pending.Store(fmt.Sprintf("%s:%s", item.Type, item.ID), item)

	// Update version
	ds.versions.Store(item.Type, item.Version)
}

// RemoveSyncItem removes a synchronized item
func (ds *DataSync) RemoveSyncItem(itemType, itemID string) {
	ds.pending.Delete(fmt.Sprintf("%s:%s", itemType, itemID))
}

// syncLoop performs periodic synchronization
func (ds *DataSync) syncLoop() {
	ticker := time.NewTicker(ds.config.SyncInterval)
	defer ticker.Stop()

	// Initial sync
	ds.performSync()

	for {
		select {
		case <-ds.ctx.Done():
			return
		case <-ticker.C:
			ds.performSync()
		}
	}
}

// retryLoop retries failed sync items
func (ds *DataSync) retryLoop() {
	ticker := time.NewTicker(ds.config.RetryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ds.ctx.Done():
			return
		case <-ticker.C:
			ds.retryFailedItems()
		}
	}
}

// performSync sends pending items to all peers
func (ds *DataSync) performSync() {
	nodes := ds.discovery.GetActiveNodes()

	// Collect pending items
	var items []*SyncItem
	ds.pending.Range(func(key, value interface{}) bool {
		items = append(items, value.(*SyncItem))
		if len(items) >= ds.config.BatchSize {
			return false
		}
		return true
	})

	if len(items) == 0 {
		return
	}

	for _, node := range nodes {
		if node.ID == ds.config.NodeID {
			continue // Skip self
		}

		go func(node *NodeInfo) {
			ds.syncToNode(node, items)
		}(node)
	}
}

// syncToNode sends sync items to a specific node
func (ds *DataSync) syncToNode(node *NodeInfo, items []*SyncItem) {
	url := fmt.Sprintf("http://%s:%d/cluster/sync", node.Address, node.HTTPPort)

	data, err := json.Marshal(items)
	if err != nil {
		log.Printf("Failed to marshal sync items: %v", err)
		return
	}

	req, err := http.NewRequestWithContext(ds.ctx, "POST", url, nil)
	if err != nil {
		return
	}
	req.Body = io.NopCloser(bytes.NewReader(data))
	defer req.Body.Close()

	resp, err := ds.client.Do(req)
	if err != nil {
		log.Printf("Sync failed to %s: %v", node.ID, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		// Mark items as synced
		for _, item := range items {
			ds.RemoveSyncItem(item.Type, item.ID)
		}
	}
}

// retryFailedItems retries items that failed to sync
func (ds *DataSync) retryFailedItems() {
	nodes := ds.discovery.GetActiveNodes()
	if len(nodes) == 0 {
		return
	}

	// Collect failed items
	var failedItems []*SyncItem
	ds.pending.Range(func(key, value interface{}) bool {
		item := value.(*SyncItem)
		if item.Timestamp.Before(time.Now().Add(-5 * time.Minute)) {
			failedItems = append(failedItems, item)
		}
		return true
	})

	if len(failedItems) == 0 {
		return
	}

	log.Printf("Retrying %d failed sync items", len(failedItems))
	ds.performSync()
}

// ReceiveSync handles incoming sync data
func (ds *DataSync) ReceiveSync(items []*SyncItem) error {
	for _, item := range items {
		// Check version for conflict resolution
		currentVersion, ok := ds.versions.Load(item.Type)
		if ok && currentVersion.(int64) >= item.Version {
			continue // Already have newer version
		}

		// Process the item based on type
		switch item.Type {
		case "agent":
			ds.processAgentSync(item)
		case "task":
			ds.processTaskSync(item)
		case "message":
			ds.processMessageSync(item)
		}

		// Update version
		ds.versions.Store(item.Type, item.Version)
	}
	return nil
}

// processAgentSync processes synced agent data
func (ds *DataSync) processAgentSync(item *SyncItem) {
	var node Node
	if err := json.Unmarshal(item.Data, &node); err != nil {
		log.Printf("Failed to unmarshal agent sync: %v", err)
		return
	}
	ds.discovery.RegisterNode(&node)
}

// processTaskSync processes synced task data
func (ds *DataSync) processTaskSync(item *SyncItem) {
	// Task sync logic will be implemented by the task manager
	log.Printf("Received task sync: %s", item.ID)
}

// processMessageSync processes synced message data
func (ds *DataSync) processMessageSync(item *SyncItem) {
	// Message sync logic will be implemented by the message router
	log.Printf("Received message sync: %s", item.ID)
}

// GetSyncVersion returns the current sync version for a type
func (ds *DataSync) GetSyncVersion(itemType string) int64 {
	if v, ok := ds.versions.Load(itemType); ok {
		return v.(int64)
	}
	return 0
}

// GetPendingCount returns the number of pending sync items
func (ds *DataSync) GetPendingCount() int {
	count := 0
	ds.pending.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}