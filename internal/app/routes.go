package app

import (
	"github.com/go-chi/chi/v5"
)

func NewRouter(baseURL string, app *Application) *chi.Mux {
	r := chi.NewRouter()
	r.Use(app.gzipCompress)
	r.Use(app.loggerMW)

	r.Post(baseURL+"/", app.CreateShortURL)
	r.Get(baseURL+"/{id}", app.GetFullURL)
	r.Get(baseURL+"/ping", app.Ping)

	r.Post(baseURL+"/api/shorten", app.CreateShortURLJSON)
	r.Post(baseURL+"/api/shorten/batch", app.BatchCreateShortURLJSON)
	return r
}
