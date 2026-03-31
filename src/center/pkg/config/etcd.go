// Package config - etcd配置中心
package config

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdConfig etcd配置
type EtcdConfig struct {
	Endpoints        []string      `json:"endpoints"`         // etcd节点地址
	DialTimeout      time.Duration `json:"dial_timeout"`      // 连接超时
	RequestTimeout   time.Duration `json:"request_timeout"`   // 请求超时
	Username         string        `json:"username"`          // 用户名
	Password         string        `json:"password"`          // 密码
	TLS              *TLSConfig    `json:"tls,omitempty"`     // TLS配置
	AutoSyncInterval time.Duration `json:"auto_sync_interval"` // 自动同步间隔
	Prefix           string        `json:"prefix"`            // 键前缀
}

// TLSConfig TLS配置
type TLSConfig struct {
	CertFile   string `json:"cert_file"`
	KeyFile    string `json:"key_file"`
	CAFile     string `json:"ca_file"`
	ServerName string `json:"server_name"`
}

// ConfigItem 配置项
type ConfigItem struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Version   int64     `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	TTL       int64     `json:"ttl,omitempty"` // 租约TTL(秒)，0表示永不过期
	LeaseID   int64     `json:"lease_id,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
}

// ServiceNode 服务节点
type ServiceNode struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Address   string            `json:"address"`
	Port      int               `json:"port"`
	Metadata  map[string]string `json:"metadata"`
	Status    string            `json:"status"` // online, offline, maintenance
	UpdatedAt time.Time         `json:"updated_at"`
	TTL       int64             `json:"ttl"`
}

// WatchEvent 监听事件
type WatchEvent struct {
	Type    string      `json:"type"` // PUT, DELETE
	Key     string      `json:"key"`
	Value   string      `json:"value,omitempty"`
	OldValue string     `json:"old_value,omitempty"`
	NewItem *ConfigItem `json:"new_item,omitempty"`
	OldItem *ConfigItem `json:"old_item,omitempty"`
}

// ConfigStats 配置中心统计
type ConfigStats struct {
	TotalKeys      int64         `json:"total_keys"`
	Watchers       int           `json:"watchers"`
	Services       int           `json:"services"`
	Leases         int           `json:"leases"`
	LastSyncTime   time.Time     `json:"last_sync_time"`
	Operations     int64         `json:"operations"`
	Errors         int64         `json:"errors"`
	AvgLatency     time.Duration `json:"avg_latency"`
	Leader         string        `json:"leader,omitempty"`
	ClusterHealth  string        `json:"cluster_health"` // healthy, degraded, unhealthy
}

// EtcdCenter etcd配置中心
type EtcdCenter struct {
	config    EtcdConfig
	client    *clientv3.Client
	kv        clientv3.KV
	watchers  map[string]context.CancelFunc
	services  map[string]*ServiceNode
	stats     *ConfigStats
	cache     map[string]*ConfigItem
	mu        sync.RWMutex
	running   bool
}

// NewEtcdCenter 创建etcd配置中心
func NewEtcdCenter(config EtcdConfig) *EtcdCenter {
	// 默认配置
	if len(config.Endpoints) == 0 {
		config.Endpoints = []string{"localhost:2379"}
	}
	if config.DialTimeout == 0 {
		config.DialTimeout = 5 * time.Second
	}
	if config.RequestTimeout == 0 {
		config.RequestTimeout = 3 * time.Second
	}
	if config.AutoSyncInterval == 0 {
		config.AutoSyncInterval = 30 * time.Second
	}

	return &EtcdCenter{
		config:   config,
		watchers: make(map[string]context.CancelFunc),
		services: make(map[string]*ServiceNode),
		stats:    &ConfigStats{},
		cache:    make(map[string]*ConfigItem),
	}
}

