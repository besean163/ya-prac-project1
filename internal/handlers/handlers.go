package handlers

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Storage interface {
	SetValue(t, name, value string) error
	GetValue(t, name string) (string, error)
	ToString() string
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
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	mType := chi.URLParam(r, "metric_type")
	mName := chi.URLParam(r, "metric_name")
	mValue := chi.URLParam(r, "metric_value")

	err := s.storage.SetValue(mType, mName, mValue)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
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
			log.Println(err)
			log.Println("here")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write([]byte(v))
	} else {
		w.Write([]byte(s.storage.ToString()))

	}
}
