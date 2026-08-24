-- KYC 实名认证分销平台 数据库初始化脚本

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- 删除已存在的数据库（如果存在）
DROP DATABASE IF EXISTS `kyc_platform`;

-- 创建数据库
CREATE DATABASE `kyc_platform` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `kyc_platform`;

-- ----------------------------
-- 1. 平台用户表
-- ----------------------------
DROP TABLE IF EXISTS `platform_user`;
CREATE TABLE `platform_user` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `phone` varchar(20) NOT NULL COMMENT '手机号',
  `password_hash` varchar(255) NOT NULL COMMENT '密码哈希',
  `balance` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '账户余额',
  `api_key` varchar(64) NOT NULL COMMENT 'API Key',
  `api_secret` varchar(64) NOT NULL COMMENT 'API Secret',
  `is_kyc_verified` tinyint(1) NOT NULL DEFAULT '0' COMMENT '账户实名状态：0-未实名 1-已实名',
  `kyc_name` varchar(100) DEFAULT NULL COMMENT '实名认证姓名（加密存储）',
  `kyc_id_card` varchar(100) DEFAULT NULL COMMENT '实名认证身份证号（加密存储）',
  `kyc_price` decimal(10,2) NOT NULL DEFAULT '1.00' COMMENT 'KYC认证单价（元），默认为系统价格，可单独调整',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态：1-正常 0-禁用',
  `last_login_at` datetime DEFAULT NULL COMMENT '最后登录时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_phone` (`phone`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台用户表';

-- ----------------------------
-- 2. 管理员用户表
-- ----------------------------
DROP TABLE IF EXISTS `admin_user`;
CREATE TABLE `admin_user` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `username` varchar(50) NOT NULL COMMENT '管理员用户名',
  `password_hash` varchar(255) NOT NULL COMMENT '密码哈希',
  `nickname` varchar(50) DEFAULT NULL COMMENT '管理员昵称',
  `status` tinyint(1) NOT NULL DEFAULT '1' COMMENT '状态：1-正常 0-禁用',
  `last_login_at` datetime DEFAULT NULL COMMENT '最后登录时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员用户表';

-- 插入默认管理员账号
-- 用户名: admin
-- 密码: Admin@2026KYC (请在首次登录后立即修改！)
-- 密码哈希算法: bcrypt (cost=10)

-- 使用 REPLACE INTO 避免重复键冲突
REPLACE INTO `admin_user` (`username`, `password_hash`, `nickname`, `status`) 
VALUES ('admin', '$2a$10$J98RpDuJVy4nNUN7LLjNgu0mEYQmO.z5yZy29LwVyFaWwg8JjprRW', '管理员', 1);