// Connect 连接etcd
func (ec *EtcdCenter) Connect(ctx context.Context) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:        ec.config.Endpoints,
		DialTimeout:      ec.config.DialTimeout,
		Username:         ec.config.Username,
		Password:         ec.config.Password,
		AutoSyncInterval: ec.config.AutoSyncInterval,
	})
	if err != nil {
		return fmt.Errorf("连接etcd失败: %w", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(ctx, ec.config.DialTimeout)
	defer cancel()

	_, err = cli.Get(ctx, "test-connection")
	if err != nil && err != context.DeadlineExceeded {
		cli.Close()
		return fmt.Errorf("etcd连接测试失败: %w", err)
	}

	ec.client = cli
	ec.kv = cli

	ec.mu.Lock()
	ec.running = true
	ec.mu.Unlock()

	// 启动后台同步
	go ec.syncLoop(ctx)

	return nil
}

// Disconnect 断开连接
func (ec *EtcdCenter) Disconnect() {
	ec.mu.Lock()
	ec.running = false
	// 取消所有监听
	for key, cancel := range ec.watchers {
		cancel()
		delete(ec.watchers, key)
	}
	ec.mu.Unlock()

	if ec.client != nil {
		ec.client.Close()
	}
}

// === 配置管理 ===

// Get 获取配置
func (ec *EtcdCenter) Get(ctx context.Context, key string) (*ConfigItem, error) {
	fullKey := ec.prefixKey(key)

	start := time.Now()
	resp, err := ec.client.Get(ctx, fullKey)
	ec.updateLatency(time.Since(start))

	if err != nil {
		ec.mu.Lock()
		ec.stats.Errors++
		ec.mu.Unlock()
		return nil, fmt.Errorf("获取配置失败: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("配置不存在: %s", key)
	}

	kv := resp.Kvs[0]
	item := &ConfigItem{
		Key:       key,
		Value:     string(kv.Value),
		Version:   kv.Version,
		UpdatedAt: time.Unix(0, kv.ModRevision*int64(time.Millisecond)),
		LeaseID:   int64(kv.Lease),
	}

	ec.mu.Lock()
	ec.cache[key] = item
	ec.stats.Operations++
	ec.mu.Unlock()

	return item, nil
}

// Set 设置配置
func (ec *EtcdCenter) Set(ctx context.Context, key string, value string, ttl int64) error {
	fullKey := ec.prefixKey(key)

	var opts []clientv3.OpOption
	if ttl > 0 {
		// 创建租约
		lease, err := ec.client.Grant(ctx, ttl)
		if err != nil {
			return fmt.Errorf("创建租约失败: %w", err)
		}
		opts = append(opts, clientv3.WithLease(lease.ID))
	}

	start := time.Now()
	_, err := ec.client.Put(ctx, fullKey, value, opts...)
	ec.updateLatency(time.Since(start))

	if err != nil {
		ec.mu.Lock()
		ec.stats.Errors++
		ec.mu.Unlock()
		return fmt.Errorf("设置配置失败: %w", err)
	}

	// 更新缓存
	ec.mu.Lock()
	ec.cache[key] = &ConfigItem{
		Key:       key,
		Value:     value,
		UpdatedAt: time.Now(),
		TTL:       ttl,
	}
	ec.stats.Operations++
	ec.stats.TotalKeys++
	ec.mu.Unlock()

	return nil
}

// Delete 删除配置
func (ec *EtcdCenter) Delete(ctx context.Context, key string) error {
	fullKey := ec.prefixKey(key)

	start := time.Now()
	_, err := ec.client.Delete(ctx, fullKey)
	ec.updateLatency(time.Since(start))

	if err != nil {
		ec.mu.Lock()
		ec.stats.Errors++
		ec.mu.Unlock()
		return fmt.Errorf("删除配置失败: %w", err)
	}

	ec.mu.Lock()
	delete(ec.cache, key)
	ec.stats.Operations++
	ec.stats.TotalKeys--
	ec.mu.Unlock()

	return nil
}

// GetPrefix 获取前缀所有配置
func (ec *EtcdCenter) GetPrefix(ctx context.Context, prefix string) ([]*ConfigItem, error) {
	fullPrefix := ec.prefixKey(prefix)

	resp, err := ec.client.Get(ctx, fullPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("获取配置列表失败: %w", err)
	}

	items := make([]*ConfigItem, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		if ec.config.Prefix != "" {
			key = key[len(ec.config.Prefix)+1:]
		}
		items = append(items, &ConfigItem{
			Key:       key,
			Value:     string(kv.Value),
			Version:   kv.Version,
			UpdatedAt: time.Unix(0, kv.ModRevision*int64(time.Millisecond)),
			LeaseID:   int64(kv.Lease),
		})
	}

	ec.mu.Lock()
	ec.stats.Operations++
	ec.mu.Unlock()

	return items, nil
}

// SetMulti 批量设置
func (ec *EtcdCenter) SetMulti(ctx context.Context, items map[string]string) error {
	if len(items) == 0 {
		return nil
	}

	ops := make([]clientv3.Op, 0, len(items))
	for key, value := range items {
		fullKey := ec.prefixKey(key)
		ops = append(ops, clientv3.OpPut(fullKey, value))
	}

	// 事务执行
	_, err := ec.client.Txn(ctx).Then(ops...).Commit()
	if err != nil {
		return fmt.Errorf("批量设置失败: %w", err)
	}

	// 更新缓存
	ec.mu.Lock()
	for key, value := range items {
		ec.cache[key] = &ConfigItem{
			Key:       key,
			Value:     value,
			UpdatedAt: time.Now(),
		}
	}
	ec.stats.Operations++
	ec.stats.TotalKeys += int64(len(items))
	ec.mu.Unlock()

	return nil
}

// DeleteMulti 批量删除
func (ec *EtcdCenter) DeleteMulti(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	ops := make([]clientv3.Op, 0, len(keys))
	for _, key := range keys {
		fullKey := ec.prefixKey(key)
		ops = append(ops, clientv3.OpDelete(fullKey))
	}

	_, err := ec.client.Txn(ctx).Then(ops...).Commit()
	if err != nil {
		return fmt.Errorf("批量删除失败: %w", err)
	}

	ec.mu.Lock()
	for _, key := range keys {
		delete(ec.cache, key)
	}
	ec.stats.Operations++
	ec.stats.TotalKeys -= int64(len(keys))
	ec.mu.Unlock()

	return nil
}

// === 配置监听 ===

// Watch 监听配置变化
func (ec *EtcdCenter) Watch(ctx context.Context, key string, handler func(event *WatchEvent)) error {
	fullKey := ec.prefixKey(key)

	ec.mu.Lock()
	if _, exists := ec.watchers[key]; exists {
		ec.mu.Unlock()
		return fmt.Errorf("已存在监听: %s", key)
	}

	watchCtx, cancel := context.WithCancel(ctx)
	ec.watchers[key] = cancel
	ec.stats.Watchers++
	ec.mu.Unlock()

	// 获取当前值
	current, _ := ec.Get(ctx, key)

	// 开始监听
	watchChan := ec.client.Watch(watchCtx, fullKey)

	go func() {
		defer func() {
			ec.mu.Lock()
			delete(ec.watchers, key)
			ec.stats.Watchers--
			ec.mu.Unlock()
		}()

		for {
			select {
			case <-watchCtx.Done():
				return
			case resp := <-watchChan:
				if resp.Err() != nil {
					continue
				}

				for _, ev := range resp.Events {
					event := &WatchEvent{
						Type: ev.Type.String(),
						Key:  key,
					}

					if ev.Type == clientv3.EventTypePut {
						event.Value = string(ev.Kv.Value)
						event.NewItem = &ConfigItem{
							Key:     key,
							Value:   string(ev.Kv.Value),
							Version: ev.Kv.Version,
						}
						if current != nil {
							event.OldValue = current.Value
							event.OldItem = current
						}

						// 更新缓存
						ec.mu.Lock()
						ec.cache[key] = event.NewItem
						ec.mu.Unlock()
					} else if ev.Type == clientv3.EventTypeDelete {
						if current != nil {
							event.OldValue = current.Value
							event.OldItem = current
						}
						ec.mu.Lock()
						delete(ec.cache, key)
						ec.mu.Unlock()
					}

					handler(event)
				}
			}
		}
	}()

	return nil
}

// WatchPrefix 监听前缀变化
func (ec *EtcdCenter) WatchPrefix(ctx context.Context, prefix string, handler func(event *WatchEvent)) error {
	fullPrefix := ec.prefixKey(prefix)
	watchKey := "prefix:" + prefix

	ec.mu.Lock()
	if _, exists := ec.watchers[watchKey]; exists {
		ec.mu.Unlock()
		return fmt.Errorf("已存在前缀监听: %s", prefix)
	}

	watchCtx, cancel := context.WithCancel(ctx)
	ec.watchers[watchKey] = cancel
	ec.stats.Watchers++
	ec.mu.Unlock()

	watchChan := ec.client.Watch(watchCtx, fullPrefix, clientv3.WithPrefix())

	go func() {
		defer func() {
			ec.mu.Lock()
			delete(ec.watchers, watchKey)
			ec.stats.Watchers--
			ec.mu.Unlock()
		}()

		for {
			select {
			case <-watchCtx.Done():
				return
			case resp := <-watchChan:
				if resp.Err() != nil {
					continue
				}

				for _, ev := range resp.Events {
					key := string(ev.Kv.Key)
					if ec.config.Prefix != "" {
						key = key[len(ec.config.Prefix)+1:]
					}

					event := &WatchEvent{
						Type: ev.Type.String(),
						Key:  key,
					}

					if ev.Type == clientv3.EventTypePut {
						event.Value = string(ev.Kv.Value)
						event.NewItem = &ConfigItem{
							Key:     key,
							Value:   string(ev.Kv.Value),
							Version: ev.Kv.Version,
						}
					}

					handler(event)
				}
			}
		}
	}()

	return nil
}

// Unwatch 取消监听
func (ec *EtcdCenter) Unwatch(key string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	if cancel, exists := ec.watchers[key]; exists {
		cancel()
		delete(ec.watchers, key)
		ec.stats.Watchers--
	}
}

// === 服务发现 ===

// RegisterService 注册服务
func (ec *EtcdCenter) RegisterService(ctx context.Context, service *ServiceNode) error {
	if service.ID == "" {
		service.ID = fmt.Sprintf("svc-%d", time.Now().UnixNano())
	}
	service.UpdatedAt = time.Now()
	service.Status = "online"

	key := ec.serviceKey(service.Name, service.ID)
	data, err := json.Marshal(service)
	if err != nil {
		return err
	}

	// 使用租约实现心跳
	ttl := service.TTL
	if ttl == 0 {
		ttl = 10 // 默认10秒
	}

	lease, err := ec.client.Grant(ctx, ttl)
	if err != nil {
		return fmt.Errorf("创建服务租约失败: %w", err)
	}

	_, err = ec.client.Put(ctx, key, string(data), clientv3.WithLease(lease.ID))
	if err != nil {
		return fmt.Errorf("注册服务失败: %w", err)
	}

	// 保存租约ID
	service.LeaseID = int64(lease.ID)

	ec.mu.Lock()
	ec.services[service.ID] = service
	ec.stats.Services++
	ec.mu.Unlock()

	// 保持心跳
	go ec.keepAlive(ctx, lease.ID, service.ID)

	return nil
}

// DeregisterService 注销服务
func (ec *EtcdCenter) DeregisterService(ctx context.Context, name, id string) error {
	key := ec.serviceKey(name, id)

	_, err := ec.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("注销服务失败: %w", err)
	}

	ec.mu.Lock()
	delete(ec.services, id)
	ec.stats.Services--
	ec.mu.Unlock()

	return nil
}

// DiscoverService 发现服务
func (ec *EtcdCenter) DiscoverService(ctx context.Context, name string) ([]*ServiceNode, error) {
	prefix := ec.servicePrefix(name)

	resp, err := ec.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("发现服务失败: %w", err)
	}

	services := make([]*ServiceNode, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var service ServiceNode
		if err := json.Unmarshal(kv.Value, &service); err != nil {
			continue
		}
		services = append(services, &service)
	}

	return services, nil
}

