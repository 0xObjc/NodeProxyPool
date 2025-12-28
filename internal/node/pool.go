package node

import (
	"sort"
	"strings"
	"sync"
)

// Pool 节点池
type Pool struct {
	nodes       map[string]*Node   // ID -> Node
	nodesByType map[string][]*Node // Type -> Nodes
	mu          sync.RWMutex
}

// NewPool 创建节点池
func NewPool() *Pool {
	return &Pool{
		nodes:       make(map[string]*Node),
		nodesByType: make(map[string][]*Node),
	}
}

// Add 添加节点
func (p *Pool) Add(node *Node) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.nodes[node.ID] = node

	// 更新类型索引
	p.nodesByType[node.Type] = append(p.nodesByType[node.Type], node)
}

// UpdateNodes 批量更新节点(替换所有节点)
func (p *Pool) UpdateNodes(nodes []*Node) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 清空现有节点
	p.nodes = make(map[string]*Node)
	p.nodesByType = make(map[string][]*Node)

	// 添加新节点
	for _, node := range nodes {
		p.nodes[node.ID] = node
		p.nodesByType[node.Type] = append(p.nodesByType[node.Type], node)
	}
}

// Get 获取节点
func (p *Pool) Get(id string) (*Node, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	node, ok := p.nodes[id]
	return node, ok
}

// GetAll 获取所有节点
func (p *Pool) GetAll() []*Node {
	p.mu.RLock()
	defer p.mu.RUnlock()

	nodes := make([]*Node, 0, len(p.nodes))
	for _, node := range p.nodes {
		nodes = append(nodes, node)
	}

	return nodes
}

// Filter 过滤节点
func (p *Pool) Filter(opts FilterOptions) []*Node {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var result []*Node

	for _, node := range p.nodes {
		// 过滤可用性
		if opts.Available && !node.Available {
			continue
		}

		// 过滤延迟
		if opts.MaxDelay > 0 && (node.Delay < 0 || node.Delay > opts.MaxDelay) {
			continue
		}

		// 过滤类型
		if opts.NodeType != "" && node.Type != opts.NodeType {
			continue
		}

		// 过滤地区(从节点名称中匹配)
		if opts.Region != "" && !strings.Contains(node.Name, opts.Region) {
			continue
		}

		result = append(result, node)
	}

	// 按延迟排序(延迟低的在前)
	sort.Slice(result, func(i, j int) bool {
		// 不可用节点排到最后
		if result[i].Delay < 0 {
			return false
		}
		if result[j].Delay < 0 {
			return true
		}
		return result[i].Delay < result[j].Delay
	})

	return result
}

// Count 获取节点总数
func (p *Pool) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.nodes)
}

// CountAvailable 获取可用节点数
func (p *Pool) CountAvailable() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	count := 0
	for _, node := range p.nodes {
		if node.Available {
			count++
		}
	}
	return count
}

// CountByType 按类型统计节点数
func (p *Pool) CountByType() map[string]int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make(map[string]int)
	for nodeType, nodes := range p.nodesByType {
		result[nodeType] = len(nodes)
	}
	return result
}

// AvgDelay 计算平均延迟(仅可用节点)
func (p *Pool) AvgDelay() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	totalDelay := 0
	count := 0

	for _, node := range p.nodes {
		if node.Available && node.Delay > 0 {
			totalDelay += node.Delay
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return totalDelay / count
}
