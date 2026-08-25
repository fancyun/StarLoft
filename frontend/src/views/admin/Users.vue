<template>
  <div class="admin-users">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>用户管理</span>
          <el-button type="primary" @click="showRegisterDialog">手动注册用户</el-button>
        </div>
      </template>

      <div class="filter-bar">
        <el-input
          v-model="filters.phone"
          placeholder="搜索手机号"
          style="width: 200px"
          clearable
          @clear="handleSearch"
        >
          <template #append>
            <el-button icon="Search" @click="handleSearch" />
          </template>
        </el-input>
        
        <el-select v-model="filters.kycStatus" placeholder="实名状态" clearable style="width: 150px" @change="handleSearch">
          <el-option label="全部" :value="null" />
          <el-option label="未实名" :value="0" />
          <el-option label="已实名" :value="1" />
        </el-select>
        
        <el-date-picker
          v-model="filters.dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          @change="handleSearch"
        />
      </div>

      <el-table :data="users" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="用户ID" width="80" />
        <el-table-column prop="phone" label="手机号" width="140" />
        <el-table-column prop="balance" label="余额" width="100">
          <template #default="{ row }">¥{{ row.balance }}</template>
        </el-table-column>
        <el-table-column prop="is_kyc_verified" label="实名状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_kyc_verified ? 'success' : 'info'">
              {{ row.is_kyc_verified ? '已实名' : '未实名' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="kyc_name" label="实名信息" width="120">
          <template #default="{ row }">
            {{ row.is_kyc_verified ? maskName(row.kyc_name) : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 1 ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="注册时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column prop="last_login_at" label="最后登录" width="180">
          <template #default="{ row }">{{ row.last_login_at ? formatDateTime(row.last_login_at) : '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" width="360">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewDetail(row)">详情</el-button>
            <el-button size="small" type="info" @click="handleViewFinance(row)">财务</el-button>
            <el-button size="small" type="primary" @click="handleRecharge(row)">充值</el-button>
            <el-button size="small" type="warning" @click="handleToggleStatus(row)">
              {{ row.status === 1 ? '禁用' : '启用' }}
            </el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="loadUsers"
        @size-change="loadUsers"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>

    <!-- 用户详情对话框 -->
    <el-dialog v-model="detailDialogVisible" title="用户详情" width="800px">
      <el-descriptions :column="2" border v-if="currentUser">
        <el-descriptions-item label="用户ID">{{ currentUser.id }}</el-descriptions-item>
        <el-descriptions-item label="手机号">{{ currentUser.phone }}</el-descriptions-item>
        <el-descriptions-item label="余额">¥{{ currentUser.balance }}</el-descriptions-item>
        <el-descriptions-item label="实名状态">
          {{ currentUser.is_kyc_verified ? '已实名' : '未实名' }}
        </el-descriptions-item>
        <el-descriptions-item label="实名姓名">
            {{ currentUser.is_kyc_verified && currentUser.kyc_name ? maskName(currentUser.kyc_name) : '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="实名身份证">
            {{ currentUser.is_kyc_verified && currentUser.kyc_id_card ? maskIdCard(currentUser.kyc_id_card) : '-' }}
          </el-descriptions-item>
        <el-descriptions-item label="API Key" :span="2">
          <el-text type="info" style="font-family: monospace">{{ currentUser.api_key }}</el-text>
        </el-descriptions-item>
        <el-descriptions-item label="注册时间">{{ formatDateTime(currentUser.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="最后登录">{{ formatDateTime(currentUser.last_login_at) }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <!-- 手动注册用户对话框 -->
    <el-dialog v-model="registerDialogVisible" title="手动注册用户" width="500px">
      <el-form :model="registerForm" :rules="registerRules" ref="registerFormRef" label-width="100px">
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="registerForm.phone" placeholder="请输入11位手机号" maxlength="11" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="registerForm.password" type="password" placeholder="请输入密码（至少6位）" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="registerDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleRegisterSubmit" :loading="registerLoading">确定注册</el-button>
      </template>
    </el-dialog>

    <!-- 人工充值对话框 -->
    <el-dialog v-model="rechargeDialogVisible" title="人工充值" width="500px">
      <el-form :model="rechargeForm" :rules="rechargeRules" ref="rechargeFormRef" label-width="100px">
        <el-form-item label="用户手机号">
          <el-input :value="rechargeForm.phone" disabled />
        </el-form-item>
        <el-form-item label="当前余额">
          <el-input :value="`¥${rechargeForm.currentBalance}`" disabled />
        </el-form-item>
        <el-form-item label="充值金额" prop="amount">
          <el-input-number 
            v-model="rechargeForm.amount" 
            :precision="2" 
            :step="10" 
            :min="0.01" 
            :max="10000"
            placeholder="请输入充值金额"
            style="width: 100%"
          />
          <div style="font-size: 12px; color: #909399; margin-top: 4px;">
            单位：元，最小0.01元，最大10000元
          </div>
        </el-form-item>
        <el-form-item label="银行流水单号" prop="bank_serial_no">
          <el-input 
            v-model="rechargeForm.bank_serial_no" 
            placeholder="请输入银行流水单号（必填，用于对账）"
            maxlength="100"
          />
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input 
            v-model="rechargeForm.remark" 
            type="textarea"
            :rows="3"
            placeholder="请输入充值备注（可选）"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rechargeDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleRechargeSubmit" :loading="rechargeLoading">确定充值</el-button>
      </template>
    </el-dialog>

    <!-- 财务详情对话框 -->
    <el-dialog v-model="financeDialogVisible" title="用户财务详情" width="900px">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="用户手机号">{{ financeUser.phone }}</el-descriptions-item>
        <el-descriptions-item label="当前余额">¥{{ financeUser.balance }}</el-descriptions-item>
        <el-descriptions-item label="总充值金额">¥{{ financeStats.totalRecharge }}</el-descriptions-item>
        <el-descriptions-item label="总消费金额">¥{{ financeStats.totalConsume }}</el-descriptions-item>
        <el-descriptions-item label="总退款金额">¥{{ financeStats.totalRefund }}</el-descriptions-item>
      </el-descriptions>

      <el-tabs v-model="financeActiveTab" style="margin-top: 20px">
        <el-tab-pane label="余额流水" name="balance">
          <el-table :data="balanceLogs" style="width: 100%" max-height="400">
            <el-table-column prop="created_at" label="时间" width="180">
              <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
            </el-table-column>
            <el-table-column prop="type" label="类型" width="100">
              <template #default="{ row }">
                <el-tag :type="getBalanceTypeTag(row.type)">{{ getBalanceTypeName(row.type) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="amount" label="金额" width="120">
              <template #default="{ row }">
                <span :style="{ color: isIncome(row.type) ? '#67C23A' : '#F56C6C' }">
                  {{ isIncome(row.type) ? '+' : '-' }}¥{{ row.amount }}
                </span>
              </template>
            </el-table-column>
            <el-table-column prop="balance_after" label="余额" width="120" />
            <el-table-column prop="bank_serial_no" label="银行流水单号" width="180">
              <template #default="{ row }">{{ row.bank_serial_no || '-' }}</template>
            </el-table-column>
            <el-table-column prop="remark" label="备注" />
          </el-table>
        </el-tab-pane>
        <el-tab-pane label="认证订单" name="orders">
          <el-table :data="authOrders" style="width: 100%" max-height="400">
            <el-table-column prop="platform_biz_no" label="订单号" width="200" />
            <el-table-column prop="cost" label="金额" width="100">
              <template #default="{ row }">
                ¥{{ row.cost }}
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getOrderStatusTag(row.status)">{{ getOrderStatusName(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="180">
              <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox, FormInstance, FormRules } from 'element-plus'
import { adminAPI } from '@/api'
import { formatDateTime } from '@/utils/format'

const loading = ref(false)
interface User {
  id: number
  phone: string
  balance: number
  is_kyc_verified: boolean
  kyc_name?: string
  kyc_id_card?: string
  api_key: string
  created_at: string
  last_login_at?: string
}

const users = ref<User[]>([])
const detailDialogVisible = ref(false)
const currentUser = ref<User | null>(null)
const registerDialogVisible = ref(false)
const registerLoading = ref(false)
const registerFormRef = ref<FormInstance>()

const registerForm = reactive({
  phone: '',
  password: ''
})

const registerRules: FormRules = {
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少6位', trigger: 'blur' }
  ]
}

const rechargeDialogVisible = ref(false)
const rechargeLoading = ref(false)
const rechargeFormRef = ref<FormInstance>()

const rechargeForm = reactive({
  userId: 0,
  phone: '',
  currentBalance: 0,
  amount: undefined as number | undefined,
  bank_serial_no: '',
  remark: ''
})

const rechargeRules: FormRules = {
  amount: [
    { required: true, message: '请输入充值金额', trigger: 'blur' },
    { type: 'number', min: 0.01, max: 10000, message: '充值金额范围为0.01-10000元', trigger: 'blur' }
  ],
  bank_serial_no: [
    { required: true, message: '请输入银行流水单号', trigger: 'blur' },
    { min: 4, message: '银行流水单号至少4个字符', trigger: 'blur' }
  ]
}

// 财务详情相关
const financeDialogVisible = ref(false)
const financeActiveTab = ref('balance')
const financeUser = reactive({
  phone: '',
  balance: 0
})
const financeStats = reactive({
  totalRecharge: 0,
  totalConsume: 0,
  totalRefund: 0
})
const balanceLogs = ref([])
const authOrders = ref([])

const filters = reactive({
  phone: '',
  kycStatus: null,
  dateRange: null
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

onMounted(() => {
  loadUsers()
})

const loadUsers = async () => {
  loading.value = true
  try {
    const response: any = await adminAPI.getUsers({
      page: pagination.page,
      page_size: pagination.pageSize,
      phone: filters.phone,
      kyc_status: filters.kycStatus,
      start_date: filters.dateRange?.[0],
      end_date: filters.dateRange?.[1]
    })
    // 响应拦截器已经返回了 data，所以直接使用 response.list
    users.value = response.list || []
    pagination.total = response.total || 0
  } catch (error: any) {
    console.error('加载用户列表失败:', error)
    ElMessage.error(error.response?.data?.message || '加载用户列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadUsers()
}

const showRegisterDialog = () => {
  registerForm.phone = ''
  registerForm.password = ''
  registerDialogVisible.value = true
  // 清空表单验证
  registerFormRef.value?.clearValidate()
}

const handleRegisterSubmit = async () => {
  if (!registerFormRef.value) return
  
  await registerFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    registerLoading.value = true
    try {
      const data: any = {
        phone: registerForm.phone,
        password: registerForm.password
      }
      
      const response: any = await adminAPI.registerUser(data)
      
      ElMessage.success('用户注册成功')
      registerDialogVisible.value = false
      
      // 显示注册结果信息
      ElMessageBox.alert(
        `<div style="line-height: 1.8;">
          <p><strong>手机号：</strong>${response.phone}</p>
          <p><strong>用户ID：</strong>${response.id}</p>
          <p><strong>API Key：</strong><br/><code style="word-break: break-all;">${response.api_key || '实名认证后自动生成'}</code></p>
        </div>`,
        '注册成功',
        {
          dangerouslyUseHTMLString: true,
          confirmButtonText: '确定'
        }
      )
      
      // 刷新用户列表
      loadUsers()
    } catch (error: any) {
      ElMessage.error(error.response?.data?.message || '注册失败')
    } finally {
      registerLoading.value = false
    }
  })
}

const handleViewDetail = (row: any) => {
  currentUser.value = row
  detailDialogVisible.value = true
}

const handleToggleStatus = async (row: any) => {
  const action = row.status === 1 ? '禁用' : '启用'
  try {
    await ElMessageBox.confirm(`确认${action}该用户？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await adminAPI.updateUserStatus(row.id, { status: row.status === 1 ? 0 : 1 })
    ElMessage.success(`${action}成功`)
    loadUsers()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(`${action}失败`)
    }
  }
}

const handleDelete = async (row: any) => {
  try {
    await ElMessageBox.confirm('确认删除该用户？此操作不可恢复！', '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error'
    })
    
    await adminAPI.deleteUser(row.id)
    ElMessage.success('删除成功')
    loadUsers()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const maskName = (name: string) => {
  if (!name || name.length === 0) return ''
  if (name.length === 1) return name
  if (name.length === 2) return name[0] + '*'
  return name[0] + '*'.repeat(name.length - 2) + name[name.length - 1]
}

const maskIdCard = (idCard: string) => {
  if (!idCard || idCard.length < 8) return idCard
  return idCard.substring(0, 3) + '***********' + idCard.substring(idCard.length - 4)
}

const handleRecharge = (row: any) => {
  rechargeForm.userId = row.id
  rechargeForm.phone = row.phone
  rechargeForm.currentBalance = row.balance
  rechargeForm.amount = undefined
  rechargeForm.bank_serial_no = ''
  rechargeForm.remark = ''
  rechargeDialogVisible.value = true
  // 清空表单验证
  rechargeFormRef.value?.clearValidate()
}

const handleRechargeSubmit = async () => {
  if (!rechargeFormRef.value) return
  
  await rechargeFormRef.value.validate(async (valid) => {
    if (!valid) return
    
    rechargeLoading.value = true
    try {
      const data = {
        amount: rechargeForm.amount!,
        bank_serial_no: rechargeForm.bank_serial_no.trim(),
        remark: rechargeForm.remark || undefined
      }
      
      await adminAPI.rechargeUser(rechargeForm.userId, data)
      
      ElMessage.success('充值成功')
      rechargeDialogVisible.value = false
      
      // 刷新用户列表
      loadUsers()
    } catch (error: any) {
      ElMessage.error(error.response?.data?.message || '充值失败')
    } finally {
      rechargeLoading.value = false
    }
  })
}

const handleViewFinance = async (row: any) => {
  financeUser.phone = row.phone
  financeUser.balance = row.balance
  
  financeDialogVisible.value = true
  financeActiveTab.value = 'balance'
  
  // 加载财务统计
  try {
    const stats: any = await adminAPI.getUserFinanceStats(row.id)
    Object.assign(financeStats, stats)
  } catch (error) {
    console.error('Failed to load finance stats:', error)
  }
  
  // 加载余额流水
  try {
    const logs: any = await adminAPI.getUserBalanceLogs(row.id)
    balanceLogs.value = logs.list || logs || []
  } catch (error) {
    console.error('Failed to load balance logs:', error)
    balanceLogs.value = []
  }
  
  // 加载认证订单
  try {
    const orders: any = await adminAPI.getUserAuthOrders(row.id)
    authOrders.value = orders.list || orders || []
  } catch (error) {
    console.error('Failed to load auth orders:', error)
    authOrders.value = []
  }
}

const getBalanceTypeName = (type: number | string) => {
  const names: Record<number, string> = {
    1: '充值',
    2: '消费',
    3: '退款'
  }
  return names[Number(type)] || String(type)
}

const getBalanceTypeTag = (type: number | string) => {
  const tags: Record<number, string> = {
    1: 'success',
    2: 'danger',
    3: 'warning'
  }
  return tags[Number(type)] || 'info'
}

// 增加余额的类型（充值/退款）显示为正数，消费显示为负数
const isIncome = (type: number | string) => [1, 3].includes(Number(type))

const getOrderStatusName = (status: number) => {
  const names: Record<number, string> = {
    0: '待支付',
    1: '已支付',
    2: '认证中',
    3: '已完成',
    4: '已取消',
    5: '已失败'
  }
  return names[status] || '未知'
}

const getOrderStatusTag = (status: number) => {
  const tags: Record<number, string> = {
    0: 'info',
    1: 'success',
    2: 'warning',
    3: 'success',
    4: 'info',
    5: 'danger'
  }
  return tags[status] || 'info'
}
</script>

<style scoped>
.admin-users {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}
</style>
