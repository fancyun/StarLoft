<template>
  <div class="kyc-container">
    <div class="content">
      <div class="kyc-card">
        <!-- 状态 1：已实名（record_status=2） -->
        <div v-if="recordStatus === 2" class="verified-section">
          <el-icon class="success-icon"><SuccessFilled /></el-icon>
          <h2>您已完成实名认证</h2>
          <div class="verified-info">
            <p>姓名：{{ maskName(kycName) }}</p>
            <p>身份证号：{{ maskIDCard(kycIDCard) }}</p>
          </div>
          <div class="verified-actions">
            <el-button @click="$router.push('/user/dashboard')">返回首页</el-button>
            <el-button type="warning" @click="replaceAuth">重新实名</el-button>
          </div>
        </div>

        <!-- 状态 2：进行中（record_status=1）→ 跳转到 /kyc -->
        <div v-if="recordStatus === 1" class="processing-section">
          <el-icon class="processing-icon"><Clock /></el-icon>
          <h2>认证进行中</h2>
          <p>您有一个实名认证正在进行中，请继续完成认证</p>
          <div class="processing-actions">
            <el-button type="primary" size="large" @click="continueAuth">继续认证</el-button>
            <el-button size="large" @click="cancelAuth">取消认证</el-button>
          </div>
        </div>

        <!-- 状态 3：无记录 / 认证失败 / 已更换（record_status=-1 / 3 / 4） -->
        <div v-if="recordStatus === -1 || recordStatus === 3 || recordStatus === 4" class="auth-form-section">
          <div class="form-header">
            <h2 v-if="recordStatus === -1">账户实名认证</h2>
            <h2 v-else-if="recordStatus === 4">更换实名认证</h2>
            <h2 v-else>实名认证失败</h2>
            <p v-if="recordStatus === -1">完成实名认证后，您可以使用 API 进行业务调用</p>
            <p v-else-if="recordStatus === 4">请填写新的姓名和身份证号，重新进行实名认证</p>
            <p v-else>上次实名认证未通过，请重新填写信息后再次认证</p>
          </div>

          <el-alert
            v-if="balance < kycPrice"
            title="余额不足"
            type="warning"
            :closable="false"
            style="margin-bottom: 24px"
          >
            <template #default>
              <p>您的账户余额不足，当前余额：¥{{ balance }}，认证费用：¥{{ kycPrice }}</p>
              <el-button type="primary" size="small" @click="$router.push('/user/profile')" style="margin-top: 8px">
                立即充值
              </el-button>
            </template>
          </el-alert>

          <el-form
            :model="form"
            :rules="rules"
            ref="formRef"
            label-width="100px"
            class="auth-form"
          >
            <el-form-item label="真实姓名" prop="name">
              <el-input
                v-model="form.name"
                placeholder="请输入真实姓名"
                :disabled="balance < kycPrice"
              />
            </el-form-item>
            <el-form-item label="身份证号" prop="id_card">
              <el-input
                v-model="form.id_card"
                placeholder="请输入身份证号"
                maxlength="18"
                :disabled="balance < kycPrice"
              />
            </el-form-item>
            <div class="cost-info">
              <el-icon><InfoFilled /></el-icon>
              <span>本次认证将从您的账户余额中扣除 ¥{{ kycPrice }}</span>
            </div>
            <el-button
              type="primary"
              size="large"
              class="submit-btn"
              :loading="loading"
              :disabled="balance < kycPrice"
              @click="handleSubmit"
            >
              开始认证
            </el-button>
          </el-form>

          <div class="tips">
            <h3>认证说明</h3>
            <ul>
              <li>请确保提供的姓名和身份证号真实有效</li>
              <li>认证过程中需要进行人脸识别，请在光线充足的环境下操作</li>
              <li>认证费用将从您的账户余额中扣除</li>
              <li>认证成功后，您将获得 API 密钥，可通过 API 调用认证服务</li>
              <li>如认证超时未完成，费用将自动退还</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { userAPI } from '@/api'

const router = useRouter()

const loading = ref(false)
const recordStatus = ref(-1)  // -1=无记录, 1=进行中, 2=已实名, 3=失败
const kycName = ref('')
const kycIDCard = ref('')
const balance = ref(0)
const kycPrice = ref(1.0)
const pendingToken = ref('')

const form = reactive({
  name: '',
  id_card: '',
  return_url: window.location.origin + '/kyc'
})

const validateIdCard = (_rule: any, value: string, callback: any) => {
  const reg = /(^\d{15}$)|(^\d{18}$)|(^\d{17}(\d|X|x)$)/
  if (!reg.test(value)) {
    callback(new Error('请输入正确的身份证号'))
  } else {
    callback()
  }
}

const rules = {
  name: [
    { required: true, message: '请输入真实姓名', trigger: 'blur' },
    { min: 2, max: 20, message: '姓名长度在 2 到 20 个字符', trigger: 'blur' }
  ],
  id_card: [
    { required: true, message: '请输入身份证号', trigger: 'blur' },
  { validator: validateIdCard, trigger: 'blur' }
  ]
}

const formRef = ref()

const maskName = (name: string) => {
  if (!name) return ''
  return name.charAt(0) + '**'
}

const maskIDCard = (idCard: string) => {
  if (!idCard) return ''
  return idCard.substring(0, 3) + '***********' + idCard.substring(idCard.length - 4)
}

