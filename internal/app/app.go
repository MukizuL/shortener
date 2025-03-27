package app

import (
	"context"
	"github.com/MukizuL/shortener/internal/config"
	"github.com/MukizuL/shortener/internal/storage"
	"go.uber.org/zap"
	"os/signal"
	"syscall"
)

type repo interface {
	CreateShortURL(fullURL string) (string, error)
	GetLongURL(ID string) (string, error)
	OffloadStorage(filepath string) error
}
type Application struct {
	storage repo
	logger  *zap.Logger
}

func NewApplication(storage repo, logger *zap.Logger) *Application {
	return &Application{
		storage: storage,
		logger:  logger,
	}
}

func Run() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	params := config.GetParams()

	app := NewApplication(storage.New(params.Filepath), log)

	r := NewRouter(params.Base, app)

	err = runServer(ctx, params.Addr, r)
	if err != nil {
		panic(err)
	}

	err = app.storage.OffloadStorage(params.Filepath)
	if err != nil {
		panic(err)
	}
}
