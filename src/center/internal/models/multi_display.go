// Package models defines the MultiDisplay (多端展示) models for v5.5.0.
//
// MultiDisplay represents the multi-device display system for avatar,
// including 3D rendering, device adaptation, and display synchronization.
package models

import (
	"encoding/json"
	"time"
)

// MultiDisplayProfile represents the complete multi-device display configuration.
type MultiDisplayProfile struct {
	IdentityID string `json:"identity_id"`

	// 3D rendering settings
	RenderingSettings RenderingSettings `json:"rendering_settings"`

	// Device adaptation settings
	DeviceAdaptation DeviceAdaptation `json:"device_adaptation"`

	// Display synchronization
	DisplaySync DisplaySync `json:"display_sync"`

	// Scene rendering profiles
	SceneRenderProfiles []SceneRenderProfile `json:"scene_render_profiles"`

	// Display context
	DisplayContext DisplayContext `json:"display_context"`

	// Metadata
	Version   int64     `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RenderingSettings defines 3D rendering configuration.
type RenderingSettings struct {
	// Render engine settings
	RenderEngine     string `json:"render_engine"`     // webgl, vulkan, metal, directx
	EngineVersion    string `json:"engine_version"`
	FallbackEngine   string `json:"fallback_engine"`   // fallback when primary not available

	// Quality settings
	QualityPreset      string `json:"quality_preset"`      // low, medium, high, ultra, custom
	RenderQuality      string `json:"render_quality"`      // low, medium, high, ultra
	TextureQuality     string `json:"texture_quality"`     // low, medium, high
	ShadowQuality      string `json:"shadow_quality"`      // off, low, medium, high, ultra
	AntiAliasing       string `json:"anti_aliasing"`       // none, fxaa, smaa, taa, msaa_2x, msaa_4x, msaa_8x
	AnisotropicFiltering int  `json:"anisotropic_filtering"` // 0, 2, 4, 8, 16

	// Lighting settings
	LightingMode      string `json:"lighting_mode"`      // baked, mixed, realtime
	LightingQuality   string `json:"lighting_quality"`   // low, medium, high
	GlobalIllumination bool  `json:"global_illumination"`
	AmbientOcclusion  string `json:"ambient_occlusion"`  // none, ssao, hbao, gtao
	Reflections       string `json:"reflections"`        // none, screen_space, ray_traced

	// Animation settings
	AnimationQuality   string `json:"animation_quality"`   // low, medium, high
	AnimationBlendMode string `json:"animation_blend_mode"` // linear, additive, override
	InverseKinematics  bool   `json:"inverse_kinematics"`
	FacialRigQuality   string `json:"facial_rig_quality"`   // bones, blendshapes, both

	// Physics settings
	PhysicsEnabled     bool   `json:"physics_enabled"`
	PhysicsQuality     string `json:"physics_quality"`     // low, medium, high
	ClothSimulation    bool   `json:"cloth_simulation"`
	HairSimulation     bool   `json:"hair_simulation"`

	// Post-processing
	PostProcessingEnabled bool     `json:"post_processing_enabled"`
	Bloom                bool     `json:"bloom"`
	DepthOfField         string   `json:"depth_of_field"`   // none, low, medium, high
	MotionBlur           string   `json:"motion_blur"`     // none, low, medium, high
	ColorGrading         string   `json:"color_grading"`   // none, neutral, cinematic
	Vignette             bool     `json:"vignette"`
	ChromaticAberration  double   `json:"chromatic_aberration"` // 0-1

	// Performance settings
	TargetFPS         int    `json:"target_fps"`         // 30, 60, 120
	VSync             bool   `json:"v_sync"`
	AdaptiveQuality   bool   `json:"adaptive_quality"`   // auto-adjust quality to maintain FPS
	MaxDrawCalls      int    `json:"max_draw_calls"`
	MaxTextureMemory  int    `json:"max_texture_memory"` // MB

	// Shader settings
	ShaderComplexity   string `json:"shader_complexity"`   // simple, standard, complex
	CustomShaders      []string `json:"custom_shaders"`
}

// DeviceAdaptation defines device-specific rendering adaptation.
type DeviceAdaptation struct {
	// Adaptation mode
	AdaptationMode   string `json:"adaptation_mode"`   // auto, manual, profile
	AutoOptimize     bool   `json:"auto_optimize"`     // auto-optimize for device

	// Device profiles
	DeviceProfiles   map[string]DeviceRenderProfile `json:"device_profiles"`

	// Mobile settings
	MobileOptimizations MobileOptimizations `json:"mobile_optimizations"`

	// Desktop settings
	DesktopSettings DesktopSettings `json:"desktop_settings"`

	// VR/AR settings
	VRSettings      VRSettings   `json:"vr_settings"`
	ARSettings      ARSettings   `json:"ar_settings"`

	// Performance scaling
	PerformanceScaling PerformanceScaling `json:"performance_scaling"`

	// Battery optimization
	BatteryOptimization BatteryOptimization `json:"battery_optimization"`
}

// DeviceRenderProfile defines rendering profile for a device type.
type DeviceRenderProfile struct {
	DeviceType        string `json:"device_type"`        // phone, tablet, watch, tv, desktop, vr, ar
	DeviceTier        string `json:"device_tier"`        // low, medium, high, flagship
	MaxQuality        string `json:"max_quality"`        // maximum quality allowed
	RecommendedQuality string `json:"recommended_quality"` // recommended quality setting
	ResolutionScale   double `json:"resolution_scale"`   // 0.5-1.5, render resolution scale
	TextureLimit      int    `json:"texture_limit"`      // max texture resolution
	ShadowResolution  int    `json:"shadow_resolution"`  // shadow map resolution
	ParticleLimit     int    `json:"particle_limit"`     // max particles
	DrawDistance      double `json:"draw_distance"`      // rendering distance

	// Feature support
	SupportsRayTracing   bool `json:"supports_ray_tracing"`
	SupportsComputeShaders bool `json:"supports_compute_shaders"`
	SupportsGeometryShader bool `json:"supports_geometry_shader"`
	SupportsTessellation   bool `json:"supports_tessellation"`

	// Performance thresholds
	MinFPSForQualityUp   int `json:"min_fps_for_quality_up"`
	MaxFPSForQualityDown int `json:"max_fps_for_quality_down"`
}

// MobileOptimizations defines mobile-specific optimizations.
type MobileOptimizations struct {
	// GPU optimization
	GPUPowerMode      string `json:"gpu_power_mode"`      // low, balanced, high
	TextureCompression string `json:"texture_compression"` // etc2, astc, pvrtc
	VertexCompression  bool   `json:"vertex_compression"`
	MeshSimplification double `json:"mesh_simplification"` // 0-1

	// Memory optimization
	TextureStreaming   bool   `json:"texture_streaming"`
	LODBias           double `json:"lod_bias"`           // 0-2, higher = more aggressive LOD
	UnloadUnusedAssets bool   `json:"unload_unused_assets"`
	MemoryBudget      int    `json:"memory_budget"`      // MB

	// Thermal management
	ThermalThrottling  bool   `json:"thermal_throttling"`
	ThermalThreshold   double `json:"thermal_threshold"` // temperature in Celsius
	ReduceOnThermal    bool   `json:"reduce_on_thermal"`

	// Battery management
	BatteryAwareMode   bool   `json:"battery_aware_mode"`
	LowBatteryQuality string `json:"low_battery_quality"` // quality when battery low
	PowerSaveMode     bool   `json:"power_save_mode"`

	// Network optimization for streaming
	NetworkStreaming  bool   `json:"network_streaming"`
	StreamingQuality  string `json:"streaming_quality"`  // low, medium, high, adaptive
	CacheSize         int    `json:"cache_size"`         // MB
	PrefetchDistance  double `json:"prefetch_distance"`
}

// DesktopSettings defines desktop-specific settings.
type DesktopSettings struct {
	// Multi-monitor support
	MultiMonitorEnabled bool     `json:"multi_monitor_enabled"`
	PrimaryMonitor      int      `json:"primary_monitor"`
	MonitorLayout       string   `json:"monitor_layout"`      // single, extended, mirror

	// Window settings
	WindowMode        string `json:"window_mode"`        // windowed, borderless, fullscreen
	ResolutionWidth   int    `json:"resolution_width"`
	ResolutionHeight  int    `json:"resolution_height"`
	RefreshRate       int    `json:"refresh_rate"`
	HDR               bool   `json:"hdr"`
	WideScreenSupport bool   `json:"wide_screen_support"`

	// GPU selection
	GPUSelection      string `json:"gpu_selection"`      // auto, integrated, discrete
	VendorPreference  string `json:"vendor_preference"`  // nvidia, amd, intel, apple
	VRAMLimit         int    `json:"vram_limit"`         // MB, 0 = unlimited

	// Advanced settings
	SLICrossfire      bool   `json:"sli_crossfire"`
	MultiGPU          bool   `json:"multi_gpu"`
	RayTracingCores   int    `json:"ray_tracing_cores"`  // 0 = use all
}

// VRSettings defines VR-specific settings.
type VRSettings struct {
	// VR mode
	VREnabled        bool   `json:"vr_enabled"`
	VRPlatform       string `json:"vr_platform"`       // oculus, steamvr, pico, etc.
	VRSDK           string `json:"vr_sdk"`           // openxr, ovr, steamvr

	// Rendering
	SinglePassRendering bool   `json:"single_pass_rendering"`
	FoveatedRendering   bool   `json:"foveated_rendering"`
	FoveationLevel      double `json:"foveation_level"` // 0-1

	// Performance
	VRTargetFPS        int    `json:"vr_target_fps"`    // 72, 90, 120
	Reprojection       bool   `json:"reprojection"`
	SpaceWarp          bool   `json:"space_warp"`

	// Comfort
	ComfortMode       string `json:"comfort_mode"`       // none, mild, moderate, strong
	VignetteOnMotion  bool   `json:"vignette_on_motion"`
	SnapTurn          bool   `json:"snap_turn"`
	TeleportOnly      bool   `json:"teleport_only"`

	// Hand tracking
	HandTrackingEnabled bool   `json:"hand_tracking_enabled"`
	HandTrackingQuality string `json:"hand_tracking_quality"` // low, medium, high
	GestureRecognition  bool   `json:"gesture_recognition"`

	// Guardian
	GuardianSystem    string `json:"guardian_system"`    // auto, stationary, roomscale
	PlayAreaSize      double `json:"play_area_size"`     // square meters
}

// ARSettings defines AR-specific settings.
type ARSettings struct {
	// AR mode
	AREnabled        bool   `json:"ar_enabled"`
	ARPlatform       string `json:"ar_platform"`       // arcore, arkit, webxr
	ARSessionType    string `json:"ar_session_type"`   // world, face, image, object

	// Tracking
	WorldTracking     bool   `json:"world_tracking"`
	PlaneDetection    bool   `json:"plane_detection"`
	ObjectDetection   bool   `json:"object_detection"`
	ImageTracking     bool   `json:"image_tracking"`
	FaceTracking      bool   `json:"face_tracking"`
	BodyTracking      bool   `json:"body_tracking"`

	// Rendering
	OcclusionEnabled  bool   `json:"occlusion_enabled"`
	OcclusionQuality  string `json:"occlusion_quality"`  // low, medium, high
	LightEstimation   bool   `json:"light_estimation"`
	ShadowCasting     bool   `json:"shadow_casting"`

	// Physics
	ARPhysicsEnabled  bool   `json:"ar_physics_enabled"`
	CollisionDetection string `json:"collision_detection"` // none, basic, accurate

	// Environment
	EnvironmentProbe  bool   `json:"environment_probe"`
	ReflectionProbe   bool   `json:"reflection_probe"`
}

// PerformanceScaling defines performance scaling rules.
type PerformanceScaling struct {
	// Scaling mode
	ScalingMode      string `json:"scaling_mode"`      // fixed, dynamic, predictive
	ScalingAggression double `json:"scaling_aggression"` // 0-1, how aggressively to scale

	// Quality tiers
	QualityTiers     []QualityTier `json:"quality_tiers"`
	CurrentTier      int           `json:"current_tier"`

	// Metrics
	FPSSampleWindow int `json:"fps_sample_window"` // frames to average
	FPSTargetLow    int `json:"fps_target_low"`    // target for low quality
	FPSTargetHigh   int `json:"fps_target_high"`   // target for high quality

	// Hysteresis
	QualityUpDelay   int `json:"quality_up_delay"`   // frames before quality up
	QualityDownDelay int `json:"quality_down_delay"` // frames before quality down
}

// QualityTier defines a quality tier.
type QualityTier struct {
	TierName       string `json:"tier_name"`
	Resolution     double `json:"resolution"`     // 0.5-1.5
	TextureQuality string `json:"texture_quality"`
	ShadowQuality  string `json:"shadow_quality"`
	EffectQuality  string `json:"effect_quality"`
	DrawDistance   double `json:"draw_distance"`
	ParticleCount  int    `json:"particle_count"`
}

// BatteryOptimization defines battery optimization settings.
type BatteryOptimization struct {
	// Optimization level
	OptimizationLevel string `json:"optimization_level"` // none, low, medium, high, extreme

	// Thresholds
	HighBatteryThreshold   double `json:"high_battery_threshold"`   // 0-1, above = full quality
	MediumBatteryThreshold double `json:"medium_battery_threshold"` // 0-1, above = reduced
	LowBatteryThreshold    double `json:"low_battery_threshold"`    // 0-1, above = minimal

	// Quality adjustments
	HighBatteryQuality   string `json:"high_battery_quality"`
	MediumBatteryQuality string `json:"medium_battery_quality"`
	LowBatteryQuality    string `json:"low_battery_quality"`
	CriticalBatteryQuality string `json:"critical_battery_quality"`

	// Power saving features
	ReduceAnimations    bool `json:"reduce_animations"`
	ReduceParticles     bool `json:"reduce_particles"`
	ReducePhysics       bool `json:"reduce_physics"`
	ReduceShadows       bool `json:"reduce_shadows"`
	ReducePostProcess   bool `json:"reduce_post_process"`

	// Charging
	BoostOnCharging    bool   `json:"boost_on_charging"`
	ChargingQuality    string `json:"charging_quality"`
}

// DisplaySync defines display synchronization settings.
type DisplaySync struct {
	// Sync mode
	SyncMode       string `json:"sync_mode"`       // none, state_sync, full_sync
	SyncFrequency  int    `json:"sync_frequency"`  // ms, how often to sync
	SyncPriority   string `json:"sync_priority"`   // realtime, high, normal, low

	// State sync
	StateSyncEnabled   bool     `json:"state_sync_enabled"`
	SyncedStates       []string `json:"synced_states"`    // position, rotation, animation, expression
	InterpolationMode  string   `json:"interpolation_mode"` // none, linear, cubic
	PredictionEnabled  bool     `json:"prediction_enabled"`
	PredictionSteps    int      `json:"prediction_steps"`

	// Conflict resolution
	ConflictResolution string `json:"conflict_resolution"` // last_write, timestamp, priority, merge
	MasterDevice       string `json:"master_device"`       // which device is the master

	// Network settings
	NetworkProtocol string `json:"network_protocol"` // tcp, udp, websocket, quic
	CompressionEnabled bool `json:"compression_enabled"`
	CompressionLevel  int    `json:"compression_level"` // 0-9
	DeltaEncoding     bool   `json:"delta_encoding"`

	// Latency management
	LatencyCompensation bool   `json:"latency_compensation"`
	MaxLatency         int    `json:"max_latency"`    // ms
	BufferSize         int    `json:"buffer_size"`    // frames
	InterpolationBuffer int   `json:"interpolation_buffer"` // frames

	// Offline handling
	OfflineMode       string `json:"offline_mode"`       // pause, local_only, hybrid
	ReconnectAttempts int    `json:"reconnect_attempts"`
	ReconnectDelay    int    `json:"reconnect_delay"`    // ms
	StateRecovery     bool   `json:"state_recovery"`
}

// SceneRenderProfile defines rendering profile for a scene.
type SceneRenderProfile struct {
	SceneName        string `json:"scene_name"`
	SceneType        string `json:"scene_type"`        // home, meeting, outdoor, custom

	// Camera settings
	CameraMode       string `json:"camera_mode"`       // orbit, first_person, third_person, fixed
	CameraFOV        double `json:"camera_fov"`
	CameraNearClip   double `json:"camera_near_clip"`
	CameraFarClip    double `json:"camera_far_clip"`
	CameraPosition   Position3D `json:"camera_position"`
	CameraTarget     Position3D `json:"camera_target"`

	// Lighting
	EnvironmentLight string `json:"environment_light"` // hdri path or preset name
	SunIntensity     double `json:"sun_intensity"`
	SunColor         string `json:"sun_color"`
	AmbientColor     string `json:"ambient_color"`
	FogEnabled       bool   `json:"fog_enabled"`
	FogColor         string `json:"fog_color"`
	FogDensity       double `json:"fog_density"`

	// Background
	BackgroundType   string `json:"background_type"`   // solid, gradient, skybox, image, video
	BackgroundColor  string `json:"background_color"`
	SkyboxPath       string `json:"skybox_path"`
	BackgroundImage  string `json:"background_image"`

	// Avatar placement
	AvatarPosition   Position3D `json:"avatar_position"`
	AvatarRotation   Rotation3D `json:"avatar_rotation"`
	AvatarScale      double `json:"avatar_scale"`
	AvatarVisible    bool   `json:"avatar_visible"`
	AvatarShadow     bool   `json:"avatar_shadow"`

	// Animation
	DefaultAnimation string `json:"default_animation"`
	IdleAnimations   []string `json:"idle_animations"`
	TransitionTime   double `json:"transition_time"`

	// Quality override
	QualityOverride string `json:"quality_override"` // override default quality
	Enabled         bool   `json:"enabled"`
}

// Position3D represents a 3D position.
type Position3D struct {
	X double `json:"x"`
	Y double `json:"y"`
	Z double `json:"z"`
}

// Rotation3D represents a 3D rotation in Euler angles.
type Rotation3D struct {
	X double `json:"x"` // pitch
	Y double `json:"y"` // yaw
	Z double `json:"z"` // roll
}

// DisplayContext provides display-related context.
type DisplayContext struct {
	IdentityID string `json:"identity_id"`

	// Current display info
	CurrentDevice    string `json:"current_device"`
	DeviceType       string `json:"device_type"`
	ScreenResolution Resolution `json:"screen_resolution"`
	DisplayMode      string `json:"display_mode"`

	// Rendering state
	CurrentQuality     string `json:"current_quality"`
	CurrentFPS         double `json:"current_fps"`
	FrameTime          double `json:"frame_time"`
	DrawCalls          int    `json:"draw_calls"`
	MemoryUsage        int64  `json:"memory_usage"`    // bytes
	TextureMemory      int64  `json:"texture_memory"`
	GPUUtilization     double `json:"gpu_utilization"`

	// Performance metrics
	AverageFPS         double `json:"average_fps"`
	MinFPS             double `json:"min_fps"`
	MaxFPS             double `json:"max_fps"`
	DroppedFrames      int    `json:"dropped_frames"`
	ThermalState       string `json:"thermal_state"`   // normal, warm, hot
	BatteryLevel       double `json:"battery_level"`
	BatteryState       string `json:"battery_state"`   // unknown, unplugged, charging, full

	// Scene state
	CurrentScene       string `json:"current_scene"`
	SceneLoadTime      double `json:"scene_load_time"`
	AssetsLoaded       int    `json:"assets_loaded"`
	AssetsTotal        int    `json:"assets_total"`

	// Sync state
	SyncStatus         string `json:"sync_status"`      // synced, syncing, offline, error
	LastSyncTime       time.Time `json:"last_sync_time"`
	SyncLatency        int    `json:"sync_latency"`     // ms
	ConnectedDevices   []string `json:"connected_devices"`

	// Recommendations
	RecommendedQuality string `json:"recommended_quality"`
	RecommendedDevice  string `json:"recommended_device"`
	QualityAdjustmentNeeded bool `json:"quality_adjustment_needed"`

	// Timestamps
	Timestamp          time.Time `json:"timestamp"`
}

// Resolution represents screen resolution.
type Resolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	Scale  double `json:"scale"` // device pixel ratio
}

// ToJSON converts MultiDisplayProfile to JSON string.
func (p *MultiDisplayProfile) ToJSON() string {
	data, _ := json.Marshal(p)
	return string(data)
}

// FromJSON parses MultiDisplayProfile from JSON string.
func (p *MultiDisplayProfile) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), p)
}

// ToJSON converts DisplayContext to JSON string.
func (c *DisplayContext) ToJSON() string {
	data, _ := json.Marshal(c)
	return string(data)
}

// FromJSON parses DisplayContext from JSON string.
func (c *DisplayContext) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), c)
}