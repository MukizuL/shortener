package app

import (
	"context"
	"github.com/MukizuL/shortener/internal/dto"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

//go:generate mockgen -source=app.go -destination=mocks/app.go -package=mocksapp

type Repo interface {
	CreateShortURL(ctx context.Context, urlBase, fullURL string) (string, error)
	BatchCreateShortURL(ctx context.Context, urlBase string, data []dto.BatchRequest) ([]dto.BatchResponse, error)
	GetLongURL(ctx context.Context, ID string) (string, error)
	OffloadStorage(ctx context.Context, filepath string) error
	Ping(ctx context.Context) error
}
