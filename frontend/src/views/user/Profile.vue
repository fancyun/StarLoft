<template>
  <div class="profile-container">
    <div class="content" v-loading="pageLoading">
      <div class="profile-grid">
        <!-- 个人信息 -->
        <div class="card">
          <h3>个人信息</h3>
          <div class="info-item">
            <span class="label">手机号</span>
            <span class="value">{{ maskPhone(userInfo.phone) }}</span>
          </div>
          <div class="info-item">
            <span class="label">实名状态</span>
            <span class="value">
              <el-tag v-if="userInfo.is_kyc_verified" type="success">已实名</el-tag>
              <el-tag v-else type="warning">未实名</el-tag>
            </span>
          </div>
          <div v-if="userInfo.is_kyc_verified" class="info-item">
            <span class="label">实名信息</span>
            <span class="value">{{ maskName(userInfo.kyc_name) }} / {{ maskIDCard(userInfo.kyc_id_card) }}</span>
          </div>
          <div class="info-item">
            <span class="label">注册时间</span>
            <span class="value">{{ userInfo.created_at }}</span>
          </div>
        </div>

        <!-- 余额管理 -->
        <div class="card">
          <h3>余额管理</h3>
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
                <el-radio label="unionpay">云闪付</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-button type="primary" @click="handleRecharge" :loading="rechargeLoading">
              充值
            </el-button>
          </el-form>
        </div>

        <!-- API密钥 -->
        <div class="card">
          <h3>API 密钥</h3>
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

        <!-- 安全设置 -->
        <div class="card">
          <h3>安全设置</h3>
          <el-button @click="dialogVisible = true">修改密码</el-button>
          <el-button @click="dialogVisible = false" type="danger">退出登录</el-button>
        </div>
      </div>
    </div>

    <!-- 修改密码弹窗 -->
    <el-dialog v-model="dialogVisible" title="修改密码" width="500px">
      <el-form :model="passwordForm" label-width="100px">
        <el-form-item label="短信验证码">
          <div style="display: flex; gap: 12px;">
            <el-input v-model="passwordForm.sms_code" placeholder="请输入验证码" />
            <el-button :disabled="countdown > 0" @click="sendCode">
              {{ countdown > 0 ? `${countdown}秒` : '获取验证码' }}
            </el-button>
          </div>
        </el-form-item>
        <el-form-item label="新密码">
          <el-input
            v-model="passwordForm.new_password"
            type="password"
            placeholder="请输入新密码"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleChangePassword" :loading="passwordLoading">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { userAPI } from '@/api'
import { useUserStore } from '@/stores/user'
import { verifyCaptcha } from '@/utils/captcha'

const router = useRouter()
const userStore = useUserStore()

const userInfo = ref<any>({})
const showSecret = ref(false)
const pageLoading = ref(true)
const dialogVisible = ref(false)
const rechargeLoading = ref(false)
const passwordLoading = ref(false)
const countdown = ref(0)

const rechargeForm = reactive({
  amount: '',
  channel: 'alipay'
})

const passwordForm = reactive({
  sms_code: '',
  new_password: ''
})

const maskPhone = (phone: string) => {
  if (!phone) return ''
  return phone.substring(0, 3) + '****' + phone.substring(7)
}

const maskName = (name: string) => {
  if (!name) return ''
  return name.charAt(0) + '**'
}

const maskIDCard = (idCard: string) => {
  if (!idCard) return ''
  return idCard.substring(0, 3) + '***********' + idCard.substring(idCard.length - 4)
}

const maskSecret = (secret: string) => {
  if (!secret) return ''
  return showSecret.value ? secret : '****' + secret.substring(secret.length - 4)
}

const copyText = (text: string) => {
  navigator.clipboard.writeText(text)
  ElMessage.success('已复制到剪贴板')
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
    
    if (res.pay_url) {
      window.open(res.pay_url)
    }
    
    ElMessage.success('充值订单已创建')
  } catch (error) {
    console.error(error)
  } finally {
    rechargeLoading.value = false
  }
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

const sendCode = async () => {
  try {
    const { ticket, randstr } = await verifyCaptcha()
    await userAPI.sendSMSCode({
      phone: userInfo.value.phone,
      captcha_ticket: ticket,
      captcha_randstr: randstr,
      scene: 'change_password'
    })
    ElMessage.success('验证码已发送')
    countdown.value = 60
    const timer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) {
        clearInterval(timer)
      }
    }, 1000)
  } catch (error) {
    console.error(error)
  }
}

const handleChangePassword = async () => {
  if (!passwordForm.sms_code || !passwordForm.new_password) {
    ElMessage.warning('请填写完整信息')
    return
  }
  
  passwordLoading.value = true
  
  try {
    // 修改密码前也触发人机验证
    const { ticket, randstr } = await verifyCaptcha()
    
    await userAPI.changePassword({
      ...passwordForm,
      captcha_ticket: ticket,
      captcha_randstr: randstr
    })
    
    ElMessage.success('密码修改成功，请重新登录')
    dialogVisible.value = false
    setTimeout(() => {
      userStore.logout()
      router.push('/user/login')
    }, 1000)
  } catch (error: any) {
    if (error.message === '用户取消验证') {
      ElMessage.info('已取消修改')
    } else {
      ElMessage.error(error.response?.data?.message || error.message || '修改失败')
    }
  } finally {
    passwordLoading.value = false
  }
}

const loadUserInfo = async () => {
  try {
    const res = await userAPI.getProfile()
    userInfo.value = res
  } catch (error) {
    console.error(error)
  } finally {
    pageLoading.value = false
  }
}

onMounted(() => {
  loadUserInfo()
})
</script>

<style scoped>
.profile-container {
  min-height: 100%;
}

.content {
  max-width: 1200px;
  margin: 0 auto;
}

.profile-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: var(--gap-lg);
}

.card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 28px;
  box-shadow: var(--shadow-sm);
}

.card h3 {
  font-size: 18px;
  color: var(--text-primary);
  margin-bottom: 24px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-light);
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px dashed var(--border-light);
}
.info-item:last-child { border-bottom: none; }

.info-item .label {
  color: var(--text-muted);
  font-size: 14px;
}

.info-item .value {
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 500;
}

.balance-display {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px;
  background: linear-gradient(135deg, var(--color-primary-light) 0%, #E0E7FF 100%);
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
}
</style>
