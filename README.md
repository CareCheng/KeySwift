# KeySwift

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/React-19.2.3-61DAFB?style=flat-square&logo=react" alt="React">
  <img src="https://img.shields.io/badge/Next.js-16.1.1-000000?style=flat-square&logo=next.js" alt="Next.js">
  <img src="https://img.shields.io/badge/License-GPL%20v3-blue?style=flat-square" alt="License">
</p>

> [!CAUTION]
> **⚠️ 警告：本项目目前仍处于开发阶段（Alpha），功能尚未完全稳定。**
>
> **严禁在生产环境中使用**，否则可能导致数据丢失、资金风险或安全漏洞。仅供学习、研究和测试使用。
> 系统已重构为「核心平台 + 插件扩展」架构，部分能力以插件形式提供，README 所列功能以代码与已安装插件实际实现为准。

<p align="center">
  一个以插件化为核心设计理念的卡密销售平台：宿主保留交易主链路与治理能力，支付、发卡、客服、营销等扩展能力均通过插件交付。
</p>

<p align="center">
  <a href="#架构理念">架构理念</a> •
  <a href="#功能特性">功能特性</a> •
  <a href="#快速开始">快速开始</a> •
  <a href="#技术栈">技术栈</a> •
  <a href="#配置说明">配置说明</a> •
  <a href="#项目结构">项目结构</a> •
  <a href="#api-接口">API 接口</a> •
  <a href="#文档">文档</a> •
  <a href="#许可证">许可证</a>
</p>

---

## 架构理念

KeySwift 采用 **「宿主核心 + 插件扩展」** 的分层架构：

- **宿主核心**保留交易主链路与平台治理能力，保证系统在任何插件组合下都可独立运行。
- **扩展能力**全部通过插件交付，第三方支付、发卡交付、客服工单、营销优惠券、通知对账、主题皮肤等均以插件形式存在，按需安装启用。
- **插件不得承载**正式用户、管理员、会话、权限、订单、余额等主表，核心数据始终由宿主裁决与守护。

| 维度 | 宿主核心 | 插件承载 |
|------|----------|----------|
| 交易 | 商品基础信息、订单创建与状态裁决、默认余额支付、余额账本 | 第三方支付渠道、发卡与交付方式、商品类型扩展 |
| 平台 | 用户/管理员/会话/权限、插件发现与治理、数据库声明与审计 | 客服工单、营销优惠券、通知与外部平台集成、对账、导入导出 |
| 界面 | 后台框架与工作区基础 | 主题皮肤、后台工作区与页面组件扩展 |

插件通过声明能力（capabilities）由宿主登记，通过事件、任务、配置与受控 API 与宿主联动，不能直接修改宿主核心状态。

## 功能特性

### 🧩 插件体系（核心能力）
- 完整插件生命周期：发现 → 解析 → 校验 → 登记 → 审批 → 数据库结构处理 → 启动 → 握手 → 运行 → 停用
- 插件治理：安装 / 启用 / 禁用 / 卸载 / 多版本管理
- manifest 协议：声明能力、权限、数据库结构、兼容性与完整性校验
- 插件运行时与诊断（进程模式 + JSON stdio 协议）
- 插件前端资源托管（`/api/plugins/:plugin_id/:version/frontend/*`，支持 iframe 渲染）
- 插件分类动态展示（分类筛选项来自已安装插件真实声明）
- 插件四类形态：功能插件 `functional`、集成插件 `integration`、主题插件 `ui-theme`、工具插件 `tooling`

### 🛒 商品系统（宿主核心）
- 商品分类管理
- 商品多图展示与主图设置
- 商品详情（富文本 / 详情文件）
- 手动卡密导入与批量管理（启用 / 禁用 / 删除）
- 卡密库存统计
- 商品类型扩展可通过插件增强

### 🎫 订单系统（宿主核心）
- 订单创建与全生命周期管理
- 余额支付（宿主默认支付方式）
- 订单超时自动取消与库存回滚
- 公开订单查询（无需登录，凭订单号 + 邮箱）
- 第三方支付渠道、发卡与交付方式通过支付 / 功能插件接入

