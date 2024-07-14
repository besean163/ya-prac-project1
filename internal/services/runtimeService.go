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
	SetValue(t, name, value string) error
	GetValue(t, name string) (string, error)
	GetMetricPaths() []string
	ToString() string
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
	s.storage.SetValue("counters", "PollCount", string(1))
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
		"Alloc":        string(stat.Alloc),
		"BuckHashSys":  string(stat.BuckHashSys),
		"Frees":        string(stat.Frees),
		"GCSys":        string(stat.GCSys),
		"HeapAlloc":    string(stat.HeapAlloc),
		"HeapIdle":     string(stat.HeapIdle),
		"HeapInuse":    string(stat.HeapInuse),
		"HeapObjects":  string(stat.HeapObjects),
		"HeapReleased": string(stat.HeapReleased),
		"LastGC":       string(stat.LastGC),
		"Lookups":      string(stat.Lookups),
		"MCacheInuse":  string(stat.MCacheInuse),
		"MCacheSys":    string(stat.MCacheSys),
		"MSpanInuse":   string(stat.MSpanInuse),
		"MSpanSys":     string(stat.MSpanSys),
		"Mallocs":      string(stat.Mallocs),
		"NextGC":       string(stat.NextGC),
		"NumForcedGC":  string(stat.NumForcedGC),
		"NumGC":        string(stat.NumGC),
		"OtherSys":     string(stat.OtherSys),
		"PauseTotalNs": string(stat.PauseTotalNs),
		"StackInuse":   string(stat.StackInuse),
		"Sys":          string(stat.Sys),
		"TotalAlloc":   string(stat.TotalAlloc),
	}
}
