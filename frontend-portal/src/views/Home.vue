<template>
  <div class="home">
    <!-- Hero 区 -->
    <section class="hero">
      <div class="hero-visual" :style="{ backgroundImage: heroVisualImage }"></div>
      <div class="container hero-inner">
        <div class="hero-panel">
          <h1 class="hero-title">一站式云服务平台</h1>
          <p class="hero-sub">
            为企业提供弹性云服务器、实名认证、短信等安全可信的云产品能力，
            简单接入、稳定可靠、按需付费。
          </p>
          <div class="hero-actions">
            <a class="btn-primary" href="https://console.starloft.cn/register">免费注册</a>
            <router-link class="btn-ghost" to="/docs">了解产品服务</router-link>
          </div>
        </div>
      </div>
    </section>

    <!-- 产品服务 -->
    <section class="section">
      <div class="container">
        <div class="section-head">
          <h2>产品与服务</h2>
          <p>覆盖实名、计算、通信等基础云能力，持续扩充中</p>
        </div>
        <div class="product-grid">
          <router-link
            v-for="p in products"
            :key="p.key"
            :to="`/${p.key}`"
            class="product-card"
          >
            <span class="product-icon" v-html="p.icon"></span>
            <div class="product-info">
              <div class="product-name">
                {{ p.name }}
                <span v-if="p.status === 'coming-soon'" class="tag-coming">建设中</span>
              </div>
              <p class="product-desc">{{ p.description }}</p>
              <span class="product-more">了解详情 →</span>
            </div>
          </router-link>
        </div>
      </div>
    </section>

    <!-- 平台优势 -->
    <section class="section section-soft">
      <div class="container">
        <div class="section-head">
          <h2>平台优势</h2>
          <p>稳定可靠的基础设施，让企业专注业务本身</p>
        </div>
        <div class="advantage-grid">
          <div v-for="a in advantages" :key="a.title" class="advantage-card">
            <div class="advantage-icon" v-html="a.icon"></div>
            <h3>{{ a.title }}</h3>
            <p>{{ a.desc }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- CTA -->
    <section class="cta">
      <div class="container cta-inner">
        <h2>立即开始使用星楼云</h2>
        <p>注册账号，开通云服务，几分钟内完成对接</p>
        <a class="btn-primary" href="https://console.starloft.cn/register">注册</a>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { products } from '@/config/products'

// Hero 右侧品牌插画（Unsplash 免费直链，蓝色云主题，独立图层经蒙版从左侧平滑淡入）
const heroVisualImage = 'url("https://images.unsplash.com/photo-1667372283496-893f0b1e7c16?auto=format&fit=crop&w=1200&q=80")'

const advantages = [
  {
    title: '安全可靠',
    desc: 'HMAC-SHA256 签名鉴权，敏感数据加密存储，认证结果防伪造',
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>'
  },
  {
    title: '开放易用',
    desc: 'RESTful API 标准接口，配套 PHP 插件与完整文档，快速集成',
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>'
  },
  {
    title: '灵活计费',
    desc: '资源包 + 余额组合计费，先扣资源包、余额兜底，成本可控',
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="5" width="20" height="14" rx="2"/><line x1="2" y1="10" x2="22" y2="10"/></svg>'
  },
  {
    title: '高并发稳定',
    desc: 'Go 语言构建，Redis 缓存，支撑大规模并发认证请求',
    icon: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>'
  }
]
</script>

<style scoped>
.home {
  background: var(--bg-page);
}

/* ========== Hero ========== */
.hero {
  position: relative;
  overflow: hidden;
  background-color: #fff;
  color: var(--text-primary);
  padding: 72px 0;
  display: flex;
  align-items: center;
  min-height: 440px;
}

/* 右侧插画背景层：蒙版从左侧平滑淡入，避免出现竖直接缝 */
.hero-visual {
  position: absolute;
  top: 0;
  right: 0;
  height: 100%;
  width: 60%;
  background-position: center;
  background-size: cover;
  background-repeat: no-repeat;
  -webkit-mask-image: linear-gradient(90deg, transparent 0%, #000 40%);
  mask-image: linear-gradient(90deg, transparent 0%, #000 40%);
  pointer-events: none;
}

.hero-inner {
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
  display: flex;
  justify-content: flex-start;
}

/* 左侧内容白板：避免右侧背景图影响文字可读性 */
.hero-panel {
  max-width: 640px;
  background: rgba(255, 255, 255, 0.92);
  padding: 40px 44px;
  border-radius: 20px;
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.06);
}

.hero-title {
  color: var(--text-primary);
  font-size: 40px;
  font-weight: 700;
  letter-spacing: 1px;
  margin-bottom: 20px;
  text-align: left;
}

.hero-sub {
  color: var(--text-secondary);
  font-size: 18px;
  line-height: 1.7;
  margin-bottom: 36px;
  text-align: left;
}

.hero-actions {
  display: flex;
  justify-content: flex-start;
  gap: 16px;
}

/* 白板场景下按钮改为品牌蓝实心 / 线框，避免白底白字不可见 */
.hero-panel .btn-primary {
  background: var(--color-primary);
  color: #fff;
}
.hero-panel .btn-primary:hover {
  background: #0043b3;
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

/* ========== 产品卡片 ========== */
.product-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 24px;
}

.product-card {
  display: flex;
  align-items: flex-start;
  gap: 20px;
  padding: 28px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-xl);
  text-decoration: none;
  transition: all 0.15s;
}

.product-card:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
}

.product-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: var(--radius-lg);
  background: var(--color-primary-light);
  color: var(--color-primary);
  flex-shrink: 0;
}

.product-icon :deep(svg) {
  width: 28px;
  height: 28px;
}

.product-info {
  min-width: 0;
}

.product-name {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.product-desc {
  font-size: 14px;
  color: var(--text-secondary);
  line-height: 1.7;
  margin-bottom: 12px;
}

.product-more {
  font-size: 14px;
  color: var(--color-primary);
  font-weight: 500;
}

.tag-coming {
  display: inline-block;
  padding: 1px 8px;
  border-radius: var(--radius-sm);
  background: var(--bg-soft);
  color: var(--text-muted);
  font-size: 12px;
  font-weight: 500;
}

/* ========== 优势卡片 ========== */
.advantage-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 24px;
}

.advantage-card {
  padding: 32px 28px;
  background: var(--bg-card);
  border-radius: var(--radius-xl);
  text-align: center;
}

.advantage-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: var(--radius-lg);
  background: var(--color-primary-light);
  color: var(--color-primary);
  margin-bottom: 16px;
}

.advantage-icon :deep(svg) {
  width: 24px;
  height: 24px;
}

.advantage-card h3 {
  font-size: 17px;
  margin-bottom: 8px;
}

.advantage-card p {
  font-size: 14px;
  color: var(--text-muted);
  line-height: 1.7;
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
</style>
