package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sunnyhmz7010/meowbridge/internal/meow"
	"github.com/sunnyhmz7010/meowbridge/internal/respond"
	"github.com/sunnyhmz7010/meowbridge/internal/store"
	"github.com/sunnyhmz7010/meowbridge/internal/webhook"
)

const maxWebhookRequestBodyBytes = 1 << 20

func (api *API) handleVerifyToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	ep, err := api.deps.Store.GetEndpointByToken(r.Context(), token)
	if errors.Is(err, store.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "token not found")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	respond.OK(w, map[string]any{"name": ep.Name, "active": ep.Active})
}

func (api *API) handleWebhook(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	// 记录 webhook 请求日志
	slog.Info("webhook received",
		"token", token[:min(8, len(token))]+"...",
		"method", r.Method,
		"path", r.URL.Path,
	)

	ep, err := api.deps.Store.GetEndpointByToken(r.Context(), token)
	if errors.Is(err, store.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "token not found")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !ep.Active {
		api.writeWebhookLog(r, ep, webhook.ParsedMessage{}, webhook.FinalMessage{}, 0, "", false, "endpoint is disabled", "")
		respond.Error(w, http.StatusForbidden, "endpoint is disabled")
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxWebhookRequestBodyBytes))
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			api.writeWebhookLog(r, ep, webhook.ParsedMessage{}, webhook.FinalMessage{}, 0, "", false, "request body too large", "")
			respond.Error(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		api.writeWebhookLog(r, ep, webhook.ParsedMessage{}, webhook.FinalMessage{}, 0, "", false, "failed to read request body", "")
		respond.Error(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		api.writeWebhookLog(r, ep, webhook.ParsedMessage{}, webhook.FinalMessage{}, 0, "", false, "request body is empty", string(body))
		respond.Error(w, http.StatusBadRequest, "request body is empty")
		return
	}

	parsed, status, parseErr := api.parseWebhookRequest(r, ep, body)
	if parseErr != nil {
		api.writeWebhookLog(r, ep, parsed, webhook.FinalMessage{}, 0, "", false, parseErr.Error(), string(body))
		respond.Error(w, status, parseErr.Error())
		return
	}

	final, err := webhook.Merge(parsed, webhook.EndpointDefaults{
		DefaultTitle:  ep.DefaultTitle,
		MsgType:       ep.MsgType,
		HTMLHeight:    ep.HTMLHeight,
		DefaultURL:    ep.DefaultURL,
		DefaultImgURL: ep.DefaultImgURL,
	}, queryOverrides(r))
	if err != nil {
		api.writeWebhookLog(r, ep, parsed, webhook.FinalMessage{}, 0, "", false, err.Error(), string(body))
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	meowResp, pushErr := api.deps.MeowClient.Push(r.Context(), meow.PushRequest{
		Nickname:   ep.MeowNickname,
		Title:      final.Title,
		Msg:        final.Msg,
		URL:        final.URL,
		ImgURL:     final.ImgURL,
		MsgType:    final.MsgType,
		HTMLHeight: final.HTMLHeight,
	})
	if pushErr != nil {
		api.writeWebhookLog(r, ep, parsed, final, meowResp.StatusCode, meowResp.Body, false, pushErr.Error(), string(body))

		// 记录推送失败日志
		slog.Error("webhook push failed",
			"endpoint_id", ep.ID,
			"meow_status", meowResp.StatusCode,
			"error", pushErr.Error(),
		)

		respond.Error(w, http.StatusBadGateway, "meow upstream request failed")
		return
	}
	logID, err := api.writeWebhookLog(r, ep, parsed, final, meowResp.StatusCode, meowResp.Body, true, "", string(body))
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to create push log")
		return
	}

	// 记录推送成功日志
	slog.Info("webhook processed",
		"endpoint_id", ep.ID,
		"log_id", logID,
		"meow_status", meowResp.StatusCode,
		"success", true,
	)

	respond.WebhookOK(w, logID)
}

func (api *API) parseWebhookRequest(r *http.Request, ep store.Endpoint, body []byte) (webhook.ParsedMessage, int, error) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err == nil && mediaType == "text/plain" {
		return webhook.ParsedMessage{SourceType: "text_plain", Msg: strings.TrimSpace(string(body)), MsgType: "text"}, http.StatusOK, nil
	}
	if config, ok := endpointParserConfig(ep); ok {
		parsed, matched, err := webhook.ParseWithConfig(webhook.ParseInput{Headers: r.Header, Body: body}, config)
		if err != nil {
			return webhook.ParsedMessage{}, http.StatusBadRequest, err
		}
		if matched {
			return parsed, http.StatusOK, nil
		}
	}
	parsed, err := webhook.Parse(webhook.ParseInput{Headers: r.Header, Body: body})
	if err != nil {
		return webhook.ParsedMessage{}, http.StatusBadRequest, err
	}
	return parsed, http.StatusOK, nil
}

func endpointParserConfig(ep store.Endpoint) (webhook.ParserConfig, bool) {
	raw := strings.TrimSpace(ep.ParserConfig)
	if raw == "" {
		return webhook.ParserConfig{}, false
	}
	var config webhook.ParserConfig
	if err := json.Unmarshal([]byte(raw), &config); err != nil {
		return webhook.ParserConfig{}, false
	}
	return config, true
}

func queryOverrides(r *http.Request) webhook.QueryOverrides {
	height, _ := strconv.Atoi(r.URL.Query().Get("htmlHeight"))
	return webhook.QueryOverrides{
		Title:      r.URL.Query().Get("title"),
		MsgType:    r.URL.Query().Get("msgType"),
		HTMLHeight: height,
		URL:        r.URL.Query().Get("url"),
		ImgURL:     r.URL.Query().Get("imgUrl"),
	}
}

func (api *API) writeWebhookLog(r *http.Request, ep store.Endpoint, parsed webhook.ParsedMessage, final webhook.FinalMessage, statusCode int, responseBody string, success bool, errorMessage string, payload string) (int64, error) {
	headers, _ := json.Marshal(r.Header)
	query, _ := json.Marshal(r.URL.Query())
	return api.deps.Store.CreatePushLog(r.Context(), store.PushLogInput{
		EndpointID:       ep.ID,
		EndpointName:     ep.Name,
		Token:            ep.Token,
		SourceType:       parsed.SourceType,
		RequestMethod:    r.Method,
		RequestHeaders:   string(headers),
		RequestQuery:     string(query),
		RequestPayload:   payload,
		ParsedTitle:      final.Title,
		ParsedMsg:        final.Msg,
		ParsedMsgType:    final.MsgType,
		MeowStatusCode:   statusCode,
		MeowResponseBody: responseBody,
		Success:          success,
		ErrorMessage:     errorMessage,
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
