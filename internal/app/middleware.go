package app

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

type moddedResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (r *moddedResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.size += size
	return size, err
}

func (r *moddedResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.status = statusCode
}

func (app *Application) loggerMW(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		mRW := &moddedResponseWriter{
			ResponseWriter: w,
			status:         0,
			size:           0,
		}

		h.ServeHTTP(mRW, r)

		duration := time.Since(start)

		app.logger.Info("Request", zap.String("uri", r.RequestURI), zap.String("method", r.Method), zap.Duration("time", duration))
		app.logger.Info("Response", zap.Int("status", mRW.status), zap.Int("size", mRW.size))
	})
}
