package config

import (
	"os"
	"sync"
)

// Config 全局配置
type Config struct {
	DBConfig     DBConfig     `json:"db_config"`     // 数据库配置（从SQLite配置数据库加载）
	ServerConfig ServerConfig `json:"server_config"` // 服务器配置（运行时从主数据库加载）
	EmailConfig  EmailConfig  `json:"email_config"`  // 邮箱配置（运行时从主数据库加载）
	ConfigDir    string       `json:"-"`
	mu           sync.RWMutex
}

// EmailConfig 邮箱配置
type EmailConfig struct {
	Enabled      bool   `json:"enabled"`
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUser     string `json:"smtp_user"`
	SMTPPassword string `json:"smtp_password"`
	FromName     string `json:"from_name"`
	FromEmail    string `json:"from_email"`
	Encryption   string `json:"encryption"` // 加密方式：none/ssl/starttls
	CodeLength   int    `json:"code_length"`
}

// DBConfig 数据库配置（从SQLite配置数据库加载）
type DBConfig struct {
	Type     string `json:"type"` // mysql, postgres, sqlite
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// ServerConfig 服务器配置（运行时从主数据库加载）
type ServerConfig struct {
	Port                                     int    `json:"port"`
	UseHTTPS                                 bool   `json:"use_https"`
	CertFile                                 string `json:"cert_file"`
	KeyFile                                  string `json:"key_file"`
	AdminUsername                            string `json:"admin_username"`
	AdminPassword                            string `json:"admin_password"`
	AdminPasswordInitialized                 bool   `json:"admin_password_initialized"`
	AdminSuffix                              string `json:"admin_suffix"`
	EnableLogin                              bool   `json:"enable_login"`
	AdminHumanVerificationEnabled            bool   `json:"admin_human_verification_enabled"`
	AdminHumanVerificationProviderID         string `json:"admin_human_verification_provider_id"`
	Enable2FA                                bool   `json:"enable_2fa"`
	TOTPSecret                               string `json:"totp_secret"`
	EnableSessionTimeout                     bool   `json:"enable_session_timeout"`
	SessionTimeout                           int    `json:"session_timeout"`
	UserAllowRegister                        bool   `json:"user_allow_register"`
	UserLoginHumanVerificationEnabled        bool   `json:"user_login_human_verification_enabled"`
	UserLoginHumanVerificationProviderID     string `json:"user_login_human_verification_provider_id"`
	UserRegisterHumanVerificationEnabled     bool   `json:"user_register_human_verification_enabled"`
	UserRegisterHumanVerificationProviderID  string `json:"user_register_human_verification_provider_id"`
	UserRegisterHumanVerificationFollowLogin bool   `json:"user_register_human_verification_follow_login"`
	UserEnable2FA                            bool   `json:"user_enable_2fa"`
	UserRequireEmailVerification             bool   `json:"user_require_email_verification"`
	UserEnableSessionTimeout                 bool   `json:"user_enable_session_timeout"`
	UserSessionTimeout                       int    `json:"user_session_timeout"`
	SystemTitle                              string `json:"system_title"`
}

var (
	GlobalConfig *Config
	once         sync.Once
)

func InitConfig(configDir string) (*Config, error) {
	var err error
	once.Do(func() {
		GlobalConfig = &Config{
			ConfigDir: configDir,
		}
		os.MkdirAll(configDir, 0755)
		err = GlobalConfig.LoadAll()
	})
	return GlobalConfig, err
}

// LoadAll 加载所有配置（数据库配置从SQLite配置数据库加载，其他配置从主数据库加载）
func (c *Config) LoadAll() error {
	// 数据库配置将通过 ConfigService 从 SQLite 配置数据库加载
	// 这里只设置默认值
	c.DBConfig = DBConfig{Type: "sqlite", Database: "user_config/user_data.db", Port: 3306}

	// 设置默认的服务器配置（实际值从数据库加载）
	c.ServerConfig = ServerConfig{
		Port:                                     8080,
		AdminUsername:                            "admin",
		AdminPassword:                            "admin123",
		AdminPasswordInitialized:                 false,
		AdminSuffix:                              "manage",
		EnableLogin:                              true,
		AdminHumanVerificationEnabled:            false,
		// 默认人机验证 provider，须与 service.DefaultHumanVerificationProviderID 保持一致；
		// config 包不依赖 service 包，故此处保留字面量并以此注释标注同步约束。
		AdminHumanVerificationProviderID:         "keyswift.image_captcha",
		Enable2FA:                                false,
		EnableSessionTimeout:                     true,
		SessionTimeout:                           60,
		UserAllowRegister:                        true,
		UserLoginHumanVerificationEnabled:        false,
		UserLoginHumanVerificationProviderID:     "keyswift.image_captcha",
		UserRegisterHumanVerificationEnabled:     false,
		UserRegisterHumanVerificationProviderID:  "keyswift.image_captcha",
		UserRegisterHumanVerificationFollowLogin: true,
		UserEnable2FA:                            true,
		UserRequireEmailVerification:             false,
		UserEnableSessionTimeout:                 true,
		UserSessionTimeout:                       120,
		SystemTitle:                              "卡密购买系统",
	}

	// 设置默认的邮箱配置（实际值从数据库加载）
	c.EmailConfig = EmailConfig{
		SMTPPort:   465,
		Encryption: "ssl",
		CodeLength: 6,
	}

	return nil
}

// SetDBConfig 设置数据库配置（由 ConfigService 调用）
func (c *Config) SetDBConfig(cfg DBConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.DBConfig = cfg
}

// SetServerConfig 设置服务器配置（由 ConfigService 调用）
func (c *Config) SetServerConfig(cfg ServerConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ServerConfig = cfg
}
