# 插件数据库声明说明

插件业务表不在本目录直接保存 SQL 文件。

插件必须在 `manifest.json` 的 `database` 区块中声明：

- 命名空间
- 表键
- 物理表名
- 表类型
- 字段
- 索引
- 关系
- 结构操作声明
- 敏感级别
- 备份策略

宿主读取声明后写入 `plugin_database_tables`、`plugin_database_columns`、`plugin_database_indexes`、`plugin_database_relations` 和 `plugin_database_operations`。

插件不得直接获取主库 DSN，也不得直接修改宿主核心表。插件数据访问必须经过宿主存储网关或后续约定的受控接口。
