package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sunnyhmz7010/meowbridge/internal/meow"
	"github.com/sunnyhmz7010/meowbridge/internal/respond"
	"github.com/sunnyhmz7010/meowbridge/internal/store"
	"github.com/sunnyhmz7010/meowbridge/internal/webhook"
)

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
	ep, err := api.deps.Store.GetEndpointByToken(r.Context(), token)
	if errors.Is(err, store.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "token not found")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "internal error")
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1024*1024))
	if err != nil {
		api.writeWebhookLog(r, ep, webhook.ParsedMessage{}, webhook.FinalMessage{}, 0, "", false, "failed to read request body", "")
		respond.Error(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	if !ep.Active {
		api.writeWebhookLog(r, ep, webhook.ParsedMessage{}, webhook.FinalMessage{}, 0, "", false, "endpoint is disabled", string(body))
		respond.Error(w, http.StatusForbidden, "endpoint is disabled")
		return
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		api.writeWebhookLog(r, ep, webhook.ParsedMessage{}, webhook.FinalMessage{}, 0, "", false, "request body is empty", string(body))
		respond.Error(w, http.StatusBadRequest, "request body is empty")
		return
	}

	parsed, status, parseErr := api.parseWebhookRequest(r, body)
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
		respond.Error(w, http.StatusBadGateway, pushErr.Error())
		return
	}
	logID := api.writeWebhookLog(r, ep, parsed, final, meowResp.StatusCode, meowResp.Body, true, "", string(body))
	respond.WebhookOK(w, logID)
}

func (api *API) parseWebhookRequest(r *http.Request, body []byte) (webhook.ParsedMessage, int, error) {
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "text/plain") {
		return webhook.ParsedMessage{SourceType: "text_plain", Msg: strings.TrimSpace(string(body)), MsgType: "text"}, http.StatusOK, nil
	}
	parsed, err := webhook.Parse(webhook.ParseInput{Headers: r.Header, Body: body})
	if err != nil {
		return webhook.ParsedMessage{}, http.StatusBadRequest, err
	}
	return parsed, http.StatusOK, nil
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

func (api *API) writeWebhookLog(r *http.Request, ep store.Endpoint, parsed webhook.ParsedMessage, final webhook.FinalMessage, statusCode int, responseBody string, success bool, errorMessage string, payload string) int64 {
	headers, _ := json.Marshal(r.Header)
	query, _ := json.Marshal(r.URL.Query())
	id, _ := api.deps.Store.CreatePushLog(r.Context(), store.PushLogInput{
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
	return id
}
