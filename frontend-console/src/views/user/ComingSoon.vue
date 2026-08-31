<template>
  <div class="coming-soon">
    <div class="coming-soon-icon">🚧</div>
    <h2>{{ productName }}正在建设中</h2>
    <p>该产品功能即将上线，敬请期待</p>
    <el-button type="primary" @click="$router.push('/dashboard')">返回控制台首页</el-button>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const productName = computed(() => (route.meta.productName as string) || '产品')
</script>

<style scoped>
.coming-soon {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 120px 0;
  text-align: center;
}

.coming-soon-icon {
  font-size: 56px;
}

.coming-soon h2 {
  margin: 0;
  color: var(--text-primary);
  font-size: 22px;
}

.coming-soon p {
  margin: 0 0 8px;
  color: var(--text-muted);
  font-size: 14px;
}
</style>
