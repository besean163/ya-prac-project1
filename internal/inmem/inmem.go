package inmem

import (
	"errors"
	"ya-prac-project1/internal/metrics"
)

type MemStorage struct {
	Metrics []*metrics.Metrics
}

func NewStorage() MemStorage {
	return MemStorage{Metrics: []*metrics.Metrics{}}
}

func (m MemStorage) GetMetric(metricType, name string) *metrics.Metrics {
	var item *metrics.Metrics
	for _, metric := range m.Metrics {
		if metric.MType == metricType && metric.ID == name {
			item = metric
			break
		}
	}

	return item

}

func (m *MemStorage) GetAllMetrics() []*metrics.Metrics {
	return m.Metrics
}

func (m *MemStorage) UpdateMetric(inputMetric *metrics.Metrics) error {
	if inputMetric == nil {
		return errors.New("get empty metric")
	}

	if !inputMetric.IsValidType() {
		return errors.New("wrong metric type")
	}

	metric := m.GetMetric(inputMetric.MType, inputMetric.ID)
	if metric == nil {
		m.Metrics = append(m.Metrics, inputMetric)
	} else {
		metric.Update(inputMetric)
	}

	return nil
}

func (m MemStorage) GetRows() []string {
	result := []string{}

	for _, metric := range m.Metrics {
		result = append(result, metric.GetRow())
	}

	return result
}
