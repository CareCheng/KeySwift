package api

import (
	"encoding/json"
	"strconv"
	"time"

	"user-frontend/internal/model"
	pluginapi "user-frontend/internal/plugin"
	"user-frontend/internal/service"

	"github.com/gin-gonic/gin"
)

// ==================== 角色管理 API ====================

type roleResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Permissions []string  `json:"permissions"`
	IsSystem    bool      `json:"is_system"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type adminResponse struct {
	ID          uint       `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Nickname    string     `json:"nickname"`
	RoleID      uint       `json:"role_id"`
	RoleName    string     `json:"role_name"`
	Enable2FA   bool       `json:"enable_2fa"`
	Status      int        `json:"status"`
	LastLoginAt *time.Time `json:"last_login_at"`
	LastLoginIP string     `json:"last_login_ip"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type pluginPermissionResponse struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Group       string `json:"group"`
	PluginID    string `json:"plugin_id"`
	Scope       string `json:"scope"`
	RiskLevel   string `json:"risk_level"`
}

func buildRoleResponse(role model.AdminRole) roleResponse {
	permissions := make([]string, 0)
	if role.Permissions != "" {
		_ = json.Unmarshal([]byte(role.Permissions), &permissions)
	}
	return roleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Permissions: permissions,
		IsSystem:    role.IsSystem,
		Status:      role.Status,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

func buildRoleResponses(roles []model.AdminRole) []roleResponse {
	items := make([]roleResponse, 0, len(roles))
	for _, role := range roles {
		items = append(items, buildRoleResponse(role))
	}
	return items
}

func buildAdminResponse(admin model.Admin) adminResponse {
	roleName := ""
	if admin.Role != nil {
		roleName = admin.Role.Name
	}
	return adminResponse{
		ID:          admin.ID,
		Username:    admin.Username,
		Email:       admin.Email,
		Nickname:    admin.Nickname,
		RoleID:      admin.RoleID,
		RoleName:    roleName,
		Enable2FA:   admin.Enable2FA,
		Status:      admin.Status,
		LastLoginAt: admin.LastLoginAt,
		LastLoginIP: admin.LastLoginIP,
		CreatedAt:   admin.CreatedAt,
		UpdatedAt:   admin.UpdatedAt,
	}
}

func buildAdminResponses(admins []model.Admin) []adminResponse {
	items := make([]adminResponse, 0, len(admins))
	for _, admin := range admins {
		items = append(items, buildAdminResponse(admin))
	}
	return items
}

func pluginPermissionToModel(item pluginapi.PermissionDeclaration) model.Permission {
	group := item.Namespace
	if group == "" {
		group = "插件权限"
	} else {
		group = "插件权限：" + group
	}
	name := item.Title
	if name == "" {
		name = item.Key
	}
	return model.Permission{
		Code:        item.Key,
		Name:        name,
		Description: item.Description,
		Group:       group,
	}
}

func buildPluginPermissionResponses(items []pluginapi.PermissionDeclaration) []pluginPermissionResponse {
	permissions := make([]pluginPermissionResponse, 0, len(items))
	for _, item := range items {
		modelPerm := pluginPermissionToModel(item)
		permissions = append(permissions, pluginPermissionResponse{
			Code:        modelPerm.Code,
			Name:        modelPerm.Name,
			Description: modelPerm.Description,
			Group:       modelPerm.Group,
			PluginID:    item.Namespace,
			Scope:       item.Scope,
			RiskLevel:   item.RiskLevel,
		})
	}
	return permissions
}

func syncPluginPermissionsToRoleService() []pluginapi.PermissionDeclaration {
	if RoleSvc == nil {
		return nil
	}
	if PluginSvc == nil {
		RoleSvc.SetExtraPermissions(nil)
		return nil
	}
	pluginPermissions := PluginSvc.Permissions()
	extra := make([]model.Permission, 0, len(pluginPermissions))
	for _, item := range pluginPermissions {
		extra = append(extra, pluginPermissionToModel(item))
	}
	RoleSvc.SetExtraPermissions(extra)
	return pluginPermissions
}

// AdminGetRoles 获取所有角色
func AdminGetRoles(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	roles, err := RoleSvc.GetAllRoles()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取角色列表失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": buildRoleResponses(roles)})
}

// AdminGetRole 获取角色详情
func AdminGetRole(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的角色ID"})
		return
	}

	role, err := RoleSvc.GetRoleByID(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "角色不存在"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": buildRoleResponse(*role)})
}

// AdminCreateRole 创建角色
func AdminCreateRole(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}
	syncPluginPermissionsToRoleService()

	var req struct {
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Permissions []string `json:"permissions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	role, err := RoleSvc.CreateRole(req.Name, req.Description, req.Permissions)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}
	syncRolePermissionGrants(c, role.ID, req.Permissions)

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "创建角色", "role", "", req, GetClientIP(c), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "data": buildRoleResponse(*role)})
}

