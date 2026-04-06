package store

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/ofa/center/internal/config"
	"github.com/ofa/center/internal/models"

	pb "github.com/ofa/center/proto"

	_ "github.com/lib/pq"
)

// PostgreSQLStore - PostgreSQL 存储实现
type PostgreSQLStore struct {
	db    *sql.DB
	cache *MemoryCache // In-memory cache for hot data
	mu    sync.RWMutex
}

// MemoryCache provides in-memory caching layer
type MemoryCache struct {
	online    sync.Map // agentID -> bool
	resources sync.Map // agentID -> *models.ResourceUsage
}

// PostgreSQLConfig - PostgreSQL 配置
type PostgreSQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// NewPostgreSQLStore 创建 PostgreSQL 存储
func NewPostgreSQLStore(cfg PostgreSQLConfig) (*PostgreSQLStore, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 验证连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	store := &PostgreSQLStore{
		db:    db,
		cache: &MemoryCache{},
	}

	// 初始化表
	if err := store.initTables(); err != nil {
		return nil, fmt.Errorf("failed to init tables: %w", err)
	}

	return store, nil
}

// initTables 初始化数据库表
func (s *PostgreSQLStore) initTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS agents (
			id VARCHAR(64) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			type INTEGER NOT NULL,
			status INTEGER NOT NULL,
			device_info JSONB,
			capabilities JSONB,
			last_seen TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id VARCHAR(64) PRIMARY KEY,
			parent_task_id VARCHAR(64),
			skill_id VARCHAR(255),
			input BYTEA,
			output BYTEA,
			status INTEGER NOT NULL,
			priority INTEGER DEFAULT 0,
			target_agent VARCHAR(64),
			source_agent VARCHAR(64),
			error TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			started_at TIMESTAMP,
			completed_at TIMESTAMP,
			duration_ms BIGINT,
			timeout_ms BIGINT,
			metadata JSONB
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id VARCHAR(64) PRIMARY KEY,
			from_agent VARCHAR(64) NOT NULL,
			to_agent VARCHAR(64) NOT NULL,
			action VARCHAR(255),
			payload BYTEA,
			timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			ttl INTEGER,
			delivered BOOLEAN DEFAULT FALSE,
			headers JSONB
		)`,
		`CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status)`,
		`CREATE INDEX IF NOT EXISTS idx_agents_type ON agents(type)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_agent ON tasks(target_agent)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_to ON messages(to_agent)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_delivered ON messages(delivered)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

// Close 关闭连接
func (s *PostgreSQLStore) Close() error {
	return s.db.Close()
}

// === Agent Operations ===

// SaveAgent 保存 Agent
func (s *PostgreSQLStore) SaveAgent(ctx context.Context, agent *models.Agent) error {
	query := `
		INSERT INTO agents (id, name, type, status, device_info, capabilities, last_seen, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			type = EXCLUDED.type,
			status = EXCLUDED.status,
			device_info = EXCLUDED.device_info,
			capabilities = EXCLUDED.capabilities,
			last_seen = EXCLUDED.last_seen,
			updated_at = EXCLUDED.updated_at
	`

	_, err := s.db.ExecContext(ctx, query,
		agent.ID, agent.Name, agent.Type, agent.Status,
		agent.DeviceInfo, agent.Capabilities, agent.LastSeen, time.Now(),
	)

	return err
}

// GetAgent 获取 Agent
func (s *PostgreSQLStore) GetAgent(ctx context.Context, id string) (*models.Agent, error) {
	query := `SELECT id, name, type, status, device_info, capabilities, last_seen, created_at, updated_at
			  FROM agents WHERE id = $1`

	agent := &models.Agent{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&agent.ID, &agent.Name, &agent.Type, &agent.Status,
		&agent.DeviceInfo, &agent.Capabilities, &agent.LastSeen,
		&agent.CreatedAt, &agent.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("agent not found: %s", id)
	}

	return agent, err
}

