package mw

import (
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	contextI "github.com/MukizuL/shortener/internal/context"
	"github.com/MukizuL/shortener/internal/errs"
	"github.com/MukizuL/shortener/internal/helpers"
	jwtService "github.com/MukizuL/shortener/internal/jwt"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type userIDKey string

var key userIDKey = "userID"

type MiddlewareService struct {
	jwtService jwtService.JWTServiceInterface
	logger     *zap.Logger
}

func NewMiddlewareService(jwtService jwtService.JWTServiceInterface, logger *zap.Logger) *MiddlewareService {
	return &MiddlewareService{
		jwtService: jwtService,
		logger:     logger,
	}
}

func Provide() fx.Option {
	return fx.Provide(NewMiddlewareService)
}

// LoggerMW logs request and response.
func (s *MiddlewareService) LoggerMW(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		mRW, ok := w.(universalRW)
		if !ok {
			panic("this is not a modded ResponseWriter")
		}

		h.ServeHTTP(w, r)

		duration := time.Since(start)

		s.logger.Info("Request", zap.String("uri", r.RequestURI), zap.String("method", r.Method), zap.Duration("time", duration))
		s.logger.Info("Response", zap.Int("status", mRW.GetStatusCode()), zap.Int("size", mRW.GetSize()))
	})
}

func (s *MiddlewareService) GzipCompress(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(&moddedResponseWriter{
				ResponseWriter: w,
				status:         0,
				size:           0,
			}, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}

		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")

		h.ServeHTTP(&gzipResponseWriter{
			moddedResponseWriter: moddedResponseWriter{
				ResponseWriter: w,
				status:         0,
				size:           0,
			},
			Writer: gz,
		}, r)
	})
}

// Authorization checks for Access-token in cookie. If it's present and valid, sets userID in context.
// If token is not present, creates a new one. If token is invalid, returns an error.
func (s *MiddlewareService) Authorization(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("Access-token")
		if err != nil && !errors.Is(err, http.ErrNoCookie) {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		var token, userID string
		if errors.Is(err, http.ErrNoCookie) {
			token, userID, err = s.jwtService.CreateToken()
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		} else {
			token, userID, err = s.jwtService.ValidateToken(cookie.Value)
			if err != nil {
				if errors.Is(err, errs.ErrNotAuthorized) {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}

				if errors.Is(err, errs.ErrUnexpectedSigningMethod) {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}

				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		r = r.Clone(context.WithValue(r.Context(), contextI.UserIDContextKey, userID))

		h.ServeHTTP(w, r)

		helpers.WriteCookie(w, token)
	})
}
