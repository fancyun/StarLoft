package cron

import (
	"log"

	"github.com/robfig/cron/v3"

	"starloftrpa/internal/service"
)

// CronManager 定时任务管理器
type CronManager struct {
	cron        *cron.Cron
	authService *service.AuthService
}

// NewCronManager 创建定时任务管理器
func NewCronManager(authService *service.AuthService) *CronManager {
	return &CronManager{
		cron:        cron.New(),
		authService: authService,
	}
}

// Start 启动定时任务
func (m *CronManager) Start() error {
	// 每5分钟同步一次处理中订单，根据上游返回结果处理退款
	_, err := m.cron.AddFunc("*/5 * * * *", m.syncPendingOrders)
	if err != nil {
		return err
	}

	m.cron.Start()
	log.Println("定时任务已启动")

	return nil
}

// Stop 停止定时任务
func (m *CronManager) Stop() {
	m.cron.Stop()
	log.Println("定时任务已停止")
}

// syncPendingOrders 同步处理中订单并根据上游结果处理退款
func (m *CronManager) syncPendingOrders() {
	log.Println("开始同步处理中订单...")

	err := m.authService.SyncPendingOrders()
	if err != nil {
		log.Printf("同步处理中订单失败: %v", err)
		return
	}

	log.Println("处理中订单同步完成")
}
