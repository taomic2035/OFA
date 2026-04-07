// Package identity - PostgreSQL 身份存储 (v2.7.0)
//
// Center 是永远在线的灵魂载体，需要持久化存储确保灵魂永存。
// PostgreSQL 提供企业级数据持久化，支持身份的永久保存。
package identity

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ofa/center/internal/models"
	_ "github.com/lib/pq"
)

// PostgresConfig PostgreSQL 配置
type PostgresConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	User            string        `json:"user"`
	Password        string        `json:"password"`
	Database        string        `json:"database"`
	SSLMode         string        `json:"ssl_mode"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	Schema          string        `json:"schema"`
}

// DefaultPostgresConfig 默认配置
func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "ofa",
		Password:        "",
		Database:        "ofa_center",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		Schema:          "public",
	}
}

// PostgresStore PostgreSQL 身份存储
type PostgresStore struct {
	mu     sync.RWMutex
	db     *sql.DB
	config PostgresConfig
	cache  map[string]*models.PersonalIdentity // 本地缓存
}

// NewPostgresStore 创建 PostgreSQL 存储
func NewPostgresStore(config PostgresConfig) (*PostgresStore, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &PostgresStore{
		db:     db,
		config: config,
		cache:  make(map[string]*models.PersonalIdentity),
	}

	// 初始化表结构
	if err := store.initSchema(ctx); err != nil {
		return nil, fmt.Errorf("failed to init schema: %w", err)
	}

	return store, nil
}

// initSchema 初始化数据库表结构
func (s *PostgresStore) initSchema(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS identities (
		id VARCHAR(64) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		nickname VARCHAR(255),
		avatar_url TEXT,
		birthday DATE,
		gender VARCHAR(32),
		language VARCHAR(32),
		region VARCHAR(64),
		timezone VARCHAR(64),
		personality JSONB,
		value_system JSONB,
		interests JSONB,
		voice_profile JSONB,
		writing_style JSONB,
		decision_style VARCHAR(32),
		privacy_level VARCHAR(32),
		version BIGINT DEFAULT 1,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		stability_score FLOAT DEFAULT 0.0,
		metadata JSONB
	);

	CREATE INDEX IF NOT EXISTS idx_identities_name ON identities(name);
	CREATE INDEX IF NOT EXISTS idx_identities_updated_at ON identities(updated_at);
	CREATE INDEX IF NOT EXISTS idx_identities_version ON identities(version);
	`

	_, err := s.db.ExecContext(ctx, schema)
	return err
}

