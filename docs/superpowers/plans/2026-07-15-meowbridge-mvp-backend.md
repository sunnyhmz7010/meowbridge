# meowbridge MVP Backend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the meowbridge backend MVP: authenticated admin APIs, SQLite persistence, universal Webhook parsing, synchronous MeoW forwarding, and push logs.

**Architecture:** A Go 1.23+ single-binary HTTP service using `chi`, SQLite via `modernc.org/sqlite`, focused internal packages, and testable boundaries. Handlers orchestrate request/response flow; storage, auth, Webhook parsing, field merging, and MeoW forwarding are implemented as separate units with direct tests.

**Tech Stack:** Go 1.23+, `github.com/go-chi/chi/v5`, `modernc.org/sqlite`, `github.com/golang-jwt/jwt/v5`, `golang.org/x/crypto/bcrypt`, standard `net/http/httptest`. Resolve dependency versions at implementation time with `go get ...@latest` so the build does not pin stale preselected versions.

## Global Constraints

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
- 真正执行实现前，不在 `main` 分支直接启动较大实现；先创建隔离分支或工作区。
- 每个任务完成后运行指定测试，并使用中文 commit message 提交该任务范围内的文件。

---

## File Structure

Create and maintain these files:

- `go.mod`：Go module definition and direct dependencies.
- `go.sum`：Dependency checksums.
- `cmd/meowbridge/main.go`：Program entry, environment loading, store initialization, router setup, HTTP server startup.
- `internal/config/config.go`：Environment parsing and runtime configuration defaults.
- `internal/respond/respond.go`：Uniform JSON response helpers.
- `internal/store/store.go`：SQLite connection wrapper, schema migration, transaction helpers.
- `internal/store/models.go`：Shared persistence model structs.
- `internal/store/bootstrap.go`：Admin/settings initialization from environment.
- `internal/store/endpoints.go`：Endpoint CRUD and token lookup.
- `internal/store/logs.go`：Push log create/list/detail/cleanup queries.
- `internal/store/settings.go`：Settings get/update queries.
- `internal/auth/auth.go`：Password hashing/verification and JWT creation/validation.
- `internal/token/token.go`：Cryptographically secure token generation.
- `internal/webhook/types.go`：Parser input/output structs and final message struct.
- `internal/webhook/parsers.go`：Parser chain orchestration.
- `internal/webhook/providers.go`：Built-in provider parsers for GitHub, Jenkins, Grafana, Prometheus, Zabbix, Gotify, Emby, generic, fallback.
- `internal/webhook/merge.go`：Field precedence and final message validation.
- `internal/meow/client.go`：MeoW HTTP client and request construction.
- `internal/httpapi/router.go`：Route registration and middleware wiring.
- `internal/httpapi/admin.go`：Admin auth, endpoint, settings, log handlers.
- `internal/httpapi/webhook.go`：Public Webhook and token verification handlers.
- `internal/httpapi/types.go`：Request/response DTOs used by HTTP handlers.
- `internal/testutil/testutil.go`：Temporary test database and helper assertions.
- `README.md`：User-facing project overview, run instructions, API examples, GPL-3.0 license statement.
- `AGENTS.md`：Keep existing project requirement notes; update only if implementation-specific local commands become necessary.

Test files:

- `internal/config/config_test.go`
- `internal/respond/respond_test.go`
- `internal/store/store_test.go`
- `internal/auth/auth_test.go`
- `internal/token/token_test.go`
- `internal/webhook/parsers_test.go`
- `internal/webhook/merge_test.go`
- `internal/meow/client_test.go`
- `internal/httpapi/webhook_test.go`
- `internal/httpapi/admin_test.go`

---

### Task 1: Project Bootstrap, Config, and Response Helpers

**Files:**
- Create: `go.mod`
- Create: `cmd/meowbridge/main.go`
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`
- Create: `internal/respond/respond.go`
- Create: `internal/respond/respond_test.go`

**Interfaces:**
- Produces: `config.Load() (config.Config, error)`
- Produces: `respond.JSON(w http.ResponseWriter, status int, payload any)`
- Produces: `respond.OK(w http.ResponseWriter, data any)`
- Produces: `respond.WebhookOK(w http.ResponseWriter, logID int64)`
- Produces: `respond.Error(w http.ResponseWriter, status int, message string)`

- [ ] **Step 1: Create isolated implementation branch or worktree**

Run:

```powershell
rtk git status --short
rtk git switch -c feat/mvp-backend
```

Expected:

```text
Switched to a new branch 'feat/mvp-backend'
```

- [ ] **Step 2: Write failing config tests**

Create `internal/config/config_test.go`:

```go
package config

import (
	"testing"
	"time"
)

func TestLoadRequiresAdminPasswordAndMeowBaseURL(t *testing.T) {
	t.Setenv("ADMIN_PASSWORD", "")
	t.Setenv("MEOW_API_BASE_URL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected missing environment error")
	}
}

func TestLoadUsesDefaults(t *testing.T) {
	t.Setenv("ADMIN_PASSWORD", "secret-password")
	t.Setenv("MEOW_API_BASE_URL", "https://push.example.test")
	t.Setenv("DATABASE_PATH", "")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("JWT_SECRET", "jwt-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DatabasePath != "meowbridge.db" {
		t.Fatalf("DatabasePath = %q", cfg.DatabasePath)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q", cfg.HTTPAddr)
	}
	if cfg.LogRetentionDays != 14 {
		t.Fatalf("LogRetentionDays = %d", cfg.LogRetentionDays)
	}
	if cfg.MeowTimeout != 10*time.Second {
		t.Fatalf("MeowTimeout = %s", cfg.MeowTimeout)
	}
}
```

- [ ] **Step 3: Write failing response tests**

Create `internal/respond/respond_test.go`:

```go
package respond

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorResponseShape(t *testing.T) {
	rr := httptest.NewRecorder()
	Error(rr, http.StatusBadRequest, "bad request")

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rr.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json decode: %v", err)
	}
	if body["ok"] != false || body["error"] != "bad request" {
		t.Fatalf("body = %#v", body)
	}
}

func TestWebhookOKResponseShape(t *testing.T) {
	rr := httptest.NewRecorder()
	WebhookOK(rr, 123)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json decode: %v", err)
	}
	if body["ok"] != true || body["log_id"].(float64) != 123 {
		t.Fatalf("body = %#v", body)
	}
}
```

- [ ] **Step 4: Run tests and verify failure**

Run:

```powershell
rtk go test ./internal/config ./internal/respond
```

Expected: FAIL because packages or functions are not implemented.

- [ ] **Step 5: Create `go.mod`**

Run:

```powershell
rtk go mod init github.com/sunnyhmz7010/meowbridge
rtk go get github.com/go-chi/chi/v5@latest
rtk go get github.com/golang-jwt/jwt/v5@latest
rtk go get golang.org/x/crypto@latest
rtk go get modernc.org/sqlite@latest
rtk go mod tidy
```

Expected: `go.mod` and `go.sum` are generated. `go.mod` must contain module path and direct requirements for the selected dependencies. The exact versions are whatever `go get ...@latest` resolves on the implementation date.

Verify `go.mod` has this module line:

```go
module github.com/sunnyhmz7010/meowbridge

go 1.23
```

- [ ] **Step 6: Implement config loader**

Create `internal/config/config.go`:

```go
package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AdminPassword    string
	DatabasePath     string
	HTTPAddr         string
	JWTSecret        string
	MeowAPIBaseURL   string
	LogRetentionDays int
	MeowTimeout      time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		AdminPassword:    os.Getenv("ADMIN_PASSWORD"),
		DatabasePath:     envOrDefault("DATABASE_PATH", "meowbridge.db"),
		HTTPAddr:         envOrDefault("HTTP_ADDR", ":8080"),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		MeowAPIBaseURL:   os.Getenv("MEOW_API_BASE_URL"),
		LogRetentionDays: intEnvOrDefault("LOG_RETENTION_DAYS", 14),
		MeowTimeout:      10 * time.Second,
	}
	if cfg.AdminPassword == "" {
		return Config{}, errors.New("ADMIN_PASSWORD is required")
	}
	if cfg.JWTSecret == "" {
		return Config{}, errors.New("JWT_SECRET is required")
	}
	if cfg.MeowAPIBaseURL == "" {
		return Config{}, errors.New("MEOW_API_BASE_URL is required")
	}
	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func intEnvOrDefault(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
```

- [ ] **Step 7: Implement response helpers**

Create `internal/respond/respond.go`:

```go
package respond

import (
	"encoding/json"
	"net/http"
)

type successResponse struct {
	OK   bool `json:"ok"`
	Data any  `json:"data,omitempty"`
}

type webhookSuccessResponse struct {
	OK    bool  `json:"ok"`
	LogID int64 `json:"log_id"`
}

type errorResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func OK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, successResponse{OK: true, Data: data})
}

func WebhookOK(w http.ResponseWriter, logID int64) {
	JSON(w, http.StatusOK, webhookSuccessResponse{OK: true, LogID: logID})
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, errorResponse{OK: false, Error: message})
}
```

- [ ] **Step 8: Add minimal main**

Create `cmd/meowbridge/main.go`:

```go
package main

import (
	"log"

	"github.com/sunnyhmz7010/meowbridge/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("meowbridge starting on %s", cfg.HTTPAddr)
}
```

- [ ] **Step 9: Verify and commit**

Run:

```powershell
rtk go test ./internal/config ./internal/respond
rtk go test ./...
rtk git status --short
```

Expected: tests pass; only Task 1 files are changed.

Commit:

```powershell
rtk git add go.mod go.sum cmd internal
rtk git commit -m "初始化后端工程结构"
```

---

### Task 2: SQLite Schema, Bootstrap, and Store Foundation

**Files:**
- Create: `internal/store/models.go`
- Create: `internal/store/store.go`
- Create: `internal/store/bootstrap.go`
- Create: `internal/store/settings.go`
- Create: `internal/testutil/testutil.go`
- Create: `internal/store/store_test.go`
- Modify: `cmd/meowbridge/main.go`

**Interfaces:**
- Consumes: `config.Config`
- Produces: `store.Open(ctx context.Context, path string) (*store.Store, error)`
- Produces: `(*Store).Migrate(ctx context.Context) error`
- Produces: `(*Store).Bootstrap(ctx context.Context, opts store.BootstrapOptions) error`
- Produces: `(*Store).GetSetting(ctx context.Context, key string) (string, error)`
- Produces: `(*Store).SetSetting(ctx context.Context, key, value string) error`

- [ ] **Step 1: Write failing migration/bootstrap tests**

Create `internal/store/store_test.go`:

```go
package store

