package mihomo

import (
	"fmt"
	"sync"

	"proxyPool/pkg/logger"

	"github.com/metacubex/mihomo/adapter"
	"github.com/metacubex/mihomo/adapter/inbound"
	"github.com/metacubex/mihomo/adapter/outbound"
	C "github.com/metacubex/mihomo/constant"
	"github.com/metacubex/mihomo/hub/executor"
	"github.com/metacubex/mihomo/listener/http"
	"github.com/metacubex/mihomo/listener/socks"
	"github.com/metacubex/mihomo/tunnel"
	"go.uber.org/zap"
)

// ProxyListener 代理监听器
type ProxyListener struct {
	socksListener *socks.Listener
	httpListener  *http.Listener
	protocol      string
	port          int
}

// MihomoAdapter Mihomo适配器
type MihomoAdapter struct {
	configBuilder *ConfigBuilder
	tunnel        C.Tunnel
	listeners     map[int]*ProxyListener // port -> listener
	mu            sync.RWMutex
	proxyRegistry map[string]C.Proxy
	initialized   bool
}

// Adapter 是 MihomoAdapter 的别名，用于简化外部引用
type Adapter = MihomoAdapter

// NewAdapter 创建Mihomo适配器（简化版构造函数）
func NewAdapter(logger *zap.Logger) (*Adapter, error) {
	adapter := &MihomoAdapter{
		configBuilder: NewConfigBuilder(),
		listeners:     make(map[int]*ProxyListener),
		proxyRegistry: make(map[string]C.Proxy),
	}

	// 初始化
	if err := adapter.Init(); err != nil {
		return nil, err
	}

	return adapter, nil
}

// NewMihomoAdapter 创建Mihomo适配器
func NewMihomoAdapter() *MihomoAdapter {
	return &MihomoAdapter{
		configBuilder: NewConfigBuilder(),
		listeners:     make(map[int]*ProxyListener),
		proxyRegistry: make(map[string]C.Proxy),
	}
}

// Init 初始化Mihomo
func (m *MihomoAdapter) Init() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return nil
	}

	// 构建最小化配置并初始化Mihomo
	minimalConfig := `
mode: direct
ipv6: false
log-level: error
allow-lan: true
bind-address: "*"

dns:
  enable: false
`

	cfg, err := executor.ParseWithBytes([]byte(minimalConfig))
	if err != nil {
		return fmt.Errorf("failed to parse minimal config: %w", err)
	}

	// 注册默认DIRECT代理，供后续路由回落使用
	directOutbound := outbound.NewDirect()
	directProxy := adapter.NewProxy(directOutbound)
	if cfg.Proxies == nil {
		cfg.Proxies = make(map[string]C.Proxy)
	}
	cfg.Proxies["DIRECT"] = directProxy
	m.proxyRegistry["DIRECT"] = directProxy

	// 应用配置
	executor.ApplyConfig(cfg, true)

	// 使用全局Tunnel实例
	m.tunnel = tunnel.Tunnel

	// 将已知代理写入全局代理表
	tunnel.UpdateProxies(m.proxyRegistry, nil)

	m.initialized = true
	logger.Info("Mihomo adapter initialized")

	return nil
}

// ensureProxy 将代理注册到全局路由表
func (m *MihomoAdapter) ensureProxy(proxy C.Proxy) {
	if proxy == nil {
		return
	}
	if _, exists := m.proxyRegistry[proxy.Name()]; exists {
		return
	}

	m.proxyRegistry[proxy.Name()] = proxy
	tunnel.UpdateProxies(m.proxyRegistry, nil)
}

// CreateListener 创建代理监听器
func (m *MihomoAdapter) CreateListener(protocol string, port int, proxy C.Proxy) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return fmt.Errorf("mihomo adapter not initialized")
	}

	// 注册本次使用的代理以便路由绑定
	m.ensureProxy(proxy)

	// 检查端口是否已被使用
	if _, exists := m.listeners[port]; exists {
		return fmt.Errorf("port %d already in use", port)
	}

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	proxyListener := &ProxyListener{
		protocol: protocol,
		port:     port,
	}

	var err error

	// 根据协议类型创建监听器
	switch protocol {
	case "socks5":
		proxyListener.socksListener, err = socks.New(addr, m.tunnel,
			inbound.WithSpecialProxy(proxy.Name()))
		if err != nil {
			return fmt.Errorf("failed to create socks listener: %w", err)
		}
		logger.Info("Created SOCKS5 listener",
			zap.Int("port", port),
			zap.String("proxy", proxy.Name()),
		)

	case "http":
		proxyListener.httpListener, err = http.New(addr, m.tunnel,
			inbound.WithSpecialProxy(proxy.Name()))
		if err != nil {
			return fmt.Errorf("failed to create http listener: %w", err)
		}
		logger.Info("Created HTTP listener",
			zap.Int("port", port),
			zap.String("proxy", proxy.Name()),
		)

	case "mixed":
		// 创建SOCKS5监听器(mixed在Mihomo中是通过mixed包实现的)
		// 简化实现,先只支持socks5和http
		return fmt.Errorf("mixed protocol not yet supported, use socks5 or http")

	default:
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}

	// 保存监听器
	m.listeners[port] = proxyListener

	return nil
}

// CloseListener 关闭监听器
func (m *MihomoAdapter) CloseListener(protocol string, port int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	listener, exists := m.listeners[port]
	if !exists {
		return fmt.Errorf("listener on port %d not found", port)
	}

	// 关闭监听器
	if listener.socksListener != nil {
		if err := listener.socksListener.Close(); err != nil {
			logger.Error("Failed to close SOCKS listener",
				zap.Int("port", port),
				zap.Error(err),
			)
		}
	}

	if listener.httpListener != nil {
		if err := listener.httpListener.Close(); err != nil {
			logger.Error("Failed to close HTTP listener",
				zap.Int("port", port),
				zap.Error(err),
			)
		}
	}

	// 从map中删除
	delete(m.listeners, port)

	logger.Info("Closed listener",
		zap.Int("port", port),
		zap.String("protocol", listener.protocol),
	)

	return nil
}

// GetTunnel 获取Tunnel实例
func (m *MihomoAdapter) GetTunnel() C.Tunnel {
	return m.tunnel
}

// ListenerCount 获取当前监听器数量
func (m *MihomoAdapter) ListenerCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.listeners)
}
