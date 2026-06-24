# KeySwift

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/React-19.2.3-61DAFB?style=flat-square&logo=react" alt="React">
  <img src="https://img.shields.io/badge/Next.js-16.1.1-000000?style=flat-square&logo=next.js" alt="Next.js">
  <img src="https://img.shields.io/badge/License-GPL%20v3-blue?style=flat-square" alt="License">
</p>

> [!CAUTION]
> **⚠️ Warning: This project is still in the development stage (Alpha) and not yet fully stable.**
>
> **Do not use it in production**, otherwise it may cause data loss, financial risk, or security vulnerabilities. For learning, research, and testing only.
> The system has been refactored into a "core platform + plugin extension" architecture. Some capabilities are delivered via plugins. Features listed in this README are subject to the actual implementation in code and installed plugins.

<p align="center">
  A plugin-driven card key sales platform: the host retains the transaction main chain and governance capabilities, while payments, card delivery, customer service, marketing, and other extensions are delivered through plugins.
</p>

<p align="center">
  <a href="#architecture">Architecture</a> •
  <a href="#features">Features</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#tech-stack">Tech Stack</a> •
  <a href="#configuration">Configuration</a> •
  <a href="#project-structure">Project Structure</a> •
  <a href="#api">API</a> •
  <a href="#documentation">Documentation</a> •
  <a href="#license">License</a>
</p>

---

## Architecture

KeySwift follows a **"host core + plugin extension"** layered architecture:

- The **host core** retains the transaction main chain and platform governance, so the system runs independently under any plugin combination.
- **Extensions** are entirely delivered via plugins. Third-party payments, card delivery, customer service tickets, marketing coupons, notifications, reconciliation, and theme skins all exist as plugins, installed and enabled on demand.
- **Plugins must not host** the official user, admin, session, permission, order, or balance tables. Core data is always adjudicated and guarded by the host.

| Dimension | Host Core | Plugin-provided |
|-----------|-----------|-----------------|
| Transaction | Product basics, order creation & status adjudication, default balance payment, balance ledger | Third-party payment channels, card delivery methods, product type extensions |
| Platform | User/admin/session/permission, plugin discovery & governance, database declaration & audit | Customer service tickets, marketing coupons, notifications & external integrations, reconciliation, import/export |
| UI | Admin framework & workspace basics | Theme skins, admin workspace & page component extensions |

Plugins declare capabilities that the host registers, and interact with the host via events, tasks, configuration, and controlled APIs. They cannot directly modify host core state.

## Features

### 🧩 Plugin System (Core Capability)
- Full plugin lifecycle: discover → parse → validate → register → approve → database structure handling → start → handshake → run → stop
- Plugin governance: install / enable / disable / uninstall / multi-version management
- manifest protocol: declares capabilities, permissions, database structure, compatibility, and integrity
- Plugin runtime & diagnostics (process mode + JSON stdio protocol)
- Plugin frontend asset hosting (`/api/plugins/:plugin_id/:version/frontend/*`, supports iframe rendering)
- Dynamic category display (category filters come from real declarations of installed plugins)
- Four plugin kinds: `functional`, `integration`, `ui-theme`, `tooling`

### 🛒 Product System (Host Core)
- Product category management
- Multi-image product display & primary image
- Product details (rich text / detail file)
- Manual card key import & batch management (enable / disable / delete)
- Card key inventory statistics
- Product type extensions can be enhanced via plugins

### 🎫 Order System (Host Core)
- Order creation & full lifecycle management
- Balance payment (host default payment method)
- Auto-cancel expired orders with inventory rollback
- Public order query (no login required, by order number + email)
- Third-party payment channels and card delivery methods via payment / functional plugins

### 💳 Payment
- **Balance payment**: built into the host, works out of the box
- **Third-party payment channels** (Alipay / WeChat Pay / PayPal / Stripe / USDT / YiPay, etc.): delivered as **payment plugins**, installed on demand
- Payment password & secondary verification

### 👤 User System (Host Core)
- User registration / login / logout
- Email verification & password recovery
- Two-factor authentication (TOTP)
- User balance system
- Payment password management

### 🛡️ Human Verification (Plugin-based)
- Pluggable human verification mechanism based on plugins
- Covers user login, registration, and admin panel entry
- Switch verification methods by installing different provider plugins

### 🔀 Reverse Proxy
- Built-in reverse proxy configuration, supports forwarding specified paths to backend services

### 🔧 Admin Panel
- Dashboard statistics
- Multi-admin role permissions (RBAC)
- Product / order / user / balance management
- System settings & system management
- Plugin governance UI
- IP whitelist
- Operation log audit
- Database backup

### 🚀 Caching
- Local in-memory cache
- Cache statistics metrics
- Unified cache interface, ready to be replaced via plugins or future implementations

### 🔒 Security
- CSRF protection
- IP whitelist
- Tiered rate limiting
- Security response headers
- bcrypt password hashing
- AES-GCM sensitive data encryption
- Login failure lockout
- Payment password & secondary verification

### 📦 Built-in Official Plugins

