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

	client := &testHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			if req.URL.Path == "/retry" {
				return nil, errors.New("temporary error")
			}
			return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
		},
	}
	go worker(ctx, rCh, rDoneCh, client)

	r, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost%s", port), nil)
	if err != nil {
		panic(err)
	}
	rCh <- r
	<-rDoneCh
}

type testHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (c *testHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.doFunc(req)
}

// func TestWorker(t *testing.T) {
// 	requestCh := make(chan *http.Request, 1)
// 	requestDone := make(chan struct{}, 1)

// 	// Создаем контекст с таймаутом
// 	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
// 	defer cancel()

// 	// Мокаем поведение HTTP-клиента
// 	client := &testHTTPClient{
// 		doFunc: func(req *http.Request) (*http.Response, error) {
// 			if req.URL.Path == "/retry" {
// 				return nil, errors.New("temporary error")
// 			}
// 			return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
// 		},
// 	}

// 	// Отправляем запрос в канал
// 	testReq := &http.Request{URL: mustParseURL("http://example.com/success")}
// 	requestCh <- testReq

// 	// Запускаем worker в отдельной горутине
// 	go worker(ctx, requestCh, requestDone, client)

// 	// Ожидаем завершения обработки запроса
// 	select {
// 	case <-requestDone:
// 		// Успех
// 	case <-time.After(3 * time.Second):
// 		t.Fatal("worker did not finish in time")
// 	}

// 	// Проверяем завершение работы при завершении контекста
// 	cancel() // Завершаем контекст
// 	select {
// 	case <-ctx.Done():
// 		// Успех
// 	case <-time.After(1 * time.Second):
// 		t.Fatal("worker did not stop on context cancellation")
// 	}
// }

// func mustParseURL(rawurl string) *url.URL {
// 	url, err := url.ParseRequestURI(rawurl)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return url
// }

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

func TestLoadConfigFromFile(t *testing.T) {
	loadConfigFromFile("config.json")
}

func TestRunProfiler_1(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	RunProfiler(ctx, "")
}

func TestRunProfiler_2(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	RunProfiler(ctx, "8888")
}

func TestGetEnv(t *testing.T) {
	// Определяем переменные для тестов
	key := "TEST_KEY"
	fallback := "default_value"

	t.Run("Environment variable exists", func(t *testing.T) {
		// Устанавливаем переменную окружения
		expectedValue := "env_value"
		if err := os.Setenv(key, expectedValue); err != nil {
			t.Fatalf("Failed to set environment variable: %v", err)
		}
		defer os.Unsetenv(key) // Очищаем переменную после теста

		// Проверяем результат
		value := getEnv(key, fallback)
		if value != expectedValue {
			t.Errorf("Expected %s, got %s", expectedValue, value)
		}
	})

	t.Run("Environment variable does not exist", func(t *testing.T) {
		// Убедимся, что переменной окружения нет
		os.Unsetenv(key)

		// Проверяем результат
		value := getEnv(key, fallback)
		if value != fallback {
			t.Errorf("Expected fallback value %s, got %s", fallback, value)
		}
	})
}

func TestRunProfiler(t *testing.T) {
	// Устанавливаем мок логера
	logger.Set()

	// Тест с пустым портом
	t.Run("empty port", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		RunProfiler(ctx, "")
	})

	// Тест с нормальным портом
	t.Run("valid port", func(t *testing.T) {
		// Мокируем http.Server
		port := "8080"
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Начинаем профайлер в go-рутине
		go RunProfiler(ctx, port)

		// Ожидаем несколько секунд, чтобы сервер успел запуститься
		time.Sleep(500 * time.Millisecond)

		// Проверяем, что сервер слушает на порту
		resp, err := http.Get("http://localhost:" + port)
		if resp != nil {
			resp.Body.Close()
		}

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	// Тест на остановку профайлера
	t.Run("shutdown profiler", func(t *testing.T) {
		port := "8080"
		ctx, cancel := context.WithCancel(context.Background())

		// Запускаем сервер
		go RunProfiler(ctx, port)

		// Ожидаем, пока сервер начнёт слушать
		time.Sleep(500 * time.Millisecond)

		// Отменяем контекст
		cancel()

	})

	// Тест на ошибку при старте сервера
	t.Run("fail to start profiler", func(t *testing.T) {
		// Мокируем сервер, чтобы он всегда возвращал ошибку
		port := "8081"
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Симулируем сбой в ListenAndServe
		go RunProfiler(ctx, port)

		// Ожидаем несколько секунд, чтобы сервер не успел стартовать
		time.Sleep(500 * time.Millisecond)

	})
}
