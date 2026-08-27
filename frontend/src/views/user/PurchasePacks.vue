<template>
  <div class="packs-container">
    <div class="content" v-loading="pageLoading">
      <div class="card">
        <div class="card-head">
          <h3 class="section-title">
            <el-icon><ShoppingCart /></el-icon>
            购买资源包
          </h3>
          <el-button @click="$router.back()">返回</el-button>
        </div>
        <p class="purchase-tip">购买资源包将直接从余额扣费（请先充值再购买），购买后次数实时到账。</p>
        <div v-if="packs.length" class="pack-grid">
          <div v-for="pack in packs" :key="pack.id" class="pack-card">
            <div class="pack-header">
              <div class="pack-name">{{ pack.name }}</div>
              <el-tag v-if="pack.stock === -1" type="info" size="small">不限量</el-tag>
              <el-tag
                v-else
                :type="pack.stock > 0 ? 'warning' : 'danger'"
                size="small"
              >
                {{ pack.stock > 0 ? `库存 ${pack.stock}` : '已售罄' }}
              </el-tag>
            </div>
            <div class="pack-count">{{ pack.total_count }} 次认证</div>
            <div class="pack-desc">{{ pack.description || 'KYC 实名认证次数' }}</div>
            <div class="pack-footer">
              <span class="pack-price">¥{{ pack.price }}</span>
              <el-button
                type="primary"
                size="small"
                :disabled="pack.stock !== -1 && pack.stock <= 0"
                :loading="purchasingId === pack.id"
                @click="handlePurchase(pack)"
              >
                购买
              </el-button>
            </div>
          </div>
        </div>
        <el-empty v-else description="暂无可购买的资源包" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ShoppingCart } from '@element-plus/icons-vue'
import { userAPI } from '@/api'

const packs = ref<any[]>([])
const purchasingId = ref<number | null>(null)
const pageLoading = ref(true)

const loadPacks = async () => {
  try {
    const onSale = await userAPI.listPacks()
    packs.value = onSale.list || []
  } catch (error) {
    console.error(error)
  } finally {
    pageLoading.value = false
  }
}

const handlePurchase = (pack: any) => {
  ElMessageBox.confirm(
    `确认使用余额 ¥${pack.price} 购买「${pack.name}」？购买后将获得 ${pack.total_count} 次认证次数。`,
    '确认购买',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  )
    .then(async () => {
      purchasingId.value = pack.id
      try {
        await userAPI.purchasePack(pack.id)
        ElMessage.success('购买成功')
        await loadPacks()
      } catch (error: any) {
        const msg = error?.response?.data?.message || '购买失败'
        ElMessage.error(msg)
      } finally {
        purchasingId.value = null
      }
    })
    .catch(() => {})
}

onMounted(() => {
  loadPacks()
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

.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
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

.card-head .section-title {
  margin-bottom: 0;
  padding-bottom: 0;
  border-bottom: none;
}

.purchase-tip {
  font-size: 13px;
  color: var(--text-secondary);
  margin: 0 0 20px;
  padding: 10px 14px;
  background: var(--bg-page);
  border: 1px solid var(--border-light);
  border-radius: var(--radius-md);
}

/* 在售资源包：一行 3 个 */
.pack-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
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
  color: var(--color-danger);
}
</style>
