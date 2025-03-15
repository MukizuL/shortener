package app

import (
	"github.com/go-chi/chi/v5"
)

func NewRouter(baseURL string, app *Application) *chi.Mux {
	r := chi.NewRouter()

	r.Post(baseURL+"/", app.CreateShortURL)
	r.Get(baseURL+"/{id}", app.GetFullURL)

	return r
}
