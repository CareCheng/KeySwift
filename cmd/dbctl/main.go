// dbctl 是 KeySwift 开发期数据库显式构建工具。
//
// 普通服务启动不会创建表结构。初始化全新开发数据库时，请在 Program 目录执行：
//
//	go run ./cmd/dbctl -target all
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"user-frontend/internal/dbschema"
	"user-frontend/internal/utils"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	target := flag.String("target", "all", "构建目标：bootstrap、main、all")
	configDir := flag.String("config-dir", "user_config", "配置目录")
	schemaDir := flag.String("schema-dir", "", "schema 根目录，默认使用 Program/database")
	mainDBPath := flag.String("main-db", "", "主业务库 SQLite 路径，默认使用 <config-dir>/user_data.db")
	serverPort := flag.Int("server-port", 8080, "HTTP 服务端口")
	appVersion := flag.String("app-version", "development", "写入 schema 元数据的程序版本")
	flag.Parse()

	if *mainDBPath == "" {
		*mainDBPath = filepath.Join(*configDir, "user_data.db")
	}
	resolvedSchemaDir := dbschema.ResolveSchemaDir(*schemaDir)

	if err := os.MkdirAll(*configDir, 0755); err != nil {
		log.Fatalf("创建配置目录失败: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(*mainDBPath), 0755); err != nil {
		log.Fatalf("创建主库目录失败: %v", err)
	}

	switch *target {
	case "bootstrap":
		runBootstrap(*configDir, resolvedSchemaDir, *mainDBPath, *serverPort, *appVersion)
	case "main":
		runMain(*mainDBPath, resolvedSchemaDir, *appVersion)
	case "all":
		runBootstrap(*configDir, resolvedSchemaDir, *mainDBPath, *serverPort, *appVersion)
		runMain(*mainDBPath, resolvedSchemaDir, *appVersion)
	default:
		log.Fatalf("未知构建目标: %s", *target)
	}
}

func runBootstrap(configDir, schemaDir, mainDBPath string, serverPort int, appVersion string) {
	dbPath := filepath.Join(configDir, "db-config.db")
	db := openSQLite(dbPath)
	checksum, err := dbschema.ApplySQLiteSchema(db, dbschema.ApplyOptions{
		SchemaPath:    filepath.Join(schemaDir, "bootstrap", "sqlite", "schema.sql"),
		MetadataTable: "bootstrap_schema_metadata",
		SchemaKey:     dbschema.BootstrapSchemaKey,
		SchemaVersion: dbschema.BootstrapSchemaVersion,
		AppVersion:    appVersion,
	})
	if err != nil {
		log.Fatalf("构建配置库失败: %v", err)
	}

	encryptionKey, err := utils.GenerateAESKey(256)
	if err != nil {
		log.Fatalf("生成配置加密密钥失败: %v", err)
	}
	statement := `INSERT INTO db_configs
		(id, type, host, port, user, password, database, server_port, encryption_key, key_length, created_at, updated_at)
		VALUES (1, 'sqlite', 'localhost', 3306, '', '', ?, ?, ?, 256, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET
			type = excluded.type,
			host = excluded.host,
			port = excluded.port,
			user = excluded.user,
			database = excluded.database,
			server_port = excluded.server_port,
			encryption_key = CASE WHEN db_configs.encryption_key = '' THEN excluded.encryption_key ELSE db_configs.encryption_key END,
			key_length = 256,
			updated_at = CURRENT_TIMESTAMP`
	if err := db.Exec(statement, mainDBPath, serverPort, encryptionKey).Error; err != nil {
		log.Fatalf("写入配置库默认配置失败: %v", err)
	}
	fmt.Printf("配置库构建完成: %s\nchecksum: %s\n", dbPath, checksum)
}

func runMain(mainDBPath, schemaDir, appVersion string) {
	db := openSQLite(mainDBPath)
	checksum, err := dbschema.ApplySQLiteSchema(db, dbschema.ApplyOptions{
		SchemaPath:    filepath.Join(schemaDir, "main", "sqlite", "schema.sql"),
		SeedPath:      filepath.Join(schemaDir, "main", "sqlite", "seed.sql"),
		MetadataTable: "schema_metadata",
		SchemaKey:     dbschema.MainSchemaKey,
		SchemaVersion: dbschema.MainSchemaVersion,
		AppVersion:    appVersion,
	})
	if err != nil {
		log.Fatalf("构建主业务库失败: %v", err)
	}
	fmt.Printf("主业务库构建完成: %s\nchecksum: %s\n", mainDBPath, checksum)
}

func openSQLite(path string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("打开 SQLite 数据库失败 %s: %v", path, err)
	}
	return db
}
