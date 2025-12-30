<template>
  <div class="config-manage">
    <!-- Header -->
    <el-card class="header-card">
      <div class="header-actions">
        <div class="header-left">
          <h3>系统配置</h3>
          <span class="subtitle">管理系统运行时配置，修改后立即生效</span>
        </div>
        <div class="header-right">
          <el-button icon="RefreshRight" @click="configStore.reload()">
            从文件重载
          </el-button>
          <el-button icon="Refresh" @click="configStore.fetchConfig()">
            刷新配置
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- Configuration Tabs -->
    <el-card class="mt-20" v-loading="configStore.loading">
      <el-tabs v-model="activeTab" type="border-card">

        <!-- Proxy Configuration Tab -->
        <el-tab-pane label="代理配置" name="proxy">
          <el-form
            ref="proxyFormRef"
            :model="proxyForm"
            :rules="proxyRules"
            label-width="140px"
            class="config-form"
          >
            <el-divider content-position="left">端口范围</el-divider>
            <el-form-item label="最小端口" prop="port_range.min">
              <el-input-number
                v-model="proxyForm.port_range.min"
                :min="1024"
                :max="65535"
                :step="1"
                style="width: 200px"
              />
              <div class="form-tip">动态分配端口的起始值 (1024-65535)</div>
            </el-form-item>

            <el-form-item label="最大端口" prop="port_range.max">
              <el-input-number
                v-model="proxyForm.port_range.max"
                :min="1024"
                :max="65535"
                :step="1"
                style="width: 200px"
              />
              <div class="form-tip">动态分配端口的结束值 (1024-65535)</div>
            </el-form-item>

            <el-divider content-position="left">实例配置</el-divider>

            <el-form-item label="默认存活时间" prop="default_ttl">
              <el-input-number
                v-model="proxyForm.default_ttl"
                :min="60"
                :max="86400"
                :step="60"
                style="width: 200px"
              />
              <span class="unit-label">秒</span>
              <div class="form-tip">代理实例默认存活时间 (60-86400秒，即1分钟-24小时)</div>
            </el-form-item>

            <el-form-item label="最大实例数" prop="max_instances">
              <el-input-number
                v-model="proxyForm.max_instances"
                :min="1"
                :max="1000"
                :step="10"
                style="width: 200px"
              />
              <div class="form-tip">允许同时存在的最大代理实例数量</div>
            </el-form-item>

            <el-form-item label="默认协议" prop="default_protocol">
              <el-select v-model="proxyForm.default_protocol" style="width: 200px">
                <el-option label="SOCKS5" value="socks5" />
                <el-option label="HTTP" value="http" />
                <el-option label="Mixed" value="mixed" />
              </el-select>
              <div class="form-tip">创建代理时的默认协议类型</div>
            </el-form-item>

            <el-form-item>
              <el-button
                type="primary"
                @click="handleSaveProxy"
                :loading="configStore.saving"
              >
                保存代理配置
              </el-button>
              <el-button @click="resetProxyForm">重置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- Health Check Configuration Tab -->
        <el-tab-pane label="健康检查" name="health_check">
          <el-form
            ref="healthFormRef"
            :model="healthForm"
            :rules="healthRules"
            label-width="140px"
            class="config-form"
          >
            <el-form-item label="启用健康检查" prop="enabled">
              <el-switch v-model="healthForm.enabled" />
              <div class="form-tip">关闭后将不再定期检测节点可用性</div>
            </el-form-item>

            <el-form-item label="检查间隔" prop="interval">
              <el-input-number
                v-model="healthForm.interval"
                :min="60"
                :max="3600"
                :step="60"
                :disabled="!healthForm.enabled"
                style="width: 200px"
              />
              <span class="unit-label">秒</span>
              <div class="form-tip">每隔多久检查一次节点 (60-3600秒)</div>
            </el-form-item>

            <el-form-item label="超时时间" prop="timeout">
              <el-input-number
                v-model="healthForm.timeout"
                :min="1"
                :max="60"
                :step="1"
                :disabled="!healthForm.enabled"
                style="width: 200px"
              />
              <span class="unit-label">秒</span>
              <div class="form-tip">单次检查的超时时间 (1-60秒)</div>
            </el-form-item>

            <el-form-item label="检查地址" prop="url">
              <el-input
                v-model="healthForm.url"
                :disabled="!healthForm.enabled"
                style="width: 400px"
                placeholder="http://www.google.com/generate_204"
              />
              <div class="form-tip">用于测试节点连通性的 URL</div>
            </el-form-item>

            <el-form-item label="最大延迟" prop="max_delay">
              <el-input-number
                v-model="healthForm.max_delay"
                :min="100"
                :max="10000"
                :step="100"
                :disabled="!healthForm.enabled"
                style="width: 200px"
              />
              <span class="unit-label">毫秒</span>
              <div class="form-tip">超过此延迟的节点将被标记为不可用</div>
            </el-form-item>

            <el-form-item>
              <el-button
                type="primary"
                @click="handleSaveHealth"
                :loading="configStore.saving"
              >
                保存健康检查配置
              </el-button>
              <el-button @click="resetHealthForm">重置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- Subscription Configuration Tab -->
        <el-tab-pane label="订阅配置" name="subscription">
          <el-form
            ref="subscriptionFormRef"
            :model="subscriptionForm"
            :rules="subscriptionRules"
            label-width="140px"
            class="config-form"
          >
            <el-form-item label="更新间隔" prop="update_interval">
              <el-input-number
                v-model="subscriptionForm.update_interval"
                :min="300"
                :max="86400"
                :step="300"
                style="width: 200px"
              />
              <span class="unit-label">秒</span>
              <div class="form-tip">自动更新订阅的时间间隔 (300-86400秒，即5分钟-24小时)</div>
            </el-form-item>

            <el-form-item label="请求超时" prop="timeout">
              <el-input-number
                v-model="subscriptionForm.timeout"
                :min="5"
                :max="300"
                :step="5"
                style="width: 200px"
              />
              <span class="unit-label">秒</span>
              <div class="form-tip">获取订阅内容的超时时间 (5-300秒)</div>
            </el-form-item>

            <el-form-item>
              <el-button
                type="primary"
                @click="handleSaveSubscription"
                :loading="configStore.saving"
              >
                保存订阅配置
              </el-button>
              <el-button @click="resetSubscriptionForm">重置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- Read-only System Info Tab -->
        <el-tab-pane label="系统信息" name="system">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="服务地址">
              {{ configStore.config?.server.host }}:{{ configStore.config?.server.port }}
            </el-descriptions-item>
            <el-descriptions-item label="运行模式">
              <el-tag :type="configStore.config?.server.mode === 'release' ? 'success' : 'warning'">
                {{ configStore.config?.server.mode }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="日志级别">
              {{ configStore.config?.log.level }}
            </el-descriptions-item>
            <el-descriptions-item label="日志文件">
              {{ configStore.config?.log.file || '控制台输出' }}
            </el-descriptions-item>
          </el-descriptions>

          <el-alert
            class="mt-20"
            title="提示"
            type="info"
            :closable="false"
          >
            服务器和日志配置为只读，需要修改配置文件并重启服务
          </el-alert>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useConfigStore } from '@/stores/config'
