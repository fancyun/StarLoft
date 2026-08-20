# StarLoft KYC 实名认证系统

[![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)](https://github.com/starloft/kyc)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.20+-00ADD8.svg)](https://golang.org/)
[![Docker](https://img.shields.io/badge/docker-20.10+-2496ED.svg)](https://www.docker.com/)

**StarLoft KYC** 是一个专业的企业级实名认证系统，提供身份证三要素实名验证服务（姓名+身份证+人脸识别），支持高并发、高可用部署，具有完善的API接口和插件生态。

---

## 🌟 核心特性

- 🚀 **高性能**: Go语言开发，支持10000+ QPS
- 🔒 **安全可靠**: 多层安全防护，敏感数据加密存储
- 🔌 **易于集成**: RESTful API + 多语言SDK + PHP插件
- 📊 **完善管理**: Dashboard + 财务统计 + 数据导出
- 💰 **灵活计费**: 按量付费，支持个性化单价
- 🎨 **现代UI**: 深色主题设计，移动端自适应
- 👤 **三要素认证**: 姓名 + 身份证 + 人脸识别（活体检测）

---

## ⚡ 快速开始

### 使用Docker部署（推荐）

```bash
# 1. 克隆代码
git clone https://github.com/starloft/kyc.git
cd kyc

# 2. 配置环境变量
cp .env.example .env
nano .env  # 修改密码和密钥

# 3. 启动服务
docker-compose up -d

# 4. 访问系统
# 官网: https://kyc.starloft.cn
# API: https://kyc.starloft.cn/api
```

📖 **完整文档**: [docs/文档导航-README.md](docs/文档导航-README.md)

---

## 📚 文档中心

所有项目文档都在 `docs/` 目录下：

- **[文档导航](docs/文档导航-README.md)** - 📑 所有文档索引
- **[快速开始](docs/快速开始-QUICK_START.md)** - 🚀 5分钟部署指南
- **[项目概览](docs/项目概览-PROJECT_OVERVIEW.md)** - 📖 功能和架构介绍
- **[API接口文档](docs/api/API接口文档-API_REFERENCE.md)** - 🔌 完整API参考
- **[部署指南](docs/deployment/部署指南-DEPLOYMENT_GUIDE.md)** - 🐳 生产环境部署
- **[安全配置指南](docs/security/安全配置指南-SECURITY_CONFIG_GUIDE.md)** - 🔒 安全最佳实践

---

## 🏗️ 技术架构

```
后端: Go 1.20+ + Gin框架
数据库: MySQL 8.0 + Redis 7.0
前端: Vue 3 + Element Plus
部署: Docker + Docker Compose + Nginx
认证: FinAuth H5 Plus（三要素实名认证）
```

---

## 📦 项目结构

```
StarLoftKYC/
├── README.md              # 项目入口（本文件）
├── init.sql               # 数据库初始化
├── docker-compose.yml     # Docker编排
├── backend/               # Go后端服务
├── frontend/              # Vue前端
├── 3.7.6/                 # PHP插件（智简魔方）
└── docs/                  # 📚 完整文档目录
    ├── 文档导航-README.md
    ├── 快速开始-QUICK_START.md
    ├── 项目概览-PROJECT_OVERVIEW.md
    ├── 变更日志-CHANGELOG.md
    ├── api/               # API文档
    ├── architecture/      # 架构设计
    ├── backend/           # 后端文档
    ├── frontend/          # 前端文档
    ├── deployment/        # 部署指南
    ├── plugin/            # 插件文档
    └── security/          # 安全文档
```

---

## 🚀 主要功能

### 用户功能
- ✅ 手机号注册登录
- ✅ 身份证三要素实名认证（姓名+身份证+人脸）
- ✅ 在线充值（支付宝/微信/银联）
- ✅ 余额查询和消费记录
- ✅ 退款申请

### 管理后台
- ✅ 数据统计Dashboard
- ✅ 用户管理（搜索/详情/单价调整）
- ✅ 订单管理（认证订单/充值订单）
- ✅ 退款审批
- ✅ 财务统计和报表
- ✅ 人工充值和赠送

### 开发者功能
- ✅ RESTful API接口
- ✅ API Key认证
- ✅ Webhook回调通知
- ✅ PHP插件（智简魔方财务系统）

---

## 🔐 安全特性

- ✅ SQL注入防护
- ✅ API限流保护
- ✅ 敏感数据AES-256-GCM加密
- ✅ 密码bcrypt加密存储
- ✅ 数据库TLS连接
- ✅ JWT Token认证
- ✅ 订单权限验证

---

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

---

## 📞 技术支持

- 📧 邮箱: support@starloft.tech
- 📖 文档: [docs/文档导航-README.md](docs/文档导航-README.md)
- 🐛 问题反馈: [GitHub Issues](https://github.com/starloft/kyc/issues)

---

**版本**: v1.0.0  
**更新日期**: 2026-08-19  
**开发团队**: StarLoft Tech Team
