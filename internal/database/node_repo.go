package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"proxyPool/internal/node"
)

// NodeRepository 节点持久化
type NodeRepository struct {
	db *DB
}

// NewNodeRepository 构造
func NewNodeRepository(db *DB) *NodeRepository {
	return &NodeRepository{db: db}
}

// Upsert 插入或更新单个节点
func (r *NodeRepository) Upsert(n *node.Node, subscriptionID *int64) error {
	rec, err := toRecord(n, subscriptionID)
	if err != nil {
		return err
	}
	_, err = r.db.Conn().Exec(`
		INSERT OR REPLACE INTO nodes
		(id, name, type, server, port, raw_config_json, delay, last_check, available, active_count, total_used, subscription_source_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, COALESCE((SELECT created_at FROM nodes WHERE id = ?), CURRENT_TIMESTAMP), CURRENT_TIMESTAMP)
	`, rec.ID, rec.Name, rec.Type, rec.Server, rec.Port, rec.RawConfigJSON, rec.Delay, rec.LastCheck, boolToInt(rec.Available), rec.ActiveCount, rec.TotalUsed, rec.SubscriptionSourceID, rec.ID)
	return err
}

// UpsertBatch 批量插入/更新
func (r *NodeRepository) UpsertBatch(nodes []*node.Node, subscriptionID *int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO nodes
		(id, name, type, server, port, raw_config_json, delay, last_check, available, active_count, total_used, subscription_source_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, COALESCE((SELECT created_at FROM nodes WHERE id = ?), CURRENT_TIMESTAMP), CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, n := range nodes {
		rec, err := toRecord(n, subscriptionID)
		if err != nil {
			return err
		}
		if _, err := stmt.Exec(rec.ID, rec.Name, rec.Type, rec.Server, rec.Port, rec.RawConfigJSON, rec.Delay, rec.LastCheck, boolToInt(rec.Available), rec.ActiveCount, rec.TotalUsed, rec.SubscriptionSourceID, rec.ID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetAll 返回所有节点
func (r *NodeRepository) GetAll() ([]*node.Node, error) {
	rows, err := r.db.Conn().Query(`SELECT id, name, type, server, port, raw_config_json, delay, last_check, available, active_count, total_used, subscription_source_id, created_at, updated_at FROM nodes`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*node.Node
	for rows.Next() {
		rec, err := scanNode(rows)
		if err != nil {
			return nil, err
		}
		n, err := rec.ToNode()
		if err != nil {
			return nil, err
		}
		result = append(result, n)
	}
	return result, rows.Err()
}

// GetByID 查询单个节点
func (r *NodeRepository) GetByID(id string) (*node.Node, error) {
	row := r.db.Conn().QueryRow(`SELECT id, name, type, server, port, raw_config_json, delay, last_check, available, active_count, total_used, subscription_source_id, created_at, updated_at FROM nodes WHERE id = ?`, id)
	rec, err := scanNode(row)
	if err != nil {
		return nil, err
	}
	return rec.ToNode()
}

// UpdateHealthStatus 更新健康检查结果
func (r *NodeRepository) UpdateHealthStatus(id string, delay int, available bool) error {
	_, err := r.db.Conn().Exec(`UPDATE nodes SET delay = ?, available = ?, last_check = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, delay, boolToInt(available), time.Now(), id)
	return err
}

// UpdateUsageStats 更新使用统计
func (r *NodeRepository) UpdateUsageStats(id string, activeCount, totalUsed int) error {
	_, err := r.db.Conn().Exec(`UPDATE nodes SET active_count = ?, total_used = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, activeCount, totalUsed, id)
	return err
}

// DeleteBySubscription 删除订阅源对应节点
func (r *NodeRepository) DeleteBySubscription(subscriptionID int64) error {
	_, err := r.db.Conn().Exec(`DELETE FROM nodes WHERE subscription_source_id = ?`, subscriptionID)
	return err
}

// DeleteStale 删除更新时间早于截止时间的节点
func (r *NodeRepository) DeleteStale(cutoff time.Time) error {
	_, err := r.db.Conn().Exec(`DELETE FROM nodes WHERE updated_at < ?`, cutoff)
	return err
}

func toRecord(n *node.Node, subscriptionID *int64) (*NodeRecord, error) {
	raw := n.RawConfig
	if raw == nil {
		raw = map[string]interface{}{
			"name":   n.Name,
			"type":   n.Type,
			"server": n.Server,
			"port":   n.Port,
		}
	}
	rawJSON, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("marshal raw config: %w", err)
	}

	rec := &NodeRecord{
		ID:                   n.ID,
		Name:                 n.Name,
		Type:                 n.Type,
		Server:               n.Server,
		Port:                 n.Port,
		RawConfigJSON:        string(rawJSON),
		Delay:                n.Delay,
		Available:            n.Available,
		ActiveCount:          n.ActiveCount,
		TotalUsed:            n.TotalUsed,
		SubscriptionSourceID: subscriptionID,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
	if !n.LastCheck.IsZero() {
		rec.LastCheck = &n.LastCheck
	}
	return rec, nil
}

func scanNode(scanner interface {
	Scan(dest ...any) error
}) (*NodeRecord, error) {
	var rec NodeRecord
	var lastCheck sql.NullTime
	var available int
	var subID sql.NullInt64

	if err := scanner.Scan(&rec.ID, &rec.Name, &rec.Type, &rec.Server, &rec.Port, &rec.RawConfigJSON, &rec.Delay, &lastCheck, &available, &rec.ActiveCount, &rec.TotalUsed, &subID, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
		return nil, err
	}
	rec.Available = available == 1
	if lastCheck.Valid {
		rec.LastCheck = &lastCheck.Time
	}
	if subID.Valid {
		rec.SubscriptionSourceID = &subID.Int64
	}
	return &rec, nil
}
