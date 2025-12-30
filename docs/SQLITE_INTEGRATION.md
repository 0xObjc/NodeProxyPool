# SQLite 集成方案

## 1. 概述

本方案将 ProxyPool 从基于配置文件的架构转变为数据库支持的架构,实现运行时配置管理。在保持内存性能的同时,增加持久化、动态配置和 API 驱动的订阅源管理功能。

### 1.1 用户需求

- **订阅源管理**: 数据库 + API 管理(支持运行时 CRUD 操作)
- **节点历史**: 仅保留最新状态(不保存历史延迟记录)
- **实例历史**: 仅跟踪活跃实例(不保存使用历史)
- **动态配置**: 健康检查、代理池、订阅更新参数支持运行时修改

### 1.2 核心特性

- 混合存储架构:内存为主(性能),数据库为辅(持久化)
- 异步写入:数据库操作不阻塞主流程
- 平滑迁移:首次启动自动从 config.yaml 导入
- API 驱动:所有配置和订阅源可通过 API 管理

---

## 2. 数据库设计

### 2.1 表结构

#### subscription_sources - 订阅源表

```sql
CREATE TABLE subscription_sources (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    url TEXT NOT NULL,
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_fetch_at DATETIME,
    last_fetch_status TEXT,
    last_fetch_error TEXT,
    node_count INTEGER DEFAULT 0
);

CREATE INDEX idx_subscription_enabled ON subscription_sources(enabled);
CREATE INDEX idx_subscription_updated ON subscription_sources(updated_at);
```

**字段说明**:
- `id`: 自增主键
- `name`: 订阅源名称(唯一)
- `url`: 订阅地址
- `enabled`: 是否启用(1=启用, 0=禁用)
- `last_fetch_at`: 最后抓取时间
- `last_fetch_status`: 抓取状态(success/failed/pending)
- `last_fetch_error`: 错误信息
- `node_count`: 最后抓取的节点数量

#### nodes - 节点表

```sql
CREATE TABLE nodes (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    server TEXT NOT NULL,
    port INTEGER NOT NULL,
    raw_config_json TEXT NOT NULL,

    -- 健康状态
    delay INTEGER NOT NULL DEFAULT -1,
    last_check DATETIME,
    available INTEGER NOT NULL DEFAULT 0,

    -- 使用统计
    active_count INTEGER NOT NULL DEFAULT 0,
    total_used INTEGER NOT NULL DEFAULT 0,

    -- 元数据
    subscription_source_id INTEGER,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (subscription_source_id)
        REFERENCES subscription_sources(id) ON DELETE SET NULL
);

CREATE INDEX idx_nodes_type ON nodes(type);
CREATE INDEX idx_nodes_available ON nodes(available);
CREATE INDEX idx_nodes_delay ON nodes(delay);
CREATE INDEX idx_nodes_type_available ON nodes(type, available);
CREATE INDEX idx_nodes_subscription ON nodes(subscription_source_id);
```

**字段说明**:
- `id`: 节点唯一标识(配置哈希)
- `type`: 节点类型(vmess/trojan/ss 等)
- `raw_config_json`: 原始配置(JSON 格式)
- `delay`: 延迟(毫秒,-1 表示不可用)
- `available`: 是否可用(1=可用, 0=不可用)
- `active_count`: 当前活跃连接数
- `total_used`: 总使用次数

#### system_config - 系统配置表

```sql
CREATE TABLE system_config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    value_type TEXT NOT NULL,
    description TEXT,
    category TEXT NOT NULL,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_config_category ON system_config(category);
```

**字段说明**:
- `key`: 配置键(如 proxy.port_range.min)
- `value`: 配置值(JSON 序列化)
- `value_type`: 值类型(string/int/duration/bool)
- `category`: 分类(proxy/health_check/subscription/server/log)

#### config_history - 配置历史表(可选)

```sql
CREATE TABLE config_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    config_key TEXT NOT NULL,
    old_value TEXT,
    new_value TEXT NOT NULL,
    changed_by TEXT,
    changed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_config_history_key ON config_history(config_key);
CREATE INDEX idx_config_history_time ON config_history(changed_at);
```

### 2.2 配置键定义

