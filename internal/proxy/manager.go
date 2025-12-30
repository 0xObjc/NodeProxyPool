package proxy

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"proxyPool/internal/mihomo"
	"proxyPool/internal/node"
)

var (
	ErrNoAvailableNode = errors.New("no available node matching criteria")
	ErrInstanceNotFound = errors.New("proxy instance not found")
	ErrMaxInstancesReached = errors.New("maximum instances limit reached")
)

// CreateProxyRequest 创建代理请求
type CreateProxyRequest struct {
	Protocol  string `json:"protocol"`   // socks5/http/mixed
	TTL       int    `json:"ttl"`        // 存活时间(秒)
	MaxDelay  int    `json:"max_delay"`  // 最大延迟(ms)
	NodeType  string `json:"node_type"`  // vmess/trojan/ss
	Region    string `json:"region"`     // 地区关键字
}

// Manager 代理管理器
// 核心编排器,协调端口分配、节点选择、监听器创建
type Manager struct {
	allocator      *PortAllocator
	nodePool       *node.Pool
	mihomoAdapter  *mihomo.Adapter
	instances      map[string]*ProxyInstance
	portMap        map[int]string
	maxInstances   int
	defaultTTL     time.Duration
	defaultProtocol string
	mu             sync.RWMutex
	logger         *zap.Logger
}

// NewManager 创建代理管理器
func NewManager(
	allocator *PortAllocator,
	nodePool *node.Pool,
	mihomoAdapter *mihomo.Adapter,
	maxInstances int,
	defaultTTL time.Duration,
	defaultProtocol string,
	logger *zap.Logger,
) *Manager {
	return &Manager{
		allocator:       allocator,
		nodePool:        nodePool,
		mihomoAdapter:   mihomoAdapter,
		instances:       make(map[string]*ProxyInstance),
		portMap:         make(map[int]string),
		maxInstances:    maxInstances,
		defaultTTL:      defaultTTL,
		defaultProtocol: defaultProtocol,
		logger:          logger,
	}
}

// CreateProxy 创建代理实例
// 核心流程: 选择节点 -> 分配端口 -> 创建监听器 -> 注册实例
func (m *Manager) CreateProxy(req *CreateProxyRequest) (*ProxyInstance, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. 检查实例数量限制
	if len(m.instances) >= m.maxInstances {
		return nil, ErrMaxInstancesReached
	}

	// 2. 应用默认值
	if req.Protocol == "" {
		req.Protocol = m.defaultProtocol
	}
	if req.TTL == 0 {
		req.TTL = int(m.defaultTTL.Seconds())
	}

	// 3. 智能选择节点
	selectedNode := m.selectNode(req)
	if selectedNode == nil {
		return nil, ErrNoAvailableNode
	}

	// 4. 分配端口
	port, err := m.allocator.Allocate()
	if err != nil {
		m.logger.Error("failed to allocate port", zap.Error(err))
		return nil, err
	}

	// 5. 创建实例对象
	instance := &ProxyInstance{
		ID:        uuid.New().String(),
		Host:      "127.0.0.1",
		Port:      port,
		Protocol:  req.Protocol,
		Node:      selectedNode,
		NodeName:  selectedNode.Name,
		NodeDelay: selectedNode.Delay,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(req.TTL) * time.Second),
		TTL:       req.TTL,
	}

	// 6. 创建Mihomo监听器
	err = m.mihomoAdapter.CreateListener(req.Protocol, port, selectedNode.Proxy)
	if err != nil {
		m.allocator.Release(port)
		m.logger.Error("failed to create mihomo listener",
			zap.Error(err),
			zap.Int("port", port),
			zap.String("node", selectedNode.Name))
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	// 7. 注册实例
	m.instances[instance.ID] = instance
	m.portMap[port] = instance.ID

	m.logger.Info("proxy instance created",
		zap.String("id", instance.ID),
		zap.Int("port", port),
		zap.String("protocol", req.Protocol),
		zap.String("node", selectedNode.Name),
		zap.Int("ttl", req.TTL))

	return instance, nil
}

