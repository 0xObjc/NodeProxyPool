package subscription

import (
	"fmt"
	"sync"
	"time"

	"proxyPool/internal/config"
	"proxyPool/internal/node"
	"proxyPool/pkg/logger"

	"go.uber.org/zap"
)

// Manager 订阅管理器
type Manager struct {
	fetcher  *Fetcher
	parser   *Parser
	nodePool *node.Pool
	sources  []config.SubscriptionSource
	interval time.Duration

	stopCh chan struct{}
	mu     sync.RWMutex
}

// NewManager 创建订阅管理器
func NewManager(cfg *config.Config, nodePool *node.Pool) *Manager {
	return &Manager{
		fetcher:  NewFetcher(cfg.Subscription.Timeout),
		parser:   NewParser(),
		nodePool: nodePool,
		sources:  cfg.Subscription.Sources,
		interval: cfg.Subscription.UpdateInterval,
		stopCh:   make(chan struct{}),
	}
}

// Start 启动自动更新
func (m *Manager) Start() {
	logger.Info("Starting subscription manager",
		zap.Int("sources", len(m.sources)),
		zap.Duration("interval", m.interval),
	)

	// 立即执行一次更新
	m.UpdateAll()

	// 启动定时更新
	go func() {
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.UpdateAll()
			case <-m.stopCh:
				return
			}
		}
	}()
}

// Stop 停止管理器
func (m *Manager) Stop() {
	close(m.stopCh)
}

// UpdateAll 更新所有启用的订阅源
func (m *Manager) UpdateAll() {
	logger.Info("Updating all subscriptions...")

	var allNodes []*node.Node
	var mu sync.Mutex

	var wg sync.WaitGroup
	for _, source := range m.sources {
		if !source.Enabled {
			logger.Debug("Skipping disabled subscription",
				zap.String("name", source.Name),
			)
			continue
		}

		wg.Add(1)
		go func(src config.SubscriptionSource) {
			defer wg.Done()

			nodes, err := m.UpdateOne(src)
			if err != nil {
				logger.Error("Failed to update subscription",
					zap.String("name", src.Name),
					zap.Error(err),
				)
				return
			}

			mu.Lock()
			allNodes = append(allNodes, nodes...)
			mu.Unlock()

			logger.Info("Subscription updated",
				zap.String("name", src.Name),
				zap.Int("nodes", len(nodes)),
			)
		}(source)
	}

	wg.Wait()

	// 更新节点池
	if len(allNodes) > 0 {
		m.nodePool.UpdateNodes(allNodes)
		logger.Info("Node pool updated",
			zap.Int("total_nodes", len(allNodes)),
		)
	} else {
		logger.Warn("No nodes updated, node pool remains unchanged")
	}
}

// UpdateOne 更新单个订阅源
func (m *Manager) UpdateOne(source config.SubscriptionSource) ([]*node.Node, error) {
	// 抓取订阅内容
	content, err := m.fetcher.Fetch(source.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}

	// 解析节点
	nodes, err := m.parser.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}

	return nodes, nil
}

// GetNodePool 获取节点池
func (m *Manager) GetNodePool() *node.Pool {
	return m.nodePool
}
