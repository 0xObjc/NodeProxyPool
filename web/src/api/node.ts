import api from './index'
import type { Node, NodePoolStats } from '@/types/api'

export function getNodePool(): Promise<NodePoolStats> {
  return api.get('/nodePool')
}

export function getNodes(): Promise<Node[]> {
  return api.get('/nodes')
}

export function updateSubscription(): Promise<void> {
  return api.post('/subscription/update')
}
