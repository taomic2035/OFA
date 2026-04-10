package relationship

import (
	"testing"
	"time"

	"github.com/ofa/center/internal/models"
)

// TestRelationshipEngineCreation 测试关系引擎创建
func TestRelationshipEngineCreation(t *testing.T) {
	engine := NewRelationshipEngine()

	if engine == nil {
		t.Fatal("RelationshipEngine should not be nil")
	}
	if engine.relationshipSystems == nil {
		t.Error("relationshipSystems map should be initialized")
	}
	if engine.relationshipProfiles == nil {
		t.Error("relationshipProfiles map should be initialized")
	}
}

// TestGetOrCreateRelationshipSystem 测试获取或创建关系系统
func TestGetOrCreateRelationshipSystem(t *testing.T) {
	engine := NewRelationshipEngine()

	// 创建新关系系统
	system := engine.GetOrCreateRelationshipSystem("identity_001")
	if system == nil {
		t.Fatal("RelationshipSystem should not be nil")
	}

	// 验证默认值
	if system.Relationships == nil {
		t.Error("Relationships should be initialized")
	}
	if system.SocialNetwork == nil {
		t.Error("SocialNetwork should be initialized")
	}

	// 获取已存在的系统
	system2 := engine.GetOrCreateRelationshipSystem("identity_001")
	if system != system2 {
		t.Error("Should return same system instance")
	}
}

// TestAddRelationship 测试添加关系
func TestAddRelationship(t *testing.T) {
	engine := NewRelationshipEngine()

	// 创建监听器
	listener := &testRelationshipListener{
		relationshipAddedCount: 0,
	}
	engine.AddListener(listener)

	// 添加关系
	relationship := models.Relationship{
		PersonID:         "person_001",
		PersonName:       "张三",
		RelationshipType: "friend",
		Intimacy:         0.7,
		Trust:            0.8,
		Importance:       0.6,
	}

	err := engine.AddRelationship("identity_001", relationship)
	if err != nil {
		t.Fatalf("AddRelationship failed: %v", err)
	}

	// 验证关系添加
	system := engine.GetRelationshipSystem("identity_001")
	if len(system.Relationships) != 1 {
		t.Errorf("Should have 1 relationship, got %d", len(system.Relationships))
	}

	// 验证ID生成
	if system.Relationships[0].RelationshipID == "" {
		t.Error("RelationshipID should be generated")
	}

	// 验证默认阶段
	if system.Relationships[0].Stage != "acquaintance" {
		t.Errorf("Default Stage should be 'acquaintance', got %s", system.Relationships[0].Stage)
	}

	// 验证监听器通知
	if listener.relationshipAddedCount != 1 {
		t.Errorf("relationshipAddedCount should be 1, got %d", listener.relationshipAddedCount)
	}

	// 验证社交网络更新
	if system.SocialNetwork.TotalContacts != 1 {
		t.Errorf("TotalContacts should be 1, got %d", system.SocialNetwork.TotalContacts)
	}
}

// TestUpdateRelationship 测试更新关系
func TestUpdateRelationship(t *testing.T) {
	engine := NewRelationshipEngine()

	// 先添加关系
	relationship := models.Relationship{
		PersonID:   "person_001",
		PersonName: "张三",
		Intimacy:   0.5,
		Trust:      0.5,
	}
	engine.AddRelationship("identity_001", relationship)

	// 创建监听器
	listener := &testRelationshipListener{
		relationshipUpdatedCount: 0,
	}
	engine.AddListener(listener)

	// 获取关系ID
	system := engine.GetRelationshipSystem("identity_001")
	relID := system.Relationships[0].RelationshipID

	// 更新关系
	updated := models.Relationship{
		RelationshipID: relID,
		PersonID:       "person_001",
		PersonName:     "张三",
		Intimacy:       0.8,
		Trust:          0.9,
		Importance:     0.7,
	}

	err := engine.UpdateRelationship("identity_001", updated)
	if err != nil {
		t.Fatalf("UpdateRelationship failed: %v", err)
	}

	// 验证更新
	retrieved := engine.GetRelationship("identity_001", relID)
	if retrieved.Intimacy != 0.8 {
		t.Errorf("Intimacy should be 0.8, got %f", retrieved.Intimacy)
	}

	// 验证监听器通知
	if listener.relationshipUpdatedCount != 1 {
		t.Errorf("relationshipUpdatedCount should be 1, got %d", listener.relationshipUpdatedCount)
	}
}

