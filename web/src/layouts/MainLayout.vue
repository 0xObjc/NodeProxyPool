<template>
  <div class="main-layout">
    <el-container>
      <el-aside width="200px">
        <el-menu
          router
          :default-active="$route.path"
          class="el-menu-vertical"
        >
          <div class="logo">
            <h2>ProxyPool</h2>
          </div>
          <el-menu-item index="/dashboard">
            <el-icon><Odometer /></el-icon>
            <span>仪表盘</span>
          </el-menu-item>
          <el-menu-item index="/proxy">
            <el-icon><Connection /></el-icon>
            <span>代理管理</span>
          </el-menu-item>
          <el-menu-item index="/nodes">
            <el-icon><Share /></el-icon>
            <span>节点管理</span>
          </el-menu-item>
          <el-menu-item index="/subscription">
            <el-icon><Document /></el-icon>
            <span>订阅管理</span>
          </el-menu-item>
          <el-menu-item index="/config">
            <el-icon><Setting /></el-icon>
            <span>系统配置</span>
          </el-menu-item>
          <el-menu-item index="/monitor">
            <el-icon><Monitor /></el-icon>
            <span>实时监控</span>
          </el-menu-item>
        </el-menu>
      </el-aside>
      <el-container>
        <el-header>
          <div class="header-content">
            <span class="page-title">{{ $route.meta.title || 'ProxyPool' }}</span>
            <div style="flex: 1"></div>
            <el-button circle icon="Refresh" @click="handleRefresh" title="刷新页面" />
          </div>
        </el-header>
        <el-main>
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup lang="ts">
const handleRefresh = () => {
  location.reload()
}
</script>

<style scoped>
.main-layout {
  height: 100vh;
}
.el-container {
  height: 100%;
}
.el-aside {
  background-color: #fff;
  border-right: 1px solid #e6e6e6;
}
.el-menu {
  border-right: none;
}
.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid #e6e6e6;
  color: #409eff;
}
.el-header {
  background-color: #fff;
  border-bottom: 1px solid #e6e6e6;
  display: flex;
  align-items: center;
}
.header-content {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.el-main {
  background-color: #f5f7fa;
  padding: 20px;
}
</style>
