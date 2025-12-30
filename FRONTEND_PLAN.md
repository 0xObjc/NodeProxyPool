# ProxyPool 前端实现计划

## 概述

为 ProxyPool 项目添加 Vue 3 + Element Plus 前端管理界面，采用嵌入式部署方案（使用 Go embed），实现仪表盘、代理管理、节点管理和实时监控四大核心功能。

## 技术选型

- **前端框架**: Vue 3 (Composition API) + TypeScript
- **UI 组件库**: Element Plus
- **构建工具**: Vite
- **状态管理**: Pinia
- **路由**: Vue Router 4 (HTML5 History 模式)
- **图表库**: ECharts
- **部署方式**: Go embed 嵌入式部署
- **认证**: 无需认证（内网使用）

## 项目结构

```
proxyPool/
├── web/                          # 前端项目根目录（新建）
│   ├── src/
│   │   ├── api/                  # API 封装
│   │   │   ├── index.ts          # axios 配置
│   │   │   ├── proxy.ts          # 代理 API
│   │   │   ├── node.ts           # 节点 API
│   │   │   └── stats.ts          # 统计 API
│   │   ├── components/           # 公共组件
│   │   │   ├── CountdownTimer.vue # TTL 倒计时
│   │   │   └── StatsCard.vue     # 统计卡片
│   │   ├── layouts/
│   │   │   └── MainLayout.vue    # 主布局（侧边栏+顶栏）
│   │   ├── router/
│   │   │   └── index.ts          # 路由配置
│   │   ├── stores/               # Pinia stores
│   │   │   ├── proxy.ts
│   │   │   ├── node.ts
│   │   │   └── stats.ts
│   │   ├── types/                # TypeScript 类型
│   │   │   ├── proxy.ts
│   │   │   ├── node.ts
│   │   │   └── api.ts
│   │   ├── views/                # 页面组件
│   │   │   ├── Dashboard.vue     # 仪表盘
│   │   │   ├── ProxyManage.vue   # 代理管理
│   │   │   ├── NodeManage.vue    # 节点管理
│   │   │   └── Monitor.vue       # 实时监控
│   │   ├── App.vue
│   │   └── main.ts
│   ├── .env.development          # 开发环境配置
│   ├── .env.production           # 生产环境配置
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
├── cmd/server/main.go            # 需要修改：添加 embed 和静态文件服务
└── go.mod                        # 无需修改
```

## 实现步骤

### 第一阶段：前端项目初始化

1. **创建前端项目**
   - 在项目根目录创建 `web/` 目录
   - 初始化 Vite + Vue 3 + TypeScript 项目
   - 安装依赖：element-plus, vue-router, pinia, axios, echarts, vue-echarts

2. **配置 Vite**
   - 配置路径别名 `@` 指向 `src/`
   - 配置开发服务器代理：`/api` -> `http://localhost:8080`
   - 配置构建输出目录为 `dist/`
   - 配置代码分割（element-plus 和 echarts 单独打包）

3. **配置环境变量**
   - `.env.development`: `VITE_API_BASE_URL=/api`
   - `.env.production`: `VITE_API_BASE_URL=/api`

### 第二阶段：基础架构搭建

4. **API 封装层**
   - `api/index.ts`: 创建 axios 实例，配置拦截器、错误处理
   - `api/proxy.ts`: 封装代理相关接口（getProxy, releaseProxy, listInstances, getInstance）
   - `api/node.ts`: 封装节点相关接口（getNodes, getNodePool, updateSubscription）
   - `api/stats.ts`: 封装统计接口（getStats）

5. **TypeScript 类型定义**
   - `types/api.ts`: 定义统一的 API 响应格式 `ApiResponse<T>`
   - `types/proxy.ts`: 定义 `ProxyInstance`, `CreateProxyRequest` 等类型
   - `types/node.ts`: 定义 `Node`, `NodePool`, `FilterOptions` 等类型

6. **Pinia 状态管理**
   - `stores/proxy.ts`: 代理实例状态管理（列表、创建、释放）
   - `stores/node.ts`: 节点状态管理（列表、过滤、更新订阅）
   - `stores/stats.ts`: 统计数据状态管理

7. **路由配置**
   - 配置 4 个主要路由：`/dashboard`, `/proxy`, `/nodes`, `/monitor`
   - 使用 MainLayout 作为父路由
   - 配置 HTML5 History 模式

8. **主布局组件**
   - 使用 Element Plus 的 `el-container` 布局
   - 左侧导航栏：4 个菜单项（仪表盘、代理管理、节点管理、实时监控）
   - 顶部栏：显示系统标题和刷新按钮
   - 主内容区：`<router-view>` 渲染子路由

### 第三阶段：核心功能页面开发

