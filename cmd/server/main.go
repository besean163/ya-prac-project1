package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"ya-prac-project1/internal/handlers"
	"ya-prac-project1/internal/logger"
	"ya-prac-project1/internal/services"
	"ya-prac-project1/internal/storage/databasestorage"
	"ya-prac-project1/internal/storage/filestorage"
	"ya-prac-project1/internal/storage/inmemstorage"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	config := NewConfig()
	if err := run(config); err != nil {
		log.Fatal(err.Error())
	}
	os.Exit(0)
}

func run(config ServerConfig) error {
	err := logger.Set()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	runGracefulShutdown(cancel)

	store, err := getStorage(ctx, config, getSQLConnect(config))
	if err != nil {
		return err
	}

	metricService := services.NewMetricSaverService(store)

	h := handlers.New(metricService, getSQLConnect(config), config.HashKey)
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
		fmt.Printf("Stop server on: %s\n", config.Endpoint)
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

func getStorage(ctx context.Context, config ServerConfig, db *sql.DB) (services.SaveStorage, error) {
	if config.BaseDNS != "" {
		return databasestorage.NewStorage(db)
	} else if config.StoreFile != "" {
		return filestorage.NewStorage(ctx, config.StoreFile, config.Restore, int64(config.StoreInterval))
	}

	store := inmemstorage.NewStorage()
	return store, nil
}

func getSQLConnect(config ServerConfig) *sql.DB {
	if config.BaseDNS == "" {
		return nil
	}
	db, err := sql.Open("pgx", config.BaseDNS)
	if err != nil {
		logger.Get().Info("can't connect base.", zap.String("Error", err.Error()))
		return nil
	}

	return db
}

func runGracefulShutdown(cancel context.CancelFunc) {
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-s
		cancel()
	}()
}
