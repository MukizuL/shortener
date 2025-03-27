package storage

import (
	"encoding/json"
	"errors"
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

func New(filepath string) *MapStorage {
	storage := &MapStorage{
		storage:    make(map[string]string),
		createdURL: make(map[string]struct{}),
	}

	err := storage.LoadStorage(filepath)
	if err != nil {
		panic(err)
	}

	return storage
}

func (r *MapStorage) CreateShortURL(fullURL string) (string, error) {
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

func (r *MapStorage) GetLongURL(ID string) (string, error) {
	r.m.RLock()
	defer r.m.RUnlock()

	if val, exist := r.storage[ID]; !exist {
		return "", errs.ErrNotFound
	} else {
		return val, nil
	}
}

func (r *MapStorage) LoadStorage(filepath string) error {
	r.m.Lock()
	defer r.m.Unlock()

	file, err := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		panic("Error opening file" + err.Error())
	}
	defer file.Close()

	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	var data []models.DataObject
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

func (r *MapStorage) OffloadStorage(filepath string) error {
	r.m.Lock()
	defer r.m.Unlock()

	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic("Error opening file" + err.Error())
	}
	defer file.Close()

	var data []models.DataObject
	for k, v := range r.storage {
		data = append(data, models.DataObject{
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
