// Package service 提供业务逻辑服务
// config_db_ops.go - 数据库配置管理方法
package service

import (
	"encoding/base64"
	"fmt"

	"user-frontend/internal/config"
	"user-frontend/internal/model"
	"user-frontend/internal/utils"
)

// GetDBConfig 获取数据库配置（从SQLite配置数据库）
func (s *ConfigService) GetDBConfig() (*config.DBConfig, error) {
	if s.configDB == nil {
		return nil, fmt.Errorf("配置数据库未初始化")
	}

	var dbConfig model.DBConfigDB
	err := s.configDB.First(&dbConfig).Error
	if err != nil {
		// 数据库中没有配置，返回默认值
		return &config.DBConfig{
			Type:     "sqlite",
			Host:     "localhost",
			Port:     3306,
			Database: s.getDefaultDBPath(),
		}, nil
	}

	// 解密密码
	password := dbConfig.Password
	if password != "" {
		if decrypted, err := decryptPassword(password); err == nil {
			password = decrypted
		}
	}

	return &config.DBConfig{
		Type:     dbConfig.Type,
		Host:     dbConfig.Host,
		Port:     dbConfig.Port,
		User:     dbConfig.User,
		Password: password,
		Database: dbConfig.Database,
	}, nil
}

// SaveDBConfig 保存数据库配置（到SQLite配置数据库）
func (s *ConfigService) SaveDBConfig(cfg *config.DBConfig) error {
	if s.configDB == nil {
		return fmt.Errorf("配置数据库未初始化")
	}

	// 加密密码
	encryptedPassword := cfg.Password
	if cfg.Password != "" {
		if encrypted, err := encryptPassword(cfg.Password); err == nil {
			encryptedPassword = encrypted
		}
	}

	dbConfig := &model.DBConfigDB{
		Type:     cfg.Type,
		Host:     cfg.Host,
		Port:     cfg.Port,
		User:     cfg.User,
		Password: encryptedPassword,
		Database: cfg.Database,
	}

	var existing model.DBConfigDB
	err := s.configDB.First(&existing).Error
	if err != nil {
		return s.configDB.Create(dbConfig).Error
	}
	dbConfig.ID = existing.ID
	dbConfig.ServerPort = existing.ServerPort       // 保留端口配置
	dbConfig.EncryptionKey = existing.EncryptionKey // 保留加密密钥
	dbConfig.KeyLength = existing.KeyLength         // 保留密钥长度
	return s.configDB.Save(dbConfig).Error
}

// GetEncryptionKeyInfo 获取加密密钥信息
func (s *ConfigService) GetEncryptionKeyInfo() (keyBase64 string, keyLength int, err error) {
	if s.configDB == nil {
		return "", 0, fmt.Errorf("配置数据库未初始化")
	}

	var dbConfig model.DBConfigDB
	err = s.configDB.First(&dbConfig).Error
	if err != nil {
		return "", 256, nil // 默认256位
	}

	keyLength = dbConfig.KeyLength
	if keyLength == 0 {
		keyLength = 256
	}

	return dbConfig.EncryptionKey, keyLength, nil
}

// InitEncryptionKey 初始化加密密钥（如果不存在则自动生成）
func (s *ConfigService) InitEncryptionKey() error {
	if s.configDB == nil {
		return fmt.Errorf("配置数据库未初始化")
	}

	var dbConfig model.DBConfigDB
	err := s.configDB.First(&dbConfig).Error
	if err != nil {
		// 配置不存在，创建默认配置并生成密钥
		keyBase64, err := utils.GenerateAESKey(256)
		if err != nil {
			return fmt.Errorf("生成密钥失败: %v", err)
		}

		dbConfig = model.DBConfigDB{
			Type:          "sqlite",
			Host:          "localhost",
			Port:          3306,
			Database:      s.getDefaultDBPath(),
			EncryptionKey: keyBase64,
			KeyLength:     256,
		}
		if err := s.configDB.Create(&dbConfig).Error; err != nil {
			return fmt.Errorf("保存配置失败: %v", err)
		}

		// 设置全局密钥
		key, _ := base64.StdEncoding.DecodeString(keyBase64)
		utils.SetConfigEncryptionKey(key)
		fmt.Println("已自动生成256位AES加密密钥")
		return nil
	}

	// 配置存在，检查是否有密钥
	if dbConfig.EncryptionKey == "" {
		keyLength := dbConfig.KeyLength
		if keyLength == 0 {
			keyLength = 256
		}

		keyBase64, err := utils.GenerateAESKey(keyLength)
		if err != nil {
			return fmt.Errorf("生成密钥失败: %v", err)
		}

		dbConfig.EncryptionKey = keyBase64
		dbConfig.KeyLength = keyLength
		if err := s.configDB.Save(&dbConfig).Error; err != nil {
			return fmt.Errorf("保存密钥失败: %v", err)
		}

		key, _ := base64.StdEncoding.DecodeString(keyBase64)
		utils.SetConfigEncryptionKey(key)
		fmt.Printf("已自动生成%d位AES加密密钥\n", keyLength)
		return nil
	}

	// 已有密钥，加载到全局
	key, err := base64.StdEncoding.DecodeString(dbConfig.EncryptionKey)
	if err != nil {
		return fmt.Errorf("解析密钥失败: %v", err)
	}
	utils.SetConfigEncryptionKey(key)
	return nil
}

