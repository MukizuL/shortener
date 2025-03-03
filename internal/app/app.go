package app

import (
	"github.com/MukizuL/shortener/internal/config"
	"github.com/MukizuL/shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"log"
	"log/slog"
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

	params := config.GetParams()

	r := chi.NewRouter()

	r.Post(params.Base+"/", app.CreateShortURL)
	r.Get(params.Base+"/{id}", app.GetFullURL)

	slog.Info("Server started on " + params.Addr)
	err := http.ListenAndServe(params.Addr, r)
	if err != nil {
		log.Fatal(err)
	}
}
