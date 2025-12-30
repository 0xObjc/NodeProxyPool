import api from './index'
import type { SystemStats } from '@/types/api'

export function getStats(): Promise<SystemStats> {
  return api.get('/stats')
}
