package main

import (
	"fmt"
	"log"
	"net/http"
	"ya-prac-project1/internal/handlers"
	"ya-prac-project1/internal/inmem"
	"ya-prac-project1/internal/logger"
)

func main() {
	config := NewConfig()
	if err := run(config); err != nil {
		log.Fatalf(err.Error())
	}
}

func run(config ServerConfig) error {
	err := logger.Set()
	if err != nil {
		return err
	}
	store := inmem.NewStorage()
	h := handlers.New(store)
	h.Mount()

	fmt.Printf("Start server on: %s\n", config.Endpoint)
	return http.ListenAndServe(config.Endpoint, h)
}
