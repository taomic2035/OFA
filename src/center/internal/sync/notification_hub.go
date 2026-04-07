package sync

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

// === 跨设备通知系统 (v3.4.0) ===
//
// Center 是永远在线的灵魂载体，跨设备通知确保：
// - 通知智能分发到合适的设备
// - 支持通知分组和优先级
// - 支持通知静默和勿扰模式
// - 支持通知历史和状态追踪

// NotificationType - 通知类型
type NotificationType string

const (
	NotificationTypeMessage  NotificationType = "message"   // 消息通知
	NotificationTypeAlert    NotificationType = "alert"     // 警报通知
	NotificationTypeReminder NotificationType = "reminder"  // 提醒通知
	NotificationTypeSystem   NotificationType = "system"    // 系统通知
	NotificationTypeSocial   NotificationType = "social"    // 社交通知
	NotificationTypeHealth   NotificationType = "health"    // 健康通知
	NotificationTypeCalendar NotificationType = "calendar"  // 日历通知
	NotificationTypeCall     NotificationType = "call"      // 来电通知
)

// NotificationPriority - 通知优先级
type NotificationPriority int

const (
	NotificationPriorityMin    NotificationPriority = 0
	NotificationPriorityLow    NotificationPriority = 1
	NotificationPriorityNormal NotificationPriority = 2
	NotificationPriorityHigh   NotificationPriority = 3
	NotificationPriorityMax    NotificationPriority = 4
)

// NotificationStatus - 通知状态
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusDelivered NotificationStatus = "delivered"
	NotificationStatusRead      NotificationStatus = "read"
	NotificationStatusDismissed NotificationStatus = "dismissed"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusExpired   NotificationStatus = "expired"
)

// NotificationAction - 通知动作
type NotificationAction struct {
	ActionID  string `json:"action_id"`
	Title     string `json:"title"`
	Type      string `json:"type"`      // open, reply, dismiss, custom
	Payload   map[string]interface{} `json:"payload,omitempty"`
}

