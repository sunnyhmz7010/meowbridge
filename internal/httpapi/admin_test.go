package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sunnyhmz7010/meowbridge/internal/config"
	"github.com/sunnyhmz7010/meowbridge/internal/meow"
	"github.com/sunnyhmz7010/meowbridge/internal/store"
)

func TestAdminLoginAndEndpointCRUD(t *testing.T) {
	ctx := context.Background()
	st := newHTTPTestStore(t)
	if err := st.Bootstrap(ctx, store.BootstrapOptions{AdminPassword: "admin-password", LogRetentionDays: 14}); err != nil {
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

	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/endpoints", bytes.NewBufferString(`{"name":"GitHub","meow_nickname":"sunny","default_title":"GitHub","msg_type":"text","html_height":200,"parser_config":{"mode":"preset","preset":"github_push_minimal"},"active":true}`))
	createReq.Header.Set("Authorization", "Bearer "+loginBody.Data.Token)
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusOK {
		t.Fatalf("create status = %d body = %s", createRR.Code, createRR.Body.String())
	}
	if !strings.Contains(createRR.Body.String(), `"meow_nickname"`) || strings.Contains(createRR.Body.String(), `"MeowNickname"`) {
		t.Fatalf("create response does not use stable snake_case JSON: %s", createRR.Body.String())
	}
	if !strings.Contains(createRR.Body.String(), `"parser_config":{"mode":"preset","preset":"github_push_minimal","field_mapping":{},"default_values":{}}`) {
		t.Fatalf("create response does not expose parser_config object: %s", createRR.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/admin/endpoints", nil)
	listReq.Header.Set("Authorization", "Bearer "+loginBody.Data.Token)
	listRR := httptest.NewRecorder()
	router.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("list status = %d", listRR.Code)
	}
	if !strings.Contains(listRR.Body.String(), `"default_title"`) || strings.Contains(listRR.Body.String(), `"DefaultTitle"`) {
		t.Fatalf("list response does not use stable snake_case JSON: %s", listRR.Body.String())
	}
}

func TestAdminWebhookPresetsAndPreview(t *testing.T) {
	ctx := context.Background()
	st := newHTTPTestStore(t)
	if err := st.Bootstrap(ctx, store.BootstrapOptions{AdminPassword: "admin-password", LogRetentionDays: 14}); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	router := NewRouter(Dependencies{Store: st, Config: config.Config{JWTSecret: "jwt-secret", MeowTimeout: time.Second}, MeowClient: meow.New("http://127.0.0.1:1", time.Millisecond)})
	token := adminToken(t, router)

	presetsReq := httptest.NewRequest(http.MethodGet, "/api/admin/webhook/presets", nil)
	presetsReq.Header.Set("Authorization", "Bearer "+token)
	presetsRR := httptest.NewRecorder()
	router.ServeHTTP(presetsRR, presetsReq)
	if presetsRR.Code != http.StatusOK {
		t.Fatalf("presets status = %d body = %s", presetsRR.Code, presetsRR.Body.String())
	}
	if !strings.Contains(presetsRR.Body.String(), `"id":"github_push_minimal"`) || !strings.Contains(presetsRR.Body.String(), `"field_mapping"`) {
		t.Fatalf("presets response missing parser details: %s", presetsRR.Body.String())
	}

	previewBody := `{
		"parser_config": {"mode":"preset","preset":"github_push_minimal"},
		"payload": {"sourcecontrol":"github","service":"github","event_type":"push","hook":{"url":"https://github.com/sunnyhmz7010/meowbridge"},"ref":"refs/heads/main"}
	}`
	previewReq := httptest.NewRequest(http.MethodPost, "/api/admin/webhook/preview", bytes.NewBufferString(previewBody))
	previewReq.Header.Set("Authorization", "Bearer "+token)
	previewReq.Header.Set("Content-Type", "application/json")
	previewRR := httptest.NewRecorder()
	router.ServeHTTP(previewRR, previewReq)
	if previewRR.Code != http.StatusOK {
		t.Fatalf("preview status = %d body = %s", previewRR.Code, previewRR.Body.String())
	}
	for _, want := range []string{`"source_type":"github_push_minimal"`, `"title":"GitHub Push"`, "分支: main"} {
		if !strings.Contains(previewRR.Body.String(), want) {
			t.Fatalf("preview response missing %q: %s", want, previewRR.Body.String())
		}
	}
}

func TestAdminRejectsInvalidParserConfig(t *testing.T) {
	ctx := context.Background()
	st := newHTTPTestStore(t)
	if err := st.Bootstrap(ctx, store.BootstrapOptions{AdminPassword: "admin-password", LogRetentionDays: 14}); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	router := NewRouter(Dependencies{Store: st, Config: config.Config{JWTSecret: "jwt-secret", MeowTimeout: time.Second}, MeowClient: meow.New("http://127.0.0.1:1", time.Millisecond)})
	token := adminToken(t, router)

	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/endpoints", bytes.NewBufferString(`{"name":"Bad","meow_nickname":"sunny","parser_config":[]}`))
	createReq.Header.Set("Authorization", "Bearer "+token)
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusBadRequest {
		t.Fatalf("create status = %d body = %s", createRR.Code, createRR.Body.String())
	}

	validReq := httptest.NewRequest(http.MethodPost, "/api/admin/endpoints", bytes.NewBufferString(`{"name":"Good","meow_nickname":"sunny","parser_config":{"mode":"preset","preset":"github_push_minimal"}}`))
	validReq.Header.Set("Authorization", "Bearer "+token)
	validReq.Header.Set("Content-Type", "application/json")
	validRR := httptest.NewRecorder()
	router.ServeHTTP(validRR, validReq)
	if validRR.Code != http.StatusOK {
		t.Fatalf("valid create status = %d body = %s", validRR.Code, validRR.Body.String())
	}
	endpoints, err := st.ListEndpoints(ctx)
	if err != nil || len(endpoints) != 1 {
		t.Fatalf("ListEndpoints = %#v, %v", endpoints, err)
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/api/admin/endpoints/"+strconv.FormatInt(endpoints[0].ID, 10), bytes.NewBufferString(`{"name":"Bad","parser_config":{"mode":"preset","preset":"missing"}}`))
	updateReq.Header.Set("Authorization", "Bearer "+token)
	updateReq.Header.Set("Content-Type", "application/json")
	updateRR := httptest.NewRecorder()
	router.ServeHTTP(updateRR, updateReq)
	if updateRR.Code != http.StatusBadRequest {
		t.Fatalf("update status = %d body = %s", updateRR.Code, updateRR.Body.String())
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

func TestAdminEndpointDefaultsActiveAndPreservesMeowNickname(t *testing.T) {
	ctx := context.Background()
	st := newHTTPTestStore(t)
	if err := st.Bootstrap(ctx, store.BootstrapOptions{AdminPassword: "admin-password", LogRetentionDays: 14}); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	router := NewRouter(Dependencies{Store: st, Config: config.Config{JWTSecret: "jwt-secret", MeowTimeout: time.Second}, MeowClient: meow.New("http://127.0.0.1:1", time.Millisecond)})
	token := adminToken(t, router)

	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/endpoints", bytes.NewBufferString(`{"name":"GitHub","meow_nickname":"sunny"}`))
	createReq.Header.Set("Authorization", "Bearer "+token)
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusOK {
		t.Fatalf("create status = %d body = %s", createRR.Code, createRR.Body.String())
	}

	endpoints, err := st.ListEndpoints(ctx)
	if err != nil || len(endpoints) != 1 {
		t.Fatalf("ListEndpoints = %#v, %v", endpoints, err)
	}
	if !endpoints[0].Active {
		t.Fatal("endpoint should default to active")
	}
	if err := st.SetEndpointActive(ctx, endpoints[0].ID, false); err != nil {
		t.Fatalf("SetEndpointActive: %v", err)
	}

	updateReq := httptest.NewRequest(http.MethodPut, "/api/admin/endpoints/"+strconv.FormatInt(endpoints[0].ID, 10), bytes.NewBufferString(`{"name":"Renamed","meow_nickname":"other"}`))
	updateReq.Header.Set("Authorization", "Bearer "+token)
	updateRR := httptest.NewRecorder()
	router.ServeHTTP(updateRR, updateReq)
	if updateRR.Code != http.StatusOK {
		t.Fatalf("update status = %d body = %s", updateRR.Code, updateRR.Body.String())
	}

	updated, err := st.GetEndpoint(ctx, endpoints[0].ID)
	if err != nil {
		t.Fatalf("GetEndpoint: %v", err)
	}
	if updated.MeowNickname != "sunny" {
		t.Fatalf("meow nickname = %q, want sunny", updated.MeowNickname)
	}
	if updated.Active {
		t.Fatal("endpoint should remain inactive when update omits active")
	}
}

func TestAdminSetEndpointActiveRequiresActiveField(t *testing.T) {
	ctx := context.Background()
	st := newHTTPTestStore(t)
	if err := st.Bootstrap(ctx, store.BootstrapOptions{AdminPassword: "admin-password", LogRetentionDays: 14}); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	endpoint, err := st.CreateEndpoint(ctx, store.EndpointInput{Name: "Active", Token: "active-token", MeowNickname: "sunny", MsgType: "text", Active: true})
	if err != nil {
		t.Fatalf("CreateEndpoint: %v", err)
	}
	router := NewRouter(Dependencies{Store: st, Config: config.Config{JWTSecret: "jwt-secret", MeowTimeout: time.Second}, MeowClient: meow.New("http://127.0.0.1:1", time.Millisecond)})
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/endpoints/"+strconv.FormatInt(endpoint.ID, 10)+"/active", bytes.NewBufferString(`{}`))
	req.Header.Set("Authorization", "Bearer "+adminToken(t, router))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body = %s", rr.Code, rr.Body.String())
	}
	got, err := st.GetEndpoint(ctx, endpoint.ID)
	if err != nil {
		t.Fatalf("GetEndpoint: %v", err)
	}
	if !got.Active {
		t.Fatal("endpoint was disabled when active field was missing")
	}
}

func TestAdminPushLogListOmitsSensitiveFields(t *testing.T) {
	ctx := context.Background()
	st := newHTTPTestStore(t)
	if err := st.Bootstrap(ctx, store.BootstrapOptions{AdminPassword: "admin-password", LogRetentionDays: 14}); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	if _, err := st.CreatePushLog(ctx, store.PushLogInput{Token: "secret-token", RequestHeaders: "Authorization: secret", RequestPayload: "full payload", ParsedMsg: strings.Repeat("x", 201)}); err != nil {
		t.Fatalf("CreatePushLog: %v", err)
	}
	router := NewRouter(Dependencies{Store: st, Config: config.Config{JWTSecret: "jwt-secret", MeowTimeout: time.Second}, MeowClient: meow.New("http://127.0.0.1:1", time.Millisecond)})
	req := httptest.NewRequest(http.MethodGet, "/api/admin/push-logs", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, router))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rr.Code, rr.Body.String())
	}
	for _, sensitive := range []string{"secret-token", "Authorization: secret", "full payload"} {
		if strings.Contains(rr.Body.String(), sensitive) {
			t.Fatalf("list response exposed %q: %s", sensitive, rr.Body.String())
		}
	}
}

func TestAdminMissingPushLogReturns404(t *testing.T) {
	ctx := context.Background()
	st := newHTTPTestStore(t)
	if err := st.Bootstrap(ctx, store.BootstrapOptions{AdminPassword: "admin-password", LogRetentionDays: 14}); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	router := NewRouter(Dependencies{Store: st, Config: config.Config{JWTSecret: "jwt-secret", MeowTimeout: time.Second}, MeowClient: meow.New("http://127.0.0.1:1", time.Millisecond)})
	req := httptest.NewRequest(http.MethodGet, "/api/admin/push-logs/999", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken(t, router))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d body = %s", rr.Code, rr.Body.String())
	}
}

