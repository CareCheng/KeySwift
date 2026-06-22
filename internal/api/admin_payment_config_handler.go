package api

import "github.com/gin-gonic/gin"

// ==================== 支付配置相关 API ====================

// AdminGetPaymentConfig 获取支付配置。
func AdminGetPaymentConfig(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"config": gin.H{
			"balance": gin.H{
				"enabled": true,
				"builtin": true,
				"name":    "余额支付",
			},
		},
	})
}
