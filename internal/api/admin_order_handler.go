package api

import (
	"strconv"

	"user-frontend/internal/model"
	pluginapi "user-frontend/internal/plugin"

	"github.com/gin-gonic/gin"
)

// ==================== 订单管理相关 API ====================

// AdminGetOrders 获取订单列表
func AdminGetOrders(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if OrderSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	orders, total, err := OrderSvc.GetAllOrders(page, pageSize, nil)
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

// AdminGetOrder 获取订单详情
func AdminGetOrder(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if OrderSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	order, err := OrderSvc.GetOrderByID(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "订单不存在"})
		return
	}

	response := gin.H{"success": true, "order": order}
	if OrderKernelSvc != nil {
		if snapshot, err := OrderKernelSvc.SnapshotByOrderID(uint(id)); err == nil {
			response["trace"] = snapshot
		}
	}
	c.JSON(200, response)
}

// AdminCreatePluginMaterialOrder 由宿主接收商品插件订单素材并创建正式订单。
func AdminCreatePluginMaterialOrder(c *gin.Context) {
	if OrderKernelSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "订单内核未初始化"})
		return
	}
	var req struct {
		UserID   uint                           `json:"user_id" binding:"required"`
		Username string                         `json:"username"`
		Material pluginapi.ProductOrderMaterial `json:"material" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}
	order, err := OrderKernelSvc.CreatePluginMaterialOrder(req.UserID, req.Username, GetClientIP(c), req.Material)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true, "order": order})
}
