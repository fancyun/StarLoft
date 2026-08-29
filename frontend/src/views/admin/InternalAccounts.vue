<template>
  <div class="admin-internal-accounts">
    <div class="card">
      <div class="card-head">
        <h3 class="section-title">
          <el-icon><Lock /></el-icon>
          内部账号
        </h3>
        <div class="card-actions">
          <el-input
            v-model="keyword"
            placeholder="搜索名称 / 备注"
            clearable
            style="width: 220px"
            @keyup.enter="handleSearch"
            @clear="handleSearch"
          />
          <el-button @click="handleSearch">搜索</el-button>
          <el-button type="primary" @click="handleCreate">新增内部账号</el-button>
        </div>
      </div>

      <div class="tip">
        <el-icon><InfoFilled /></el-icon>
        <span>内部账号仅供本司其他系统调用，无需实名认证、不计费，且无法在用户端登录。API Key 与 Secret 用于调用方签名鉴权。</span>
      </div>

      <el-table :data="accounts" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column prop="remark" label="备注" min-width="160">
          <template #default="{ row }">{{ row.remark || '-' }}</template>
        </el-table-column>
        <el-table-column label="API Key" min-width="200">
          <template #default="{ row }">
            <span class="secret-text">{{ row.api_key }}</span>
            <el-button size="small" text type="primary" @click="copyText(row.api_key)">复制</el-button>
          </template>
        </el-table-column>
        <el-table-column label="API Secret" min-width="200">
          <template #default="{ row }">
            <span class="secret-text">{{ row.api_secret }}</span>
            <el-button size="small" text type="primary" @click="copyText(row.api_secret)">复制</el-button>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" width="200">
          <template #default="{ row }">
            <el-button size="small" type="primary" plain @click="handleResetAPI(row)">重置API</el-button>
            <el-button
              size="small"
              :type="row.status === 1 ? 'warning' : 'success'"
              plain
              @click="handleToggleStatus(row)"
            >
              {{ row.status === 1 ? '禁用' : '启用' }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrap">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadAccounts"
          @current-change="loadAccounts"
        />
      </div>
    </div>

    <!-- 新增内部账号对话框 -->
    <el-dialog v-model="dialogVisible" title="新增内部账号" width="520px">
      <el-form :model="form" :rules="formRules" ref="formRef" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入账号名称（标识本司系统，唯一）" maxlength="64" />
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input
            v-model="form.remark"
            type="textarea"
            :rows="3"
            placeholder="请输入备注（可选）"
            maxlength="255"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="saving">确定</el-button>
      </template>
    </el-dialog>

    <!-- 创建/重置成功展示 API Key/Secret 对话框 -->
    <el-dialog v-model="secretDialogVisible" title="API 密钥信息" width="560px">
      <el-alert
        type="warning"
        :closable="false"
        show-icon
        title="请立即复制并妥善保存，关闭后不再展示完整 Secret"
        style="margin-bottom: 16px"
      />
      <el-form label-width="90px">
        <el-form-item label="账号名称">
          <span>{{ secretData.name }}</span>
        </el-form-item>
        <el-form-item label="API Key">
          <div class="secret-row">
            <el-input :model-value="secretData.api_key" readonly />
            <el-button @click="copyText(secretData.api_key)">复制</el-button>
          </div>
        </el-form-item>
        <el-form-item label="API Secret">
          <div class="secret-row">
            <el-input :model-value="secretData.api_secret" readonly />
            <el-button @click="copyText(secretData.api_secret)">复制</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button type="primary" @click="secretDialogVisible = false">我已保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'
import { Lock, InfoFilled } from '@element-plus/icons-vue'
import { adminAPI } from '@/api'
import { formatDateTime } from '@/utils/format'

const loading = ref(false)
const saving = ref(false)
const accounts = ref<any[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')

const dialogVisible = ref(false)
const formRef = ref<FormInstance>()
const form = reactive({
  name: '',
  remark: ''
})

const secretDialogVisible = ref(false)
const secretData = reactive({
  name: '',
  api_key: '',
  api_secret: ''
})

const formRules: FormRules = {
  name: [
    { required: true, message: '请输入账号名称', trigger: 'blur' },
    { min: 1, max: 64, message: '名称长度1-64个字符', trigger: 'blur' }
  ]
}

onMounted(() => {
  loadAccounts()
})

const loadAccounts = async () => {
  loading.value = true
  try {
    const response: any = await adminAPI.getInternalAccounts({
      page: page.value,
      page_size: pageSize.value,
      keyword: keyword.value.trim()
    })
    accounts.value = response.list || []
    total.value = response.total || 0
  } catch (error: any) {
    ElMessage.error('加载内部账号失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  page.value = 1
  loadAccounts()
}

const copyText = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success('已复制')
  } catch (error: any) {
    ElMessage.error('复制失败，请手动复制')
  }
}

const handleCreate = () => {
  form.name = ''
  form.remark = ''
  dialogVisible.value = true
  formRef.value?.clearValidate()
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      const response: any = await adminAPI.createInternalAccount({
        name: form.name.trim(),
        remark: form.remark.trim()
      })
      ElMessage.success('内部账号创建成功')
      dialogVisible.value = false
      showSecret(response)
      loadAccounts()
    } catch (error: any) {
      ElMessage.error(error.response?.data?.message || '创建失败')
    } finally {
      saving.value = false
    }
  })
}

const showSecret = (data: any) => {
  secretData.name = data.name || ''
  secretData.api_key = data.api_key || ''
  secretData.api_secret = data.api_secret || ''
  secretDialogVisible.value = true
}

const handleResetAPI = async (row: any) => {
  try {
    await ElMessageBox.confirm(
      `确认重置内部账号「${row.name}」的 API Key/Secret？重置后旧密钥立即失效。`,
      '提示',
      {
        confirmButtonText: '确定重置',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    const response: any = await adminAPI.resetInternalAccountAPI(row.id)
    ElMessage.success('API 密钥重置成功')
    showSecret(response)
    loadAccounts()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.message || '重置失败')
    }
  }
}

const handleToggleStatus = async (row: any) => {
  const actionText = row.status === 1 ? '禁用' : '启用'
  try {
    await ElMessageBox.confirm(
      `确认${actionText}内部账号「${row.name}」？`,
      '提示',
      {
        confirmButtonText: `确定${actionText}`,
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    await adminAPI.updateInternalAccountStatus(row.id, {
      status: row.status === 1 ? 0 : 1
    })
    ElMessage.success(`${actionText}成功`)
    loadAccounts()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.message || `${actionText}失败`)
    }
  }
}
</script>

<style scoped>
.admin-internal-accounts {
  min-height: 100%;
}

.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.card-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tip {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 14px;
  margin-bottom: 16px;
  border-radius: var(--radius-sm);
  background: var(--bg-active);
  color: var(--text-secondary);
  font-size: 13px;
}

.secret-text {
  font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 12px;
  color: var(--text-secondary);
  margin-right: 4px;
}

.pagination-wrap {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.secret-row {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}
</style>
