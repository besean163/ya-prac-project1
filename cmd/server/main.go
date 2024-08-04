package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"ya-prac-project1/internal/handlers"
	"ya-prac-project1/internal/logger"
	"ya-prac-project1/internal/storage/filestorage"
	"ya-prac-project1/internal/storage/inmemstorage"

	"golang.org/x/sync/errgroup"
)

func main() {
	config := NewConfig()
	if err := run(config); err != nil {
		log.Fatalf(err.Error())
	}
	os.Exit(0)
}

func run(config ServerConfig) error {
	err := logger.Set()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		s := make(chan os.Signal, 1)
		signal.Notify(s, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-s
		cancel()
	}()

	store, err := getStorage(ctx, config)
	if err != nil {
		return err
	}

	// store := inmem.NewStorage()
	// storeService := inmem.NewService(&store, config.StoreFile)
	// err = storeService.Restore(config.Restore)
	// if err != nil {
	// 	return err
	// }

	// if config.StoreInterval > 0 {
	// 	go storeService.Run(config.StoreInterval)
	// }

	h := handlers.New(store, config.BaseDNS)
	h.Mount()

	srv := &http.Server{
		Addr:    config.Endpoint,
		Handler: h,
	}

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		fmt.Printf("Start server on: %s\n", config.Endpoint)
		if err = srv.ListenAndServe(); err != nil {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-gCtx.Done()
		// делаем итоговый дамп
		// storeService.Dump()

		if err = srv.Shutdown(context.Background()); err != nil {
			fmt.Printf("Start server on: %s\n", config.Endpoint)
			return err
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		fmt.Printf("exit reason: %s \n", err)
	}

	return nil
}

func getStorage(ctx context.Context, config ServerConfig) (handlers.Storage, error) {

	if config.StoreFile != "" {
		return filestorage.NewStorage(ctx, config.StoreFile, config.Restore, int64(config.StoreInterval))
	}

	store := inmemstorage.NewStorage()
	return store, nil

}
