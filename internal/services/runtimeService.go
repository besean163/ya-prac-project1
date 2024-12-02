package services

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	random "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"
	"ya-prac-project1/internal/logger"
	"ya-prac-project1/internal/metrics"

	pb "ya-prac-project1/internal/services/proto"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

func (s *RuntimeService) RunSendRequest(requestCh chan *http.Request, serverEndpoint string, key string, cryptoKey string) {
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

	buf = encryptMessage(buf, cryptoKey)
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

	cIP, err := getCurrentIP()
	if err != nil {
		logger.Get().Info("can't create request. Error: %s\n", zap.String("error", err.Error()))
		return
	}
	req.Header.Set("X-Real-IP", cIP)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	requestCh <- req
}

func (s *RuntimeService) RunSendgRPCRequest(serverEndpoint string) {
	metrics := s.storage.GetMetrics()

	targetMetrics := make([]*pb.Metric, 0)
	for _, m := range metrics {
		tm := pb.Metric{
			Type: m.MType,
			Id:   m.ID,
		}
		if m.Value != nil {
			tm.Value = m.Value
		}
		if m.Delta != nil {
			tm.Delta = m.Delta
		}
		targetMetrics = append(targetMetrics, &tm)
	}

	conn, err := grpc.NewClient(serverEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	c := pb.NewMetricSaverServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	var request pb.SaveMetricsRequest
	request.Metrics = targetMetrics
	_, err = c.UpdateMetrics(ctx, &request)
	if err != nil {
		logger.Get().Info("fail send grps reauest", zap.String("error", err.Error()))
	}

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

func encryptMessage(buf bytes.Buffer, cryptoKey string) bytes.Buffer {
	if cryptoKey == "" {
		return buf
	}

	pubKeyBytes, err := os.ReadFile(cryptoKey)
	if err != nil {
		fmt.Printf("can't read public key. Error: %s\n", err)
		return buf
	}

	block, _ := pem.Decode(pubKeyBytes)
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		fmt.Printf("failed to decode PEM block containing public key")
		return buf
	}

	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		fmt.Printf("can't parse public key. Error: %s\n", err)
		return buf
	}

	encryptedMessage, err := rsa.EncryptPKCS1v15(random.Reader, publicKey, buf.Bytes())
	if err != nil {
		fmt.Printf("can't encrypt message")
		return buf
	}

	return *bytes.NewBuffer(encryptedMessage)
}

func getCurrentIP() (string, error) {
	// Получаем список всех сетевых интерфейсов
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("error getting network interfaces: %v", err)
	}

	// Проходим по интерфейсам и находим IPv4-адрес
	for _, iface := range interfaces {
		// Пропускаем интерфейсы, которые не активны
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Получаем адреса интерфейса
		addrs, err := iface.Addrs()
		if err != nil {
			log.Printf("Error getting addresses for interface %v: %v", iface.Name, err)
			continue
		}

		// Проходим по адресам
		for _, addr := range addrs {
			// Проверяем, что это IPv4-адрес
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("no active IPv4 interface found")
}
