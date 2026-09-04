<template>
  <div class="admin-config">
    <div class="card">
      <h3 class="section-title">
        <el-icon><Setting /></el-icon>
        系统配置
      </h3>

      <el-alert
        type="info"
        :closable="false"
        style="margin: 0 0 20px"
      >
        <template #title>
          第三方业务配置（FinAuth / 腾讯云 / 支付宝 / 微信支付 / 邮件）已迁移至数据库，修改保存后即时生效，无需重启服务。
        </template>
      </el-alert>

      <!-- 平台基础 -->
      <el-form :model="configForm" label-width="150px" style="max-width: 700px; margin-bottom: 40px">
        <el-form-item label="实名认证单价">
          <el-input-number v-model="configForm.kyc_price" :min="0" :precision="2" :step="0.1" />
          <span style="margin-left: 10px; color: var(--text-muted)">元/次</span>
        </el-form-item>
      </el-form>

      <!-- FinAuth 实名认证 -->
      <h4 class="group-title">FinAuth 实名认证</h4>
      <el-form :model="configForm" label-width="150px" style="max-width: 700px; margin-bottom: 40px">
        <el-form-item label="API Key">
          <el-input v-model="configForm.finauth_api_key" show-password />
        </el-form-item>
        <el-form-item label="API Secret">
          <el-input v-model="configForm.finauth_api_secret" show-password />
        </el-form-item>
        <el-form-item label="场景 ID">
          <el-input v-model="configForm.finauth_scene_id" />
        </el-form-item>
        <el-form-item label="接口地址">
          <el-input v-model="configForm.finauth_base_url" />
        </el-form-item>
      </el-form>

      <!-- 腾讯云 -->
      <h4 class="group-title">腾讯云（短信 / 人机验证 / 邮件共用密钥）</h4>
      <el-form :model="configForm" label-width="150px" style="max-width: 700px; margin-bottom: 40px">
        <el-form-item label="SecretId">
          <el-input v-model="configForm.tencent_secret_id" show-password />
        </el-form-item>
        <el-form-item label="SecretKey">
          <el-input v-model="configForm.tencent_secret_key" show-password />
        </el-form-item>
        <el-form-item label="短信 SDKAppID">
          <el-input v-model="configForm.tencent_sms_sdk_app_id" />
        </el-form-item>
        <el-form-item label="短信签名">
          <el-input v-model="configForm.tencent_sms_sign_name" />
        </el-form-item>
        <el-form-item label="短信模板ID">
          <el-input v-model="configForm.tencent_sms_template_id" />
        </el-form-item>
        <el-form-item label="验证码 AppID">
          <el-input v-model="configForm.tencent_captcha_app_id" />
        </el-form-item>
        <el-form-item label="验证码 Secret">
          <el-input v-model="configForm.tencent_captcha_secret" show-password />
        </el-form-item>
      </el-form>

      <!-- 支付宝 -->
      <h4 class="group-title">支付宝支付</h4>
      <el-form :model="configForm" label-width="150px" style="max-width: 700px; margin-bottom: 40px">
        <el-form-item label="AppID">
          <el-input v-model="configForm.alipay_app_id" />
        </el-form-item>
        <el-form-item label="应用私钥(PEM)">
          <el-input v-model="configForm.alipay_private_key" type="textarea" :rows="4" />
        </el-form-item>
        <el-form-item label="支付宝公钥(PEM)">
          <el-input v-model="configForm.alipay_public_key" type="textarea" :rows="4" />
        </el-form-item>
      </el-form>

      <!-- 微信支付 -->
      <h4 class="group-title">微信支付</h4>
      <el-form :model="configForm" label-width="150px" style="max-width: 700px; margin-bottom: 40px">
        <el-form-item label="AppID">
          <el-input v-model="configForm.wechat_app_id" />
        </el-form-item>
        <el-form-item label="商户号">
          <el-input v-model="configForm.wechat_mch_id" />
        </el-form-item>
        <el-form-item label="APIv3 密钥">
          <el-input v-model="configForm.wechat_api_v3_key" show-password />
        </el-form-item>
        <el-form-item label="证书序列号">
          <el-input v-model="configForm.wechat_mch_serial_no" />
        </el-form-item>
        <el-form-item label="商户私钥(PEM)">
          <el-input v-model="configForm.wechat_mch_private_key" type="textarea" :rows="4" />
        </el-form-item>
        <el-form-item label="微信支付公钥(PEM)">
          <el-input v-model="configForm.wechat_platform_public_key" type="textarea" :rows="4" />
        </el-form-item>
      </el-form>

      <!-- 邮件 SES -->
      <h4 class="group-title">腾讯云 SES 邮件</h4>
      <el-form :model="configForm" label-width="150px" style="max-width: 700px; margin-bottom: 40px">
        <el-form-item label="发件人地址">
          <el-input v-model="configForm.ses_from" />
        </el-form-item>
        <el-form-item label="模板ID">
          <el-input v-model="configForm.ses_template_id" />
        </el-form-item>
        <el-form-item label="地域">
          <el-input v-model="configForm.ses_region" placeholder="如 ap-guangzhou" />
        </el-form-item>
      </el-form>

      <el-form-item label-width="150px">
        <el-button type="primary" @click="handleSave" :loading="saving" size="large">
          保存配置
        </el-button>
        <el-button @click="loadConfig" size="large">重置</el-button>
      </el-form-item>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { adminAPI } from '@/api'

const saving = ref(false)

const allKeys = [
  'kyc_price',
  'finauth_api_key',
  'finauth_api_secret',
  'finauth_scene_id',
  'finauth_base_url',
  'tencent_secret_id',
  'tencent_secret_key',
  'tencent_sms_sdk_app_id',
  'tencent_sms_sign_name',
  'tencent_sms_template_id',
  'tencent_captcha_app_id',
  'tencent_captcha_secret',
  'alipay_app_id',
  'alipay_private_key',
  'alipay_public_key',
  'wechat_app_id',
  'wechat_mch_id',
  'wechat_api_v3_key',
  'wechat_mch_serial_no',
  'wechat_mch_private_key',
  'wechat_platform_public_key',
  'ses_from',
  'ses_template_id',
  'ses_region'
]

const emptyForm = (): Record<string, string | number> => {
  const f: Record<string, string | number> = {}
  allKeys.forEach((k) => {
    f[k] = k === 'kyc_price' ? 1.0 : ''
  })
  return f
}

const configForm = reactive(emptyForm())

onMounted(() => {
  loadConfig()
})

const loadConfig = async () => {
  try {
    const response: any = await adminAPI.getConfig()
    const data = response?.data ?? response
    allKeys.forEach((k) => {
      if (data[k] !== undefined && data[k] !== null) {
        configForm[k] = data[k]
      }
    })
  } catch (error: any) {
    ElMessage.error('加载配置失败')
  }
}

const handleSave = async () => {
  saving.value = true
  try {
    const payload: Record<string, string | number> = {}
    allKeys.forEach((k) => {
      payload[k] = configForm[k]
    })
    await adminAPI.updateConfig(payload)
    ElMessage.success('保存成功，配置已即时生效')
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '保存失败')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.admin-config {
  min-height: 100%;
}

.group-title {
  margin: 0 0 16px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  border-left: 3px solid var(--el-color-primary);
  padding-left: 10px;
}
</style>