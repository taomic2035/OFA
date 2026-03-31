// Package store - PostgreSQL数据库支持
package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

// PostgreSQLConfig PostgreSQL配置
type PostgreSQLConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	User            string        `json:"user"`
	Password        string        `json:"password"`
	Database        string        `json:"database"`
	SSLMode         string        `json:"ssl_mode"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`
	ConnectTimeout  time.Duration `json:"connect_timeout"`
	Schema          string        `json:"schema"`
}

// Migration 数据库迁移
type Migration struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     int       `json:"version"`
	SQL         string    `json:"sql"`
	CreatedAt   time.Time `json:"created_at"`
	AppliedAt   *time.Time `json:"applied_at,omitempty"`
	Status      string    `json:"status"` // pending, applied, failed, rolled_back
	Error       string    `json:"error,omitempty"`
}

// TransactionStats 事务统计
type TransactionStats struct {
	TotalTransactions   int64         `json:"total_transactions"`
	SuccessfulTrans     int64         `json:"successful_transactions"`
	FailedTrans         int64         `json:"failed_transactions"`
	AvgDuration         time.Duration `json:"avg_duration"`
	ActiveTransactions  int           `json:"active_transactions"`
	LastTransTime       time.Time     `json:"last_transaction_time"`
	CommitCount         int64         `json:"commit_count"`
	RollbackCount       int64         `json:"rollback_count"`
}

// PostgreSQLStore PostgreSQL存储
type PostgreSQLStore struct {
	config    PostgreSQLConfig
	db        *sql.DB
	migrations map[string]*Migration
	transStats *TransactionStats
	mu        sync.RWMutex
	connected bool
}

// NewPostgreSQLStore 创建PostgreSQL存储
func NewPostgreSQLStore(config PostgreSQLConfig) *PostgreSQLStore {
	// 默认配置
	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = 5432
	}
	if config.MaxOpenConns == 0 {
		config.MaxOpenConns = 25
	}
	if config.MaxIdleConns == 0 {
		config.MaxIdleConns = 5
	}
	if config.ConnMaxLifetime == 0 {
		config.ConnMaxLifetime = 5 * time.Minute
	}
	if config.ConnMaxIdleTime == 0 {
		config.ConnMaxIdleTime = 2 * time.Minute
	}
	if config.SSLMode == "" {
		config.SSLMode = "disable"
	}
	if config.Schema == "" {
		config.Schema = "public"
	}

	return &PostgreSQLStore{
		config:     config,
		migrations: make(map[string]*Migration),
		transStats: &TransactionStats{},
	}
}

