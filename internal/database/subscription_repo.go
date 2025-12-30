package database

import (
	"database/sql"
	"fmt"
	"time"
)

// SubscriptionRepository 订阅源 CRUD
type SubscriptionRepository struct {
	db *DB
}

// NewSubscriptionRepository 构造
func NewSubscriptionRepository(db *DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// Create 新增订阅源
func (r *SubscriptionRepository) Create(src *SubscriptionSource) (int64, error) {
	res, err := r.db.Conn().Exec(
		`INSERT INTO subscription_sources(name, url, enabled, created_at, updated_at) VALUES(?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		src.Name, src.URL, boolToInt(src.Enabled),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Update 更新订阅源
func (r *SubscriptionRepository) Update(id int64, src *SubscriptionSource) error {
	_, err := r.db.Conn().Exec(
		`UPDATE subscription_sources SET name = ?, url = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		src.Name, src.URL, boolToInt(src.Enabled), id,
	)
	return err
}

// Delete 删除订阅源
func (r *SubscriptionRepository) Delete(id int64) error {
	_, err := r.db.Conn().Exec(`DELETE FROM subscription_sources WHERE id = ?`, id)
	return err
}

// GetByID 获取单个订阅源
func (r *SubscriptionRepository) GetByID(id int64) (*SubscriptionSource, error) {
	row := r.db.Conn().QueryRow(`SELECT id, name, url, enabled, created_at, updated_at, last_fetch_at, last_fetch_status, last_fetch_error, node_count FROM subscription_sources WHERE id = ?`, id)
	return scanSubscription(row)
}

// GetAll 获取全部订阅源
func (r *SubscriptionRepository) GetAll() ([]*SubscriptionSource, error) {
	rows, err := r.db.Conn().Query(`SELECT id, name, url, enabled, created_at, updated_at, last_fetch_at, last_fetch_status, last_fetch_error, node_count FROM subscription_sources ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*SubscriptionSource
	for rows.Next() {
		src, err := scanSubscription(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, src)
	}
	return result, rows.Err()
}

// GetEnabled 获取启用的订阅源
func (r *SubscriptionRepository) GetEnabled() ([]*SubscriptionSource, error) {
	rows, err := r.db.Conn().Query(`SELECT id, name, url, enabled, created_at, updated_at, last_fetch_at, last_fetch_status, last_fetch_error, node_count FROM subscription_sources WHERE enabled = 1 ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*SubscriptionSource
	for rows.Next() {
		src, err := scanSubscription(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, src)
	}
	return result, rows.Err()
}

// UpdateFetchStatus 更新抓取状态
func (r *SubscriptionRepository) UpdateFetchStatus(id int64, status, errMsg string, nodeCount int) error {
	_, err := r.db.Conn().Exec(
		`UPDATE subscription_sources SET last_fetch_status = ?, last_fetch_error = ?, last_fetch_at = ?, node_count = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		status, errMsg, time.Now(), nodeCount, id,
	)
	return err
}

func scanSubscription(scanner interface {
	Scan(dest ...any) error
}) (*SubscriptionSource, error) {
	var src SubscriptionSource
	var enabled int
	var lastFetch sql.NullTime
	var lastStatus sql.NullString
	var lastErr sql.NullString

	if err := scanner.Scan(
		&src.ID, &src.Name, &src.URL, &enabled,
		&src.CreatedAt, &src.UpdatedAt, &lastFetch, &lastStatus, &lastErr, &src.NodeCount,
	); err != nil {
		return nil, fmt.Errorf("scan subscription: %w", err)
	}

	src.Enabled = enabled == 1
	if lastFetch.Valid {
		src.LastFetchAt = &lastFetch.Time
	}
	if lastStatus.Valid {
		src.LastFetchStatus = lastStatus.String
	}
	if lastErr.Valid {
		src.LastFetchError = lastErr.String
	}
	return &src, nil
}
