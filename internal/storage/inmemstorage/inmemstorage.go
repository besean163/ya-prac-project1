// Package inmemstorage предоставляет хранилище в памяти
package inmemstorage

import (
	"ya-prac-project1/internal/metrics"
)

// Storage структура представляющая репозиторий
type Storage struct {
	Metrics []metrics.Metrics
}

// NewStorage создает репозиторий
func NewStorage() *Storage {
	return &Storage{
		Metrics: make([]metrics.Metrics, 0),
	}
}

// GetMetrics возвращает все метрики в репозитории
func (s *Storage) GetMetrics() []metrics.Metrics {
	return s.Metrics
}

// CreateMetrics добавляет полученные метрики в репозиторий
func (s *Storage) CreateMetrics(ms []metrics.Metrics) error {
	ems := s.GetMetrics()
	ems = append(ems, ms...)

	s.SetMetrics(ems)
	return nil
}

// UpdateMetrics обновляет полученные метрики в репозитории
func (s *Storage) UpdateMetrics(ms []metrics.Metrics) error {
	ems := s.GetMetrics()
	for _, m := range ms {
		for k := range ems {
			metric := ems[k]
			if m.GetKey() != metric.GetKey() {
				continue
			}
			ems[k] = m
		}
	}
	s.SetMetrics(ems)
	return nil
}

// SetMetrics заменяет метркии в репозитории на полученные
func (s *Storage) SetMetrics(ms []metrics.Metrics) {
	s.Metrics = ms
}
