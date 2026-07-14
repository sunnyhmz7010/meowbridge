package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

	body := &countingErrorBody{}
	req := httptest.NewRequest(http.MethodPost, "/webhook/missing", nil)
	req.Body = body
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rr.Code)
	}
	if body.reads != 0 {
		t.Fatalf("body reads = %d, want 0", body.reads)
	}
	logs, err := st.ListPushLogs(context.Background())
	if err != nil {
		t.Fatalf("ListPushLogs: %v", err)
	}
	if len(logs) != 0 {
		t.Fatalf("logs = %#v", logs)
	}
}

func TestWebhookTextPlainWithCharsetUsesQueryOverride(t *testing.T) {
	ctx := context.Background()
	st := newHTTPTestStore(t)
	endpoint, err := st.CreateEndpoint(ctx, store.EndpointInput{Name: "Text", Token: "text-token", MeowNickname: "sunny", MsgType: "text", HTMLHeight: 200, Active: true})
	if err != nil {
		t.Fatalf("CreateEndpoint: %v", err)
	}

	var received meow.PushRequest
	var receivedMsgType string
	meowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode push request: %v", err)
		}
		receivedMsgType = r.URL.Query().Get("msgType")
		w.WriteHeader(http.StatusOK)
	}))
	defer meowServer.Close()

	router := NewRouter(Dependencies{Store: st, Config: config.Config{JWTSecret: "secret", MeowTimeout: time.Second}, MeowClient: meow.New(meowServer.URL, time.Second)})
	req := httptest.NewRequest(http.MethodPost, "/webhook/text-token?msgType=markdown", strings.NewReader("plain message"))
	req.Header.Set("Content-Type", "TEXT/PLAIN; charset=utf-8")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rr.Code, rr.Body.String())
	}
	if received.Msg != "plain message" || receivedMsgType != "markdown" {
		t.Fatalf("push request = %#v", received)
	}

	logs, err := st.ListPushLogs(ctx)
	if err != nil {
		t.Fatalf("ListPushLogs: %v", err)
	}
	if len(logs) != 1 || logs[0].EndpointID != endpoint.ID || logs[0].RequestPayload != "plain message" {
		t.Fatalf("logs = %#v", logs)
	}
}

func TestWebhookRejectsOversizedBodyWithoutForwardingOrLoggingPayload(t *testing.T) {
	ctx := context.Background()
	st := newHTTPTestStore(t)
	_, err := st.CreateEndpoint(ctx, store.EndpointInput{Name: "Large", Token: "large-token", MeowNickname: "sunny", MsgType: "text", HTMLHeight: 200, Active: true})
	if err != nil {
		t.Fatalf("CreateEndpoint: %v", err)
	}

	upstreamCalls := 0
	meowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamCalls++
		w.WriteHeader(http.StatusOK)
	}))
	defer meowServer.Close()

	payload := strings.Repeat("x", 1024*1024+1)
	router := NewRouter(Dependencies{Store: st, Config: config.Config{JWTSecret: "secret", MeowTimeout: time.Second}, MeowClient: meow.New(meowServer.URL, time.Second)})
	req := httptest.NewRequest(http.MethodPost, "/webhook/large-token", strings.NewReader(payload))
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d body = %s", rr.Code, rr.Body.String())
	}
	if upstreamCalls != 0 {
		t.Fatalf("upstream calls = %d, want 0", upstreamCalls)
	}
	logs, err := st.ListPushLogs(ctx)
	if err != nil {
		t.Fatalf("ListPushLogs: %v", err)
	}
	if len(logs) != 1 || logs[0].RequestPayload != "" {
		t.Fatalf("logs = %#v", logs)
	}
}

func TestDisabledWebhookDoesNotReadOrLogBody(t *testing.T) {
	ctx := context.Background()
	st := newHTTPTestStore(t)
	_, err := st.CreateEndpoint(ctx, store.EndpointInput{Name: "Disabled", Token: "disabled-token", MeowNickname: "sunny", MsgType: "text", HTMLHeight: 200, Active: false})
	if err != nil {
		t.Fatalf("CreateEndpoint: %v", err)
	}

	router := NewRouter(Dependencies{Store: st, Config: config.Config{JWTSecret: "secret", MeowTimeout: time.Second}, MeowClient: meow.New("http://127.0.0.1:1", time.Millisecond)})
	body := &countingErrorBody{}
	req := httptest.NewRequest(http.MethodPost, "/webhook/disabled-token", nil)
	req.Body = body
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d body = %s", rr.Code, rr.Body.String())
	}
	if body.reads != 0 {
		t.Fatalf("body reads = %d, want 0", body.reads)
	}
	logs, err := st.ListPushLogs(ctx)
	if err != nil {
		t.Fatalf("ListPushLogs: %v", err)
	}
	if len(logs) != 1 || logs[0].RequestPayload != "" || logs[0].ErrorMessage != "endpoint is disabled" {
		t.Fatalf("logs = %#v", logs)
	}
}

