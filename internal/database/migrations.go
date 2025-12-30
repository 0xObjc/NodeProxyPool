package database

import (
	"database/sql"
	"fmt"
)

// runMigrations 创建 schema 与索引
func runMigrations(conn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS subscription_sources (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			url TEXT NOT NULL,
			enabled INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			last_fetch_at DATETIME,
			last_fetch_status TEXT,
			last_fetch_error TEXT,
			node_count INTEGER DEFAULT 0
		);`,
		`CREATE INDEX IF NOT EXISTS idx_subscription_enabled ON subscription_sources(enabled);`,
		`CREATE INDEX IF NOT EXISTS idx_subscription_updated ON subscription_sources(updated_at);`,
		`CREATE TABLE IF NOT EXISTS nodes (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			server TEXT NOT NULL,
			port INTEGER NOT NULL,
			raw_config_json TEXT NOT NULL,
			delay INTEGER NOT NULL DEFAULT -1,
			last_check DATETIME,
			available INTEGER NOT NULL DEFAULT 0,
			active_count INTEGER NOT NULL DEFAULT 0,
			total_used INTEGER NOT NULL DEFAULT 0,
			subscription_source_id INTEGER,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (subscription_source_id) REFERENCES subscription_sources(id) ON DELETE SET NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_nodes_type ON nodes(type);`,
		`CREATE INDEX IF NOT EXISTS idx_nodes_available ON nodes(available);`,
		`CREATE INDEX IF NOT EXISTS idx_nodes_delay ON nodes(delay);`,
		`CREATE INDEX IF NOT EXISTS idx_nodes_type_available ON nodes(type, available);`,
		`CREATE INDEX IF NOT EXISTS idx_nodes_subscription ON nodes(subscription_source_id);`,
		`CREATE TABLE IF NOT EXISTS system_config (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			value_type TEXT NOT NULL,
			description TEXT,
			category TEXT NOT NULL,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, stmt := range stmts {
		if _, err := conn.Exec(stmt); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

// SeedData 首次导入使用的数据
type SeedData struct {
	Sources []SubscriptionSource
	Configs map[string]ConfigValue
}

// seedFromConfig 将初始配置导入数据库
func seedFromConfig(conn *sql.DB, seed *SeedData) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 订阅源
	subStmt, err := tx.Prepare(`INSERT OR IGNORE INTO subscription_sources(name, url, enabled) VALUES (?, ?, ?);`)
	if err != nil {
		return err
	}
	defer subStmt.Close()

	for _, src := range seed.Sources {
		if _, err := subStmt.Exec(src.Name, src.URL, boolToInt(src.Enabled)); err != nil {
			return fmt.Errorf("seed subscription %s failed: %w", src.Name, err)
		}
	}

	// 系统配置
	confStmt, err := tx.Prepare(`INSERT OR REPLACE INTO system_config(key, value, value_type, description, category, updated_at) VALUES(?, ?, ?, ?, ?, CURRENT_TIMESTAMP);`)
	if err != nil {
		return err
	}
	defer confStmt.Close()

	for key, val := range seed.Configs {
		if _, err := confStmt.Exec(key, val.Value, val.Type, val.Description, val.Category); err != nil {
			return fmt.Errorf("seed config %s failed: %w", key, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
