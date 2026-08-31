<template>
  <div class="markdown-body">
    <h1>API 文档</h1>
    <p class="lead">星楼云平台对外提供基于 API Key + HMAC-SHA256 签名鉴权的 RESTful 接口。</p>

    <h2>版本列表</h2>
    <table>
      <thead>
        <tr>
          <th>版本</th>
          <th>状态</th>
          <th>说明</th>
          <th>文档</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>v1</td>
          <td>当前版本</td>
          <td>提供创建认证订单、查询认证结果、查询余额等接口</td>
          <td><router-link to="/docs/api/v1">查看文档</router-link></td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts"></script>
