// Package display provides the Multi-Device Display Engine (v5.5.0).
//
// The DisplayEngine manages multi-device display including
// 3D rendering, device adaptation, and display synchronization.
package display

import (
	"sync"
	"time"

	"ofa/center/internal/models"
)

// DisplayEngine manages multi-device display.
type DisplayEngine struct {
	mu sync.RWMutex

	// Profiles
	profiles map[string]*models.MultiDisplayProfile

	// Contexts
	contexts map[string]*models.DisplayContext

	// Device connections
	connectedDevices map[string][]string // identityID -> deviceIDs

	// Quality presets
	qualityPresets map[string]QualityPreset

	// Device profiles
	deviceProfiles map[string]models.DeviceRenderProfile
}

// QualityPreset defines a quality preset configuration.
type QualityPreset struct {
	Name              string
	RenderQuality     string
	TextureQuality    string
	ShadowQuality     string
	AntiAliasing      string
	TargetFPS         int
	DrawDistance      float64
	ParticleLimit     int
	PostProcessing    bool
}

// NewDisplayEngine creates a new DisplayEngine.
func NewDisplayEngine() *DisplayEngine {
	engine := &DisplayEngine{
		profiles:         make(map[string]*models.MultiDisplayProfile),
		contexts:         make(map[string]*models.DisplayContext),
		connectedDevices: make(map[string][]string),
		qualityPresets:   make(map[string]QualityPreset),
		deviceProfiles:   make(map[string]models.DeviceRenderProfile),
	}

	// Initialize default quality presets
	engine.initQualityPresets()

	// Initialize device profiles
	engine.initDeviceProfiles()

	return engine
}

// initQualityPresets initializes default quality presets.
func (e *DisplayEngine) initQualityPresets() {
	e.qualityPresets["low"] = QualityPreset{
		Name:           "low",
		RenderQuality:  "low",
		TextureQuality: "low",
		ShadowQuality:  "off",
		AntiAliasing:   "none",
		TargetFPS:      30,
		DrawDistance:   50,
		ParticleLimit:  100,
		PostProcessing: false,
	}

	e.qualityPresets["medium"] = QualityPreset{
		Name:           "medium",
		RenderQuality:  "medium",
		TextureQuality: "medium",
		ShadowQuality:  "low",
		AntiAliasing:   "fxaa",
		TargetFPS:      60,
		DrawDistance:   100,
		ParticleLimit:  500,
		PostProcessing: true,
	}

	e.qualityPresets["high"] = QualityPreset{
		Name:           "high",
		RenderQuality:  "high",
		TextureQuality: "high",
		ShadowQuality:  "medium",
		AntiAliasing:   "smaa",
		TargetFPS:      60,
		DrawDistance:   200,
		ParticleLimit:  1000,
		PostProcessing: true,
	}

	e.qualityPresets["ultra"] = QualityPreset{
		Name:           "ultra",
		RenderQuality:  "ultra",
		TextureQuality: "high",
		ShadowQuality:  "high",
		AntiAliasing:   "taa",
		TargetFPS:      60,
		DrawDistance:   500,
		ParticleLimit:  2000,
		PostProcessing: true,
	}
}

