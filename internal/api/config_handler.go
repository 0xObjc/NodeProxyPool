package api

import (
	"net/http"
	"time"

	"proxyPool/internal/config"

	"github.com/gin-gonic/gin"
)

// GetAllConfig 返回所有配置
func (h *Handler) GetAllConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    h.dynamicCfg.Current(),
	})
}

// GetConfigByCategory 按分类返回配置
func (h *Handler) GetConfigByCategory(c *gin.Context) {
	category := c.Param("category")
	cfg := h.dynamicCfg.Current()
	switch category {
	case "proxy":
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": cfg.Proxy})
	case "health_check":
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": cfg.HealthCheck})
	case "subscription":
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": cfg.Subscription})
	case "server":
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": cfg.Server})
	case "log":
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": cfg.Log})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "unknown category"})
	}
}

// UpdateProxyConfigRequest 更新代理配置请求
type UpdateProxyConfigRequest struct {
	PortRange struct {
		Min int `json:"min" binding:"required"`
		Max int `json:"max" binding:"required"`
	} `json:"port_range" binding:"required"`
	DefaultTTL      int    `json:"default_ttl" binding:"required"` // 秒
	MaxInstances    int    `json:"max_instances" binding:"required"`
	DefaultProtocol string `json:"default_protocol" binding:"required"`
}

// UpdateProxyConfig 更新代理配置
func (h *Handler) UpdateProxyConfig(c *gin.Context) {
	var req UpdateProxyConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	cfg := &config.ProxyConfig{
		PortRange: config.PortRangeConfig{
			Min: req.PortRange.Min,
			Max: req.PortRange.Max,
		},
		DefaultTTL:      time.Duration(req.DefaultTTL) * time.Second,
		MaxInstances:    req.MaxInstances,
		DefaultProtocol: req.DefaultProtocol,
	}

	if err := h.dynamicCfg.UpdateProxyConfig(cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// UpdateHealthCheckConfigRequest 更新健康检查配置请求
type UpdateHealthCheckConfigRequest struct {
	Enabled  bool   `json:"enabled"`
	Interval int    `json:"interval"` // 秒
	Timeout  int    `json:"timeout"`  // 秒
	URL      string `json:"url"`
	MaxDelay int    `json:"max_delay"`
}

// UpdateHealthCheckConfig 更新健康检查配置
func (h *Handler) UpdateHealthCheckConfig(c *gin.Context) {
	var req UpdateHealthCheckConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	cfg := &config.HealthCheckConfig{
		Enabled:  req.Enabled,
		Interval: time.Duration(req.Interval) * time.Second,
		Timeout:  time.Duration(req.Timeout) * time.Second,
		URL:      req.URL,
		MaxDelay: req.MaxDelay,
	}

	if err := h.dynamicCfg.UpdateHealthCheckConfig(cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// UpdateSubscriptionConfigRequest 更新订阅配置请求
type UpdateSubscriptionConfigRequest struct {
	UpdateInterval int `json:"update_interval"` // 秒
	Timeout        int `json:"timeout"`         // 秒
}

// UpdateSubscriptionConfig 更新订阅配置
func (h *Handler) UpdateSubscriptionConfig(c *gin.Context) {
	var req UpdateSubscriptionConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	cfg := &config.SubscriptionConfig{
		UpdateInterval: time.Duration(req.UpdateInterval) * time.Second,
		Timeout:        time.Duration(req.Timeout) * time.Second,
	}

	if err := h.dynamicCfg.UpdateSubscriptionConfig(cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	h.subManager.UpdateConfig(cfg)

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// ReloadConfig 手动重载配置
func (h *Handler) ReloadConfig(c *gin.Context) {
	if err := h.dynamicCfg.Reload(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	h.subManager.UpdateConfig(&h.dynamicCfg.Current().Subscription)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}
