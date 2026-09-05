<template>
  <div class="markdown-body">
    <h1>API v1 文档</h1>
    <p class="lead">
      本文档介绍星楼云平台 API v1 的鉴权方式与接口调用方法。所有接口均使用
      <strong>API Key + HMAC-SHA256 签名</strong> 鉴权。
    </p>

    <h2>1. 基础信息</h2>
    <ul>
      <li>接口地址：<code>https://www.starloft.cn/api</code></li>
      <li>请求格式：<code>application/json</code></li>
      <li>鉴权方式：请求头携带 API Key 与签名（见「请求鉴权」）</li>
    </ul>

    <h2>2. 开通前提（必读）</h2>
    <p>调用本 API 前，账户必须同时满足以下三个条件：</p>
    <ol>
      <li>
        <strong>完成企业实名认证。</strong>在星楼云控制台完成「企业实名认证」，
        使账户处于企业实名状态（<code>is_kyc_verified=2</code>）。
      </li>
      <li>
        <strong>开通 KYC 服务。</strong>在控制台服务管理中开通 <strong>KYC 实名认证 API</strong> 服务
        （对应用户侧接口 <code>POST /service/open</code>，入参 <code>service_code=kyc</code>）。
        开通为幂等操作，可重复调用。
      </li>
      <li>
        <strong>获取 API Key 与 API Secret。</strong>两者均可在控制台「API 密钥管理」中查看
        （API Key 注册时自动生成且唯一；API Secret 用于请求签名与回调验签，仅企业实名成功后在系统中下发）。
      </li>
    </ol>
    <p>
      未完成企业实名或未开通 KYC 服务时直接调用 <code>/api/kyc/*</code> 接口，
      将返回 <code>code: 403</code> 与「服务未开通」提示。
    </p>

    <h2>3. 请求鉴权</h2>
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
          <td>控制台「API 密钥管理」中获取的 API Key</td>
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

    <h2>4. 接口列表</h2>

    <h3>4.1 创建认证订单</h3>
    <p><span class="method post">POST</span><code>/api/kyc/start</code></p>
    <p>发起一次实名认证（API 业务调用，始终计费）。由平台统一生成 20 位唯一订单号 <code>biz_no</code> 并返回认证跳转地址（<code>auth_url</code>）。认证完成后用户浏览器将直接跳转回你填写的 <code>return_url</code>，不再经过平台中转页。</p>

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
          <td>认证完成后用户浏览器直接跳转地址（原样透传至上游）</td>
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
    <p>说明：由平台统一生成订单号 <code>biz_no</code>，调用方无需（也无需自行生成）传入业务单号；后续用返回的 <code>biz_no</code> 查询结果或关联回调。</p>

    <h4>请求示例</h4>
    <pre><code>{
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
    "biz_no": "46382671905182934716",
    "auth_url": "https://auth.finauth.com/verify?token=xxx",
    "expired_time": 1234567890,
    "expired_in": 900
  }
}</code></pre>

    <h3>4.2 查询认证结果</h3>
    <p><span class="method post">POST</span><code>/api/kyc/result</code></p>
    <p>根据创建认证订单时返回的平台订单号 <code>biz_no</code> 查询认证结果（单号二选一填写）。</p>

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
          <td>创建认证订单时返回的 20 位平台订单号</td>
        </tr>
      </tbody>
    </table>

    <h4>请求示例</h4>
    <pre><code>{
  "biz_no": "46382671905182934716"
}</code></pre>

    <h4>响应示例</h4>
    <pre><code>{
  "code": 0,
  "message": "success",
  "data": {
    "biz_no": "46382671905182934716",
    "status": 2,
    "result_code": "1000",
    "result_message": "SUCCESS",
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

    <h3>4.3 查询余额</h3>
    <p><span class="method post">POST</span><code>/api/kyc/balance/query</code></p>
    <p>查询当前 API Key 所属账户的余额与平台 KYC 认证单价（统一按平台价格扣费，无个人价格区分）。</p>

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

    <h2>5. 回调说明（notify_url 与 return_url）</h2>
    <p>
      发起认证时需提供 <code>return_url</code>（同步跳转）与 <code>notify_url</code>（异步通知），
      用于认证结果回流。两者触发时机与数据格式如下。
    </p>

    <h3>5.1 异步通知（notify_url）</h3>
    <p>
      认证产生最终结果（成功或失败）后，平台会以 <strong>POST</strong> 方式向你的
      <code>notify_url</code> 发送通知，<code>Content-Type: application/json</code>。
      通知携带 <code>sign</code>（HMAC-SHA256 签名，基于你的 API Secret 生成），
      下游应校验签名后再落地结果，防止伪造回调；同时可依据 <code>biz_no</code>（平台订单号）关联订单。
    </p>

    <h4>请求体</h4>
    <pre><code>{
  "biz_no": "46382671905182934716",
  "status": 2,
  "result_code": "1000",
  "result_message": "SUCCESS",
  "cost": 1.50,
  "sign": "a1b2c3d4e5f6..."
}</code></pre>

    <h4>字段说明</h4>
    <table>
      <thead>
        <tr>
          <th>字段</th>
          <th>类型</th>
          <th>说明</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td><code>biz_no</code></td>
          <td>string</td>
          <td>平台订单号（创建认证订单时返回，用于关联订单）</td>
        </tr>
        <tr>
          <td><code>status</code></td>
          <td>int</td>
          <td>订单状态：<code>2</code> 认证成功 / <code>3</code> 认证失败（详见「status 状态说明」）</td>
        </tr>
        <tr>
          <td><code>result_code</code></td>
          <td>string</td>
          <td>上游认证结果码，见下方「result_code 说明」</td>
        </tr>
        <tr>
          <td><code>result_message</code></td>
          <td>string</td>
          <td>结果说明（英文常量，如 SUCCESS）</td>
        </tr>
        <tr>
          <td><code>cost</code></td>
          <td>number</td>
          <td>本次认证扣费金额（元）</td>
        </tr>
        <tr>
          <td><code>sign</code></td>
          <td>string</td>
          <td>回调签名，用于防止伪造回调（算法见下方「签名算法」）</td>
        </tr>
      </tbody>
    </table>

    <h4>签名算法</h4>
    <p>
      取 <code>biz_no</code>、<code>cost</code>、<code>result_code</code>、
      <code>result_message</code>、<code>status</code> 五个字段，
      按 key 字典序拼接为原始字符串（不做 URL 编码），以 API Secret 计算 HMAC-SHA256：
    </p>
    <pre><code>canonical = "biz_no=xxx&amp;cost=1.50&amp;result_code=1000&amp;result_message=xxx&amp;status=2"
sign = 小写hex( HMAC-SHA256(api_secret, canonical) )</code></pre>
    <p>其中 <code>cost</code> 固定保留两位小数（如 <code>1.50</code>），<code>status</code> 为整数。校验失败应拒绝该通知（返回非 2xx），并依赖轮询接口兜底。</p>

    <h4>result_code 说明</h4>
    <table>
      <thead>
        <tr>
          <th>result_code</th>
          <th>result_message</th>
          <th>含义</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td><code>1000</code></td>
          <td>SUCCESS</td>
          <td>认证通过</td>
        </tr>
        <tr>
          <td><code>2000</code></td>
          <td>PASS_LIVING_NOT_THE_SAME</td>
          <td>活体通过，但与身份证非同一人</td>
        </tr>
        <tr>
          <td><code>3000</code></td>
          <td>NO_ID_CARD_NUMBER / ID_NUMBER_NAME_NOT_MATCH / NO_FACE_FOUND / NO_ID_PHOTO / PHOTO_FORMAT_ERROR</td>
          <td>证号、姓名不匹配或未检出人脸等（认证不通过）</td>
        </tr>
        <tr>
          <td><code>3000</code></td>
          <td>DATA_SOURCE_ERROR / INTERNAL_ERROR</td>
          <td>数据源或服务器临时异常</td>
        </tr>
        <tr>
          <td><code>4000</code></td>
          <td>FAIL_LIVING_FACE_ATTACK</td>
          <td>活体攻击 / 活体检测失败</td>
        </tr>
        <tr>
          <td><code>6000</code></td>
          <td>FAILED / CANCELLED / TIMEOUT</td>
          <td>流程异常结束 / 用户取消 / 超时</td>
        </tr>
        <tr>
          <td><code>6100</code></td>
          <td>SUPPORT_ERROR / PERMISSIONS_ERROR / OTHER_ERROR</td>
          <td>webRTC / 摄像头权限等问题</td>
        </tr>
      </tbody>
    </table>
    <p>
      <code>6000</code> / <code>6100</code> 为不计费结果，平台会自动退还预扣金额；
      <code>NOT_STARTED</code> / <code>PROCESSING</code> 表示认证尚未完结，不会作为最终通知发送。
    </p>

    <h4>响应要求</h4>
    <p>下游收到通知后返回 HTTP 200～299 即可（平台校验 2xx 视为通知成功，否则后续重试）。</p>

    <h3>5.2 同步跳转（return_url）</h3>
    <p>
      认证结果确定后，平台会将用户浏览器以 <strong>GET</strong> 方式 302 跳转回你发起认证时填写的
      <code>return_url</code>，且<strong>不附加任何参数</strong>（跳转地址与你传入的 return_url 完全一致，
      认证失败同样跳回该 return_url）。
      下游应在该页面根据自己记录的 <code>biz_no</code> 调用「4.2 查询认证结果」接口获取最终状态，再展示结果。
    </p>
    <p>示例：</p>
    <pre><code>https://yourdomain.com/certification/starloft_kyc/result</code></pre>
    <p>
      说明：<code>return_url</code> 用于用户侧页面回流，<code>notify_url</code> 用于服务端结果落地，
      两者配合使用；若 <code>notify_url</code> 未及时到达，可主动调用查询接口兜底。
    </p>

    <h2>6. 调用示例（curl）</h2>
    <p>以查询余额为例：</p>
    <pre><code>BODY='{}'
SIGN=$(printf '%s' "$BODY" | openssl dgst -sha256 -hmac "your_api_secret" | awk '{print $2}')
TS=$(date +%s)

curl -X POST "https://www.starloft.cn/api/kyc/balance/query" \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: your_api_key" \
  -H "X-Sign: $SIGN" \
  -H "X-Sign-Version: hmac_sha256" \
  -H "X-Timestamp: $TS" \
  -d "$BODY"</code></pre>

    <h2>7. 错误码</h2>
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
          <td>权限不足（用户被禁用、未完成企业实名，或 KYC 服务未开通）</td>
        </tr>
        <tr>
          <td><code>404</code></td>
          <td>订单不存在</td>
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