// Package service 提供业务逻辑服务
// config_system.go - 系统配置管理方法
package service

import (
	"encoding/json"
	"fmt"

	"user-frontend/internal/config"
	"user-frontend/internal/model"
)

const (
	DefaultHumanVerificationProviderID = "keyswift.image_captcha"
	humanVerificationConfiguredKey     = "human_verification_configured"
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

// SystemConfig 系统配置结构。
type SystemConfig struct {
	SystemTitle                              string   `json:"system_title"`
	AdminSuffix                              string   `json:"admin_suffix"`
	EnableLogin                              bool     `json:"enable_login"`
	AdminHumanVerificationEnabled            bool     `json:"admin_human_verification_enabled"`
	AdminHumanVerificationProviderID         string   `json:"admin_human_verification_provider_id"`
	AdminUsername                            string   `json:"admin_username"`
	AdminPassword                            string   `json:"admin_password"`
	AdminPasswordInitialized                 bool     `json:"admin_password_initialized"`
	Enable2FA                                bool     `json:"enable_2fa"`
	TOTPSecret                               string   `json:"totp_secret"`
	EnableSessionTimeout                     bool     `json:"enable_session_timeout"`
	SessionTimeout                           int      `json:"session_timeout"`
	UserAllowRegister                        bool     `json:"user_allow_register"`
	UserLoginHumanVerificationEnabled        bool     `json:"user_login_human_verification_enabled"`
	UserLoginHumanVerificationProviderID     string   `json:"user_login_human_verification_provider_id"`
	UserRegisterHumanVerificationEnabled     bool     `json:"user_register_human_verification_enabled"`
	UserRegisterHumanVerificationProviderID  string   `json:"user_register_human_verification_provider_id"`
	UserRegisterHumanVerificationFollowLogin bool     `json:"user_register_human_verification_follow_login"`
	UserEnable2FA                            bool     `json:"user_enable_2fa"`
	UserRequireEmailVerification             bool     `json:"user_require_email_verification"`
	UserEnableSessionTimeout                 bool     `json:"user_enable_session_timeout"`
	UserSessionTimeout                       int      `json:"user_session_timeout"`
	EnableWhitelist                          bool     `json:"enable_whitelist"`
	IPWhitelist                              []string `json:"ip_whitelist"`
}

func defaultProviderID(value string) string {
	if value != "" {
		return value
	}
	return DefaultHumanVerificationProviderID
}

func systemConfigFromGlobal(globalCfg config.ServerConfig) *SystemConfig {
	return &SystemConfig{
		SystemTitle:                              getStringOrDefault(globalCfg.SystemTitle, "卡密购买系统"),
		AdminSuffix:                              getStringOrDefault(globalCfg.AdminSuffix, "manage"),
		EnableLogin:                              globalCfg.EnableLogin,
		AdminHumanVerificationEnabled:            globalCfg.AdminHumanVerificationEnabled,
		AdminHumanVerificationProviderID:         defaultProviderID(globalCfg.AdminHumanVerificationProviderID),
		AdminUsername:                            getStringOrDefault(globalCfg.AdminUsername, "admin"),
		AdminPassword:                            getStringOrDefault(globalCfg.AdminPassword, "admin123"),
		AdminPasswordInitialized:                 globalCfg.AdminPasswordInitialized,
		Enable2FA:                                globalCfg.Enable2FA,
		TOTPSecret:                               globalCfg.TOTPSecret,
		EnableSessionTimeout:                     globalCfg.EnableSessionTimeout,
		SessionTimeout:                           normalizeTimeoutMinutes(globalCfg.SessionTimeout, 60),
		UserAllowRegister:                        globalCfg.UserAllowRegister,
		UserLoginHumanVerificationEnabled:        globalCfg.UserLoginHumanVerificationEnabled,
		UserLoginHumanVerificationProviderID:     defaultProviderID(globalCfg.UserLoginHumanVerificationProviderID),
		UserRegisterHumanVerificationEnabled:     globalCfg.UserRegisterHumanVerificationEnabled,
		UserRegisterHumanVerificationProviderID:  defaultProviderID(globalCfg.UserRegisterHumanVerificationProviderID),
		UserRegisterHumanVerificationFollowLogin: globalCfg.UserRegisterHumanVerificationFollowLogin,
		UserEnable2FA:                            globalCfg.UserEnable2FA,
		UserRequireEmailVerification:             globalCfg.UserRequireEmailVerification,
		UserEnableSessionTimeout:                 globalCfg.UserEnableSessionTimeout,
		UserSessionTimeout:                       normalizeTimeoutMinutes(globalCfg.UserSessionTimeout, 120),
		EnableWhitelist:                          false,
		IPWhitelist:                              []string{},
	}
}

// GetSystemConfig 获取系统配置。
func (s *ConfigService) GetSystemConfig() (*SystemConfig, error) {
	if s.repo == nil {
		return systemConfigFromGlobal(config.GlobalConfig.ServerConfig), nil
	}

	dbConfig, err := s.repo.GetSystemConfig()
	if err != nil {
		return systemConfigFromGlobal(config.GlobalConfig.ServerConfig), nil
	}

	var ipWhitelist []string
	if dbConfig.IPWhitelist != "" {
		_ = json.Unmarshal([]byte(dbConfig.IPWhitelist), &ipWhitelist)
	}

	cfg := normalizeSystemConfig(&SystemConfig{
		SystemTitle:                              dbConfig.SystemTitle,
		AdminSuffix:                              dbConfig.AdminSuffix,
		EnableLogin:                              dbConfig.EnableLogin,
		AdminHumanVerificationEnabled:            dbConfig.AdminHumanVerificationEnabled,
		AdminHumanVerificationProviderID:         dbConfig.AdminHumanVerificationProviderID,
		AdminUsername:                            dbConfig.AdminUsername,
		AdminPassword:                            dbConfig.AdminPassword,
		AdminPasswordInitialized:                 dbConfig.AdminPasswordInitialized,
		Enable2FA:                                dbConfig.Enable2FA,
		TOTPSecret:                               dbConfig.TOTPSecret,
		EnableSessionTimeout:                     dbConfig.EnableSessionTimeout,
		SessionTimeout:                           dbConfig.SessionTimeout,
		UserAllowRegister:                        dbConfig.UserAllowRegister,
		UserLoginHumanVerificationEnabled:        dbConfig.UserLoginHumanVerificationEnabled,
		UserLoginHumanVerificationProviderID:     dbConfig.UserLoginHumanVerificationProviderID,
		UserRegisterHumanVerificationEnabled:     dbConfig.UserRegisterHumanVerificationEnabled,
		UserRegisterHumanVerificationProviderID:  dbConfig.UserRegisterHumanVerificationProviderID,
		UserRegisterHumanVerificationFollowLogin: dbConfig.UserRegisterHumanVerificationFollowLogin,
		UserEnable2FA:                            dbConfig.UserEnable2FA,
		UserRequireEmailVerification:             dbConfig.UserRequireEmailVerification,
		UserEnableSessionTimeout:                 dbConfig.UserEnableSessionTimeout,
		UserSessionTimeout:                       dbConfig.UserSessionTimeout,
		EnableWhitelist:                          dbConfig.EnableWhitelist,
		IPWhitelist:                              ipWhitelist,
	})
	if !s.HumanVerificationConfigured() {
		disableHumanVerification(cfg)
	}
	return cfg, nil
}

// SaveSystemConfig 保存系统配置。
func (s *ConfigService) SaveSystemConfig(cfg *SystemConfig) error {
	if s.repo == nil {
		return fmt.Errorf("数据库未连接")
	}
	cfg = normalizeSystemConfig(cfg)

	ipWhitelistJSON := "[]"
	if len(cfg.IPWhitelist) > 0 {
		if data, err := json.Marshal(cfg.IPWhitelist); err == nil {
			ipWhitelistJSON = string(data)
		}
	}

	dbConfig := &model.SystemConfigDB{
		SystemTitle:                              cfg.SystemTitle,
		AdminSuffix:                              cfg.AdminSuffix,
		EnableLogin:                              cfg.EnableLogin,
		AdminHumanVerificationEnabled:            cfg.AdminHumanVerificationEnabled,
		AdminHumanVerificationProviderID:         cfg.AdminHumanVerificationProviderID,
		AdminUsername:                            cfg.AdminUsername,
		AdminPassword:                            cfg.AdminPassword,
		AdminPasswordInitialized:                 cfg.AdminPasswordInitialized,
		Enable2FA:                                cfg.Enable2FA,
		TOTPSecret:                               cfg.TOTPSecret,
		EnableSessionTimeout:                     cfg.EnableSessionTimeout,
		SessionTimeout:                           normalizeTimeoutMinutes(cfg.SessionTimeout, 60),
		UserAllowRegister:                        cfg.UserAllowRegister,
		UserLoginHumanVerificationEnabled:        cfg.UserLoginHumanVerificationEnabled,
		UserLoginHumanVerificationProviderID:     cfg.UserLoginHumanVerificationProviderID,
		UserRegisterHumanVerificationEnabled:     cfg.UserRegisterHumanVerificationEnabled,
		UserRegisterHumanVerificationProviderID:  cfg.UserRegisterHumanVerificationProviderID,
		UserRegisterHumanVerificationFollowLogin: cfg.UserRegisterHumanVerificationFollowLogin,
		UserEnable2FA:                            cfg.UserEnable2FA,
		UserRequireEmailVerification:             cfg.UserRequireEmailVerification,
		UserEnableSessionTimeout:                 cfg.UserEnableSessionTimeout,
		UserSessionTimeout:                       normalizeTimeoutMinutes(cfg.UserSessionTimeout, 120),
		EnableWhitelist:                          cfg.EnableWhitelist,
		IPWhitelist:                              ipWhitelistJSON,
	}
	if err := s.repo.SaveSystemConfig(dbConfig); err != nil {
		return err
	}
	if config.GlobalConfig != nil {
		config.GlobalConfig.SetServerConfig(serverConfigFromSystemConfig(cfg))
	}
	return nil
}

func serverConfigFromSystemConfig(cfg *SystemConfig) config.ServerConfig {
	current := config.GlobalConfig.ServerConfig
	return config.ServerConfig{
		Port:                                     current.Port,
		UseHTTPS:                                 current.UseHTTPS,
		CertFile:                                 current.CertFile,
		KeyFile:                                  current.KeyFile,
		AdminUsername:                            cfg.AdminUsername,
		AdminPassword:                            cfg.AdminPassword,
		AdminPasswordInitialized:                 cfg.AdminPasswordInitialized,
		AdminSuffix:                              cfg.AdminSuffix,
		SystemTitle:                              cfg.SystemTitle,
		EnableLogin:                              cfg.EnableLogin,
		AdminHumanVerificationEnabled:            cfg.AdminHumanVerificationEnabled,
		AdminHumanVerificationProviderID:         cfg.AdminHumanVerificationProviderID,
		Enable2FA:                                cfg.Enable2FA,
		TOTPSecret:                               cfg.TOTPSecret,
		EnableSessionTimeout:                     cfg.EnableSessionTimeout,
		SessionTimeout:                           normalizeTimeoutMinutes(cfg.SessionTimeout, 60),
		UserAllowRegister:                        cfg.UserAllowRegister,
		UserLoginHumanVerificationEnabled:        cfg.UserLoginHumanVerificationEnabled,
		UserLoginHumanVerificationProviderID:     cfg.UserLoginHumanVerificationProviderID,
		UserRegisterHumanVerificationEnabled:     cfg.UserRegisterHumanVerificationEnabled,
		UserRegisterHumanVerificationProviderID:  cfg.UserRegisterHumanVerificationProviderID,
		UserRegisterHumanVerificationFollowLogin: cfg.UserRegisterHumanVerificationFollowLogin,
		UserEnable2FA:                            cfg.UserEnable2FA,
		UserRequireEmailVerification:             cfg.UserRequireEmailVerification,
		UserEnableSessionTimeout:                 cfg.UserEnableSessionTimeout,
		UserSessionTimeout:                       normalizeTimeoutMinutes(cfg.UserSessionTimeout, 120),
	}
}

// UpdateSystemTitle 更新系统标题。
func (s *ConfigService) UpdateSystemTitle(title string) error {
	cfg, err := s.GetSystemConfig()
	if err != nil {
		return err
	}
	cfg.SystemTitle = title
	return s.SaveSystemConfig(cfg)
}

// UpdateAdminSuffix 更新管理后台路径后缀。
func (s *ConfigService) UpdateAdminSuffix(suffix string) error {
	cfg, err := s.GetSystemConfig()
	if err != nil {
		return err
	}
	cfg.AdminSuffix = suffix
	return s.SaveSystemConfig(cfg)
}

// UpdateSecuritySettings 更新安全设置。
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

// GetWhitelistConfig 获取白名单配置。
func (s *ConfigService) GetWhitelistConfig() (bool, []string, error) {
	cfg, err := s.GetSystemConfig()
	if err != nil {
		return false, nil, err
	}
	return cfg.EnableWhitelist, cfg.IPWhitelist, nil
}

// UpdateWhitelistConfig 更新白名单配置。
func (s *ConfigService) UpdateWhitelistConfig(enabled bool, whitelist []string) error {
	cfg, err := s.GetSystemConfig()
	if err != nil {
		return err
	}
	cfg.EnableWhitelist = enabled
	cfg.IPWhitelist = whitelist
	return s.SaveSystemConfig(cfg)
}

// IsIPInWhitelist 检查 IP 是否在白名单中。
func (s *ConfigService) IsIPInWhitelist(ip string) bool {
	cfg, err := s.GetSystemConfig()
	if err != nil || !cfg.EnableWhitelist {
		return true
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
	cfg.AdminHumanVerificationProviderID = defaultProviderID(cfg.AdminHumanVerificationProviderID)
	cfg.UserLoginHumanVerificationProviderID = defaultProviderID(cfg.UserLoginHumanVerificationProviderID)
	if cfg.UserRegisterHumanVerificationFollowLogin {
		cfg.UserRegisterHumanVerificationProviderID = cfg.UserLoginHumanVerificationProviderID
	} else {
		cfg.UserRegisterHumanVerificationProviderID = defaultProviderID(cfg.UserRegisterHumanVerificationProviderID)
	}
	cfg.SessionTimeout = normalizeTimeoutMinutes(cfg.SessionTimeout, 60)
	cfg.UserSessionTimeout = normalizeTimeoutMinutes(cfg.UserSessionTimeout, 120)
	if cfg.IPWhitelist == nil {
		cfg.IPWhitelist = []string{}
	}
	return cfg
}

func disableHumanVerification(cfg *SystemConfig) {
	cfg.AdminHumanVerificationEnabled = false
	cfg.UserLoginHumanVerificationEnabled = false
	cfg.UserRegisterHumanVerificationEnabled = false
}

// HumanVerificationConfigured 判断管理员是否已经显式保存过人机验证策略。
func (s *ConfigService) HumanVerificationConfigured() bool {
	if s == nil || s.repo == nil {
		return false
	}
	value, err := s.repo.GetSetting(humanVerificationConfiguredKey)
	return err == nil && value == "true"
}

// MarkHumanVerificationConfigured 标记人机验证策略已由后台显式保存。
func (s *ConfigService) MarkHumanVerificationConfigured() error {
	if s == nil || s.repo == nil {
		return fmt.Errorf("数据库未连接")
	}
	return s.repo.SetSetting(humanVerificationConfiguredKey, "true", "后台登录配置已显式保存人机验证策略")
}

// NeedsInitialSetup 检查是否需要初始化设置。
func (s *ConfigService) NeedsInitialSetup() bool {
	if s.repo != nil {
		dbConfig, err := s.repo.GetSystemConfig()
		if err != nil {
			return true
		}
		return !dbConfig.AdminPasswordInitialized
	}
	cfg := config.GlobalConfig.ServerConfig
	return !cfg.AdminPasswordInitialized
}

// SetInitialPassword 设置初始管理员密码。
func (s *ConfigService) SetInitialPassword(newPassword string) error {
	if !s.NeedsInitialSetup() {
		return fmt.Errorf("初始密码已设置，无法重复设置")
	}
	if len(newPassword) < 6 {
		return fmt.Errorf("密码长度至少6位")
	}

	cfg, err := s.GetSystemConfig()
	if err != nil {
		cfg = &SystemConfig{
			SystemTitle:                              "卡密购买系统",
			AdminSuffix:                              "manage",
			EnableLogin:                              true,
			AdminHumanVerificationEnabled:            false,
			AdminHumanVerificationProviderID:         DefaultHumanVerificationProviderID,
			AdminUsername:                            "admin",
			AdminPassword:                            newPassword,
			AdminPasswordInitialized:                 true,
			Enable2FA:                                false,
			EnableSessionTimeout:                     true,
			SessionTimeout:                           60,
			UserAllowRegister:                        true,
			UserLoginHumanVerificationEnabled:        false,
			UserLoginHumanVerificationProviderID:     DefaultHumanVerificationProviderID,
			UserRegisterHumanVerificationEnabled:     false,
			UserRegisterHumanVerificationProviderID:  DefaultHumanVerificationProviderID,
			UserRegisterHumanVerificationFollowLogin: true,
			UserEnable2FA:                            true,
			UserEnableSessionTimeout:                 true,
			UserSessionTimeout:                       120,
			EnableWhitelist:                          false,
			IPWhitelist:                              []string{},
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
