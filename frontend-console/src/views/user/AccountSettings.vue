<template>
  <div class="account-container">
    <div class="content">
      <div class="card">
        <h3 class="section-title">
          <el-icon><Lock /></el-icon>
          修改密码
        </h3>

        <el-form :model="passwordForm" label-width="100px" class="password-form">
          <el-form-item label="短信验证码">
            <el-input v-model="passwordForm.sms_code" placeholder="请输入短信验证码" />
          </el-form-item>
          <el-form-item label="新密码">
            <el-input v-model="passwordForm.new_password" type="password" placeholder="请输入新密码" show-password />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleChangePassword" :loading="passwordLoading">
              确认修改
            </el-button>
          </el-form-item>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { userAPI } from '@/api'
import { verifyCaptcha } from '@/utils/captcha'

const passwordLoading = ref(false)

const passwordForm = reactive({
  sms_code: '',
  new_password: ''
})

const handleChangePassword = async () => {
  if (!passwordForm.sms_code || !passwordForm.new_password) {
    ElMessage.warning('请填写完整信息')
    return
  }

  passwordLoading.value = true
  try {
    const { ticket, randstr } = await verifyCaptcha()
    await userAPI.changePassword({
      sms_code: passwordForm.sms_code,
      new_password: passwordForm.new_password,
      captcha_ticket: ticket,
      captcha_randstr: randstr
    })
    ElMessage.success('密码修改成功')
    passwordForm.sms_code = ''
    passwordForm.new_password = ''
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || error.message || '修改失败')
  } finally {
    passwordLoading.value = false
  }
}
</script>

<style scoped>
.account-container {
  min-height: 100%;
}

.content {
  max-width: 700px;
  margin: 0 auto;
}

.password-form {
  max-width: 480px;
}
</style>