// initDeviceProfiles initializes default device profiles.
func (e *DisplayEngine) initDeviceProfiles() {
	// Phone - low tier
	e.deviceProfiles["phone_low"] = models.DeviceRenderProfile{
		DeviceType:         "phone",
		DeviceTier:         "low",
		MaxQuality:         "low",
		RecommendedQuality: "low",
		ResolutionScale:    0.75,
		TextureLimit:       512,
		ShadowResolution:   512,
		ParticleLimit:      100,
		DrawDistance:       50,
		MinFPSForQualityUp: 28,
		MaxFPSForQualityDown: 20,
	}

	// Phone - mid tier
	e.deviceProfiles["phone_mid"] = models.DeviceRenderProfile{
		DeviceType:         "phone",
		DeviceTier:         "medium",
		MaxQuality:         "medium",
		RecommendedQuality: "medium",
		ResolutionScale:    1.0,
		TextureLimit:       1024,
		ShadowResolution:   1024,
		ParticleLimit:      500,
		DrawDistance:       100,
		MinFPSForQualityUp: 55,
		MaxFPSForQualityDown: 45,
	}

	// Phone - high tier
	e.deviceProfiles["phone_high"] = models.DeviceRenderProfile{
		DeviceType:         "phone",
		DeviceTier:         "high",
		MaxQuality:         "high",
		RecommendedQuality: "high",
		ResolutionScale:    1.0,
		TextureLimit:       2048,
		ShadowResolution:   2048,
		ParticleLimit:      1000,
		DrawDistance:       200,
		SupportsRayTracing: false,
		MinFPSForQualityUp: 55,
		MaxFPSForQualityDown: 45,
	}

	// Tablet
	e.deviceProfiles["tablet"] = models.DeviceRenderProfile{
		DeviceType:         "tablet",
		DeviceTier:         "medium",
		MaxQuality:         "high",
		RecommendedQuality: "medium",
		ResolutionScale:    1.0,
		TextureLimit:       2048,
		ShadowResolution:   1024,
		ParticleLimit:      1000,
		DrawDistance:       150,
		MinFPSForQualityUp: 55,
		MaxFPSForQualityDown: 45,
	}

	// Watch
	e.deviceProfiles["watch"] = models.DeviceRenderProfile{
		DeviceType:         "watch",
		DeviceTier:         "low",
		MaxQuality:         "low",
		RecommendedQuality: "low",
		ResolutionScale:    0.5,
		TextureLimit:       256,
		ShadowResolution:   0,
		ParticleLimit:      50,
		DrawDistance:       20,
		MinFPSForQualityUp: 28,
		MaxFPSForQualityDown: 20,
	}

	// Desktop
	e.deviceProfiles["desktop"] = models.DeviceRenderProfile{
		DeviceType:         "desktop",
		DeviceTier:         "high",
		MaxQuality:         "ultra",
		RecommendedQuality: "high",
		ResolutionScale:    1.0,
		TextureLimit:       4096,
		ShadowResolution:   4096,
		ParticleLimit:      5000,
		DrawDistance:       500,
		SupportsRayTracing: true,
		SupportsComputeShaders: true,
		MinFPSForQualityUp: 55,
		MaxFPSForQualityDown: 45,
	}

	// VR
	e.deviceProfiles["vr"] = models.DeviceRenderProfile{
		DeviceType:         "vr",
		DeviceTier:         "high",
		MaxQuality:         "high",
		RecommendedQuality: "medium",
		ResolutionScale:    1.0,
		TextureLimit:       2048,
		ShadowResolution:   1024,
		ParticleLimit:      500,
		DrawDistance:       100,
		MinFPSForQualityUp: 88,
		MaxFPSForQualityDown: 85,
	}
}

// === Profile Management ===

// GetProfile returns the multi-display profile for an identity.
func (e *DisplayEngine) GetProfile(identityID string) *models.MultiDisplayProfile {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if profile, ok := e.profiles[identityID]; ok {
		return profile
	}
	return nil
}

