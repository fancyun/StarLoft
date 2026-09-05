package database

import (
	"database/sql"
	"fmt"
	"log"
	"starloftrpa/internal/config"
	"starloftrpa/internal/model"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var DB *sql.DB

func Init(cfg config.DatabaseConfig) error {
	// 构建DSN（连接系统库），添加TLS支持
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&tls=preferred",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// 设置连接池参数
	DB.SetMaxIdleConns(cfg.MaxIdleConns)
	DB.SetMaxOpenConns(cfg.MaxOpenConns)
	DB.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 使用 GORM AutoMigrate 自动创建/更新表结构（替代 init.sql 建表）
	if err := autoMigrate(); err != nil {
		return fmt.Errorf("failed to auto migrate database: %w", err)
	}

	// 清理旧结构：迁移存量 API 密钥到新 api 表，并删除 user 表 api_key/api_secret 列与废弃的 user_service 表
	if err := migrateLegacyRemoval(); err != nil {
		return fmt.Errorf("failed to migrate legacy schema: %w", err)
	}

	log.Println("Database connected successfully with TLS support")
	return nil
}

// autoMigrate 复用现有数据库连接执行 AutoMigrate，按模型自动创建表结构
func autoMigrate() error {
	gormDB, err := gorm.Open(mysql.New(mysql.Config{Conn: DB}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
		// AutoMigrate 的 information_schema 内省查询在云数据库上耗时较长，屏蔽慢查询日志避免启动时刷屏
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return err
	}
	// 系统库（starloft_sys）：位于默认连接库，AutoMigrate 内省可直接识别已存在的表并仅补列
	migrateModels := []interface{}{
		&model.User{},
		&model.AdminUser{},
		&model.ApiKey{},
		&model.KycPersonal{},
		&model.KycEnterprise{},
		&model.BalanceLog{},
		&model.PaymentOrder{},
		&model.UserLoginLog{},
		&model.AdminLoginLog{},
	}
	// 实名认证产品库（starloft_kyc）：跨库表 AutoMigrate 会按当前库匹配全限定表名，
	// 导致已存在表被误判为不存在而重复 CREATE（报 1050）。这里用原生 SQL 按实际库检测，
	// 表已存在（由 init.sql 创建）则跳过迁移，仅迁移缺失的表
	kycTables := []struct {
		model interface{}
		table string
	}{
		{&model.AuthOrder{}, "auth_order"},
		{&model.ResourcePack{}, "resource_pack"},
		{&model.UserResourcePack{}, "user_resource_pack"},
	}
	for _, t := range kycTables {
		exists, err := tableExists(model.KycDB, t.table)
		if err != nil {
			return err
		}
		if !exists {
			migrateModels = append(migrateModels, t.model)
		}
	}
	return gormDB.AutoMigrate(migrateModels...)
}

// tableExists 判断指定库下表是否存在（跨库存在性检查基于原生查询，避免 AutoMigrate 按当前库误判）
func tableExists(schemaName, tableName string) (bool, error) {
	var n int
	err := DB.QueryRow(
		`SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = ?`,
		schemaName, tableName,
	).Scan(&n)
	return n > 0, err
}

// columnExists 判断指定库表中某列是否存在
func columnExists(schemaName, tableName, columnName string) (bool, error) {
	var n int
	err := DB.QueryRow(
		`SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = ? AND table_name = ? AND column_name = ?`,
		schemaName, tableName, columnName,
	).Scan(&n)
	return n > 0, err
}

// migrateLegacyRemoval 移除已废弃的旧结构（仅在对应表/列仍存在时执行，具备幂等性）：
//   - 将 user 表存量 api_key/api_secret 迁入新的 api 表（permission 置为 all）
//   - 删除 user 表 api_key/api_secret 列
//   - 删除废弃的 user_service 表
func migrateLegacyRemoval() error {
	// 数据迁移：存量 API 密钥迁入 api 表
	if ex, err := columnExists(model.SysDB, "user", "api_key"); err != nil {
		return err
	} else if ex {
		if _, err := DB.Exec(`INSERT INTO ` + model.SysDB + `.api (user_id, api_key, api_secret, permission, created_at, updated_at)
			SELECT id, api_key, api_secret, 'all', NOW(), NOW()
			FROM ` + model.SysDB + `.user WHERE api_key != ''
			ON DUPLICATE KEY UPDATE api_secret = VALUES(api_secret)`); err != nil {
			return err
		}
	}

	// 删除 user 表 api_key / api_secret 列
	for _, col := range []string{"api_key", "api_secret"} {
		if ex, err := columnExists(model.SysDB, "user", col); err != nil {
			return err
		} else if ex {
			if _, err := DB.Exec(`ALTER TABLE ` + model.SysDB + `.user DROP COLUMN ` + col); err != nil {
				return err
			}
		}
	}

	// 删除废弃的 user_service 表
	if ex, err := tableExists(model.SysDB, "user_service"); err != nil {
		return err
	} else if ex {
		if _, err := DB.Exec(`DROP TABLE ` + model.SysDB + `.user_service`); err != nil {
			return err
		}
	}

	return nil
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
