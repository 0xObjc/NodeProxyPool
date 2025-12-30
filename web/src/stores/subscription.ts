import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  listSubscriptions,
  createSubscription,
  updateSubscription,
  deleteSubscription,
  testSubscription
} from '@/api/subscription'
import type { Subscription, CreateSubscriptionRequest, UpdateSubscriptionRequest } from '@/types/api'
import { ElMessage } from 'element-plus'

export const useSubscriptionStore = defineStore('subscription', () => {
  const subscriptions = ref<Subscription[]>([])
  const loading = ref(false)
  const testing = ref<Record<number, boolean>>({})

  // Fetch all subscriptions
  async function fetchSubscriptions() {
    loading.value = true
    try {
      const data = await listSubscriptions()
      subscriptions.value = data || []
    } catch (error) {
      console.error('Failed to fetch subscriptions:', error)
      ElMessage.error('获取订阅列表失败')
    } finally {
      loading.value = false
    }
  }

  // Create new subscription
  async function create(data: CreateSubscriptionRequest) {
    try {
      await createSubscription(data)
      ElMessage.success('创建订阅成功')
      await fetchSubscriptions()
      return true
    } catch (error) {
      console.error('Failed to create subscription:', error)
      return false
    }
  }

  // Update subscription
  async function update(id: number, data: UpdateSubscriptionRequest) {
    try {
      await updateSubscription(id, data)
      ElMessage.success('更新订阅成功')
      await fetchSubscriptions()
      return true
    } catch (error) {
      console.error('Failed to update subscription:', error)
      return false
    }
  }

  // Delete subscription
  async function remove(id: number) {
    try {
      await deleteSubscription(id)
      ElMessage.success('删除订阅成功')
      await fetchSubscriptions()
    } catch (error) {
      console.error('Failed to delete subscription:', error)
    }
  }

  // Test subscription
  async function test(id: number) {
    testing.value[id] = true
    try {
      const result = await testSubscription(id)
      ElMessage.success(`测试成功！解析到 ${result.nodes} 个节点`)
      await fetchSubscriptions()
    } catch (error) {
      console.error('Failed to test subscription:', error)
    } finally {
      testing.value[id] = false
    }
  }

  // Toggle subscription enabled status
  async function toggleEnabled(subscription: Subscription) {
    await update(subscription.id, {
      name: subscription.name,
      url: subscription.url,
      enabled: !subscription.enabled
    })
  }

  return {
    subscriptions,
    loading,
    testing,
    fetchSubscriptions,
    create,
    update,
    remove,
    test,
    toggleEnabled
  }
})