import (
	"context"
	"testing"
)

func TestMigrateCreatesCoreTables(t *testing.T) {
	ctx := context.Background()
	st, cleanup := openTestStore(t)
	defer cleanup()

	for _, table := range []string{"admin_users", "endpoints", "settings", "push_logs"} {
		var name string
		err := st.db.QueryRowContext(ctx, `SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name)
		if err != nil {
			t.Fatalf("table %s was not created: %v", table, err)
		}
	}
}

func TestBootstrapCreatesAdminAndSettingsOnce(t *testing.T) {
	ctx := context.Background()
	st, cleanup := openTestStore(t)
	defer cleanup()

	err := st.Bootstrap(ctx, BootstrapOptions{
		AdminPassword:    "first-password",
		MeowAPIBaseURL:   "https://meow.example.test",
		LogRetentionDays: 14,
	})
	if err != nil {
		t.Fatalf("Bootstrap first: %v", err)
	}

	err = st.Bootstrap(ctx, BootstrapOptions{
		AdminPassword:    "second-password",
		MeowAPIBaseURL:   "https://changed.example.test",
		LogRetentionDays: 30,
	})
	if err != nil {
		t.Fatalf("Bootstrap second: %v", err)
	}

	var adminCount int
	if err := st.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM admin_users`).Scan(&adminCount); err != nil {
		t.Fatalf("count admins: %v", err)
	}
	if adminCount != 1 {
		t.Fatalf("adminCount = %d", adminCount)
	}

	baseURL, err := st.GetSetting(ctx, "meow_api_base_url")
	if err != nil {
		t.Fatalf("GetSetting: %v", err)
	}
	if baseURL != "https://meow.example.test" {
		t.Fatalf("baseURL = %q", baseURL)
	}
}

func openTestStore(t *testing.T) (*Store, func()) {
	t.Helper()
	ctx := context.Background()
	st, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := st.Migrate(ctx); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return st, func() { _ = st.Close() }
}
```

- [ ] **Step 2: Run tests and verify failure**

Run:

```powershell
rtk go test ./internal/store
```

Expected: FAIL because store types do not exist.

- [ ] **Step 3: Implement models**

Create `internal/store/models.go`:

```go
package store

import "time"

type Endpoint struct {
	ID            int64
	Name          string
	Token         string
	MeowNickname  string
	DefaultTitle  string
	MsgType       string
	HTMLHeight    int
	DefaultURL    string
	DefaultImgURL string
	Active        bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Setting struct {
	Key       string
	Value     string
	UpdatedAt time.Time
}

type PushLog struct {
	ID               int64
	EndpointID       int64
	EndpointName     string
	Token            string
	SourceType       string
	RequestMethod    string
	RequestHeaders   string
	RequestQuery     string
	RequestPayload   string
	ParsedTitle      string
	ParsedMsg        string
	ParsedMsgType    string
	MeowStatusCode   int
	MeowResponseBody string
	Success          bool
	ErrorMessage     string
	CreatedAt        time.Time
}
```

- [ ] **Step 4: Implement store and schema migration**

Create `internal/store/store.go`:

```go
package store

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(ctx context.Context, path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS admin_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			password_hash TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS endpoints (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			token TEXT NOT NULL UNIQUE,
			meow_nickname TEXT NOT NULL,
			default_title TEXT NOT NULL DEFAULT '',
			msg_type TEXT NOT NULL DEFAULT 'text',
			html_height INTEGER NOT NULL DEFAULT 200,
			default_url TEXT NOT NULL DEFAULT '',
			default_img_url TEXT NOT NULL DEFAULT '',
			active INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS push_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			endpoint_id INTEGER NOT NULL,
			endpoint_name TEXT NOT NULL,
			token TEXT NOT NULL,
			source_type TEXT NOT NULL,
			request_method TEXT NOT NULL,
			request_headers TEXT NOT NULL,
			request_query TEXT NOT NULL,
			request_payload TEXT NOT NULL,
			parsed_title TEXT NOT NULL,
			parsed_msg TEXT NOT NULL,
			parsed_msg_type TEXT NOT NULL,
			meow_status_code INTEGER NOT NULL DEFAULT 0,
			meow_response_body TEXT NOT NULL DEFAULT '',
			success INTEGER NOT NULL DEFAULT 0,
			error_message TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_push_logs_endpoint_id_created_at ON push_logs(endpoint_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_push_logs_created_at ON push_logs(created_at DESC)`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 5: Implement bootstrap and settings**

Create `internal/store/settings.go`:

```go
package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")

func (s *Store) GetSetting(ctx context.Context, key string) (string, error) {
	var value string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return value, err
}

func (s *Store) SetSetting(ctx context.Context, key, value string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO settings(key, value, updated_at)
		VALUES(?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
	`, key, value, time.Now().UTC())
	return err
}
```

Create `internal/store/bootstrap.go`:

```go
package store

import (
	"context"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type BootstrapOptions struct {
	AdminPassword    string
	MeowAPIBaseURL   string
	LogRetentionDays int
}

func (s *Store) Bootstrap(ctx context.Context, opts BootstrapOptions) error {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM admin_users`).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		hash, err := bcrypt.GenerateFromPassword([]byte(opts.AdminPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		if _, err := s.db.ExecContext(ctx, `
			INSERT INTO admin_users(password_hash, created_at, updated_at)
			VALUES(?, ?, ?)
		`, string(hash), time.Now().UTC(), time.Now().UTC()); err != nil {
			return err
		}
	}
	if err := s.insertSettingIfMissing(ctx, "meow_api_base_url", opts.MeowAPIBaseURL); err != nil {
		return err
	}
	return s.insertSettingIfMissing(ctx, "log_retention_days", strconv.Itoa(opts.LogRetentionDays))
}

func (s *Store) insertSettingIfMissing(ctx context.Context, key, value string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO settings(key, value, updated_at)
		VALUES(?, ?, ?)
		ON CONFLICT(key) DO NOTHING
	`, key, value, time.Now().UTC())
	return err
}
```

- [ ] **Step 6: Wire store initialization in main**

Modify `cmd/meowbridge/main.go`:

```go
package main

import (
	"context"
	"log"

	"github.com/sunnyhmz7010/meowbridge/internal/config"
	"github.com/sunnyhmz7010/meowbridge/internal/store"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	st, err := store.Open(ctx, cfg.DatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	if err := st.Migrate(ctx); err != nil {
		log.Fatal(err)
	}
	if err := st.Bootstrap(ctx, store.BootstrapOptions{
		AdminPassword:    cfg.AdminPassword,
		MeowAPIBaseURL:   cfg.MeowAPIBaseURL,
		LogRetentionDays: cfg.LogRetentionDays,
	}); err != nil {
		log.Fatal(err)
	}

	log.Printf("meowbridge starting on %s", cfg.HTTPAddr)
}
```

- [ ] **Step 7: Verify and commit**

Run:

```powershell
rtk go test ./internal/store ./...
rtk git status --short
```

Expected: all tests pass; only Task 2 files and `cmd/meowbridge/main.go` are changed.

Commit:

```powershell
rtk git add cmd internal go.mod go.sum
rtk git commit -m "添加数据库迁移和启动初始化"
```

---

### Task 3: Auth, JWT, and Secure Token Generation

**Files:**
- Create: `internal/auth/auth.go`
- Create: `internal/auth/auth_test.go`
- Create: `internal/token/token.go`
- Create: `internal/token/token_test.go`
- Modify: `internal/store/bootstrap.go`

**Interfaces:**
- Produces: `auth.HashPassword(password string) (string, error)`
- Produces: `auth.VerifyPassword(hash, password string) bool`
- Produces: `auth.IssueJWT(secret string, ttl time.Duration) (string, error)`
- Produces: `auth.VerifyJWT(secret, raw string) error`
- Produces: `token.Generate() (string, error)`

- [ ] **Step 1: Write failing auth tests**

Create `internal/auth/auth_test.go`:

```go
package auth

import (
	"testing"
	"time"
)

func TestPasswordHashAndVerify(t *testing.T) {
	hash, err := HashPassword("secret-password")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "secret-password" {
		t.Fatal("password was stored in plain text")
	}
	if !VerifyPassword(hash, "secret-password") {
		t.Fatal("expected password to verify")
	}
	if VerifyPassword(hash, "wrong-password") {
		t.Fatal("wrong password verified")
	}
}

func TestJWTIssueAndVerify(t *testing.T) {
	raw, err := IssueJWT("jwt-secret", time.Hour)
	if err != nil {
		t.Fatalf("IssueJWT: %v", err)
	}
	if err := VerifyJWT("jwt-secret", raw); err != nil {
		t.Fatalf("VerifyJWT: %v", err)
	}
	if err := VerifyJWT("other-secret", raw); err == nil {
		t.Fatal("expected wrong secret to fail")
	}
}
```

- [ ] **Step 2: Write failing token tests**

Create `internal/token/token_test.go`:

```go
package token

import "testing"

func TestGenerateReturnsUniqueURLSafeTokens(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		value, err := Generate()
		if err != nil {
			t.Fatalf("Generate: %v", err)
		}
		if len(value) < 32 {
			t.Fatalf("token too short: %q", value)
		}
		if seen[value] {
			t.Fatalf("duplicate token: %q", value)
		}
		seen[value] = true
	}
}
```

- [ ] **Step 3: Run tests and verify failure**

Run:

```powershell
rtk go test ./internal/auth ./internal/token
```

Expected: FAIL because packages are not implemented.

- [ ] **Step 4: Implement auth**

Create `internal/auth/auth.go`:

```go
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func VerifyPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func IssueJWT(secret string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Subject:   "admin",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func VerifyJWT(secret, raw string) error {
	parsed, err := jwt.ParseWithClaims(raw, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return err
	}
	if !parsed.Valid {
		return errors.New("invalid token")
	}
	return nil
}
```

- [ ] **Step 5: Implement token generation**

Create `internal/token/token.go`:

```go
package token

import (
	"crypto/rand"
	"encoding/base64"
)

func Generate() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
```

- [ ] **Step 6: Reuse auth hashing in bootstrap**

Modify `internal/store/bootstrap.go` to use `auth.HashPassword`:

```go
hash, err := auth.HashPassword(opts.AdminPassword)
if err != nil {
	return err
}
```

Add import:

```go
"github.com/sunnyhmz7010/meowbridge/internal/auth"
```

Remove direct `golang.org/x/crypto/bcrypt` import from `internal/store/bootstrap.go`.

- [ ] **Step 7: Verify and commit**

Run:

```powershell
rtk go test ./internal/auth ./internal/token ./internal/store ./...
rtk git status --short
```

Expected: all tests pass.

Commit:

```powershell
rtk git add internal go.mod go.sum
rtk git commit -m "添加管理员鉴权和安全令牌生成"
```

---

### Task 4: Endpoint CRUD and Settings Store

**Files:**
- Create: `internal/store/endpoints.go`
- Modify: `internal/store/settings.go`
- Modify: `internal/store/store_test.go`

**Interfaces:**
- Produces: `store.CreateEndpoint(ctx, input store.EndpointInput) (store.Endpoint, error)`
- Produces: `store.ListEndpoints(ctx) ([]store.Endpoint, error)`
- Produces: `store.GetEndpoint(ctx, id int64) (store.Endpoint, error)`
- Produces: `store.GetEndpointByToken(ctx, token string) (store.Endpoint, error)`
- Produces: `store.UpdateEndpoint(ctx, id int64, input store.EndpointUpdate) (store.Endpoint, error)`
- Produces: `store.SetEndpointActive(ctx, id int64, active bool) error`
- Produces: `store.ResetEndpointToken(ctx, id int64, newToken string) (store.Endpoint, error)`
- Produces: `store.DeleteEndpoint(ctx, id int64) error`
- Produces: `store.ListSettings(ctx) (map[string]string, error)`

- [ ] **Step 1: Add failing endpoint tests**

Append to `internal/store/store_test.go`:

```go
func TestEndpointCRUDKeepsMeowNicknameImmutable(t *testing.T) {
	ctx := context.Background()
	st, cleanup := openTestStore(t)
	defer cleanup()

	created, err := st.CreateEndpoint(ctx, EndpointInput{
		Name:          "GitHub",
		Token:         "token-1",
		MeowNickname:  "sunny",
		DefaultTitle:  "Default",
		MsgType:       "markdown",
		HTMLHeight:    300,
		DefaultURL:    "https://example.test",
		DefaultImgURL: "https://example.test/icon.png",
		Active:        true,
	})
	if err != nil {
		t.Fatalf("CreateEndpoint: %v", err)
	}

	updated, err := st.UpdateEndpoint(ctx, created.ID, EndpointUpdate{
		Name:          "GitHub Updated",
		DefaultTitle:  "Changed",
		MsgType:       "text",
		HTMLHeight:    200,
		DefaultURL:    "",
		DefaultImgURL: "",
		Active:        true,
	})
	if err != nil {
		t.Fatalf("UpdateEndpoint: %v", err)
	}
	if updated.MeowNickname != "sunny" {
		t.Fatalf("MeowNickname changed to %q", updated.MeowNickname)
	}

	if err := st.SetEndpointActive(ctx, created.ID, false); err != nil {
		t.Fatalf("SetEndpointActive: %v", err)
	}
	byToken, err := st.GetEndpointByToken(ctx, "token-1")
	if err != nil {
		t.Fatalf("GetEndpointByToken: %v", err)
	}
	if byToken.Active {
		t.Fatal("expected inactive endpoint")
	}

	reset, err := st.ResetEndpointToken(ctx, created.ID, "token-2")
	if err != nil {
		t.Fatalf("ResetEndpointToken: %v", err)
	}
	if reset.Token != "token-2" {
		t.Fatalf("Token = %q", reset.Token)
	}
}
```

- [ ] **Step 2: Run tests and verify failure**

Run:

```powershell
rtk go test ./internal/store
```

Expected: FAIL because endpoint methods do not exist.

- [ ] **Step 3: Implement endpoint inputs and CRUD**

Create `internal/store/endpoints.go`:

```go
package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type EndpointInput struct {
	Name          string
	Token         string
	MeowNickname  string
	DefaultTitle  string
	MsgType       string
	HTMLHeight    int
	DefaultURL    string
	DefaultImgURL string
	Active        bool
}

type EndpointUpdate struct {
	Name          string
	DefaultTitle  string
	MsgType       string
	HTMLHeight    int
	DefaultURL    string
	DefaultImgURL string
	Active        bool
}

func (s *Store) CreateEndpoint(ctx context.Context, input EndpointInput) (Endpoint, error) {
	now := time.Now().UTC()
	active := boolToInt(input.Active)
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO endpoints(name, token, meow_nickname, default_title, msg_type, html_height, default_url, default_img_url, active, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, input.Name, input.Token, input.MeowNickname, input.DefaultTitle, input.MsgType, input.HTMLHeight, input.DefaultURL, input.DefaultImgURL, active, now, now)
	if err != nil {
		return Endpoint{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Endpoint{}, err
	}
	return s.GetEndpoint(ctx, id)
}

func (s *Store) ListEndpoints(ctx context.Context) ([]Endpoint, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, token, meow_nickname, default_title, msg_type, html_height, default_url, default_img_url, active, created_at, updated_at
		FROM endpoints ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var endpoints []Endpoint
	for rows.Next() {
		ep, err := scanEndpoint(rows)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, ep)
	}
	return endpoints, rows.Err()
}

func (s *Store) GetEndpoint(ctx context.Context, id int64) (Endpoint, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, token, meow_nickname, default_title, msg_type, html_height, default_url, default_img_url, active, created_at, updated_at
		FROM endpoints WHERE id = ?
	`, id)
	return scanEndpoint(row)
}

func (s *Store) GetEndpointByToken(ctx context.Context, token string) (Endpoint, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, token, meow_nickname, default_title, msg_type, html_height, default_url, default_img_url, active, created_at, updated_at
		FROM endpoints WHERE token = ?
	`, token)
	return scanEndpoint(row)
}

func (s *Store) UpdateEndpoint(ctx context.Context, id int64, input EndpointUpdate) (Endpoint, error) {
	_, err := s.db.ExecContext(ctx, `
		UPDATE endpoints
		SET name = ?, default_title = ?, msg_type = ?, html_height = ?, default_url = ?, default_img_url = ?, active = ?, updated_at = ?
		WHERE id = ?
	`, input.Name, input.DefaultTitle, input.MsgType, input.HTMLHeight, input.DefaultURL, input.DefaultImgURL, boolToInt(input.Active), time.Now().UTC(), id)
	if err != nil {
		return Endpoint{}, err
	}
	return s.GetEndpoint(ctx, id)
}

func (s *Store) SetEndpointActive(ctx context.Context, id int64, active bool) error {
	_, err := s.db.ExecContext(ctx, `UPDATE endpoints SET active = ?, updated_at = ? WHERE id = ?`, boolToInt(active), time.Now().UTC(), id)
	return err
}

func (s *Store) ResetEndpointToken(ctx context.Context, id int64, newToken string) (Endpoint, error) {
	_, err := s.db.ExecContext(ctx, `UPDATE endpoints SET token = ?, updated_at = ? WHERE id = ?`, newToken, time.Now().UTC(), id)
	if err != nil {
		return Endpoint{}, err
	}
	return s.GetEndpoint(ctx, id)
}

func (s *Store) DeleteEndpoint(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM endpoints WHERE id = ?`, id)
	return err
}

type endpointScanner interface {
	Scan(dest ...any) error
}

func scanEndpoint(row endpointScanner) (Endpoint, error) {
	var ep Endpoint
	var active int
	err := row.Scan(&ep.ID, &ep.Name, &ep.Token, &ep.MeowNickname, &ep.DefaultTitle, &ep.MsgType, &ep.HTMLHeight, &ep.DefaultURL, &ep.DefaultImgURL, &active, &ep.CreatedAt, &ep.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Endpoint{}, ErrNotFound
	}
	if err != nil {
		return Endpoint{}, err
	}
	ep.Active = active == 1
	return ep, nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
```

- [ ] **Step 4: Add settings listing**

Append to `internal/store/settings.go`:

```go
func (s *Store) ListSettings(ctx context.Context) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT key, value FROM settings ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		values[key] = value
	}
	return values, rows.Err()
}
```

- [ ] **Step 5: Verify and commit**

Run:

```powershell
rtk go test ./internal/store ./...
rtk git status --short
```

Expected: all tests pass.

Commit:

```powershell
rtk git add internal
rtk git commit -m "添加推送接口存储能力"
```

---

### Task 5: Webhook Parser Chain and Field Merge

**Files:**
- Create: `internal/webhook/types.go`
- Create: `internal/webhook/parsers.go`
- Create: `internal/webhook/providers.go`
- Create: `internal/webhook/merge.go`
- Create: `internal/webhook/parsers_test.go`
- Create: `internal/webhook/merge_test.go`

**Interfaces:**
- Produces: `webhook.Parse(input webhook.ParseInput) (webhook.ParsedMessage, error)`
- Produces: `webhook.Merge(parsed webhook.ParsedMessage, endpoint webhook.EndpointDefaults, query webhook.QueryOverrides) (webhook.FinalMessage, error)`

- [ ] **Step 1: Write failing parser tests**

Create `internal/webhook/parsers_test.go`:

```go
package webhook

import (
	"net/http"
	"testing"
)

func TestParseGitHubPullRequest(t *testing.T) {
	payload := []byte(`{"action":"opened","repository":{"full_name":"sunny/meowbridge","html_url":"https://github.com/sunny/meowbridge"},"pull_request":{"title":"Add webhook","body":"Adds support","html_url":"https://github.com/sunny/meowbridge/pull/1"}}`)
	parsed, err := Parse(ParseInput{
		Headers: http.Header{"X-GitHub-Event": []string{"pull_request"}},
		Body:    payload,
	})
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if parsed.SourceType != "github_pr" || parsed.Title != "Add webhook" || parsed.Msg != "Adds support" {
		t.Fatalf("parsed = %#v", parsed)
	}
}

func TestParseKnownProvidersAndFallback(t *testing.T) {
	cases := []struct {
		name       string
		body       string
		wantSource string
	}{
		{"github_action", `{"workflow_run":{"event":"push","head_commit":{"message":"build passed"},"html_url":"https://github.test/run"}}`, "github_action"},
		{"github", `{"action":"push","repository":{"full_name":"sunny/meowbridge","html_url":"https://github.test/repo"}}`, "github"},
		{"jenkins", `{"project":{"name":"build"},"build":{"full_display_url":"https://jenkins.test/1"}}`, "jenkins"},
		{"grafana", `{"alerts":[{"labels":{"alertname":"CPUHigh"},"annotations":{"message":"CPU high"}}],"externalURL":"https://grafana.test"}`, "grafana"},
		{"prometheus", `{"receiver":"default","alerts":[{"labels":{"alertname":"DiskFull"},"annotations":{"description":"disk full"}}]}`, "prometheus"},
		{"zabbix", `{"trigger":{"description":"Host down"},"event":{"description":"host unavailable"}}`, "zabbix"},
		{"gotify", `{"title":"Gotify title","message":"Gotify message"}`, "gotify"},
		{"emby", `{"Title":"Playback started","Description":"Movie"}`, "emby"},
		{"generic", `{"title":"Generic title","message":"Generic message"}`, "generic"},
		{"fallback", `{"unexpected":{"nested":true}}`, "fallback"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := Parse(ParseInput{Headers: http.Header{}, Body: []byte(tc.body)})
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if parsed.SourceType != tc.wantSource {
				t.Fatalf("SourceType = %q", parsed.SourceType)
			}
			if parsed.Msg == "" {
				t.Fatalf("Msg is empty: %#v", parsed)
			}
		})
	}
}
```

- [ ] **Step 2: Write failing merge tests**

Create `internal/webhook/merge_test.go`:

```go
package webhook

import "testing"

func TestMergeFieldPrecedence(t *testing.T) {
	final, err := Merge(
		ParsedMessage{
			Title:   "parsed title",
			Msg:     "parsed msg",
			URL:     "https://parsed.test",
			ImgURL:  "https://parsed.test/icon.png",
			MsgType: "markdown",
		},
		EndpointDefaults{
			DefaultTitle:  "default title",
			MsgType:       "text",
			HTMLHeight:    200,
			DefaultURL:    "https://default.test",
			DefaultImgURL: "https://default.test/icon.png",
		},
		QueryOverrides{
			Title:      "query title",
			MsgType:    "html",
			HTMLHeight: 500,
		},
	)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if final.Title != "query title" || final.Msg != "parsed msg" || final.MsgType != "html" || final.HTMLHeight != 500 {
		t.Fatalf("final = %#v", final)
	}
}

func TestMergeRejectsEmptyMessage(t *testing.T) {
	_, err := Merge(ParsedMessage{}, EndpointDefaults{}, QueryOverrides{})
	if err == nil {
		t.Fatal("expected empty message error")
	}
}
```

- [ ] **Step 3: Run tests and verify failure**

Run:

```powershell
rtk go test ./internal/webhook
```

Expected: FAIL because webhook package does not exist.

- [ ] **Step 4: Implement webhook types**

Create `internal/webhook/types.go`:

```go
package webhook

import "net/http"

type ParseInput struct {
	Headers http.Header
	Body    []byte
}

type ParsedMessage struct {
	SourceType string
	Title      string
	Msg        string
	URL        string
	ImgURL     string
	MsgType    string
}

type EndpointDefaults struct {
	DefaultTitle  string
	MsgType       string
	HTMLHeight    int
	DefaultURL    string
	DefaultImgURL string
}

type QueryOverrides struct {
	Title      string
	MsgType    string
	HTMLHeight int
	URL        string
	ImgURL     string
}

type FinalMessage struct {
	Title      string
	Msg        string
	URL        string
	ImgURL     string
	MsgType    string
	HTMLHeight int
}
```

- [ ] **Step 5: Implement parser chain**

Create `internal/webhook/parsers.go`:

```go
package webhook

import (
	"encoding/json"
	"errors"
)

type parser func(ParseInput, map[string]any) (ParsedMessage, bool)

func Parse(input ParseInput) (ParsedMessage, error) {
	var payload map[string]any
	if err := json.Unmarshal(input.Body, &payload); err != nil {
		return ParsedMessage{}, errors.New("invalid json payload")
	}
	for _, parse := range []parser{
		parseGitHubPR,
		parseGitHubAction,
		parseGitHub,
		parseJenkins,
		parseGrafana,
		parsePrometheus,
		parseZabbix,
		parseGotify,
		parseEmby,
		parseGeneric,
		parseFallback,
	} {
		if parsed, ok := parse(input, payload); ok {
			return parsed, nil
		}
	}
	return ParsedMessage{}, errors.New("payload could not be parsed")
}
```

- [ ] **Step 6: Implement provider parsers**

Create `internal/webhook/providers.go` with focused helpers:

```go
package webhook

import (
	"encoding/json"
	"strings"
)

func parseGitHubPR(input ParseInput, payload map[string]any) (ParsedMessage, bool) {
	if input.Headers.Get("X-GitHub-Event") != "pull_request" && payload["pull_request"] == nil {
		return ParsedMessage{}, false
	}
	pr, _ := payload["pull_request"].(map[string]any)
	return ParsedMessage{
		SourceType: "github_pr",
		Title:      stringValue(pr, "title"),
		Msg:        firstNonEmpty(stringValue(pr, "body"), stringValue(payload, "action")),
		URL:        stringValue(pr, "html_url"),
		MsgType:    "markdown",
	}, true
}

func parseGitHubAction(input ParseInput, payload map[string]any) (ParsedMessage, bool) {
	run, ok := payload["workflow_run"].(map[string]any)
	if !ok {
		return ParsedMessage{}, false
	}
	commit, _ := run["head_commit"].(map[string]any)
	return ParsedMessage{
		SourceType: "github_action",
		Title:      firstNonEmpty(stringValue(run, "event"), "GitHub Actions"),
		Msg:        firstNonEmpty(stringValue(commit, "message"), stringValue(run, "name")),
		URL:        stringValue(run, "html_url"),
		MsgType:    "markdown",
	}, true
}

func parseGitHub(input ParseInput, payload map[string]any) (ParsedMessage, bool) {
	repo, ok := payload["repository"].(map[string]any)
	if !ok && input.Headers.Get("X-GitHub-Event") == "" {
		return ParsedMessage{}, false
	}
	return ParsedMessage{
		SourceType: "github",
		Title:      firstNonEmpty(stringValue(payload, "action"), input.Headers.Get("X-GitHub-Event"), "GitHub Webhook"),
		Msg:        firstNonEmpty(stringValue(repo, "full_name"), compactJSON(payload)),
		URL:        stringValue(repo, "html_url"),
		MsgType:    "text",
	}, true
}

func parseJenkins(input ParseInput, payload map[string]any) (ParsedMessage, bool) {
	project, ok := payload["project"].(map[string]any)
	if !ok {
		return ParsedMessage{}, false
	}
	build, _ := payload["build"].(map[string]any)
	return ParsedMessage{SourceType: "jenkins", Title: stringValue(project, "name"), Msg: firstNonEmpty(stringValue(build, "full_display_url"), compactJSON(payload)), URL: stringValue(build, "full_display_url"), MsgType: "text"}, true
}

func parseGrafana(input ParseInput, payload map[string]any) (ParsedMessage, bool) {
	if payload["alerts"] == nil || payload["receiver"] != nil {
		return ParsedMessage{}, false
	}
	return parseAlertPayload("grafana", payload, "message")
}

func parsePrometheus(input ParseInput, payload map[string]any) (ParsedMessage, bool) {
	if payload["alerts"] == nil || payload["receiver"] == nil {
		return ParsedMessage{}, false
	}
	return parseAlertPayload("prometheus", payload, "description")
}

func parseAlertPayload(source string, payload map[string]any, annotationKey string) (ParsedMessage, bool) {
	alert := firstAlert(payload)
	labels, _ := alert["labels"].(map[string]any)
	annotations, _ := alert["annotations"].(map[string]any)
	return ParsedMessage{SourceType: source, Title: stringValue(labels, "alertname"), Msg: firstNonEmpty(stringValue(annotations, annotationKey), compactJSON(payload)), URL: stringValue(payload, "externalURL"), MsgType: "markdown"}, true
}

func parseZabbix(input ParseInput, payload map[string]any) (ParsedMessage, bool) {
	trigger, ok := payload["trigger"].(map[string]any)
	if !ok {
		return ParsedMessage{}, false
	}
	event, _ := payload["event"].(map[string]any)
	return ParsedMessage{SourceType: "zabbix", Title: stringValue(trigger, "description"), Msg: firstNonEmpty(stringValue(event, "description"), compactJSON(payload)), MsgType: "markdown"}, true
}

func parseGotify(input ParseInput, payload map[string]any) (ParsedMessage, bool) {
	if _, ok := payload["message"]; !ok {
		return ParsedMessage{}, false
	}
	return ParsedMessage{SourceType: "gotify", Title: stringValue(payload, "title"), Msg: stringValue(payload, "message"), MsgType: "markdown"}, true
}

func parseEmby(input ParseInput, payload map[string]any) (ParsedMessage, bool) {
	if _, ok := payload["Title"]; !ok {
		return ParsedMessage{}, false
	}
	return ParsedMessage{SourceType: "emby", Title: stringValue(payload, "Title"), Msg: firstNonEmpty(stringValue(payload, "Description"), compactJSON(payload)), MsgType: "text"}, true
}

func parseGeneric(input ParseInput, payload map[string]any) (ParsedMessage, bool) {
	msg := firstNonEmpty(stringValue(payload, "msg"), stringValue(payload, "message"), stringValue(payload, "text"), stringValue(payload, "content"))
	if msg == "" {
		return ParsedMessage{}, false
	}
	return ParsedMessage{SourceType: "generic", Title: stringValue(payload, "title"), Msg: msg, URL: stringValue(payload, "url"), ImgURL: stringValue(payload, "imgUrl"), MsgType: stringValue(payload, "msgType")}, true
}

func parseFallback(input ParseInput, payload map[string]any) (ParsedMessage, bool) {
	return ParsedMessage{SourceType: "fallback", Title: "Webhook", Msg: prettyJSON(payload), MsgType: "markdown"}, true
}

func firstAlert(payload map[string]any) map[string]any {
	alerts, _ := payload["alerts"].([]any)
	if len(alerts) == 0 {
		return map[string]any{}
	}
	first, _ := alerts[0].(map[string]any)
	return first
}

func stringValue(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	value, _ := m[key].(string)
	return strings.TrimSpace(value)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func compactJSON(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func prettyJSON(value any) string {
	data, _ := json.MarshalIndent(value, "", "  ")
	return string(data)
}
```

- [ ] **Step 7: Implement merge rules**

Create `internal/webhook/merge.go`:

```go
package webhook

import (
	"errors"
	"strings"
)

func Merge(parsed ParsedMessage, endpoint EndpointDefaults, query QueryOverrides) (FinalMessage, error) {
	msg := strings.TrimSpace(parsed.Msg)
	if msg == "" {
		return FinalMessage{}, errors.New("message is required")
	}
	final := FinalMessage{
		Title:      firstNonEmpty(query.Title, parsed.Title, endpoint.DefaultTitle, "Meow"),
		Msg:        msg,
		URL:        firstNonEmpty(query.URL, parsed.URL, endpoint.DefaultURL),
		ImgURL:     firstNonEmpty(query.ImgURL, parsed.ImgURL, endpoint.DefaultImgURL),
		MsgType:    firstNonEmpty(query.MsgType, parsed.MsgType, endpoint.MsgType, "text"),
		HTMLHeight: firstPositive(query.HTMLHeight, endpoint.HTMLHeight, 200),
	}
	return final, nil
}

func firstPositive(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}
```

- [ ] **Step 8: Verify and commit**

Run:

```powershell
rtk go test ./internal/webhook ./...
rtk git status --short
```

Expected: all tests pass.

Commit:

```powershell
rtk git add internal
rtk git commit -m "添加通用 Webhook 解析器链"
```

---

### Task 6: MeoW Client and Push Log Store

**Files:**
- Create: `internal/meow/client.go`
- Create: `internal/meow/client_test.go`
- Create: `internal/store/logs.go`
- Modify: `internal/store/store_test.go`

**Interfaces:**
- Produces: `meow.Client.Push(ctx context.Context, req meow.PushRequest) (meow.PushResponse, error)`
- Produces: `store.CreatePushLog(ctx, input store.PushLogInput) (int64, error)`
- Produces: `store.ListPushLogs(ctx) ([]store.PushLog, error)`
- Produces: `store.GetPushLog(ctx, id int64) (store.PushLog, error)`
- Produces: `store.CleanupPushLogs(ctx, before time.Time) (int64, error)`

- [ ] **Step 1: Write failing MeoW client tests**

Create `internal/meow/client_test.go`:

```go
package meow

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPushSendsJSONToNicknamePath(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.URL.Query().Get("msgType") != "html" || r.URL.Query().Get("htmlHeight") != "500" {
			t.Fatalf("query = %s", r.URL.RawQuery)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := New(server.URL, time.Second)
	resp, err := client.Push(context.Background(), PushRequest{
		Nickname:   "sunny",
		Title:      "title",
		Msg:        "message",
		MsgType:    "html",
		HTMLHeight: 500,
	})
	if err != nil {
		t.Fatalf("Push: %v", err)
	}
	if gotPath != "/sunny" {
		t.Fatalf("path = %q", gotPath)
	}
	if resp.StatusCode != http.StatusOK || resp.Body != `{"ok":true}` {
		t.Fatalf("resp = %#v", resp)
	}
}

func TestPushTreatsNon2xxAsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("bad gateway"))
	}))
	defer server.Close()

	client := New(server.URL, time.Second)
	resp, err := client.Push(context.Background(), PushRequest{Nickname: "sunny", Msg: "message", MsgType: "text"})
	if err == nil {
		t.Fatal("expected upstream error")
	}
	if resp.StatusCode != http.StatusBadGateway || resp.Body != "bad gateway" {
		t.Fatalf("resp = %#v", resp)
	}
}
```

- [ ] **Step 2: Add failing push log tests**

Append to `internal/store/store_test.go`:

```go
func TestPushLogCreateListDetailCleanup(t *testing.T) {
	ctx := context.Background()
	st, cleanup := openTestStore(t)
	defer cleanup()

	id, err := st.CreatePushLog(ctx, PushLogInput{
		EndpointID:       1,
		EndpointName:     "GitHub",
		Token:            "token-1",
		SourceType:       "github",
		RequestMethod:    "POST",
		RequestHeaders:   `{"content-type":["application/json"]}`,
		RequestQuery:     `{"title":["override"]}`,
		RequestPayload:   `{"message":"hello"}`,
		ParsedTitle:      "title",
		ParsedMsg:        "message",
		ParsedMsgType:    "text",
		MeowStatusCode:   200,
		MeowResponseBody: "ok",
		Success:          true,
		ErrorMessage:     "",
	})
	if err != nil {
		t.Fatalf("CreatePushLog: %v", err)
	}
	log, err := st.GetPushLog(ctx, id)
	if err != nil {
		t.Fatalf("GetPushLog: %v", err)
	}
	if log.RequestPayload == "" || !log.Success {
		t.Fatalf("log = %#v", log)
	}
}
```

- [ ] **Step 3: Run tests and verify failure**

Run:

```powershell
rtk go test ./internal/meow ./internal/store
```

Expected: FAIL because client and log methods are missing.

- [ ] **Step 4: Implement MeoW client**

Create `internal/meow/client.go`:

```go
package meow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const maxResponseBytes = 16 * 1024

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type PushRequest struct {
	Nickname   string
	Title      string
	Msg        string
	URL        string
	ImgURL     string
	MsgType    string
	HTMLHeight int
}

