# 独立式插件开发手册

适用范围：`Program/plugins/` 下所有插件。  
当前版本：2026-06-23。  
目标：确保所有正式插件具备独立源码、独立构建脚本、清晰前后端边界、完整配置声明和可直接安装的 `.ksplugin.zip` 包。

## 1. 总原则

KeySwift 插件必须是独立插件，不得只是宿主中的硬编码分支或声明壳。

所有插件必须遵守：

1. 插件业务逻辑写在插件目录内。
2. 插件自身声明 manifest、配置 schema、权限、前端贡献、后端入口和数据声明。
3. 宿主只负责插件发现、安装、启停治理、配置读取、权限治理、统一协议调用、菜单挂载和审计。
4. 宿主不得为了某个插件硬编码该插件专属业务算法、第三方请求、页面逻辑或私有状态。
5. 插件必须能通过自己的构建脚本生成可安装包。

## 当前宿主插件能力基线

本节是当前程序代码已经支持或已经约定的插件规范基线，所有插件类型都必须按此基线设计，不限于人机验证插件。

1. 插件发现目录：宿主扫描标准插件目录下的 `plugins/<plugin-id>/releases/<version>/manifest.json`。
2. 插件安装入口：后台插件管理页通过选择 `.ksplugin.zip` 安装包调用 `/api/admin/plugins/install`，宿主解压到标准插件目录后立即刷新扫描结果。
3. 插件刷新入口：后台插件管理页调用 `/api/admin/plugins/refresh` 重新扫描磁盘插件并登记 manifest、配置 schema、权限、前端贡献和数据库声明。
4. 插件启停入口：后台通过 `/api/admin/plugin/{pluginID}/enable` 和 `/api/admin/plugin/{pluginID}/disable` 显式启停插件；扫描发现插件只能进入已发现或停用状态，不得隐式启用。
5. 插件详情入口：后台通过 `/api/admin/plugin/{pluginID}`、`/bindings`、`/migrations`、`/database`、`/runtime-records`、`/permission-definitions` 读取治理状态。
6. 插件配置入口：后台通过 `/api/admin/plugins/config-schemas` 读取 schema，通过 `/api/admin/plugin/{pluginID}/config-values` 读取和保存配置值。
7. 配置落库位置：配置 schema 进入 `plugin_config_schemas`，配置值进入 `plugin_config_values`，历史修订进入 `plugin_config_revisions`。
8. 插件注册表：插件安装、启停、健康、校验、manifest 快照登记在 `plugin_registry` 和 `plugin_versions`。
9. 插件权限：manifest `permissions` 会进入宿主权限定义和角色授权体系，后台菜单和敏感操作必须绑定权限。
10. 插件前端：manifest `frontend` 声明页面、菜单、表单、视图、settings page、bundle 或 iframe，宿主只按声明挂载。
11. 插件数据库：manifest `database` 声明插件自有表，宿主登记到 `plugin_database_declarations`、`plugin_database_tables`、`plugin_database_columns`、`plugin_database_indexes`、`plugin_database_relations`、`plugin_database_operations`。
12. 插件运行态：独立进程插件以 manifest `backend.entryExecutable`、`backend.controlProtocol`、`backend.dataProtocol` 和 `package.binaries` 为准，当前人机验证插件使用 `json-stdio`。
13. 插件完整性：manifest `integrity.checksumFile` 必须指向 `checksums.json`，安装包中的 manifest、配置 schema、二进制和前端产物都应纳入校验。
14. 插件默认状态：首次安装或扫描发现的插件默认不启用；管理员必须在后台明确启用后，插件能力才可参与业务流程。

## manifest 完整字段基线

当前宿主协议结构以 `Program/internal/plugin/types.go` 为准。正式插件的 `manifest.json` 应按插件能力完整声明以下字段，未使用的能力可为空对象或空数组，但不得省略必填基础字段。

