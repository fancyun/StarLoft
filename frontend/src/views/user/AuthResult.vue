<template>
  <div class="result-container">
    <div class="result-card">
      <div v-if="status === 'loading'" class="status-section">
        <el-icon class="loading-icon"><Loading /></el-icon>
        <h2>正在查询认证结果...</h2>
        <p>请稍候</p>
      </div>

      <div v-else-if="status === 'success'" class="status-section success">
        <el-icon class="status-icon"><SuccessFilled /></el-icon>
        <h2>认证成功</h2>
        <p>您的实名认证已完成</p>
      </div>

      <div v-else-if="status === 'failed'" class="status-section failed">
        <el-icon class="status-icon"><CircleCloseFilled /></el-icon>
        <h2>认证失败</h2>
        <p>{{ errorMessage }}</p>
      </div>

      <div v-else class="status-section processing">
        <el-icon class="status-icon"><Clock /></el-icon>
        <h2>认证处理中</h2>
        <p>系统正在处理您的认证信息，请稍候...</p>
      </div>

      <div class="countdown">
        {{ countdown }} 秒后自动跳转到首页
      </div>

      <el-button type="primary" @click="goHome">返回首页</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { userAPI } from '@/api'

const router = useRouter()
const route = useRoute()

const status = ref('loading')
const errorMessage = ref('')
const countdown = ref(3)

const goHome = () => {
  router.push('/user/dashboard')
}

const checkStatus = async () => {
  try {
    const res = await userAPI.getKYCStatus()
    if (res.is_kyc_verified) {
      status.value = 'success'
    } else {
      status.value = 'processing'
    }
  } catch (error: any) {
    status.value = 'failed'
    errorMessage.value = error.message || '认证失败'
  }
}

onMounted(async () => {
  const bizId = route.query.biz_id as string
  
  if (bizId) {
    // 轮询查询状态
    let pollCount = 0
    const maxPoll = 15
    
    const poll = setInterval(async () => {
      await checkStatus()
      pollCount++
      
      if (status.value !== 'loading' && status.value !== 'processing') {
        clearInterval(poll)
        startCountdown()
      }
      
      if (pollCount >= maxPoll) {
        clearInterval(poll)
        status.value = 'processing'
        startCountdown()
      }
    }, 2000)
  } else {
    await checkStatus()
    startCountdown()
  }
})

const startCountdown = () => {
  const timer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      clearInterval(timer)
      goHome()
    }
  }, 1000)
}
</script>

<style scoped>
.result-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background:
    radial-gradient(circle at 20% 0%, var(--color-primary-light) 0%, transparent 50%),
    radial-gradient(circle at 100% 80%, #E0E7FF 0%, transparent 45%),
    var(--bg-page);
  padding: 20px;
}

.result-card {
  width: 100%;
  max-width: 500px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-xl);
  padding: 48px 40px;
  text-align: center;
  box-shadow: var(--shadow-lg);
}

.status-section {
  margin-bottom: 32px;
}

.loading-icon {
  width: 72px; height: 72px;
  color: var(--color-primary);
  animation: spin 1s linear infinite;
}

.status-icon {
  width: 72px; height: 72px;
  margin-bottom: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  font-size: 40px;
}

.success .status-icon {
  color: var(--color-success);
  background: var(--color-success-light);
}

.failed .status-icon {
  color: var(--color-danger);
  background: var(--color-danger-light);
}

.processing .status-icon {
  color: var(--color-warning);
  background: var(--color-warning-light);
}

.status-section h2 {
  font-size: 24px;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.status-section p {
  color: var(--text-muted);
  font-size: 14px;
}

.countdown {
  color: var(--text-muted);
  font-size: 14px;
  margin-bottom: 24px;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to   { transform: rotate(360deg); }
}
</style>
