# StarLoft KYC 插件 - 集成文档

## 概述

本插件为智简魔方（ZJMF）3.7.6 IDC在线服务器售卖系统提供实名认证功能，通过对接 StarLoft KYC 实名认证系统，实现身份证三要素认证（姓名+身份证+人脸识别）。

## 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                      智简魔方 3.7.6 系统                         │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │         ZjmfMfcwPlugin (插件主类)                         │  │
│  │  - personal()      个人实名认证                          │  │
│  │  - getStatus()     查询认证状态                          │  │
│  │  - collectionInfo() 自定义字段                           │  │
│  └──────────────────┬───────────────────────────────────────┘  │
│                     │                                           │
│                     ↓                                           │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │           KycSdk (对接逻辑类)                            │  │
│  │  - generateSign()    生成HMAC签名                        │  │
│  │  - startKyc()        创建认证订单                        │  │
│  │  - queryResult()     查询认证结果                        │  │
│  │  - queryBalance()    查询余额                            │  │
│  └──────────────────┬───────────────────────────────────────┘  │
└────────────────────┼────────────────────────────────────────────┘
                     │ HTTPS + HMAC-SHA256 签名
                     │
                     ↓
┌─────────────────────────────────────────────────────────────────┐
│              StarLoft KYC 实名认证系统                           │
│                                                                 │
│  API接口:                                                        │
│  - POST /kyc/start          创建认证订单                        │
│  - POST /kyc/result         查询认证结果                        │
│  - POST /kyc/balance/query  查询余额                            │
│                                                                 │
│  ↓ 对接上游                                                      │
│  FinAuth H5 Plus (三要素实名认证)                                │
└─────────────────────────────────────────────────────────────────┘
```

## 文件结构

```
zjmf_mfcw/
├── ZjmfMfcwPlugin.php          # 插件主类
│   ├── install()               # 安装方法
│   ├── uninstall()             # 卸载方法
│   ├── personal()              # 个人实名认证
│   ├── company()               # 企业认证（暂不支持）
│   ├── collectionInfo()        # 自定义字段
│   └── getStatus()             # 查询认证状态
│
├── logic/
│   └── KycSdk.php              # KYC SDK
│       ├── generateSign()      # 生成HMAC签名
│       ├── request()           # HTTP请求封装
│       ├── startKyc()          # 创建认证订单
│       ├── queryResult()       # 查询认证结果
│       └── queryBalance()      # 查询余额
│
├── config/
│   └── config.php              # 配置文件
│
├── plugin.json                 # 插件元数据
├── README.md                   # 完整文档
└── QUICK_START.md              # 快速开始
```

## 核心流程

### 1. 认证流程

```
用户提交认证
    ↓
ZjmfMfcwPlugin::personal()
    ↓
KycSdk::startKyc()
    ├─ 生成HMAC签名
    ├─ 构建请求参数
    └─ 发送POST请求到 /kyc/start
    ↓
KYC系统创建订单
    ├─ 验证API Key与HMAC签名
    ├─ 扣除用户余额
    ├─ 调用FinAuth接口
    └─ 返回认证URL
    ↓
用户跳转到认证页面
    ├─ 人脸识别
    └─ 活体检测
    ↓
认证完成
    ├─ 回调notify_url（后端）
    └─ 跳转return_url（前端）
    ↓
系统轮询查询状态
KycSdk::queryResult()
    ↓
更新认证状态
updatePersonalCertifiStatus()
```

### 2. HMAC签名

每个请求均通过以下方式鉴权（见 `KycSdk::request()`）：

```php
// 原始请求体
$body = json_encode($data);

// 时间戳：Unix 秒
$timestamp = time();

// 签名：hex(HMAC-SHA256(api_secret, body))，结果为小写十六进制
$sign = hash_hmac('sha256', $body, $this->apiSecret);

// 请求头
// X-Api-Key:      <api_key>
// X-Sign:         <sign>
// X-Sign-Version: hmac_sha256
// X-Timestamp:    <timestamp>
```

后端校验：
- `X-Sign-Version` 必须为 `hmac_sha256`
- 时间戳必须在 ±5 分钟内
- 签名基于原始请求体，与发送内容完全一致

### 3. 状态轮询机制

```
认证提交后，系统自动轮询查询状态
    ↓
每隔 N 秒调用 getStatus()
    ↓
KycSdk::queryResult()
    ↓
根据后端 order.status（0待认证 1认证中 2成功 3失败 4已取消 5超时）映射为插件本地状态：
- status=1: 认证成功，停止轮询
- status=2: 认证失败，停止轮询
- status=4: 认证中，继续轮询
```

## API接口详解

所有接口均需携带鉴权请求头。以下示例中 `<sign>` 由 `hex(HMAC-SHA256(api_secret, 请求体))` 计算得到。

### 1. 创建认证订单 (POST /kyc/start)

**请求头：**
```
Content-Type: application/json
X-Api-Key: <your_api_key>
X-Sign: <sign>
X-Sign-Version: hmac_sha256
X-Timestamp: <unix_timestamp>
```

**请求体：**
```json
{
  "biz_no": "12938475602193847560",
  "name": "张三",
  "id_card": "110101199001011234",
  "return_url": "https://yourdomain.com/certification/zjmf_mfcw/result?uid=1",
  "notify_url": "https://yourdomain.com/certification/zjmf_mfcw/callback?uid=1",
  "biz_extra_data": "{\"uid\":1}"
}
```

**响应：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "platform_biz_no": "46382671905182934716",
    "auth_url": "https://auth.finauth.com/verify?token=xxx",
    "expired_time": 1234567890,
    "expired_in": 900
  }
}
```

