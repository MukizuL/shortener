package router

import (
	"expvar"
	"net/http"
	"net/http/pprof"

	"github.com/MukizuL/shortener/internal/config"
	"github.com/MukizuL/shortener/internal/controller"
	mw "github.com/MukizuL/shortener/internal/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
)

// NewRouter initializes new chi.Mux with routes.
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

	r.Mount("/debug", Profiler())

	return r
}

// Profiler creates http.Handler with pprof's routes.
func Profiler() http.Handler {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.RequestURI+"/pprof/", http.StatusMovedPermanently)
	})
	r.HandleFunc("/pprof", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, r.RequestURI+"/", http.StatusMovedPermanently)
	})

	r.HandleFunc("/pprof/*", pprof.Index)
	r.HandleFunc("/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/pprof/profile", pprof.Profile)
	r.HandleFunc("/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/pprof/trace", pprof.Trace)
	r.Handle("/vars", expvar.Handler())

	r.Handle("/pprof/goroutine", pprof.Handler("goroutine"))
	r.Handle("/pprof/threadcreate", pprof.Handler("threadcreate"))
	r.Handle("/pprof/mutex", pprof.Handler("mutex"))
	r.Handle("/pprof/heap", pprof.Handler("heap"))
	r.Handle("/pprof/block", pprof.Handler("block"))
	r.Handle("/pprof/allocs", pprof.Handler("allocs"))

	return r
}

func Provide() fx.Option {
	return fx.Provide(NewRouter)
}
