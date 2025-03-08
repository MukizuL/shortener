package storage

import (
	"github.com/MukizuL/shortener/internal/errs"
	"github.com/MukizuL/shortener/internal/helpers"
	"sync"
)

type MapStorage struct {
	storage    map[string]string
	createdURL map[string]struct{}
	m          sync.RWMutex
}

func New() *MapStorage {
	return &MapStorage{
		storage:    make(map[string]string),
		createdURL: make(map[string]struct{}),
	}
}

func (r *MapStorage) CreateShortURL(fullURL string) (string, error) {
	r.m.Lock()
	defer r.m.Unlock()

	if _, exist := r.createdURL[string(fullURL)]; exist {
		return "", errs.ErrDuplicate
	}

	r.createdURL[string(fullURL)] = struct{}{}

	ID := helpers.RandomString(6)
	shortURL := "http://localhost:8080/" + ID

	r.storage[ID] = string(fullURL)

	return shortURL, nil
}

func (r *MapStorage) GetLongURL(ID string) (string, error) {
	r.m.RLock()
	defer r.m.RUnlock()

	if val, exist := r.storage[ID]; !exist {
		return "", errs.ErrNotFound
	} else {
		return val, nil
	}
}
