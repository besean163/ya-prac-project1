package inmemstorage

import (
	"errors"
	"strconv"
	"ya-prac-project1/internal/metrics"
)

// const (
// 	metricTypeGauge   = "gauge"
// 	metricTypeCounter = "counter"
// )

/* допустимые типы для хранения */
// var availableTypes = []string{
// 	metricTypeGauge,
// 	metricTypeCounter,
// }

// type gauge float64
// type counter int64

// type MemStorage struct {
// 	Gauges   map[string]gauge
// 	Counters map[string]counter
// }

// func NewStorage() MemStorage {
// 	return MemStorage{Gauges: map[string]gauge{}, Counters: map[string]counter{}}
// }

// func (m MemStorage) SetValue(metricType, name, value string) error {
// 	err := checkWrongType(metricType)
// 	if err != nil {
// 		return err
// 	}

// 	switch metricType {
// 	case metricTypeGauge:
// 		i, err := strconv.ParseFloat(value, 64)
// 		if err != nil {
// 			return err
// 		}
// 		m.Gauges[name] = gauge(i)
// 	case metricTypeCounter:
// 		i, err := strconv.Atoi(value)
// 		if err != nil {
// 			return err
// 		}
// 		m.Counters[name] += counter(i)

// 	default:
// 		return errors.New("not correct type")
// 	}
// 	return nil
// }

// func (m MemStorage) GetValue(metricType, name string) (string, error) {
// 	var err error
// 	err = checkWrongType(metricType)
// 	if err != nil {
// 		return "", err
// 	}

// 	value := ""
// 	switch metricType {
// 	case metricTypeGauge:
// 		v, ok := m.Gauges[name]
// 		if ok {
// 			value = fmt.Sprint(v)
// 		}

// 	case metricTypeCounter:
// 		v, ok := m.Counters[name]
// 		if ok {
// 			value = fmt.Sprint(v)
// 		}
// 	}
// 	if value == "" {
// 		err = errors.New("not found metric")
// 	}
// 	return value, err
// }

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

// func (m MemStorage) GetMetrics() []metrics.Metrics {
// 	result := []metrics.Metrics{}
// 	for k, v := range m.Gauges {
// 		value := float64(v)
// 		metric := metrics.Metrics{}
// 		metric.MType = metricTypeGauge
// 		metric.ID = k
// 		metric.Value = &value
// 		result = append(result, metric)
// 	}

// 	for k, v := range m.Counters {
// 		delta := int64(v)
// 		metric := metrics.Metrics{}
// 		metric.MType = metricTypeCounter
// 		metric.ID = k
// 		metric.Delta = &delta
// 		result = append(result, metric)
// 	}
// 	return result
// }

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
