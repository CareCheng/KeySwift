package api

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ==================== CSRF 保护 ====================

var (
	csrfTokens   = make(map[string]*CSRFToken)
	csrfTokensMu sync.RWMutex
	csrfSecret   []byte
)

// CSRFToken CSRF令牌
type CSRFToken struct {
	Token     string
	SessionID string
	ExpiresAt time.Time
}

const (
	CSRFTokenExpiry   = 2 * time.Hour
	CSRFHeaderName    = "X-CSRF-Token"
	CSRFCookieName    = "csrf_token"
	CSRFFormFieldName = "_csrf"
)

func init() {
	// 生成CSRF密钥
	csrfSecret = make([]byte, 32)
	rand.Read(csrfSecret)
}

// GenerateCSRFToken 生成CSRF令牌
func GenerateCSRFToken(sessionID string) string {
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	csrfTokensMu.Lock()
	csrfTokens[token] = &CSRFToken{
		Token:     token,
		SessionID: sessionID,
		ExpiresAt: time.Now().Add(CSRFTokenExpiry),
	}
	csrfTokensMu.Unlock()

	return token
}

// ValidateCSRFToken 验证CSRF令牌
func ValidateCSRFToken(token, sessionID string) bool {
	if token == "" {
		return false
	}

	csrfTokensMu.RLock()
	csrfData, exists := csrfTokens[token]
	csrfTokensMu.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(csrfData.ExpiresAt) {
		csrfTokensMu.Lock()
		delete(csrfTokens, token)
		csrfTokensMu.Unlock()
		return false
	}

	// 验证会话匹配
	return csrfData.SessionID == sessionID
}

// CleanupExpiredCSRFTokens 清理过期的CSRF令牌
func CleanupExpiredCSRFTokens() {
	now := time.Now()
	csrfTokensMu.Lock()
	for token, data := range csrfTokens {
		if now.After(data.ExpiresAt) {
			delete(csrfTokens, token)
		}
	}
	csrfTokensMu.Unlock()
}

