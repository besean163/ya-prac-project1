package main

import (
	"fmt"
	"log"
	"net/http"
	"ya-prac-project1/internal/handlers"
	"ya-prac-project1/internal/inmem"
)

func main() {
	config := NewConfig()
	if err := run(config); err != nil {
		log.Fatalf(err.Error())
	}
}

func run(config ServerConfig) error {
	store := inmem.NewStorage()
	h := handlers.New(store)

	router := h.GetHandler()

	fmt.Printf("Start server on: %s\n", config.Endpoint)
	return http.ListenAndServe(config.Endpoint, router)
}