#### proxy 分类
- `proxy.port_range.min` (int): 端口范围最小值
- `proxy.port_range.max` (int): 端口范围最大值
- `proxy.default_ttl` (duration): 默认 TTL(秒)
- `proxy.max_instances` (int): 最大实例数
- `proxy.default_protocol` (string): 默认协议(socks5/http/mixed)

#### health_check 分类
- `health_check.enabled` (bool): 是否启用健康检查
- `health_check.interval` (duration): 检查间隔(秒)
- `health_check.timeout` (duration): 检查超时(秒)
- `health_check.url` (string): 测试 URL
- `health_check.max_delay` (int): 最大延迟(毫秒)

#### subscription 分类
- `subscription.update_interval` (duration): 更新间隔(秒)
- `subscription.timeout` (duration): 请求超时(秒)

#### server 分类
- `server.host` (string): 监听地址
- `server.port` (int): 监听端口
- `server.mode` (string): 运行模式(debug/release)

#### log 分类
- `log.level` (string): 日志级别(debug/info/warn/error)
- `log.file` (string): 日志文件路径

---

## 3. 架构设计

### 3.1 包结构

创建新包 `internal/database/`:

```
internal/database/
├── db.go                    # 数据库连接和初始化
├── migrations.go            # Schema 迁移
├── models.go                # 数据库模型结构体
├── subscription_repo.go     # 订阅源 CRUD 操作
├── node_repo.go            # 节点持久化操作
└── config_repo.go          # 动态配置操作
```

### 3.2 核心接口

#### DB 接口 (db.go)

```go
type DB struct {
    conn *sql.DB
    mu   sync.RWMutex
}

func New(dbPath string) (*DB, error)
func (db *DB) Close() error
func (db *DB) Begin() (*sql.Tx, error)
func (db *DB) Migrate() error
func (db *DB) IsFirstRun() bool
func (db *DB) SeedFromConfig(cfg *config.Config) error
```

#### SubscriptionRepository (subscription_repo.go)

```go
type SubscriptionRepository struct {
    db *DB
}

func (r *SubscriptionRepository) Create(source *SubscriptionSource) error
func (r *SubscriptionRepository) Update(id int64, source *SubscriptionSource) error
func (r *SubscriptionRepository) Delete(id int64) error
func (r *SubscriptionRepository) GetByID(id int64) (*SubscriptionSource, error)
func (r *SubscriptionRepository) GetAll() ([]*SubscriptionSource, error)
func (r *SubscriptionRepository) GetEnabled() ([]*SubscriptionSource, error)
func (r *SubscriptionRepository) UpdateFetchStatus(id int64, status, error string, nodeCount int) error
```

#### NodeRepository (node_repo.go)

```go
type NodeRepository struct {
    db *DB
}

func (r *NodeRepository) Upsert(node *Node) error
func (r *NodeRepository) UpsertBatch(nodes []*Node) error
func (r *NodeRepository) GetAll() ([]*Node, error)
func (r *NodeRepository) GetByID(id string) (*Node, error)
func (r *NodeRepository) UpdateHealthStatus(id string, delay int, available bool) error
func (r *NodeRepository) UpdateUsageStats(id string, activeCount, totalUsed int) error
func (r *NodeRepository) DeleteBySubscription(subscriptionID int64) error
func (r *NodeRepository) DeleteStale(cutoffTime time.Time) error
```

#### ConfigRepository (config_repo.go)

```go
type ConfigRepository struct {
    db *DB
}

func (r *ConfigRepository) Get(key string) (string, error)
func (r *ConfigRepository) GetByCategory(category string) (map[string]string, error)
func (r *ConfigRepository) GetAll() (map[string]string, error)
func (r *ConfigRepository) Set(key, value, valueType, category string) error
func (r *ConfigRepository) SetBatch(configs map[string]ConfigValue) error
func (r *ConfigRepository) InitializeDefaults(cfg *config.Config) error
```

### 3.3 数据库连接配置

```go
// internal/database/db.go

func New(dbPath string) (*DB, error) {
    // 使用 WAL 模式提高并发性能
    conn, err := sql.Open("sqlite3",
        dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
    if err != nil {
        return nil, err
    }

    // 连接池设置
    conn.SetMaxOpenConns(25)  // SQLite WAL 模式支持良好并发
    conn.SetMaxIdleConns(5)
    conn.SetConnMaxLifetime(time.Hour)

    return &DB{conn: conn}, nil
}
```