// CSRFMiddleware CSRF保护中间件
func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过安全方法（GET, HEAD, OPTIONS）
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// 跳过公开API（不需要CSRF保护的接口）
		path := c.Request.URL.Path
		skipPaths := []string{
			"/api/user/login",
			"/api/user/register",
			"/api/captcha",
			"/api/user/email/send_code",
			"/api/user/forgot",
			"/health",
			"/api/health",
		}
		for _, skip := range skipPaths {
			if strings.HasPrefix(path, skip) {
				c.Next()
				return
			}
		}

		// 获取会话ID
		sessionID, _ := c.Cookie("user_session")
		if sessionID == "" {
			sessionID, _ = c.Cookie("admin_session")
		}

		// 如果没有会话，跳过CSRF检查（未登录用户）
		if sessionID == "" {
			c.Next()
			return
		}

		// 从请求头或表单获取CSRF令牌
		token := c.GetHeader(CSRFHeaderName)
		if token == "" {
			token = c.PostForm(CSRFFormFieldName)
		}
		if token == "" {
			token, _ = c.Cookie(CSRFCookieName)
		}

		// 验证令牌
		if !ValidateCSRFToken(token, sessionID) {
			c.JSON(403, gin.H{"success": false, "error": "CSRF验证失败，请刷新页面重试"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SetCSRFCookie 设置CSRF Cookie
func SetCSRFCookie(c *gin.Context, sessionID string) string {
	token := GenerateCSRFToken(sessionID)
	// 设置CSRF Cookie（允许JavaScript读取）
	c.SetCookie(CSRFCookieName, token, int(CSRFTokenExpiry.Seconds()), "/", "", isSecureMode(), false)
	return token
}

// ==================== 安全响应头 ====================

// SecurityHeadersMiddleware 安全响应头中间件
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止点击劫持
		c.Header("X-Frame-Options", "SAMEORIGIN")

		// 防止MIME类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")

		// XSS保护
		c.Header("X-XSS-Protection", "1; mode=block")

		// 引用策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 内容安全策略（CSP）
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://api.qrserver.com; style-src 'self' 'unsafe-inline'; img-src 'self' data: https: blob:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'self';")

		// 权限策略
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// HTTPS严格传输安全（仅在HTTPS模式下）
		if isSecureMode() {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// ==================== API 限流 ====================

var (
	apiRateLimits   = make(map[string]*APIRateLimit)
	apiRateLimitsMu sync.RWMutex
)

// APIRateLimit API限流记录
type APIRateLimit struct {
	Count       int
	WindowStart time.Time
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Window      time.Duration
	MaxRequests int
}

// 不同API的限流配置
var rateLimitConfigs = map[string]RateLimitConfig{
	"login":         {Window: time.Minute, MaxRequests: 10},
	"register":      {Window: time.Minute, MaxRequests: 5},
	"email_code":    {Window: time.Minute, MaxRequests: 3},
	"forgot":        {Window: time.Minute, MaxRequests: 5},
	"public_browse": {Window: time.Minute, MaxRequests: 600},
	"user_status":   {Window: time.Minute, MaxRequests: 240},
	"api_default":   {Window: time.Minute, MaxRequests: 180},
	"admin_api":     {Window: time.Minute, MaxRequests: 180},
	"payment":       {Window: time.Minute, MaxRequests: 60},
	"balance_pay":   {Window: time.Minute, MaxRequests: 10}, // 余额支付限制
	"pay_password":  {Window: time.Minute, MaxRequests: 10}, // 支付密码操作限制
}

// RateLimitMiddleware API限流中间件
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipRateLimit(c) {
			c.Next()
			return
		}

		// 确定限流类型
		limitType := getRateLimitType(c)
		config := rateLimitConfigs[limitType]
		key := getRateLimitKey(c, limitType)

		allowed, remaining := checkRateLimit(key, config)

		// 设置限流响应头
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.MaxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if !allowed {
			retryAfter := int(config.Window.Seconds())
			RenderErrorPage(c, 429, "请求过于频繁，请稍后再试", retryAfter)
			c.Abort()
			return
		}

		c.Next()
	}
}

func shouldSkipRateLimit(c *gin.Context) bool {
	path := c.Request.URL.Path
	method := c.Request.Method

	if method == http.MethodOptions || method == http.MethodHead {
		return true
	}

	if path == "/" || path == "/health" || path == "/api/health" {
		return true
	}

	skipPrefixes := []string{
		"/static/",
		"/_next/",
		"/product-files/",
		"/uploads/",
	}
	for _, prefix := range skipPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	if !strings.HasPrefix(path, "/api/") && method == http.MethodGet {
		return true
	}

	return false
}

func getRateLimitType(c *gin.Context) string {
	path := c.Request.URL.Path
	method := c.Request.Method

	if method == http.MethodGet && isPublicBrowseAPI(path) {
		return "public_browse"
	}
	if method == http.MethodGet && isUserStatusAPI(path) {
		return "user_status"
	}

	limitType := "api_default"
	if strings.Contains(path, "/login") {
		limitType = "login"
	} else if strings.Contains(path, "/register") {
		limitType = "register"
	} else if strings.Contains(path, "/email/send_code") {
		limitType = "email_code"
	} else if strings.Contains(path, "/forgot") {
		limitType = "forgot"
	} else if strings.HasPrefix(path, "/api/admin") {
		limitType = "admin_api"
	} else if strings.Contains(path, "/pay/balance") {
		limitType = "balance_pay"
	} else if strings.Contains(path, "/pay-password") {
		limitType = "pay_password"
	} else if strings.Contains(path, "/payment") || strings.Contains(path, "/paypal") {
		limitType = "payment"
	}

	return limitType
}

func isPublicBrowseAPI(path string) bool {
	publicBrowsePaths := []string{
		"/api/products",
		"/api/product/",
		"/api/categories",
		"/api/payment/methods",
		"/api/captcha",
	}
	for _, publicPath := range publicBrowsePaths {
		if path == publicPath || strings.HasPrefix(path, publicPath) {
			return true
		}
	}
	return false
}

func isUserStatusAPI(path string) bool {
	statusPaths := []string{
		"/api/user/info",
		"/api/user/orders",
		"/api/user/kamis",
		"/api/user/balance",
		"/api/user/balance/logs",
		"/api/user/pay-password/status",
		"/api/user/2fa/status",
		"/api/user/email/code_length",
		"/api/user/2fa/info",
		"/api/order/detail/",
	}
	for _, statusPath := range statusPaths {
		if path == statusPath || strings.HasPrefix(path, statusPath) {
			return true
		}
	}
	return false
}

func getRateLimitKey(c *gin.Context, limitType string) string {
	if sessionID, err := c.Cookie("user_session"); err == nil && sessionID != "" {
		return "user:" + sessionID + ":" + limitType
	}
	if sessionID, err := c.Cookie("admin_session"); err == nil && sessionID != "" {
		return "admin:" + sessionID + ":" + limitType
	}

	return "ip:" + c.ClientIP() + ":" + limitType
}

// checkRateLimit 检查限流
func checkRateLimit(key string, config RateLimitConfig) (bool, int) {
	apiRateLimitsMu.Lock()
	defer apiRateLimitsMu.Unlock()

	now := time.Now()
	limit, exists := apiRateLimits[key]

	if !exists || now.Sub(limit.WindowStart) > config.Window {
		apiRateLimits[key] = &APIRateLimit{
			Count:       1,
			WindowStart: now,
		}
		return true, config.MaxRequests - 1
	}

	if limit.Count >= config.MaxRequests {
		return false, 0
	}

	limit.Count++
	return true, config.MaxRequests - limit.Count
}

// CleanupExpiredRateLimits 清理过期的限流记录
func CleanupExpiredRateLimits() {
	apiRateLimitsMu.Lock()
	defer apiRateLimitsMu.Unlock()

	now := time.Now()
	for key, limit := range apiRateLimits {
		if now.Sub(limit.WindowStart) > time.Hour {
			delete(apiRateLimits, key)
		}
	}
}

// ==================== 请求签名验证 ====================

// RequestSignature 请求签名数据
type RequestSignature struct {
	Timestamp int64  `json:"_ts"`
	Nonce     string `json:"_nonce"`
	Signature string `json:"_sig"`
}

// SignatureMiddleware 请求签名验证中间件（用于高安全性API）
func SignatureMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 仅对特定高安全性API启用
		path := c.Request.URL.Path
		requireSignature := strings.Contains(path, "/payment") ||
			strings.Contains(path, "/withdraw") ||
			strings.Contains(path, "/transfer")

		if !requireSignature {
			c.Next()
			return
		}

		timestamp := c.GetHeader("X-Timestamp")
		nonce := c.GetHeader("X-Nonce")
		signature := c.GetHeader("X-Signature")

		if timestamp == "" || nonce == "" || signature == "" {
			c.JSON(400, gin.H{"success": false, "error": "缺少签名参数"})
			c.Abort()
			return
		}

		// 验证时间戳（5分钟有效期）
		ts, err := parseTimestamp(timestamp)
		if err != nil || time.Since(time.Unix(ts, 0)) > 5*time.Minute {
			c.JSON(400, gin.H{"success": false, "error": "请求已过期"})
			c.Abort()
			return
		}

		// 验证签名
		sessionID, _ := c.Cookie("user_session")
		expectedSig := generateRequestSignature(c.Request.Method, path, timestamp, nonce, sessionID)
		if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
			c.JSON(400, gin.H{"success": false, "error": "签名验证失败"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// generateRequestSignature 生成请求签名
func generateRequestSignature(method, path, timestamp, nonce, sessionID string) string {
	data := method + "|" + path + "|" + timestamp + "|" + nonce + "|" + sessionID
	h := hmac.New(sha256.New, csrfSecret)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func parseTimestamp(ts string) (int64, error) {
	var timestamp int64
	_, err := strings.NewReader(ts).Read([]byte{})
	if err != nil {
		return 0, err
	}
	// 简单解析
	for _, c := range ts {
		if c >= '0' && c <= '9' {
			timestamp = timestamp*10 + int64(c-'0')
		}
	}
	return timestamp, nil
}

// ==================== Cookie 安全增强 ====================

// isSecureMode 检查是否为安全模式（HTTPS）
func isSecureMode() bool {
	// 可以通过环境变量或配置控制
	// 生产环境应返回true
	return false // 开发环境默认false
}

// SetSecureCookie 设置安全Cookie
func SetSecureCookie(c *gin.Context, name, value string, maxAge int, httpOnly bool) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(name, value, maxAge, "/", "", isSecureMode(), httpOnly)
}

// ==================== IP 黑名单 ====================

var (
	ipBlacklist   = make(map[string]time.Time)
	ipBlacklistMu sync.RWMutex
)

// AddToBlacklist 添加IP到黑名单
func AddToBlacklist(ip string, duration time.Duration) {
	ipBlacklistMu.Lock()
	ipBlacklist[ip] = time.Now().Add(duration)
	ipBlacklistMu.Unlock()
}

// IsBlacklisted 检查IP是否在黑名单中
func IsBlacklisted(ip string) bool {
	ipBlacklistMu.RLock()
	expiry, exists := ipBlacklist[ip]
	ipBlacklistMu.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(expiry) {
		ipBlacklistMu.Lock()
		delete(ipBlacklist, ip)
		ipBlacklistMu.Unlock()
		return false
	}

	return true
}

// IPBlacklistMiddleware IP黑名单中间件
func IPBlacklistMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsBlacklisted(c.ClientIP()) {
			RenderErrorPage(c, 403, "您的 IP 已被列入黑名单，访问被拒绝", 0)
			c.Abort()
			return
		}
		c.Next()
	}
}

// CleanupExpiredBlacklist 清理过期的黑名单
func CleanupExpiredBlacklist() {
	ipBlacklistMu.Lock()
	defer ipBlacklistMu.Unlock()

	now := time.Now()
	for ip, expiry := range ipBlacklist {
		if now.After(expiry) {
			delete(ipBlacklist, ip)
		}
	}
}

// BlacklistEntry 黑名单条目
type BlacklistEntry struct {
	IP        string `json:"ip"`
	ExpiresAt string `json:"expires_at"`
	Remaining int64  `json:"remaining"` // 剩余秒数
}

// GetBlacklistEntries 获取所有黑名单条目
func GetBlacklistEntries() []BlacklistEntry {
	ipBlacklistMu.RLock()
	defer ipBlacklistMu.RUnlock()

	now := time.Now()
	entries := make([]BlacklistEntry, 0)

	for ip, expiry := range ipBlacklist {
		if now.Before(expiry) {
			entries = append(entries, BlacklistEntry{
				IP:        ip,
				ExpiresAt: expiry.Format("2006-01-02 15:04:05"),
				Remaining: int64(expiry.Sub(now).Seconds()),
			})
		}
	}

	return entries
}

// RemoveFromBlacklist 从黑名单中移除IP
func RemoveFromBlacklist(ip string) bool {
	ipBlacklistMu.Lock()
	defer ipBlacklistMu.Unlock()

	if _, exists := ipBlacklist[ip]; exists {
		delete(ipBlacklist, ip)
		return true
	}
	return false
}

// ClearBlacklist 清空所有黑名单
func ClearBlacklist() int {
	ipBlacklistMu.Lock()
	defer ipBlacklistMu.Unlock()

	count := len(ipBlacklist)
	ipBlacklist = make(map[string]time.Time)
	return count
}

// ==================== IP 白名单 ====================

// IPWhitelistMiddleware IP白名单中间件（仅用于管理后台）
func IPWhitelistMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否启用白名单
		if ConfigSvc == nil {
			c.Next()
			return
		}

		enabled, whitelist, err := ConfigSvc.GetWhitelistConfig()
		if err != nil || !enabled || len(whitelist) == 0 {
			c.Next()
			return
		}

		// 检查IP是否在白名单中
		clientIP := c.ClientIP()
		for _, ip := range whitelist {
			if ip == clientIP {
				c.Next()
				return
			}
		}

		// IP不在白名单中，拒绝访问
		RenderErrorPage(c, 403, "您的 IP 不在白名单中，访问被拒绝", 0)
		c.Abort()
	}
}

// ==================== 安全清理任务 ====================

// StartSecurityCleanupTask 启动安全清理定时任务
func StartSecurityCleanupTask() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			CleanupExpiredCSRFTokens()
			CleanupExpiredRateLimits()
			CleanupExpiredBlacklist()
		}
	}()
}

// ==================== CSRF API ====================

// GetCSRFToken 获取CSRF令牌API
func GetCSRFToken(c *gin.Context) {
	sessionID, _ := c.Cookie("user_session")
	if sessionID == "" {
		sessionID, _ = c.Cookie("admin_session")
	}
	if sessionID == "" {
		// 为未登录用户生成临时会话ID
		sessionID = "anonymous_" + generateRandomString(16)
	}

	token := SetCSRFCookie(c, sessionID)
	c.JSON(200, gin.H{
		"success": true,
		"token":   token,
	})
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)[:length]
}