// CreateProfile creates a new multi-display profile.
func (e *DisplayEngine) CreateProfile(identityID string) *models.MultiDisplayProfile {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile := &models.MultiDisplayProfile{
		IdentityID: identityID,
		RenderingSettings: models.RenderingSettings{
			RenderEngine:         "webgl",
			QualityPreset:        "medium",
			RenderQuality:        "medium",
			TextureQuality:       "medium",
			ShadowQuality:        "medium",
			AntiAliasing:         "smaa",
			LightingMode:         "mixed",
			GlobalIllumination:   false,
			AmbientOcclusion:     "ssao",
			AnimationQuality:     "medium",
			InverseKinematics:    true,
			PhysicsEnabled:       true,
			PhysicsQuality:       "medium",
			PostProcessingEnabled: true,
			TargetFPS:            60,
			VSync:                true,
			AdaptiveQuality:      true,
			ShaderComplexity:     "standard",
		},
		DeviceAdaptation: models.DeviceAdaptation{
			AdaptationMode:   "auto",
			AutoOptimize:     true,
			DeviceProfiles:   make(map[string]models.DeviceRenderProfile),
			PerformanceScaling: models.PerformanceScaling{
				ScalingMode:     "dynamic",
				FPSSampleWindow: 60,
				FPSTargetLow:    30,
				FPSTargetHigh:   60,
			},
		},
		DisplaySync: models.DisplaySync{
			SyncMode:           "state_sync",
			SyncFrequency:      100,
			SyncPriority:       "high",
			StateSyncEnabled:   true,
			SyncedStates:       []string{"position", "rotation", "animation", "expression"},
			InterpolationMode:  "linear",
			PredictionEnabled:  true,
			PredictionSteps:    2,
			ConflictResolution: "timestamp",
			NetworkProtocol:    "websocket",
			CompressionEnabled: true,
			DeltaEncoding:      true,
		},
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	e.profiles[identityID] = profile
	return profile
}

// UpdateProfile updates the multi-display profile.
func (e *DisplayEngine) UpdateProfile(identityID string, profile *models.MultiDisplayProfile) {
	e.mu.Lock()
	defer e.mu.Unlock()

	profile.Version++
	profile.UpdatedAt = time.Now()
	e.profiles[identityID] = profile
}

// UpdateRenderingSettings updates rendering settings.
func (e *DisplayEngine) UpdateRenderingSettings(identityID string, settings models.RenderingSettings) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if profile, ok := e.profiles[identityID]; ok {
		profile.RenderingSettings = settings
		profile.Version++
		profile.UpdatedAt = time.Now()
	}
}

// UpdateDeviceAdaptation updates device adaptation settings.
func (e *DisplayEngine) UpdateDeviceAdaptation(identityID string, adaptation models.DeviceAdaptation) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if profile, ok := e.profiles[identityID]; ok {
		profile.DeviceAdaptation = adaptation
		profile.Version++
		profile.UpdatedAt = time.Now()
	}
}

// UpdateDisplaySync updates display sync settings.
func (e *DisplayEngine) UpdateDisplaySync(identityID string, sync models.DisplaySync) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if profile, ok := e.profiles[identityID]; ok {
		profile.DisplaySync = sync
		profile.Version++
		profile.UpdatedAt = time.Now()
	}
}

// === Quality Management ===

