package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"ya-prac-project1/internal/handlers"
	"ya-prac-project1/internal/logger"
	"ya-prac-project1/internal/metrics"

	"github.com/stretchr/testify/assert"
)

type StorageMock struct {
}

func (mock *StorageMock) SetValue(t, name, value string) error {
	return nil
}

func (mock *StorageMock) GetValue(t, name string) (string, error) {
	return "20", nil
}

func (mock StorageMock) GetRows() []string {
	return []string{"testname: 20"}
}

func (mock StorageMock) GetMetric(metricType, name string) *metrics.Metrics {
	value := float64(20)
	return &metrics.Metrics{
		MType: metrics.MetricTypeGauge,
		ID:    "testname",
		Value: &value,
	}
}
func (mock StorageMock) UpdateMetric(metric *metrics.Metrics) error {
	return nil
}

func TestUpdateMetrics(t *testing.T) {
	store := StorageMock{}
	h := handlers.New(&store)

	tests := []struct {
		code       int
		method     string
		path       string
		body       string
		checkValue bool
		result     string
	}{
		{
			code:       200,
			method:     http.MethodPost,
			path:       "/update",
			body:       `{"id":"testname","type":"gauge","value":20}`,
			checkValue: false,
			result:     "",
		},
		{
			code:       405,
			method:     http.MethodGet,
			path:       "/update",
			body:       `{"id":"testname","type":"gauge","value":20}`,
			checkValue: false,
			result:     "",
		},
		{
			code:       200,
			method:     http.MethodGet,
			path:       "/value",
			body:       `{"id":"testname","type":"gauge"}`,
			checkValue: true,
			result:     `{"id":"testname","type":"gauge","value":20}`,
		},
		{
			code:       200,
			method:     http.MethodGet,
			path:       "/",
			checkValue: true,
			result:     `<!DOCTYPE html><html><head><title>Report</title></head><body><div>testname: 20</div></body></html>`,
		},
	}

	for _, test := range tests {
		logger.Set()
		t.Run(test.path, func(t *testing.T) {
			r := bytes.NewReader([]byte(test.body))
			req, _ := http.NewRequest(test.method, test.path, r)
			rr := httptest.NewRecorder()

			h.Mount()
			h.ServeHTTP(rr, req)

			assert.Equal(t, test.code, rr.Code)

			if test.checkValue {
				answer, _ := io.ReadAll(rr.Body)
				assert.Equal(t, test.result, string(answer))
			}
		})
	}
}
