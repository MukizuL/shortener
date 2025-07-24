package server

import (
	"context"
	"crypto/tls"
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
		Addr:         cfg.Addr,
		Handler:      r,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}

			if !cfg.HTTPS {
				logger.Info("Starting HTTP server", zap.String("addr", srv.Addr))

				go srv.Serve(ln)

				return nil
			} else {
				logger.Info("Starting HTTPS server", zap.String("addr", srv.Addr))

				tlsConfig := &tls.Config{
					CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256}, // They have assembly implementation
					MinVersion:       tls.VersionTLS12,
					CipherSuites: []uint16{
						tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
						tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
						tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
						tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
						tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
						tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
					},
				}

				srv.TLSConfig = tlsConfig

				go srv.ServeTLS(ln, cfg.Cert, cfg.PK)

				return nil
			}
		},
		OnStop: func(ctx context.Context) error {
			err := storage.OffloadStorage(ctx, cfg.Filepath)
			if err != nil {
				return err
			}

			return srv.Shutdown(ctx)
		},
	})

	return srv
}

func Provide() fx.Option {
	return fx.Provide(newHTTPServer)
}
