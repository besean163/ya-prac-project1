package databasestorage

import (
	"database/sql"
	"fmt"
	"ya-prac-project1/internal/metrics"
)

var database *sql.DB

type Storage struct {
}

func NewStorage(dsn string) (*Storage, error) {
	initDB(dsn)
	prepare()

	return &Storage{}, nil
}

func (s *Storage) SetValue(metricType, name, value string) error {
	return nil
}

func (s *Storage) GetValue(metricType, name string) (string, error) {
	return "", nil
}

func (s *Storage) GetMetrics() []*metrics.Metrics {
	return []*metrics.Metrics{}
}

func initDB(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("base init fail: %s", err)
	}
	database = db
	return nil
}

func prepare() {

}