// TestRemoveRelationship 测试移除关系
func TestRemoveRelationship(t *testing.T) {
	engine := NewRelationshipEngine()

	// 添加两个关系
	engine.AddRelationship("identity_001", models.Relationship{PersonID: "person_001", PersonName: "张三"})
	engine.AddRelationship("identity_001", models.Relationship{PersonID: "person_002", PersonName: "李四"})

	// 创建监听器
	listener := &testRelationshipListener{
		relationshipEndedCount: 0,
	}
	engine.AddListener(listener)

	// 获取关系ID
	system := engine.GetRelationshipSystem("identity_001")
	relID := system.Relationships[0].RelationshipID

	// 移除关系
	err := engine.RemoveRelationship("identity_001", relID)
	if err != nil {
		t.Fatalf("RemoveRelationship failed: %v", err)
	}

	// 验证移除
	system = engine.GetRelationshipSystem("identity_001")
	if len(system.Relationships) != 1 {
		t.Errorf("Should have 1 relationship after removal, got %d", len(system.Relationships))
	}

	// 验证监听器通知
	if listener.relationshipEndedCount != 1 {
		t.Errorf("relationshipEndedCount should be 1, got %d", listener.relationshipEndedCount)
	}

	// 验证社交网络更新
	if system.SocialNetwork.TotalContacts != 1 {
		t.Errorf("TotalContacts should be 1 after removal, got %d", system.SocialNetwork.TotalContacts)
	}
}

// TestGetRelationship 测试获取特定关系
func TestGetRelationship(t *testing.T) {
	engine := NewRelationshipEngine()

	relationship := models.Relationship{
		PersonID:   "person_001",
		PersonName: "张三",
		Intimacy:   0.7,
	}
	engine.AddRelationship("identity_001", relationship)

	system := engine.GetRelationshipSystem("identity_001")
	relID := system.Relationships[0].RelationshipID

	// 通过ID获取
	retrieved := engine.GetRelationship("identity_001", relID)
	if retrieved == nil {
		t.Fatal("Relationship should exist")
	}
	if retrieved.PersonName != "张三" {
		t.Errorf("PersonName should be '张三', got %s", retrieved.PersonName)
	}

	// 通过PersonID获取
	retrieved2 := engine.GetRelationshipByPersonID("identity_001", "person_001")
	if retrieved2 == nil {
		t.Fatal("Relationship should exist by PersonID")
	}
}

// TestGetRelationshipsByType 测试按类型获取关系
func TestGetRelationshipsByType(t *testing.T) {
	engine := NewRelationshipEngine()

	// 添加不同类型关系
	relationships := []models.Relationship{
		{PersonID: "person_001", RelationshipType: "friend"},
		{PersonID: "person_002", RelationshipType: "family"},
		{PersonID: "person_003", RelationshipType: "friend"},
		{PersonID: "person_004", RelationshipType: "colleague"},
	}

	for _, rel := range relationships {
		engine.AddRelationship("identity_001", rel)
	}

	// 获取朋友关系
	friends := engine.GetRelationshipsByType("identity_001", "friend")
	if len(friends) != 2 {
		t.Errorf("Should have 2 friends, got %d", len(friends))
	}

	// 获取家庭关系
	family := engine.GetRelationshipsByType("identity_001", "family")
	if len(family) != 1 {
		t.Errorf("Should have 1 family, got %d", len(family))
	}
}

