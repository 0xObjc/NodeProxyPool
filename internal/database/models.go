package database

import (
	"encoding/json"
	"time"

	"proxyPool/internal/node"

	"github.com/metacubex/mihomo/adapter"
)

// SubscriptionSource 订阅源行
type SubscriptionSource struct {
	ID              int64
	Name            string
	URL             string
	Enabled         bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	LastFetchAt     *time.Time
	LastFetchStatus string
	LastFetchError  string
	NodeCount       int
}

// NodeRecord 节点持久化行
type NodeRecord struct {
	ID                   string
	Name                 string
	Type                 string
	Server               string
	Port                 int
	RawConfigJSON        string
	Delay                int
	LastCheck            *time.Time
	Available            bool
	ActiveCount          int
	TotalUsed            int
	SubscriptionSourceID *int64
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// ToNode 转为运行时节点
func (n *NodeRecord) ToNode() (*node.Node, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(n.RawConfigJSON), &raw); err != nil {
		return nil, err
	}

	node := &node.Node{
		ID:          n.ID,
		Name:        n.Name,
		Type:        n.Type,
		Server:      n.Server,
		Port:        n.Port,
		RawConfig:   raw,
		Delay:       n.Delay,
		Available:   n.Available,
		ActiveCount: n.ActiveCount,
		TotalUsed:   n.TotalUsed,
	}

	if n.LastCheck != nil {
		node.LastCheck = *n.LastCheck
	}

	if proxy, err := adapter.ParseProxy(raw); err == nil {
		node.Proxy = proxy
	}

	return node, nil
}

// SystemConfig 系统配置行
type SystemConfig struct {
	Key         string
	Value       string
	ValueType   string
	Description string
	Category    string
	UpdatedAt   time.Time
}

// ConfigValue 表示单个配置项值
type ConfigValue struct {
	Value       string
	Type        string
	Category    string
	Description string
}
