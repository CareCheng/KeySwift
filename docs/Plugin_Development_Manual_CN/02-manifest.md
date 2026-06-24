# 02 manifest 协议

## 1. 基本要求

每个插件版本必须包含 `manifest.json`。

必填字段：

- `manifestVersion`
- `id`
- `version`
- `pluginKind`
- `identity`
- `compatibility`
- `package`
- `integrity`

涉及功能扩展时按需声明：

- `permissions`
- `backend`
- `database`
- `frontend`
- `ui`
- `observability`
- `operations`

## 2. 最小示例

```json
{
  "manifestVersion": "1.0.0",
  "id": "payment.alipay",
  "version": "1.0.0",
  "pluginKind": "integration",
  "identity": {
    "name": "payment-alipay",
    "displayName": "支付宝支付",
    "description": "提供支付宝支付渠道接入",
    "categories": ["payment"],
    "keywords": ["payment", "alipay"],
    "tags": ["支付", "支付宝"]
  },
  "compatibility": {
    "hostApp": "KeySwift",
    "minHostVersion": "0.1.0",
    "backendProtocol": "1.0.0",
    "frontendProtocol": "1.0.0",
    "handshakeProtocol": "1.0.0",
    "supportedPlatforms": ["windows", "linux"],
    "supportedArchitectures": ["amd64"],
    "runtimeModes": ["process"],
    "databaseModes": ["host-main-db"]
  },
  "package": {
    "packageFormat": "keyswift-plugin",
    "installMode": "local-directory",
    "distributionType": "local",
    "entryStrategy": "process",
    "binaries": [
      {
        "platform": "windows",
        "arch": "amd64",
        "path": "bin/windows-amd64/payment-alipay.exe"
      }
    ]
  },
  "integrity": {
    "enabled": true,
    "hashAlgorithm": "sha256",
    "checksumFile": "checksums.json",
    "verifyOnInstall": true,
    "verifyOnEnable": true
  },
  "permissions": [],
  "backend": {},
  "database": {
    "namespace": "payment_alipay",
    "storageMode": "host-main-db",
    "tables": []
  }
}
```

## 3. 身份与分类声明

`identity` 描述插件的人类可读信息，也承载后台展示、检索和动态筛选所需的分类元数据。

字段：

| 字段 | 说明 |
| --- | --- |
| `name` | 插件名称，建议与插件 ID 或包名保持可识别关系 |
| `displayName` | 后台展示名称 |
| `description` | 插件用途说明 |
| `categories` | 插件业务分类，用于后台插件管理页动态筛选 |
| `keywords` | 搜索关键词，用于后台检索增强 |
| `tags` | 展示标签，可使用中文 |
| `author` | 作者 |
| `organization` | 组织 |
| `license` | 许可证 |
| `homepage` | 首页 |
| `repository` | 代码仓库 |
| `support` | 支持入口 |

### 3.1 分类字段规则

分类统一写入：

```json
"identity": {
  "categories": ["payment"]
}
```

规则：

- `categories` 必须是字符串数组。
- 分类值使用稳定小写 slug。
- 推荐格式为 `^[a-z][a-z0-9-]{1,63}$`。
- 一个插件可以声明多个分类，第一个分类视为主分类。
- 后台插件管理页根据当前已安装插件真实声明的分类动态生成筛选项。
- 推荐分类表只作为插件开发规范，不是后台固定筛选枚举。
- 未声明分类的插件不应导致扫描失败，但只能在“全部分类”下展示。
- `tags` 可以使用中文展示标签，但 `categories` 不应使用中文。
- `keywords` 用于搜索增强，不等同于分类。

推荐写法：

```json
"identity": {
  "name": "payment-alipay",
  "displayName": "支付宝支付",
  "description": "提供支付宝支付渠道接入",
  "categories": ["payment"],
  "keywords": ["payment", "alipay"],
  "tags": ["支付", "支付宝"]
}
```

禁止写法：

```json
"identity": {
  "categories": ["支付", "第一阶段", "plugin_runtime_todo"]
}
```

### 3.2 推荐初始分类

| 分类值 | 中文展示建议 | 适用范围 |
| --- | --- | --- |
| `payment` | 支付 | 支付渠道、回调、对账、退款 |
| `fulfillment` | 发卡交付 | 卡密发放、库存锁定、交付回传 |
| `security` | 安全 | 人机验证、风控、登录保护 |
| `human-verification` | 人机验证 | 图片验证码、Turnstile、其他验证 provider |
| `customer-support` | 客服支持 | 工单、在线客服、消息处理 |
| `marketing` | 营销 | 优惠券、活动、推广、积分扩展 |
| `notification` | 通知 | 邮件、短信、Webhook、机器人通知 |
| `analytics` | 分析统计 | 报表、监控、运营分析 |
| `tooling` | 工具维护 | 诊断、导入导出、批处理 |
| `ui-theme` | 主题外观 | 主题 token、皮肤、组件覆盖 |

## 4. 权限声明

插件权限必须写入 `permissions`。

字段：

| 字段 | 说明 |
| --- | --- |
| `key` | 权限点，建议格式为 `plugin.<namespace>.<resource>.<action>` |
| `title` | 后台展示名称 |
| `description` | 权限说明 |
| `scope` | 权限作用范围 |
| `namespace` | 插件命名空间 |
| `kind` | `page`、`action`、`config`、`data` |
| `riskLevel` | `low`、`medium`、`high`、`critical` |
| `defaultVisibility` | 默认是否对普通管理员可见 |

## 5. 后端声明

`backend` 用于声明插件后端入口、路由、事件、任务和结构操作。

注意：

- 插件路由由宿主代理或挂载。
- 插件回调必须经过宿主鉴权和限流。
- 插件异步任务必须有幂等键。
- 数据库结构不要只写在 `backend.migrations`，必须同步写入 `database`。

## 6. 前端声明

`frontend` 用于声明插件页面、菜单、表单和视图。

页面声明必须包含：

- `id`
- `area`
- `path`
- `title`
- `viewId`
- `renderMode`
- `permissionKeys`

菜单声明必须包含：

- `id`
- `targetPageId`
- `title`
- `defaultGroup`
- `order`
- `permissionKeys`

## 7. 主题声明

主题插件使用 `ui`。

字段：

| 字段 | 说明 |
| --- | --- |
| `enabled` | 是否启用 UI 贡献 |
| `uiKind` | `theme`、`component-override`、`layout-skin` |
| `themeScope` | `admin`、`user`、`global` |
| `tokenExtensions` | 主题 token 扩展 |
| `componentOverridesRef` | 组件覆盖引用 |
| `layoutSkinRef` | 布局皮肤引用 |
| `iconPackRef` | 图标包引用 |

主题插件不得修改业务状态，只能贡献 UI 资源和渲染配置。

[下一章：数据库开发规范](./03-database-development.md)
