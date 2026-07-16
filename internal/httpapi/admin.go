package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sunnyhmz7010/meowbridge/internal/auth"
	"github.com/sunnyhmz7010/meowbridge/internal/respond"
	"github.com/sunnyhmz7010/meowbridge/internal/store"
	"github.com/sunnyhmz7010/meowbridge/internal/token"
	"github.com/sunnyhmz7010/meowbridge/internal/webhook"
)

type pushLogListItem struct {
	ID             int64     `json:"id"`
	EndpointID     int64     `json:"endpoint_id"`
	EndpointName   string    `json:"endpoint_name"`
	SourceType     string    `json:"source_type"`
	ParsedTitle    string    `json:"parsed_title"`
	ParsedMsg      string    `json:"parsed_msg"`
	ParsedMsgType  string    `json:"parsed_msg_type"`
	MeowStatusCode int       `json:"meow_status_code"`
	Success        bool      `json:"success"`
	ErrorMessage   string    `json:"error_message"`
	CreatedAt      time.Time `json:"created_at"`
}

type loginRequest struct {
	Password string `json:"password"`
}

type setupRequest struct {
	Password string `json:"password"`
}

type endpointRequest struct {
	Name          string           `json:"name"`
	MeowNickname  string           `json:"meow_nickname"`
	DefaultTitle  string           `json:"default_title"`
	MsgType       string           `json:"msg_type"`
	HTMLHeight    int              `json:"html_height"`
	DefaultURL    string           `json:"default_url"`
	DefaultImgURL string           `json:"default_img_url"`
	ParserConfig  *json.RawMessage `json:"parser_config"`
	Active        *bool            `json:"active"`
}

