package main

import (
	"fmt"
	"net/http"
	"ya-prac-project1/internal/handlers"
	"ya-prac-project1/internal/inmem"

	"github.com/go-chi/chi/v5"
)

func main() {
	parseFlags()
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	store := inmem.NewStorage()
	h := handlers.New(store)

	router := chi.NewRouter()
	router.Route("/", func(r chi.Router) {
		r.Get("/", h.GetMetrics)
		r.Get("/value/{metric_type}/{metric_name}", h.GetMetrics)
		r.Post("/update/{metric_type}/{metric_name}/{metric_value}", h.UpdateMetrics)
	})

	fmt.Printf("Start server on: %s\n", serverEndpointFlag)
	return http.ListenAndServe(serverEndpointFlag, router)
}
