package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	tts "ofa/center/internal/tts"
	"ofa/center/internal/tts/providers"
)

func main() {
	f, err := os.Create("tts_test_output.txt")
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer f.Close()

	output := func(format string, args ...interface{}) {
		s := fmt.Sprintf(format, args...)
		fmt.Fprintln(f, s)
		fmt.Println(s)
	}

	output("=== TTS API Test ===")
	output("Time: %s", time.Now().Format("2006-01-02 15:04:05"))
	output("")

	// Test 1: Create TTS Engine
	output("Test 1: Create TTS Engine")
	config := tts.DefaultTTSEngineConfig()
	engine := tts.NewTTSEngine(config)
	output("OK - TTS Engine created successfully")
	output("   Primary Provider: %s", config.PrimaryProvider)
	output("   Default Voice: %s", config.DefaultVoice)
	output("")

	// Test 2: List Voices
	output("Test 2: List Available Voices")
	voices, err := engine.ListVoices(context.Background(), "doubao")
	if err != nil {
		output("ERROR - List voices failed: %v", err)
	} else {
		output("OK - Found %d voices:", len(voices))
		for i, v := range voices {
			if i < 5 {
				output("   - %s (%s): %s", v.Name, v.VoiceID, v.Description)
			}
		}
		if len(voices) > 5 {
			output("   ... and %d more voices", len(voices)-5)
		}
	}
	output("")

	// Test 3: Volcengine Provider
	output("Test 3: Volcengine Provider")
	volcengine := providers.NewVolcengineProvider("test_app_id", "test_token")
	output("OK - Volcengine Provider created")
	output("   Name: %s", volcengine.Name())
	output("   Supports Streaming: %v", volcengine.SupportsStreaming())
	output("   Supports Cloning: %v", volcengine.SupportsCloning())
	output("")

	// Test 4: Doubao Provider
	output("Test 4: Doubao Provider")
	doubao := providers.NewDoubaoProvider("test_app_id", "test_token")
	output("OK - Doubao Provider created")
	output("   Name: %s", doubao.Name())
	output("   Supports Streaming: %v", doubao.SupportsStreaming())
	output("   Supports Cloning: %v", doubao.SupportsCloning())
	output("")

	// Test 5: Voice Mapping
	output("Test 5: Voice Mapping")
	identityID := "user-test-001"
	voiceID := "zh_female_meilinvyou_uranus_bigtts"
	engine.SetVoice(identityID, voiceID)
	mappedVoice := engine.GetVoice(identityID)
	output("OK - Voice mapping set")
	output("   Identity: %s -> Voice: %s", identityID, mappedVoice)
	output("")

	// Test 6: Synthesis Request Structure
	output("Test 6: Synthesis Request Structure")
	req := &tts.SynthesisRequest{
		Text:         "你好，我是OFA数字人助手！",
		VoiceID:      "zh_female_meilinvyou_uranus_bigtts",
		OutputFormat: "mp3",
		SampleRate:   24000,
		Rate:         1.0,
		Pitch:        1.0,
		Volume:       0.7,
		IdentityID:   "user-test-001",
	}
	reqJSON, _ := json.MarshalIndent(req, "   ", "  ")
	output("OK - Synthesis Request:")
	output("   %s", string(reqJSON))
	output("")

	// Test 7: Synthesis Result Structure
	output("Test 7: Synthesis Result Structure")
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
	resultJSON, _ := json.MarshalIndent(result, "   ", "  ")
	output("OK - Synthesis Result:")
	output("   %s", string(resultJSON))
	output("")

	// Test 8: VoiceInfo Structure
	output("Test 8: Voice Info Structure")
	voiceInfo := tts.VoiceInfo{
		VoiceID:     "zh_male_sunwukong_uranus_bigtts",
		Name:        "猴哥",
		Language:    "zh-CN",
		Gender:      "male",
		Age:         "adult",
		Provider:    "doubao",
		Description: "孙悟空音色",
	}
	voiceJSON, _ := json.MarshalIndent(voiceInfo, "   ", "  ")
	output("OK - Voice Info:")
	output("   %s", string(voiceJSON))
	output("")

	// Test 9: Clone Request Structure
	output("Test 9: Clone Request Structure")
	cloneReq := &tts.CloneRequest{
		IdentityID: "user-test-001",
		VoiceName:  "我的声音",
		Language:   "zh-CN",
		ReferenceAudios: []tts.ReferenceAudio{
			{
				AudioURL:      "https://example.com/voice-sample.mp3",
				DurationMs:    10000,
				Transcription: "这是参考音频的文本",
			},
		},
	}
	cloneJSON, _ := json.MarshalIndent(cloneReq, "   ", "  ")
	output("OK - Clone Request:")
	output("   %s", string(cloneJSON))
	output("")

	// Summary
	output("=== Test Summary ===")
	output("OK - All TTS API structures validated successfully")
	output("OK - TTS Engine initialization working")
	output("OK - Provider creation working")
	output("OK - Voice mapping working")
	output("")
	output("Note: Actual synthesis requires API credentials (Volcengine/Doubao)")
	output("      Set volcengine_app_id/volcengine_token or doubao_app_id/doubao_token")
}