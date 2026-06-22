package model

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"user-frontend/internal/config"
	"user-frontend/internal/dbschema"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB
var DBConnected bool

func InitDB(cfg *config.DBConfig) error {
	var dialector gorm.Dialector
	shouldInitSQLite := false

	switch cfg.Type {
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			cfg.Host, cfg.User, cfg.Password, cfg.Database, cfg.Port)
		dialector = postgres.Open(dsn)
	case "sqlite":
		dir := filepath.Dir(cfg.Database)
		if dir != "." && dir != "" {
			os.MkdirAll(dir, 0755)
		}
		var err error
		shouldInitSQLite, err = dbschema.ShouldInitializeSQLite(cfg.Database)
		if err != nil {
			DBConnected = false
			return err
		}
		dialector = sqlite.Open(cfg.Database)
	case "mysql":
		fallthrough
	default:
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
		dialector = mysql.Open(dsn)
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Error,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	var err error
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		DBConnected = false
		return err
	}

	if cfg.Type == "sqlite" && shouldInitSQLite {
		if _, err := dbschema.ApplyEmbeddedMainSchema(DB, "runtime"); err != nil {
			DBConnected = false
			return err
		}
	}

	// 已存在主库只校验当前基线，不执行旧库兼容迁移。
	if err := dbschema.ValidateMainSchema(DB); err != nil {
		DBConnected = false
		if cfg.Type == "sqlite" {
			return fmt.Errorf("%w；请删除主业务数据库文件后重启: %s", err, cfg.Database)
		}
		return fmt.Errorf("%w；请使用当前基线重新初始化主业务数据库", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		DBConnected = false
		return err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DBConnected = true
	return nil
}

// TestConnection 测试数据库连接
func TestConnection(cfg *config.DBConfig) error {
	var dialector gorm.Dialector

	switch cfg.Type {
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
			cfg.Host, cfg.User, cfg.Password, cfg.Database, cfg.Port)
		dialector = postgres.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(cfg.Database)
	case "mysql":
		fallthrough
	default:
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
		dialector = mysql.Open(dsn)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	return sqlDB.Ping()
}
