package handlers

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	hpprof "net/http/pprof"
	"yaprakticum-go-track2/internal/handlers/middleware"
	"yaprakticum-go-track2/internal/prom"
)

// Router of the Server
func Router(h Handlers, pm *prom.CustomPromMetrics) chi.Router {

	r := chi.NewRouter()
	r.Use(middleware.WithRSA(h.cfg), middleware.GzipHandler, middleware.WithLogging, middleware.Prom(pm))
	r.Route("/", func(r chi.Router) {
		r.Get("/", h.GetAllMetricsHandler)
		r.Route("/updates", func(r chi.Router) {
			r.Post("/", h.MultiMetricsUpdateHandlerREST)
		})
		r.Route("/update", func(r chi.Router) {
			r.Post("/", h.MetricsUpdateHandlerREST)
			r.Post("/{type}", func(res http.ResponseWriter, req *http.Request) {
				http.Error(res, "Not enough args (No name)", http.StatusNotFound)
			})
			r.Post("/{type}/{name}", func(res http.ResponseWriter, req *http.Request) {
				http.Error(res, "Not enough args (No value)", http.StatusBadRequest)
			})
			r.Post("/{type}/{name}/{value}", h.MetricUpdateHandler)

		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/{type}/{name}", h.GetMetricHandler)
			r.Post("/", h.GetMetricHandlerREST)
		})
		r.Route("/ping", func(r chi.Router) {
			r.Get("/", h.PingHandler)
		})
		r.Route("/debug", func(r chi.Router) {
			r.Route("/pprof", func(r chi.Router) {
				r.Get("/", hpprof.Index)
				r.Get("/heap", hpprof.Index)
				r.Get("/profile", hpprof.Profile)
			})
		})
	})
	return r

}