// Notification - 通知模型
type Notification struct {
	NotificationID string                 `json:"notification_id"`
	IdentityID     string                 `json:"identity_id"`
	Type           NotificationType       `json:"type"`
	Priority       NotificationPriority   `json:"priority"`
	Title          string                 `json:"title"`
	Body           string                 `json:"body"`
	Icon           string                 `json:"icon,omitempty"`
	Image          string                 `json:"image,omitempty"`
	Sound          string                 `json:"sound,omitempty"`
	Vibrate        bool                   `json:"vibrate"`
	Actions        []NotificationAction   `json:"actions,omitempty"`

	// 目标设备
	TargetDevices  []string               `json:"target_devices,omitempty"` // 指定设备
	TargetScenes   []SceneType            `json:"target_scenes,omitempty"`  // 指定场景
	ExcludeDevices []string               `json:"exclude_devices,omitempty"`

	// 状态追踪
	Status         NotificationStatus     `json:"status"`
	DeliveredTo    []string               `json:"delivered_to,omitempty"`  // 已送达设备
	ReadBy         []string               `json:"read_by,omitempty"`       // 已读设备
	DismissedBy    []string               `json:"dismissed_by,omitempty"`  // 已忽略设备

	// 时间信息
	CreatedAt      time.Time              `json:"created_at"`
	ExpiresAt      *time.Time             `json:"expires_at,omitempty"`
	DeliveredAt    *time.Time             `json:"delivered_at,omitempty"`
	ReadAt         *time.Time             `json:"read_at,omitempty"`

	// 元数据
	SourceApp      string                 `json:"source_app,omitempty"`
	SourceAgent    string                 `json:"source_agent,omitempty"`
	Category       string                 `json:"category,omitempty"`
	GroupID        string                 `json:"group_id,omitempty"`      // 分组ID
	ThreadID       string                 `json:"thread_id,omitempty"`     // 会话ID

	// 扩展数据
	Data           map[string]interface{} `json:"data,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// NotificationConfig - 通知配置
type NotificationConfig struct {
	// 默认过期时间
	DefaultTTL time.Duration
	// 最大通知历史
	MaxHistory int
	// 勿扰模式配置
	QuietHoursStart int // 22 (22:00)
	QuietHoursEnd   int // 7  (07:00)
}

// DefaultNotificationConfig 默认配置
func DefaultNotificationConfig() NotificationConfig {
	return NotificationConfig{
		DefaultTTL:      24 * time.Hour,
		MaxHistory:      1000,
		QuietHoursStart: 22,
		QuietHoursEnd:   7,
	}
}

// NotificationHub - 通知中心
type NotificationHub struct {
	mu sync.RWMutex

	// 通知存储
	notifications map[string]*Notification

	// 身份通知索引: identityId -> []notificationId
	identityNotifications map[string][]string

	// 分组索引: groupId -> []notificationId
	groupNotifications map[string][]string

	// 配置
	config NotificationConfig

	// 设备管理器
	deviceManager *DeviceManager

	// 消息总线
	messageBus *MessageBus

	// 状态管理器
	stateManager *StateSyncManager

	// 场景路由器
	sceneRouter *SceneRouter

	// 通知监听器
	listeners []NotificationListener
}

// NotificationListener - 通知监听器
type NotificationListener interface {
	OnNotificationCreated(notification *Notification)
	OnNotificationDelivered(notification *Notification, agentID string)
	OnNotificationRead(notification *Notification, agentID string)
	OnNotificationDismissed(notification *Notification, agentID string)
	OnNotificationExpired(notification *Notification)
}

// NewNotificationHub 创建通知中心
func NewNotificationHub(config NotificationConfig) *NotificationHub {
	if config.MaxHistory == 0 {
		config = DefaultNotificationConfig()
	}

	return &NotificationHub{
		notifications:          make(map[string]*Notification),
		identityNotifications:  make(map[string][]string),
		groupNotifications:     make(map[string][]string),
		config:                 config,
		listeners:              make([]NotificationListener, 0),
	}
}

// SetDeviceManager 设置设备管理器
func (h *NotificationHub) SetDeviceManager(dm *DeviceManager) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.deviceManager = dm
}

// SetMessageBus 设置消息总线
func (h *NotificationHub) SetMessageBus(mb *MessageBus) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.messageBus = mb
}

// SetStateManager 设置状态管理器
func (h *NotificationHub) SetStateManager(sm *StateSyncManager) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.stateManager = sm
}

// SetSceneRouter 设置场景路由器
func (h *NotificationHub) SetSceneRouter(sr *SceneRouter) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sceneRouter = sr
}

// === 通知创建与发送 ===

// CreateNotification 创建通知
func (h *NotificationHub) CreateNotification(notification *Notification) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 设置默认值
	if notification.NotificationID == "" {
		notification.NotificationID = generateNotificationID()
	}
	if notification.Status == "" {
		notification.Status = NotificationStatusPending
	}
	notification.CreatedAt = time.Now()

	// 设置过期时间
	if notification.ExpiresAt == nil {
		expires := notification.CreatedAt.Add(h.config.DefaultTTL)
		notification.ExpiresAt = &expires
	}

	// 存储通知
	h.notifications[notification.NotificationID] = notification

	// 更新索引
	h.identityNotifications[notification.IdentityID] = append(
		h.identityNotifications[notification.IdentityID],
		notification.NotificationID,
	)

	if notification.GroupID != "" {
		h.groupNotifications[notification.GroupID] = append(
			h.groupNotifications[notification.GroupID],
			notification.NotificationID,
		)
	}

	// 通知监听器
	go h.notifyCreated(notification)

	log.Printf("Notification created: %s (type=%s, priority=%d)",
		notification.NotificationID, notification.Type, notification.Priority)

	return nil
}

// SendNotification 发送通知
func (h *NotificationHub) SendNotification(notificationID string) error {
	h.mu.RLock()
	notification, exists := h.notifications[notificationID]
	h.mu.RUnlock()

	if !exists {
		return ErrNotificationNotFound
	}

	// 确定目标设备
	targetAgents := h.determineTargets(notification)
	if len(targetAgents) == 0 {
		return ErrNoTargetDevices
	}

	// 发送到目标设备
	for _, agentID := range targetAgents {
		go h.deliverToAgent(notification, agentID)
	}

	return nil
}

// CreateAndSend 创建并发送通知
func (h *NotificationHub) CreateAndSend(notification *Notification) error {
	err := h.CreateNotification(notification)
	if err != nil {
		return err
	}

	return h.SendNotification(notification.NotificationID)
}

// determineTargets 确定目标设备
func (h *NotificationHub) determineTargets(notification *Notification) []string {
	var targets []string

	// 1. 指定设备优先
	if len(notification.TargetDevices) > 0 {
		return notification.filterOnlineDevices(notification.TargetDevices, h.deviceManager)
	}

	// 2. 使用场景路由
	if h.sceneRouter != nil && len(notification.TargetScenes) > 0 {
		ctx := &RoutingContext{
			IdentityID:  notification.IdentityID,
			MessageType: MessageType(string(notification.Type)),
			Priority:    int(notification.Priority),
		}

		for _, scene := range notification.TargetScenes {
			ctx.Scene = scene
			result := h.sceneRouter.Route(ctx)
			targets = append(targets, result.TargetAgents...)
		}

		return uniqueStrings(targets)
	}

	// 3. 默认广播到所有在线设备
	if h.deviceManager != nil {
		devices := h.deviceManager.GetActiveDevices(notification.IdentityID)
		for _, d := range devices {
			// 排除指定设备
			if !containsString(notification.ExcludeDevices, d.AgentID) {
				targets = append(targets, d.AgentID)
			}
		}
	}

	// 4. 场景感知优化
	if h.stateManager != nil && h.sceneRouter != nil {
		targets = h.optimizeTargets(notification, targets)
	}

	return targets
}

// optimizeTargets 优化目标设备
func (h *NotificationHub) optimizeTargets(notification *Notification, targets []string) []string {
	if h.stateManager == nil || h.sceneRouter == nil {
		return targets
	}

	var optimized []string

	for _, agentID := range targets {
		state := h.stateManager.GetDeviceState(agentID)
		if state == nil {
			continue
		}

		// 检查勿扰模式
		if h.isQuietHours() && notification.Priority < NotificationPriorityHigh {
			// 低优先级通知在勿扰时段不发送
			continue
		}

		// 睡眠场景只发送高优先级通知
		if state.Scene == string(SceneSleeping) && notification.Priority < NotificationPriorityHigh {
			continue
		}

		// 会议场景过滤社交通知
		if state.Scene == string(SceneMeeting) && notification.Type == NotificationTypeSocial {
			if notification.Priority < NotificationPriorityHigh {
				continue
			}
		}

		optimized = append(optimized, agentID)
	}

	// 如果所有设备都被过滤，选择主设备
	if len(optimized) == 0 && notification.Priority >= NotificationPriorityHigh {
		if h.deviceManager != nil {
			primary := h.deviceManager.GetPrimaryDevice(notification.IdentityID)
			if primary != nil {
				optimized = append(optimized, primary.AgentID)
			}
		}
	}

	return optimized
}

// isQuietHours 检查是否在勿扰时段
func (h *NotificationHub) isQuietHours() bool {
	hour := time.Now().Hour()

	if h.config.QuietHoursStart > h.config.QuietHoursEnd {
		// 跨午夜: 22:00 - 07:00
		return hour >= h.config.QuietHoursStart || hour < h.config.QuietHoursEnd
	}

	// 同一天: 14:00 - 17:00
	return hour >= h.config.QuietHoursStart && hour < h.config.QuietHoursEnd
}

// deliverToAgent 发送通知到设备
func (h *NotificationHub) deliverToAgent(notification *Notification, agentID string) {
	if h.messageBus == nil {
		return
	}

	msg := &Message{
		ID:         generateMessageID(),
		FromAgent:  "center",
		ToAgent:    agentID,
		IdentityID: notification.IdentityID,
		Type:       MessageTypeNotification,
		Priority:   MessagePriority(notification.Priority),
		Payload:    notification.ToMap(),
		CreatedAt:  time.Now(),
	}

	h.messageBus.Send(msg)

	// 更新送达状态
	h.mu.Lock()
	notification.DeliveredTo = append(notification.DeliveredTo, agentID)
	if notification.Status == NotificationStatusPending {
		notification.Status = NotificationStatusDelivered
		now := time.Now()
		notification.DeliveredAt = &now
	}
	h.mu.Unlock()

	// 通知监听器
	go h.notifyDelivered(notification, agentID)
}

// === 通知状态更新 ===

// MarkAsRead 标记为已读
func (h *NotificationHub) MarkAsRead(notificationID, agentID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	notification, exists := h.notifications[notificationID]
	if !exists {
		return ErrNotificationNotFound
	}

	notification.ReadBy = append(notification.ReadBy, agentID)
	notification.Status = NotificationStatusRead
	now := time.Now()
	notification.ReadAt = &now

	go h.notifyRead(notification, agentID)

	return nil
}

// MarkAsDismissed 标记为已忽略
func (h *NotificationHub) MarkAsDismissed(notificationID, agentID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	notification, exists := h.notifications[notificationID]
	if !exists {
		return ErrNotificationNotFound
	}

	notification.DismissedBy = append(notification.DismissedBy, agentID)
	notification.Status = NotificationStatusDismissed

	go h.notifyDismissed(notification, agentID)

	return nil
}

// MarkAllAsRead 标记所有为已读
func (h *NotificationHub) MarkAllAsRead(identityID, agentID string) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	count := 0
	for _, notifID := range h.identityNotifications[identityID] {
		notification := h.notifications[notifID]
		if notification != nil && notification.Status != NotificationStatusRead {
			notification.Status = NotificationStatusRead
			notification.ReadBy = append(notification.ReadBy, agentID)
			now := time.Now()
			notification.ReadAt = &now
			count++
		}
	}

	return count
}

// === 通知查询 ===

// GetNotification 获取通知
func (h *NotificationHub) GetNotification(notificationID string) *Notification {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.notifications[notificationID]
}

// GetIdentityNotifications 获取身份通知列表
func (h *NotificationHub) GetIdentityNotifications(identityID string, limit int) []*Notification {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var result []*Notification
	ids := h.identityNotifications[identityID]

	// 从最新开始
	for i := len(ids) - 1; i >= 0 && len(result) < limit; i-- {
		if notification := h.notifications[ids[i]]; notification != nil {
			result = append(result, notification)
		}
	}

	return result
}

// GetUnreadNotifications 获取未读通知
func (h *NotificationHub) GetUnreadNotifications(identityID string) []*Notification {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var result []*Notification
	for _, notifID := range h.identityNotifications[identityID] {
		if notification := h.notifications[notifID]; notification != nil {
			if notification.Status != NotificationStatusRead && notification.Status != NotificationStatusDismissed {
				result = append(result, notification)
			}
		}
	}

	return result
}

// GetGroupNotifications 获取分组通知
func (h *NotificationHub) GetGroupNotifications(groupID string) []*Notification {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var result []*Notification
	for _, notifID := range h.groupNotifications[groupID] {
		if notification := h.notifications[notifID]; notification != nil {
			result = append(result, notification)
		}
	}

	return result
}

// === 通知管理 ===

// CancelNotification 取消通知
func (h *NotificationHub) CancelNotification(notificationID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	notification, exists := h.notifications[notificationID]
	if !exists {
		return ErrNotificationNotFound
	}

	notification.Status = NotificationStatusDismissed

	return nil
}

// DeleteNotification 删除通知
func (h *NotificationHub) DeleteNotification(notificationID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	notification, exists := h.notifications[notificationID]
	if !exists {
		return
	}

	// 从索引移除
	delete(h.notifications, notificationID)

	if notification != nil {
		// 从身份索引移除
		var newIdentityNotifs []string
		for _, id := range h.identityNotifications[notification.IdentityID] {
			if id != notificationID {
				newIdentityNotifs = append(newIdentityNotifs, id)
			}
		}
		h.identityNotifications[notification.IdentityID] = newIdentityNotifs

		// 从分组索引移除
		if notification.GroupID != "" {
			var newGroupNotifs []string
			for _, id := range h.groupNotifications[notification.GroupID] {
				if id != notificationID {
					newGroupNotifs = append(newGroupNotifs, id)
				}
			}
			h.groupNotifications[notification.GroupID] = newGroupNotifs
		}
	}
}

// CleanupExpired 清理过期通知
func (h *NotificationHub) CleanupExpired() int {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	expired := make([]string, 0)

	for id, notification := range h.notifications {
		if notification.ExpiresAt != nil && now.After(*notification.ExpiresAt) {
			notification.Status = NotificationStatusExpired
			expired = append(expired, id)
			go h.notifyExpired(notification)
		}
	}

	for _, id := range expired {
		h.DeleteNotification(id)
	}

	return len(expired)
}

// === 监听器管理 ===

func (h *NotificationHub) AddListener(listener NotificationListener) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.listeners = append(h.listeners, listener)
}

func (h *NotificationHub) notifyCreated(notification *Notification) {
	for _, l := range h.listeners {
		l.OnNotificationCreated(notification)
	}
}

func (h *NotificationHub) notifyDelivered(notification *Notification, agentID string) {
	for _, l := range h.listeners {
		l.OnNotificationDelivered(notification, agentID)
	}
}

func (h *NotificationHub) notifyRead(notification *Notification, agentID string) {
	for _, l := range h.listeners {
		l.OnNotificationRead(notification, agentID)
	}
}

func (h *NotificationHub) notifyDismissed(notification *Notification, agentID string) {
	for _, l := range h.listeners {
		l.OnNotificationDismissed(notification, agentID)
	}
}

func (h *NotificationHub) notifyExpired(notification *Notification) {
	for _, l := range h.listeners {
		l.OnNotificationExpired(notification)
	}
}

// === 统计 ===

// NotificationStats 通知统计
type NotificationStats struct {
	TotalNotifications int            `json:"total_notifications"`
	Pending            int            `json:"pending"`
	Delivered          int            `json:"delivered"`
	Read               int            `json:"read"`
	Dismissed          int            `json:"dismissed"`
	ByType             map[string]int `json:"by_type"`
	ByPriority         map[int]int    `json:"by_priority"`
}

// GetStats 获取统计
func (h *NotificationHub) GetStats(identityID string) *NotificationStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := &NotificationStats{
		ByType:     make(map[string]int),
		ByPriority: make(map[int]int),
	}

	for _, notifID := range h.identityNotifications[identityID] {
		if notification := h.notifications[notifID]; notification != nil {
			stats.TotalNotifications++

			switch notification.Status {
			case NotificationStatusPending:
				stats.Pending++
			case NotificationStatusDelivered:
				stats.Delivered++
			case NotificationStatusRead:
				stats.Read++
			case NotificationStatusDismissed:
				stats.Dismissed++
			}

			stats.ByType[string(notification.Type)]++
			stats.ByPriority[int(notification.Priority)]++
		}
	}

	return stats
}

// === 辅助方法 ===

func (n *Notification) filterOnlineDevices(devices []string, dm *DeviceManager) []string {
	if dm == nil {
		return devices
	}

	var online []string
	for _, agentID := range devices {
		if dm.IsDeviceOnline(agentID) {
			online = append(online, agentID)
		}
	}
	return online
}

// ToMap 转换为 Map
func (n *Notification) ToMap() map[string]interface{} {
	m := map[string]interface{}{
		"notification_id": n.NotificationID,
		"identity_id":     n.IdentityID,
		"type":            n.Type,
		"priority":        n.Priority,
		"title":           n.Title,
		"body":            n.Body,
		"status":          n.Status,
		"vibrate":         n.Vibrate,
		"created_at":      n.CreatedAt,
	}

	if n.Icon != "" {
		m["icon"] = n.Icon
	}
	if n.Image != "" {
		m["image"] = n.Image
	}
	if n.Sound != "" {
		m["sound"] = n.Sound
	}
	if n.SourceApp != "" {
		m["source_app"] = n.SourceApp
	}
	if n.Category != "" {
		m["category"] = n.Category
	}
	if n.GroupID != "" {
		m["group_id"] = n.GroupID
	}
	if n.Data != nil {
		m["data"] = n.Data
	}

	return m
}

// ToJSON 序列化
func (n *Notification) ToJSON() ([]byte, error) {
	return json.Marshal(n)
}

// FromJSON 解析
func NotificationFromJSON(data []byte) (*Notification, error) {
	var n Notification
	err := json.Unmarshal(data, &n)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func generateNotificationID() string {
	return "notif_" + time.Now().Format("20060102150405") + "_" + randomString(6)
}

func uniqueStrings(s []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

func containsString(s []string, e string) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

// 错误定义
var (
	ErrNotificationNotFound = &NotificationError{Code: "notification_not_found", Message: "Notification not found"}
	ErrNoTargetDevices      = &NotificationError{Code: "no_target_devices", Message: "No target devices available"}
)

// NotificationError 通知错误
type NotificationError struct {
	Code    string
	Message string
}

func (e *NotificationError) Error() string {
	return e.Message
}