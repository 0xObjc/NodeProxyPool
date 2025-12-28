package config

import "time"

// Config 应用配置
type Config struct {
	Server       ServerConfig       `yaml:"server"`
	Proxy        ProxyConfig        `yaml:"proxy"`
	Subscription SubscriptionConfig `yaml:"subscription"`
	HealthCheck  HealthCheckConfig  `yaml:"health_check"`
	Log          LogConfig          `yaml:"log"`
}

// ServerConfig HTTP服务器配置
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"` // debug/release
}

// ProxyConfig 代理配置
type ProxyConfig struct {
	PortRange       PortRangeConfig `yaml:"port_range"`
	DefaultTTL      time.Duration   `yaml:"default_ttl"`
	MaxInstances    int             `yaml:"max_instances"`
	DefaultProtocol string          `yaml:"default_protocol"` // socks5/http/mixed
}

// PortRangeConfig 端口范围配置
type PortRangeConfig struct {
	Min int `yaml:"min"`
	Max int `yaml:"max"`
}

// SubscriptionConfig 订阅配置
type SubscriptionConfig struct {
	Sources        []SubscriptionSource `yaml:"sources"`
	UpdateInterval time.Duration        `yaml:"update_interval"`
	Timeout        time.Duration        `yaml:"timeout"`
}

// SubscriptionSource 订阅源
type SubscriptionSource struct {
	Name    string `yaml:"name"`
	URL     string `yaml:"url"`
	Enabled bool   `yaml:"enabled"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
	URL      string        `yaml:"url"`
	MaxDelay int           `yaml:"max_delay"` // 最大延迟(ms)
}

// LogConfig 日志配置
type LogConfig struct {
	Level string `yaml:"level"` // debug/info/warn/error
	File  string `yaml:"file"`
}