1. `manifestVersion`：插件 manifest 协议版本，当前为 `1.0.0`。
2. `id`：全局唯一插件 ID，必须与目录和配置 schema 的 `pluginId` 一致。
3. `version`：插件版本，必须与 release 目录一致。
4. `pluginKind`：插件类型，可使用 `functional`、`ui-theme`、`integration`、`tooling`。
5. `identity`：名称、显示名、描述、作者、许可证、主页、仓库、支持地址、标签、分类、发布时间等人类可读信息。
6. `compatibility`：宿主名、宿主版本、前端/后端/握手协议、支持平台、支持架构、运行模式、数据库模式和宿主特性依赖。
7. `package`：包格式、安装模式、分发类型、入口策略、默认二进制、平台二进制、前端资源、文档、本地化、资源和摘要引用。
8. `integrity`：哈希算法、checksum 文件、签名文件、包摘要、manifest 摘要、安装/启用/启动前校验策略、篡改策略和未签名策略。
9. `lifecycle`：自动启动、启动/注册/就绪/停止超时、心跳、健康检查、重载、升级、回滚、排空和崩溃退避策略。
10. `dependencies`：插件依赖、宿主能力依赖、外部服务依赖、系统要求、冲突插件和迁移依赖。
11. `permissions`：插件页面、操作、配置、数据访问等权限点。
12. `capabilities`：后端、前端、数据、宿主、安全、可观测和实验能力。
13. `backend`：后端入口、控制协议、数据协议、路由、webhook、事件、任务、消费者、迁移、配置 schema 引用和服务契约。
14. `database`：命名空间、存储模式、表、字段、索引、关系和结构操作声明。
15. `frontend`：页面、菜单、表单、视图、settings page、工作区提示、bundle、iframe 和前端权限。
16. `ui`：主题、组件覆盖、布局皮肤、图标包和 UI 激活策略。
17. `observability`：健康探针、指标、日志通道、审计事件、诊断包、追踪标签和支持命令。
18. `operations`：安装、启用、停用、升级、回滚、卸载钩子和维护任务。
19. `metadata`：非协议关键的展示和说明元数据。
20. `extensions`：按命名空间扩展的能力声明，例如 `humanVerification`。

## 2. 标准目录

```text
Program/plugins/<plugin-id>/
├── AGENTS.md                         # 可选，插件内更细规则
├── build.ps1                         # 必需，Windows 构建与打包
├── build.sh                          # 必需，Linux/macOS 构建与打包
├── go.mod / package.json / Cargo.toml # 按插件技术栈选择
├── src/                              # 后端、命令行、服务端或 worker 源码
├── frontend/                         # 有自定义页面、组件或前端扩展时必需
├── assets/                           # 可选，静态资源
├── docs/                             # 可选，插件自己的开发与使用说明
├── releases/
│   └── <version>/
│       ├── manifest.json
│       ├── settings.schema.json
│       ├── checksums.json
│       ├── bin/<os>_<arch>/<plugin-binary>
│       └── frontend/                 # 有前端 bundle 时放构建产物
└── dist/packages/<plugin-id>-<version>-<os>_<arch>.ksplugin.zip
```

说明：

1. `src/` 承载插件后端或执行逻辑。
2. `frontend/` 承载插件自定义前端源码或静态前端入口；如果插件没有任何用户可见或后台可见交互，才可以在 manifest 中明确 `frontend.enabled=false`。
3. `settings.schema.json` 是所有需要配置的插件必须提供的配置契约，即使插件没有自定义配置页面，也应通过 schema 让宿主渲染配置表单。

## 3. 前端自包含规范

插件只要提供任何页面、菜单、配置 UI、业务组件或终端用户可见交互，就必须自包含前端契约。自包含前端契约不等于所有插件都必须提供自定义前端 bundle；简单配置表单或宿主已有通用组件可以由 manifest、`settings.schema.json` 和扩展元数据驱动宿主渲染。

插件前端能力分三类：

