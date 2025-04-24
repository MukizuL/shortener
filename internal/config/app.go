package config

import (
	"github.com/MukizuL/shortener/internal/storage/pgstorage"
	"go.uber.org/fx"
)

type ApplicationConfig struct {
	fx.Out

	PostgresConfig pgstorage.PostgresConfig
}

func newAppConfig() ApplicationConfig {
	return ApplicationConfig{
		PostgresConfig: &pgstorage.PostgresParams{DSN: ""},
	}
}

func Provide() fx.Option {
	return fx.Provide(newAppConfig)
}
