package respond

import (
	"encoding/json"
	"net/http"
)

type successResponse struct {
	OK   bool `json:"ok"`
	Data any  `json:"data,omitempty"`
}

type webhookSuccessResponse struct {
	OK    bool  `json:"ok"`
	LogID int64 `json:"log_id"`
}

type errorResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func OK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, successResponse{OK: true, Data: data})
}

func WebhookOK(w http.ResponseWriter, logID int64) {
	JSON(w, http.StatusOK, webhookSuccessResponse{OK: true, LogID: logID})
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, errorResponse{OK: false, Error: message})
}
