import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getStats } from '@/api/stats'
import type { SystemStats } from '@/types/api'

export const useStatsStore = defineStore('stats', () => {
  const stats = ref<SystemStats | null>(null)

  async function fetchStats() {
    try {
      const data = await getStats()
      stats.value = data
    } catch (error) {
      console.error(error)
    }
  }

  return {
    stats,
    fetchStats
  }
})
