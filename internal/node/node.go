package node

import (
	"time"

	C "github.com/metacubex/mihomo/constant"
)

// Node 代理节点
type Node struct {
	ID        string                 `json:"id"`         // 唯一标识
	Name      string                 `json:"name"`       // 节点名称
	Type      string                 `json:"type"`       // vmess/trojan/ss等
	Server    string                 `json:"server"`     // 服务器地址
	Port      int                    `json:"port"`       // 服务器端口
	RawConfig map[string]interface{} `json:"raw_config"` // 原始配置
	Proxy     C.Proxy                `json:"-"`          // Mihomo代理对象

	// 健康状态
	Delay     int       `json:"delay"`      // 延迟(ms),-1表示不可用
	LastCheck time.Time `json:"last_check"` // 上次检测时间
	Available bool      `json:"available"`  // 是否可用

	// 使用统计
	ActiveCount int `json:"active_count"` // 当前活跃连接数
	TotalUsed   int `json:"total_used"`   // 总使用次数
}

// FilterOptions 节点过滤选项
type FilterOptions struct {
	MaxDelay  int    // 最大延迟(ms), 0表示不限制
	NodeType  string // 节点类型, 空表示不限制
	Region    string // 地区关键字, 空表示不限制
	Available bool   // 是否只返回可用节点
}
