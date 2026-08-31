# StarLoft KYC 实名认证插件（智简魔方业务系统 v10）

## 插件简介

对接 StarLoft KYC 系统的实名认证插件，适用于**智简魔方业务系统 v10**。支持身份证三要素实名认证（姓名+身份证+人脸识别），提供完整的 API 对接、状态轮询与异步回调功能。

## 功能特性

- ✅ 身份证三要素实名认证（姓名+身份证+人脸识别）
- ✅ 按 v10 实名认证接口规范开发（`ZjmfV10Person` / `ZjmfV10CollectionInfo`）
- ✅ 支持自动扣费 / 免费认证次数
- ✅ 支持强制实名认证（由 v10 后台开启）
- ✅ 认证结果异步回调（notify_url）+ 前台状态轮询
- ✅ 安全的 HMAC-SHA256 签名认证
- ✅ 支持年龄限制（可配置最低实名年龄）
- ✅ 幂等保护：认证中任务自动复用，避免重复发单扣费
- ✅ 轮询上限保护：网络异常/任务丢失时自动终止，杜绝无限轮询

## 目录结构

```
zjmf_v10/
├── ZjmfV10.php                    # 插件入口文件（命名空间 certification\zjmf_v10）
├── config.php                     # 插件配置项（后台「实名认证 → 接口管理 → 配置」展示）
├── controller/
│   └── IndexController.php        # 外部回调控制器
│       ├── notifyHandle()         # 异步通知处理（KYC 平台推送认证结果）
│       ├── result()               # 认证完成回跳页
│       └── status()               # 状态查询（AJAX 轮询）
└── logic/
    └── KycSdk.php                 # StarLoft KYC SDK（API 通信 + HMAC 签名 + 错误分类）
```

## 安装步骤

### 1. 上传插件文件

将 `zjmf_v10` 文件夹上传到：

```
/public/plugins/certification/zjmf_v10/
```

### 2. 后台安装

1. 登录 v10 管理后台
2. 进入 `实名认证` → `接口管理`
3. 找到 `StarLoft KYC实名认证`
4. 点击 `安装`，然后点击 `配置`

### 3. 配置插件

| 配置项 | 必填 | 说明 | 示例 |
|--------|------|------|------|
| API地址 | ✅ | StarLoft KYC 系统的 API 地址 | `https://www.starloft.cn/api/v1` |
| API Key | ✅ | 在 KYC 系统后台获取 | `your_api_key_here` |
| API Secret | ✅ | 在 KYC 系统后台获取，用于 HMAC 签名 | `your_api_secret_here` |
| 单次认证费用 | - | 每次认证费用（元） | `0`（不扣费）或 `2.00` |
| 免费认证次数 | - | 每个用户免费次数 | `0`（无免费）或 `3` |
| 最低实名年龄 | - | 要求的最低年龄（周岁），0 表示不限 | `16` |
| 认证完成回跳地址 | - | 用户完成认证后浏览器回跳地址；留空使用插件内置结果页 | 可留空 |

> 说明：`amount`（单次认证费用）与 `free`（免费认证次数）为 v10 实名认证系统必需字段，由 v10 后台统一处理扣费逻辑。

### 4. 获取 API 密钥

登录 StarLoft KYC 系统后台：

1. 完成账户实名认证
2. 进入 `用户中心` → `API密钥管理`
3. 复制 API Key 和 API Secret

## 使用方法

### 用户实名认证

用户在 v10 会员中心进入 `实名认证` 页面：

1. 选择 `StarLoft KYC实名认证` 认证方式
2. 填写姓名和身份证号
3. 点击提交，插件创建认证任务并跳转到认证页面
4. 完成人脸识别
5. 认证结果自动通过回调/轮询同步，用户返回会员中心查看状态

### 认证状态说明

| 状态 | 含义 |
|------|------|
| 0 | 待认证 |
| 1 | 认证通过 |
| 2 | 认证失败 |
| 4 | 认证中 |

## 回调地址

插件内置两个外部访问地址：

| 地址 | 用途 |
|------|------|
| `/certification/zjmf_v10/index/notifyHandle` | 异步通知（StarLoft KYC 平台推送认证结果） |
| `/certification/zjmf_v10/index/result` | 认证完成回跳页 |

> `notify_url` 由插件在创建认证任务时自动生成并传给 KYC 平台，无需手工配置。

### 异步通知格式

StarLoft KYC 平台在认证终态时向 `notify_url` 推送：