// WatchService 监听服务变化
func (ec *EtcdCenter) WatchService(ctx context.Context, name string, handler func(event *WatchEvent, service *ServiceNode)) error {
	prefix := ec.servicePrefix(name)
	watchKey := "service:" + name

	ec.mu.Lock()
	if _, exists := ec.watchers[watchKey]; exists {
		ec.mu.Unlock()
		return fmt.Errorf("已存在服务监听: %s", name)
	}

	watchCtx, cancel := context.WithCancel(ctx)
	ec.watchers[watchKey] = cancel
	ec.stats.Watchers++
	ec.mu.Unlock()

	watchChan := ec.client.Watch(watchCtx, prefix, clientv3.WithPrefix())

	go func() {
		defer func() {
			ec.mu.Lock()
			delete(ec.watchers, watchKey)
			ec.stats.Watchers--
			ec.mu.Unlock()
		}()

		for {
			select {
			case <-watchCtx.Done():
				return
			case resp := <-watchChan:
				if resp.Err() != nil {
					continue
				}

				for _, ev := range resp.Events {
					event := &WatchEvent{
						Type: ev.Type.String(),
					}

					var service ServiceNode
					if len(ev.Kv.Value) > 0 {
						json.Unmarshal(ev.Kv.Value, &service)
					}

					handler(event, &service)
				}
			}
		}
	}()

	return nil
}

