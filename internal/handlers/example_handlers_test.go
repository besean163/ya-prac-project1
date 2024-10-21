package handlers_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"ya-prac-project1/internal/handlers"
	"ya-prac-project1/internal/logger"
	"ya-prac-project1/internal/metrics"
	"ya-prac-project1/internal/services"
	"ya-prac-project1/internal/storage/inmemstorage"
)

func ExampleServerHandler_GetMetrics() {
	storage := inmemstorage.NewStorage()
	storage.SetMetrics([]metrics.Metrics{
		metrics.NewMetric("testname", metrics.MetricTypeGauge, "20"),
	})
	service := services.NewMetricSaverService(storage)
	logger.Set()
	h := handlers.New(service, nil, "")
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	h.Mount()
	h.ServeHTTP(rr, req)

	answer, _ := io.ReadAll(rr.Body)
	fmt.Println(string(answer))

	// Output:
	// <!DOCTYPE html><html><head><title>Report</title></head><body><div>testname: 20</div></body></html>
}

func ExampleServerHandler_GetMetrics_second() {
	storage := inmemstorage.NewStorage()
	storage.SetMetrics([]metrics.Metrics{
		metrics.NewMetric("testname", metrics.MetricTypeGauge, "20"),
	})
	service := services.NewMetricSaverService(storage)
	logger.Set()
	h := handlers.New(service, nil, "")
	req, _ := http.NewRequest(http.MethodGet, "/value/gauge/testname", nil)
	rr := httptest.NewRecorder()

	h.Mount()
	h.ServeHTTP(rr, req)

	answer, _ := io.ReadAll(rr.Body)
	fmt.Println(string(answer))

	// Output:
	// 20
}

func ExampleServerHandler_GetMetrics_third() {
	storage := inmemstorage.NewStorage()
	storage.SetMetrics([]metrics.Metrics{
		metrics.NewMetric("testname", metrics.MetricTypeGauge, "20"),
	})
	service := services.NewMetricSaverService(storage)
	logger.Set()
	h := handlers.New(service, nil, "")
	body := bytes.NewReader([]byte(`{"id":"testname","type":"gauge"}`))
	req, _ := http.NewRequest(http.MethodPost, "/value/", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Mount()
	h.ServeHTTP(rr, req)

	answer, _ := io.ReadAll(rr.Body)
	fmt.Println(string(answer))

	// Output:
	// {"id":"testname","type":"gauge","value":20}
}

func ExampleServerHandler_UpdateMetrics() {
	storage := inmemstorage.NewStorage()
	service := services.NewMetricSaverService(storage)
	logger.Set()
	h := handlers.New(service, nil, "")
	h.Mount()

	req, _ := http.NewRequest(http.MethodPost, "/update/gauge/testname/20", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	req, _ = http.NewRequest(http.MethodGet, "/value/gauge/testname", nil)
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	answer, _ := io.ReadAll(rr.Body)
	fmt.Println(string(answer))

	// Output:
	// 20
}

func ExampleServerHandler_UpdateMetrics_second() {
	storage := inmemstorage.NewStorage()
	service := services.NewMetricSaverService(storage)
	logger.Set()
	h := handlers.New(service, nil, "")
	h.Mount()

	body := bytes.NewReader([]byte(`{"id": "testname","type": "gauge","value": 20}`))
	req, _ := http.NewRequest(http.MethodPost, "/update/", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	req, _ = http.NewRequest(http.MethodGet, "/value/gauge/testname", nil)
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	answer, _ := io.ReadAll(rr.Body)
	fmt.Println(string(answer))

	// Output:
	// 20
}

func ExampleServerHandler_UpdateBatchMetrics_second() {
	storage := inmemstorage.NewStorage()
	service := services.NewMetricSaverService(storage)
	logger.Set()
	h := handlers.New(service, nil, "")
	h.Mount()

	body := bytes.NewReader([]byte(`[{"id": "testname","type": "gauge","value": 20},{"id": "testname2","type": "gauge","value": 30}]`))
	req, _ := http.NewRequest(http.MethodPost, "/updates/", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	req, _ = http.NewRequest(http.MethodGet, "/value/gauge/testname", nil)
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	answer, _ := io.ReadAll(rr.Body)
	fmt.Println(string(answer))

	req, _ = http.NewRequest(http.MethodGet, "/value/gauge/testname2", nil)
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	answer, _ = io.ReadAll(rr.Body)
	fmt.Println(string(answer))

	// Output:
	// 20
	// 30
}
