// 统一 API 响应格式
export interface ApiResponse<T> {
  code: number
  message: string
  data?: T
}

// 节点信息
export interface Node {
  id: string
  name: string
  type: string
  delay: number
  available: boolean
  last_check: string
}

// 节点池概览
export interface NodePoolStats {
  total_nodes: number
  available_nodes: number
  nodes_by_type: Record<string, number>
  avg_delay: number
}

// 代理实例
export interface ProxyInstance {
  instance_id: string
  host: string
  port: number
  protocol: string
  node_name: string
  node_delay: number
  created_at: string
  expires_at: string
  ttl: number
  remaining: number
}

// 创建代理请求参数
export interface CreateProxyRequest {
  protocol?: 'socks5' | 'http' | 'mixed'
  ttl?: number
  max_delay?: number
  node_type?: string
  region?: string
}

// 释放代理请求参数
export interface ReleaseProxyRequest {
  instance_id: string
}

// 系统统计信息
export interface SystemStats {
  total_instances: number
  max_instances: number
  available_slots: number
  ports_used: number
  ports_available: number
}

// ============ Subscription Types ============
export interface Subscription {
  id: number
  name: string
  url: string
  enabled: boolean
  created_at: string
  updated_at: string
  last_fetch_at?: string
  last_fetch_status?: string
  last_fetch_error?: string
  node_count: number
}

export interface CreateSubscriptionRequest {
  name: string
  url: string
  enabled: boolean
}

export interface UpdateSubscriptionRequest {
  name: string
  url: string
  enabled: boolean
}

export interface TestSubscriptionResponse {
  nodes: number
}

// ============ Configuration Types ============
export interface SystemConfig {
  proxy: ProxyConfig
  health_check: HealthCheckConfig
  subscription: SubscriptionConfig
  server: ServerConfig
  log: LogConfig
}

export interface ProxyConfig {
  port_range: {
    min: number
    max: number
  }
  default_ttl: number  // seconds
  max_instances: number
  default_protocol: string
}

export interface HealthCheckConfig {
  enabled: boolean
  interval: number  // seconds
  timeout: number   // seconds
  url: string
  max_delay: number
}

export interface SubscriptionConfig {
  update_interval: number  // seconds
  timeout: number          // seconds
}

export interface ServerConfig {
  host: string
  port: number
  mode: string
}

export interface LogConfig {
  level: string
  file: string
}

export interface UpdateProxyConfigRequest {
  port_range: {
    min: number
    max: number
  }
  default_ttl: number
  max_instances: number
  default_protocol: string
}

export interface UpdateHealthCheckConfigRequest {
  enabled: boolean
  interval: number
  timeout: number
  url: string
  max_delay: number
}

export interface UpdateSubscriptionConfigRequest {
  update_interval: number
  timeout: number
}
