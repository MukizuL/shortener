package app

import (
	"github.com/MukizuL/shortener/internal/helpers"
	"io"
	"net/http"
)

func (app *application) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	url, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// TODO: validate URL

	app.m.Lock()

	if _, exist := app.createdURL[string(url)]; exist {
		http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		return
	}

	app.createdURL[string(url)] = struct{}{}

	ID := helpers.RandomString(app.seededRand, 6)
	shortURL := "http://localhost:8080/" + ID

	app.storage[ID] = string(url)

	app.m.Unlock()

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (app *application) GetFullURL(w http.ResponseWriter, r *http.Request) {
	ID := r.PathValue("id")
	if ID == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if val, exist := app.storage[ID]; !exist {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(val))
	}
}
