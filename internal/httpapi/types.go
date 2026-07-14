package httpapi

import (
	"github.com/sunnyhmz7010/meowbridge/internal/config"
	"github.com/sunnyhmz7010/meowbridge/internal/meow"
	"github.com/sunnyhmz7010/meowbridge/internal/store"
)

type Dependencies struct {
	Store      *store.Store
	Config     config.Config
	MeowClient *meow.Client
}

type API struct {
	deps Dependencies
}