type PushResponse struct {
	StatusCode int
	Body       string
}

func New(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Push(ctx context.Context, req PushRequest) (PushResponse, error) {
	target, err := url.Parse(c.baseURL + "/" + url.PathEscape(req.Nickname))
	if err != nil {
		return PushResponse{}, err
	}
	query := target.Query()
	query.Set("msgType", req.MsgType)
	if req.MsgType == "html" && req.HTMLHeight > 0 {
		query.Set("htmlHeight", strconv.Itoa(req.HTMLHeight))
	}
	target.RawQuery = query.Encode()

	body := map[string]string{
		"title": req.Title,
		"msg":   req.Msg,
	}
	if req.URL != "" {
		body["url"] = req.URL
	}
	if req.ImgURL != "" {
		body["imgUrl"] = req.ImgURL
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return PushResponse{}, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, target.String(), bytes.NewReader(encoded))
	if err != nil {
		return PushResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return PushResponse{}, err
	}
	defer httpResp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(httpResp.Body, maxResponseBytes))
	resp := PushResponse{StatusCode: httpResp.StatusCode, Body: string(respBody)}
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return resp, fmt.Errorf("meow upstream returned %d", httpResp.StatusCode)
	}
	return resp, nil
}
```

- [ ] **Step 5: Implement push log store**

Create `internal/store/logs.go`:

```go
package store