func TestAdminSettingsAndPasswordChangesPersist(t *testing.T) {
	ctx := context.Background()
	st := newHTTPTestStore(t)
	if err := st.Bootstrap(ctx, store.BootstrapOptions{AdminPassword: "admin-password", LogRetentionDays: 14}); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	router := NewRouter(Dependencies{Store: st, Config: config.Config{JWTSecret: "jwt-secret", MeowTimeout: time.Second}, MeowClient: meow.New("http://127.0.0.1:1", time.Millisecond)})
	token := adminToken(t, router)

	settingsReq := httptest.NewRequest(http.MethodPut, "/api/admin/settings", bytes.NewBufferString(`{"meow_api_base_url":"https://new-meow.example.test","log_retention_days":"30"}`))
	settingsReq.Header.Set("Authorization", "Bearer "+token)
	settingsRR := httptest.NewRecorder()
	router.ServeHTTP(settingsRR, settingsReq)
	if settingsRR.Code != http.StatusOK {
		t.Fatalf("settings status = %d body = %s", settingsRR.Code, settingsRR.Body.String())
	}
	got, err := st.GetSetting(ctx, "log_retention_days")
	if err != nil || got != "30" {
		t.Fatalf("log_retention_days = %q, %v; want 30", got, err)
	}
	if _, err := st.GetSetting(ctx, "meow_api_base_url"); err != store.ErrNotFound {
		t.Fatalf("meow_api_base_url error = %v, want ErrNotFound", err)
	}

	passwordReq := httptest.NewRequest(http.MethodPost, "/api/admin/change-password", bytes.NewBufferString(`{"old_password":"admin-password","new_password":"new-password"}`))
	passwordReq.Header.Set("Authorization", "Bearer "+token)
	passwordRR := httptest.NewRecorder()
	router.ServeHTTP(passwordRR, passwordReq)
	if passwordRR.Code != http.StatusOK {
		t.Fatalf("password status = %d body = %s", passwordRR.Code, passwordRR.Body.String())
	}

	loginReq := httptest.NewRequest(http.MethodPost, "/api/admin/login", bytes.NewBufferString(`{"password":"new-password"}`))
	loginRR := httptest.NewRecorder()
	router.ServeHTTP(loginRR, loginReq)
	if loginRR.Code != http.StatusOK {
		t.Fatalf("new password login status = %d body = %s", loginRR.Code, loginRR.Body.String())
	}
}

