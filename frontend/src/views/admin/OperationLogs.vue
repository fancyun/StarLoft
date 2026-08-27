<template>
  <div class="admin-operation-logs">
    <div class="card">
      <div class="card-head">
        <h3 class="section-title">
          <el-icon><Document /></el-icon>
          操作日志
        </h3>
      </div>

      <div class="filter-bar">
        <el-input
          v-model="filters.adminName"
          placeholder="搜索管理员用户名"
          style="width: 200px"
          clearable
          @clear="handleSearch"
          @keyup.enter="handleSearch"
        >
          <template #append>
            <el-button icon="Search" @click="handleSearch" />
          </template>
        </el-input>

        <el-select v-model="filters.operation" placeholder="操作类型" clearable style="width: 160px" @change="handleSearch">
          <el-option label="全部" :value="null" />
          <el-option label="登录" value="login" />
          <el-option label="注册" value="register" />
          <el-option label="禁用/启用用户" value="update_user_status" />
          <el-option label="删除用户" value="delete_user" />
          <el-option label="充值" value="recharge" />
          <el-option label="修改配置" value="update_config" />
          <el-option label="资源包" value="resource_pack" />
          <el-option label="修改密码" value="change_password" />
        </el-select>

        <el-button @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
      </div>

      <el-table :data="logs" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="admin_name" label="管理员" width="140">
          <template #default="{ row }">{{ row.admin_name || '-' }}</template>
        </el-table-column>
        <el-table-column prop="operation" label="操作类型" width="150">
          <template #default="{ row }">
            <el-tag :type="getOperationTag(row.operation)" size="small">
              {{ getOperationName(row.operation) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="resource_type" label="资源类型" width="120">
          <template #default="{ row }">{{ getResourceTypeName(row.resource_type) }}</template>
        </el-table-column>
        <el-table-column prop="resource_id" label="资源ID" width="100">
          <template #default="{ row }">{{ row.resource_id || '-' }}</template>
        </el-table-column>
        <el-table-column prop="details" label="详情" min-width="240" show-overflow-tooltip>
          <template #default="{ row }">{{ row.details || '-' }}</template>
        </el-table-column>
        <el-table-column prop="ip_address" label="IP地址" width="140">
          <template #default="{ row }">{{ row.ip_address || '-' }}</template>
        </el-table-column>
        <el-table-column prop="created_at" label="操作时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="loadLogs"
        @size-change="loadLogs"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { adminAPI } from '@/api'
import { formatDateTime } from '@/utils/format'

const loading = ref(false)

const filters = reactive({
  adminName: '',
  operation: null as string | null
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const logs = ref<any[]>([])

onMounted(() => {
  loadLogs()
})

const loadLogs = async () => {
  loading.value = true
  try {
    const response: any = await adminAPI.getLogs({
      page: pagination.page,
      page_size: pagination.pageSize,
      admin_name: filters.adminName || undefined,
      operation: filters.operation || undefined
    })
    logs.value = response.list || []
    pagination.total = response.total || 0
  } catch (error: any) {
    console.error('加载操作日志失败:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadLogs()
}

const handleReset = () => {
  filters.adminName = ''
  filters.operation = null
  handleSearch()
}

const getOperationName = (op: string) => {
  const names: Record<string, string> = {
    login: '登录',
    register: '注册用户',
    update_user_status: '禁用/启用用户',
    delete_user: '删除用户',
    recharge: '人工充值',
    update_config: '修改配置',
    resource_pack: '资源包操作',
    change_password: '修改密码'
  }
  return names[op] || op
}

const getOperationTag = (op: string) => {
  const tags: Record<string, string> = {
    login: 'info',
    register: 'success',
    update_user_status: 'warning',
    delete_user: 'danger',
    recharge: 'success',
    update_config: 'warning',
    resource_pack: 'primary',
    change_password: 'info'
  }
  return tags[op] || 'info'
}

const getResourceTypeName = (type: string) => {
  const names: Record<string, string> = {
    user: '用户',
    config: '系统配置',
    resource_pack: '资源包',
    order: '订单',
    balance: '余额',
    system: '系统'
  }
  return names[type] || type || '-'
}
</script>

<style scoped>
.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.filter-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 20px;
}
</style>