import (
	"context"
	"time"
)

type PushLogInput struct {
	EndpointID       int64
	EndpointName     string
	Token            string
	SourceType       string
	RequestMethod    string
	RequestHeaders   string
	RequestQuery     string
	RequestPayload   string
	ParsedTitle      string
	ParsedMsg        string
	ParsedMsgType    string
	MeowStatusCode   int
	MeowResponseBody string
	Success          bool
	ErrorMessage     string
}

func (s *Store) CreatePushLog(ctx context.Context, input PushLogInput) (int64, error) {
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO push_logs(endpoint_id, endpoint_name, token, source_type, request_method, request_headers, request_query, request_payload, parsed_title, parsed_msg, parsed_msg_type, meow_status_code, meow_response_body, success, error_message, created_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, input.EndpointID, input.EndpointName, input.Token, input.SourceType, input.RequestMethod, input.RequestHeaders, input.RequestQuery, input.RequestPayload, input.ParsedTitle, input.ParsedMsg, input.ParsedMsgType, input.MeowStatusCode, input.MeowResponseBody, boolToInt(input.Success), input.ErrorMessage, time.Now().UTC())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) GetPushLog(ctx context.Context, id int64) (PushLog, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, endpoint_id, endpoint_name, token, source_type, request_method, request_headers, request_query, request_payload, parsed_title, parsed_msg, parsed_msg_type, meow_status_code, meow_response_body, success, error_message, created_at
		FROM push_logs WHERE id = ?
	`, id)
	return scanPushLog(row)
}

func (s *Store) ListPushLogs(ctx context.Context) ([]PushLog, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, endpoint_id, endpoint_name, token, source_type, request_method, request_headers, request_query, request_payload, parsed_title, parsed_msg, parsed_msg_type, meow_status_code, meow_response_body, success, error_message, created_at
		FROM push_logs ORDER BY id DESC LIMIT 200
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []PushLog
	for rows.Next() {
		log, err := scanPushLog(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

func (s *Store) CleanupPushLogs(ctx context.Context, before time.Time) (int64, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM push_logs WHERE created_at < ?`, before.UTC())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

