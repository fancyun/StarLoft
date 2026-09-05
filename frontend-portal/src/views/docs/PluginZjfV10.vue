<template>
  <div class="markdown-body">
    <h1>智简魔方业务系统 v10 · KYC 实名认证插件</h1>
    <p class="lead">
      对接 StarLoft KYC 系统的实名认证插件，适用于<strong>智简魔方业务系统 v10</strong>。
      支持身份证三要素实名认证（姓名 + 身份证 + 人脸识别），并提供完整的 API 对接、状态轮询与异步回调功能。
    </p>

    <h2>1. 功能特性</h2>
    <ul>
      <li>身份证三要素实名认证（姓名 + 身份证 + 人脸识别）</li>
      <li>按 v10 实名认证接口规范开发（<code>ZjmfV10Person</code> / <code>ZjmfV10CollectionInfo</code>）</li>
      <li>支持自动扣费 / 免费认证次数</li>
      <li>支持强制实名认证（由 v10 后台开启）</li>
      <li>认证结果异步回调（<code>notify_url</code>）+ 前台状态轮询</li>
      <li>安全的 HMAC-SHA256 签名认证</li>
      <li>支持年龄限制（可配置最低实名年龄）</li>
      <li>幂等保护：认证中任务自动复用，避免重复发单扣费</li>
      <li>轮询上限保护：网络异常/任务丢失时自动终止，杜绝无限轮询</li>
    </ul>

    <h2>2. 目录结构</h2>
    <pre><code>zjmf_v10/
├── ZjmfV10.php                    # 插件入口文件（命名空间 certification\zjmf_v10）
├── config.php                     # 插件配置项（后台「实名认证 → 接口管理 → 配置」展示）
├── controller/
│   └── IndexController.php        # 外部回调控制器
│       ├── notifyHandle()         # 异步通知处理（KYC 平台推送认证结果）
│       ├── result()               # 认证完成回跳页
│       └── status()               # 状态查询（AJAX 轮询）
└── logic/
    └── KycSdk.php                 # StarLoft KYC SDK（API 通信 + HMAC 签名 + 错误分类）</code></pre>

    <h2>3. 安装步骤</h2>
    <h3>3.1 上传插件文件</h3>
    <p>将 <code>zjmf_v10</code> 文件夹上传到：</p>
    <pre><code>/public/plugins/certification/zjmf_v10/</code></pre>

    <h3>3.2 后台安装</h3>
    <ol>
      <li>登录 v10 管理后台</li>
      <li>进入 <code>实名认证</code> → <code>接口管理</code></li>
      <li>找到 <code>StarLoft KYC实名认证</code></li>
      <li>点击 <code>安装</code>，然后点击 <code>配置</code></li>
    </ol>

    <h3>3.3 配置插件</h3>
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
          <td><code>API地址</code></td>
          <td>是</td>
          <td>StarLoft KYC 系统的 API 地址</td>
          <td><code>https://www.starloft.cn/api</code></td>
        </tr>
        <tr>
          <td><code>API Key</code></td>
          <td>是</td>
          <td>在 KYC 系统后台获取</td>
          <td><code>your_api_key_here</code></td>
        </tr>
        <tr>
          <td><code>API Secret</code></td>
          <td>是</td>
          <td>在 KYC 系统后台获取，用于 HMAC 签名</td>
          <td><code>your_api_secret_here</code></td>
        </tr>
        <tr>
          <td><code>单次认证费用</code></td>
          <td>否</td>
          <td>每次认证费用（元），0 表示不扣费</td>
          <td><code>0</code> 或 <code>2.00</code></td>
        </tr>
        <tr>
          <td><code>免费认证次数</code></td>
          <td>否</td>
          <td>每个用户免费次数，0 表示无免费</td>
          <td><code>0</code> 或 <code>3</code></td>
        </tr>
        <tr>
          <td><code>最低实名年龄</code></td>
          <td>否</td>
          <td>要求的最低年龄（周岁），0 表示不限</td>
          <td><code>16</code></td>
        </tr>
        <tr>
          <td><code>认证完成回跳地址</code></td>
          <td>否</td>
          <td>用户完成认证后浏览器回跳地址；留空使用插件内置结果页</td>
          <td>可留空</td>
        </tr>
      </tbody>
    </table>
    <div class="notice">
      <code>amount</code>（单次认证费用）与 <code>free</code>（免费认证次数）为 v10 实名认证系统必需字段，由 v10 后台统一处理扣费逻辑。
    </div>

    <h3>3.4 获取 API 密钥</h3>
    <ol>
      <li>完成账户实名认证</li>
      <li>进入 <code>用户中心</code> → <code>API密钥管理</code></li>
      <li>复制 API Key 和 API Secret</li>
    </ol>

    <h2>4. 使用方法</h2>
    <h3>4.1 用户实名认证</h3>
    <p>用户在 v10 会员中心进入 <code>实名认证</code> 页面：</p>
    <ol>
      <li>选择 <code>StarLoft KYC实名认证</code> 认证方式</li>
      <li>填写姓名和身份证号</li>
      <li>点击提交，插件创建认证任务并跳转到认证页面</li>
      <li>完成人脸识别</li>
      <li>认证结果自动通过回调/轮询同步，用户返回会员中心查看状态</li>
    </ol>

    <h3>4.2 认证状态说明</h3>
    <table>
      <thead>
        <tr>
          <th>状态</th>
          <th>含义</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td><code>0</code></td>
          <td>待认证</td>
        </tr>
        <tr>
          <td><code>1</code></td>
          <td>认证通过</td>
        </tr>
        <tr>
          <td><code>2</code></td>
          <td>认证失败</td>
        </tr>
        <tr>
          <td><code>4</code></td>
          <td>认证中</td>
        </tr>
      </tbody>
    </table>

    <h2>5. 回调地址</h2>
    <p>插件内置两个外部访问地址：</p>
    <table>
      <thead>
        <tr>
          <th>地址</th>
          <th>用途</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td><code>/certification/zjmf_v10/index/notifyHandle</code></td>
          <td>异步通知（StarLoft KYC 平台推送认证结果）</td>
        </tr>
        <tr>
          <td><code>/certification/zjmf_v10/index/result</code></td>
          <td>认证完成回跳页</td>
        </tr>
      </tbody>
    </table>
    <p><code>notify_url</code> 由插件在创建认证任务时自动生成并传给 KYC 平台，无需手工配置。</p>

    <h3>5.1 异步通知格式</h3>
    <p>StarLoft KYC 平台在认证终态时向 <code>notify_url</code> 推送：</p>
    <pre><code>POST /certification/zjmf_v10/index/notifyHandle
