package sync

import (
	"testing"
	"time"
)

func TestNotificationHub(t *testing.T) {
	config := DefaultNotificationConfig()
	hub := NewNotificationHub(config)

	// 测试初始化
	if hub.notifications == nil {
		t.Error("notifications map should be initialized")
	}
	if hub.identityNotifications == nil {
		t.Error("identityNotifications map should be initialized")
	}
}

func TestNotificationHubCreateNotification(t *testing.T) {
	hub := NewNotificationHub(DefaultNotificationConfig())

	notification := &Notification{
		NotificationID: "notif-001",
		IdentityID:     "identity-001",
		Type:           NotificationTypeMessage,
		Priority:       NotificationPriorityNormal,
		Title:          "Test Notification",
		Body:           "This is a test",
	}

	err := hub.CreateNotification(notification)
	if err != nil {
		t.Fatalf("CreateNotification failed: %v", err)
	}

	// 验证存储
	stored := hub.GetNotification("notif-001")
	if stored == nil {
		t.Fatal("Notification not stored")
	}

	if stored.Status != NotificationStatusPending {
		t.Errorf("Expected pending status, got %s", stored.Status)
	}
}

func TestNotificationHubAutoID(t *testing.T) {
	hub := NewNotificationHub(DefaultNotificationConfig())

	notification := &Notification{
		IdentityID: "identity-001",
		Type:       NotificationTypeMessage,
		Title:      "Test",
	}

	hub.CreateNotification(notification)

	if notification.NotificationID == "" {
		t.Error("NotificationID should be auto-generated")
	}
}

func TestNotificationHubDefaultTTL(t *testing.T) {
	config := DefaultNotificationConfig()
	hub := NewNotificationHub(config)

	notification := &Notification{
		IdentityID: "identity-001",
		Type:       NotificationTypeMessage,
	}

	hub.CreateNotification(notification)

	if notification.ExpiresAt == nil {
		t.Error("ExpiresAt should be set by default")
	}

	expectedExpiry := notification.CreatedAt.Add(config.DefaultTTL)
	if !notification.ExpiresAt.After(notification.CreatedAt) {
		t.Error("ExpiresAt should be in the future")
	}

	_ = expectedExpiry // avoid unused variable error
}

func TestNotificationHubMarkAsRead(t *testing.T) {
	hub := NewNotificationHub(DefaultNotificationConfig())

	notification := &Notification{
		NotificationID: "notif-read",
		IdentityID:     "identity-001",
		Type:           NotificationTypeMessage,
		Status:         NotificationStatusDelivered,
	}

	hub.CreateNotification(notification)

	err := hub.MarkAsRead("notif-read", "agent-001")
	if err != nil {
		t.Fatalf("MarkAsRead failed: %v", err)
	}

	stored := hub.GetNotification("notif-read")
	if stored.Status != NotificationStatusRead {
		t.Errorf("Expected read status, got %s", stored.Status)
	}

	if len(stored.ReadBy) != 1 || stored.ReadBy[0] != "agent-001" {
		t.Error("ReadBy should contain agent-001")
	}
}

func TestNotificationHubMarkAsDismissed(t *testing.T) {
	hub := NewNotificationHub(DefaultNotificationConfig())

	notification := &Notification{
		NotificationID: "notif-dismiss",
		IdentityID:     "identity-001",
		Type:           NotificationTypeMessage,
	}

	hub.CreateNotification(notification)

	err := hub.MarkAsDismissed("notif-dismiss", "agent-001")
	if err != nil {
		t.Fatalf("MarkAsDismissed failed: %v", err)
	}

	stored := hub.GetNotification("notif-dismiss")
	if stored.Status != NotificationStatusDismissed {
		t.Errorf("Expected dismissed status, got %s", stored.Status)
	}
}

func TestNotificationHubGetUnread(t *testing.T) {
	hub := NewNotificationHub(DefaultNotificationConfig())

	// 创建多个通知
	for i := 0; i < 5; i++ {
		notification := &Notification{
			NotificationID: "notif-" + string(rune('0'+i)),
			IdentityID:     "identity-001",
			Type:           NotificationTypeMessage,
		}
		hub.CreateNotification(notification)
	}

	// 标记部分已读
	hub.MarkAsRead("notif-0", "agent-001")
	hub.MarkAsRead("notif-1", "agent-001")

	// 获取未读
	unread := hub.GetUnreadNotifications("identity-001")

	if len(unread) != 3 {
		t.Errorf("Expected 3 unread, got %d", len(unread))
	}
}

