package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	netUrl "net/url"
	"time"

	contextI "github.com/MukizuL/shortener/internal/context"
	"github.com/MukizuL/shortener/internal/dto"
	"github.com/MukizuL/shortener/internal/errs"
	"github.com/MukizuL/shortener/internal/helpers"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// CreateShortURL godoc
//
//	@Summary		Creates short url
//	@Description	If cookie with access token is not provided, creates a new token with new userID.
//	@Tags			default
//	@Accept			text/plain
//	@Produce		text/plain
//	@Param			Cookie	header		string		false	"Cookie with access token"
//	@Param			URL		body		string		true	"URL to shorten"
//	@Success		201		body		string		"Short url"
//	@Header			201		{string}	Set-cookie	"Access token"
//	@Failure		400		{string}	string		"Wrong URL schema"
//	@Failure		409		{string}	string		"URL already exists"
//	@Failure		422		{string}	string		"Not a URL"
//	@Failure		500		{string}	string		"Internal Server Error"
//	@Router			/ [post]
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

// GetFullURL godoc
//
//	@Summary	Redirects to original URL
//	@Tags		default
//	@Produce	text/html
//	@Param		Cookie	header	string	true	"Cookie with access token"
//	@Param		ID		query	string	true	"Short URL ID"
//	@Success	307
//	@Header		307	{string}	Location	"Original URL"
//	@Failure	400	{string}	string		"ID is not present"
//	@Failure	404	{string}	string		"URL not Found"
//	@Failure	410	{string}	string		"URL deleted"
//	@Failure	500	{string}	string		"Internal Server Error"
//	@Router		/:id [get]
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

// GetURLs godoc
//
//	@Summary	Returns array of user URLs
//	@Tags		json
//	@Produce	application/json
//	@Param		Cookie	header		string			true	"Cookie with access token"
//	@Success	200		{object}	[]dto.URLPair	"Array of URLs"
//	@Failure	500		{string}	string			"Internal Server Error"
//	@Router		/api/user/urls [get]
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

// DeleteURLs godoc
//
//	@Summary	Deletes user URLs
//	@Tags		json
//	@Accept		application/json
//	@Produce	application/json
//	@Param		Cookie	header		string		true	"Cookie with access token"
//	@Param		URLs	body		[]string	true	"URLs to delete"
//	@Success	202		{string}	string		"Accepted"
//	@Failure	401		{string}	string		"URL doesn't belong to user"
//	@Failure	500		{string}	string		"Internal Server Error"
//	@Router		/api/user/urls [delete]
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

// CreateShortURLJSON godoc
//
//	@Summary		Creates short URL
//	@Description	If cookie with access token is not provided, creates a new token with new userID.
//	@Tags			json
//	@Accept			application/json
//	@Produce		application/json
//	@Param			Cookie	header		string				false	"Cookie with access token"
//	@Param			URL		body		dto.Request			true	"URL to shorten"
//	@Success		201		body		dto.Response		"Short url"
//	@Header			201		{string}	Set-cookie			"Access token"
//	@Failure		400		{object}	dto.ErrorResponse	"Wrong URL schema"
//	@Failure		409		{object}	dto.ErrorResponse	"URL already exists"
//	@Failure		422		{object}	dto.ErrorResponse	"Not a URL"
//	@Failure		500		{object}	dto.ErrorResponse	"Internal Server Error"
//	@Router			/api/shorten [post]
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

// BatchCreateShortURLJSON godoc
//
//	@Summary		Creates a batch of short URLs
//	@Description	If cookie with access token is not provided, creates a new token with new userID.
//	@Tags			json
//	@Accept			application/json
//	@Produce		application/json
//	@Param			Cookie	header		string				false	"Cookie with access token"
//	@Param			URL		body		[]dto.BatchRequest	true	"URLs to shorten"
//	@Success		201		body		[]dto.BatchResponse	"Short urls"
//	@Header			201		{string}	Set-cookie			"Access token"
//	@Failure		400		{object}	dto.ErrorResponse	"Wrong URL schema"
//	@Failure		409		{object}	dto.ErrorResponse	"URL already exists"
//	@Failure		422		{object}	dto.ErrorResponse	"Not a URL"
//	@Failure		500		{object}	dto.ErrorResponse	"Internal Server Error"
//	@Router			/api/shorten/batch [post]
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
