package culture

import (
	"testing"
	"time"

	"github.com/ofa/center/internal/models"
)

// TestRegionalCultureEngineCreation 测试地域文化引擎创建
func TestRegionalCultureEngineCreation(t *testing.T) {
	engine := NewRegionalCultureEngine()

	if engine == nil {
		t.Fatal("RegionalCultureEngine should not be nil")
	}
	if engine.regionalCultures == nil {
		t.Error("regionalCultures map should be initialized")
	}
	if engine.cultureProfiles == nil {
		t.Error("cultureProfiles map should be initialized")
	}
}

// TestGetOrCreateRegionalCulture 测试获取或创建地域文化
func TestGetOrCreateRegionalCulture(t *testing.T) {
	engine := NewRegionalCultureEngine()

	// 创建新地域文化
	culture := engine.GetOrCreateRegionalCulture("identity_001")
	if culture == nil {
		t.Fatal("RegionalCulture should not be nil")
	}

	// 验证默认值
	if culture.Collectivism != 0.5 {
		t.Errorf("Default Collectivism should be 0.5, got %f", culture.Collectivism)
	}

	// 获取已存在的文化
	culture2 := engine.GetOrCreateRegionalCulture("identity_001")
	if culture != culture2 {
		t.Error("Should return same culture instance")
	}
}

// TestUpdateRegionalCulture 测试更新地域文化
func TestRegionalCultureEngine_Update(t *testing.T) {
	engine := NewRegionalCultureEngine()

	culture := models.NewRegionalCulture()
	culture.Province = "广东"
	culture.City = "广州"
	culture.Collectivism = 0.7
	culture.TraditionOriented = 0.6

	err := engine.UpdateRegionalCulture("identity_001", culture)
	if err != nil {
		t.Fatalf("UpdateRegionalCulture failed: %v", err)
	}

	// 验证更新
	retrieved := engine.GetRegionalCulture("identity_001")
	if retrieved == nil {
		t.Fatal("RegionalCulture should exist after update")
	}
	if retrieved.Province != "广东" {
		t.Errorf("Province should be '广东', got %s", retrieved.Province)
	}
	if retrieved.Collectivism != 0.7 {
		t.Errorf("Collectivism should be 0.7, got %f", retrieved.Collectivism)
	}
}

// TestSetLocation 测试设置地理位置
func TestSetLocation(t *testing.T) {
	engine := NewRegionalCultureEngine()

	err := engine.SetLocation("identity_001", "浙江", "杭州", "一线城市")
	if err != nil {
		t.Fatalf("SetLocation failed: %v", err)
	}

	culture := engine.GetRegionalCulture("identity_001")
	if culture == nil {
		t.Fatal("Culture should exist")
	}
	if culture.Province != "浙江" {
		t.Errorf("Province should be '浙江', got %s", culture.Province)
	}
	if culture.City != "杭州" {
		t.Errorf("City should be '杭州', got %s", culture.City)
	}
	if culture.CityTier != "一线城市" {
		t.Errorf("CityTier should be '一线城市', got %s", culture.CityTier)
	}

	// 验证预设文化应用
	if culture.Collectivism == 0.5 {
		t.Error("Collectivism should be updated by preset for '浙江'")
	}
}

// TestSetDialect 测试设置方言
func TestSetDialect(t *testing.T) {
	engine := NewRegionalCultureEngine()

	err := engine.SetDialect("identity_001", "粤语", 0.8)
	if err != nil {
		t.Fatalf("SetDialect failed: %v", err)
	}

	culture := engine.GetRegionalCulture("identity_001")
	if culture.Dialect != "粤语" {
		t.Errorf("Dialect should be '粤语', got %s", culture.Dialect)
	}
	if culture.DialectProficiency != 0.8 {
		t.Errorf("DialectProficiency should be 0.8, got %f", culture.DialectProficiency)
	}
}

