package mapstorage

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/MukizuL/shortener/internal/dto"
	"github.com/MukizuL/shortener/internal/errs"
	"github.com/MukizuL/shortener/internal/helpers"
	"github.com/MukizuL/shortener/internal/models"
	"go.uber.org/zap"
	"io"
	"os"
	"sync"
)

type MapStorage struct {
	storage    map[string]string
	createdURL map[string]string
	m          sync.RWMutex
	logger     *zap.Logger
}

func New(ctx context.Context, filepath string, logger *zap.Logger) (*MapStorage, error) {
	storage := &MapStorage{
		storage:    make(map[string]string),
		createdURL: make(map[string]string),
		logger:     logger,
	}

	err := storage.LoadStorage(ctx, filepath)
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (r *MapStorage) CreateShortURL(ctx context.Context, fullURL string) (string, error) {
	r.m.Lock()
	defer r.m.Unlock()

	if v, exist := r.createdURL[fullURL]; exist {
		return "http://localhost:8080/" + v, errs.ErrDuplicate
	}

	ID := helpers.RandomString(6)
	shortURL := "http://localhost:8080/" + ID

	r.createdURL[fullURL] = ID

	r.storage[ID] = fullURL

	return shortURL, nil
}

func (r *MapStorage) BatchCreateShortURL(ctx context.Context, data []dto.BatchRequest) ([]dto.BatchResponse, error) {
	r.m.Lock()
	defer r.m.Unlock()

	result := make([]dto.BatchResponse, 0, len(data))

	for _, v := range data {
		if _, exist := r.createdURL[v.OriginalURL]; exist {
			return nil, errs.ErrDuplicate
		}

		ID := helpers.RandomString(6)
		shortURL := "http://localhost:8080/" + ID

		r.createdURL[v.OriginalURL] = ID

		result = append(result, dto.BatchResponse{CorrelationID: v.CorrelationID, ShortURL: shortURL})

		r.storage[ID] = v.OriginalURL
	}

	return result, nil
}

func (r *MapStorage) GetLongURL(ctx context.Context, ID string) (string, error) {
	r.m.RLock()
	defer r.m.RUnlock()

	if val, exist := r.storage[ID]; !exist {
		return "", errs.ErrNotFound
	} else {
		return val, nil
	}
}

func (r *MapStorage) LoadStorage(ctx context.Context, filepath string) error {
	r.m.Lock()
	defer r.m.Unlock()

	file, err := os.Open(filepath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		r.logger.Error("mapstorage:LoadStorage Error opening file", zap.Error(err))
		return errs.ErrInternalServerError
	}

	defer file.Close()

	var data []models.Urls
	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}

		r.logger.Error("mapstorage:LoadStorage Error opening file", zap.Error(err))
		return errs.ErrInternalServerError
	}

	for _, entry := range data {
		r.storage[entry.ShortURL] = entry.OriginalURL
		r.createdURL[entry.OriginalURL] = entry.ShortURL
	}

	return nil
}

func (r *MapStorage) OffloadStorage(ctx context.Context, filepath string) error {
	r.m.Lock()
	defer r.m.Unlock()

	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		r.logger.Error("mapstorage:OffloadStorage Error opening file", zap.Error(err))
		return errs.ErrInternalServerError
	}
	defer file.Close()

	var data []models.Urls
	for k, v := range r.storage {
		data = append(data, models.Urls{
			ShortURL:    k,
			OriginalURL: v,
		})
	}

	err = json.NewEncoder(file).Encode(&data)
	if err != nil {
		r.logger.Error("mapstorage:OffloadStorage Error encoding data", zap.Error(err))
		return errs.ErrInternalServerError
	}

	return nil
}

func (r *MapStorage) Ping(ctx context.Context) error {
	return nil
}
