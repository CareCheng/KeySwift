// Package api 提供 HTTP API 处理器
// router.go - 路由注册
package api

import (
	"user-frontend/internal/config"
	"user-frontend/internal/static"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(r *gin.Engine, cfg *config.Config) {
	// 全局安全中间件
	r.Use(SecurityHeadersMiddleware())
	r.Use(IPBlacklistMiddleware())
	r.Use(RateLimitMiddleware())

	// 启动安全清理任务
	StartSecurityCleanupTask()

	// 静态文件服务（自动选择嵌入式或外部模式）
	// 注意：SetupStaticRoutes 内部已处理 /product-files 和 /uploads 路由
	static.SetupStaticRoutes(r)

	// CSRF令牌API
	r.GET("/api/csrf-token", GetCSRFToken)
	r.GET("/api/auth/config", PublicAuthConfig)

	// 注册各模块路由
	registerUserRoutes(r)
	registerOrderRoutes(r)
	registerPaymentRoutes(r)
	registerProductRoutes(r)
	registerAdminRoutes(r, cfg)
	registerSPARoutes(r, cfg)
}

// registerUserRoutes 注册用户相关路由
func registerUserRoutes(r *gin.Engine) {
	userAPI := r.Group("/api/user")
	{
		// 认证相关
		userAPI.POST("/register", UserRegister)
		userAPI.POST("/login", UserLogin)
		userAPI.POST("/logout", UserLogout)
		userAPI.GET("/info", AuthRequired(), UserInfo)
		userAPI.PUT("/info", AuthRequired(), UpdateUserInfo)
		userAPI.POST("/password", AuthRequired(), UpdatePassword)
		userAPI.GET("/orders", AuthRequired(), UserOrders)

		// 2FA相关
		userAPI.POST("/2fa/enable", AuthRequired(), Enable2FA)
		userAPI.POST("/2fa/disable", AuthRequired(), Disable2FA)
		userAPI.GET("/2fa/status", AuthRequired(), Get2FAStatus)
		userAPI.POST("/2fa/generate", AuthRequired(), Generate2FASecret)
		userAPI.POST("/2fa/preference", AuthRequired(), Set2FAPreference)
		userAPI.GET("/2fa/info", Get2FAInfo)
		userAPI.POST("/2fa/verify_login", Verify2FALogin)
		userAPI.POST("/2fa/enable_email", AuthRequired(), Enable2FAEmail)
		userAPI.POST("/2fa/verify_totp", AuthRequired(), VerifyTOTP)

		// 邮箱验证
		userAPI.POST("/email/send_code", SendEmailCode)
		userAPI.POST("/email/verify", VerifyEmailCode)
		userAPI.POST("/email/verify_only", VerifyEmailCodeOnly) // 仅验证不消耗
		userAPI.GET("/email/code_length", GetEmailCodeLength)   // 获取验证码长度
		userAPI.POST("/email/bind", AuthRequired(), BindEmail)

		// 忘记密码
		userAPI.POST("/forgot/check", ForgotPasswordCheck)
		userAPI.POST("/forgot/verify", ForgotPasswordVerify)
		userAPI.POST("/forgot/reset", ForgotPasswordReset)

		// 余额系统
		userAPI.GET("/balance", AuthRequired(), GetMyBalance)
		userAPI.GET("/balance/logs", AuthRequired(), GetMyBalanceLogs)

		// 支付密码
		userAPI.GET("/pay-password/status", AuthRequired(), GetPayPasswordStatus)
		userAPI.POST("/pay-password/set", AuthRequired(), SetPayPassword)
		userAPI.POST("/pay-password/update", AuthRequired(), UpdatePayPassword)
		userAPI.POST("/pay-password/reset", AuthRequired(), ResetPayPassword)
		userAPI.POST("/pay-password/verify", AuthRequired(), VerifyPayPassword)
		userAPI.POST("/pay-password/send-reset-code", AuthRequired(), SendResetPayPasswordCode)

		// 卡密
		userAPI.GET("/kamis", AuthRequired(), GetUserKamis)
	}
}

// registerOrderRoutes 注册订单相关路由
func registerOrderRoutes(r *gin.Engine) {
	orderAPI := r.Group("/api/order")
	{
		orderAPI.POST("/create", AuthRequired(), CreateOrder)
		orderAPI.GET("/detail/:order_no", AuthRequired(), OrderDetail)
		orderAPI.POST("/cancel", AuthRequired(), CancelOrder)
		orderAPI.POST("/pay/balance", AuthRequired(), PayOrderWithBalance)
	}

	// 订单查询（未登录，通过订单号+邮箱）
	r.POST("/api/order/query", QueryOrderPublic)
}

// registerPaymentRoutes 注册支付相关路由
func registerPaymentRoutes(r *gin.Engine) {
	r.GET("/api/payment/methods", GetPaymentMethods)
}

// registerProductRoutes 注册商品相关路由
func registerProductRoutes(r *gin.Engine) {
	// 商品API（公开）
	r.GET("/api/products", GetProducts)
	r.GET("/api/product/:id", GetProduct)
	r.GET("/api/product/:id/images", GetProductImages)
	r.GET("/api/product/:id/detail-file", GetProductDetailFile)
	r.GET("/api/categories", GetCategories)

	// 健康检查
	r.GET("/health", HealthCheck)
	r.GET("/api/health", HealthCheck)

	// 验证码
	r.GET("/api/captcha", CaptchaHandler)
	r.POST("/api/captcha/verify", VerifyCaptcha)
}

// registerAdminRoutes 注册管理后台路由
func registerAdminRoutes(r *gin.Engine, cfg *config.Config) {
	adminSuffix := cfg.ServerConfig.AdminSuffix

	// 初始化设置检查（无需认证）
	r.GET("/"+adminSuffix+"/check-setup", CheckInitialSetup)
	r.POST("/"+adminSuffix+"/setup", SetInitialPassword)

	// 管理员登录相关
	r.POST("/"+adminSuffix+"/login", AdminLogin)
	r.POST("/"+adminSuffix+"/totp", AdminVerifyTOTP)
	r.POST("/"+adminSuffix+"/logout", AdminLogout)

	// 管理员信息API
	r.GET("/api/admin/info", AdminAuthRequired(), AdminInfo)

	// 获取管理后台入口配置
	r.GET("/api/admin/suffix", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true, "suffix": adminSuffix})
	})

	// 管理API
	adminAPI := r.Group("/api/admin")
	adminAPI.Use(IPWhitelistMiddleware())
	adminAPI.Use(AdminAuthRequired())
	{
		// 仪表盘
		adminAPI.GET("/dashboard", PermissionRequired("dashboard:view"), AdminDashboard)

		// 商品管理
		registerAdminProductRoutes(adminAPI)

		// 订单管理
		registerAdminOrderRoutes(adminAPI)

		// 用户管理
		registerAdminUserRoutes(adminAPI)

		// 余额管理
		registerAdminBalanceRoutes(adminAPI)

		// 系统设置
		registerAdminSettingsRoutes(adminAPI)

		// 系统管理
		registerAdminSystemRoutes(adminAPI)

		// 插件治理
		registerAdminPluginRoutes(adminAPI)
	}
}

