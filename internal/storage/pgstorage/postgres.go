package pgstorage

import "go.uber.org/fx"

type PostgresService struct {
	cfg PostgresConfig
}

type PostgresServiceIn struct {
	fx.In

	Cfg PostgresConfig
}

func NewPostgresService(in PostgresServiceIn) *PostgresService {
	return &PostgresService{
		cfg: in.Cfg,
	}
}

func ProvidePostgres() fx.Option {
	return fx.Provide(
		NewPostgresService,
	)
}
