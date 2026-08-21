# StarLoft KYC 实名认证插件 - 使用文档

## 插件简介

StarLoft KYC 是一个对接 StarLoft KYC 系统的实名认证插件，适用于智简魔方（ZJMF）3.7.6+ 财务系统。支持身份证三要素实名认证（姓名+身份证+人脸识别），提供完整的API对接功能。

## 功能特性

- ✅ 身份证三要素实名认证（姓名+身份证+人脸识别）
- ✅ 自动对接KYC后端系统
- ✅ 支持自动扣费
- ✅ 支持强制实名认证
- ✅ 认证结果自动回调
- ✅ 支持认证状态轮询
- ✅ 安全的HMAC-SHA256签名认证
- ✅ 支持年龄限制（可配置最低实名年龄）

## 安装步骤

### 1. 上传插件文件

将 `starloft_kyc` 文件夹上传到以下目录：
```
/public/plugins/certification/starloft_kyc/
```

### 2. 确认文件结构

确保以下文件结构完整：
```
starloft_kyc/
├── StarloftKycPlugin.php       # 插件主类
├── README.md                   # 本文档
├── config/
│   └── config.php              # 配置文件
└── logic/
    └── KycSdk.php              # KYC SDK
```

### 3. 后台安装

1. 登录后台管理面板
2. 进入 `系统设置` → `实名认证设置` → `接口设置`
3. 找到 `StarLoft KYC实名认证`
4. 点击 `安装` 按钮

### 4. 配置插件

安装成功后，点击"配置"，填写以下信息：

| 配置项 | 必填 | 说明 | 示例 |
|--------|------|------|------|
| API地址 | ✅ | KYC系统的API地址 | `https://kyc.starloft.cn/api/v1` |
| API Key | ✅ | 在KYC系统后台获取 | `your_api_key_here` |
| API Secret | ✅ | 在KYC系统后台获取，用于HMAC签名 | `your_api_secret_here` |
| 单次认证费用 | - | 每次认证费用（元） | `0`（不扣费）或 `2.00` |
| 免费认证次数 | - | 每个用户免费次数 | `0`（无免费）或 `3` |
| 自动扣费 | - | 是否自动扣除费用 | 启用/禁用 |
| 强制实名 | - | 购买前是否必须实名 | 启用/禁用 |
| 最低实名年龄 | - | 要求的最低年龄（周岁），根据身份证出生日期计算，0表示不限 | `16` |

### 5. 获取API密钥

登录 StarLoft KYC 系统后台：
1. 进入 `用户中心`
2. 点击 `API密钥管理`
3. 复制 API Key 和 API Secret

### 6. 测试连接

配置完成后，可以通过查询余额接口测试连接是否正常。

## 使用方法

### 用户端使用

**访问地址：**
```
https://你的域名/certification/starloft_kyc
```

**认证流程：**
1. 用户登录后访问实名认证页面
2. 填写真实姓名和身份证号
3. 点击提交认证
4. 跳转到KYC认证页面进行人脸识别
5. 完成认证后自动返回

### 认证状态

- **待认证(0)**: 尚未提交认证
- **认证中(4)**: 认证请求已提交，正在处理
- **认证成功(1)**: 认证通过
- **认证失败(2)**: 认证未通过，可重新提交

## API接口说明

插件对接以下KYC系统API，所有接口均使用 **API Key + HMAC-SHA256 签名** 鉴权。

### 请求鉴权（每个请求必带）

每个请求需携带以下 4 个请求头：

```
X-Api-Key: <你的API Key>
X-Sign: <签名>
X-Sign-Version: hmac_sha256
X-Timestamp: <当前Unix时间戳（秒）>
```

**签名算法：**

```
sign = hex(HMAC-SHA256(api_secret, 原始请求体))
```

其中「原始请求体」为 POST 发送的 JSON 字符串（GET 请求为当前查询字符串），与 `CURLOPT_POSTFIELDS` 完全一致。PHP 写法：

```php
$sign = hash_hmac('sha256', $body, $this->apiSecret); // 小写十六进制
```

### 1. 创建认证订单