// AdminUpdateRole 更新角色
func AdminUpdateRole(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}
	syncPluginPermissionsToRoleService()

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的角色ID"})
		return
	}

	var req struct {
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Permissions []string `json:"permissions"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := RoleSvc.UpdateRole(uint(id), req.Name, req.Description, req.Permissions); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}
	syncRolePermissionGrants(c, uint(id), req.Permissions)

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "更新角色", "role", c.Param("id"), req, GetClientIP(c), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "更新成功"})
}

// AdminDeleteRole 删除角色
func AdminDeleteRole(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的角色ID"})
		return
	}

	// 获取角色信息用于日志
	role, _ := RoleSvc.GetRoleByID(uint(id))
	roleName := ""
	if role != nil {
		roleName = role.Name
	}

	if err := RoleSvc.DeleteRole(uint(id)); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "删除角色", "role", c.Param("id"), gin.H{"name": roleName}, GetClientIP(c), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "删除成功"})
}

// AdminGetPermissions 获取所有权限定义
func AdminGetPermissions(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	pluginPermissions := syncPluginPermissionsToRoleService()
	permissions := RoleSvc.GetAllPermissions()
	groups := RoleSvc.GetPermissionGroups()
	var definitions []model.PermissionDefinition
	if GovernanceSvc != nil {
		if records, err := GovernanceSvc.ListPermissionDefinitions(); err == nil {
			definitions = records
		}
	}

	c.JSON(200, gin.H{
		"success":            true,
		"permissions":        permissions,
		"definitions":        definitions,
		"host_permissions":   model.AllPermissions,
		"plugin_permissions": buildPluginPermissionResponses(pluginPermissions),
		"groups":             groups,
		"templates":          model.PermissionTemplates,
	})
}

// AdminGrantSubjectDataScope 为主体授予数据范围。
func AdminGrantSubjectDataScope(c *gin.Context) {
	if GovernanceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "治理服务未初始化"})
		return
	}
	var req struct {
		SubjectID     string `json:"subject_id" binding:"required"`
		SubjectType   string `json:"subject_type"`
		ResourceType  string `json:"resource_type" binding:"required"`
		ScopeType     string `json:"scope_type" binding:"required"`
		ScopeValue    string `json:"scope_value"`
		OwnerPluginID string `json:"owner_plugin_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}
	subjectID := currentSubjectID(c)
	err := GovernanceSvc.GrantSubjectDataScope(req.SubjectID, req.SubjectType, service.DataScope{
		ResourceType:  req.ResourceType,
		ScopeType:     req.ScopeType,
		ScopeValue:    req.ScopeValue,
		OwnerPluginID: req.OwnerPluginID,
	}, subjectID)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}
	_ = GovernanceSvc.RecordAudit(service.AuditInput{
		ActorSubjectID: subjectID,
		Action:         "subject.data_scope.grant",
		ResourceType:   "subject",
		ResourceID:     req.SubjectID,
		RiskLevel:      "high",
		IP:             GetClientIP(c),
		UserAgent:      c.GetHeader("User-Agent"),
		Payload:        req,
	})
	c.JSON(200, gin.H{"success": true})
}

// AdminGetSubjectDataScopes 获取主体数据范围。
func AdminGetSubjectDataScopes(c *gin.Context) {
	if GovernanceSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "治理服务未初始化"})
		return
	}
	scopes, err := GovernanceSvc.ListSubjectDataScopes(c.Query("subject_id"))
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "读取数据范围失败"})
		return
	}
	c.JSON(200, gin.H{"success": true, "scopes": scopes})
}

func syncRolePermissionGrants(c *gin.Context, roleID uint, permissions []string) {
	if GovernanceSvc == nil {
		return
	}
	subjectID := currentSubjectID(c)
	_ = GovernanceSvc.SyncRolePermissionGrants(roleID, permissions, subjectID)
}

// ==================== 管理员管理 API ====================

