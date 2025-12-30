# SQLite 集成方案 - 第二部分

## 6. API 端点设计

### 6.1 订阅源管理 API

**新文件**: `internal/api/subscription_handler.go`

#### 数据结构

```go
// 订阅源响应结构
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

// 创建订阅源请求
type CreateSubscriptionRequest struct {
    Name    string `json:"name" binding:"required"`
    URL     string `json:"url" binding:"required,url"`
    Enabled bool   `json:"enabled"`
}

// 更新订阅源请求
type UpdateSubscriptionRequest struct {
    Name    *string `json:"name"`
    URL     *string `json:"url" binding:"omitempty,url"`
    Enabled *bool   `json:"enabled"`
}
```

#### API 端点

**1. GET /api/subscriptions - 列出所有订阅源**

响应示例:
```json
{
    "code": 0,
    "message": "success",
    "data": [
        {
            "id": 1,
            "name": "主订阅",
            "url": "https://example.com/subscribe",
            "enabled": true,
            "created_at": "2024-01-01T00:00:00Z",
            "updated_at": "2024-01-01T00:00:00Z",
            "last_fetch_at": "2024-01-01T12:00:00Z",
            "last_fetch_status": "success",
            "node_count": 150
        }
    ]
}
```

**2. POST /api/subscriptions - 创建订阅源**

请求示例:
```json
{
    "name": "新订阅",
    "url": "https://example.com/subscribe",
    "enabled": true
}
```

响应示例:
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": 2,
        "name": "新订阅",
        "url": "https://example.com/subscribe",
        "enabled": true,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z",
        "node_count": 0
    }
}
```

**3. GET /api/subscriptions/:id - 获取单个订阅源**

响应格式同上。

**4. PUT /api/subscriptions/:id - 更新订阅源**

请求示例:
```json
{
    "enabled": false
}
```

响应示例:
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "id": 1,
        "name": "主订阅",
        "url": "https://example.com/subscribe",
        "enabled": false,
        "updated_at": "2024-01-01T13:00:00Z"
    }
}
```

**5. DELETE /api/subscriptions/:id - 删除订阅源**

响应示例:
```json
{
    "code": 0,
    "message": "success"
}
```

**6. POST /api/subscriptions/:id/test - 测试订阅源**

响应示例:
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "node_count": 150,
        "error": ""
    }
}
```

### 6.2 动态配置 API

**新文件**: `internal/api/config_handler.go`

#### 数据结构

```go
// 代理配置更新请求
type UpdateProxyConfigRequest struct {
    PortRangeMin    *int    `json:"port_range_min" binding:"omitempty,min=1024,max=65535"`
    PortRangeMax    *int    `json:"port_range_max" binding:"omitempty,min=1024,max=65535"`
    DefaultTTL      *int    `json:"default_ttl" binding:"omitempty,min=60"` // 秒
    MaxInstances    *int    `json:"max_instances" binding:"omitempty,min=1"`
    DefaultProtocol *string `json:"default_protocol" binding:"omitempty,oneof=socks5 http mixed"`
}

// 健康检查配置更新请求
type UpdateHealthCheckConfigRequest struct {
    Enabled  *bool   `json:"enabled"`
    Interval *int    `json:"interval" binding:"omitempty,min=60"`    // 秒
    Timeout  *int    `json:"timeout" binding:"omitempty,min=1"`      // 秒
    URL      *string `json:"url" binding:"omitempty,url"`
    MaxDelay *int    `json:"max_delay" binding:"omitempty,min=0"`    // 毫秒
}

// 订阅配置更新请求
type UpdateSubscriptionConfigRequest struct {
    UpdateInterval *int `json:"update_interval" binding:"omitempty,min=300"` // 秒
    Timeout        *int `json:"timeout" binding:"omitempty,min=5"`           // 秒
}
```

#### API 端点

**1. GET /api/config - 获取所有配置**

响应示例:
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "proxy": {
            "port_range_min": 20000,
            "port_range_max": 30000,
            "default_ttl": 1800,
            "max_instances": 100,
            "default_protocol": "socks5"
        },
        "health_check": {
            "enabled": true,
            "interval": 300,
            "timeout": 10,
            "url": "https://www.google.com/generate_204",
            "max_delay": 1000
        },
        "subscription": {
            "update_interval": 21600,
            "timeout": 30
        }
    }
}
```

