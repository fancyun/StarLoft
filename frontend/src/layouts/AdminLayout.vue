<template>
  <div class="admin-layout">
    <el-container>
      <el-aside width="220px">
        <div class="logo">
          <h2>管理后台</h2>
        </div>
        <el-menu
          :default-active="activeMenu"
          router
          :collapse="isCollapse"
        >
          <el-menu-item index="/admin/dashboard">
            <el-icon><Odometer /></el-icon>
            <span>数据统计</span>
          </el-menu-item>
          <el-menu-item index="/admin/users">
            <el-icon><User /></el-icon>
            <span>用户管理</span>
          </el-menu-item>
          <el-menu-item index="/admin/orders">
            <el-icon><Document /></el-icon>
            <span>订单管理</span>
          </el-menu-item>
          <el-menu-item index="/admin/config">
            <el-icon><Setting /></el-icon>
            <span>系统配置</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <el-container>
        <el-header>
          <div class="header-left">
            <el-icon class="collapse-icon" @click="toggleCollapse">
              <Fold v-if="!isCollapse" />
              <Expand v-else />
            </el-icon>
          </div>
          <div class="header-right">
            <el-dropdown @command="handleCommand">
              <span class="user-dropdown">
                <el-icon><UserFilled /></el-icon>
                <span>{{ adminName }}</span>
                <el-icon><ArrowDown /></el-icon>
              </span>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="logout">退出登录</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </el-header>

        <el-main>
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { 
  Odometer, 
  User, 
  Document, 
  Setting, 
  UserFilled, 
  ArrowDown, 
  Fold, 
  Expand 
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const isCollapse = ref(false)

const activeMenu = computed(() => route.path)
const adminName = computed(() => userStore.adminInfo?.username || '管理员')

const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value
}

const handleCommand = (command: string) => {
  if (command === 'logout') {
    userStore.clearAdminInfo()
    ElMessage.success('已退出登录')
    router.push('/admin/login')
  }
}
</script>

<style scoped>
.admin-layout {
  height: 100vh;
}

.el-container {
  height: 100%;
}

.el-aside {
  background: var(--bg-card);
  border-right: 1px solid var(--border-color);
  box-shadow: 2px 0 6px rgba(17, 24, 39, 0.03);
  transition: width 0.2s;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  border-bottom: 1px solid var(--border-light);
  background: var(--bg-card);
}

.logo h2 {
  color: var(--text-primary);
  font-size: 18px;
  margin: 0;
  font-weight: 700;
  white-space: nowrap;
}

.el-menu {
  border-right: none;
  background: transparent;
}

:deep(.el-menu-item) {
  color: var(--text-secondary);
  margin: 4px 10px;
  border-radius: var(--radius-sm);
  height: 44px;
  line-height: 44px;
}

:deep(.el-menu-item:hover) {
  background: var(--bg-hover) !important;
  color: var(--text-primary) !important;
}

:deep(.el-menu-item.is-active) {
  background: var(--bg-active) !important;
  color: var(--color-primary) !important;
  font-weight: 600;
}

:deep(.el-menu-item .el-icon) {
  margin-right: 10px;
}

.el-header {
  background: var(--bg-card);
  border-bottom: 1px solid var(--border-color);
  box-shadow: 0 1px 4px rgba(17, 24, 39, 0.03);
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 24px;
  height: 60px;
}

.header-left {
  display: flex;
  align-items: center;
}

.collapse-icon {
  font-size: 20px;
  cursor: pointer;
  color: var(--text-secondary);
  transition: color 0.15s, transform 0.2s;
  padding: 6px 8px;
  border-radius: 6px;
}

.collapse-icon:hover {
  color: var(--color-primary);
  background: var(--bg-hover);
}

.header-right {
  display: flex;
  align-items: center;
}

.user-dropdown {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  transition: background 0.15s;
  color: var(--text-secondary);
}

.user-dropdown:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.el-main {
  background: var(--bg-page);
  padding: 0;
  overflow-y: auto;
}
</style>
