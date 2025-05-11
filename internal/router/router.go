package router

import (
	"github.com/MukizuL/shortener/internal/config"
	"github.com/MukizuL/shortener/internal/controller"
	mw "github.com/MukizuL/shortener/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
)

func NewRouter(cfg *config.Config, mw *mw.MiddlewareService, c *controller.Controller) *chi.Mux {
	r := chi.NewRouter()
	r.Use(mw.GzipCompress)
	r.Use(mw.LoggerMW)

	r.With(mw.Authorization).Post(cfg.Base+"/", c.CreateShortURL)
	r.Get(cfg.Base+"/{id}", c.GetFullURL)
	r.Get(cfg.Base+"/ping", c.Ping)

	r.With(mw.Authorization).Get(cfg.Base+"/api/user/urls", c.GetURLs)
	r.With(mw.Authorization).Delete(cfg.Base+"/api/user/urls", c.DeleteURLs)
	r.With(mw.Authorization).Post(cfg.Base+"/api/shorten", c.CreateShortURLJSON)
	r.With(mw.Authorization).Post(cfg.Base+"/api/shorten/batch", c.BatchCreateShortURLJSON)

	return r
}

func Provide() fx.Option {
	return fx.Provide(NewRouter)
}
