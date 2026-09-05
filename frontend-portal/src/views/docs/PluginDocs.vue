<template>
  <div class="markdown-body">
    <h1>插件使用教程</h1>
    <p class="lead">以下为星楼云各产品提供的下游系统对接插件，选择你的系统查看对应的安装与使用教程。</p>

    <h2>实名认证 · KYC 插件</h2>
    <table>
      <thead>
        <tr>
          <th>插件</th>
          <th>适用系统</th>
          <th>说明</th>
          <th>教程</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td><code>zjmf_mfcw</code></td>
          <td>智简魔方·财务版（3.7.6+）</td>
          <td>身份证三要素实名认证（姓名 + 身份证 + 人脸识别）</td>
          <td><router-link to="/docs/plugin/zjmf_mfcw">查看教程</router-link></td>
        </tr>
        <tr>
          <td><code>zjmf_v10</code></td>
          <td>智简魔方业务系统 v10</td>
          <td>身份证三要素实名认证（姓名 + 身份证 + 人脸识别），支持自动扣费 / 免费次数 / 强制实名</td>
          <td><router-link to="/docs/plugin/zjmf_v10">查看教程</router-link></td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts"></script>
