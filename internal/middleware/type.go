package mw

import (
	"io"
	"net/http"
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
