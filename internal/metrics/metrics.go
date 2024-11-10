// Package metrics предоставляет структуру метрикик и методы работы с ней
package metrics

import (
	"errors"
	"fmt"
	"strconv"
)

const (
	// MetricTypeGauge — тип метрики содержащей значение с плавающей точкой
	MetricTypeGauge = "gauge"
	// MetricTypeCounter — тип метрики содержащей целое значение
	MetricTypeCounter = "counter"
)

// Metrics представляет структуру метрики
type Metrics struct {
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	ID    string   `json:"id"`
	MType string   `json:"type"`
}

// NewMetric создает новую метрику
func NewMetric(name, mType, value string) Metrics {
	item := Metrics{ID: name, MType: mType}
	item.SetValue(value)
	return item
}

// SetValue обновляет значение метрики
func (m *Metrics) SetValue(value string) error {
	switch m.MType {
	case MetricTypeGauge:
		i, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		m.Value = &i
	case MetricTypeCounter:
		i, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		delta := int64(i)
		if m.Delta != nil {
			delta = delta + *m.Delta
		}
		m.Delta = &delta
	}
	return nil
}

// GetValue возвращает значение метрики
func (m *Metrics) GetValue() string {
	switch m.MType {
	case MetricTypeGauge:
		if m.Value != nil {
			return fmt.Sprintf("%v", *m.Value)
		}
	case MetricTypeCounter:
		if m.Delta != nil {
			return fmt.Sprintf("%v", *m.Delta)
		}
	}
	return ""
}

// Validate валидирует метрку, проверяет ее тип
func (m Metrics) Validate() error {
	if m.MType != MetricTypeGauge &&
		m.MType != MetricTypeCounter {
		return errors.New("wrong metric type")
	}

	return nil
}

// GetKey получает уникальный ключ метрики
func (m Metrics) GetKey() string {
	return fmt.Sprintf("%s_%s", m.MType, m.ID)
}

// GetInfo получает информацию о метрике в строковом формате
func (m Metrics) GetInfo() string {
	return fmt.Sprintf("%s - %s - %s", m.ID, m.MType, m.GetValue())
}
