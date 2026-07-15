<div align="center">
  <h1>meowbridge</h1>
  <p>把标准 Webhook 稳定转发到 MeoW 推送。</p>
</div>

<p align="center">
  <img src="https://img.shields.io/github/v/tag/sunnyhmz7010/meowbridge?label=Release&color=3b82f6" alt="Release" />
  <img src="https://img.shields.io/github/license/sunnyhmz7010/meowbridge?color=10b981" alt="License" />
  <a href="https://github.com/sunnyhmz7010/meowbridge/actions/workflows/ci.yaml"><img src="https://img.shields.io/github/actions/workflow/status/sunnyhmz7010/meowbridge/ci.yaml?branch=main&label=CI" alt="CI" /></a>
</p>
<p align="center">
  <a href="https://github.com/sunnyhmz7010/meowbridge/issues">反馈问题</a> · <a href="https://github.com/sunnyhmz7010/meowbridge/archive/refs/heads/main.zip">下载源码</a>
</p>

---

## ✨ 为什么做这个服务

MeoW Push 仅支持一个昵称作为唯一标识，多场景共用时容易混淆，且昵称是公开链接的一部分，一旦泄露牵连全部服务对象。meowbridge 提供独立 token 入口，把 GitHub、Grafana、Prometheus、Jenkins、Zabbix、Gotify、Emby 等常见 Webhook 自动解析并转发到指定 MeoW 昵称。

## 🚀 核心能力

### 🌐 通用 Webhook 入口

- 外部服务填写 `/webhook/{token}` 即可推送
- 纯文本请求直接以文本内容推送
- 未识别 payload 格式化完整 JSON 兜底，优先保证消息不丢

### 🧠 内置常见服务解析器

- 支持 9 种预设解析器：GitHub PR、GitHub Actions、GitHub、Jenkins、Grafana 告警、Prometheus Alertmanager、Zabbix、Gotify、Emby
- 自动提取标题、正文、链接和消息类型
- 通用 fallback 按 `title`/`msg`/`url`/`imgUrl` 等字段名匹配

### 📊 配置优先级

- 请求参数 / 映射值优先于接口预设，接口预设优先于 MeoW 默认值
- 支持 `msgType` / `url` / `imgUrl` / `htmlHeight` 等 query 参数实时覆盖

### 🛠 管理后台

- 浏览器中管理推送入口、推送日志和全局设置
- 管理员密码登录、JWT Bearer 鉴权
- 推送入口支持创建、启停、删除、重置 token

## ⚡ 快速开始

### 📋 运行要求

- Go 1.23 或更高版本
- 可访问的 MeoW API Base URL
- 用于管理后台的管理员密码和 JWT Secret

### 🚀 本地运行

```powershell
$env:ADMIN_PASSWORD="change-me"
$env:JWT_SECRET="replace-with-long-random-secret"
$env:MEOW_API_BASE_URL="https://api.chuckfang.com"
go run ./cmd/meowbridge
```

### 📦 Docker Compose

首次部署前请修改 `compose.yaml` 中的 `ADMIN_PASSWORD`、`JWT_SECRET` 和 `MEOW_API_BASE_URL`。

```bash
docker compose up -d
```

默认 Compose 会监听 `8080`，并将 SQLite 数据持久化到 `meowbridge-data` volume。

### 🐳 Docker 镜像

```bash
docker run -d \
  --name meowbridge \
  -p 8080:8080 \
  -v meowbridge-data:/data \
  -e ADMIN_PASSWORD=change-me \
  -e JWT_SECRET="replace-with-long-random-secret" \
  -e MEOW_API_BASE_URL="https://api.chuckfang.com" \
  ghcr.io/sunnyhmz7010/meowbridge:latest
```

## 📖 使用说明

启动后访问 `http://localhost:8080/admin/` 进入管理后台。首次启动使用 `ADMIN_PASSWORD` 环境变量设置管理员密码。

创建推送入口后，将生成的地址填入外部服务：

```text
https://your-domain.example/webhook/{token}
```

支持 query 参数覆盖：

| 参数 | 说明 |
|------|------|
| `msgType` | 覆盖消息类型：text / html / markdown |
| `url` | 覆盖跳转链接 |
| `imgUrl` | 覆盖通知图标 URL |

推送示例：

