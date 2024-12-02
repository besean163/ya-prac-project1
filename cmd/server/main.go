package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
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
	"google.golang.org/grpc"

	_ "net/http/pprof"
	pb "ya-prac-project1/internal/services/proto"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	showBuildInfo()
	config := NewConfig()
	if err := run(config); err != nil {
		log.Fatal(err.Error())
	}
}

func run(config ServerConfig) error {
	err := logger.Set()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	runGracefulShutdown(cancel)
	RunProfiler(ctx, config.Profiler)

	store, err := getStorage(ctx, config, getSQLConnect(config))
	if err != nil {
		return err
	}

	metricService := services.NewMetricSaverService(store)

	RungRPCServer(ctx, config.GRPCEndpoint, metricService)

	h := handlers.New(metricService, getSQLConnect(config), config.HashKey, config.CryptoKey, config.TrustedSubnet)
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
	signal.Notify(s, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-s
		cancel()
	}()
}

func RunProfiler(ctx context.Context, port string) {
	if port == "" {
		return
	}

	ps := &http.Server{
		Addr: port,
	}

	go func() {
		err := ps.ListenAndServe()
		if err != nil {
			logger.Get().Info("Fail start prof server", zap.String("error", err.Error()))
		}
	}()

	go func() {
		<-ctx.Done()
		logger.Get().Info("Stop prof server")
		if err := ps.Shutdown(context.Background()); err != nil {
			logger.Get().Info("Fail shutdown prof server", zap.String("error", err.Error()))
		}
	}()
}

func showBuildInfo() {
	na := "N/A"
	if buildVersion == "" {
		buildVersion = na
	}

	if buildDate == "" {
		buildDate = na
	}

	if buildCommit == "" {
		buildCommit = na
	}

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}

func RungRPCServer(ctx context.Context, port string, ms pb.MetricSaverServiceServer) {
	server := grpc.NewServer()
	listen, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	pb.RegisterMetricSaverServiceServer(server, ms)

	go func() {
		server.Serve(listen)
	}()

	go func() {
		<-ctx.Done()
		server.GracefulStop()
	}()
}
