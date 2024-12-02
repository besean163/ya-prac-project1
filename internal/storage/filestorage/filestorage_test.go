package filestorage

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
	"ya-prac-project1/internal/metrics"
	"ya-prac-project1/internal/storage/inmemstorage"

	"github.com/stretchr/testify/assert"
)

func TestNewStorage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s, _ := NewStorage(ctx, "test", false, 1)

	assert.Equal(t, "test", s.FilePath)
	// чтобы воркеры внутри успели запуститься
	time.Sleep(time.Millisecond * 100)
}

func TestNewStorage2(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s, _ := NewStorage(ctx, "test", true, 1)

	assert.Equal(t, "test", s.FilePath)
	// чтобы воркеры внутри успели запуститься
	time.Sleep(time.Millisecond * 100)
}

func TestDump_Success(t *testing.T) {
	// Создаем временный файл для теста
	tempFile, err := os.CreateTemp("", "test_dump")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name()) // Удаляем файл после теста

	// Инициализируем Storage с временным файлом
	store := &Storage{
		Storage:  inmemstorage.NewStorage(),
		FilePath: tempFile.Name(),
	}

	// Устанавливаем метрики
	metricsToSave := []metrics.Metrics{
		{MType: "gauge", ID: "metric1", Value: floatPtr(123.45)},
		{MType: "counter", ID: "metric2", Delta: int64Ptr(10)},
	}
	store.Storage.SetMetrics(metricsToSave)

	// Вызываем dump
	err = store.dump()
	assert.NoError(t, err)

	// Проверяем содержимое файла
	fileContent, err := os.ReadFile(tempFile.Name())
	assert.NoError(t, err)

	// Сравниваем сохраненные метрики с содержимым файла
	var savedMetrics []metrics.Metrics
	lines := splitLines(string(fileContent))
	for _, line := range lines {
		var m metrics.Metrics
		err := json.Unmarshal([]byte(line), &m)
		assert.NoError(t, err)
		savedMetrics = append(savedMetrics, m)
	}

	assert.Equal(t, metricsToSave, savedMetrics)
}

func TestDump_FileOpenError(t *testing.T) {
	// Создаем Storage с недоступным файлом
	store := &Storage{
		Storage:  inmemstorage.NewStorage(),
		FilePath: "/nonexistent/path/test_dump",
	}

	// Вызываем dump
	err := store.dump()
	assert.Error(t, err)
}

func TestDump_JSONMarshalError(t *testing.T) {
	// Создаем временный файл для теста
	tempFile, err := os.CreateTemp("", "test_dump")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name()) // Удаляем файл после теста

	// Инициализируем Storage с временным файлом
	store := &Storage{
		Storage:  inmemstorage.NewStorage(),
		FilePath: tempFile.Name(),
	}

	// Устанавливаем метрику с некорректным типом данных для маршализации
	store.Storage.SetMetrics([]metrics.Metrics{
		{MType: "invalid_type", ID: "metric1", Value: nil},
	})

	// Вызываем dump
	err = store.dump()
	assert.NoError(t, err)
}

func splitLines(content string) []string {
	lines := []string{}
	for _, line := range strings.Split(content, "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func floatPtr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}
