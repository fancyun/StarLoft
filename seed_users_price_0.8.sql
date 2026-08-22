-- ============================================================
-- StarLoft KYC 数据初始化脚本
--   1) 修改系统实名单价为 0.8 元
--   2) 插入五个用户（实名单价均为 0.8 元）
--
-- 密码采用 bcrypt(cost=10) 哈希，与后端注册逻辑完全一致，
-- 可直接用于登录。
-- 注意：bcrypt 哈希中包含 "$" 字符，务必保持单引号字符串原样。
-- ============================================================

-- 1. 修改系统实名单价（config_value 存字符串）
UPDATE `system_config`
SET `config_value` = '0.80',
    `updated_at`   = NOW()
WHERE `config_key` = 'kyc_price';

-- 2. 插入五个用户
-- 字段说明：
--   balance=0.00（初始余额）
--   api_key / api_secret 用 UUID 生成 64 位唯一串（如需更强的随机密钥，
--     可登录后在后台“重新生成密钥”）
--   is_kyc_verified=0（未实名），status=1（正常）
-- 幂等：若手机号已存在，则更新密码、单价、状态，不动已有 api_key/api_secret
INSERT INTO `platform_user`
  (`phone`, `password_hash`, `balance`, `api_key`, `api_secret`, `is_kyc_verified`, `kyc_price`, `status`, `created_at`, `updated_at`)
VALUES
  ('18731515851', '$2a$10$toCHI4bzbMBQlXhbl62rxeECvC2EuoPKsubCZccFZzHuGUO2n4xXy', 0.00, LOWER(CONCAT(REPLACE(UUID(),'-',''),REPLACE(UUID(),'-',''))), LOWER(CONCAT(REPLACE(UUID(),'-',''),REPLACE(UUID(),'-',''))), 0, 0.80, 1, DATE_SUB(NOW(), INTERVAL 4 SECOND), NOW()),
  ('17608923285', '$2a$10$iCTF/bfkn.8qZZ9yD.kZXO71bm7x9MAcR6ERSBPDVVRjU9sTxpM7m', 0.00, LOWER(CONCAT(REPLACE(UUID(),'-',''),REPLACE(UUID(),'-',''))), LOWER(CONCAT(REPLACE(UUID(),'-',''),REPLACE(UUID(),'-',''))), 0, 0.80, 1, DATE_SUB(NOW(), INTERVAL 3 SECOND), NOW()),
  ('15190522390', '$2a$10$2i4kt2Rx2.WbPU5N4DL1We02XeVkYcx9Hy5nk3jWNu0cOql1.C5DO', 0.00, LOWER(CONCAT(REPLACE(UUID(),'-',''),REPLACE(UUID(),'-',''))), LOWER(CONCAT(REPLACE(UUID(),'-',''),REPLACE(UUID(),'-',''))), 0, 0.80, 1, DATE_SUB(NOW(), INTERVAL 2 SECOND), NOW()),
  ('19894632113', '$2a$10$KgEN1VtY30TUW6g967j0uePlBTnTk7kGkiQK34S9qqKvMdG6vYEE6', 0.00, LOWER(CONCAT(REPLACE(UUID(),'-',''),REPLACE(UUID(),'-',''))), LOWER(CONCAT(REPLACE(UUID(),'-',''),REPLACE(UUID(),'-',''))), 0, 0.80, 1, DATE_SUB(NOW(), INTERVAL 1 SECOND), NOW()),
  ('15093559700', '$2a$10$.7uxWn54h9q33YKdirvMj.1LyTkvcfj49GwqBJS8.DrEVIdIR8F8G', 0.00, LOWER(CONCAT(REPLACE(UUID(),'-',''),REPLACE(UUID(),'-',''))), LOWER(CONCAT(REPLACE(UUID(),'-',''),REPLACE(UUID(),'-',''))), 0, 0.80, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE
  `password_hash` = VALUES(`password_hash`),
  `kyc_price`     = VALUES(`kyc_price`),
  `status`        = VALUES(`status`),
  `updated_at`    = NOW();