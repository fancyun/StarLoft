<template>
  <div class="kyc-verify-container">
    <div class="verify-card">
      <!-- 缺少 biz_no 参数 -->
      <div v-if="!bizNo" class="status-section">
        <el-icon class="result-icon"><WarningFilled /></el-icon>
        <h2>缺少参数</h2>
        <p>请从正确入口进入此页面</p>
        <el-button type="primary" size="large" @click="$router.push('/login')">去登录</el-button>
      </div>

      <!-- 加载中 -->
      <div v-else-if="checking" class="status-section">
        <div class="spinner"></div>
        <h2>正在获取认证结果...</h2>
      </div>

      <!-- 认证成功（已完成验证） -->
      <div v-else-if="resultStatus === 2" class="status-section success">
        <el-icon class="result-icon success-icon"><SuccessFilled /></el-icon>
        <h2>已完成验证</h2>
        <p>{{ countdown }} 秒后自动返回</p>
        <el-button type="primary" size="large" @click="goBack">立即返回</el-button>
      </div>

      <!-- 认证失败 -->
      <div v-else-if="resultStatus === 3" class="status-section failed">
        <el-icon class="result-icon failed-icon"><CircleCloseFilled /></el-icon>
        <h2>实名认证未通过</h2>
        <p v-if="resultMessage">原因：{{ resultMessage }}</p>
        <p>{{ countdown }} 秒后自动返回</p>
        <el-button type="primary" size="large" @click="goBack">立即返回</el-button>
      </div>

      <!-- 认证中：手动查询按钮 + 继续认证入口 -->
      <div v-else class="status-section">
        <el-icon class="result-icon"><Clock /></el-icon>
        <h2>认证处理中</h2>
        <p>如您已完成人脸验证，请点击下方按钮手动查询结果</p>
        <el-button type="primary" size="large" :loading="checking" @click="fetchPublicResult">
          我已完成验证，点击此处
        </el-button>

        <!-- 继续认证入口（存在上游 token 时展示） -->
        <div v-if="upToken" class="continue-auth">
          <div class="divider"><span>或继续完成认证</span></div>
          <template v-if="!isMobile">
            <div class="qr-wrapper">
              <canvas ref="qrCanvas"></canvas>
            </div>
            <p class="link-label">复制下方链接在手机浏览器中打开：</p>
            <div class="link-box" @click="copyLink" style="cursor: pointer">
              <span class="auth-link" :title="authUrl">{{ authUrl }}</span>
            </div>
          </template>
          <el-button v-else type="warning" size="large" @click="redirectToAuth">前往认证</el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { isMobileOrTablet } from '@/utils/device'
import { publicAPI } from '@/api'
import QRCode from 'qrcode'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()
const bizNo = ref((route.query.biz_no as string) || '')
const isMobile = ref(false)
const qrCanvas = ref<HTMLCanvasElement>()

const checking = ref(true)
const resultStatus = ref(0) // 0=认证中, 2=成功, 3=失败
const resultMessage = ref('')
const returnUrl = ref('')
const upToken = ref('')
const countdown = ref(5)
let countdownTimer: ReturnType<typeof setInterval> | null = null

const authUrl = computed(() => `https://api.yljz.com/finauth/lite/do?token=${upToken.value}`)

function redirectToAuth() {
  window.location.href = authUrl.value
}

function copyLink() {
  navigator.clipboard.writeText(authUrl.value).then(() => {
    ElMessage.success('链接已复制到剪贴板')
  }).catch(() => {
    ElMessage.error('复制失败，请手动复制')
  })
}

function goBack() {
  // 认证完成/失败后，优先返回下游的 return_url（下游最终用户场景）
  if (returnUrl.value) {
    window.location.href = returnUrl.value
  } else {
    // 下游未提供 return_url 时，退回浏览器上一页，避免错误跳转到平台登录页
    if (window.history.length > 1) {
      router.back()
    } else {
      router.push('/')
    }
  }
}

function startCountdown() {
  countdown.value = 5
  if (countdownTimer) clearInterval(countdownTimer)
  countdownTimer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      if (countdownTimer) clearInterval(countdownTimer)
      goBack()
    }
  }, 1000)
}

async function showQR() {
  await nextTick()
  if (qrCanvas.value) {
    QRCode.toCanvas(qrCanvas.value, authUrl.value, {
      width: 256,
      margin: 2,
      color: { dark: '#111827', light: '#FFFFFF' }
    })
  }
}

async function fetchPublicResult() {
  if (!bizNo.value) return
  checking.value = true
  try {
    const res: any = await publicAPI.getKycResult(bizNo.value)
    if (res.return_url) returnUrl.value = res.return_url
    if (res.up_token) upToken.value = res.up_token

    if (res.status === 2) {
      resultStatus.value = 2
    } else if (res.status === 3 || res.status === 4 || res.status === 5) {
      resultStatus.value = 3
      resultMessage.value = res.result_message || ''
    } else {
      resultStatus.value = 0
    }
  } catch (error) {
    resultStatus.value = 0
  } finally {
    checking.value = false
  }

  if (resultStatus.value === 2 || resultStatus.value === 3) {
    startCountdown()
  } else if (upToken.value && !isMobile.value) {
    showQR()
  }
}

onMounted(async () => {
  isMobile.value = isMobileOrTablet()
  if (bizNo.value) {
    await fetchPublicResult()
  } else {
    checking.value = false
  }
})

onBeforeUnmount(() => {
  if (countdownTimer) clearInterval(countdownTimer)
})
</script>

<style scoped>
.kyc-verify-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background:
    radial-gradient(circle at 20% 0%, var(--color-primary-light) 0%, transparent 50%),
    radial-gradient(circle at 100% 80%, #e0e7ff 0%, transparent 45%),
    var(--bg-page);
  padding: 20px;
}

.verify-card {
  width: 100%;
  max-width: 480px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-xl);
  padding: 48px 40px;
  box-shadow: var(--shadow-lg);
  text-align: center;
}

.spinner {
  width: 48px; height: 48px;
  border: 3px solid var(--border-color);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 24px;
}
@keyframes spin { to { transform: rotate(360deg); } }

.status-section { padding: 20px 0; }
.status-section h2 { font-size: 22px; font-weight: 600; color: var(--text-primary); margin-bottom: 12px; }
.status-section p { font-size: 14px; color: var(--text-muted); margin-bottom: 24px; }
.result-icon { font-size: 56px; margin-bottom: 16px; }
.success-icon { color: var(--color-success); }
.failed-icon { color: var(--color-danger); }

.continue-auth { margin-top: 32px; }
.divider {
  display: flex;
  align-items: center;
  gap: 12px;
  color: var(--text-muted);
  font-size: 13px;
  margin-bottom: 20px;
}
.divider::before,
.divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: var(--border-color);
}

.qr-wrapper {
  display: inline-flex;
  padding: 16px;
  border: 2px dashed var(--border-color);
  border-radius: 12px;
  margin-bottom: 20px;
  background: var(--bg-soft);
}
.qr-wrapper canvas { display: block; }
.link-label { font-size: 13px; color: var(--text-muted); margin-bottom: 12px; }
.link-box {
  padding: 12px 16px;
  background: var(--bg-soft);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
}
.auth-link {
  display: block;
  font-size: 13px;
  color: var(--color-primary);
  text-decoration: none;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.auth-link:hover { color: var(--color-primary-hover); text-decoration: underline; }

@media (max-width: 480px) {
  .verify-card { padding: 32px 24px; }
}
</style>