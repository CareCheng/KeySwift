package dbschema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestApplyEmbeddedMainSchemaRecordsRevisionAndValidates(t *testing.T) {
	db := openTestSQLite(t, filepath.Join(t.TempDir(), "main.db"))

	checksum, err := ApplyEmbeddedMainSchema(db, "test")
	if err != nil {
		t.Fatalf("初始化主业务库失败: %v", err)
	}
	if strings.TrimSpace(checksum) == "" {
		t.Fatal("初始化后应返回 schema checksum")
	}
	if err := ValidateMainSchema(db); err != nil {
		t.Fatalf("当前主业务库基线应通过校验: %v", err)
	}

	var count int64
	if err := db.Table(SchemaRevisionTable).
		Where("schema_key = ? AND version = ? AND direction = ?", MainSchemaKey, MainSchemaVersion, SchemaDirectionBase).
		Count(&count).Error; err != nil {
		t.Fatalf("读取 schema 变更记录失败: %v", err)
	}
	if count != 1 {
		t.Fatalf("当前主业务库基线应写入 1 条变更记录，实际 %d", count)
	}
}

func TestApplyEmbeddedBootstrapSchemaRecordsRevisionAndValidates(t *testing.T) {
	db := openTestSQLite(t, filepath.Join(t.TempDir(), "config.db"))

	checksum, err := ApplyEmbeddedBootstrapSchema(db, "test")
	if err != nil {
		t.Fatalf("初始化配置库失败: %v", err)
	}
	if strings.TrimSpace(checksum) == "" {
		t.Fatal("初始化后应返回 schema checksum")
	}
	if err := ValidateBootstrapSchema(db); err != nil {
		t.Fatalf("当前配置库基线应通过校验: %v", err)
	}

	var count int64
	if err := db.Table(SchemaRevisionTable).
		Where("schema_key = ? AND version = ? AND direction = ?", BootstrapSchemaKey, BootstrapSchemaVersion, SchemaDirectionBase).
		Count(&count).Error; err != nil {
		t.Fatalf("读取配置库 schema 变更记录失败: %v", err)
	}
	if count != 1 {
		t.Fatalf("当前配置库基线应写入 1 条变更记录，实际 %d", count)
	}
}

func TestValidateMainSchemaRejectsExistingDatabaseWithoutMetadata(t *testing.T) {
	db := openTestSQLite(t, filepath.Join(t.TempDir(), "legacy.db"))
	if err := db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT)").Error; err != nil {
		t.Fatalf("创建旧库模拟表失败: %v", err)
	}

	err := ValidateMainSchema(db)
	if err == nil {
		t.Fatal("缺少 schema 元数据的旧库不应通过校验")
	}
	if !strings.Contains(err.Error(), "请删除对应数据库文件后重启") {
		t.Fatalf("旧库错误提示应要求删除重建，实际: %v", err)
	}
}

func TestValidateMainSchemaRejectsMetadataChecksumMismatch(t *testing.T) {
	db := openTestSQLite(t, filepath.Join(t.TempDir(), "main.db"))
	if _, err := ApplyEmbeddedMainSchema(db, "test"); err != nil {
		t.Fatalf("初始化主业务库失败: %v", err)
	}
	if err := db.Exec("UPDATE schema_metadata SET schema_checksum = ? WHERE schema_key = ?", "tampered", MainSchemaKey).Error; err != nil {
		t.Fatalf("篡改 schema 元数据失败: %v", err)
	}

	err := ValidateMainSchema(db)
	if err == nil {
		t.Fatal("schema 元数据 checksum 不匹配时不应通过校验")
	}
	if !strings.Contains(err.Error(), "数据库结构 checksum 不匹配") {
		t.Fatalf("checksum 错误提示不符合预期: %v", err)
	}
}

func TestValidateMainSchemaRejectsRevisionChecksumMismatch(t *testing.T) {
	db := openTestSQLite(t, filepath.Join(t.TempDir(), "main.db"))
	if _, err := ApplyEmbeddedMainSchema(db, "test"); err != nil {
		t.Fatalf("初始化主业务库失败: %v", err)
	}
	if err := db.Exec(
		"UPDATE schema_revisions SET checksum = ? WHERE schema_key = ? AND version = ?",
		"tampered",
		MainSchemaKey,
		MainSchemaVersion,
	).Error; err != nil {
		t.Fatalf("篡改 schema 变更记录失败: %v", err)
	}

	err := ValidateMainSchema(db)
	if err == nil {
		t.Fatal("schema 变更记录 checksum 不匹配时不应通过校验")
	}
	if !strings.Contains(err.Error(), "数据库结构变更记录 checksum 不匹配") {
		t.Fatalf("变更记录 checksum 错误提示不符合预期: %v", err)
	}
}

func TestShouldInitializeSQLite(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing.db")
	shouldInit, err := ShouldInitializeSQLite(missingPath)
	if err != nil {
		t.Fatalf("检查缺失 SQLite 文件失败: %v", err)
	}
	if !shouldInit {
		t.Fatal("缺失 SQLite 文件应触发初始化")
	}

	emptyPath := filepath.Join(t.TempDir(), "empty.db")
	if err := os.WriteFile(emptyPath, nil, 0600); err != nil {
		t.Fatalf("创建空 SQLite 文件失败: %v", err)
	}
	shouldInit, err = ShouldInitializeSQLite(emptyPath)
	if err != nil {
		t.Fatalf("检查空 SQLite 文件失败: %v", err)
	}
	if !shouldInit {
		t.Fatal("0 字节 SQLite 文件应触发初始化")
	}

	initializedPath := filepath.Join(t.TempDir(), "initialized.db")
	initializedDB := openTestSQLite(t, initializedPath)
	if err := initializedDB.Exec("CREATE TABLE initialized_marker (id INTEGER PRIMARY KEY AUTOINCREMENT)").Error; err != nil {
		t.Fatalf("写入已初始化 SQLite 文件失败: %v", err)
	}
	shouldInit, err = ShouldInitializeSQLite(initializedPath)
	if err != nil {
		t.Fatalf("检查已初始化 SQLite 文件失败: %v", err)
	}
	if shouldInit {
		t.Fatal("已存在非空 SQLite 文件不应触发初始化")
	}
}

func openTestSQLite(t *testing.T, path string) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("打开测试 SQLite 数据库失败: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}
