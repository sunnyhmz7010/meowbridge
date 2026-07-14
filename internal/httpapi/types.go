package httpapi

import (
	"context"

	"github.com/sunnyhmz7010/meowbridge/internal/config"
	"github.com/sunnyhmz7010/meowbridge/internal/meow"
	"github.com/sunnyhmz7010/meowbridge/internal/store"
)

type webhookStore interface {
	GetEndpointByToken(ctx context.Context, token string) (store.Endpoint, error)
	CreatePushLog(ctx context.Context, input store.PushLogInput) (int64, error)
}

type Dependencies struct {
	Store      webhookStore
	Config     config.Config
	MeowClient *meow.Client
}

type API struct {
	deps Dependencies
}