// ReleaseProxy 释放代理实例
// 关闭监听器 -> 释放端口 -> 删除实例
func (m *Manager) ReleaseProxy(instanceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.instances[instanceID]
	if !exists {
		return ErrInstanceNotFound
	}

	// 1. 关闭Mihomo监听器
	err := m.mihomoAdapter.CloseListener(instance.Protocol, instance.Port)
	if err != nil {
		m.logger.Warn("failed to close mihomo listener",
			zap.Error(err),
			zap.Int("port", instance.Port))
	}

	// 2. 释放端口
	m.allocator.Release(instance.Port)

	// 3. 删除实例
	delete(m.instances, instanceID)
	delete(m.portMap, instance.Port)

	m.logger.Info("proxy instance released",
		zap.String("id", instanceID),
		zap.Int("port", instance.Port))

	return nil
}

// GetInstance 获取实例信息
func (m *Manager) GetInstance(instanceID string) (*ProxyInstance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instance, exists := m.instances[instanceID]
	if !exists {
		return nil, ErrInstanceNotFound
	}

	return instance, nil
}

// ListInstances 列出所有活跃实例
func (m *Manager) ListInstances() []*ProxyInstance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instances := make([]*ProxyInstance, 0, len(m.instances))
	for _, instance := range m.instances {
		instances = append(instances, instance)
	}

	return instances
}

// GetStats 获取统计信息
func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"total_instances":     len(m.instances),
		"max_instances":       m.maxInstances,
		"available_slots":     m.maxInstances - len(m.instances),
		"ports_used":          m.allocator.GetUsedCount(),
		"ports_available":     m.allocator.GetAvailableCount(),
	}
}

// UpdateConfig 动态更新核心配置
func (m *Manager) UpdateConfig(maxInstances int, defaultTTL time.Duration, defaultProtocol string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.maxInstances = maxInstances
	m.defaultTTL = defaultTTL
	if defaultProtocol != "" {
		m.defaultProtocol = defaultProtocol
	}
}

// selectNode 智能选择节点
// 按延迟、类型、地区等条件过滤,返回最优节点
func (m *Manager) selectNode(req *CreateProxyRequest) *node.Node {
	allNodes := m.nodePool.GetAll()
	if len(allNodes) == 0 {
		return nil
	}

	// 过滤可用节点
	var candidates []*node.Node
	for _, n := range allNodes {
		if n.Available {
			if matchesFilters(n, req) {
				candidates = append(candidates, n)
			}
		}
	}

	// 如无可用节点，降级为忽略可用性再尝试（避免健康检查全失败导致完全不可用）
	if len(candidates) == 0 {
		for _, n := range allNodes {
			if matchesFilters(n, req) {
				candidates = append(candidates, n)
			}
		}
		if len(candidates) == 0 {
			return nil
		}
	}

	// 选择延迟最低的节点
	bestNode := candidates[0]
	for _, n := range candidates[1:] {
		if n.Delay < bestNode.Delay {
			bestNode = n
		}
	}

	return bestNode
}

// contains 检查字符串是否包含子串(不区分大小写)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
		 len(s) > 0 && len(substr) > 0 &&
		 containsIgnoreCase(s, substr))
}

// matchesFilters 应用延迟/类型/地区过滤；对未检测的延迟(-1)不会过滤
func matchesFilters(n *node.Node, req *CreateProxyRequest) bool {
	// 延迟过滤
	if req.MaxDelay > 0 && n.Delay > req.MaxDelay {
		return false
	}

	// 类型过滤
	if req.NodeType != "" && n.Type != req.NodeType {
		return false
	}

	// 地区过滤 (简单字符串匹配)
	if req.Region != "" {
		if !contains(n.Name, req.Region) {
			return false
		}
	}
	return true
}

func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}
