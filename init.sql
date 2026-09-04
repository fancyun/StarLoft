-- ============================================================
-- StarLoft 星楼云 · 数据库初始化脚本（分库架构）
-- ============================================================
-- 分库设计：同一 MySQL 实例下的多个 database
--   starloft_sys  系统库：用户/管理员/实名记录/余额流水/充值订单/系统配置/登录日志
--   starloft_kyc  实名认证产品库：认证订单/资源包
--   starloft_cs   云服务器产品库（预留，暂不建表）
--   starloft_sms  短信服务产品库（预留，暂不建表）
--
-- 使用方式：
--   1. 本脚本会先删除旧库再创建新库与全部表结构，并插入初始化数据，直接执行即可彻底重建
--   2. 脚本包含完整的建库、建表与初始化数据（删除旧库后重建，避免旧表/旧字段残留）
--   3. 数据库账号需拥有上述 4 个库的读写权限（在云数据库控制台创建库并授权）
--   4. 后端启动时的 GORM AutoMigrate 会自动同步表结构（只增不减），两者不冲突
--
-- 注意：本脚本会 DROP DATABASE 永久删除旧库数据！仅限确认无存量数据需要保留时执行！
-- ============================================================

SET NAMES utf8mb4;

-- ------------------------------------------------------------
-- 0. 删除旧库以彻底重建（避免旧表/旧字段残留，AutoMigrate 只增不减无法清理）
-- ------------------------------------------------------------
DROP DATABASE IF EXISTS `starloft_sys`;
DROP DATABASE IF EXISTS `starloft_kyc`;
DROP DATABASE IF EXISTS `starloft_cs`;
DROP DATABASE IF EXISTS `starloft_sms`;

-- ------------------------------------------------------------
-- 1. 创建数据库
-- ------------------------------------------------------------
CREATE DATABASE IF NOT EXISTS `starloft_sys` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
CREATE DATABASE IF NOT EXISTS `starloft_kyc` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
CREATE DATABASE IF NOT EXISTS `starloft_cs`  DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
CREATE DATABASE IF NOT EXISTS `starloft_sms` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;

-- ============================================================
-- 系统库 starloft_sys
-- ============================================================
USE `starloft_sys`;