```
POST /certification/zjmf_v10/index/notifyHandle
Content-Type: application/json

{
    "biz_no":          "12938475602193847560",
    "platform_biz_no": "46382671905182934716",
    "status":          2,
    "result_code":     "1000",
    "result_message":  "认证成功",
    "cost":            1.50,
    "sign":            "HMAC-SHA256 签名（防伪造，插件必校验）"
}
```

字段说明：

- `biz_no`：下游业务订单号
- `platform_biz_no`：平台流水号（插件保存为认证记录的 `certify_id`）
- `status`：0待认证 1认证中 2成功 3失败 4已取消 5超时
- `result_code` / `result_message`：上游结果码与说明
- `cost`：本次认证扣费金额
- `sign`：回调签名（HMAC-SHA256），插件校验不通过会拒绝（返回 401），防止伪造回调

**签名算法**（与插件 `verifyNotifySign()` 一致）：

1. 取 `biz_no` / `cost` / `platform_biz_no` / `result_code` / `result_message` / `status` 六个字段
2. 按 key 字典序拼接为原始字符串（不做 URL 编码）：`k1=v1&k2=v2&...`
3. `sign = 小写hex( HMAC-SHA256(api_secret, 拼接串) )`

其中 `cost` 固定保留两位小数（如 `1.50`），`status` 为整数。`api_secret` 即插件配置的 API Secret。

插件收到通知后先校验 `sign`，校验通过再按 `platform_biz_no` 定位认证记录并更新状态：
`status=2 → 通过`、`status=3/4/5 → 失败`。

## API 接口说明

插件通过 `logic/KycSdk.php` 对接以下 StarLoft KYC API，所有请求使用 **API Key + HMAC-SHA256 签名** 鉴权。

### 请求鉴权（每个请求必带 4 个请求头）

```
X-Api-Key: <你的API Key>
X-Sign: <签名>
X-Sign-Version: hmac_sha256
X-Timestamp: <当前Unix时间戳（秒）>
```

签名算法：

```
sign = hex(HMAC-SHA256(api_secret, 原始请求体))
```

PHP 写法：

```php
$sign = hash_hmac('sha256', $body, $this->apiSecret); // 小写十六进制
```

### 1. 创建认证订单

```
POST /api/v1/kyc/start

{
    "biz_no":         "12938475602193847560",   // 业务订单号（唯一）
    "name":           "张三",                              // 真实姓名
    "id_card":        "110101199001011234",               // 身份证号
    "return_url":     "https://yourdomain.com/certification/zjmf_v10/index/result",
    "notify_url":     "https://yourdomain.com/certification/zjmf_v10/index/notifyHandle",
    "biz_extra_data": "{\"uid\":1}"                        // 业务扩展数据
}
```

响应：

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "platform_biz_no": "46382671905182934716",
        "auth_url": "https://auth.finauth.com/verify?token=xxx",
        "expired_time": 1737000900,
        "expired_in": 900
    }
}
```

### 2. 查询认证结果

```
POST /api/v1/kyc/result

{
    "platform_biz_no": "46382671905182934716"
}
```

响应：

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

### 3. 查询用户余额

```
POST /api/v1/kyc/balance/query

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

### Q1: 安装后找不到插件？

- 确认目录名正确：`/public/plugins/certification/zjmf_v10/`
- 确认入口文件名正确：`ZjmfV10.php`（目录名大驼峰 + .php）
- 清空 v10 运行时缓存后重新进入 `实名认证 → 接口管理`

### Q2: 认证一直显示"认证中"？

- 用户未完成人脸识别
- 检查 `notify_url`（`/certification/zjmf_v10/index/notifyHandle`）能否从外网访问
- 插件会按 2s/次自动轮询查询状态，认证完成后自动更新
- 若长时间无结果，可检查 KYC 后台该订单的状态

### Q3: API 连接失败 / 鉴权失败？

- 检查 API 地址是否正确（包含 `/api/v1`）
- 检查 API Key / Secret 是否正确
- 确认 KYC 系统后台已完成实名认证并生成 API 密钥
- 确认服务器可以访问 KYC 系统、时间戳与服务器时间同步（±5 分钟内）

### Q4: 认证费用如何计算？

- 费用由 KYC 系统设定（单次认证费用）
- 插件配置的 `amount` / `free` 由 v10 系统用于用户扣费 / 免费次数控制
- 可在 KYC 后台查看当前单价

## 技术支持

- **插件版本**: v1.0.0
- **作者**: StarLoft
- **兼容版本**: 智简魔方业务系统 v10
- **文档**: https://docs.starloft.cn/kyc/plugin/v10
- **问题反馈**: https://github.com/starloft/kyc/issues

## 许可协议

本插件遵循 MIT 许可协议。

---

© 2026 StarLoft. All Rights Reserved.
