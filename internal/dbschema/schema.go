// Package dbschema 提供数据库显式构建、首次空库初始化与启动校验能力。
//
// 普通服务启动只允许在数据库文件不存在时创建当前干净基线。
// 如果数据库文件已存在，则只校验 schema 元数据，不补旧字段、不做兼容迁移。
// 数据库初始化请使用 Program/docs/Plugin_Development_Manual_CN/03-database-development.md
// 和 Program/database 下的 schema 文件。
package dbschema

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	embeddeddb "user-frontend/database"

	"gorm.io/gorm"
)

const (
	BootstrapSchemaKey     = "bootstrap"
	BootstrapSchemaVersion = "2026.06.20.1"
	MainSchemaKey          = "main"
	MainSchemaVersion      = "2026.06.22.1"
)

// Metadata 是配置库和主库共用的 schema 元数据结构。
type Metadata struct {
	SchemaKey      string
	SchemaVersion  string
	SchemaChecksum string
	AppVersion     string
}

// ValidateBootstrapSchema 校验配置库结构元数据。
func ValidateBootstrapSchema(db *gorm.DB) error {
	return ValidateSchema(db, "bootstrap_schema_metadata", BootstrapSchemaKey, BootstrapSchemaVersion)
}

// ValidateMainSchema 校验主业务库结构元数据。
func ValidateMainSchema(db *gorm.DB) error {
	return ValidateSchema(db, "schema_metadata", MainSchemaKey, MainSchemaVersion)
}

// ValidateSchema 只读取结构元数据，不创建、不迁移、不补偿。
func ValidateSchema(db *gorm.DB, metadataTable, schemaKey, expectedVersion string) error {
	if db == nil {
		return errors.New("数据库连接为空")
	}

	var metadata Metadata
	query := fmt.Sprintf(
		"SELECT schema_key, schema_version, schema_checksum, app_version FROM %s WHERE schema_key = ? LIMIT 1",
		metadataTable,
	)
	if err := db.Raw(query, schemaKey).Scan(&metadata).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such table") {
			return metadataMissingError(schemaKey)
		}
		return fmt.Errorf("读取数据库结构元数据失败: %w", err)
	}
	if strings.TrimSpace(metadata.SchemaKey) == "" {
		return metadataMissingError(schemaKey)
	}
	if metadata.SchemaVersion != expectedVersion {
		return fmt.Errorf("数据库结构版本不匹配: 当前 %s，程序需要 %s", metadata.SchemaVersion, expectedVersion)
	}
	if strings.TrimSpace(metadata.SchemaChecksum) == "" {
		return errors.New("数据库结构 checksum 为空")
	}
	return nil
}

// ApplyOptions 描述一次显式 schema 构建任务。
type ApplyOptions struct {
	SchemaPath    string
	SeedPath      string
	MetadataTable string
	SchemaKey     string
	SchemaVersion string
	AppVersion    string
}

// ApplyContentOptions 描述一次基于内存 SQL 内容的 schema 构建任务。
type ApplyContentOptions struct {
	SchemaSQL     string
	SeedSQL       string
	MetadataTable string
	SchemaKey     string
	SchemaVersion string
	AppVersion    string
}

// ShouldInitializeSQLite 判断 SQLite 文件是否需要按当前基线初始化。
func ShouldInitializeSQLite(path string) (bool, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	if info.IsDir() {
		return false, fmt.Errorf("数据库路径是目录: %s", path)
	}
	return info.Size() == 0, nil
}

// ApplySQLiteSchema 执行 SQLite schema 和 seed 文件，并写入 schema 元数据。
func ApplySQLiteSchema(db *gorm.DB, options ApplyOptions) (string, error) {
	if strings.TrimSpace(options.SchemaPath) == "" {
		return "", errors.New("schema 文件路径不能为空")
	}

	schemaSQL, _, err := ReadSQLFile(options.SchemaPath)
	if err != nil {
		return "", err
	}
	seedSQL := ""
	if strings.TrimSpace(options.SeedPath) != "" {
		seedSQL, _, err = ReadSQLFile(options.SeedPath)
		if err != nil {
			return "", err
		}
	}
	return ApplySQLiteSchemaContent(db, ApplyContentOptions{
		SchemaSQL:     schemaSQL,
		SeedSQL:       seedSQL,
		MetadataTable: options.MetadataTable,
		SchemaKey:     options.SchemaKey,
		SchemaVersion: options.SchemaVersion,
		AppVersion:    options.AppVersion,
	})
}

// ApplyEmbeddedBootstrapSchema 使用内嵌 SQL 初始化全新的配置库。
func ApplyEmbeddedBootstrapSchema(db *gorm.DB, appVersion string) (string, error) {
	schemaSQL, _, err := ReadEmbeddedSQL("bootstrap/sqlite/schema.sql")
	if err != nil {
		return "", err
	}
	return ApplySQLiteSchemaContent(db, ApplyContentOptions{
		SchemaSQL:     schemaSQL,
		MetadataTable: "bootstrap_schema_metadata",
		SchemaKey:     BootstrapSchemaKey,
		SchemaVersion: BootstrapSchemaVersion,
		AppVersion:    appVersion,
	})
}