---

## 4. 迁移策略

### 4.1 首次启动流程

```
1. 检测数据库文件是否存在
   ↓ 不存在
2. 创建数据库文件
   ↓
3. 运行 Schema 迁移(创建所有表和索引)
   ↓
4. 从 config.yaml 读取配置
   ↓
5. 将订阅源插入 subscription_sources 表
   ↓
6. 将所有配置值插入 system_config 表
   ↓
7. 记录日志:"Database initialized from config.yaml"
   ↓
8. 继续正常启动流程(使用数据库作为数据源)
```

### 4.2 后续启动流程

```
1. 打开数据库连接
   ↓
2. 从 system_config 表加载所有配置
   ↓
3. 重构 config.Config 结构体
   ↓
4. 验证配置有效性
   ↓
5. 从 subscription_sources 表加载订阅源
   ↓
6. 从 nodes 表加载节点到内存
   ↓
7. 启动所有服务
```

### 4.3 迁移代码示例

```go
// cmd/server/main.go

func main() {
    // 初始化数据库
    db, err := database.New("./data/proxypool.db")
    if err != nil {
        logger.Fatal("Failed to initialize database", zap.Error(err))
    }
    defer db.Close()

    // 运行迁移
    if err := db.Migrate(); err != nil {
        logger.Fatal("Failed to run migrations", zap.Error(err))
    }

    // 检查是否首次运行
    if db.IsFirstRun() {
        logger.Info("First run detected, seeding from config.yaml")

        // 从 YAML 加载配置
        cfg, err := config.Load(*configFile)
        if err != nil {
            logger.Fatal("Failed to load config", zap.Error(err))
        }

        // 导入到数据库
        if err := db.SeedFromConfig(cfg); err != nil {
            logger.Fatal("Failed to seed database", zap.Error(err))
        }

        logger.Info("Database initialized successfully")
    }

    // 从数据库加载配置
    dynamicConfig, err := config.NewDynamicConfig(db)
    if err != nil {
        logger.Fatal("Failed to load dynamic config", zap.Error(err))
    }

    cfg := dynamicConfig.Get()

    // 继续初始化其他组件...
}
```

### 4.4 向后兼容性

**策略**: 数据库优先(Database-First)

- 首次运行后,数据库成为唯一数据源
- config.yaml 仅用于初始化,之后成为只读参考
- 所有后续配置更改通过 API 进行
- 清晰的迁移路径,避免混淆

**用户提示**:
```
INFO: Database initialized from config.yaml
INFO: Future configuration changes should be made via API
INFO: config.yaml will no longer be read on subsequent startups
```

---

## 5. 组件集成

### 5.1 订阅管理器集成

**文件**: `internal/subscription/manager.go`

**修改内容**:

```go
type Manager struct {
    fetcher  *Fetcher
    parser   *Parser
    nodePool *node.Pool
    db       *database.DB                          // 新增
    subRepo  *database.SubscriptionRepository      // 新增
    interval time.Duration
    stopCh   chan struct{}
    mu       sync.RWMutex
}

func NewManager(
    cfg *config.Config,
    nodePool *node.Pool,
    db *database.DB,  // 新增参数
) *Manager {
    return &Manager{
        fetcher:  NewFetcher(cfg.Subscription.Timeout),
        parser:   NewParser(),
        nodePool: nodePool,
        db:       db,
        subRepo:  database.NewSubscriptionRepository(db),
        interval: cfg.Subscription.UpdateInterval,
        stopCh:   make(chan struct{}),
    }
}

func (m *Manager) UpdateAll() {
    // 从数据库加载启用的订阅源
    sources, err := m.subRepo.GetEnabled()
    if err != nil {
        logger.Error("Failed to load subscription sources", zap.Error(err))
        return
    }

    // 并行抓取所有订阅源
    var wg sync.WaitGroup
    var mu sync.Mutex
    var allNodes []*node.Node

    for _, source := range sources {
        wg.Add(1)
        go func(src *database.SubscriptionSource) {
            defer wg.Done()

            // 抓取订阅
            content, err := m.fetcher.Fetch(src.URL)
            if err != nil {
                logger.Error("Failed to fetch subscription",
                    zap.String("name", src.Name),
                    zap.Error(err))
                m.subRepo.UpdateFetchStatus(src.ID, "failed", err.Error(), 0)
                return
            }

            // 解析节点
            nodes, err := m.parser.Parse(content)
            if err != nil {
                logger.Error("Failed to parse subscription",
                    zap.String("name", src.Name),
                    zap.Error(err))
                m.subRepo.UpdateFetchStatus(src.ID, "failed", err.Error(), 0)
                return
            }

            // 更新抓取状态
            m.subRepo.UpdateFetchStatus(src.ID, "success", "", len(nodes))

            mu.Lock()
            allNodes = append(allNodes, nodes...)
            mu.Unlock()
        }(source)
    }

    wg.Wait()

    // 更新节点池
    m.nodePool.UpdateNodes(allNodes)
}

// 新增:动态重载订阅源
func (m *Manager) ReloadSources() error {
    m.mu.Lock()
    defer m.mu.Unlock()

    // 触发立即更新
    go m.UpdateAll()
    return nil
}
```

