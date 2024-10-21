package inmemstorage

import (
	"ya-prac-project1/internal/metrics"
)

type Storage struct {
	Metrics []metrics.Metrics
}

func NewStorage() *Storage {
	return &Storage{
		Metrics: make([]metrics.Metrics, 0),
	}
}

func (s *Storage) GetMetrics() []metrics.Metrics {
	return s.Metrics
}

func (s *Storage) CreateMetrics(ms []metrics.Metrics) error {
	ems := s.GetMetrics()
	ems = append(ems, ms...)

	s.SetMetrics(ems)
	return nil
}

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

func (s *Storage) SetMetrics(ms []metrics.Metrics) {
	s.Metrics = ms
}