func TestAdminRejectsInvalidRetentionDaysAndEmptyNewPassword(t *testing.T) {
	ctx := context.Background()
	st := newHTTPTestStore(t)
	if err := st.Bootstrap(ctx, store.BootstrapOptions{AdminPassword: "admin-password", LogRetentionDays: 14}); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	router := NewRouter(Dependencies{Store: st, Config: config.Config{JWTSecret: "jwt-secret", MeowTimeout: time.Second}, MeowClient: meow.New("http://127.0.0.1:1", time.Millisecond)})
	token := adminToken(t, router)

	settingsReq := httptest.NewRequest(http.MethodPut, "/api/admin/settings", bytes.NewBufferString(`{"log_retention_days":"0"}`))
	settingsReq.Header.Set("Authorization", "Bearer "+token)
	settingsRR := httptest.NewRecorder()
	router.ServeHTTP(settingsRR, settingsReq)
	if settingsRR.Code != http.StatusBadRequest {
		t.Fatalf("settings status = %d body = %s", settingsRR.Code, settingsRR.Body.String())
	}
	retentionDays, err := st.GetSetting(ctx, "log_retention_days")
	if err != nil || retentionDays != "14" {
		t.Fatalf("log_retention_days = %q, %v; want 14", retentionDays, err)
	}

	passwordReq := httptest.NewRequest(http.MethodPost, "/api/admin/change-password", bytes.NewBufferString(`{"old_password":"admin-password","new_password":""}`))
	passwordReq.Header.Set("Authorization", "Bearer "+token)
	passwordRR := httptest.NewRecorder()
	router.ServeHTTP(passwordRR, passwordReq)
	if passwordRR.Code != http.StatusBadRequest {
		t.Fatalf("password status = %d body = %s", passwordRR.Code, passwordRR.Body.String())
	}
}

