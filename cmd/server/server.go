package main

import "github.com/go-chi/chi/v5"

type Server struct {
	Router *chi.Mux
}

func CreateServer() *Server {
	s := &Server{}
	s.Router = chi.NewRouter()
	return s
}

func (s *Server) MountHandlers(store MetricStorage) {
	s.Router.Route("/", func(r chi.Router) {
		r.Get("/", GetMetrics(store))
		r.Get("/value/{metric_type}/{metric_name}", GetMetrics(store))
		r.Post("/update/{metric_type}/{metric_name}/{metric_value}", UpdateMetrics(store))
	})

}
