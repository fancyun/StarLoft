# StarLoft KYC 插件（v10） - 快速开始指南

## 📦 安装步骤

### 1. 上传插件

将 `zjmf_v10` 文件夹上传到：

```
/public/plugins/certification/zjmf_v10/
```

### 2. 后台安装

1. 登录 v10 管理后台
2. 进入 `实名认证` → `接口管理`
3. 找到 `StarLoft KYC实名认证`
4. 点击 `安装`

### 3. 配置插件

点击 `配置`，填写以下信息：

```
API地址: https://kyc.starloft.cn/api/v1
API Key: your_api_key_here
API Secret: your_api_secret_here
```

### 4. 获取 API 密钥

登录 [KYC系统](https://kyc.starloft.cn/admin/login)：

- 完成账户实名认证
- 进入 `用户中心` → `API密钥管理`
- 复制 API Key 和 API Secret

### 5. 测试连接

配置完成后可通过余额查询接口测试连接。

## 🚀 使用方法

### 用户实名认证

用户在 v10 会员中心 `实名认证` 页面：

1. 选择 `StarLoft KYC实名认证` 方式
2. 填写姓名和身份证号
3. 提交后跳转到认证页面完成人脸识别
4. 认证结果自动同步，返回会员中心查看状态

## ⚙️ 配置说明

### 常见配置场景

**场景1：仅提供服务，不强制**
```
单次认证费用: 0 或 2.00
自动扣费: 启用
强制实名: 禁用
```

**场景2：购买前必须实名**
```
单次认证费用: 0
自动扣费: 启用
强制实名: 启用  ← 关键配置
```

**场景3：提供免费次数**
```
单次认证费用: 2.00
免费认证次数: 3  ← 每人3次免费
自动扣费: 启用
强制实名: 禁用
```

## 🔧 故障排查

### API 连接失败

**检查清单：**
- ✅ API 地址正确（包含 `/api/v1`）
- ✅ API Key 和 Secret 正确
- ✅ KYC 系统运行正常
- ✅ 服务器可访问 KYC 系统
- ✅ 服务器时间与标准时间同步（签名时间戳 ±5 分钟）

**测试方法：**
```bash
curl -I https://kyc.starloft.cn/api/v1/
```

### 认证一直显示"认证中"

**原因：**
- 用户未完成认证
- 回调地址不可访问

**解决：**
- 系统会自动轮询查询状态
- 确认 `/certification/zjmf_v10/index/notifyHandle` 可从外网访问
- 用户完成认证后会自动更新

## 📚 文档链接

- [完整使用文档](README.md)
- [API接口文档](https://docs.starloft.cn/api)
- [KYC系统文档](https://docs.starloft.cn)

## 💡 技术支持

- 版本: v1.0.0
- 作者: StarLoft
- 日期: 2026-08-27
- 邮箱: support@starloft.tech

---

**祝使用愉快！** 🎉
