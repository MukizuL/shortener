package mapstorage

import (
	"github.com/MukizuL/shortener/internal/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"sync"
)

type MapStorage struct {
	storage    map[string]string
	createdURL map[string]string
	userLink   map[string]map[string]struct{}
	m          sync.RWMutex
	logger     *zap.Logger
}

func newMapStorage(cfg *config.Config, logger *zap.Logger) (*MapStorage, error) {
	storage := &MapStorage{
		storage:    make(map[string]string),
		createdURL: make(map[string]string),
		userLink:   make(map[string]map[string]struct{}),
		logger:     logger,
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
