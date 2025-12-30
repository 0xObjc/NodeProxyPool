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
