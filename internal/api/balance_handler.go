package api

import (
	"strconv"

	"user-frontend/internal/service"

	"github.com/gin-gonic/gin"
)

// ==================== 用户端余额 API ====================

// GetMyBalance 获取我的余额
func GetMyBalance(c *gin.Context) {
	if BalanceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, _ := c.Get("user_id")
	balance, err := BalanceSvc.GetUserBalance(userID.(uint))
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取余额失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": balance})
}

// GetMyBalanceLogs 获取我的余额变动记录
func GetMyBalanceLogs(c *gin.Context) {
	if BalanceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID, _ := c.Get("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	logType := c.Query("type")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := BalanceSvc.GetBalanceLogs(userID.(uint), page, pageSize, logType)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取记录失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    logs,
		"total":   total,
		"page":    page,
		"pages":   (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// ==================== 管理员余额 API ====================

// AdminGetBalances 管理员获取用户余额列表
func AdminGetBalances(c *gin.Context) {
	if BalanceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	balances, total, err := BalanceSvc.AdminGetAllBalances(page, pageSize, keyword)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取余额列表失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    balances,
		"total":   total,
		"page":    page,
		"pages":   (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// AdminGetBalanceLogs 管理员获取余额变动记录
func AdminGetBalanceLogs(c *gin.Context) {
	if BalanceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)
	logType := c.Query("type")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := BalanceSvc.AdminGetBalanceLogs(page, pageSize, uint(userID), logType)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取记录失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    logs,
		"total":   total,
		"page":    page,
		"pages":   (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// AdminAdjustBalance 管理员调整用户余额
func AdminAdjustBalance(c *gin.Context) {
	if BalanceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		UserID uint    `json:"user_id" binding:"required"`
		Amount float64 `json:"amount" binding:"required"`
		Remark string  `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if req.Amount == 0 {
		c.JSON(400, gin.H{"success": false, "error": "调整金额不能为0"})
		return
	}

	// 构建操作者信息（安全获取 admin_id）
	var adminIDVal uint = 0
	if adminID, exists := c.Get("admin_id"); exists && adminID != nil {
		adminIDVal = adminID.(uint)
	}
	operator := &service.OperatorInfo{
		OperatorID:   adminIDVal,
		OperatorType: "admin",
		ClientIP:     GetClientIP(c),
	}

	if err := BalanceSvc.AdjustBalance(req.UserID, req.Amount, req.Remark, operator); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil && adminUsername != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "调整余额", "balance", strconv.FormatUint(uint64(req.UserID), 10), req, GetClientIP(c), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "调整成功"})
}

// AdminGetBalanceStats 管理员获取余额统计
func AdminGetBalanceStats(c *gin.Context) {
	if BalanceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	stats, err := BalanceSvc.GetBalanceStats()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取统计失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": stats})
}

// ==================== 余额支付 API ====================

// PayOrderWithBalance 使用余额支付订单
// 安全修复：使用冻结-扣除模式确保原子性
// 流程：1.验证支付密码 -> 2.冻结余额 -> 3.处理订单 -> 4.扣除冻结金额
// 如果步骤3失败，会自动解冻余额
func PayOrderWithBalance(c *gin.Context) {
	if BalanceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "余额服务未初始化"})
		return
	}

	if OrderSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "订单服务未初始化"})
		return
	}

	if PayPasswordSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "支付密码服务未初始化"})
		return
	}

	userID, _ := c.Get("user_id")

	var req struct {
		OrderNo     string `json:"order_no" binding:"required"`
		PayPassword string `json:"pay_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误，请提供订单号和支付密码"})
		return
	}

	// 验证支付密码
	if err := PayPasswordSvc.VerifyPayPassword(userID.(uint), req.PayPassword); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 获取订单信息
	order, err := OrderSvc.GetOrderByOrderNo(req.OrderNo)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "订单不存在"})
		return
	}

	// 验证订单归属
	if order.UserID != userID.(uint) {
		c.JSON(403, gin.H{"success": false, "error": "无权操作此订单"})
		return
	}

	// 验证订单状态
	if order.Status != 0 { // 0 = 待支付
		c.JSON(400, gin.H{"success": false, "error": "订单状态不正确，无法支付"})
		return
	}

	// 获取用户余额
	balance, err := BalanceSvc.GetUserBalance(userID.(uint))
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取余额失败"})
		return
	}

	// 检查余额是否充足
	if balance.Balance < order.Price {
		c.JSON(400, gin.H{"success": false, "error": "账户余额不足"})
		return
	}

	// 构建用户操作者信息
	operator := &service.OperatorInfo{
		OperatorID:   userID.(uint),
		OperatorType: "user",
		ClientIP:     GetClientIP(c),
	}

	// 步骤1：冻结余额（原子操作）
	err = BalanceSvc.Freeze(userID.(uint), order.Price, order.OrderNo, "购买商品冻结: "+order.ProductName, operator)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "余额冻结失败: " + err.Error()})
		return
	}

	// 步骤2：处理订单支付
	_, err = OrderSvc.ProcessPayment(order.OrderNo, "balance", "BAL"+order.OrderNo)
	if err != nil {
		// 订单处理失败，解冻余额
		unfreezeErr := BalanceSvc.Unfreeze(userID.(uint), order.Price, order.OrderNo, "支付失败解冻", operator)
		if unfreezeErr != nil {
			// 解冻失败，记录错误日志，需要人工处理
			c.JSON(500, gin.H{
				"success": false,
				"error":   "支付处理失败且解冻失败，请联系管理员处理。订单号: " + order.OrderNo,
			})
			return
		}
		c.JSON(500, gin.H{"success": false, "error": "支付处理失败: " + err.Error()})
		return
	}

	// 步骤3：扣除冻结金额（确认消费）
	err = BalanceSvc.DeductFrozen(userID.(uint), order.Price, order.OrderNo, "购买商品: "+order.ProductName, operator)
	if err != nil {
		// 这种情况理论上不应该发生，因为冻结金额已经存在
		// 但如果发生了，订单已经完成，需要人工处理
		c.JSON(200, gin.H{
			"success": true,
			"message": "支付成功，但余额扣除异常，请联系管理员处理",
			"warning": "余额扣除异常: " + err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "支付成功",
	})
}