// registerAdminProductRoutes 注册管理后台商品相关路由
func registerAdminProductRoutes(adminAPI *gin.RouterGroup) {
	adminAPI.GET("/products", PermissionRequired("product:view"), AdminGetProducts)
	adminAPI.POST("/product", PermissionRequired("product:create"), AdminCreateProduct)
	adminAPI.PUT("/product/:id", PermissionRequired("product:edit"), AdminUpdateProduct)
	adminAPI.DELETE("/product/:id", PermissionRequired("product:delete"), AdminDeleteProduct)
	adminAPI.GET("/product/:id/images", PermissionRequired("product:view"), AdminGetProductImages)
	adminAPI.POST("/product/:id/images", PermissionRequired("product:edit"), AdminUploadProductImages)
	adminAPI.DELETE("/product/:id/images/:image_id", PermissionRequired("product:edit"), AdminDeleteProductImages)
	adminAPI.POST("/product/:id/images/primary", PermissionRequired("product:edit"), AdminSetPrimaryImage)
	adminAPI.PUT("/product/:id/images/order", PermissionRequired("product:edit"), AdminUpdateImageOrder)
	adminAPI.POST("/product/:id/detail-file", PermissionRequired("product:edit"), SaveProductDetailFile)
	adminAPI.POST("/product/:id/detail-image", PermissionRequired("product:edit"), UploadProductDetailImage)

	// 分类管理
	adminAPI.GET("/categories", PermissionRequired("category:view"), AdminGetCategories)
	adminAPI.POST("/category", PermissionRequired("category:create"), AdminCreateCategory)
	adminAPI.PUT("/category/:id", PermissionRequired("category:edit"), AdminUpdateCategory)
	adminAPI.DELETE("/category/:id", PermissionRequired("category:delete"), AdminDeleteCategory)

	// 手动卡密管理
	adminAPI.POST("/product/:id/kami/import", PermissionRequired("product:edit"), AdminImportKami)
	adminAPI.GET("/product/:id/kami", PermissionRequired("product:view"), AdminGetProductKamis)
	adminAPI.GET("/product/:id/kami/stats", PermissionRequired("product:view"), AdminGetKamiStats)
	adminAPI.DELETE("/kami/:id", PermissionRequired("product:edit"), AdminDeleteKami)
	adminAPI.POST("/kami/:id/disable", PermissionRequired("product:edit"), AdminDisableKami)
	adminAPI.POST("/kami/:id/enable", PermissionRequired("product:edit"), AdminEnableKami)
	adminAPI.POST("/kami/batch-delete", PermissionRequired("product:edit"), AdminBatchDeleteKamis)
}

