package tgproxy

import (
	"context"
	"io"
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

	// 1. Token 校验
	ep, err := h.deps.Store.GetEndpointByToken(r.Context(), token)
	if err != nil {
		RespondTGError(w, http.StatusUnauthorized, "Unauthorized: invalid token")
		return
	}

	// 2. 接口状态校验
	if !ep.Active {
		RespondTGError(w, http.StatusForbidden, "Forbidden: endpoint is disabled")
		return
	}

	// 3. 只处理 sendMessage，其他方法伪造成功
	if method != "sendMessage" && method != "sendPhoto" && method != "sendDocument" {
		RespondTGSuccess(w, "")
		return
	}

	// 4. 读取请求体
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		RespondTGError(w, http.StatusBadRequest, "Bad Request: cannot read body")
		return
	}

	// 5. 解析 TG 请求
	msg, err := ParseTGRequest(body, method)
	if err != nil {
		RespondTGError(w, http.StatusBadRequest, "Bad Request: invalid JSON")
		return
	}

	// 6. 获取消息内容
	content := msg.GetContent()
	if content == "" {
		RespondTGError(w, http.StatusBadRequest, "Bad Request: no text or caption")
		return
	}

	// 7. 格式转换
	convertedContent, msgType := ConvertTGFormat(content, msg.ParseMode)

	// 8. 推送到 MeoW
	pushReq := meow.PushRequest{
		Nickname: ep.MeowNickname,
		Msg:      convertedContent,
		MsgType:  msgType,
		Title:    ep.DefaultTitle,
		URL:      ep.DefaultURL,
		ImgURL:   ep.DefaultImgURL,
	}

	resp, retryCount, pushErr := h.deps.MeowClient.PushWithRetry(r.Context(), pushReq)

	// 9. 记录日志
	if pushErr != nil || resp.StatusCode >= 400 {
		slog.Error("tg proxy push failed",
			"endpoint_id", ep.ID,
			"method", method,
			"retry_count", retryCount,
			"meow_status", resp.StatusCode,
			"error", pushErr,
		)
	} else {
		slog.Info("tg proxy result",
			"endpoint_id", ep.ID,
			"method", method,
			"retry_count", retryCount,
			"meow_status", resp.StatusCode,
			"success", true,
		)
	}

	// 10. 伪造成功响应
	RespondTGSuccess(w, content)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
