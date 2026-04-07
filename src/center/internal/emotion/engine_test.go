package emotion

import (
	"context"
	"testing"
	"time"

	"github.com/taomic2035/OFA/src/center/internal/models"
)

// TestEmotionEngineCreation 测试情绪引擎创建
func TestEmotionEngineCreation(t *testing.T) {
	config := EmotionConfig{
		DefaultDecayRate: 0.02,
	}
	engine := NewEmotionEngine(config)

	if engine == nil {
		t.Fatal("EmotionEngine should not be nil")
	}
	if engine.config.DefaultDecayRate != 0.02 {
		t.Errorf("DefaultDecayRate should be 0.02, got %f", engine.config.DefaultDecayRate)
	}
}

// TestGetOrCreateEmotion 测试获取或创建情绪
func TestGetOrCreateEmotion(t *testing.T) {
	engine := NewEmotionEngine(EmotionConfig{})

	// 创建新情绪
	emotion := engine.GetOrCreateEmotion("identity_001")
	if emotion == nil {
		t.Fatal("Emotion should not be nil")
	}

	// 验证默认值
	if emotion.Joy != 0.5 {
		t.Errorf("Default Joy should be 0.5, got %f", emotion.Joy)
	}
	if emotion.CurrentMood != "neutral" {
		t.Errorf("Default mood should be neutral, got %s", emotion.CurrentMood)
	}

	// 获取已存在的情绪
	emotion2 := engine.GetOrCreateEmotion("identity_001")
	if emotion != emotion2 {
		t.Error("Should return same emotion instance")
	}
}

// TestTriggerEmotion 测试触发情绪
func TestTriggerEmotion(t *testing.T) {
	engine := NewEmotionEngine(EmotionConfig{})

	trigger := models.EmotionTrigger{
		TriggerID:   "trigger_001",
		TriggerType: "event",
		TriggerDesc: "获得成就",
		EmotionType: "joy",
		Intensity:   0.5,
		Duration:    30,
		Timestamp:   time.Now(),
	}

	emotion, err := engine.TriggerEmotion(context.Background(), "identity_001", trigger)
	if err != nil {
		t.Fatalf("TriggerEmotion failed: %v", err)
	}

	// 验证情绪增加
	if emotion.Joy < 0.5 {
		t.Errorf("Joy should be >= 0.5 after trigger, got %f", emotion.Joy)
	}

	// 验证心境更新
	if emotion.CurrentMood != "happy" && emotion.CurrentMood != "excited" {
		t.Errorf("Mood should be happy or excited, got %s", emotion.CurrentMood)
	}

	// 验证触发历史
	if len(emotion.RecentTriggers) == 0 {
		t.Error("RecentTriggers should not be empty")
	}
}

// TestTriggerEmotionByEvent 测试事件触发情绪
func TestTriggerEmotionByEvent(t *testing.T) {
	engine := NewEmotionEngine(EmotionConfig{})

	emotion, err := engine.TriggerEmotionByEvent(context.Background(), "identity_001", "achievement", "完成项目", nil)
	if err != nil {
		t.Fatalf("TriggerEmotionByEvent failed: %v", err)
	}

	// achievement 应触发 joy
	if emotion.Joy < 0.5 {
		t.Errorf("Achievement should increase Joy, got %f", emotion.Joy)
	}

	// 测试负面事件
	emotion2, err := engine.TriggerEmotionByEvent(context.Background(), "identity_001", "failure", "项目失败", nil)
	if err != nil {
		t.Fatalf("TriggerEmotionByEvent for failure failed: %v", err)
	}

	// failure 应触发 sadness
	if emotion2.Sadness < 0.3 {
		t.Errorf("Failure should increase Sadness, got %f", emotion2.Sadness)
	}
}

// TestDecayEmotion 测试情绪衰减
func TestDecayEmotion(t *testing.T) {
	engine := NewEmotionEngine(EmotionConfig{
		DefaultDecayRate: 0.1, // 更快的衰减便于测试
	})

	// 先触发一个强情绪
	trigger := models.EmotionTrigger{
		TriggerID:   "trigger_001",
		TriggerType: "event",
		TriggerDesc: "好消息",
		EmotionType: "joy",
		Intensity:   0.8,
		Duration:    30,
		Timestamp:   time.Now(),
	}

	emotion, _ := engine.TriggerEmotion(context.Background(), "identity_001", trigger)
	initialJoy := emotion.Joy

	// 衰减
	emotion, _ = engine.DecayEmotion("identity_001", 30)

	// 验证衰减
	if emotion.Joy >= initialJoy {
		t.Errorf("Joy should decrease after decay, initial=%f, after=%f", initialJoy, emotion.Joy)
	}

	// 验证持续时间增加
	if emotion.Duration < 30 {
		t.Errorf("Duration should increase, got %d", emotion.Duration)
	}
}