// TestGetCloseRelationships 测试获取亲密关系
func TestGetCloseRelationships(t *testing.T) {
	engine := NewRelationshipEngine()

	// 添加不同亲密度的关系
	relationships := []models.Relationship{
		{PersonID: "person_001", Intimacy: 0.9, Trust: 0.8}, // 亲密
		{PersonID: "person_002", Intimacy: 0.3, Trust: 0.4}, // 不亲密
		{PersonID: "person_003", Intimacy: 0.7, Trust: 0.7}, // 亲密
	}

	for _, rel := range relationships {
		engine.AddRelationship("identity_001", rel)
	}

	closeRels := engine.GetCloseRelationships("identity_001")
	if len(closeRels) != 2 {
		t.Errorf("Should have 2 close relationships, got %d", len(closeRels))
	}
}

// TestRecordInteraction 测试记录互动
func TestRecordInteraction(t *testing.T) {
	engine := NewRelationshipEngine()

	// 添加关系
	engine.AddRelationship("identity_001", models.Relationship{
		PersonID:   "person_001",
		PersonName: "张三",
		Intimacy:   0.5,
		Trust:      0.5,
	})

	// 创建监听器
	listener := &testRelationshipListener{
		interactionRecordedCount: 0,
	}
	engine.AddListener(listener)

	// 获取关系ID
	system := engine.GetRelationshipSystem("identity_001")
	relID := system.Relationships[0].RelationshipID

	// 记录互动
	interaction := models.RelationshipInteraction{
		RelationshipID:  relID,
		InteractionType: "conversation",
		Satisfaction:    0.8,
		IntimacyChange:  0.1,
		TrustChange:     0.1,
	}

	err := engine.RecordInteraction("identity_001", interaction)
	if err != nil {
		t.Fatalf("RecordInteraction failed: %v", err)
	}

	// 验证互动统计更新
	retrieved := engine.GetRelationship("identity_001", relID)
	if retrieved.TotalInteractions != 1 {
		t.Errorf("TotalInteractions should be 1, got %d", retrieved.TotalInteractions)
	}
	if retrieved.PositiveInteractions != 1 {
		t.Errorf("PositiveInteractions should be 1, got %d", retrieved.PositiveInteractions)
	}

	// 验证关系属性更新
	if retrieved.Intimacy < 0.5 {
		t.Errorf("Intimacy should increase, got %f", retrieved.Intimacy)
	}
	if retrieved.Trust < 0.5 {
		t.Errorf("Trust should increase, got %f", retrieved.Trust)
	}

	// 验证监听器通知
	if listener.interactionRecordedCount != 1 {
		t.Errorf("interactionRecordedCount should be 1, got %d", listener.interactionRecordedCount)
	}
}

// TestRecordConflict 测试记录冲突
func TestRecordConflict(t *testing.T) {
	engine := NewRelationshipEngine()

	// 添加关系
	engine.AddRelationship("identity_001", models.Relationship{
		PersonID:   "person_001",
		PersonName: "张三",
	})

	// 创建监听器
	listener := &testRelationshipListener{
		conflictOccurredCount: 0,
	}
	engine.AddListener(listener)

	// 获取关系ID
	system := engine.GetRelationshipSystem("identity_001")
	relID := system.Relationships[0].RelationshipID

	// 记录冲突
	err := engine.RecordConflict("identity_001", relID, "意见不合", 0.6)
	if err != nil {
		t.Fatalf("RecordConflict failed: %v", err)
	}

	// 验证冲突级别更新
	retrieved := engine.GetRelationship("identity_001", relID)
	if retrieved.ConflictLevel != 0.6 {
		t.Errorf("ConflictLevel should be 0.6, got %f", retrieved.ConflictLevel)
	}
	if retrieved.Trend != "fluctuating" {
		t.Errorf("Trend should be 'fluctuating', got %s", retrieved.Trend)
	}

	// 验证监听器通知
	if listener.conflictOccurredCount != 1 {
		t.Errorf("conflictOccurredCount should be 1, got %d", listener.conflictOccurredCount)
	}
}

