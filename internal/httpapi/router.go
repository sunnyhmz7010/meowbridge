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
	r.Route("/api/admin", func(r chi.Router) {
		r.Post("/login", api.handleLogin)
		r.Group(func(r chi.Router) {
			r.Use(api.requireAdmin)
			r.Get("/endpoints", api.handleListEndpoints)
			r.Post("/endpoints", api.handleCreateEndpoint)
			r.Get("/endpoints/{id}", api.handleGetEndpoint)
			r.Put("/endpoints/{id}", api.handleUpdateEndpoint)
			r.Delete("/endpoints/{id}", api.handleDeleteEndpoint)
			r.Post("/endpoints/{id}/reset-token", api.handleResetEndpointToken)
			r.Patch("/endpoints/{id}/active", api.handleSetEndpointActive)
			r.Get("/push-logs", api.handleListPushLogs)
			r.Get("/push-logs/{id}", api.handleGetPushLog)
			r.Delete("/push-logs", api.handleCleanupPushLogs)
			r.Get("/settings", api.handleGetSettings)
			r.Put("/settings", api.handleUpdateSettings)
			r.Post("/change-password", api.handleChangePassword)
			r.Get("/webhook/presets", api.handleWebhookPresets)
		})
	})
	return r
}
