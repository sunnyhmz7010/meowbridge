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
