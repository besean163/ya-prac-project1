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
	"ya-prac-project1/internal/inmem"
	"ya-prac-project1/internal/logger"

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

	store := inmem.NewStorage()
	storeService := inmem.NewService(&store, config.StoreFile)
	err = storeService.Restore(config.Restore)
	if err != nil {
		return err
	}

	if config.StoreInterval > 0 {
		go storeService.Run(config.StoreInterval)
	}

	h := handlers.New(store, config.BaseDNS)
	h.Mount()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		s := make(chan os.Signal, 1)
		signal.Notify(s, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-s
		cancel()
	}()

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
		storeService.Dump()

		if err = srv.Shutdown(context.Background()); err != nil {
			fmt.Printf("Start server on: %s\n", config.Endpoint)
			return err
		}

		return nil
	})

	// нашел этот кусок кода в интернете, понимаю его туманно
	// смысл в обработке входящих сигналов системы для отсрочки выхода из программы чтобы сделать итоговый дамп
	// stopped := make(chan struct{})
	// go func() {
	// 	sigint := make(chan os.Signal, 1)
	// 	signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	// 	<-sigint
	// 	ctx, cancel := context.WithCancel(context.Background())
	// 	defer cancel()
	// 	if err := srv.Shutdown(ctx); err != nil {
	// 		log.Printf("HTTP server shutdown err: %v", err)
	// 	}

	// 	storeService.Dump()
	// 	close(stopped)
	// }()

	// fmt.Printf("Start server on: %s\n", config.Endpoint)
	// if err = srv.ListenAndServe(); err != nil {
	// 	return err
	// }

	if err := g.Wait(); err != nil {
		fmt.Printf("exit reason: %s \n", err)
	}

	// <-stopped
	return nil
}
