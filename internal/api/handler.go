package api

import (
	"proxyPool/internal/config"
	"proxyPool/internal/database"
	"proxyPool/internal/proxy"
	"proxyPool/internal/subscription"

	"github.com/gin-gonic/gin"
)

// Handler API处理器
type Handler struct {
	subManager   *subscription.Manager
	proxyManager *proxy.Manager
	dynamicCfg   *config.DynamicConfig
	subRepo      *database.SubscriptionRepository
}

// NewHandler 创建API处理器
func NewHandler(subManager *subscription.Manager, proxyManager *proxy.Manager, dynamicCfg *config.DynamicConfig, subRepo *database.SubscriptionRepository) *Handler {
	return &Handler{
		subManager:   subManager,
		proxyManager: proxyManager,
		dynamicCfg:   dynamicCfg,
		subRepo:      subRepo,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.POST("/getProxy", h.GetProxy)
		api.POST("/releaseProxy", h.ReleaseProxy)
		api.GET("/getInstance/:id", h.GetInstance)
		api.GET("/listInstances", h.ListInstances)
		api.GET("/stats", h.GetStats)
		api.GET("/nodePool", h.GetNodePool)
		api.GET("/nodes", h.GetNodes)
		api.POST("/subscription/update", h.TriggerSubscriptionUpdate)

		// 订阅源管理
		api.GET("/subscriptions", h.ListSubscriptions)
		api.POST("/subscriptions", h.CreateSubscription)
		api.GET("/subscriptions/:id", h.GetSubscription)
		api.PUT("/subscriptions/:id", h.UpdateSubscription)
		api.DELETE("/subscriptions/:id", h.DeleteSubscription)
		api.POST("/subscriptions/:id/test", h.TestSubscription)

		// 动态配置
		api.GET("/config", h.GetAllConfig)
		api.GET("/config/:category", h.GetConfigByCategory)
		api.PUT("/config/proxy", h.UpdateProxyConfig)
		api.PUT("/config/health-check", h.UpdateHealthCheckConfig)
		api.PUT("/config/subscription", h.UpdateSubscriptionConfig)
		api.POST("/config/reload", h.ReloadConfig)
	}
}

// GetNodePool 获取节点池状态
func (h *Handler) GetNodePool(c *gin.Context) {
	pool := h.subManager.GetNodePool()

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"total_nodes":     pool.Count(),
			"available_nodes": pool.CountAvailable(),
			"nodes_by_type":   pool.CountByType(),
			"avg_delay":       pool.AvgDelay(),
		},
	})
}

// TriggerSubscriptionUpdate 手动更新订阅
func (h *Handler) TriggerSubscriptionUpdate(c *gin.Context) {
	go h.subManager.UpdateAll()

	c.JSON(200, gin.H{
		"code":    0,
		"message": "Update task started",
	})
}

// GetProxy 获取代理
func (h *Handler) GetProxy(c *gin.Context) {
	var req proxy.CreateProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	instance, err := h.proxyManager.CreateProxy(&req)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    instance.ToResponse(),
	})
}

// ReleaseProxy 释放代理
func (h *Handler) ReleaseProxy(c *gin.Context) {
	var req struct {
		InstanceID string `json:"instance_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "Invalid request: " + err.Error(),
		})
		return
	}

	err := h.proxyManager.ReleaseProxy(req.InstanceID)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
	})
}

// GetInstance 获取实例信息
func (h *Handler) GetInstance(c *gin.Context) {
	instanceID := c.Param("id")

	instance, err := h.proxyManager.GetInstance(instanceID)
	if err != nil {
		c.JSON(404, gin.H{
			"code":    404,
			"message": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    instance.ToResponse(),
	})
}

// ListInstances 列出所有实例
func (h *Handler) ListInstances(c *gin.Context) {
	instances := h.proxyManager.ListInstances()

	data := make([]map[string]interface{}, 0, len(instances))
	for _, instance := range instances {
		data = append(data, instance.ToResponse())
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

// GetStats 获取统计信息
func (h *Handler) GetStats(c *gin.Context) {
	stats := h.proxyManager.GetStats()

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}

// GetNodes 获取节点列表
func (h *Handler) GetNodes(c *gin.Context) {
	pool := h.subManager.GetNodePool()
	nodes := pool.GetAll()

	data := make([]map[string]interface{}, 0, len(nodes))
	for _, node := range nodes {
		data = append(data, map[string]interface{}{
			"id":         node.ID,
			"name":       node.Name,
			"type":       node.Type,
			"delay":      node.Delay,
			"available":  node.Available,
			"last_check": node.LastCheck.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}
