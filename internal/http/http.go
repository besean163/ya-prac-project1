package http

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

type Server struct {
	Router  *chi.Mux
	storage Storage
}

func NewServer(storage Storage) *Server {
	s := &Server{}
	s.Router = chi.NewRouter()
	s.storage = storage
	return s
}

func (s *Server) mountHandlers() {
	s.Router.Route("/", func(r chi.Router) {
		r.Get("/", s.GetMetrics)
		r.Get("/value/{metric_type}/{metric_name}", s.GetMetrics)
		r.Post("/update/{metric_type}/{metric_name}/{metric_value}", s.GetMetrics)
	})
}

func (s *Server) UpdateMetrics(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) GetMetrics(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "metric_type")
	mName := chi.URLParam(r, "metric_name")

	if mType != "" && mName != "" {
		v, err := s.storage.GetValue(mType, mName)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write([]byte(v))
	} else {
		w.Write([]byte(s.storage.ToString()))

	}
}

func (s *Server) Start(serverEndpointFlag string) error {
	s.mountHandlers()

	log.Printf("Start server on: %s\n", serverEndpointFlag)
	return http.ListenAndServe(serverEndpointFlag, s.Router)
}
