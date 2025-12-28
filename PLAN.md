# 动态代理池实现计划

## 项目概述

将机场订阅转化为**多入口、API驱动、带生命周期管理**的本地动态代理池。通过Golang + Mihomo内核实现低成本的商用动态IP池体验。

## 核心功能

1. **订阅解析**: 自动抓取机场订阅,解析Vmess/Trojan/Shadowsocks等协议节点
2. **动态分配**: 通过`getProxy` API动态开启独立的Socks5/HTTP端口
3. **TTL控制**: 固定存活时间(默认30分钟),到期自动关闭释放资源
4. **健康检查**: 定期测试节点延迟和可用性,自动过滤失效节点
5. **智能选择**: 支持按延迟、类型、地区等条件过滤节点

## 技术架构

### 技术栈
- **语言**: Golang 1.21+
- **代理内核**: Mihomo (作为Go库集成,非独立进程)
- **Web框架**: Gin
- **配置**: YAML
- **日志**: Zap

### 核心模块

```
┌─────────────────────────────────────────────────────────────┐
│                        HTTP API Server                       │
│              POST /api/getProxy, releaseProxy               │
└──────────────────┬──────────────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────────────┐
│                    Proxy Manager                             │
│   - 创建/释放代理实例                                          │
│   - 协调端口分配、节点选择、监听器创建                          │
└──────┬──────────────────────┬──────────────────────┬────────┘
       │                      │                      │
┌──────▼────────┐    ┌───────▼────────┐    ┌────────▼────────┐
│Port Allocator │    │  Node Pool     │    │ Mihomo Adapter  │
│ 端口池管理     │    │  节点池管理     │    │ 内核适配层       │
│ 20000-30000   │    │  健康检查       │    │ 监听器创建       │
└───────────────┘    └────────┬───────┘    └─────────────────┘
                              │
                     ┌────────▼────────┐
                     │ Subscription    │
                     │ Manager         │
                     │ 订阅抓取/解析    │
                     └─────────────────┘
```

### 关键数据流

**getProxy请求流程:**
```
1. API收到请求 (延迟<500ms, TTL=1800s)
2. ProxyManager从NodePool筛选节点 (按延迟/可用性过滤)
3. PortAllocator分配可用端口 (如20001)
4. MihomoAdapter创建监听器 (绑定节点到端口)
5. 返回代理信息 {host: "127.0.0.1", port: 20001}
6. TTLCleaner在到期时自动清理
```

## 项目结构

```
proxyPool/
├── cmd/server/main.go              # 应用入口
├── internal/
│   ├── config/                     # 配置管理
│   │   ├── config.go              # 配置结构
│   │   └── loader.go              # YAML加载
│   ├── subscription/               # 订阅管理
│   │   ├── fetcher.go             # HTTP抓取
│   │   ├── parser.go              # 节点解析(Mihomo convert)
│   │   └── manager.go             # 自动更新
│   ├── node/                       # 节点管理
│   │   ├── node.go                # 节点结构
│   │   ├── pool.go                # 节点池
│   │   └── health_checker.go     # 健康检查
│   ├── proxy/                      # 代理管理
│   │   ├── port_allocator.go     # 端口分配
│   │   ├── proxy_instance.go     # 代理实例
│   │   ├── manager.go             # 核心管理器
│   │   └── ttl_cleaner.go        # TTL清理
│   ├── mihomo/                     # Mihomo集成
│   │   ├── adapter.go             # 核心适配器
│   │   ├── listener.go            # 监听器管理
│   │   └── config_builder.go     # 配置构建
│   └── api/                        # API服务
│       ├── handler.go             # 路由处理
│       ├── middleware.go          # 中间件
│       └── server.go              # HTTP服务器
├── pkg/
│   ├── logger/logger.go           # 日志封装
│   └── utils/                     # 工具函数
├── configs/config.yaml            # 默认配置
├── go.mod
└── README.md
```

## 关键实现细节

### 1. Mihomo集成 - 多端口实现

**核心挑战**: 如何为每个getProxy请求创建独立的监听端口并绑定到特定节点?

**实现方案** (`internal/mihomo/adapter.go`):

