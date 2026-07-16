# First Run Admin Setup Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Docker 部署不再要求 `ADMIN_PASSWORD`，首次打开管理后台即可设置管理员密码。

**Architecture:** 后端允许空管理员启动，提供公开的 setup 状态与一次性初始化接口；前端登录页根据状态展示“首次初始化”或“登录”；初始化成功后复用现有 JWT 登录态进入后台。

**Tech Stack:** Go/chi/SQLite/bcrypt/JWT，Vue 3/Vite/Vitest。

## Global Constraints

- 保留 `/webhook/{token}` 行为不变。
- `ADMIN_PASSWORD` 仅作为可选无人值守初始化方式；不再是 Docker 必填环境变量。
- 初始化接口仅在 `admin_users` 为空时可用；管理员存在后必须返回禁止重复初始化。
- 用户文档保持一行 Docker 命令优先，不引入 `.env` 或 Compose。

---

### Task 1: 后端启动与 setup API

**Files:**
- Modify: `internal/store/bootstrap.go`
- Modify: `internal/store/store_test.go`
- Modify: `internal/httpapi/types.go`
- Modify: `internal/httpapi/admin.go`
- Modify: `internal/httpapi/admin_test.go`
- Modify: `internal/httpapi/router.go`

**Interfaces:**
- Produces: `Store.AdminExists(ctx context.Context) (bool, error)`
- Produces: `Store.CreateInitialAdmin(ctx context.Context, password string) error`
- Produces: `GET /api/admin/setup` returning `{ "needs_setup": boolean }`
- Produces: `POST /api/admin/setup` accepting `{ "password": string }` and returning `{ "token": string }`

- [ ] Add failing store tests for bootstrap without password and one-time admin creation.
- [ ] Implement store methods and update `Bootstrap` to skip admin creation when no password is provided.
- [ ] Add failing HTTP API tests for setup status, first setup, and repeat setup rejection.
- [ ] Implement setup handlers and public routes.
- [ ] Run `rtk go test ./internal/store ./internal/httpapi`.

### Task 2: 前端初始化模式

**Files:**
- Modify: `web/src/api/types.ts`
- Modify: `web/src/api/client.ts`
- Modify: `web/src/api/client.test.ts`
- Modify: `web/src/stores/auth.ts`
- Modify: `web/src/pages/LoginPage.vue`
- Modify: `web/src/pages/LoginPage.test.ts`

**Interfaces:**
- Consumes: `GET /api/admin/setup`
- Consumes: `POST /api/admin/setup`
- Produces: `authStore.setup(password: string): Promise<void>`

- [ ] Add failing API client tests for setup status and setup token.
- [ ] Implement client and auth store methods.
- [ ] Add failing login page test for setup copy and setup submission.
- [ ] Implement setup-mode login page.
- [ ] Run `rtk npm run test -- src/api/client.test.ts src/pages/LoginPage.test.ts`.

### Task 3: 文档、AGENTS 与构建产物

**Files:**
- Modify: `README.md`
- Modify: `AGENTS.md`
- Modify: `internal/webui/dist/**`

- [ ] Update README Docker examples to one-line command without `ADMIN_PASSWORD`.
- [ ] Update AGENTS project conventions for first-run setup.
- [ ] Run `rtk npm run build` in `web` and sync `web/dist` to `internal/webui/dist`.
- [ ] Run full verification: `rtk go test ./...`, `rtk npm run test`, `rtk npm run type-check`, `rtk npm run build`, `rtk git diff --check`.
