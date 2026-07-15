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

## ✨ 为什么做这个应用

MeoW Push 只有昵称作为入口标识，多场景共用时容易混淆，也不利于隔离泄露风险。meowbridge 提供独立 token 入口，把 Webhook 自动解析并转发到指定 MeoW 。

## 🚀 核心能力

- 通用 Webhook 入口：外部服务填写 `/webhook/{token}` 即可推送。
- 内置常见服务解析器：自动提取标题、正文、链接和消息类型。
- 未识别 payload 兜底：格式化完整 JSON 推送，优先保证消息不丢。
- 管理 API：创建、启停、删除、重置推送入口。
- 浏览器管理后台：登录后管理推送入口、日志和 MeoW 设置。
- 推送日志：记录原始 payload、解析结果、MeoW 响应和失败原因。

## ⚡ 快速开始

### 前置要求

- Go 1.23 或更高版本
- 可访问的 MeoW API Base URL
- 用于管理后台的管理员密码和 JWT Secret

### 安装与运行

```powershell
$env:ADMIN_PASSWORD="change-me"
$env:JWT_SECRET="replace-with-long-random-secret"
$env:MEOW_API_BASE_URL="https://api.chuckfang.com"
go run ./cmd/meowbridge
```

### Docker Compose

首次部署前请修改 `compose.yaml` 中的 `ADMIN_PASSWORD`、`JWT_SECRET` 和 `MEOW_API_BASE_URL`。

```bash
docker compose up -d
```

默认 Compose 会监听 `8080`，并将 SQLite 数据持久化到 `meowbridge-data` volume。

### Docker 镜像

```bash
docker run -d \
  --name meowbridge \
  -p 8080:8080 \
  -v meowbridge-data:/data \
  -e ADMIN_PASSWORD=change-me \
  -e JWT_SECRET=replace-with-long-random-secret \
  -e MEOW_API_BASE_URL=https://api.chuckfang.com \
  ghcr.io/sunnyhmz7010/meowbridge:latest
```

### Release 下载

版本发布页提供 Linux `amd64` 和 `arm64` 二进制包，以及 `checksums.txt` 校验文件。下载后解压并按环境变量说明启动即可。

## 📖 使用说明

启动服务后访问 `http://localhost:8080/admin/` 进入管理后台。首次启动使用 `ADMIN_PASSWORD` 环境变量设置管理员密码。

创建 endpoint 后，将生成的 Webhook 地址填入外部服务：

```text
https://your-domain.example/webhook/{token}
```

纯文本推送：

```powershell
Invoke-RestMethod -Method Post -Uri "http://localhost:8080/webhook/{token}" -ContentType "text/plain" -Body "hello meowbridge"
```

JSON 推送：

```powershell
Invoke-RestMethod -Method Post -Uri "http://localhost:8080/webhook/{token}" -ContentType "application/json" -Body '{"title":"Build","message":"passed"}'
```

## 🧠 功能细节

Webhook 请求会按解析器链处理：GitHub Pull Request、GitHub Actions、GitHub、Jenkins、Grafana、Prometheus Alertmanager、Zabbix、Gotify、Emby、Generic、Fallback。字段优先级为 query 覆盖、解析器输出、endpoint 默认值、MeoW 默认值。

管理后台通过 `/admin/` 访问，前端资源内嵌在 Go 二进制中。生产运行时不需要单独启动前端服务。

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
- GitHub Actions

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
├── web/                # Vue 管理后台源码
├── compose.yaml        # Docker Compose 单服务部署示例
├── Dockerfile          # 容器镜像构建配置
├── SECURITY.md
├── LICENSE
└── README.md
```

## 👨‍💻 本地开发

### 环境

- Go 1.23+
- Node.js 20+ 与 npm

### 命令

```bash
go test ./...
cd web
npm install
npm run test
npm run type-check
npm run build
```

## 🔐 安全报告

如果发现安全问题，请不要公开披露细节。请优先参考仓库中的 [SECURITY.md](./SECURITY.md) 提交安全报告。

## 📄 许可证

本项目基于 [GPL-3.0](./LICENSE) 开源。

## ⭐ 星标历史

[![Star History Chart](https://api.star-history.com/svg?repos=sunnyhmz7010/meowbridge)](https://star-history.com/#sunnyhmz7010/meowbridge)

<div align="center">
  <sub>Built with ❤️ by Sunny</sub>
</div>
