<template>
  <div class="admin-orders">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>订单管理</span>
        </div>
      </template>

      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="认证订单" name="auth">
          <div class="filter-bar">
            <el-input
              v-model="authFilters.platformBizNo"
              placeholder="平台流水号"
              style="width: 200px"
              clearable
            />
            <el-select v-model="authFilters.status" placeholder="订单状态" clearable style="width: 150px">
              <el-option label="待认证" :value="0" />
              <el-option label="认证中" :value="1" />
              <el-option label="已完成" :value="2" />
              <el-option label="失败" :value="3" />
              <el-option label="已退款" :value="5" />
            </el-select>
            <el-date-picker
              v-model="authFilters.dateRange"
              type="daterange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
            />
            <el-button type="primary" icon="Search" @click="loadAuthOrders">搜索</el-button>
          </div>

          <el-table :data="authOrders" style="width: 100%" v-loading="loading">
            <el-table-column prop="platform_biz_no" label="平台流水号" width="180" />
            <el-table-column prop="biz_no" label="业务流水号" width="180" />
            <el-table-column prop="user_phone" label="用户手机号" width="140" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getAuthStatusType(row.status)">{{ getAuthStatusText(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="cost" label="金额" width="100">
              <template #default="{ row }">¥{{ row.cost }}</template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="180" />
            <el-table-column label="操作" fixed="right" width="120">
              <template #default="{ row }">
                <el-button size="small" @click="viewAuthDetail(row)">详情</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="authPagination.page"
            v-model:page-size="authPagination.pageSize"
            :total="authPagination.total"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            @current-change="loadAuthOrders"
            @size-change="loadAuthOrders"
            style="margin-top: 20px; justify-content: flex-end"
          />
        </el-tab-pane>

        <el-tab-pane label="支付订单" name="payment">
          <div class="filter-bar">
            <el-input
              v-model="paymentFilters.payOrderNo"
              placeholder="支付流水号"
              style="width: 200px"
              clearable
            />
            <el-select v-model="paymentFilters.channel" placeholder="支付渠道" clearable style="width: 150px">
              <el-option label="支付宝" value="alipay" />
              <el-option label="微信支付" value="wechat" />
              <el-option label="云闪付" value="unionpay" />
            </el-select>
            <el-select v-model="paymentFilters.status" placeholder="订单状态" clearable style="width: 150px">
              <el-option label="待支付" :value="0" />
              <el-option label="已支付" :value="1" />
              <el-option label="已退款" :value="2" />
              <el-option label="已关闭" :value="3" />
            </el-select>
            <el-button type="primary" icon="Search" @click="loadPaymentOrders">搜索</el-button>
          </div>

          <el-table :data="paymentOrders" style="width: 100%" v-loading="loading">
            <el-table-column prop="pay_order_no" label="支付流水号" width="180" />
            <el-table-column prop="user_phone" label="用户手机号" width="140" />
            <el-table-column prop="channel" label="支付渠道" width="100">
              <template #default="{ row }">
                <el-tag>{{ getChannelText(row.channel) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="amount" label="充值金额" width="100">
              <template #default="{ row }">¥{{ row.amount }}</template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getPayStatusType(row.status)">{{ getPayStatusText(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="paid_at" label="支付时间" width="180">
              <template #default="{ row }">{{ row.paid_at || '-' }}</template>
            </el-table-column>
            <el-table-column label="操作" fixed="right" width="120">
              <template #default="{ row }">
                <el-button size="small" @click="viewPaymentDetail(row)">详情</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="paymentPagination.page"
            v-model:page-size="paymentPagination.pageSize"
            :total="paymentPagination.total"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            @current-change="loadPaymentOrders"
            @size-change="loadPaymentOrders"
            style="margin-top: 20px; justify-content: flex-end"
          />
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { adminAPI } from '@/api'

const loading = ref(false)
const activeTab = ref('auth')
const authOrders = ref([])
const paymentOrders = ref([])

const authFilters = reactive({
  platformBizNo: '',
  status: null,
  dateRange: null
})

const paymentFilters = reactive({
  payOrderNo: '',
  channel: '',
  status: null
})

const authPagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const paymentPagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

onMounted(() => {
  loadAuthOrders()
})

const handleTabChange = (tab: string) => {
  if (tab === 'auth') {
    loadAuthOrders()
  } else {
    loadPaymentOrders()
  }
}

const loadAuthOrders = async () => {
  loading.value = true
  try {
    const response: any = await adminAPI.getOrders({
      page: authPagination.page,
      page_size: authPagination.pageSize,
      platform_biz_no: authFilters.platformBizNo,
      status: authFilters.status,
      start_date: authFilters.dateRange?.[0],
      end_date: authFilters.dateRange?.[1]
    })
    authOrders.value = response.orders || response.list || []
    authPagination.total = response.total || 0
  } catch (error: any) {
    ElMessage.error('加载认证订单失败')
  } finally {
    loading.value = false
  }
}

const loadPaymentOrders = async () => {
  loading.value = true
  try {
    const response: any = await adminAPI.getPayments({
      page: paymentPagination.page,
      page_size: paymentPagination.pageSize,
      pay_order_no: paymentFilters.payOrderNo,
      channel: paymentFilters.channel,
      status: paymentFilters.status
    })
    paymentOrders.value = response.orders || response.list || []
    paymentPagination.total = response.total || 0
  } catch (error: any) {
    ElMessage.error('加载支付订单失败')
  } finally {
    loading.value = false
  }
}

// 姓名脱敏函数（保留供将来使用）
// const _maskName = (name: string) => {
//   if (!name || name.length === 0) return ''
//   if (name.length === 1) return name
//   if (name.length === 2) return name[0] + '*'
//   return name[0] + '*'.repeat(name.length - 2) + name[name.length - 1]
// }

const getAuthStatusType = (status: number) => {
  const map: Record<number, string> = { 0: 'info', 1: 'warning', 2: 'success', 3: 'danger', 5: 'info' }
  return map[status] || 'info'
}

const getAuthStatusText = (status: number) => {
  const map: Record<number, string> = { 0: '待认证', 1: '认证中', 2: '已完成', 3: '失败', 5: '已退款' }
  return map[status] || '未知'
}

const getChannelText = (channel: string) => {
  const map: Record<string, string> = { alipay: '支付宝', wechat: '微信支付', unionpay: '云闪付' }
  return map[channel] || channel
}

const getPayStatusType = (status: number) => {
  const map: Record<number, string> = { 0: 'warning', 1: 'success', 2: 'info', 3: 'info' }
  return map[status] || 'info'
}

const getPayStatusText = (status: number) => {
  const map: Record<number, string> = { 0: '待支付', 1: '已支付', 2: '已退款', 3: '已关闭' }
  return map[status] || '未知'
}

const viewAuthDetail = (_row: any) => {
  ElMessage.info('查看认证订单详情')
}

const viewPaymentDetail = (_row: any) => {
  ElMessage.info('查看支付订单详情')
}
</script>

<style scoped>
.admin-orders {
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
