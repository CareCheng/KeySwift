# KeySwift 插件开发手册

分析基线时间：2026-06-20

本手册面向 KeySwift 独立插件开发者、宿主接口维护者和后台实施人员，说明插件包结构、manifest 协议、数据库声明、宿主对接、安全审计和验收流程。

当前项目处于开发阶段，插件和宿主只面向最新协议前进，不做旧协议兼容。

## 目录

| 章节 | 文件 | 说明 |
| --- | --- | --- |
| 01 | [插件开发总览](./01-overview.md) | 插件边界、目录结构、生命周期、宿主职责 |
| 02 | [manifest 协议](./02-manifest.md) | 插件包声明字段、能力声明、前后端入口 |
| 03 | [数据库开发规范](./03-database-development.md) | 插件数据库声明、命名、字段、索引、关系、结构执行 |
| 04 | [宿主对接规范](./04-host-integration.md) | 权限、菜单、页面、配置、事件、任务、事实回传 |
| 05 | [安全与审计规范](./05-security-and-audit.md) | 密钥、敏感字段、高危权限、审计、完整性校验 |
| 06 | [开发调试与验收](./06-development-and-acceptance.md) | 本地开发、打包、检查、提交和验收清单 |
| 07 | [独立式插件通用规范](./06-independent-human-verification-plugins.md) | 所有插件的独立源码、前后端边界、配置表单、构建脚本、安装包格式和验收清单 |

说明：`06-independent-human-verification-plugins.md` 是历史文件名，当前内容已经升级为所有独立式插件的通用规范，并保留人机验证插件专项补充协议。

## 先读结论

| 主题 | 规则 |
| --- | --- |
| 插件形态 | 独立插件包，宿主扫描 `plugins/<plugin_id>/releases/<version>/manifest.json` |
| 数据库 | 插件只能声明自己的表，由宿主登记、校验、显式创建和审计 |
| 数据访问 | 插件不得直接获取主库 DSN，后续通过宿主存储网关或受控接口访问 |
| 权限 | 插件权限必须在 manifest 中声明，并进入宿主权限体系 |
| 页面 | 插件页面、菜单、配置页通过 manifest 挂载到宿主后台 |
| 分类 | 插件分类写入 `identity.categories`，后台插件管理页只根据当前已安装插件真实声明的分类动态生成筛选项 |
| 前端架构 | 插件前端强制与主程序同栈：Next.js + React + TypeScript + Tailwind + Zustand 静态导出，不得用其他框架或裸 HTML 替代；版本须跟随主程序升级保持一致，不兼容即停用；build 脚本负责 npm 安装 + next build 静态导出 + 产物落位 + checksums；iframe 插件经"主题 token 字典" postMessage 机制适配宿主任意主题 |
| 支付 | 宿主默认只保留余额支付，其他支付方式应作为支付插件接入 |
| 商品类型 | 基础商品保留在宿主，扩展商品类型由插件声明能力与数据表 |
| 主题 | 主题类插件通过 `ui` 声明 token、组件覆盖和布局皮肤 |

## 推荐阅读顺序

1. 开发新插件：先看 `01-overview.md`，再看 `02-manifest.md` 和 `03-database-development.md`；插件分类规范见 `01-overview.md` 的“插件分类”和 `02-manifest.md` 的“身份与分类声明”。
2. 开发支付、发卡、商品类型插件：重点阅读 `03-database-development.md` 和 `04-host-integration.md`。
3. 开发主题插件：重点阅读 `02-manifest.md` 中的 `ui` 声明和 `04-host-integration.md` 中的前端挂载。
4. 开发任意独立插件：重点阅读 `06-independent-human-verification-plugins.md`；人机验证插件还需关注其中的专项补充协议。
5. 发布插件前：按 `06-development-and-acceptance.md` 完成验收。

## 对应代码位置

| 能力 | 代码位置 |
| --- | --- |
| 插件协议结构 | `Program/internal/plugin/types.go` |
| 插件发现与 manifest 校验 | `Program/internal/plugin/discovery.go` |
| 插件注册中心 | `Program/internal/plugin/registry.go` |
| 插件治理模型 | `Program/internal/model/plugin.go` |
| 插件仓储 | `Program/internal/repository/repository.go` |
| 插件服务 | `Program/internal/service/plugin_service.go` |
| 插件后台 API | `Program/internal/api/plugin_handler.go` |
| 数据库显式构建 | `Program/cmd/dbctl/main.go` |
| 数据库 schema | `Program/database` |

[下一章：插件开发总览](./01-overview.md)
