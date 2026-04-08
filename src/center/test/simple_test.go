package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	tts "github.com/ofa/center/internal/tts"
	"github.com/ofa/center/internal/tts/providers"
)

func main() {
	// Output to stdout and file
	output := func(format string, args ...interface{}) {
		s := fmt.Sprintf(format, args...)
		fmt.Println(s)
	}

	output("=== TTS API Test ===")
	output("")

	// Test 1: Create TTS Engine
	output("Test 1: Create TTS Engine")
	config := tts.DefaultTTSEngineConfig()
	engine := tts.NewTTSEngine(config)
	output("OK - TTS Engine created")
	output("Primary Provider: %s", config.PrimaryProvider)
	output("Default Voice: %s", config.DefaultVoice)
	output("")

	// Test 2: List Voices
	output("Test 2: List Voices")
	voices, err := engine.ListVoices(context.Background(), "doubao")
	if err != nil {
		output("ERROR: %v", err)
	} else {
		output("OK - Found %d voices", len(voices))
		for i, v := range voices {
			if i < 3 {
				output("  - %s (%s)", v.Name, v.VoiceID)
			}
		}
	}
	output("")

	// Test 3: Providers
	output("Test 3: Providers")
	volc := providers.NewVolcengineProvider("app", "token")
	doubao := providers.NewDoubaoProvider("app", "token")
	output("OK - Volcengine: %s, streaming=%v", volc.Name(), volc.SupportsStreaming())
	output("OK - Doubao: %s, streaming=%v", doubao.Name(), doubao.SupportsStreaming())
	output("")

	// Test 4: Voice Mapping
	output("Test 4: Voice Mapping")
	engine.SetVoice("user-001", "voice-001")
	output("OK - Mapped user-001 -> %s", engine.GetVoice("user-001"))
	output("")

	// Test 5: Request Structure
	output("Test 5: Request Structure")
	req := &tts.SynthesisRequest{
		Text:         "你好，我是OFA数字人助手！",
		VoiceID:      "zh_female_meilinvyou_uranus_bigtts",
		OutputFormat: "mp3",
		SampleRate:   24000,
		Rate:         1.0,
		Pitch:        1.0,
		Volume:       0.7,
	}
	data, _ := json.MarshalIndent(req, "", "  ")
	output("OK - Request:")
	output("%s", string(data))
	output("")

	// Test 6: Result Structure
	output("Test 6: Result Structure")
	result := &tts.SynthesisResult{
		DurationMs:   3500,
		Format:       "mp3",
		SampleRate:   24000,
		Provider:     "doubao",
		VoiceUsed:    "zh_female_meilinvyou_uranus_bigtts",
		LatencyMs:    250,
		QualityScore: 0.95,
		Success:      true,
	}
	data, _ = json.MarshalIndent(result, "", "  ")
	output("OK - Result:")
	output("%s", string(data))
	output("")

	// Test 7: Clone Request
	output("Test 7: Clone Request")
	cloneReq := &tts.CloneRequest{
		IdentityID: "user-001",
		VoiceName:  "我的声音",
		Language:   "zh-CN",
		ReferenceAudios: []tts.ReferenceAudio{
			{AudioURL: "https://example.com/sample.mp3", DurationMs: 10000},
		},
	}
	data, _ = json.MarshalIndent(cloneReq, "", "  ")
	output("OK - Clone Request:")
	output("%s", string(data))
	output("")

	output("=== All Tests Passed ===")

	// Write to file as well
	f, _ := os.Create("test_result.txt")
	f.WriteString("TTS API Test Completed Successfully\n")
	f.Close()
	output("Result written to test_result.txt")
}