const handleSubmit = async () => {
  await formRef.value.validate()

  if (balance.value < kycPrice.value) {
    ElMessage.warning('余额不足，请先充值')
    return
  }

  loading.value = true
  try {
    const res = await userAPI.startKYC({
      name: form.name,
      id_card: form.id_card,
      return_url: window.location.origin + '/user/kyc/result'
    })
    
    // 跳转到认证页面
    if (res && res.auth_url) {
      window.location.href = res.auth_url
    }
  } catch (error) {
    console.error(error)
  } finally {
    loading.value = false
  }
}

const continueAuth = () => {
  if (pendingToken.value) {
    router.push({ path: '/kyc', query: { token: pendingToken.value } })
  } else {
    router.push({ path: '/kyc' })
  }
}

const cancelAuth = async () => {
  try {
    await ElMessageBox.confirm('确定要取消当前认证吗？取消后可重新填写信息进行认证。', '取消认证', {
      confirmButtonText: '确定取消',
      cancelButtonText: '我再想想',
      type: 'warning'
    })
    await userAPI.cancelKYC()
    ElMessage.success('认证已取消')
    recordStatus.value = 3
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('取消失败，请稍后重试')
    }
  }
}

const replaceAuth = async () => {
  try {
    await ElMessageBox.confirm('更换实名后，您当前的实名信息将失效，需要重新填写姓名和身份证号进行认证。确定继续吗？', '更换实名', {
      confirmButtonText: '确定更换',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await userAPI.replaceKYC()
    ElMessage.success('已提交更换申请，请重新填写认证信息')
    loadData()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败，请稍后重试')
    }
  }
}

// Token提取函数（保留供将来使用）
// const _extractToken = (url: string): string | null => {
//   try {
//     return new URL(url).searchParams.get('token')
//   } catch {
//     const match = url.match(/token=([^&]+)/)
//     return match ? match[1] : null
//   }
// }

const loadData = async () => {
  try {
    const [statusRes, profileRes] = await Promise.all([
      userAPI.getKYCStatus(),
      userAPI.getProfile()
    ])

    const data = statusRes as any
    recordStatus.value = data.record_status ?? -1
    pendingToken.value = data.pending_token || ''

    if (data.kyc_name) kycName.value = data.kyc_name
    if (data.kyc_id_card) kycIDCard.value = data.kyc_id_card

    balance.value = profileRes.balance
    kycPrice.value = profileRes.kyc_price || 1.0
  } catch (error) {
    console.error(error)
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.kyc-container {
  min-height: 100%;
}
.content {
  max-width: 800px;
  margin: 0 auto;
}
.kyc-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 40px;
  box-shadow: var(--shadow-sm);
}

/* 已实名 */
.verified-section {
  text-align: center;
}
.success-icon {
  width: 72px; height: 72px;
  border-radius: 50%;
  background: var(--color-success-light);
  color: var(--color-success);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 40px;
  margin-bottom: 16px;
}
.verified-section h2 {
  font-size: 24px;
  color: var(--text-primary);
  margin-bottom: 24px;
}
.verified-info {
  margin-bottom: 32px;
  padding: 16px 24px;
  background: var(--bg-soft);
  border-radius: var(--radius-md);
  display: inline-block;
}
.verified-info p {
  color: var(--text-secondary);
  font-size: 14px;
  margin-bottom: 8px;
}
.verified-info p:last-child { margin-bottom: 0; }
.verified-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}

/* 重新实名 */
.reauth-section {
  margin-top: 24px;
}

/* 进行中 */
.processing-section {
  text-align: center;
  padding: 40px 0;
}
.processing-icon {
  width: 72px; height: 72px;
  border-radius: 50%;
  background: var(--color-warning-light);
  color: var(--color-warning);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 40px;
  margin-bottom: 16px;
}
.processing-section h2 {
  font-size: 24px;
  color: var(--text-primary);
  margin-bottom: 12px;
}
.processing-section p {
  color: var(--text-muted);
  font-size: 14px;
  margin-bottom: 24px;
}
.processing-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}

/* 未认证/失败 */
.form-header {
  text-align: center;
  margin-bottom: 32px;
}
.form-header h2 {
  font-size: 24px;
  color: var(--text-primary);
  margin-bottom: 8px;
}
.form-header p {
  color: var(--text-muted);
  font-size: 14px;
}
.auth-form {
  margin-bottom: 32px;
}
.cost-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 18px;
  background: var(--color-primary-light);
  border: 1px solid #BFDBFE;
  border-radius: var(--radius-md);
  margin-bottom: 24px;
  color: var(--color-primary);
  font-size: 14px;
  font-weight: 500;
}
.submit-btn {
  width: 100%;
  height: 48px;
  font-size: 16px;
}
.form-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}
.tips {
  padding: 24px;
  background: var(--bg-soft);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
}
.tips h3 {
  font-size: 16px;
  color: var(--text-primary);
  margin-bottom: 16px;
}
.tips ul {
  list-style: none;
  padding: 0;
}
.tips li {
  color: var(--text-secondary);
  font-size: 13px;
  margin-bottom: 8px;
  padding-left: 16px;
  position: relative;
}
.tips li::before {
  content: '•';
  position: absolute;
  left: 0;
  color: var(--color-primary);
  font-size: 16px;
  line-height: 1.4;
}
</style>