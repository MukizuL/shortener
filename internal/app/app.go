package app

import (
	"context"
	"errors"
	"github.com/MukizuL/shortener/internal/config"
	"github.com/MukizuL/shortener/internal/storage/map_storage"
	"github.com/MukizuL/shortener/internal/storage/pg_storage"
	"github.com/golang-migrate/migrate/v4"
	"path/filepath"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
	"os/signal"
	"syscall"
)

//go:generate mockgen -source=app.go -destination=mocks/app.go -package=mocksapp

type repo interface {
	CreateShortURL(ctx context.Context, fullURL string) (string, error)
	GetLongURL(ctx context.Context, ID string) (string, error)
	OffloadStorage(ctx context.Context, filepath string) error
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

	var repository repo
	if params.DSN != "" {
		err = Migrate(params.DSN)
		if err != nil {
			return err
		}

		db, err := pg_storage.New(ctx, params.DSN)
		if err != nil {
			return err
		}
		defer db.Close()

		repository = db

	} else {
		repository, err = map_storage.New(ctx, params.Filepath)
		if err != nil {
			return err
		}
	}

	app := NewApplication(repository, log)

	r := NewRouter(params.Base, app)

	err = runServer(ctx, params.Addr, r)
	if err != nil {
		return err
	}

	err = app.storage.OffloadStorage(ctx, params.Filepath)
	if err != nil {
		return err
	}

	return nil
}

func Migrate(DSN string) error {
	_, err := filepath.Abs("./migrations")
	if err != nil {
		return err
	}

	m, err := migrate.New("file://migrations", DSN+"?sslmode=disable")
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
