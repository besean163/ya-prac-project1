package services

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strings"
	"time"
	"ya-prac-project1/internal/metrics"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"golang.org/x/sync/errgroup"
)

const (
	retryAttempts    = 3
	waitSec          = 1
	waitSecIncrement = 2
)

type Storage interface {
	SetValue(metricType, name, value string) error
	GetValue(metricType, name string) (string, error)
	GetMetrics() []metrics.Metrics
	SetMetric(metric *metrics.Metrics) error
}

type RuntimeService struct {
	storage Storage
}

func NewRuntimeService(storage Storage) RuntimeService {
	return RuntimeService{storage: storage}
}

func (service RuntimeService) Run(errGroup *errgroup.Group, poolInterval int) {

	metricsCh := make(chan metrics.Metrics, 10)

	errGroup.Go(func() error {
		return updateRuntimeMetrics(metricsCh, poolInterval)
	})

	errGroup.Go(func() error {
		return updateExtraRuntimeMetrics(metricsCh, poolInterval)
	})

	errGroup.Go(func() error {
		for {
			metric := <-metricsCh
			err := service.storage.SetMetric(&metric)
			if err != nil {
				return err
			}
		}
	})

}

func updateRuntimeMetrics(metricsCh chan metrics.Metrics, poolInterval int) error {
	for {
		rMetrics := getRuntimeMetrics()

		for name, value := range rMetrics {
			metric := metrics.Metrics{
				MType: metrics.MetricTypeGauge,
				ID:    name,
			}
			err := metric.SetValue(value)
			if err != nil {
				return err
			}

			metricsCh <- metric
		}

		pcm, err := getPoolCountMetric()
		if err != nil {
			return err
		}
		metricsCh <- pcm

		rm, err := getRandomValueMetirc()
		if err != nil {
			return err
		}

		metricsCh <- rm

		time.Sleep(time.Duration(poolInterval) * time.Second)
	}
}

func getPoolCountMetric() (metrics.Metrics, error) {
	m := metrics.Metrics{
		ID:    "PollCount",
		MType: "counter",
	}
	err := m.SetValue(fmt.Sprint(1))
	if err != nil {
		return m, err
	}
	return m, nil
}

func getRandomValueMetirc() (metrics.Metrics, error) {
	m := metrics.Metrics{
		ID:    "RandomValue",
		MType: "gauge",
	}

	rand.New(rand.NewSource(time.Now().Unix()))

	err := m.SetValue(fmt.Sprint(rand.Float64()))
	if err != nil {
		return m, err
	}
	return m, nil
}

func updateExtraRuntimeMetrics(metricsCh chan metrics.Metrics, poolInterval int) error {
	for {
		rMetrics, err := getExtraRuntimeMetrics()
		if err != nil {
			return err
		}
		for name, value := range rMetrics {
			metric := metrics.Metrics{
				MType: "gauge",
				ID:    name,
			}
			err := metric.SetValue(value)
			if err != nil {
				return err
			}

			metricsCh <- metric
		}

		time.Sleep(time.Duration(poolInterval) * time.Second)
	}
}

func (s *RuntimeService) UpdateMetrics() {
	metrics := getRuntimeMetrics()
	for name, value := range metrics {
		s.storage.SetValue("gauge", name, value)
	}

	rand.New(rand.NewSource(time.Now().Unix()))
	s.storage.SetValue("gauge", "RandomValue", fmt.Sprint(rand.Float64()))
	s.storage.SetValue("counter", "PollCount", fmt.Sprint(1))
}

func (s *RuntimeService) RunSendRequest(requestCh chan *http.Request, serverEndpoint string, key string) {
	metrics := s.storage.GetMetrics()

	b, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("encode error. Error: %s\n", err)
		return
	}

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err = gw.Write(b); err != nil {
		log.Printf("compress error. Error: %s\n", err)
		return
	}
	gw.Close()

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/updates/", serverEndpoint), &buf)
	if err != nil {
		fmt.Printf("can't create request. Error: %s\n", err)
		return
	}

	if key != "" {
		h := hmac.New(sha256.New, []byte(key))
		h.Write(buf.Bytes())
		sign := hex.EncodeToString(h.Sum(nil))
		req.Header.Set("HashSHA256", sign)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	requestCh <- req
}

