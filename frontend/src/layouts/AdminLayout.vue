<template>
  <div class="admin-layout">
    <!-- 顶部 Header（一级，全宽置顶，与用户端一致） -->
    <el-header>
      <div class="header-left">
        <el-icon class="collapse-icon" @click="toggleCollapse">
          <Fold v-if="!isCollapse" />
          <Expand v-else />
        </el-icon>
        <router-link to="/admin/dashboard" class="header-title">
          <span class="brand-logo">SL</span>
          <span>管理后台</span>
        </router-link>
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
              <el-dropdown-item command="changePassword">修改密码</el-dropdown-item>
              <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-header>

    <div class="layout-body">
      <!-- 左侧侧边栏（二级，位于 Header 下方，与用户端一致） -->
      <el-aside width="200px">
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
          <el-menu-item index="/admin/packs">
            <el-icon><Box /></el-icon>
            <span>资源包管理</span>
          </el-menu-item>
          <el-menu-item index="/admin/config">
            <el-icon><Setting /></el-icon>
            <span>系统配置</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <!-- 右侧主内容区 -->
      <el-main>
        <router-view />
      </el-main>
    </div>

    <!-- 修改密码对话框 -->
    <el-dialog v-model="passwordDialogVisible" title="修改密码" width="420px" :close-on-click-modal="false">
      <el-form
        ref="passwordFormRef"
        :model="passwordForm"
        :rules="passwordRules"
        label-width="80px"
      >
        <el-form-item label="旧密码" prop="old_password">
          <el-input
            v-model="passwordForm.old_password"
            type="password"
            placeholder="请输入当前密码"
            show-password
          />
        </el-form-item>
        <el-form-item label="新密码" prop="new_password">
          <el-input
            v-model="passwordForm.new_password"
            type="password"
            placeholder="至少6位"
            show-password
          />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirm_password">
          <el-input
            v-model="passwordForm.confirm_password"
            type="password"
            placeholder="再次输入新密码"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="passwordDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="passwordLoading" @click="handleChangePassword">
          确定修改
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { 
  Odometer, 
  User, 
  Document, 
  Box, 
  Setting, 
  UserFilled, 
  ArrowDown, 
  Fold, 
  Expand 
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { adminAPI } from '@/api'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const isCollapse = ref(false)

const activeMenu = computed(() => route.path)
const adminName = computed(() => userStore.adminInfo?.username || '管理员')

const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value
}

// ===== 修改密码 =====
const passwordDialogVisible = ref(false)
const passwordLoading = ref(false)
const passwordFormRef = ref<FormInstance>()

const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const passwordRules: FormRules = {
  old_password: [{ required: true, message: '请输入当前密码', trigger: 'blur' }],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '新密码至少6位', trigger: 'blur' }
  ],
  confirm_password: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    {
      validator: (_rule, value, callback) => {
        if (value !== passwordForm.new_password) {
          callback(new Error('两次输入的新密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

const handleChangePassword = async () => {
  if (!passwordFormRef.value) return
  await passwordFormRef.value.validate(async (valid) => {
    if (!valid) return
    passwordLoading.value = true
    try {
      await adminAPI.changePassword({
        old_password: passwordForm.old_password,
        new_password: passwordForm.new_password
      })
      ElMessage.success('密码修改成功')
      passwordDialogVisible.value = false
      passwordForm.old_password = ''
      passwordForm.new_password = ''
      passwordForm.confirm_password = ''
    } catch (error) {
      // 错误消息已由请求拦截器统一提示
    } finally {
      passwordLoading.value = false
    }
  })
}

const handleCommand = (command: string) => {
  if (command === 'changePassword') {
    passwordForm.old_password = ''
    passwordForm.new_password = ''
    passwordForm.confirm_password = ''
    passwordFormRef.value?.clearValidate()
    passwordDialogVisible.value = true
  } else if (command === 'logout') {
    userStore.clearAdminInfo()
    ElMessage.success('已退出登录')
    router.push('/admin/login')
  }
}
</script>

<style scoped>
.admin-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

/* 主体（Header 下方：左侧边栏 + 右侧内容区，与用户端 layout-body 一致） */
.layout-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.el-aside {
  background: var(--bg-card);
  border-right: 1px solid var(--border-color);
  transition: width 0.2s;
  flex-shrink: 0;
  overflow-y: auto;
}

/* 头部品牌标识（与用户端 header-title 一致） */
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

.el-menu {
  border-right: none;
  background: transparent;
  padding: 12px 0;
}

:deep(.el-menu-item) {
  color: var(--text-secondary);
  margin: 0 8px 2px;
  border-radius: var(--radius-md);
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
  margin-right: 12px;
}

.el-header {
  background: var(--bg-card);
  border-bottom: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 24px;
  height: 60px;
  z-index: 100;
  flex-shrink: 0;
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
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}
</style>