9. **仪表盘页面 (Dashboard.vue)**
   - 顶部统计卡片：活跃实例数、可用节点数、平均延迟、端口使用率
   - 节点类型分布饼图（ECharts）
   - 实例使用率进度条
   - 最近创建的实例列表（最多显示 5 条）
   - 数据刷新：每 5 秒自动轮询

10. **代理管理页面 (ProxyManage.vue)**
    - 创建代理表单：
      - 协议类型选择（socks5/http/mixed）
      - TTL 设置（60-3600 秒）
      - 最大延迟过滤（0=不限制）
      - 节点类型过滤（vmess/trojan/ss）
      - 地区过滤（关键字匹配）
    - 活跃代理列表表格：
      - 显示实例 ID、代理地址、节点名称、延迟、剩余时间
      - 操作按钮：复制地址、手动释放
      - TTL 倒计时组件（每秒更新，颜色根据剩余时间变化）
    - 数据刷新：每 3 秒自动轮询

11. **节点管理页面 (NodeManage.vue)**
    - 顶部操作栏：
      - 手动更新订阅按钮
      - 节点过滤器（类型、状态、关键字搜索）
    - 节点健康状态饼图（可用/不可用）
    - 节点列表表格：
      - 显示节点名称、类型、延迟、状态、最后检测时间
      - 延迟列使用颜色标签（<200ms 绿色，200-500ms 黄色，>500ms 红色）
      - 支持按延迟排序
    - 数据刷新：每 10 秒自动轮询

12. **实时监控页面 (Monitor.vue)**
    - 实例数量趋势折线图（保留最近 20 个数据点）
    - 节点延迟分布柱状图（显示前 10 个最快节点）
    - 端口使用率仪表盘（ECharts Gauge）
    - 活跃实例实时列表（自动高亮新创建的实例）
    - 数据刷新：每 3 秒自动轮询

13. **公共组件开发**
    - `CountdownTimer.vue`: TTL 倒计时组件
      - 接收 `expiresAt` 时间戳
      - 每秒更新剩余时间
      - 根据剩余时间显示不同颜色（<60s 红色，<300s 橙色，其他绿色）
    - `StatsCard.vue`: 统计卡片组件
      - 显示标题、数值、图标
      - 支持自定义颜色主题

### 第四阶段：Go 后端改造

14. **修改 cmd/server/main.go**
    - 添加 `embed` 导入和 `//go:embed ../../web/dist` 指令
    - 创建 `setupStaticFiles()` 函数：
      - 使用 `fs.Sub()` 获取 `web/dist` 子目录
      - 注册 `/assets/*filepath` 路由处理静态资源
      - 配置 `NoRoute` 处理器返回 `index.html`（支持 SPA 路由）
      - API 404 特殊处理（返回 JSON 格式错误）
    - 创建 `corsMiddleware()` 函数（仅开发环境启用）：
      - 允许 `http://localhost:5173` 跨域
      - 支持 OPTIONS 预检请求
    - 在 main 函数中调用：
      - 先注册 API 路由（优先级高）
      - 再调用 `setupStaticFiles()`（优先级低）

15. **路由优先级确保**
    - API 路由 `/api/*` 优先匹配
    - 静态资源 `/assets/*` 次之
    - SPA 路由通过 `NoRoute` 兜底返回 `index.html`

### 第五阶段：构建和测试

16. **前端构建**
    - 执行 `cd web && npm run build`
    - 验证 `web/dist/` 目录生成
    - 检查 `dist/index.html` 和 `dist/assets/` 存在

17. **Go 重新编译**
    - 执行 `go build ./cmd/server`
    - embed 会自动将 `web/dist/` 打包进二进制文件

18. **集成测试**
    - 启动服务：`./server -c configs/config.yaml`
    - 访问 `http://localhost:8080/` 验证前端加载
    - 测试所有页面路由是否正常
    - 测试 API 调用是否正常
    - 测试前端路由刷新是否正常（SPA History 模式）

19. **开发环境测试**
    - 启动后端：`go run cmd/server/main.go -c configs/config.yaml`
    - 启动前端：`cd web && npm run dev`
    - 访问 `http://localhost:5173`
    - 验证 Vite 代理是否正常工作
    - 验证 CORS 是否正确配置

## 关键技术细节

### API 响应格式统一处理

所有 API 返回格式：
```typescript
interface ApiResponse<T> {
  code: number      // 0=成功, 400=请求错误, 404=未找到, 500=服务器错误
  message: string
  data?: T
}
```

### 详细 API 文档

#### 1. 代理管理 (Proxy Management)

