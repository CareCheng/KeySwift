# 04 宿主对接规范

## 1. 对接原则

插件和宿主的关系是“插件提交能力和事实，宿主登记、校验、裁决正式状态”。

插件不得直接修改：

- 用户状态。
- 管理员状态。
- 会话状态。
- 权限授权。
- 订单状态。
- 支付状态。
- 交付状态。
- 余额账本。
- 插件治理状态。

## 2. 权限对接

插件权限写入 manifest 的 `permissions`。

推荐命名：

```text
plugin.<namespace>.<resource>.<action>
```

示例：

```text
plugin.payment_alipay.config.view
plugin.payment_alipay.config.update
plugin.payment_alipay.callback.view
plugin.fulfillment_manual_kami.card.import
plugin.fulfillment_manual_kami.card.export
```

## 3. 菜单和页面对接

插件通过 `frontend.pages` 和 `frontend.menus` 对接后台。

规则：

- 页面必须绑定权限。
- 菜单只负责入口，不负责权限裁决。
- 页面路径不得覆盖宿主核心页面。
- 配置页应优先挂在插件管理或对应业务域下。

## 4. 配置对接

简单配置使用 `ConfigSchema`。

插件运行配置应优先由宿主统一落库到 `plugin_config_values`，并由宿主在首次同步或安装后按 schema 自动初始化默认结构。只有当插件确实存在独立业务数据表且该表不属于宿主统一配置表时，才额外声明自己的 `config` 表。

配置字段规则：

- 密钥字段必须标记 `secret`。
- 配置变更必须审计。
- 高风险配置必须绑定高风险权限。
- 插件不得把宿主密钥复制到自己的表中。
- 默认配置字段应在宿主首次同步插件 schema 后自动写入 `plugin_config_values`，普通字段写入 `value_json`，敏感字段只在用户保存后写入 `secret_json`。
- 插件运行时应只把宿主数据库中的配置视为唯一事实源，不得再从 release 文件、环境变量或其他外部配置位置读取业务配置。

## 5. 事件对接

插件可以声明生产或消费事件。

典型事件：

| 事件 | 说明 |
| --- | --- |
| `order.created` | 宿主订单创建 |
| `order.payment.paid` | 宿主确认订单已支付 |
| `fulfillment.requested` | 宿主请求交付 |
| `fulfillment.succeeded` | 插件提交交付成功事实 |
| `payment.fact.accepted` | 支付插件提交支付事实 |

插件消费事件后，应写入自己的业务表并向宿主回传事实。

## 6. 任务对接

长耗时或可重试操作必须使用任务方式。

适用场景：

- 发卡。
- 对账。
- 退款同步。
- 外部服务同步。
- 批量导入导出。

任务规则：

- 必须声明 `jobType`。
- 必须支持幂等。
- 必须记录失败原因。
- 必须声明重试策略。

## 7. 支付插件对接

支付插件只提交支付事实，不直接改订单。

流程：

1. 宿主创建订单。
2. 用户选择支付插件渠道。
3. 宿主创建支付尝试。
4. 支付插件发起通道支付。
5. 支付插件接收回调并记录插件交易事实。
6. 支付插件向宿主提交支付事实。
7. 宿主校验幂等、金额、订单、状态后裁决支付状态。

## 8. 发卡插件对接

发卡插件只提交交付事实，不直接改订单完成状态。

流程：

1. 宿主确认订单已支付。
2. 宿主创建交付任务。
3. 发卡插件领取任务。
4. 插件锁定库存并写入自己的交付表。
5. 插件提交交付成功或失败事实。
6. 宿主写入交付摘要并裁决订单状态。

## 9. 主题插件对接

主题插件通过 `ui` 声明：

- `themeScope`
- `tokenExtensions`
- `componentOverridesRef`
- `layoutSkinRef`
- `iconPackRef`

主题插件只能影响渲染，不得影响业务数据和权限裁决。

[下一章：安全与审计规范](./05-security-and-audit.md)
