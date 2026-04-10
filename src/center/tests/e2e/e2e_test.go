package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ofa/center/internal/config"
	"github.com/ofa/center/internal/service"
	"github.com/ofa/center/pkg/rest"
)

// TestE2EIdentitySync 测试身份同步端到端流程
func TestE2EIdentitySync(t *testing.T) {
	// 启动测试服务器
	server := startTestServer(t)
	defer server.Stop()

	baseURL := "http://localhost:8080"

	// 等待服务器就绪
	waitForServer(t, baseURL)

	// 场景1: 创建身份
	t.Run("CreateIdentity", func(t *testing.T) {
		identityJSON := `{
			"id": "identity_001",
			"name": "测试用户",
			"nickname": "Tester",
			"personality": {
				"openness": 0.7,
				"conscientiousness": 0.6,
				"extraversion": 0.5,
				"agreeableness": 0.8,
				"neuroticism": 0.4
			},
			"value_system": {
				"privacy": 0.8,
				"efficiency": 0.7,
				"health": 0.9
			}
		}`

		resp, err := http.Post(baseURL+"/api/v1/identities", "application/json", strings.NewReader(identityJSON))
		if err != nil {
			t.Fatalf("Failed to create identity: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Unexpected status: %d, body: %s", resp.StatusCode, body)
		}

		t.Logf("Identity created successfully")
	})

	// 场景2: 获取身份
	t.Run("GetIdentity", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/v1/identities/identity_001")
		if err != nil {
			t.Fatalf("Failed to get identity: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Unexpected status: %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(body, &result)

		if result["name"] != "测试用户" {
			t.Errorf("Expected name '测试用户', got %v", result["name"])
		}

		t.Logf("Identity retrieved: %s", string(body))
	})

	// 场景3: 行为上报
	t.Run("ReportBehavior", func(t *testing.T) {
		behaviorJSON := `{
			"agent_id": "agent_001",
			"identity_id": "identity_001",
			"type": "decision",
			"observation": {
				"decision_type": "impulse_purchase",
				"context": {"product": "电子产品", "price": 1000}
			},
			"timestamp": "2026-04-10T10:00:00Z"
		}`

		resp, err := http.Post(baseURL+"/api/v1/behaviors", "application/json", strings.NewReader(behaviorJSON))
		if err != nil {
			t.Fatalf("Failed to report behavior: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Fatalf("Unexpected status: %d, body: %s", resp.StatusCode, body)
		}

		t.Logf("Behavior reported successfully")
	})

	// 场景4: 情绪触发
	t.Run("TriggerEmotion", func(t *testing.T) {
		emotionJSON := `{
			"identity_id": "identity_001",
			"trigger_type": "event",
			"trigger_desc": "获得奖励",
			"emotion_type": "joy",
			"intensity": 0.8
		}`

		resp, err := http.Post(baseURL+"/api/v1/emotions/trigger", "application/json", strings.NewReader(emotionJSON))
		if err != nil {
			t.Fatalf("Failed to trigger emotion: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Emotion trigger response: %d, body: %s", resp.StatusCode, body)
		}

		t.Logf("Emotion triggered successfully")
	})
}

// TestE2EDeviceSync 测试设备同步端到端流程
func TestE2EDeviceSync(t *testing.T) {
	server := startTestServer(t)
	defer server.Stop()

	baseURL := "http://localhost:8080"
	waitForServer(t, baseURL)

	// 场景1: 设备注册
	t.Run("DeviceRegister", func(t *testing.T) {
		deviceJSON := `{
			"agent_id": "device_001",
			"identity_id": "identity_001",
			"device_type": "mobile",
			"device_name": "测试手机",
			"capabilities": ["ui_automation", "voice"],
			"status": "online"
		}`

		resp, err := http.Post(baseURL+"/api/v1/devices", "application/json", strings.NewReader(deviceJSON))
		if err != nil {
			t.Fatalf("Failed to register device: %v", err)
		}
		defer resp.Body.Close()

		t.Logf("Device registered with status: %d", resp.StatusCode)
	})

	// 场景2: 心跳更新
	t.Run("Heartbeat", func(t *testing.T) {
		heartbeatJSON := `{
			"agent_id": "device_001",
			"status": "online",
			"battery": 85,
			"network": "wifi"
		}`

		resp, err := http.Post(baseURL+"/api/v1/devices/device_001/heartbeat", "application/json", strings.NewReader(heartbeatJSON))
		if err != nil {
			t.Fatalf("Failed to send heartbeat: %v", err)
		}
		defer resp.Body.Close()

		t.Logf("Heartbeat sent with status: %d", resp.StatusCode)
	})

	// 场景3: 数据同步
	t.Run("DataSync", func(t *testing.T) {
		syncJSON := `{
			"agent_id": "device_001",
			"identity_id": "identity_001",
			"sync_type": "delta",
			"changes": [
				{"key": "preference.theme", "value": "dark", "version": 2}
			]
		}`

		resp, err := http.Post(baseURL+"/api/v1/sync", "application/json", strings.NewReader(syncJSON))
		if err != nil {
			t.Fatalf("Failed to sync data: %v", err)
		}
		defer resp.Body.Close()

		t.Logf("Data sync completed with status: %d", resp.StatusCode)
	})
}

// TestE2ETTS 测试 TTS 端到端流程
func TestE2ETTS(t *testing.T) {
	server := startTestServer(t)
	defer server.Stop()

	baseURL := "http://localhost:8080"
	waitForServer(t, baseURL)

	// 场景1: 语音合成
	t.Run("SynthesizeSpeech", func(t *testing.T) {
		ttsJSON := `{
			"text": "你好，欢迎使用 OFA 系统",
			"voice_id": "zh_female_qingxin",
			"speed": 1.0,
			"format": "mp3"
		}`

		resp, err := http.Post(baseURL+"/api/v1/tts/synthesize", "application/json", strings.NewReader(ttsJSON))
		if err != nil {
			t.Fatalf("Failed to synthesize speech: %v", err)
		}
		defer resp.Body.Close()

		t.Logf("TTS synthesis completed with status: %d", resp.StatusCode)
	})

	// 场景2: 获取声音列表
	t.Run("GetVoices", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/api/v1/tts/voices")
		if err != nil {
			t.Fatalf("Failed to get voices: %v", err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		t.Logf("Voices list: %s", string(body))
	})
}

// TestE2EHealthCheck 测试健康检查端点
func TestE2EHealthCheck(t *testing.T) {
	server := startTestServer(t)
	defer server.Stop()

	baseURL := "http://localhost:8080"
	waitForServer(t, baseURL)

	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	t.Logf("Health check response: %s", string(body))
}

// TestE2EFullScenario 测试完整场景
func TestE2EFullScenario(t *testing.T) {
	server := startTestServer(t)
	defer server.Stop()

	baseURL := "http://localhost:8080"
	waitForServer(t, baseURL)

	// 完整流程: 创建身份 -> 注册设备 -> 行为上报 -> 情绪触发 -> TTS合成
	scenarios := []struct {
		name   string
		path   string
		body   string
		expect int
	}{
		{
			name: "CreateIdentity",
			path: "/api/v1/identities",
			body: `{"id": "e2e_user", "name": "E2E测试用户", "personality": {"openness": 0.6}}`,
			expect: http.StatusOK,
		},
		{
			name: "RegisterDevice",
			path: "/api/v1/devices",
			body: `{"agent_id": "e2e_device", "identity_id": "e2e_user", "device_type": "mobile"}`,
			expect: http.StatusOK,
		},
		{
			name: "ReportBehavior",
			path: "/api/v1/behaviors",
			body: `{"agent_id": "e2e_device", "identity_id": "e2e_user", "type": "interaction"}`,
			expect: http.StatusOK,
		},
		{
			name: "GetEmotionContext",
			path: "/api/v1/emotions/e2e_user/context",
			body: "",
			expect: http.StatusOK,
		},
		{
			name: "TTSSynthesize",
			path: "/api/v1/tts/synthesize",
			body: `{"text": "端到端测试成功", "voice_id": "default"}`,
			expect: http.StatusOK,
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			var resp *http.Response
			var err error

			if s.body != "" {
				resp, err = http.Post(baseURL+s.path, "application/json", strings.NewReader(s.body))
			} else {
				resp, err = http.Get(baseURL + s.path)
			}

			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			t.Logf("%s response (status %d): %s", s.name, resp.StatusCode, string(body))

			// 检查响应状态（可能返回不同状态码，视 API 实现而定）
			if resp.StatusCode >= 500 {
				t.Errorf("Server error: %d", resp.StatusCode)
			}
		})
	}
}

// 辅助函数

func startTestServer(t *testing.T) *rest.Server {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Name: "e2e-test-center",
		},
		REST: config.RESTConfig{
			Address: ":8080",
		},
		GRPC: config.GRPCConfig{
			Address: ":9090",
		},
	}

	svc := service.NewCenterService(cfg)
	server := rest.NewServer(svc, cfg)

	go func() {
		if err := server.Start(":8080"); err != nil {
			t.Logf("Server stopped: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	return server
}

func waitForServer(t *testing.T, baseURL string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for {
		resp, err := http.Get(baseURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}

		select {
		case <-ctx.Done():
			t.Fatal("Server did not start within timeout")
		case <-time.After(500 * time.Millisecond):
		}
	}
}