### 5.2 节点池集成

**文件**: `internal/node/pool.go`

**修改内容**:

```go
type Pool struct {
    nodes       map[string]*Node
    nodesByType map[string][]*Node
    db          *database.DB                  // 新增
    nodeRepo    *database.NodeRepository      // 新增
    mu          sync.RWMutex
}

func NewPool(db *database.DB) *Pool {
    return &Pool{
        nodes:       make(map[string]*Node),
        nodesByType: make(map[string][]*Node),
        db:          db,
        nodeRepo:    database.NewNodeRepository(db),
    }
}

// 新增:从数据库加载节点
func (p *Pool) LoadFromDatabase() error {
    p.mu.Lock()
    defer p.mu.Unlock()

    dbNodes, err := p.nodeRepo.GetAll()
    if err != nil {
        return fmt.Errorf("failed to load nodes from database: %w", err)
    }

    for _, dbNode := range dbNodes {
        node := p.convertFromDBNode(dbNode)
        p.nodes[node.ID] = node
        p.nodesByType[node.Type] = append(p.nodesByType[node.Type], node)
    }

    logger.Info("Loaded nodes from database", zap.Int("count", len(dbNodes)))
    return nil
}

// 修改:更新节点时持久化到数据库
func (p *Pool) UpdateNodes(nodes []*Node) {
    p.mu.Lock()
    defer p.mu.Unlock()

    // 更新内存(现有逻辑)
    p.nodes = make(map[string]*Node)
    p.nodesByType = make(map[string][]*Node)

    for _, node := range nodes {
        p.nodes[node.ID] = node
        p.nodesByType[node.Type] = append(p.nodesByType[node.Type], node)
    }

    logger.Info("Updated node pool", zap.Int("total", len(nodes)))

    // 异步持久化到数据库
    go func() {
        if err := p.nodeRepo.UpsertBatch(nodes); err != nil {
            logger.Error("Failed to persist nodes to database", zap.Error(err))
        } else {
            logger.Debug("Persisted nodes to database", zap.Int("count", len(nodes)))
        }
    }()
}

// 辅助方法:转换数据库节点到内存节点
func (p *Pool) convertFromDBNode(dbNode *database.Node) *Node {
    // 反序列化 raw_config_json
    var rawConfig map[string]interface{}
    json.Unmarshal([]byte(dbNode.RawConfigJSON), &rawConfig)

    return &Node{
        ID:          dbNode.ID,
        Name:        dbNode.Name,
        Type:        dbNode.Type,
        Server:      dbNode.Server,
        Port:        dbNode.Port,
        RawConfig:   rawConfig,
        Delay:       dbNode.Delay,
        LastCheck:   dbNode.LastCheck,
        Available:   dbNode.Available,
        ActiveCount: dbNode.ActiveCount,
        TotalUsed:   dbNode.TotalUsed,
    }
}
```

### 5.3 健康检查器集成

**文件**: `internal/node/health_checker.go`

**修改内容**:

