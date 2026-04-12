// health_test.go
// OFA Center Health Check Tests (v9.3.0)

package health

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewHealthChecker(t *testing.T) {
	hc := NewHealthChecker("v9.3.0")
	if hc == nil {
		t.Fatal("HealthChecker should not be nil")
	}
	if hc.version != "v9.3.0" {
		t.Errorf("Version should be v9.3.0, got %s", hc.version)
	}
	if len(hc.components) != 0 {
		t.Errorf("Initial components should be empty")
	}
}

func TestRegisterComponent(t *testing.T) {
	hc := NewHealthChecker("v9.3.0")
	checker := NewDatabaseHealthChecker("postgres", func(ctx context.Context) error {
		return nil
	})
	hc.RegisterComponent(checker)

	if len(hc.components) != 1 {
		t.Errorf("Should have 1 component")
	}
	if _, ok := hc.components["postgres"]; !ok {
		t.Errorf("Should have postgres component")
	}
}

func TestCheckAllHealthy(t *testing.T) {
	hc := NewHealthChecker("v9.3.0")

	hc.RegisterComponent(NewDatabaseHealthChecker("postgres", func(ctx context.Context) error {
		return nil
	}))
	hc.RegisterComponent(NewRedisHealthChecker(func(ctx context.Context) error {
		return nil
	}))

	result := hc.Check(context.Background())

	if result.Status != HealthStatusHealthy {
		t.Errorf("Should be healthy, got %s", result.Status)
	}
	if len(result.Components) != 2 {
		t.Errorf("Should have 2 components")
	}
}

func TestCheckWithUnhealthy(t *testing.T) {
	hc := NewHealthChecker("v9.3.0")

	hc.RegisterComponent(NewDatabaseHealthChecker("postgres", func(ctx context.Context) error {
		return errors.New("connection failed")
	}))

	result := hc.Check(context.Background())

	if result.Status != HealthStatusUnhealthy {
		t.Errorf("Should be unhealthy, got %s", result.Status)
	}
}

func TestCheckWithDegraded(t *testing.T) {
	hc := NewHealthChecker("v9.3.0")

	hc.RegisterComponent(NewWebSocketHealthChecker(100, func() int {
		return 95 // 95% usage = degraded
	}))

	result := hc.Check(context.Background())

	if result.Status != HealthStatusDegraded {
		t.Errorf("Should be degraded, got %s", result.Status)
	}
}

func TestSelfHealing(t *testing.T) {
	hc := NewHealthChecker("v9.3.0")

	// 创建一个可以自愈的检查器
	callCount := 0
	hc.RegisterComponent(NewRedisHealthChecker(func(ctx context.Context) error {
		callCount++
		if callCount == 1 {
			return errors.New("first call fails")
		}
		return nil // second call (heal) succeeds
	}))

	result := hc.Check(context.Background())

	// 自愈后应该恢复健康
	for _, comp := range result.Components {
		if comp.Name == "redis" {
			if comp.HealAttempt != 1 {
				t.Errorf("Should have 1 heal attempt")
			}
		}
	}
}

func TestAlertManager(t *testing.T) {
	am := NewAlertManager()

	alertReceived := false
	am.AddListener(func(alert Alert) {
		alertReceived = true
	})

	am.SendAlert(Alert{
		Component: "test",
		Level:     AlertLevelWarning,
		Message:   "Test alert",
		Time:      time.Now(),
	})

	if !alertReceived {
		t.Error("Alert listener should have been called")
	}

	pending := am.GetPendingAlerts()
	if len(pending) != 1 {
		t.Errorf("Should have 1 pending alert")
	}

	am.ResolveAlert("test")
	pending = am.GetPendingAlerts()
	if len(pending) != 0 {
		t.Errorf("Should have 0 pending alerts after resolve")
	}
}