import type { FormInstance, FormRules } from 'element-plus'
import type {
  UpdateProxyConfigRequest,
  UpdateHealthCheckConfigRequest,
  UpdateSubscriptionConfigRequest
} from '@/types/api'

const configStore = useConfigStore()

// Tab state
const activeTab = ref('proxy')

// Form refs
const proxyFormRef = ref<FormInstance>()
const healthFormRef = ref<FormInstance>()
const subscriptionFormRef = ref<FormInstance>()

// Proxy form
const proxyForm = ref<UpdateProxyConfigRequest>({
  port_range: { min: 20000, max: 30000 },
  default_ttl: 1800,
  max_instances: 100,
  default_protocol: 'socks5'
})

const proxyRules: FormRules = {
  'port_range.min': [
    { required: true, message: '请输入最小端口', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value >= proxyForm.value.port_range.max) {
          callback(new Error('最小端口必须小于最大端口'))
        } else {
          callback()
        }
      },
      trigger: 'change'
    }
  ],
  'port_range.max': [
    { required: true, message: '请输入最大端口', trigger: 'blur' }
  ],
  default_ttl: [
    { required: true, message: '请输入默认存活时间', trigger: 'blur' }
  ],
  max_instances: [
    { required: true, message: '请输入最大实例数', trigger: 'blur' }
  ],
  default_protocol: [
    { required: true, message: '请选择默认协议', trigger: 'change' }
  ]
}