| Plugin | Kind | Description |
|--------|------|-------------|
| `keyswift.image_captcha` | integration | Image captcha human verification provider |
| `cloudflare.turnstile` | integration | Cloudflare Turnstile human verification provider |
| `example-diagnostics` | tooling | Plugin diagnostics example, demonstrating runtime & diagnostics |

> Third-party payments, customer service tickets, marketing coupons, notifications, reconciliation, and theme skins will be provided as standalone plugins. See the plugin repository and installed plugins for details.

## Quick Start

### Requirements

- **Go 1.25+** (backend build)
- **Node.js 18+** (frontend build)
- **Supported platforms**: Windows, Linux, macOS (x64 & ARM64)

### 🛠️ Build

This project provides cross-platform build scripts supporting **Windows**, **Linux**, and **macOS**, in both **x64 (AMD64)** and **ARM64** architectures.

#### Windows (PowerShell)

```powershell
# 1. Default build (Windows/amd64)
.\build.ps1

# 2. Embed mode (single-file deployment, recommended)
.\build.ps1 -Embed

# 3. Cross-compile other platforms
.\build.ps1 -Linux               # Build Linux version
.\build.ps1 -Mac                 # Build macOS version
.\build.ps1 -All                 # Build all platforms (Win/Lin/Mac)

# 4. Build ARM architecture (e.g. Surface Pro X, Raspberry Pi, Apple M1/M2)
.\build.ps1 -Arm                 # Build ARM64 for the current platform
.\build.ps1 -All -Arm            # Build ARM64 for all platforms
.\build.ps1 -All -Arm -X64       # Build all architectures for all platforms
```

#### Linux / macOS (Bash)

```bash
# 1. Default build (current system)
./build.sh

# 2. Embed mode (single-file deployment, recommended)
./build.sh --embed

# 3. Cross-compile other platforms
./build.sh --linux               # Build Linux version
./build.sh --mac                 # Build macOS version
./build.sh --win                 # Build Windows version

# 4. Build ARM architecture
./build.sh --arm                 # Build ARM64 architecture
./build.sh --all --arm --x64     # Build all architectures for all platforms
```

### Run

Build artifacts are located in the `dist/` directory.

```bash
# Windows
.\dist\windows_amd64\UserFrontend.exe

# Linux
./dist/linux_amd64/UserFrontend

# macOS
./dist/macos_arm64/UserFrontend
```

### Access URLs

| Page | URL |
|------|-----|
| User Frontend | http://localhost:8080/ |
| Admin Panel | http://localhost:8080/manage |

> The admin panel path suffix (default `manage`) can be changed in system settings.

### Default Account

On first access to the admin panel, the system will guide you to set the administrator password.

- Default username: `admin`
- Password: Set via the initialization wizard on first startup

> ⚠️ **Security Notice**: Please set a strong password with letters, numbers, and special characters!

## Tech Stack

| Component | Technology |
|-----------|------------|
| Backend Framework | Go 1.25 + Gin 1.11 |
| ORM | GORM 1.31 |
| Database | SQLite / MySQL / PostgreSQL |
| Cache | Local in-memory cache |
| Frontend Framework | React 19 + Next.js 16 + TypeScript 5 |
| Styling | Tailwind CSS 3.4 |
| State Management | Zustand 5 |
| Plugin Protocol | Process mode + JSON stdio |
| Authentication | Session + Cookie + TOTP |
| Encryption | bcrypt + AES-GCM |

## Configuration

### Database

The system defaults to SQLite and runs without extra configuration; you can switch to MySQL or PostgreSQL in the initialization wizard.

- **SQLite** (default): data file at `user_config/user_data.db`
- **MySQL**: recommended for production
- **PostgreSQL**: advanced feature support

Database connection info is configured via the initialization wizard on first startup and stored in the config database (`user_config/db-config.db`). When the main business database is unavailable, it automatically falls back to local SQLite.

### Environment Variables

