package migration

import (
	"database/sql"
	"embed"

	"github.com/MukizuL/shortener/internal/config"
	"github.com/pressly/goose/v3"
	"go.uber.org/fx"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed "migrations/*.sql"
var embedMigrations embed.FS

type Migrator struct{}

func newMigrator(cfg *config.Config) (*Migrator, error) {
	if cfg.DSN == "" {
		return &Migrator{}, nil
	}

	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}

	if cfg.Debug {
		// Should not be in release
		goose.Reset(db, "migrations")
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return nil, err
	}

	return &Migrator{}, nil
}

func Provide() fx.Option {
	return fx.Provide(newMigrator)
}