### 💳 支付能力
- **余额支付**：宿主内置，开箱即用
- **第三方支付渠道**（支付宝 / 微信 / PayPal / Stripe / USDT / 易支付等）：以**支付插件**形式提供，按需安装
- 支付密码与二次校验

### 👤 用户系统（宿主核心）
- 用户注册 / 登录 / 登出
- 邮箱验证与找回密码
- 两步验证（TOTP）
- 用户余额系统
- 支付密码管理

### 🛡️ 人机验证（插件化）
- 基于插件的可插拔人机验证机制
- 验证范围覆盖用户登录、注册及管理后台入口
- 通过安装不同 provider 插件切换验证方式

### 🔀 反向代理
- 内置反向代理配置能力，支持将指定路径转发至后端服务

### 🔧 管理后台
- 仪表盘数据统计
- 多管理员角色权限（RBAC）
- 商品 / 订单 / 用户 / 余额管理
- 系统设置与系统管理
- 插件治理界面
- IP 白名单
- 操作日志审计
- 数据库备份

### 🚀 缓存系统
- 本地内存缓存
- 缓存统计指标
- 统一缓存接口，便于通过插件或后续实现替换

### 🔒 安全特性
- CSRF 保护
- IP 白名单
- 分级速率限制
- 安全响应头
- 密码 bcrypt 加密
- 敏感数据 AES-GCM 加密
- 登录失败锁定
- 支付密码与二次验证

### 📦 当前内置官方插件

| 插件 | 类型 | 说明 |
|------|------|------|
| `keyswift.image_captcha` | integration | 图片验证码人机验证 provider |
| `cloudflare.turnstile` | integration | Cloudflare Turnstile 人机验证 provider |
| `example-diagnostics` | tooling | 插件诊断示例，演示插件运行时与诊断能力 |

> 第三方支付、客服工单、营销优惠券、通知对账、主题皮肤等能力将以独立插件形式陆续提供，具体以插件仓库与已安装插件为准。

## 快速开始

### 环境要求

- **Go 1.25+**（后端编译）
- **Node.js 18+**（前端构建）
- **支持平台**：Windows、Linux、macOS（x64 与 ARM64）

### 🛠️ 构建指南

本项目提供跨平台构建脚本，支持 **Windows**、**Linux** 和 **macOS**，以及 **x64 (AMD64)** 和 **ARM64** 架构。

#### Windows (PowerShell)

```powershell
# 1. 默认构建 (Windows/amd64)
.\build.ps1

# 2. 嵌入模式 (单文件部署，推荐)
.\build.ps1 -Embed

# 3. 交叉编译其他平台
.\build.ps1 -Linux               # 构建 Linux 版本
.\build.ps1 -Mac                 # 构建 macOS 版本
.\build.ps1 -All                 # 构建全平台 (Win/Lin/Mac)

# 4. 构建 ARM 架构 (如 Surface Pro X, Raspberry Pi, Apple M1/M2)
.\build.ps1 -Arm                 # 仅构建当前平台的 ARM64 版本
.\build.ps1 -All -Arm            # 构建全平台的 ARM64 版本
.\build.ps1 -All -Arm -X64       # 构建全平台的所有架构版本
```

#### Linux / macOS (Bash)

```bash
# 1. 默认构建 (当前系统)
./build.sh

# 2. 嵌入模式 (单文件部署，推荐)
./build.sh --embed

# 3. 交叉编译其他平台
./build.sh --linux               # 构建 Linux 版本
./build.sh --mac                 # 构建 macOS 版本
./build.sh --win                 # 构建 Windows 版本

# 4. 构建 ARM 架构
./build.sh --arm                 # 构建 ARM64 架构
./build.sh --all --arm --x64     # 构建全平台的所有架构版本
```

### 运行

构建产物位于 `dist/` 目录下。

```bash
# Windows
.\dist\windows_amd64\UserFrontend.exe

# Linux
./dist/linux_amd64/UserFrontend

# macOS
./dist/macos_arm64/UserFrontend
```

### 访问地址

| 页面 | 地址 |
|------|------|
| 用户前台 | http://localhost:8080/ |
| 管理后台 | http://localhost:8080/manage |

