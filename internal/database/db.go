package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB 封装 SQLite 连接
type DB struct {
	conn     *sql.DB
	mu       sync.RWMutex
	firstRun bool
}

// New 创建数据库连接并配置 WAL/连接池
func New(dbPath string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create db dir: %w", err)
	}

	_, err := os.Stat(dbPath)
	firstRun := os.IsNotExist(err)

	conn, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(time.Hour)

	return &DB{conn: conn, firstRun: firstRun}, nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	return db.conn.Close()
}

// Begin 开启事务
func (db *DB) Begin() (*sql.Tx, error) {
	return db.conn.Begin()
}

// Conn 返回底层连接(只读)
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// Migrate 执行 schema 创建
func (db *DB) Migrate() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	return runMigrations(db.conn)
}

// IsFirstRun 是否首次运行(数据库文件此前不存在)
func (db *DB) IsFirstRun() bool {
	return db.firstRun
}

// SeedFromConfig 首次启动时从配置写入默认数据
func (db *DB) SeedFromConfig(seed *SeedData) error {
	return seedFromConfig(db.conn, seed)
}
