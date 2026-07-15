package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sunnyhmz7010/meowbridge/internal/config"
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

func TestAdminRouteDisabledWhenUIIsNotBuilt(t *testing.T) {
	router := NewRouter(Dependencies{Config: config.Config{JWTSecret: "secret"}})

	req := httptest.NewRequest(http.MethodGet, "/admin/logs/1", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}
