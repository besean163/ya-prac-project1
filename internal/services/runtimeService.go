package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"
	"ya-prac-project1/internal/logger"
	"ya-prac-project1/internal/metrics"
	metrs "ya-prac-project1/internal/metrics"

	"go.uber.org/zap"
)

type Storage interface {
	GetMetric(metricType, name string) *metrics.Metrics
	UpdateMetric(metric *metrics.Metrics) error
	GetAllMetrics() []*metrics.Metrics
}

type RuntimeService struct {
	storage Storage
}

func NewRuntimeService(storage Storage) RuntimeService {
	return RuntimeService{storage: storage}
}

func (s *RuntimeService) UpdateMetrics() error {
	metrics := getRuntimeMetrics()
	for name, value := range metrics {
		metric, err := metrs.New(metrs.MetricTypeGauge, name, value)
		if err != nil {
			return err
		}
		s.storage.UpdateMetric(metric)
	}

	rand.New(rand.NewSource(time.Now().Unix()))
	randMetric, err := metrs.New(metrs.MetricTypeGauge, "RandomValue", fmt.Sprint(rand.Float64()))
	if err != nil {
		return err
	}
	s.storage.UpdateMetric(randMetric)

	poolMetric, err := metrs.New(metrs.MetricTypeCounter, "PollCount", fmt.Sprint(1))
	if err != nil {
		return err
	}
	s.storage.UpdateMetric(poolMetric)

	return nil
}

func (s *RuntimeService) SendMetrics(serverEndpoint string) {
	for _, metric := range s.storage.GetAllMetrics() {
		makeUpdateRequest(serverEndpoint, metric)
	}
}

func makeUpdateRequest(serverEndpoint string, metric *metrics.Metrics) {
	b, err := json.Marshal(metric)
	if err != nil {
		logger.Get().Info("can marshal", zap.Error(err))
		return
	}
	br := bytes.NewReader(b)
	client := http.Client{}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/update/", serverEndpoint), br)
	if err != nil {
		logger.Get().Info("can't create request", zap.Error(err))
		return
	}

	response, err := client.Do(req)
	if err != nil {
		logger.Get().Info("call error", zap.Error(err))
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		logger.Get().Info(
			"Error write metrics",
			zap.String("path", req.URL.Path),
			zap.Int("code", response.StatusCode),
		)
	}

}

func getRuntimeMetrics() map[string]string {
	stat := runtime.MemStats{}
	runtime.ReadMemStats(&stat)
	return map[string]string{
		"Alloc":        fmt.Sprint(stat.Alloc),
		"BuckHashSys":  fmt.Sprint(stat.BuckHashSys),
		"Frees":        fmt.Sprint(stat.Frees),
		"GCSys":        fmt.Sprint(stat.GCSys),
		"HeapAlloc":    fmt.Sprint(stat.HeapAlloc),
		"HeapIdle":     fmt.Sprint(stat.HeapIdle),
		"HeapInuse":    fmt.Sprint(stat.HeapInuse),
		"HeapObjects":  fmt.Sprint(stat.HeapObjects),
		"HeapReleased": fmt.Sprint(stat.HeapReleased),
		"LastGC":       fmt.Sprint(stat.LastGC),
		"Lookups":      fmt.Sprint(stat.Lookups),
		"MCacheInuse":  fmt.Sprint(stat.MCacheInuse),
		"MCacheSys":    fmt.Sprint(stat.MCacheSys),
		"MSpanInuse":   fmt.Sprint(stat.MSpanInuse),
		"MSpanSys":     fmt.Sprint(stat.MSpanSys),
		"Mallocs":      fmt.Sprint(stat.Mallocs),
		"NextGC":       fmt.Sprint(stat.NextGC),
		"NumForcedGC":  fmt.Sprint(stat.NumForcedGC),
		"NumGC":        fmt.Sprint(stat.NumGC),
		"OtherSys":     fmt.Sprint(stat.OtherSys),
		"PauseTotalNs": fmt.Sprint(stat.PauseTotalNs),
		"StackInuse":   fmt.Sprint(stat.StackInuse),
		"Sys":          fmt.Sprint(stat.Sys),
		"TotalAlloc":   fmt.Sprint(stat.TotalAlloc),
	}
}
