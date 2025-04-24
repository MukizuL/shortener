package mw

import (
	"compress/gzip"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
	"time"
)

type MiddlewareService struct {
	logger *zap.Logger
}

func NewMiddlewareService(logger *zap.Logger) *MiddlewareService {
	return &MiddlewareService{
		logger: logger,
	}
}

func Provide() fx.Option {
	return fx.Provide(NewMiddlewareService)
}

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
