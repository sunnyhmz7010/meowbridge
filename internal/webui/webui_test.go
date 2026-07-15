package webui

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestHandlerServesIndex(t *testing.T) {
	handler := NewHandler(fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>admin</html>")},
	}, true)

	req := httptest.NewRequest(http.MethodGet, "/admin/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "admin") {
		t.Fatalf("body = %q, want admin html", rr.Body.String())
	}
}

func TestHandlerFallsBackToIndexForSPARoute(t *testing.T) {
	handler := NewHandler(fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>admin</html>")},
	}, true)

	req := httptest.NewRequest(http.MethodGet, "/admin/logs/1", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "admin") {
		t.Fatalf("body = %q, want admin html", rr.Body.String())
	}
}

func TestHandlerDoesNotFallbackForMissingAssets(t *testing.T) {
	handler := NewHandler(fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>admin</html>")},
	}, true)

	req := httptest.NewRequest(http.MethodGet, "/admin/assets/missing.js", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestHandlerRejectsPathTraversal(t *testing.T) {
	handler := NewHandler(fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>admin</html>")},
		"secret.txt": &fstest.MapFile{Data: []byte("secret")},
	}, true)

	req := httptest.NewRequest(http.MethodGet, "/admin/../secret.txt", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
	if strings.Contains(rr.Body.String(), "secret") {
		t.Fatalf("body leaked secret: %q", rr.Body.String())
	}
}

func TestDisabledHandlerReturnsNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/admin/", nil)
	rr := httptest.NewRecorder()
	DisabledHandler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestSubOrDisabledRequiresIndex(t *testing.T) {
	root := fstest.MapFS{
		"dist/.gitkeep": &fstest.MapFile{Data: []byte{}},
	}

	sub, ok := SubOrDisabled(root, "dist")
	if ok || sub != nil {
		t.Fatalf("SubOrDisabled() = (%v, %v), want disabled", sub, ok)
	}
}

func TestSubOrDisabledEnablesWhenIndexExists(t *testing.T) {
	root := fstest.MapFS{
		"dist/index.html": &fstest.MapFile{Data: []byte("admin")},
	}

	sub, ok := SubOrDisabled(root, "dist")
	if !ok {
		t.Fatal("SubOrDisabled() disabled, want enabled")
	}
	if _, err := fs.Stat(sub, "index.html"); err != nil {
		t.Fatalf("index.html missing in sub fs: %v", err)
	}
}
