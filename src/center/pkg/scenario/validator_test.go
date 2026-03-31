// Package scenario - 场景集成测试
package scenario

import (
	"context"
	"testing"
	"time"
)

func TestAllScenarios(t *testing.T) {
	config := ScenarioConfig{
		MockMode:     true,
		Timeout:      30 * time.Second,
		TestDuration: 10 * time.Second,
	}

	validator := NewScenarioValidator(config)
	ctx := context.Background()

	report := validator.RunAllScenarios(ctx)

	if report.Summary.ScenariosTotal != 4 {
		t.Errorf("预期4个场景, 实际: %d", report.Summary.ScenariosTotal)
	}

	if report.Summary.PassRate != 100.0 {
		t.Errorf("预期通过率100%%, 实际: %.2f%%", report.Summary.PassRate)
	}

	t.Log(report.PrintReport())
}

func TestCrossDeviceScenario(t *testing.T) {
	config := ScenarioConfig{MockMode: true}
	validator := NewScenarioValidator(config)

	result, err := validator.RunScenario(context.Background(), "cross_device")
	if err != nil {
		t.Fatalf("运行失败: %v", err)
	}

	if result.Status != "passed" {
		t.Errorf("预期通过, 实际: %s", result.Status)
	}
}

func TestSmartHomeScenario(t *testing.T) {
	config := ScenarioConfig{MockMode: true}
	validator := NewScenarioValidator(config)

	result, err := validator.RunScenario(context.Background(), "smart_home")
	if err != nil {
		t.Fatalf("运行失败: %v", err)
	}

	if result.Status != "passed" {
		t.Errorf("预期通过, 实际: %s", result.Status)
	}
}

func TestDistributedAIScenario(t *testing.T) {
	config := ScenarioConfig{MockMode: true}
	validator := NewScenarioValidator(config)

	result, err := validator.RunScenario(context.Background(), "distributed_ai")
	if err != nil {
		t.Fatalf("运行失败: %v", err)
	}

	if result.Status != "passed" {
		t.Errorf("预期通过, 实际: %s", result.Status)
	}
}

func TestPrivacyComputingScenario(t *testing.T) {
	config := ScenarioConfig{MockMode: true}
	validator := NewScenarioValidator(config)

	result, err := validator.RunScenario(context.Background(), "privacy_computing")
	if err != nil {
		t.Fatalf("运行失败: %v", err)
	}

	if result.Status != "passed" {
		t.Errorf("预期通过, 实际: %s", result.Status)
	}
}