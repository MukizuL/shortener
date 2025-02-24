package storage

import (
	"github.com/MukizuL/shortener/internal/errs"
	"github.com/MukizuL/shortener/internal/helpers"
	"sync"
)

type MapStorage struct {
	storage    map[string]string
	createdURL map[string]struct{}
	m          sync.Mutex
}

func New() *MapStorage {
	return &MapStorage{
		storage:    make(map[string]string),
		createdURL: make(map[string]struct{}),
	}
}

func (r *MapStorage) Create(fullURL string) (string, error) {
	r.m.Lock()

	if _, exist := r.createdURL[string(fullURL)]; exist {
		return "", errs.ErrDuplicate
	}

	r.createdURL[string(fullURL)] = struct{}{}

	ID := helpers.RandomString(6)
	shortURL := "http://localhost:8080/" + ID

	r.storage[ID] = string(fullURL)

	r.m.Unlock()

	return shortURL, nil
}

func (r *MapStorage) Get(ID string) (string, error) {
	if val, exist := r.storage[ID]; !exist {
		return "", errs.ErrNotFound
	} else {
		return val, nil
	}
}