// ListAgents 列出 Agents
func (s *PostgreSQLStore) ListAgents(ctx context.Context, agentType pb.AgentType, status pb.AgentStatus, page, pageSize int) ([]*models.Agent, int, error) {
	// Count query
	countQuery := `SELECT COUNT(*) FROM agents WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if agentType != 0 {
		countQuery += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, agentType)
		argIdx++
	}
	if status != 0 {
		countQuery += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
	}

	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Select query
	selectQuery := `SELECT id, name, type, status, device_info, capabilities, last_seen, created_at, updated_at
					FROM agents WHERE 1=1`

	if agentType != 0 {
		selectQuery += fmt.Sprintf(" AND type = $%d", argIdx)
	}
	if status != 0 {
		selectQuery += fmt.Sprintf(" AND status = $%d", argIdx)
	}

	selectQuery += " ORDER BY updated_at DESC"
	if pageSize > 0 {
		selectQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, page*pageSize)
	}

	rows, err := s.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var agents []*models.Agent
	for rows.Next() {
		agent := &models.Agent{}
		if err := rows.Scan(
			&agent.ID, &agent.Name, &agent.Type, &agent.Status,
			&agent.DeviceInfo, &agent.Capabilities, &agent.LastSeen,
			&agent.CreatedAt, &agent.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		agents = append(agents, agent)
	}

	return agents, total, nil
}

// DeleteAgent 删除 Agent
func (s *PostgreSQLStore) DeleteAgent(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM agents WHERE id = $1", id)
	return err
}

// === Task Operations ===

// SaveTask 保存 Task
func (s *PostgreSQLStore) SaveTask(ctx context.Context, task *models.Task) error {
	query := `
		INSERT INTO tasks (id, parent_task_id, skill_id, input, output, status, priority,
			target_agent, source_agent, error, created_at, started_at, completed_at, duration_ms, timeout_ms, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (id) DO UPDATE SET
			output = EXCLUDED.output,
			status = EXCLUDED.status,
			error = EXCLUDED.error,
			started_at = EXCLUDED.started_at,
			completed_at = EXCLUDED.completed_at,
			duration_ms = EXCLUDED.duration_ms
	`

	_, err := s.db.ExecContext(ctx, query,
		task.ID, task.ParentTaskID, task.SkillID, task.Input, task.Output,
		task.Status, task.Priority, task.TargetAgent, task.SourceAgent,
		task.Error, task.CreatedAt, task.StartedAt, task.CompletedAt,
		task.DurationMS, task.TimeoutMS, task.Metadata,
	)

	return err
}

// GetTask 获取 Task
func (s *PostgreSQLStore) GetTask(ctx context.Context, id string) (*models.Task, error) {
	query := `SELECT id, parent_task_id, skill_id, input, output, status, priority,
			  target_agent, source_agent, error, created_at, started_at, completed_at, duration_ms, timeout_ms, metadata
			  FROM tasks WHERE id = $1`

	task := &models.Task{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.ParentTaskID, &task.SkillID, &task.Input, &task.Output,
		&task.Status, &task.Priority, &task.TargetAgent, &task.SourceAgent,
		&task.Error, &task.CreatedAt, &task.StartedAt, &task.CompletedAt,
		&task.DurationMS, &task.TimeoutMS, &task.Metadata,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found: %s", id)
	}

	return task, err
}

// ListTasks 列出 Tasks
func (s *PostgreSQLStore) ListTasks(ctx context.Context, status pb.TaskStatus, agentID string, page, pageSize int) ([]*models.Task, int, error) {
	countQuery := `SELECT COUNT(*) FROM tasks WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if status != 0 {
		countQuery += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	if agentID != "" {
		countQuery += fmt.Sprintf(" AND target_agent = $%d", argIdx)
		args = append(args, agentID)
	}

	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	selectQuery := `SELECT id, parent_task_id, skill_id, input, output, status, priority,
					target_agent, source_agent, error, created_at, started_at, completed_at, duration_ms, timeout_ms, metadata
					FROM tasks WHERE 1=1`

	if status != 0 {
		selectQuery += fmt.Sprintf(" AND status = $%d", argIdx)
	}
	if agentID != "" {
		selectQuery += fmt.Sprintf(" AND target_agent = $%d", argIdx)
	}

	selectQuery += " ORDER BY created_at DESC"
	if pageSize > 0 {
		selectQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, page*pageSize)
	}

	rows, err := s.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		if err := rows.Scan(
			&task.ID, &task.ParentTaskID, &task.SkillID, &task.Input, &task.Output,
			&task.Status, &task.Priority, &task.TargetAgent, &task.SourceAgent,
			&task.Error, &task.CreatedAt, &task.StartedAt, &task.CompletedAt,
			&task.DurationMS, &task.TimeoutMS, &task.Metadata,
		); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, task)
	}

	return tasks, total, nil
}