func TestDegradationManager(t *testing.T) {
	dm := NewDegradationManager()

	strategy := dm.GetStrategy(HealthStatusDegraded)
	if strategy == nil {
		t.Fatal("Degraded strategy should not be nil")
	}
	if strategy.Name != "degraded_mode" {
		t.Errorf("Strategy name should be degraded_mode")
	}

	strategy = dm.GetStrategy(HealthStatusUnhealthy)
	if strategy == nil {
		t.Fatal("Unhealthy strategy should not be nil")
	}
	if strategy.Name != "emergency_mode" {
		t.Errorf("Strategy name should be emergency_mode")
	}
}

func TestDatabaseHealthChecker(t *testing.T) {
	checker := NewDatabaseHealthChecker("postgres", func(ctx context.Context) error {
		return nil
	})

	if checker.Name() != "postgres" {
		t.Errorf("Name should be postgres")
	}
	if checker.IsHealable() {
		t.Error("Database should not be healable")
	}

	health, err := checker.Check(context.Background())
	if err != nil {
		t.Errorf("Check should not error")
	}
	if health.Status != HealthStatusHealthy {
		t.Errorf("Should be healthy")
	}
}

func TestRedisHealthChecker(t *testing.T) {
	checker := NewRedisHealthChecker(func(ctx context.Context) error {
		return nil
	})

	if checker.Name() != "redis" {
		t.Errorf("Name should be redis")
	}
	if !checker.IsHealable() {
		t.Error("Redis should be healable")
	}

	health, err := checker.Check(context.Background())
	if err != nil {
		t.Errorf("Check should not error")
	}
	if health.Status != HealthStatusHealthy {
		t.Errorf("Should be healthy")
	}
}

func TestWebSocketHealthChecker(t *testing.T) {
	checker := NewWebSocketHealthChecker(100, func() int {
		return 50
	})

	health, err := checker.Check(context.Background())
	if err != nil {
		t.Errorf("Check should not error")
	}
	if health.Status != HealthStatusHealthy {
		t.Errorf("50% usage should be healthy")
	}

	checker2 := NewWebSocketHealthChecker(100, func() int {
		return 91
	})
	health2, _ := checker2.Check(context.Background())
	if health2.Status != HealthStatusDegraded {
		t.Errorf("91% usage should be degraded")
	}
}

func TestMemoryHealthChecker(t *testing.T) {
	checker := NewMemoryHealthChecker(1000, func() int {
		return 500
	})

	health, err := checker.Check(context.Background())
	if err != nil {
		t.Errorf("Check should not error")
	}
	if health.Status != HealthStatusHealthy {
		t.Errorf("50% memory should be healthy")
	}

	checker2 := NewMemoryHealthChecker(1000, func() int {
		return 850
	})
	health2, _ := checker2.Check(context.Background())
	if health2.Status != HealthStatusDegraded {
		t.Errorf("85% memory should be degraded")
	}

	checker3 := NewMemoryHealthChecker(1000, func() int {
		return 960
	})
	health3, _ := checker3.Check(context.Background())
	if health3.Status != HealthStatusUnhealthy {
		t.Errorf("96% memory should be unhealthy")
	}
}

func TestLLMHealthChecker(t *testing.T) {
	checker := NewLLMHealthChecker(func(ctx context.Context) error {
		return nil
	})

	health, err := checker.Check(context.Background())
	if err != nil {
		t.Errorf("Check should not error")
	}
	if health.Status != HealthStatusHealthy {
		t.Errorf("Should be healthy")
	}

	checker2 := NewLLMHealthChecker(func(ctx context.Context) error {
		return errors.New("LLM unavailable")
	})
	health2, _ := checker2.Check(context.Background())
	if health2.Status != HealthStatusDegraded {
		t.Errorf("LLM unavailable should be degraded (not unhealthy)")
	}
}

func TestHealthCheckResult(t *testing.T) {
	hc := NewHealthChecker("v9.3.0")
	result := hc.Check(context.Background())

	if result.Version != "v9.3.0" {
		t.Errorf("Version should be v9.3.0")
	}
	if result.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
	if result.Uptime < 0 {
		t.Error("Uptime should be positive")
	}
}