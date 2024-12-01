package main

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"
	"ya-prac-project1/internal/logger"
	"ya-prac-project1/internal/storage/filestorage"
	"ya-prac-project1/internal/storage/inmemstorage"

	_ "net/http/pprof"

	"github.com/stretchr/testify/assert"
)

func TestGetStorage_inmemory(t *testing.T) {
	c := ServerConfig{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s, err := getStorage(ctx, c, nil)

	assert.Nil(t, err)
	assert.IsType(t, &inmemstorage.Storage{}, s)
}

func TestGetStorage_file(t *testing.T) {
	c := ServerConfig{
		StoreFile: "test",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s, err := getStorage(ctx, c, nil)

	assert.Nil(t, err)
	assert.IsType(t, &filestorage.Storage{}, s)
}

func TestRunProfiler(t *testing.T) {
	_ = logger.Set()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	RunProfiler(ctx, ":9898")
}

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
	os.Setenv("STORE_INTERVAL", "10")
	os.Setenv("FILE_STORAGE_PATH", "test_store_file")
	os.Setenv("RESTORE", "false")
	os.Setenv("DATABASE_DSN", "dns_row")
	os.Setenv("KEY", "test_key")

	c := NewConfig()
	assert.Equal(t, ":8081", c.Endpoint)
	assert.Equal(t, 10, c.StoreInterval)
	assert.Equal(t, "test_store_file", c.StoreFile)
	assert.Equal(t, false, c.Restore)
	assert.Equal(t, "dns_row", c.BaseDNS)
	assert.Equal(t, "test_key", c.HashKey)
	assert.Equal(t, "", c.Profiler)
}

func TestShowBuildInfo(t *testing.T) {
	showBuildInfo()
}

func TestLoadConfigFromFile(t *testing.T) {
	loadConfigFromFile("config.json")
}

func TestRungRPCServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(time.Second * 1)
		cancel()
	}()

	RungRPCServer(ctx, ":8888", nil)
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