// ApplyEmbeddedMainSchema 使用内嵌 SQL 初始化全新的主业务库。
func ApplyEmbeddedMainSchema(db *gorm.DB, appVersion string) (string, error) {
	schemaSQL, _, err := ReadEmbeddedSQL("main/sqlite/schema.sql")
	if err != nil {
		return "", err
	}
	seedSQL, _, err := ReadEmbeddedSQL("main/sqlite/seed.sql")
	if err != nil {
		return "", err
	}
	return ApplySQLiteSchemaContent(db, ApplyContentOptions{
		SchemaSQL:     schemaSQL,
		SeedSQL:       seedSQL,
		MetadataTable: "schema_metadata",
		SchemaKey:     MainSchemaKey,
		SchemaVersion: MainSchemaVersion,
		AppVersion:    appVersion,
	})
}

// ApplySQLiteSchemaContent 执行 SQLite schema 内容和 seed 内容，并写入 schema 元数据。
func ApplySQLiteSchemaContent(db *gorm.DB, options ApplyContentOptions) (string, error) {
	if db == nil {
		return "", errors.New("数据库连接为空")
	}
	if strings.TrimSpace(options.SchemaSQL) == "" {
		return "", errors.New("schema 内容不能为空")
	}
	if strings.TrimSpace(options.MetadataTable) == "" || strings.TrimSpace(options.SchemaKey) == "" {
		return "", errors.New("schema 元数据配置不能为空")
	}
	if strings.TrimSpace(options.SchemaVersion) == "" {
		return "", errors.New("schema 版本不能为空")
	}
	if strings.TrimSpace(options.AppVersion) == "" {
		options.AppVersion = "development"
	}

	checksum := SQLChecksum(options.SchemaSQL)
	if err := ExecSQLScript(db, options.SchemaSQL); err != nil {
		return "", err
	}
	if strings.TrimSpace(options.SeedSQL) != "" {
		if err := ExecSQLScript(db, options.SeedSQL); err != nil {
			return "", err
		}
	}

	statement := fmt.Sprintf(`INSERT INTO %s
		(schema_key, schema_version, schema_checksum, app_version, initialized_at, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(schema_key) DO UPDATE SET
			schema_version = excluded.schema_version,
			schema_checksum = excluded.schema_checksum,
			app_version = excluded.app_version,
			updated_at = CURRENT_TIMESTAMP`, options.MetadataTable)
	if err := db.Exec(statement, options.SchemaKey, options.SchemaVersion, checksum, options.AppVersion).Error; err != nil {
		return "", fmt.Errorf("写入 schema 元数据失败: %w", err)
	}
	return checksum, nil
}

// SQLChecksum 计算 SQL 内容 SHA-256 指纹。
func SQLChecksum(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

// ReadSQLFile 读取 SQL 文件并返回 SHA-256 指纹。
func ReadSQLFile(path string) (string, string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("读取 SQL 文件失败 %s: %w", path, err)
	}
	return string(content), SQLChecksum(string(content)), nil
}

// ReadEmbeddedSQL 读取内嵌 SQL 文件并返回 SHA-256 指纹。
func ReadEmbeddedSQL(path string) (string, string, error) {
	content, err := embeddeddb.Files.ReadFile(filepath.ToSlash(path))
	if err != nil {
		return "", "", fmt.Errorf("读取内嵌 SQL 失败 %s: %w", path, err)
	}
	return string(content), SQLChecksum(string(content)), nil
}

func metadataMissingError(schemaKey string) error {
	return fmt.Errorf("数据库已存在但缺少 %s 结构元数据，可能是旧库或非当前基线库；请删除对应数据库文件后重启，程序会按当前基线重新创建", schemaKey)
}

// ExecSQLScript 逐条执行 SQL 脚本。
func ExecSQLScript(db *gorm.DB, script string) error {
	for _, statement := range SplitSQLStatements(script) {
		if strings.TrimSpace(statement) == "" {
			continue
		}
		if err := db.Exec(statement).Error; err != nil {
			return fmt.Errorf("执行 SQL 失败: %w\nSQL: %s", err, statement)
		}
	}
	return nil
}

// SplitSQLStatements 按分号拆分 SQL，同时忽略以 -- 开头的整行注释。
func SplitSQLStatements(script string) []string {
	scanner := bufio.NewScanner(strings.NewReader(script))
	var builder strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}
		builder.WriteString(line)
		builder.WriteByte('\n')
	}
	parts := strings.Split(builder.String(), ";")
	statements := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			statements = append(statements, trimmed)
		}
	}
	return statements
}

// ResolveSchemaDir 返回数据库 schema 目录。
func ResolveSchemaDir(explicit string) string {
	if strings.TrimSpace(explicit) != "" {
		return filepath.Clean(explicit)
	}
	return filepath.Clean("database")
}
