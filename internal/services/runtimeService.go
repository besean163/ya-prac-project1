package services

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"
	"ya-prac-project1/internal/logger"
	"ya-prac-project1/internal/metrics"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"
)

// Storage структура представляющая интерфейс репозитория для работы с сервисом RuntimeService
type Storage interface {
	GetMetrics() []metrics.Metrics
	SetMetrics(metrics []metrics.Metrics)
}

// RuntimeService структура представляющая сервис для получения рантайм метрик
type RuntimeService struct {
	storage Storage
}

// NewRuntimeService создает сервис
func NewRuntimeService(storage Storage) RuntimeService {
	return RuntimeService{storage: storage}
}

// Run запускает работу сервиса
func (s RuntimeService) Run(ctx context.Context, poolInterval int) {
	go s.updateRuntimeMetrics(ctx, poolInterval)
}

func (s RuntimeService) updateRuntimeMetrics(ctx context.Context, poolInterval int) {
	ticker := time.NewTicker(time.Duration(poolInterval) * time.Second)
	for {
		select {
		case <-ctx.Done():
			logger.Get().Info("updateRuntimeMetrics stopped")
			return
		case <-ticker.C:
			rMetrics := getRuntimeMetrics()

			pcm, err := getPoolCountMetric()
			if err != nil {
				logger.Get().Info("updateRuntimeMetrics getPoolCountMetric error", zap.String("error", err.Error()))
				continue
			}
			rMetrics = append(rMetrics, pcm)

			rm, err := getRandomValueMetirc()
			if err != nil {
				logger.Get().Info("updateRuntimeMetrics getRandomValueMetirc error", zap.String("error", err.Error()))
				continue
			}
			rMetrics = append(rMetrics, rm)

			s.storage.SetMetrics(rMetrics)
		}
	}
}

func getPoolCountMetric() (metrics.Metrics, error) {
	m := metrics.Metrics{ID: "PollCount", MType: metrics.MetricTypeCounter}
	err := m.SetValue(fmt.Sprint(1))
	if err != nil {
		return m, err
	}
	return m, nil
}

func getRandomValueMetirc() (metrics.Metrics, error) {
	m := metrics.Metrics{ID: "RandomValue", MType: metrics.MetricTypeGauge}
	rand.New(rand.NewSource(time.Now().Unix()))
	err := m.SetValue(fmt.Sprint(rand.Float64()))
	if err != nil {
		return m, err
	}
	return m, nil
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

func getRuntimeMetrics() []metrics.Metrics {
	// основные метрики
	stat := runtime.MemStats{}
	runtime.ReadMemStats(&stat)
	m := map[string]string{
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

	// дополнительный метрики
	cpuUtilization, err := cpu.Counts(false)
	if err != nil {
		logger.Get().Info("can't get cpuUtilization. skip")
	} else {
		m["CPUutilization1"] = fmt.Sprint(cpuUtilization)
		virtMem, err := mem.VirtualMemory()
		if err != nil {
			logger.Get().Info("can't get virtMem. skip")
		} else {
			m["TotalMemory"] = fmt.Sprint(virtMem.Total)
			m["FreeMemory"] = fmt.Sprint(virtMem.Free)
		}
	}

	items := []metrics.Metrics{}
	for id, value := range m {
		metric := metrics.Metrics{ID: id, MType: metrics.MetricTypeGauge}
		err := metric.SetValue(value)
		if err != nil {
			logger.Get().Info("can't set runtime metric value. skip", zap.String("error", err.Error()))
			continue
		}
		items = append(items, metric)
	}

	return items
}
