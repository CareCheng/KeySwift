package api

import (
	"strconv"
	"time"

	"user-frontend/internal/model"
	"user-frontend/internal/repository"
	"user-frontend/internal/service"

	"github.com/gin-gonic/gin"
)

// ==================== 分类管理 ====================

// AdminGetCategories 获取分类列表
func AdminGetCategories(c *gin.Context) {
	if CategorySvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	categories, err := CategorySvc.GetAllCategories(false)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "categories": categories})
}

// AdminCreateCategory 创建分类
func AdminCreateCategory(c *gin.Context) {
	if CategorySvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Name      string `json:"name" binding:"required"`
		Icon      string `json:"icon"`
		SortOrder int    `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	category, err := CategorySvc.CreateCategory(req.Name, req.Icon, req.SortOrder)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(c.GetString("admin_username"), "create", "category", strconv.Itoa(int(category.ID)), req.Name, GetClientIP(c), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "category": category})
}

// AdminUpdateCategory 更新分类
func AdminUpdateCategory(c *gin.Context) {
	if CategorySvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	var req struct {
		Name      string `json:"name"`
		Icon      string `json:"icon"`
		SortOrder int    `json:"sort_order"`
		Status    int    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	category, err := CategorySvc.UpdateCategory(uint(id), req.Name, req.Icon, req.SortOrder, req.Status)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(c.GetString("admin_username"), "update", "category", idStr, req.Name, GetClientIP(c), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "category": category})
}

// AdminDeleteCategory 删除分类
func AdminDeleteCategory(c *gin.Context) {
	if CategorySvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)

	if err := CategorySvc.DeleteCategory(uint(id)); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(c.GetString("admin_username"), "delete", "category", idStr, "", GetClientIP(c), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "分类已删除"})
}

// ==================== 操作日志 ====================

// AdminGetOperationLogs 获取操作日志
// 支持按日期查询，日志从加密的CSV文件中读取
func AdminGetOperationLogs(c *gin.Context) {
	if LogSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	userType := c.Query("user_type")
	action := c.Query("action")
	category := c.Query("category")
	date := c.Query("date") // 日期参数，格式: YYYY-MM-DD

	logs, total, err := LogSvc.GetOperationLogs(date, page, pageSize, userType, action, category)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 计算总页数
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	c.JSON(200, gin.H{
		"success":     true,
		"logs":        logs,
		"total":       total,
		"total_pages": totalPages,
		"page":        page,
	})
}

// AdminGetLogDates 获取可用的日志日期列表
func AdminGetLogDates(c *gin.Context) {
	if LogSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	dates, err := LogSvc.GetAvailableLogDates()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"dates":   dates,
	})
}

// AdminGetLogConfig 获取日志配置
func AdminGetLogConfig(c *gin.Context) {
	if LogSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	config := LogSvc.GetLogConfig()
	c.JSON(200, gin.H{
		"success": true,
		"config":  config,
	})
}

// AdminUpdateLogConfig 更新日志配置
func AdminUpdateLogConfig(c *gin.Context) {
	if LogSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		EnableUserLog  bool `json:"enable_user_log"`
		EnableAdminLog bool `json:"enable_admin_log"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	config := service.LogConfig{
		EnableUserLog:  req.EnableUserLog,
		EnableAdminLog: req.EnableAdminLog,
	}

	if err := LogSvc.UpdateLogConfig(config); err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "日志配置已更新",
	})
}

// ==================== 统计图表 ====================

// AdminGetStatsChart 获取统计图表数据
func AdminGetStatsChart(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	if OrderSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	if days <= 0 || days > 90 {
		days = 7
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	stats, err := OrderSvc.GetOrderStatsByDateRange(startDate, endDate)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"stats":   stats,
		"days":    days,
	})
}

// ==================== 订单搜索 ====================

// AdminSearchOrders 搜索订单（增强版）
func AdminSearchOrders(c *gin.Context) {
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

	params := &repository.OrderSearchParams{
		OrderNo:  c.Query("order_no"),
		Username: c.Query("username"),
	}

	if statusStr := c.Query("status"); statusStr != "" {
		status, _ := strconv.Atoi(statusStr)
		params.Status = &status
	}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		t, _ := time.Parse("2006-01-02", startDateStr)
		params.StartDate = &t
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		t, _ := time.Parse("2006-01-02", endDateStr)
		t = t.Add(24*time.Hour - time.Second) // 包含当天
		params.EndDate = &t
	}

	orders, total, err := OrderSvc.SearchOrders(params, page, pageSize)
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

// ==================== IP黑名单管理 ====================

// AdminGetBlacklist 获取IP黑名单列表
func AdminGetBlacklist(c *gin.Context) {
	entries := GetBlacklistEntries()
	c.JSON(200, gin.H{
		"success":   true,
		"blacklist": entries,
		"total":     len(entries),
	})
}

// AdminRemoveFromBlacklist 从黑名单中移除指定IP
func AdminRemoveFromBlacklist(c *gin.Context) {
	ip := c.Param("ip")
	if ip == "" {
		c.JSON(400, gin.H{"success": false, "error": "IP地址不能为空"})
		return
	}

	if RemoveFromBlacklist(ip) {
		// 记录操作日志
		if LogSvc != nil {
			LogSvc.LogAdminActionSimple("admin", "remove_blacklist", "blacklist", ip, "移除IP黑名单: "+ip, GetClientIP(c), c.GetHeader("User-Agent"))
		}
		c.JSON(200, gin.H{"success": true, "message": "已从黑名单中移除"})
	} else {
		c.JSON(404, gin.H{"success": false, "error": "该IP不在黑名单中"})
	}
}

// AdminClearBlacklist 清空所有黑名单
func AdminClearBlacklist(c *gin.Context) {
	count := ClearBlacklist()

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple("admin", "clear_blacklist", "blacklist", "", "清空IP黑名单", GetClientIP(c), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "已清空黑名单",
		"count":   count,
	})
}

// ==================== IP白名单管理 ====================

// AdminGetWhitelist 获取IP白名单配置
func AdminGetWhitelist(c *gin.Context) {
	if ConfigSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	enabled, whitelist, err := ConfigSvc.GetWhitelistConfig()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success":   true,
		"enabled":   enabled,
		"whitelist": whitelist,
	})
}

// AdminSaveWhitelist 保存IP白名单配置
func AdminSaveWhitelist(c *gin.Context) {
	if ConfigSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Enabled   bool     `json:"enabled"`
		Whitelist []string `json:"whitelist"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 保存白名单配置
	if err := ConfigSvc.UpdateWhitelistConfig(req.Enabled, req.Whitelist); err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		action := "update_whitelist"
		detail := "更新IP白名单配置"
		if req.Enabled {
			detail += "（已启用）"
		} else {
			detail += "（已禁用）"
		}
		LogSvc.LogAdminActionSimple(c.GetString("admin_username"), action, "whitelist", "", detail, GetClientIP(c), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "白名单配置已保存",
	})
}
