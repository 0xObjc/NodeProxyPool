import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '@/layouts/MainLayout.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: MainLayout,
      redirect: '/dashboard',
      children: [
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: () => import('@/views/Dashboard.vue')
        },
        {
          path: 'proxy',
          name: 'ProxyManage',
          component: () => import('@/views/ProxyManage.vue')
        },
        {
          path: 'nodes',
          name: 'NodeManage',
          component: () => import('@/views/NodeManage.vue')
        },
        {
          path: 'monitor',
          name: 'Monitor',
          component: () => import('@/views/Monitor.vue')
        }
      ]
    }
  ]
})

export default router