// TestResolveConflict 测试解决冲突
func TestResolveConflict(t *testing.T) {
	engine := NewRelationshipEngine()

	// 添加关系并记录冲突
	engine.AddRelationship("identity_001", models.Relationship{PersonID: "person_001"})
	system := engine.GetRelationshipSystem("identity_001")
	relID := system.Relationships[0].RelationshipID
	engine.RecordConflict("identity_001", relID, "测试", 0.7)

	// 解决冲突
	err := engine.ResolveConflict("identity_001", relID, "collaborative")
	if err != nil {
		t.Fatalf("ResolveConflict failed: %v", err)
	}

	// 验证冲突级别降低
	retrieved := engine.GetRelationship("identity_001", relID)
	if retrieved.ConflictLevel > 0.5 {
		t.Errorf("ConflictLevel should decrease, got %f", retrieved.ConflictLevel)
	}

	// 验证信任增加（协作解决）
	if retrieved.Trust < 0.1 {
		t.Errorf("Trust should increase after collaborative resolution, got %f", retrieved.Trust)
	}

	// 验证正面互动增加
	if retrieved.PositiveInteractions != 1 {
		t.Errorf("PositiveInteractions should be 1, got %d", retrieved.PositiveInteractions)
	}
}

// TestCreateCircle 测试创建社交圈
func TestCreateCircle(t *testing.T) {
	engine := NewRelationshipEngine()

	circle := models.SocialCircle{
		CircleName:   "同事圈",
		CircleType:   "work",
		Description:  "工作相关联系人",
	}

	err := engine.CreateCircle("identity_001", circle)
	if err != nil {
		t.Fatalf("CreateCircle failed: %v", err)
	}

	// 验证社交圈创建
	system := engine.GetRelationshipSystem("identity_001")
	if len(system.SocialNetwork.Circles) != 1 {
		t.Errorf("Should have 1 circle, got %d", len(system.SocialNetwork.Circles))
	}

	// 验证ID生成
	if system.SocialNetwork.Circles[0].CircleID == "" {
		t.Error("CircleID should be generated")
	}
}

// TestAddToCircle 测试添加到社交圈
func TestAddToCircle(t *testing.T) {
	engine := NewRelationshipEngine()

	// 创建社交圈
	engine.CreateCircle("identity_001", models.SocialCircle{CircleName: "朋友圈"})

	// 获取社交圈ID
	system := engine.GetRelationshipSystem("identity_001")
	circleID := system.SocialNetwork.Circles[0].CircleID

	// 添加成员
	err := engine.AddToCircle("identity_001", circleID, "person_001")
	if err != nil {
		t.Fatalf("AddToCircle failed: %v", err)
	}

	// 验证成员添加
	system = engine.GetRelationshipSystem("identity_001")
	if len(system.SocialNetwork.Circles[0].Members) != 1 {
		t.Errorf("Should have 1 member, got %d", len(system.SocialNetwork.Circles[0].Members))
	}

	// 添加重复成员
	engine.AddToCircle("identity_001", circleID, "person_001")

	system = engine.GetRelationshipSystem("identity_001")
	if len(system.SocialNetwork.Circles[0].Members) != 1 {
		t.Errorf("Should still have 1 member (no duplicates), got %d", len(system.SocialNetwork.Circles[0].Members))
	}
}

