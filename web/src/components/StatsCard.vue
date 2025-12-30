<template>
  <el-card class="stats-card" :body-style="{ padding: '20px' }">
    <div class="stats-content">
      <div class="stats-icon" :style="{ backgroundColor: bgColor, color: iconColor }">
        <el-icon><component :is="icon" /></el-icon>
      </div>
      <div class="stats-info">
        <div class="stats-title">{{ title }}</div>
        <div class="stats-value">{{ value }}</div>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  value: string | number
  icon: string
  bgColor?: string
  iconColor?: string
}>()
</script>

<style scoped>
.stats-card {
  border-radius: 8px;
  border: none;
  box-shadow: 0 2px 12px 0 rgba(0,0,0,0.05);
  transition: transform 0.3s;
}
.stats-card:hover {
  transform: translateY(-5px);
}
.stats-content {
  display: flex;
  align-items: center;
}
.stats-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  margin-right: 16px;
}
.stats-title {
  font-size: 14px;
  color: #909399;
  margin-bottom: 4px;
}
.stats-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
}
</style>
