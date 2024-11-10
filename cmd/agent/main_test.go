package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"
	"ya-prac-project1/internal/logger"

	"github.com/stretchr/testify/assert"
)

func TestRunGracefulShutdown(t *testing.T) {
	_ = logger.Set()

	go func() {
		timer := time.NewTimer(time.Millisecond * 500)
		<-timer.C
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	ctx, cancel := context.WithCancel(context.Background())
	runGracefulShutdown(cancel)

	assert.IsType(t, struct{}{}, <-ctx.Done())
}

func TestNewConfig(t *testing.T) {
	os.Setenv("ADDRESS", ":8081")
	os.Setenv("REPORT_INTERVAL", "1")
	os.Setenv("POLL_INTERVAL", "2")
	os.Setenv("KEY", "test_key")
	os.Setenv("RATE_LIMIT", "3")

	c := NewConfig()
	assert.Equal(t, ":8081", c.Endpoint)
	assert.Equal(t, 1, c.ReportInterval)
	assert.Equal(t, 2, c.PoolInterval)
	assert.Equal(t, "test_key", c.HashKey)
	assert.Equal(t, 3, c.RateLimit)
}

func TestWorker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := ":8181"
	s := http.Server{
		Addr: port,
		Handler: http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		),
	}
	go s.ListenAndServe()
	defer s.Shutdown(ctx)

	rCh := make(chan *http.Request, 1)
	rDoneCh := make(chan struct{}, 1)

	go worker(ctx, rCh, rDoneCh)

	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost%s", port), nil)
	if err != nil {
		panic(err)
	}
	rCh <- r
	<-rDoneCh
}

func TestNeedRetry(t *testing.T) {
	var err error
	assert.False(t, needRetry(err))

	err = errors.New("some error")
	assert.False(t, needRetry(err))

	err = errors.New("connection refused")
	assert.True(t, needRetry(err))
}

func TestShowBuildInfo(t *testing.T) {
	showBuildInfo()
}
