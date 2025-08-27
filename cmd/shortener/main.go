// Package shortener starts server for shortening URLs.

package main

import (
	"fmt"
	"net/http"

	"github.com/MukizuL/shortener/internal/controller"
	"github.com/MukizuL/shortener/internal/interceptor"
	"github.com/MukizuL/shortener/internal/migration"
	"go.uber.org/fx/fxevent"
	"google.golang.org/grpc"

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

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	fx.New(
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		createApp(),
		fx.Invoke(func(*http.Server, *grpc.Server) {}),
	).Run()
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
		interceptor.Provide(),

		pgstorage.Provide(),
		mapstorage.Provide(),
		storage.Provide(),
		migration.Provide(),
	)
}
