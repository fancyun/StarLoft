<template>
  <div class="markdown-body">
    <h1>智简魔方·财务版（zjmf_mfcw）实名认证插件</h1>
    <p class="lead">
      平台标识：<code>zjmf_mfcw</code>，插件名称：<code>starloft_kyc</code>。为智简魔方·财务版（3.7.6+）系统
      提供身份证三要素实名认证（姓名 + 身份证 + 人脸识别），对接星楼云实名认证系统。
    </p>

    <h2>1. 功能特性</h2>
    <ul>
      <li>身份证三要素实名认证（姓名 + 身份证 + 人脸识别）</li>
      <li>自动对接星楼云后端系统</li>
      <li>支持自动扣费与强制实名</li>
      <li>支持年龄限制（可配置最低实名年龄）</li>
      <li>认证结果自动回调与状态轮询</li>
      <li>安全的 HMAC-SHA256 签名认证</li>
    </ul>

    <h2>2. 安装步骤</h2>
    <h3>2.1 上传插件文件</h3>
    <p>将 <code>starloft_kyc</code> 文件夹上传到以下目录：</p>
    <pre><code>/public/plugins/certification/starloft_kyc/</code></pre>

    <h3>2.2 后台安装</h3>
    <ol>
      <li>登录后台管理面板</li>
      <li>进入「系统设置」→「实名认证设置」→「接口设置」</li>
      <li>找到「StarLoft KYC 实名认证」</li>
      <li>点击「安装」按钮</li>
    </ol>

    <h3>2.3 配置插件</h3>
    <p>安装成功后点击「配置」，填写以下信息：</p>
    <table>
      <thead>
        <tr>
          <th>配置项</th>
          <th>必填</th>
          <th>说明</th>
          <th>示例</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>API 地址</td>
          <td>是</td>
          <td>KYC 系统的 API 地址</td>
          <td><code>https://www.starloft.cn/api/v1</code></td>
        </tr>
        <tr>
          <td>API Key</td>
          <td>是</td>
          <td>在 KYC 系统后台获取</td>
          <td><code>your_api_key_here</code></td>
        </tr>
        <tr>
          <td>API Secret</td>
          <td>是</td>
          <td>用于 HMAC 签名</td>
          <td><code>your_api_secret_here</code></td>
        </tr>
        <tr>
          <td>单次认证费用</td>
          <td>否</td>
          <td>每次认证费用（元）</td>
          <td><code>0</code> 或 <code>2.00</code></td>
        </tr>
        <tr>
          <td>免费认证次数</td>
          <td>否</td>
          <td>每个用户免费次数</td>
          <td><code>0</code> 或 <code>3</code></td>
        </tr>
        <tr>
          <td>自动扣费</td>
          <td>否</td>
          <td>是否自动扣除费用</td>
          <td>启用 / 禁用</td>
        </tr>
        <tr>
          <td>强制实名</td>
          <td>否</td>
          <td>购买前是否必须实名</td>
          <td>启用 / 禁用</td>
        </tr>
        <tr>
          <td>最低实名年龄</td>
          <td>否</td>
          <td>要求的最低年龄（周岁），根据身份证出生日期计算，0 表示不限</td>
          <td><code>16</code></td>
        </tr>
      </tbody>
    </table>

    <h3>2.4 获取 API 密钥</h3>
    <p>登录星楼云控制台：进入「用户中心」→「API 密钥管理」→ 复制 API Key 和 API Secret。</p>

    <h2>3. 使用方法</h2>
    <h3>3.1 用户实名认证</h3>
    <p>访问地址：</p>
    <pre><code>https://你的域名/certification/starloft_kyc</code></pre>
    <p>认证流程：填写姓名与身份证号 → 提交后跳转认证页面 → 完成人脸识别 → 自动返回并更新状态。</p>

    <h3>3.2 常见配置场景</h3>
    <table>
      <thead>
        <tr>
          <th>场景</th>
          <th>建议配置</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>仅提供服务，不强制</td>
          <td>自动扣费启用、强制实名禁用</td>
        </tr>
        <tr>
          <td>购买前必须实名</td>
          <td>自动扣费启用、<strong>强制实名启用</strong></td>
        </tr>
        <tr>
          <td>提供免费次数</td>
          <td>自动扣费启用、设置免费次数</td>
        </tr>
        <tr>
          <td>限制实名年龄</td>
          <td>设置「最低实名年龄」为所需周岁（如 16）</td>
        </tr>
      </tbody>
    </table>

    <h2>4. 对接接口</h2>
    <p>插件对接以下接口，鉴权方式详见 <router-link to="/docs/api/v1">API v1 文档</router-link>：</p>
    <ul>
      <li><span class="method post">POST</span> <code>/api/v1/kyc/start</code> — 创建认证订单</li>
      <li><span class="method post">POST</span> <code>/api/v1/kyc/result</code> — 查询认证结果</li>
      <li><span class="method post">POST</span> <code>/api/v1/kyc/balance/query</code> — 查询余额</li>
    </ul>

    <h2>5. 故障排查</h2>
    <h3>5.1 API 连接失败</h3>
    <ul>
      <li>确认 API 地址正确（包含 <code>/api/v1</code>）</li>
      <li>确认 API Key 与 Secret 正确</li>
      <li>确认 KYC 系统运行正常、服务器可访问</li>
    </ul>

    <h3>5.2 认证一直显示「认证中」</h3>
    <ul>
      <li>用户未完成人脸识别，等待用户完成</li>
      <li>系统会自动轮询查询状态</li>
      <li>检查 <code>notify_url</code> 是否可被外网访问</li>
    </ul>

    <h2>6. 环境要求</h2>
    <ul>
      <li>PHP &gt;= 7.0</li>
      <li>curl 扩展</li>
      <li>json 扩展</li>
      <li>SSL 支持</li>
    </ul>

    <h2>7. 技术支持</h2>
    <ul>
      <li>插件版本：v1.0.0</li>
      <li>兼容版本：智简魔方·财务版 3.7.6+</li>
      <li>作者：StarLoft</li>
    </ul>
  </div>
</template>

<script setup lang="ts"></script>