// TestGetDecisionContext 测试获取决策上下文
func TestGetDecisionContext(t *testing.T) {
	engine := NewRelationshipEngine()

	// 添加关系
	relationships := []models.Relationship{
		{PersonID: "person_001", PersonName: "亲密朋友", Intimacy: 0.9, Trust: 0.8, Importance: 0.8},
		{PersonID: "person_002", PersonName: "普通朋友", Intimacy: 0.5, Trust: 0.5, Importance: 0.4},
	}
	for _, rel := range relationships {
		engine.AddRelationship("identity_001", rel)
	}

	// 更新画像
	profile := models.NewRelationshipProfile("identity_001")
	profile.AttachmentStyle = models.AttachmentStyleProfile{
		StyleName:        "secure",
		AnxietyLevel:     0.2,
		AvoidanceLevel:   0.1,
		SeparationAnxiety: 0.2,
	}
	profile.SocialStyle = models.SocialStyleProfile{
		SocialEnergy:     0.7,
		GroupPreference:  0.5,
		DeepTalkPreference: 0.8,
	}
	engine.UpdateRelationshipProfile("identity_001", profile)

	// 获取上下文
	context := engine.GetDecisionContext("identity_001")

	if context == nil {
		t.Fatal("Context should not be nil")
	}

	// 验证上下文字段
	if context.IdentityID != "identity_001" {
		t.Errorf("IdentityID should be identity_001, got %s", context.IdentityID)
	}

	// 验证网络摘要
	if context.NetworkSummary.TotalContacts != 2 {
		t.Errorf("TotalContacts should be 2, got %d", context.NetworkSummary.TotalContacts)
	}
	if context.NetworkSummary.CloseContacts != 1 {
		t.Errorf("CloseContacts should be 1, got %d", context.NetworkSummary.CloseContacts)
	}

	// 验证决策影响
	if context.DecisionInfluence.SocialApproach == 0 {
		t.Error("SocialApproach should be calculated")
	}
	if context.DecisionInfluence.SecurityLevel == 0 {
		t.Error("SecurityLevel should be calculated")
	}

	// 验证社交建议
	if !context.SocialGuidance.ShouldSocialize {
		t.Error("ShouldSocialize should be true when SocialEnergy > 0.4")
	}
	if context.SocialGuidance.SocialEnergyLevel != 0.7 {
		t.Errorf("SocialEnergyLevel should be 0.7, got %f", context.SocialGuidance.SocialEnergyLevel)
	}
}

// TestUpdateAttachmentStyle 测试更新依恋风格
func TestUpdateAttachmentStyle(t *testing.T) {
	engine := NewRelationshipEngine()

	style := models.AttachmentStyleProfile{
		StyleName:       "anxious",
		AnxietyLevel:    0.7,
		AvoidanceLevel:  0.2,
	}

	err := engine.UpdateAttachmentStyle("identity_001", style)
	if err != nil {
		t.Fatalf("UpdateAttachmentStyle failed: %v", err)
	}

	profile := engine.GetRelationshipProfile("identity_001")
	if profile == nil {
		t.Fatal("Profile should exist")
	}
	if profile.AttachmentStyle.StyleName != "anxious" {
		t.Errorf("StyleName should be 'anxious', got %s", profile.AttachmentStyle.StyleName)
	}
}

// TestUpdateSocialStyle 测试更新社交风格
func TestUpdateSocialStyle(t *testing.T) {
	engine := NewRelationshipEngine()

	style := models.SocialStyleProfile{
		SocialEnergy:      0.8,
		GroupPreference:   0.6,
		SmallTalkComfort:  0.7,
	}

	err := engine.UpdateSocialStyle("identity_001", style)
	if err != nil {
		t.Fatalf("UpdateSocialStyle failed: %v", err)
	}

	profile := engine.GetRelationshipProfile("identity_001")
	if profile == nil {
		t.Fatal("Profile should exist")
	}
	if profile.SocialStyle.SocialEnergy != 0.8 {
		t.Errorf("SocialEnergy should be 0.8, got %f", profile.SocialStyle.SocialEnergy)
	}
}

