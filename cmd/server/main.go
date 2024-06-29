package main

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
)

type gauge float64
type counter int64

var store *MemStorage

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
		m.Gauges[name] = m.Gauges[name] + gauge(i)
	}
	return nil
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

func UpdateMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	args := strings.Split(r.URL.Path, "/")

	if len(args) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(args) < 5 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if args[1] != "update" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	t := args[2]
	if !slices.Contains(availableMetricTypes, t) {
		w.WriteHeader(http.StatusBadRequest)
	}
	name := args[3]
	value := args[4]

	err := store.SetValue(t, name, value)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)
}

func run() error {
	store = new(MemStorage)
	store.Gauges = map[string]gauge{}
	store.Counters = map[string]counter{}

	m := http.NewServeMux()
	m.HandleFunc("/", UpdateMetrics)

	return http.ListenAndServe(":8080", m)
}
