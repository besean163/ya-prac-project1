// Package handlers предоставляет работу по маршрутизации входящих запросов
package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"ya-prac-project1/internal/metrics"

	"github.com/go-chi/chi/v5"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// MetricService представляет интерфейс сервиса работы с метриками
type MetricService interface {
	GetMetric(metricType, name string) (metrics.Metrics, error)
	GetMetrics() []metrics.Metrics
	SaveMetric(m metrics.Metrics) error
	SaveMetrics(ms []metrics.Metrics) error
}

// ServerHandler представляет структуру сервера
type ServerHandler struct {
	metricService MetricService
	database      *sql.DB
	handler       *chi.Mux
	hashKey       string
}

// New создает новый экземпляр сервера
func New(metricService MetricService, db *sql.DB, hashKey string) *ServerHandler {
	s := &ServerHandler{}
	s.metricService = metricService
	s.database = db
	s.hashKey = hashKey
	return s
}

// UpdateMetrics обновляет метрики в привязаном сервисе метрик
func (s *ServerHandler) UpdateMetrics(w http.ResponseWriter, r *http.Request) {
	if hasJSONHeader(r) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		metric := metrics.Metrics{}
		if err := json.Unmarshal(body, &metric); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		err = s.metricService.SaveMetric(metric)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	} else {
		mType := chi.URLParam(r, "metric_type")
		mName := chi.URLParam(r, "metric_name")
		mValue := chi.URLParam(r, "metric_value")
		metric := metrics.Metrics{
			MType: mType,
			ID:    mName,
		}
		err := metric.SetValue(mValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = s.metricService.SaveMetric(metric)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// UpdateBatchMetrics обновляет пук метрик в привязаном сервисе метрик
func (s *ServerHandler) UpdateBatchMetrics(w http.ResponseWriter, r *http.Request) {
	if hasJSONHeader(r) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		metrics := []metrics.Metrics{}
		if err := json.Unmarshal(body, &metrics); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		err = s.metricService.SaveMetrics(metrics)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// GetMetrics получает запрашиваемые метрики из сервиса метрик
func (s *ServerHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && hasJSONHeader(r) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		metric := metrics.Metrics{}
		if err := json.Unmarshal(body, &metric); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		metric, err = s.metricService.GetMetric(metric.MType, metric.ID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		body, err = json.Marshal(metric)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	} else {
		mType := chi.URLParam(r, "metric_type")
		mName := chi.URLParam(r, "metric_name")
		if mType != "" && mName != "" {
			metric, err := s.metricService.GetMetric(mType, mName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			w.Write([]byte(metric.GetValue()))
		} else {
			w.Header().Set("Content-Type", "text/html")

			metrics := s.metricService.GetMetrics()
			rows := make([]string, 0)
			for _, metric := range metrics {
				row := fmt.Sprintf("%s: %s", metric.ID, metric.GetValue())
				rows = append(rows, row)

			}
			w.Write([]byte(getMetricPage(rows)))
		}
	}
}

// Ping тестовый роут на проверку подключения к бд
func (s *ServerHandler) Ping(w http.ResponseWriter, r *http.Request) {
	var err error
	if s.database == nil {
		err = errors.New("no database connection")
	}

	if err == nil {
		err = s.database.PingContext(context.Background())
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getMetricPage выводит страницу со всеми имеющимися метриками
func getMetricPage(rows []string) string {
	page := `<!DOCTYPE html><html><head><title>Report</title></head><body>`

	for _, row := range rows {
		page += fmt.Sprintf("<div>%s</div>", row)
	}

	page += `</body></html>`
	return page
}

// Mount монтирует маршруты на сервер
func (s *ServerHandler) Mount() {
	router := chi.NewRouter()
	router.Route("/", func(r chi.Router) {
		r.Get("/ping", s.Ping)
		r.Get("/", s.GetMetrics)
		r.Get("/value/{metric_type}/{metric_name}", s.GetMetrics)
		r.Post("/update/{metric_type}/{metric_name}/{metric_value}", s.UpdateMetrics)
		r.Post("/update/", s.UpdateMetrics)
		r.Post("/value/", s.GetMetrics)
		r.Post("/updates/", s.UpdateBatchMetrics)
	})
	s.handler = router
}

// ServeHTTP обрабатывает входящий запрос
func (s *ServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	h := logMiddleware(zipMiddleware(s.handler))

	if s.hashKey != "" {
		h = hashKeyMiddleware(h, s.hashKey)
	}

	h.ServeHTTP(w, r)
}

func hasJSONHeader(r *http.Request) bool {
	for header, values := range r.Header {
		if header != "Content-Type" {
			continue
		}
		for _, value := range values {
			if value == "application/json" {
				return true
			}
		}
	}
	return false
}