type pushLogScanner interface {
	Scan(dest ...any) error
}

func scanPushLog(row pushLogScanner) (PushLog, error) {
	var log PushLog
	var success int
	err := row.Scan(&log.ID, &log.EndpointID, &log.EndpointName, &log.Token, &log.SourceType, &log.RequestMethod, &log.RequestHeaders, &log.RequestQuery, &log.RequestPayload, &log.ParsedTitle, &log.ParsedMsg, &log.ParsedMsgType, &log.MeowStatusCode, &log.MeowResponseBody, &success, &log.ErrorMessage, &log.CreatedAt)
	if err != nil {
		return PushLog{}, err
	}
	log.Success = success == 1
	return log, nil
}
```

- [ ] **Step 6: Verify and commit**

Run:

```powershell
rtk go test ./internal/meow ./internal/store ./...
rtk git status --short
```

Expected: all tests pass.

Commit:

```powershell
rtk git add internal
rtk git commit -m "添加 MeoW 客户端和推送日志"
```

---

### Task 7: Public Webhook HTTP Handler and Router

**Files:**
- Create: `internal/httpapi/types.go`
- Create: `internal/httpapi/router.go`
- Create: `internal/httpapi/webhook.go`
- Create: `internal/httpapi/webhook_test.go`
- Modify: `cmd/meowbridge/main.go`

**Interfaces:**
- Consumes: store endpoint/log methods, webhook parser/merge, meow client.
- Produces: `httpapi.NewRouter(deps httpapi.Dependencies) http.Handler`
- Produces public routes: `POST /webhook/{token}`, `GET /verify/{token}`

- [ ] **Step 1: Write failing Webhook integration tests**

Create `internal/httpapi/webhook_test.go`:

```go
package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sunnyhmz7010/meowbridge/internal/config"
	"github.com/sunnyhmz7010/meowbridge/internal/meow"
	"github.com/sunnyhmz7010/meowbridge/internal/store"
)

