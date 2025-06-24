package main

import (
	"github.com/MukizuL/shortener/internal/controller"
	"net/http"

	"github.com/MukizuL/shortener/internal/config"
	jwtService "github.com/MukizuL/shortener/internal/jwt"
	mw "github.com/MukizuL/shortener/internal/middleware"
	"github.com/MukizuL/shortener/internal/router"
	"github.com/MukizuL/shortener/internal/server"
	"github.com/MukizuL/shortener/internal/storage"
	"github.com/MukizuL/shortener/internal/storage/mapstorage"
	"github.com/MukizuL/shortener/internal/storage/pgstorage"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

//	@title			Shortener API
//	@version		1.0
//	@description	This is a url shortening server.

//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							cookie
//	@name						Access-token

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
		jwtService.Provide(),

		pgstorage.Provide(),
		mapstorage.Provide(),
		storage.Provide(),
	)
}