// registerAdminOrderRoutes 注册管理后台订单相关路由
func registerAdminOrderRoutes(adminAPI *gin.RouterGroup) {
	adminAPI.GET("/orders", PermissionRequired("order:view"), AdminGetOrders)
	adminAPI.GET("/orders/search", PermissionRequired("order:view"), AdminSearchOrders)
	adminAPI.GET("/order/:id", PermissionRequired("order:view"), AdminGetOrder)
}

// registerAdminUserRoutes 注册管理后台用户相关路由
func registerAdminUserRoutes(adminAPI *gin.RouterGroup) {
	adminAPI.GET("/users", PermissionRequired("user:view"), AdminGetUsers)
	adminAPI.PUT("/user/:id/status", PermissionRequired("user:edit"), AdminUpdateUserStatus)
}

// registerAdminBalanceRoutes 注册管理后台余额相关路由。
func registerAdminBalanceRoutes(adminAPI *gin.RouterGroup) {
	adminAPI.GET("/balances", PermissionRequired("balance:view"), AdminGetBalances)
	adminAPI.GET("/balance/logs", PermissionRequired("balance:view"), AdminGetBalanceLogs)
	adminAPI.GET("/balance/stats", PermissionRequired("balance:view"), AdminGetBalanceStats)
	adminAPI.POST("/balance/adjust", PermissionRequired("balance:adjust"), AdminAdjustBalance)
}