func TestWebhookSuccessWritesLog(t *testing.T) {
	ctx := context.Background()
	st := newHTTPTestStore(t)
	endpoint, err := st.CreateEndpoint(ctx, store.EndpointInput{Name: "GitHub", Token: "token-1", MeowNickname: "sunny", MsgType: "text", HTMLHeight: 200, Active: true})
	if err != nil {
		t.Fatalf("CreateEndpoint: %v", err)
	}
	if endpoint.ID == 0 {
		t.Fatal("endpoint id was not set")
	}

	meowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer meowServer.Close()

	router := NewRouter(Dependencies{
		Store:      st,
		Config:     config.Config{JWTSecret: "secret", MeowTimeout: time.Second},
		MeowClient: meow.New(meowServer.URL, time.Second),
	})

	req := httptest.NewRequest(http.MethodPost, "/webhook/token-1", bytes.NewBufferString(`{"message":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rr.Code, rr.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if body["ok"] != true || body["log_id"].(float64) == 0 {
		t.Fatalf("body = %#v", body)
	}
}

func TestWebhookReturns404ForUnknownToken(t *testing.T) {
	st := newHTTPTestStore(t)
	router := NewRouter(Dependencies{Store: st, Config: config.Config{JWTSecret: "secret", MeowTimeout: time.Second}, MeowClient: meow.New("http://127.0.0.1:1", time.Millisecond)})

	req := httptest.NewRequest(http.MethodPost, "/webhook/missing", bytes.NewBufferString(`{"message":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rr.Code)
	}
}

func newHTTPTestStore(t *testing.T) *store.Store {
	t.Helper()
	ctx := context.Background()
	st, err := store.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	if err := st.Migrate(ctx); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return st
}
```

- [ ] **Step 2: Run tests and verify failure**

Run:

```powershell
rtk go test ./internal/httpapi
```

Expected: FAIL because HTTP API package does not exist.

- [ ] **Step 3: Implement HTTP API dependency types and router**

Create `internal/httpapi/types.go`:

```go
package httpapi

import (
	"github.com/sunnyhmz7010/meowbridge/internal/config"
	"github.com/sunnyhmz7010/meowbridge/internal/meow"
	"github.com/sunnyhmz7010/meowbridge/internal/store"
)

type Dependencies struct {
	Store      *store.Store
	Config     config.Config
	MeowClient *meow.Client
}

type API struct {
	deps Dependencies
}
```

Create `internal/httpapi/router.go`:

```go
package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(deps Dependencies) http.Handler {
	api := &API{deps: deps}
	r := chi.NewRouter()
	r.Post("/webhook/{token}", api.handleWebhook)
	r.Get("/verify/{token}", api.handleVerifyToken)
	return r
}
```

- [ ] **Step 4: Implement Webhook handler**

Create `internal/httpapi/webhook.go`:

```go
package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sunnyhmz7010/meowbridge/internal/meow"
	"github.com/sunnyhmz7010/meowbridge/internal/respond"
	"github.com/sunnyhmz7010/meowbridge/internal/store"
	"github.com/sunnyhmz7010/meowbridge/internal/webhook"
)

func (api *API) handleVerifyToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	ep, err := api.deps.Store.GetEndpointByToken(r.Context(), token)
	if errors.Is(err, store.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "token not found")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	respond.OK(w, map[string]any{"name": ep.Name, "active": ep.Active})
}

func (api *API) handleWebhook(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	ep, err := api.deps.Store.GetEndpointByToken(r.Context(), token)
	if errors.Is(err, store.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "token not found")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !ep.Active {
		api.writeWebhookLog(r, ep, webhook.ParsedMessage{}, webhook.FinalMessage{}, 0, "", false, "endpoint is disabled", "")
		respond.Error(w, http.StatusForbidden, "endpoint is disabled")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1024*1024))
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		api.writeWebhookLog(r, ep, webhook.ParsedMessage{}, webhook.FinalMessage{}, 0, "", false, "request body is empty", "")
		respond.Error(w, http.StatusBadRequest, "request body is empty")
		return
	}

	parsed, status, parseErr := api.parseWebhookRequest(r, body)
	if parseErr != nil {
		api.writeWebhookLog(r, ep, parsed, webhook.FinalMessage{}, 0, "", false, parseErr.Error(), string(body))
		respond.Error(w, status, parseErr.Error())
		return
	}

	final, err := webhook.Merge(parsed, webhook.EndpointDefaults{
		DefaultTitle:  ep.DefaultTitle,
		MsgType:       ep.MsgType,
		HTMLHeight:    ep.HTMLHeight,
		DefaultURL:    ep.DefaultURL,
		DefaultImgURL: ep.DefaultImgURL,
	}, queryOverrides(r))
	if err != nil {
		api.writeWebhookLog(r, ep, parsed, webhook.FinalMessage{}, 0, "", false, err.Error(), string(body))
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	meowResp, pushErr := api.deps.MeowClient.Push(r.Context(), meow.PushRequest{
		Nickname:   ep.MeowNickname,
		Title:      final.Title,
		Msg:        final.Msg,
		URL:        final.URL,
		ImgURL:     final.ImgURL,
		MsgType:    final.MsgType,
		HTMLHeight: final.HTMLHeight,
	})
	if pushErr != nil {
		logID := api.writeWebhookLog(r, ep, parsed, final, meowResp.StatusCode, meowResp.Body, false, pushErr.Error(), string(body))
		_ = logID
		respond.Error(w, http.StatusBadGateway, pushErr.Error())
		return
	}
	logID := api.writeWebhookLog(r, ep, parsed, final, meowResp.StatusCode, meowResp.Body, true, "", string(body))
	respond.WebhookOK(w, logID)
}

func (api *API) parseWebhookRequest(r *http.Request, body []byte) (webhook.ParsedMessage, int, error) {
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "text/plain") {
		return webhook.ParsedMessage{SourceType: "text_plain", Msg: strings.TrimSpace(string(body)), MsgType: "text"}, http.StatusOK, nil
	}
	parsed, err := webhook.Parse(webhook.ParseInput{Headers: r.Header, Body: body})
	if err != nil {
		return webhook.ParsedMessage{}, http.StatusBadRequest, err
	}
	return parsed, http.StatusOK, nil
}

func queryOverrides(r *http.Request) webhook.QueryOverrides {
	height, _ := strconv.Atoi(r.URL.Query().Get("htmlHeight"))
	return webhook.QueryOverrides{
		Title:      r.URL.Query().Get("title"),
		MsgType:    r.URL.Query().Get("msgType"),
		HTMLHeight: height,
		URL:        r.URL.Query().Get("url"),
		ImgURL:     r.URL.Query().Get("imgUrl"),
	}
}