// TestPropagateEmotion 测试情绪传播
func TestPropagateEmotion(t *testing.T) {
	engine := NewEmotionEngine(EmotionConfig{})

	// 创建情绪状态
	emotion := engine.GetOrCreateEmotion("identity_001")
	emotion.Joy = 0.8
	emotion.Fear = 0.1

	// 获取传播影响
	influence := engine.PropagateEmotion("identity_001")

	if influence == nil {
		t.Fatal("Influence should not be nil")
	}

	// 验证风险偏好
	if influence["risk_tolerance"] < 0.5 {
		t.Errorf("High Joy should increase risk_tolerance, got %f", influence["risk_tolerance"])
	}

	// 验证社交倾向
	if influence["social_tendency"] < 0.2 {
		t.Errorf("High Joy and Love should increase social_tendency, got %f", influence["social_tendency"])
	}
}

// TestGetEmotionContext 测试获取情绪决策上下文
func TestGetEmotionContext(t *testing.T) {
	engine := NewEmotionEngine(EmotionConfig{})

	// 创建情绪和欲望
	engine.GetOrCreateEmotion("identity_001")
	engine.GetOrCreateDesire("identity_001")

	// 触发情绪
	trigger := models.EmotionTrigger{
		TriggerID:   "trigger_001",
		TriggerType: "event",
		TriggerDesc: "测试事件",
		EmotionType: "joy",
		Intensity:   0.6,
		Duration:    30,
		Timestamp:   time.Now(),
	}
	engine.TriggerEmotion(context.Background(), "identity_001", trigger)

	// 获取上下文
	context := engine.GetEmotionContext("identity_001")

	if context == nil {
		t.Fatal("Context should not be nil")
	}

	// 验证上下文字段
	if context.IdentityID != "identity_001" {
		t.Errorf("IdentityID should be identity_001, got %s", context.IdentityID)
	}
	if context.DominantEmotion != "joy" {
		t.Errorf("DominantEmotion should be joy, got %s", context.DominantEmotion)
	}
	if context.InfluenceFactors == nil {
		t.Error("InfluenceFactors should not be nil")
	}
	if context.SatisfactionLevel == 0 {
		t.Error("SatisfactionLevel should be calculated")
	}
}

// TestDesireManagement 测试欲望管理
func TestDesireManagement(t *testing.T) {
	engine := NewEmotionEngine(EmotionConfig{})

	// 创建欲望
	desire := engine.GetOrCreateDesire("identity_001")
	if desire == nil {
		t.Fatal("Desire should not be nil")
	}

	// 验证默认值
	if desire.Physiological != 0.7 {
		t.Errorf("Default Physiological should be 0.7, got %f", desire.Physiological)
	}

	// 触发欲望
	desire, _ = engine.TriggerDesire("identity_001", "esteem", 0.5, "获得认可")
	if desire.PrimaryDesire != "esteem" {
		t.Errorf("PrimaryDesire should be esteem, got %s", desire.PrimaryDesire)
	}

	// 满足欲望
	desire, _ = engine.SatisfyDesire("identity_001", "esteem", 0.4)
	if desire.Esteem < 0.4 {
		t.Errorf("Esteem should increase after satisfaction, got %f", desire.Esteem)
	}
}

// TestDesireDecay 测试欲望衰减
func TestDesireDecay(t *testing.T) {
	engine := NewEmotionEngine(EmotionConfig{})

	// 创建欲望
	desire := engine.GetOrCreateDesire("identity_001")
	desire.Physiological = 0.8
	desire.Safety = 0.7

	initialPhysio := desire.Physiological

	// 衰减
	desire, _ = engine.DecayDesire("identity_001", 30)

	// 验证衰减
	if desire.Physiological >= initialPhysio {
		t.Errorf("Physiological should decrease after decay, initial=%f, after=%f", initialPhysio, desire.Physiological)
	}
}

// TestEmotionProfile 测试情绪画像
func TestEmotionProfile(t *testing.T) {
	engine := NewEmotionEngine(EmotionConfig{})

	// 获取画像
	profile := engine.GetEmotionProfile("identity_001")
	if profile == nil {
		t.Fatal("Profile should be created with emotion")
	}

	// 验证默认触发敏感度
	if profile.TriggerSensitivity["positive_event"] != 0.5 {
		t.Errorf("TriggerSensitivity for positive_event should be 0.5")
	}

	// 更新画像
	newProfile := models.NewEmotionProfile("identity_001")
	newProfile.BaselineJoy = 0.7
	newProfile.EmotionalRange = 0.4

	err := engine.UpdateEmotionProfile("identity_001", newProfile)
	if err != nil {
		t.Fatalf("UpdateEmotionProfile failed: %v", err)
	}

	// 验证更新
	updated := engine.GetEmotionProfile("identity_001")
	if updated.BaselineJoy != 0.7 {
		t.Errorf("BaselineJoy should be 0.7, got %f", updated.BaselineJoy)
	}
}

