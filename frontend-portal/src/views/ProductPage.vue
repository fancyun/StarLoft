<template>
  <div class="product-page">
    <div v-if="product">
      <!-- 产品 Hero -->
      <section class="hero" :style="heroBg">
        <div class="container hero-inner">
          <div class="hero-panel">
            <div class="hero-icon" v-html="product.icon"></div>
            <h1 class="hero-title">{{ product.name }}</h1>
            <p class="hero-english">{{ product.english }}</p>
            <p class="hero-tagline">{{ product.tagline }}</p>
            <p class="hero-desc">{{ product.description }}</p>
            <div class="hero-actions">
              <a
                v-if="product.status === 'available'"
                class="btn-primary"
                :href="product.consolePath"
              >立即使用</a>
              <span v-else class="btn-primary btn-disabled">即将上线</span>
              <router-link class="btn-ghost" to="/docs/api/v1">查看 API 文档</router-link>
            </div>
          </div>
        </div>
      </section>

      <!-- 产品特性 -->
      <section class="section">
        <div class="container">
          <div class="section-head">
            <h2>产品特性</h2>
            <p>核心能力一览</p>
          </div>
          <div class="feature-grid">
            <div v-for="f in product.features" :key="f.title" class="feature-card">
              <h3>{{ f.title }}</h3>
              <p>{{ f.desc }}</p>
            </div>
          </div>
        </div>
      </section>

      <!-- 典型场景 -->
      <section class="section section-soft">
        <div class="container">
          <div class="section-head">
            <h2>典型场景</h2>
            <p>适用于多种业务形态</p>
          </div>
          <div class="scenario-list">
            <span v-for="s in product.scenarios" :key="s" class="scenario-item">{{ s }}</span>
          </div>
        </div>
      </section>

      <!-- CTA -->
      <section class="cta">
        <div class="container cta-inner">
          <h2>开始使用{{ product.name }}</h2>
          <p>注册账号即可体验，按需购买、灵活计费</p>
          <a class="btn-primary" href="https://console.starloft.cn/register">免费注册</a>
        </div>
      </section>
    </div>
    <div v-else class="container not-found">
      <h2>产品不存在</h2>
      <router-link to="/">返回首页</router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { productByKey } from '@/config/products'

const route = useRoute()
const product = computed(() => productByKey(String(route.meta.product)))

// Hero 背景：按产品展示不同插画，左侧白色渐变留白 + 白板文字，右侧铺图（阿里云/腾讯云风格）
const heroImageByKey: Record<string, string> = {
  kyc: 'identity+verification+face+recognition+illustration%2C+blue+abstract%2C+secure+badge%2C+minimal%2C+clean+white+background%2C+right+aligned',
  cs: 'cloud+server+data+center+illustration%2C+blue+gradient%2C+minimal%2C+clean+white+background%2C+right+aligned',
  sms: 'message+and+notification+communication+illustration%2C+blue+gradient%2C+minimal%2C+clean+white+background%2C+right+aligned'
}
const IMAGE_ENDPOINT = 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image'
const heroBg = computed(() => {
  const prompt = heroImageByKey[String(route.meta.product)] || heroImageByKey.kyc
  return {
    backgroundImage: `linear-gradient(90deg, #ffffff 0%, rgba(255,255,255,0.94) 34%, rgba(255,255,255,0) 60%), url("${IMAGE_ENDPOINT}?prompt=${prompt}&image_size=landscape_16_9")`,
    backgroundSize: '100% 100%, 60% auto',
    backgroundPosition: '0 0, right center',
    backgroundRepeat: 'no-repeat'
  }
})
</script>

<style scoped>
.product-page {
  background: var(--bg-page);
}

/* ========== Hero ========== */
.hero {
  background-color: #fff;
  color: var(--text-primary);
  padding: 72px 0;
  display: flex;
  align-items: center;
  min-height: 440px;
}

.hero-inner {
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
  display: flex;
  justify-content: flex-start;
  text-align: left;
}