// === Message Operations ===

// SaveMessage 保存 Message
func (s *PostgreSQLStore) SaveMessage(ctx context.Context, msg *models.Message) error {
	query := `
		INSERT INTO messages (id, from_agent, to_agent, action, payload, timestamp, ttl, delivered, headers)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := s.db.ExecContext(ctx, query,
		msg.ID, msg.FromAgent, msg.ToAgent, msg.Action, msg.Payload,
		msg.Timestamp, msg.TTL, false, msg.Headers,
	)

	return err
}

// GetPendingMessages 获取待发送消息
func (s *PostgreSQLStore) GetPendingMessages(ctx context.Context, agentID string) ([]*models.Message, error) {
	query := `SELECT id, from_agent, to_agent, action, payload, timestamp, ttl, headers
			  FROM messages WHERE to_agent = $1 AND delivered = FALSE
			  ORDER BY timestamp ASC`

	rows, err := s.db.QueryContext(ctx, query, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		msg := &models.Message{}
		if err := rows.Scan(
			&msg.ID, &msg.FromAgent, &msg.ToAgent, &msg.Action, &msg.Payload,
			&msg.Timestamp, &msg.TTL, &msg.Headers,
		); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// MarkMessageDelivered 标记消息已送达
func (s *PostgreSQLStore) MarkMessageDelivered(ctx context.Context, msgID string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE messages SET delivered = TRUE WHERE id = $1", msgID)
	return err
}

// === Cache Operations ===

// SetAgentOnline sets agent online status
func (s *PostgreSQLStore) SetAgentOnline(ctx context.Context, agentID string, ttl time.Duration) error {
	s.cache.online.Store(agentID, true)
	return nil
}

// IsAgentOnline checks if agent is online
func (s *PostgreSQLStore) IsAgentOnline(ctx context.Context, agentID string) bool {
	v, ok := s.cache.online.Load(agentID)
	return ok && v.(bool)
}

// SetAgentResources stores agent resources
func (s *PostgreSQLStore) SetAgentResources(ctx context.Context, agentID string, resources *models.ResourceUsage) error {
	s.cache.resources.Store(agentID, resources)
	return nil
}

// GetAgentResources retrieves agent resources
func (s *PostgreSQLStore) GetAgentResources(ctx context.Context, agentID string) (*models.ResourceUsage, error) {
	v, ok := s.cache.resources.Load(agentID)
	if !ok {
		return nil, ErrNotFound
	}
	return v.(*models.ResourceUsage), nil
}

// NewPostgreSQLStoreFromConfig creates PostgreSQL store from config
func NewPostgreSQLStoreFromConfig(cfg *config.Config) (*PostgreSQLStore, error) {
	return NewPostgreSQLStore(PostgreSQLConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Database: cfg.Database.Database,
		SSLMode:  "disable",
	})
}