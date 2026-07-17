package tgproxy

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sunnyhmz7010/meowbridge/internal/meow"
	"github.com/sunnyhmz7010/meowbridge/internal/store"
)

type StoreInterface interface {
	GetEndpointByToken(ctx context.Context, token string) (store.Endpoint, error)
}

type Dependencies struct {
	Store      StoreInterface
	MeowClient *meow.Client
}

type Handler struct {
	deps *Dependencies
}

func NewHandler(deps *Dependencies) *Handler {
	return &Handler{deps: deps}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/tg-proxy/bot{token}/{method}", h.handleTGMethod)
	r.Get("/tg-proxy/bot{token}/{method}", h.handleTGMethod)
}

func (h *Handler) handleTGMethod(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	method := chi.URLParam(r, "method")

	// 记录请求日志
	slog.Info("tg proxy request",
		"token", token[:min(8, len(token))]+"...",
		"method", method,
		"http_method", r.Method,
	)

	// 后续实现
	_ = token
	_ = method
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
