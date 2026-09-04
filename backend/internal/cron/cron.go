package cron

import (
	"log"

	"github.com/robfig/cron/v3"

	"starloftrpa/internal/service"
)

// CronManager 定时任务管理器
type CronManager struct {
	cron           *cron.Cron
	authService    *service.AuthService
	balanceService *service.BalanceService
}

// NewCronManager 创建定时任务管理器
func NewCronManager(authService *service.AuthService, balanceService *service.BalanceService) *CronManager {
	return &CronManager{
		cron:           cron.New(),
		authService:    authService,
		balanceService: balanceService,
	}
}

// Start 启动定时任务
func (m *CronManager) Start() error {
	// 每5分钟同步一次处理中订单，根据上游返回结果处理退款
	_, err := m.cron.AddFunc("*/5 * * * *", m.syncPendingOrders)
	if err != nil {
		return err
	}

	// 每分钟释放一次超时未支付的资源包订单（释放占用的库存并退还余额支付部分）
	_, err = m.cron.AddFunc("*/1 * * * *", m.releaseExpiredResourcePacks)
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

// releaseExpiredResourcePacks 释放超时未支付的资源包订单（库存 + 余额支付部分）
func (m *CronManager) releaseExpiredResourcePacks() {
	log.Println("开始释放超时资源包订单...")

	err := m.balanceService.ReleaseExpiredResourcePackOrders()
	if err != nil {
		log.Printf("释放超时资源包订单失败: %v", err)
		return
	}

	log.Println("超时资源包订单释放完成")
}
