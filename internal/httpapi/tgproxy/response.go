package tgproxy

import (
	"encoding/json"
	"net/http"
	"time"
)

type TGSuccessResponse struct {
	OK     bool        `json:"ok"`
	Result interface{} `json:"result"`
}

type TGErrorResponse struct {
	OK          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
}

type TGMessageResult struct {
	MessageID int64       `json:"message_id"`
	Date      int64       `json:"date"`
	Chat      interface{} `json:"chat"`
	Text      string      `json:"text"`
}

type TGChat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

func RespondTGSuccess(w http.ResponseWriter, content string) {
	resp := TGSuccessResponse{
		OK: true,
		Result: TGMessageResult{
			MessageID: time.Now().Unix(),
			Date:      time.Now().Unix(),
			Chat:      TGChat{ID: 0, Type: "private"},
			Text:      content,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func RespondTGError(w http.ResponseWriter, status int, description string) {
	resp := TGErrorResponse{
		OK:          false,
		ErrorCode:   status,
		Description: description,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // TG API 总是返回 200，错误在 body 中
	json.NewEncoder(w).Encode(resp)
}