**2. GET /api/config/:category - 按分类获取配置**

分类: `proxy`, `health_check`, `subscription`, `server`, `log`

响应示例:
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "enabled": true,
        "interval": 300,
        "timeout": 10,
        "url": "https://www.google.com/generate_204",
        "max_delay": 1000
    }
}
```

**3. PUT /api/config/proxy - 更新代理配置**

请求示例:
```json
{
    "max_instances": 200,
    "default_ttl": 3600
}
```

响应示例:
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "port_range_min": 20000,
        "port_range_max": 30000,
        "default_ttl": 3600,
        "max_instances": 200,
        "default_protocol": "socks5"
    }
}
```

**4. PUT /api/config/health-check - 更新健康检查配置**

请求示例:
```json
{
    "interval": 600,
    "max_delay": 2000
}
```

响应示例:
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "enabled": true,
        "interval": 600,
        "timeout": 10,
        "url": "https://www.google.com/generate_204",
        "max_delay": 2000
    }
}
```

**5. PUT /api/config/subscription - 更新订阅配置**

请求示例:
```json
{
    "update_interval": 43200
}
```

响应示例:
```json
{
    "code": 0,
    "message": "success",
    "data": {
        "update_interval": 43200,
        "timeout": 30
    }
}
```

**6. POST /api/config/reload - 重新加载配置**

触发所有组件重新加载配置。

响应示例:
```json
{
    "code": 0,
    "message": "Configuration reloaded successfully"
}
```

### 6.3 路由注册

**修改文件**: `internal/api/handler.go`

```go
func (h *Handler) RegisterRoutes(r *gin.Engine) {
    api := r.Group("/api")
    {
        // 现有路由
        api.GET("/health", h.Health)
        api.GET("/nodePool", h.GetNodePool)
        api.GET("/nodes", h.ListNodes)
        api.POST("/getProxy", h.GetProxy)
        api.POST("/releaseProxy", h.ReleaseProxy)
        api.GET("/getInstance/:id", h.GetInstance)
        api.GET("/listInstances", h.ListInstances)
        api.GET("/stats", h.GetStats)
        api.POST("/subscription/update", h.UpdateSubscription)

        // 新增:订阅源管理
        api.GET("/subscriptions", h.ListSubscriptions)
        api.POST("/subscriptions", h.CreateSubscription)
        api.GET("/subscriptions/:id", h.GetSubscription)
        api.PUT("/subscriptions/:id", h.UpdateSubscription)
        api.DELETE("/subscriptions/:id", h.DeleteSubscription)
        api.POST("/subscriptions/:id/test", h.TestSubscription)

        // 新增:动态配置管理
        api.GET("/config", h.GetAllConfig)
        api.GET("/config/:category", h.GetConfigByCategory)
        api.PUT("/config/proxy", h.UpdateProxyConfig)
        api.PUT("/config/health-check", h.UpdateHealthCheckConfig)
        api.PUT("/config/subscription", h.UpdateSubscriptionConfig)
        api.POST("/config/reload", h.ReloadConfig)
    }
}
```

---

## 7. 实施步骤

### Phase 1: 数据库基础层(优先级:最高)

**目标**: 建立数据库连接和 Schema

**任务**:
1. 创建 `internal/database/` 目录
2. 实现 `db.go`:
   - 数据库连接管理
   - WAL 模式配置
   - 连接池设置
3. 实现 `migrations.go`:
   - Schema 创建(所有表和索引)
   - 版本管理
   - `SeedFromConfig()` 方法
4. 实现 `models.go`:
   - 定义所有数据库模型结构体
5. 添加依赖:
   ```bash
   go get github.com/mattn/go-sqlite3
   ```

**验证**:
- 数据库文件成功创建
- 所有表和索引正确创建
- 从 config.yaml 导入数据成功

### Phase 2: Repository 层(优先级:高)

**目标**: 实现数据访问层

**任务**:
1. 实现 `subscription_repo.go`:
   - Create, Update, Delete, GetByID, GetAll, GetEnabled
   - UpdateFetchStatus
2. 实现 `node_repo.go`:
   - Upsert, UpsertBatch, GetAll, GetByID
   - UpdateHealthStatus, UpdateUsageStats
   - DeleteBySubscription, DeleteStale
3. 实现 `config_repo.go`:
   - Get, GetByCategory, GetAll
   - Set, SetBatch
   - InitializeDefaults

**验证**:
- 单元测试覆盖所有 CRUD 操作
- 批量操作性能测试
- 并发访问测试

### Phase 3: 订阅源管理(优先级:高)

**目标**: 订阅源从数据库加载和 API 管理

**任务**:
1. 修改 `internal/subscription/manager.go`:
   - 添加 `db` 和 `subRepo` 字段
   - 修改 `UpdateAll()` 从数据库加载订阅源
   - 添加 `ReloadSources()` 方法
2. 创建 `internal/api/subscription_handler.go`:
   - 实现所有订阅源 CRUD 端点
   - 实现测试订阅端点
3. 修改 `cmd/server/main.go`:
   - 初始化数据库
   - 传递数据库到 subscription.Manager
4. 修改 `internal/api/handler.go`:
   - 注册订阅源管理路由

**验证**:
- 订阅源从数据库正确加载
- API 端点功能正常
- 添加/删除订阅源后立即生效

### Phase 4: 节点持久化(优先级:中)

**目标**: 节点数据持久化到数据库

**任务**:
1. 修改 `internal/node/pool.go`:
   - 添加 `db` 和 `nodeRepo` 字段
   - 实现 `LoadFromDatabase()` 方法
   - 修改 `UpdateNodes()` 异步持久化
   - 实现 `convertFromDBNode()` 辅助方法
2. 修改 `internal/node/health_checker.go`:
   - 添加 `nodeRepo` 字段
   - 修改 `checkNode()` 异步持久化健康状态
3. 修改 `cmd/server/main.go`:
   - 传递数据库到 node.Pool 和 health_checker
   - 启动时调用 `pool.LoadFromDatabase()`

**验证**:
- 节点在重启后正确恢复
- 健康检查结果持久化
- 性能无明显下降

### Phase 5: 动态配置系统(优先级:中)

**目标**: 配置运行时动态修改

**任务**:
1. 创建 `internal/config/dynamic.go`:
   - 实现 `DynamicConfig` 结构体
   - 实现 `Reload()` 方法
   - 实现 `UpdateProxyConfig()`, `UpdateHealthCheckConfig()`, `UpdateSubscriptionConfig()`
   - 实现配置变更回调机制
2. 创建 `internal/api/config_handler.go`:
   - 实现所有配置更新端点
   - 实现配置重载端点
3. 修改 `cmd/server/main.go`:
   - 使用 `DynamicConfig` 替代静态配置
   - 设置配置变更回调
4. 修改 `internal/api/handler.go`:
   - 注册配置管理路由

**验证**:
- 配置更新通过 API 成功
- 组件接收到配置变更通知
- 配置在重启后保持

### Phase 6: 集成测试(优先级:高)

**目标**: 端到端测试

**任务**:
1. 测试完整启动流程:
   - 首次启动(数据库初始化)
   - 后续启动(从数据库加载)
2. 测试订阅源管理:
   - CRUD 操作
   - 订阅源测试
   - 节点更新
3. 测试动态配置:
   - 配置更新
   - 组件响应
   - 配置持久化
4. 性能测试:
   - 数据库操作延迟
   - 批量操作性能
   - 并发访问性能

**验证**:
- 所有功能正常工作
- 性能满足要求
- 无数据丢失或损坏

### Phase 7: 文档和清理(优先级:中)

**目标**: 完善文档和代码清理

**任务**:
1. 更新 `CLAUDE.md`:
   - 添加数据库架构说明
   - 添加 API 端点文档
   - 更新数据流说明
2. 创建用户迁移指南
3. 代码审查和清理:
   - 移除未使用的代码
   - 统一错误处理
   - 添加必要的注释
4. 性能优化:
   - 调整连接池参数
   - 优化批量操作
   - 添加必要的索引

**验证**:
- 文档完整准确
- 代码质量良好
- 性能达到预期

---

## 8. 性能优化建议

### 8.1 数据库层面

**连接池配置**:
```go
conn.SetMaxOpenConns(25)  // 根据并发量调整
conn.SetMaxIdleConns(5)   // 保持少量空闲连接
conn.SetConnMaxLifetime(time.Hour)
```

**批量操作**:
```go
// 节点批量插入使用事务
func (r *NodeRepository) UpsertBatch(nodes []*Node) error {
    tx, err := r.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    stmt, err := tx.Prepare(`
        INSERT OR REPLACE INTO nodes
        (id, name, type, server, port, raw_config_json, ...)
        VALUES (?, ?, ?, ?, ?, ?, ...)
    `)
    if err != nil {
        return err
    }
    defer stmt.Close()

    for _, node := range nodes {
        _, err := stmt.Exec(node.ID, node.Name, ...)
        if err != nil {
            return err
        }
    }

    return tx.Commit()
}
```

**索引优化**:
- `nodes(type, available)` - 用于节点筛选
- `nodes(delay)` - 用于延迟排序
- `subscription_sources(enabled)` - 用于加载启用的订阅源

### 8.2 应用层面

**异步写入**:
```go
// 健康检查结果异步持久化
go func() {
    err := h.nodeRepo.UpdateHealthStatus(n.ID, n.Delay, n.Available)
    if err != nil {
        h.logger.Debug("Failed to persist health status", zap.Error(err))
    }
}()
```

**批量更新缓冲**:
```go
// 收集多个健康检查结果,批量更新
type HealthStatusBuffer struct {
    updates []HealthUpdate
    mu      sync.Mutex
    ticker  *time.Ticker
}