// SaveIdentity 保存身份（灵魂持久化）
func (s *PostgresStore) SaveIdentity(ctx context.Context, identity *models.PersonalIdentity) error {
	if identity == nil || identity.ID == "" {
		return fmt.Errorf("invalid identity")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 序列化复杂字段
	personalityJSON, _ := json.Marshal(identity.Personality)
	valueSystemJSON, _ := json.Marshal(identity.ValueSystem)
	interestsJSON, _ := json.Marshal(identity.Interests)
	voiceProfileJSON, _ := json.Marshal(identity.VoiceProfile)
	writingStyleJSON, _ := json.Marshal(identity.WritingStyle)
	metadataJSON, _ := json.Marshal(identity.Metadata)

	// 更新时间
	identity.UpdatedAt = time.Now()
	if identity.CreatedAt.IsZero() {
		identity.CreatedAt = time.Now()
	}

	// 使用 UPSERT (INSERT ... ON CONFLICT)
	query := `
		INSERT INTO identities (
			id, name, nickname, avatar_url, birthday, gender, language, region, timezone,
			personality, value_system, interests, voice_profile, writing_style,
			decision_style, privacy_level, version, created_at, updated_at,
			stability_score, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			nickname = EXCLUDED.nickname,
			avatar_url = EXCLUDED.avatar_url,
			birthday = EXCLUDED.birthday,
			gender = EXCLUDED.gender,
			language = EXCLUDED.language,
			region = EXCLUDED.region,
			timezone = EXCLUDED.timezone,
			personality = EXCLUDED.personality,
			value_system = EXCLUDED.value_system,
			interests = EXCLUDED.interests,
			voice_profile = EXCLUDED.voice_profile,
			writing_style = EXCLUDED.writing_style,
			decision_style = EXCLUDED.decision_style,
			privacy_level = EXCLUDED.privacy_level,
			version = EXCLUDED.version,
			updated_at = EXCLUDED.updated_at,
			stability_score = EXCLUDED.stability_score,
			metadata = EXCLUDED.metadata
	`

	_, err := s.db.ExecContext(ctx, query,
		identity.ID, identity.Name, identity.Nickname, identity.AvatarURL,
		identity.Birthday, identity.Gender, identity.Language, identity.Region, identity.Timezone,
		personalityJSON, valueSystemJSON, interestsJSON, voiceProfileJSON, writingStyleJSON,
		identity.DecisionStyle, identity.PrivacyLevel, identity.Version,
		identity.CreatedAt, identity.UpdatedAt,
		identity.StabilityScore, metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to save identity: %w", err)
	}

	// 更新缓存
	s.cache[identity.ID] = identity

	return nil
}

// GetIdentity 获取身份
func (s *PostgresStore) GetIdentity(ctx context.Context, id string) (*models.PersonalIdentity, error) {
	if id == "" {
		return nil, fmt.Errorf("empty id")
	}

	// 先查缓存
	s.mu.RLock()
	if identity, ok := s.cache[id]; ok {
		s.mu.RUnlock()
		return identity, nil
	}
	s.mu.RUnlock()

	// 从数据库查询
	query := `
		SELECT id, name, nickname, avatar_url, birthday, gender, language, region, timezone,
		       personality, value_system, interests, voice_profile, writing_style,
		       decision_style, privacy_level, version, created_at, updated_at,
		       stability_score, metadata
		FROM identities WHERE id = $1
	`

	row := s.db.QueryRowContext(ctx, query, id)

	identity := &models.PersonalIdentity{}
	var personalityJSON, valueSystemJSON, interestsJSON, voiceProfileJSON, writingStyleJSON, metadataJSON []byte

	err := row.Scan(
		&identity.ID, &identity.Name, &identity.Nickname, &identity.AvatarURL,
		&identity.Birthday, &identity.Gender, &identity.Language, &identity.Region, &identity.Timezone,
		&personalityJSON, &valueSystemJSON, &interestsJSON, &voiceProfileJSON, &writingStyleJSON,
		&identity.DecisionStyle, &identity.PrivacyLevel, &identity.Version,
		&identity.CreatedAt, &identity.UpdatedAt,
		&identity.StabilityScore, &metadataJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("identity not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get identity: %w", err)
	}

	// 解析复杂字段
	if len(personalityJSON) > 0 {
		json.Unmarshal(personalityJSON, &identity.Personality)
	}
	if len(valueSystemJSON) > 0 {
		json.Unmarshal(valueSystemJSON, &identity.ValueSystem)
	}
	if len(interestsJSON) > 0 {
		json.Unmarshal(interestsJSON, &identity.Interests)
	}
	if len(voiceProfileJSON) > 0 {
		json.Unmarshal(voiceProfileJSON, &identity.VoiceProfile)
	}
	if len(writingStyleJSON) > 0 {
		json.Unmarshal(writingStyleJSON, &identity.WritingStyle)
	}
	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &identity.Metadata)
	}

	// 缓存
	s.mu.Lock()
	s.cache[id] = identity
	s.mu.Unlock()

	return identity, nil
}

