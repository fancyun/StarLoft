package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"starloftrpa/internal/repository"
)

// serviceRequirement 服务开通资格要求：serviceCode -> 需满足的实名等级（1-个人 2-企业）
// 后续新增服务只需在此登记资格要求，无需改动中间件
var serviceRequirement = map[string]int{
	"kyc": 2, // KYC API 服务要求企业实名
}

// serviceCatalog 服务目录（前端展示用）：code 名称、资格说明
var serviceCatalog = []gin.H{
	{"service_code": "kyc", "name": "KYC 实名认证 API", "requirement": "企业实名"},
}

// ServiceHandler 服务开通处理
type ServiceHandler struct {
	userRepo        *repository.UserRepository
	userServiceRepo *repository.UserServiceRepository
}

func NewServiceHandler(userRepo *repository.UserRepository, userServiceRepo *repository.UserServiceRepository) *ServiceHandler {
	return &ServiceHandler{
		userRepo:        userRepo,
		userServiceRepo: userServiceRepo,
	}
}

// OpenService 开通服务（POST /service/open，入参 service_code）
// 幂等：已开通直接返回成功；未开通则校验资格后落库
func (h *ServiceHandler) OpenService(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req struct {
		ServiceCode string `json:"service_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "invalid request parameters"})
		return
	}

	need, ok := serviceRequirement[req.ServiceCode]
	if !ok {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的服务标识"})
		return
	}

	// 幂等：已开通直接返回成功
	opened, err := h.userServiceRepo.IsOpen(userID, req.ServiceCode)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询服务开通状态失败"})
		return
	}
	if opened {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "服务已开通", "data": gin.H{"opened": true}})
		return
	}

	// 校验实名资格
	user, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "user not found"})
		return
	}
	if user.IsKYCVerified < need {
		if need == 2 {
			c.JSON(http.StatusOK, gin.H{"code": 403, "message": "该服务需企业实名才能开通，请先完成企业实名认证"})
		} else {
			c.JSON(http.StatusOK, gin.H{"code": 403, "message": "该服务需完成实名认证后才能开通"})
		}
		return
	}

	if _, err := h.userServiceRepo.Create(userID, req.ServiceCode); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "开通服务失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "服务开通成功", "data": gin.H{"opened": true}})
}

// ServiceSummary 服务开通概览（前端展示每个服务的开通状态与资格说明）
func (h *ServiceHandler) ServiceSummary(c *gin.Context) {
	userID := c.GetInt64("user_id")

	user, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 404, "message": "user not found"})
		return
	}

	list := make([]gin.H, 0, len(serviceCatalog))
	for _, svc := range serviceCatalog {
		code := svc["service_code"].(string)
		opened, _ := h.userServiceRepo.IsOpen(userID, code)
		item := gin.H{
			"service_code": code,
			"name":         svc["name"],
			"requirement":  svc["requirement"],
			"opened":       opened,
		}
		// 附加是否满足资格（用于前端「可开通」判断）
		if need, ok := serviceRequirement[code]; ok {
			item["eligible"] = user.IsKYCVerified >= need
		}
		list = append(list, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"is_kyc_verified": user.IsKYCVerified,
			"services":        list,
		},
	})
}
