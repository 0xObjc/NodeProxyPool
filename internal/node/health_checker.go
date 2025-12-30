package node

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// HealthChecker 健康检查器
type HealthChecker struct {
	pool          *Pool
	checkInterval time.Duration
	checkTimeout  time.Duration
	testURL       string
	maxDelay      int
	logger        *zap.Logger
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	mu            sync.Mutex
	started       bool
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(
	pool *Pool,
	checkInterval time.Duration,
	checkTimeout time.Duration,
	testURL string,
	maxDelay int,
	logger *zap.Logger,
) *HealthChecker {
	return &HealthChecker{
		pool:          pool,
		checkInterval: checkInterval,
		checkTimeout:  checkTimeout,
		testURL:       testURL,
		maxDelay:      maxDelay,
		logger:        logger,
	}
}

// Start 启动健康检查
func (h *HealthChecker) Start() {
	h.mu.Lock()
	if h.started {
		h.mu.Unlock()
		return
	}
	h.ctx, h.cancel = context.WithCancel(context.Background())
	h.started = true
	h.mu.Unlock()

	h.logger.Info("Health checker started",
		zap.Duration("interval", h.checkInterval),
		zap.String("test_url", h.testURL))

	h.wg.Add(1)
	go h.run()
}

// Stop 停止健康检查
func (h *HealthChecker) Stop() {
	h.mu.Lock()
	if !h.started {
		h.mu.Unlock()
		return
	}
	cancel := h.cancel
	h.started = false
	h.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	h.wg.Wait()
	h.logger.Info("Health checker stopped")
}

// UpdateConfig 动态更新配置并重启
func (h *HealthChecker) UpdateConfig(interval, timeout time.Duration, url string, maxDelay int) {
	h.mu.Lock()
	running := h.started
	h.checkInterval = interval
	h.checkTimeout = timeout
	h.testURL = url
	h.maxDelay = maxDelay
	h.mu.Unlock()

	if running {
		h.Stop()
	}
	h.Start()
}

// run 运行健康检查循环
func (h *HealthChecker) run() {
	defer h.wg.Done()

	ticker := time.NewTicker(h.checkInterval)
	defer ticker.Stop()

	// 启动时立即执行一次检查
	h.checkAllNodes()

	for {
		select {
		case <-ticker.C:
			h.checkAllNodes()
		case <-h.ctx.Done():
			return
		}
	}
}

// checkAllNodes 检查所有节点
func (h *HealthChecker) checkAllNodes() {
	nodes := h.pool.GetAll()
	if len(nodes) == 0 {
		return
	}

	h.logger.Debug("Starting health check", zap.Int("node_count", len(nodes)))

	start := time.Now()
	var wg sync.WaitGroup

	// 并发检查所有节点
	for _, node := range nodes {
		wg.Add(1)
		go func(n *Node) {
			defer wg.Done()
			h.checkNode(n)
		}(node)
	}

	wg.Wait()

	// 统计结果
	available := 0
	for _, node := range nodes {
		if node.Available {
			available++
		}
	}

	h.logger.Info("Health check completed",
		zap.Int("total", len(nodes)),
		zap.Int("available", available),
		zap.Int("unavailable", len(nodes)-available),
		zap.Duration("duration", time.Since(start)))
}

// checkNode 检查单个节点
func (h *HealthChecker) checkNode(n *Node) {
	ctx, cancel := context.WithTimeout(h.ctx, h.checkTimeout)
	defer cancel()

	start := time.Now()
	// URLTest返回uint16类型的延迟(毫秒)
	delay, err := n.Proxy.URLTest(ctx, h.testURL, nil)

	if err != nil {
		n.Available = false
		n.Delay = -1
		n.LastCheck = time.Now()
		h.logger.Debug("Node check failed",
			zap.String("node", n.Name),
			zap.Error(err))
		return
	}

	delayMs := int(delay)
	n.Delay = delayMs
	n.LastCheck = time.Now()

	// 根据延迟判断是否可用
	if h.maxDelay > 0 && delayMs > h.maxDelay {
		n.Available = false
		h.logger.Debug("Node delay too high",
			zap.String("node", n.Name),
			zap.Int("delay", delayMs),
			zap.Int("max_delay", h.maxDelay))
	} else {
		n.Available = true
		h.logger.Debug("Node check success",
			zap.String("node", n.Name),
			zap.Int("delay", delayMs),
			zap.Duration("duration", time.Since(start)))
	}
}
