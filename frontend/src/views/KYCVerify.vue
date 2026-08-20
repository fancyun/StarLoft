<template>
  <div class="kyc-verify-container">
    <!-- 有 token：跳转到上游认证 -->
    <div v-if="token" class="verify-card">
      <div v-if="!isMobile" class="desktop-section">
        <div class="qr-header">
          <div class="header-icon">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <rect x="3" y="3" width="7" height="7" rx="1"/>
              <rect x="14" y="3" width="7" height="7" rx="1"/>
              <rect x="3" y="14" width="7" height="7" rx="1"/>
              <rect x="14" y="14" width="7" height="7" rx="1"/>
            </svg>
          </div>
          <h2>扫码完成实名认证</h2>
          <p>请使用手机系统浏览器扫描二维码完成认证</p>
        </div>
        <div class="qr-wrapper">
          <canvas ref="qrCanvas"></canvas>
        </div>
        <div class="qr-link">
          <p class="link-label">或复制下方链接在手机浏览器中打开：</p>
	          <div class="link-box" @click="copyLink" style="cursor: pointer">
	            <span class="auth-link" :title="authUrl">{{ authUrl }}</span>
	          </div>
        </div>
      </div>

      <div v-else class="mobile-section">
        <div class="redirecting">
          <div class="spinner"></div>
          <h2>正在跳转到认证页面...</h2>
          <p>如未自动跳转，请点击下方按钮</p>
          <el-button type="primary" size="large" @click="redirectToAuth">前往认证</el-button>
        </div>
      </div>
    </div>

    <!-- 无 token：返回结果展示 -->
    <div v-else class="verify-card">
      <!-- 加载中 -->
      <div v-if="checking" class="status-section">
        <div class="spinner"></div>
        <h2>正在获取认证结果...</h2>
      </div>

      <!-- 认证成功 -->
      <div v-else-if="resultStatus === 2" class="status-section success">
        <el-icon class="result-icon success-icon"><SuccessFilled /></el-icon>
        <h2>实名认证成功</h2>
        <div class="result-info">
          <p>姓名：{{ maskName(resultName) }}</p>
          <p>身份证号：{{ maskIDCard(resultIDCard) }}</p>
        </div>
        <el-button type="primary" size="large" @click="goBack">返回</el-button>
      </div>

      <!-- 认证失败 -->
      <div v-else-if="resultStatus === 3" class="status-section failed">
        <el-icon class="result-icon failed-icon"><CircleCloseFilled /></el-icon>
        <h2>实名认证未通过</h2>
        <p v-if="resultMessage">原因：{{ resultMessage }}</p>
        <p v-else>请重新认证</p>
        <div class="status-actions">
          <el-button size="large" @click="goBack">返回</el-button>
          <el-button type="primary" size="large" @click="retryAuth">重新认证</el-button>
        </div>
      </div>

      <!-- 无法获取结果（未登录） -->
      <div v-else class="status-section">
        <div class="error-icon">
          <svg width="56" height="56" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"/>
            <line x1="12" y1="8" x2="12" y2="12"/>
            <line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
        </div>
        <h2>无法获取认证结果</h2>
        <p>请登录后查看认证状态</p>
        <el-button type="primary" size="large" @click="$router.push('/login')">去登录</el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { isMobileOrTablet } from '@/utils/device'
import { userAPI } from '@/api'
import QRCode from 'qrcode'
import { ElMessage } from 'element-plus'

const route = useRoute()
const router = useRouter()
const token = ref((route.query.token as string) || '')
const isMobile = ref(false)
const qrCanvas = ref<HTMLCanvasElement>()

const checking = ref(true)
const resultStatus = ref(0)  // 0=未知, 2=成功, 3=失败
const resultName = ref('')
const resultIDCard = ref('')
const resultMessage = ref('')
const returnUrl = ref('')
const hasResult = ref(false)

const authUrl = `https://api.yljz.com/finauth/lite/do?token=${token.value}`

function redirectToAuth() {
  window.location.href = authUrl
}

function copyLink() {
  navigator.clipboard.writeText(authUrl).then(() => {
    ElMessage.success('链接已复制到剪贴板')
  }).catch(() => {
    ElMessage.error('复制失败，请手动复制')
  })
}

function maskName(name: string) {
  if (!name) return ''
  return name.charAt(0) + '**'
}

function maskIDCard(idCard: string) {
  if (!idCard) return ''
  return idCard.substring(0, 3) + '***********' + idCard.substring(idCard.length - 4)
}

function goBack() {
  // 优先跳转到下游的 return_url，否则跳转到 /user/kyc
  if (returnUrl.value) {
    window.location.href = returnUrl.value
  } else {
    router.push('/user/kyc')
  }
}

function retryAuth() {
  router.push('/user/kyc')
}

