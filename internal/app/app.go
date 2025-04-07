package app

import (
	"context"
	"github.com/MukizuL/shortener/internal/config"
	"github.com/MukizuL/shortener/internal/storage/postgres"
	"go.uber.org/zap"
	"os/signal"
	"syscall"
)

//go:generate mockgen -source=app.go -destination=mocks/app.go -package=mocksapp

type repo interface {
	CreateShortURL(fullURL string) (string, error)
	GetLongURL(ID string) (string, error)
	OffloadStorage(filepath string) error
	Ping(ctx context.Context) error
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

func Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	defer log.Sync()

	params := config.GetParams()

	repository, err := postgres.New(ctx, params.DSN)
	if err != nil {
		return err
	}
	defer repository.Close()

	//repository, err := map_storage.New(params.Filepath)
	//if err != nil {
	//	return err
	//}

	app := NewApplication(repository, log)

	r := NewRouter(params.Base, app)

	err = runServer(ctx, params.Addr, r)
	if err != nil {
		return err
	}

	err = app.storage.OffloadStorage(params.Filepath)
	if err != nil {
		return err
	}

	return nil
}
