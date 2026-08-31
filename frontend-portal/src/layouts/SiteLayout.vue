<template>
  <div class="site-layout">
    <header class="site-header">
      <div class="container header-inner">
        <router-link to="/" class="brand">
          <span class="brand-logo">SL</span>
          <span class="brand-text">星楼云</span>
          <span class="brand-sub">StarLoft</span>
        </router-link>

        <nav class="site-nav">
          <router-link to="/" class="nav-item" :class="{ active: $route.path === '/' }">首页</router-link>
          <router-link to="/kyc" class="nav-item" :class="{ active: $route.path === '/kyc' }">实名认证</router-link>
          <router-link to="/cs" class="nav-item" :class="{ active: $route.path === '/cs' }">云服务器</router-link>
          <router-link to="/sms" class="nav-item" :class="{ active: $route.path === '/sms' }">短信服务</router-link>
          <router-link to="/docs" class="nav-item" :class="{ active: $route.path.startsWith('/docs') }">文档中心</router-link>
        </nav>

        <div class="header-actions">
          <a class="action-link" href="https://console.starloft.cn/login">登录</a>
          <a class="action-btn" href="https://console.starloft.cn/register">立即注册</a>
        </div>
      </div>
    </header>

    <main class="site-main">
      <router-view />
    </main>

    <footer class="site-footer">
      <div class="container footer-inner">
        <div class="footer-cols">
          <div class="footer-col">
            <div class="footer-brand">
              <span class="brand-logo">SL</span>
              <span>星楼云</span>
            </div>
            <p class="footer-desc">一站式云服务平台，为企业提供安全、稳定、易用的云产品能力。</p>
          </div>
          <div class="footer-col">
            <div class="footer-title">产品服务</div>
            <router-link to="/kyc" class="footer-link">实名认证</router-link>
            <router-link to="/cs" class="footer-link">云服务器</router-link>
            <router-link to="/sms" class="footer-link">短信服务</router-link>
          </div>
          <div class="footer-col">
            <div class="footer-title">开发文档</div>
            <router-link to="/docs" class="footer-link">文档总览</router-link>
            <router-link to="/docs/api/v1" class="footer-link">API v1 文档</router-link>
            <router-link to="/docs/plugin" class="footer-link">插件教程</router-link>
          </div>
          <div class="footer-col">
            <div class="footer-title">资源入口</div>
            <a class="footer-link" href="https://console.starloft.cn">云控制台</a>
            <a class="footer-link" href="https://console.starloft.cn/login">控制台登录</a>
            <a class="footer-link" href="https://console.starloft.cn/register">账号注册</a>
          </div>
          <div class="footer-col">
            <div class="footer-title">联系我们</div>
            <a class="footer-link" href="mailto:support@starloft.tech">support@starloft.tech</a>
          </div>
        </div>
        <div class="footer-bottom">
          <span>© 2026 星楼云 StarLoft · 保留所有权利</span>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts"></script>

<style scoped>
.site-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 24px;
}

/* ========== 顶部导航（白底 + 下边框） ========== */
.site-header {
  position: sticky;
  top: 0;
  z-index: 100;
  background: var(--bg-card);
  border-bottom: 1px solid var(--border-color);
}

.header-inner {
  display: flex;
  align-items: center;
  height: 60px;
  gap: 32px;
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
  color: var(--text-primary);
}

.brand-text {
  font-size: 18px;
  font-weight: 700;
  letter-spacing: 0.5px;
}

.brand-sub {
  font-size: 12px;
  color: var(--text-muted);
  letter-spacing: 0.5px;
  margin-top: 8px;
}

.site-nav {
  display: flex;
  align-items: center;
  gap: 4px;
  flex: 1;
}

.nav-item {
  padding: 8px 14px;
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: 14px;
  font-weight: 500;
  text-decoration: none;
  transition: all 0.15s;
}

.nav-item:hover {
  color: var(--text-primary);
  background: var(--bg-hover);
}

.nav-item.active {
  color: var(--color-primary);
  background: var(--bg-active);
  font-weight: 600;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.action-link {
  color: var(--text-secondary);
  font-size: 14px;
  text-decoration: none;
}

.action-link:hover {
  color: var(--color-primary);
}

.action-btn {
  padding: 8px 18px;
  border-radius: var(--radius-md);
  background: var(--color-primary);
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  text-decoration: none;
  transition: background 0.15s;
}

.action-btn:hover {
  background: var(--color-primary-hover);
  color: #fff;
}

/* ========== 主体 ========== */
.site-main {
  flex: 1;
}

/* ========== 页脚 ========== */
.site-footer {
  background: var(--bg-card);
  border-top: 1px solid var(--border-color);
  padding: 40px 0 0;
}

.footer-inner {
  padding-bottom: 24px;
}

.footer-cols {
  display: grid;
  grid-template-columns: 2fr 1fr 1fr 1fr 1fr;
  gap: 32px;
}

.footer-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 12px;
}

.footer-desc {
  font-size: 13px;
  color: var(--text-muted);
  line-height: 1.7;
}

.footer-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 12px;
}

.footer-link {
  display: block;
  padding: 4px 0;
  font-size: 13px;
  color: var(--text-secondary);
  text-decoration: none;
  transition: color 0.15s;
}

.footer-link:hover {
  color: var(--color-primary);
}

.footer-bottom {
  margin-top: 32px;
  padding-top: 16px;
  border-top: 1px solid var(--border-light);
  text-align: center;
  font-size: 12px;
  color: var(--text-muted);
}

@media (max-width: 900px) {
  .footer-cols {
    grid-template-columns: 1fr 1fr;
  }
  .brand-sub {
    display: none;
  }
  .header-actions {
    gap: 8px;
  }
  .nav-item {
    padding: 8px 8px;
  }
}
</style>