```go
type HealthChecker struct {
    pool          *Pool
    nodeRepo      *database.NodeRepository      // 新增
    checkInterval time.Duration
    checkTimeout  time.Duration
    testURL       string
    maxDelay      int
    logger        *zap.Logger
    ctx           context.Context
    cancel        context.CancelFunc
    wg            sync.WaitGroup
}

func NewHealthChecker(
    pool *Pool,
    db *database.DB,  // 新增参数
    checkInterval time.Duration,
    checkTimeout time.Duration,
    testURL string,
    maxDelay int,
    logger *zap.Logger,
) *HealthChecker {
    ctx, cancel := context.WithCancel(context.Background())
    return &HealthChecker{
        pool:          pool,
        nodeRepo:      database.NewNodeRepository(db),
        checkInterval: checkInterval,
        checkTimeout:  checkTimeout,
        testURL:       testURL,
        maxDelay:      maxDelay,
        logger:        logger,
        ctx:           ctx,
        cancel:        cancel,
    }
}

func (h *HealthChecker) checkNode(n *Node) {
    ctx, cancel := context.WithTimeout(h.ctx, h.checkTimeout)
    defer cancel()

    // 执行健康检查
    delay, err := n.Proxy.URLTest(ctx, h.testURL, nil)

    delayMs := -1
    if err == nil {
        delayMs = int(delay)
    }

    // 更新内存状态(现有逻辑)
    n.Delay = delayMs
    n.LastCheck = time.Now()
    n.Available = (delayMs > 0 && delayMs <= h.maxDelay)

    h.logger.Debug("Health check completed",
        zap.String("node", n.Name),
        zap.Int("delay", delayMs),
        zap.Bool("available", n.Available))

    // 异步持久化到数据库
    go func() {
        err := h.nodeRepo.UpdateHealthStatus(n.ID, n.Delay, n.Available)
        if err != nil {
            h.logger.Debug("Failed to persist health status",
                zap.String("node", n.ID),
                zap.Error(err))
        }
    }()
}
```

### 5.4 动态配置系统

**新文件**: `internal/config/dynamic.go`

