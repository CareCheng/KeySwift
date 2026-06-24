package api

import (
	"fmt"

	"user-frontend/internal/config"
	"user-frontend/internal/model"

	"github.com/gin-gonic/gin"
)

// ==================== 数据库配置相关 API ====================

// AdminGetDBConfig 获取数据库配置
func AdminGetDBConfig(c *gin.Context) {
	cfg := config.GlobalConfig

	// 获取加密密钥信息
	var encryptionKey string
	var keyLength int
	if DBConfigSvc != nil {
		encryptionKey, keyLength, _ = DBConfigSvc.GetEncryptionKeyInfo()
	}
	if keyLength == 0 {
		keyLength = 256
	}

	// 对于SQLite数据库，将完整路径转换为相对路径显示
	dbPath := cfg.DBConfig.Database
	if cfg.DBConfig.Type == "sqlite" {
		dbPath = toRelativePath(dbPath)
	}

	c.JSON(200, gin.H{
		"success": true,
		"config": gin.H{
			"type":           cfg.DBConfig.Type,
			"host":           cfg.DBConfig.Host,
			"port":           cfg.DBConfig.Port,
			"user":           cfg.DBConfig.User,
			"database":       dbPath,
			"connected":      model.DBConnected,
			"encryption_key": encryptionKey,
			"key_length":     keyLength,
		},
	})
}

// AdminSaveDBConfig 保存数据库配置
func AdminSaveDBConfig(c *gin.Context) {
	var req struct {
		Type     string `json:"type" binding:"required"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 检查数据库配置服务是否初始化
	if DBConfigSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "配置服务未初始化"})
		return
	}

	// 获取当前配置
	currentCfg, _ := DBConfigSvc.GetDBConfig()

	// 对于SQLite数据库，将相对路径转换为完整路径存储
	dbPath := req.Database
	if req.Type == "sqlite" {
		dbPath = toAbsolutePath(req.Database)
	}

	// 更新配置
	newCfg := &config.DBConfig{
		Type:     req.Type,
		Host:     req.Host,
		Port:     req.Port,
		User:     req.User,
		Password: req.Password,
		Database: dbPath,
	}

	// 如果没有提供新密码，保留原密码
	if req.Password == "" && currentCfg != nil {
		newCfg.Password = currentCfg.Password
	}

	// 保存到SQLite配置数据库
	if err := DBConfigSvc.SaveDBConfig(newCfg); err != nil {
		c.JSON(500, gin.H{"success": false, "error": "保存配置失败: " + err.Error()})
		return
	}

	// 更新全局配置
	config.GlobalConfig.SetDBConfig(*newCfg)

	c.JSON(200, gin.H{"success": true, "message": "数据库配置已保存，请重启程序生效"})
}

// AdminTestDBConnection 测试数据库连接
func AdminTestDBConnection(c *gin.Context) {
	var req struct {
		Type     string `json:"type" binding:"required"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Database string `json:"database" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	testCfg := &config.DBConfig{
		Type:     req.Type,
		Host:     req.Host,
		Port:     req.Port,
		User:     req.User,
		Password: req.Password,
		Database: req.Database,
	}

	if err := model.TestConnection(testCfg); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "连接失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "连接成功"})
}

// AdminResetEncryptionKey 重置加密密钥（危险操作）
func AdminResetEncryptionKey(c *gin.Context) {
	var req struct {
		KeyLength   int    `json:"key_length"`
		Confirm     string `json:"confirm" binding:"required"`
		ConfirmText string `json:"confirm_text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 二级确认验证
	if req.Confirm != "RESET_KEY" || req.ConfirmText != "我确认重置密钥并了解数据将永久丢失" {
		c.JSON(400, gin.H{"success": false, "error": "确认信息不正确，请仔细阅读警告信息"})
		return
	}

	if DBConfigSvc == nil {
		c.JSON(500, gin.H{"success": false, "error": "配置服务未初始化"})
		return
	}

	// 验证密钥长度
	keyLength := req.KeyLength
	if keyLength != 128 && keyLength != 192 && keyLength != 256 {
		keyLength = 256
	}

	// 重置密钥
	newKey, err := DBConfigSvc.ResetEncryptionKey(keyLength)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "error": "重置密钥失败: " + err.Error()})
		return
	}

	// 记录操作日志
	if LogSvc != nil {
		LogSvc.LogAdminActionSimple("system", "reset_encryption_key", "security", "",
			fmt.Sprintf("重置了%d位AES加密密钥", keyLength), GetClientIP(c), c.GetHeader("User-Agent"))
	}

	c.JSON(200, gin.H{
		"success":        true,
		"message":        "加密密钥已重置，之前使用旧密钥加密的数据将无法解密",
		"encryption_key": newKey,
		"key_length":     keyLength,
	})
}
