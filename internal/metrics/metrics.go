package metrics

import (
	"errors"
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

func NewMetric(name, mType, value string) Metrics {
	item := Metrics{ID: name, MType: mType}
	item.SetValue(value)
	return item
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
		if m.Delta != nil {
			delta = delta + *m.Delta
		}
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

func (m Metrics) Validate() error {
	if m.MType != MetricTypeGauge &&
		m.MType != MetricTypeCounter {
		return errors.New("wrong metric type")
	}

	return nil
}

func (m Metrics) GetKey() string {
	return fmt.Sprintf("%s_%s", m.MType, m.ID)
}

func (m Metrics) GetInfo() string {
	return fmt.Sprintf("%s - %s - %s", m.ID, m.MType, m.GetValue())
}
