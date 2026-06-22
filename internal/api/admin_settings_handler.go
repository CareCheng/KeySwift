package api

import (
	"user-frontend/internal/config"
	"user-frontend/internal/model"
	"user-frontend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
)

// ==================== 系统设置相关 API ====================

// AdminGetSettings 获取系统设置（从数据库）
func AdminGetSettings(c *gin.Context) {
	if ConfigSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	sysCfg, err := ConfigSvc.GetSystemConfig()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "获取配置失败"})
		return
	}

	// 获取服务器端口配置
	serverPort, _ := ConfigSvc.GetServerPort()
	if serverPort <= 0 {
		serverPort = 8080
	}

	c.JSON(200, gin.H{
		"success": true,
		"settings": gin.H{
			"system_title":                    sysCfg.SystemTitle,
			"admin_suffix":                    sysCfg.AdminSuffix,
			"enable_login":                    sysCfg.EnableLogin,
			"enable_captcha":                  sysCfg.EnableCaptcha,
			"admin_username":                  sysCfg.AdminUsername,
			"enable_2fa":                      sysCfg.Enable2FA,
			"totp_secret":                     sysCfg.TOTPSecret,
			"enable_session_timeout":          sysCfg.EnableSessionTimeout,
			"session_timeout":                 sysCfg.SessionTimeout,
			"user_allow_register":             sysCfg.UserAllowRegister,
			"user_enable_captcha":             sysCfg.UserEnableCaptcha,
			"user_enable_2fa":                 sysCfg.UserEnable2FA,
			"user_require_email_verification": sysCfg.UserRequireEmailVerification,
			"user_enable_session_timeout":     sysCfg.UserEnableSessionTimeout,
			"user_session_timeout":            sysCfg.UserSessionTimeout,
			"server_port":                     serverPort,
		},
	})
}

