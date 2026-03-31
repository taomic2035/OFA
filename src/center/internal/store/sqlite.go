package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ofa/center/internal/config"
	"github.com/ofa/center/internal/models"

	pb "github.com/ofa/center/proto"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStore implements StoreInterface using SQLite
type SQLiteStore struct {
	db       *sql.DB
	cfg      *config.Config
	cache    *MemoryCache // In-memory cache for hot data
	mu       sync.RWMutex
}

// MemoryCache provides in-memory caching layer
type MemoryCache struct {
	online    sync.Map // agentID -> bool
	resources sync.Map // agentID -> *models.ResourceUsage
}

// NewSQLiteStore creates a new SQLite store
func NewSQLiteStore(cfg *config.Config) (*SQLiteStore, error) {
	dbPath := cfg.Database.Database
	if dbPath == "" {
		dbPath = "data/ofa.db"
	}

	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	store := &SQLiteStore{
		db:    db,
		cfg:   cfg,
		cache: &MemoryCache{},
	}

	// Run migrations
	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return store, nil
}

// migrate creates database tables
func (s *SQLiteStore) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS agents (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type INTEGER NOT NULL DEFAULT 1,
			status INTEGER NOT NULL DEFAULT 1,
			capabilities TEXT,
			metadata TEXT,
			last_heartbeat DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			parent_id TEXT,
			source_agent TEXT,
			target_agent TEXT,
			skill_id TEXT NOT NULL,
			input BLOB,
			output BLOB,
			status INTEGER NOT NULL DEFAULT 1,
			priority INTEGER NOT NULL DEFAULT 0,
			error TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			started_at DATETIME,
			completed_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			msg_type INTEGER NOT NULL DEFAULT 1,
			from_agent TEXT NOT NULL,
			to_agent TEXT NOT NULL,
			action TEXT NOT NULL,
			payload BLOB,
			status INTEGER NOT NULL DEFAULT 1,
			ttl INTEGER DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			delivered_at DATETIME
		)`,
		`CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status)`,
		`CREATE INDEX IF NOT EXISTS idx_agents_type ON agents(type)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_agent ON tasks(target_agent, source_agent)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_to_agent ON messages(to_agent, status)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// ===== Agent Store =====

// SaveAgent saves or updates an agent
func (s *SQLiteStore) SaveAgent(ctx context.Context, agent *models.Agent) error {
	capabilitiesJSON, _ := json.Marshal(agent.Capabilities)
	metadataJSON, _ := json.Marshal(agent.Metadata)

	query := `INSERT INTO agents (id, name, type, status, capabilities, metadata, last_heartbeat, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
		name = excluded.name,
		type = excluded.type,
		status = excluded.status,
		capabilities = excluded.capabilities,
		metadata = excluded.metadata,
		last_heartbeat = excluded.last_heartbeat,
		updated_at = excluded.updated_at`

	_, err := s.db.ExecContext(ctx, query,
		agent.ID, agent.Name, int(agent.Type), int(agent.Status),
		capabilitiesJSON, metadataJSON, agent.LastHeartbeat, time.Now())

	return err
}

// GetAgent retrieves an agent by ID
func (s *SQLiteStore) GetAgent(ctx context.Context, id string) (*models.Agent, error) {
	query := `SELECT id, name, type, status, capabilities, metadata, last_heartbeat, created_at, updated_at
		FROM agents WHERE id = ?`

	var agent models.Agent
	var capabilitiesJSON, metadataJSON []byte
	var lastHeartbeat sql.NullTime

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&agent.ID, &agent.Name, &agent.Type, &agent.Status,
		&capabilitiesJSON, &metadataJSON, &lastHeartbeat,
		&agent.CreatedAt, &agent.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if lastHeartbeat.Valid {
		agent.LastHeartbeat = lastHeartbeat.Time
	}

	json.Unmarshal(capabilitiesJSON, &agent.Capabilities)
	json.Unmarshal(metadataJSON, &agent.Metadata)

	return &agent, nil
}

