# StarLoft 星楼云 · 云服务与实名认证平台

[![Version](https://img.shields.io/badge/version-1.12.0-blue.svg)](https://github.com/starloft/kyc)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.20+-00ADD8.svg)](https://golang.org/)
[![Docker](https://img.shields.io/badge/docker-20.10+-2496ED.svg)](https://www.docker.com/)

**星楼云** 是一个多产品云平台（对标腾讯云 / 阿里云架构）：平台门户 + 独立产品（实名认证、云服务器、短信服务等）+ 统一控制台与账户系统。当前已上线**实名认证**产品（身份证三要素 + 人脸识别），支持高并发、高可用部署，具有完善的 API 接口和插件生态。

---

## 🌟 核心特性

- 🚀 **多产品平台**: 门户聚合产品入口，每个产品有独立 URL 前缀（`/kyc` `/cs` `/sms`），横向可扩展新云产品
- 🔒 **安全可靠**: 多层安全防护，HMAC-SHA256 签名鉴权，敏感数据加密存储
- 🔌 **易于集成**: RESTful API + PHP插件（智简魔方财务版 / v10）
- 📊 **完善管理**: 后台 Dashboard + 用户/订单/资源包管理
- 💰 **灵活计费**: 资源包 + 余额组合计费（有资源包先扣资源包，否则按平台单价扣余额）；账户实名免费
- 🎨 **现代UI**: 腾讯云官网设计风格，移动端自适应
- 👤 **三要素认证**: 姓名 + 身份证 + 人脸识别（活体检测）

---

## 🏗️ 平台架构

```
www.starloft.cn       门户站点（frontend-portal）：平台首页 / 产品页 / 文档中心
console.starloft.cn   控制台（frontend）：登录注册、产品控制台、管理后台
└─ /api/ 反向代理      → backend 后端服务（Go + Gin）
```

- 门户聚合平台产品入口；每个产品页可通过 URL 访问（实名认证 `/kyc`、云服务器 `/cs`、短信服务 `/sms`）
- 门户与控制台共用同一账户体系，登录 / 注册 / 实名认证均位于控制台
- 旧域名 `kyc.starloft.cn` 端口 80 自动重定向至门户

---

## ⚡ 快速开始

### 使用Docker部署（推荐）

> 说明：MySQL 使用云数据库服务（不包含在容器编排中），容器仅包含 Redis、后端、前端（Nginx 多站点）。

```bash
# 1. 克隆代码
git clone https://github.com/starloft/kyc.git
cd kyc

# 2. 导入数据库初始化脚本（云数据库 MySQL 8.0+，分库架构）
# 脚本会先删除旧库再创建 4 个库（starloft_sys 系统库 / starloft_kyc 实名认证库 / starloft_cs、starloft_sms 预留产品库）与全部表结构，
# 并插入初始化数据（默认管理员、系统配置、资源包套餐）。
# 注意：脚本会 DROP DATABASE 永久删除旧库数据，仅限确认无存量数据需要保留时执行！
# 连接账号需拥有上述 4 个库的读写权限。
mysql -h <DB_HOST> -u <DB_USER> -p < init.sql

# 3. 配置环境变量
cp .env.example .env
nano .env  # 填写数据库/Redis/密钥/上游认证等配置

# 4. 放置证书（两个站点各一份）
# 将 www.starloft.cn 的证书私钥与完整证书链分别命名为 privkey.pem、fullchain.pem 放入 ./certs/www.starloft.cn/
# 将 console.starloft.cn 的证书放入 ./certs/console.starloft.cn/
# 证书由证书服务商（如腾讯云）签发后手动放入，可开启服务商自动续费，到期后替换对应子目录文件并重启前端加载
mkdir -p certs/www.starloft.cn certs/console.starloft.cn

# 5. 启动服务（前端镜像一次构建门户与控制台两个应用）
docker compose up -d

# 6. 访问系统
# 门户:  https://www.starloft.cn
# 控制台: https://console.starloft.cn
# API:   https://www.starloft.cn/api
# 文档:  https://www.starloft.cn/docs
```

### 新增一个云产品

1. 在 `frontend-portal/src/config/products.ts` 登记产品信息（key / 名称 / 特性 / 场景 / 控制台地址）
2. 在 `frontend-portal/src/router/index.ts` 注册对应产品路由（复用 `ProductPage.vue`，`meta.product` 指定 key）
3. 在 `frontend/src/router/index.ts` 为控制台增加对应产品路由与侧边栏入口
4. 重新构建前端镜像即可上线（`docker compose up -d --build frontend`）

---

## 📚 文档中心

- **[在线文档中心](https://www.starloft.cn/docs)** - 📑 门户内嵌文档（API v1 文档、插件教程），部署后可直接访问
- **[更新文档](更新文档.md)** - 📋 版本变更日志
- **API v1 文档** - 🔌 [在线版](https://www.starloft.cn/docs/api/v1)（鉴权方式、签名算法、接口说明）
- **插件文档（智简魔方财务版）** - 🔌 [plugin/zjmf_mfcw](plugin/zjmf_mfcw/README.md)
- **插件文档（智简魔方业务系统 v10）** - 🔌 [plugin/zjmf_v10](plugin/zjmf_v10/README.md)

---

## 🏗️ 技术架构

```
后端: Go 1.20+ + Gin框架
数据库: MySQL 8.0（云服务）
缓存: Redis 7.0
前端: Vue 3 + Vite（门户 frontend-portal / 控制台 frontend）
部署: Docker + Docker Compose + Nginx（TLS终止，多站点分发）
认证: FinAuth H5 Plus（三要素实名认证，HMAC-SHA256 签名）
短信/验证码: 腾讯云 SMS + 腾讯云天御验证码
支付: 支付宝（电脑网站支付）+ 微信支付（Native支付）
```

---

## 📦 项目结构

```
StarLoftKYC/
├── README.md              # 项目入口（本文件）
├── 更新文档.md            # 版本变更日志
├── init.sql               # 数据库初始化脚本（分库：创建 4 个库 + 全部表结构 + 初始化数据）
├── docker-compose.yml     # Docker编排
├── .env.example           # 环境变量模板
├── backend/               # Go后端服务
│   └── internal/
│       ├── handler/       # HTTP处理器
│       ├── service/       # 业务逻辑
│       ├── repository/    # 数据访问
│       ├── model/         # 数据模型
│       ├── router/        # 路由
│       └── upstream/      # 上游FinAuth/支付宝/微信支付客户端
├── frontend/              # 控制台前端（console.starloft.cn）
│   └── src/
│       ├── views/         # 页面（user用户端 / admin管理端）
│       ├── layouts/       # 布局
│       └── api/           # 接口封装
├── frontend-portal/       # 门户前端（www.starloft.cn）
│   └── src/
│       ├── config/        # 产品目录配置（products.ts）
│       ├── views/         # 页面（首页 / 产品页 / docs文档中心）
│       └── layouts/       # 站点布局
└── plugin/
    ├── zjmf_mfcw/         # 智简魔方财务版 PHP插件
    └── zjmf_v10/          # 智简魔方业务系统 v10 PHP插件
```

---

## 🚀 主要功能

### 平台门户（www.starloft.cn）
- ✅ 平台首页（产品聚合、平台优势）
- ✅ 产品页：实名认证（/kyc）、云服务器（/cs）、短信服务（/sms）
- ✅ 文档中心（/docs）：API 文档、插件教程

### 用户功能（console.starloft.cn）
- ✅ 手机号注册登录
- ✅ 账户实名认证（Web端免费，实名信息永久绑定不可修改；终身累计失败达上限后转为计费）
- ✅ 资源包购买（余额扣费，需先充值；库存限量）
- ✅ 在线充值（支付宝/微信支付）
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
- ✅ PHP 插件（智简魔方财务系统 / v10）

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
- 📖 文档: [在线文档中心](https://www.starloft.cn/docs)
- 🐛 问题反馈: [GitHub Issues](https://github.com/starloft/kyc/issues)

---

**版本**: v1.10.0  
**更新日期**: 2026-08-31  
**开发团队**: StarLoft Tech Team
