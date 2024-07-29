package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"ya-prac-project1/internal/handlers"
	"ya-prac-project1/internal/inmem"
	"ya-prac-project1/internal/logger"
)

func main() {
	config := NewConfig()
	if err := run(config); err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Println("here")
}

func run(config ServerConfig) error {
	err := logger.Set()
	if err != nil {
		return err
	}

	store := inmem.NewStorage()
	storeService := inmem.NewService(&store, config.StoreFile)
	err = storeService.Restore(config.Restore)
	if err != nil {
		return err
	}

	if config.StoreInterval > 0 {
		go storeService.Run(config.StoreInterval)
	}

	h := handlers.New(store)
	h.Mount()

	srv := &http.Server{
		Addr:    config.Endpoint,
		Handler: h,
	}

	stopped := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-sigint
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("HTTP server shutdown err: %v", err)
		}

		storeService.Dump()
		close(stopped)
	}()

	fmt.Printf("Start server on: %s\n", config.Endpoint)
	return srv.ListenAndServe()
}
