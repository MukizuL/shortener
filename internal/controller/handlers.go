package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

	cookie, err := r.Cookie("Access-token")
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var token, userID string
	if errors.Is(err, http.ErrNoCookie) {
		token, userID, err = c.jwtService.CreateToken()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	} else {
		token, userID, err = c.jwtService.ValidateToken(cookie.Value)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	shortURL, err := c.storage.CreateShortURL(ctx, userID, fmt.Sprintf("http://%s/", r.Host), url.String())
	if err != nil {
		if errors.Is(err, errs.ErrDuplicate) {
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	helpers.WriteCookie(w, token)

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
		if errors.Is(err, errs.ErrNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
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

	cookie, err := r.Cookie("Access-token")
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	_, userID, err := c.jwtService.ValidateToken(cookie.Value)
	if err != nil {
		c.logger.Error("GetURLs Failed to validate token", zap.Error(err))

		if errors.Is(err, errs.ErrNotAuthorized) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		if errors.Is(err, errs.ErrUnexpectedSigningMethod) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := c.storage.GetUserURLs(ctx, userID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	helpers.WriteJSON(w, http.StatusOK, data)
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

	cookie, err := r.Cookie("Access-token")
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var token, userID string
	if errors.Is(err, http.ErrNoCookie) {
		token, userID, err = c.jwtService.CreateToken()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	} else {
		token, userID, err = c.jwtService.ValidateToken(cookie.Value)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

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

	helpers.WriteCookie(w, token)

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

	cookie, err := r.Cookie("Access-token")
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	var token, userID string
	if errors.Is(err, http.ErrNoCookie) {
		token, userID, err = c.jwtService.CreateToken()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	} else {
		token, userID, err = c.jwtService.ValidateToken(cookie.Value)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	response, err := c.storage.BatchCreateShortURL(ctx, userID, fmt.Sprintf("http://%s/", r.Host), req)
	if err != nil {
		if errors.Is(err, errs.ErrDuplicate) {
			helpers.WriteJSON(w, http.StatusConflict, &dto.ErrorResponse{Err: http.StatusText(http.StatusConflict)})
			return
		}

		helpers.WriteJSON(w, http.StatusInternalServerError, &dto.ErrorResponse{Err: http.StatusText(http.StatusInternalServerError)})
		return
	}

	helpers.WriteCookie(w, token)

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
