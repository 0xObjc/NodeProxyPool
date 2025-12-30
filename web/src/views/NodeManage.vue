<template>
  <div class="node-manage">
    <el-row :gutter="20" style="margin-bottom: 20px;">
      <el-col :span="8">
        <el-card header="节点健康状态">
          <div class="chart-container">
             <v-chart class="chart" :option="pieOption" autoresize />
          </div>
        </el-card>
      </el-col>
      <el-col :span="16">
        <el-card header="节点过滤" style="height: 100%">
           <el-form :inline="true">
             <el-form-item label="类型">
               <el-select v-model="filterType" placeholder="所有类型" clearable style="width: 120px">
                 <el-option label="VMess" value="vmess" />
                 <el-option label="Trojan" value="trojan" />
                 <el-option label="Shadowsocks" value="ss" />
               </el-select>
             </el-form-item>
             <el-form-item label="状态">
               <el-select v-model="filterStatus" placeholder="所有状态" clearable style="width: 120px">
                 <el-option label="可用" value="available" />
                 <el-option label="不可用" value="unavailable" />
               </el-select>
             </el-form-item>
             <el-form-item label="搜索">
               <el-input v-model="search" placeholder="节点名称..." prefix-icon="Search" clearable />
             </el-form-item>
             <el-form-item>
                <el-button type="primary" icon="Refresh" @click="nodeStore.triggerUpdate">
                  更新订阅
                </el-button>
             </el-form-item>
           </el-form>
           <div class="stats-summary">
             共 {{ nodeStore.nodes.length }} 个节点，筛选出 {{ filteredNodes.length }} 个
           </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card>
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <span>节点列表</span>
          </div>
        </div>
      </template>

      <el-table
        :data="filteredNodes"
        style="width: 100%"
        v-loading="nodeStore.loading"
        :default-sort="{ prop: 'delay', order: 'ascending' }"
      >
        <el-table-column prop="name" label="节点名称" show-overflow-tooltip />
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small" effect="plain">{{ row.type.toUpperCase() }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="delay" label="延迟" width="120" sortable>
          <template #default="{ row }">
            <span :class="getDelayClass(row.delay)">
              {{ row.delay === -1 ? '超时' : row.delay + 'ms' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="available" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.available ? 'success' : 'danger'" size="small">
              {{ row.available ? '可用' : '不可用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_check" label="最后检测" width="180" />
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useNodeStore } from '@/stores/node'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { PieChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import VChart from 'vue-echarts'

use([CanvasRenderer, PieChart, TitleComponent, TooltipComponent, LegendComponent])

const nodeStore = useNodeStore()
const search = ref('')
const filterType = ref('')
const filterStatus = ref('')
const timer = ref<any>(null)

const filteredNodes = computed(() => {
  return nodeStore.nodes.filter(node => {
    // Search filter
    if (search.value && !node.name.toLowerCase().includes(search.value.toLowerCase())) {
      return false
    }
    // Type filter
    if (filterType.value && node.type !== filterType.value) {
      return false
    }
    // Status filter
    if (filterStatus.value) {
      const isAvailable = filterStatus.value === 'available'
      if (node.available !== isAvailable) return false
    }
    return true
  })
})

const pieOption = computed(() => {
  const available = nodeStore.nodes.filter(n => n.available).length
  const unavailable = nodeStore.nodes.length - available
  
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
        name: '节点状态',
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
        data: [
          { value: available, name: '可用', itemStyle: { color: '#67c23a' } },
          { value: unavailable, name: '不可用', itemStyle: { color: '#f56c6c' } }
        ]
      }
    ]
  }
})

const getDelayClass = (delay: number) => {
  if (delay === -1 || delay > 1000) return 'text-danger'
  if (delay < 200) return 'text-success'
  if (delay < 500) return 'text-warning'
  return 'text-danger'
}

onMounted(() => {
  nodeStore.fetchNodes()
  timer.value = setInterval(() => nodeStore.fetchNodes(), 10000)
})

onUnmounted(() => {
  if (timer.value) clearInterval(timer.value)
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.ml-10 { margin-left: 10px; }
.text-success { color: #67c23a; font-weight: bold; }
.text-warning { color: #e6a23c; font-weight: bold; }
.text-danger { color: #f56c6c; font-weight: bold; }
.chart-container {
  height: 200px;
}
.chart {
  height: 100%;
}
.stats-summary {
  margin-top: 10px;
  color: #909399;
  font-size: 13px;
}
</style>