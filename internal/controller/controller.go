package controller

import (
	"github.com/MukizuL/shortener/internal/storage"
	pb "github.com/MukizuL/shortener/proto"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Controller struct {
	storage storage.Repo
	logger  *zap.Logger
	pb.UnimplementedShortenerServer
}

func newController(storage storage.Repo, logger *zap.Logger) *Controller {
	return &Controller{
		storage: storage,
		logger:  logger,
	}
}

func Provide() fx.Option {
	return fx.Provide(newController)
}
