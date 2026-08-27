import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'

const routes: Array<RouteRecordRaw> = [
  {
    path: '/',
    name: 'Home',
    component: () => import('@/views/Home.vue')
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/user/Login.vue')
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('@/views/user/Register.vue')
  },
  // 文档中心（公开，无需登录）
  {
    path: '/docs',
    component: () => import('@/views/docs/DocsLayout.vue'),
    children: [
      {
        path: '',
        name: 'DocsHome',
        component: () => import('@/views/docs/DocsHome.vue')
      },
      {
        path: 'api',
        name: 'DocsApi',
        component: () => import('@/views/docs/ApiDocs.vue')
      },
      {
        path: 'api/v1',
        name: 'DocsApiV1',
        component: () => import('@/views/docs/ApiV1.vue')
      },
      {
        path: 'plugin',
        name: 'DocsPlugin',
        component: () => import('@/views/docs/PluginDocs.vue')
      },
      {
        path: 'plugin/zjmf_mfcw',
        name: 'DocsPluginZjfMfcw',
        component: () => import('@/views/docs/PluginZjfMfcw.vue')
      }
    ]
  },
  // 用户功能（需登录）
  {
    path: '/user',
    component: () => import('@/views/user/UserLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/user/Dashboard.vue')
      },
      {
        path: 'kyc',
        name: 'KYC',
        component: () => import('@/views/user/KYC.vue')
      },
      {
        path: 'balance',
        name: 'Balance',
        component: () => import('@/views/user/Balance.vue')
      },
      {
        path: 'packs',
        name: 'ResourcePacks',
        component: () => import('@/views/user/ResourcePacks.vue')
      },
      {
        path: 'packs/buy',
        name: 'PurchasePacks',
        component: () => import('@/views/user/PurchasePacks.vue')
      },
      {
        path: 'records',
        name: 'Records',
        component: () => import('@/views/user/Records.vue')
      },
      {
        path: 'settings',
        name: 'AccountSettings',
        component: () => import('@/views/user/AccountSettings.vue')
      },
      {
        path: 'api',
        name: 'APIManagement',
        component: () => import('@/views/user/APIManagement.vue')
      }
    ]
  },
  // 管理后台路由
  {
    path: '/admin/login',
    name: 'AdminLogin',
    component: () => import('@/views/admin/Login.vue')
  },
  {
    path: '/admin',
    component: () => import('@/layouts/AdminLayout.vue'),
    meta: { requiresAdmin: true },
    redirect: '/admin/dashboard', // 默认重定向到dashboard
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
      },
      {
        path: 'config',
        name: 'AdminConfig',
        component: () => import('@/views/admin/Config.vue')
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
  const userStore = useUserStore()
  const requiresAuth = to.meta.requiresAuth
  const requiresAdmin = to.meta.requiresAdmin

  // 管理员页面未登录检查
  if (requiresAdmin && !userStore.adminToken) {
    // 需要管理员权限但未登录
    next('/admin/login')
  } else if (requiresAuth && !userStore.token) {
    // 普通用户页面需要登录但未登录
    next('/login')
  } else if (!requiresAuth && !requiresAdmin && userStore.token && (to.path === '/login' || to.path === '/register' || to.path === '/')) {
    // 普通用户已登录访问首页/登录/注册页，跳转到用户 dashboard
    next('/user/dashboard')
  } else if (to.path === '/admin/login' && userStore.adminToken) {
    // 管理员已登录访问登录页，跳转到管理后台
    next('/admin/dashboard')
  } else {
    next()
  }
})

export default router
