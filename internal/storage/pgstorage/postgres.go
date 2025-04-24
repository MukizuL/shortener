package pgstorage

import (
	"context"
	"github.com/MukizuL/shortener/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type PGStorage struct {
	conn   *pgxpool.Pool
	logger *zap.Logger
}

func newPGStorage(cfg *config.Config, logger *zap.Logger) *PGStorage {
	dbpool, err := pgxpool.New(context.TODO(), cfg.DSN)
	if err != nil {
		panic(err)
	}

	return &PGStorage{
		conn:   dbpool,
		logger: logger,
	}
}

func Provide() fx.Option {
	return fx.Provide(newPGStorage)
}
