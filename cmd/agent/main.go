package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

type gauge float64
type counter int64

type MetricSource interface {
	GetMetrics() map[string]gauge
}

type Storage interface {
	Update()
	getGauges() map[string]float64
	getCounters() map[string]int64
}

type RuntimeSource struct {
	stat runtime.MemStats
}

func (s RuntimeSource) GetMetrics() map[string]gauge {
	runtime.ReadMemStats(&s.stat)
	return map[string]gauge{
		"Alloc":        gauge(s.stat.Alloc),
		"BuckHashSys":  gauge(s.stat.BuckHashSys),
		"Frees":        gauge(s.stat.Frees),
		"GCSys":        gauge(s.stat.GCSys),
		"HeapAlloc":    gauge(s.stat.HeapAlloc),
		"HeapIdle":     gauge(s.stat.HeapIdle),
		"HeapInuse":    gauge(s.stat.HeapInuse),
		"HeapObjects":  gauge(s.stat.HeapObjects),
		"HeapReleased": gauge(s.stat.HeapReleased),
		"LastGC":       gauge(s.stat.LastGC),
		"Lookups":      gauge(s.stat.Lookups),
		"MCacheInuse":  gauge(s.stat.MCacheInuse),
		"MCacheSys":    gauge(s.stat.MCacheSys),
		"MSpanInuse":   gauge(s.stat.MSpanInuse),
		"MSpanSys":     gauge(s.stat.MSpanSys),
		"Mallocs":      gauge(s.stat.Mallocs),
		"NextGC":       gauge(s.stat.NextGC),
		"NumForcedGC":  gauge(s.stat.NumForcedGC),
		"NumGC":        gauge(s.stat.NumGC),
		"OtherSys":     gauge(s.stat.OtherSys),
		"PauseTotalNs": gauge(s.stat.PauseTotalNs),
		"StackInuse":   gauge(s.stat.StackInuse),
		"Sys":          gauge(s.stat.Sys),
		"TotalAlloc":   gauge(s.stat.TotalAlloc),
	}
}

const (
	poolInterval   = 1
	reportInterval = 2
)

func main() {
	ms := RuntimeSource{stat: runtime.MemStats{}}
	mp := &MetricStorage{Source: ms, Metrics: map[string]gauge{}, Counters: map[string]counter{}}

	go func() {
		for {
			mp.Update()
			time.Sleep(poolInterval * time.Second)
		}
	}()

	for {
		time.Sleep(reportInterval * time.Second)
		SendMetrics(mp)
	}
}

func SendMetrics(s Storage) {
	for k, v := range s.getGauges() {
		makeUpdateRequest(fmt.Sprintf("gauge/%s/%v", k, v))
	}

	for k, v := range s.getCounters() {
		makeUpdateRequest(fmt.Sprintf("counter/%s/%v", k, v))
	}
}

func makeUpdateRequest(path string) {
	updatePath := "/update"
	client := http.Client{}
	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:8080", nil)
	if err != nil {
		fmt.Printf("can't create request. Error: %s\n", err)
		return
	}

	req.URL.Path = fmt.Sprintf("%s/%s", updatePath, path)
	response, err := client.Do(req)
	if err != nil {
		// fmt.Printf("call error. Error: %s\n", err)
		return
	}

	if response.StatusCode != http.StatusOK {
		fmt.Println("Error write metrics")
		fmt.Println("Path:", req.URL.Path)
		fmt.Println("Code:", response.StatusCode)
	}

}

type MetricStorage struct {
	Source   MetricSource
	Metrics  map[string]gauge
	Counters map[string]counter
}

func (mp *MetricStorage) Update() {
	rand.New(rand.NewSource(time.Now().Unix()))
	mp.Metrics = mp.Source.GetMetrics()
	mp.Metrics["RandomValue"] = gauge(rand.Float64())

	count, ok := mp.Counters["PollCount"]
	if !ok {
		count = 0
	}
	count++
	mp.Counters = map[string]counter{
		"PollCount": count,
	}
}

func (mp *MetricStorage) getGauges() map[string]float64 {
	result := map[string]float64{}
	for k, v := range mp.Metrics {
		result[k] = float64(v)
	}
	return result
}
func (mp *MetricStorage) getCounters() map[string]int64 {
	result := map[string]int64{}
	for k, v := range mp.Counters {
		result[k] = int64(v)
	}
	return result
}
