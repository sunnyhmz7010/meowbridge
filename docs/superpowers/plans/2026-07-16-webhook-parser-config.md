# Webhook Parser Config Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 meowbridge 增加 magicpush 式入站解析配置：每个 endpoint 可选择预设解析器或自定义字段映射，并在后台预览解析结果。

**Architecture:** endpoint 保存 `parser_config` JSON。Webhook 请求处理时优先使用 endpoint 解析配置，未配置时继续使用现有内置解析器链，最后 fallback 为完整 JSON。前端在 Endpoint 编辑页暴露解析器配置与 payload 预览，不改变公开 `/webhook/{token}` URL。

**Tech Stack:** Go 1.26、SQLite、Vue 3、TypeScript、Vite；不新增运行时依赖。

## Global Constraints

- 保持 `/webhook/{token}` 标准入口不变。
- `MEOW_API_BASE_URL` 保持代码内置，不新增环境变量。
- Docker Compose 保持移除状态。
- 不引入外部 JSONPath 依赖，只实现 `$.a.b[0].c` 最小语法。
- 未识别 payload 继续 fallback 为完整 JSON，优先保证消息不丢。
- 管理后台保持现代简约风格，并支持亮色、暗色、跟随系统。
- 公共 README 只写用户需要的解析器使用说明，不写 AI 协作内容。

---

### Task 1: 后端解析配置模型与 JSONPath

**Files:**
- Modify: `internal/webhook/types.go`
- Create: `internal/webhook/jsonpath.go`
- Create: `internal/webhook/config_parser.go`
- Test: `internal/webhook/jsonpath_test.go`
- Test: `internal/webhook/config_parser_test.go`

**Interfaces:**
- Produce `ParserConfig` with fields `Mode`, `Preset`, `FieldMapping`, `DefaultValues`.
- Produce `ParseWithConfig(input ParseInput, config ParserConfig) (ParsedMessage, bool, error)`.
- Produce `ParseJSONPath(source any, path string) (string, bool)`.
- Existing `Parse(input ParseInput)` remains available.

**Steps:**
- [ ] Write failing JSONPath tests for object paths, array indexes, missing fields, and unsupported syntax.
- [ ] Run `rtk go test ./internal/webhook -run TestParseJSONPath`.
- [ ] Implement `ParseJSONPath`.
- [ ] Run JSONPath tests until they pass.
- [ ] Write failing config parser tests for literal + JSONPath concatenation, escape handling, `github_push_minimal` preset, and invalid config fallback behavior.
- [ ] Run `rtk go test ./internal/webhook -run 'TestParseWithConfig|TestPreset'`.
- [ ] Implement `ParserConfig`, preset definitions, and config parser.
- [ ] Run `rtk go test ./internal/webhook`.

### Task 2: Store schema、endpoint API 与 webhook 链路接入

**Files:**
- Modify: `internal/store/store.go`
- Modify: `internal/store/models.go`
- Modify: `internal/store/endpoints.go`
- Modify: `internal/store/store_test.go`
- Modify: `internal/httpapi/admin.go`
- Modify: `internal/httpapi/webhook.go`
- Modify: `internal/httpapi/admin_test.go`
- Modify: `internal/httpapi/webhook_test.go`

**Interfaces:**
- Add endpoint JSON field `parser_config`.
- Add admin endpoints:
  - `GET /api/admin/webhook/presets`
  - `POST /api/admin/webhook/preview`
- `handleWebhook` decodes endpoint `parser_config` and tries configured parser before built-in parser.

**Steps:**
- [ ] Write failing store test proving `parser_config` is saved, listed, loaded, and migrated for existing DBs.
- [ ] Run `rtk go test ./internal/store -run ParserConfig`.
- [ ] Implement schema column and store CRUD support.
- [ ] Run store tests.
- [ ] Write failing webhook test using the reported PowerShell/GitHub minimal payload with endpoint parser config.
- [ ] Run `rtk go test ./internal/httpapi -run Minimal`.
- [ ] Wire parser config into webhook request handling.
- [ ] Write failing admin API tests for presets and preview.
- [ ] Implement admin API handlers.
- [ ] Run `rtk go test ./internal/httpapi`.

### Task 3: 管理后台解析器配置 UI

**Files:**
- Modify: `web/src/api/types.ts`
- Modify: `web/src/api/client.ts`
- Modify: `web/src/pages/EndpointFormPage.vue`
- Modify: `web/src/pages/EndpointsPage.vue`
- Test: `web/src/api/client.test.ts`

**Interfaces:**
- Add `ParserConfig`, `WebhookPreset`, `WebhookPreviewRequest`, `WebhookPreviewResult` types.
- Add client methods `getWebhookPresets()` and `previewWebhook(payload)`.
- Endpoint form edits `parser_config` as JSON-compatible structured data.

**Steps:**
- [ ] Write failing frontend API normalization test for `parser_config`.
- [ ] Run `rtk npm run test -- src/api/client.test.ts`.
- [ ] Add frontend types and client methods.
- [ ] Add Endpoint form section for automatic/preset/custom parsing.
- [ ] Add payload preview area using `/api/admin/webhook/preview`.
- [ ] Add parser status to endpoint list.
- [ ] Run `rtk npm run test`.
- [ ] Run `rtk npm run type-check`.
- [ ] Run `rtk npm run build`.
- [ ] Copy `web/dist` to `internal/webui/dist`.

### Task 4: 文档、项目说明与全量验证

**Files:**
- Modify: `README.md`
- Modify: `AGENTS.md`

**Steps:**
- [ ] Update README with concise parser config usage and one example for the minimal GitHub payload.
- [ ] Update AGENTS with project-level parser config rule.
- [ ] Run `rtk go test ./...`.
- [ ] Run `rtk npm run test`, `rtk npm run type-check`, and `rtk npm run build` in `web`.
- [ ] Confirm `rtk git status --short` only contains intended files.
- [ ] Commit with Chinese commit message.
- [ ] Push `main` if verification passes.
