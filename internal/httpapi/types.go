package httpapi

import (
	"context"
	"time"

	"github.com/sunnyhmz7010/meowbridge/internal/config"
	"github.com/sunnyhmz7010/meowbridge/internal/meow"
	"github.com/sunnyhmz7010/meowbridge/internal/store"
)

type webhookStore interface {
	GetEndpointByToken(ctx context.Context, token string) (store.Endpoint, error)
	CreatePushLog(ctx context.Context, input store.PushLogInput) (int64, error)
}

type apiStore interface {
	webhookStore
	AdminPasswordHash(ctx context.Context) (string, error)
	UpdateAdminPasswordHash(ctx context.Context, hash string) error
	CreateEndpoint(ctx context.Context, input store.EndpointInput) (store.Endpoint, error)
	ListEndpoints(ctx context.Context) ([]store.Endpoint, error)
	GetEndpoint(ctx context.Context, id int64) (store.Endpoint, error)
	UpdateEndpoint(ctx context.Context, id int64, input store.EndpointUpdate) (store.Endpoint, error)
	DeleteEndpoint(ctx context.Context, id int64) error
	ResetEndpointToken(ctx context.Context, id int64, newToken string) (store.Endpoint, error)
	SetEndpointActive(ctx context.Context, id int64, active bool) error
	ListPushLogs(ctx context.Context) ([]store.PushLog, error)
	GetPushLog(ctx context.Context, id int64) (store.PushLog, error)
	CleanupPushLogs(ctx context.Context, before time.Time) (int64, error)
	ListSettings(ctx context.Context) (map[string]string, error)
	SetSetting(ctx context.Context, key, value string) error
}

type Dependencies struct {
	Store      apiStore
	Config     config.Config
	MeowClient *meow.Client
}

type API struct {
	deps Dependencies
}
