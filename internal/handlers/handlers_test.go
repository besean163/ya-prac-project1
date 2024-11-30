package handlers_test

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"ya-prac-project1/internal/handlers"
	mock "ya-prac-project1/internal/handlers/mocks"
	"ya-prac-project1/internal/logger"
	"ya-prac-project1/internal/metrics"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mock.NewMockMetricService(ctrl)

	value := new(float64)
	*value = 20
	store.EXPECT().SaveMetrics(gomock.Any()).Return(nil).AnyTimes()
	store.EXPECT().SaveMetric(gomock.Any()).Return(nil).AnyTimes()
	store.EXPECT().GetMetric("gauge", "testname").Return(metrics.Metrics{ID: "testname", MType: "gauge", Value: value}, nil).AnyTimes()
	store.EXPECT().GetMetric("gauge", "test_name").Return(metrics.Metrics{ID: "test_name", MType: "gauge", Value: value}, nil).AnyTimes()
	store.EXPECT().GetMetrics().Return([]metrics.Metrics{
		{
			MType: "gauge",
			ID:    "testname",
			Value: value,
		},
	}).AnyTimes()

	h := handlers.New(store, nil, "", "")

	tests := []struct {
		method     string
		path       string
		body       string
		result     string
		checkValue bool
		code       int
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
			body:       `{"id":"test_name","type":"gauge"}`,
			checkValue: true,
			result:     `{"value":20,"id":"test_name","type":"gauge"}`,
		},
		{
			code:       200,
			method:     http.MethodPost,
			path:       "/updates/",
			body:       `[{"id": "test_name","type": "gauge"}]`,
			checkValue: false,
			result:     ``,
		},
		{
			code:       500,
			method:     http.MethodPost,
			path:       "/updates/",
			body:       `[{"id": "test_name","type": "gauge"}`,
			checkValue: false,
			result:     ``,
		},
		{
			code:       500,
			method:     http.MethodGet,
			path:       "/ping",
			checkValue: false,
			result:     ``,
		},
	}

	for _, test := range tests {
		logger.Set()
		t.Run(test.path, func(t *testing.T) {
			var b io.Reader
			if test.body != "" {
				b = bytes.NewReader([]byte(test.body))
			}
			req, _ := http.NewRequest(test.method, test.path, b)
			if test.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			rr := httptest.NewRecorder()

			h.Mount()
			h.ServeHTTP(rr, req)

			assert.Equal(t, test.code, rr.Code)

			if test.checkValue {
				answer, _ := io.ReadAll(rr.Body)
				fmt.Println(string(answer))
				assert.Equal(t, test.result, string(answer))
			}
		})
	}
}

func TestGzipCompression(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mock.NewMockMetricService(ctrl)

	value := new(float64)
	*value = 20
	store.EXPECT().SaveMetrics(gomock.Any()).Return(nil).AnyTimes()
	store.EXPECT().SaveMetric(gomock.Any()).Return(nil).AnyTimes()
	store.EXPECT().GetMetric("gauge", "testname").Return(metrics.Metrics{ID: "testname", MType: "gauge", Value: value}, nil).AnyTimes()
	store.EXPECT().GetMetrics().Return([]metrics.Metrics{
		{
			MType: "gauge",
			ID:    "testname",
			Value: value,
		},
	}).AnyTimes()

	h := handlers.New(store, nil, "", "")

	valueResponse := "20"
	t.Run("value", func(t *testing.T) {

		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(valueResponse))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		req, _ := http.NewRequest(http.MethodGet, "/value/gauge/testname", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rr := httptest.NewRecorder()

		h.Mount()
		h.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		respBody := bytes.NewBuffer(nil)
		zr, err := gzip.NewReader(rr.Body)
		require.NoError(t, err)
		zr.Read(respBody.Bytes())

		b, err := io.ReadAll(zr)
		require.NoError(t, err)
		require.JSONEq(t, valueResponse, string(b))
	})

	allValuesResponse := "<!DOCTYPE html><html><head><title>Report</title></head><body><div>testname: 20</div></body></html>"
	t.Run("all_values", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(allValuesResponse))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rr := httptest.NewRecorder()

		h.Mount()
		h.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		respBody := bytes.NewBuffer(nil)
		zr, err := gzip.NewReader(rr.Body)
		require.NoError(t, err)
		zr.Read(respBody.Bytes())

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.Equal(t, allValuesResponse, string(b))
	})
}