func getRuntimeMetrics() map[string]string {
	stat := runtime.MemStats{}
	runtime.ReadMemStats(&stat)
	return map[string]string{
		"Alloc":         fmt.Sprint(stat.Alloc),
		"BuckHashSys":   fmt.Sprint(stat.BuckHashSys),
		"Frees":         fmt.Sprint(stat.Frees),
		"GCSys":         fmt.Sprint(stat.GCSys),
		"HeapAlloc":     fmt.Sprint(stat.HeapAlloc),
		"HeapIdle":      fmt.Sprint(stat.HeapIdle),
		"HeapInuse":     fmt.Sprint(stat.HeapInuse),
		"HeapObjects":   fmt.Sprint(stat.HeapObjects),
		"HeapReleased":  fmt.Sprint(stat.HeapReleased),
		"LastGC":        fmt.Sprint(stat.LastGC),
		"Lookups":       fmt.Sprint(stat.Lookups),
		"MCacheInuse":   fmt.Sprint(stat.MCacheInuse),
		"MCacheSys":     fmt.Sprint(stat.MCacheSys),
		"MSpanInuse":    fmt.Sprint(stat.MSpanInuse),
		"MSpanSys":      fmt.Sprint(stat.MSpanSys),
		"Mallocs":       fmt.Sprint(stat.Mallocs),
		"NextGC":        fmt.Sprint(stat.NextGC),
		"NumForcedGC":   fmt.Sprint(stat.NumForcedGC),
		"NumGC":         fmt.Sprint(stat.NumGC),
		"OtherSys":      fmt.Sprint(stat.OtherSys),
		"PauseTotalNs":  fmt.Sprint(stat.PauseTotalNs),
		"StackInuse":    fmt.Sprint(stat.StackInuse),
		"Sys":           fmt.Sprint(stat.Sys),
		"TotalAlloc":    fmt.Sprint(stat.TotalAlloc),
		"StackSys":      fmt.Sprint(stat.StackSys),
		"HeapSys":       fmt.Sprint(stat.HeapSys),
		"GCCPUFraction": fmt.Sprint(stat.GCCPUFraction),
	}
}

func getExtraRuntimeMetrics() (map[string]string, error) {
	cpuUtilization, err := cpu.Counts(false)
	if err != nil {
		return nil, err
	}
	virtMem, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"CPUutilization1": fmt.Sprint(cpuUtilization),
		"TotalMemory":     fmt.Sprint(virtMem.Total),
		"FreeMemory":      fmt.Sprint(virtMem.Free),
	}, nil
}

// возвращает функцию которая следит за количеством повторов и определяет их надобность
// работает за счет замыкания, т.е. передаем в параметры создающей функции количество попыток и каждую следующую заддержку и эти параметры используем при каждом вызове функции
func getRetryFunc(attempts, secDelta, waitDelta int) func(err error) bool {
	attempt := 0
	return func(err error) bool {
		attempt++

		// первый запуск
		if attempt == 1 && attempt <= attempts {
			return true
		}

		if err == nil {
			return false
		}

		if strings.Contains(err.Error(), "connection refused") {
			time.Sleep(time.Duration(secDelta) * time.Second)
			secDelta += waitDelta
			return attempt <= attempts
		}

		// если дошли сюда, то попытки закончились
		return false
	}
}

func (s *RuntimeService) SendMetrics(serverEndpoint string, key string) {
	metrics := s.storage.GetMetrics()
	client := http.Client{}

	b, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("encode error. Error: %s\n", err)
		return
	}

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err = gw.Write(b); err != nil {
		log.Printf("compress error. Error: %s\n", err)
		return
	}
	gw.Close()

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/updates/", serverEndpoint), &buf)
	if err != nil {
		fmt.Printf("can't create request. Error: %s\n", err)
		return
	}

	if key != "" {
		h := hmac.New(sha256.New, []byte(key))
		h.Write(buf.Bytes())
		sign := hex.EncodeToString(h.Sum(nil))
		req.Header.Set("HashSHA256", sign)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	retry := getRetryFunc(retryAttempts, waitSec, waitSecIncrement)
	var response *http.Response
	for retry(err) {
		response, err = client.Do(req)
		if err == nil {
			defer response.Body.Close()
		}
	}

	if err != nil {
		log.Printf("call error. Error: %s\n", err)
		return
	}

	if response.StatusCode != http.StatusOK {
		log.Println("Error write metrics")
		log.Println("Path:", req.URL.Path)
		log.Println("Code:", response.StatusCode)
	}

}
