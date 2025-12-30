<template>
  <div class="dashboard">
    <!-- 顶部统计卡片 -->
    <el-row :gutter="20">
      <el-col :span="6">
        <StatsCard
          title="活跃代理实例"
          :value="statsStore.stats?.total_instances || 0"
          icon="Connection"
          bg-color="#ecf5ff"
          icon-color="#409eff"
        />
      </el-col>
      <el-col :span="6">
        <StatsCard
          title="可用节点数"
          :value="nodeStore.stats?.available_nodes || 0"
          icon="Share"
          bg-color="#f0f9eb"
          icon-color="#67c23a"
        />
      </el-col>
      <el-col :span="6">
        <StatsCard
          title="平均延迟"
          :value="(nodeStore.stats?.avg_delay || 0) + 'ms'"
          icon="Timer"
          bg-color="#fdf6ec"
          icon-color="#e6a23c"
        />
      </el-col>
      <el-col :span="6">
        <StatsCard
          title="端口使用率"
          :value="usageRate + '%'"
          icon="PieChart"
          bg-color="#fef0f0"
          icon-color="#f56c6c"
        />
      </el-col>
    </el-row>

    <el-row :gutter="20" class="mt-20">
      <el-col :span="24">
        <el-card header="实例使用率 (容量)">
          <el-progress
            :text-inside="true"
            :stroke-width="20"
            :percentage="instanceUsagePercentage"
            :status="instanceUsageStatus"
          />
          <div style="margin-top: 5px; font-size: 12px; color: #909399; text-align: right;">
             {{ statsStore.stats?.total_instances || 0 }} / {{ statsStore.stats?.max_instances || 0 }} 实例
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="mt-20">
      <!-- 节点分布饼图 -->
      <el-col :span="10">
        <el-card header="节点类型分布">
          <div class="chart-container">
            <v-chart class="chart" :option="pieOption" autoresize />
          </div>
        </el-card>
      </el-col>
      
      <!-- 最近实例列表 -->
      <el-col :span="14">
        <el-card header="最近创建的实例">
          <el-table :data="recentInstances" style="width: 100%" v-loading="proxyStore.loading">
            <el-table-column prop="port" label="端口" width="100">
              <template #default="{ row }">
                <el-tag size="small">{{ row.port }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="node_name" label="节点" show-overflow-tooltip />
            <el-table-column prop="node_delay" label="延迟" width="100">
              <template #default="{ row }">
                <span :class="getDelayClass(row.node_delay)">{{ row.node_delay }}ms</span>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="180">
              <template #default="{ row }">
                {{ formatTime(row.created_at) }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, computed, ref } from 'vue'
import { useStatsStore } from '@/stores/stats'
import { useNodeStore } from '@/stores/node'
import { useProxyStore } from '@/stores/proxy'
import StatsCard from '@/components/StatsCard.vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { PieChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import VChart from 'vue-echarts'

use([CanvasRenderer, PieChart, TitleComponent, TooltipComponent, LegendComponent])

const statsStore = useStatsStore()
const nodeStore = useNodeStore()
const proxyStore = useProxyStore()

const timer = ref<any>(null)

const usageRate = computed(() => {
  if (!statsStore.stats) return 0
  const total = statsStore.stats.ports_used + statsStore.stats.ports_available
  if (total === 0) return 0
  return Math.round((statsStore.stats.ports_used / total) * 100)
})

const instanceUsagePercentage = computed(() => {
  if (!statsStore.stats || statsStore.stats.max_instances === 0) return 0
  return Math.min(100, Math.round((statsStore.stats.total_instances / statsStore.stats.max_instances) * 100))
})

const instanceUsageStatus = computed(() => {
  const p = instanceUsagePercentage.value
  if (p >= 90) return 'exception'
  if (p >= 70) return 'warning'
  return 'success'
})

const recentInstances = computed(() => {
  return [...proxyStore.instances]
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
    .slice(0, 5)
})

const pieOption = computed(() => {
  const data = Object.entries(nodeStore.stats?.nodes_by_type || {}).map(([name, value]) => ({
    name,
    value
  }))
  
  return {
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c} ({d}%)'
    },
    legend: {
      bottom: '0',
      left: 'center'
    },
    series: [
      {
        name: '节点类型',
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 10,
          borderColor: '#fff',
          borderWidth: 2
        },
        label: {
          show: false,
          position: 'center'
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 20,
            fontWeight: 'bold'
          }
        },
        labelLine: {
          show: false
        },
        data: data
      }
    ]
  }
})

const fetchData = () => {
  statsStore.fetchStats()
  nodeStore.fetchStats()
  proxyStore.fetchInstances()
}

const getDelayClass = (delay: number) => {
  if (delay < 200) return 'text-success'
  if (delay < 500) return 'text-warning'
  return 'text-danger'
}

const formatTime = (timeStr: string) => {
  return new Date(timeStr).toLocaleString()
}

onMounted(() => {
  fetchData()
  timer.value = setInterval(fetchData, 5000)
})

onUnmounted(() => {
  if (timer.value) clearInterval(timer.value)
})
</script>

<style scoped>
.mt-20 {
  margin-top: 20px;
}
.chart-container {
  height: 300px;
}
.chart {
  height: 100%;
}
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-danger { color: #f56c6c; }
</style>