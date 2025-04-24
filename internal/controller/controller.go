package controller

import (
	"github.com/MukizuL/shortener/internal/app"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Controller struct {
	storage app.Repo
	logger  *zap.Logger
}

func newController(storage app.Repo, logger *zap.Logger) *Controller {
	return &Controller{
		storage: storage,
		logger:  logger,
	}
}

func Provide() fx.Option {
	return fx.Provide(newController)
}
