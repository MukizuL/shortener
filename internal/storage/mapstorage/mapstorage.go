package mapstorage

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/MukizuL/shortener/internal/dto"
	"github.com/MukizuL/shortener/internal/errs"
	"github.com/MukizuL/shortener/internal/helpers"
	"github.com/MukizuL/shortener/internal/models"
	"go.uber.org/zap"
)

func (s *MapStorage) CreateShortURL(ctx context.Context, userID, urlBase, fullURL string) (string, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if v, exist := s.ShortURLStorage[fullURL]; exist {
		return urlBase + v, errs.ErrDuplicate
	}

	ID := helpers.RandomString(6)
	shortURL := urlBase + ID

	s.ShortURLStorage[fullURL] = ID

	s.FullURLStorage[ID] = fullURL

	if _, ok := s.UserLinkStorage[userID]; !ok {
		s.UserLinkStorage[userID] = make(map[string]string)
	}

	s.UserLinkStorage[userID][ID] = fullURL

	return shortURL, nil
}

func (s *MapStorage) BatchCreateShortURL(ctx context.Context, userID, urlBase string, data []dto.BatchRequest) ([]dto.BatchResponse, error) {
	s.m.Lock()
	defer s.m.Unlock()

	result := make([]dto.BatchResponse, 0, len(data))

	for _, v := range data {
		if _, exist := s.ShortURLStorage[v.OriginalURL]; exist {
			return nil, errs.ErrDuplicate
		}

		ID := helpers.RandomString(6)
		shortURL := urlBase + ID

		s.ShortURLStorage[v.OriginalURL] = ID

		result = append(result, dto.BatchResponse{CorrelationID: v.CorrelationID, ShortURL: shortURL})

		s.FullURLStorage[ID] = v.OriginalURL

		if _, ok := s.UserLinkStorage[userID]; !ok {
			s.UserLinkStorage[userID] = make(map[string]string)
		}

		s.UserLinkStorage[userID][ID] = v.OriginalURL
	}

	return result, nil
}

func (s *MapStorage) GetLongURL(ctx context.Context, ID string) (string, error) {
	s.m.RLock()
	defer s.m.RUnlock()

	if val, exist := s.FullURLStorage[ID]; !exist {
		return "", errs.ErrURLNotFound
	} else {
		return val, nil
	}
}

func (s *MapStorage) GetUserURLs(ctx context.Context, userID string) ([]dto.URLPair, error) {
	s.m.RLock()
	defer s.m.RUnlock()

	var result []dto.URLPair

	data, ok := s.UserLinkStorage[userID]
	if !ok {
		return nil, errs.ErrURLNotFound
	}

	for k := range data {
		fullURL := s.FullURLStorage[k]
		pair := dto.URLPair{
			ShortURL:    k,
			OriginalURL: fullURL,
		}

		result = append(result, pair)
	}

	return result, nil
}

func (s *MapStorage) DeleteURLs(ctx context.Context, userID string, urls []string) error {
	userURLs, ok := s.UserLinkStorage[userID]
	if !ok {
		return errs.ErrURLNotFound
	}

	s.m.Lock()
	defer s.m.Unlock()

	for _, url := range urls {
		if _, ok = userURLs[url]; !ok {
			return errs.ErrUserMismatch
		}

		fullURL := s.FullURLStorage[url]

		delete(s.UserLinkStorage[userID], url)
		delete(s.FullURLStorage, fullURL)
		delete(s.ShortURLStorage, url)
	}

	return nil
}

func (s *MapStorage) GetStats(ctx context.Context) (int, int, error) {
	s.m.Lock()
	defer s.m.Unlock()

	return len(s.ShortURLStorage), len(s.UserLinkStorage), nil
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
		if _, ok := s.UserLinkStorage[entry.UserID]; !ok {
			s.UserLinkStorage[entry.UserID] = make(map[string]string)
		}

		s.FullURLStorage[entry.ShortURL] = entry.OriginalURL
		s.ShortURLStorage[entry.OriginalURL] = entry.ShortURL
		s.UserLinkStorage[entry.UserID][entry.ShortURL] = entry.OriginalURL
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
	for k, v := range s.UserLinkStorage {
		for kInner, vInner := range v {
			data = append(data, models.Urls{
				UserID:      k,
				ShortURL:    kInner,
				OriginalURL: vInner,
			})
		}
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