// AdminSaveSettings 保存系统设置（到数据库）
func AdminSaveSettings(c *gin.Context) {
	if ConfigSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	var req struct {
		SystemTitle string `json:"system_title"`
		AdminSuffix string `json:"admin_suffix"`
		ServerPort  int    `json:"server_port"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 获取当前配置，如果获取失败则使用默认配置
	sysCfg, err := ConfigSvc.GetSystemConfig()
	if err != nil || sysCfg == nil {
		sysCfg = &service.SystemConfig{
			SystemTitle:                  config.GlobalConfig.ServerConfig.SystemTitle,
			AdminSuffix:                  config.GlobalConfig.ServerConfig.AdminSuffix,
			EnableLogin:                  config.GlobalConfig.ServerConfig.EnableLogin,
			EnableCaptcha:                config.GlobalConfig.ServerConfig.EnableCaptcha,
			AdminUsername:                config.GlobalConfig.ServerConfig.AdminUsername,
			AdminPassword:                config.GlobalConfig.ServerConfig.AdminPassword,
			Enable2FA:                    config.GlobalConfig.ServerConfig.Enable2FA,
			TOTPSecret:                   config.GlobalConfig.ServerConfig.TOTPSecret,
			EnableSessionTimeout:         config.GlobalConfig.ServerConfig.EnableSessionTimeout,
			SessionTimeout:               normalizeSettingsTimeout(config.GlobalConfig.ServerConfig.SessionTimeout, 60),
			UserAllowRegister:            config.GlobalConfig.ServerConfig.UserAllowRegister,
			UserEnableCaptcha:            config.GlobalConfig.ServerConfig.UserEnableCaptcha,
			UserEnable2FA:                config.GlobalConfig.ServerConfig.UserEnable2FA,
			UserRequireEmailVerification: config.GlobalConfig.ServerConfig.UserRequireEmailVerification,
			UserEnableSessionTimeout:     config.GlobalConfig.ServerConfig.UserEnableSessionTimeout,
			UserSessionTimeout:           normalizeSettingsTimeout(config.GlobalConfig.ServerConfig.UserSessionTimeout, 120),
			EnableWhitelist:              false,
			IPWhitelist:                  []string{},
		}
	}

	// 更新基本设置字段
	if req.SystemTitle != "" {
		sysCfg.SystemTitle = req.SystemTitle
	}
	if req.AdminSuffix != "" {
		sysCfg.AdminSuffix = req.AdminSuffix
	}
	if err := ConfigSvc.SaveSystemConfig(sysCfg); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "保存配置失败: " + err.Error()})
		return
	}

	// 保存服务器端口配置
	needRestart := false
	if req.ServerPort > 0 && req.ServerPort <= 65535 {
		currentPort, _ := ConfigSvc.GetServerPort()
		if currentPort != req.ServerPort {
			if err := ConfigSvc.SaveServerPort(req.ServerPort); err != nil {
				c.JSON(500, gin.H{"success": false, "error": "保存端口配置失败: " + err.Error()})
				return
			}
			needRestart = true
		}
	}

	// 同步更新全局配置
	config.GlobalConfig.ServerConfig.SystemTitle = sysCfg.SystemTitle
	config.GlobalConfig.ServerConfig.AdminSuffix = sysCfg.AdminSuffix

	message := "设置已保存"
	if needRestart {
		message = "设置已保存，端口更改需要重启程序后生效"
	}
	c.JSON(200, gin.H{"success": true, "message": message, "need_restart": needRestart})
}

// AdminSaveSecuritySettings 保存安全设置（到数据库）
func AdminSaveSecuritySettings(c *gin.Context) {
	if ConfigSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "数据库未连接"})
		return
	}

	var req struct {
		EnableLogin                  bool   `json:"enable_login"`
		EnableCaptcha                bool   `json:"enable_captcha"`
		AdminUsername                string `json:"admin_username"`
		AdminPassword                string `json:"admin_password"`
		Enable2FA                    bool   `json:"enable_2fa"`
		TOTPSecret                   string `json:"totp_secret"`
		EnableSessionTimeout         bool   `json:"enable_session_timeout"`
		SessionTimeout               int    `json:"session_timeout"`
		UserAllowRegister            bool   `json:"user_allow_register"`
		UserEnableCaptcha            bool   `json:"user_enable_captcha"`
		UserEnable2FA                bool   `json:"user_enable_2fa"`
		UserRequireEmailVerification bool   `json:"user_require_email_verification"`
		UserEnableSessionTimeout     bool   `json:"user_enable_session_timeout"`
		UserSessionTimeout           int    `json:"user_session_timeout"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 获取当前配置，如果获取失败则使用默认配置
	sysCfg, err := ConfigSvc.GetSystemConfig()
	if err != nil || sysCfg == nil {
		sysCfg = &service.SystemConfig{
			SystemTitle:              config.GlobalConfig.ServerConfig.SystemTitle,
			AdminSuffix:              config.GlobalConfig.ServerConfig.AdminSuffix,
			EnableLogin:              true,
			EnableCaptcha:            true,
			AdminUsername:            "admin",
			AdminPassword:            "admin123",
			Enable2FA:                false,
			EnableSessionTimeout:     true,
			SessionTimeout:           60,
			UserAllowRegister:        true,
			UserEnableCaptcha:        true,
			UserEnable2FA:            true,
			UserEnableSessionTimeout: true,
			UserSessionTimeout:       120,
			EnableWhitelist:          false,
			IPWhitelist:              []string{},
		}
	}

	// 检查新用户名是否与 admins 表中的用户名冲突
	if req.AdminUsername != "" && req.AdminUsername != sysCfg.AdminUsername {
		if model.DBConnected && RoleSvc != nil {
			if _, err := RoleSvc.GetAdminByUsername(req.AdminUsername); err == nil {
				c.JSON(400, gin.H{"success": false, "error": "用户名与多管理员系统中的账户冲突，请使用其他用户名"})
				return
			}
		}
	}

	// 更新安全相关字段
	sysCfg.EnableLogin = req.EnableLogin
	sysCfg.EnableCaptcha = req.EnableLogin && req.EnableCaptcha
	if req.AdminUsername != "" {
		sysCfg.AdminUsername = req.AdminUsername
	}
	if req.AdminPassword != "" {
		sysCfg.AdminPassword = req.AdminPassword
	}
	sysCfg.Enable2FA = req.EnableLogin && req.Enable2FA
	sysCfg.TOTPSecret = req.TOTPSecret
	sysCfg.EnableSessionTimeout = req.EnableLogin && req.EnableSessionTimeout
	sysCfg.SessionTimeout = normalizeSettingsTimeout(req.SessionTimeout, 60)
	sysCfg.UserAllowRegister = req.UserAllowRegister
	sysCfg.UserEnableCaptcha = req.UserEnableCaptcha
	sysCfg.UserEnable2FA = req.UserEnable2FA
	sysCfg.UserRequireEmailVerification = req.UserRequireEmailVerification && config.GlobalConfig.EmailConfig.Enabled
	sysCfg.UserEnableSessionTimeout = req.UserEnableSessionTimeout
	sysCfg.UserSessionTimeout = normalizeSettingsTimeout(req.UserSessionTimeout, 120)

	if err := ConfigSvc.SaveSystemConfig(sysCfg); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "保存配置失败: " + err.Error()})
		return
	}

	// 同步更新全局配置
	config.GlobalConfig.ServerConfig.EnableLogin = sysCfg.EnableLogin
	config.GlobalConfig.ServerConfig.EnableCaptcha = sysCfg.EnableCaptcha
	config.GlobalConfig.ServerConfig.AdminUsername = sysCfg.AdminUsername
	if req.AdminPassword != "" {
		config.GlobalConfig.ServerConfig.AdminPassword = sysCfg.AdminPassword
	}
	config.GlobalConfig.ServerConfig.Enable2FA = sysCfg.Enable2FA
	config.GlobalConfig.ServerConfig.TOTPSecret = sysCfg.TOTPSecret
	config.GlobalConfig.ServerConfig.EnableSessionTimeout = sysCfg.EnableSessionTimeout
	config.GlobalConfig.ServerConfig.SessionTimeout = sysCfg.SessionTimeout
	config.GlobalConfig.ServerConfig.UserAllowRegister = sysCfg.UserAllowRegister
	config.GlobalConfig.ServerConfig.UserEnableCaptcha = sysCfg.UserEnableCaptcha
	config.GlobalConfig.ServerConfig.UserEnable2FA = sysCfg.UserEnable2FA
	config.GlobalConfig.ServerConfig.UserRequireEmailVerification = sysCfg.UserRequireEmailVerification
	config.GlobalConfig.ServerConfig.UserEnableSessionTimeout = sysCfg.UserEnableSessionTimeout
	config.GlobalConfig.ServerConfig.UserSessionTimeout = sysCfg.UserSessionTimeout

	c.JSON(200, gin.H{"success": true, "message": "安全设置已保存"})
}

