package inmem

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"time"
	"ya-prac-project1/internal/logger"
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
	Gauges       map[string]gauge
	Counters     map[string]counter
	filePath     string
	dumpInterval int
}

func NewStorage(filePath string, restore bool, dumpInterval int) (MemStorage, error) {
	storage := MemStorage{
		Gauges:       map[string]gauge{},
		Counters:     map[string]counter{},
		dumpInterval: dumpInterval,
		filePath:     filePath,
	}

	// восстанавливаем метрики по надобности
	if restore {
		if err := storage.Restore(); err != nil {
			return storage, err
		}
	}

	// запускаем дампер если задан интервал
	if storage.dumpInterval > 0 {
		go func() {
			for {
				time.Sleep(time.Second * time.Duration(dumpInterval))
				storage.Dump()
			}
		}()
	}

	return storage, nil
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

	if m.dumpInterval == 0 {
		m.Dump()
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
		metric.MType = metricTypeCounter
		metric.ID = k
		metric.Delta = &delta
		result = append(result, metric)
	}
	return result
}

func (m MemStorage) SetMetrics(metrics []metrics.Metrics) error {
	for _, metric := range metrics {
		err := m.SetValue(metric.MType, metric.ID, metric.GetValue())
		if err != nil {
			return err
		}
	}
	return nil
}

func (s MemStorage) Restore() error {
	if s.filePath == "" {
		return nil
	}
	file, err := os.OpenFile(s.filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	buf := bufio.NewScanner(file)
	items := []metrics.Metrics{}

	for {
		if !buf.Scan() {
			break
		}
		fmt.Println(string(buf.Bytes()))
		data := buf.Bytes()
		item := metrics.Metrics{}
		err := json.Unmarshal(data, &item)
		if err != nil {
			logger.Get().Info(err.Error())
			continue
		}
		items = append(items, item)
	}

	s.SetMetrics(items)
	return nil
}

func (s MemStorage) Dump() error {
	if s.filePath == "" {
		return nil
	}

	file, err := os.OpenFile(s.filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	items := s.GetMetrics()
	for _, item := range items {
		row, err := json.Marshal(item)
		if err != nil {
			return err
		}

		file.Write(row)
		file.WriteString("\n")
	}

	return nil
}