```go
// 为每个代理实例创建独立监听器
func (m *MihomoAdapter) CreateListener(
    protocol string,
    port int,
    proxy C.Proxy,
) error {
    // 1. 构建最小化配置,只包含当前代理节点
    config := m.buildMinimalConfig(proxy)

    // 2. 应用配置到Mihomo
    executor.ApplyConfig(config, false)

    // 3. 创建对应协议的监听器
    switch protocol {
    case "socks5":
        listener.ReCreateSocks(port, m.tunnel)
    case "http":
        listener.ReCreateHTTP(port, m.tunnel)
    case "mixed":
        listener.ReCreateMixed(port, m.tunnel)
    }
}
```

**关键API**:
- `github.com/metacubex/mihomo/hub/executor` - 配置解析与应用
- `github.com/metacubex/mihomo/listener` - 动态监听器创建
- `github.com/metacubex/mihomo/adapter` - 代理适配器

### 2. 端口分配器 - 避免冲突

**三重保护机制** (`internal/proxy/port_allocator.go`):

```go
func (p *PortAllocator) Allocate() (int, error) {
    p.mu.Lock()
    defer p.mu.Unlock()

    // 遍历端口池 (20000-30000)
    for port := p.minPort; port <= p.maxPort; port++ {
        // 1. 检查内存状态
        if !p.usedPorts[port] {
            // 2. 实际绑定测试
            if p.isPortAvailable(port) {
                // 3. 标记已使用
                p.usedPorts[port] = true
                return port, nil
            }
        }
    }
    return 0, ErrNoAvailablePort
}
```

### 3. 代理管理器 - 核心编排

**CreateProxy流程** (`internal/proxy/manager.go`):

```go
func (m *Manager) CreateProxy(req *CreateProxyRequest) (*ProxyInstance, error) {
    // 1. 智能选择节点 (按延迟/类型/地区过滤)
    node := m.selectNode(req)

    // 2. 分配端口
    port := m.allocator.Allocate()

    // 3. 创建实例对象
    instance := &ProxyInstance{
        ID:        uuid.New().String(),
        Port:      port,
        Node:      node,
        ExpiresAt: time.Now().Add(req.TTL),
    }

    // 4. 创建Mihomo监听器
    m.mihomoAdapter.CreateListener(req.Protocol, port, node.Proxy)

    // 5. 注册实例
    m.instances[instance.ID] = instance
    m.portMap[port] = instance.ID

    return instance
}
```

### 4. 订阅解析 - Mihomo转换

**使用Mihomo内置解析器** (`internal/subscription/parser.go`):

```go
import "github.com/metacubex/mihomo/adapter/provider/convert"

func (p *Parser) Parse(content []byte) ([]*Node, error) {
    // 1. Base64解码订阅内容
    decoded, _ := base64.StdEncoding.DecodeString(string(content))

    // 2. 使用Mihomo解析器 (支持vmess/trojan/ss等)
    proxies, err := convert.ConvertsV2Ray(decoded)

    // 3. 转换为节点对象
    for _, proxyConfig := range proxies {
        node := &Node{
            Name:      proxyConfig["name"].(string),
            Type:      proxyConfig["type"].(string),
            RawConfig: proxyConfig,
        }
        // 4. 创建Mihomo Proxy对象
        node.Proxy, _ = adapter.ParseProxy(proxyConfig)
    }
}
```

### 5. TTL自动清理

**定时扫描到期实例** (`internal/proxy/ttl_cleaner.go`):

```go
func (c *TTLCleaner) Start() {
    ticker := time.NewTicker(30 * time.Second) // 每30秒检查
    go func() {
        for range ticker.C {
            now := time.Now()
            for id, instance := range c.manager.instances {
                if now.After(instance.ExpiresAt) {
                    c.manager.ReleaseProxy(id) // 关闭监听器+释放端口
                }
            }
        }
    }()
}
```

### 6. 健康检查

**并发测试节点延迟** (`internal/node/health_checker.go`):

```go
func (h *HealthChecker) checkNode(n *Node) (int, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    start := time.Now()
    // 使用Mihomo的URLTest测试节点
    _, err := n.Proxy.URLTest(ctx, "https://www.google.com/generate_204")

    delay := int(time.Since(start).Milliseconds())
    n.Delay = delay
    n.Available = (err == nil)
    return delay, err
}
```

## API接口设计

### 1. 获取代理 `POST /api/getProxy`

**请求:**
```json
{
  "protocol": "socks5",      // socks5/http/mixed (可选,默认socks5)
  "ttl": 1800,               // 秒 (可选,默认配置值)
  "max_delay": 500,          // 最大延迟ms (可选,过滤节点)
  "node_type": "vmess",      // vmess/trojan/ss (可选)
  "region": "HK"             // 地区关键字 (可选)
}
```

