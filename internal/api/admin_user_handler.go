package api

import (
	"strconv"

	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== 用户管理相关 API ====================

type adminUserResponse struct {
	ID               uint    `json:"id"`
	Username         string  `json:"username"`
	Email            string  `json:"email"`
	Phone            string  `json:"phone"`
	EmailVerified    bool    `json:"email_verified"`
	Enable2FA        bool    `json:"enable_2fa"`
	PayPasswordSet   bool    `json:"pay_password_set"`
	Status           int     `json:"status"`
	LastLoginAt      any     `json:"last_login_at"`
	LastLoginIP      string  `json:"last_login_ip"`
	CreatedAt        any     `json:"created_at"`
	OrderCount       int64   `json:"order_count"`
	PaidOrderCount   int64   `json:"paid_order_count"`
	AvailableBalance float64 `json:"available_balance"`
}

func buildAdminUserResponses(users []model.User) []adminUserResponse {
	items := make([]adminUserResponse, 0, len(users))
	for _, user := range users {
		var orderCount int64
		var paidOrderCount int64
		var balance model.UserBalance
		if model.DBConnected && model.DB != nil {
			model.DB.Model(&model.Order{}).Where("user_id = ?", user.ID).Count(&orderCount)
			model.DB.Model(&model.Order{}).Where("user_id = ? AND status >= ?", user.ID, model.OrderStatusPaid).Count(&paidOrderCount)
			_ = model.DB.Where("user_id = ?", user.ID).First(&balance).Error
		}
		items = append(items, adminUserResponse{
			ID:               user.ID,
			Username:         user.Username,
			Email:            user.Email,
			Phone:            user.Phone,
			EmailVerified:    user.EmailVerified,
			Enable2FA:        user.Enable2FA,
			PayPasswordSet:   user.PayPasswordSet,
			Status:           user.Status,
			LastLoginAt:      user.LastLoginAt,
			LastLoginIP:      user.LastLoginIP,
			CreatedAt:        user.CreatedAt,
			OrderCount:       orderCount,
			PaidOrderCount:   paidOrderCount,
			AvailableBalance: balance.Balance,
		})
	}
	return items
}

// AdminGetUsers 获取用户列表
func AdminGetUsers(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if UserSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	users, total, err := UserSvc.GetAllUsers(page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success":     true,
		"users":       buildAdminUserResponses(users),
		"total":       total,
		"page":        page,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// AdminUpdateUserStatus 更新用户状态
func AdminUpdateUserStatus(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if UserSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var req struct {
		Status int `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := UserSvc.UpdateUserStatus(uint(id), req.Status); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "状态已更新"})
}