func TestWebhookMeowFailureReturns502AndWritesFailureLog(t *testing.T) {
	ctx := context.Background()
	st := newHTTPTestStore(t)
	_, err := st.CreateEndpoint(ctx, store.EndpointInput{Name: "Failure", Token: "failure-token", MeowNickname: "sunny", MsgType: "text", HTMLHeight: 200, Active: true})
	if err != nil {
		t.Fatalf("CreateEndpoint: %v", err)
	}

	meowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "meow unavailable", http.StatusServiceUnavailable)
	}))
	defer meowServer.Close()

	router := NewRouter(Dependencies{Store: st, Config: config.Config{JWTSecret: "secret", MeowTimeout: time.Second}, MeowClient: meow.New(meowServer.URL, time.Second)})
	req := httptest.NewRequest(http.MethodPost, "/webhook/failure-token", strings.NewReader("failed push"))
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("status = %d body = %s", rr.Code, rr.Body.String())
	}
	var responseBody map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("json: %v", err)
	}
	if responseBody["error"] != "meow upstream request failed" {
		t.Fatalf("public error = %q", responseBody["error"])
	}
	logs, err := st.ListPushLogs(ctx)
	if err != nil {
		t.Fatalf("ListPushLogs: %v", err)
	}
	if len(logs) != 1 || logs[0].Success || logs[0].MeowStatusCode != http.StatusServiceUnavailable || logs[0].RequestPayload != "failed push" || logs[0].ErrorMessage != "meow upstream returned 503" {
		t.Fatalf("logs = %#v", logs)
	}
}

func TestWebhookReturns500WhenSuccessLogWriteFails(t *testing.T) {
	ctx := context.Background()
	st := newHTTPTestStore(t)
	_, err := st.CreateEndpoint(ctx, store.EndpointInput{Name: "Log failure", Token: "log-failure-token", MeowNickname: "sunny", MsgType: "text", HTMLHeight: 200, Active: true})
	if err != nil {
		t.Fatalf("CreateEndpoint: %v", err)
	}

	meowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer meowServer.Close()

	router := NewRouter(Dependencies{Store: failingPushLogStore{Store: st}, Config: config.Config{JWTSecret: "secret", MeowTimeout: time.Second}, MeowClient: meow.New(meowServer.URL, time.Second)})
	req := httptest.NewRequest(http.MethodPost, "/webhook/log-failure-token", strings.NewReader("log failure"))
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d body = %s", rr.Code, rr.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v", err)
	}
	if body["ok"] != false {
		t.Fatalf("body = %#v", body)
	}
}

func TestWebhookMeowFailureReturns502WhenLogWriteFails(t *testing.T) {
	ctx := context.Background()
	st := newHTTPTestStore(t)
	_, err := st.CreateEndpoint(ctx, store.EndpointInput{Name: "Both failures", Token: "both-failures-token", MeowNickname: "sunny", MsgType: "text", HTMLHeight: 200, Active: true})
	if err != nil {
		t.Fatalf("CreateEndpoint: %v", err)
	}

	meowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "meow unavailable", http.StatusServiceUnavailable)
	}))
	defer meowServer.Close()

	router := NewRouter(Dependencies{Store: failingPushLogStore{Store: st}, Config: config.Config{JWTSecret: "secret", MeowTimeout: time.Second}, MeowClient: meow.New(meowServer.URL, time.Second)})
	req := httptest.NewRequest(http.MethodPost, "/webhook/both-failures-token", strings.NewReader("failed push"))
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("status = %d body = %s", rr.Code, rr.Body.String())
	}
}

type failingPushLogStore struct {
	*store.Store
}

type countingErrorBody struct {
	reads int
}

func (b *countingErrorBody) Read([]byte) (int, error) {
	b.reads++
	return 0, errors.New("body must not be read")
}

func (*countingErrorBody) Close() error { return nil }

func (failingPushLogStore) CreatePushLog(context.Context, store.PushLogInput) (int64, error) {
	return 0, errors.New("push log unavailable")
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