// keepAlive 保持心跳
func (ec *EtcdCenter) keepAlive(ctx context.Context, leaseID int64, serviceID string) {
	ch, err := ec.client.KeepAlive(ctx, clientv3.LeaseID(leaseID))
	if err != nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-ch:
			if !ok {
				// 租约过期，重新注册
				ec.mu.Lock()
				if service, exists := ec.services[serviceID]; exists {
					service.Status = "offline"
				}
				ec.mu.Unlock()
				return
			}
		}
	}
}

// === 租约管理 ===

// Grant 创建租约
func (ec *EtcdCenter) Grant(ctx context.Context, ttl int64) (int64, error) {
	resp, err := ec.client.Grant(ctx, ttl)
	if err != nil {
		return 0, err
	}

	ec.mu.Lock()
	ec.stats.Leases++
	ec.mu.Unlock()

	return int64(resp.ID), nil
}

// Revoke 撤销租约
func (ec *EtcdCenter) Revoke(ctx context.Context, leaseID int64) error {
	_, err := ec.client.Revoke(ctx, clientv3.LeaseID(leaseID))
	if err != nil {
		return err
	}

	ec.mu.Lock()
	ec.stats.Leases--
	ec.mu.Unlock()

	return nil
}

// KeepAlive 保持租约
func (ec *EtcdCenter) KeepAlive(ctx context.Context, leaseID int64) error {
	ch, err := ec.client.KeepAlive(ctx, clientv3.LeaseID(leaseID))
	if err != nil {
		return err
	}

	// 消费响应
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-ch:
				if !ok {
					return
				}
			}
		}
	}()

	return nil
}