Environment variables override runtime behavior (logging, CORS, rate limiting, cookies, etc.) and are not used for database connection configuration.

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_ENV` | Runtime environment (development / production / testing) | `development` |
| `SECURE_COOKIE` | Whether cookies use the Secure flag | `false` (dev) |
| `COOKIE_DOMAIN` | Cookie scope | - |
| `LOG_LEVEL` | Log level (debug / info / warn / error) | `debug` |
| `LOG_FORMAT` | Log format (json / text) | `text` |
| `LOG_OUTPUT` | Log output (stdout / file / both) | `stdout` |
| `LOG_FILE` | Log file path | `logs/app.log` |
| `ENABLE_DEBUG` | Enable debug mode | `true` (dev) |
| `ENABLE_PPROF` | Enable pprof profiling | `false` |
| `ENABLE_SQL_LOG` | Enable SQL logging | `true` (dev) |
| `ENABLE_REQUEST_LOG` | Enable request logging | `true` |
| `ALLOW_ORIGINS` | Allowed CORS origins (comma-separated) | `*` (dev) |
| `ALLOW_CREDENTIALS` | Whether credentials are allowed | `true` |
| `RATE_LIMIT_ENABLED` | Enable rate limiting | `false` (dev) |
| `MAX_REQUEST_BODY` | Max request body size (bytes) | `10485760` |

## Project Structure

```
KeySwift/
├── cmd/
│   ├── server/              # Entry point
│   └── dbctl/               # Database maintenance tool
├── internal/
│   ├── api/                 # HTTP API handlers & routes
│   ├── cache/               # Cache layer (local in-memory cache + metrics)
│   ├── config/              # Configuration & environment variables
│   ├── dbschema/            # Database schema embedding
│   ├── model/               # Data models
│   ├── plugin/              # Plugin registry & governance
│   ├── repository/          # Data access layer
│   ├── service/             # Business logic
│   ├── static/              # Static asset embedding
│   └── utils/               # Utilities
├── database/                # Built-in database schema & seed data
│   ├── bootstrap/sqlite/
│   └── main/sqlite/
├── plugins/                 # Official built-in plugin sources
│   ├── cloudflare.turnstile/
│   ├── keyswift.image_captcha/
│   └── example-diagnostics/
├── docs/                    # Public technical documentation
│   └── Plugin_Development_Manual_CN/   # Plugin development manual
├── web/                     # Frontend source (React + Next.js, static export)
│   └── src/
│       ├── app/             # Page routes
│       ├── components/      # React components
│       ├── hooks/           # Custom hooks
│       └── lib/             # Utilities, API wrappers
├── build.ps1                # Windows build script
└── build.sh                 # Linux / macOS build script
```

## API

The system provides a RESTful API with the following main modules:

| Module | Path | Description |
|--------|------|-------------|
| Public | `/api/csrf-token`, `/api/auth/config` | CSRF token, public auth config |
| Human Verification | `/api/human-verification/challenge` | Human verification challenge |
| Plugin Frontend | `/api/plugins/:plugin_id/:version/frontend/*` | Plugin frontend assets |
| User | `/api/user/*` | Registration, login, profile, balance, etc. |
| Product | `/api/products`, `/api/product/:id`, `/api/categories` | Product list, detail, categories |
| Order | `/api/order/*`, `/api/order/query` | Order create, detail, cancel, public query |
| Payment | `/api/payment/methods` | Available payment methods (incl. enabled payment plugins) |
| Admin | `/:suffix/*`, `/api/admin/*` | Admin login & management API (RBAC) |
| Health | `/health`, `/api/health` | Service health check |

> Admin APIs are protected by admin authentication and permission checks. See `internal/api/router.go` for permission items.

## Documentation

- 📖 [Plugin Development Manual (Chinese)](docs/Plugin_Development_Manual_CN/README.md) - Plugin boundaries, manifest protocol, database development, host integration, security audit & acceptance

## Deployment

### Binary Deployment (Recommended)

Build with "embed mode" to get a single executable containing frontend assets for simple deployment.

```bash
# 1. Build (Windows example)
.\build.ps1 -Embed

# 2. Deploy
# Upload the generated UserFrontend.exe to the server and run it directly
```

### Data Directories

The following directories are generated next to the executable at runtime. Mind backup and permissions:

| Directory | Purpose |
|-----------|---------|
| `user_config/` | Config database & business database (SQLite mode) |
| `logs/` | Runtime logs |
| `backups/` | Database backups |
| `Product/` | Uploaded resources such as product images |

## FAQ

<details>
<summary>How to reset the admin password?</summary>

Delete the `user_config/db-config.db` file and restart the program to re-enter the initialization wizard.

</details>

<details>
<summary>Which databases are supported?</summary>

- **SQLite** (default): suitable for small deployments, no extra configuration
- **MySQL**: recommended for production
- **PostgreSQL**: advanced feature support

The database type is configured in the initialization wizard on first startup.

</details>

<details>
<summary>How to back up data?</summary>

1. **Automatic backup**: Admin panel → System Settings → Data Backup, configure scheduled automatic backup
2. **Manual backup**: click the "Backup Now" button
3. **Database file**: in SQLite mode, directly copy `user_config/user_data.db`

</details>

<details>
<summary>How to integrate third-party payments or customer service?</summary>

These capabilities are delivered as plugins. In the admin panel → Plugin Governance, install and enable the corresponding plugin (payment plugin, customer service plugin, etc.) and configure it per the plugin manifest. To develop custom plugins, see the [Plugin Development Manual](docs/Plugin_Development_Manual_CN/README.md).

</details>

## Contributing

Issues and Pull Requests are welcome!

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the [GNU General Public License v3.0](LICENSE).

```
KeySwift - Card Key Sales Management System
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

## Acknowledgments

- [Gin](https://github.com/gin-gonic/gin) - High-performance HTTP web framework
- [GORM](https://gorm.io/) - Go ORM library
- [Next.js](https://nextjs.org/) - React framework
- [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS framework
- [Zustand](https://github.com/pmndrs/zustand) - Lightweight state management

---

<p align="center">
  If this project helps you, please give it a ⭐ Star!
</p>
<p align="center">
  Made with ❤️ by CareCheng with 海绵
</p>