```go
package config

import (
    "fmt"
    "proxyPool/internal/database"
    "sync"
    "time"
)

// DynamicConfig 包装 Config 并提供运行时重载能力
type DynamicConfig struct {
    current    *Config
    db         *database.DB
    configRepo *database.ConfigRepository
    mu         sync.RWMutex

    // 配置变更回调
    onProxyConfigChange       func(*ProxyConfig)
    onHealthCheckConfigChange func(*HealthCheckConfig)
    onSubscriptionConfigChange func(*SubscriptionConfig)
}

func NewDynamicConfig(db *database.DB) (*DynamicConfig, error) {
    dc := &DynamicConfig{
        db:         db,
        configRepo: database.NewConfigRepository(db),
    }

    // 加载初始配置
    if err := dc.Reload(); err != nil {
        return nil, err
    }

    return dc, nil
}

// Get 获取当前配置(线程安全)
func (dc *DynamicConfig) Get() *Config {
    dc.mu.RLock()
    defer dc.mu.RUnlock()
    return dc.current
}

// Reload 从数据库重新加载配置
func (dc *DynamicConfig) Reload() error {
    dc.mu.Lock()
    defer dc.mu.Unlock()

    // 从数据库加载所有配置
    configs, err := dc.configRepo.GetAll()
    if err != nil {
        return fmt.Errorf("failed to load configs: %w", err)
    }

    // 重构 Config 结构体
    newConfig := dc.buildConfigFromMap(configs)

    // 验证配置
    if err := validate(newConfig); err != nil {
        return fmt.Errorf("invalid config: %w", err)
    }

    dc.current = newConfig
    return nil
}

// UpdateProxyConfig 更新代理配置
func (dc *DynamicConfig) UpdateProxyConfig(cfg *ProxyConfig) error {
    // 验证
    if err := validateProxyConfig(cfg); err != nil {
        return fmt.Errorf("invalid proxy config: %w", err)
    }

    // 持久化到数据库
    configs := map[string]database.ConfigValue{
        "proxy.port_range.min": {
            Value: fmt.Sprintf("%d", cfg.PortRange.Min),
            Type:  "int",
        },
        "proxy.port_range.max": {
            Value: fmt.Sprintf("%d", cfg.PortRange.Max),
            Type:  "int",
        },
        "proxy.default_ttl": {
            Value: fmt.Sprintf("%d", int(cfg.DefaultTTL.Seconds())),
            Type:  "duration",
        },
        "proxy.max_instances": {
            Value: fmt.Sprintf("%d", cfg.MaxInstances),
            Type:  "int",
        },
        "proxy.default_protocol": {
            Value: cfg.DefaultProtocol,
            Type:  "string",
        },
    }

    if err := dc.configRepo.SetBatch(configs); err != nil {
        return fmt.Errorf("failed to persist config: %w", err)
    }

    // 重新加载并触发回调
    if err := dc.Reload(); err != nil {
        return err
    }

    if dc.onProxyConfigChange != nil {
        dc.onProxyConfigChange(cfg)
    }

    return nil
}

// UpdateHealthCheckConfig 更新健康检查配置
func (dc *DynamicConfig) UpdateHealthCheckConfig(cfg *HealthCheckConfig) error {
    // 验证
    if err := validateHealthCheckConfig(cfg); err != nil {
        return fmt.Errorf("invalid health check config: %w", err)
    }

    // 持久化到数据库
    configs := map[string]database.ConfigValue{
        "health_check.enabled": {
            Value: fmt.Sprintf("%t", cfg.Enabled),
            Type:  "bool",
        },
        "health_check.interval": {
            Value: fmt.Sprintf("%d", int(cfg.Interval.Seconds())),
            Type:  "duration",
        },
        "health_check.timeout": {
            Value: fmt.Sprintf("%d", int(cfg.Timeout.Seconds())),
            Type:  "duration",
        },
        "health_check.url": {
            Value: cfg.URL,
            Type:  "string",
        },
        "health_check.max_delay": {
            Value: fmt.Sprintf("%d", cfg.MaxDelay),
            Type:  "int",
        },
    }

    if err := dc.configRepo.SetBatch(configs); err != nil {
        return fmt.Errorf("failed to persist config: %w", err)
    }

    // 重新加载并触发回调
    if err := dc.Reload(); err != nil {
        return err
    }

    if dc.onHealthCheckConfigChange != nil {
        dc.onHealthCheckConfigChange(cfg)
    }

    return nil
}

// UpdateSubscriptionConfig 更新订阅配置
func (dc *DynamicConfig) UpdateSubscriptionConfig(cfg *SubscriptionConfig) error {
    // 验证
    if err := validateSubscriptionConfig(cfg); err != nil {
        return fmt.Errorf("invalid subscription config: %w", err)
    }

    // 持久化到数据库
    configs := map[string]database.ConfigValue{
        "subscription.update_interval": {
            Value: fmt.Sprintf("%d", int(cfg.UpdateInterval.Seconds())),
            Type:  "duration",
        },
        "subscription.timeout": {
            Value: fmt.Sprintf("%d", int(cfg.Timeout.Seconds())),
            Type:  "duration",
        },
    }

    if err := dc.configRepo.SetBatch(configs); err != nil {
        return fmt.Errorf("failed to persist config: %w", err)
    }

    // 重新加载并触发回调
    if err := dc.Reload(); err != nil {
        return err
    }

    if dc.onSubscriptionConfigChange != nil {
        dc.onSubscriptionConfigChange(cfg)
    }

    return nil
}

// SetProxyConfigChangeCallback 设置代理配置变更回调
func (dc *DynamicConfig) SetProxyConfigChangeCallback(fn func(*ProxyConfig)) {
    dc.mu.Lock()
    defer dc.mu.Unlock()
    dc.onProxyConfigChange = fn
}

// SetHealthCheckConfigChangeCallback 设置健康检查配置变更回调
func (dc *DynamicConfig) SetHealthCheckConfigChangeCallback(fn func(*HealthCheckConfig)) {
    dc.mu.Lock()
    defer dc.mu.Unlock()
    dc.onHealthCheckConfigChange = fn
}

// SetSubscriptionConfigChangeCallback 设置订阅配置变更回调
func (dc *DynamicConfig) SetSubscriptionConfigChangeCallback(fn func(*SubscriptionConfig)) {
    dc.mu.Lock()
    defer dc.mu.Unlock()
    dc.onSubscriptionConfigChange = fn
}

// buildConfigFromMap 从 map 重构 Config 结构体
func (dc *DynamicConfig) buildConfigFromMap(configs map[string]string) *Config {
    // 实现配置重构逻辑
    // 从 key-value map 构建完整的 Config 结构体
    // ...
}
```