// registerAdminSettingsRoutes 注册管理后台设置相关路由
func registerAdminSettingsRoutes(adminAPI *gin.RouterGroup) {
	// 系统设置
	adminAPI.GET("/settings", PermissionRequired("settings:view"), AdminGetSettings)
	adminAPI.POST("/settings", PermissionRequired("settings:edit"), AdminSaveSettings)
	adminAPI.POST("/settings/security", PermissionRequired("settings:security"), AdminSaveSecuritySettings)

	// 数据库配置
	adminAPI.GET("/db/config", PermissionRequired("settings:database"), AdminGetDBConfig)
	adminAPI.POST("/db/config", PermissionRequired("settings:database"), AdminSaveDBConfig)
	adminAPI.POST("/db/test", PermissionRequired("settings:database"), AdminTestDBConnection)
	adminAPI.POST("/db/reset-key", PermissionRequired("settings:database"), AdminResetEncryptionKey)

	// 2FA设置
	adminAPI.POST("/2fa/enable", PermissionRequired("settings:security"), AdminEnable2FA)
	adminAPI.POST("/2fa/disable", PermissionRequired("settings:security"), AdminDisable2FA)
	adminAPI.GET("/2fa/status", PermissionRequired("settings:security"), AdminGet2FAStatus)
	adminAPI.POST("/2fa/generate", PermissionRequired("settings:security"), AdminGenerate2FASecret)
	adminAPI.POST("/2fa/verify", PermissionRequired("settings:security"), AdminVerify2FACode)

	// 支付配置
	adminAPI.GET("/payment/config", PermissionRequired("settings:payment"), AdminGetPaymentConfig)

	// 邮箱配置
	adminAPI.GET("/email/config", PermissionRequired("settings:email"), AdminGetEmailConfig)
	adminAPI.POST("/email/config", PermissionRequired("settings:email"), AdminSaveEmailConfig)
	adminAPI.POST("/email/test", PermissionRequired("settings:email"), AdminTestEmail)
}

// registerAdminSystemRoutes 注册管理后台系统管理路由
func registerAdminSystemRoutes(adminAPI *gin.RouterGroup) {
	// IP黑名单管理
	adminAPI.GET("/blacklist", PermissionRequired("settings:security"), AdminGetBlacklist)
	adminAPI.DELETE("/blacklist/:ip", PermissionRequired("settings:security"), AdminRemoveFromBlacklist)
	adminAPI.DELETE("/blacklist", PermissionRequired("settings:security"), AdminClearBlacklist)

	// IP白名单管理
	adminAPI.GET("/whitelist", PermissionRequired("settings:security"), AdminGetWhitelist)
	adminAPI.POST("/whitelist", PermissionRequired("settings:security"), AdminSaveWhitelist)

	// 角色权限管理
	adminAPI.GET("/roles", PermissionRequired("role:view"), AdminGetRoles)
	adminAPI.GET("/role/:id", PermissionRequired("role:view"), AdminGetRole)
	adminAPI.POST("/role", PermissionRequired("role:create"), AdminCreateRole)
	adminAPI.PUT("/role/:id", PermissionRequired("role:edit"), AdminUpdateRole)
	adminAPI.DELETE("/role/:id", PermissionRequired("role:delete"), AdminDeleteRole)
	adminAPI.GET("/permissions", PermissionRequired("role:view"), AdminGetPermissions)
	adminAPI.GET("/admins", PermissionRequired("admin:view"), AdminGetAdmins)
	adminAPI.GET("/admin/:id", PermissionRequired("admin:view"), AdminGetAdmin)
	adminAPI.POST("/admin", PermissionRequired("admin:create"), AdminCreateAdmin)
	adminAPI.PUT("/admin/:id", PermissionRequired("admin:edit"), AdminUpdateAdmin)
	adminAPI.PUT("/admin/:id/password", PermissionRequired("admin:edit"), AdminUpdateAdminPassword)
	adminAPI.DELETE("/admin/:id", PermissionRequired("admin:delete"), AdminDeleteAdmin)
	adminAPI.GET("/my-permissions", AdminGetMyPermissions)
}

