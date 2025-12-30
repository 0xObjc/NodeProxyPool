package config

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"proxyPool/internal/database"
)

// DynamicConfig 支持运行时重载与持久化
type DynamicConfig struct {
	current    *Config
	db         *database.DB
	configRepo *database.ConfigRepository
	mu         sync.RWMutex

	onProxyConfigChange        func(*ProxyConfig)
	onHealthCheckConfigChange  func(*HealthCheckConfig)
	onSubscriptionConfigChange func(*SubscriptionConfig)
}

// NewDynamicConfig 从数据库加载配置
func NewDynamicConfig(db *database.DB) (*DynamicConfig, error) {
	dc := &DynamicConfig{
		db:         db,
		configRepo: database.NewConfigRepository(db),
	}
	if err := dc.Reload(); err != nil {
		return nil, err
	}
	return dc, nil
}

// Current 返回当前配置快照(只读引用)
func (dc *DynamicConfig) Current() *Config {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	return dc.current
}

// Reload 从数据库重载配置
func (dc *DynamicConfig) Reload() error {
	cfg, err := dc.loadConfigFromDB()
	if err != nil {
		return fmt.Errorf("load config from db failed: %w", err)
	}
	if err := validate(cfg); err != nil {
		return fmt.Errorf("invalid config from db: %w", err)
	}

	dc.mu.Lock()
	dc.current = cfg
	dc.mu.Unlock()

	if dc.onProxyConfigChange != nil {
		dc.onProxyConfigChange(&cfg.Proxy)
	}
	if dc.onHealthCheckConfigChange != nil {
		dc.onHealthCheckConfigChange(&cfg.HealthCheck)
	}
	if dc.onSubscriptionConfigChange != nil {
		dc.onSubscriptionConfigChange(&cfg.Subscription)
	}
	return nil
}

// UpdateProxyConfig 更新代理配置
func (dc *DynamicConfig) UpdateProxyConfig(cfg *ProxyConfig) error {
	merged := dc.clone()
	merged.Proxy = *cfg
	if err := validate(merged); err != nil {
		return fmt.Errorf("invalid proxy config: %w", err)
	}

	configs := map[string]database.ConfigValue{
		"proxy.port_range.min":   {Value: fmt.Sprintf("%d", cfg.PortRange.Min), Type: "int", Category: "proxy"},
		"proxy.port_range.max":   {Value: fmt.Sprintf("%d", cfg.PortRange.Max), Type: "int", Category: "proxy"},
		"proxy.default_ttl":      {Value: fmt.Sprintf("%d", int64(cfg.DefaultTTL/time.Second)), Type: "duration", Category: "proxy"},
		"proxy.max_instances":    {Value: fmt.Sprintf("%d", cfg.MaxInstances), Type: "int", Category: "proxy"},
		"proxy.default_protocol": {Value: cfg.DefaultProtocol, Type: "string", Category: "proxy"},
	}
	if err := dc.configRepo.SetBatch(configs); err != nil {
		return err
	}
	return dc.Reload()
}

// UpdateHealthCheckConfig 更新健康检查配置
func (dc *DynamicConfig) UpdateHealthCheckConfig(cfg *HealthCheckConfig) error {
	merged := dc.clone()
	merged.HealthCheck = *cfg
	if err := validate(merged); err != nil {
		return fmt.Errorf("invalid health check config: %w", err)
	}

	configs := map[string]database.ConfigValue{
		"health_check.enabled":   {Value: fmt.Sprintf("%t", cfg.Enabled), Type: "bool", Category: "health_check"},
		"health_check.interval":  {Value: fmt.Sprintf("%d", int64(cfg.Interval/time.Second)), Type: "duration", Category: "health_check"},
		"health_check.timeout":   {Value: fmt.Sprintf("%d", int64(cfg.Timeout/time.Second)), Type: "duration", Category: "health_check"},
		"health_check.url":       {Value: cfg.URL, Type: "string", Category: "health_check"},
		"health_check.max_delay": {Value: fmt.Sprintf("%d", cfg.MaxDelay), Type: "int", Category: "health_check"},
	}
	if err := dc.configRepo.SetBatch(configs); err != nil {
		return err
	}
	return dc.Reload()
}