func TestAdminSettingsIgnoresMeowAPIBaseURL(t *testing.T) {
	ctx := context.Background()
	st := newHTTPTestStore(t)
	if err := st.Bootstrap(ctx, store.BootstrapOptions{AdminPassword: "admin-password", LogRetentionDays: 14}); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	router := NewRouter(Dependencies{Store: st, Config: config.Config{JWTSecret: "jwt-secret", MeowTimeout: time.Second}, MeowClient: meow.New("http://127.0.0.1:1", time.Millisecond)})
	token := adminToken(t, router)

	body, err := json.Marshal(map[string]string{"meow_api_base_url": "https://new-meow.example.test"})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rr.Code, rr.Body.String())
	}
	if _, err := st.GetSetting(ctx, "meow_api_base_url"); err != store.ErrNotFound {
		t.Fatalf("meow_api_base_url error = %v, want ErrNotFound", err)
	}
}

func TestAdminSettingsDoesNotExposeLegacyMeowAPIBaseURL(t *testing.T) {
	ctx := context.Background()
	st := newHTTPTestStore(t)
	if err := st.Bootstrap(ctx, store.BootstrapOptions{AdminPassword: "admin-password", LogRetentionDays: 14}); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	if err := st.SetSetting(ctx, "meow_api_base_url", "https://legacy.example.test"); err != nil {
		t.Fatalf("SetSetting legacy: %v", err)
	}
	router := NewRouter(Dependencies{Store: st, Config: config.Config{JWTSecret: "jwt-secret", MeowTimeout: time.Second}, MeowClient: meow.New("http://127.0.0.1:1", time.Millisecond)})
	token := adminToken(t, router)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/settings", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rr.Code, rr.Body.String())
	}
	if strings.Contains(rr.Body.String(), "meow_api_base_url") || strings.Contains(rr.Body.String(), "legacy.example.test") {
		t.Fatalf("settings response exposed legacy MeoW URL: %s", rr.Body.String())
	}
}

func adminToken(t *testing.T, router http.Handler) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/login", bytes.NewBufferString(`{"password":"admin-password"}`))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("login status = %d body = %s", rr.Code, rr.Body.String())
	}
	var body struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("login json: %v", err)
	}
	return body.Data.Token
}
