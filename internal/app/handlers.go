package app

import (
	"encoding/json"
	"errors"
	"github.com/MukizuL/shortener/internal/dto"
	"github.com/MukizuL/shortener/internal/errs"
	"github.com/MukizuL/shortener/internal/helpers"
	"github.com/go-chi/chi/v5"
	"io"
	"log/slog"
	"net/http"
	url2 "net/url"
)

func (app *Application) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	rawURL, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	url, err := url2.ParseRequestURI(string(rawURL))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	if url.Scheme != "http" && url.Scheme != "https" || url.Host == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	shortURL, err := app.storage.CreateShortURL(url.String())
	if err != nil {
		if errors.Is(err, errs.ErrDuplicate) {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write([]byte(shortURL))
	if err != nil {
		slog.Error("Error in handler", "func", "CreateShortURL", "err", err.Error())
	}
}

func (app *Application) GetFullURL(w http.ResponseWriter, r *http.Request) {
	ID := chi.URLParam(r, "id")
	if ID == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	fullURL, err := app.storage.GetLongURL(ID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fullURL, http.StatusTemporaryRedirect)
}

func (app *Application) CreateShortURLJSON(w http.ResponseWriter, r *http.Request) {
	var in dto.Request
	err := json.NewDecoder(r.Body).Decode(&in)
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, &dto.ErrorResponse{Err: http.StatusText(http.StatusInternalServerError)})
		return
	}

	url, err := url2.ParseRequestURI(in.FullURL)
	if err != nil {
		helpers.WriteJSON(w, http.StatusUnprocessableEntity, &dto.ErrorResponse{Err: http.StatusText(http.StatusUnprocessableEntity)})
		return
	}

	if url.Scheme != "http" && url.Scheme != "https" || url.Host == "" {
		helpers.WriteJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Err: http.StatusText(http.StatusBadRequest)})
		return
	}

	shortURL, err := app.storage.CreateShortURL(url.String())
	if err != nil {
		if errors.Is(err, errs.ErrDuplicate) {
			helpers.WriteJSON(w, http.StatusConflict, &dto.ErrorResponse{Err: http.StatusText(http.StatusConflict)})
			return
		}

		helpers.WriteJSON(w, http.StatusInternalServerError, &dto.ErrorResponse{Err: http.StatusText(http.StatusInternalServerError)})
		return
	}

	out := &dto.Response{Result: shortURL}

	helpers.WriteJSON(w, http.StatusCreated, out)
}