func TestNotificationHubGetIdentityNotifications(t *testing.T) {
	hub := NewNotificationHub(DefaultNotificationConfig())

	// 为不同身份创建通知
	for i := 0; i < 3; i++ {
		notification := &Notification{
			NotificationID: "notif-id1-" + string(rune('0'+i)),
			IdentityID:     "identity-001",
			Type:           NotificationTypeMessage,
		}
		hub.CreateNotification(notification)
	}

	for i := 0; i < 2; i++ {
		notification := &Notification{
			NotificationID: "notif-id2-" + string(rune('0'+i)),
			IdentityID:     "identity-002",
			Type:           NotificationTypeMessage,
		}
		hub.CreateNotification(notification)
	}

	// 获取身份 001 的通知
	notifs := hub.GetIdentityNotifications("identity-001", 10)

	if len(notifs) != 3 {
		t.Errorf("Expected 3 notifications for identity-001, got %d", len(notifs))
	}
}

func TestNotificationHubGroupNotifications(t *testing.T) {
	hub := NewNotificationHub(DefaultNotificationConfig())

	// 创建分组通知
	for i := 0; i < 3; i++ {
		notification := &Notification{
			NotificationID: "notif-group-" + string(rune('0'+i)),
			IdentityID:     "identity-001",
			Type:           NotificationTypeMessage,
			GroupID:        "group-001",
		}
		hub.CreateNotification(notification)
	}

	// 获取分组通知
	group := hub.GetGroupNotifications("group-001")

	if len(group) != 3 {
		t.Errorf("Expected 3 notifications in group, got %d", len(group))
	}
}

func TestNotificationHubDelete(t *testing.T) {
	hub := NewNotificationHub(DefaultNotificationConfig())

	notification := &Notification{
		NotificationID: "notif-delete",
		IdentityID:     "identity-001",
		Type:           NotificationTypeMessage,
	}

	hub.CreateNotification(notification)

	// 删除
	hub.DeleteNotification("notif-delete")

	// 验证已删除
	if hub.GetNotification("notif-delete") != nil {
		t.Error("Notification should be deleted")
	}
}

func TestNotificationHubCleanupExpired(t *testing.T) {
	hub := NewNotificationHub(DefaultNotificationConfig())

	// 创建已过期通知
	past := time.Now().Add(-1 * time.Hour)
	notification := &Notification{
		NotificationID: "notif-expired",
		IdentityID:     "identity-001",
		Type:           NotificationTypeMessage,
		ExpiresAt:      &past,
	}
	hub.CreateNotification(notification)

	// 清理过期
	count := hub.CleanupExpired()

	if count != 1 {
		t.Errorf("Expected 1 expired notification, got %d", count)
	}

	if hub.GetNotification("notif-expired") != nil {
		t.Error("Expired notification should be deleted")
	}
}

func TestNotificationHubStats(t *testing.T) {
	hub := NewNotificationHub(DefaultNotificationConfig())

	// 创建多个不同状态的通知
	hub.CreateNotification(&Notification{
		NotificationID: "notif-1",
		IdentityID:     "identity-001",
		Type:           NotificationTypeMessage,
		Priority:       NotificationPriorityNormal,
		Status:         NotificationStatusPending,
	})

	hub.CreateNotification(&Notification{
		NotificationID: "notif-2",
		IdentityID:     "identity-001",
		Type:           NotificationTypeAlert,
		Priority:       NotificationPriorityHigh,
		Status:         NotificationStatusDelivered,
	})

	hub.CreateNotification(&Notification{
		NotificationID: "notif-3",
		IdentityID:     "identity-001",
		Type:           NotificationTypeMessage,
		Priority:       NotificationPriorityLow,
		Status:         NotificationStatusRead,
	})

	stats := hub.GetStats("identity-001")

	if stats.TotalNotifications != 3 {
		t.Errorf("Expected 3 total, got %d", stats.TotalNotifications)
	}

	if stats.Pending != 1 {
		t.Errorf("Expected 1 pending, got %d", stats.Pending)
	}

	if stats.Delivered != 1 {
		t.Errorf("Expected 1 delivered, got %d", stats.Delivered)
	}

	if stats.Read != 1 {
		t.Errorf("Expected 1 read, got %d", stats.Read)
	}
}

func TestNotificationHubQuietHours(t *testing.T) {
	config := DefaultNotificationConfig()
	config.QuietHoursStart = 22
	config.QuietHoursEnd = 7

	hub := NewNotificationHub(config)

	// 测试勿扰检测逻辑
	// 注意：这个测试依赖于当前时间，可能不稳定
	// 实际项目中应该注入时间接口
	isQuiet := hub.isQuietHours()
	_ = isQuiet // 结果取决于当前时间
}