-- ----------------------------
-- 3. 认证订单表
-- ----------------------------
DROP TABLE IF EXISTS `auth_order`;
CREATE TABLE `auth_order` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `platform_biz_no` varchar(50) NOT NULL COMMENT '平台业务流水号',
  `biz_no` varchar(50) DEFAULT NULL COMMENT '用户业务流水号',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `return_url` varchar(500) DEFAULT NULL COMMENT '认证完成后跳转的URL',
  `notify_url` varchar(500) DEFAULT NULL COMMENT '异步通知回调URL',
  `biz_extra_data` text COMMENT '额外业务数据',
  `up_token` varchar(100) DEFAULT NULL COMMENT '上游返回的token',
  `up_biz_id` varchar(50) DEFAULT NULL COMMENT '上游返回的biz_id',
  `up_request_id` varchar(50) DEFAULT NULL COMMENT '上游返回的request_id',
  `result_code` varchar(20) DEFAULT NULL COMMENT '认证结果码',
  `result_message` varchar(255) DEFAULT NULL COMMENT '认证结果消息',
  `result_data` text COMMENT '认证结果完整数据（JSON）',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态：0-待认证 1-认证中 2-已完成 3-失败 4-已取消 5-超时（已退款）',
  `cost` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '本次认证消耗金额',
  `is_refunded` tinyint(1) NOT NULL DEFAULT '0' COMMENT '超时是否已退款：0-否 1-是',
  `notify_times` int(11) NOT NULL DEFAULT '0' COMMENT '回调用户次数',
  `notify_status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '回调用户状态：0-待通知 1-通知成功 2-通知失败',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `finished_at` datetime DEFAULT NULL COMMENT '完成时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_platform_biz_no` (`platform_biz_no`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_biz_no` (`biz_no`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='认证订单表';

-- ----------------------------
-- 4. 余额流水表
-- ----------------------------
DROP TABLE IF EXISTS `balance_log`;
CREATE TABLE `balance_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `order_id` bigint(20) DEFAULT NULL COMMENT '关联订单ID（认证订单或支付订单）',
  `type` tinyint(1) NOT NULL COMMENT '类型：1-充值 2-消费 3-退款',
  `amount` decimal(10,2) NOT NULL COMMENT '变动金额',
  `balance_before` decimal(10,2) NOT NULL COMMENT '变动前余额',
  `balance_after` decimal(10,2) NOT NULL COMMENT '变动后余额',
  `remark` varchar(255) DEFAULT NULL COMMENT '备注',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='余额流水表';

-- ----------------------------
-- 5. 回调记录表
-- ----------------------------
DROP TABLE IF EXISTS `callback_log`;
CREATE TABLE `callback_log` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `order_id` bigint(20) NOT NULL COMMENT '关联订单ID',
  `direction` tinyint(1) NOT NULL COMMENT '方向：1-上游回调平台 2-平台回调用户',
  `request_body` text COMMENT '回调请求体',
  `response_body` text COMMENT '回调响应体',
  `http_status` int(11) DEFAULT NULL COMMENT 'HTTP状态码',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态：0-待处理 1-成功 2-失败',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='回调记录表';

-- ----------------------------
-- 6. 支付订单表
-- ----------------------------
DROP TABLE IF EXISTS `payment_order`;
CREATE TABLE `payment_order` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `pay_order_no` varchar(50) NOT NULL COMMENT '支付流水号',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `amount` decimal(10,2) NOT NULL COMMENT '充值金额（元）',
  `channel` varchar(20) NOT NULL COMMENT '支付渠道：alipay-支付宝 wechat-微信 unionpay-云闪付',
  `channel_trade_no` varchar(100) DEFAULT NULL COMMENT '银联商务交易号（seqId）',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '状态：0-待支付 1-已支付 2-已退款 3-已关闭',
  `expire_time` datetime NOT NULL COMMENT '支付过期时间',
  `paid_at` datetime DEFAULT NULL COMMENT '支付完成时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `refund_status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '退款状态：0-未退款 1-退款审核中 2-已退款 3-退款驳回',
  `refund_amount` decimal(10,2) DEFAULT NULL COMMENT '实际退款金额',
  `refunded_at` datetime DEFAULT NULL COMMENT '退款完成时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_pay_order_no` (`pay_order_no`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付订单表';

-- ----------------------------
-- 7. 系统配置表
-- ----------------------------
DROP TABLE IF EXISTS `system_config`;
CREATE TABLE `system_config` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `config_key` varchar(50) NOT NULL COMMENT '配置键',
  `config_value` text NOT NULL COMMENT '配置值（JSON格式）',
  `description` varchar(255) DEFAULT NULL COMMENT '配置说明',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_config_key` (`config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';

-- 插入默认配置（密钥类配置统一在 .env 环境变量中管理）

-- 使用 REPLACE INTO 避免重复键冲突
REPLACE INTO `system_config` (`config_key`, `config_value`, `description`) VALUES
('kyc_price', '1.00', 'KYC单价');

-- ----------------------------
-- 8. 账户实名认证记录表
-- ----------------------------
DROP TABLE IF EXISTS `kyc_record`;
CREATE TABLE `kyc_record` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `user_id` bigint(20) NOT NULL COMMENT '用户ID',
  `auth_order_id` bigint(20) DEFAULT NULL COMMENT '关联认证订单ID（通过下单完成实名时关联）',
  `name` varchar(50) NOT NULL COMMENT '实名姓名',
  `id_card` varchar(18) NOT NULL COMMENT '身份证号',
  `status` tinyint(1) NOT NULL DEFAULT '0' COMMENT '认证状态：0-待认证 1-认证中 2-认证成功 3-认证失败',
  `result_code` varchar(20) DEFAULT NULL COMMENT '上游认证结果码',
  `result_message` varchar(255) DEFAULT NULL COMMENT '认证结果消息',
  `result_data` text COMMENT '认证结果完整数据（JSON）',
  `verified_at` datetime DEFAULT NULL COMMENT '认证通过时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_auth_order_id` (`auth_order_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='账户实名认证记录表';

-- ----------------------------
-- 9. 用户API调用记录表
-- ----------------------------
DROP TABLE IF EXISTS `kyc_api_record`;
CREATE TABLE `kyc_api_record` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  `user_id` bigint(20) NOT NULL COMMENT '调用用户ID',
  `auth_order_id` bigint(20) DEFAULT NULL COMMENT '关联认证订单ID',
  `api_type` varchar(50) NOT NULL COMMENT 'API类型：get_token / get_result / create_order',
  `request_data` text COMMENT '请求数据（脱敏后）',
  `response_data` text COMMENT '响应数据',
  `http_status` int(11) NOT NULL DEFAULT '0' COMMENT 'HTTP状态码',
  `cost` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '本次调用消耗金额',
  `duration_ms` int(11) NOT NULL DEFAULT '0' COMMENT '接口耗时（毫秒）',
  `error_message` text COMMENT '错误信息',
  `ip_address` varchar(50) DEFAULT NULL COMMENT '请求IP地址',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_auth_order_id` (`auth_order_id`),
  KEY `idx_api_type` (`api_type`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户API调用记录表';

SET FOREIGN_KEY_CHECKS = 1;