async function fetchResult() {
  checking.value = true
  try {
    const statusRes: any = await userAPI.getKYCStatus()
    const recordStatus = statusRes.record_status

    // 保存 return_url（来自下游 API 调用方）
    if (statusRes.return_url) {
      returnUrl.value = statusRes.return_url
    }

    if (recordStatus === 2) {
      resultStatus.value = 2
      resultName.value = statusRes.record_name || statusRes.kyc_name || ''
      resultIDCard.value = statusRes.record_id_card || statusRes.kyc_id_card || ''
    } else if (recordStatus === 3) {
      resultStatus.value = 3
      resultMessage.value = statusRes.result_message || ''
    } else {
      // 仍进行中，等待后重试
      setTimeout(() => fetchResult(), 2000)
      return
    }
  } catch (error) {
    console.error('获取认证结果失败:', error)
    resultStatus.value = 0
  } finally {
    checking.value = false
  }

  // 结果出来后，自动跳转到下游 return_url（等待 3 秒让用户看到结果）
  setTimeout(() => {
    goBack()
  }, 3000)
}

async function syncAndShowResult() {
  checking.value = true
  try {
    const syncRes: any = await userAPI.syncKYC()
    if (syncRes.synced && syncRes.order_status >= 2) {
      // 上游已有结果，直接展示
      hasResult.value = true
      const statusRes: any = await userAPI.getKYCStatus()
      const recordStatus = statusRes.record_status

      if (statusRes.return_url) {
        returnUrl.value = statusRes.return_url
      }

      if (recordStatus === 2) {
        resultStatus.value = 2
        resultName.value = statusRes.record_name || statusRes.kyc_name || ''
        resultIDCard.value = statusRes.record_id_card || statusRes.kyc_id_card || ''
      } else if (recordStatus === 3) {
        resultStatus.value = 3
        resultMessage.value = statusRes.result_message || ''
      }

      // 3 秒后自动跳转
      setTimeout(() => goBack(), 3000)
      return
    }
    // 上游无结果，继续展示二维码
  } catch (error) {
    console.error('同步上游结果失败:', error)
  } finally {
    checking.value = false
  }
}

onMounted(async () => {
  isMobile.value = isMobileOrTablet()

  if (token.value) {
    // 有 token：先同步上游结果
    await syncAndShowResult()

    // 如果已有结果，不再展示二维码
    if (hasResult.value) return

    // 无结果，展示二维码
    if (isMobile.value) {
      setTimeout(() => redirectToAuth(), 1000)
    } else {
      await nextTick()
      if (qrCanvas.value) {
        QRCode.toCanvas(qrCanvas.value, authUrl, {
          width: 256,
          margin: 2,
          color: { dark: '#111827', light: '#FFFFFF' }
        })
      }
    }
  } else {
    // 无 token：从上游返回，自动回调结果
    fetchResult()
  }
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

/* 桌面端：二维码 */
.qr-header { margin-bottom: 32px; }
.header-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 64px; height: 64px;
  border-radius: 16px;
  background: var(--color-primary-light);
  color: var(--color-primary);
  margin-bottom: 16px;
}
.qr-header h2 { font-size: 22px; font-weight: 700; color: var(--text-primary); margin-bottom: 8px; }
.qr-header p { font-size: 14px; color: var(--text-muted); }
.qr-wrapper {
  display: inline-flex;
  padding: 16px;
  border: 2px dashed var(--border-color);
  border-radius: 12px;
  margin-bottom: 24px;
  background: var(--bg-soft);
}
.qr-wrapper canvas { display: block; }
.qr-link { text-align: center; }
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

/* 移动端：跳转中 */
.mobile-section { padding: 20px 0; }
.redirecting { text-align: center; }
.spinner {
  width: 48px; height: 48px;
  border: 3px solid var(--border-color);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 24px;
}
@keyframes spin { to { transform: rotate(360deg); } }
.redirecting h2 { font-size: 20px; font-weight: 600; color: var(--text-primary); margin-bottom: 8px; }
.redirecting p { font-size: 14px; color: var(--text-muted); margin-bottom: 24px; }

/* 结果展示 */
.status-section { padding: 20px 0; }
.status-section h2 { font-size: 22px; font-weight: 600; color: var(--text-primary); margin-bottom: 12px; }
.status-section p { font-size: 14px; color: var(--text-muted); margin-bottom: 24px; }
.result-icon { font-size: 56px; margin-bottom: 16px; }
.success-icon { color: var(--color-success); }
.failed-icon { color: var(--color-danger); }
.result-info {
  margin-bottom: 24px;
  padding: 16px 24px;
  background: var(--bg-soft);
  border-radius: var(--radius-md);
  display: inline-block;
}
.result-info p { color: var(--text-secondary); font-size: 14px; margin-bottom: 6px; }
.result-info p:last-child { margin-bottom: 0; }
.status-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}
.error-icon {
  color: var(--color-warning);
  margin-bottom: 16px;
}

@media (max-width: 480px) {
  .verify-card { padding: 32px 24px; }
}
</style>