// ListAgents lists agents with filters
func (s *SQLiteStore) ListAgents(ctx context.Context, agentType pb.AgentType, status pb.AgentStatus, page, pageSize int) ([]*models.Agent, int, error) {
	// Build query
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if agentType != 0 {
		whereClause += " AND type = ?"
		args = append(args, int(agentType))
	}
	if status != 0 {
		whereClause += " AND status = ?"
		args = append(args, int(status))
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM agents " + whereClause
	var total int
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := 0
	if page > 0 && pageSize > 0 {
		offset = (page - 1) * pageSize
	}

	query := `SELECT id, name, type, status, capabilities, metadata, last_heartbeat, created_at, updated_at
		FROM agents ` + whereClause + ` ORDER BY updated_at DESC`

	if pageSize > 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var agents []*models.Agent
	for rows.Next() {
		var agent models.Agent
		var capabilitiesJSON, metadataJSON []byte
		var lastHeartbeat sql.NullTime

		err := rows.Scan(
			&agent.ID, &agent.Name, &agent.Type, &agent.Status,
			&capabilitiesJSON, &metadataJSON, &lastHeartbeat,
			&agent.CreatedAt, &agent.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}

		if lastHeartbeat.Valid {
			agent.LastHeartbeat = lastHeartbeat.Time
		}

		json.Unmarshal(capabilitiesJSON, &agent.Capabilities)
		json.Unmarshal(metadataJSON, &agent.Metadata)
		agents = append(agents, &agent)
	}

	return agents, total, nil
}

// DeleteAgent deletes an agent
func (s *SQLiteStore) DeleteAgent(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM agents WHERE id = ?", id)
	return err
}

// ===== Task Store =====

// SaveTask saves or updates a task
func (s *SQLiteStore) SaveTask(ctx context.Context, task *models.Task) error {
	query := `INSERT INTO tasks (id, parent_id, source_agent, target_agent, skill_id, input, output, status, priority, error, started_at, completed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
		output = excluded.output,
		status = excluded.status,
		error = excluded.error,
		started_at = excluded.started_at,
		completed_at = excluded.completed_at`

	var startedAt, completedAt interface{}
	if !task.StartedAt.IsZero() {
		startedAt = task.StartedAt
	}
	if !task.CompletedAt.IsZero() {
		completedAt = task.CompletedAt
	}

	_, err := s.db.ExecContext(ctx, query,
		task.ID, task.ParentTaskID, task.SourceAgent, task.TargetAgent,
		task.SkillID, task.Input, task.Output, int(task.Status), task.Priority,
		task.Error, startedAt, completedAt)

	return err
}

// GetTask retrieves a task by ID
func (s *SQLiteStore) GetTask(ctx context.Context, id string) (*models.Task, error) {
	query := `SELECT id, parent_id, source_agent, target_agent, skill_id, input, output, status, priority, error, created_at, started_at, completed_at
		FROM tasks WHERE id = ?`

	var task models.Task
	var startedAt, completedAt sql.NullTime

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.ParentTaskID, &task.SourceAgent, &task.TargetAgent,
		&task.SkillID, &task.Input, &task.Output, &task.Status, &task.Priority,
		&task.Error, &task.CreatedAt, &startedAt, &completedAt)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if startedAt.Valid {
		task.StartedAt = startedAt.Time
	}
	if completedAt.Valid {
		task.CompletedAt = completedAt.Time
	}

	return &task, nil
}

// ListTasks lists tasks with filters
func (s *SQLiteStore) ListTasks(ctx context.Context, status pb.TaskStatus, agentID string, page, pageSize int) ([]*models.Task, int, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if status != 0 {
		whereClause += " AND status = ?"
		args = append(args, int(status))
	}
	if agentID != "" {
		whereClause += " AND (target_agent = ? OR source_agent = ?)"
		args = append(args, agentID, agentID)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM tasks " + whereClause
	var total int
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := 0
	if page > 0 && pageSize > 0 {
		offset = (page - 1) * pageSize
	}

	query := `SELECT id, parent_id, source_agent, target_agent, skill_id, input, output, status, priority, error, created_at, started_at, completed_at
		FROM tasks ` + whereClause + ` ORDER BY created_at DESC`

	if pageSize > 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		var task models.Task
		var startedAt, completedAt sql.NullTime

		err := rows.Scan(
			&task.ID, &task.ParentTaskID, &task.SourceAgent, &task.TargetAgent,
			&task.SkillID, &task.Input, &task.Output, &task.Status, &task.Priority,
			&task.Error, &task.CreatedAt, &startedAt, &completedAt)
		if err != nil {
			return nil, 0, err
		}

		if startedAt.Valid {
			task.StartedAt = startedAt.Time
		}
		if completedAt.Valid {
			task.CompletedAt = completedAt.Time
		}
		tasks = append(tasks, &task)
	}

	return tasks, total, nil
}

// ===== Message Store =====

// SaveMessage saves a message
func (s *SQLiteStore) SaveMessage(ctx context.Context, msg *models.Message) error {
	query := `INSERT INTO messages (id, msg_type, from_agent, to_agent, action, payload, status, ttl)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.ExecContext(ctx, query,
		msg.ID, int(msg.Type), msg.FromAgent, msg.ToAgent,
		msg.Action, msg.Payload, int(msg.Status), msg.TTL)

	return err
}

// GetPendingMessages retrieves undelivered messages for an agent
func (s *SQLiteStore) GetPendingMessages(ctx context.Context, agentID string) ([]*models.Message, error) {
	query := `SELECT id, msg_type, from_agent, to_agent, action, payload, status, ttl, created_at
		FROM messages WHERE to_agent = ? AND status = 1 ORDER BY created_at ASC`

	rows, err := s.db.QueryContext(ctx, query, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(
			&msg.ID, &msg.Type, &msg.FromAgent, &msg.ToAgent,
			&msg.Action, &msg.Payload, &msg.Status, &msg.TTL, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}

// MarkMessageDelivered marks a message as delivered
func (s *SQLiteStore) MarkMessageDelivered(ctx context.Context, msgID string) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE messages SET status = 2, delivered_at = ? WHERE id = ?",
		time.Now(), msgID)
	return err
}

// ===== Cache Operations =====

// SetAgentOnline sets agent online status
func (s *SQLiteStore) SetAgentOnline(ctx context.Context, agentID string, ttl time.Duration) error {
	s.cache.online.Store(agentID, true)
	return nil
}

// IsAgentOnline checks if agent is online
func (s *SQLiteStore) IsAgentOnline(ctx context.Context, agentID string) bool {
	v, ok := s.cache.online.Load(agentID)
	return ok && v.(bool)
}

// SetAgentResources stores agent resources
func (s *SQLiteStore) SetAgentResources(ctx context.Context, agentID string, resources *models.ResourceUsage) error {
	s.cache.resources.Store(agentID, resources)
	return nil
}

// GetAgentResources retrieves agent resources
func (s *SQLiteStore) GetAgentResources(ctx context.Context, agentID string) (*models.ResourceUsage, error) {
	v, ok := s.cache.resources.Load(agentID)
	if !ok {
		return nil, ErrNotFound
	}
	return v.(*models.ResourceUsage), nil
}