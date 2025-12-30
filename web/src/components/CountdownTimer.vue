<template>
  <span :class="['countdown-timer', colorClass]">
    {{ formattedTime }}
  </span>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

const props = defineProps<{
  expiresAt: string
}>()

const remaining = ref(0)
let timer: any = null

const updateRemaining = () => {
  const now = new Date().getTime()
  const end = new Date(props.expiresAt).getTime()
  remaining.value = Math.max(0, Math.floor((end - now) / 1000))
}

const formattedTime = computed(() => {
  const s = remaining.value
  const m = Math.floor(s / 60)
  const sec = s % 60
  return `${m}:${sec.toString().padStart(2, '0')}`
})

const colorClass = computed(() => {
  if (remaining.value < 60) return 'color-danger'
  if (remaining.value < 300) return 'color-warning'
  return 'color-success'
})

onMounted(() => {
  updateRemaining()
  timer = setInterval(updateRemaining, 1000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
.countdown-timer {
  font-family: monospace;
  font-weight: bold;
}
.color-success { color: #67c23a; }
.color-warning { color: #e6a23c; }
.color-danger { color: #f56c6c; }
</style>