// Health check form
const healthForm = ref<UpdateHealthCheckConfigRequest>({
  enabled: true,
  interval: 300,
  timeout: 10,
  url: 'http://www.google.com/generate_204',
  max_delay: 1000
})

const healthRules: FormRules = {
  interval: [
    { required: true, message: '请输入检查间隔', trigger: 'blur' }
  ],
  timeout: [
    { required: true, message: '请输入超时时间', trigger: 'blur' }
  ],
  url: [
    { required: true, message: '请输入检查地址', trigger: 'blur' },
    { pattern: /^https?:\/\/.+/, message: '请输入有效的 URL', trigger: 'blur' }
  ],
  max_delay: [
    { required: true, message: '请输入最大延迟', trigger: 'blur' }
  ]
}

// Subscription form
const subscriptionForm = ref<UpdateSubscriptionConfigRequest>({
  update_interval: 21600,
  timeout: 30
})

const subscriptionRules: FormRules = {
  update_interval: [
    { required: true, message: '请输入更新间隔', trigger: 'blur' }
  ],
  timeout: [
    { required: true, message: '请输入请求超时', trigger: 'blur' }
  ]
}

// Watch config changes and update forms
watch(() => configStore.config, (newConfig) => {
  if (newConfig) {
    proxyForm.value = {
      port_range: {
        min: newConfig.proxy.port_range.min,
        max: newConfig.proxy.port_range.max
      },
      default_ttl: newConfig.proxy.default_ttl,
      max_instances: newConfig.proxy.max_instances,
      default_protocol: newConfig.proxy.default_protocol
    }
    healthForm.value = {
      enabled: newConfig.health_check.enabled,
      interval: newConfig.health_check.interval,
      timeout: newConfig.health_check.timeout,
      url: newConfig.health_check.url,
      max_delay: newConfig.health_check.max_delay
    }
    subscriptionForm.value = {
      update_interval: newConfig.subscription.update_interval,
      timeout: newConfig.subscription.timeout
    }
  }
}, { immediate: true })

// Handlers
const handleSaveProxy = async () => {
  if (!proxyFormRef.value) return
  await proxyFormRef.value.validate(async (valid) => {
    if (valid) {
      await configStore.updateProxy(proxyForm.value)
    }
  })
}

const handleSaveHealth = async () => {
  if (!healthFormRef.value) return
  await healthFormRef.value.validate(async (valid) => {
    if (valid) {
      await configStore.updateHealthCheck(healthForm.value)
    }
  })
}

const handleSaveSubscription = async () => {
  if (!subscriptionFormRef.value) return
  await subscriptionFormRef.value.validate(async (valid) => {
    if (valid) {
      await configStore.updateSubscriptionConfig(subscriptionForm.value)
    }
  })
}

const resetProxyForm = () => {
  if (configStore.config) {
    proxyForm.value = {
      port_range: {
        min: configStore.config.proxy.port_range.min,
        max: configStore.config.proxy.port_range.max
      },
      default_ttl: configStore.config.proxy.default_ttl,
      max_instances: configStore.config.proxy.max_instances,
      default_protocol: configStore.config.proxy.default_protocol
    }
  }
}

const resetHealthForm = () => {
  if (configStore.config) {
    healthForm.value = {
      enabled: configStore.config.health_check.enabled,
      interval: configStore.config.health_check.interval,
      timeout: configStore.config.health_check.timeout,
      url: configStore.config.health_check.url,
      max_delay: configStore.config.health_check.max_delay
    }
  }
}

const resetSubscriptionForm = () => {
  if (configStore.config) {
    subscriptionForm.value = {
      update_interval: configStore.config.subscription.update_interval,
      timeout: configStore.config.subscription.timeout
    }
  }
}

// Lifecycle
onMounted(() => {
  configStore.fetchConfig()
})
</script>

<style scoped>
.config-manage {
  padding: 0;
}

.header-card {
  margin-bottom: 0;
}

.header-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left h3 {
  margin: 0 0 5px 0;
  font-size: 20px;
  color: #303133;
}

.subtitle {
  font-size: 13px;
  color: #909399;
}

.header-right {
  display: flex;
  gap: 10px;
}

.mt-20 {
  margin-top: 20px;
}

.config-form {
  max-width: 800px;
}

.unit-label {
  margin-left: 10px;
  color: #909399;
  font-size: 14px;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  line-height: 1.5;
  margin-top: 4px;
}
</style>