// TestSetCulturalDimensions 测试设置文化维度
func TestSetCulturalDimensions(t *testing.T) {
	engine := NewRegionalCultureEngine()

	dimensions := map[string]float64{
		"collectivism":        0.7,
		"tradition_oriented":  0.6,
		"innovation_oriented": 0.8,
		"power_distance":      0.5,
		"expression_level":    0.7,
	}

	err := engine.SetCulturalDimensions("identity_001", dimensions)
	if err != nil {
		t.Fatalf("SetCulturalDimensions failed: %v", err)
	}

	culture := engine.GetRegionalCulture("identity_001")
	if culture.Collectivism != 0.7 {
		t.Errorf("Collectivism should be 0.7, got %f", culture.Collectivism)
	}
	if culture.TraditionOriented != 0.6 {
		t.Errorf("TraditionOriented should be 0.6, got %f", culture.TraditionOriented)
	}
	if culture.InnovationOriented != 0.8 {
		t.Errorf("InnovationOriented should be 0.8, got %f", culture.InnovationOriented)
	}
}

// TestSetCommunicationStyle 测试设置沟通风格
func TestSetCommunicationStyle(t *testing.T) {
	engine := NewRegionalCultureEngine()

	err := engine.SetCommunicationStyle("identity_001", "委婉")
	if err != nil {
		t.Fatalf("SetCommunicationStyle failed: %v", err)
	}

	culture := engine.GetRegionalCulture("identity_001")
	if culture.CommunicationStyle != "委婉" {
		t.Errorf("CommunicationStyle should be '委婉', got %s", culture.CommunicationStyle)
	}
}

// TestSetSocialStyle 测试设置社交风格
func TestSetSocialStyle(t *testing.T) {
	engine := NewRegionalCultureEngine()

	err := engine.SetSocialStyle("identity_001", "内敛")
	if err != nil {
		t.Fatalf("SetSocialStyle failed: %v", err)
	}

	culture := engine.GetRegionalCulture("identity_001")
	if culture.SocialStyle != "内敛" {
		t.Errorf("SocialStyle should be '内敛', got %s", culture.SocialStyle)
	}
}

// TestAddRegionalTrait 测试添加地域性格特征
func TestAddRegionalTrait(t *testing.T) {
	engine := NewRegionalCultureEngine()

	err := engine.AddRegionalTrait("identity_001", "务实")
	if err != nil {
		t.Fatalf("AddRegionalTrait failed: %v", err)
	}

	culture := engine.GetRegionalCulture("identity_001")
	if len(culture.RegionalTraits) != 1 {
		t.Errorf("Should have 1 trait, got %d", len(culture.RegionalTraits))
	}

	// 添加重复特征
	err = engine.AddRegionalTrait("identity_001", "务实")
	if err != nil {
		t.Fatalf("AddRegionalTrait (duplicate) failed: %v", err)
	}

	culture = engine.GetRegionalCulture("identity_001")
	if len(culture.RegionalTraits) != 1 {
		t.Errorf("Should still have 1 trait (no duplicates), got %d", len(culture.RegionalTraits))
	}

	// 添加新特征
	err = engine.AddRegionalTrait("identity_001", "开放")
	if err != nil {
		t.Fatalf("AddRegionalTrait failed: %v", err)
	}

	culture = engine.GetRegionalCulture("identity_001")
	if len(culture.RegionalTraits) != 2 {
		t.Errorf("Should have 2 traits, got %d", len(culture.RegionalTraits))
	}
}

// TestAddCustom 测试添加习俗
func TestAddCustom(t *testing.T) {
	engine := NewRegionalCultureEngine()

	err := engine.AddCustom("identity_001", "喝茶")
	if err != nil {
		t.Fatalf("AddCustom failed: %v", err)
	}

	culture := engine.GetRegionalCulture("identity_001")
	if len(culture.Customs) != 1 {
		t.Errorf("Should have 1 custom, got %d", len(culture.Customs))
	}
}

// TestAddMigration 测试添加迁移记录
func TestAddMigration(t *testing.T) {
	engine := NewRegionalCultureEngine()

	// 先设置初始位置
	engine.SetLocation("identity_001", "广东", "广州", "一线城市")

	migration := models.Migration{
		MigrationID:    "migration_001",
		FromProvince:   "广东",
		FromCity:       "广州",
		ToProvince:     "北京",
		ToCity:         "北京",
		MigrationDate:  time.Now(),
		Reason:         "工作",
		AdaptationTime: 6,
	}

	err := engine.AddMigration("identity_001", migration)
	if err != nil {
		t.Fatalf("AddMigration failed: %v", err)
	}

	// 验证迁移历史
	culture := engine.GetRegionalCulture("identity_001")
	if len(culture.MigrationHistory) != 1 {
		t.Errorf("Should have 1 migration, got %d", len(culture.MigrationHistory))
	}

	// 文化适应能力应增加
	if culture.CulturalAdaptation < 0.5 {
		t.Errorf("CulturalAdaptation should increase after migration, got %f", culture.CulturalAdaptation)
	}
}