func (b *HealthStatusBuffer) Add(update HealthUpdate) {
    b.mu.Lock()
    b.updates = append(b.updates, update)
    b.mu.Unlock()
}

func (b *HealthStatusBuffer) Flush() {
    b.mu.Lock()
    updates := b.updates
    b.updates = nil
    b.mu.Unlock()

    if len(updates) > 0 {
        b.repo.UpdateHealthStatusBatch(updates)
    }
}
```

**内存优先策略**:
- 节点池保持在内存中,数据库仅作持久化
- 读操作优先从内存获取
- 写操作异步持久化到数据库

### 8.3 监控指标

建议添加以下监控指标:

```go
// 数据库操作延迟
dbOperationDuration := prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name: "db_operation_duration_seconds",
        Help: "Database operation duration",
    },
    []string{"operation", "table"},
)

// 数据库连接池状态
dbPoolStats := prometheus.NewGaugeVec(
    prometheus.GaugeOpts{
        Name: "db_pool_connections",
        Help: "Database connection pool stats",
    },
    []string{"state"}, // open, idle, in_use
)
```

---

## 9. 错误处理和恢复

### 9.1 数据库初始化失败

```go
db, err := database.New("./data/proxypool.db")
if err != nil {
    logger.Fatal("Failed to initialize database", zap.Error(err))
    // 无法继续运行,退出程序
}
```

### 9.2 运行时数据库错误

**读取失败**:
```go
sources, err := m.subRepo.GetEnabled()
if err != nil {
    logger.Error("Failed to load subscription sources", zap.Error(err))
    // 继续使用内存中的数据
    return
}
```

**写入失败**:
```go
go func() {
    err := h.nodeRepo.UpdateHealthStatus(n.ID, n.Delay, n.Available)
    if err != nil {
        // 记录错误,但不影响主流程
        h.logger.Debug("Failed to persist health status", zap.Error(err))

        // 可选:重试机制
        for i := 0; i < 3; i++ {
            time.Sleep(time.Second * time.Duration(i+1))
            if err := h.nodeRepo.UpdateHealthStatus(n.ID, n.Delay, n.Available); err == nil {
                break
            }
        }
    }
}()
```

### 9.3 配置验证失败

```go
func (dc *DynamicConfig) UpdateProxyConfig(cfg *ProxyConfig) error {
    // 验证配置
    if err := validateProxyConfig(cfg); err != nil {
        return fmt.Errorf("invalid proxy config: %w", err)
    }

    // 持久化前再次验证
    if cfg.PortRange.Min >= cfg.PortRange.Max {
        return errors.New("port_range_min must be less than port_range_max")
    }

    // 持久化到数据库
    if err := dc.configRepo.SetBatch(configs); err != nil {
        return fmt.Errorf("failed to persist config: %w", err)
    }

    // 重新加载
    if err := dc.Reload(); err != nil {
        // 回滚?或者保持旧配置
        return fmt.Errorf("failed to reload config: %w", err)
    }

    return nil
}
```

---

## 10. 用户迁移指南

### 10.1 首次升级步骤

1. **备份现有配置**:
   ```bash
   cp configs/config.yaml configs/config.yaml.backup
   ```

2. **停止旧版本服务**:
   ```bash
   # 如果服务正在运行
   pkill -f "./server"
   ```

3. **更新代码**:
   ```bash
   git pull
   go mod tidy
   go build ./cmd/server
   ```

4. **首次启动**(自动初始化数据库):
   ```bash
   ./server -c configs/config.yaml
   ```

5. **查看日志确认**:
   ```
   INFO: First run detected, seeding from config.yaml
   INFO: Database initialized successfully
   INFO: Loaded 150 nodes from database
   INFO: Server started on 0.0.0.0:8080
   ```

### 10.2 日常使用

**查看订阅源**:
```bash
curl http://localhost:8080/api/subscriptions
```

**添加订阅源**:
```bash
curl -X POST http://localhost:8080/api/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "name": "新订阅",
    "url": "https://example.com/subscribe",
    "enabled": true
  }'
