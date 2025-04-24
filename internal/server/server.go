package server

import (
	"context"
	"errors"
	"github.com/MukizuL/shortener/internal/app"
	"github.com/MukizuL/shortener/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"net"
	"net/http"
	"path/filepath"
	"time"
)

func newHTTPServer(lc fx.Lifecycle, cfg *config.Config, r *chi.Mux, logger *zap.Logger, storage app.Repo) *http.Server {
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