// TestRecordCulturalAdaptation 测试记录文化适应
func TestRecordCulturalAdaptation(t *testing.T) {
	engine := NewRegionalCultureEngine()

	adaptation := models.CulturalAdaptation{
		AdaptationID:   "adaptation_001",
		SourceCulture:  "南方",
		TargetCulture:  "北方",
		AdaptationType: "饮食习惯",
		Success:        0.8,
		Challenges:     []string{"口味差异"},
	}

	err := engine.RecordCulturalAdaptation("identity_001", adaptation)
	if err != nil {
		t.Fatalf("RecordCulturalAdaptation failed: %v", err)
	}

	// 验证适应历史
	profile := engine.GetCultureProfile("identity_001")
	if profile == nil {
		t.Fatal("Profile should exist")
	}
	if len(profile.AdaptationHistory) != 1 {
		t.Errorf("Should have 1 adaptation record, got %d", len(profile.AdaptationHistory))
	}

	// 文化适应能力应增加
	culture := engine.GetRegionalCulture("identity_001")
	if culture.CulturalAdaptation < 0.5 {
		t.Errorf("CulturalAdaptation should increase, got %f", culture.CulturalAdaptation)
	}
}

// TestGetDecisionContext 测试获取决策上下文
func TestGetDecisionContext(t *testing.T) {
	engine := NewRegionalCultureEngine()

	// 设置完整地域文化
	culture := engine.GetOrCreateRegionalCulture("identity_001")
	culture.Province = "浙江"
	culture.City = "杭州"
	culture.CityTier = "一线城市"
	culture.Collectivism = 0.6
	culture.TraditionOriented = 0.5
	culture.InnovationOriented = 0.7
	culture.CommunicationStyle = "直接"
	culture.SocialStyle = "开放"
	engine.UpdateRegionalCulture("identity_001", culture)

	// 获取上下文
	context := engine.GetDecisionContext("identity_001")

	if context == nil {
		t.Fatal("Context should not be nil")
	}

	// 验证上下文字段
	if context.IdentityID != "identity_001" {
		t.Errorf("IdentityID should be identity_001, got %s", context.IdentityID)
	}
	if context.InfluenceFactors == nil {
		t.Error("InfluenceFactors should be calculated")
	}

	// 验证文化倾向
	if context.CulturalTendencies.GroupDecisionPreference == 0 {
		t.Error("GroupDecisionPreference should be calculated")
	}
	if context.CulturalTendencies.TraditionRespect == 0 {
		t.Error("TraditionRespect should be calculated")
	}
	if context.CulturalTendencies.Directness == 0 {
		t.Error("Directness should be calculated")
	}

	// 验证文化背景
	if context.CulturalBackground.Province != "浙江" {
		t.Errorf("Province should be '浙江', got %s", context.CulturalBackground.Province)
	}
	if context.CulturalBackground.City != "杭州" {
		t.Errorf("City should be '杭州', got %s", context.CulturalBackground.City)
	}
}

// TestCompareCultures 测试比较文化差异
func TestCompareCultures(t *testing.T) {
	engine := NewRegionalCultureEngine()

	// 设置两个身份的文化
	engine.SetLocation("identity_001", "浙江", "杭州", "一线城市")
	engine.SetLocation("identity_002", "四川", "成都", "新一线城市")

	comparison := engine.CompareCultures("identity_001", "identity_002")

	if comparison == nil {
		t.Fatal("Comparison should not be nil")
	}

	// 验证比较字段
	if comparison.Identity1 != "identity_001" {
		t.Errorf("Identity1 should be identity_001, got %s", comparison.Identity1)
	}
	if comparison.Identity2 != "identity_002" {
		t.Errorf("Identity2 should be identity_002, got %s", comparison.Identity2)
	}

	// 验证兼容性计算
	if comparison.Compatibility < 0 || comparison.Compatibility > 1 {
		t.Errorf("Compatibility should be between 0 and 1, got %f", comparison.Compatibility)
	}
}

