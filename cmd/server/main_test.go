package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateMetrics(t *testing.T) {
	tests := []struct {
		code    int
		method  string
		path    string
		storage *MemStorage
	}{
		{
			code:    200,
			method:  http.MethodPost,
			path:    "/update/gauge/testname/20",
			storage: &MemStorage{Gauges: map[string]gauge{}, Counters: map[string]counter{}},
		},
		{
			code:    405,
			method:  http.MethodGet,
			path:    "/update/gauge/testname/20",
			storage: &MemStorage{Gauges: map[string]gauge{}, Counters: map[string]counter{}},
		},
		{
			code:    404,
			method:  http.MethodPost,
			path:    "/update/gauge/20",
			storage: &MemStorage{Gauges: map[string]gauge{}, Counters: map[string]counter{}},
		},
		{
			code:    400,
			method:  http.MethodPost,
			path:    "/",
			storage: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			r := httptest.NewRequest(test.method, test.path, nil)
			w := httptest.NewRecorder()

			UpdateMetrics(test.storage)(w, r)

			assert.Equal(t, test.code, w.Code)
		})
	}
}
