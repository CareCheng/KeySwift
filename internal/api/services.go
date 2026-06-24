// Package api 提供 HTTP API 处理器
// services.go - 服务变量声明和初始化
package api

import (
	"context"
	"time"

	"user-frontend/internal/config"
	"user-frontend/internal/model"
	"user-frontend/internal/repository"
	"user-frontend/internal/service"
)

// ==================== 核心服务 ====================
var (
	UserSvc              *service.UserService              // 用户服务
	OrderSvc             *service.OrderService             // 订单服务
	OrderKernelSvc       *service.OrderKernelService       // 宿主订单状态裁决内核
	ProductSvc           *service.ProductService           // 商品服务
	EmailSvc             *service.EmailService             // 邮箱服务
	ConfigSvc            *service.ConfigService            // 配置服务（主数据库存储）
	DBConfigSvc          *service.ConfigService            // 数据库配置服务（SQLite配置数据库）
	SecuritySvc          *service.SecurityService          // 安全服务
	LogSvc               *service.LogService               // 日志服务
	CategorySvc          *service.CategoryService          // 分类服务
	SessionSvc           *service.SessionService           // 会话服务（数据库持久化）
	ManualKamiSvc        *service.ManualKamiService        // 手动卡密服务
	BalanceSvc           *service.BalanceService           // 余额服务
	PayPasswordSvc       *service.PayPasswordService       // 支付密码服务
	RoleSvc              *service.RoleService              // 角色权限服务
	SubjectSvc           *service.SubjectService           // 统一主体上下文服务
	GovernanceSvc        *service.GovernanceService        // 权限、审计、事件和任务治理服务
	ProductImageSvc      *service.ProductImageService      // 商品图片服务
	PluginSvc            *service.PluginService            // 插件治理服务
	PluginRuntimeSvc     *service.PluginRuntimeService     // 插件运行时服务
	HumanVerificationSvc *service.HumanVerificationService // 人机验证服务
)

// InitDBConfigService 初始化数据库配置服务（在主数据库初始化之前调用）
func InitDBConfigService(configSvc *service.ConfigService) {
	DBConfigSvc = configSvc
}

// InitServices 初始化所有服务
func InitServices(cfg *config.Config) {
	// 设置黑名单回调
	service.SetBlacklistCallback(func(key string, duration time.Duration) {
		AddToBlacklist(key, duration)
	})

	if model.DBConnected {
		repo := repository.NewRepository(model.DB)

		// 初始化角色和治理底座
		initGovernanceBaseServices(repo)

		// 初始化核心服务
		initCoreServices(repo, cfg)
		initSupportServices(repo)
		initPluginService(repo, cfg)

		if ConfigSvc != nil && !ConfigSvc.NeedsInitialSetup() && RoleSvc != nil {
			ensureDefaultAdmin(cfg)
		}

		// 启动核心清理任务
		go startScheduledTasks()
	}
}

// initCoreServices 初始化核心服务
func initCoreServices(repo *repository.Repository, cfg *config.Config) {
	UserSvc = service.NewUserService(repo)
	ProductSvc = service.NewProductService(repo)

	// 复用DBConfigSvc并设置repo，而不是创建新的ConfigService
	if DBConfigSvc != nil {
		DBConfigSvc.SetRepo(repo)
		ConfigSvc = DBConfigSvc
	} else {
		ConfigSvc = service.NewConfigService(repo)
	}

	// 从数据库加载系统配置
	loadSystemConfig(cfg)

	// 从数据库加载邮箱配置
	loadEmailConfig(cfg)
	EmailSvc = service.NewEmailService(repo, &cfg.EmailConfig)

	// 订单服务
	OrderKernelSvc = service.NewOrderKernelService(repo, GovernanceSvc)
	OrderSvc = service.NewOrderService(repo, cfg)
	OrderSvc.SetConfigService(ConfigSvc)
	OrderSvc.SetOrderKernelService(OrderKernelSvc)

	// 初始化安全服务
	SecuritySvc = service.NewSecurityService(repo)

	// 初始化日志服务（文件存储版本，不再使用数据库）
	LogSvc = service.NewLogService()

	// 初始化分类服务
	CategorySvc = service.NewCategoryService(repo)

	// 初始化会话服务
	SessionSvc = service.NewSessionService(repo)

	// 初始化手动卡密服务
	ManualKamiSvc = service.NewManualKamiService(repo)
	OrderSvc.SetManualKamiService(ManualKamiSvc)
}

