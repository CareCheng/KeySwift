// Package service 提供业务逻辑服务
// config_system.go - 系统配置管理方法
package service

import (
	"encoding/json"
	"fmt"

	"user-frontend/internal/config"
	"user-frontend/internal/model"
)

// normalizeTimeoutMinutes 统一限制后台配置中的会话超时范围。
func normalizeTimeoutMinutes(value, defaultValue int) int {
	if value <= 0 {
		return defaultValue
	}
	if value < 5 {
		return 5
	}
	if value > 1440 {
		return 1440
	}
	return value
}

// SystemConfig 系统配置结构
type SystemConfig struct {
	SystemTitle                  string   `json:"system_title"`
	AdminSuffix                  string   `json:"admin_suffix"`
	EnableLogin                  bool     `json:"enable_login"`
	EnableCaptcha                bool     `json:"enable_captcha"`
	AdminUsername                string   `json:"admin_username"`
	AdminPassword                string   `json:"admin_password"`
	AdminPasswordInitialized     bool     `json:"admin_password_initialized"`
	Enable2FA                    bool     `json:"enable_2fa"`
	TOTPSecret                   string   `json:"totp_secret"`
	EnableSessionTimeout         bool     `json:"enable_session_timeout"`
	SessionTimeout               int      `json:"session_timeout"`
	UserAllowRegister            bool     `json:"user_allow_register"`
	UserEnableCaptcha            bool     `json:"user_enable_captcha"`
	UserEnable2FA                bool     `json:"user_enable_2fa"`
	UserRequireEmailVerification bool     `json:"user_require_email_verification"`
	UserEnableSessionTimeout     bool     `json:"user_enable_session_timeout"`
	UserSessionTimeout           int      `json:"user_session_timeout"`
	EnableWhitelist              bool     `json:"enable_whitelist"`
	IPWhitelist                  []string `json:"ip_whitelist"`
}

