package tgproxy

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sunnyhmz7010/meowbridge/internal/meow"
	"github.com/sunnyhmz7010/meowbridge/internal/respond"
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

	// 读取请求体
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		respond.ErrorCode(w, http.StatusBadRequest, "INVALID_PAYLOAD", "cannot read request body")
		return
	}

	// 解析 TG 请求
	msg, err := ParseTGRequest(body, method)
	if err != nil {
		respond.ErrorCode(w, http.StatusBadRequest, "INVALID_PAYLOAD", "invalid JSON payload")
		return
	}

	// 获取消息内容
	content := msg.GetContent()
	if content == "" {
		respond.ErrorCode(w, http.StatusBadRequest, "MISSING_FIELD", "no text or caption field")
		return
	}

	// 记录解析结果
	slog.Info("tg message parsed",
		"method", method,
		"parse_mode", msg.ParseMode,
		"content_length", len(content),
	)

	// 后续实现格式转换和推送
	_ = content
	_ = msg.ParseMode
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