// === 事务支持 ===

// Txn 事务操作
func (ec *EtcdCenter) Txn(ctx context.Context, compares []string, thenOps []clientv3.Op, elseOps []clientv3.Op) error {
	txn := ec.client.Txn(ctx)

	// 添加比较条件
	for _, cmp := range compares {
		txn.If(clientv3.Compare(clientv3.Value(cmp), "=", ""))
	}

	// 添加操作
	if len(thenOps) > 0 {
		txn.Then(thenOps...)
	}
	if len(elseOps) > 0 {
		txn.Else(elseOps...)
	}

	_, err := txn.Commit()
	return err
}

// CompareAndSwap 比较并交换
func (ec *EtcdCenter) CompareAndSwap(ctx context.Context, key string, oldValue, newValue string) (bool, error) {
	fullKey := ec.prefixKey(key)

	txn := ec.client.Txn(ctx).
		If(clientv3.Compare(clientv3.Value(fullKey), "=", oldValue)).
		Then(clientv3.OpPut(fullKey, newValue))

	resp, err := txn.Commit()
	if err != nil {
		return false, err
	}

	return resp.Succeeded, nil
}

// === 辅助方法 ===

func (ec *EtcdCenter) prefixKey(key string) string {
	if ec.config.Prefix == "" {
		return key
	}
	return ec.config.Prefix + "/" + key
}

