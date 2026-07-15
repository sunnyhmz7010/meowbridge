# meowbridge Vue 管理后台 MVP 设计规格

## 背景

meowbridge 后端 MVP 已提供完整的管理 API，但当前只能通过 HTTP 客户端手动调用。下一步需要提供一个轻量、可嵌入单二进制的 Vue 管理后台，让单用户管理员完成日常配置和排障。

管理后台的目标不是做复杂运营系统，而是把后端已具备的能力可视化：

- 登录并持有管理员 JWT。
- 创建、编辑、启停、删除和重置 Webhook endpoint。
- 复制外部服务可直接填写的标准 Webhook URL。
- 查看推送日志列表与详情。
- 修改 MeoW API Base URL、日志保留天数和管理员密码。

## 设计目标

1. 保持部署简单：前端构建产物由 Go `embed` 内嵌，生产环境只发布一个后端二进制。
2. 保持 API 简单：复用现有 `/api/admin/*`，不为前端新增平行业务接口。
3. 保持标准 Webhook 语义：UI 展示和复制的入口仍是 `/webhook/{token}`。
4. 保持 MVP 边界：不做多用户、角色权限、复杂图表、实时推送、主题系统或 i18n。
5. 保持安全基线：token 不主动全量展示在列表页；日志详情明确提示 payload 可能包含敏感信息。

## 技术选择

推荐方案：Vue 3 + TypeScript + Vite + Tailwind CSS，构建产物放入 `web/dist/`，由 Go 后端通过 `embed.FS` 提供静态资源和 SPA fallback。

选择理由：

- Vue 3 与 Vite 是当前主流轻量组合，适合快速构建管理后台。
- TypeScript 能减少 API 字段名漂移，后续维护成本低。
- Tailwind CSS 适合快速做一致、简洁的后台界面，不引入重型 UI 框架。
- Go embed 保持单二进制部署，不增加 Nginx、CDN 或跨域配置要求。

不采用的方案：

- 前后端分离部署：开发灵活，但会引入额外部署、跨域、版本同步成本。
- 服务端模板：依赖少，但交互状态复杂度会转移到后端模板和原生 JS，长期维护性较差。

## 用户路径

### 登录

管理员访问 `/admin/` 进入登录页，输入密码后调用：

`POST /api/admin/login`

成功后将 `data.token` 保存到 `localStorage`，后续管理请求统一携带：

`Authorization: Bearer <token>`

收到 `401` 时清除本地 token 并跳转登录页。

### Endpoint 管理

Endpoint 列表页展示：

- 名称
- MeoW nickname
- 默认标题
- 消息类型
- 启用状态
- Webhook URL 复制按钮
- 编辑、启停、重置 token、删除操作

创建表单字段对应后端 `endpointRequest`：

- `name`：必填。
- `meow_nickname`：创建时必填；编辑时不可修改。
- `default_title`：可选。
- `msg_type`：`text` / `html` / `markdown`，默认 `text`。
- `html_height`：正整数，默认 `200`。
- `default_url`：可选。
- `default_img_url`：可选。
- `active`：默认启用。

Webhook URL 由浏览器当前 origin 和 endpoint token 组合：

`{window.location.origin}/webhook/{token}`

### 推送日志

日志列表页调用：

`GET /api/admin/push-logs`

列表展示：

- 时间
- endpoint 名称
- source type
- 解析标题
- 解析消息摘要
- MeoW 状态码
- 成功/失败状态
- 错误信息摘要

点击日志进入详情页，调用：

`GET /api/admin/push-logs/{id}`

详情展示请求方法、headers、query、payload、解析字段、MeoW 响应、错误信息。payload、headers 和响应体使用 `<pre>` 或只读文本块展示，不使用 `innerHTML`。

清理日志按钮调用：

`DELETE /api/admin/push-logs`

按钮文案明确说明按当前 `log_retention_days` 清理。

### 设置

设置页包含三块：

1. MeoW 设置：读取和更新 `meow_api_base_url`。
2. 日志设置：读取和更新 `log_retention_days`。
3. 修改密码：输入旧密码和新密码，调用 `POST /api/admin/change-password`。

设置更新调用：

`PUT /api/admin/settings`

只提交被修改的 key，避免覆盖其它设置。

## 前端结构

建议新增 `web/` 目录：

