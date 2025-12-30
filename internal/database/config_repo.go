package database

import "database/sql"

// ConfigRepository system_config 操作
type ConfigRepository struct {
	db *DB
}

// NewConfigRepository 构造
func NewConfigRepository(db *DB) *ConfigRepository {
	return &ConfigRepository{db: db}
}

// Get 返回指定 key 的值
func (r *ConfigRepository) Get(key string) (string, error) {
	row := r.db.Conn().QueryRow(`SELECT value FROM system_config WHERE key = ?`, key)
	var value string
	if err := row.Scan(&value); err != nil {
		return "", err
	}
	return value, nil
}

// GetByCategory 返回指定分类下所有配置
func (r *ConfigRepository) GetByCategory(category string) (map[string]string, error) {
	rows, err := r.db.Conn().Query(`SELECT key, value FROM system_config WHERE category = ?`, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		result[k] = v
	}
	return result, rows.Err()
}

// GetAll 返回所有配置
func (r *ConfigRepository) GetAll() (map[string]string, error) {
	rows, err := r.db.Conn().Query(`SELECT key, value FROM system_config`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		result[k] = v
	}
	return result, rows.Err()
}

// Set 插入/更新单个配置
func (r *ConfigRepository) Set(key, value, valueType, category string) error {
	_, err := r.db.Conn().Exec(`INSERT OR REPLACE INTO system_config(key, value, value_type, category, updated_at) VALUES(?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		key, value, valueType, category)
	return err
}

// SetBatch 批量写入配置
func (r *ConfigRepository) SetBatch(configs map[string]ConfigValue) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT OR REPLACE INTO system_config(key, value, value_type, description, category, updated_at) VALUES(?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for key, cfg := range configs {
		if _, err := stmt.Exec(key, cfg.Value, cfg.Type, cfg.Description, cfg.Category); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// InitializeDefaults 使用 YAML 配置写入默认值(如果不存在则插入)
// List 列出所有配置行
func (r *ConfigRepository) List() ([]SystemConfig, error) {
	rows, err := r.db.Conn().Query(`SELECT key, value, value_type, description, category, updated_at FROM system_config`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []SystemConfig
	for rows.Next() {
		var cfg SystemConfig
		if err := rows.Scan(&cfg.Key, &cfg.Value, &cfg.ValueType, &cfg.Description, &cfg.Category, &cfg.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, cfg)
	}
	return result, rows.Err()
}

// GetRaw 返回完整行
func (r *ConfigRepository) GetRaw(key string) (*SystemConfig, error) {
	row := r.db.Conn().QueryRow(`SELECT key, value, value_type, description, category, updated_at FROM system_config WHERE key = ?`, key)
	var cfg SystemConfig
	if err := row.Scan(&cfg.Key, &cfg.Value, &cfg.ValueType, &cfg.Description, &cfg.Category, &cfg.UpdatedAt); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// InitializeDefaultsIfMissing 在缺失时写入
func (r *ConfigRepository) InitializeDefaultsIfMissing(configs map[string]ConfigValue) error {
	existing, err := r.GetAll()
	if err != nil {
		return err
	}
	toWrite := make(map[string]ConfigValue)
	for key, val := range configs {
		if _, ok := existing[key]; !ok {
			toWrite[key] = val
		}
	}
	if len(toWrite) == 0 {
		return nil
	}
	return r.SetBatch(toWrite)
}

// RawTx 允许外部在事务中执行自定义逻辑
func (r *ConfigRepository) RawTx(fn func(tx *sql.Tx) error) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
