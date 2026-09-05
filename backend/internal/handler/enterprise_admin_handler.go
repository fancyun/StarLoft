package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"starloftrpa/internal/service"
)

// EnterpriseAdminHandler 企业实名管理（管理后台）
type EnterpriseAdminHandler struct {
	authService *service.AuthService
}

func NewEnterpriseAdminHandler(authService *service.AuthService) *EnterpriseAdminHandler {
	return &EnterpriseAdminHandler{authService: authService}
}

// VerifyEnterprise 后台人工企业实名（=公户验证）：录入企业名称与统一社会信用代码即开通
func (h *EnterpriseAdminHandler) VerifyEnterprise(c *gin.Context) {
	adminID := c.GetInt64("user_id") // 管理员 JWT 与用户共用 user_id 上下文键

	var req struct {
		UserID      int64  `json:"user_id" binding:"required"`
		CompanyName string `json:"company_name" binding:"required"`
		CreditCode  string `json:"credit_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "请填写用户、企业名称与统一社会信用代码"})
		return
	}

	if err := h.authService.AdminCreateEnterpriseManual(req.UserID, req.CompanyName, req.CreditCode, adminID); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "企业实名开通失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "企业实名已开通"})
}

// ListEnterpriseRecords 企业实名记录列表（分页）
func (h *EnterpriseAdminHandler) ListEnterpriseRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	list, total, err := h.authService.ListEnterpriseRecords(page, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "查询企业实名记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":      list,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}