> 管理后台路径后缀（默认 `manage`）可在系统配置中修改。

### 默认账户

首次访问管理后台时，系统会引导您设置管理员密码。

- 默认用户名：`admin`
- 密码：首次启动时通过初始化向导设置

> ⚠️ **安全提示**：请设置强密码，建议包含字母、数字和特殊字符！

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端框架 | Go 1.25 + Gin 1.11 |
| ORM | GORM 1.31 |
| 数据库 | SQLite / MySQL / PostgreSQL |
| 缓存 | 本地内存缓存 |
| 前端框架 | React 19 + Next.js 16 + TypeScript 5 |
| 样式 | Tailwind CSS 3.4 |
| 状态管理 | Zustand 5 |
| 插件协议 | 进程模式 + JSON stdio |
| 认证 | Session + Cookie + TOTP |
| 加密 | bcrypt + AES-GCM |

## 配置说明

### 数据库配置

系统默认使用 SQLite，无需额外配置即可启动；亦可在初始化向导中切换为 MySQL 或 PostgreSQL。

- **SQLite**（默认）：数据文件位于 `user_config/user_data.db`
- **MySQL**：推荐用于生产环境
- **PostgreSQL**：高级功能支持

数据库连接信息通过首次启动的初始化向导配置，保存在配置数据库（`user_config/db-config.db`）中，主业务库不可用时自动回退到本地 SQLite。

### 环境变量

环境变量用于覆盖运行时行为（日志、CORS、限流、Cookie 等），不用于数据库连接配置。

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `APP_ENV` | 运行环境（development / production / testing） | `development` |
| `SECURE_COOKIE` | Cookie 是否启用 Secure 标志 | 开发环境 `false` |
| `COOKIE_DOMAIN` | Cookie 作用域 | - |
| `LOG_LEVEL` | 日志级别（debug / info / warn / error） | `debug` |
| `LOG_FORMAT` | 日志格式（json / text） | `text` |
| `LOG_OUTPUT` | 日志输出（stdout / file / both） | `stdout` |
| `LOG_FILE` | 日志文件路径 | `logs/app.log` |
| `ENABLE_DEBUG` | 启用调试模式 | `true`（开发） |
| `ENABLE_PPROF` | 启用 pprof 性能分析 | `false` |
| `ENABLE_SQL_LOG` | 启用 SQL 日志 | `true`（开发） |
| `ENABLE_REQUEST_LOG` | 启用请求日志 | `true` |
| `ALLOW_ORIGINS` | 允许的跨域来源（逗号分隔） | `*`（开发） |
| `ALLOW_CREDENTIALS` | 是否允许携带凭证 | `true` |
| `RATE_LIMIT_ENABLED` | 启用速率限制 | `false`（开发） |
| `MAX_REQUEST_BODY` | 最大请求体大小（字节） | `10485760` |

## 项目结构

```
KeySwift/
├── cmd/
│   ├── server/              # 程序入口
│   └── dbctl/               # 数据库维护工具
├── internal/
│   ├── api/                 # HTTP API 处理层与路由
│   ├── cache/               # 缓存层（本地内存缓存 + 指标）
│   ├── config/              # 配置定义与环境变量
│   ├── dbschema/            # 数据库 Schema 嵌入
│   ├── model/               # 数据模型
│   ├── plugin/              # 插件注册与治理
│   ├── repository/          # 数据访问层
│   ├── service/             # 业务逻辑层
│   ├── static/              # 静态资源嵌入
│   └── utils/               # 工具函数
├── database/                # 内置数据库 Schema 与种子数据
│   ├── bootstrap/sqlite/
│   └── main/sqlite/
├── plugins/                 # 官方内置插件源码
│   ├── cloudflare.turnstile/
│   ├── keyswift.image_captcha/
│   └── example-diagnostics/
├── docs/                    # 公开技术文档
│   └── Plugin_Development_Manual_CN/   # 插件开发手册
├── web/                     # 前端源码 (React + Next.js，静态导出)
│   └── src/
│       ├── app/             # 页面路由
│       ├── components/      # React 组件
│       ├── hooks/           # 自定义 Hooks
│       └── lib/             # 工具库、API 封装
├── build.ps1                # Windows 构建脚本
└── build.sh                 # Linux / macOS 构建脚本
```