// SetQualityPreset sets the quality preset.
func (e *DisplayEngine) SetQualityPreset(identityID string, preset string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	if profile, ok := e.profiles[identityID]; ok {
		if p, exists := e.qualityPresets[preset]; exists {
			profile.RenderingSettings.QualityPreset = preset
			profile.RenderingSettings.RenderQuality = p.RenderQuality
			profile.RenderingSettings.TextureQuality = p.TextureQuality
			profile.RenderingSettings.ShadowQuality = p.ShadowQuality
			profile.RenderingSettings.AntiAliasing = p.AntiAliasing
			profile.RenderingSettings.TargetFPS = p.TargetFPS
			profile.Version++
			profile.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// GetRecommendedQuality returns the recommended quality for a device.
func (e *DisplayEngine) GetRecommendedQuality(deviceType string, deviceTier string) string {
	profileKey := deviceType + "_" + deviceTier
	if profile, ok := e.deviceProfiles[profileKey]; ok {
		return profile.RecommendedQuality
	}
	return "medium"
}

// AdjustQualityForPerformance adjusts quality based on performance.
func (e *DisplayEngine) AdjustQualityForPerformance(identityID string, currentFPS float64) string {
	e.mu.Lock()
	defer e.mu.Unlock()

	if profile, ok := e.profiles[identityID]; ok {
		targetFPS := float64(profile.RenderingSettings.TargetFPS)

		// If FPS is too low, reduce quality
		if currentFPS < targetFPS*0.8 {
			currentQuality := profile.RenderingSettings.QualityPreset
			switch currentQuality {
			case "ultra":
				profile.RenderingSettings.QualityPreset = "high"
				return "high"
			case "high":
				profile.RenderingSettings.QualityPreset = "medium"
				return "medium"
			case "medium":
				profile.RenderingSettings.QualityPreset = "low"
				return "low"
			}
		}

		// If FPS is high enough, consider increasing quality
		if currentFPS > targetFPS*1.1 && profile.RenderingSettings.AdaptiveQuality {
			currentQuality := profile.RenderingSettings.QualityPreset
			switch currentQuality {
			case "low":
				profile.RenderingSettings.QualityPreset = "medium"
				return "medium"
			case "medium":
				profile.RenderingSettings.QualityPreset = "high"
				return "high"
			case "high":
				profile.RenderingSettings.QualityPreset = "ultra"
				return "ultra"
			}
		}

		return profile.RenderingSettings.QualityPreset
	}
	return "medium"
}

// === Device Management ===

// RegisterDevice registers a device for an identity.
func (e *DisplayEngine) RegisterDevice(identityID string, deviceID string, deviceType string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	devices := e.connectedDevices[identityID]
	for _, d := range devices {
		if d == deviceID {
			return // already registered
		}
	}
	e.connectedDevices[identityID] = append(devices, deviceID)
}

// UnregisterDevice unregisters a device.
func (e *DisplayEngine) UnregisterDevice(identityID string, deviceID string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	devices := e.connectedDevices[identityID]
	for i, d := range devices {
		if d == deviceID {
			e.connectedDevices[identityID] = append(devices[:i], devices[i+1:]...)
			return
		}
	}
}

// GetConnectedDevices returns connected devices for an identity.
func (e *DisplayEngine) GetConnectedDevices(identityID string) []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.connectedDevices[identityID]
}

// GetDeviceProfile returns the device render profile.
func (e *DisplayEngine) GetDeviceProfile(deviceType string, deviceTier string) *models.DeviceRenderProfile {
	e.mu.RLock()
	defer e.mu.RUnlock()

	profileKey := deviceType + "_" + deviceTier
	if profile, ok := e.deviceProfiles[profileKey]; ok {
		return &profile
	}

	// Return default profile
	if profile, ok := e.deviceProfiles["phone_mid"]; ok {
		return &profile
	}
	return nil
}

// === Scene Management ===

// AddSceneRenderProfile adds a scene render profile.
func (e *DisplayEngine) AddSceneRenderProfile(identityID string, profile models.SceneRenderProfile) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if p, ok := e.profiles[identityID]; ok {
		p.SceneRenderProfiles = append(p.SceneRenderProfiles, profile)
		p.Version++
		p.UpdatedAt = time.Now()
	}
}

// GetSceneRenderProfile returns the scene render profile.
func (e *DisplayEngine) GetSceneRenderProfile(identityID string, sceneName string) *models.SceneRenderProfile {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if profile, ok := e.profiles[identityID]; ok {
		for _, scene := range profile.SceneRenderProfiles {
			if scene.SceneName == sceneName {
				return &scene
			}
		}
	}
	return nil
}

// === Sync Management ===

// SyncState syncs display state across devices.
func (e *DisplayEngine) SyncState(identityID string, state map[string]interface{}) map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Get sync settings
	syncConfig := models.DisplaySync{
		SyncMode:          "state_sync",
		InterpolationMode: "linear",
	}

	if profile, ok := e.profiles[identityID]; ok {
		syncConfig = profile.DisplaySync
	}

	// Apply sync configuration
	result := make(map[string]interface{})
	for _, key := range syncConfig.SyncedStates {
		if val, ok := state[key]; ok {
			result[key] = val
		}
	}

	// Add sync metadata
	result["_sync_time"] = time.Now().UnixMilli()
	result["_sync_mode"] = syncConfig.SyncMode

	return result
}

