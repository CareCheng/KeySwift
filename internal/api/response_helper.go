package api

import (
	"net/http"
	"reflect"
	"strconv"

	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== 响应辅助函数 ====================
// 本文件提供统一的 API 响应处理函数，减少 Handler 中的重复代码
// 使用规范：
// - 所有成功响应使用 Success* 系列函数
// - 所有错误响应使用 Error* 系列函数
// - 分页响应使用 PagedResponse 函数

// ==================== 成功响应 ====================

// SuccessResponse 返回成功响应（带数据）
// 参数：
//   - c: Gin 上下文
//   - data: 响应数据（可以是任意类型）
//
// 响应格式：{"success": true, "data": data}
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// SuccessMessage 返回成功响应（仅消息）
// 参数：
//   - c: Gin 上下文
//   - message: 成功消息
//
// 响应格式：{"success": true, "message": message}
func SuccessMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
	})
}

// SuccessWithMessageAndData 返回成功响应（消息和数据）
// 参数：
//   - c: Gin 上下文
//   - message: 成功消息
//   - data: 响应数据
//
// 响应格式：{"success": true, "message": message, "data": data}
func SuccessWithMessageAndData(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// ==================== 错误响应 ====================

// ErrorResponse 返回错误响应
// 参数：
//   - c: Gin 上下文
//   - code: HTTP 状态码
//   - message: 错误消息
//
// 响应格式：{"success": false, "error": message}
func ErrorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"success": false,
		"error":   message,
	})
}

// BadRequestError 返回 400 错误
func BadRequestError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusBadRequest, message)
}

// UnauthorizedError 返回 401 错误
func UnauthorizedError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusUnauthorized, message)
}

// ForbiddenError 返回 403 错误
func ForbiddenError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusForbidden, message)
}

// NotFoundError 返回 404 错误
func NotFoundError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusNotFound, message)
}

// InternalServerError 返回 500 错误
func InternalServerError(c *gin.Context, message string) {
	ErrorResponse(c, http.StatusInternalServerError, message)
}

// ParamError 返回参数错误响应（400）
// 这是最常用的错误响应，用于请求参数校验失败
func ParamError(c *gin.Context) {
	BadRequestError(c, "参数错误")
}

// ParamErrorWithMessage 返回带详细信息的参数错误响应
func ParamErrorWithMessage(c *gin.Context, message string) {
	BadRequestError(c, message)
}

// ServiceNotInitializedError 返回服务未初始化错误
func ServiceNotInitializedError(c *gin.Context) {
	InternalServerError(c, "服务未初始化")
}

// DatabaseNotConnectedError 返回数据库未连接错误
func DatabaseNotConnectedError(c *gin.Context) {
	InternalServerError(c, "数据库未连接")
}

// ==================== 分页响应 ====================

// PaginationParams 分页参数
type PaginationParams struct {
	Page     int // 当前页码（从1开始）
	PageSize int // 每页数量
}

// GetOffset 获取数据库查询偏移量
func (p PaginationParams) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetPagination 从请求中获取分页参数
// 参数：
//   - c: Gin 上下文
//   - defaultPageSize: 默认每页数量（如果为0则默认20）
//   - maxPageSize: 最大每页数量（如果为0则默认100）
//
// 返回：分页参数结构体
func GetPagination(c *gin.Context, defaultPageSize, maxPageSize int) PaginationParams {
	if defaultPageSize <= 0 {
		defaultPageSize = 20
	}
	if maxPageSize <= 0 {
		maxPageSize = 100
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", strconv.Itoa(defaultPageSize)))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > maxPageSize {
		pageSize = defaultPageSize
	}

	return PaginationParams{Page: page, PageSize: pageSize}
}

// CalculateTotalPages 计算总页数
func CalculateTotalPages(total int64, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	return int((total + int64(pageSize) - 1) / int64(pageSize))
}