1. 宿主通用渲染：插件只提供 manifest、`settings.schema.json`、`extensions` 元数据，由宿主已有组件渲染。
2. 插件自定义页面：插件提供 `frontend/pages`、`frontend/menus` 或 `frontend/settingsPages` 声明，并提供前端 bundle。
3. 插件自定义组件：插件提供 bundle 或 iframe 入口，由宿主按 manifest 挂载到指定区域。

### 前端架构选型（强制与主程序同栈）

插件前端**必须**使用与主程序相同的前端架构：Next.js + React + TypeScript + Tailwind CSS + Zustand，并采用静态导出模式。主程序前端（`Program/web`）即此栈，插件前端不得使用其他框架（Vue/Svelte/Angular/Preact/原生 HTML 工程等）替代，也不得以"轻交互""单文件足够"为由回避同栈要求。

主程序当前前端版本基线（插件应与之保持大版本兼容）：

- Next.js（静态导出，`output: export`）
- React
- TypeScript
- Tailwind CSS
- Zustand
- Framer Motion（按需）

### 前端版本同步约束

插件前端必须与主程序前端保持版本一致，跟随主程序升级，不得长期停留在旧版本或自行选定偏离的版本。

1. **大版本跟随**：主程序前端 Next.js / React / Tailwind / Zustand 等核心依赖升级大版本时，所有维护中的插件前端必须同步升级到与主程序一致的版本，不得继续使用旧大版本。
2. **小版本/补丁对齐**：插件前端应尽量与主程序保持相同的小版本与补丁版本，至少保证同一大版本下的兼容范围；出现破坏性变更时以主程序为准。
3. **升级责任**：主程序前端版本变更后，宿主侧应在新版本落定后同步更新本手册与 `Program/plugins/AGENTS.md` 中的版本基线；插件开发者须在合理周期内完成各自插件前端的版本对齐与回归验证，未对齐的插件不得发布新版本。
4. **构建产物一致性**：版本跟随不仅指 `package.json` 依赖版本，还包括静态导出配置、Tailwind 配置约定、主题 token 命名等构建期约定；主程序调整这些约定时，插件前端须同步调整，保证产物形态与主程序一致。
5. **不兼容即停用**：插件前端未跟随升级且已与主程序前端产生不兼容时，宿主可按治理策略停用该插件前端贡献，直至插件完成版本对齐；不得为兼容旧插件而回退或冻结主程序前端版本。
6. **新插件门槛**：新建插件前端必须以当前主程序前端版本为基线起步，禁止以历史版本为基础开发。

按交互形态落地：

1. **宿主通用渲染**：插件只提供 manifest、`settings.schema.json`、`extensions` 元数据，由宿主已有组件渲染，无插件前端代码，自然满足同栈要求。
2. **插件自定义页面 / 自定义组件**：插件 `frontend/` 必须是一个独立的 Next.js 工程（或与主程序共用同一套构建约定的工程），使用 React + TypeScript + Tailwind 编写，构建产物为静态文件，放入 `releases/<version>/frontend/`。iframe 形态的自定义组件也必须以同栈构建产物挂载，不得用裸 HTML 文件充当。
3. **iframe 嵌入组件**：即便体量很小，也必须以 Next.js + React + Tailwind 工程产出，构建后挂载；不得用单个手写 `widget.html` 作为正式产物。

架构选型约束：

1. 插件前端必须自包含为独立 Next.js 工程，不得反向依赖宿主 `Program/web` 的源码、组件、构建配置或运行时模块；可在插件 `frontend/` 内复用与主程序一致的约定（目录结构、主题 token、Tailwind 配置），但需自带一份，不直接引用宿主工程。
2. 插件前端产物必须可由插件自己的 `build.ps1` / `build.sh` 独立构建生成（`next build` 静态导出），不得要求在宿主前端工程内构建。
3. iframe 形态的插件组件运行在独立文档内，**无法继承宿主 CSS 变量与 Tailwind 类**。视觉协调通过"主题 token 字典"机制实现（见下节）：构建时插件前端自带一份与主程序一致的默认 token 兜底；运行时宿主经标准 `postMessage` 下发当前主题 token 包，插件前端写入自身 `:root` 跟随宿主主题。不得硬编码宿主内部类名或假定宿主样式环境，也不得只切换 `data-theme` 主题名而不消费 token 值。
4. 插件前端不得引入对宿主路由、宿主全局状态、宿主接口鉴权 cookie 之外的隐式耦合；需要宿主数据时走 manifest 声明的接口或扩展协议。
5. 选型应匹配实际复杂度，但复杂度只影响是否需要 bundle，不改变强制同栈要求：禁止用其他框架或裸 HTML 替代 Next.js + React + Tailwind。

