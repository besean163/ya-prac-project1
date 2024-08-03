package metrics

import (
	"fmt"
	"strconv"
)

const (
	MetricTypeGauge   = "gauge"
	MetricTypeCounter = "counter"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

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
		m.Delta = &delta
	}
	return nil
}

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
