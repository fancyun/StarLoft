<template>
  <div class="admin-packs">
    <div class="card">
      <div class="card-head">
        <h3 class="section-title">
          <el-icon><Box /></el-icon>
          资源包管理
        </h3>
        <el-button type="primary" @click="handleCreate">新增资源包</el-button>
      </div>

      <el-table :data="packs" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="名称" min-width="140" />
        <el-table-column prop="total_count" label="认证次数" width="110" />
        <el-table-column prop="price" label="售价" width="100">
          <template #default="{ row }">¥{{ row.price }}</template>
        </el-table-column>
        <el-table-column prop="stock" label="库存" width="110">
          <template #default="{ row }">
            <el-tag v-if="row.stock === -1" type="info">不限量</el-tag>
            <el-tag v-else :type="row.stock > 0 ? 'warning' : 'danger'">{{ row.stock }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '上架' : '下架' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="160">
          <template #default="{ row }">{{ row.description || '-' }}</template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" width="180">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">下架</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 新增/编辑资源包对话框 -->
    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑资源包' : '新增资源包'" width="520px">
      <el-form :model="form" :rules="formRules" ref="formRef" label-width="100px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入资源包名称" maxlength="100" />
        </el-form-item>
        <el-form-item label="认证次数" prop="total_count">
          <el-input-number
            v-model="form.total_count"
            :min="1"
            :max="100000"
            :step="10"
            style="width: 100%"
          />
          <div style="font-size: 12px; color: var(--text-muted); margin-top: 4px;">
            每个资源包包含的 KYC 认证次数
          </div>
        </el-form-item>
        <el-form-item label="售价" prop="price">
          <el-input-number
            v-model="form.price"
            :precision="2"
            :step="10"
            :min="0.01"
            :max="100000"
            style="width: 100%"
          />
          <div style="font-size: 12px; color: var(--text-muted); margin-top: 4px;">
            单位：元，购买时从用户余额扣费
          </div>
        </el-form-item>
        <el-form-item label="库存" prop="stock">
          <el-input-number
            v-model="form.stock"
            :min="-1"
            :max="1000000"
            style="width: 100%"
          />
          <div style="font-size: 12px; color: var(--text-muted); margin-top: 4px;">
            -1 表示不限量，>= 0 表示限量剩余库存
          </div>
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">上架</el-radio>
            <el-radio :value="0">下架</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="请输入资源包描述（可选）"
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'
import { adminAPI } from '@/api'
import { formatDateTime } from '@/utils/format'

const loading = ref(false)
const saving = ref(false)
const packs = ref<any[]>([])
const dialogVisible = ref(false)
const formRef = ref<FormInstance>()

const form = reactive({
  id: 0,
  name: '',
  total_count: 100,
  price: 100,
  stock: -1,
  status: 1,
  description: ''
})

const formRules: FormRules = {
  name: [
    { required: true, message: '请输入资源包名称', trigger: 'blur' },
    { min: 1, max: 100, message: '名称长度1-100个字符', trigger: 'blur' }
  ],
  total_count: [
    { required: true, message: '请输入认证次数', trigger: 'blur' }
  ],
  price: [
    { required: true, message: '请输入售价', trigger: 'blur' }
  ]
}

onMounted(() => {
  loadPacks()
})

const loadPacks = async () => {
  loading.value = true
  try {
    const response: any = await adminAPI.getPacks()
    packs.value = response.list || []
  } catch (error: any) {
    ElMessage.error('加载资源包失败')
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  form.id = 0
  form.name = ''
  form.total_count = 100
  form.price = 100
  form.stock = -1
  form.status = 1
  form.description = ''
}

const handleCreate = () => {
  resetForm()
  dialogVisible.value = true
  formRef.value?.clearValidate()
}

const handleEdit = (row: any) => {
  form.id = row.id
  form.name = row.name
  form.total_count = row.total_count
  form.price = row.price
  form.stock = row.stock
  form.status = row.status
  form.description = row.description || ''
  dialogVisible.value = true
  formRef.value?.clearValidate()
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      const data = {
        name: form.name.trim(),
        total_count: form.total_count,
        price: form.price,
        stock: form.stock,
        status: form.status,
        description: form.description
      }
      if (form.id) {
        await adminAPI.updatePack(form.id, data)
        ElMessage.success('资源包更新成功')
      } else {
        await adminAPI.createPack(data)
        ElMessage.success('资源包创建成功')
      }
      dialogVisible.value = false
      loadPacks()
    } catch (error: any) {
      ElMessage.error(error.response?.data?.message || '保存失败')
    } finally {
      saving.value = false
    }
  })
}

const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm(
      `确认下架资源包「${row.name}」？已售出的资源包不受影响。`,
      '提示',
      {
        confirmButtonText: '确定下架',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    await adminAPI.deletePack(row.id)
    ElMessage.success('资源包已下架')
    loadPacks()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.message || '下架失败')
    }
  }
}
</script>

<style scoped>
.admin-packs {
  min-height: 100%;
}
</style>
