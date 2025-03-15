package app

import (
	"context"
	"github.com/MukizuL/shortener/internal/config"
	"github.com/MukizuL/shortener/internal/storage"
	"log"
	"os/signal"
	"syscall"
)

type repo interface {
	CreateShortURL(fullURL string) (string, error)
	GetLongURL(ID string) (string, error)
}
type Application struct {
	storage repo
}

func NewApplication(storage repo) *Application {
	return &Application{
		storage: storage,
	}
}

func Run() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app := NewApplication(storage.New())

	params := config.GetParams()

	r := NewRouter(params.Base, app)

	err := runServer(ctx, params.Addr, r)
	if err != nil {
		log.Fatal(err)
	}
}
