<template>
  <div class="markdown-body">
    <h1>文档总览</h1>
    <p class="lead">欢迎使用星楼云实名认证服务。以下文档将帮助你快速完成对接与使用。</p>

    <h2>文档目录</h2>

    <div class="doc-cards">
      <router-link to="/docs/api" class="doc-card">
        <div class="doc-card-icon">API</div>
        <div class="doc-card-body">
          <h3>API 文档</h3>
          <p>面向开发者，介绍平台 API 接口的鉴权方式、调用方法、请求与响应示例以及错误码。</p>
        </div>
      </router-link>

      <router-link to="/docs/plugin" class="doc-card">
        <div class="doc-card-icon">插件</div>
        <div class="doc-card-body">
          <h3>插件使用教程</h3>
          <p>面向智简魔方等系统的管理员，介绍插件的安装、配置与使用方式。</p>
        </div>
      </router-link>
    </div>
  </div>
</template>

<script setup lang="ts"></script>

<style scoped>
.doc-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 20px;
  margin-top: 20px;
}

.doc-card {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  padding: 24px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  background: var(--bg-card);
  text-decoration: none;
  color: inherit;
  transition: all 0.15s;
}

.doc-card:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-md);
}

.doc-card-icon {
  flex-shrink: 0;
  width: 48px;
  height: 48px;
  border-radius: 10px;
  background: var(--color-primary-light);
  color: var(--color-primary);
  font-weight: 700;
  font-size: 13px;
  display: flex;
  align-items: center;
  justify-content: center;
  white-space: nowrap;
}

.doc-card-body h3 {
  margin: 0 0 6px;
  color: var(--text-primary);
  font-size: 17px;
}

.doc-card-body p {
  margin: 0;
  font-size: 14px;
  color: var(--text-muted);
  line-height: 1.6;
}
</style>
