<template>
  <div class="api-container">
    <div class="content">
      <div class="card" v-loading="pageLoading">
        <h3 class="section-title">
          <el-icon><Key /></el-icon>
          API 密钥
        </h3>
        <el-alert
          v-if="!pageLoading && isKycVerified !== 1"
          title="请先完成实名认证后再开通 API"
          type="warning"
          :closable="false"
          show-icon
          class="kyc-alert"
        >
          <p>API Key 已随注册自动生成；API Secret 需完成账户实名认证后自动下发，实名通过后即可调用 API。</p>
          <el-button type="primary" size="small" @click="$router.push('/kyc')">前往实名认证</el-button>
        </el-alert>
        <template v-if="!pageLoading">
          <div class="api-item">
            <span class="label">API Key</span>
            <div class="api-value">
              <code>{{ userInfo.api_key }}</code>
              <el-button link @click="copyText(userInfo.api_key)">复制</el-button>
            </div>
          </div>
          <template v-if="isKycVerified === 1">
            <div class="api-item">
              <span class="label">API Secret</span>
              <div class="api-value">
                <code>{{ maskSecret(userInfo.api_secret) }}</code>
                <el-button link @click="showSecret = !showSecret">
                  {{ showSecret ? '隐藏' : '显示' }}
                </el-button>
                <el-button link @click="copyText(userInfo.api_secret)">复制</el-button>
              </div>
            </div>
            <el-button @click="handleResetAPIKey" type="warning">重置 API 密钥</el-button>
          </template>
        </template>
      </div>

      <div class="card">
        <h3 class="section-title">
          <el-icon><Document /></el-icon>
          接口文档
        </h3>
        <p class="doc-tip">完整的接口调用方法与示例请前往文档中心查看。</p>
        <el-button type="primary" @click="goApiDocs">查看 API 文档</el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { userAPI } from '@/api'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()

const userInfo = ref<any>({})
const showSecret = ref(false)
const pageLoading = ref(true)
const isKycVerified = computed(() => (userStore.isKYCVerified ? 1 : 0))

const maskSecret = (secret: string) => {
  if (!secret) return ''
  return showSecret.value ? secret : '****' + secret.substring(secret.length - 4)
}

const copyText = (text: string) => {
  if (!text) {
    ElMessage.warning('暂无内容可复制')
    return
  }
  navigator.clipboard.writeText(text)
  ElMessage.success('已复制到剪贴板')
}

const goApiDocs = () => {
  window.open('https://www.starloft.cn/docs/api/v1', '_blank')
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
  } catch (error: any) {
    if (error?.response?.data?.message) {
      ElMessage.error(error.response.data.message)
    } else if (typeof error === 'string' && error !== 'cancel') {
      ElMessage.error(error)
    }
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

.doc-tip {
  color: var(--text-muted);
  font-size: 14px;
  margin-bottom: 16px;
}

.kyc-alert {
  margin-bottom: 8px;
}
.kyc-alert p {
  margin: 4px 0 12px;
  font-size: 13px;
  line-height: 1.6;
}
</style>