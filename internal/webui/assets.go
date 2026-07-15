package webui

import (
	"embed"
	"net/http"
)

//go:embed dist
var assets embed.FS

func HandlerFromEmbeddedAssets() http.Handler {
	fsys, enabled := SubOrDisabled(assets, "dist")
	return NewHandler(fsys, enabled)
}
