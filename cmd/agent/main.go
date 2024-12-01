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
			worker(gCtx, requestCh, requestDone)
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

func worker(ctx context.Context, requestCh chan *http.Request, requestDone chan struct{}) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("worker stopped")
			return
		case req := <-requestCh:
			client := http.Client{}
			var err error

			response, err := client.Do(req)
			if response != nil && response.Body != nil {
				response.Body.Close()
			}
			attempt := 1
			secDelta := waitSec
			for response == nil && err != nil && needRetry(err) && attempt <= retryAttempts {
				log.Printf("get error, need try again, wait %d sec", secDelta)
				stop := false
				ticker := time.NewTicker(time.Duration(secDelta) * time.Second)
				select {
				case <-ticker.C:
					response, err = client.Do(req)
					if response != nil {
						response.Body.Close()
					}
				case <-ctx.Done():
					stop = true
				}

				if stop {
					break
				}

				attempt++
				secDelta += waitSecIncrement
			}

			if response != nil {
				if response.StatusCode != http.StatusOK {
					log.Println("Error write metrics")
					log.Println("Path:", req.URL.Path)
					log.Println("Code:", response.StatusCode)
				}
			}

			if err != nil {
				log.Printf("call error. Error: %s\n", err)
			}

			requestDone <- struct{}{}
		}

	}

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
