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

type HTTPFxIn struct {
	fx.In

	Lc       fx.Lifecycle
	Cfg      *config.Config
	R        *chi.Mux
	Logger   *zap.Logger
	Storage  storage.Repo
	Migrator *migration.Migrator
}

func newHTTPServer(in HTTPFxIn) *http.Server {
	srv := &http.Server{
		Addr:         in.Cfg.Addr,
		Handler:      in.R,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	in.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}

			if !in.Cfg.HTTPS {
				in.Logger.Info("Starting HTTP server", zap.String("addr", srv.Addr))

				go srv.Serve(ln)

				return nil
			} else {
				in.Logger.Info("Starting HTTPS server", zap.String("addr", srv.Addr))

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

				go srv.ServeTLS(ln, in.Cfg.Cert, in.Cfg.PK)

				return nil
			}
		},
		OnStop: func(ctx context.Context) error {
			err := in.Storage.OffloadStorage(ctx, in.Cfg.Filepath)
			if err != nil {
				return err
			}

			return srv.Shutdown(ctx)
		},
	})

	return srv
}

func Provide() fx.Option {
	return fx.Provide(newHTTPServer, newGRPCServer)
}
