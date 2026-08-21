<template>
  <div class="markdown-body">
    <h1>API v1 文档</h1>
    <p class="lead">
      本文档介绍星楼KYC 平台 API v1 的鉴权方式与接口调用方法。所有接口均使用
      <strong>API Key + HMAC-SHA256 签名</strong> 鉴权。
    </p>

    <h2>1. 基础信息</h2>
    <ul>
      <li>接口地址：<code>https://kyc.starloft.cn/api/v1</code></li>
      <li>请求格式：<code>application/json</code></li>
      <li>鉴权方式：请求头携带 API Key 与签名（见下）</li>
    </ul>

    <h2>2. 请求鉴权</h2>
    <p>每个请求必须携带以下 4 个请求头：</p>
    <table>
      <thead>
        <tr>
          <th>请求头</th>
          <th>说明</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td><code>X-Api-Key</code></td>
          <td>在用户中心「API 密钥管理」中获取的 API Key</td>
        </tr>
        <tr>
          <td><code>X-Sign</code></td>
          <td>签名，算法见下方</td>
        </tr>
        <tr>
          <td><code>X-Sign-Version</code></td>
          <td>固定值 <code>hmac_sha256</code></td>
        </tr>
        <tr>
          <td><code>X-Timestamp</code></td>
          <td>当前 Unix 时间戳（秒），允许 ±5 分钟误差（防重放）</td>
        </tr>
      </tbody>
    </table>

    <h3>签名算法</h3>
    <p>签名为对<strong>原始请求体</strong>（POST 的 JSON 字符串，与发送内容完全一致）进行 HMAC-SHA256 运算后的小写十六进制字符串：</p>
    <pre><code>sign = hex(HMAC-SHA256(api_secret, 原始请求体))</code></pre>

    <p>PHP 示例：</p>
    <pre><code>$sign = hash_hmac('sha256', $body, $apiSecret); // 小写十六进制</code></pre>

    <h2>3. 接口列表</h2>

    <h3>3.1 创建认证订单</h3>
    <p><span class="method post">POST</span><code>/api/v1/kyc/start</code></p>
    <p>发起一次实名认证。平台将扣取余额、创建订单并返回认证跳转地址。</p>

    <h4>请求参数</h4>
    <table>
      <thead>
        <tr>
          <th>参数</th>
          <th>类型</th>
          <th>必填</th>
          <th>说明</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td><code>biz_no</code></td>
          <td>string</td>
          <td>是</td>
          <td>业务订单号（下游系统内唯一）</td>
        </tr>
        <tr>
          <td><code>name</code></td>
          <td>string</td>
          <td>是</td>
          <td>真实姓名</td>
        </tr>
        <tr>
          <td><code>id_card</code></td>
          <td>string</td>
          <td>是</td>
          <td>身份证号</td>
        </tr>
        <tr>
          <td><code>return_url</code></td>
          <td>string</td>
          <td>是</td>
          <td>认证完成后前端跳转地址</td>
        </tr>
        <tr>
          <td><code>notify_url</code></td>
          <td>string</td>
          <td>是</td>
          <td>认证结果异步通知地址</td>
        </tr>
        <tr>
          <td><code>biz_extra_data</code></td>
          <td>string</td>
          <td>否</td>
          <td>业务扩展数据（原样回传）</td>
        </tr>
      </tbody>
    </table>

    <h4>请求示例</h4>
    <pre><code>{
  "biz_no": "ZJMF20260119001",
  "name": "张三",
  "id_card": "110101199001011234",
  "return_url": "https://yourdomain.com/certification/starloft_kyc/result",
  "notify_url": "https://yourdomain.com/certification/starloft_kyc/callback",
  "biz_extra_data": "{\"uid\":1001}"
}</code></pre>

    <h4>响应示例</h4>
    <pre><code>{
  "code": 0,
  "message": "success",
  "data": {
    "platform_biz_no": "ZJMF20260119001_1001_1234567890",
    "auth_url": "https://auth.finauth.com/verify?token=xxx",
    "expired_time": 1234567890,
    "expired_in": 900
  }
}</code></pre>

    <h3>3.2 查询认证结果</h3>
    <p><span class="method post">POST</span><code>/api/v1/kyc/result</code></p>
    <p>根据业务订单号或平台流水号查询认证结果（二选一）。</p>

    <h4>请求参数</h4>
    <table>
      <thead>
        <tr>
          <th>参数</th>
          <th>类型</th>
          <th>必填</th>
          <th>说明</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td><code>biz_no</code></td>
          <td>string</td>
          <td>二选一</td>
          <td>业务订单号</td>
        </tr>
        <tr>
          <td><code>platform_biz_no</code></td>
          <td>string</td>
          <td>二选一</td>
          <td>平台流水号</td>
        </tr>
      </tbody>
    </table>

    <h4>请求示例</h4>
    <pre><code>{
  "platform_biz_no": "ZJMF20260119001_1001_1234567890"
}</code></pre>

    <h4>响应示例</h4>
    <pre><code>{
  "code": 0,
  "message": "success",
  "data": {
    "platform_biz_no": "ZJMF20260119001_1001_1234567890",
    "biz_no": "ZJMF20260119001",
    "status": 2,
    "result_code": "1000",
    "result_message": "认证成功",
    "cost": 1.50
  }
}</code></pre>

    <h4>status 状态说明</h4>
    <table>
      <thead>
        <tr>
          <th>值</th>
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
          <td>认证中</td>
        </tr>
        <tr>
          <td><code>2</code></td>
          <td>认证成功</td>
        </tr>
        <tr>
          <td><code>3</code></td>
          <td>认证失败</td>
        </tr>
        <tr>
          <td><code>5</code></td>
          <td>超时 / 已退款</td>
        </tr>
      </tbody>
    </table>

    <h3>3.3 查询余额</h3>
    <p><span class="method post">POST</span><code>/api/v1/kyc/balance/query</code></p>
    <p>查询当前 API Key 所属账户的余额与实名认证单价。</p>

    <h4>请求示例</h4>
    <p>请求体为空 JSON：</p>
    <pre><code>{}</code></pre>

    <h4>响应示例</h4>
    <pre><code>{
  "code": 0,
  "message": "success",
  "data": {
    "balance": 100.50,
    "kyc_price": 1.50
  }
}</code></pre>

    <h2>4. 调用示例（curl）</h2>
    <p>以查询余额为例：</p>
    <pre><code>BODY='{}'
SIGN=$(printf '%s' "$BODY" | openssl dgst -sha256 -hmac "your_api_secret" | awk '{print $2}')
TS=$(date +%s)

curl -X POST "https://kyc.starloft.cn/api/v1/kyc/balance/query" \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: your_api_key" \
  -H "X-Sign: $SIGN" \
  -H "X-Sign-Version: hmac_sha256" \
  -H "X-Timestamp: $TS" \
  -d "$BODY"</code></pre>

    <h2>5. 错误码</h2>
    <table>
      <thead>
        <tr>
          <th>code</th>
          <th>说明</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td><code>0</code></td>
          <td>成功</td>
        </tr>
        <tr>
          <td><code>400</code></td>
          <td>参数错误 / 余额不足</td>
        </tr>
        <tr>
          <td><code>401</code></td>
          <td>鉴权失败（API Key 无效、签名错误或时间戳过期）</td>
        </tr>
        <tr>
          <td><code>403</code></td>
          <td>权限不足（用户被禁用或未完成实名认证）</td>
        </tr>
        <tr>
          <td><code>404</code></td>
          <td>用户或订单不存在</td>
        </tr>
        <tr>
          <td><code>500</code></td>
          <td>系统内部错误</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts"></script>