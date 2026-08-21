<template>
  <div class="docs-layout">
    <header class="docs-header">
      <div class="docs-header-inner">
        <router-link to="/docs" class="brand">
          <span class="brand-mark">K</span>
          <span class="brand-text">星楼KYC 文档中心</span>
        </router-link>
        <a href="/" class="home-link">返回首页</a>
      </div>
    </header>

    <div class="docs-body">
      <aside class="docs-sidebar">
        <nav class="docs-nav">
          <div v-for="group in navGroups" :key="group.title" class="nav-group">
            <div class="nav-group-title">{{ group.title }}</div>
            <router-link
              v-for="item in group.items"
              :key="item.path"
              :to="item.path"
              class="nav-item"
              :class="{ active: $route.path === item.path }"
            >
              {{ item.label }}
            </router-link>
          </div>
        </nav>
      </aside>

      <main class="docs-main">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
const navGroups = [
  {
    title: '开始',
    items: [{ label: '文档总览', path: '/docs' }]
  },
  {
    title: 'API 文档',
    items: [
      { label: '版本列表', path: '/docs/api' },
      { label: 'API v1', path: '/docs/api/v1' }
    ]
  },
  {
    title: '插件教程',
    items: [
      { label: '插件列表', path: '/docs/plugin' },
      { label: '智简魔方·财务版', path: '/docs/plugin/zjmf_mfcw' }
    ]
  }
]
</script>

<style scoped>
.docs-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.docs-header {
  height: 56px;
  background: #1a1a2e;
  flex-shrink: 0;
  position: sticky;
  top: 0;
  z-index: 100;
}

.docs-header-inner {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 24px;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #fff;
  text-decoration: none;
  font-weight: 700;
  font-size: 16px;
}

.brand-mark {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  background: var(--color-primary);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 800;
}

.home-link {
  color: rgba(255, 255, 255, 0.7);
  text-decoration: none;
  font-size: 13px;
  transition: color 0.2s;
}

.home-link:hover {
  color: #fff;
}

.docs-body {
  flex: 1;
  display: flex;
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
}

.docs-sidebar {
  width: 220px;
  flex-shrink: 0;
  border-right: 1px solid var(--border-color);
  background: var(--bg-card);
  padding: 24px 0;
}

.docs-nav {
  position: sticky;
  top: 80px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 0 12px;
}

.nav-group-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin: 0 12px 8px;
}

.nav-item {
  display: block;
  padding: 8px 12px;
  border-radius: 6px;
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 14px;
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

.docs-main {
  flex: 1;
  padding: 40px 48px;
  min-width: 0;
}

@media (max-width: 768px) {
  .docs-sidebar {
    display: none;
  }
  .docs-main {
    padding: 24px 20px;
  }
}
</style>

<style>
/* 文档正文通用样式（由 DocsLayout 统一提供，供子页面复用） */
.markdown-body {
  color: var(--text-secondary);
  line-height: 1.75;
  font-size: 15px;
  max-width: 860px;
}

.markdown-body h1 {
  font-size: 28px;
  margin: 0 0 8px;
  color: var(--text-primary);
}

.markdown-body .lead {
  font-size: 16px;
  color: var(--text-muted);
  margin: 0 0 24px;
}

.markdown-body h2 {
  font-size: 22px;
  margin: 36px 0 16px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-color);
  color: var(--text-primary);
}

.markdown-body h3 {
  font-size: 18px;
  margin: 24px 0 12px;
  color: var(--text-primary);
}

.markdown-body p {
  margin: 12px 0;
}

.markdown-body ul,
.markdown-body ol {
  margin: 12px 0;
  padding-left: 24px;
}

.markdown-body li {
  margin: 6px 0;
}

.markdown-body code {
  background: var(--bg-soft);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'JetBrains Mono', 'Courier New', monospace;
  font-size: 13px;
  color: #1f2937;
}

.markdown-body pre {
  background: #0f172a;
  color: #e2e8f0;
  padding: 16px 20px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 16px 0;
  font-size: 13px;
  line-height: 1.6;
}

.markdown-body pre code {
  background: transparent;
  color: inherit;
  padding: 0;
  font-size: 13px;
}

.markdown-body table {
  width: 100%;
  border-collapse: collapse;
  margin: 16px 0;
}

.markdown-body th,
.markdown-body td {
  border: 1px solid var(--border-color);
  padding: 10px 12px;
  text-align: left;
  vertical-align: top;
}

.markdown-body th {
  background: var(--bg-soft);
  font-weight: 600;
  color: var(--text-primary);
}

.markdown-body .method {
  display: inline-block;
  font-weight: 700;
  font-family: 'JetBrains Mono', 'Courier New', monospace;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  margin-right: 8px;
}

.markdown-body .method.post {
  background: #ecfdf5;
  color: #059669;
}

.markdown-body .method.get {
  background: #eff6ff;
  color: #2563eb;
}
</style>