// GetSystemConfig 获取系统配置
func (s *ConfigService) GetSystemConfig() (*SystemConfig, error) {
	// 检查 repo 是否已初始化
	if s.repo == nil {
		globalCfg := config.GlobalConfig.ServerConfig
		return &SystemConfig{
			SystemTitle:                  getStringOrDefault(globalCfg.SystemTitle, "卡密购买系统"),
			AdminSuffix:                  getStringOrDefault(globalCfg.AdminSuffix, "manage"),
			EnableLogin:                  globalCfg.EnableLogin,
			EnableCaptcha:                globalCfg.EnableCaptcha,
			AdminUsername:                getStringOrDefault(globalCfg.AdminUsername, "admin"),
			AdminPassword:                getStringOrDefault(globalCfg.AdminPassword, "admin123"),
			AdminPasswordInitialized:     globalCfg.AdminPasswordInitialized,
			Enable2FA:                    globalCfg.Enable2FA,
			TOTPSecret:                   globalCfg.TOTPSecret,
			EnableSessionTimeout:         globalCfg.EnableSessionTimeout,
			SessionTimeout:               normalizeTimeoutMinutes(globalCfg.SessionTimeout, 60),
			UserAllowRegister:            globalCfg.UserAllowRegister,
			UserEnableCaptcha:            globalCfg.UserEnableCaptcha,
			UserEnable2FA:                globalCfg.UserEnable2FA,
			UserRequireEmailVerification: globalCfg.UserRequireEmailVerification,
			UserEnableSessionTimeout:     globalCfg.UserEnableSessionTimeout,
			UserSessionTimeout:           normalizeTimeoutMinutes(globalCfg.UserSessionTimeout, 120),
			EnableWhitelist:              false,
			IPWhitelist:                  []string{},
		}, nil
	}

	dbConfig, err := s.repo.GetSystemConfig()
	if err != nil {
		globalCfg := config.GlobalConfig.ServerConfig
		return &SystemConfig{
			SystemTitle:                  getStringOrDefault(globalCfg.SystemTitle, "卡密购买系统"),
			AdminSuffix:                  getStringOrDefault(globalCfg.AdminSuffix, "manage"),
			EnableLogin:                  globalCfg.EnableLogin,
			EnableCaptcha:                globalCfg.EnableCaptcha,
			AdminUsername:                getStringOrDefault(globalCfg.AdminUsername, "admin"),
			AdminPassword:                getStringOrDefault(globalCfg.AdminPassword, "admin123"),
			AdminPasswordInitialized:     globalCfg.AdminPasswordInitialized,
			Enable2FA:                    globalCfg.Enable2FA,
			TOTPSecret:                   globalCfg.TOTPSecret,
			EnableSessionTimeout:         globalCfg.EnableSessionTimeout,
			SessionTimeout:               normalizeTimeoutMinutes(globalCfg.SessionTimeout, 60),
			UserAllowRegister:            globalCfg.UserAllowRegister,
			UserEnableCaptcha:            globalCfg.UserEnableCaptcha,
			UserEnable2FA:                globalCfg.UserEnable2FA,
			UserRequireEmailVerification: globalCfg.UserRequireEmailVerification,
			UserEnableSessionTimeout:     globalCfg.UserEnableSessionTimeout,
			UserSessionTimeout:           normalizeTimeoutMinutes(globalCfg.UserSessionTimeout, 120),
			EnableWhitelist:              false,
			IPWhitelist:                  []string{},
		}, nil
	}

	// 解析IP白名单JSON
	var ipWhitelist []string
	if dbConfig.IPWhitelist != "" {
		json.Unmarshal([]byte(dbConfig.IPWhitelist), &ipWhitelist)
	}

	return &SystemConfig{
		SystemTitle:                  dbConfig.SystemTitle,
		AdminSuffix:                  dbConfig.AdminSuffix,
		EnableLogin:                  dbConfig.EnableLogin,
		EnableCaptcha:                dbConfig.EnableCaptcha,
		AdminUsername:                dbConfig.AdminUsername,
		AdminPassword:                dbConfig.AdminPassword,
		AdminPasswordInitialized:     dbConfig.AdminPasswordInitialized,
		Enable2FA:                    dbConfig.Enable2FA,
		TOTPSecret:                   dbConfig.TOTPSecret,
		EnableSessionTimeout:         dbConfig.EnableSessionTimeout,
		SessionTimeout:               dbConfig.SessionTimeout,
		UserAllowRegister:            dbConfig.UserAllowRegister,
		UserEnableCaptcha:            dbConfig.UserEnableCaptcha,
		UserEnable2FA:                dbConfig.UserEnable2FA,
		UserRequireEmailVerification: dbConfig.UserRequireEmailVerification,
		UserEnableSessionTimeout:     dbConfig.UserEnableSessionTimeout,
		UserSessionTimeout:           dbConfig.UserSessionTimeout,
		EnableWhitelist:              dbConfig.EnableWhitelist,
		IPWhitelist:                  ipWhitelist,
	}, nil
}