### 既有插件迁移说明

`keyswift.image_captcha`、`cloudflare.turnstile` 的前端已从裸 `widget.html` 迁移为独立 Next.js + React + TypeScript + Tailwind 工程（`frontend/` 目录），产物经 `next build` 静态导出后落到 `releases/<version>/frontend/widget.html` 与 `_next/`，作为新插件前端工程化的范本。其他既有插件若仍为裸 HTML 形态，须按上述同栈要求迁移。

### 主题 token 字典

iframe 插件前端无法继承宿主 CSS 变量，为适配宿主任意主题（含主题插件自定义配色），宿主与插件 iframe 前端之间约定一组**固定的主题 token 字段**，作为标准主题契约。

字段清单（字段名固定，主题插件只能改值，不得增删字段名；新增字段属于协议变更，须同步所有插件前端）：

| 字段名 | 含义 |
| --- | --- |
| `bg-input` | 输入框背景色 |
| `border-color` | 输入框/容器边框色 |
| `text-primary` | 主要文字色 |
| `text-secondary` | 次要文字色（标签等） |
| `text-muted` | 弱化文字色（提示、图标） |
| `text-placeholder` | 输入框 placeholder 色 |
| `primary-500` | 主色（焦点边框、强调） |
| `primary-ring` | 焦点 ring 半透明色 |

契约规则：

1. **宿主侧**：`Program/web/src/lib/themeTokens.ts` 定义字段清单与读取函数；从当前生效的 CSS 变量（`globals.css` 的 `:root` / `[data-theme]`）读取 token 值，在挂载 iframe 的 `init` 消息与主题切换的 `theme` 消息中携带 `theme_tokens` 字段下发。主题切换时仅下发 token，不重置插件业务状态（如验证码挑战）。
2. **插件前端侧**：各自 `frontend/src/lib/themeTokens.ts` 定义对称的契约与默认值；收到 `theme_tokens` 后调用 `applyThemeTokens` 写入 `document.documentElement` 的 `:root` 内联 style，缺字段用内置默认值（深色）兜底，避免白屏。`globals.css` 的 `:root` 仅保留默认值，不再用 `[data-theme]` 选择器切换。
3. **主题插件责任**：主题类插件（manifest `ui` 声明）改宿主配色时，必须通过覆盖上述 CSS 变量生效；宿主读取逻辑自动把新值下发给所有 iframe 插件前端。主题插件不得只改视觉而不更新这些标准变量，否则 iframe 插件无法跟随。
4. **兜底**：插件前端在收到 token 前先用 `:root` 默认值渲染，收到后热替换；宿主读取失败（SSR 或无 document）时下发空对象，由插件默认值兜底。
5. **样式引用**：插件前端 CSS 必须用 `var(--字段名)` 引用这些 token（如 `border-color: var(--border-color)`、焦点态用 `var(--primary-500)` / `var(--primary-ring)`），不得硬编码色值，否则无法跟随主题。

当前 `keyswift.image_captcha`、`cloudflare.turnstile` 已按本机制接入，可作为范本。

必须声明的 manifest 字段：

```json
{
  "frontend": {
    "enabled": true,
    "renderMode": "iframe",
    "pages": [],
    "menus": [],
    "settingsPages": [],
    "bundleEntries": [],
    "iframeEntries": [
      {
        "id": "example.widget",
        "entry": "frontend/widget.html",
        "sandboxPolicy": "allow-scripts allow-same-origin allow-forms"
      }
    ]
  }
}
```

