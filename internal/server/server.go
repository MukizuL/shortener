package server

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/MukizuL/shortener/internal/config"
	"github.com/MukizuL/shortener/internal/migration"
	"github.com/MukizuL/shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func newHTTPServer(lc fx.Lifecycle, cfg *config.Config, r *chi.Mux, logger *zap.Logger, storage storage.Repo, migrator *migration.Migrator) *http.Server {
	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
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