// TestListener 测试监听器完整功能
func TestRelationshipEngine_Listener(t *testing.T) {
	engine := NewRelationshipEngine()

	listener := &testRelationshipListener{
		relationshipAddedCount:    0,
		relationshipUpdatedCount:  0,
		relationshipEndedCount:    0,
		interactionRecordedCount:  0,
		conflictOccurredCount:     0,
	}

	engine.AddListener(listener)

	// 添加关系
	engine.AddRelationship("identity_001", models.Relationship{PersonID: "person_001"})

	if listener.relationshipAddedCount != 1 {
		t.Errorf("relationshipAddedCount should be 1, got %d", listener.relationshipAddedCount)
	}

	// 更新关系
	system := engine.GetRelationshipSystem("identity_001")
	relID := system.Relationships[0].RelationshipID
	engine.UpdateRelationship("identity_001", models.Relationship{RelationshipID: relID})

	if listener.relationshipUpdatedCount != 1 {
		t.Errorf("relationshipUpdatedCount should be 1, got %d", listener.relationshipUpdatedCount)
	}

	// 记录互动
	engine.RecordInteraction("identity_001", models.RelationshipInteraction{RelationshipID: relID})

	if listener.interactionRecordedCount != 1 {
		t.Errorf("interactionRecordedCount should be 1, got %d", listener.interactionRecordedCount)
	}

	// 移除监听器
	engine.RemoveListener(listener)

	// 再次添加关系
	engine.AddRelationship("identity_001", models.Relationship{PersonID: "person_002"})

	if listener.relationshipAddedCount != 1 {
		t.Errorf("relationshipAddedCount should still be 1 after removal, got %d", listener.relationshipAddedCount)
	}
}

// testRelationshipListener 测试监听器
type testRelationshipListener struct {
	relationshipAddedCount    int
	relationshipUpdatedCount  int
	relationshipEndedCount    int
	interactionRecordedCount  int
	conflictOccurredCount     int
}

func (l *testRelationshipListener) OnRelationshipAdded(identityID string, relationship models.Relationship) {
	l.relationshipAddedCount++
}

func (l *testRelationshipListener) OnRelationshipUpdated(identityID string, relationship models.Relationship) {
	l.relationshipUpdatedCount++
}

func (l *testRelationshipListener) OnRelationshipEnded(identityID string, relationshipID string) {
	l.relationshipEndedCount++
}

func (l *testRelationshipListener) OnInteractionRecorded(identityID string, interaction models.RelationshipInteraction) {
	l.interactionRecordedCount++
}

func (l *testRelationshipListener) OnConflictOccurred(identityID string, relationshipID string, conflictLevel float64) {
	l.conflictOccurredCount++
}

// TestRelationshipIsClose 测试关系亲密判断
func TestRelationshipIsClose(t *testing.T) {
	relationship := models.Relationship{
		Intimacy: 0.7,
		Trust:    0.8,
	}

	if !relationship.IsClose() {
		t.Error("Relationship with high intimacy and trust should be close")
	}

	relationship.Intimacy = 0.3
	relationship.Trust = 0.4

	if relationship.IsClose() {
		t.Error("Relationship with low intimacy and trust should not be close")
	}
}

// TestRelationshipIsSupportive 测试关系支持判断
func TestRelationshipIsSupportive(t *testing.T) {
	relationship := models.Relationship{
		SupportGiven:    0.8,
		SupportReceived: 0.7,
	}

	if !relationship.IsSupportive() {
		t.Error("Relationship with high support should be supportive")
	}

	relationship.SupportGiven = 0.2
	relationship.SupportReceived = 0.3

	if relationship.IsSupportive() {
		t.Error("Relationship with low support should not be supportive")
	}
}

// TestRelationshipGetHealth 测试关系健康度计算
func TestRelationshipGetHealth(t *testing.T) {
	relationship := models.Relationship{
		Intimacy:        0.8,
		Trust:           0.9,
		ConflictLevel:   0.2,
		PositiveInteractions: 10,
		NegativeInteractions: 2,
	}

	health := relationship.GetRelationshipHealth()

	if health < 0.7 || health > 0.95 {
		t.Errorf("RelationshipHealth should be between 0.7 and 0.95, got %f", health)
	}
}

// TestSocialNetworkGetOverallHealth 测试社交网络健康度计算
func TestSocialNetworkGetOverallHealth(t *testing.T) {
	network := models.NewSocialNetwork()
	network.TotalContacts = 10
	network.CloseContacts = 3
	network.SupportContacts = 5
	network.StrongTies = 3
	network.WeakTies = 7
	network.SocialCapital = 0.6

	health := network.GetOverallSocialHealth()

	if health < 0.3 || health > 0.8 {
		t.Errorf("OverallSocialHealth should be reasonable, got %f", health)
	}
}