# 图片人机验证插件前端

本目录是插件的 **Next.js + React + TypeScript + Tailwind** 前端工程，与主程序 `Program/web` 同栈，遵循《插件开发手册》"前端架构选型（强制与主程序同栈）"要求。

## 工程结构

```text
frontend/
├── package.json          # 依赖版本与主程序一致，跟随主程序升级
├── next.config.js        # output: export 静态导出
├── tsconfig.json
├── tailwind.config.ts    # 设计 token 与主程序一致
├── postcss.config.js
└── src/
    ├── app/
    │   ├── layout.tsx        # 根布局，data-theme 由宿主消息驱动
    │   ├── globals.css       # 主题 token（与主程序 globals.css 一致）
    │   └── widget/page.tsx   # 单页路由 widget，静态导出为 widget.html
    ├── components/
    │   └── ImageCaptchaWidget.tsx   # 图片验证码 React 组件
    └── lib/
        └── protocol.ts       # 宿主 ↔ iframe 标准消息协议类型
```

## 构建产物

- `npm run build` 执行 `next build` 静态导出，产物在 `frontend/out/`。
- 构建脚本（`build.ps1` / `build.sh`）将 `out/widget.html` 与 `out/_next/` 静态资源拷贝到 `releases/<version>/frontend/`，最终由宿主 `/api/plugins/<id>/<ver>/frontend/widget.html` 提供 iframe 加载。
- manifest `frontend.iframeEntries[].entry` 与 `extensions.humanVerification.frontendUrl` 均指向 `frontend/widget.html`，与构建产物路径一致，无需修改 manifest。

## 与宿主的交互

1. 宿主 `HumanVerificationWidget` 挂载 iframe 并通过 `postMessage` 下发 `init` / `theme` 消息。
2. 组件接收 `init` 后调用 `/api/human-verification/challenge` 创建图片挑战并展示。
3. 用户输入答案后通过 `change` 消息回传 payload；宿主在登录/注册提交时携带该校验结果。
4. 主题切换时宿主下发 `theme` 消息，组件切换 `data-theme`，不刷新验证码。

## 开发调试

```bash
cd frontend
npm ci
npm run dev      # 本地开发，http://localhost:3000/widget
npm run build    # 静态导出到 out/
npm run typecheck
```

## 边界说明

本插件仅需 iframe widget，不提供独立后台菜单页。后台配置入口为宿主插件管理详情页的通用配置表单（由 `settings.schema.json` 驱动）。若后续需要图片样式预览、风险统计或独立配置页面，须在本工程内新增页面/路由，并在 manifest `frontend` 中声明 `pages`、`menus` 或 `settingsPages`。