// TestEmotionHistory 测试情绪历史
func TestEmotionHistory(t *testing.T) {
	engine := NewEmotionEngine(EmotionConfig{
		MaxHistoryStates: 10,
	})

	// 触发多次情绪
	for i := 0; i < 15; i++ {
		trigger := models.EmotionTrigger{
			TriggerID:   "trigger_" + string(rune(i)),
			TriggerType: "event",
			TriggerDesc: "事件",
			EmotionType: "joy",
			Intensity:   0.3,
			Duration:    10,
			Timestamp:   time.Now(),
		}
		engine.TriggerEmotion(context.Background(), "identity_001", trigger)
	}

	// 获取历史
	history := engine.GetEmotionHistory("identity_001", 5)

	// 验证限制
	if len(history) > 10 {
		t.Errorf("History should be limited to MaxHistoryStates, got %d", len(history))
	}
	if len(history) < 5 {
		t.Errorf("Should return at least 5 states, got %d", len(history))
	}
}

// TestListener 测试监听器
func TestListener(t *testing.T) {
	engine := NewEmotionEngine(EmotionConfig{})

	// 创建测试监听器
	listener := &testListener{
		triggeredCount: 0,
		decayedCount:   0,
		moodChangedCount: 0,
		desireChangedCount: 0,
	}

	engine.AddListener(listener)

	// 触发情绪
	trigger := models.EmotionTrigger{
		TriggerID:   "trigger_001",
		TriggerType: "event",
		TriggerDesc: "测试",
		EmotionType: "joy",
		Intensity:   0.6,
		Duration:    30,
		Timestamp:   time.Now(),
	}
	engine.TriggerEmotion(context.Background(), "identity_001", trigger)

	// 验证监听器被调用
	if listener.triggeredCount != 1 {
		t.Errorf("triggeredCount should be 1, got %d", listener.triggeredCount)
	}

	// 衰减
	engine.DecayEmotion("identity_001", 30)

	if listener.decayedCount != 1 {
		t.Errorf("decayedCount should be 1, got %d", listener.decayedCount)
	}

	// 触发欲望
	engine.TriggerDesire("identity_001", "esteem", 0.5, "测试")

	if listener.desireChangedCount != 1 {
		t.Errorf("desireChangedCount should be 1, got %d", listener.desireChangedCount)
	}

	// 移除监听器
	engine.RemoveListener(listener)

	// 再次触发
	engine.TriggerEmotion(context.Background(), "identity_001", trigger)

	if listener.triggeredCount != 1 {
		t.Errorf("triggeredCount should still be 1 after removal, got %d", listener.triggeredCount)
	}
}

// testListener 测试监听器
type testListener struct {
	triggeredCount    int
	decayedCount      int
	moodChangedCount  int
	desireChangedCount int
}

func (l *testListener) OnEmotionTriggered(identityID string, emotion *models.Emotion, trigger models.EmotionTrigger) {
	l.triggeredCount++
}

func (l *testListener) OnEmotionDecayed(identityID string, emotion *models.Emotion) {
	l.decayedCount++
}

func (l *testListener) OnMoodChanged(identityID string, oldMood, newMood string) {
	l.moodChangedCount++
}

func (l *testListener) OnDesireChanged(identityID string, desire *models.Desire) {
	l.desireChangedCount++
}

// TestEmotionNormalize 测试情绪归一化
func TestEmotionNormalize(t *testing.T) {
	emotion := models.NewEmotion()
	emotion.Joy = 1.5  // 超出范围
	emotion.Anger = -0.3 // 低于范围

	emotion.Normalize()

	if emotion.Joy != 1.0 {
		t.Errorf("Joy should be normalized to 1.0, got %f", emotion.Joy)
	}
	if emotion.Anger != 0 {
		t.Errorf("Anger should be normalized to 0, got %f", emotion.Anger)
	}
}

// TestGetDominantEmotion 测试获取主导情绪
func TestGetDominantEmotion(t *testing.T) {
	emotion := models.NewEmotion()
	emotion.Joy = 0.3
	emotion.Anger = 0.7
	emotion.Sadness = 0.2

	dominant := emotion.GetDominantEmotion()

	if dominant != "anger" {
		t.Errorf("Dominant emotion should be anger, got %s", dominant)
	}
}

// TestDesireSatisfactionCalculation 测试欲望满足度计算
func TestDesireSatisfactionCalculation(t *testing.T) {
	desire := models.NewDesire()
	desire.Physiological = 0.9
	desire.Safety = 0.8
	desire.LoveBelonging = 0.5
	desire.Esteem = 0.3
	desire.SelfActualization = 0.1

	// 计算整体满足度
	satisfaction := desire.CalculateOverallSatisfaction()

	if satisfaction < 0.5 || satisfaction > 0.8 {
		t.Errorf("Overall satisfaction should be between 0.5 and 0.8, got %f", satisfaction)
	}

	// 计算挫折程度
	frustration := desire.CalculateFrustrationLevel()

	if frustration < 0.2 {
		t.Errorf("Frustration should be significant due to low Esteem, got %f", frustration)
	}

	// 获取主导需求
	dominant := desire.GetDominantNeed()

	if dominant.Name != "自我实现" && dominant.Name != "尊重需求" {
		t.Errorf("Dominant need should be self_actualization or esteem, got %s", dominant.Name)
	}
}