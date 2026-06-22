package model

import (
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConfigDB 配置数据库连接（SQLite）
var ConfigDB *gorm.DB

// DBConfigDB 数据库配置（存储在SQLite配置数据库中）
type DBConfigDB struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Type          string    `gorm:"type:varchar(50);default:sqlite" json:"type"`
	Host          string    `gorm:"type:varchar(200)" json:"host"`
	Port          int       `gorm:"default:3306" json:"port"`
	User          string    `gorm:"type:varchar(100)" json:"user"`
	Password      string    `gorm:"type:varchar(255)" json:"password"`
	Database      string    `gorm:"type:varchar(200)" json:"database"`
	ServerPort    int       `gorm:"default:8080" json:"server_port"`         // 服务器监听端口
	EncryptionKey string    `gorm:"type:varchar(100)" json:"encryption_key"` // AES加密密钥（Base64编码）
	KeyLength     int       `gorm:"default:256" json:"key_length"`           // 密钥长度：128/192/256位
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (DBConfigDB) TableName() string {
	return "db_configs"
}

// InitConfigDB 初始化配置数据库（SQLite）
func InitConfigDB(configDir string) error {
	// 确保配置目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// 配置数据库路径
	dbPath := filepath.Join(configDir, "db-config.db")

	var err error
	ConfigDB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}

	// 按当前模型创建或同步配置表结构。
	if err := ConfigDB.AutoMigrate(&DBConfigDB{}); err != nil {
		return err
	}

	return nil
}
