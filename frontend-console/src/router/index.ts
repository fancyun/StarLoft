import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'

// 控制台应用路由（部署于 console.starloft.cn）
// 产品化结构：实名认证 /kyc、云服务器 /cs、短信服务 /sms 等均为顶层产品路径
const routes: Array<RouteRecordRaw> = [
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
  // 控制台（需登录）
  {
    path: '/',
    component: () => import('@/views/user/UserLayout.vue'),
    meta: { requiresAuth: true },
    redirect: '/dashboard',
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
        path: 'kyc/records',
        name: 'Records',
        component: () => import('@/views/user/Records.vue')
      },
      {
        path: 'kyc/packs',
        name: 'ResourcePacks',
        component: () => import('@/views/user/ResourcePacks.vue')
      },
      {
        path: 'kyc/packs/buy',
        name: 'PurchasePacks',
        component: () => import('@/views/user/PurchasePacks.vue')
      },
      {
        path: 'cs',
        name: 'CloudServer',
        component: () => import('@/views/user/ComingSoon.vue'),
        meta: { productName: '云服务器' }
      },
      {
        path: 'sms',
        name: 'SMS',
        component: () => import('@/views/user/ComingSoon.vue'),
        meta: { productName: '短信服务' }
      },
      {
        path: 'balance',
        name: 'Balance',
        component: () => import('@/views/user/Balance.vue')
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
  } else if (!requiresAuth && !requiresAdmin && userStore.token && (to.path === '/login' || to.path === '/register')) {
    // 普通用户已登录访问登录/注册页，跳转到控制台首页
    next('/dashboard')
  } else if (to.path === '/admin/login' && userStore.adminToken) {
    // 管理员已登录访问登录页，跳转到管理后台
    next('/admin/dashboard')
  } else {
    next()
  }
})

export default router