// ResetEncryptionKey 重置加密密钥（危险操作！会导致已加密数据无法解密）
func (s *ConfigService) ResetEncryptionKey(keyLength int) (string, error) {
	if s.configDB == nil {
		return "", fmt.Errorf("配置数据库未初始化")
	}

	// 验证密钥长度
	if keyLength != 128 && keyLength != 192 && keyLength != 256 {
		keyLength = 256
	}

	// 生成新密钥
	keyBase64, err := utils.GenerateAESKey(keyLength)
	if err != nil {
		return "", fmt.Errorf("生成密钥失败: %v", err)
	}

	// 更新数据库
	var dbConfig model.DBConfigDB
	err = s.configDB.First(&dbConfig).Error
	if err != nil {
		dbConfig = model.DBConfigDB{
			Type:          "sqlite",
			Host:          "localhost",
			Port:          3306,
			Database:      s.getDefaultDBPath(),
			EncryptionKey: keyBase64,
			KeyLength:     keyLength,
		}
		if err := s.configDB.Create(&dbConfig).Error; err != nil {
			return "", fmt.Errorf("保存配置失败: %v", err)
		}
	} else {
		dbConfig.EncryptionKey = keyBase64
		dbConfig.KeyLength = keyLength
		if err := s.configDB.Save(&dbConfig).Error; err != nil {
			return "", fmt.Errorf("保存密钥失败: %v", err)
		}
	}

	// 更新全局密钥
	key, _ := base64.StdEncoding.DecodeString(keyBase64)
	utils.SetConfigEncryptionKey(key)

	return keyBase64, nil
}

// GetServerPort 获取服务器端口配置（从SQLite配置数据库）
func (s *ConfigService) GetServerPort() (int, error) {
	if s.configDB == nil {
		return 8080, fmt.Errorf("配置数据库未初始化")
	}

	var dbConfig model.DBConfigDB
	err := s.configDB.First(&dbConfig).Error
	if err != nil {
		return 8080, nil // 默认端口
	}

	if dbConfig.ServerPort <= 0 {
		return 8080, nil
	}
	return dbConfig.ServerPort, nil
}

// SaveServerPort 保存服务器端口配置（到SQLite配置数据库）
func (s *ConfigService) SaveServerPort(port int) error {
	if s.configDB == nil {
		return fmt.Errorf("配置数据库未初始化")
	}

	var existing model.DBConfigDB
	err := s.configDB.First(&existing).Error
	if err != nil {
		dbConfig := &model.DBConfigDB{
			Type:       "sqlite",
			Host:       "localhost",
			Port:       3306,
			Database:   s.getDefaultDBPath(),
			ServerPort: port,
		}
		return s.configDB.Create(dbConfig).Error
	}
	return s.configDB.Model(&existing).Update("server_port", port).Error
}

// LoadDBConfigToGlobal 从SQLite配置数据库加载数据库配置到全局配置
func (s *ConfigService) LoadDBConfigToGlobal() {
	s.ensureDBConfig()
	if dbCfg, err := s.GetDBConfig(); err == nil {
		config.GlobalConfig.SetDBConfig(*dbCfg)
	}
}

// ensureDBConfig 确保SQLite配置数据库中有数据库配置
func (s *ConfigService) ensureDBConfig() {
	if s.configDB == nil {
		return
	}

	var count int64
	s.configDB.Model(&model.DBConfigDB{}).Count(&count)
	if count == 0 {
		defaultConfig := &config.DBConfig{
			Type:     "sqlite",
			Host:     "localhost",
			Port:     3306,
			Database: s.getDefaultDBPath(),
		}
		if err := s.SaveDBConfig(defaultConfig); err != nil {
			fmt.Printf("创建默认数据库配置失败: %v\n", err)
		} else {
			fmt.Println("已创建默认数据库配置")
		}
	}
}
