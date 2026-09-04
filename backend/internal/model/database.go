package model

// 分库库名（逻辑分库：同一 MySQL 实例下的多个 database，事务可跨库，连接默认库为系统库）
// 库名写死为固定常量，需与 database/init.sql 的建库名一致，不允许通过配置修改
var (
	SysDB = "starloft_sys" // 系统库：用户/管理员/实名记录/余额流水/充值订单/系统配置/登录日志
	KycDB = "starloft_kyc" // 实名认证产品库：认证订单/资源包
)
