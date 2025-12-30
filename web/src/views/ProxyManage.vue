<template>
  <div class="proxy-manage">
    <el-row :gutter="20">
      <!-- 创建表单 -->
      <el-col :span="8">
        <el-card header="创建新代理">
          <el-form :model="form" label-width="80px">
            <el-form-item label="协议">
              <el-select v-model="form.protocol" placeholder="选择协议" style="width: 100%">
                <el-option label="SOCKS5" value="socks5" />
                <el-option label="HTTP" value="http" />
                <el-option label="Mixed" value="mixed" />
              </el-select>
            </el-form-item>
            <el-form-item label="存活时间">
              <el-input-number v-model="form.ttl" :min="60" :max="3600" style="width: 100%" />
              <div class="form-tip">单位：秒 (60-3600)</div>
            </el-form-item>
            <el-form-item label="最大延迟">
              <el-input-number v-model="form.max_delay" :min="0" style="width: 100%" />
              <div class="form-tip">0 表示不限制延迟</div>
            </el-form-item>
            <el-form-item label="节点类型">
              <el-select v-model="form.node_type" placeholder="所有类型" clearable style="width: 100%">
                <el-option label="VMess" value="vmess" />
                <el-option label="Trojan" value="trojan" />
                <el-option label="Shadowsocks" value="ss" />
              </el-select>
            </el-form-item>
            <el-form-item label="地区">
              <el-input v-model="form.region" placeholder="地区关键字，如：香港" clearable />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="handleCreate" :loading="creating" style="width: 100%">
                立即创建
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </el-col>

      <!-- 实例列表 -->
      <el-col :span="16">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>活跃代理实例</span>
              <el-button @click="proxyStore.fetchInstances()" circle icon="Refresh" />
            </div>
          </template>
          
          <el-table :data="proxyStore.instances" style="width: 100%" v-loading="proxyStore.loading">
            <el-table-column prop="instance_id" label="ID" width="100">
              <template #default="{ row }">
                <el-tooltip :content="row.instance_id" placement="top">
                  <span>{{ row.instance_id.substring(0, 8) }}</span>
                </el-tooltip>
              </template>
            </el-table-column>
            <el-table-column label="地址" width="180">
              <template #default="{ row }">
                <el-link type="primary" @click="copyAddress(row)">
                  {{ row.host }}:{{ row.port }}
                </el-link>
              </template>
            </el-table-column>
            <el-table-column prop="protocol" label="协议" width="80">
              <template #default="{ row }">
                <el-tag size="small">{{ row.protocol.toUpperCase() }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="node_name" label="后端节点" show-overflow-tooltip />
            <el-table-column prop="node_delay" label="节点延迟" width="100">
               <template #default="{ row }">
                <span :class="getDelayClass(row.node_delay)">{{ row.node_delay }}ms</span>
              </template>
            </el-table-column>
            <el-table-column label="剩余时间" width="100">
              <template #default="{ row }">
                <CountdownTimer :expires-at="row.expires_at" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="100" fixed="right">
              <template #default="{ row }">
                <el-popconfirm title="确定要释放此实例吗？" @confirm="proxyStore.removeInstance(row.instance_id)">
                  <template #reference>
                    <el-button type="danger" size="small" icon="Delete" circle />
                  </template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useProxyStore } from '@/stores/proxy'
import type { CreateProxyRequest } from '@/types/api'
import CountdownTimer from '@/components/CountdownTimer.vue'
import { ElMessage } from 'element-plus'

const proxyStore = useProxyStore()
const creating = ref(false)
const timer = ref<any>(null)

const form = ref<CreateProxyRequest>({
  protocol: 'socks5',
  ttl: 300,
  max_delay: 0,
  node_type: '',
  region: ''
})

const handleCreate = async () => {
  creating.value = true
  try {
    await proxyStore.createInstance(form.value)
  } finally {
    creating.value = false
  }
}

const copyAddress = (row: any) => {
  const addr = `${row.host}:${row.port}`
  navigator.clipboard.writeText(addr)
  ElMessage.success('地址已复制到剪贴板')
}

const getDelayClass = (delay: number) => {
  if (delay < 200) return 'text-success'
  if (delay < 500) return 'text-warning'
  return 'text-danger'
}

onMounted(() => {
  proxyStore.fetchInstances()
  timer.value = setInterval(() => proxyStore.fetchInstances(), 3000)
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
.form-tip {
  font-size: 12px;
  color: #909399;
  line-height: 1.5;
  margin-top: 4px;
}
.text-success { color: #67c23a; }
.text-warning { color: #e6a23c; }
.text-danger { color: #f56c6c; }
</style>