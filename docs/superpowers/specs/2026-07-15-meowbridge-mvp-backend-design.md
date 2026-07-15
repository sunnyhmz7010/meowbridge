# meowbridge 一期 MVP 后端设计

## 目标

一期 MVP 聚焦后端核心链路：提供通用 Webhook 接收接口，将外部服务的 Webhook 请求自动解析并同步转发到 MeoW Push API。外部服务应尽量只需要填写 `https://<meowbridge-host>/webhook/{token}`，无需为每个服务手工配置字段映射。

本阶段不实现 Vue 管理后台、Telegram Bot API 劫持、队列重试、内置限流和用户自定义 JSONPath 映射。这些能力保留为后续独立设计。

## 已确认约束

- 后端语言：Go 1.23+。
- HTTP 路由：`chi`。
- 数据库：SQLite，驱动使用 `modernc.org/sqlite`，避免 CGO。
- 管理鉴权：JWT Bearer。
- 管理员初始化：首次启动从 `ADMIN_PASSWORD` 初始化单管理员密码，写入数据库后不再依赖该环境变量。
- 全局设置初始化：首次启动从环境变量初始化 MeoW API 地址、日志保留天数等设置，之后通过管理 API 修改并持久化。
- Webhook token：数据库明文保存，后台可展示和复制。
- 推送日志：保存完整原始 payload。
- MeoW 转发：同步调用一次，不做重试。
- Webhook 响应：统一返回 meowbridge JSON，不透传 MeoW 响应作为公开契约。
- 频率限制：一期不内置，交给反向代理、WAF 或部署层处理。
- Telegram 劫持：一期不包含，只保留二期边界。

## 架构

服务采用 Go 单体结构，SQLite 内置迁移，适合单二进制自部署。

建议模块划分：

- `cmd/meowbridge`：程序入口，加载配置、初始化数据库、注册路由、启动 HTTP 服务。
- `internal/http`：HTTP handler 和中间件，负责请求解析、JWT 鉴权和统一响应。
- `internal/auth`：管理员密码哈希、登录校验、JWT 签发与验证。
- `internal/store`：SQLite 表结构、迁移和 CRUD。
- `internal/webhook`：Webhook 解析器链，将不同 payload 转成统一推送消息。
- `internal/meow`：MeoW API 客户端，负责构造请求并同步转发。
- `internal/logging`：推送日志写入和查询逻辑。实现上可与 store 复用，但业务边界保持清楚。

核心数据流：

```text
外部服务 Webhook
  -> POST /webhook/{token}
  -> 查找 endpoint 并检查 active
  -> 解析 payload
  -> 合并 query 覆盖值、endpoint 默认值、MeoW 默认值
  -> 同步调用 MeoW API
  -> 写入 push_logs
  -> 返回 meowbridge JSON
```

管理数据流：

```text
管理员登录
  -> 获取 JWT
  -> 管理 endpoints / settings / push_logs
```

handler 只做请求编排，不承载具体解析逻辑。Webhook 解析器、字段合并和 MeoW 客户端都应能独立测试。

## 数据模型

### `admin_users`

单管理员表。

字段：

- `id`
- `password_hash`
- `created_at`
- `updated_at`

启动时如果不存在管理员，则读取 `ADMIN_PASSWORD` 并写入哈希。若已存在管理员，忽略环境变量中的管理员密码。

### `endpoints`

推送入口配置表。

字段：

- `id`
- `name`
- `token`
- `meow_nickname`
- `default_title`
- `msg_type`
- `html_height`
- `default_url`
- `default_img_url`
- `active`
- `created_at`
- `updated_at`

约束：

- `token` 明文保存，并建立唯一索引。
- `meow_nickname` 创建后不可修改。
- 支持创建、列表、详情、更新、删除、启停和重置 token。
- 删除 endpoint 不级联删除历史日志。

### `settings`

全局设置表。

字段：

- `key`
- `value`
- `updated_at`

建议一期设置项：

- `meow_api_base_url`
- `log_retention_days`

`meow_api_base_url` 首次从环境变量 `MEOW_API_BASE_URL` 初始化。`log_retention_days` 默认值为 14 天。

