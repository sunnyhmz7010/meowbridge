package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sunnyhmz7010/meowbridge/internal/config"
	"github.com/sunnyhmz7010/meowbridge/internal/meow"
)

func TestRootRedirectsToAdmin(t *testing.T) {
	router := NewRouter(Dependencies{Config: config.Config{JWTSecret: "secret"}})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusFound)
	}
	if location := rr.Header().Get("Location"); location != "/admin/" {
		t.Fatalf("Location = %q, want /admin/", location)
	}
}

func TestAdminRouteDoesNotCaptureAPIRoute(t *testing.T) {
	router := NewRouter(Dependencies{Config: config.Config{JWTSecret: "secret"}})

	req := httptest.NewRequest(http.MethodGet, "/api/admin/endpoints", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAdminRouteDoesNotCaptureWebhookRoute(t *testing.T) {
	st := newHTTPTestStore(t)
	router := NewRouter(Dependencies{Store: st, Config: config.Config{JWTSecret: "secret", MeowTimeout: time.Second}, MeowClient: meow.New("http://127.0.0.1:1", time.Millisecond)})

	req := httptest.NewRequest(http.MethodPost, "/webhook/missing", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
	assertTokenNotFoundJSON(t, rr)
}

func TestAdminRouteDoesNotCaptureVerifyRoute(t *testing.T) {
	st := newHTTPTestStore(t)
	router := NewRouter(Dependencies{Store: st, Config: config.Config{JWTSecret: "secret", MeowTimeout: time.Second}, MeowClient: meow.New("http://127.0.0.1:1", time.Millisecond)})

	req := httptest.NewRequest(http.MethodGet, "/verify/missing", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
	assertTokenNotFoundJSON(t, rr)
}

func TestAdminRouteServesEmbeddedUI(t *testing.T) {
	router := NewRouter(Dependencies{Config: config.Config{JWTSecret: "secret"}})

	req := httptest.NewRequest(http.MethodGet, "/admin/logs/1", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if contentType := rr.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("Content-Type = %q, want text/html", contentType)
	}
	if body := rr.Body.String(); !strings.Contains(body, `id="app"`) {
		t.Fatalf("body missing app mount: %q", body)
	}
}

func assertTokenNotFoundJSON(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()
	if contentType := rr.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", contentType)
	}
	if body := rr.Body.String(); !strings.Contains(body, `"error":"token not found"`) {
		t.Fatalf("body = %q, want token not found JSON error", body)
	}
}