// DeleteIdentity 删除身份
func (s *PostgresStore) DeleteIdentity(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("empty id")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	query := `DELETE FROM identities WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete identity: %w", err)
	}

	// 从缓存移除
	delete(s.cache, id)

	return nil
}

// ListIdentities 列出所有身份
func (s *PostgresStore) ListIdentities(ctx context.Context, page, pageSize int) ([]*models.PersonalIdentity, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 获取总数
	var total int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM identities`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count identities: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := `
		SELECT id, name, nickname, avatar_url, birthday, gender, language, region, timezone,
		       personality, value_system, interests, voice_profile, writing_style,
		       decision_style, privacy_level, version, created_at, updated_at,
		       stability_score, metadata
		FROM identities
		ORDER BY updated_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list identities: %w", err)
	}
	defer rows.Close()

	identities := make([]*models.PersonalIdentity, 0, pageSize)
	for rows.Next() {
		identity := &models.PersonalIdentity{}
		var personalityJSON, valueSystemJSON, interestsJSON, voiceProfileJSON, writingStyleJSON, metadataJSON []byte

		err := rows.Scan(
			&identity.ID, &identity.Name, &identity.Nickname, &identity.AvatarURL,
			&identity.Birthday, &identity.Gender, &identity.Language, &identity.Region, &identity.Timezone,
			&personalityJSON, &valueSystemJSON, &interestsJSON, &voiceProfileJSON, &writingStyleJSON,
			&identity.DecisionStyle, &identity.PrivacyLevel, &identity.Version,
			&identity.CreatedAt, &identity.UpdatedAt,
			&identity.StabilityScore, &metadataJSON,
		)
		if err != nil {
			continue
		}

		// 解析复杂字段
		if len(personalityJSON) > 0 {
			json.Unmarshal(personalityJSON, &identity.Personality)
		}
		if len(valueSystemJSON) > 0 {
			json.Unmarshal(valueSystemJSON, &identity.ValueSystem)
		}
		if len(interestsJSON) > 0 {
			json.Unmarshal(interestsJSON, &identity.Interests)
		}
		if len(voiceProfileJSON) > 0 {
			json.Unmarshal(voiceProfileJSON, &identity.VoiceProfile)
		}
		if len(writingStyleJSON) > 0 {
			json.Unmarshal(writingStyleJSON, &identity.WritingStyle)
		}
		if len(metadataJSON) > 0 {
			json.Unmarshal(metadataJSON, &identity.Metadata)
		}

		identities = append(identities, identity)
	}

	return identities, total, nil
}

// Close 关闭连接
func (s *PostgresStore) Close() error {
	return s.db.Close()
}

// === 纠偏查询 ===

// GetIdentitiesByVersion 获取指定版本的身份
func (s *PostgresStore) GetIdentitiesByVersion(ctx context.Context, minVersion int64) ([]*models.PersonalIdentity, error) {
	query := `
		SELECT id, name, version, updated_at
		FROM identities
		WHERE version >= $1
		ORDER BY version DESC
	`

	rows, err := s.db.QueryContext(ctx, query, minVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to query by version: %w", err)
	}
	defer rows.Close()

	identities := make([]*models.PersonalIdentity, 0)
	for rows.Next() {
		identity := &models.PersonalIdentity{}
		err := rows.Scan(&identity.ID, &identity.Name, &identity.Version, &identity.UpdatedAt)
		if err != nil {
			continue
		}
		identities = append(identities, identity)
	}

	return identities, nil
}

// GetLatestVersion 获取最新版本号
func (s *PostgresStore) GetLatestVersion(ctx context.Context, id string) (int64, error) {
	query := `SELECT version FROM identities WHERE id = $1`
	var version int64
	err := s.db.QueryRowContext(ctx, query, id).Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return version, nil
}

// === 备份相关 ===

// BackupIdentity 备份身份到 JSON
func (s *PostgresStore) BackupIdentity(ctx context.Context, id string) ([]byte, error) {
	identity, err := s.GetIdentity(ctx, id)
	if err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(identity, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to backup identity: %w", err)
	}

	return data, nil
}

// RestoreIdentity 从备份恢复身份
func (s *PostgresStore) RestoreIdentity(ctx context.Context, data []byte) error {
	var identity models.PersonalIdentity
	if err := json.Unmarshal(data, &identity); err != nil {
		return fmt.Errorf("failed to parse backup: %w", err)
	}

	return s.SaveIdentity(ctx, &identity)
}