// loadSystemConfig 从数据库加载系统配置
func loadSystemConfig(cfg *config.Config) {
	if sysCfg, err := ConfigSvc.GetSystemConfig(); err == nil && sysCfg.SystemTitle != "" {
		cfg.ServerConfig.SystemTitle = sysCfg.SystemTitle
		cfg.ServerConfig.AdminSuffix = sysCfg.AdminSuffix
		cfg.ServerConfig.EnableLogin = sysCfg.EnableLogin
		cfg.ServerConfig.AdminHumanVerificationEnabled = sysCfg.AdminHumanVerificationEnabled
		cfg.ServerConfig.AdminHumanVerificationProviderID = sysCfg.AdminHumanVerificationProviderID
		cfg.ServerConfig.AdminUsername = sysCfg.AdminUsername
		if sysCfg.AdminPassword != "" {
			cfg.ServerConfig.AdminPassword = sysCfg.AdminPassword
		}
		cfg.ServerConfig.AdminPasswordInitialized = sysCfg.AdminPasswordInitialized
		cfg.ServerConfig.Enable2FA = sysCfg.Enable2FA
		cfg.ServerConfig.TOTPSecret = sysCfg.TOTPSecret
		cfg.ServerConfig.EnableSessionTimeout = sysCfg.EnableSessionTimeout
		cfg.ServerConfig.SessionTimeout = sysCfg.SessionTimeout
		cfg.ServerConfig.UserAllowRegister = sysCfg.UserAllowRegister
		cfg.ServerConfig.UserLoginHumanVerificationEnabled = sysCfg.UserLoginHumanVerificationEnabled
		cfg.ServerConfig.UserLoginHumanVerificationProviderID = sysCfg.UserLoginHumanVerificationProviderID
		cfg.ServerConfig.UserRegisterHumanVerificationEnabled = sysCfg.UserRegisterHumanVerificationEnabled
		cfg.ServerConfig.UserRegisterHumanVerificationProviderID = sysCfg.UserRegisterHumanVerificationProviderID
		cfg.ServerConfig.UserRegisterHumanVerificationFollowLogin = sysCfg.UserRegisterHumanVerificationFollowLogin
		cfg.ServerConfig.UserEnable2FA = sysCfg.UserEnable2FA
		cfg.ServerConfig.UserRequireEmailVerification = sysCfg.UserRequireEmailVerification
		cfg.ServerConfig.UserEnableSessionTimeout = sysCfg.UserEnableSessionTimeout
		cfg.ServerConfig.UserSessionTimeout = sysCfg.UserSessionTimeout
	}
}

// loadEmailConfig 从数据库加载邮箱配置
func loadEmailConfig(cfg *config.Config) {
	if emailCfg, err := ConfigSvc.GetEmailConfig(); err == nil {
		cfg.EmailConfig = *emailCfg
	}
}

// initGovernanceBaseServices 初始化角色和治理底座。
func initGovernanceBaseServices(repo *repository.Repository) {
	// 角色权限服务
	RoleSvc = service.NewRoleService(repo)
	_ = RoleSvc.InitDefaultRoles()

	// 治理服务
	GovernanceSvc = service.NewGovernanceService(repo)
	_ = GovernanceSvc.SyncHostPermissions(model.AllPermissions)
}

// initSupportServices 初始化依赖核心服务的支撑服务。
func initSupportServices(repo *repository.Repository) {
	// 余额服务
	BalanceSvc = service.NewBalanceService(repo)

	// 支付密码服务
	PayPasswordSvc = service.NewPayPasswordService(repo)

	// 统一主体上下文服务
	SubjectSvc = service.NewSubjectService(SessionSvc, UserSvc, RoleSvc)

	// 商品图片服务
	ProductImageSvc = service.NewProductImageService(repo)
}

func ensureDefaultAdmin(cfg *config.Config) {
	if RoleSvc == nil {
		return
	}
	if _, err := RoleSvc.GetAdminByUsername(cfg.ServerConfig.AdminUsername); err == nil {
		return
	}
	_ = RoleSvc.CreateSuperAdmin(cfg.ServerConfig.AdminUsername, cfg.ServerConfig.AdminPassword)
}

// initPluginService 初始化宿主插件治理服务。
func initPluginService(repo *repository.Repository, cfg *config.Config) {
	pluginRoot := service.DefaultPluginRoot(cfg.ConfigDir)
	PluginSvc = service.NewPluginService(repo, pluginRoot)
	PluginSvc.SetGovernanceService(GovernanceSvc)
	PluginRuntimeSvc = service.NewPluginRuntimeService(repo, PluginSvc, GovernanceSvc)
	HumanVerificationSvc = service.NewHumanVerificationService(PluginSvc, GovernanceSvc)
	if _, err := PluginSvc.Refresh(context.Background()); err != nil {
		// 插件目录为空或不可用不应阻断宿主启动，后台接口会暴露当前状态。
		return
	}
	if RoleSvc != nil {
		extra := make([]model.Permission, 0)
		for _, item := range PluginSvc.Permissions() {
			name := item.Title
			if name == "" {
				name = item.Key
			}
			group := item.Namespace
			if group == "" {
				group = "插件权限"
			} else {
				group = "插件权限：" + group
			}
			extra = append(extra, model.Permission{
				Code:        item.Key,
				Name:        name,
				Description: item.Description,
				Group:       group,
			})
		}
		RoleSvc.SetExtraPermissions(extra)
	}
	if GovernanceSvc != nil {
		_ = GovernanceSvc.RegisterPermissions(PluginSvc.PermissionDefinitionInputs())
		_ = GovernanceSvc.SyncPluginConfigSchemas(PluginSvc.ConfigSchemas())
	}
}

// startScheduledTasks 启动定时任务
func startScheduledTasks() {
	// 每分钟执行一次
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		// 取消过期订单（30分钟未支付）
		if OrderSvc != nil {
			OrderSvc.CancelExpiredOrders(30)
		}
		// 清理安全服务过期记录
		if SecuritySvc != nil {
			SecuritySvc.CleanupExpiredRecords()
		}
		// 清理过期会话（数据库）
		if SessionSvc != nil {
			SessionSvc.CleanupExpiredSessions()
		}
		// 清理过期令牌
		CleanupExpiredTokens()
	}
}
