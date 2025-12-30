package subscription

import (
	"fmt"
	"sync"
	"time"

	"proxyPool/internal/config"
	"proxyPool/internal/database"
	"proxyPool/internal/node"
	"proxyPool/pkg/logger"

	"go.uber.org/zap"
)

// Manager 订阅管理器
type Manager struct {
	fetcher  *Fetcher
	parser   *Parser
	nodePool *node.Pool
	db       *database.DB
	subRepo  *database.SubscriptionRepository
	nodeRepo *database.NodeRepository
	interval time.Duration

	stopCh chan struct{}
	mu     sync.RWMutex
}

// NewManager 创建订阅管理器
func NewManager(cfg *config.DynamicConfig, db *database.DB, nodePool *node.Pool, subRepo *database.SubscriptionRepository, nodeRepo *database.NodeRepository) *Manager {
	subCfg := cfg.Current().Subscription
	return &Manager{
		fetcher:  NewFetcher(subCfg.Timeout),
		parser:   NewParser(),
		nodePool: nodePool,
		db:       db,
		subRepo:  subRepo,
		nodeRepo: nodeRepo,
		interval: subCfg.UpdateInterval,
		stopCh:   make(chan struct{}),
	}
}

// Start 启动自动更新
func (m *Manager) Start() {
	logger.Info("Starting subscription manager",
		zap.Duration("interval", m.interval),
	)

	m.UpdateAll()

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

	sources, err := m.loadSources()
	if err != nil {
		logger.Error("Failed to load subscription sources", zap.Error(err))
		return
	}

	var wg sync.WaitGroup
	for _, source := range sources {
		if !source.Enabled {
			continue
		}

		wg.Add(1)
		go func(src *database.SubscriptionSource) {
			defer wg.Done()

			nodes, err := m.UpdateOne(src)
			if err != nil {
				logger.Error("Failed to update subscription",
					zap.String("name", src.Name),
					zap.Error(err),
				)
				_ = m.subRepo.UpdateFetchStatus(src.ID, "failed", err.Error(), 0)
				return
			}

			if err := m.nodeRepo.UpsertBatch(nodes, &src.ID); err != nil {
				logger.Warn("Persist nodes failed", zap.Error(err))
			}

			mu.Lock()
			allNodes = append(allNodes, nodes...)
			mu.Unlock()

			_ = m.subRepo.UpdateFetchStatus(src.ID, "success", "", len(nodes))
			logger.Info("Subscription updated",
				zap.String("name", src.Name),
				zap.Int("nodes", len(nodes)),
			)
		}(source)
	}

	wg.Wait()

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
func (m *Manager) UpdateOne(source *database.SubscriptionSource) ([]*node.Node, error) {
	content, err := m.fetcher.Fetch(source.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}

	nodes, err := m.parser.Parse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}

	return nodes, nil
}

// ReloadSources 从数据库刷新订阅源列表
func (m *Manager) ReloadSources() error {
	_, err := m.loadSources()
	return err
}

// TestSubscription 只抓取不落库
func (m *Manager) TestSubscription(source *database.SubscriptionSource) ([]*node.Node, error) {
	return m.UpdateOne(source)
}

// UpdateConfig 更新订阅相关配置
func (m *Manager) UpdateConfig(cfg *config.SubscriptionConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.interval = cfg.UpdateInterval
	m.fetcher = NewFetcher(cfg.Timeout)
}

// GetNodePool 获取节点池
func (m *Manager) GetNodePool() *node.Pool {
	return m.nodePool
}

func (m *Manager) loadSources() ([]*database.SubscriptionSource, error) {
	sources, err := m.subRepo.GetEnabled()
	if err != nil {
		return nil, err
	}
	if len(sources) == 0 {
		logger.Warn("No enabled subscription sources in database")
	}
	return sources, nil
}