```powershell
Invoke-RestMethod -Method Post -Uri "http://localhost:8080/webhook/{token}" -ContentType "text/plain" -Body "hello meowbridge"
```

```powershell
Invoke-RestMethod -Method Post -Uri "http://localhost:8080/webhook/{token}" -ContentType "application/json" -Body '{"title":"Build","message":"passed"}'
```

## 📸 截图预览

### 推送入口管理

列表页展示所有入口的 ID、名称、Token、当前状态。可点击启停开关切换启用/禁用状态。

### 推送日志

列表页展示最近 200 条推送记录，显示来源服务、解析结果、推送状态和错误信息。

### 全局设置

配置 MeoW API 地址、日志保留天数，以及修改管理员密码。

## 🧠 功能细节

### 消息合并优先级

```
P1 — 请求参数 / Webhook 映射（最高）
P2 — 推送接口预设值
P3 — MeoW 默认值（最低）
```

| 参数 | P1 请求/映射 | P2 接口预设 | P3 MeoW 默认 |
|------|------------|------------|------------|
| `meow_nickname` | — | ✅ 创建后不可修改 | — |
| `title` | ✅ | ✅ | `"Meow"` |
| `msg` | ✅ 必须提供 | — | —（不允许默认） |
| `msgType` | ✅ | ✅ | `"text"` |
| `htmlHeight` | ✅ (query) | ✅ | `200` |
| `url` | ✅ | ✅ | 不传 |
| `imgUrl` | ✅ | ✅ | 不传 |

### 预设解析器

| 来源 | 展示名称 | 推荐 msgType | title 映射 | msg 映射 |
|------|---------|-------------|-----------|---------|
| `github_pr` | GitHub PullRequest | markdown | `$.pull_request.title` | `$.pull_request.body` |
| `github_action` | GitHub Actions | markdown | `$.workflow_run.event` | `$.workflow_run.head_commit.message` |
| `github` | GitHub Webhook | text | `$.action` | `$.repository.full_name` |
| `jenkins` | Jenkins | text | `$.project.name` | `$.build.full_display_url` |
| `grafana` | Grafana 告警 | markdown | `$.alerts[0].labels.alertname` | `$.alerts[0].annotations.message` |
| `prometheus` | Prometheus AlertManager | markdown | `$.alerts[0].labels.alertname` | `$.alerts[0].annotations.description` |
| `zabbix` | Zabbix | markdown | `$.trigger.description` | `$.event.description` |
| `gotify` | Gotify | markdown | `$.title` | `$.message` |
| `emby` | Emby | text | `$.Title` | `$.Description` |

## 🧱 技术栈

- Go 1.23+
- chi
- SQLite
- modernc.org/sqlite
- JWT Bearer
- Vue 3
- Vite
- TypeScript
- Docker

## 🗂️ 项目结构

```text
meowbridge/
├── cmd/
│   └── meowbridge/     # 服务入口与 HTTP Server 配置
├── internal/
│   ├── auth/           # 管理员密码与 JWT 鉴权
│   ├── config/         # 环境变量配置
│   ├── httpapi/        # Webhook、管理 API 与路由
│   ├── meow/           # MeoW 推送客户端
│   ├── respond/        # 统一 HTTP 响应
│   ├── store/          # SQLite 数据访问
│   ├── token/          # Webhook token 生成
│   ├── webhook/        # Payload 解析与消息合并
│   └── webui/          # 管理后台内嵌静态资源
├── web/                # Vue 3 管理后台源码
├── compose.yaml        # Docker Compose 单服务部署示例
├── Dockerfile          # 容器镜像构建配置
├── SECURITY.md
├── LICENSE
└── README.md
```

## 👨‍💻 本地开发

### 🧰 环境

- Go 1.23+
- Node.js 20+ 与 npm

### 🔨 命令

```bash
go test ./...
cd web
npm install
npm run test
npm run type-check
npm run build
```

## 🔐 安全报告

如发现安全问题，请不要公开披露细节。请优先参考仓库中的 [SECURITY.md](./SECURITY.md) 提交安全报告。

## 📄 许可证

本项目基于 [GPL-3.0](./LICENSE) 开源。

## ⭐ 星标历史

  [![Star History Chart](https://api.star-history.com/svg?repos=sunnyhmz7010/meowbridge)](https://star-history.com/#sunnyhmz7010/meowbridge)

Built with ❤️ by Sunny
