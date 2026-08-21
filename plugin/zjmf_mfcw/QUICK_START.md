# StarLoft KYC 插件 - 快速开始指南

## 📦 安装步骤

### 1. 上传插件

将 `starloft_kyc` 文件夹上传到：
```
/public/plugins/certification/starloft_kyc/
```

### 2. 后台安装

1. 登录后台：`系统设置` → `实名认证设置` → `接口设置`
2. 找到 `StarLoft KYC实名认证`
3. 点击 `安装`

### 3. 配置插件

点击"配置"，填写以下信息：

```
API地址: https://kyc.starloft.cn/api/v1
API Key: your_api_key_here
API Secret: your_api_secret_here
```

### 4. 获取API密钥

登录 [KYC系统](https://kyc.starloft.cn/admin/login)：
- 进入 `用户中心` → `API密钥管理`
- 复制 API Key 和 API Secret

### 5. 测试连接

配置完成后可通过余额查询接口测试连接。

## 🚀 使用方法

### 用户实名认证

访问地址：
```
https://你的域名/certification/starloft_kyc
```

认证流程：
1. 填写姓名和身份证号
2. 提交后跳转到认证页面
3. 完成人脸识别
4. 自动返回并更新状态

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

### API连接失败

**检查清单：**
- ✅ API地址正确（包含 `/api/v1`）
- ✅ API Key和Secret正确
- ✅ KYC系统运行正常
- ✅ 服务器可访问KYC系统

**测试方法：**
```bash
curl -I https://kyc.starloft.cn/api/v1/
```

### 认证一直显示"认证中"

**原因：**
- 用户未完成认证
- 认证接口响应慢

**解决：**
- 系统会自动轮询查询状态
- 用户完成认证后会自动更新

## 📚 文档链接

- [完整使用文档](README.md)
- [API接口文档](https://docs.starloft.cn/api)
- [KYC系统文档](https://docs.starloft.cn)

## 💡 技术支持

- 版本: v1.0.0
- 作者: StarLoft
- 日期: 2026-08-20
- 邮箱: support@starloft.tech

---

**祝使用愉快！** 🎉
