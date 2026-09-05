import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

// 门户应用路由（部署于 www.starloft.cn）
// 产品页共用 ProductPage.vue，通过 meta.product 指定产品 key（/kyc /cs /sms 等）
const routes: Array<RouteRecordRaw> = [
  {
    path: '/',
    component: () => import('@/layouts/SiteLayout.vue'),
    children: [
      {
        path: '',
        name: 'Home',
        component: () => import('@/views/Home.vue')
      },
      {
        path: 'kyc',
        name: 'KYC',
        component: () => import('@/views/ProductPage.vue'),
        meta: { product: 'kyc' }
      },
      {
        path: 'cs',
        name: 'CloudServer',
        component: () => import('@/views/ProductPage.vue'),
        meta: { product: 'cs' }
      },
      {
        path: 'sms',
        name: 'SMS',
        component: () => import('@/views/ProductPage.vue'),
        meta: { product: 'sms' }
      }
    ]
  },
  // 文档中心
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
        name: 'ApiDocs',
        component: () => import('@/views/docs/ApiDocs.vue')
      },
      {
        path: 'api/v1',
        name: 'ApiV1',
        component: () => import('@/views/docs/ApiV1.vue')
      },
      {
        path: 'plugin',
        name: 'PluginDocs',
        component: () => import('@/views/docs/PluginDocs.vue')
      },
      {
        path: 'plugin/zjmf_mfcw',
        name: 'PluginZjfMfcw',
        component: () => import('@/views/docs/PluginZjfMfcw.vue')
      },
      {
        path: 'plugin/zjmf_v10',
        name: 'PluginZjfV10',
        component: () => import('@/views/docs/PluginZjfV10.vue')
      }
    ]
  },
  // 未匹配路由回退首页
  {
    path: '/:pathMatch(.*)*',
    redirect: '/'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() {
    return { top: 0 }
  }
})

export default router
