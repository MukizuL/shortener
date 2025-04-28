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
)

func (s *MapStorage) CreateShortURL(ctx context.Context, userID, urlBase, fullURL string) (string, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if v, exist := s.createdURL[fullURL]; exist {
		return urlBase + v, errs.ErrDuplicate
	}

	ID := helpers.RandomString(6)
	shortURL := urlBase + ID

	s.createdURL[fullURL] = ID

	s.storage[ID] = fullURL

	s.userLink[userID][ID] = struct{}{}

	return shortURL, nil
}

func (s *MapStorage) BatchCreateShortURL(ctx context.Context, userID, urlBase string, data []dto.BatchRequest) ([]dto.BatchResponse, error) {
	s.m.Lock()
	defer s.m.Unlock()

	result := make([]dto.BatchResponse, 0, len(data))

	for _, v := range data {
		if _, exist := s.createdURL[v.OriginalURL]; exist {
			return nil, errs.ErrDuplicate
		}

		ID := helpers.RandomString(6)
		shortURL := urlBase + ID

		s.createdURL[v.OriginalURL] = ID

		result = append(result, dto.BatchResponse{CorrelationID: v.CorrelationID, ShortURL: shortURL})

		s.storage[ID] = v.OriginalURL

		s.userLink[userID][ID] = struct{}{}
	}

	return result, nil
}

func (s *MapStorage) GetLongURL(ctx context.Context, ID string) (string, error) {
	s.m.RLock()
	defer s.m.RUnlock()

	if val, exist := s.storage[ID]; !exist {
		return "", errs.ErrNotFound
	} else {
		return val, nil
	}
}

func (s *MapStorage) GetUserURLs(ctx context.Context, userID string) ([]dto.URLPair, error) {
	s.m.RLock()
	defer s.m.RUnlock()

	var result []dto.URLPair

	data, ok := s.userLink[userID]
	if !ok {
		return nil, errs.ErrNotFound
	}

	for k := range data {
		fullURL := s.storage[k]
		pair := dto.URLPair{
			ShortURL:    k,
			OriginalURL: fullURL,
		}

		result = append(result, pair)
	}

	return result, nil
}

func (s *MapStorage) DeleteURLs(ctx context.Context, userID string, urls []string) error {
	userURLs, ok := s.userLink[userID]
	if !ok {
		return errs.ErrNotFound
	}

	for _, url := range urls {
		if _, ok = userURLs[url]; !ok {
			return errs.ErrUserMismatch
		}

		fullURL := s.storage[url]

		delete(s.userLink[userID], url)
		delete(s.storage, fullURL)
		delete(s.createdURL, url)
	}

	return nil
}

func (s *MapStorage) LoadStorage(filepath string) error {
	s.m.Lock()
	defer s.m.Unlock()

	file, err := os.Open(filepath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		s.logger.Error("mapstorage:LoadStorage Error opening file", zap.Error(err))
		return errs.ErrInternalServerError
	}

	defer file.Close()

	var data []models.Urls
	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}

		s.logger.Error("mapstorage:LoadStorage Error opening file", zap.Error(err))
		return errs.ErrInternalServerError
	}

	for _, entry := range data {
		s.storage[entry.ShortURL] = entry.OriginalURL
		s.createdURL[entry.OriginalURL] = entry.ShortURL
	}

	return nil
}

func (s *MapStorage) OffloadStorage(ctx context.Context, filepath string) error {
	s.m.Lock()
	defer s.m.Unlock()

	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		s.logger.Error("mapstorage:OffloadStorage Error opening file", zap.Error(err))
		return errs.ErrInternalServerError
	}
	defer file.Close()

	var data []models.Urls
	for k, v := range s.storage {
		data = append(data, models.Urls{
			ShortURL:    k,
			OriginalURL: v,
		})
	}

	err = json.NewEncoder(file).Encode(&data)
	if err != nil {
		s.logger.Error("mapstorage:OffloadStorage Error encoding data", zap.Error(err))
		return errs.ErrInternalServerError
	}

	return nil
}

func (s *MapStorage) Ping(ctx context.Context) error {
	return nil
}