// PagedResponse 返回分页响应
// 参数：
//   - c: Gin 上下文
//   - data: 分页数据列表
//   - total: 总记录数
//   - page: 当前页码
//   - pageSize: 每页数量
//
// 响应格式：{"success": true, "data": data, "total": total, "page": page, "page_size": pageSize, "pages": pages}
func PagedResponse(c *gin.Context, data interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"data":      data,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"pages":     CalculateTotalPages(total, pageSize),
	})
}

// PagedResponseWithParams 使用 PaginationParams 返回分页响应
func PagedResponseWithParams(c *gin.Context, data interface{}, total int64, params PaginationParams) {
	PagedResponse(c, data, total, params.Page, params.PageSize)
}

// ==================== 服务检查 ====================

// CheckService 检查服务是否已初始化
// 参数：
//   - c: Gin 上下文
//   - svc: 服务实例
//
// 返回：
//   - bool: 服务是否可用（如果不可用已自动返回错误响应）
//
// 使用示例：
//
//	if !CheckService(c, BalanceSvc) {
//	    return
//	}
func CheckService(c *gin.Context, svc interface{}) bool {
	if svc == nil || reflect.ValueOf(svc).IsNil() {
		ServiceNotInitializedError(c)
		return false
	}
	return true
}

// CheckDBConnected 检查数据库是否已连接
// 参数：
//   - c: Gin 上下文
//
// 返回：
//   - bool: 数据库是否已连接（如果未连接已自动返回错误响应）
//
// 使用示例：
//
//	if !CheckDBConnected(c) {
//	    return
//	}
func CheckDBConnected(c *gin.Context) bool {
	if !model.DBConnected {
		DatabaseNotConnectedError(c)
		return false
	}
	return true
}

// ==================== 用户认证辅助 ====================

// GetUserID 从上下文获取用户ID
// 参数：
//   - c: Gin 上下文
//
// 返回：
//   - uint: 用户ID
//   - bool: 是否成功获取（如果失败已自动返回 401 错误响应）
//
// 使用示例：
//
//	userID, ok := GetUserID(c)
//	if !ok {
//	    return
//	}
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		UnauthorizedError(c, "请先登录")
		return 0, false
	}
	uid, ok := userID.(uint)
	if !ok {
		UnauthorizedError(c, "用户信息无效")
		return 0, false
	}
	return uid, true
}

// GetUserIDOptional 从上下文获取用户ID（可选，不返回错误）
// 适用于支持游客访问的接口
// 返回：
//   - uint: 用户ID（未登录返回 0）
//   - bool: 是否已登录
func GetUserIDOptional(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	uid, ok := userID.(uint)
	if !ok {
		return 0, false
	}
	return uid, true
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) string {
	username, exists := c.Get("username")
	if !exists {
		return ""
	}
	name, ok := username.(string)
	if !ok {
		return ""
	}
	return name
}

// GetAdminUsername 从上下文获取管理员用户名
func GetAdminUsername(c *gin.Context) string {
	username, exists := c.Get("admin_username")
	if !exists {
		return ""
	}
	name, ok := username.(string)
	if !ok {
		return ""
	}
	return name
}

// GetAdminRole 从上下文获取管理员角色
func GetAdminRole(c *gin.Context) string {
	role, exists := c.Get("admin_role")
	if !exists {
		return ""
	}
	r, ok := role.(string)
	if !ok {
		return ""
	}
	return r
}

// ==================== 参数解析辅助 ====================

// ParseUintParam 解析 URL 路径参数为 uint
// 参数：
//   - c: Gin 上下文
//   - paramName: 参数名称（如 "id"）
//
// 返回：
//   - uint: 解析后的值
//   - bool: 是否成功（如果失败已自动返回错误响应）
//
// 使用示例：
//
//	id, ok := ParseUintParam(c, "id")
//	if !ok {
//	    return
//	}
func ParseUintParam(c *gin.Context, paramName string) (uint, bool) {
	paramValue := c.Param(paramName)
	id, err := strconv.ParseUint(paramValue, 10, 32)
	if err != nil {
		BadRequestError(c, "无效的"+paramName)
		return 0, false
	}
	return uint(id), true
}

