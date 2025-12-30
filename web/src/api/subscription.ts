import api from './index'
import type {
  Subscription,
  CreateSubscriptionRequest,
  UpdateSubscriptionRequest,
  TestSubscriptionResponse
} from '@/types/api'

// List all subscriptions
export function listSubscriptions(): Promise<Subscription[]> {
  return api.get('/subscriptions')
}

// Create new subscription
export function createSubscription(data: CreateSubscriptionRequest): Promise<Subscription> {
  return api.post('/subscriptions', data)
}

// Get single subscription
export function getSubscription(id: number): Promise<Subscription> {
  return api.get(`/subscriptions/${id}`)
}

// Update subscription
export function updateSubscription(id: number, data: UpdateSubscriptionRequest): Promise<Subscription> {
  return api.put(`/subscriptions/${id}`, data)
}

// Delete subscription
export function deleteSubscription(id: number): Promise<void> {
  return api.delete(`/subscriptions/${id}`)
}

// Test subscription connectivity
export function testSubscription(id: number): Promise<TestSubscriptionResponse> {
  return api.post(`/subscriptions/${id}/test`)
}
