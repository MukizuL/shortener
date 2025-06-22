package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	contextI "github.com/MukizuL/shortener/internal/context"
	"github.com/MukizuL/shortener/internal/dto"
	"github.com/MukizuL/shortener/internal/errs"
	"github.com/MukizuL/shortener/internal/helpers"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"io"
	"net/http"
	netUrl "net/url"
	"time"
)

func (c *Controller) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	rawURL, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	url, err := netUrl.ParseRequestURI(string(rawURL))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	if url.Scheme != "http" && url.Scheme != "https" || url.Host == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(contextI.UserIDContextKey).(string)

	shortURL, err := c.storage.CreateShortURL(ctx, userID, fmt.Sprintf("http://%s/", r.Host), url.String())
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
		c.logger.Error("Error in handler CreateShortURL", zap.Error(err))
	}
}

func (c *Controller) GetFullURL(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	ID := chi.URLParam(r, "id")
	if ID == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	fullURL, err := c.storage.GetLongURL(ctx, ID)
	if err != nil {
		if errors.Is(err, errs.ErrURLNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		if errors.Is(err, errs.ErrGone) {
			http.Error(w, http.StatusText(http.StatusGone), http.StatusGone)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fullURL, http.StatusTemporaryRedirect)
}

func (c *Controller) GetURLs(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	userID := r.Context().Value(contextI.UserIDContextKey).(string)

	data, err := c.storage.GetUserURLs(ctx, userID)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, data)
}

func (c *Controller) DeleteURLs(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	userID := r.Context().Value(contextI.UserIDContextKey).(string)

	var urls []string

	err := json.NewDecoder(r.Body).Decode(&urls)
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, &dto.ErrorResponse{Err: http.StatusText(http.StatusInternalServerError)})
		return
	}

	err = c.storage.DeleteURLs(ctx, userID, urls)
	if err != nil {
		if errors.Is(err, errs.ErrUserMismatch) {
			helpers.WriteJSON(w, http.StatusUnauthorized, &dto.ErrorResponse{Err: http.StatusText(http.StatusUnauthorized)})
			return
		}

		helpers.WriteJSON(w, http.StatusInternalServerError, &dto.ErrorResponse{Err: http.StatusText(http.StatusInternalServerError)})
		return
	}

	helpers.WriteJSON(w, http.StatusAccepted, http.StatusText(http.StatusAccepted))
}

func (c *Controller) CreateShortURLJSON(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var req dto.Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, &dto.ErrorResponse{Err: http.StatusText(http.StatusInternalServerError)})
		return
	}

	url, err := netUrl.ParseRequestURI(req.FullURL)
	if err != nil {
		helpers.WriteJSON(w, http.StatusUnprocessableEntity, &dto.ErrorResponse{Err: http.StatusText(http.StatusUnprocessableEntity)})
		return
	}

	if url.Scheme != "http" && url.Scheme != "https" || url.Host == "" {
		helpers.WriteJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Err: http.StatusText(http.StatusBadRequest)})
		return
	}

	userID := r.Context().Value(contextI.UserIDContextKey).(string)

	shortURL, err := c.storage.CreateShortURL(ctx, userID, fmt.Sprintf("http://%s/", r.Host), url.String())
	if err != nil {
		if errors.Is(err, errs.ErrDuplicate) {
			helpers.WriteJSON(w, http.StatusConflict, &dto.Response{Result: shortURL})
			return
		}

		helpers.WriteJSON(w, http.StatusInternalServerError, &dto.ErrorResponse{Err: http.StatusText(http.StatusInternalServerError)})
		return
	}

	out := &dto.Response{Result: shortURL}

	helpers.WriteJSON(w, http.StatusCreated, out)
}

func (c *Controller) BatchCreateShortURLJSON(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var req []dto.BatchRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		helpers.WriteJSON(w, http.StatusInternalServerError, &dto.ErrorResponse{Err: http.StatusText(http.StatusInternalServerError)})
		return
	}

	for _, v := range req {
		url, err := netUrl.ParseRequestURI(v.OriginalURL)
		if err != nil {
			helpers.WriteJSON(w, http.StatusUnprocessableEntity, &dto.ErrorResponse{Err: http.StatusText(http.StatusUnprocessableEntity)})
			return
		}

		if url.Scheme != "http" && url.Scheme != "https" || url.Host == "" {
			helpers.WriteJSON(w, http.StatusBadRequest, &dto.ErrorResponse{Err: http.StatusText(http.StatusBadRequest)})
			return
		}
	}

	userID := r.Context().Value(contextI.UserIDContextKey).(string)

	response, err := c.storage.BatchCreateShortURL(ctx, userID, fmt.Sprintf("http://%s/", r.Host), req)
	if err != nil {
		if errors.Is(err, errs.ErrDuplicate) {
			helpers.WriteJSON(w, http.StatusConflict, &dto.ErrorResponse{Err: http.StatusText(http.StatusConflict)})
			return
		}

		helpers.WriteJSON(w, http.StatusInternalServerError, &dto.ErrorResponse{Err: http.StatusText(http.StatusInternalServerError)})
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, response)
}

func (c *Controller) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	err := c.storage.Ping(ctx)
	if err != nil {
		c.logger.Error("Error in Ping handler", zap.Error(err))
		helpers.WriteJSON(w, http.StatusInternalServerError, &dto.ErrorResponse{Err: http.StatusText(http.StatusInternalServerError)})
		return
	}

	out := &dto.Response{Result: "Pong"}

	helpers.WriteJSON(w, http.StatusOK, out)
}