// UpdateSubscriptionConfig 更新订阅配置
func (dc *DynamicConfig) UpdateSubscriptionConfig(cfg *SubscriptionConfig) error {
	merged := dc.clone()
	merged.Subscription = *cfg
	if err := validate(merged); err != nil {
		return fmt.Errorf("invalid subscription config: %w", err)
	}

	configs := map[string]database.ConfigValue{
		"subscription.update_interval": {Value: fmt.Sprintf("%d", int64(cfg.UpdateInterval/time.Second)), Type: "duration", Category: "subscription"},
		"subscription.timeout":         {Value: fmt.Sprintf("%d", int64(cfg.Timeout/time.Second)), Type: "duration", Category: "subscription"},
	}
	if err := dc.configRepo.SetBatch(configs); err != nil {
		return err
	}
	return dc.Reload()
}

// SetOnProxyConfigChange 注册回调
func (dc *DynamicConfig) SetOnProxyConfigChange(cb func(*ProxyConfig)) {
	dc.onProxyConfigChange = cb
}

// SetOnHealthCheckConfigChange 注册回调
func (dc *DynamicConfig) SetOnHealthCheckConfigChange(cb func(*HealthCheckConfig)) {
	dc.onHealthCheckConfigChange = cb
}

// SetOnSubscriptionConfigChange 注册回调
func (dc *DynamicConfig) SetOnSubscriptionConfigChange(cb func(*SubscriptionConfig)) {
	dc.onSubscriptionConfigChange = cb
}

func (dc *DynamicConfig) clone() *Config {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	copyCfg := *dc.current
	copyCfg.Proxy = dc.current.Proxy
	copyCfg.Subscription = dc.current.Subscription
	copyCfg.HealthCheck = dc.current.HealthCheck
	copyCfg.Server = dc.current.Server
	copyCfg.Log = dc.current.Log
	return &copyCfg
}

func (dc *DynamicConfig) loadConfigFromDB() (*Config, error) {
	items, err := dc.configRepo.List()
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	for _, item := range items {
		if err := applyConfigValue(cfg, item); err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

func applyConfigValue(cfg *Config, item database.SystemConfig) error {
	switch item.Key {
	case "server.host":
		cfg.Server.Host = item.Value
	case "server.port":
		v, err := strconv.Atoi(item.Value)
		if err != nil {
			return err
		}
		cfg.Server.Port = v
	case "server.mode":
		cfg.Server.Mode = item.Value
	case "log.level":
		cfg.Log.Level = item.Value
	case "log.file":
		cfg.Log.File = item.Value
	case "proxy.port_range.min":
		v, err := strconv.Atoi(item.Value)
		if err != nil {
			return err
		}
		cfg.Proxy.PortRange.Min = v
	case "proxy.port_range.max":
		v, err := strconv.Atoi(item.Value)
		if err != nil {
			return err
		}
		cfg.Proxy.PortRange.Max = v
	case "proxy.default_ttl":
		d, err := parseDurationSeconds(item.Value)
		if err != nil {
			return err
		}
		cfg.Proxy.DefaultTTL = d
	case "proxy.max_instances":
		v, err := strconv.Atoi(item.Value)
		if err != nil {
			return err
		}
		cfg.Proxy.MaxInstances = v
	case "proxy.default_protocol":
		cfg.Proxy.DefaultProtocol = item.Value
	case "subscription.update_interval":
		d, err := parseDurationSeconds(item.Value)
		if err != nil {
			return err
		}
		cfg.Subscription.UpdateInterval = d
	case "subscription.timeout":
		d, err := parseDurationSeconds(item.Value)
		if err != nil {
			return err
		}
		cfg.Subscription.Timeout = d
	case "health_check.enabled":
		cfg.HealthCheck.Enabled = item.Value == "true" || item.Value == "1"
	case "health_check.interval":
		d, err := parseDurationSeconds(item.Value)
		if err != nil {
			return err
		}
		cfg.HealthCheck.Interval = d
	case "health_check.timeout":
		d, err := parseDurationSeconds(item.Value)
		if err != nil {
			return err
		}
		cfg.HealthCheck.Timeout = d
	case "health_check.url":
		cfg.HealthCheck.URL = item.Value
	case "health_check.max_delay":
		v, err := strconv.Atoi(item.Value)
		if err != nil {
			return err
		}
		cfg.HealthCheck.MaxDelay = v
	default:
		// 忽略未知 key
	}
	return nil
}

func parseDurationSeconds(value string) (time.Duration, error) {
	secs, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return time.Duration(secs) * time.Second, nil
}
