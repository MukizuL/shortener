package mw

import (
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/MukizuL/shortener/internal/config"
	contextI "github.com/MukizuL/shortener/internal/context"
	"github.com/MukizuL/shortener/internal/errs"
	"github.com/MukizuL/shortener/internal/helpers"
	jwtService "github.com/MukizuL/shortener/internal/jwt"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type MiddlewareService struct {
	jwtService jwtService.JWTServiceInterface
	cfg        *config.Config
	logger     *zap.Logger
}

func newMiddlewareService(jwtService jwtService.JWTServiceInterface, cfg *config.Config, logger *zap.Logger) *MiddlewareService {
	return &MiddlewareService{
		jwtService: jwtService,
		cfg:        cfg,
		logger:     logger,
	}
}

func Provide() fx.Option {
	return fx.Provide(newMiddlewareService)
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

		gzW, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}

		defer gzW.Close()

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzR, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer gzR.Close()

			r.Body = gzR

			r.Body.Close()
		}

		w.Header().Set("Content-Encoding", "gzip")

		h.ServeHTTP(&gzipResponseWriter{
			moddedResponseWriter: moddedResponseWriter{
				ResponseWriter: w,
				status:         0,
				size:           0,
			},
			Writer: gzW,
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
			token, userID, err = s.jwtService.CreateOrValidateToken("")
		} else {
			token, userID, err = s.jwtService.CreateOrValidateToken(cookie.Value)
		}
		if err != nil {
			switch {
			case errors.Is(err, errs.ErrNotAuthorized), errors.Is(err, errs.ErrUnexpectedSigningMethod):
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		r = r.Clone(context.WithValue(r.Context(), contextI.UserIDContextKey, userID))

		helpers.WriteCookie(w, token)

		h.ServeHTTP(w, r)
	})
}

func (s *MiddlewareService) IsTrustedCIDR(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.TrustedCIDR == "" {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		realIP := r.Header.Get("X-Real-IP")
		if realIP == "" {
			ipPort := strings.Split(r.RemoteAddr, ":")
			realIP = ipPort[0]
		}

		IP := net.ParseIP(realIP)

		_, subnet, err := net.ParseCIDR(s.cfg.TrustedCIDR)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if subnet.Contains(IP) {
			h.ServeHTTP(w, r)
		} else {
			s.logger.Warn("IP is not trusted", zap.String("ip", IP.String()))
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
	})
}
