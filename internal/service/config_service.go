// Package service 提供业务逻辑服务
// config_service.go - 配置服务核心定义和邮箱配置
package service

import (
	"user-frontend/internal/config"
	"user-frontend/internal/model"
	"user-frontend/internal/repository"
	"user-frontend/internal/utils"

	"gorm.io/gorm"
)

// encryptPassword 加密密码
func encryptPassword(password string) (string, error) {
	return utils.AESEncrypt(password)
}

// decryptPassword 解密密码
func decryptPassword(encrypted string) (string, error) {
	return utils.AESDecrypt(encrypted)
}

// getStringOrDefault 获取字符串值，如果为空则返回默认值
func getStringOrDefault(value, defaultValue string) string {
	if value != "" {
		return value
	}
	return defaultValue
}

// ConfigService 配置服务（从数据库读写配置）
type ConfigService struct {
	repo      *repository.Repository
	configDB  *gorm.DB // SQLite配置数据库
	mainDB    *gorm.DB // 主数据库
	configDir string   // 配置目录路径
}

// NewConfigService 创建配置服务
func NewConfigService(repo *repository.Repository) *ConfigService {
	return &ConfigService{repo: repo}
}

// InitConfigService 初始化配置服务（使用配置数据库）
func InitConfigService(configDB *gorm.DB) *ConfigService {
	return &ConfigService{configDB: configDB}
}

// InitConfigServiceWithDir 初始化配置服务（使用配置数据库和配置目录）
func InitConfigServiceWithDir(configDB *gorm.DB, configDir string) *ConfigService {
	return &ConfigService{configDB: configDB, configDir: configDir}
}

// SetConfigDir 设置配置目录
func (s *ConfigService) SetConfigDir(configDir string) {
	s.configDir = configDir
}

// getDefaultDBPath 获取默认数据库路径
func (s *ConfigService) getDefaultDBPath() string {
	if s.configDir != "" {
		return s.configDir + "/user_data.db"
	}
	return "user_config/user_data.db"
}

// SetMainDB 设置主数据库（在主数据库初始化后调用）
func (s *ConfigService) SetMainDB(mainDB *gorm.DB) {
	s.mainDB = mainDB
}

// SetRepo 设置仓库（在主数据库初始化后调用）
func (s *ConfigService) SetRepo(repo *repository.Repository) {
	s.repo = repo
}

// ==================== 邮箱配置 ====================

// GetEmailConfig 获取邮箱配置
func (s *ConfigService) GetEmailConfig() (*config.EmailConfig, error) {
	dbConfig, err := s.repo.GetEmailConfig()
	if err != nil {
		// 数据库中没有配置，返回默认值
		return &config.EmailConfig{
			SMTPPort:   465,
			Encryption: "ssl",
			CodeLength: 6,
		}, nil
	}

	return &config.EmailConfig{
		Enabled:      dbConfig.Enabled,
		SMTPHost:     dbConfig.SMTPHost,
		SMTPPort:     dbConfig.SMTPPort,
		SMTPUser:     dbConfig.SMTPUser,
		SMTPPassword: dbConfig.SMTPPassword,
		FromName:     dbConfig.FromName,
		FromEmail:    dbConfig.FromEmail,
		Encryption:   dbConfig.Encryption,
		CodeLength:   dbConfig.CodeLength,
	}, nil
}

// SaveEmailConfig 保存邮箱配置
func (s *ConfigService) SaveEmailConfig(cfg *config.EmailConfig) error {
	dbConfig := &model.EmailConfigDB{
		Enabled:      cfg.Enabled,
		SMTPHost:     cfg.SMTPHost,
		SMTPPort:     cfg.SMTPPort,
		SMTPUser:     cfg.SMTPUser,
		SMTPPassword: cfg.SMTPPassword,
		FromName:     cfg.FromName,
		FromEmail:    cfg.FromEmail,
		Encryption:   cfg.Encryption,
		CodeLength:   cfg.CodeLength,
	}
	return s.repo.SaveEmailConfig(dbConfig)
}
