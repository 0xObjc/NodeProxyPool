import api from './index'
import type { CreateProxyRequest, ProxyInstance, ReleaseProxyRequest } from '@/types/api'

export function getProxy(data: CreateProxyRequest): Promise<ProxyInstance> {
  return api.post('/getProxy', data)
}

export function releaseProxy(data: ReleaseProxyRequest): Promise<void> {
  return api.post('/releaseProxy', data)
}

export function listInstances(): Promise<ProxyInstance[]> {
  return api.get('/listInstances')
}

export function getInstance(id: string): Promise<ProxyInstance> {
  return api.get(`/getInstance/${id}`)
}