前端插件要求：

1. 插件页面不得依赖宿主硬编码路由才能访问，必须通过 manifest 声明挂载位置。
2. 插件菜单必须声明权限，不能默认对所有后台角色可见。
3. 插件配置页面如果只是表单，优先使用 `settings.schema.json` 交给宿主渲染。
4. 插件需要复杂配置体验时，必须提供自定义 settings page 或 bundle。
5. 插件前端构建产物必须进入 `releases/<version>/frontend/`，并进入 `checksums.json`。
6. 插件不得把开发阶段、插件化阶段、后续规划等内部语义展示给最终用户。
7. 后台插件管理页必须提供通用 schema 表单渲染与保存能力；插件不提供自定义配置页时，配置仍必须可通过该通用入口完成。

## 4. 后端自包含规范

插件只要提供业务处理、第三方接口调用、后台任务、事件消费、支付、发卡、人机验证、同步等能力，就必须自包含后端逻辑。

后端插件必须声明：

```json
{
  "backend": {
    "entryExecutable": "bin/windows_amd64/plugin.exe",
    "controlProtocol": "json-stdio",
    "dataProtocol": "json-stdio",
    "routes": [],
    "events": [],
    "jobs": [],
    "settingsRef": "settings.schema.json"
  }
}
```

后端插件要求：

1. 第三方 API 请求必须在插件内完成，不能写入宿主核心。
2. 插件私有状态必须写入插件数据目录或插件声明的数据表。
3. 插件异常必须返回明确错误，不能要求宿主静默放行。
4. 插件二进制必须按平台放入 `bin/<os>_<arch>/`。
5. 插件运行协议必须在 manifest 中明确声明。

## 5. 配置与配置页面

所有需要配置的插件都必须提供 `settings.schema.json`。

插件运行配置必须由宿主写入并读取主数据库统一插件配置表，当前标准表为 `plugin_config_values`，配置 schema 登记在 `plugin_config_schemas`。宿主在插件安装、首次刷新或启动同步后，应根据 schema 自动为该插件创建默认配置记录，并在 schema 新增普通字段时自动补齐缺失字段。插件不得要求管理员手工编辑外部配置文件、release 目录文件、环境变量或宿主源码来完成业务配置。

配置 schema 至少说明：

1. 配置分组。
2. 字段 key。
3. 字段类型。
4. 是否必填。
5. 默认值。
6. 是否敏感。
7. 字段说明。
8. 存储位置，默认使用 `storage.mode=host-main-db`、`storage.table=plugin_config_values`。

敏感配置要求：

1. secret、token、私钥、服务端密钥必须标记 `secret: true`。
2. 敏感值不得出现在公开配置、前端 payload、普通日志或审计摘要中。
3. 插件缺少必填配置时必须返回 `missing_config` 或等效不可用状态。

配置页面规则：

1. 简单配置使用宿主配置表单渲染 `settings.schema.json`。
2. 复杂配置使用插件自定义 settings page。
3. 自定义 settings page 也必须以 schema 作为后端配置契约，不能只靠前端私有字段。
4. 后台插件管理页保存配置时必须写入主数据库插件配置表，不得写入插件包目录或临时文件。
5. 插件首次同步后，宿主应先写入默认配置记录；普通字段先占位，敏感字段未填写时不得伪装成已保存密钥。

当前宿主配置链路：

1. 插件在 manifest `backend.settingsRef` 中声明 `settings.schema.json`。
2. 插件扫描时宿主读取 schema，并登记到 `plugin_config_schemas`。
3. 后台插件管理页读取 `/api/admin/plugins/config-schemas`，按 schema 渲染通用配置表单。
4. 后台读取 `/api/admin/plugin/{pluginID}/config-values` 展示当前配置。
5. 后台保存 `/api/admin/plugin/{pluginID}/config-values`，普通字段写入 `value_json`，敏感字段写入 `secret_json`。
6. 插件运行时通过宿主传入的配置上下文或宿主标准查询能力读取配置，不得自行读取 release 目录外部配置文件。
7. 插件返回 `public_config` 时只能返回公开字段，例如前端 `site_key`、主题、尺寸，不能返回 `secret_key`、token、私钥等敏感值。
8. 配置唯一事实源是 `plugin_config_values`，宿主不得在多个位置并行维护同一业务配置。

