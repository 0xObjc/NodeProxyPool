package proxy

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

var (
	ErrNoAvailablePort = errors.New("no available port in range")
	ErrInvalidPort     = errors.New("invalid port number")
)

// PortAllocator 端口分配器
// 负责管理端口池,避免端口冲突
type PortAllocator struct {
	minPort   int
	maxPort   int
	usedPorts map[int]bool
	mu        sync.RWMutex
}

// NewPortAllocator 创建端口分配器
func NewPortAllocator(minPort, maxPort int) (*PortAllocator, error) {
	if minPort < 1024 || minPort > 65535 {
		return nil, ErrInvalidPort
	}
	if maxPort < minPort || maxPort > 65535 {
		return nil, ErrInvalidPort
	}

	return &PortAllocator{
		minPort:   minPort,
		maxPort:   maxPort,
		usedPorts: make(map[int]bool),
	}, nil
}

// Allocate 分配一个可用端口
// 三重保护机制:
// 1. 检查内存状态
// 2. 实际绑定测试
// 3. 标记已使用
func (p *PortAllocator) Allocate() (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 遍历端口池
	for port := p.minPort; port <= p.maxPort; port++ {
		// 1. 检查内存状态
		if p.usedPorts[port] {
			continue
		}

		// 2. 实际绑定测试
		if !p.isPortAvailable(port) {
			continue
		}

		// 3. 标记已使用
		p.usedPorts[port] = true
		return port, nil
	}

	return 0, ErrNoAvailablePort
}

// Release 释放端口
func (p *PortAllocator) Release(port int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if port < p.minPort || port > p.maxPort {
		return ErrInvalidPort
	}

	delete(p.usedPorts, port)
	return nil
}

// IsUsed 检查端口是否已被使用
func (p *PortAllocator) IsUsed(port int) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.usedPorts[port]
}

// GetUsedCount 获取已使用端口数量
func (p *PortAllocator) GetUsedCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.usedPorts)
}

// GetAvailableCount 获取可用端口数量
func (p *PortAllocator) GetAvailableCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	total := p.maxPort - p.minPort + 1
	return total - len(p.usedPorts)
}

// isPortAvailable 测试端口是否可用
// 通过实际尝试绑定来验证
func (p *PortAllocator) isPortAvailable(port int) bool {
	// 测试 TCP 端口
	addr := net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", port))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}
