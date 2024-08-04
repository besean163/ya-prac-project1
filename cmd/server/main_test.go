package main

import (
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

func (mock *StorageMock) GetMetrics() []*metrics.Metrics {
	v := new(float64)
	*v = 20
	return []*metrics.Metrics{
		{
			ID:    "testname",
			MType: "gauge",
			Value: v,
		},
	}
}

func (mock StorageMock) GetRows() []string {
	return []string{"testname: 20"}
}

func TestUpdateMetrics(t *testing.T) {
	store := StorageMock{}
	h := handlers.New(&store, "")

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
			path:       "/update/gauge/testname/20",
			checkValue: false,
			result:     "",
		},
		{
			code:       405,
			method:     http.MethodGet,
			path:       "/update/gauge/testname/20",
			checkValue: false,
			result:     "",
		},
		{
			code:       404,
			method:     http.MethodPost,
			path:       "/update/gauge/20",
			checkValue: false,
			result:     "",
		},
		{
			code:       200,
			method:     http.MethodGet,
			path:       "/value/gauge/testname",
			checkValue: true,
			result:     "20",
		},
		{
			code:       200,
			method:     http.MethodGet,
			path:       "/",
			checkValue: true,
			result:     `<!DOCTYPE html><html><head><title>Report</title></head><body><div>testname: 20</div></body></html>`,
		},
		{
			code:       200,
			method:     http.MethodPost,
			path:       "/update/",
			body:       `{"id": "test_name","type": "gauge","value": 20}`,
			checkValue: false,
			result:     "",
		},
		{
			code:       200,
			method:     http.MethodPost,
			path:       "/value/",
			body:       `{"id": "test_name","type": "gauge"}`,
			checkValue: false,
			result:     `{"id": "test_name","type": "gauge","value": 20}`,
		},
	}

	for _, test := range tests {
		logger.Set()
		t.Run(test.path, func(t *testing.T) {
			req, _ := http.NewRequest(test.method, test.path, nil)
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
