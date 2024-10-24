// Package filestorage предоставляет хранилище в файле
package filestorage

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"time"
	"ya-prac-project1/internal/logger"
	"ya-prac-project1/internal/metrics"
	"ya-prac-project1/internal/storage/inmemstorage"
)

// Storage структура представляющая репозиторий
type Storage struct {
	*inmemstorage.Storage
	FilePath string
}

// NewStorage создает репозиторий
func NewStorage(ctx context.Context, path string, restore bool, dumpInterval int64) (*Storage, error) {
	store := Storage{
		Storage:  inmemstorage.NewStorage(),
		FilePath: path,
	}

	if restore {
		if err := store.restore(); err != nil {
			return nil, err
		}
	}

	store.runIntervalDumper(ctx, dumpInterval)
	store.runDumper(ctx)

	return &store, nil
}

func (s *Storage) restore() error {
	file, err := os.OpenFile(s.FilePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	buf := bufio.NewScanner(file)
	items := []metrics.Metrics{}

	for {
		if !buf.Scan() {
			break
		}
		data := buf.Bytes()
		item := metrics.Metrics{}
		err := json.Unmarshal(data, &item)
		if err != nil {
			logger.Get().Info(err.Error())
			continue
		}
		items = append(items, item)
	}

	s.Storage.SetMetrics(items)
	return nil
}

func (s *Storage) dump() error {
	file, err := os.OpenFile(s.FilePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	items := s.Storage.GetMetrics()
	for _, item := range items {
		row, err := json.Marshal(item)
		if err != nil {
			return err
		}

		file.Write(row)
		file.WriteString("\n")
	}

	return nil
}

func (s *Storage) runDumper(ctx context.Context) {
	go func(ctx context.Context) {
		<-ctx.Done()
		s.dump()
	}(ctx)
}

func (s *Storage) runIntervalDumper(ctx context.Context, interval int64) {
	if interval == 0 {
		return
	}

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(interval) * time.Second):
				s.dump()
			}
		}
	}(ctx)
}