// registerAdminPluginRoutes 注册插件治理路由。
func registerAdminPluginRoutes(adminAPI *gin.RouterGroup) {
	adminAPI.GET("/plugins/summary", PermissionRequired("plugin:view"), AdminGetPluginSummary)
	adminAPI.GET("/plugins", PermissionRequired("plugin:view"), AdminListPlugins)
	adminAPI.POST("/plugins/refresh", PermissionRequired("plugin:manage"), AdminRefreshPlugins)
	adminAPI.GET("/plugins/frontend", AdminGetPluginFrontendContribution)
	adminAPI.GET("/plugins/permissions", PermissionRequired("plugin:view"), AdminGetPluginPermissions)
	adminAPI.GET("/plugins/config-schemas", PermissionRequired("plugin:view"), AdminGetPluginConfigSchemas)
	adminAPI.GET("/plugin/:id", PermissionRequired("plugin:view"), AdminGetPluginDetail)
	adminAPI.GET("/plugin/:id/bindings", PermissionRequired("plugin:view"), AdminGetPluginBindings)
	adminAPI.GET("/plugin/:id/migrations", PermissionRequired("plugin:view"), AdminGetPluginMigrations)
	adminAPI.GET("/plugin/:id/configs", PermissionRequired("plugin:view"), AdminGetPluginConfigs)
	adminAPI.GET("/plugin/:id/database", PermissionRequired("plugin:view"), AdminGetPluginDatabaseTables)
	adminAPI.POST("/plugin/:id/enable", PermissionRequired("plugin:manage"), AdminEnablePlugin)
	adminAPI.POST("/plugin/:id/disable", PermissionRequired("plugin:manage"), AdminDisablePlugin)
}

// registerSPARoutes 注册 SPA 前端页面路由
func registerSPARoutes(r *gin.Engine, cfg *config.Config) {
	adminSuffix := cfg.ServerConfig.AdminSuffix

	// 用户前台页面
	r.GET("/", ServeReactPage("index.html"))
	r.GET("/login", ServeReactPage("login/index.html"))
	r.GET("/login/", ServeReactPage("login/index.html"))
	r.GET("/register", ServeReactPage("register/index.html"))
	r.GET("/register/", ServeReactPage("register/index.html"))
	r.GET("/forgot", ServeReactPage("forgot/index.html"))
	r.GET("/forgot/", ServeReactPage("forgot/index.html"))
	r.GET("/verify", ServeReactPage("verify/index.html"))
	r.GET("/verify/", ServeReactPage("verify/index.html"))
	r.GET("/products", ServeReactPage("products/index.html"))
	r.GET("/products/", ServeReactPage("products/index.html"))
	r.GET("/product", ServeReactPage("product/index.html"))
	r.GET("/product/", ServeReactPage("product/index.html"))
	r.GET("/order/detail", ServeReactPage("order/detail/index.html"))
	r.GET("/order/detail/", ServeReactPage("order/detail/index.html"))
	r.GET("/user", ServeReactPage("user/index.html"))
	r.GET("/user/", ServeReactPage("user/index.html"))

	// 支付相关页面
	r.GET("/payment", ServeReactPage("payment/index.html"))
	r.GET("/payment/", ServeReactPage("payment/index.html"))
	r.GET("/payment/result", ServeReactPage("payment/result/index.html"))
	r.GET("/payment/result/", ServeReactPage("payment/result/index.html"))

	// 管理后台页面（使用动态后缀）
	r.GET("/"+adminSuffix, ServeReactPage("admin/index.html"))
	r.GET("/"+adminSuffix+"/", ServeReactPage("admin/index.html"))
	r.GET("/"+adminSuffix+"/login", ServeReactPage("admin/login/index.html"))
	r.GET("/"+adminSuffix+"/login/", ServeReactPage("admin/login/index.html"))
	r.GET("/"+adminSuffix+"/totp", ServeReactPage("admin/totp/index.html"))
	r.GET("/"+adminSuffix+"/totp/", ServeReactPage("admin/totp/index.html"))
	r.GET("/"+adminSuffix+"/setup", ServeReactPage("admin/setup/index.html"))
	r.GET("/"+adminSuffix+"/setup/", ServeReactPage("admin/setup/index.html"))
}

// ServeReactPage 返回React静态页面的处理函数
// 自动支持嵌入式和外部资源模式
func ServeReactPage(pagePath string) gin.HandlerFunc {
	return static.ServeEmbeddedPage(pagePath)
}
