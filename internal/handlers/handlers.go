package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"ya-prac-project1/internal/logger"
	"ya-prac-project1/internal/metrics"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage interface {
	SetValue(t, name, value string) error
	GetValue(t, name string) (string, error)
	GetMetrics() []*metrics.Metrics
	// GetRows() []string
}

type ServerHandler struct {
	storage Storage
	baseDNS string
	handler *chi.Mux
}

func New(storage Storage, baseDNS string) *ServerHandler {
	s := &ServerHandler{}
	s.storage = storage
	s.baseDNS = baseDNS
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
			http.Error(w, err.Error(), http.StatusNotFound)
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
			w.Header().Set("Content-Type", "text/html")

			metrics := s.storage.GetMetrics()
			rows := make([]string, 0)
			for _, metric := range metrics {
				row := fmt.Sprintf("%s: %s", metric.ID, metric.GetValue())
				rows = append(rows, row)

			}
			w.Write([]byte(getMetricPage(rows)))
		}
	}
}

func (s *ServerHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if s.baseDNS == "" {
		return
	}
	db, err := sql.Open("pgx", s.baseDNS)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = db.PingContext(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
		r.Get("/ping", s.Ping)
		r.Get("/", s.GetMetrics)
		r.Get("/value/{metric_type}/{metric_name}", s.GetMetrics)
		r.Post("/update/{metric_type}/{metric_name}/{metric_value}", s.UpdateMetrics)
		r.Post("/update/", s.UpdateMetrics)
		r.Post("/value/", s.GetMetrics)
	})
	s.handler = router
}

func (s *ServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logMiddleware(zipMiddleware(s.handler)).ServeHTTP(w, r)
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

func zipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportGzip := strings.Contains(acceptEncoding, "gzip")
		if supportGzip {
			w.Header().Set("Content-Encoding", "gzip")
			cw := newZipWriter(w)
			ow = cw
			defer cw.Close()
		}

		contentType := r.Header.Get("Content-Type")
		contentEncoding := r.Header.Get("Content-Encoding")
		sendGzip := strings.Contains(contentEncoding, "gzip")
		if sendGzip && (contentType == "application/json" || contentType == "text/html") {
			cr, err := newZipReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}
		h.ServeHTTP(ow, r)
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