func (ec *EtcdCenter) serviceKey(name, id string) string {
	return ec.prefixKey("services/" + name + "/" + id)
}

func (ec *EtcdCenter) servicePrefix(name string) string {
	return ec.prefixKey("services/" + name + "/")
}

func (ec *EtcdCenter) updateLatency(d time.Duration) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	// 计算平均延迟
	ec.stats.Operations++
	if ec.stats.AvgLatency == 0 {
		ec.stats.AvgLatency = d
	} else {
		ec.stats.AvgLatency = (ec.stats.AvgLatency + d) / 2
	}
}

// syncLoop 同步循环
func (ec *EtcdCenter) syncLoop(ctx context.Context) {
	ticker := time.NewTicker(ec.config.AutoSyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ec.syncStats(ctx)
		}
	}
}

// syncStats 同步统计信息
func (ec *EtcdCenter) syncStats(ctx context.Context) {
	// 获取集群状态
	if ec.client != nil {
		resp, err := ec.client.Get(ctx, ec.prefixKey(""), clientv3.WithPrefix(), clientv3.WithCountOnly())
		if err == nil {
			ec.mu.Lock()
			ec.stats.TotalKeys = int64(resp.Count)
			ec.stats.LastSyncTime = time.Now()
			ec.mu.Unlock()
		}
	}
}

// GetStatistics 获取统计信息
func (ec *EtcdCenter) GetStatistics() map[string]interface{} {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	return map[string]interface{}{
		"total_keys":     ec.stats.TotalKeys,
		"watchers":       ec.stats.Watchers,
		"services":       ec.stats.Services,
		"leases":         ec.stats.Leases,
		"operations":     ec.stats.Operations,
		"errors":         ec.stats.Errors,
		"avg_latency_ms": ec.stats.AvgLatency.Milliseconds(),
		"last_sync":      ec.stats.LastSyncTime,
		"cache_size":     len(ec.cache),
	}
}

// GetClusterInfo 获取集群信息
func (ec *EtcdCenter) GetClusterInfo(ctx context.Context) (map[string]interface{}, error) {
	if ec.client == nil {
		return nil, fmt.Errorf("未连接到etcd")
	}

	// 获取成员列表
	membersResp, err := ec.client.MemberList(ctx)
	if err != nil {
		return nil, err
	}

	members := make([]map[string]interface{}, 0)
	for _, m := range membersResp.Members {
		members = append(members, map[string]interface{}{
			"id":         m.ID,
			"name":       m.Name,
			"peer_urls":  m.PeerURLs,
			"client_urls": m.ClientURLs,
		})
	}

	// 获取集群状态
	statusResp, err := ec.client.Status(ctx, ec.config.Endpoints[0])
	if err == nil {
		ec.mu.Lock()
		ec.stats.Leader = fmt.Sprintf("%d", statusResp.Leader)
		ec.mu.Unlock()
	}

	return map[string]interface{}{
		"members":     members,
		"leader":      ec.stats.Leader,
		"endpoints":   ec.config.Endpoints,
		"connected":   ec.running,
	}, nil
}

// HealthCheck 健康检查
func (ec *EtcdCenter) HealthCheck(ctx context.Context) error {
	if ec.client == nil {
		return fmt.Errorf("etcd未连接")
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := ec.client.Get(ctx, "health-check")
	if err != nil && err != context.DeadlineExceeded {
		return fmt.Errorf("etcd健康检查失败: %w", err)
	}

	return nil
}

// IsConnected 检查连接状态
func (ec *EtcdCenter) IsConnected() bool {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	return ec.running
}

// GetClient 获取etcd客户端
func (ec *EtcdCenter) GetClient() *clientv3.Client {
	return ec.client
}

// Clear 清空配置
func (ec *EtcdCenter) Clear(ctx context.Context) error {
	prefix := ec.config.Prefix
	if prefix == "" {
		prefix = ""
	}

	_, err := ec.client.Delete(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	ec.mu.Lock()
	ec.cache = make(map[string]*ConfigItem)
	ec.stats.TotalKeys = 0
	ec.mu.Unlock()

	return nil
}