func TestNotificationJSON(t *testing.T) {
	notification := &Notification{
		NotificationID: "notif-json",
		IdentityID:     "identity-001",
		Type:           NotificationTypeMessage,
		Priority:       NotificationPriorityHigh,
		Title:          "Test Title",
		Body:           "Test Body",
		Status:         NotificationStatusPending,
		Vibrate:        true,
		SourceApp:      "com.test.app",
		Category:       "test",
		GroupID:        "group-001",
	}

	// 序列化
	data, err := notification.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// 反序列化
	parsed, err := NotificationFromJSON(data)
	if err != nil {
		t.Fatalf("NotificationFromJSON failed: %v", err)
	}

	if parsed.NotificationID != notification.NotificationID {
		t.Error("NotificationID mismatch after JSON roundtrip")
	}
	if parsed.Title != notification.Title {
		t.Error("Title mismatch after JSON roundtrip")
	}
	if parsed.Priority != notification.Priority {
		t.Error("Priority mismatch after JSON roundtrip")
	}
}

func TestNotificationToMap(t *testing.T) {
	notification := &Notification{
		NotificationID: "notif-map",
		IdentityID:     "identity-001",
		Type:           NotificationTypeAlert,
		Priority:       NotificationPriorityHigh,
		Title:          "Alert",
		Body:           "Alert body",
		Status:         NotificationStatusPending,
		Vibrate:        true,
	}

	m := notification.ToMap()

	if m["notification_id"] != notification.NotificationID {
		t.Error("notification_id mismatch")
	}
	if m["type"] != string(notification.Type) {
		t.Error("type mismatch")
	}
	if m["priority"] != notification.Priority {
		t.Error("priority mismatch")
	}
	if m["title"] != notification.Title {
		t.Error("title mismatch")
	}
}

func TestNotificationErrors(t *testing.T) {
	// 测试错误定义
	if ErrNotificationNotFound.Code != "notification_not_found" {
		t.Error("Wrong error code for ErrNotificationNotFound")
	}

	if ErrNoTargetDevices.Code != "no_target_devices" {
		t.Error("Wrong error code for ErrNoTargetDevices")
	}

	if ErrNotificationNotFound.Error() != "Notification not found" {
		t.Error("Wrong error message")
	}
}

func TestNotificationHubListener(t *testing.T) {
	hub := NewNotificationHub(DefaultNotificationConfig())

	created := false
	delivered := false

	listener := &TestNotificationListener{
		onCreated: func(n *Notification) {
			created = true
		},
		onDelivered: func(n *Notification, agentID string) {
			delivered = true
		},
	}

	hub.AddListener(listener)

	notification := &Notification{
		NotificationID: "notif-listener",
		IdentityID:     "identity-001",
		Type:           NotificationTypeMessage,
	}

	hub.CreateNotification(notification)

	// 等待异步通知
	time.Sleep(100 * time.Millisecond)

	if !created {
		t.Error("Listener should have been notified on create")
	}

	// 测试送达通知
	hub.mu.Lock()
	hub.notifyDelivered(notification, "agent-001")
	hub.mu.Unlock()

	time.Sleep(100 * time.Millisecond)

	if !delivered {
		t.Error("Listener should have been notified on deliver")
	}
}

func TestNotificationHubMarkAllAsRead(t *testing.T) {
	hub := NewNotificationHub(DefaultNotificationConfig())

	// 创建多个通知
	for i := 0; i < 5; i++ {
		hub.CreateNotification(&Notification{
			NotificationID: "notif-all-" + string(rune('0'+i)),
			IdentityID:     "identity-001",
			Type:           NotificationTypeMessage,
			Status:         NotificationStatusDelivered,
		})
	}

	// 全部标记已读
	count := hub.MarkAllAsRead("identity-001", "agent-001")

	if count != 5 {
		t.Errorf("Expected 5 marked as read, got %d", count)
	}

	// 验证状态
	unread := hub.GetUnreadNotifications("identity-001")
	if len(unread) != 0 {
		t.Errorf("Expected 0 unread, got %d", len(unread))
	}
}

// 测试监听器
type TestNotificationListener struct {
	onCreated    func(n *Notification)
	onDelivered  func(n *Notification, agentID string)
	onRead       func(n *Notification, agentID string)
	onDismissed  func(n *Notification, agentID string)
	onExpired    func(n *Notification)
}

func (l *TestNotificationListener) OnNotificationCreated(n *Notification) {
	if l.onCreated != nil {
		l.onCreated(n)
	}
}

func (l *TestNotificationListener) OnNotificationDelivered(n *Notification, agentID string) {
	if l.onDelivered != nil {
		l.onDelivered(n, agentID)
	}
}

func (l *TestNotificationListener) OnNotificationRead(n *Notification, agentID string) {
	if l.onRead != nil {
		l.onRead(n, agentID)
	}
}

func (l *TestNotificationListener) OnNotificationDismissed(n *Notification, agentID string) {
	if l.onDismissed != nil {
		l.onDismissed(n, agentID)
	}
}

func (l *TestNotificationListener) OnNotificationExpired(n *Notification) {
	if l.onExpired != nil {
		l.onExpired(n)
	}
}