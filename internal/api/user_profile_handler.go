// Package api 提供 HTTP API 处理器
// user_profile_handler.go - 用户信息管理处理器
package api

import (
	"time"

	"github.com/gin-gonic/gin"
)

// UserInfo 获取用户信息
func UserInfo(c *gin.Context) {
	if UserSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")
	user, err := UserSvc.GetUserByID(userID)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "用户不存在"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"user": gin.H{
			"id":             user.ID,
			"username":       user.Username,
			"email":          user.Email,
			"email_verified": user.EmailVerified,
			"phone":          user.Phone,
			"enable_2fa":     user.Enable2FA,
			"created_at":     user.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}

// UpdateUserInfo 更新用户信息
func UpdateUserInfo(c *gin.Context) {
	if UserSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")

	var req struct {
		Phone string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	user, err := UserSvc.GetUserByID(userID)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "用户不存在"})
		return
	}

	user.Phone = req.Phone
	if err := UserSvc.UpdateUser(user); err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "信息已更新"})
}

// UpdatePassword 修改密码
func UpdatePassword(c *gin.Context) {
	if UserSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if len(req.NewPassword) < 6 {
		c.JSON(400, gin.H{"success": false, "error": "新密码长度至少6位"})
		return
	}

	if err := UserSvc.UpdatePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "密码修改成功"})
}

// UserOrders 获取用户订单
func UserOrders(c *gin.Context) {
	if OrderSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")
	page := 1
	pageSize := 10

	orders, total, err := OrderSvc.GetUserOrders(userID, page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"orders":  orders,
		"total":   total,
		"page":    page,
	})
}

// GetUserKamis 获取用户已购卡密
func GetUserKamis(c *gin.Context) {
	if OrderSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")
	orders, _, err := OrderSvc.GetUserOrders(userID, 1, 1000)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	type userKamiItem struct {
		OrderNo      string     `json:"order_no"`
		ProductName  string     `json:"product_name"`
		KamiCode     string     `json:"kami_code"`
		Quantity     int        `json:"quantity"`
		Status       int        `json:"status"`
		PaymentTime  *time.Time `json:"payment_time,omitempty"`
		CreatedAt    time.Time  `json:"created_at"`
		UpdatedAt    time.Time  `json:"updated_at"`
	}

	kamis := make([]userKamiItem, 0, len(orders))
	for _, order := range orders {
		if order.KamiCode == "" || order.Status != 2 {
			continue
		}
		kamis = append(kamis, userKamiItem{
			OrderNo:     order.OrderNo,
			ProductName: order.ProductName,
			KamiCode:    order.KamiCode,
			Quantity:    order.Quantity,
			Status:      order.Status,
			PaymentTime: order.PaymentTime,
			CreatedAt:   order.CreatedAt,
			UpdatedAt:   order.UpdatedAt,
		})
	}

	c.JSON(200, gin.H{
		"success": true,
		"kamis":   kamis,
	})
}