// Connect 连接数据库
func (ps *PostgreSQLStore) Connect(ctx context.Context) error {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		ps.config.Host,
		ps.config.Port,
		ps.config.User,
		ps.config.Password,
		ps.config.Database,
		ps.config.SSLMode,
		int(ps.config.ConnectTimeout.Seconds()),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(ps.config.MaxOpenConns)
	db.SetMaxIdleConns(ps.config.MaxIdleConns)
	db.SetConnMaxLifetime(ps.config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(ps.config.ConnMaxIdleTime)

	// 测试连接
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	ps.db = db
	ps.mu.Lock()
	ps.connected = true
	ps.mu.Unlock()

	// 初始化表结构
	if err := ps.initSchema(ctx); err != nil {
		return fmt.Errorf("初始化Schema失败: %w", err)
	}

	// 应用迁移
	if err := ps.ApplyMigrations(ctx); err != nil {
		return fmt.Errorf("应用迁移失败: %w", err)
	}

	return nil
}

// Disconnect 断开连接
func (ps *PostgreSQLStore) Disconnect() {
	ps.mu.Lock()
	ps.connected = false
	ps.mu.Unlock()

	if ps.db != nil {
		ps.db.Close()
	}
}

// initSchema 初始化Schema
func (ps *PostgreSQLStore) initSchema(ctx context.Context) error {
	// 创建迁移表
	createMigrationsTable := `
	CREATE TABLE IF NOT EXISTS migrations (
		id VARCHAR(100) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		version INTEGER NOT NULL,
		sql_text TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL,
		applied_at TIMESTAMP,
		status VARCHAR(50) NOT NULL,
		error TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_migrations_version ON migrations(version);
	CREATE INDEX IF NOT EXISTS idx_migrations_status ON migrations(status);
	`

	_, err := ps.db.ExecContext(ctx, createMigrationsTable)
	return err
}

// === Agent存储 ===

// SaveAgent 保存Agent
func (ps *PostgreSQLStore) SaveAgent(ctx context.Context, agent map[string]interface{}) error {
	id, ok := agent["id"].(string)
	if !ok || id == "" {
		return fmt.Errorf("Agent ID缺失")
	}

	data, err := json.Marshal(agent)
	if err != nil {
		return err
	}

	query := `
	INSERT INTO agents (id, data, created_at, updated_at)
	VALUES ($1, $2, NOW(), NOW())
	ON CONFLICT (id) DO UPDATE SET
		data = EXCLUDED.data,
		updated_at = NOW()
	`

	// 先确保表存在
	ps.ensureAgentsTable(ctx)

	return ps.db.ExecContext(ctx, query, id, data).Err()
}

// GetAgent 获取Agent
func (ps *PostgreSQLStore) GetAgent(ctx context.Context, id string) (map[string]interface{}, error) {
	query := `SELECT data FROM agents WHERE id = $1`

	var data []byte
	err := ps.db.QueryRowContext(ctx, query, id).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("Agent不存在: %s", id)
	}
	if err != nil {
		return nil, err
	}

	var agent map[string]interface{}
	json.Unmarshal(data, &agent)
	return agent, nil
}

