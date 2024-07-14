package services

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

type Storage interface {
	SetValue(metricType, name, value string) error
	GetValue(metricType, name string) (string, error)
	GetMetricPaths() []string
}

type RuntimeService struct {
	storage Storage
}

func NewRuntimeService(storage Storage) RuntimeService {
	return RuntimeService{storage: storage}
}

func (s *RuntimeService) UpdateMetrics() {
	metrics := getRuntimeMetrics()
	for name, value := range metrics {
		s.storage.SetValue("gauge", name, value)
	}

	rand.New(rand.NewSource(time.Now().Unix()))
	s.storage.SetValue("gauge", "RandomValue", fmt.Sprint(rand.Float64()))
	s.storage.SetValue("counters", "PollCount", fmt.Sprint(1))
}

func (s *RuntimeService) SendMetrics(serverEndpoint string) {
	paths := s.storage.GetMetricPaths()
	for _, path := range paths {
		makeUpdateRequest(path, serverEndpoint)
	}
}

func makeUpdateRequest(path string, serverEndpoint string) {
	updatePath := "/update"
	client := http.Client{}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s", serverEndpoint), nil)
	if err != nil {
		fmt.Printf("can't create request. Error: %s\n", err)
		return
	}

	req.URL.Path = fmt.Sprintf("%s/%s", updatePath, path)
	response, err := client.Do(req)
	if err != nil {
		log.Printf("call error. Error: %s\n", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Println("Error write metrics")
		log.Println("Path:", req.URL.Path)
		log.Println("Code:", response.StatusCode)
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
