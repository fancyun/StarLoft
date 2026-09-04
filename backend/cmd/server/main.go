package main

import (
	"log"
	"starloftrpa/internal/config"
	"starloftrpa/internal/cron"
	"starloftrpa/internal/database"
	"starloftrpa/internal/redis"
	"starloftrpa/internal/router"
	"starloftrpa/internal/utils"
)

func main() {
	// 加载配置（从环境变量）
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化文件日志（写入 /app/logs，可按 LOG_DIR 覆盖）
	if err := utils.InitLoggers(cfg.Log.Dir); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	// 初始化数据库
	if err := database.Init(cfg.Database); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer database.Close()

	// 初始化 Redis
	if err := redis.Init(&cfg.Redis); err != nil {
		log.Fatalf("初始化 Redis 失败: %v", err)
	}
	defer redis.Close()
	log.Println("Redis 连接成功")

	// 初始化路由
	r, authService, balanceService := router.Setup(cfg)

	// 启动定时任务
	cronManager := cron.NewCronManager(authService, balanceService)
	if err := cronManager.Start(); err != nil {
		log.Fatalf("启动定时任务失败: %v", err)
	}
	defer cronManager.Stop()

	// 启动服务器
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("服务器启动在 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
