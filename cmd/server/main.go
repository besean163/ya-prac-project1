package main

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type gauge float64
type counter int64

type MemStorage struct {
	Gauges   map[string]gauge
	Counters map[string]counter
}

func (m *MemStorage) SetValue(t, name, value string) error {
	switch t {
	case "gauge":
		i, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		m.Gauges[name] = gauge(i)
	case "counter":
		i, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		v, ok := m.Counters[name]
		if ok {
			m.Counters[name] = v + counter(i)
		} else {
			m.Counters[name] = counter(i)
		}

	default:
		return errors.New("not correct type")
	}
	return nil
}

func (m *MemStorage) GetValue(t, name string) (string, error) {
	value := ""
	var err error
	switch t {
	case "gauge":
		v, ok := m.Gauges[name]
		if ok {
			value = fmt.Sprint(v)
		}

	case "counter":
		v, ok := m.Counters[name]
		if ok {
			value = fmt.Sprint(v)
		}
	}
	if value == "" {
		err = errors.New("not found metric")
	}
	return value, err
}

func (m *MemStorage) ToStringValues() string {
	result := ""

	for k, v := range m.Gauges {
		result += fmt.Sprintf("%s: %s\n", k, fmt.Sprint(v))
	}

	for k, v := range m.Counters {
		result += fmt.Sprintf("%s: %s\n", k, fmt.Sprint(v))
	}

	return result
}

var availableMetricTypes = []string{
	"gauge",
	"counter",
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func UpdateMetrics(ms *MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		mType := chi.URLParam(r, "metric_type")
		mName := chi.URLParam(r, "metric_name")
		mValue := chi.URLParam(r, "metric_value")

		if !slices.Contains(availableMetricTypes, mType) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err := ms.SetValue(mType, mName, mValue)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func GetMetrics(ms *MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, "metric_type")
		mName := chi.URLParam(r, "metric_name")

		if mType != "" && mName != "" {
			v, err := ms.GetValue(mType, mName)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Write([]byte(v))
		} else {
			w.Write([]byte(ms.ToStringValues()))

		}
	}
}

func run() error {
	store := NewStorage()
	s := CreateServer()
	s.MountHandlers(store)

	return http.ListenAndServe(":8080", s.Router)
}

type Server struct {
	Router *chi.Mux
}

func CreateServer() *Server {
	s := &Server{}
	s.Router = chi.NewRouter()
	return s
}

func (s *Server) MountHandlers(store *MemStorage) {
	s.Router.Route("/", func(r chi.Router) {
		r.Get("/", GetMetrics(store))
		r.Get("/value/{metric_type}/{metric_name}", GetMetrics(store))
		r.Post("/update/{metric_type}/{metric_name}/{metric_value}", UpdateMetrics(store))
	})

}

func NewStorage() *MemStorage {
	return &MemStorage{Gauges: map[string]gauge{}, Counters: map[string]counter{}}
}
