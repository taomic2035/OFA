package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	GRPC     GRPCConfig     `yaml:"grpc"`
	REST     RESTConfig     `yaml:"rest"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Etcd     EtcdConfig     `yaml:"etcd"`
	Scheduler SchedulerConfig `yaml:"scheduler"`
	Agent    AgentConfig    `yaml:"agent"`
	TTS      TTSConfig      `yaml:"tts"` // v5.6.2
}

type ServerConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type GRPCConfig struct {
	Address string `yaml:"address"`
}

type RESTConfig struct {
	Address string `yaml:"address"`
}

type DatabaseConfig struct {
	Type     string `yaml:"type"`     // "memory", "sqlite", "postgres", "hybrid"
	Host     string `yaml:"host"`     // PostgreSQL host
	Port     int    `yaml:"port"`     // PostgreSQL port (default: 5432)
	User     string `yaml:"user"`     // PostgreSQL user
	Password string `yaml:"password"` // PostgreSQL password
	Database string `yaml:"database"` // Database name or SQLite file path
	SSLMode  string `yaml:"ssl_mode"` // PostgreSQL SSL mode (disable, require, verify-ca, verify-full)
}

type RedisConfig struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type EtcdConfig struct {
	Endpoints []string `yaml:"endpoints"`
}

type SchedulerConfig struct {
	DefaultStrategy string        `yaml:"default_strategy"`
	MaxConcurrent   int           `yaml:"max_concurrent"`
	TaskTimeout     time.Duration `yaml:"task_timeout"`
}

type AgentConfig struct {
	HeartbeatInterval time.Duration `yaml:"heartbeat_interval"`
	HeartbeatTimeout  time.Duration `yaml:"heartbeat_timeout"`
	OfflineThreshold  time.Duration `yaml:"offline_threshold"`
}

// TTSConfig holds TTS engine configuration (v5.6.2).
type TTSConfig struct {
	PrimaryProvider   string  `yaml:"primary_provider"`
	FallbackProvider  string  `yaml:"fallback_provider"`
	EnableCache       bool    `yaml:"enable_cache"`
	CacheSizeMB       int     `yaml:"cache_size_mb"`
	DefaultVoice      string  `yaml:"default_voice"`
	DefaultFormat     string  `yaml:"default_format"`
	DefaultSampleRate int     `yaml:"default_sample_rate"`
	DefaultRate       float64 `yaml:"default_rate"`
	DefaultPitch      float64 `yaml:"default_pitch"`
	DefaultVolume     float64 `yaml:"default_volume"`
	VolcengineAppID   string  `yaml:"volcengine_app_id"`
	VolcengineToken   string  `yaml:"volcengine_token"`
	DoubaoAppID       string  `yaml:"doubao_app_id"`
	DoubaoToken       string  `yaml:"doubao_token"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set defaults
	if cfg.GRPC.Address == "" {
		cfg.GRPC.Address = ":9090"
	}
	if cfg.REST.Address == "" {
		cfg.REST.Address = ":8080"
	}
	if cfg.Agent.HeartbeatInterval == 0 {
		cfg.Agent.HeartbeatInterval = 30 * time.Second
	}
	if cfg.Agent.HeartbeatTimeout == 0 {
		cfg.Agent.HeartbeatTimeout = 60 * time.Second
	}
	if cfg.Scheduler.DefaultStrategy == "" {
		cfg.Scheduler.DefaultStrategy = "hybrid"
	}
	if cfg.Scheduler.MaxConcurrent == 0 {
		cfg.Scheduler.MaxConcurrent = 1000
	}

	return &cfg, nil
}