package api

import (
	"log"
	"time"

	"user-frontend/internal/config"
	"user-frontend/internal/model"
	"user-frontend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
)

// ==================== 管理员认证相关 API ====================

// CheckInitialSetup 检查是否需要初始化设置
func CheckInitialSetup(c *gin.Context) {
	needsSetup := false

	if ConfigSvc != nil {
		needsSetup = ConfigSvc.NeedsInitialSetup()
		log.Printf("[CheckInitialSetup] ConfigSvc存在, needsSetup=%v", needsSetup)
	} else {
		// ConfigSvc 未初始化，检查全局配置
		cfg := config.GlobalConfig.ServerConfig
		needsSetup = !cfg.AdminPasswordInitialized
		log.Printf("[CheckInitialSetup] ConfigSvc为nil, 检查GlobalConfig, needsSetup=%v", needsSetup)
	}

	c.JSON(200, gin.H{
		"success":     true,
		"needs_setup": needsSetup,
	})
}

// SetInitialPassword 设置初始管理员密码
func SetInitialPassword(c *gin.Context) {
	var req struct {
		Password        string `json:"password" binding:"required,min=6"`
		ConfirmPassword string `json:"confirm_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "密码长度至少6位"})
		return
	}

	if req.Password != req.ConfirmPassword {
		c.JSON(400, gin.H{"success": false, "error": "两次输入的密码不一致"})
		return
	}

	// 检查是否需要初始化
	needsSetup := false
	if ConfigSvc != nil {
		needsSetup = ConfigSvc.NeedsInitialSetup()
	} else {
		cfg := config.GlobalConfig.ServerConfig
		needsSetup = !cfg.AdminPasswordInitialized
	}

	if !needsSetup {
		c.JSON(400, gin.H{"success": false, "error": "初始密码已设置，无法重复设置"})
		return
	}

	// 设置密码到配置
	if ConfigSvc != nil {
		if err := ConfigSvc.SetInitialPassword(req.Password); err != nil {
			c.JSON(500, gin.H{"success": false, "error": err.Error()})
			return
		}
	} else {
		// 更新全局配置（内存中）
		config.GlobalConfig.ServerConfig.AdminPassword = req.Password
		config.GlobalConfig.ServerConfig.AdminPasswordInitialized = true
	}

	if model.DBConnected {
		if RoleSvc != nil {
			existingAdmin, _ := RoleSvc.GetAdminByUsername("admin")
			if existingAdmin == nil {
				if err := RoleSvc.CreateSuperAdmin("admin", req.Password); err != nil {
					c.JSON(200, gin.H{
						"success": true,
						"message": "管理员密码设置成功（注意：数据库管理员创建失败）",
						"warning": err.Error(),
					})
					return
				}
			} else {
				if err := RoleSvc.UpdateAdminPassword(existingAdmin.ID, req.Password); err != nil {
					c.JSON(200, gin.H{
						"success": true,
						"message": "管理员密码设置成功（注意：数据库密码更新失败）",
						"warning": err.Error(),
					})
					return
				}
			}
		}
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "管理员密码设置成功",
	})
}

// AdminLoginPage 管理员登录页面
func AdminLoginPage(c *gin.Context) {
	c.HTML(200, "admin_login.html", gin.H{
		"title": "管理员登录",
	})
}

func adminSessionPolicy(remember bool) (time.Duration, int) {
	serverCfg := config.GlobalConfig.ServerConfig
	if remember {
		return 24 * time.Hour, 86400
	}
	if !serverCfg.EnableSessionTimeout {
		return service.LongLivedSessionDuration, int(service.LongLivedSessionDuration.Seconds())
	}
	timeout := serverCfg.SessionTimeout
	if timeout <= 0 {
		timeout = 60
	}
	if timeout < 5 {
		timeout = 5
	}
	if timeout > 1440 {
		timeout = 1440
	}
	duration := time.Duration(timeout) * time.Minute
	return duration, int(duration.Seconds())
}

// AdminTOTPPage 管理员TOTP验证页面
func AdminTOTPPage(c *gin.Context) {
	c.HTML(200, "admin_totp.html", gin.H{
		"title": "两步验证",
	})
}

// AdminLogin 管理员登录
func AdminLogin(c *gin.Context) {
	// 检查是否需要初始化设置
	if ConfigSvc != nil && ConfigSvc.NeedsInitialSetup() {
		c.JSON(400, gin.H{"success": false, "error": "请先完成初始化设置", "needs_setup": true})
		return
	}

	if !model.DBConnected {
		// 数据库未连接时使用配置文件中的管理员账号
		var req struct {
			Username          string                            `json:"username" binding:"required"`
			Password          string                            `json:"password" binding:"required"`
			HumanVerification *service.HumanVerificationPayload `json:"human_verification"`
			Remember          bool                              `json:"remember"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"success": false, "error": "参数错误"})
			return
		}

		if !verifyHumanVerificationForRequest(c, service.HumanScopeAdminLogin, req.HumanVerification) {
			return
		}

		cfg := config.GlobalConfig
		if req.Username != cfg.ServerConfig.AdminUsername || req.Password != cfg.ServerConfig.AdminPassword {
			c.JSON(400, gin.H{"success": false, "error": "用户名或密码错误"})
			return
		}

		// 创建会话（数据库持久化）
		if SessionSvc == nil {
			c.JSON(500, gin.H{"success": false, "error": "会话服务未初始化"})
			return
		}
		sessionDuration, cookieMaxAge := adminSessionPolicy(req.Remember)
		sessionID, err := SessionSvc.CreateAdminSessionWithDuration(req.Username, "super_admin", GetClientIP(c), c.GetHeader("User-Agent"), sessionDuration)
		if err != nil {
			c.JSON(500, gin.H{"success": false, "error": "创建会话失败"})
			return
		}

		// 如果未启用2FA则直接验证通过
		if !cfg.ServerConfig.Enable2FA {
			SessionSvc.SetAdminSessionVerified(sessionID)
		}

		SetSecureCookie(c, "admin_session", sessionID, cookieMaxAge, true)
		SetCSRFCookie(c, sessionID)

		if cfg.ServerConfig.Enable2FA {
			c.JSON(200, gin.H{
				"success":      true,
				"require_totp": true,
				"message":      "请完成两步验证",
			})
			return
		}

		c.JSON(200, gin.H{"success": true, "message": "登录成功"})
		return
	}

	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Username          string                            `json:"username" binding:"required"`
		Password          string                            `json:"password" binding:"required"`
		HumanVerification *service.HumanVerificationPayload `json:"human_verification"`
		Remember          bool                              `json:"remember"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if !verifyHumanVerificationForRequest(c, service.HumanScopeAdminLogin, req.HumanVerification) {
		return
	}

	admin, err := RoleSvc.VerifyAdminPassword(req.Username, req.Password)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "用户名或密码错误"})
		return
	}
	adminRole := "admin"
	if admin.Role != nil {
		adminRole = admin.Role.Name
	}
	_ = RoleSvc.UpdateAdminLoginInfo(admin.ID, GetClientIP(c))

	// 创建会话（数据库持久化）
	if SessionSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "会话服务未初始化"})
		return
	}
	sessionDuration, cookieMaxAge := adminSessionPolicy(req.Remember)
	sessionID, err := SessionSvc.CreateAdminSessionWithDuration(admin.Username, adminRole, GetClientIP(c), c.GetHeader("User-Agent"), sessionDuration)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "创建会话失败"})
		return
	}

	// 如果未启用2FA则直接验证通过
	if !admin.Enable2FA {
		SessionSvc.SetAdminSessionVerified(sessionID)
	}

	SetSecureCookie(c, "admin_session", sessionID, cookieMaxAge, true)
	SetCSRFCookie(c, sessionID)

	if admin.Enable2FA {
		c.JSON(200, gin.H{
			"success":      true,
			"require_totp": true,
			"message":      "请完成两步验证",
		})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "登录成功"})
}

// AdminVerifyTOTP 管理员TOTP验证
func AdminVerifyTOTP(c *gin.Context) {
	sessionID, _ := c.Cookie("admin_session")
	if sessionID == "" {
		c.JSON(401, gin.H{"success": false, "error": "请先登录"})
		return
	}

	if SessionSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "会话服务未初始化"})
		return
	}

	session, err := SessionSvc.GetAdminSession(sessionID)
	if err != nil {
		c.JSON(401, gin.H{"success": false, "error": "会话已过期"})
		return
	}

	var req struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 获取TOTP密钥
	var totpSecret string
	if model.DBConnected {
		if RoleSvc == nil {
			c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
			return
		}
		_, secret, err := RoleSvc.GetAdmin2FAStatus(session.Username)
		if err != nil {
			c.JSON(400, gin.H{"success": false, "error": "两步验证状态不存在"})
			return
		}
		totpSecret = secret
	} else {
		totpSecret = config.GlobalConfig.ServerConfig.TOTPSecret
	}

	if !totp.Validate(req.Code, totpSecret) {
		c.JSON(400, gin.H{"success": false, "error": "验证码错误"})
		return
	}

	// 更新会话状态
	SessionSvc.SetAdminSessionVerified(sessionID)

	c.JSON(200, gin.H{"success": true, "message": "验证成功"})
}

// AdminLogout 管理员登出
func AdminLogout(c *gin.Context) {
	sessionID, _ := c.Cookie("admin_session")
	if sessionID != "" && SessionSvc != nil {
		SessionSvc.DeleteAdminSession(sessionID)
	}

	// 清除Cookie
	clearConfiguredCookie(c, "admin_session", true)
	clearConfiguredCookie(c, "csrf_token", false)

	c.JSON(200, gin.H{"success": true, "message": "已退出登录"})
}

// AdminInfo 获取当前管理员信息
func AdminInfo(c *gin.Context) {
	username := c.GetString("admin_username")
	role := c.GetString("admin_role")

	if username == "" {
		c.JSON(401, gin.H{"success": false, "error": "未登录"})
		return
	}

	// 获取权限列表
	var permissions []string
	isSuperAdmin := false

	// 超级管理员（系统配置账户或 super_admin 角色）拥有所有权限
	if role == "super_admin" {
		isSuperAdmin = true
		for _, p := range model.AllPermissions {
			permissions = append(permissions, p.Code)
		}
	} else if model.DBConnected && RoleSvc != nil {
		// 从数据库获取管理员权限
		admin, err := RoleSvc.GetAdminByUsername(username)
		if err == nil && admin != nil {
			perms, err := RoleSvc.GetRolePermissions(admin.RoleID)
			if err == nil {
				permissions = perms
			}
			// 检查是否是超级管理员角色
			if admin.Role != nil && admin.Role.Name == "super_admin" {
				isSuperAdmin = true
				permissions = make([]string, 0, len(model.AllPermissions))
				for _, p := range model.AllPermissions {
					permissions = append(permissions, p.Code)
				}
			}
		}
	}

	// 如果没有获取到权限，给予基本的仪表盘查看权限
	if len(permissions) == 0 && !isSuperAdmin {
		permissions = []string{"dashboard:view"}
	}

	c.JSON(200, gin.H{
		"success": true,
		"admin": gin.H{
			"username":       username,
			"role":           role,
			"permissions":    permissions,
			"is_super_admin": isSuperAdmin,
		},
	})
}

// AdminEnable2FA 启用管理员2FA
func AdminEnable2FA(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	username := c.GetString("admin_username")

	var req struct {
		Secret string `json:"secret" binding:"required"`
		Code   string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if !totp.Validate(req.Code, req.Secret) {
		c.JSON(400, gin.H{"success": false, "error": "验证码错误"})
		return
	}

	if err := RoleSvc.EnableAdmin2FA(username, req.Secret); err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "两步验证已启用"})
}

// AdminDisable2FA 禁用管理员2FA
func AdminDisable2FA(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	username := c.GetString("admin_username")

	var req struct {
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if _, err := RoleSvc.VerifyAdminPassword(username, req.Password); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "密码错误"})
		return
	}
	if err := RoleSvc.DisableAdmin2FA(username); err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "两步验证已禁用"})
}

// AdminGet2FAStatus 获取管理员2FA状态
func AdminGet2FAStatus(c *gin.Context) {
	if !model.DBConnected {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}
	if RoleSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	username := c.GetString("admin_username")
	enabled, _, _ := RoleSvc.GetAdmin2FAStatus(username)

	c.JSON(200, gin.H{
		"success": true,
		"enabled": enabled,
	})
}
