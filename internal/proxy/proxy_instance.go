package proxy

import (
	"time"

	"proxyPool/internal/node"
)

// ProxyInstance 代理实例
// 表示一个活跃的代理连接
type ProxyInstance struct {
	ID        string       `json:"instance_id"`
	Host      string       `json:"host"`
	Port      int          `json:"port"`
	Protocol  string       `json:"protocol"`
	Node      *node.Node   `json:"-"`
	NodeName  string       `json:"node_name"`
	NodeDelay int          `json:"node_delay"`
	CreatedAt time.Time    `json:"created_at"`
	ExpiresAt time.Time    `json:"expires_at"`
	TTL       int          `json:"ttl"`
}

// IsExpired 检查实例是否已过期
func (p *ProxyInstance) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

// RemainingTime 获取剩余时间(秒)
func (p *ProxyInstance) RemainingTime() int {
	remaining := time.Until(p.ExpiresAt).Seconds()
	if remaining < 0 {
		return 0
	}
	return int(remaining)
}

// ToResponse 转换为API响应格式
func (p *ProxyInstance) ToResponse() map[string]interface{} {
	return map[string]interface{}{
		"instance_id": p.ID,
		"host":        p.Host,
		"port":        p.Port,
		"protocol":    p.Protocol,
		"node_name":   p.NodeName,
		"node_delay":  p.NodeDelay,
		"created_at":  p.CreatedAt.Format(time.RFC3339),
		"expires_at":  p.ExpiresAt.Format(time.RFC3339),
		"ttl":         p.TTL,
		"remaining":   p.RemainingTime(),
	}
}