```text
web/
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
└── src/
    ├── main.ts
    ├── App.vue
    ├── style.css
    ├── api/
    │   ├── client.ts
    │   └── types.ts
    ├── router/
    │   └── index.ts
    ├── stores/
    │   └── auth.ts
    ├── components/
    │   ├── AppLayout.vue
    │   ├── ConfirmDialog.vue
    │   ├── EmptyState.vue
    │   └── ToastHost.vue
    └── pages/
        ├── LoginPage.vue
        ├── EndpointsPage.vue
        ├── EndpointFormPage.vue
        ├── LogsPage.vue
        ├── LogDetailPage.vue
        └── SettingsPage.vue
```

边界说明：

- `api/client.ts` 只处理请求、响应解包、错误转换和 Bearer token 注入。
- `api/types.ts` 定义与后端 JSON 字段一致的 TypeScript 类型。
- `stores/auth.ts` 管理 token、登录状态、登出和 401 处理。
- `pages/*` 只做页面状态和表单交互，不直接拼底层 fetch 逻辑。
- `components/*` 放跨页面可复用 UI，不引入复杂组件库。

## 后端静态资源设计

新增 `internal/webui/`：

```text
internal/webui/
├── embed.go
└── embed_disabled.go
```

推荐使用构建标签处理“前端尚未构建”的开发场景：

- 默认构建：如果 `web/dist` 不存在，Go 后端仍可测试和运行 API。
- release 构建：构建前端后，通过 `go:embed` 内嵌 `web/dist`。

路由规则：

- `/api/*` 永远走 API，不进入 SPA fallback。
- `/webhook/{token}` 和 `/verify/{token}` 永远走现有后端路由。
- `/admin/*` 返回前端静态资源；未知子路径返回 `index.html`。
- `/assets/*` 返回构建产物中的静态资源。
- `/` 可重定向到 `/admin/`，但不影响公开 Webhook URL。

## API 契约

前端统一处理后端响应格式：

成功：

```json
{ "ok": true, "data": {} }
```

失败：

```json
{ "ok": false, "error": "message" }
```

Webhook 成功响应 `{ "ok": true, "log_id": 1 }` 不属于管理后台 API client 的核心路径。

TypeScript 类型必须使用后端 JSON 字段名：

- `meow_nickname`
- `default_title`
- `msg_type`
- `html_height`
- `default_url`
- `default_img_url`
- `created_at`
- `updated_at`

## 错误处理

- 登录失败：显示“密码错误或凭证无效”。
- 401：清除 token，跳转 `/admin/login`。
- 400：展示后端返回的 `error`。
- 404：endpoint 或日志详情页显示不存在状态。
- 5xx：展示通用失败提示，不暴露额外调试信息。
- 网络错误：提示服务不可达，保留当前页面状态。

所有危险操作必须二次确认：

- 删除 endpoint。
- 重置 token。
- 清理日志。
- 修改密码。

## 安全与隐私

- token 存储在 `localStorage` 是 MVP 取舍；后续可改为 HttpOnly Cookie。
- 列表页不展示完整 Webhook token，只提供复制完整 URL 的按钮。
- 日志详情中的 payload、headers、query、MeoW 响应以纯文本方式渲染，禁止使用 `v-html`。
- 退出登录必须清除本地 token。
- 修改密码成功后建议强制重新登录，避免旧 token 继续被误用。

## 验证要求

前端：

- `npm run type-check`
- `npm run build`

后端：

- `go test ./...`

关键测试点：

- 未登录访问 `/admin/endpoints` 跳转登录页。
- 登录成功后能进入 endpoint 列表。
- API client 对 `{ ok: false, error }` 抛出可展示错误。
- 401 响应触发登出。
- Go 静态资源路由不会吞掉 `/api/*`、`/webhook/*`、`/verify/*`。
- SPA fallback 能让 `/admin/logs/1` 刷新后仍返回前端入口。

## 非目标

本阶段不做：

- 多用户、RBAC、注册流程。
- WebSocket 或实时日志。
- 图表统计。
- 移动端深度适配。
- i18n。
- Telegram Bot API 劫持。
- Docker、CI、Release 构建流水线。

## 验收标准

1. 管理员可以通过浏览器完成登录、endpoint 管理、日志查看、设置修改和改密。
2. 外部服务可复制得到标准 Webhook URL：`/webhook/{token}`。
3. 前端构建产物可由 Go 后端在 `/admin/` 下提供。
4. API 认证失败、业务错误和网络错误都有明确提示。
5. 前端构建和后端测试均通过。