func (api *API) writeWebhookLog(r *http.Request, ep store.Endpoint, parsed webhook.ParsedMessage, final webhook.FinalMessage, statusCode int, responseBody string, success bool, errorMessage string, payload string) int64 {
	headers, _ := json.Marshal(r.Header)
	query, _ := json.Marshal(r.URL.Query())
	id, _ := api.deps.Store.CreatePushLog(r.Context(), store.PushLogInput{
		EndpointID:       ep.ID,
		EndpointName:     ep.Name,
		Token:            ep.Token,
		SourceType:       parsed.SourceType,
		RequestMethod:    r.Method,
		RequestHeaders:   string(headers),
		RequestQuery:     string(query),
		RequestPayload:   payload,
		ParsedTitle:      final.Title,
		ParsedMsg:        final.Msg,
		ParsedMsgType:    final.MsgType,
		MeowStatusCode:   statusCode,
		MeowResponseBody: responseBody,
		Success:          success,
		ErrorMessage:     errorMessage,
	})
	return id
}
```

- [ ] **Step 5: Wire router into main**

Modify `cmd/meowbridge/main.go` after bootstrap:

```go
meowClient := meow.New(cfg.MeowAPIBaseURL, cfg.MeowTimeout)
router := httpapi.NewRouter(httpapi.Dependencies{
	Store:      st,
	Config:     cfg,
	MeowClient: meowClient,
})
log.Printf("meowbridge starting on %s", cfg.HTTPAddr)
if err := http.ListenAndServe(cfg.HTTPAddr, router); err != nil {
	log.Fatal(err)
}
```

Add imports:

```go
"net/http"

"github.com/sunnyhmz7010/meowbridge/internal/httpapi"
"github.com/sunnyhmz7010/meowbridge/internal/meow"
```

- [ ] **Step 6: Verify and commit**

Run:

```powershell
rtk go test ./internal/httpapi ./...
rtk git status --short
```

Expected: all tests pass.

Commit:

```powershell
rtk git add cmd internal go.mod go.sum
rtk git commit -m "添加公开 Webhook 推送接口"
```

---

### Task 8: Admin API Handlers

**Files:**
- Create: `internal/httpapi/admin.go`
- Create: `internal/httpapi/admin_test.go`
- Modify: `internal/httpapi/router.go`
- Modify: `internal/store/bootstrap.go`

**Interfaces:**
- Consumes: `auth`, `token`, `store` endpoint/settings/log APIs.
- Produces authenticated admin routes listed in the design spec.

- [ ] **Step 1: Write failing admin API tests**

Create `internal/httpapi/admin_test.go`:

```go
package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sunnyhmz7010/meowbridge/internal/config"
	"github.com/sunnyhmz7010/meowbridge/internal/meow"
	"github.com/sunnyhmz7010/meowbridge/internal/store"
)

func TestAdminLoginAndEndpointCRUD(t *testing.T) {
	ctx := context.Background()
	st := newHTTPTestStore(t)
	if err := st.Bootstrap(ctx, store.BootstrapOptions{AdminPassword: "admin-password", MeowAPIBaseURL: "https://meow.example.test", LogRetentionDays: 14}); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	router := NewRouter(Dependencies{Store: st, Config: config.Config{JWTSecret: "jwt-secret", MeowTimeout: time.Second}, MeowClient: meow.New("http://127.0.0.1:1", time.Millisecond)})

	loginReq := httptest.NewRequest(http.MethodPost, "/api/admin/login", bytes.NewBufferString(`{"password":"admin-password"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRR := httptest.NewRecorder()
	router.ServeHTTP(loginRR, loginReq)
	if loginRR.Code != http.StatusOK {
		t.Fatalf("login status = %d body = %s", loginRR.Code, loginRR.Body.String())
	}
	var loginBody struct {
		OK   bool `json:"ok"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginRR.Body.Bytes(), &loginBody); err != nil {
		t.Fatalf("login json: %v", err)
	}
	if loginBody.Data.Token == "" {
		t.Fatal("missing token")
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/endpoints", bytes.NewBufferString(`{"name":"GitHub","meow_nickname":"sunny","default_title":"GitHub","msg_type":"text","html_height":200,"active":true}`))
	createReq.Header.Set("Authorization", "Bearer "+loginBody.Data.Token)
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusOK {
		t.Fatalf("create status = %d body = %s", createRR.Code, createRR.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/admin/endpoints", nil)
	listReq.Header.Set("Authorization", "Bearer "+loginBody.Data.Token)
	listRR := httptest.NewRecorder()
	router.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("list status = %d", listRR.Code)
	}
}

func TestAdminRoutesRequireJWT(t *testing.T) {
	st := newHTTPTestStore(t)
	router := NewRouter(Dependencies{Store: st, Config: config.Config{JWTSecret: "jwt-secret", MeowTimeout: time.Second}, MeowClient: meow.New("http://127.0.0.1:1", time.Millisecond)})

	req := httptest.NewRequest(http.MethodGet, "/api/admin/endpoints", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", rr.Code)
	}
}
```

- [ ] **Step 2: Run tests and verify failure**

Run:

```powershell
rtk go test ./internal/httpapi
```

Expected: FAIL because admin routes are missing.

- [ ] **Step 3: Add admin route group**

Modify `internal/httpapi/router.go`:

```go
r.Route("/api/admin", func(r chi.Router) {
	r.Post("/login", api.handleLogin)
	r.Group(func(r chi.Router) {
		r.Use(api.requireAdmin)
		r.Get("/endpoints", api.handleListEndpoints)
		r.Post("/endpoints", api.handleCreateEndpoint)
		r.Get("/endpoints/{id}", api.handleGetEndpoint)
		r.Put("/endpoints/{id}", api.handleUpdateEndpoint)
		r.Delete("/endpoints/{id}", api.handleDeleteEndpoint)
		r.Post("/endpoints/{id}/reset-token", api.handleResetEndpointToken)
		r.Patch("/endpoints/{id}/active", api.handleSetEndpointActive)
		r.Get("/push-logs", api.handleListPushLogs)
		r.Get("/push-logs/{id}", api.handleGetPushLog)
		r.Delete("/push-logs", api.handleCleanupPushLogs)
		r.Get("/settings", api.handleGetSettings)
		r.Put("/settings", api.handleUpdateSettings)
		r.Post("/change-password", api.handleChangePassword)
		r.Get("/webhook/presets", api.handleWebhookPresets)
	})
})
```

- [ ] **Step 4: Implement admin auth and core endpoint handlers**

Create `internal/httpapi/admin.go` with these concrete handlers:

```go
package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sunnyhmz7010/meowbridge/internal/auth"
	"github.com/sunnyhmz7010/meowbridge/internal/respond"
	"github.com/sunnyhmz7010/meowbridge/internal/store"
	"github.com/sunnyhmz7010/meowbridge/internal/token"
)

type loginRequest struct {
	Password string `json:"password"`
}

type endpointRequest struct {
	Name          string `json:"name"`
	MeowNickname  string `json:"meow_nickname"`
	DefaultTitle  string `json:"default_title"`
	MsgType       string `json:"msg_type"`
	HTMLHeight    int    `json:"html_height"`
	DefaultURL    string `json:"default_url"`
	DefaultImgURL string `json:"default_img_url"`
	Active        bool   `json:"active"`
}

func (api *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	hash, err := api.deps.Store.AdminPasswordHash(r.Context())
	if err != nil || !auth.VerifyPassword(hash, req.Password) {
		respond.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	raw, err := auth.IssueJWT(api.deps.Config.JWTSecret, 24*time.Hour)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to issue token")
		return
	}
	respond.OK(w, map[string]string{"token": raw})
}

func (api *API) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		raw := strings.TrimPrefix(authHeader, "Bearer ")
		if raw == authHeader || raw == "" {
			respond.Error(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		if err := auth.VerifyJWT(api.deps.Config.JWTSecret, raw); err != nil {
			respond.Error(w, http.StatusUnauthorized, "invalid token")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (api *API) handleCreateEndpoint(w http.ResponseWriter, r *http.Request) {
	var req endpointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Name == "" || req.MeowNickname == "" {
		respond.Error(w, http.StatusBadRequest, "name and meow_nickname are required")
		return
	}
	tok, err := token.Generate()
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	ep, err := api.deps.Store.CreateEndpoint(r.Context(), store.EndpointInput{
		Name:          req.Name,
		Token:         tok,
		MeowNickname:  req.MeowNickname,
		DefaultTitle:  req.DefaultTitle,
		MsgType:       defaultString(req.MsgType, "text"),
		HTMLHeight:    defaultInt(req.HTMLHeight, 200),
		DefaultURL:    req.DefaultURL,
		DefaultImgURL: req.DefaultImgURL,
		Active:        req.Active,
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to create endpoint")
		return
	}
	respond.OK(w, ep)
}

func (api *API) handleListEndpoints(w http.ResponseWriter, r *http.Request) {
	endpoints, err := api.deps.Store.ListEndpoints(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to list endpoints")
		return
	}
	respond.OK(w, endpoints)
}

func endpointID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func defaultInt(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}
```

- [ ] **Step 5: Add remaining handlers**

Append complete remaining handlers to `internal/httpapi/admin.go`:

```go
func (api *API) handleGetEndpoint(w http.ResponseWriter, r *http.Request) {
	id, err := endpointID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid endpoint id")
		return
	}
	ep, err := api.deps.Store.GetEndpoint(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "endpoint not found")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to get endpoint")
		return
	}
	respond.OK(w, ep)
}

func (api *API) handleUpdateEndpoint(w http.ResponseWriter, r *http.Request) {
	id, err := endpointID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid endpoint id")
		return
	}
	var req endpointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	ep, err := api.deps.Store.UpdateEndpoint(r.Context(), id, store.EndpointUpdate{
		Name:          req.Name,
		DefaultTitle:  req.DefaultTitle,
		MsgType:       defaultString(req.MsgType, "text"),
		HTMLHeight:    defaultInt(req.HTMLHeight, 200),
		DefaultURL:    req.DefaultURL,
		DefaultImgURL: req.DefaultImgURL,
		Active:        req.Active,
	})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to update endpoint")
		return
	}
	respond.OK(w, ep)
}

func (api *API) handleDeleteEndpoint(w http.ResponseWriter, r *http.Request) {
	id, err := endpointID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid endpoint id")
		return
	}
	if err := api.deps.Store.DeleteEndpoint(r.Context(), id); err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to delete endpoint")
		return
	}
	respond.OK(w, map[string]bool{"deleted": true})
}

