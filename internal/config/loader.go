package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var globalConfig *Config

// Load 从文件加载配置
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 验证配置
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	globalConfig = &cfg
	return &cfg, nil
}

// Get 获取全局配置
func Get() *Config {
	return globalConfig
}

// validate 验证配置有效性
func validate(cfg *Config) error {
	// 验证服务器配置
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}

	// 验证端口范围
	if cfg.Proxy.PortRange.Min <= 0 || cfg.Proxy.PortRange.Max > 65535 {
		return fmt.Errorf("invalid proxy port range: %d-%d",
			cfg.Proxy.PortRange.Min, cfg.Proxy.PortRange.Max)
	}
	if cfg.Proxy.PortRange.Min >= cfg.Proxy.PortRange.Max {
		return fmt.Errorf("invalid proxy port range: min >= max")
	}

	// 验证最大实例数
	if cfg.Proxy.MaxInstances <= 0 {
		return fmt.Errorf("max_instances must be positive")
	}

	// 验证协议类型
	switch cfg.Proxy.DefaultProtocol {
	case "socks5", "http", "mixed":
		// 合法
	default:
		return fmt.Errorf("invalid default protocol: %s", cfg.Proxy.DefaultProtocol)
	}

	// 验证TTL
	if cfg.Proxy.DefaultTTL <= 0 {
		return fmt.Errorf("default_ttl must be positive")
	}

	return nil
}
