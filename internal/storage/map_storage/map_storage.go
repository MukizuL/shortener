package map_storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MukizuL/shortener/internal/errs"
	"github.com/MukizuL/shortener/internal/helpers"
	"github.com/MukizuL/shortener/internal/models"
	"io"
	"os"
	"sync"
)

type MapStorage struct {
	storage    map[string]string
	createdURL map[string]struct{}
	m          sync.RWMutex
}

func New(ctx context.Context, filepath string) (*MapStorage, error) {
	storage := &MapStorage{
		storage:    make(map[string]string),
		createdURL: make(map[string]struct{}),
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

	if _, exist := r.createdURL[fullURL]; exist {
		return "", errs.ErrDuplicate
	}

	r.createdURL[fullURL] = struct{}{}

	ID := helpers.RandomString(6)
	shortURL := "http://localhost:8080/" + ID

	r.storage[ID] = fullURL

	return shortURL, nil
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
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("open storage file: %w", err)
	}

	defer file.Close()

	var data []models.Urls
	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}

		return err
	}

	for _, entry := range data {
		r.storage[entry.ShortURL] = entry.OriginalURL
		r.createdURL[entry.OriginalURL] = struct{}{}
	}

	return nil
}

func (r *MapStorage) OffloadStorage(ctx context.Context, filepath string) error {
	r.m.Lock()
	defer r.m.Unlock()

	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic("Error opening file" + err.Error())
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
		return err
	}

	return nil
}

func (r *MapStorage) Ping(ctx context.Context) error {
	return nil
}
