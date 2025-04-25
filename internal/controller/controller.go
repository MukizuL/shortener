package controller

import (
	jwtService "github.com/MukizuL/shortener/internal/jwt"
	"github.com/MukizuL/shortener/internal/storage"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Controller struct {
	jwtService jwtService.JWTServiceInterface
	storage    storage.Repo
	logger     *zap.Logger
}

func newController(jwtService jwtService.JWTServiceInterface, storage storage.Repo, logger *zap.Logger) *Controller {
	return &Controller{
		jwtService: jwtService,
		storage:    storage,
		logger:     logger,
	}
}

func Provide() fx.Option {
	return fx.Provide(newController)
}