-- 平台用户
CREATE TABLE IF NOT EXISTS `platform_user` (
  `id`             bigint       NOT NULL AUTO_INCREMENT,
  `phone`          varchar(20)  NOT NULL,
  `username`       varchar(50)  NOT NULL,
  `email`          varchar(100) NOT NULL,
  `password_hash`  varchar(255) NOT NULL,
  `balance`        decimal(10,2) NOT NULL DEFAULT 0,
  `api_key`        varchar(64)  NOT NULL,
  `api_secret`     varchar(64)  NOT NULL,
  `is_kyc_verified` tinyint     NOT NULL DEFAULT 0,
  `kyc_name`       varchar(100) NULL,
  `kyc_id_card`    varchar(100) NULL,
  `kyc_price`      decimal(10,2) NOT NULL DEFAULT 1 COMMENT '已废弃的个人KYC单价，仅保留字段兼容',
  `status`         tinyint      NOT NULL DEFAULT 1,
  `last_login_at`  datetime(3)  NULL,
  `created_at`     datetime(3)  NOT NULL,
  `updated_at`     datetime(3)  NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_platform_user_phone` (`phone`),
  UNIQUE KEY `idx_platform_user_username` (`username`),
  UNIQUE KEY `idx_platform_user_email` (`email`),
  UNIQUE KEY `idx_platform_user_api_key` (`api_key`),
  KEY `idx_platform_user_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='平台用户';

-- 管理员
CREATE TABLE IF NOT EXISTS `admin_user` (
  `id`            bigint      NOT NULL AUTO_INCREMENT,
  `username`      varchar(50) NOT NULL,
  `password_hash` varchar(255) NOT NULL,
  `nickname`      varchar(50) NULL,
  `status`        tinyint     NOT NULL DEFAULT 1,
  `last_login_at` datetime(3) NULL,
  `created_at`    datetime(3) NOT NULL,
  `updated_at`    datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_admin_user_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员';

-- 账户实名认证记录
CREATE TABLE IF NOT EXISTS `kyc_record` (
  `id`            bigint       NOT NULL AUTO_INCREMENT,
  `user_id`       bigint       NOT NULL,
  `source`        tinyint      NOT NULL DEFAULT 2 COMMENT '来源：1-账户实名 2-API调用',
  `auth_order_id` bigint       NULL,
  `biz_no`        varchar(50)  NOT NULL COMMENT '全平台唯一业务流水号（平台调用时随机生成）',
  `return_url`    varchar(500) NULL,
  `notify_url`    varchar(500) NULL,
  `biz_extra_data` text        NULL,
  `up_token`      varchar(100) NULL,
  `up_biz_id`     varchar(50) NULL,
  `up_request_id` varchar(50) NULL,
  `name`          varchar(50) NOT NULL,
  `id_card`       varchar(18) NOT NULL,
  `status`        tinyint      NOT NULL DEFAULT 0 COMMENT '0-待认证 1-认证中 2-成功 3-失败 4-已更换',
  `result_code`   varchar(20) NULL,
  `result_message` varchar(255) NULL,
  `result_data`   text         NULL,
  `verified_at`   datetime(3) NULL,
  `created_at`    datetime(3) NOT NULL,
  `updated_at`    datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_kyc_record_user_id` (`user_id`),
  KEY `idx_kyc_record_source` (`source`),
  UNIQUE KEY `idx_kyc_record_biz_no` (`biz_no`),
  KEY `idx_kyc_record_up_biz_id` (`up_biz_id`),
  KEY `idx_kyc_record_auth_order_id` (`auth_order_id`),
  KEY `idx_kyc_record_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='账户实名认证记录';

-- 余额流水
CREATE TABLE IF NOT EXISTS `balance_log` (
  `id`             bigint       NOT NULL AUTO_INCREMENT,
  `user_id`        bigint       NOT NULL,
  `order_id`       bigint       NULL COMMENT '关联订单ID（认证订单或支付订单）',
  `type`           tinyint      NOT NULL COMMENT '1-充值 2-消费 3-退款',
  `amount`         decimal(10,2) NOT NULL,
  `balance_before` decimal(10,2) NOT NULL,
  `balance_after`  decimal(10,2) NOT NULL,
  `bank_serial_no` varchar(100) NULL,
  `remark`         varchar(255) NULL,
  `created_at`     datetime(3)  NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_balance_log_user_id` (`user_id`),
  KEY `idx_balance_log_order_id` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='余额流水';

-- 支付订单（充值订单）
CREATE TABLE IF NOT EXISTS `payment_order` (
  `id`               bigint       NOT NULL AUTO_INCREMENT,
  `pay_order_no`     varchar(50)  NOT NULL,
  `user_id`          bigint       NOT NULL,
  `amount`           decimal(10,2) NOT NULL,
  `channel`          varchar(20)  NOT NULL COMMENT 'alipay-支付宝 wechat-微信',
  `channel_trade_no` varchar(100) NULL,
  `status`           tinyint      NOT NULL DEFAULT 0 COMMENT '0-待支付 1-已支付 2-已退款 3-已关闭',
  `refund_status`    tinyint      NOT NULL DEFAULT 0 COMMENT '0-未退款 1-部分退款 2-全额退款',
  `refund_amount`    decimal(10,2) NULL DEFAULT 0,
  `expire_time`      datetime(3)  NULL,
  `paid_at`          datetime(3)  NULL,
  `refunded_at`      datetime(3)  NULL,
  `created_at`       datetime(3)  NOT NULL,
  `updated_at`       datetime(3)  NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_payment_order_pay_order_no` (`pay_order_no`),
  KEY `idx_payment_order_user_id` (`user_id`),
  KEY `idx_payment_order_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='支付订单';

-- 系统配置
CREATE TABLE IF NOT EXISTS `system_config` (
  `id`           bigint       NOT NULL AUTO_INCREMENT,
  `config_key`   varchar(50)  NOT NULL,
  `config_value` text         NOT NULL COMMENT '配置值（JSON格式）',
  `description`  varchar(255) NULL,
  `created_at`   datetime(3)  NOT NULL,
  `updated_at`   datetime(3)  NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_system_config_config_key` (`config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统配置';

-- 用户登录记录
CREATE TABLE IF NOT EXISTS `user_login_log` (
  `id`          bigint       NOT NULL AUTO_INCREMENT,
  `user_id`     bigint       NULL,
  `account`     varchar(100) NOT NULL COMMENT '登录账号（手机号/用户名/邮箱）',
  `login_type`  varchar(20)  NOT NULL DEFAULT 'password' COMMENT 'password-密码 sms_code-短信验证码',
  `ip`          varchar(50)  NULL,
  `user_agent`  varchar(500) NULL,
  `status`      tinyint      NOT NULL DEFAULT 1 COMMENT '1-成功 0-失败',
  `fail_reason` varchar(255) NULL,
  `created_at`  datetime(3)  NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_login_log_user_id` (`user_id`),
  KEY `idx_user_login_log_account` (`account`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户登录记录';

-- 管理员登录记录
CREATE TABLE IF NOT EXISTS `admin_login_log` (
  `id`          bigint       NOT NULL AUTO_INCREMENT,
  `admin_id`    bigint       NULL,
  `username`    varchar(50)  NOT NULL,
  `ip`          varchar(50)  NULL,
  `user_agent`  varchar(500) NULL,
  `status`      tinyint      NOT NULL DEFAULT 1 COMMENT '1-成功 0-失败',
  `fail_reason` varchar(255) NULL,
  `created_at`  datetime(3)  NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_admin_login_log_admin_id` (`admin_id`),
  KEY `idx_admin_login_log_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员登录记录';

-- ============================================================
-- 实名认证产品库 starloft_kyc
-- ============================================================
USE `starloft_kyc`;

-- 认证订单
CREATE TABLE IF NOT EXISTS `auth_order` (
  `id`              bigint       NOT NULL AUTO_INCREMENT,
  `biz_no`          varchar(50)  NOT NULL COMMENT '全平台唯一业务流水号（平台调用时随机生成）',
  `user_id`         bigint       NOT NULL,
  `return_url`      varchar(500) NULL,
  `notify_url`      varchar(500) NULL,
  `biz_extra_data`  text         NULL,
  `up_token`        varchar(100) NULL,
  `up_biz_id`       varchar(50)  NULL,
  `up_request_id`   varchar(50)  NULL,
  `result_code`     varchar(20)  NULL,
  `result_message`  varchar(255) NULL,
  `result_data`     text         NULL,
  `status`          tinyint      NOT NULL DEFAULT 0 COMMENT '0-待认证 1-认证中 2-已完成 3-失败 4-已取消 5-超时（已退款）',
  `cost`            decimal(10,2) NOT NULL DEFAULT 0,
  `source`          tinyint      NOT NULL DEFAULT 2 COMMENT '1-账户实名 2-API调用',
  `pay_type`        tinyint      NOT NULL DEFAULT 0 COMMENT '0-免费 1-余额 2-资源包',
  `user_pack_id`    bigint       NULL,
  `is_refunded`     tinyint      NOT NULL DEFAULT 0,
  `notify_times`    bigint       NOT NULL DEFAULT 0,
  `notify_status`   tinyint      NOT NULL DEFAULT 0 COMMENT '0-待通知 1-通知成功 2-通知失败',
  `created_at`      datetime(3)  NOT NULL,
  `updated_at`      datetime(3)  NOT NULL,
  `finished_at`     datetime(3)  NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_auth_order_biz_no` (`biz_no`),
  KEY `idx_auth_order_user_id` (`user_id`),
  KEY `idx_auth_order_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='认证订单';

-- 资源包定义
CREATE TABLE IF NOT EXISTS `resource_pack` (
  `id`          bigint       NOT NULL AUTO_INCREMENT,
  `name`        varchar(100) NOT NULL,
  `total_count` bigint       NOT NULL,
  `price`       decimal(10,2) NOT NULL,
  `stock`       bigint       NOT NULL DEFAULT -1 COMMENT '-1 不限量，>=0 限量剩余库存',
  `status`      tinyint      NOT NULL DEFAULT 1 COMMENT '1-上架 0-下架',
  `description` varchar(255) NULL,
  `created_at`  datetime(3)  NOT NULL,
  `updated_at`  datetime(3)  NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_resource_pack_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='资源包';

-- 用户已购资源包
CREATE TABLE IF NOT EXISTS `user_resource_pack` (
  `id`             bigint      NOT NULL AUTO_INCREMENT,
  `user_id`        bigint      NOT NULL,
  `pack_id`        bigint      NOT NULL,
  `pack_name`      varchar(100) NOT NULL COMMENT '资源包名称（快照）',
  `total_count`    bigint      NOT NULL,
  `remaining_count` bigint     NOT NULL,
  `status`         tinyint     NOT NULL DEFAULT 1 COMMENT '1-有效 0-已耗尽/禁用',
  `created_at`     datetime(3) NOT NULL,
  `updated_at`     datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_user_resource_pack_user_id` (`user_id`),
  KEY `idx_user_resource_pack_user_status` (`user_id`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户资源包';

-- ============================================================
-- 初始化数据
-- ============================================================

-- 默认管理员
-- 用户名: admin  密码: kigga9aj（首次登录后请立即修改！） 密码哈希: bcrypt cost=10
-- 使用 REPLACE INTO 避免重复键冲突
REPLACE INTO `starloft_sys`.`admin_user` (`username`, `password_hash`, `nickname`, `status`, `created_at`, `updated_at`)
VALUES ('admin', '$2a$10$U8urafJV5ggvfeHunkalMuH4Wkdq7gWK1kI5X2q076ZWEJwGn5Dce', '管理员', 1, NOW(3), NOW(3));

-- 默认系统配置（第三方业务配置已迁移至数据库，由管理员在后台修改，无需重启；
-- 值为空时兜底使用 .env 环境变量/JWT/Redis 等基础设施配置仍在 .env 管理）
REPLACE INTO `starloft_sys`.`system_config` (`config_key`, `config_value`, `description`, `created_at`, `updated_at`) VALUES
('kyc_price', '1.00', 'KYC单价', NOW(3), NOW(3)),
('finauth_api_key', '', 'FinAuth API Key', NOW(3), NOW(3)),
('finauth_api_secret', '', 'FinAuth API Secret', NOW(3), NOW(3)),
('finauth_scene_id', '', 'FinAuth 场景ID', NOW(3), NOW(3)),
('finauth_base_url', '', 'FinAuth 接口地址', NOW(3), NOW(3)),
('tencent_secret_id', '', '腾讯云 SecretId（短信/验证码/邮件共用）', NOW(3), NOW(3)),
('tencent_secret_key', '', '腾讯云 SecretKey（短信/验证码/邮件共用）', NOW(3), NOW(3)),
('tencent_sms_sdk_app_id', '', '腾讯云短信 SDKAppID', NOW(3), NOW(3)),
('tencent_sms_sign_name', '', '腾讯云短信签名', NOW(3), NOW(3)),
('tencent_sms_template_id', '', '腾讯云短信模板ID', NOW(3), NOW(3)),
('tencent_captcha_app_id', '', '腾讯云验证码 AppID', NOW(3), NOW(3)),
('tencent_captcha_secret', '', '腾讯云验证码 SecretKey', NOW(3), NOW(3)),
('alipay_app_id', '', '支付宝应用AppID', NOW(3), NOW(3)),
('alipay_private_key', '', '支付宝应用私钥（PEM）', NOW(3), NOW(3)),
('alipay_public_key', '', '支付宝公钥（PEM）', NOW(3), NOW(3)),
('wechat_app_id', '', '微信支付绑定AppID', NOW(3), NOW(3)),
('wechat_mch_id', '', '微信支付商户号', NOW(3), NOW(3)),
('wechat_api_v3_key', '', '微信支付 APIv3 密钥', NOW(3), NOW(3)),
('wechat_mch_serial_no', '', '微信支付商户API证书序列号', NOW(3), NOW(3)),
('wechat_mch_private_key', '', '微信支付商户API私钥（PEM）', NOW(3), NOW(3)),
('wechat_platform_public_key', '', '微信支付公钥（PEM）', NOW(3), NOW(3)),
('ses_from', '', '腾讯云 SES 发件人地址', NOW(3), NOW(3)),
('ses_template_id', '', '腾讯云 SES 模板ID', NOW(3), NOW(3)),
('ses_region', '', '腾讯云 SES 地域（如 ap-guangzhou）', NOW(3), NOW(3));

-- 资源包种子数据（单价分档：0.7/0.6/0.5/0.4 元每次，库存 -1 不限量）
INSERT INTO `starloft_kyc`.`resource_pack` (`name`, `total_count`, `price`, `stock`, `status`, `description`, `created_at`, `updated_at`) VALUES
('认证100次套餐', 100, 70.00, -1, 1, '单价0.7元/次', NOW(3), NOW(3)),
('认证200次套餐', 200, 140.00, -1, 1, '单价0.7元/次', NOW(3), NOW(3)),
('认证500次套餐', 500, 350.00, -1, 1, '单价0.7元/次', NOW(3), NOW(3)),
('认证1000次套餐', 1000, 600.00, -1, 1, '单价0.6元/次', NOW(3), NOW(3)),
('认证2000次套餐', 2000, 1200.00, -1, 1, '单价0.6元/次', NOW(3), NOW(3)),
('认证5000次套餐', 5000, 3000.00, -1, 1, '单价0.6元/次', NOW(3), NOW(3)),
('认证1万次套餐', 10000, 5000.00, -1, 1, '单价0.5元/次', NOW(3), NOW(3)),
('认证2万次套餐', 20000, 10000.00, -1, 1, '单价0.5元/次', NOW(3), NOW(3)),
('认证5万次套餐', 50000, 25000.00, -1, 1, '单价0.5元/次', NOW(3), NOW(3)),
('认证10万次套餐', 100000, 40000.00, -1, 1, '单价0.4元/次', NOW(3), NOW(3)),
('认证20万次套餐', 200000, 80000.00, -1, 1, '单价0.4元/次', NOW(3), NOW(3)),
('认证50万次套餐', 500000, 200000.00, -1, 1, '单价0.4元/次', NOW(3), NOW(3));