Content-Type: application/json

{
    "biz_no":         "46382671905182934716",
    "status":          2,
    "result_code":     "1000",
    "result_message":  "认证成功",
    "cost":            1.50,
    "sign":            "HMAC-SHA256 签名（防伪造，插件必校验）"
}</code></pre>

    <p>字段说明：</p>
    <ul>
      <li><code>biz_no</code>：全平台唯一流水号（插件保存为认证记录的 <code>certify_id</code>）</li>
      <li><code>status</code>：0 待认证 / 1 认证中 / 2 成功 / 3 失败 / 4 已取消 / 5 超时</li>
      <li><code>result_code</code> / <code>result_message</code>：上游结果码与说明</li>
      <li><code>cost</code>：本次认证扣费金额</li>
      <li><code>sign</code>：回调签名（HMAC-SHA256），插件校验不通过会拒绝（返回 401），防止伪造回调</li>
    </ul>

    <p><strong>签名算法</strong>（与插件 <code>verifyNotifySign()</code> 一致）：</p>
    <ol>
      <li>取 <code>biz_no</code> / <code>cost</code> / <code>result_code</code> / <code>result_message</code> / <code>status</code> 五个字段</li>
      <li>按 key 字典序拼接为原始字符串（不做 URL 编码）：<code>k1=v1&amp;k2=v2&amp;...</code></li>
      <li><code>sign = 小写hex( HMAC-SHA256(api_secret, 拼接串) )</code></li>
    </ol>
    <p>
      其中 <code>cost</code> 固定保留两位小数（如 <code>1.50</code>），<code>status</code> 为整数。
      <code>api_secret</code> 即插件配置的 API Secret。
    </p>
    <p>
      插件收到通知后先校验 <code>sign</code>，校验通过再按 <code>biz_no</code>（即 <code>certify_id</code>）定位认证记录并更新状态：
      <code>status=2 → 通过</code>、<code>status=3/4/5 → 失败</code>。
    </p>

    <h2>6. API 接口说明</h2>
    <p>
      插件通过 <code>logic/KycSdk.php</code> 对接以下 StarLoft KYC API，所有请求使用
      <strong>API Key + HMAC-SHA256 签名</strong> 鉴权。
    </p>

    <h3>6.1 请求鉴权（每个请求必带 4 个请求头）</h3>
    <pre><code>X-Api-Key: &lt;你的API Key&gt;
