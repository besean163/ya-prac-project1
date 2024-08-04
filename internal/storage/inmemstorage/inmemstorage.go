package inmemstorage

import (
	"errors"
	"strconv"
	"ya-prac-project1/internal/metrics"
)

type Storage struct {
	Metrics []*metrics.Metrics
}

func NewStorage() *Storage {
	return &Storage{
		Metrics: make([]*metrics.Metrics, 0),
	}
}

func (s *Storage) SetValue(metricType, name, value string) error {
	err := checkWrongType(metricType)
	if err != nil {
		return err
	}

	metric := s.GetMetric(metricType, name)
	if metric == nil {
		metric = &metrics.Metrics{
			MType: metricType,
			ID:    name,
		}
		s.Metrics = append(s.Metrics, metric)
	}

	switch metricType {
	case metrics.MetricTypeGauge:
		i, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		metric.Value = &i
	case metrics.MetricTypeCounter:
		i, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		delta := int64(i)
		if metric.Delta == nil {
			metric.Delta = &delta
		} else {
			delta := *metric.Delta + int64(i)
			metric.Delta = &delta
		}
	}
	return nil
}

func (s *Storage) GetValue(metricType, name string) (string, error) {
	err := checkWrongType(metricType)
	if err != nil {
		return "", err
	}

	metric := s.GetMetric(metricType, name)
	if metric == nil {
		return "", errors.New("not found metric")
	}

	return metric.GetValue(), nil
}

func checkWrongType(metricType string) error {
	if metricType != metrics.MetricTypeGauge &&
		metricType != metrics.MetricTypeCounter {
		return errors.New("wrong type")
	}

	return nil
}

func (s *Storage) GetMetric(metricType, name string) *metrics.Metrics {
	for _, metric := range s.Metrics {
		if metric.MType != metricType || metric.ID != name {
			continue
		}
		return metric
	}
	return nil
}

func (s *Storage) GetMetrics() []*metrics.Metrics {
	return s.Metrics
}

func (s *Storage) SetMetrics(metrics []metrics.Metrics) error {
	for _, metric := range metrics {
		err := s.SetValue(metric.MType, metric.ID, metric.GetValue())
		if err != nil {
			return err
		}
	}
	return nil
}
