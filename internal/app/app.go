package app

import (
	"github.com/MukizuL/shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"sync"
)

type repo interface {
	Create(fullURL string) (string, error)
	Get(ID string) (string, error)
}
type application struct {
	storage repo
	m       sync.Mutex
}

func Run() {
	app := &application{
		storage: storage.New(),
	}

	r := chi.NewRouter()

	r.Post("/", app.CreateShortURL)
	r.Get("/{id}", app.GetFullURL)

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}
