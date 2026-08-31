<template>
  <div class="admin-config">
    <div class="card">
      <h3 class="section-title">
        <el-icon><Setting /></el-icon>
        系统配置
      </h3>

      <el-form :model="configForm" label-width="150px" style="max-width: 600px">
        <el-form-item label="实名认证单价">
          <el-input-number v-model="configForm.kyc_price" :min="0" :precision="2" :step="0.1" />
          <span style="margin-left: 10px; color: var(--text-muted)">元/次</span>
        </el-form-item>

        <el-alert
          type="info"
          :closable="false"
          style="margin-bottom: 20px"
        >
          <template #title>
            密钥类配置请在 .env 环境变量中管理，修改后需重启服务生效。
          </template>
        </el-alert>

        <el-form-item>
          <el-button type="primary" @click="handleSave" :loading="saving" size="large">
            保存配置
          </el-button>
          <el-button @click="loadConfig" size="large">重置</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { adminAPI } from '@/api'

const saving = ref(false)

const configForm = reactive({
  kyc_price: 1.0
})

onMounted(() => {
  loadConfig()
})

const loadConfig = async () => {
  try {
    const response: any = await adminAPI.getConfig()
    Object.assign(configForm, response)
  } catch (error: any) {
    ElMessage.error('加载配置失败')
  }
}

const handleSave = async () => {
  saving.value = true
  try {
    await adminAPI.updateConfig({
      kyc_price: configForm.kyc_price
    })
    ElMessage.success('保存成功')
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
</style>