// ParseIntParam 解析 URL 路径参数为 int
func ParseIntParam(c *gin.Context, paramName string) (int, bool) {
	paramValue := c.Param(paramName)
	id, err := strconv.Atoi(paramValue)
	if err != nil {
		BadRequestError(c, "无效的"+paramName)
		return 0, false
	}
	return id, true
}

// ParseQueryUint 解析查询参数为 uint
// 如果参数不存在返回默认值
func ParseQueryUint(c *gin.Context, paramName string, defaultValue uint) uint {
	paramValue := c.Query(paramName)
	if paramValue == "" {
		return defaultValue
	}
	id, err := strconv.ParseUint(paramValue, 10, 32)
	if err != nil {
		return defaultValue
	}
	return uint(id)
}

// ParseQueryInt 解析查询参数为 int
// 如果参数不存在返回默认值
func ParseQueryInt(c *gin.Context, paramName string, defaultValue int) int {
	paramValue := c.Query(paramName)
	if paramValue == "" {
		return defaultValue
	}
	id, err := strconv.Atoi(paramValue)
	if err != nil {
		return defaultValue
	}
	return id
}

// ==================== 操作日志辅助 ====================

// LogAdminOperation 记录管理员操作日志
// 参数：
//   - c: Gin 上下文
//   - action: 操作名称
//   - resourceType: 资源类型
//   - resourceID: 资源ID
//   - detail: 详细信息（可以是任意类型，会被 JSON 序列化）
func LogAdminOperation(c *gin.Context, action, resourceType, resourceID string, detail interface{}) {
	if LogSvc == nil {
		return
	}
	adminUsername := GetAdminUsername(c)
	LogSvc.LogAdminActionSimple(adminUsername, action, resourceType, resourceID, detail, GetClientIP(c), c.GetHeader("User-Agent"))
}

// LogUserOperation 记录用户操作日志
// 参数：
//   - c: Gin 上下文
//   - action: 操作名称
//   - resourceType: 资源类型
//   - resourceID: 资源ID
//   - detail: 详细信息
func LogUserOperation(c *gin.Context, action, resourceType, resourceID string, detail interface{}) {
	if LogSvc == nil {
		return
	}
	userID, ok := GetUserIDOptional(c)
	if !ok {
		return
	}
	username := GetUsername(c)
	LogSvc.LogUserActionSimple(userID, username, action, resourceType, resourceID, detail, GetClientIP(c), c.GetHeader("User-Agent"))
}

// ==================== JSON 绑定辅助 ====================

// BindJSON 绑定 JSON 请求体
// 参数：
//   - c: Gin 上下文
//   - req: 请求结构体指针
//
// 返回：
//   - bool: 是否成功（如果失败已自动返回错误响应）
//
// 使用示例：
//
//	var req struct {
//	    Name string `json:"name" binding:"required"`
//	}
//	if !BindJSON(c, &req) {
//	    return
//	}
func BindJSON(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		ParamError(c)
		return false
	}
	return true
}

// BindJSONWithError 绑定 JSON 请求体（返回详细错误）
func BindJSONWithError(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		ParamErrorWithMessage(c, "参数错误: "+err.Error())
		return false
	}
	return true
}

// ==================== 资源归属验证 ====================

// VerifyResourceOwner 验证资源是否属于当前用户
// 参数：
//   - c: Gin 上下文
//   - ownerID: 资源所有者ID
//   - userID: 当前用户ID
//
// 返回：
//   - bool: 是否属于当前用户（如果不属于已自动返回 403 错误响应）
func VerifyResourceOwner(c *gin.Context, ownerID, userID uint) bool {
	if ownerID != userID {
		ForbiddenError(c, "无权操作此资源")
		return false
	}
	return true
}

// ==================== 通用响应结构 ====================

// Response 通用响应结构
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// PagedData 分页数据结构
type PagedData struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// NewPagedData 创建分页数据结构
func NewPagedData(items interface{}, total int64, page, pageSize int) *PagedData {
	return &PagedData{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: CalculateTotalPages(total, pageSize),
	}
}