## API 接口

系统提供 RESTful API，主要模块如下：

| 模块 | 路径 | 说明 |
|------|------|------|
| 公共 | `/api/csrf-token`、`/api/auth/config` | CSRF 令牌、公开认证配置 |
| 人机验证 | `/api/human-verification/challenge` | 人机验证挑战 |
| 插件前端 | `/api/plugins/:plugin_id/:version/frontend/*` | 插件前端资源 |
| 用户 | `/api/user/*` | 注册、登录、资料、余额等 |
| 商品 | `/api/products`、`/api/product/:id`、`/api/categories` | 商品列表、详情、分类 |
| 订单 | `/api/order/*`、`/api/order/query` | 订单创建、详情、取消、公开查询 |
| 支付 | `/api/payment/methods` | 可用支付方式（含已启用支付插件） |
| 管理后台 | `/:suffix/*`、`/api/admin/*` | 后台登录与管理 API（RBAC） |
| 健康检查 | `/health`、`/api/health` | 服务健康检查 |

> 管理后台 API 受管理员认证与权限校验保护，具体权限项见 `internal/api/router.go`。

## 文档

- 📖 [插件开发手册（中文）](docs/Plugin_Development_Manual_CN/README.md) - 插件边界、清单协议、数据库开发、宿主集成、安全审计与验收

## 部署方式

### 二进制部署（推荐）

使用"嵌入模式"构建后，将获得一个包含前端资源的单一可执行文件，部署简单。

```bash
# 1. 构建 (以 Windows 为例)
.\build.ps1 -Embed

# 2. 部署
# 将生成的 UserFrontend.exe 上传到服务器即可直接运行
```

### 数据目录

运行时会在程序同目录生成以下目录，请注意备份与权限：

| 目录 | 用途 |
|------|------|
| `user_config/` | 配置数据库与业务数据库（SQLite 模式） |
| `logs/` | 运行日志 |
| `backups/` | 数据库备份 |
| `Product/` | 商品图片等上传资源 |

## 常见问题

<details>
<summary>如何重置管理员密码？</summary>

删除 `user_config/db-config.db` 文件，重启程序后会重新进入初始化向导。

</details>

<details>
<summary>支持哪些数据库？</summary>

- **SQLite**（默认）：适合小型部署，无需额外配置
- **MySQL**：推荐用于生产环境
- **PostgreSQL**：高级功能支持

数据库类型在首次启动的初始化向导中配置。

</details>

<details>
<summary>如何备份数据？</summary>

1. **自动备份**：管理后台 → 系统设置 → 数据备份，可配置定时自动备份
2. **手动备份**：点击"立即备份"按钮
3. **数据库文件**：SQLite 模式下，直接复制 `user_config/user_data.db`

</details>

<details>
<summary>如何接入第三方支付或客服等功能？</summary>

这些能力均以插件形式提供。在管理后台 → 插件治理中安装并启用对应插件（支付插件、客服插件等），按插件 manifest 声明完成配置即可。开发自定义插件请参阅 [插件开发手册](docs/Plugin_Development_Manual_CN/README.md)。

</details>

## 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

## 许可证

本项目采用 [GNU General Public License v3.0](LICENSE) 开源许可证。

```
KeySwift - 卡密销售管理系统
Copyright (C) 2025


This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
```

## 致谢

- [Gin](https://github.com/gin-gonic/gin) - 高性能 HTTP Web 框架
- [GORM](https://gorm.io/) - Go 语言 ORM 库
- [Next.js](https://nextjs.org/) - React 全栈框架
- [Tailwind CSS](https://tailwindcss.com/) - 实用优先的 CSS 框架
- [Zustand](https://github.com/pmndrs/zustand) - 轻量级状态管理

---

<p align="center">
  如果这个项目对你有帮助，请给一个 ⭐ Star！
</p>
<p align="center">
  Made with ❤️ by CareCheng with 海绵
</p>
