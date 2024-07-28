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
	SetValue(t, name, value string) error
	GetValue(t, name string) (string, error)
	GetRows() []string
}

type ServerHandler struct {
	storage Storage
	handler *chi.Mux
}

func New(storage Storage) *ServerHandler {
	s := &ServerHandler{}
	s.storage = storage
	return s
}

func (s *ServerHandler) UpdateMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if hasJSONHeader(r) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		metric := metrics.Metrics{}
		if err := json.Unmarshal(body, &metric); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		err = s.storage.SetValue(metric.MType, metric.ID, metric.GetValue())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		mType := chi.URLParam(r, "metric_type")
		mName := chi.URLParam(r, "metric_name")
		mValue := chi.URLParam(r, "metric_value")

		err := s.storage.SetValue(mType, mName, mValue)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

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
		v, err := s.storage.GetValue(metric.MType, metric.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		metric.SetValue(v)

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
			v, err := s.storage.GetValue(mType, mName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			w.Write([]byte(v))
		} else {
			w.Write([]byte(getMetricPage(s.storage.GetRows())))
		}
	}
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
		r.Get("/", s.GetMetrics)
		r.Get("/value/{metric_type}/{metric_name}", s.GetMetrics)
		r.Post("/update/{metric_type}/{metric_name}/{metric_value}", s.UpdateMetrics)
		r.Post("/update/", s.UpdateMetrics)
		r.Post("/value/", s.GetMetrics)
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
