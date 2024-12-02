// Package services предоставляет сервисы для работы с метриками
package services

import (
	"context"
	"fmt"
	"ya-prac-project1/internal/logger"
	"ya-prac-project1/internal/metrics"
	pb "ya-prac-project1/internal/services/proto"

	"go.uber.org/zap"
)

// SaveStorage структура представляющая интерфейс репозитория для работы с сервисом MetricSaverService
type SaveStorage interface {
	GetMetrics() []metrics.Metrics
	CreateMetrics([]metrics.Metrics) error
	UpdateMetrics([]metrics.Metrics) error
}

// MetricSaverService структура представляющая сервис для хранения метрик
type MetricSaverService struct {
	storage SaveStorage
}

// NewMetricSaverService создает сервис
func NewMetricSaverService(storage SaveStorage) *MetricSaverService {
	return &MetricSaverService{
		storage: storage,
	}
}

// GetMetric получает метрику по имени и типу. Возвращает ошибку в случае если не находит запрашиваемую метрику
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

// GetMetrics отдает все метрики которы есть в репозитории сервиса
func (s *MetricSaverService) GetMetrics() []metrics.Metrics {
	return s.storage.GetMetrics()
}

// SaveMetric сохраняет входную метрику
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

// SaveMetrics сохраняет набор входных метрик
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

func (s *MetricSaverService) UpdateMetrics(ctx context.Context, in *pb.SaveMetricsRequest) (*pb.SaveMetricsResponse, error) {
	var response pb.SaveMetricsResponse

	storeMetrics := make([]metrics.Metrics, 0)
	for _, m := range in.Metrics {
		sm := metrics.Metrics{
			MType: m.Type,
			ID:    m.Id,
		}
		if m.Value != nil {
			sm.Value = m.Value
		}
		if m.Delta != nil {
			sm.Delta = m.Delta
		}
		storeMetrics = append(storeMetrics, sm)
	}

	err := s.SaveMetrics(storeMetrics)
	if err != nil {
		logger.Get().Info("server grpc error", zap.String("error", err.Error()))
		response.Error = err.Error()
	} else {
		logger.Get().Info("grpc success")
	}

	return &response, err
}
