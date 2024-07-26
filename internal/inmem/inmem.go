package inmem

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
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

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type gauge float64
type counter int64

type MemStorage struct {
	Metrics []*Metrics
}

func NewStorage() MemStorage {
	return MemStorage{Metrics: []*Metrics{}}
}

func (m MemStorage) SetValue(metricType, name, value string) error {
	err := checkWrongType(metricType)
	if err != nil {
		return err
	}

	metric, err := m.GetMetric(metricType, name)
	if err != nil {
		return err
	}

	switch metricType {
	case metricTypeGauge:
		i, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		metric.Value = &i

		m.Metrics = append(m.Metrics, metric)
	case metricTypeCounter:
		i, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		// m.Counters[name] += counter(i)
		// metric.Value = &i
		delta := *metric.Delta + int64(i)

		metric.Delta = &delta

	default:
		return errors.New("not correct type")
	}
	return nil
}

func (m *MemStorage) GetMetric(metricType, name string) (*Metrics, error) {
	err := checkWrongType(metricType)
	if err != nil {
		return nil, err
	}
	var item *Metrics
	for _, metric := range m.Metrics {
		if metric.MType == metricType && metric.ID == name {
			item = metric
			break
		}
	}

	if item != nil {
		item = &Metrics{
			ID:    name,
			MType: metricType,
		}
	}

	return item, err

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

func (m MemStorage) GetMetricPaths() []string {
	paths := []string{}
	for k, v := range m.Gauges {
		paths = append(paths, fmt.Sprintf("gauge/%s/%v", k, v))
	}

	for k, v := range m.Counters {
		paths = append(paths, fmt.Sprintf("counter/%s/%v", k, v))
	}
	return paths
}