// SaveSystemConfig 保存系统配置
func (s *ConfigService) SaveSystemConfig(cfg *SystemConfig) error {
	if s.repo == nil {
		return fmt.Errorf("数据库未连接")
	}

	// 序列化IP白名单为JSON
	ipWhitelistJSON := "[]"
	if len(cfg.IPWhitelist) > 0 {
		if data, err := json.Marshal(cfg.IPWhitelist); err == nil {
			ipWhitelistJSON = string(data)
		}
	}

	dbConfig := &model.SystemConfigDB{
		SystemTitle:                  cfg.SystemTitle,
		AdminSuffix:                  cfg.AdminSuffix,
		EnableLogin:                  cfg.EnableLogin,
		EnableCaptcha:                cfg.EnableCaptcha,
		AdminUsername:                cfg.AdminUsername,
		AdminPassword:                cfg.AdminPassword,
		AdminPasswordInitialized:     cfg.AdminPasswordInitialized,
		Enable2FA:                    cfg.Enable2FA,
		TOTPSecret:                   cfg.TOTPSecret,
		EnableSessionTimeout:         cfg.EnableSessionTimeout,
		SessionTimeout:               normalizeTimeoutMinutes(cfg.SessionTimeout, 60),
		UserAllowRegister:            cfg.UserAllowRegister,
		UserEnableCaptcha:            cfg.UserEnableCaptcha,
		UserEnable2FA:                cfg.UserEnable2FA,
		UserRequireEmailVerification: cfg.UserRequireEmailVerification,
		UserEnableSessionTimeout:     cfg.UserEnableSessionTimeout,
		UserSessionTimeout:           normalizeTimeoutMinutes(cfg.UserSessionTimeout, 120),
		EnableWhitelist:              cfg.EnableWhitelist,
		IPWhitelist:                  ipWhitelistJSON,
	}

	err := s.repo.SaveSystemConfig(dbConfig)
	if err != nil {
		return err
	}

	// 同步更新内存中的 GlobalConfig
	if config.GlobalConfig != nil {
		config.GlobalConfig.SetServerConfig(config.ServerConfig{
			Port:                         config.GlobalConfig.ServerConfig.Port,
			UseHTTPS:                     config.GlobalConfig.ServerConfig.UseHTTPS,
			CertFile:                     config.GlobalConfig.ServerConfig.CertFile,
			KeyFile:                      config.GlobalConfig.ServerConfig.KeyFile,
			AdminUsername:                cfg.AdminUsername,
			AdminPassword:                cfg.AdminPassword,
			AdminPasswordInitialized:     cfg.AdminPasswordInitialized,
			AdminSuffix:                  cfg.AdminSuffix,
			SystemTitle:                  cfg.SystemTitle,
			EnableLogin:                  cfg.EnableLogin,
			EnableCaptcha:                cfg.EnableCaptcha,
			Enable2FA:                    cfg.Enable2FA,
			TOTPSecret:                   cfg.TOTPSecret,
			EnableSessionTimeout:         cfg.EnableSessionTimeout,
			SessionTimeout:               normalizeTimeoutMinutes(cfg.SessionTimeout, 60),
			UserAllowRegister:            cfg.UserAllowRegister,
			UserEnableCaptcha:            cfg.UserEnableCaptcha,
			UserEnable2FA:                cfg.UserEnable2FA,
			UserRequireEmailVerification: cfg.UserRequireEmailVerification,
			UserEnableSessionTimeout:     cfg.UserEnableSessionTimeout,
			UserSessionTimeout:           normalizeTimeoutMinutes(cfg.UserSessionTimeout, 120),
		})
	}

	return nil
}

// UpdateSystemTitle 更新系统标题
func (s *ConfigService) UpdateSystemTitle(title string) error {
	cfg, err := s.GetSystemConfig()
	if err != nil {
		return err
	}
	cfg.SystemTitle = title
	return s.SaveSystemConfig(cfg)
}

// UpdateAdminSuffix 更新管理后台路径后缀
func (s *ConfigService) UpdateAdminSuffix(suffix string) error {
	cfg, err := s.GetSystemConfig()
	if err != nil {
		return err
	}
	cfg.AdminSuffix = suffix
	return s.SaveSystemConfig(cfg)
}

// UpdateSecuritySettings 更新安全设置
func (s *ConfigService) UpdateSecuritySettings(enableLogin bool, adminUsername, adminPassword string, enable2FA bool, totpSecret string) error {
	cfg, err := s.GetSystemConfig()
	if err != nil {
		return err
	}
	cfg.EnableLogin = enableLogin
	if adminUsername != "" {
		cfg.AdminUsername = adminUsername
	}
	if adminPassword != "" {
		cfg.AdminPassword = adminPassword
		cfg.AdminPasswordInitialized = true
	}
	cfg.Enable2FA = enable2FA
	cfg.TOTPSecret = totpSecret
	return s.SaveSystemConfig(cfg)
}

