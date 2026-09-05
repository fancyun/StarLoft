import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useAdminStore } from '@/stores/admin'

// 管理后台独立应用路由（部署于 admin.starloft.cn）
const routes: Array<RouteRecordRaw> = [
  {
    path: '/login',
    name: 'AdminLogin',
    component: () => import('@/views/admin/Login.vue')
  },
  {
    path: '/',
    component: () => import('@/layouts/AdminLayout.vue'),
    meta: { requiresAdmin: true },
    redirect: '/dashboard', // 默认重定向到dashboard
    children: [
      {
        path: 'dashboard',
        name: 'AdminDashboard',
        component: () => import('@/views/admin/Dashboard.vue')
      },
      {
        path: 'users',
        name: 'AdminUsers',
        component: () => import('@/views/admin/Users.vue')
      },
      {
        path: 'orders',
        name: 'AdminOrders',
        component: () => import('@/views/admin/Orders.vue')
      },
      {
        path: 'packs',
        name: 'AdminResourcePacks',
        component: () => import('@/views/admin/ResourcePacks.vue')
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, _from, next) => {
  const adminStore = useAdminStore()
  const requiresAdmin = to.meta.requiresAdmin

  // 管理员页面未登录检查
  if (requiresAdmin && !adminStore.adminToken) {
    next('/login')
  } else if (to.path === '/login' && adminStore.adminToken) {
    // 管理员已登录访问登录页，跳转到管理后台
    next('/dashboard')
  } else {
    next()
  }
})

export default router