.hero-panel {
  max-width: 640px;
  background: rgba(255, 255, 255, 0.92);
  padding: 40px 44px;
  border-radius: 20px;
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.06);
}

.hero-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 72px;
  height: 72px;
  border-radius: 16px;
  background: var(--color-primary-light);
  color: var(--color-primary);
  margin-bottom: 20px;
}

.hero-icon :deep(svg) {
  width: 36px;
  height: 36px;
}

.hero-title {
  color: var(--text-primary);
  font-size: 36px;
  font-weight: 700;
  margin-bottom: 4px;
}

.hero-english {
  color: var(--text-muted);
  font-size: 14px;
  letter-spacing: 1px;
  margin-bottom: 16px;
}

.hero-tagline {
  color: var(--text-primary);
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 16px;
}

.hero-desc {
  color: var(--text-secondary);
  font-size: 15px;
  line-height: 1.8;
  max-width: 600px;
  margin-bottom: 36px;
}

.hero-actions {
  display: flex;
  justify-content: flex-start;
  gap: 16px;
}

/* 白板场景下按钮改为品牌蓝实心 / 线框 */
.hero-panel .btn-primary {
  background: var(--color-primary);
  color: #fff;
}
.hero-panel .btn-primary:hover {
  background: #0043b3;
}
.hero-panel .btn-primary.btn-disabled {
  background: var(--bg-soft);
  color: var(--text-muted);
}
.hero-panel .btn-ghost {
  border-color: var(--color-primary);
  color: var(--color-primary);
}
.hero-panel .btn-ghost:hover {
  background: var(--color-primary-light);
  color: var(--color-primary);
}

.btn-primary {
  display: inline-block;
  padding: 12px 32px;
  border-radius: var(--radius-md);
  background: #fff;
  color: var(--color-primary);
  font-size: 16px;
  font-weight: 600;
  text-decoration: none;
  transition: all 0.15s;
}

.btn-primary:hover {
  background: var(--bg-soft);
  color: var(--color-primary-active);
}

.btn-disabled {
  cursor: not-allowed;
  opacity: 0.85;
}

.btn-ghost {
  display: inline-block;
  padding: 12px 32px;
  border-radius: var(--radius-md);
  border: 1px solid rgba(255, 255, 255, 0.7);
  color: #fff;
  font-size: 16px;
  font-weight: 500;
  text-decoration: none;
  transition: all 0.15s;
}

.btn-ghost:hover {
  background: rgba(255, 255, 255, 0.15);
  color: #fff;
}

/* ========== 通用区块 ========== */
.section {
  padding: 72px 0;
}

.section-soft {
  background: var(--bg-soft);
}

.section-head {
  text-align: center;
  margin-bottom: 48px;
}

.section-head h2 {
  font-size: 28px;
  margin-bottom: 8px;
}

.section-head p {
  color: var(--text-muted);
  font-size: 15px;
}

/* ========== 特性卡片 ========== */
.feature-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 24px;
}

.feature-card {
  padding: 28px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-xl);
  transition: all 0.15s;
}

.feature-card:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-md);
}

.feature-card h3 {
  font-size: 17px;
  margin-bottom: 10px;
  color: var(--color-primary);
}

.feature-card p {
  font-size: 14px;
  color: var(--text-secondary);
  line-height: 1.7;
}

/* ========== 场景 ========== */
.scenario-list {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 16px;
}

.scenario-item {
  padding: 12px 28px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 999px;
  color: var(--text-secondary);
  font-size: 15px;
  font-weight: 500;
}

/* ========== CTA ========== */
.cta {
  background: var(--bg-card);
  border-top: 1px solid var(--border-color);
  padding: 64px 0;
}

.cta-inner {
  text-align: center;
}

.cta-inner h2 {
  font-size: 28px;
  margin-bottom: 8px;
}

.cta-inner p {
  color: var(--text-muted);
  font-size: 15px;
  margin-bottom: 32px;
}

/* ========== 兜底 ========== */
.not-found {
  padding: 120px 0;
  text-align: center;
}

.not-found h2 {
  margin-bottom: 16px;
}
</style>
