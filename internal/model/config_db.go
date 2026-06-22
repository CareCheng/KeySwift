package model

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"user-frontend/internal/dbschema"

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
	shouldInit, err := dbschema.ShouldInitializeSQLite(dbPath)
	if err != nil {
		return err
	}

	ConfigDB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}

	if shouldInit {
		if _, err := dbschema.ApplyEmbeddedBootstrapSchema(ConfigDB, "runtime"); err != nil {
			return err
		}
	}

	// 已存在配置库只校验当前基线，不执行旧库兼容迁移。
	if err := dbschema.ValidateBootstrapSchema(ConfigDB); err != nil {
		return fmt.Errorf("%w；请删除配置数据库文件后重启: %s", err, dbPath)
	}
	return nil
}