func (api *API) handleResetEndpointToken(w http.ResponseWriter, r *http.Request) {
	id, err := endpointID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid endpoint id")
		return
	}
	newToken, err := token.Generate()
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	ep, err := api.deps.Store.ResetEndpointToken(r.Context(), id, newToken)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to reset token")
		return
	}
	respond.OK(w, ep)
}

func (api *API) handleSetEndpointActive(w http.ResponseWriter, r *http.Request) {
	id, err := endpointID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid endpoint id")
		return
	}
	var req struct {
		Active bool `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := api.deps.Store.SetEndpointActive(r.Context(), id, req.Active); err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to update active state")
		return
	}
	respond.OK(w, map[string]bool{"active": req.Active})
}

func (api *API) handleListPushLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := api.deps.Store.ListPushLogs(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to list push logs")
		return
	}
	respond.OK(w, logs)
}

func (api *API) handleGetPushLog(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid log id")
		return
	}
	log, err := api.deps.Store.GetPushLog(r.Context(), id)
	if err != nil {
		respond.Error(w, http.StatusNotFound, "push log not found")
		return
	}
	respond.OK(w, log)
}

func (api *API) handleCleanupPushLogs(w http.ResponseWriter, r *http.Request) {
	values, err := api.deps.Store.ListSettings(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	days, _ := strconv.Atoi(values["log_retention_days"])
	if days <= 0 {
		days = 14
	}
	deleted, err := api.deps.Store.CleanupPushLogs(r.Context(), time.Now().UTC().AddDate(0, 0, -days))
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to cleanup push logs")
		return
	}
	respond.OK(w, map[string]int64{"deleted": deleted})
}

func (api *API) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	values, err := api.deps.Store.ListSettings(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to list settings")
		return
	}
	respond.OK(w, values)
}

func (api *API) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var values map[string]string
	if err := json.NewDecoder(r.Body).Decode(&values); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	for _, key := range []string{"meow_api_base_url", "log_retention_days"} {
		if value, ok := values[key]; ok {
			if err := api.deps.Store.SetSetting(r.Context(), key, value); err != nil {
				respond.Error(w, http.StatusInternalServerError, "failed to update settings")
				return
			}
		}
	}
	respond.OK(w, map[string]bool{"updated": true})
}

func (api *API) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	hash, err := api.deps.Store.AdminPasswordHash(r.Context())
	if err != nil || !auth.VerifyPassword(hash, req.OldPassword) {
		respond.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	newHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to hash password")
		return
	}
	if err := api.deps.Store.UpdateAdminPasswordHash(r.Context(), newHash); err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to change password")
		return
	}
	respond.OK(w, map[string]bool{"changed": true})
}

func (api *API) handleWebhookPresets(w http.ResponseWriter, r *http.Request) {
	respond.OK(w, []map[string]string{
		{"source_type": "github_pr", "name": "GitHub Pull Request"},
		{"source_type": "github_action", "name": "GitHub Actions"},
		{"source_type": "github", "name": "GitHub Webhook"},
		{"source_type": "jenkins", "name": "Jenkins"},
		{"source_type": "grafana", "name": "Grafana"},
		{"source_type": "prometheus", "name": "Prometheus Alertmanager"},
		{"source_type": "zabbix", "name": "Zabbix"},
		{"source_type": "gotify", "name": "Gotify"},
		{"source_type": "emby", "name": "Emby"},
		{"source_type": "generic", "name": "Generic"},
	})
}
```

- [ ] **Step 6: Add admin password store methods**

Append to `internal/store/bootstrap.go`:

```go
func (s *Store) AdminPasswordHash(ctx context.Context) (string, error) {
	var hash string
	err := s.db.QueryRowContext(ctx, `SELECT password_hash FROM admin_users ORDER BY id ASC LIMIT 1`).Scan(&hash)
	return hash, err
}

func (s *Store) UpdateAdminPasswordHash(ctx context.Context, hash string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE admin_users SET password_hash = ?, updated_at = ? WHERE id = (SELECT id FROM admin_users ORDER BY id ASC LIMIT 1)`, hash, time.Now().UTC())
	return err
}
```

- [ ] **Step 7: Verify and commit**

Run:

```powershell
rtk go test ./internal/httpapi ./...
rtk git status --short
```

Expected: all tests pass.

Commit:

```powershell
rtk git add internal go.mod go.sum
rtk git commit -m "添加管理后台 API"
```

---

### Task 9: User Documentation and Final Verification

**Files:**
- Create: `README.md`
- Modify: `AGENTS.md` only if implementation-specific local commands need to be recorded.

**Interfaces:**
- Consumes: completed backend MVP.
- Produces: user-facing README with run instructions and examples.

- [ ] **Step 1: Write README using required public skeleton**

Create `README.md` with concise user-facing content. Use this exact high-level skeleton and adapt repository URLs to `sunnyhmz7010/meowbridge`:

```markdown
<div align="center">
  <h1>meowbridge</h1>
  <p>把常见 Webhook 自动转发到 MeoW 推送。</p>
</div>

<p align="center">
  <img alt="License" src="https://img.shields.io/github/license/sunnyhmz7010/meowbridge?color=blue" />
  <img alt="Go" src="https://img.shields.io/badge/Go-1.23%2B-00ADD8" />
</p>
<p align="center">[反馈问题](https://github.com/sunnyhmz7010/meowbridge/issues) · [下载源码](https://github.com/sunnyhmz7010/meowbridge/archive/refs/heads/main.zip)</p>

---

## ✨ 为什么做这个应用

MeoW Push 只有 nickname 作为入口标识，多场景共用时容易混淆，也不利于隔离泄露风险。meowbridge 提供独立 token 入口，把 GitHub、Grafana、Prometheus、Jenkins、Zabbix、Gotify、Emby 等常见 Webhook 自动解析并转发到指定 MeoW nickname。

## 🚀 核心能力

- 通用 Webhook 入口：外部服务填写 `/webhook/{token}` 即可推送。
- 内置常见服务解析器：自动提取标题、正文、链接和消息类型。
- 未识别 payload 兜底：格式化完整 JSON 推送，优先保证消息不丢。
- 管理 API：创建、启停、删除、重置推送入口。
- 推送日志：记录原始 payload、解析结果、MeoW 响应和失败原因。

## ⚡ 快速开始

```powershell
$env:ADMIN_PASSWORD="change-me"
$env:JWT_SECRET="replace-with-long-random-secret"
$env:MEOW_API_BASE_URL="https://api.chuckfang.com"
go run ./cmd/meowbridge
```

## 📖 使用说明

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

## 🧱 技术栈

- Go 1.23+
- chi
- SQLite
- modernc.org/sqlite
- JWT Bearer

## 🗂️ 项目结构

```text
.
├── cmd/meowbridge
├── internal/auth
├── internal/config
├── internal/httpapi
├── internal/meow
├── internal/respond
├── internal/store
├── internal/token
└── internal/webhook
```

## 👨‍💻 本地开发

```powershell
go test ./...
go run ./cmd/meowbridge
```

## 🔐 安全报告

如果发现安全问题，请不要公开披露细节。请优先参考仓库中的 [SECURITY.md](./SECURITY.md) 提交安全报告。

## 📄 许可证

本项目基于 [GNU General Public License v3.0](./LICENSE) 开源。

## ⭐ 星标历史

[![Star History Chart](https://api.star-history.com/svg?repos=sunnyhmz7010/meowbridge)](https://star-history.com/#sunnyhmz7010/meowbridge)

<div align="center">
  <sub>Built with ❤️ by Sunny</sub>
</div>
```

- [ ] **Step 2: Run full verification**

Run:

```powershell
rtk gofmt -w cmd internal
rtk go test ./...
$env:ADMIN_PASSWORD=""
$env:JWT_SECRET=""
$env:MEOW_API_BASE_URL=""
rtk go run ./cmd/meowbridge
```

Expected:

- `gofmt` completes without output.
- `go test ./...` passes.
- `go run` fails fast with a clear missing environment error when `ADMIN_PASSWORD`, `JWT_SECRET`, and `MEOW_API_BASE_URL` are unset. Do not leave a long-running server process active during this verification step.

- [ ] **Step 3: Inspect git scope**

Run:

```powershell
rtk git status --short
rtk git diff --stat
```

Expected: only backend MVP source files, tests, `go.mod`, `go.sum`, and `README.md` are changed since the feature branch started.

- [ ] **Step 4: Commit documentation and final verification**

Commit:

```powershell
rtk git add README.md AGENTS.md
rtk git commit -m "完善后端 MVP 使用文档"
```

- [ ] **Step 5: Final acceptance checklist**

Run:

```powershell
rtk go test ./...
rtk git log --oneline -9
rtk git status --short
```

Expected:

- `go test ./...` passes.
- Recent commits include one commit per task.
- `git status --short` has no output.

---

## Self-Review

Spec coverage:

- Go 1.23+, chi, SQLite modernc, JWT, single-binary structure: covered by Tasks 1, 2, 7, 8.
- Admin password initialization from `ADMIN_PASSWORD`: covered by Task 2 and Task 8.
- Settings initialization and mutation: covered by Task 2 and Task 8.
- Plaintext token storage and endpoint CRUD/reset/active/delete: covered by Task 4 and Task 8.
- Universal Webhook endpoint, JSON and `text/plain`: covered by Task 5 and Task 7.
- 10 built-in provider parsers plus generic and fallback: covered by Task 5.
- Field precedence and required `msg`: covered by Task 5.
- Synchronous MeoW forwarding and non-2xx handling: covered by Task 6 and Task 7.
- Push logs with full payload and response truncation: covered by Task 6 and Task 7.
- No built-in rate limiting, no queue retry, no Telegram proxy, no Vue frontend: preserved by task scope.
- `go test ./...` acceptance: covered by Task 9.

Type consistency:

- `store.Endpoint`, `webhook.EndpointDefaults`, `webhook.FinalMessage`, and `meow.PushRequest` use matching field names for title, message, URL, image URL, message type, and HTML height.
- HTTP handlers consume only interfaces created in earlier tasks.
- Admin route paths match the design document.