// AdminGetAdmins 获取管理员列表
func AdminGetAdmins(c *gin.Context) {
	if RoleSvc == nil {
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

	admins, total, err := RoleSvc.GetAllAdmins(page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取管理员列表失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    buildAdminResponses(admins),
		"total":   total,
		"page":    page,
		"pages":   (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// AdminGetAdmin 获取管理员详情
func AdminGetAdmin(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的管理员ID"})
		return
	}

	admin, err := RoleSvc.GetAdminByID(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "管理员不存在"})
		return
	}

	c.JSON(200, gin.H{"success": true, "data": buildAdminResponse(*admin)})
}

// AdminCreateAdmin 创建管理员
func AdminCreateAdmin(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
		Email    string `json:"email"`
		Nickname string `json:"nickname"`
		RoleID   uint   `json:"role_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误，密码至少6位"})
		return
	}

	admin, err := RoleSvc.CreateAdmin(req.Username, req.Password, req.Email, req.Nickname, req.RoleID)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "创建管理员", "admin", "", gin.H{"username": req.Username, "role_id": req.RoleID}, GetClientIP(c), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "data": buildAdminResponse(*admin)})
}

// AdminUpdateAdmin 更新管理员
func AdminUpdateAdmin(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的管理员ID"})
		return
	}

	var req struct {
		Email    string `json:"email"`
		Nickname string `json:"nickname"`
		RoleID   uint   `json:"role_id" binding:"required"`
		Status   int    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := RoleSvc.UpdateAdmin(uint(id), req.Email, req.Nickname, req.RoleID, req.Status); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "更新管理员", "admin", c.Param("id"), req, GetClientIP(c), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "更新成功"})
}

// AdminUpdateAdminPassword 更新管理员密码
func AdminUpdateAdminPassword(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的管理员ID"})
		return
	}

	var req struct {
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "密码至少6位"})
		return
	}

	if err := RoleSvc.UpdateAdminPassword(uint(id), req.Password); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "重置密码", "admin", c.Param("id"), nil, GetClientIP(c), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "密码更新成功"})
}

// AdminDeleteAdmin 删除管理员
func AdminDeleteAdmin(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "无效的管理员ID"})
		return
	}

	// 获取管理员信息用于日志
	admin, _ := RoleSvc.GetAdminByID(uint(id))
	adminName := ""
	if admin != nil {
		adminName = admin.Username
	}

	if err := RoleSvc.DeleteAdmin(uint(id)); err != nil {
		c.JSON(400, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	adminUsername, _ := c.Get("admin_username")
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple(adminUsername.(string), "删除管理员", "admin", c.Param("id"), gin.H{"username": adminName}, GetClientIP(c), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "删除成功"})
}

// AdminGetMyPermissions 获取当前管理员的权限
func AdminGetMyPermissions(c *gin.Context) {
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	adminUsername, exists := c.Get("admin_username")
	if !exists {
		c.JSON(401, gin.H{"success": false, "error": "未登录"})
		return
	}

	admin, err := RoleSvc.GetAdminByUsername(adminUsername.(string))
	if err != nil {
		c.JSON(404, gin.H{"success": false, "error": "管理员不存在"})
		return
	}

	permissions, err := RoleSvc.GetAdminPermissions(admin.ID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取权限失败"})
		return
	}

	c.JSON(200, gin.H{
		"success":     true,
		"permissions": permissions,
		"role":        admin.Role,
	})
}

// PermissionRequired 权限检查中间件
func PermissionRequired(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if subject, ok := CurrentSubjectContext(c); ok {
			if subjectHasRole(subject, "super_admin") || subjectHasPermission(subject, permission) {
				c.Next()
				return
			}
			c.JSON(403, gin.H{"success": false, "error": "无权限执行此操作"})
			c.Abort()
			return
		}

		adminUsername, exists := c.Get("admin_username")
		if !exists {
			c.JSON(401, gin.H{"success": false, "error": "未登录"})
			c.Abort()
			return
		}

		// 获取管理员角色
		adminRole, exists := c.Get("admin_role")
		if exists && adminRole.(string) == "super_admin" {
			// 超级管理员拥有所有权限
			c.Next()
			return
		}

		if RoleSvc == nil {
			c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
			c.Abort()
			return
		}

		admin, err := RoleSvc.GetAdminByUsername(adminUsername.(string))
		if err != nil {
			c.JSON(403, gin.H{"success": false, "error": "无权限访问"})
			c.Abort()
			return
		}

		if !RoleSvc.AdminHasPermission(admin.ID, permission) {
			c.JSON(403, gin.H{"success": false, "error": "无权限执行此操作"})
			c.Abort()
			return
		}

		c.Next()
	}
}