### 2. 查询认证结果 (POST /kyc/result)

**请求体：**
```json
{
  "platform_biz_no": "46382671905182934716"
}
```

**响应：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "platform_biz_no": "46382671905182934716",
    "biz_no": "12938475602193847560",
    "status": 2,
    "result_code": "1000",
    "result_message": "认证成功",
    "cost": 1.50
  }
}
```

### 3. 查询余额 (POST /kyc/balance/query)

**响应：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "balance": 100.50,
    "kyc_price": 1.50
  }
}
```

## 配置参数说明

### 系统字段

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| amount | text | 单次认证费用（元） | 0 |
| free | text | 免费认证次数 | 0 |

### 插件字段

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| api_url | text | ✅ | KYC系统API地址 |
| api_key | text | ✅ | API密钥 |
| api_secret | password | ✅ | API密钥签名（HMAC-SHA256） |
| auto_deduct | radio | - | 是否自动扣费 |
| require_verify | radio | - | 是否强制实名 |

## 错误处理

### 错误码对照表

| 错误码 | 说明 | 处理方式 |
|--------|------|----------|
| 0 | 成功 | 正常处理 |
| -1 | 系统错误 | 显示错误信息 |
| 400 | 参数错误 | 检查请求参数 |
| 401 | 未认证 | 检查API Key、Secret、签名、时间戳 |
| 403 | 权限不足 | 检查API权限 |
| 1006 | 余额不足 | 提示用户充值 |

### 异常处理流程

```php
try {
    // 调用KYC接口
    $result = $sdk->startKyc($params);
    
    if ($result['code'] === 0) {
        // 成功处理
    } else {
        // 业务错误处理
        updatePersonalCertifiStatus([
            'status' => 2,
            'auth_fail' => $result['message']
        ]);
    }
} catch (\Exception $e) {
    // 系统异常处理
    updatePersonalCertifiStatus([
        'status' => 2,
        'auth_fail' => '系统错误: ' . $e->getMessage()
    ]);
}
```

## 安全特性

### 1. HMAC-SHA256签名认证

- 使用HMAC-SHA256算法签名
- 时间戳防重放（±5分钟）
- API Secret不在网络传输，仅用于签名

### 2. HTTPS传输

- 所有API请求使用HTTPS
- 验证SSL证书有效性

### 3. 数据加密

- 身份证信息在KYC系统加密存储
- 使用AES-256-GCM加密算法

## 性能优化

### 1. 连接复用

```php
// 设置连接超时
curl_setopt($ch, CURLOPT_TIMEOUT, 30);
curl_setopt($ch, CURLOPT_CONNECTTIMEOUT, 10);
```

### 2. 状态轮询优化

- 避免频繁轮询
- 建议间隔3-5秒查询一次
- 最多轮询20次

### 3. 错误重试

- 网络错误自动重试1次
- 使用指数退避策略

## 测试方法

### 1. 单元测试

```php
// 测试余额查询
$sdk = new KycSdk($config);
$result = $sdk->queryBalance();
print_r($result);
```

### 2. 集成测试

```bash
# 测试API连接（查询余额，body 为空 JSON）
BODY='{}'
SIGN=$(printf '%s' "$BODY" | openssl dgst -sha256 -hmac "your_api_secret" | awk '{print $2}')
TS=$(date +%s)

curl -X POST "https://kyc.starloft.cn/api/v1/kyc/balance/query" \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: your_api_key" \
  -H "X-Sign: $SIGN" \
  -H "X-Sign-Version: hmac_sha256" \
  -H "X-Timestamp: $TS" \
  -d "$BODY"
```

### 3. 测试数据

```
姓名: 张三
身份证: 110101199001011234
```

## 部署建议

### 1. 环境要求

- PHP >= 7.0
- curl扩展
- json扩展
- SSL支持

### 2. 权限设置

```bash
# 设置文件权限
chmod 755 zjmf_mfcw/
chmod 644 zjmf_mfcw/*.php
chmod 644 zjmf_mfcw/config/*.php
```

### 3. 日志配置

建议启用PHP错误日志，便于排查问题：

```php
ini_set('log_errors', 'On');
ini_set('error_log', '/var/log/php_errors.log');
```

## 常见问题

### Q: 如何调试插件？

在插件代码中添加日志：

```php
error_log('KYC Request: ' . json_encode($params));
error_log('KYC Response: ' . json_encode($result));
```

### Q: 如何验证签名？

使用命令行工具核对签名：

```bash
printf '%s' "$BODY" | openssl dgst -sha256 -hmac "your_api_secret"
```

输出应与 `X-Sign` 头一致（均为小写十六进制）。

### Q: 认证费用如何计算？

KYC系统按次计费，每次认证扣除用户配置的单价（默认1.00元）。

## 技术支持

- **文档**: https://docs.starloft.cn
- **API参考**: [API接口文档](../../../docs/api/API接口文档-API_REFERENCE.md)
- **问题反馈**: https://github.com/starloft/kyc/issues
- **邮箱**: support@starloft.tech

## 更新记录

### v1.0.0 (2026-08-20)

- 首次发布
- 实现个人实名认证
- HMAC-SHA256签名认证
- 状态轮询机制
- 完善的错误处理

---

© 2026 StarLoft. All Rights Reserved.