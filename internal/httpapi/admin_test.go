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

func TestAdminEndpointDefaultsActiveAndPreservesMeowNickname(t *testing.T) {
	ctx := context.Background()
	st := newHTTPTestStore(t)
	if err := st.Bootstrap(ctx, store.BootstrapOptions{AdminPassword: "admin-password", MeowAPIBaseURL: "https://meow.example.test", LogRetentionDays: 14}); err != nil {
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
	if err := st.Bootstrap(ctx, store.BootstrapOptions{AdminPassword: "admin-password", MeowAPIBaseURL: "https://meow.example.test", LogRetentionDays: 14}); err != nil {
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
	if err := st.Bootstrap(ctx, store.BootstrapOptions{AdminPassword: "admin-password", MeowAPIBaseURL: "https://meow.example.test", LogRetentionDays: 14}); err != nil {
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
	if err := st.Bootstrap(ctx, store.BootstrapOptions{AdminPassword: "admin-password", MeowAPIBaseURL: "https://meow.example.test", LogRetentionDays: 14}); err != nil {
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
	if err := st.Bootstrap(ctx, store.BootstrapOptions{AdminPassword: "admin-password", MeowAPIBaseURL: "https://meow.example.test", LogRetentionDays: 14}); err != nil {
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
	for key, want := range map[string]string{"meow_api_base_url": "https://new-meow.example.test", "log_retention_days": "30"} {
		got, err := st.GetSetting(ctx, key)
		if err != nil || got != want {
			t.Fatalf("GetSetting(%s) = %q, %v; want %q", key, got, err, want)
		}
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

func TestAdminSettingsUpdateChangesNextWebhookPushTarget(t *testing.T) {
	ctx := context.Background()
	st := newHTTPTestStore(t)
	firstCalls := 0
	firstServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		firstCalls++
		w.WriteHeader(http.StatusOK)
	}))
	defer firstServer.Close()
	secondCalls := 0
	secondServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secondCalls++
		w.WriteHeader(http.StatusOK)
	}))
	defer secondServer.Close()

	if err := st.Bootstrap(ctx, store.BootstrapOptions{AdminPassword: "admin-password", MeowAPIBaseURL: firstServer.URL, LogRetentionDays: 14}); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	_, err := st.CreateEndpoint(ctx, store.EndpointInput{Name: "Dynamic", Token: "dynamic-token", MeowNickname: "sunny", MsgType: "text", HTMLHeight: 200, Active: true})
	if err != nil {
		t.Fatalf("CreateEndpoint: %v", err)
	}
	router := NewRouter(Dependencies{
		Store:  st,
		Config: config.Config{JWTSecret: "jwt-secret", MeowTimeout: time.Second},
		MeowClient: meow.NewWithBaseURLProvider(func(ctx context.Context) (string, error) {
			return st.GetSetting(ctx, "meow_api_base_url")
		}, time.Second),
	})

	postWebhook := func() {
		t.Helper()
		req := httptest.NewRequest(http.MethodPost, "/webhook/dynamic-token", strings.NewReader("dynamic push"))
		req.Header.Set("Content-Type", "text/plain")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("webhook status = %d body = %s", rr.Code, rr.Body.String())
		}
	}

	postWebhook()
	token := adminToken(t, router)
	settingsBody, err := json.Marshal(map[string]string{"meow_api_base_url": secondServer.URL})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	settingsReq := httptest.NewRequest(http.MethodPut, "/api/admin/settings", bytes.NewReader(settingsBody))
	settingsReq.Header.Set("Authorization", "Bearer "+token)
	settingsRR := httptest.NewRecorder()
	router.ServeHTTP(settingsRR, settingsReq)
	if settingsRR.Code != http.StatusOK {
		t.Fatalf("settings status = %d body = %s", settingsRR.Code, settingsRR.Body.String())
	}
	postWebhook()

	if firstCalls != 1 || secondCalls != 1 {
		t.Fatalf("push calls: first = %d, second = %d", firstCalls, secondCalls)
	}
}

func TestAdminRejectsInvalidRetentionDaysAndEmptyNewPassword(t *testing.T) {
	ctx := context.Background()
	st := newHTTPTestStore(t)
	if err := st.Bootstrap(ctx, store.BootstrapOptions{AdminPassword: "admin-password", MeowAPIBaseURL: "https://meow.example.test", LogRetentionDays: 14}); err != nil {
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

func TestAdminRejectsInvalidMeowAPIBaseURL(t *testing.T) {
	ctx := context.Background()
	st := newHTTPTestStore(t)
	const originalURL = "https://meow.example.test"
	if err := st.Bootstrap(ctx, store.BootstrapOptions{AdminPassword: "admin-password", MeowAPIBaseURL: originalURL, LogRetentionDays: 14}); err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	router := NewRouter(Dependencies{Store: st, Config: config.Config{JWTSecret: "jwt-secret", MeowTimeout: time.Second}, MeowClient: meow.New("http://127.0.0.1:1", time.Millisecond)})
	token := adminToken(t, router)

	for _, value := range []string{"", "/relative", "ftp://meow.example.test", "https:///missing-host"} {
		t.Run(value, func(t *testing.T) {
			body, err := json.Marshal(map[string]string{"meow_api_base_url": value})
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			req := httptest.NewRequest(http.MethodPut, "/api/admin/settings", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+token)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("value %q status = %d body = %s", value, rr.Code, rr.Body.String())
			}
			got, err := st.GetSetting(ctx, "meow_api_base_url")
			if err != nil || got != originalURL {
				t.Fatalf("stored URL = %q, %v; want %q", got, err, originalURL)
			}
		})
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
