<div align="center">
  <h1>meowbridge</h1>
  <p>把标准 Webhook 稳定转发到 MeoW 推送服务，提供独立 token 入口和智能解析能力</p>
</div>

<p align="center">
  <a href="https://github.com/sunnyhmz7010/meowbridge/releases"><img src="https://img.shields.io/github/v/release/sunnyhmz7010/meowbridge?label=Release&color=3b82f6" alt="Release" /></a>
  <a href="https://github.com/sunnyhmz7010/meowbridge/blob/main/LICENSE"><img src="https://img.shields.io/github/license/sunnyhmz7010/meowbridge?color=10b981" alt="License" /></a>
  <a href="https://github.com/sunnyhmz7010/meowbridge/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/sunnyhmz7010/meowbridge/ci.yml?branch=main&label=CI" alt="CI" /></a>
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

- 支持常见服务解析器：GitHub PR、GitHub Actions、GitHub、Jenkins、Grafana 告警、Prometheus Alertmanager、Zabbix、Gotify、Emby，并可为单个入口选择 GitHub 简化 Push 等预设或自定义映射
- 自动提取标题、正文、链接和消息类型
- 通用 fallback 按 `title`/`msg`/`url`/`imgUrl` 等字段名匹配

### 🧩 可配置解析器

- 每个推送入口可选择自动解析、预设解析器或自定义字段映射
- 自定义映射支持简单 JSONPath，例如 `$.hook.url`、`$.alerts[0].labels.alertname`
- 管理后台支持粘贴实际 payload 预览解析结果

### 📊 配置优先级

- 请求参数 / 映射值优先于接口预设，接口预设优先于 MeoW 默认值
- 支持 `msgType` / `url` / `imgUrl` / `htmlHeight` 等 query 参数实时覆盖

### 🛠 管理后台

- 浏览器中管理推送入口、推送日志和日志保留设置
- 管理员密码登录、JWT Bearer 鉴权
- 推送入口支持创建、启停、删除、重置 token

## ⚡ 快速开始

### 📋 前置要求

- Go 1.26 或更高版本
- 管理员密码在首次打开管理后台时设置
- 可选 JWT Secret；未提供时会自动生成并持久化到 SQLite

### 📦 安装与运行

```powershell
go run ./cmd/meowbridge
```

启动后访问 `http://localhost:8080/admin/`，按页面提示设置管理员密码。

默认监听 `8080`。如需修改端口，请设置不带冒号的 `HTTP_PORT`：

```powershell
$env:HTTP_PORT="9090"
```

### 🐳 Docker 镜像

Linux / macOS / Git Bash：

```bash
docker run -d --name meowbridge -p 8080:8080 -v meowbridge-data:/data ghcr.io/sunnyhmz7010/meowbridge:latest
```

PowerShell：

```powershell
docker run -d --name meowbridge -p 8080:8080 -v meowbridge-data:/data ghcr.io/sunnyhmz7010/meowbridge:latest
```

启动后访问 `http://localhost:8080/admin/`，首次打开页面会要求设置管理员密码。`JWT_SECRET` 可以省略；省略时服务会自动生成并保存到 SQLite。显式传入更适合迁移实例或多次重建容器时保持登录密钥稳定。

## 📖 使用说明

启动后访问 `http://localhost:8080/admin/` 进入管理后台。首次打开页面时设置管理员密码；如果已经存在管理员，则直接进入登录页。

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

### 📌 配置 Webhook 解析器

在管理后台编辑 Endpoint，进入「Webhook 解析」区域：

- `自动解析`：使用内置解析器链，未识别时发送完整 JSON。
- `预设解析器`：适合已知来源，例如 GitHub 简化 Push、GitHub Actions、Grafana、Prometheus。
- `自定义字段映射`：按行配置字面量或 JSONPath，组合出标题、消息、URL 和消息类型。

例如下面这种精简 GitHub payload：

```json
{
  "event_type": "push",
  "hook": {
    "url": "https://github.com/sunnyhmz7010/meowbridge"
  },
  "ref": "refs/heads/main",
  "service": "github",
  "sourcecontrol": "github"
}
```

可直接选择 `GitHub 简化 Push` 预设，消息会提取仓库 URL、分支、事件和来源，而不是把完整 JSON 原样推送。

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
| `github_push_minimal` | GitHub 简化 Push | markdown | 固定 `GitHub Push` | `$.hook.url` / `$.ref` / `$.event_type` |
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

- Go 1.26+
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
├── Dockerfile          # 容器镜像构建配置
├── SECURITY.md
├── LICENSE
└── README.md
```

## 👨‍💻 本地开发

### 🧰 环境

- Go 1.26+
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

<a href="https://www.star-history.com/?repos=sunnyhmz7010%2Fmeowbridge&type=date&legend=top-left">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=sunnyhmz7010/meowbridge&type=date&theme=dark&legend=top-left&sealed_token=8CLTHqfYRDrHubgQ1gdfE4NnD72hlouhH3kw694XvS6S2Yi1bvynYqQgX8QI9fPWl8W3bQ-k11lFdb6rtX6u9Uqf7TtpQ_iVp80sQAh3bMljWd1AzcMRC-Z6b4hKn4quWfNRDyKNCZPrSPGX-1CEQT_siwGIhENaEgyEVYrZSKPbX2Wo6IBwsROeHBg2" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=sunnyhmz7010/meowbridge&type=date&legend=top-left&sealed_token=8CLTHqfYRDrHubgQ1gdfE4NnD72hlouhH3kw694XvS6S2Yi1bvynYqQgX8QI9fPWl8W3bQ-k11lFdb6rtX6u9Uqf7TtpQ_iVp80sQAh3bMljWd1AzcMRC-Z6b4hKn4quWfNRDyKNCZPrSPGX-1CEQT_siwGIhENaEgyEVYrZSKPbX2Wo6IBwsROeHBg2" />
   <img alt="Star History Chart" src="https://api.star-history.com/chart?repos=sunnyhmz7010/meowbridge&type=date&legend=top-left&sealed_token=8CLTHqfYRDrHubgQ1gdfE4NnD72hlouhH3kw694XvS6S2Yi1bvynYqQgX8QI9fPWl8W3bQ-k11lFdb6rtX6u9Uqf7TtpQ_iVp80sQAh3bMljWd1AzcMRC-Z6b4hKn4quWfNRDyKNCZPrSPGX-1CEQT_siwGIhENaEgyEVYrZSKPbX2Wo6IBwsROeHBg2" />
 </picture>
</a>

<div align="center">
  <sub>Built with ❤️ by Sunny</sub>
</div>
