<template>
  <div class="user-layout">
    <!-- 顶部蓝色 Header -->
    <header class="layout-header">
      <div class="header-left">
        <router-link to="/dashboard" class="header-title">
          <span class="brand-logo">SL</span>
          <span>星楼云控制台</span>
        </router-link>
      </div>
      <div class="header-right">
        <router-link to="/dashboard" class="user-name">
          <span>{{ userInfo?.phone || '用户' }}</span>
        </router-link>
      </div>
    </header>

    <div class="layout-body">
      <!-- 左侧黑色侧边栏 -->
      <aside class="layout-sidebar">
        <nav class="sidebar-nav">
          <router-link
            to="/dashboard"
            class="sidebar-item"
            :class="{ active: $route.path === '/dashboard' }"
          >
            <svg class="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
              <polyline points="9 22 9 12 15 12 15 22"/>
            </svg>
            <span>控制台首页</span>
          </router-link>

          <div class="sidebar-group">产品服务</div>
          <router-link
            to="/kyc"
            class="sidebar-item"
            :class="{ active: $route.path.startsWith('/kyc') }"
          >
            <svg class="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
            </svg>
            <span>实名认证</span>
          </router-link>
          <router-link
            to="/cs"
            class="sidebar-item"
            :class="{ active: $route.path.startsWith('/cs') }"
          >
            <svg class="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
              <line x1="8" y1="21" x2="16" y2="21"/>
              <line x1="12" y1="17" x2="12" y2="21"/>
            </svg>
            <span>云服务器</span>
          </router-link>
          <router-link
            to="/sms"
            class="sidebar-item"
            :class="{ active: $route.path.startsWith('/sms') }"
          >
            <svg class="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"/>
            </svg>
            <span>短信服务</span>
          </router-link>

          <div class="sidebar-group">资源与账户</div>
          <router-link
            to="/kyc/records"
            class="sidebar-item"
            :class="{ active: $route.path === '/kyc/records' }"
          >
            <svg class="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
              <line x1="16" y1="2" x2="16" y2="6"/>
              <line x1="8" y1="2" x2="8" y2="6"/>
              <line x1="3" y1="10" x2="21" y2="10"/>
            </svg>
            <span>认证记录</span>
          </router-link>
          <router-link
            to="/kyc/packs"
            class="sidebar-item"
            :class="{ active: $route.path.startsWith('/kyc/packs') }"
          >
            <svg class="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
              <polyline points="3.27 6.96 12 12.01 20.73 6.96"/>
              <line x1="12" y1="22.08" x2="12" y2="12"/>
            </svg>
            <span>资源包</span>
          </router-link>
          <router-link
            to="/balance"
            class="sidebar-item"
            :class="{ active: $route.path === '/balance' }"
          >
            <svg class="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="2" y="5" width="20" height="14" rx="2"/>
              <line x1="2" y1="10" x2="22" y2="10"/>
            </svg>
            <span>余额管理</span>
          </router-link>
          <router-link
            to="/api"
            class="sidebar-item"
            :class="{ active: $route.path === '/api' }"
          >
            <svg class="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="16 18 22 12 16 6"/>
              <polyline points="8 6 2 12 8 18"/>
            </svg>
            <span>API 管理</span>
          </router-link>
          <router-link
            to="/settings"
            class="sidebar-item"
            :class="{ active: $route.path === '/settings' }"
          >
            <svg class="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="3"/>
              <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
            </svg>
            <span>账户设置</span>
          </router-link>

          <div class="sidebar-footer">
            <button class="sidebar-item logout-btn" @click="handleLogout">
              <svg class="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
                <polyline points="16 17 21 12 16 7"/>
                <line x1="21" y1="12" x2="9" y2="12"/>
              </svg>
              <span>退出登录</span>
            </button>
          </div>
        </nav>
      </aside>

      <!-- 右侧白色主内容区 -->
      <main class="layout-main">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

const userInfo = computed(() => userStore.userInfo)

function handleLogout() {
  userStore.clearAuth()
  ElMessage.success('已退出登录')
  router.push('/login')
}
</script>

<style scoped>
.user-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

/* ========== 顶部 Header（腾讯云控制台风格：白底+下边框） ========== */
.layout-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
  padding: 0 24px;
  background: var(--bg-card);
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
  z-index: 100;
}

.header-left {
  display: flex;
  align-items: center;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
  color: var(--text-primary);
  font-size: 18px;
  font-weight: 700;
  letter-spacing: 0.5px;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-name {
  display: flex;
  align-items: center;
  gap: 8px;
  text-decoration: none;
  color: var(--text-secondary);
  font-size: 14px;
  padding: 6px 14px;
  border-radius: 6px;
  transition: background 0.2s;
}

.user-name:hover {
  background: var(--bg-hover);
  color: var(--color-primary);
}

/* ========== 主体 ========== */
.layout-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}

/* ========== 左侧白色侧边栏 ========== */
.layout-sidebar {
  width: 200px;
  background: var(--bg-card);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  overflow-y: auto;
}

.sidebar-nav {
  flex: 1;
  padding: 12px 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.sidebar-group {
  padding: 12px 24px 6px;
  font-size: 12px;
  color: var(--text-muted);
  letter-spacing: 1px;
}

.sidebar-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  margin: 0 8px;
  border-radius: 6px;
  text-decoration: none;
  color: var(--text-secondary);
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s;
}

.sidebar-item:hover {
  color: var(--text-primary);
  background: var(--bg-hover);
}

.sidebar-item.active {
  color: var(--color-primary);
  background: var(--bg-active);
  font-weight: 600;
}

.sidebar-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.sidebar-footer {
  padding: 8px 0 16px;
  border-top: 1px solid var(--border-light);
  margin: 0 8px;
}

.logout-btn {
  width: 100%;
  background: none;
  border: none;
  cursor: pointer;
  font-family: inherit;
  color: var(--text-muted);
}

.logout-btn:hover {
  color: var(--color-danger);
  background: var(--color-danger-light);
}

/* ========== 右侧主内容区 ========== */
.layout-main {
  flex: 1;
  background: var(--bg-page);
  overflow-y: auto;
  padding: 24px;
}
</style>