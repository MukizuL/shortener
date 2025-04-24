package main

import (
	"github.com/MukizuL/shortener/internal/config"
	"go.uber.org/fx"
)

func main() {
	//err := app.Run()
	//if err != nil {
	//	log.Fatal(err)
	//}
	fx.New(createApp()).Run()
}

func createApp() fx.Option {
	return fx.Options(
		config.Provide(),
	)
}
