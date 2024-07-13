package main

import (
	"fmt"
	"net/http"
	"ya-prac-project1/internal/inmem"

	"github.com/go-chi/chi/v5"
)

type MetricStorage interface {
	SetValue(t, name, value string) error
	GetValue(t, name string) (string, error)
	ToString() string
}

func main() {
	parseFlags()
	if err := run(); err != nil {
		panic(err)
	}
}

func UpdateMetrics(ms MetricStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		mType := chi.URLParam(r, "metric_type")
		mName := chi.URLParam(r, "metric_name")
		mValue := chi.URLParam(r, "metric_value")

		err := ms.SetValue(mType, mName, mValue)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func GetMetrics(ms MetricStorage) http.HandlerFunc {
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
			w.Write([]byte(ms.ToString()))

		}
	}
}

func run() error {
	store := inmem.NewStorage()
	s := CreateServer()
	s.MountHandlers(store)

	fmt.Printf("Start server on: %s\n", serverEndpointFlag)
	return http.ListenAndServe(serverEndpointFlag, s.Router)
}


