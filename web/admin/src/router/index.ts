import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import Layout from '@/layout/index.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { title: '登录' },
  },
  {
    path: '/',
    component: Layout,
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/index.vue'),
        meta: { title: '仪表盘', icon: 'Odometer' },
      },
      {
        path: 'devices',
        name: 'Devices',
        component: () => import('@/views/devices/index.vue'),
        meta: { title: '设备管理', icon: 'Monitor' },
      },
      {
        path: 'devices/:id',
        name: 'DeviceDetail',
        component: () => import('@/views/devices/detail.vue'),
        meta: { title: '设备详情', hidden: true },
      },
      {
        path: 'commands',
        name: 'Commands',
        component: () => import('@/views/commands/index.vue'),
        meta: { title: '指令管理', icon: 'Operation' },
      },
      {
        path: 'alerts',
        name: 'Alerts',
        component: () => import('@/views/alerts/index.vue'),
        meta: { title: '告警管理', icon: 'Bell' },
      },
      {
        path: 'sites',
        name: 'Sites',
        component: () => import('@/views/sites/index.vue'),
        meta: { title: '站点管理', icon: 'Location' },
      },
      {
        path: 'protocols',
        name: 'Protocols',
        component: () => import('@/views/protocols/index.vue'),
        meta: { title: '协议管理', icon: 'Connection' },
      },
      {
        path: 'reports',
        name: 'Reports',
        component: () => import('@/views/reports/index.vue'),
        meta: { title: '数据报表', icon: 'DataAnalysis' },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
