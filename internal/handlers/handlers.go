package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"ya-prac-project1/internal/logger"
	"ya-prac-project1/internal/metrics"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Storage interface {
	GetMetric(metricType, name string) *metrics.Metrics
	UpdateMetric(metric *metrics.Metrics) error
	GetRows() []string
}

type ServerHandler struct {
	storage Storage
	handler *chi.Mux
}

type LogData struct {
	URI    string
	Method string
	Status int
	Size   int
}

type LogResponse struct {
	http.ResponseWriter
	Data LogData
}

func (r *LogResponse) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.Data.Size += size
	return size, err
}

func (r *LogResponse) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.Data.Status = statusCode
}

func New(storage Storage) *ServerHandler {
	s := &ServerHandler{}
	s.storage = storage
	return s
}

func (s *ServerHandler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	metric := &metrics.Metrics{}
	err = json.Unmarshal(body, metric)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = s.storage.UpdateMetric(metric)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (s *ServerHandler) GetMetric(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	metric := &metrics.Metrics{}
	err = json.Unmarshal(body, metric)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	existMetric := s.storage.GetMetric(metric.MType, metric.ID)
	if existMetric == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	b, err := json.Marshal(existMetric)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *ServerHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(getMetricPage(s.storage.GetRows())))
	w.WriteHeader(http.StatusOK)
}

func getMetricPage(rows []string) string {
	page := `<!DOCTYPE html><html><head><title>Report</title></head><body>`

	for _, row := range rows {
		page += fmt.Sprintf("<div>%s</div>", row)
	}

	page += `</body></html>`
	return page
}

func (s *ServerHandler) Mount() {
	router := chi.NewRouter()
	router.Route("/", func(r chi.Router) {
		r.Get("/", s.GetAllMetrics)
		r.Post("/value/", s.GetMetric)
		r.Post("/update/", s.UpdateMetric)
	})
	s.handler = router
}

func (s *ServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logMiddleware(s.handler).ServeHTTP(w, r)
}

func logMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ld := LogData{}
		lResponseWriter := &LogResponse{
			ResponseWriter: w,
			Data:           ld,
		}
		start := time.Now()
		h.ServeHTTP(lResponseWriter, r)
		duration := time.Since(start)

		lResponseWriter.Data.Method = r.Method
		lResponseWriter.Data.URI = r.RequestURI

		logger.Get().Info(
			"get request",
			zap.String("method", lResponseWriter.Data.Method),
			zap.String("uri", lResponseWriter.Data.URI),
			zap.Int("status", lResponseWriter.Data.Status),
			zap.Int("size", lResponseWriter.Data.Size),
			zap.Duration("time", duration),
		)
	})
}
