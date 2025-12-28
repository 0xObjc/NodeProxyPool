package mihomo

import (
	"fmt"

	"github.com/metacubex/mihomo/config"
	C "github.com/metacubex/mihomo/constant"
)

// ConfigBuilder Mihomo配置构建器
type ConfigBuilder struct{}

// NewConfigBuilder 创建配置构建器
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{}
}

// BuildMinimal 构建最小化配置(包含单个代理节点)
func (b *ConfigBuilder) BuildMinimal(proxy C.Proxy) (*config.Config, error) {
	// 构建最小YAML配置
	yamlConfig := fmt.Sprintf(`
mode: direct
ipv6: false
log-level: error

dns:
  enable: false

proxies:
  - name: %s
    type: direct

`, proxy.Name())

	// 解析配置
	cfg, err := config.Parse([]byte(yamlConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// 直接使用传入的proxy对象
	cfg.Proxies = map[string]C.Proxy{
		proxy.Name(): proxy,
	}

	return cfg, nil
}
