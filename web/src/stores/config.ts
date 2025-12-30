import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  getAllConfig,
  updateProxyConfig,
  updateHealthCheckConfig,
  updateSubscriptionConfig,
  reloadConfig
} from '@/api/config'
import type {
  SystemConfig,
  UpdateProxyConfigRequest,
  UpdateHealthCheckConfigRequest,
  UpdateSubscriptionConfigRequest
} from '@/types/api'
import { ElMessage } from 'element-plus'

export const useConfigStore = defineStore('config', () => {
  const config = ref<SystemConfig | null>(null)
  const loading = ref(false)
  const saving = ref(false)

  // Fetch all configuration
  async function fetchConfig() {
    loading.value = true
    try {
      const data = await getAllConfig()
      config.value = data
    } catch (error) {
      console.error('Failed to fetch config:', error)
      ElMessage.error('获取配置失败')
    } finally {
      loading.value = false
    }
  }

  // Update proxy configuration
  async function updateProxy(data: UpdateProxyConfigRequest) {
    saving.value = true
    try {
      await updateProxyConfig(data)
      ElMessage.success('代理配置更新成功')
      await fetchConfig()
      return true
    } catch (error) {
      console.error('Failed to update proxy config:', error)
      return false
    } finally {
      saving.value = false
    }
  }

  // Update health check configuration
  async function updateHealthCheck(data: UpdateHealthCheckConfigRequest) {
    saving.value = true
    try {
      await updateHealthCheckConfig(data)
      ElMessage.success('健康检查配置更新成功')
      await fetchConfig()
      return true
    } catch (error) {
      console.error('Failed to update health check config:', error)
      return false
    } finally {
      saving.value = false
    }
  }

  // Update subscription configuration
  async function updateSubscriptionConfig(data: UpdateSubscriptionConfigRequest) {
    saving.value = true
    try {
      await updateSubscriptionConfig(data)
      ElMessage.success('订阅配置更新成功')
      await fetchConfig()
      return true
    } catch (error) {
      console.error('Failed to update subscription config:', error)
      return false
    } finally {
      saving.value = false
    }
  }

  // Reload configuration from file
  async function reload() {
    loading.value = true
    try {
      await reloadConfig()
      ElMessage.success('配置重载成功')
      await fetchConfig()
    } catch (error) {
      console.error('Failed to reload config:', error)
    } finally {
      loading.value = false
    }
  }

  return {
    config,
    loading,
    saving,
    fetchConfig,
    updateProxy,
    updateHealthCheck,
    updateSubscriptionConfig,
    reload
  }
})
