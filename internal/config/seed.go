package config

import (
	"fmt"
	"time"

	"proxyPool/internal/database"
)

// BuildSeedData 将配置转换为数据库初始化所需数据
func BuildSeedData(cfg *Config) *database.SeedData {
	return &database.SeedData{
		Sources: convertSources(cfg.Subscription.Sources),
		Configs: buildConfigValues(cfg),
	}
}

func convertSources(srcs []SubscriptionSource) []database.SubscriptionSource {
	result := make([]database.SubscriptionSource, 0, len(srcs))
	for _, s := range srcs {
		result = append(result, database.SubscriptionSource{
			Name:    s.Name,
			URL:     s.URL,
			Enabled: s.Enabled,
		})
	}
	return result
}

func buildConfigValues(cfg *Config) map[string]database.ConfigValue {
	return map[string]database.ConfigValue{
		"server.host": {
			Value:       cfg.Server.Host,
			Type:        "string",
			Category:    "server",
			Description: "HTTP 监听地址",
		},
		"server.port": {
			Value:       fmt.Sprintf("%d", cfg.Server.Port),
			Type:        "int",
			Category:    "server",
			Description: "HTTP 监听端口",
		},
		"server.mode": {
			Value:       cfg.Server.Mode,
			Type:        "string",
			Category:    "server",
			Description: "运行模式",
		},
		"log.level": {
			Value:       cfg.Log.Level,
			Type:        "string",
			Category:    "log",
			Description: "日志级别",
		},
		"log.file": {
			Value:       cfg.Log.File,
			Type:        "string",
			Category:    "log",
			Description: "日志文件",
		},
		"proxy.port_range.min": {
			Value:       fmt.Sprintf("%d", cfg.Proxy.PortRange.Min),
			Type:        "int",
			Category:    "proxy",
			Description: "端口池起始",
		},
		"proxy.port_range.max": {
			Value:       fmt.Sprintf("%d", cfg.Proxy.PortRange.Max),
			Type:        "int",
			Category:    "proxy",
			Description: "端口池结束",
		},
		"proxy.default_ttl": {
			Value:       fmt.Sprintf("%d", int64(cfg.Proxy.DefaultTTL/time.Second)),
			Type:        "duration",
			Category:    "proxy",
			Description: "默认存活秒数",
		},
		"proxy.max_instances": {
			Value:       fmt.Sprintf("%d", cfg.Proxy.MaxInstances),
			Type:        "int",
			Category:    "proxy",
			Description: "最大实例数",
		},
		"proxy.default_protocol": {
			Value:       cfg.Proxy.DefaultProtocol,
			Type:        "string",
			Category:    "proxy",
			Description: "默认协议",
		},
		"subscription.update_interval": {
			Value:       fmt.Sprintf("%d", int64(cfg.Subscription.UpdateInterval/time.Second)),
			Type:        "duration",
			Category:    "subscription",
			Description: "订阅更新间隔秒",
		},
		"subscription.timeout": {
			Value:       fmt.Sprintf("%d", int64(cfg.Subscription.Timeout/time.Second)),
			Type:        "duration",
			Category:    "subscription",
			Description: "订阅抓取超时秒",
		},
		"health_check.enabled": {
			Value:       fmt.Sprintf("%t", cfg.HealthCheck.Enabled),
			Type:        "bool",
			Category:    "health_check",
			Description: "健康检查开关",
		},
		"health_check.interval": {
			Value:       fmt.Sprintf("%d", int64(cfg.HealthCheck.Interval/time.Second)),
			Type:        "duration",
			Category:    "health_check",
			Description: "健康检查间隔秒",
		},
		"health_check.timeout": {
			Value:       fmt.Sprintf("%d", int64(cfg.HealthCheck.Timeout/time.Second)),
			Type:        "duration",
			Category:    "health_check",
			Description: "健康检查超时秒",
		},
		"health_check.url": {
			Value:       cfg.HealthCheck.URL,
			Type:        "string",
			Category:    "health_check",
			Description: "健康检查 URL",
		},
		"health_check.max_delay": {
			Value:       fmt.Sprintf("%d", cfg.HealthCheck.MaxDelay),
			Type:        "int",
			Category:    "health_check",
			Description: "最大延迟阈值",
		},
	}
}