// PublicAuthConfig 返回前台和后台登录页需要的非敏感认证配置。
func PublicAuthConfig(c *gin.Context) {
	serverCfg := config.GlobalConfig.ServerConfig
	emailEnabled := config.GlobalConfig.EmailConfig.Enabled
	c.JSON(200, gin.H{
		"success": true,
		"config": gin.H{
			"admin_enable_login":              serverCfg.EnableLogin,
			"admin_enable_captcha":            serverCfg.EnableLogin && serverCfg.EnableCaptcha,
			"user_allow_register":             serverCfg.UserAllowRegister,
			"user_enable_captcha":             serverCfg.UserEnableCaptcha,
			"user_enable_2fa":                 serverCfg.UserEnable2FA,
			"user_require_email_verification": serverCfg.UserRequireEmailVerification && emailEnabled,
			"email_enabled":                   emailEnabled,
		},
	})
}

func normalizeSettingsTimeout(value, fallback int) int {
	if value <= 0 {
		value = fallback
	}
	if value < 5 {
		return 5
	}
	if value > 1440 {
		return 1440
	}
	return value
}

// AdminGenerate2FASecret 生成2FA密钥
func AdminGenerate2FASecret(c *gin.Context) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      config.GlobalConfig.ServerConfig.SystemTitle,
		AccountName: config.GlobalConfig.ServerConfig.AdminUsername,
	})
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "生成密钥失败"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"secret":  key.Secret(),
		"url":     key.URL(),
	})
}

// AdminVerify2FACode 验证2FA验证码
func AdminVerify2FACode(c *gin.Context) {
	var req struct {
		Code   string `json:"code" binding:"required"`
		Secret string `json:"secret" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if totp.Validate(req.Code, req.Secret) {
		c.JSON(200, gin.H{"success": true, "message": "验证通过"})
	} else {
		c.JSON(400, gin.H{"success": false, "error": "验证码错误"})
	}
}