// ResolveConflict resolves display state conflicts.
func (e *DisplayEngine) ResolveConflict(identityID string, states []map[string]interface{}) map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	resolution := "timestamp"
	if profile, ok := e.profiles[identityID]; ok {
		resolution = profile.DisplaySync.ConflictResolution
	}

	switch resolution {
	case "last_write":
		if len(states) > 0 {
			return states[len(states)-1]
		}
	case "timestamp":
		// Find state with latest timestamp
		var latest map[string]interface{}
		var latestTime int64
		for _, state := range states {
			if t, ok := state["_sync_time"].(int64); ok {
				if t > latestTime {
					latestTime = t
					latest = state
				}
			}
		}
		if latest != nil {
			return latest
		}
	case "priority":
		// Find state from master device
		if profile, ok := e.profiles[identityID]; ok {
			masterDevice := profile.DisplaySync.MasterDevice
			for _, state := range states {
				if deviceID, ok := state["_device_id"].(string); ok {
					if deviceID == masterDevice {
						return state
					}
				}
			}
		}
	}

	// Default: return first state
	if len(states) > 0 {
		return states[0]
	}
	return nil
}

// === Decision Context ===

// GetDecisionContext returns the display decision context.
func (e *DisplayEngine) GetDecisionContext(identityID string) *models.DisplayContext {
	e.mu.RLock()
	defer e.mu.RUnlock()

	context := &models.DisplayContext{
		IdentityID:    identityID,
		CurrentQuality: "medium",
		CurrentFPS:    60.0,
		SyncStatus:    "synced",
		Timestamp:     time.Now(),
	}

	if profile, ok := e.profiles[identityID]; ok {
		context.CurrentQuality = profile.RenderingSettings.QualityPreset

		// Generate recommendations
		context.RecommendedQuality = e.generateQualityRecommendation(profile)
		context.QualityAdjustmentNeeded = e.checkQualityAdjustment(profile)
	}

	// Add connected devices
	context.ConnectedDevices = e.connectedDevices[identityID]

	e.contexts[identityID] = context
	return context
}

// generateQualityRecommendation generates a quality recommendation.
func (e *DisplayEngine) generateQualityRecommendation(profile *models.MultiDisplayProfile) string {
	// Check if adaptive quality is enabled
	if !profile.RenderingSettings.AdaptiveQuality {
		return profile.RenderingSettings.QualityPreset
	}

	// Get device profile
	deviceType := "phone"
	deviceTier := "medium"

	if deviceProfile := e.deviceProfiles[deviceType+"_"+deviceTier]; deviceProfile.DeviceType != "" {
		return deviceProfile.RecommendedQuality
	}

	return "medium"
}

// checkQualityAdjustment checks if quality adjustment is needed.
func (e *DisplayEngine) checkQualityAdjustment(profile *models.MultiDisplayProfile) bool {
	// Simple heuristic: check if current quality differs from recommended
	recommended := e.generateQualityRecommendation(profile)
	return profile.RenderingSettings.QualityPreset != recommended
}

// UpdatePerformanceMetrics updates performance metrics.
func (e *DisplayEngine) UpdatePerformanceMetrics(identityID string, fps float64, memory int64, gpuUtil float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if context, ok := e.contexts[identityID]; ok {
		context.CurrentFPS = fps
		context.MemoryUsage = memory
		context.GPUUtilization = gpuUtil
		context.Timestamp = time.Now()
	}
}

// UpdateSyncStatus updates sync status.
func (e *DisplayEngine) UpdateSyncStatus(identityID string, status string, latency int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if context, ok := e.contexts[identityID]; ok {
		context.SyncStatus = status
		context.SyncLatency = latency
		context.LastSyncTime = time.Now()
	}
}