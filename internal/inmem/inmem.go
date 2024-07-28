package inmem

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"ya-prac-project1/internal/metrics"
)

const (
	metricTypeGauge   = "gauge"
	metricTypeCounter = "counter"
)

/* допустимые типы для хранения */
var availableTypes = []string{
	metricTypeGauge,
	metricTypeCounter,
}

type gauge float64
type counter int64

type MemStorage struct {
	Gauges   map[string]gauge
	Counters map[string]counter
}

func NewStorage() MemStorage {
	return MemStorage{Gauges: map[string]gauge{}, Counters: map[string]counter{}}
}

func (m MemStorage) SetValue(metricType, name, value string) error {
	err := checkWrongType(metricType)
	if err != nil {
		return err
	}

	switch metricType {
	case metricTypeGauge:
		i, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		m.Gauges[name] = gauge(i)
	case metricTypeCounter:
		i, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		m.Counters[name] += counter(i)

	default:
		return errors.New("not correct type")
	}
	return nil
}

func (m MemStorage) GetValue(metricType, name string) (string, error) {
	var err error
	err = checkWrongType(metricType)
	if err != nil {
		return "", err
	}

	value := ""
	switch metricType {
	case metricTypeGauge:
		v, ok := m.Gauges[name]
		if ok {
			value = fmt.Sprint(v)
		}

	case metricTypeCounter:
		v, ok := m.Counters[name]
		if ok {
			value = fmt.Sprint(v)
		}
	}
	if value == "" {
		err = errors.New("not found metric")
	}
	return value, err
}

func (m MemStorage) GetRows() []string {
	result := []string{}

	for k, v := range m.Gauges {
		row := fmt.Sprintf("%s: %s\n", k, fmt.Sprint(v))
		result = append(result, row)
	}

	for k, v := range m.Counters {
		row := fmt.Sprintf("%s: %s\n", k, fmt.Sprint(v))
		result = append(result, row)
	}

	return result
}

func checkWrongType(t string) error {
	if !slices.Contains(availableTypes, t) {
		return errors.New("wrong type")
	}
	return nil
}

func (m MemStorage) GetMetrics() []metrics.Metrics {
	result := []metrics.Metrics{}
	for k, v := range m.Gauges {
		value := float64(v)
		metric := metrics.Metrics{}
		metric.MType = metricTypeGauge
		metric.ID = k
		metric.Value = &value
		result = append(result, metric)
	}

	for k, v := range m.Counters {
		delta := int64(v)
		metric := metrics.Metrics{}
		metric.MType = metricTypeGauge
		metric.ID = k
		metric.Delta = &delta
		result = append(result, metric)
	}
	return result
}
