package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"ya-prac-project1/internal/logger"
	"ya-prac-project1/internal/services"
	"ya-prac-project1/internal/storage/inmemstorage"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	_ "net/http/pprof"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

const (
	retryAttempts    = 3
	waitSec          = 1
	waitSecIncrement = 2
)

type HTTPClientInterface interface {
	Do(*http.Request) (*http.Response, error)
}

func main() {
	showBuildInfo()
	err := logger.Set()
	if err != nil {
		log.Fatalf("logger error: %s", err.Error())
	}

	c := NewConfig()
	ctx, cancel := context.WithCancel(context.Background())
	runGracefulShutdown(cancel)
	RunProfiler(ctx, c.Profiler)

	errGroup, gCtx := errgroup.WithContext(ctx)
	requestCh := make(chan *http.Request, 1)
	requestDone := make(chan struct{}, 1)

	storage := inmemstorage.NewStorage()
	service := services.NewRuntimeService(storage)

	service.Run(gCtx, c.PoolInterval)

	for count := 0; count < c.RateLimit; count++ {
		errGroup.Go(func() error {
			client := &http.Client{}
			worker(gCtx, requestCh, requestDone, client)
			return nil
		})
	}

	errGroup.Go(func() error {
		// для первого запуска
		requestDone <- struct{}{}
		for {
			select {
			case <-gCtx.Done():
				log.Printf("send request stopped")
				return nil
			case <-requestDone:
				ticker := time.NewTicker(time.Duration(c.ReportInterval) * time.Second)
				select {
				case <-gCtx.Done():
					log.Printf("send request stopped")
					return nil
				case <-ticker.C:
					service.RunSendRequest(requestCh, c.Endpoint, c.HashKey, c.CryptoKey)
					service.RunSendgRPCRequest(c.GRPCEndpoint)
				}
			}
		}
	})

	if err := errGroup.Wait(); err != nil {
		logger.Get().Info("agent error", zap.String("error", err.Error()))
	}
	close(requestCh)
	log.Printf("full stopped")
}

func needRetry(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "connection refused")
}

func runGracefulShutdown(cancel context.CancelFunc) {
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-s
		logger.Get().Info("start shutdown")
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