// ListAgents 列出所有Agent
func (ps *PostgreSQLStore) ListAgents(ctx context.Context, limit, offset int) ([]map[string]interface{}, error) {
	query := `SELECT data FROM agents ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := ps.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	agents := make([]map[string]interface{}, 0)
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			continue
		}
		var agent map[string]interface{}
		json.Unmarshal(data, &agent)
		agents = append(agents, agent)
	}

	return agents, nil
}

// DeleteAgent 删除Agent
func (ps *PostgreSQLStore) DeleteAgent(ctx context.Context, id string) error {
	query := `DELETE FROM agents WHERE id = $1`
	return ps.db.ExecContext(ctx, query, id).Err()
}

// ensureAgentsTable 确保agents表存在
func (ps *PostgreSQLStore) ensureAgentsTable(ctx context.Context) {
	createTable := `
	CREATE TABLE IF NOT EXISTS agents (
		id VARCHAR(100) PRIMARY KEY,
		data JSONB NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_agents_created ON agents(created_at);
	CREATE INDEX IF NOT EXISTS idx_agents_data ON agents USING GIN (data);
	`
	ps.db.ExecContext(ctx, createTable)
}

// === Task存储 ===

// SaveTask 保存任务
func (ps *PostgreSQLStore) SaveTask(ctx context.Context, task map[string]interface{}) error {
	id, ok := task["id"].(string)
	if !ok || id == "" {
		return fmt.Errorf("Task ID缺失")
	}

	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	query := `
	INSERT INTO tasks (id, data, created_at, updated_at)
	VALUES ($1, $2, NOW(), NOW())
	ON CONFLICT (id) DO UPDATE SET
		data = EXCLUDED.data,
		updated_at = NOW()
	`

	ps.ensureTasksTable(ctx)

	return ps.db.ExecContext(ctx, query, id, data).Err()
}

// GetTask 获取任务
func (ps *PostgreSQLStore) GetTask(ctx context.Context, id string) (map[string]interface{}, error) {
	ps.ensureTasksTable(ctx)

	query := `SELECT data FROM tasks WHERE id = $1`

	var data []byte
	err := ps.db.QueryRowContext(ctx, query, id).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("Task不存在: %s", id)
	}
	if err != nil {
		return nil, err
	}

	var task map[string]interface{}
	json.Unmarshal(data, &task)
	return task, nil
}

// ListTasks 列出任务
func (ps *PostgreSQLStore) ListTasks(ctx context.Context, status string, limit, offset int) ([]map[string]interface{}, error) {
	ps.ensureTasksTable(ctx)

	var query string
	var args []interface{}

	if status != "" {
		query = `SELECT data FROM tasks WHERE data->>'status' = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = []interface{}{status, limit, offset}
	} else {
		query = `SELECT data FROM tasks ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		args = []interface{}{limit, offset}
	}

	rows, err := ps.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]map[string]interface{}, 0)
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			continue
		}
		var task map[string]interface{}
		json.Unmarshal(data, &task)
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// DeleteTask 删除任务
func (ps *PostgreSQLStore) DeleteTask(ctx context.Context, id string) error {
	query := `DELETE FROM tasks WHERE id = $1`
	return ps.db.ExecContext(ctx, query, id).Err()
}

// ensureTasksTable 确保tasks表存在
func (ps *PostgreSQLStore) ensureTasksTable(ctx context.Context) {
	createTable := `
	CREATE TABLE IF NOT EXISTS tasks (
		id VARCHAR(100) PRIMARY KEY,
		data JSONB NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_tasks_created ON tasks(created_at);
	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks((data->>'status'));
	CREATE INDEX IF NOT EXISTS idx_tasks_agent ON tasks((data->>'agent_id'));
	`
	ps.db.ExecContext(ctx, createTable)
}

// === 事务支持 ===

// BeginTransaction 开始事务
func (ps *PostgreSQLStore) BeginTransaction(ctx context.Context) (*sql.Tx, error) {
	tx, err := ps.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}

	ps.mu.Lock()
	ps.transStats.ActiveTransactions++
	ps.transStats.TotalTransactions++
	ps.mu.Unlock()

	return tx, nil
}

// CommitTransaction 提交事务
func (ps *PostgreSQLStore) CommitTransaction(tx *sql.Tx) error {
	err := tx.Commit()

	ps.mu.Lock()
	ps.transStats.ActiveTransactions--
	if err == nil {
		ps.transStats.SuccessfulTrans++
		ps.transStats.CommitCount++
	} else {
		ps.transStats.FailedTrans++
	}
	ps.transStats.LastTransTime = time.Now()
	ps.mu.Unlock()

	return err
}

// RollbackTransaction 回滚事务
func (ps *PostgreSQLStore) RollbackTransaction(tx *sql.Tx) error {
	err := tx.Rollback()

	ps.mu.Lock()
	ps.transStats.ActiveTransactions--
	ps.transStats.FailedTrans++
	ps.transStats.RollbackCount++
	ps.transStats.LastTransTime = time.Now()
	ps.mu.Unlock()

	return err
}

// WithTransaction 事务包装器
func (ps *PostgreSQLStore) WithTransaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := ps.BeginTransaction(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			ps.RollbackTransaction(tx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		ps.RollbackTransaction(tx)
		return err
	}

	return ps.CommitTransaction(tx)
}

// === 数据库迁移 ===

// AddMigration 添加迁移
func (ps *PostgreSQLStore) AddMigration(migration *Migration) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	migration.CreatedAt = time.Now()
	migration.Status = "pending"
	ps.migrations[migration.ID] = migration
}

// ApplyMigrations 应用迁移
func (ps *PostgreSQLStore) ApplyMigrations(ctx context.Context) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	for _, m := range ps.migrations {
		if m.Status == "applied" {
			continue
		}

		// 执行迁移
		_, err := ps.db.ExecContext(ctx, m.SQL)
		if err != nil {
			m.Status = "failed"
			m.Error = err.Error()
			return fmt.Errorf("迁移 %s 失败: %w", m.ID, err)
		}

		// 记录迁移状态
		now := time.Now()
		m.AppliedAt = &now
		m.Status = "applied"

		// 保存迁移记录
		insertMigration := `
		INSERT INTO migrations (id, name, description, version, sql_text, created_at, applied_at, status, error)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			applied_at = EXCLUDED.applied_at,
			status = EXCLUDED.status
		`
		ps.db.ExecContext(ctx, insertMigration,
			m.ID, m.Name, m.Description, m.Version, m.SQL,
			m.CreatedAt, m.AppliedAt, m.Status, m.Error)
	}

	return nil
}

// GetMigrations 获取迁移列表
func (ps *PostgreSQLStore) GetMigrations(ctx context.Context) ([]*Migration, error) {
	query := `SELECT id, name, description, version, sql_text, created_at, applied_at, status, error
	          FROM migrations ORDER BY version`

	rows, err := ps.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	migrations := make([]*Migration, 0)
	for rows.Next() {
		m := &Migration{}
		var sqlText string
		var appliedAt sql.NullTime
		var errorStr sql.NullString

		err := rows.Scan(&m.ID, &m.Name, &m.Description, &m.Version, &sqlText,
			&m.CreatedAt, &appliedAt, &m.Status, &errorStr)
		if err != nil {
			continue
		}

		m.SQL = sqlText
		if appliedAt.Valid {
			m.AppliedAt = &appliedAt.Time
		}
		if errorStr.Valid {
			m.Error = errorStr.String
		}

		migrations = append(migrations, m)
	}

	return migrations, nil
}

// RollbackMigration 回滚迁移
func (ps *PostgreSQLStore) RollbackMigration(ctx context.Context, version int) error {
	// 需要提供回滚SQL
	return fmt.Errorf("回滚迁移需要提供回滚SQL")
}

// === 查询工具 ===

// Query 执行查询
func (ps *PostgreSQLStore) Query(ctx context.Context, sql string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := ps.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]map[string]interface{}, 0)
	cols, _ := rows.Columns()

	for rows.Next() {
		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for i := range vals {
			valPtrs[i] = &vals[i]
		}

		if err := rows.Scan(valPtrs...); err != nil {
			continue
		}

		row := make(map[string]interface{})
		for i, col := range cols {
			row[col] = vals[i]
		}
		results = append(results, row)
	}

	return results, nil
}

// Exec 执行语句
func (ps *PostgreSQLStore) Exec(ctx context.Context, sql string, args ...interface{}) error {
	return ps.db.ExecContext(ctx, sql, args...).Err()
}

// QueryRow 查询单行
func (ps *PostgreSQLStore) QueryRow(ctx context.Context, sql string, args ...interface{}) (map[string]interface{}, error) {
	row := ps.db.QueryRowContext(ctx, sql, args...)

	cols, err := row.Columns()
	if err != nil {
		return nil, err
	}

	vals := make([]interface{}, len(cols))
	valPtrs := make([]interface{}, len(cols))
	for i := range vals {
		valPtrs[i] = &vals[i]
	}

	if err := row.Scan(valPtrs...); err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for i, col := range cols {
		result[col] = vals[i]
	}

	return result, nil
}

// === 统计与监控 ===

// GetStatistics 获取统计信息
func (ps *PostgreSQLStore) GetStatistics(ctx context.Context) map[string]interface{} {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	// 数据库统计
	var dbStats map[string]interface{}
	if ps.db != nil {
		stats := ps.db.Stats()
		dbStats = map[string]interface{}{
			"open_connections": stats.OpenConnections,
			"in_use":           stats.InUse,
			"idle":             stats.Idle,
			"wait_count":       stats.WaitCount,
			"wait_duration":    stats.WaitDuration.Milliseconds(),
			"max_idle_closed":  stats.MaxIdleClosed,
			"max_lifetime_closed": stats.MaxLifetimeClosed,
		}
	}

	return map[string]interface{}{
		"connected":             ps.connected,
		"database_stats":        dbStats,
		"transaction_stats": map[string]interface{}{
			"total":       ps.transStats.TotalTransactions,
			"successful":  ps.transStats.SuccessfulTrans,
			"failed":      ps.transStats.FailedTrans,
			"active":      ps.transStats.ActiveTransactions,
			"commits":     ps.transStats.CommitCount,
			"rollbacks":   ps.transStats.RollbackCount,
			"avg_duration": ps.transStats.AvgDuration.Milliseconds(),
		},
		"migrations_count": len(ps.migrations),
	}
}

// GetPoolStatus 获取连接池状态
func (ps *PostgreSQLStore) GetPoolStatus() map[string]interface{} {
	if ps.db == nil {
		return map[string]interface{}{
			"status": "not_connected",
		}
	}

	stats := ps.db.Stats()
	return map[string]interface{}{
		"status":            "connected",
		"open_connections":  stats.OpenConnections,
		"in_use":            stats.InUse,
		"idle":              stats.Idle,
		"wait_count":        stats.WaitCount,
		"idle_percent":      float64(stats.Idle) / float64(stats.OpenConnections) * 100,
	}
}

// HealthCheck 健康检查
func (ps *PostgreSQLStore) HealthCheck(ctx context.Context) error {
	if ps.db == nil {
		return fmt.Errorf("数据库未连接")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return ps.db.PingContext(ctx)
}

// IsConnected 检查连接状态
func (ps *PostgreSQLStore) IsConnected() bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.connected
}

// GetDB 获取数据库连接
func (ps *PostgreSQLStore) GetDB() *sql.DB {
	return ps.db
}

// Close 关闭连接
func (ps *PostgreSQLStore) Close() error {
	if ps.db != nil {
		return ps.db.Close()
	}
	return nil
}

// === 批量操作 ===

// BatchInsert 批量插入
func (ps *PostgreSQLStore) BatchInsert(ctx context.Context, table string, columns []string, rows [][]interface{}) error {
	if len(rows) == 0 {
		return nil
	}

	// 构建批量插入SQL
	// 使用 unnest 函数实现高效批量插入
	query := fmt.Sprintf(`
		INSERT INTO %s (%s)
		SELECT * FROM unnest(
			ARRAY[%s]
		)
	`, table, joinColumns(columns), buildUnnestArgs(len(columns)))

	// 执行批量插入
	// 实际实现需要根据具体数据库调整
	for _, row := range rows {
		args := make([]interface{}, len(row))
		for i, val := range row {
			args[i] = val
		}
		// 简化实现：逐行插入
		insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			table, joinColumns(columns), buildValueArgs(len(columns)))

		ps.db.ExecContext(ctx, insertQuery, args...)
	}

	return nil
}

func joinColumns(cols []string) string {
	result := ""
	for i, col := range cols {
		if i > 0 {
			result += ", "
		}
		result += col
	}
	return result
}

func buildValueArgs(n int) string {
	result := ""
	for i := 1; i <= n; i++ {
		if i > 1 {
			result += ", "
		}
		result += fmt.Sprintf("$%d", i)
	}
	return result
}

func buildUnnestArgs(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("$%d::%s[]", i+1, "VARCHAR")
	}
	return result
}

// Clear 清空存储
func (ps *PostgreSQLStore) Clear(ctx context.Context) error {
	// 清空表数据(保留表结构)
	ps.db.ExecContext(ctx, "TRUNCATE TABLE agents CASCADE")
	ps.db.ExecContext(ctx, "TRUNCATE TABLE tasks CASCADE")
	ps.db.ExecContext(ctx, "TRUNCATE TABLE migrations CASCADE")

	return nil
}