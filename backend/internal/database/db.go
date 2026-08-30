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
	"gorm.io/gorm/schema"
)

var DB *sql.DB

func Init(cfg config.DatabaseConfig) error {
	// 构建DSN，添加TLS支持
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
	return gormDB.AutoMigrate(
		&model.PlatformUser{},
		&model.AdminUser{},
		&model.AuthOrder{},
		&model.BalanceLog{},
		&model.PaymentOrder{},
		&model.SystemConfig{},
		&model.KycRecord{},
		&model.ResourcePack{},
		&model.UserResourcePack{},
		&model.InternalAccount{},
	)
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
