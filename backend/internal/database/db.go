package database

import (
	"database/sql"
	"fmt"
	"log"
	"starloftrpa/internal/config"
	"time"

	_ "github.com/go-sql-driver/mysql"
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

	log.Println("Database connected successfully with TLS support")
	return nil
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