```

**禁用订阅源**:
```bash
curl -X PUT http://localhost:8080/api/subscriptions/1 \
  -H "Content-Type: application/json" \
  -d '{"enabled": false}'
```

**更新健康检查配置**:
```bash
curl -X PUT http://localhost:8080/api/config/health-check \
  -H "Content-Type: application/json" \
  -d '{
    "interval": 600,
    "max_delay": 2000
  }'
```

**查看当前配置**:
```bash
curl http://localhost:8080/api/config
```

### 10.3 故障排查

**问题**: 数据库文件损坏

**解决**:
```bash
# 删除数据库文件
rm ./data/proxypool.db

# 重新启动(会重新初始化)
./server -c configs/config.yaml
```

**问题**: 配置更新不生效

**解决**:
```bash
# 手动触发配置重载
curl -X POST http://localhost:8080/api/config/reload
```

**问题**: 节点数据丢失

**解决**:
```bash
# 手动触发订阅更新
curl -X POST http://localhost:8080/api/subscription/update
```

---

## 11. 总结

### 11.1 核心优势

1. **灵活性**: 订阅源和配置可通过 API 动态管理,无需重启
2. **持久化**: 节点数据和配置持久化,重启后快速恢复
3. **性能**: 混合存储架构,内存优先,数据库辅助
4. **可维护性**: 清晰的分层架构,易于扩展和维护

### 11.2 技术亮点

- SQLite WAL 模式提供良好并发性能
- 异步写入避免阻塞主流程
- Repository 模式提供清晰的数据访问层
- 动态配置系统支持运行时修改

### 11.3 后续扩展方向

1. **Web 管理界面**: 基于 API 构建前端管理界面
2. **统计分析**: 利用数据库数据进行节点使用分析
3. **多用户支持**: 添加用户认证和权限管理
4. **配置模板**: 支持配置预设和快速切换
5. **数据导出**: 支持节点数据和配置导出/导入

### 11.4 注意事项

- 首次启动后,config.yaml 不再被读取
- 所有配置更改通过 API 进行
- 数据库文件位于 `./data/proxypool.db`
- 定期备份数据库文件
- 监控数据库文件大小,必要时清理历史数据
