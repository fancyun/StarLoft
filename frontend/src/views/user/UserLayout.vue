<template>
  <div class="user-layout">
    <!-- 顶部蓝色 Header -->
    <header class="layout-header">
      <div class="header-left">
        <router-link to="/user/dashboard" class="header-title">
          <svg class="title-icon" viewBox="0 0 32 32" fill="none">
            <rect width="32" height="32" rx="8" fill="white" fill-opacity="0.2"/>
            <path d="M16 8L20 12L18 16L22 20L18 22L14 18L16 16L12 12L16 8Z" fill="white"/>
          </svg>
          <span>星楼KYC</span>
        </router-link>
      </div>
      <div class="header-right">
        <router-link to="/user/dashboard" class="user-name">
          <span>{{ userInfo?.phone || '用户' }}</span>
        </router-link>
      </div>
    </header>

    <div class="layout-body">
      <!-- 左侧黑色侧边栏 -->
      <aside class="layout-sidebar">
        <nav class="sidebar-nav">
          <router-link
            to="/user/dashboard"
            class="sidebar-item"
            :class="{ active: $route.path === '/user/dashboard' }"
          >
            <svg class="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
              <polyline points="9 22 9 12 15 12 15 22"/>
            </svg>
            <span>首页</span>
          </router-link>
          <router-link
            to="/user/settings"
            class="sidebar-item"
            :class="{ active: $route.path === '/user/settings' }"
          >
            <svg class="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="3"/>
              <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
            </svg>
            <span>账户设置</span>
          </router-link>
          <router-link
            to="/user/kyc"
            class="sidebar-item"
            :class="{ active: $route.path === '/user/kyc' }"
          >
            <svg class="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
            </svg>
            <span>实名认证</span>
          </router-link>
          <router-link
            to="/user/api"
            class="sidebar-item"
            :class="{ active: $route.path === '/user/api' }"
          >
            <svg class="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="16 18 22 12 16 6"/>
              <polyline points="8 6 2 12 8 18"/>
            </svg>
            <span>API 管理</span>
          </router-link>
          <router-link
            to="/user/balance"
            class="sidebar-item"
            :class="{ active: $route.path === '/user/balance' }"
          >
            <svg class="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="2" y="5" width="20" height="14" rx="2"/>
              <line x1="2" y1="10" x2="22" y2="10"/>
            </svg>
            <span>余额管理</span>
          </router-link>
          <router-link
            to="/user/records"
            class="sidebar-item"
            :class="{ active: $route.path === '/user/records' }"
          >
            <svg class="sidebar-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
              <line x1="16" y1="2" x2="16" y2="6"/>
              <line x1="8" y1="2" x2="8" y2="6"/>
              <line x1="3" y1="10" x2="21" y2="10"/>
            </svg>
            <span>认证记录</span>
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

/* ========== 顶部蓝色 Header ========== */
.layout-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
  padding: 0 24px;
  background: linear-gradient(135deg, #1e3a5f, #2563EB);
  flex-shrink: 0;
  z-index: 100;
  box-shadow: 0 2px 8px rgba(0,0,0,0.15);
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
  color: white;
  font-size: 18px;
  font-weight: 700;
  letter-spacing: 0.5px;
}

.title-icon {
  width: 28px;
  height: 28px;
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
  color: rgba(255,255,255,0.9);
  font-size: 14px;
  padding: 6px 14px;
  border-radius: 8px;
  transition: background 0.2s;
}

.user-name:hover {
  background: rgba(255,255,255,0.12);
  color: white;
}

/* ========== 左侧黑色侧边栏 ========== */
.layout-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}

/* ========== 左侧黑色侧边栏 ========== */
.layout-sidebar {
  width: 200px;
  background: #1a1a2e;
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

.sidebar-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;
  margin: 0 8px;
  border-radius: 8px;
  text-decoration: none;
  color: rgba(255,255,255,0.55);
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s;
}

.sidebar-item:hover {
  color: rgba(255,255,255,0.85);
  background: rgba(255,255,255,0.06);
}

.sidebar-item.active {
  color: white;
  background: rgba(37, 99, 235, 0.35);
}

.sidebar-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.sidebar-footer {
  padding: 8px 0 16px;
  border-top: 1px solid rgba(255,255,255,0.08);
  margin: 0 8px;
}

.logout-btn {
  width: 100%;
  background: none;
  border: none;
  cursor: pointer;
  font-family: inherit;
  color: rgba(255,255,255,0.45);
}

.logout-btn:hover {
  color: rgba(255,255,255,0.75);
  background: rgba(255,255,255,0.06);
}

/* ========== 右侧白色主内容 ========== */
.layout-main {
  flex: 1;
  background: #f5f6fa;
  overflow-y: auto;
  padding: 24px;
}
</style>