**响应:**
```json
{
  "code": 0,
  "data": {
    "instance_id": "550e8400-e29b-41d4-a716-446655440000",
    "host": "127.0.0.1",
    "port": 20001,
    "protocol": "socks5",
    "node_name": "香港-01",
    "node_delay": 123,
    "expires_at": "2025-12-28T10:30:00Z",
    "ttl": 1800
  }
}
```

### 2. 释放代理 `POST /api/releaseProxy`

**请求:**
```json
{
  "instance_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 3. 查询实例 `GET /api/getInstance/:id`

返回实例详情、剩余时间、流量统计等。

### 4. 列出所有实例 `GET /api/listInstances`

返回当前活跃的所有代理实例列表。

### 5. 节点池状态 `GET /api/nodePool`

返回节点总数、可用数、按类型分布、平均延迟等。

### 6. 手动更新订阅 `POST /api/subscription/update`

触发立即更新所有订阅源。

## 配置文件

**`configs/config.yaml`:**

```yaml
server:
  host: 0.0.0.0
  port: 8080
  mode: release

proxy:
  port_range:
    min: 20000              # 端口池起始
    max: 30000              # 端口池结束
  default_ttl: 30m          # 默认存活时间
  max_instances: 100        # 最大并发实例数
  default_protocol: socks5

subscription:
  sources:
    - name: "机场1"
      url: "https://example.com/sub?token=xxx"
      enabled: true
    - name: "机场2"
      url: "https://example2.com/api/v1/client/subscribe?token=yyy"
      enabled: true
  update_interval: 6h       # 自动更新间隔
  timeout: 30s

health_check:
  enabled: true
  interval: 5m              # 检测间隔
  timeout: 10s
  url: "https://www.google.com/generate_204"
  max_delay: 1000           # 超过此延迟标记为不可用

log:
  level: info
  file: "./logs/proxypool.log"
