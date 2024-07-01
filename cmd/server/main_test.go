package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateMetrics(t *testing.T) {
	storage := &MemStorage{Gauges: map[string]gauge{}, Counters: map[string]counter{}}
	srv := CreateServer()
	srv.MountHandlers(storage)

	tests := []struct {
		code   int
		method string
		path   string
	}{
		{
			code:   200,
			method: http.MethodPost,
			path:   "/update/gauge/testname/20",
		},
		{
			code:   405,
			method: http.MethodGet,
			path:   "/update/gauge/testname/20",
		},
		{
			code:   404,
			method: http.MethodPost,
			path:   "/update/gauge/20",
		},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			req, _ := http.NewRequest(test.method, test.path, nil)
			resp := executeRequest(req, srv)

			assert.Equal(t, test.code, resp.Code)
		})
	}
}

func executeRequest(req *http.Request, s *Server) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)

	return rr
}
