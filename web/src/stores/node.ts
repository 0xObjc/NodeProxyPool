import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getNodes, getNodePool, updateSubscription } from '@/api/node'
import type { Node, NodePoolStats } from '@/types/api'
import { ElMessage } from 'element-plus'

export const useNodeStore = defineStore('node', () => {
  const nodes = ref<Node[]>([])
  const stats = ref<NodePoolStats | null>(null)
  const loading = ref(false)

  async function fetchNodes() {
    loading.value = true
    try {
      const data = await getNodes()
      nodes.value = data || []
    } catch (error) {
      console.error(error)
    } finally {
      loading.value = false
    }
  }

  async function fetchStats() {
    try {
      const data = await getNodePool()
      stats.value = data
    } catch (error) {
      console.error(error)
    }
  }

  async function triggerUpdate() {
    try {
      await updateSubscription()
      ElMessage.success('已触发订阅更新')
    } catch (error) {
      console.error(error)
    }
  }

  return {
    nodes,
    stats,
    loading,
    fetchNodes,
    fetchStats,
    triggerUpdate
  }
})
