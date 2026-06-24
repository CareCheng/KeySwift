// Package database 内嵌当前数据库基线 SQL。
//
// 这些 SQL 文件既是开发期显式构建来源，也是首次启动空库初始化来源。
// 已存在数据库只做结构元数据校验，不在运行时补旧字段或执行兼容迁移。
package database

import "embed"

// Files 保存当前配置库和主业务库的 SQLite 基线 SQL。
//
//go:embed bootstrap/sqlite/schema.sql main/sqlite/schema.sql main/sqlite/seed.sql
var Files embed.FS
