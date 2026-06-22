// Package api 提供 HTTP API 处理器
// user_2fa_handler.go - 用户两步验证管理处理器
package api

import (
	"sync"
	"time"

	"user-frontend/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
)

// 登录验证令牌存储
var (
	loginTokens   = make(map[string]*LoginToken)
	loginTokensMu sync.RWMutex
)

// LoginToken 登录验证令牌
type LoginToken struct {
	UserID          uint
	Username        string
	Email           string
	PreferEmailAuth bool
	ExpiresAt       time.Time
}

// Enable2FA 启用两步验证
func Enable2FA(c *gin.Context) {
	if !config.GlobalConfig.ServerConfig.UserEnable2FA {
		c.JSON(403, gin.H{"success": false, "error": "用户两步验证暂未开放"})
		return
	}

	if UserSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")
	username := c.GetString("username")

	var req struct {
		Secret string `json:"secret" binding:"required"`
		Code   string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 验证TOTP码
	if !totp.Validate(req.Code, req.Secret) {
		c.JSON(400, gin.H{"success": false, "error": "验证码错误"})
		return
	}

	if err := UserSvc.Enable2FA(userID, req.Secret); err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogUserActionSimple(userID, username, "enable_2fa", "user", "", nil, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{"success": true, "message": "两步验证已启用"})
}

// Disable2FA 禁用两步验证
func Disable2FA(c *gin.Context) {
	if UserSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")

	var req struct {
		TOTPCode  string `json:"totp_code"`
		EmailCode string `json:"email_code"`
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

	// 根据当前验证方式验证
	verified := false
	if user.PreferEmailAuth || user.TOTPSecret == "" {
		// 邮箱验证
		if req.EmailCode == "" {
			c.JSON(400, gin.H{"success": false, "error": "请输入邮箱验证码"})
			return
		}
		if EmailSvc != nil && EmailSvc.VerifyCode(user.Email, req.EmailCode, "disable_2fa") {
			verified = true
		} else {
			c.JSON(400, gin.H{"success": false, "error": "邮箱验证码错误或已过期"})
			return
		}
	} else {
		// TOTP验证
		if req.TOTPCode == "" {
			c.JSON(400, gin.H{"success": false, "error": "请输入动态口令"})
			return
		}
		if totp.Validate(req.TOTPCode, user.TOTPSecret) {
			verified = true
		} else {
			c.JSON(400, gin.H{"success": false, "error": "动态口令错误"})
			return
		}
	}

	if verified {
		if err := UserSvc.Disable2FA(userID); err != nil {
			c.JSON(500, gin.H{"success": false, "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"success": true, "message": "两步验证已禁用"})
	}
}

// Get2FAStatus 获取2FA状态
func Get2FAStatus(c *gin.Context) {
	if UserSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")

	user, err := UserSvc.GetUserByID(userID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success":           true,
		"enabled":           user.Enable2FA,
		"prefer_email_auth": user.PreferEmailAuth,
		"email_verified":    user.EmailVerified,
		"has_totp":          user.TOTPSecret != "",
	})
}

// Generate2FASecret 生成2FA密钥
func Generate2FASecret(c *gin.Context) {
	if !config.GlobalConfig.ServerConfig.UserEnable2FA {
		c.JSON(403, gin.H{"success": false, "error": "用户两步验证暂未开放"})
		return
	}

	username := c.GetString("username")

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      config.GlobalConfig.ServerConfig.SystemTitle,
		AccountName: username,
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

// Set2FAPreference 设置2FA验证方式偏好
func Set2FAPreference(c *gin.Context) {
	if UserSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")

	var req struct {
		PreferEmailAuth bool `json:"prefer_email_auth"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	if err := UserSvc.SetPreferEmailAuth(userID, req.PreferEmailAuth); err != nil {
		c.JSON(500, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "设置已保存"})
}

// Get2FAInfo 获取登录验证信息
func Get2FAInfo(c *gin.Context) {
	if UserSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	token := c.Query("token")
	if token == "" {
		c.JSON(400, gin.H{"success": false, "error": "无效的验证请求"})
		return
	}

	loginTokensMu.RLock()
	tokenData, exists := loginTokens[token]
	loginTokensMu.RUnlock()

	if !exists {
		c.JSON(400, gin.H{"success": false, "error": "验证信息已过期"})
		return
	}

	if time.Now().After(tokenData.ExpiresAt) {
		loginTokensMu.Lock()
		delete(loginTokens, token)
		loginTokensMu.Unlock()
		c.JSON(400, gin.H{"success": false, "error": "验证信息已过期"})
		return
	}

	// 获取用户信息以判断是否设置了TOTP
	user, err := UserSvc.GetUserByID(tokenData.UserID)
	hasTOTP := false
	if err == nil && user.TOTPSecret != "" {
		hasTOTP = true
	}

	c.JSON(200, gin.H{
		"success":      true,
		"username":     tokenData.Username,
		"email":        tokenData.Email,
		"masked_email": maskEmail(tokenData.Email),
		"prefer_email": tokenData.PreferEmailAuth,
		"has_totp":     hasTOTP,
	})
}

// Verify2FAPage 二次验证页面
func Verify2FAPage(c *gin.Context) {
	c.HTML(200, "verify_2fa.html", gin.H{
		"title": "二次验证",
	})
}

// Verify2FALogin 完成二次验证登录
func Verify2FALogin(c *gin.Context) {
	if UserSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	var req struct {
		Token     string `json:"token" binding:"required"`
		TOTPCode  string `json:"totp_code"`
		EmailCode string `json:"email_code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	loginTokensMu.RLock()
	tokenData, exists := loginTokens[req.Token]
	loginTokensMu.RUnlock()

	if !exists {
		c.JSON(400, gin.H{"success": false, "error": "验证信息已过期，请重新登录"})
		return
	}

	if time.Now().After(tokenData.ExpiresAt) {
		loginTokensMu.Lock()
		delete(loginTokens, req.Token)
		loginTokensMu.Unlock()
		c.JSON(400, gin.H{"success": false, "error": "验证信息已过期，请重新登录"})
		return
	}

	// 获取用户信息
	user, err := UserSvc.GetUserByID(tokenData.UserID)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "error": "用户不存在"})
		return
	}

	// 验证
	verified := false
	if req.TOTPCode != "" {
		// TOTP验证
		if totp.Validate(req.TOTPCode, user.TOTPSecret) {
			verified = true
		} else {
			c.JSON(400, gin.H{"success": false, "error": "动态口令错误"})
			return
		}
	} else if req.EmailCode != "" {
		// 邮箱验证码验证
		if EmailSvc != nil && EmailSvc.VerifyCode(user.Email, req.EmailCode, "login") {
			verified = true
		} else {
			c.JSON(400, gin.H{"success": false, "error": "邮箱验证码错误或已过期"})
			return
		}
	} else {
		c.JSON(400, gin.H{"success": false, "error": "请提供验证码"})
		return
	}

	if verified {
		// 删除登录令牌
		loginTokensMu.Lock()
		delete(loginTokens, req.Token)
		loginTokensMu.Unlock()

		// 创建会话（数据库持久化）
		if SessionSvc == nil {
			c.JSON(500, gin.H{"success": false, "error": "会话服务未初始化"})
			return
		}
		sessionDuration, cookieMaxAge := userSessionPolicy(false)
		sessionID, err := SessionSvc.CreateUserSessionWithDuration(user.ID, user.Username, c.ClientIP(), c.GetHeader("User-Agent"), sessionDuration)
		if err != nil {
			c.JSON(500, gin.H{"success": false, "error": "创建会话失败"})
			return
		}

		SetSecureCookie(c, "user_session", sessionID, cookieMaxAge, true)

		// 设置CSRF令牌
		csrfToken := SetCSRFCookie(c, sessionID)

		c.JSON(200, gin.H{
			"success":    true,
			"message":    "登录成功",
			"csrf_token": csrfToken,
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
			},
		})
	}
}

// Enable2FAEmail 启用邮箱方式的两步验证
func Enable2FAEmail(c *gin.Context) {
	if !config.GlobalConfig.ServerConfig.UserEnable2FA {
		c.JSON(403, gin.H{"success": false, "error": "用户两步验证暂未开放"})
		return
	}

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

	// 检查邮箱是否已验证
	if !user.EmailVerified || user.Email == "" {
		c.JSON(400, gin.H{"success": false, "error": "请先验证邮箱"})
		return
	}

	// 启用两步验证（邮箱方式）
	user.Enable2FA = true
	user.PreferEmailAuth = true

	if err := UserSvc.UpdateUser(user); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "启用失败"})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "邮箱两步验证已启用"})
}

// VerifyTOTP 验证TOTP码（用于高风险操作验证）
func VerifyTOTP(c *gin.Context) {
	if UserSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "服务未初始化"})
		return
	}

	userID := c.GetUint("user_id")

	var req struct {
		Code string `json:"code" binding:"required"`
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

	if user.TOTPSecret == "" {
		c.JSON(400, gin.H{"success": false, "error": "未设置动态口令"})
		return
	}

	if !totp.Validate(req.Code, user.TOTPSecret) {
		c.JSON(400, gin.H{"success": false, "error": "动态口令错误"})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "验证成功"})
}