```

## 实现步骤

### Phase 1: 基础框架 (首要任务)

**目标**: 建立项目骨架

**任务**:
1. 初始化Go module:
   ```bash
   go mod init proxyPool
   go get github.com/metacubex/mihomo
   go get github.com/gin-gonic/gin
   go get gopkg.in/yaml.v3
   go get github.com/google/uuid
   go get go.uber.org/zap
   ```

2. 创建目录结构 (按照上述项目结构)

3. 实现配置模块:
   - `internal/config/config.go` - 定义Config结构
   - `internal/config/loader.go` - 加载YAML配置
   - `configs/config.yaml` - 创建示例配置

4. 实现日志模块:
   - `pkg/logger/logger.go` - Zap日志封装

5. 实现应用入口:
   - `cmd/server/main.go` - 初始化流程、启动HTTP服务

**关键文件**:
- `cmd/server/main.go`
- `internal/config/config.go`
- `internal/config/loader.go`
- `configs/config.yaml`

### Phase 2: 订阅和节点管理

**目标**: 实现订阅抓取和节点解析

**任务**:
1. 实现订阅抓取:
   - `internal/subscription/fetcher.go` - HTTP客户端,下载订阅内容

2. 实现节点解析:
   - `internal/subscription/parser.go` - 集成Mihomo的`convert.ConvertsV2Ray`

3. 实现节点池:
   - `internal/node/node.go` - Node结构定义
   - `internal/node/pool.go` - 节点存储、查询、过滤

4. 实现订阅管理:
   - `internal/subscription/manager.go` - 整合抓取和解析,定期更新

5. 实现API:
   - `GET /api/nodePool` - 查看节点状态
   - `POST /api/subscription/update` - 手动更新

**关键文件**:
- `internal/subscription/parser.go` ⭐ (使用Mihomo解析)
- `internal/node/pool.go`
- `internal/subscription/manager.go`

### Phase 3: Mihomo集成 (核心难点)

**目标**: 集成Mihomo内核,实现代理功能

**任务**:
1. 深入研究Mihomo:
   - 阅读executor/listener/adapter包源码
   - 测试多端口创建是否支持
   - 确定节点绑定方案

2. 实现Mihomo适配器:
   - `internal/mihomo/adapter.go` - CreateListener/CloseListener
   - `internal/mihomo/config_builder.go` - 构建Mihomo配置

3. 单节点测试:
   - 创建固定端口的socks5代理
   - 验证流量转发

4. 多端口测试:
   - 同时创建多个端口
   - 验证每个端口绑定正确节点

**关键文件**:
- `internal/mihomo/adapter.go` ⭐⭐⭐ (核心适配层)
- `internal/mihomo/config_builder.go`

**技术风险**:
- 如Mihomo不支持多端口绑定不同节点,需要调整方案
- 备选方案: 使用selector + API动态切换,或自实现socks5服务器

### Phase 4: 代理管理核心

**目标**: 完整的代理生命周期管理

**任务**:
1. 实现端口分配:
   - `internal/proxy/port_allocator.go` - 端口池管理

2. 实现代理实例:
   - `internal/proxy/proxy_instance.go` - ProxyInstance结构

3. 实现代理管理器:
   - `internal/proxy/manager.go` - CreateProxy/ReleaseProxy核心逻辑

4. 实现TTL清理:
   - `internal/proxy/ttl_cleaner.go` - 定时扫描过期实例

5. 实现API:
   - `POST /api/getProxy`
   - `POST /api/releaseProxy`
   - `GET /api/getInstance/:id`
   - `GET /api/listInstances`

**关键文件**:
- `internal/proxy/manager.go` ⭐⭐⭐ (核心编排)
- `internal/proxy/port_allocator.go` ⭐
- `internal/proxy/ttl_cleaner.go`

### Phase 5: 健康检查

**目标**: 节点健康监测和自动过滤

**任务**:
1. 实现健康检查器:
   - `internal/node/health_checker.go` - 使用Mihomo的URLTest

2. 集成到节点选择:
   - ProxyManager选择节点时过滤不可用节点
   - 按延迟排序

3. 实现API:
   - `GET /api/nodes` - 节点列表(含健康状态)

**关键文件**:
- `internal/node/health_checker.go`

### Phase 6: 完善和优化

**目标**: 错误处理、性能优化、文档

**任务**:
1. 错误处理完善
2. 参数验证
3. 并发优化(锁粒度)
4. 优雅关闭
5. 压力测试
6. 文档编写(API文档、README)

## 关键文件优先级

**P0 (最关键)**:
- `internal/mihomo/adapter.go` - Mihomo集成,决定方案可行性
- `internal/proxy/manager.go` - 核心编排逻辑
- `cmd/server/main.go` - 应用入口

**P1 (重要)**:
- `internal/subscription/parser.go` - 节点数据来源
- `internal/proxy/port_allocator.go` - 资源管理
- `internal/node/pool.go` - 节点存储

**P2 (增强)**:
- `internal/proxy/ttl_cleaner.go` - 自动化管理
- `internal/node/health_checker.go` - 健康监测
- `internal/api/handler.go` - API接口

## 技术风险和应对

### 风险1: Mihomo多端口支持限制

**风险**: Mihomo可能不支持同时创建多个独立监听器指向不同节点

**应对**:
- Phase 3深入测试验证
- 备选方案A: 使用全局配置 + selector组 + API动态切换
- 备选方案B: 自实现socks5服务器,调用Mihomo的Proxy对象转发
- 备选方案C: 每个实例fork独立Mihomo进程(最后手段,资源开销大)

### 风险2: 性能瓶颈

**应对**:
- 严格限制max_instances (如100)
- 优化锁粒度(读写锁)
- 性能profiling
- 对象池减少GC

### 风险3: 端口资源耗尽

**应对**:
- 合理默认TTL (不要过长)
- 监控端口使用率
- 支持扩展端口范围

## 预期成果

完成后可实现:

1. ✅ 自动抓取和解析机场订阅
2. ✅ 一个API调用获得独立代理端口
3. ✅ 自动健康检查和节点过滤
4. ✅ 固定TTL自动清理
5. ✅ 支持100+并发代理实例
6. ✅ 延迟<500ms、地区、协议等智能过滤

**使用示例**:

```bash
# 获取代理
curl -X POST http://localhost:8080/api/getProxy \
  -d '{"max_delay": 300, "region": "HK"}'

# 响应: {"port": 20001, "host": "127.0.0.1"}

# 使用代理
curl -x socks5://127.0.0.1:20001 https://api.ipify.org

# 30分钟后自动释放
```

## 下一步行动

**立即开始 Phase 1**:
1. ✅ 创建目录结构
2. ✅ 初始化Go module
3. ✅ 实现基础配置和日志
4. ✅ 搭建HTTP服务器框架
5. ✅ 验证基础框架可运行

**预计时间**: 总计14-18天完成Phase 1-6