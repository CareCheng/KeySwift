package model

import (
	"path/filepath"
	"testing"

	"user-frontend/internal/config"
	"user-frontend/internal/dbschema"

	"gorm.io/gorm"
)

func TestInitConfigDBCreatesCurrentBaselineAndValidatesRepeatStart(t *testing.T) {
	configDir := filepath.Join(t.TempDir(), "config")

	if err := InitConfigDB(configDir); err != nil {
		t.Fatalf("首次初始化配置库失败: %v", err)
	}
	assertSchemaRevisionCount(t, ConfigDB, dbschema.BootstrapSchemaKey, dbschema.BootstrapSchemaVersion)
	closeGormDB(t, ConfigDB)

	if err := InitConfigDB(configDir); err != nil {
		t.Fatalf("重复启动配置库校验失败: %v", err)
	}
	assertSchemaRevisionCount(t, ConfigDB, dbschema.BootstrapSchemaKey, dbschema.BootstrapSchemaVersion)
	closeGormDB(t, ConfigDB)
}

func TestInitDBCreatesCurrentBaselineAndValidatesRepeatStart(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "user_data.db")
	cfg := &config.DBConfig{
		Type:     "sqlite",
		Database: dbPath,
	}

	if err := InitDB(cfg); err != nil {
		t.Fatalf("首次初始化主业务库失败: %v", err)
	}
	assertSchemaRevisionCount(t, DB, dbschema.MainSchemaKey, dbschema.MainSchemaVersion)
	closeGormDB(t, DB)

	if err := InitDB(cfg); err != nil {
		t.Fatalf("重复启动主业务库校验失败: %v", err)
	}
	assertSchemaRevisionCount(t, DB, dbschema.MainSchemaKey, dbschema.MainSchemaVersion)
	closeGormDB(t, DB)
}

func assertSchemaRevisionCount(t *testing.T, db *gorm.DB, schemaKey, version string) {
	t.Helper()

	var count int64
	if err := db.Table(dbschema.SchemaRevisionTable).
		Where("schema_key = ? AND version = ? AND direction = ?", schemaKey, version, dbschema.SchemaDirectionBase).
		Count(&count).Error; err != nil {
		t.Fatalf("读取 schema 变更记录失败: %v", err)
	}
	if count != 1 {
		t.Fatalf("%s 当前基线应写入 1 条 schema 变更记录，实际 %d", schemaKey, count)
	}
}

func closeGormDB(t *testing.T, db *gorm.DB) {
	t.Helper()

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("读取底层数据库连接失败: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("关闭数据库连接失败: %v", err)
	}
}
