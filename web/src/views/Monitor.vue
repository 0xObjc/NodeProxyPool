<template>
  <div class="monitor">
    <el-row :gutter="20">
      <el-col :span="16">
        <el-card header="实例数量趋势 (实时)">
          <div class="chart-container">
            <v-chart class="chart" :option="lineOption" autoresize />
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card header="节点延迟 TOP 10">
          <div class="chart-container">
            <v-chart class="chart" :option="barOption" autoresize />
          </div>
        </el-card>
      </el-col>
    </el-row>
    
    <el-row :gutter="20" class="mt-20">
       <el-col :span="24">
         <el-card header="端口使用率">
            <el-progress 
              :percentage="usagePercentage" 
              :stroke-width="15" 
              striped 
              striped-flow 
              :status="usageStatus"
            />
         </el-card>
       </el-col>
    </el-row>

    <el-row :gutter="20" class="mt-20">
      <el-col :span="24">
        <el-card header="活跃实例流">
          <el-table 
            :data="proxyStore.instances" 
            style="width: 100%"
            :row-class-name="tableRowClassName"
          >
            <el-table-column prop="instance_id" label="实例ID" width="120">
              <template #default="{ row }">
                <code>{{ row.instance_id.substring(0, 8) }}</code>
              </template>
            </el-table-column>
            <el-table-column prop="port" label="监听端口" width="100" />
            <el-table-column prop="node_name" label="目标节点" />
            <el-table-column prop="created_at" label="创建于">
              <template #default="{ row }">
                {{ new Date(row.created_at).toLocaleTimeString() }}
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default>
                <el-tag size="small" type="success" effect="dark">RUNNING</el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useStatsStore } from '@/stores/stats'
import { useProxyStore } from '@/stores/proxy'
import { useNodeStore } from '@/stores/node'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, TitleComponent } from 'echarts/components'
import VChart from 'vue-echarts'

use([CanvasRenderer, LineChart, BarChart, GridComponent, TooltipComponent, TitleComponent])

const statsStore = useStatsStore()
const proxyStore = useProxyStore()
const nodeStore = useNodeStore()

const history = ref<{time: string, count: number}[]>([])
const timer = ref<any>(null)

const lineOption = computed(() => ({
  tooltip: { trigger: 'axis' },
  xAxis: {
    type: 'category',
    data: history.value.map(h => h.time),
    boundaryGap: false
  },
  yAxis: { type: 'value', minInterval: 1 },
  series: [{
    data: history.value.map(h => h.count),
    type: 'line',
    smooth: true,
    areaStyle: { opacity: 0.2 },
    itemStyle: { color: '#409eff' }
  }]
}))

const barOption = computed(() => {
  // Top 10 fastest nodes
  const sorted = [...nodeStore.nodes]
    .filter(n => n.delay > 0)
    .sort((a, b) => a.delay - b.delay)
    .slice(0, 10)
  
  return {
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'value', name: 'ms' },
    yAxis: { type: 'category', data: sorted.map(n => n.name).reverse() },
    series: [
      {
        name: '延迟',
        type: 'bar',
        data: sorted.map(n => n.delay).reverse(),
        itemStyle: { color: '#67c23a' }
      }
    ]
  }
})

const usagePercentage = computed(() => {
  const stats = statsStore.stats
  if (!stats) return 0
  const total = stats.ports_used + stats.ports_available
  if (total === 0) return 0
  return Math.round((stats.ports_used / total) * 100)
})

const usageStatus = computed(() => {
  if (usagePercentage.value >= 90) return 'exception'
  if (usagePercentage.value >= 70) return 'warning'
  return 'success'
})

const tableRowClassName = ({ row }: { row: any }) => {
  // Highlight if created in last 10 seconds
  const created = new Date(row.created_at).getTime()
  const now = new Date().getTime()
  if (now - created < 10000) {
    return 'new-row'
  }
  return ''
}

const updateData = async () => {
  await statsStore.fetchStats()
  await proxyStore.fetchInstances()
  await nodeStore.fetchNodes() // Need nodes for bar chart
  
  const now = new Date().toLocaleTimeString()
  history.value.push({
    time: now,
    count: statsStore.stats?.total_instances || 0
  })
  
  if (history.value.length > 20) {
    history.value.shift()
  }
}

onMounted(() => {
  updateData()
  timer.value = setInterval(updateData, 3000)
})

onUnmounted(() => {
  if (timer.value) clearInterval(timer.value)
})
</script>

<style scoped>
.mt-20 { margin-top: 20px; }
.chart-container { height: 300px; }
.chart { height: 100%; }

:deep(.el-table .new-row) {
  animation: flash 2s infinite;
  background-color: #f0f9eb;
}

@keyframes flash {
  0% { background-color: #f0f9eb; }
  50% { background-color: #ffffff; }
  100% { background-color: #f0f9eb; }
}
</style>