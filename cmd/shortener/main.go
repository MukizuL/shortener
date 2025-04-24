package main

import (
	"github.com/MukizuL/shortener/internal/config"
	"github.com/MukizuL/shortener/internal/controller"
	mw "github.com/MukizuL/shortener/internal/middleware"
	"github.com/MukizuL/shortener/internal/router"
	"github.com/MukizuL/shortener/internal/server"
	"github.com/MukizuL/shortener/internal/storage"
	"github.com/MukizuL/shortener/internal/storage/mapstorage"
	"github.com/MukizuL/shortener/internal/storage/pgstorage"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	fx.New(createApp(), fx.Invoke(func(*http.Server) {})).Run()
}

func createApp() fx.Option {
	return fx.Options(
		config.Provide(),
		fx.Provide(zap.NewDevelopment),
		mw.Provide(),
		controller.Provide(),
		router.Provide(),
		server.Provide(),

		pgstorage.Provide(),
		mapstorage.Provide(),
		storage.Provide(),
	)
}
