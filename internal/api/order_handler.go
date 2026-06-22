package api

import (
	"strconv"

	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

// GetProducts 获取商品列表
func GetProducts(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if ProductSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	products, err := ProductSvc.GetAllProducts(true)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success":  true,
		"products": buildProductResponses(products),
	})
}

// GetProduct 获取单个商品
func GetProduct(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if ProductSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的商品ID"})
		return
	}

	product, err := ProductSvc.GetProductByID(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "商品不存在"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"product": buildProductResponse(product),
	})
}

// CreateOrder 创建订单
func CreateOrder(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if OrderSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")
	username := c.GetString("username")

	var req struct {
		ProductID uint `json:"product_id" binding:"required"`
		Quantity  int  `json:"quantity"` // 购买数量，默认为1
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 数量默认为1，最小为1
	quantity := req.Quantity
	if quantity < 1 {
		quantity = 1
	}

	order, err := OrderSvc.CreateOrderWithQuantity(userID, username, req.ProductID, quantity, c.ClientIP())
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success":  true,
		"message":  "订单创建成功",
		"order_no": order.OrderNo,
		"order":    order,
	})
}

// OrderDetail 订单详情
func OrderDetail(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if OrderSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")
	orderNo := c.Param("order_no")

	order, err := OrderSvc.GetOrderByOrderNo(orderNo)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "订单不存在"})
		return
	}

	if order.UserID != userID {
		c.JSON(403, gin.H{"success": false, "error": "无权查看此订单"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"order":   order,
	})
}

// CancelOrder 取消订单
func CancelOrder(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if OrderSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")

	var req struct {
		OrderNo string `json:"order_no" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := OrderSvc.CancelOrder(req.OrderNo, userID); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "订单已取消"})
}
