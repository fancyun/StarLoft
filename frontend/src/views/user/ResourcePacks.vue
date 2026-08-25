<template>
  <div class="packs-container">
    <div class="content" v-loading="pageLoading">
      <!-- 余额概览 -->
      <div class="card">
        <h3 class="section-title">
          <el-icon><Wallet /></el-icon>
          资源包
        </h3>
        <div class="balance-display">
          <span class="balance-label">当前余额</span>
          <span class="balance-amount">¥{{ userInfo.balance }}</span>
          <span class="balance-tip">购买资源包将直接从余额扣费，请先充值再购买</span>
        </div>
      </div>

      <!-- 我的资源包 -->
      <div class="card" v-if="myPacks.length">
        <h3 class="section-title">
          <el-icon><Box /></el-icon>
          我的资源包
        </h3>
        <div class="my-pack-list">
          <div
            v-for="pack in myPacks"
            :key="pack.id"
            class="my-pack-item"
            :class="{ empty: pack.remaining_count <= 0 }"
          >
            <div class="my-pack-info">
              <div class="my-pack-name">{{ pack.pack_name }}</div>
              <div class="my-pack-count">
                剩余 {{ pack.remaining_count }} / {{ pack.total_count }} 次
              </div>
            </div>
            <el-tag :type="pack.remaining_count > 0 ? 'success' : 'info'">
              {{ pack.remaining_count > 0 ? '可用' : '已用完' }}
            </el-tag>
          </div>
        </div>
      </div>

      <!-- 在售资源包 -->
      <div class="card" v-if="packs.length">
        <h3 class="section-title">
          <el-icon><ShoppingCart /></el-icon>
          在售资源包
        </h3>
        <div class="pack-grid">
          <div v-for="pack in packs" :key="pack.id" class="pack-card">
            <div class="pack-header">
              <div class="pack-name">{{ pack.name }}</div>
              <el-tag v-if="pack.stock === -1" type="info" size="small">不限量</el-tag>
              <el-tag v-else type="warning" size="small">库存 {{ pack.stock }}</el-tag>
            </div>
            <div class="pack-count">{{ pack.total_count }} 次认证</div>
            <div class="pack-desc">{{ pack.description || 'KYC 实名认证次数' }}</div>
            <div class="pack-footer">
              <span class="pack-price">¥{{ pack.price }}</span>
              <el-button
                type="primary"
                size="small"
                :loading="purchasingId === pack.id"
                @click="handlePurchase(pack)"
              >
                购买
              </el-button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Wallet, Box, ShoppingCart } from '@element-plus/icons-vue'
import { userAPI } from '@/api'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const userInfo = ref<any>({})
const packs = ref<any[]>([])
const myPacks = ref<any[]>([])
const purchasingId = ref<number | null>(null)
const pageLoading = ref(true)

const loadData = async () => {
  try {
    const [profile, onSale, mine] = await Promise.all([
      userAPI.getProfile(),
      userAPI.listPacks(),
      userAPI.myPacks()
    ])
    userInfo.value = profile
    userStore.setUserInfo(profile)
    packs.value = onSale.list || []
    myPacks.value = mine.list || []
  } catch (error) {
    console.error(error)
  } finally {
    pageLoading.value = false
  }
}

const handlePurchase = async (pack: any) => {
  purchasingId.value = pack.id
  try {
    await userAPI.purchasePack(pack.id)
    ElMessage.success('购买成功')
    await loadData()
  } catch (error: any) {
    const msg = error?.response?.data?.message || '购买失败'
    ElMessage.error(msg)
  } finally {
    purchasingId.value = null
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.packs-container {
  min-height: 100%;
}

.content {
  max-width: 900px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 24px;
  box-shadow: var(--shadow-sm);
}

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

.balance-display {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 24px;
  background: linear-gradient(135deg, var(--color-primary-light) 0%, #E0E7FF 100%);
  border-radius: var(--radius-md);
}

.balance-label {
  color: var(--text-secondary);
  font-size: 14px;
  margin-bottom: 8px;
}

.balance-amount {
  color: var(--text-primary);
  font-size: 32px;
  font-weight: 700;
}

.balance-tip {
  margin-top: 8px;
  color: var(--text-muted);
  font-size: 12px;
}

/* 我的资源包 */
.my-pack-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.my-pack-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
  background: var(--bg-page);
}

.my-pack-item.empty {
  opacity: 0.6;
}

.my-pack-name {
  font-weight: 600;
  color: var(--text-primary);
}

.my-pack-count {
  font-size: 13px;
  color: var(--text-secondary);
  margin-top: 4px;
}

/* 在售资源包 */
.pack-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 16px;
}

.pack-card {
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
  padding: 20px;
  background: var(--bg-page);
  display: flex;
  flex-direction: column;
  gap: 8px;
  transition: box-shadow 0.15s;
}

.pack-card:hover {
  box-shadow: var(--shadow-md);
}

.pack-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.pack-name {
  font-weight: 700;
  color: var(--text-primary);
  font-size: 16px;
}

.pack-count {
  font-size: 20px;
  font-weight: 700;
  color: var(--color-primary);
}

.pack-desc {
  font-size: 13px;
  color: var(--text-secondary);
  min-height: 36px;
}

.pack-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: auto;
  padding-top: 8px;
  border-top: 1px solid var(--border-light);
}

.pack-price {
  font-size: 18px;
  font-weight: 700;
  color: #F56C6C;
}
</style>