X-Sign: &lt;签名&gt;
X-Sign-Version: hmac_sha256
X-Timestamp: &lt;当前Unix时间戳（秒）&gt;</code></pre>
    <p>签名算法：</p>
    <pre><code>sign = hex(HMAC-SHA256(api_secret, 原始请求体))</code></pre>
    <p>PHP 写法：</p>
    <pre><code>$sign = hash_hmac('sha256', $body, $this->apiSecret); // 小写十六进制</code></pre>

    <h3>6.2 创建认证订单</h3>
    <pre><code>POST /api/kyc/start

{
    "name":           "张三",                              // 真实姓名
    "id_card":        "110101199001011234",               // 身份证号
    "return_url":     "https://yourdomain.com/certification/zjmf_v10/index/result",
    "notify_url":     "https://yourdomain.com/certification/zjmf_v10/index/notifyHandle",
    "biz_extra_data": "{\"uid\":1}"                        // 业务扩展数据
}</code></pre>
    <div class="notice"><code>biz_no</code> 由平台随机生成（20 位数字）并在响应下发给下游，无需（也不应）由下游传入。</div>
    <p>响应：</p>
    <pre><code>{
    "code": 0,
    "message": "success",
    "data": {
        "biz_no": "46382671905182934716",
        "auth_url": "https://auth.finauth.com/verify?token=xxx",
        "expired_time": 1737000900,
        "expired_in": 900
    }
}</code></pre>

    <h3>6.3 查询认证结果</h3>
    <pre><code>POST /api/kyc/result

{
    "biz_no": "46382671905182934716"
}</code></pre>
    <p>响应：</p>
    <pre><code>{
    "code": 0,
    "message": "success",
    "data": {
        "biz_no": "46382671905182934716",
        "status": 2,
        "result_code": "1000",
        "result_message": "认证成功",
        "cost": 1.50
    }
}</code></pre>

    <h3>6.4 查询用户余额</h3>
    <pre><code>POST /api/kyc/balance/query</code></pre>
    <p>响应：</p>
    <pre><code>{
    "code": 0,
    "message": "success",
    "data": {
        "balance": 100.50,
        "kyc_price": 1.50
    }
}</code></pre>

    <h2>7. 常见问题</h2>

    <h3>Q1: 安装后找不到插件？</h3>
    <ul>
      <li>确认目录名正确：<code>/public/plugins/certification/zjmf_v10/</code></li>
      <li>确认入口文件名正确：<code>ZjmfV10.php</code>（目录名大驼峰 + .php）</li>
      <li>清空 v10 运行时缓存后重新进入 <code>实名认证 → 接口管理</code></li>
    </ul>

    <h3>Q2: 认证一直显示"认证中"？</h3>
    <ul>
      <li>用户未完成人脸识别</li>
      <li>检查 <code>notify_url</code>（<code>/certification/zjmf_v10/index/notifyHandle</code>）能否从外网访问</li>
      <li>插件会按 2s/次自动轮询查询状态，认证完成后自动更新</li>
      <li>若长时间无结果，可检查 KYC 后台该订单的状态</li>
    </ul>

    <h3>Q3: API 连接失败 / 鉴权失败？</h3>
    <ul>
      <li>检查 API 地址是否正确（包含 <code>/api</code>）</li>
      <li>检查 API Key / Secret 是否正确</li>
      <li>确认 KYC 系统后台已完成实名认证并生成 API 密钥</li>
      <li>确认服务器可访问 KYC 系统、时间戳与服务器时间同步（±5 分钟内）</li>
    </ul>

    <h3>Q4: 认证费用如何计算？</h3>
    <ul>
      <li>费用由 KYC 系统设定（单次认证费用）</li>
      <li>插件配置的 <code>amount</code> / <code>free</code> 由 v10 系统用于用户扣费 / 免费次数控制</li>
      <li>可在 KYC 后台查看当前单价</li>
    </ul>

    <h2>8. 技术支持</h2>
    <table>
      <tbody>
        <tr>
          <td>插件版本</td>
          <td>v1.0.0</td>
        </tr>
        <tr>
          <td>作者</td>
          <td>StarLoft</td>
        </tr>
        <tr>
          <td>兼容版本</td>
          <td>智简魔方业务系统 v10</td>
        </tr>
        <tr>
          <td>文档</td>
          <td><code>https://docs.starloft.cn/kyc/plugin/v10</code></td>
        </tr>
        <tr>
          <td>问题反馈</td>
          <td><code>https://github.com/starloft/kyc/issues</code></td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts"></script>