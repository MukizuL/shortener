package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"github.com/MukizuL/shortener/internal/config"
	"github.com/MukizuL/shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func newHTTPServer(lc fx.Lifecycle, cfg *config.Config, r *chi.Mux, logger *zap.Logger, storage storage.Repo) *http.Server {
	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if cfg.DSN != "" {
				err := Migrate(cfg.DSN)
				if err != nil {
					return err
				}
			}

			logger.Info("Starting HTTP server", zap.String("addr", srv.Addr))

			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}

			go srv.Serve(ln)

			return nil
		},
		OnStop: func(ctx context.Context) error {
			err := storage.OffloadStorage(ctx, cfg.Filepath)
			if err != nil {
				return err
			}

			time.Sleep(2 * time.Second)
			return srv.Shutdown(ctx)
		},
	})

	return srv
}

func Provide() fx.Option {
	return fx.Provide(newHTTPServer)
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