- **创建代理实例**
  - **接口**: `POST /api/getProxy`
  - **请求体**:
    ```json
    {
      "protocol": "socks5", // socks5/http/mixed (可选, 默认 socks5)
      "ttl": 300,           // 存活时间(秒) (可选, 默认 3600)
      "max_delay": 500,     // 最大延迟(ms) (可选, 0 为不限制)
      "node_type": "vmess", // 节点类型: vmess/trojan/ss (可选)
      "region": "香港"      // 地区关键字过滤 (可选)
    }
    ```
  - **响应数据 (data)**:
    ```json
    {
      "instance_id": "uuid",
      "host": "127.0.0.1",
      "port": 18081,
      "protocol": "socks5",
      "node_name": "香港 01",
      "node_delay": 120,
      "created_at": "2023-10-27T10:00:00Z",
      "expires_at": "2023-10-27T10:05:00Z",
      "ttl": 300,
      "remaining": 299
    }
    ```

- **释放代理实例**
  - **接口**: `POST /api/releaseProxy`
  - **请求体**:
    ```json
    {
      "instance_id": "uuid"
    }
    ```

- **列出所有活跃实例**
  - **接口**: `GET /api/listInstances`
  - **响应数据 (data)**: `ProxyInstance[]`

- **获取单个实例详情**
  - **接口**: `GET /api/getInstance/:id`

#### 2. 节点管理 (Node Management)

- **获取节点池概览**
  - **接口**: `GET /api/nodePool`
  - **响应数据 (data)**:
    ```json
    {
      "total_nodes": 100,
      "available_nodes": 85,
      "nodes_by_type": {
        "vmess": 50,
        "trojan": 30,
        "ss": 20
      },
      "avg_delay": 150
    }
    ```

- **获取节点列表**
  - **接口**: `GET /api/nodes`
  - **响应数据 (data)**:
    ```json
    [
      {
        "id": "node-hash",
        "name": "节点名称",
        "type": "vmess",
        "delay": 120,
        "available": true,
        "last_check": "2023-10-27 10:00:00"
      }
    ]
    ```

- **触发订阅更新**
  - **接口**: `POST /api/subscription/update`
  - **说明**: 异步操作，立即返回。

#### 3. 系统统计 (System Stats)

- **获取运行统计**
  - **接口**: `GET /api/stats`
  - **响应数据 (data)**:
    ```json
    {
      "total_instances": 5,      // 当前活跃实例数
      "max_instances": 100,     // 最大实例限制
      "available_slots": 95,    // 剩余可用名额
      "ports_used": 5,          // 已占用的端口数
      "ports_available": 995    // 剩余可用端口数
    }
    ```

- **健康检查**
  - **接口**: `GET /health` (不带 /api 前缀)
  - **响应**: `{"status": "ok", "message": "..."}`

axios 拦截器统一处理：
- 成功响应：返回 `response.data.data`
- 错误响应：抛出 `response.data.message` 或默认错误信息

### 轮询策略

所有页面离开时必须清除定时器，避免内存泄漏：
```typescript
onMounted(() => {
  fetchData()
  timer.value = setInterval(fetchData, interval)
})

onUnmounted(() => {
  if (timer.value) clearInterval(timer.value)
})
```

### SPA 路由支持

Go 后端 `NoRoute` 处理器必须：
1. 检查路径是否以 `/api/` 开头，是则返回 404 JSON
2. 否则返回 `index.html`，让前端路由接管

### 性能优化

- 使用 `<keep-alive>` 缓存路由组件
- ECharts 图表使用 `shallowRef` 减少响应式开销
- 代码分割：element-plus 和 echarts 单独打包
- 限制历史数据点数量（如监控页面最多保留 50 个数据点）

## 需要修改的文件

### 新增文件（约 30+ 个）
- `web/` 整个目录（前端项目）

### 修改文件
- `/Users/peter/Documents/工作/业务/proxyPool/cmd/server/main.go`
  - 添加 embed 导入和指令
  - 添加 `setupStaticFiles()` 函数
  - 添加 `corsMiddleware()` 函数
  - 修改 main 函数调用顺序

### 无需修改
- 所有 `internal/` 目录下的文件（API 已完整）
- `configs/config.yaml`（配置已满足需求）
- `go.mod`（无需新增依赖）

## 开发和部署流程

### 开发环境
```bash
# 终端 1：启动后端
go run cmd/server/main.go -c configs/config.yaml

# 终端 2：启动前端
cd web
npm install
npm run dev

# 访问 http://localhost:5173
```

### 生产环境
```bash
# 1. 构建前端
cd web
npm run build

# 2. 构建 Go（自动嵌入前端）
cd ..
go build ./cmd/server

# 3. 运行
./server -c configs/config.yaml

# 访问 http://localhost:8080
```

## 预期效果

- 单一二进制文件包含前后端
- 无需单独部署前端服务器
- 无需配置 nginx 反向代理
- 开发环境支持热重载
- 生产环境零配置部署