type endpointResponse struct {
	ID            int64           `json:"id"`
	Name          string          `json:"name"`
	Token         string          `json:"token"`
	MeowNickname  string          `json:"meow_nickname"`
	DefaultTitle  string          `json:"default_title"`
	MsgType       string          `json:"msg_type"`
	HTMLHeight    int             `json:"html_height"`
	DefaultURL    string          `json:"default_url"`
	DefaultImgURL string          `json:"default_img_url"`
	ParserConfig  json.RawMessage `json:"parser_config"`
	Active        bool            `json:"active"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type webhookPreviewRequest struct {
	ParserConfig webhook.ParserConfig `json:"parser_config"`
	Payload      json.RawMessage      `json:"payload"`
}

func (api *API) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	hash, err := api.deps.Store.AdminPasswordHash(r.Context())
	if err != nil || !auth.VerifyPassword(hash, req.Password) {
		respond.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	raw, err := auth.IssueJWT(api.deps.Config.JWTSecret, 24*time.Hour)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to issue token")
		return
	}
	respond.OK(w, map[string]string{"token": raw})
}

func (api *API) handleSetupStatus(w http.ResponseWriter, r *http.Request) {
	exists, err := api.deps.Store.AdminExists(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to load setup status")
		return
	}
	respond.OK(w, map[string]bool{"needs_setup": !exists})
}

func (api *API) handleSetup(w http.ResponseWriter, r *http.Request) {
	if !isJSONRequest(r) {
		respond.Error(w, http.StatusUnsupportedMediaType, "content-type must be application/json")
		return
	}
	if isCrossSiteRequest(r) {
		respond.Error(w, http.StatusForbidden, "cross-site setup is not allowed")
		return
	}

	exists, err := api.deps.Store.AdminExists(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to load setup status")
		return
	}
	if exists {
		respond.Error(w, http.StatusConflict, "admin already initialized")
		return
	}

	var req setupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	if strings.TrimSpace(req.Password) == "" {
		respond.Error(w, http.StatusBadRequest, "password is required")
		return
	}
	if err := api.deps.Store.CreateInitialAdmin(r.Context(), req.Password); errors.Is(err, store.ErrAdminAlreadyInitialized) {
		respond.Error(w, http.StatusConflict, "admin already initialized")
		return
	} else if errors.Is(err, store.ErrBlankAdminPassword) {
		respond.Error(w, http.StatusBadRequest, "password is required")
		return
	} else if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to initialize admin")
		return
	}
	raw, err := auth.IssueJWT(api.deps.Config.JWTSecret, 24*time.Hour)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to issue token")
		return
	}
	respond.OK(w, map[string]string{"token": raw})
}

func isJSONRequest(r *http.Request) bool {
	contentType := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
	return contentType == "application/json" || strings.HasPrefix(contentType, "application/json;")
}

func isCrossSiteRequest(r *http.Request) bool {
	if strings.EqualFold(r.Header.Get("Sec-Fetch-Site"), "cross-site") {
		return true
	}
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return false
	}
	return !sameOrigin(origin, r)
}

func sameOrigin(origin string, r *http.Request) bool {
	httpOrigin := "http://" + r.Host
	httpsOrigin := "https://" + r.Host
	return strings.EqualFold(origin, httpOrigin) || strings.EqualFold(origin, httpsOrigin)
}

func (api *API) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		raw := strings.TrimPrefix(authHeader, "Bearer ")
		if raw == authHeader || raw == "" {
			respond.Error(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		if err := auth.VerifyJWT(api.deps.Config.JWTSecret, raw); err != nil {
			respond.Error(w, http.StatusUnauthorized, "invalid token")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (api *API) handleCreateEndpoint(w http.ResponseWriter, r *http.Request) {
	var req endpointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Name == "" || req.MeowNickname == "" {
		respond.Error(w, http.StatusBadRequest, "name and meow_nickname are required")
		return
	}
	parserConfig, err := rawParserConfig(req.ParserConfig)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	tok, err := token.Generate()
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	ep, err := api.deps.Store.CreateEndpoint(r.Context(), store.EndpointInput{Name: req.Name, Token: tok, MeowNickname: req.MeowNickname, DefaultTitle: req.DefaultTitle, MsgType: defaultString(req.MsgType, "text"), HTMLHeight: defaultInt(req.HTMLHeight, 200), DefaultURL: req.DefaultURL, DefaultImgURL: req.DefaultImgURL, ParserConfig: parserConfig, Active: defaultBool(req.Active, true)})
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to create endpoint")
		return
	}
	respond.OK(w, endpointToResponse(ep))
}

func (api *API) handleListEndpoints(w http.ResponseWriter, r *http.Request) {
	endpoints, err := api.deps.Store.ListEndpoints(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to list endpoints")
		return
	}
	items := make([]endpointResponse, 0, len(endpoints))
	for _, ep := range endpoints {
		items = append(items, endpointToResponse(ep))
	}
	respond.OK(w, items)
}

func endpointID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}
func rawParserConfig(value *json.RawMessage) (string, error) {
	if value == nil {
		return "", nil
	}
	return webhook.NormalizeParserConfig(*value)
}
func parserConfigJSON(value string) json.RawMessage {
	value = strings.TrimSpace(value)
	if value == "" {
		return json.RawMessage("null")
	}
	normalized, err := webhook.NormalizeParserConfig([]byte(value))
	if err != nil || normalized == "" || !json.Valid([]byte(normalized)) {
		return json.RawMessage("null")
	}
	return json.RawMessage(normalized)
}
func endpointToResponse(ep store.Endpoint) endpointResponse {
	return endpointResponse{
		ID:            ep.ID,
		Name:          ep.Name,
		Token:         ep.Token,
		MeowNickname:  ep.MeowNickname,
		DefaultTitle:  ep.DefaultTitle,
		MsgType:       ep.MsgType,
		HTMLHeight:    ep.HTMLHeight,
		DefaultURL:    ep.DefaultURL,
		DefaultImgURL: ep.DefaultImgURL,
		ParserConfig:  parserConfigJSON(ep.ParserConfig),
		Active:        ep.Active,
		CreatedAt:     ep.CreatedAt,
		UpdatedAt:     ep.UpdatedAt,
	}
}
func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
func defaultInt(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}
func defaultBool(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}
func truncateString(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}

func (api *API) handleGetEndpoint(w http.ResponseWriter, r *http.Request) {
	id, err := endpointID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid endpoint id")
		return
	}
	ep, err := api.deps.Store.GetEndpoint(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "endpoint not found")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to get endpoint")
		return
	}
	respond.OK(w, endpointToResponse(ep))
}

func (api *API) handleUpdateEndpoint(w http.ResponseWriter, r *http.Request) {
	id, err := endpointID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid endpoint id")
		return
	}
	var req endpointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	existing, err := api.deps.Store.GetEndpoint(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "endpoint not found")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to get endpoint")
		return
	}
	active := existing.Active
	if req.Active != nil {
		active = *req.Active
	}
	parserConfig := existing.ParserConfig
	if req.ParserConfig != nil {
		parserConfig, err = rawParserConfig(req.ParserConfig)
		if err != nil {
			respond.Error(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	ep, err := api.deps.Store.UpdateEndpoint(r.Context(), id, store.EndpointUpdate{Name: req.Name, DefaultTitle: req.DefaultTitle, MsgType: defaultString(req.MsgType, "text"), HTMLHeight: defaultInt(req.HTMLHeight, 200), DefaultURL: req.DefaultURL, DefaultImgURL: req.DefaultImgURL, ParserConfig: parserConfig, Active: active})
	if errors.Is(err, store.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "endpoint not found")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to update endpoint")
		return
	}
	respond.OK(w, endpointToResponse(ep))
}

func (api *API) handleDeleteEndpoint(w http.ResponseWriter, r *http.Request) {
	id, err := endpointID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid endpoint id")
		return
	}
	if err := api.deps.Store.DeleteEndpoint(r.Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "endpoint not found")
		} else {
			respond.Error(w, http.StatusInternalServerError, "failed to delete endpoint")
		}
		return
	}
	respond.OK(w, map[string]bool{"deleted": true})
}

func (api *API) handleResetEndpointToken(w http.ResponseWriter, r *http.Request) {
	id, err := endpointID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid endpoint id")
		return
	}
	newToken, err := token.Generate()
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	ep, err := api.deps.Store.ResetEndpointToken(r.Context(), id, newToken)
	if errors.Is(err, store.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "endpoint not found")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to reset token")
		return
	}
	respond.OK(w, endpointToResponse(ep))
}

func (api *API) handleSetEndpointActive(w http.ResponseWriter, r *http.Request) {
	id, err := endpointID(r)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid endpoint id")
		return
	}
	var req struct {
		Active *bool `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Active == nil {
		respond.Error(w, http.StatusBadRequest, "active is required")
		return
	}
	if err := api.deps.Store.SetEndpointActive(r.Context(), id, *req.Active); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respond.Error(w, http.StatusNotFound, "endpoint not found")
		} else {
			respond.Error(w, http.StatusInternalServerError, "failed to update active state")
		}
		return
	}
	respond.OK(w, map[string]bool{"active": *req.Active})
}