### `push_logs`

推送日志表。

字段：

- `id`
- `endpoint_id`
- `endpoint_name`
- `token`
- `source_type`
- `request_method`
- `request_headers`
- `request_query`
- `request_payload`
- `parsed_title`
- `parsed_msg`
- `parsed_msg_type`
- `meow_status_code`
- `meow_response_body`
- `success`
- `error_message`
- `created_at`

说明：

- 保存完整原始 payload，便于排查和后续新增解析器。
- 日志包含请求时的 endpoint 名称和 token 快照，endpoint 删除后仍可追溯。
- 日志含敏感信息风险，必须只允许管理员 JWT 访问。
- 列表接口默认不返回 token，详情接口可返回完整日志。
- 支持按保留天数清理历史日志。

## Webhook 接口设计

公开入口：

```text
POST /webhook/{token}
GET  /verify/{token}
```

`POST /webhook/{token}` 支持：

- `application/json`：读取完整 JSON，进入解析器链。
- `text/plain`：body 直接作为 `msg`。

可选 query 参数：

- `title`
- `msgType`
- `htmlHeight`
- `url`
- `imgUrl`

query 参数用于轻量覆盖，不要求外部服务必须配置。

## Webhook 解析器链

JSON payload 按解析器链处理。解析器只负责识别 payload 和抽取候选字段，不访问数据库、不调用 MeoW、不处理 HTTP 响应。

统一输出结构：

- `source_type`
- `title`
- `msg`
- `url`
- `img_url`
- `msg_type`

解析器顺序：

1. `github_pr`
2. `github_action`
3. `github`
4. `jenkins`
5. `grafana`
6. `prometheus`
7. `zabbix`
8. `gotify`
9. `emby`
10. `generic`
11. `fallback`

更具体的解析器排在通用解析器前面，避免被提前吞掉。

识别策略：

- 优先使用 Header，例如 `X-GitHub-Event`。
- 其次使用 payload 结构特征，例如 `pull_request`、`workflow_run`、`repository`、`alerts`、`trigger`、`Title` 和 `Description`。
- 多个解析器匹配时，以链顺序为准。

兜底策略：

- `generic` 尝试提取常见字段：`msg`、`message`、`text`、`content`、`title`、`url`、`imgUrl`、`msgType`。
- `fallback` 将完整 JSON 格式化为文本或 Markdown 消息，保证未识别 payload 不丢失。

## 字段合并规则

最终推送字段按以下优先级合并：

1. query 覆盖：`title`、`msgType`、`htmlHeight`、`url`、`imgUrl`
2. 解析器输出
3. endpoint 默认值：`default_title`、`msg_type`、`html_height`、`default_url`、`default_img_url`
4. MeoW 默认值：`title=Meow`、`msgType=text`、`htmlHeight=200`

`msg` 是唯一必须字段：

- JSON 请求如果没有生成 `msg`，则使用 `fallback` 的格式化完整 JSON。
- `text/plain` 请求如果 body 为空，返回 `400` 并记录失败日志。

MeoW 调用时：

- `msgType=html` 时传 `htmlHeight`。
- 非 HTML 消息不传或忽略 `htmlHeight`。
- `url` 和 `imgUrl` 为空时不传。

## 管理 API

管理 API 使用 `Authorization: Bearer <token>`。

接口：

- `POST /api/admin/login`
- `GET /api/admin/endpoints`
- `POST /api/admin/endpoints`
- `GET /api/admin/endpoints/:id`
- `PUT /api/admin/endpoints/:id`
- `DELETE /api/admin/endpoints/:id`
- `POST /api/admin/endpoints/:id/reset-token`
- `PATCH /api/admin/endpoints/:id/active`
- `GET /api/admin/push-logs`
- `GET /api/admin/push-logs/:id`
- `DELETE /api/admin/push-logs`
- `GET /api/admin/settings`
- `PUT /api/admin/settings`
- `POST /api/admin/change-password`
- `GET /api/admin/webhook/presets`

