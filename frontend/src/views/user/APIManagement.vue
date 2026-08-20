<template>
  <div class="api-container">
    <div class="content">
      <div class="card" v-loading="pageLoading">
        <h3 class="section-title">
          <el-icon><Key /></el-icon>
          API 密钥
        </h3>
        <div class="api-item">
          <span class="label">API Key</span>
          <div class="api-value">
            <code>{{ userInfo.api_key }}</code>
            <el-button link @click="copyText(userInfo.api_key)">复制</el-button>
          </div>
        </div>
        <div class="api-item">
          <span class="label">API Secret</span>
          <div class="api-value">
            <code>{{ maskSecret(userInfo.api_secret) }}</code>
            <el-button link @click="showSecret = !showSecret">
              {{ showSecret ? '隐藏' : '显示' }}
            </el-button>
          </div>
        </div>
        <el-button @click="handleResetAPIKey" type="warning">重置 API 密钥</el-button>
      </div>

      <div class="card">
        <h3 class="section-title">
          <el-icon><Document /></el-icon>
          接口文档
        </h3>
        <div class="api-doc-item">
          <span class="doc-label">认证接口</span>
          <code class="doc-url">POST /api/v1/auth/verify</code>
        </div>
        <div class="api-doc-item">
          <span class="doc-label">结果查询</span>
          <code class="doc-url">GET /api/v1/auth/result</code>
        </div>
        <div class="api-doc-item">
          <span class="doc-label">认证方式</span>
          <div class="doc-desc">
            在请求头中携带 <code>Authorization: Bearer &lt;your_api_key&gt;</code>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { userAPI } from '@/api'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()

const userInfo = ref<any>({})
const showSecret = ref(false)
const pageLoading = ref(true)

const maskSecret = (secret: string) => {
  if (!secret) return ''
  return showSecret.value ? secret : '****' + secret.substring(secret.length - 4)
}

const copyText = (text: string) => {
  navigator.clipboard.writeText(text)
  ElMessage.success('已复制到剪贴板')
}

const handleResetAPIKey = async () => {
  try {
    await ElMessageBox.confirm('重置 API 密钥后，旧密钥将失效，确定要重置吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = await userAPI.resetAPIKey()
    userInfo.value.api_key = res.api_key
    userInfo.value.api_secret = res.api_secret
    ElMessage.success('API 密钥已重置')
  } catch (error) {
    console.error(error)
  }
}

const loadData = async () => {
  try {
    const profile = await userAPI.getProfile()
    userInfo.value = profile
    userStore.setUserInfo(profile)
  } catch (error) {
    console.error(error)
  } finally {
    pageLoading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.api-container {
  min-height: 100%;
}

.content {
  max-width: 700px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 24px;
  box-shadow: var(--shadow-sm);
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  color: var(--text-primary);
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-light);
}

.api-item {
  margin-bottom: 20px;
}

.api-item .label {
  display: block;
  color: var(--text-muted);
  font-size: 14px;
  margin-bottom: 8px;
}

.api-value {
  display: flex;
  align-items: center;
  gap: 12px;
}

.api-value code {
  flex: 1;
  padding: 10px 12px;
  background: var(--bg-soft);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-sm);
  color: var(--color-primary);
  font-family: 'Courier New', monospace;
  font-size: 13px;
  word-break: break-all;
}

.api-doc-item {
  padding: 10px 0;
  border-bottom: 1px dashed var(--border-light);
}

.api-doc-item:last-child { border-bottom: none; }

.doc-label {
  display: block;
  color: var(--text-muted);
  font-size: 13px;
  margin-bottom: 6px;
}

.doc-url {
  padding: 6px 10px;
  background: var(--bg-soft);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-sm);
  color: var(--color-primary);
  font-family: 'Courier New', monospace;
  font-size: 13px;
}

.doc-desc {
  color: var(--text-secondary);
  font-size: 13px;
  line-height: 1.6;
}

.doc-desc code {
  padding: 2px 6px;
  background: var(--bg-soft);
  border-radius: var(--radius-sm);
  font-family: 'Courier New', monospace;
  font-size: 12px;
  color: var(--color-primary);
}
</style>