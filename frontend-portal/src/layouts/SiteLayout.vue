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
          <div class="nav-dropdown" :class="{ open: menuOpen }">
            <span class="nav-item" @click="toggleMenu" :class="{ active: ['/kyc','/cs','/sms'].some(p => $route.path.startsWith(p)) }">
              产品
              <i class="dropdown-caret"></i>
            </span>
            <div class="dropdown-menu" :class="{ 'menu-open': menuOpen }">
              <router-link to="/kyc" class="dropdown-item" @click="closeMenu">实名认证</router-link>
              <router-link to="/cs" class="dropdown-item" @click="closeMenu">云服务器</router-link>
              <router-link to="/sms" class="dropdown-item" @click="closeMenu">短信服务</router-link>
            </div>
          </div>
          <router-link to="/docs" class="nav-item" :class="{ active: $route.path.startsWith('/docs') }">文档中心</router-link>
        </nav>

        <div class="header-actions">
          <a class="action-btn action-btn-plain" href="https://console.starloft.cn/login">登录</a>
          <a class="action-btn" href="https://console.starloft.cn/register">注册</a>
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
            <div class="footer-title">公司信息</div>
            <span class="footer-link">上海星楼网络科技有限公司</span>
            <span class="footer-link">上海市崇明区东平镇东冉路547号</span>
          </div>
          <div class="footer-col">
            <div class="footer-title">联系我们</div>
            <a class="footer-link" href="mailto:fancy@starloft.cn">fancy@starloft.cn</a>
            <a class="footer-link" href="tel:13472507077">13472507077</a>
            <span class="footer-link">微信：StarLoftCoLtd</span>
            <span class="footer-link">QQ：2735124804</span>
            <span class="footer-link">服务时间：周一至周日 8:00 - 24:00</span>
          </div>
        </div>
        <div class="footer-bottom">
          <span>© {{ currentYear }} 星楼云 StarLoft · 保留所有权利</span>
          <a class="icp-link" href="https://beian.miit.gov.cn/" target="_blank" rel="noopener">沪ICP备2026043262号</a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

// 版权年份动态取当前年
const currentYear = new Date().getFullYear()

// 产品下拉：桌面端 hover 展开，移动端触屏点击展开（menuOpen）
const menuOpen = ref(false)

const toggleMenu = () => {
  menuOpen.value = !menuOpen.value
}

const closeMenu = () => {
  menuOpen.value = false
}

// 点击下拉区域之外时收起菜单
const onClickOutside = (e: MouseEvent) => {
  if (!(e.target as HTMLElement).closest('.nav-dropdown')) {
    closeMenu()
  }
}

onMounted(() => document.addEventListener('click', onClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', onClickOutside))

// 路由变化时收起菜单
watch(() => route.path, closeMenu)
</script>

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

/* 产品下拉菜单（hover 展开） */
.nav-dropdown {
  position: relative;
}

.nav-dropdown .nav-item {
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.dropdown-caret {
  width: 0;
  height: 0;
  border-left: 4px solid transparent;
  border-right: 4px solid transparent;
  border-top: 5px solid currentColor;
  transition: transform 0.15s;
}

.nav-dropdown:hover .dropdown-caret,
.nav-dropdown.open .dropdown-caret {
  transform: rotate(180deg);
}

.dropdown-menu {
  position: absolute;
  top: 100%;
  left: 0;
  min-width: 160px;
  max-height: 70vh;   /* 限制最大高度，超出后内部滚动，避免产品增多时下拉溢出屏幕 */
  overflow-y: auto;
  padding: 6px;
  background: var(--bg-panel, #fff);
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: var(--radius-md);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
  opacity: 0;
  visibility: hidden;
  transform: translateY(6px);
  transition: opacity 0.15s, transform 0.15s, visibility 0.15s;
  z-index: 100;
}

/* hover（桌面端）与点击展开（移动端触屏）两种方式均可显示菜单 */
.nav-dropdown:hover .dropdown-menu,
.nav-dropdown:focus-within .dropdown-menu,
.nav-dropdown .dropdown-menu.menu-open {
  opacity: 1;
  visibility: visible;
  transform: translateY(0);
}

.dropdown-item {
  display: block;
  padding: 9px 14px;
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-size: 14px;
  text-decoration: none;
  transition: all 0.15s;
}

.dropdown-item:hover {
  color: var(--color-primary);
  background: var(--bg-hover);
}

.dropdown-item.router-link-active {
  color: var(--color-primary);
  background: var(--bg-active);
  font-weight: 600;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 8px 18px;
  border-radius: var(--radius-md);
  border: 1px solid transparent;
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

/* 登录：与注册按钮同尺寸，采用描边风格突出主次 */
.action-btn.action-btn-plain {
  background: transparent;
  border-color: var(--color-primary);
  color: var(--color-primary);
}

.action-btn.action-btn-plain:hover {
  background: var(--bg-active);
  color: var(--color-primary);
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
  /* 6 列拆成两行（3x2），避免每列过窄 */
  grid-template-columns: repeat(3, 1fr);
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

.icp-link {
  margin-left: 12px;
  color: var(--text-muted);
  text-decoration: none;
  transition: color 0.15s;
}

.icp-link:hover {
  color: var(--color-primary);
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