一期不实现 refresh token。JWT 过期后重新登录。

## 响应格式

普通成功响应：

```json
{
  "ok": true,
  "data": {}
}
```

Webhook 成功响应：

```json
{
  "ok": true,
  "log_id": 123
}
```

失败响应：

```json
{
  "ok": false,
  "error": "message"
}
```

HTTP 状态码仍表达成败：

- token 不存在：`404`
- endpoint 已禁用：`403`
- body 为空或格式错误：`400`
- 管理 JWT 缺失或无效：`401`
- 管理参数校验失败：`400`
- MeoW 上游网络错误或非 2xx：`502`
- 未预期内部错误：`500`

## 推送日志策略

`/webhook/{token}` 成功或失败时尽量写入 `push_logs`。

写日志规则：

- token 不存在时不落库，直接返回 `404`，避免匿名扫描制造日志噪声。
- endpoint 禁用、payload 错误、MeoW 失败和推送成功都落库，因为可以关联 endpoint。
- MeoW 响应体保存最多 16 KiB，避免异常大响应撑爆日志。
- 请求 header 保存时应避免在普通列表接口直接展示。

## MeoW 客户端策略

MeoW API 地址来自 settings 中的 `meow_api_base_url`。

调用策略：

- 同步调用一次。
- 默认超时 10 秒。
- 非 2xx 视为失败。
- 保存 MeoW 状态码和最多 16 KiB 的响应体到日志。
- 不将 MeoW 原始响应透传给调用方。

MeoW 请求构造以官方接口为准：目标 nickname 来自 endpoint 的 `meow_nickname`，消息字段来自字段合并结果。

## 测试策略

### 单元测试

Webhook 解析器链：

- 10 类来源样例 payload 都能识别出 `source_type`。
- 每类至少验证 `title`、`msg`、`msg_type`、`url` 的关键字段。
- 未识别 JSON 会 fallback 为格式化 JSON。
- `text/plain` body 直接生成 `msg`。

字段合并：

- query 覆盖解析器输出。
- 解析器输出覆盖 endpoint 默认值。
- endpoint 默认值覆盖 MeoW 默认值。
- 空 `msg` 返回错误。

endpoint 校验：

- token 不存在。
- endpoint 禁用。
- token 存在且 active。

### 集成测试

- 使用临时 SQLite 初始化数据库和迁移。
- 使用 `httptest` mock MeoW API。
- MeoW 返回 2xx 时 webhook 返回 `ok=true`。
- MeoW 返回 500 时 webhook 返回 `502`。
- 网络错误时 webhook 返回 `502`。
- 成功和失败都写入 `push_logs`。
- 首次启动从 `ADMIN_PASSWORD` 初始化管理员。
- 登录成功返回 JWT。
- 未带 JWT 的管理接口返回 `401`。
- endpoint CRUD、启停、重置 token 可用。

## 验收标准

- 能通过环境变量启动服务并初始化管理员和 MeoW 地址。
- 能通过管理 API 创建 endpoint。
- 将 `/webhook/{token}` 填到 GitHub、Grafana、Prometheus 等服务的 Webhook 配置后，不做字段映射也能收到 MeoW 推送。
- 未识别 payload 不丢失，推送格式化 JSON。
- MeoW 调用失败时调用方收到非 2xx，日志可查失败原因。
- `go test ./...` 通过。

## 风险与后续边界

- token 明文存储和日志保存完整 payload 会扩大数据库泄露后的影响面。MVP 通过管理员鉴权、日志保留清理和避免普通响应泄露 token 降低风险。
- Webhook payload 没有统一标准。MVP 的目标是“尽量自动解析常见服务，未知服务不丢消息”，不是保证所有服务都有精炼标题和正文。
- 不内置限流意味着公开 Webhook URL 泄露后可能被滥用。部署文档应建议在反向代理或 WAF 层做限流。
- 不做重试意味着 MeoW 短暂故障会导致本次推送失败。后续如需高可靠投递，应单独设计队列、重试、幂等和死信机制。
- Telegram Bot API 劫持属于二期功能，应独立规格化，不混入一期 MVP。
