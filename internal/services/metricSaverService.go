package services

import (
	"fmt"
	"ya-prac-project1/internal/metrics"
)

type SaveStorage interface {
	GetMetrics() []metrics.Metrics
	CreateMetrics([]metrics.Metrics) error
	UpdateMetrics([]metrics.Metrics) error
}

type MetricSaverService struct {
	storage SaveStorage
}

func NewMetricSaverService(storage SaveStorage) *MetricSaverService {
	return &MetricSaverService{
		storage: storage,
	}
}

func (s *MetricSaverService) GetMetric(metricType, name string) (metrics.Metrics, error) {
	metric := metrics.Metrics{
		MType: metricType,
		ID:    name,
	}

	metricsMap := s.getMetricsKeyMap()
	metric, ok := metricsMap[metric.GetKey()]
	if !ok {
		return metric, fmt.Errorf("metric not found")
	}

	return metric, nil
}

func (s *MetricSaverService) GetMetrics() []metrics.Metrics {
	return s.storage.GetMetrics()
}

func (s *MetricSaverService) SaveMetric(m metrics.Metrics) error {
	err := m.Validate()
	if err != nil {
		return err
	}

	var createMetrics []metrics.Metrics
	var updateMetrics []metrics.Metrics

	metricsMap := s.getMetricsKeyMap()
	em, ok := metricsMap[m.GetKey()]
	if !ok {
		createMetrics = append(createMetrics, m)
	} else {
		err := em.SetValue(m.GetValue())
		if err != nil {
			return err
		}
		updateMetrics = append(updateMetrics, em)
	}

	if len(createMetrics) > 0 {
		err := s.storage.CreateMetrics(createMetrics)
		if err != nil {
			return err
		}
	}

	if len(updateMetrics) > 0 {
		err := s.storage.UpdateMetrics(updateMetrics)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *MetricSaverService) SaveMetrics(ms []metrics.Metrics) error {
	var createMetrics []metrics.Metrics
	var updateMetrics []metrics.Metrics
	metricsMap := s.getMetricsKeyMap()
	for _, m := range ms {
		err := m.Validate()
		if err != nil {
			return err
		}
		em, ok := metricsMap[m.GetKey()]
		if !ok {
			createMetrics = append(createMetrics, m)
			metricsMap[m.GetKey()] = m
		} else {
			err := em.SetValue(m.GetValue())
			if err != nil {
				return err
			}
			updateMetrics = append(updateMetrics, em)
		}

	}

	if len(createMetrics) > 0 {
		err := s.storage.CreateMetrics(createMetrics)
		if err != nil {
			return err
		}
	}

	if len(updateMetrics) > 0 {
		err := s.storage.UpdateMetrics(updateMetrics)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *MetricSaverService) getMetricsKeyMap() map[string]metrics.Metrics {
	m := make(map[string]metrics.Metrics)
	for _, metric := range s.GetMetrics() {
		m[metric.GetKey()] = metric
	}
	return m
}