```
POST /api/v1/kyc/start

请求头:
Content-Type: application/json
X-Api-Key: <your_api_key>
X-Sign: <sign>
X-Sign-Version: hmac_sha256
X-Timestamp: <unix_timestamp>

请求参数:
{
    "biz_no": "ZJMF20260119001",           // 业务订单号（唯一）
    "name": "张三",                        // 真实姓名
    "id_card": "110101199001011234",      // 身份证号
    "return_url": "https://yourdomain.com/certification/starloft_kyc/result",
    "notify_url": "https://yourdomain.com/certification/starloft_kyc/callback",
    "biz_extra_data": "{\"uid\":1001}"    // 业务扩展数据
}

响应数据:
{
    "code": 0,
    "message": "success",
    "data": {
        "platform_biz_no": "ZJMF20260119001_1001_1234567890",  // 平台流水号
        "auth_url": "https://auth.finauth.com/verify?token=xxx", // 认证页面URL
        "expired_time": 1234567890,        // 过期时间（Unix秒）
        "expired_in": 900                  // 有效期（秒）
    }
}
```

### 2. 查询认证结果

```
POST /api/v1/kyc/result

请求头:
Content-Type: application/json
X-Api-Key: <your_api_key>
X-Sign: <sign>
X-Sign-Version: hmac_sha256
X-Timestamp: <unix_timestamp>

请求参数:
{
    "platform_biz_no": "ZJMF20260119001_1001_1234567890"  // 平台流水号
}

响应数据:
{
    "code": 0,
    "message": "success",
    "data": {
        "platform_biz_no": "ZJMF20260119001_1001_1234567890",
        "biz_no": "ZJMF20260119001",
        "status": 2,                      // 0待认证 1认证中 2成功 3失败 4已取消 5超时
        "result_code": "1000",
        "result_message": "认证成功",
        "cost": 1.50
    }
}
```

### 3. 查询用户余额

```
POST /api/v1/kyc/balance/query

请求头:
Content-Type: application/json
X-Api-Key: <your_api_key>
X-Sign: <sign>
X-Sign-Version: hmac_sha256
X-Timestamp: <unix_timestamp>

响应数据:
{
    "code": 0,
    "message": "success",
    "data": {
        "balance": 100.50,
        "kyc_price": 1.50
    }
}
```

## 常见问题

### Q1: API连接失败怎么办？

**解决方法：**
1. 检查API地址是否正确（注意HTTP/HTTPS）
2. 检查API Key和Secret是否正确
3. 确认KYC后端服务是否正常运行
4. 检查服务器防火墙设置
5. 确认服务器可以访问KYC系统

**测试命令（以查询余额为例，body 为空 JSON）：**
```bash
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

### Q2: 认证一直显示"认证中"状态？

**原因：**
- 用户未完成人脸识别
- 回调地址不可访问
- 网络延迟

**解决方案：**
1. 等待用户完成认证
2. 系统会自动轮询查询状态
3. 检查notify_url是否可以从外网访问

### Q3: 如何设置强制实名？

在插件配置中启用"强制实名"，用户在购买产品前必须完成实名认证。

### Q4: 认证费用如何计算？

- 费用由KYC系统设定
- 可在插件配置中设置每次认证费用
- 支持设置免费认证次数
- 可在KYC后台查看当前单价

### Q5: 如何调试插件？

1. 检查PHP错误日志
2. 查看curl请求是否成功
3. 打印$result变量查看API响应
4. 确认签名是否正确、时间戳是否在有效期内

### Q6: 如何设置最低实名年龄？

在插件配置中填写「最低实名年龄」（默认 `16` 周岁）。提交认证时，插件会根据身份证号中的出生日期自动计算年龄，未达到设定年龄的用户将被拒绝，无法发起认证；设置为 `0` 表示不限年龄。

## 安全建议

1. **保护API密钥**: 
   - 不要将API Secret泄露给他人
   - 不要提交到Git仓库

2. **定期更换密钥**: 
   - 建议每3-6个月更换一次API密钥

3. **使用HTTPS**: 
   - 生产环境务必使用HTTPS协议
   - 确保回调地址也是HTTPS

4. **访问控制**: 
   - 限制后台管理页面的访问权限
   - 启用IP白名单（如果KYC系统支持）

5. **数据加密**: 
   - 敏感数据传输使用HTTPS
   - KYC系统会对身份信息加密存储

## 技术支持

- **插件版本**: v1.0.0
- **作者**: StarLoft
- **更新日期**: 2026-08-20
- **兼容版本**: 智简魔方 3.7.6+
- **文档**: https://docs.starloft.cn/kyc/plugin
- **问题反馈**: https://github.com/starloft/kyc/issues

## 更新日志

### v1.0.0 (2026-08-20)
- ✨ 首次发布
- ✅ 身份证三要素实名认证
- ✅ 自动对接KYC系统
- ✅ 支持认证状态轮询
- ✅ HMAC-SHA256签名认证
- ✅ 完善的错误处理

## 许可协议

本插件遵循 MIT 许可协议。

---

© 2026 StarLoft. All Rights Reserved.