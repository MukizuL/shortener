package app

import (
	"github.com/MukizuL/shortener/internal/storage"
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

	mux := http.NewServeMux()

	mux.HandleFunc("POST /", app.CreateShortURL)
	mux.HandleFunc("GET /{id}", app.GetFullURL)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