// GetWhitelistConfig 获取白名单配置
func (s *ConfigService) GetWhitelistConfig() (bool, []string, error) {
	cfg, err := s.GetSystemConfig()
	if err != nil {
		return false, nil, err
	}
	return cfg.EnableWhitelist, cfg.IPWhitelist, nil
}

// UpdateWhitelistConfig 更新白名单配置
func (s *ConfigService) UpdateWhitelistConfig(enabled bool, whitelist []string) error {
	cfg, err := s.GetSystemConfig()
	if err != nil {
		return err
	}
	cfg.EnableWhitelist = enabled
	cfg.IPWhitelist = whitelist
	return s.SaveSystemConfig(cfg)
}

// IsIPInWhitelist 检查IP是否在白名单中
func (s *ConfigService) IsIPInWhitelist(ip string) bool {
	cfg, err := s.GetSystemConfig()
	if err != nil || !cfg.EnableWhitelist {
		return true // 白名单未启用时，所有IP都允许
	}

	for _, whiteIP := range cfg.IPWhitelist {
		if whiteIP == ip {
			return true
		}
	}
	return false
}

func normalizeSystemConfig(cfg *SystemConfig) *SystemConfig {
	if cfg == nil {
		cfg = &SystemConfig{}
	}
	if cfg.SystemTitle == "" {
		cfg.SystemTitle = "卡密购买系统"
	}
	if cfg.AdminSuffix == "" {
		cfg.AdminSuffix = "manage"
	}
	if cfg.AdminUsername == "" {
		cfg.AdminUsername = "admin"
	}
	cfg.SessionTimeout = normalizeTimeoutMinutes(cfg.SessionTimeout, 60)
	cfg.UserSessionTimeout = normalizeTimeoutMinutes(cfg.UserSessionTimeout, 120)
	if cfg.IPWhitelist == nil {
		cfg.IPWhitelist = []string{}
	}
	return cfg
}

// NeedsInitialSetup 检查是否需要初始化设置（首次启动）
// 返回 true 表示需要设置初始密码
func (s *ConfigService) NeedsInitialSetup() bool {
	// 首先检查数据库中是否有系统配置记录
	if s.repo != nil {
		dbConfig, err := s.repo.GetSystemConfig()
		if err != nil {
			// 数据库没有配置记录，需要初始化
			return true
		}
		return !dbConfig.AdminPasswordInitialized
	}

	// repo 未初始化，检查全局配置
	cfg := config.GlobalConfig.ServerConfig
	return !cfg.AdminPasswordInitialized
}

// SetInitialPassword 设置初始管理员密码
// 只有在初始密码尚未完成用户设置时才允许设置。
func (s *ConfigService) SetInitialPassword(newPassword string) error {
	if !s.NeedsInitialSetup() {
		return fmt.Errorf("初始密码已设置，无法重复设置")
	}

	if len(newPassword) < 6 {
		return fmt.Errorf("密码长度至少6位")
	}

	cfg, err := s.GetSystemConfig()
	if err != nil {
		// 创建新配置
		cfg = &SystemConfig{
			SystemTitle:              "卡密购买系统",
			AdminSuffix:              "manage",
			EnableLogin:              true,
			EnableCaptcha:            true,
			AdminUsername:            "admin",
			AdminPassword:            newPassword,
			AdminPasswordInitialized: true,
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
	} else {
		cfg = normalizeSystemConfig(cfg)
		cfg.AdminPassword = newPassword
		cfg.AdminPasswordInitialized = true
	}

	if err := s.SaveSystemConfig(cfg); err != nil {
		return err
	}

	if s.NeedsInitialSetup() {
		return fmt.Errorf("初始密码已保存，但初始化状态未生效")
	}

	return nil
}
