package proxy

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// TTLCleaner TTL清理器
// 定期扫描过期实例并自动清理
type TTLCleaner struct {
	manager       *Manager
	checkInterval time.Duration
	logger        *zap.Logger
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewTTLCleaner 创建TTL清理器
func NewTTLCleaner(manager *Manager, checkInterval time.Duration, logger *zap.Logger) *TTLCleaner {
	ctx, cancel := context.WithCancel(context.Background())
	return &TTLCleaner{
		manager:       manager,
		checkInterval: checkInterval,
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start 启动清理器
func (c *TTLCleaner) Start() {
	c.logger.Info("TTL cleaner started", zap.Duration("interval", c.checkInterval))

	go func() {
		ticker := time.NewTicker(c.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.cleanExpiredInstances()
			case <-c.ctx.Done():
				c.logger.Info("TTL cleaner stopped")
				return
			}
		}
	}()
}

// Stop 停止清理器
func (c *TTLCleaner) Stop() {
	c.cancel()
}

// cleanExpiredInstances 清理过期实例
func (c *TTLCleaner) cleanExpiredInstances() {
	instances := c.manager.ListInstances()
	now := time.Now()
	expiredCount := 0

	for _, instance := range instances {
		if now.After(instance.ExpiresAt) {
			err := c.manager.ReleaseProxy(instance.ID)
			if err != nil {
				c.logger.Error("failed to release expired instance",
					zap.String("id", instance.ID),
					zap.Error(err))
			} else {
				expiredCount++
			}
		}
	}

	if expiredCount > 0 {
		c.logger.Info("cleaned expired instances",
			zap.Int("count", expiredCount),
			zap.Int("remaining", len(instances)-expiredCount))
	}
}
