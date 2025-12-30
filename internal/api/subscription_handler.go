package api

import (
	"net/http"
	"strconv"
	"time"

	"proxyPool/internal/database"

	"github.com/gin-gonic/gin"
)

// SubscriptionResponse 订阅源响应结构
type SubscriptionResponse struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	URL             string     `json:"url"`
	Enabled         bool       `json:"enabled"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	LastFetchAt     *time.Time `json:"last_fetch_at,omitempty"`
	LastFetchStatus string     `json:"last_fetch_status,omitempty"`
	LastFetchError  string     `json:"last_fetch_error,omitempty"`
	NodeCount       int        `json:"node_count"`
}

// CreateSubscriptionRequest 创建请求
type CreateSubscriptionRequest struct {
	Name    string `json:"name" binding:"required"`
	URL     string `json:"url" binding:"required"`
	Enabled bool   `json:"enabled"`
}

// UpdateSubscriptionRequest 更新请求
type UpdateSubscriptionRequest struct {
	Name    string `json:"name" binding:"required"`
	URL     string `json:"url" binding:"required"`
	Enabled bool   `json:"enabled"`
}

// ListSubscriptions 列出订阅源
func (h *Handler) ListSubscriptions(c *gin.Context) {
	items, err := h.subRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	resp := make([]SubscriptionResponse, 0, len(items))
	for _, it := range items {
		resp = append(resp, toSubscriptionResponse(it))
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": resp})
}

// CreateSubscription 新增订阅源
func (h *Handler) CreateSubscription(c *gin.Context) {
	var req CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	id, err := h.subRepo.Create(&database.SubscriptionSource{
		Name:    req.Name,
		URL:     req.URL,
		Enabled: req.Enabled,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	_ = h.subManager.ReloadSources()

	item, _ := h.subRepo.GetByID(id)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": toSubscriptionResponse(item)})
}

// GetSubscription 获取单个订阅源
func (h *Handler) GetSubscription(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid id"})
		return
	}

	item, err := h.subRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": toSubscriptionResponse(item)})
}

// UpdateSubscription 更新订阅源
func (h *Handler) UpdateSubscription(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid id"})
		return
	}

	var req UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	err = h.subRepo.Update(id, &database.SubscriptionSource{
		Name:    req.Name,
		URL:     req.URL,
		Enabled: req.Enabled,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	_ = h.subManager.ReloadSources()

	item, _ := h.subRepo.GetByID(id)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": toSubscriptionResponse(item)})
}

// DeleteSubscription 删除订阅源
func (h *Handler) DeleteSubscription(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid id"})
		return
	}

	if err := h.subRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	_ = h.subManager.ReloadSources()
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// TestSubscription 测试订阅源可用性
func (h *Handler) TestSubscription(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid id"})
		return
	}

	item, err := h.subRepo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": err.Error()})
		return
	}

	nodes, err := h.subManager.TestSubscription(item)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"nodes": len(nodes),
		},
	})
}

func toSubscriptionResponse(s *database.SubscriptionSource) SubscriptionResponse {
	return SubscriptionResponse{
		ID:              s.ID,
		Name:            s.Name,
		URL:             s.URL,
		Enabled:         s.Enabled,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
		LastFetchAt:     s.LastFetchAt,
		LastFetchStatus: s.LastFetchStatus,
		LastFetchError:  s.LastFetchError,
		NodeCount:       s.NodeCount,
	}
}

func parseID(raw string) (int64, error) {
	return strconv.ParseInt(raw, 10, 64)
}
