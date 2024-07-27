package metrics

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
)

const (
	MetricTypeGauge   = "gauge"
	MetricTypeCounter = "counter"
)

var availableTypes = []string{
	MetricTypeGauge,
	MetricTypeCounter,
}

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func New(mType, id, value string) (*Metrics, error) {
	if err := checkWrongType(mType); err != nil {
		return nil, err
	}

	metric := &Metrics{
		ID:    id,
		MType: mType,
	}

	if err := metric.setValue(value); err != nil {
		return nil, err
	}

	return metric, nil
}

func (metric *Metrics) setValue(value string) error {
	if value == "" {
		return nil
	}

	switch metric.MType {
	case MetricTypeGauge:
		i, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		metric.Value = &i

	case MetricTypeCounter:
		i, err := strconv.Atoi(value)
		if err != nil {
			return err
		}

		delta := int64(i)
		metric.Delta = &delta

	default:
		return errors.New("not correct type")
	}
	return nil
}

func (metric *Metrics) Update(inputMetric *Metrics) {
	if metric.Value != nil && metric.MType == MetricTypeGauge {
		metric.Value = inputMetric.Value
	} else if metric.Delta != nil && metric.MType == MetricTypeCounter {
		delta := *metric.Delta + *inputMetric.Delta
		metric.Delta = &delta
	}
}

func (metric *Metrics) GetRow() string {
	value := ""
	if metric.Delta != nil {
		value = fmt.Sprintf("%v", *metric.Delta)
	} else if metric.Value != nil {
		value = fmt.Sprintf("%v", *metric.Value)
	}

	return fmt.Sprintf("%s: %s\n", metric.ID, value)
}

func (metric *Metrics) IsValidType() bool {
	return checkWrongType(metric.MType) == nil
}

func checkWrongType(t string) error {
	if !slices.Contains(availableTypes, t) {
		return errors.New("wrong metric type")
	}
	return nil
}
