package mapstorage

import (
	"github.com/MukizuL/shortener/internal/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"sync"
)

type MapStorage struct {
	FullURLStorage  map[string]string            // FullURLStorage[ShortURL]FullURL
	ShortURLStorage map[string]string            // ShortURLStorage[FullURL]ShortURL
	UserLinkStorage map[string]map[string]string // UserLinkStorage[UserID][ShortURL]FullURL
	m               sync.RWMutex
	logger          *zap.Logger
}

func newMapStorage(cfg *config.Config, logger *zap.Logger) (*MapStorage, error) {
	storage := &MapStorage{
		FullURLStorage:  make(map[string]string),
		ShortURLStorage: make(map[string]string),
		UserLinkStorage: make(map[string]map[string]string),
		logger:          logger,
	}

	err := storage.LoadStorage(cfg.Filepath)
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func Provide() fx.Option {
	return fx.Provide(newMapStorage)
}
