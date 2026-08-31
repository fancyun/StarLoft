<template>
  <div class="balance-container">
    <div class="content">
      <div class="card" v-loading="pageLoading">
        <h3 class="section-title">
          <el-icon><Wallet /></el-icon>
          余额管理
        </h3>
        <div class="balance-display">
          <span class="balance-label">当前余额</span>
          <span class="balance-amount">¥{{ userInfo.balance }}</span>
        </div>

        <el-form :model="rechargeForm" class="recharge-form">
          <el-form-item label="充值金额">
            <el-input
              v-model="rechargeForm.amount"
              placeholder="请输入充值金额"
              type="number"
            >
              <template #append>元</template>
            </el-input>
          </el-form-item>
          <el-form-item label="支付方式">
            <el-radio-group v-model="rechargeForm.channel">
              <el-radio label="alipay">支付宝</el-radio>
              <el-radio label="wechat">微信支付</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-button type="primary" @click="handleRecharge" :loading="rechargeLoading">
            充值
          </el-button>
        </el-form>

      </div>

      <el-dialog
        v-model="qrDialogVisible"
        title="微信扫码支付"
        width="360px"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        @closed="stopPolling"
      >
        <div class="qr-panel">
          <img :src="qrUrl" alt="微信支付二维码" />
          <p>请使用微信扫描二维码完成支付</p>
          <p class="qr-amount">支付金额：¥{{ rechargeForm.amount }}</p>
        </div>
      </el-dialog>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import { userAPI } from '@/api'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()

const userInfo = ref<any>({})
const rechargeLoading = ref(false)
const pageLoading = ref(true)

const qrDialogVisible = ref(false)
const qrUrl = ref('')

let pollTimer: ReturnType<typeof setInterval> | null = null
let pollCount = 0
const POLL_MAX = 120 // 最多轮询约5分钟

const rechargeForm = reactive({
  amount: '',
  channel: 'alipay'
})

const stopPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
  pollCount = 0
}

const startPolling = (payOrderNo: string) => {
  stopPolling()
  pollTimer = setInterval(async () => {
    pollCount++
    if (pollCount > POLL_MAX) {
      stopPolling()
      return
    }
    try {
      const res = await userAPI.getRechargeResult({ pay_order_no: payOrderNo })
      if (res.status === 1) {
        stopPolling()
        qrDialogVisible.value = false
        ElMessage.success('充值成功')
        loadData()
      }
    } catch (error) {
      // 轮询失败忽略，继续尝试
    }
  }, 2500)
}

const handleRecharge = async () => {
  if (!rechargeForm.amount || Number(rechargeForm.amount) <= 0) {
    ElMessage.warning('请输入正确的充值金额')
    return
  }
  rechargeLoading.value = true
  try {
    const res = await userAPI.createRecharge({
      amount: Number(rechargeForm.amount),
      channel: rechargeForm.channel
    })
    if (rechargeForm.channel === 'wechat') {
      qrUrl.value = res.qr_url
      qrDialogVisible.value = true
    } else {
      window.open(res.pay_url, '_blank')
    }
    startPolling(res.pay_order_no)
  } catch (error) {
    console.error(error)
  } finally {
    rechargeLoading.value = false
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

onBeforeUnmount(() => {
  stopPolling()
})
</script>

<style scoped>
.balance-container {
  min-height: 100%;
}

.content {
  max-width: 700px;
  margin: 0 auto;
}

.balance-display {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px;
  background: linear-gradient(135deg, var(--color-primary-light) 0%, #D6E8FF 100%);
  border-radius: var(--radius-md);
  margin-bottom: 24px;
}

.balance-label {
  color: var(--text-secondary);
  font-size: 14px;
  margin-bottom: 8px;
}

.balance-amount {
  color: var(--text-primary);
  font-size: 32px;
  font-weight: 700;
}

.recharge-form {
  margin-top: 8px;
}

.qr-panel {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px 0;
}

.qr-panel img {
  width: 280px;
  height: 280px;
}

.qr-panel p {
  margin: 8px 0 0;
  color: var(--text-secondary);
  font-size: 14px;
}

.qr-panel .qr-amount {
  color: var(--text-primary);
  font-weight: 600;
}
</style>