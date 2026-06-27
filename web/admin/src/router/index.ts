import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'
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
        path: 'devices/unregistered',
        name: 'UnregisteredDevices',
        component: () => import('@/views/devices/unregistered.vue'),
        meta: { title: '未注册设备', icon: 'Warning' },
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
      {
        path: 'logs',
        name: 'Logs',
        component: () => import('@/views/logs/index.vue'),
        meta: { title: '实时日志', icon: 'Tickets' },
      },
      {
        path: 'system',
        name: 'System',
        component: () => import('@/views/system/index.vue'),
        meta: { title: '系统监控', icon: 'Setting' },
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/users/index.vue'),
        meta: { title: '用户管理', icon: 'User', roles: ['admin', 'super_admin'] },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 路由守卫：未登录跳转登录页
router.beforeEach((to, _from, next) => {
  if (to.path === '/login') {
    next()
    return
  }

  const userStore = useUserStore()
  if (!userStore.isLoggedIn()) {
    next('/login')
    return
  }

  // 角色权限检查
  const requiredRoles = to.meta.roles as string[] | undefined
  if (requiredRoles && requiredRoles.length > 0) {
    const userRole = userStore.user?.role
    if (!userRole || !requiredRoles.includes(userRole)) {
      next('/dashboard')
      return
    }
  }

  next()
})

export default router