// TestListener 测试监听器
func TestRegionalCultureEngine_Listener(t *testing.T) {
	engine := NewRegionalCultureEngine()

	// 创建测试监听器
	listener := &testCultureListener{
		cultureChangedCount:      0,
		migrationAddedCount:      0,
		adaptationRecordedCount:  0,
	}

	engine.AddListener(listener)

	// 更新文化
	culture := models.NewRegionalCulture()
	culture.Province = "广东"
	engine.UpdateRegionalCulture("identity_001", culture)

	if listener.cultureChangedCount != 1 {
		t.Errorf("cultureChangedCount should be 1, got %d", listener.cultureChangedCount)
	}

	// 添加迁移
	migration := models.Migration{
		MigrationID:   "migration_001",
		FromProvince:  "广东",
		ToProvince:    "北京",
		MigrationDate: time.Now(),
	}
	engine.AddMigration("identity_001", migration)

	if listener.migrationAddedCount != 1 {
		t.Errorf("migrationAddedCount should be 1, got %d", listener.migrationAddedCount)
	}

	// 记录适应
	adaptation := models.CulturalAdaptation{
		AdaptationID:  "adaptation_001",
		Success:       0.8,
	}
	engine.RecordCulturalAdaptation("identity_001", adaptation)

	if listener.adaptationRecordedCount != 1 {
		t.Errorf("adaptationRecordedCount should be 1, got %d", listener.adaptationRecordedCount)
	}

	// 移除监听器
	engine.RemoveListener(listener)

	// 再次更新
	engine.UpdateRegionalCulture("identity_001", culture)

	if listener.cultureChangedCount != 1 {
		t.Errorf("cultureChangedCount should still be 1 after removal, got %d", listener.cultureChangedCount)
	}
}

// testCultureListener 测试监听器
type testCultureListener struct {
	cultureChangedCount     int
	migrationAddedCount     int
	adaptationRecordedCount int
}

func (l *testCultureListener) OnRegionalCultureChanged(identityID string, culture *models.RegionalCulture) {
	l.cultureChangedCount++
}

func (l *testCultureListener) OnMigrationAdded(identityID string, migration models.Migration) {
	l.migrationAddedCount++
}

func (l *testCultureListener) OnCulturalAdaptationRecorded(identityID string, adaptation models.CulturalAdaptation) {
	l.adaptationRecordedCount++
}

// TestRegionalCultureNormalize 测试地域文化归一化
func TestRegionalCultureNormalize(t *testing.T) {
	culture := models.NewRegionalCulture()
	culture.Collectivism = 1.5
	culture.TraditionOriented = -0.2

	culture.Normalize()

	if culture.Collectivism != 1.0 {
		t.Errorf("Collectivism should be normalized to 1.0, got %f", culture.Collectivism)
	}
	if culture.TraditionOriented != 0 {
		t.Errorf("TraditionOriented should be normalized to 0, got %f", culture.TraditionOriented)
	}
}

// TestIsMetropolitan 测试是否是大城市
func TestIsMetropolitan(t *testing.T) {
	culture := models.NewRegionalCulture()
	culture.CityTier = "一线城市"

	if !culture.IsMetropolitan() {
		t.Error("一线城市 should be metropolitan")
	}

	culture.CityTier = "二线城市"
	if culture.IsMetropolitan() {
		t.Error("二线城市 should not be metropolitan")
	}
}

// TestHasMigrationHistory 测试是否有迁移历史
func TestHasMigrationHistory(t *testing.T) {
	culture := models.NewRegionalCulture()
	culture.Native = true

	if culture.HasMigrationHistory() {
		t.Error("Native identity should not have migration history")
	}

	culture.Native = false
	culture.MigrationHistory = []models.Migration{
		{MigrationID: "migration_001"},
	}

	if !culture.HasMigrationHistory() {
		t.Error("Should have migration history")
	}
}