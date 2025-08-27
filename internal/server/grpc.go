package server

import (
	"context"
	"net"

	"github.com/MukizuL/shortener/internal/config"
	"github.com/MukizuL/shortener/internal/controller"
	"github.com/MukizuL/shortener/internal/interceptor"
	"github.com/MukizuL/shortener/internal/migration"
	"github.com/MukizuL/shortener/internal/storage"
	pb "github.com/MukizuL/shortener/proto"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GRPCFxIn struct {
	fx.In

	Lc          fx.Lifecycle
	Ctrl        *controller.Controller
	Cfg         *config.Config
	Logger      *zap.Logger
	Interceptor *interceptor.Service
	Storage     storage.Repo
	Migrator    *migration.Migrator
}

func newGRPCServer(in GRPCFxIn) (*grpc.Server, error) {
	if in.Cfg.GRPCPort == "" {
		in.Logger.Info("GRPC server is disabled")
		return nil, nil
	}

	ln, err := net.Listen("tcp", in.Cfg.GRPCPort)
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			in.Interceptor.Logger,
			in.Interceptor.Auth,
			in.Interceptor.IsTrustedCIDR,
		),
	)

	pb.RegisterShortenerServer(s, in.Ctrl)

	in.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			in.Logger.Info("Starting GRPC server", zap.String("port", in.Cfg.GRPCPort))
			go s.Serve(ln)

			return nil
		},
		OnStop: func(ctx context.Context) error {
			go s.GracefulStop()

			return nil
		},
	})

	return s, nil
}
