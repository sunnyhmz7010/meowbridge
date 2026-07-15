# AGENTS.md

本文件是 meowbridge 仓库的项目级协作说明。它只记录开发约定、架构边界和验证方式；产品需求、实施规格和阶段计划应放在 `docs/superpowers/` 下。

## 项目定位

meowbridge 是一个 Go 后端服务，用于接收标准 Webhook 请求并转发到 MeoW Push。外部服务只需要在 Webhook URL 中填写本服务提供的 `/webhook/{token}` 地址，不要求对方为 meowbridge 做定制适配。

当前核心能力：

- 管理员登录与 JWT Bearer 鉴权。
- 推送 endpoint 的创建、启停、删除、重置 token。
- 标准 Webhook 入口：`POST /webhook/{token}`。
- 常见服务 payload 自动解析；未知 JSON 使用格式化兜底消息。
- SQLite 持久化 endpoint、设置和推送日志。
- 同步调用 MeoW API 并记录响应结果。

## 技术栈

- Go 1.23+
- chi
- modernc.org/sqlite
- JWT Bearer
- SQLite

避免新增运行时依赖。确需新增依赖时，先说明必要性、替代方案和维护成本。

## 目录分层

- `cmd/meowbridge/`：服务启动、配置加载、HTTP Server 超时设置。
- `internal/config/`：环境变量解析和启动配置。
- `internal/auth/`：管理员密码校验和 JWT 签发/验证。
- `internal/httpapi/`：公开 Webhook 路由、管理 API、统一请求/响应边界。
- `internal/store/`：SQLite schema、迁移、数据访问。
- `internal/webhook/`：Webhook provider 解析器、字段合并、fallback 逻辑。
- `internal/meow/`：MeoW API 客户端。
- `internal/respond/`：统一 JSON 响应结构。
- `internal/webui/`：Go embed 管理后台静态资源与 `/admin/` SPA 路由处理。
- `web/`：Vue 3 + Vite 管理后台源码，构建产物复制到 `internal/webui/dist/` 后由 Go embed 提供。
- `docs/superpowers/specs/`：设计规格。
- `docs/superpowers/plans/`：可执行实施计划。

## 开发规则

- 默认在 `main` 分支小步提交并直接推送，除非用户明确要求分支、PR 或隔离工作区。
- 发布稳定版本时先在 GitHub 创建空的 Release，然后推送 `v*` tag 触发 Docker 镜像推送；发布 Release 时 CD workflow 会自动上传 Linux 二进制包。
- 修改前先确认目标、影响文件和验证方式；不要顺手重构无关代码。
- 保持后端 API 与 README 示例一致，接口变更必须同步更新文档。
- Webhook 入口必须保持“标准 Webhook URL 可直接填写”的产品语义，不引入要求外部服务定制字段的设计。
- 未识别 payload 要尽量可读地保留原始信息，优先保证消息不丢。
- 不在 README 中写 AI 协作、内部计划、交接记录或发布流程。
- 不提交本地数据库、临时工作区、密钥、token、私有配置或生成的敏感文件。

## 安全边界

- Webhook token 当前按明文存储处理；日志和管理 API 不应泄露 token 全量值。
- Webhook 原始 payload 可能含敏感信息，新增日志展示能力时默认脱敏或限制可见范围。
- 公开 Webhook 响应不得暴露 MeoW 上游内部错误细节。
- HTTP 请求体必须保持大小限制，避免公开入口导致内存或磁盘资源耗尽。
- 后端模板或未来前端渲染用户输入时必须进行 HTML 转义，防止 XSS。

## 本地运行

```powershell
$env:ADMIN_PASSWORD="change-me"
$env:JWT_SECRET="replace-with-long-random-secret"
$env:MEOW_API_BASE_URL="https://api.chuckfang.com"
go run ./cmd/meowbridge
```

可选环境变量：

- `HTTP_ADDR`：HTTP 监听地址，默认 `:8080`。
- `DATABASE_PATH`：SQLite 数据库路径，默认 `meowbridge.db`。

## 验证命令

```powershell
go test ./...
```

涉及前端变更时运行：

```powershell
cd web
npm run test
npm run type-check
npm run build
cd ..
```

涉及启动流程、配置、路由、鉴权、Webhook 解析、MeoW 客户端、SQLite store 或 Go embed 静态资源的变更，至少运行 `go test ./...`。只改文档时可跳过测试，但交付说明必须明确说明未运行测试的原因。

`web/dist/` 是本地前端构建产物，保持忽略；用于发布的内嵌资源位于 `internal/webui/dist/`。

## 发布流程

- 先在 GitHub 创建 Release（Draft），然后打 `v*` tag，触发 Docker 镜像推送。
- 发布 Release 时 CD workflow 会自动上传 Linux amd64/arm64 二进制包。

## 提交约定

- commit message 使用中文。
- 每次提交只包含一个清晰目标。
- 提交前检查 `git status --short`，确认没有带入无关文件。
- 推送前运行与改动风险匹配的验证命令。
