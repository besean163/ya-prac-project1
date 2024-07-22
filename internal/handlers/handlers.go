package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Storage interface {
	SetValue(t, name, value string) error
	GetValue(t, name string) (string, error)
	GetRows() []string
}

type ServerHandler struct {
	storage Storage
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

	mType := chi.URLParam(r, "metric_type")
	mName := chi.URLParam(r, "metric_name")
	mValue := chi.URLParam(r, "metric_value")

	err := s.storage.SetValue(mType, mName, mValue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *ServerHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
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

func getMetricPage(rows []string) string {
	page := `<!DOCTYPE html><html><head><title>Report</title></head><body>`

	for _, row := range rows {
		page += fmt.Sprintf("<div>%s</div>", row)
	}

	page += `</body></html>`
	return page
}

func (s *ServerHandler) GetHandler() http.Handler {
	router := chi.NewRouter()
	router.Route("/", func(r chi.Router) {
		r.Get("/", s.GetMetrics)
		r.Get("/value/{metric_type}/{metric_name}", s.GetMetrics)
		r.Post("/update/{metric_type}/{metric_name}/{metric_value}", s.UpdateMetrics)
	})
	return router
}
