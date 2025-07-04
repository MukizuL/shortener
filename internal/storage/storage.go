package storage

import (
	"context"

	"github.com/MukizuL/shortener/internal/config"
	"github.com/MukizuL/shortener/internal/dto"
	"github.com/MukizuL/shortener/internal/storage/mapstorage"
	"github.com/MukizuL/shortener/internal/storage/pgstorage"
	"go.uber.org/fx"
)

//go:generate mockgen -source=storage.go -destination=mocks/storage.go -package=mockstorage

type Repo interface {
	CreateShortURL(ctx context.Context, userID, urlBase, fullURL string) (string, error)
	BatchCreateShortURL(ctx context.Context, userID, urlBase string, data []dto.BatchRequest) ([]dto.BatchResponse, error)
	GetLongURL(ctx context.Context, ID string) (string, error)
	GetUserURLs(ctx context.Context, userID string) ([]dto.URLPair, error)
	DeleteURLs(ctx context.Context, userID string, urls []string) error
	OffloadStorage(ctx context.Context, filepath string) error
	Ping(ctx context.Context) error
}

type Repository struct {
	r Repo
}

func newRepository(cfg *config.Config, m *mapstorage.MapStorage, p *pgstorage.PGStorage) Repo {
	if cfg.DSN == "" {
		return m
	}

	return p
}

func Provide() fx.Option {
	return fx.Provide(newRepository)
}
