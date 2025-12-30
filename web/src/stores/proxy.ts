import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getProxy, releaseProxy, listInstances } from '@/api/proxy'
import type { ProxyInstance, CreateProxyRequest } from '@/types/api'
import { ElMessage } from 'element-plus'

export const useProxyStore = defineStore('proxy', () => {
  const instances = ref<ProxyInstance[]>([])
  const loading = ref(false)

  async function fetchInstances() {
    loading.value = true
    try {
      const data = await listInstances()
      // 如果后端返回的是 null (空切片可能转为null)，处理为空数组
      instances.value = data || []
    } catch (error) {
      console.error(error)
    } finally {
      loading.value = false
    }
  }

  async function createInstance(req: CreateProxyRequest) {
    try {
      await getProxy(req)
      ElMessage.success('创建代理成功')
      await fetchInstances()
    } catch (error) {
      console.error(error)
      throw error
    }
  }

  async function removeInstance(id: string) {
    try {
      await releaseProxy({ instance_id: id })
      ElMessage.success('释放代理成功')
      await fetchInstances()
    } catch (error) {
      console.error(error)
    }
  }

  return {
    instances,
    loading,
    fetchInstances,
    createInstance,
    removeInstance
  }
})
