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

type gauge float64
type counter int64

type MemStorage struct {
	Gauges   map[string]gauge
	Counters map[string]counter
}

func NewStorage() *MemStorage {
	return &MemStorage{Gauges: map[string]gauge{}, Counters: map[string]counter{}}
}

func (m *MemStorage) SetValue(t, name, value string) error {
	err := checkWrongType(t)
	if err != nil {
		return err
	}

	switch t {
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
		v, ok := m.Counters[name]
		if ok {
			m.Counters[name] = v + counter(i)
		} else {
			m.Counters[name] = counter(i)
		}

	default:
		return errors.New("not correct type")
	}
	return nil
}

func (m *MemStorage) GetValue(t, name string) (string, error) {
	var err error
	err = checkWrongType(t)
	if err != nil {
		return "", err
	}

	value := ""
	switch t {
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

func (m MemStorage) ToString() string {
	result := ""

	for k, v := range m.Gauges {
		result += fmt.Sprintf("%s: %s\n", k, fmt.Sprint(v))
	}

	for k, v := range m.Counters {
		result += fmt.Sprintf("%s: %s\n", k, fmt.Sprint(v))
	}

	return result
}

func checkWrongType(t string) error {
	if !slices.Contains(availableTypes, t) {
		return errors.New("wrong type")
	}
	return nil
}
