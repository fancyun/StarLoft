<template>
  <div class="dashboard-container">
    <div class="content" v-loading="pageLoading">
      <!-- 状态卡片 -->
      <div class="status-cards">
        <div class="card">
          <div class="card-header">
            <el-icon class="icon"><User /></el-icon>
            <span>个人信息</span>
          </div>
          <div class="card-body info-card-body">
            <div class="info-list">
              <div class="info-row">
                <span class="info-label">手机号</span>
                <span class="info-value">{{ maskPhone(userInfo.phone) }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">邮箱</span>
                <span class="info-value">{{ userInfo.email || '未设置' }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">API 认证单价</span>
                <span class="info-value">¥{{ userInfo.kyc_price }}</span>
              </div>
              <div class="info-row">
                <span class="info-label">注册时间</span>
                <span class="info-value">{{ formatDate(userInfo.created_at) }}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="card">
          <div class="card-header">
            <el-icon class="icon"><Wallet /></el-icon>
            <span>账户余额</span>
          </div>
          <div class="card-body">
            <div class="amount">¥{{ userInfo.balance }}</div>
            <el-button type="primary" size="small" @click="$router.push('/user/balance')">充值</el-button>
          </div>
        </div>

        <div class="card">
          <div class="card-header">
            <el-icon class="icon"><Key /></el-icon>
            <span>API 密钥</span>
          </div>
          <div class="card-body">
            <div class="api-key">{{ maskAPIKey(userInfo.api_key) }}</div>
            <el-button size="small" @click="copyAPIKey">复制</el-button>
          </div>
        </div>
      </div>

      <!-- 认证调用趋势图表 -->
      <div class="card">
        <h3 class="section-title">
          <el-icon><TrendCharts /></el-icon>
          认证调用趋势（近30天）
        </h3>
        <div ref="callChartRef" style="height: 300px"></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import * as echarts from 'echarts'
import { userAPI } from '@/api'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()

const userInfo = ref<any>({})
const callChartRef = ref()
const pageLoading = ref(true)

let callChartInstance: ReturnType<typeof echarts.init> | null = null
let disposed = false

// 在 TS 里读取 CSS 变量，保证图表色和主题同步
const var_color_primary = computed(() =>
  getComputedStyle(document.documentElement).getPropertyValue('--color-primary').trim() || '#2563EB'
)

// ===== 工具函数 =====
const maskPhone = (phone: string) => {
  if (!phone) return ''
  return phone.substring(0, 3) + '****' + phone.substring(7)
}

const maskAPIKey = (key: string) => {
  if (!key) return '未设置'
  return key.substring(0, 8) + '...' + key.substring(key.length - 4)
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}/${m}/${day}`
}

const copyAPIKey = () => {
  navigator.clipboard.writeText(userInfo.value.api_key)
  ElMessage.success('API Key 已复制到剪贴板')
}

const loadData = async () => {
  try {
    const profile = await userAPI.getProfile()
    userInfo.value = profile
    userStore.setUserInfo(profile)
  } catch (error) {
    console.error(error)
  } finally {
    pageLoading.value = false
  }
}

const loadCallsChart = async () => {
  try {
    const stats = await userAPI.getCallStats()

    // 组件已卸载，则跳过渲染，避免操作已销毁的 DOM
    if (disposed) return
    if (!callChartRef.value) return

    callChartInstance = echarts.init(callChartRef.value)
    callChartInstance.setOption({
      color: [var_color_primary.value],
      xAxis: { type: 'category', data: stats.dates, axisLine: { lineStyle: { color: '#E5E7EB' } } },
      yAxis: { type: 'value', minInterval: 1, splitLine: { lineStyle: { color: '#F3F4F6' } } },
      grid: { left: 40, right: 20, top: 20, bottom: 30 },
      series: [{
        data: stats.counts, type: 'line', smooth: true, symbolSize: 6,
        lineStyle: { width: 3 },
        areaStyle: {
          color: {
            type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(37, 99, 235, 0.25)' },
              { offset: 1, color: 'rgba(37, 99, 235, 0.02)' }
            ]
          }
        }
      }],
      tooltip: { trigger: 'axis' }
    })
  } catch (error: any) {
    if (disposed) return
    console.error('加载认证调用统计失败:', error)
  }
}

onMounted(() => {
  loadData()
  loadCallsChart()
})

onBeforeUnmount(() => {
  disposed = true
  callChartInstance?.dispose()
})
</script>

<style scoped>
.dashboard-container {
  min-height: 100%;
}

.content {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* ===== 状态卡片 ===== */
.status-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 20px;
}

.card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 24px;
  box-shadow: var(--shadow-sm);
  transition: box-shadow 0.15s;
}
.card:hover { box-shadow: var(--shadow-md); }

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  color: var(--text-primary);
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-light);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  color: var(--text-muted);
  font-size: 14px;
}

.card-header .icon {
  width: 36px; height: 36px;
  border-radius: 10px;
  background: var(--color-primary-light);
  color: var(--color-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
}

.card-body {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.amount {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
}

/* ===== 个人信息卡片 ===== */
.info-card-body {
  flex-direction: column;
  align-items: stretch;
}

.info-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 0;
}

.info-label {
  color: var(--text-muted);
  font-size: 13px;
}

.info-value {
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 500;
}

.api-key {
  font-family: 'Courier New', monospace;
  color: var(--text-secondary);
  font-size: 14px;
}
</style>