<template>
  <div class="subscription-manage">
    <!-- Header with Actions -->
    <el-card class="header-card">
      <div class="header-actions">
        <div class="header-left">
          <h3>订阅源管理</h3>
          <span class="subtitle">管理代理订阅源，支持多个订阅地址</span>
        </div>
        <div class="header-right">
          <el-button type="primary" icon="Plus" @click="handleCreate">
            添加订阅源
          </el-button>
          <el-button icon="Refresh" @click="subscriptionStore.fetchSubscriptions()">
            刷新列表
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- Statistics Cards -->
    <el-row :gutter="20" class="mt-20">
      <el-col :span="6">
        <el-card>
          <div class="stat-card">
            <el-icon class="stat-icon" color="#409eff"><Document /></el-icon>
            <div class="stat-content">
              <div class="stat-value">{{ subscriptionStore.subscriptions.length }}</div>
              <div class="stat-label">总订阅数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-card">
            <el-icon class="stat-icon" color="#67c23a"><CircleCheck /></el-icon>
            <div class="stat-content">
              <div class="stat-value">{{ enabledCount }}</div>
              <div class="stat-label">已启用</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-card">
            <el-icon class="stat-icon" color="#e6a23c"><Share /></el-icon>
            <div class="stat-content">
              <div class="stat-value">{{ totalNodes }}</div>
              <div class="stat-label">总节点数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-card">
            <el-icon class="stat-icon" color="#f56c6c"><Clock /></el-icon>
            <div class="stat-content">
              <div class="stat-value">{{ recentFetchCount }}</div>
              <div class="stat-label">最近更新</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Subscription Table -->
    <el-card class="mt-20">
      <el-table
        :data="subscriptionStore.subscriptions"
        v-loading="subscriptionStore.loading"
        style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />

        <el-table-column prop="name" label="名称" width="150">
          <template #default="{ row }">
            <el-tag>{{ row.name }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="url" label="订阅地址" show-overflow-tooltip>
          <template #default="{ row }">
            <el-tooltip :content="row.url" placement="top">
              <span class="url-text">{{ truncateUrl(row.url) }}</span>
            </el-tooltip>
          </template>
        </el-table-column>

        <el-table-column prop="enabled" label="状态" width="100">
          <template #default="{ row }">
            <el-switch
              v-model="row.enabled"
              @change="subscriptionStore.toggleEnabled(row)"
              :loading="subscriptionStore.loading"
            />
          </template>
        </el-table-column>

        <el-table-column prop="node_count" label="节点数" width="100">
          <template #default="{ row }">
            <el-tag type="success" size="small">{{ row.node_count }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="last_fetch_status" label="最后状态" width="120">
          <template #default="{ row }">
            <el-tag
              v-if="row.last_fetch_status"
              :type="getStatusType(row.last_fetch_status)"
              size="small"
            >
              {{ row.last_fetch_status }}
            </el-tag>
            <span v-else class="text-muted">未获取</span>
          </template>
        </el-table-column>

        <el-table-column prop="last_fetch_at" label="最后更新" width="180">
          <template #default="{ row }">
            <span v-if="row.last_fetch_at">{{ formatTime(row.last_fetch_at) }}</span>
            <span v-else class="text-muted">从未更新</span>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="240" fixed="right">
          <template #default="{ row }">
            <el-button
              size="small"
              icon="VideoPlay"
              @click="handleTest(row)"
              :loading="subscriptionStore.testing[row.id]"
            >
              测试
            </el-button>
            <el-button
              size="small"
              icon="Edit"
              @click="handleEdit(row)"
            >
              编辑
            </el-button>
            <el-popconfirm
              title="确定要删除此订阅源吗？"
              confirm-button-text="确定"
              cancel-button-text="取消"
              @confirm="subscriptionStore.remove(row.id)"
            >
              <template #reference>
                <el-button size="small" type="danger" icon="Delete">
                  删除
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '添加订阅源' : '编辑订阅源'"
      width="600px"
      @close="handleDialogClose"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item label="订阅名称" prop="name">
          <el-input
            v-model="form.name"
            placeholder="例如：主订阅、备用订阅"
            clearable
          />
        </el-form-item>

        <el-form-item label="订阅地址" prop="url">
          <el-input
            v-model="form.url"
            type="textarea"
            :rows="3"
            placeholder="输入订阅链接地址"
            clearable
          />
          <div class="form-tip">
            支持标准的 Base64 订阅链接
          </div>
        </el-form-item>

        <el-form-item label="启用状态" prop="enabled">
          <el-switch v-model="form.enabled" />
          <div class="form-tip">
            启用后将自动定期获取节点
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          {{ dialogMode === 'create' ? '创建' : '保存' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useSubscriptionStore } from '@/stores/subscription'
import type { Subscription, CreateSubscriptionRequest } from '@/types/api'
import type { FormInstance, FormRules } from 'element-plus'

const subscriptionStore = useSubscriptionStore()

// Dialog state
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const currentEditId = ref<number | null>(null)
const submitting = ref(false)

// Form state
const formRef = ref<FormInstance>()
const form = ref<CreateSubscriptionRequest>({
  name: '',
  url: '',
  enabled: true
})

// Form validation rules
const formRules: FormRules = {
  name: [
    { required: true, message: '请输入订阅名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  url: [
    { required: true, message: '请输入订阅地址', trigger: 'blur' },
    {
      pattern: /^https?:\/\/.+/,
      message: '请输入有效的 HTTP/HTTPS 地址',
      trigger: 'blur'
    }
  ]
}

// Computed statistics
const enabledCount = computed(() =>
  subscriptionStore.subscriptions.filter(s => s.enabled).length
)

const totalNodes = computed(() =>
  subscriptionStore.subscriptions.reduce((sum, s) => sum + s.node_count, 0)
)

const recentFetchCount = computed(() => {
  const oneHourAgo = new Date(Date.now() - 3600000)
  return subscriptionStore.subscriptions.filter(s =>
    s.last_fetch_at && new Date(s.last_fetch_at) > oneHourAgo
  ).length
})

// Handlers
const handleCreate = () => {
  dialogMode.value = 'create'
  currentEditId.value = null
  form.value = {
    name: '',
    url: '',
    enabled: true
  }
  dialogVisible.value = true
}

const handleEdit = (subscription: Subscription) => {
  dialogMode.value = 'edit'
  currentEditId.value = subscription.id
  form.value = {
    name: subscription.name,
    url: subscription.url,
    enabled: subscription.enabled
  }
  dialogVisible.value = true
}

const handleTest = (subscription: Subscription) => {
  subscriptionStore.test(subscription.id)
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      let success = false
      if (dialogMode.value === 'create') {
        success = await subscriptionStore.create(form.value)
      } else if (currentEditId.value !== null) {
        success = await subscriptionStore.update(currentEditId.value, form.value)
      }

      if (success) {
        dialogVisible.value = false
      }
    } finally {
      submitting.value = false
    }
  })
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
}

// Utility functions
const truncateUrl = (url: string) => {
  if (url.length <= 60) return url
  return url.substring(0, 57) + '...'
}

const getStatusType = (status: string) => {
  if (status === 'success') return 'success'
  if (status === 'failed') return 'danger'
  return 'info'
}

const formatTime = (timeStr: string) => {
  return new Date(timeStr).toLocaleString('zh-CN')
}

// Lifecycle
onMounted(() => {
  subscriptionStore.fetchSubscriptions()
})
</script>

<style scoped>
.subscription-manage {
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

.stat-card {
  display: flex;
  align-items: center;
  gap: 15px;
}

.stat-icon {
  font-size: 40px;
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
  line-height: 1.2;
}

.stat-label {
  font-size: 13px;
  color: #909399;
  margin-top: 4px;
}

.url-text {
  color: #409eff;
  cursor: pointer;
}

.text-muted {
  color: #909399;
  font-size: 13px;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  line-height: 1.5;
  margin-top: 4px;
}
</style>
