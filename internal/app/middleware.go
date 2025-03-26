package app

import (
	"compress/gzip"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
	"time"
)

type universalRW interface {
	GetStatusCode() int
	GetSize() int
}
type moddedResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (r *moddedResponseWriter) GetStatusCode() int {
	return r.status
}

func (r *moddedResponseWriter) GetSize() int {
	return r.size
}

func (r *moddedResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.size += size
	return size, err
}

type gzipResponseWriter struct {
	moddedResponseWriter
	Writer io.Writer
}

func (r *gzipResponseWriter) Write(b []byte) (int, error) {
	size, err := r.Writer.Write(b)
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

		mRW, ok := w.(universalRW)
		if !ok {
			panic("this is not a modded ResponseWriter")
		}

		h.ServeHTTP(w, r)

		duration := time.Since(start)

		app.logger.Info("Request", zap.String("uri", r.RequestURI), zap.String("method", r.Method), zap.Duration("time", duration))
		app.logger.Info("Response", zap.Int("status", mRW.GetStatusCode()), zap.Int("size", mRW.GetSize()))
	})
}

func (app *Application) gzipCompress(h http.Handler) http.Handler {
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
