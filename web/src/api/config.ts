import api from './index'
import type {
  SystemConfig,
  UpdateProxyConfigRequest,
  UpdateHealthCheckConfigRequest,
  UpdateSubscriptionConfigRequest
} from '@/types/api'

// Get all configuration
export function getAllConfig(): Promise<SystemConfig> {
  return api.get('/config')
}

// Get configuration by category
export function getConfigByCategory(category: 'proxy' | 'health_check' | 'subscription' | 'server' | 'log'): Promise<any> {
  return api.get(`/config/${category}`)
}

// Update proxy configuration
export function updateProxyConfig(data: UpdateProxyConfigRequest): Promise<void> {
  return api.put('/config/proxy', data)
}

// Update health check configuration
export function updateHealthCheckConfig(data: UpdateHealthCheckConfigRequest): Promise<void> {
  return api.put('/config/health-check', data)
}

// Update subscription configuration
export function updateSubscriptionConfig(data: UpdateSubscriptionConfigRequest): Promise<void> {
  return api.put('/config/subscription', data)
}

// Reload configuration from file
export function reloadConfig(): Promise<void> {
  return api.post('/config/reload')
}
