package webui

import (
	"bytes"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

type Handler struct {
	fsys    fs.FS
	enabled bool
}

func NewHandler(fsys fs.FS, enabled bool) http.Handler {
	return Handler{fsys: fsys, enabled: enabled}
}

func DisabledHandler() http.Handler {
	return NewHandler(nil, false)
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.enabled || h.fsys == nil {
		http.Error(w, "admin UI is not built", http.StatusNotFound)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/admin/")
	if name == "" {
		name = "index.html"
	}
	if hasPathTraversal(name) {
		http.NotFound(w, r)
		return
	}
	name = path.Clean("/" + name)[1:]
	if strings.HasPrefix(name, "..") {
		http.NotFound(w, r)
		return
	}

	if isFile(h.fsys, name) {
		h.serveFile(w, r, name)
		return
	}

	if strings.HasPrefix(name, "assets/") {
		http.NotFound(w, r)
		return
	}

	h.serveFile(w, r, "index.html")
}

func isFile(fsys fs.FS, name string) bool {
	info, err := fs.Stat(fsys, name)
	return err == nil && !info.IsDir()
}

func hasPathTraversal(name string) bool {
	for _, segment := range strings.Split(name, "/") {
		if segment == ".." {
			return true
		}
	}
	return false
}

func (h Handler) serveFile(w http.ResponseWriter, r *http.Request, name string) {
	data, err := fs.ReadFile(h.fsys, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(data))
}

func HasIndex(fsys fs.FS) bool {
	if fsys == nil {
		return false
	}
	info, err := fs.Stat(fsys, "index.html")
	return err == nil && !info.IsDir()
}

func SubOrDisabled(root fs.FS, dir string) (fs.FS, bool) {
	sub, err := fs.Sub(root, dir)
	if err != nil {
		return nil, false
	}
	if !HasIndex(sub) {
		return nil, false
	}
	return sub, true
}