## 6. 数据与数据库

插件需要持久化业务数据时，必须二选一：

1. 使用 `KEYSWIFT_PLUGIN_DATA_DIR` 保存插件私有文件状态。
2. 在 manifest `database` 中声明插件表，由宿主登记、校验和治理。

禁止事项：

1. 禁止插件直接复用宿主旧表字段作为兼容层。
2. 禁止插件直接读写未声明的宿主业务表。
3. 禁止把插件业务数据散落到宿主源码目录。

## 7. 构建脚本要求

每个插件必须提供：

1. `build.ps1`
2. `build.sh`

`build.ps1` 必须支持：

```text
-Windows
-Linux
-Mac
-All
-Arm
-X64
-Clean
-SkipPause
-SkipFrontend      # 跳过前端构建，仅构建后端二进制（用于无 Node/离线环境）
```

`build.sh` 必须支持：

```text
--windows
--linux
--mac
--all
--arm
--x64
--clean
--skip-frontend    # 跳过前端构建，仅构建后端二进制（用于无 Node/离线环境）
```

构建脚本必须完成：

1. 编译后端二进制。
2. 构建或同步前端产物（如果插件有前端），具体流程见下方"前端构建处理"。
3. 将产物写入 `releases/<version>/`。
4. 生成或更新 `checksums.json`。
5. 输出 `.ksplugin.zip` 安装包。
6. 安装包内部包含 `<plugin-id>/releases/<version>/...`，可直接解压到 `Program/plugins/`。

### 前端构建处理

插件前端为独立 Next.js 工程（`frontend/`），与主程序同栈。构建脚本对前端的完整处理流程如下，默认执行，`-SkipFrontend` / `--skip-frontend` 时跳过：

1. **前置检查**：确认 `frontend/package.json` 存在；确认 `npm` 可用。缺失则构建失败并提示，不得静默继续。
2. **安装依赖**：在 `frontend/` 目录执行 `npm ci`（有 `package-lock.json` 时）或 `npm install`（首次生成 lockfile）。依赖版本必须与主程序 `Program/web/package.json` 一致并跟随升级。
3. **静态导出**：执行 `npm run build`（即 `next build`），产物输出到 `frontend/out/`，`output: export` 模式生成 `widget.html` 与 `_next/` 静态资源目录。
4. **产物落位**：清空 `releases/<version>/frontend/`，拷贝：
   - `frontend/out/widget.html` → `releases/<version>/frontend/widget.html`（manifest `iframeEntries[].entry` 与 `extensions.humanVerification.frontendUrl` 均指向此路径）。
   - `frontend/out/_next/` → `releases/<version>/frontend/_next/`（widget.html 引用的 JS/CSS 静态资源）。
   - **不拷贝** Next 默认产物 `404.html`、`_not-found.*`、`widget.txt` 等，避免污染插件包。
5. **资源路径对齐**：`frontend/next.config.js` 必须设置 `assetPrefix` 为宿主 serve 该插件前端的完整前缀 `/api/plugins/<plugin-id>/<version>/frontend`，使产物内 `_next` 资源引用解析到宿主实际 serve 路径，避免 iframe 内 404。插件 id 或版本变更时须同步更新 `assetPrefix`。
6. **跳过前端**：`-SkipFrontend` / `--skip-frontend` 时保留既有 `releases/<version>/frontend/` 产物（若存在），仅构建后端二进制；适用于无 Node 环境的 CI 或离线构建，正式发布前必须补一次完整前端构建。
7. **完整性纳入**：`releases/<version>/frontend/` 下所有文件（`widget.html` 与 `_next/**`）必须写入 `checksums.json`，与二进制、manifest、settings schema 一并校验。

