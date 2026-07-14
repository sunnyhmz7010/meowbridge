package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(deps Dependencies) http.Handler {
	api := &API{deps: deps}
	r := chi.NewRouter()
	r.Post("/webhook/{token}", api.handleWebhook)
	r.Get("/verify/{token}", api.handleVerifyToken)
	return r
}
