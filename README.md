# StarLoft KYC 实名认证系统

[![Version](https://img.shields.io/badge/version-1.5.7-blue.svg)](https://github.com/starloft/kyc)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.20+-00ADD8.svg)](https://golang.org/)
[![Docker](https://img.shields.io/badge/docker-20.10+-2496ED.svg)](https://www.docker.com/)

**StarLoft KYC** 是一个专业的企业级实名认证系统，提供身份证三要素实名验证服务（姓名+身份证+人脸识别），支持高并发、高可用部署，具有完善的API接口和插件生态。

---

## 🌟 核心特性

- 🚀 **高性能**: Go语言开发，支持高并发
- 🔒 **安全可靠**: 多层安全防护，HMAC-SHA256 签名鉴权，敏感数据加密存储
- 🔌 **易于集成**: RESTful API + PHP插件（智简魔方财务版）
- 📊 **完善管理**: 后台 Dashboard + 用户/订单/资源包管理
- 💰 **灵活计费**: 资源包 + 余额组合计费（有资源包先扣资源包，否则按平台单价扣余额）；账户实名免费
- 🎨 **现代UI**: 深色主题设计，移动端自适应
- 👤 **三要素认证**: 姓名 + 身份证 + 人脸识别（活体检测）

---

## ⚡ 快速开始

### 使用Docker部署（推荐）

> 说明：MySQL 使用云数据库服务（不包含在容器编排中），容器仅包含 Redis、后端、前端（Nginx）。

```bash
# 1. 克隆代码
git clone https://github.com/starloft/kyc.git
cd kyc

# 2. 导入数据库初始化脚本（云数据库 MySQL 8.0+）
mysql -h <DB_HOST> -u <DB_USER> -p <DB_NAME> < init.sql

# 3. 配置环境变量
cp .env.example .env
nano .env  # 填写数据库/Redis/密钥/上游认证等配置

# 4. 放置 SSL 证书（用于 HTTPS）
mkdir -p certs
# 将证书分别命名为 fullchain.pem 和 privkey.pem 放入 certs/

# 5. 启动服务
docker-compose up -d

# 6. 访问系统
# 官网: https://kyc.starloft.cn
# API:  https://kyc.starloft.cn/api
# 文档: https://kyc.starloft.cn/docs
```

---

## 📚 文档中心

- **[在线文档中心](/docs)** - 📑 平台内嵌文档（API v1 文档、插件教程），部署后可直接访问
- **[更新文档](更新文档.md)** - 📋 版本变更日志
- **API v1 文档** - 🔌 [在线版](/docs/api/v1)（鉴权方式、签名算法、接口说明）
- **插件文档（智简魔方财务版）** - 🔌 [plugin/zjmf_mfcw](plugin/zjmf_mfcw/README.md)
- **插件文档（智简魔方业务系统 v10）** - 🔌 [plugin/zjmf_v10](plugin/zjmf_v10/README.md)

---

## 🏗️ 技术架构

```
后端: Go 1.20+ + Gin框架
数据库: MySQL 8.0（云服务）
缓存: Redis 7.0
前端: Vue 3 + Element Plus
部署: Docker + Docker Compose + Nginx（TLS终止）
认证: FinAuth H5 Plus（三要素实名认证，HMAC-SHA256 签名）
短信/验证码: 腾讯云 SMS + 腾讯云天御验证码
支付: 银联天天付
```

---

## 📦 项目结构

```
StarLoftKYC/
├── README.md              # 项目入口（本文件）
├── 更新文档.md            # 版本变更日志
├── init.sql               # 数据库初始化脚本
├── docker-compose.yml     # Docker编排
├── .env.example           # 环境变量模板
├── backend/               # Go后端服务
│   └── internal/
│       ├── handler/       # HTTP处理器
│       ├── service/       # 业务逻辑
│       ├── repository/    # 数据访问
│       ├── model/         # 数据模型
│       ├── router/        # 路由
│       └── upstream/      # 上游FinAuth/银联客户端
├── frontend/              # Vue前端
│   └── src/
│       ├── views/         # 页面（user用户端 / admin管理端 / docs文档中心）
│       └── api/           # 接口封装
└── plugin/
    ├── zjmf_mfcw/         # 智简魔方财务版 PHP插件
    └── zjmf_v10/          # 智简魔方业务系统 v10 PHP插件
```

---

## 🚀 主要功能

### 用户功能
- ✅ 手机号注册登录
- ✅ 账户实名认证（Web端免费，实名信息永久绑定不可修改；终身累计失败达上限后转为计费）
- ✅ 资源包购买（余额扣费，需先充值；库存限量）
- ✅ 在线充值（银联支付）
- ✅ 余额查询、消费记录
- ✅ API 密钥管理（Key 注册后自动生成，Secret 实名认证后下发）

### 管理后台
- ✅ 数据统计 Dashboard
- ✅ 用户管理（搜索/详情/状态管理）
- ✅ 认证订单管理（订单详情/失败原因）
- ✅ 资源包管理（创建/库存/上架下架）
- ✅ 人工充值（需填写银行流水单号）
- ✅ 系统配置（KYC 单价等）

### 开发者功能
- ✅ RESTful API 接口（API Key + HMAC-SHA256 签名）
- ✅ Webhook 回调通知（notify_url）
- ✅ PHP 插件（智简魔方财务系统）

---

## 🔐 安全特性

- ✅ SQL注入防护
- ✅ API限流保护
- ✅ 敏感数据AES-256-GCM加密
- ✅ 密码bcrypt加密存储
- ✅ 数据库TLS连接
- ✅ JWT Token认证
- ✅ HMAC-SHA256 请求签名校验（防篡改/防重放）
- ✅ 订单权限验证

---

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

---

## 📞 技术支持

- 📧 邮箱: support@starloft.tech
- 📖 文档: [在线文档中心](/docs)
- 🐛 问题反馈: [GitHub Issues](https://github.com/starloft/kyc/issues)

---

**版本**: v1.5.4  
**更新日期**: 2026-08-28  
**开发团队**: StarLoft Tech Team