## 8. 安装包格式

安装包内部结构必须是：

```text
<plugin-id>/releases/<version>/manifest.json
<plugin-id>/releases/<version>/settings.schema.json
<plugin-id>/releases/<version>/checksums.json
<plugin-id>/releases/<version>/bin/<os>_<arch>/<plugin-binary>
<plugin-id>/releases/<version>/frontend/<bundle-files>
```

如果插件没有任何前端贡献，可以不包含 `frontend/`。

## 9. 人机验证插件补充协议

人机验证插件是通用插件规范下的一类安全插件。

必须声明：

```json
{
  "capabilities": {
    "security": ["human_verification.provider"]
  },
  "extensions": {
    "humanVerification": {
      "providerId": "plugin.id",
      "providerType": "provider_type",
      "displayName": "显示名称",
      "renderMode": "plugin_iframe",
      "frontendUrl": "/api/plugins/plugin.id/1.0.0/frontend/widget.html",
      "frontendHeight": 96,
      "supportedScopes": ["admin_login", "user_login", "user_register"],
      "invokeProtocol": "json-stdio"
    }
  }
}
```

必须支持动作：

1. `public_config`
2. `config_status`
3. `health`
4. `create_challenge`
5. `verify`

## 10. 当前人机验证插件的配置入口结论

`keyswift.image_captcha` 和 `cloudflare.turnstile` 当前不需要独立后台菜单页，但必须有后台配置入口。

当前配置入口定义如下：

1. 配置入口位于后台插件管理详情页的通用配置表单，而不是单独新增后台菜单。
2. 配置表单由插件 `settings.schema.json` 驱动渲染，保存时写入主数据库 `plugin_config_values`。
3. `cloudflare.turnstile` 必须通过该入口配置 `site_key`、`secret_key`、`theme`、`size`、`verify_timeout_ms`、`fail_open` 等字段。
4. `secret_key` 必须作为敏感字段进入 `secret_json`，不得写入前端公开配置、manifest、release 文件或外部配置文件。
5. 插件前端 widget 必须自包含在 `frontend/widget.html`，并在 manifest `frontend.iframeEntries` 与 `extensions.humanVerification.frontendUrl` 中声明。
6. 宿主 `HumanVerificationWidget` 只负责挂载 iframe 和收发标准 `postMessage`，不得硬编码图片验证码、Turnstile 或其他 provider 的具体渲染逻辑。
7. 插件后端必须从宿主主数据库配置链路取得配置，不能加载插件目录外的配置文件，也不能要求管理员手工编辑文件。
8. 插件安装或首次同步后，宿主应先写入默认配置记录；普通字段先占位，敏感字段只在管理员填写后写入，不得把空密钥伪装成已配置。

如果后续要提供更复杂能力，例如 Turnstile 域名检测、密钥连通性测试、统计面板、自定义样式预览，则必须为插件新增 `frontend/` 源码、manifest 前端贡献和前端构建产物。此时仍必须保留 `settings.schema.json` 作为后端配置契约，自定义页面只能作为更复杂的交互入口。

## 11. 禁止事项

1. 禁止新增正式 `host-adapter` 插件。
2. 禁止只新增 `manifest.json` 和 `settings.schema.json`，却没有源码和构建脚本。
3. 禁止把插件专属逻辑写入 `Program/internal/`。
4. 禁止插件前端页面依赖宿主硬编码入口。
5. 禁止让最终用户看到“插件化阶段”“后续接入”“内部适配器”等开发过程语义。

## 12. 验收清单

插件交付前必须验证：

1. `build.ps1 -Windows -SkipPause` 可生成安装包。
2. `build.sh --linux` 在 Linux 环境可生成安装包。
3. `checksums.json` 覆盖 manifest、配置 schema、二进制和前端产物。
4. manifest 不包含正式 `host-adapter` 运行形态。
5. 插件声明的前端页面、菜单、配置页和后端能力与实际产物一致。
6. 宿主 `go test ./...` 通过。
7. 前端 `npm run build` 通过。