func (api *API) handleListPushLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := api.deps.Store.ListPushLogs(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to list push logs")
		return
	}
	items := make([]pushLogListItem, 0, len(logs))
	for _, log := range logs {
		items = append(items, pushLogListItem{ID: log.ID, EndpointID: log.EndpointID, EndpointName: log.EndpointName, SourceType: log.SourceType, ParsedTitle: log.ParsedTitle, ParsedMsg: truncateString(log.ParsedMsg, 200), ParsedMsgType: log.ParsedMsgType, MeowStatusCode: log.MeowStatusCode, Success: log.Success, ErrorMessage: log.ErrorMessage, CreatedAt: log.CreatedAt})
	}
	respond.OK(w, items)
}

func (api *API) handleGetPushLog(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid log id")
		return
	}
	log, err := api.deps.Store.GetPushLog(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		respond.Error(w, http.StatusNotFound, "push log not found")
		return
	}
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to get push log")
		return
	}
	respond.OK(w, log)
}

func (api *API) handleCleanupPushLogs(w http.ResponseWriter, r *http.Request) {
	values, err := api.deps.Store.ListSettings(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	days, _ := strconv.Atoi(values["log_retention_days"])
	if days <= 0 {
		days = 14
	}
	deleted, err := api.deps.Store.CleanupPushLogs(r.Context(), time.Now().UTC().AddDate(0, 0, -days))
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to cleanup push logs")
		return
	}
	respond.OK(w, map[string]int64{"deleted": deleted})
}

func (api *API) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	values, err := api.deps.Store.ListSettings(r.Context())
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to list settings")
		return
	}
	respond.OK(w, publicSettings(values))
}

func (api *API) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var values map[string]string
	if err := json.NewDecoder(r.Body).Decode(&values); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	if value, ok := values["log_retention_days"]; ok {
		days, err := strconv.Atoi(value)
		if err != nil || days <= 0 {
			respond.Error(w, http.StatusBadRequest, "log_retention_days must be a positive integer")
			return
		}
	}
	if value, ok := values["log_retention_days"]; ok {
		if err := api.deps.Store.SetSetting(r.Context(), "log_retention_days", value); err != nil {
			respond.Error(w, http.StatusInternalServerError, "failed to update settings")
			return
		}
	}
	respond.OK(w, map[string]bool{"updated": true})
}

func publicSettings(values map[string]string) map[string]string {
	return map[string]string{
		"log_retention_days": values["log_retention_days"],
	}
}

func (api *API) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	if strings.TrimSpace(req.NewPassword) == "" {
		respond.Error(w, http.StatusBadRequest, "new_password is required")
		return
	}
	hash, err := api.deps.Store.AdminPasswordHash(r.Context())
	if err != nil || !auth.VerifyPassword(hash, req.OldPassword) {
		respond.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	newHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to hash password")
		return
	}
	if err := api.deps.Store.UpdateAdminPasswordHash(r.Context(), newHash); err != nil {
		respond.Error(w, http.StatusInternalServerError, "failed to change password")
		return
	}
	respond.OK(w, map[string]bool{"changed": true})
}

func (api *API) handleWebhookPresets(w http.ResponseWriter, r *http.Request) {
	respond.OK(w, webhook.ParserPresets())
}

func (api *API) handleWebhookPreview(w http.ResponseWriter, r *http.Request) {
	var req webhookPreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, http.StatusBadRequest, "invalid json")
		return
	}
	if len(req.Payload) == 0 || !json.Valid(req.Payload) {
		respond.Error(w, http.StatusBadRequest, "payload must be valid json")
		return
	}
	input := webhook.ParseInput{Headers: r.Header, Body: req.Payload}
	parsed, matched, err := webhook.ParseWithConfig(input, req.ParserConfig)
	if err != nil {
		respond.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	if !matched {
		parsed, err = webhook.Parse(input)
		if err != nil {
			respond.Error(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	respond.OK(w, parsed)
}
