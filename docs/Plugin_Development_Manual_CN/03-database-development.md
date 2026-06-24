# 03 数据库开发规范

## 1. 总原则

插件可以声明自己的数据库表，但不能直接控制宿主数据库。

强制规则：

- 插件表必须在 `manifest.database` 中声明。
- 插件表必须以 `plugin_<namespace>_<table_key>` 命名。
- 插件表由宿主登记、校验、显式创建和审计。
- 插件不得直接获取主库 DSN。
- 插件不得直接修改宿主核心表。
- 插件不得访问其他插件表。
- 当前开发阶段只向最新结构前进，不做旧表旧字段兼容。

## 2. database 顶层结构

```json
{
  "database": {
    "namespace": "payment_alipay",
    "storageMode": "host-main-db",
    "tables": [],
    "extensions": {}
  }
}
```

字段：

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| `namespace` | 是 | 插件数据库命名空间，只允许小写字母、数字和下划线 |
| `storageMode` | 是 | 当前固定为 `host-main-db` |
| `tables` | 是 | 插件声明的表 |
| `extensions` | 否 | 非关键扩展信息 |

## 3. 表声明

```json
{
  "tableKey": "channels",
  "physicalName": "plugin_payment_alipay_channels",
  "tableKind": "config",
  "schemaVersion": "1.0.0",
  "schemaChecksum": "sha256:...",
  "description": "支付宝渠道配置表",
  "sensitivity": "sensitive",
  "createPolicy": "on_enable",
  "dropPolicy": "manual_only",
  "backupPolicy": "encrypted",
  "retentionPolicy": "retain",
  "columns": [],
  "indexes": [],
  "relations": [],
  "operations": []
}
```

表类型：

| tableKind | 说明 |
| --- | --- |
| `config` | 插件配置、渠道、策略、模板 |
| `data` | 插件正式业务数据 |
| `runtime` | 插件运行态、游标、同步状态 |
| `cache` | 可清理缓存 |
| `audit` | 插件审计补充，不替代宿主审计 |

## 4. 字段声明

```json
{
  "columnKey": "order_no",
  "columnName": "order_no",
  "dbType": "TEXT",
  "logicalType": "host_order_ref",
  "nullable": false,
  "defaultValue": null,
  "primaryKey": false,
  "autoIncrement": false,
  "unique": false,
  "indexed": true,
  "encrypted": false,
  "secret": false,
  "referenceType": "host_resource",
  "referenceTarget": "order.order_no",
  "description": "关联宿主正式订单号"
}
```

字段规则：

- `columnName` 必须是 `snake_case`。
- `dbType` 必须明确，不允许空值。
- 敏感字段必须设置 `secret` 或 `encrypted`。
- 高频查询字段必须声明索引。
- 引用宿主资源时必须写 `referenceType` 和 `referenceTarget`。
- 不得用 JSON 字段承载正式状态、金额、权限、订单号等核心查询字段。

## 5. 索引声明

```json
{
  "indexKey": "order_no",
  "indexName": "idx_plugin_payment_alipay_transactions_order_no",
  "columns": ["order_no"],
  "unique": false
}
```

索引规则：

- 索引名必须包含插件命名空间。
- 唯一索引必须说明业务唯一性来源。
- 不允许声明跨插件表索引。

## 6. 关系声明

```json
{
  "relationKey": "order",
  "localColumn": "order_no",
  "targetResourceType": "host.order",
  "targetKey": "order_no",
  "relationType": "many_to_one",
  "required": true,
  "onDeletePolicy": "restrict"
}
```

关系规则：

- 插件通过稳定业务键引用宿主资源。
- 插件不得依赖宿主内部自增 ID 作为跨边界唯一事实。
- 插件不能通过关系声明获得修改宿主状态的权限。

## 7. 结构操作

插件可以声明结构操作，但执行权属于宿主。

```json
{
  "operationId": "create_channels_v1",
  "operationType": "create_table",
  "path": "database/sqlite/create_channels.sql",
  "checksum": "sha256:...",
  "requiresReview": true
}
```

执行流程：

1. 宿主发现 manifest。
2. 宿主校验 database 声明。
3. 宿主写入数据库治理表。
4. 后台展示待执行结构操作。
5. 管理员显式执行。
6. 宿主写入操作记录。
7. 插件表状态变为可用。

## 8. 宿主治理表

宿主会登记以下表：

- `plugin_database_tables`
- `plugin_database_columns`
- `plugin_database_indexes`
- `plugin_database_relations`
- `plugin_database_operations`

插件开发者应确保 manifest 中的声明足够完整，使后台可以展示、审核和执行。

## 9. 支付插件示例表

| 表 | 类型 | 说明 |
| --- | --- | --- |
| `plugin_payment_alipay_channels` | config | 支付渠道配置 |
| `plugin_payment_alipay_transactions` | data | 通道交易事实 |
| `plugin_payment_alipay_callbacks` | data | 原始回调摘要 |
| `plugin_payment_alipay_reconcile_records` | data | 对账记录 |
| `plugin_payment_alipay_refunds` | data | 通道退款事实 |

## 10. 手动卡密插件示例表

| 表 | 类型 | 说明 |
| --- | --- | --- |
| `plugin_fulfillment_manual_kami_batches` | data | 导入批次 |
| `plugin_fulfillment_manual_kami_cards` | data | 卡密库存 |
| `plugin_fulfillment_manual_kami_reservations` | data | 库存锁定 |
| `plugin_fulfillment_manual_kami_deliveries` | data | 发卡记录 |
| `plugin_fulfillment_manual_kami_settings` | config | 发卡策略 |

[下一章：宿主对接规范](./04-host-integration.md)
