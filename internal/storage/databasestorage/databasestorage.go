package databasestorage

import (
	"database/sql"
	"errors"
	"fmt"
	"ya-prac-project1/internal/logger"
	"ya-prac-project1/internal/metrics"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

var database *sql.DB

type Storage struct {
}

func NewStorage(dsn string) (*Storage, error) {

	if err := initDB(dsn); err != nil {
		return nil, err
	}
	prepareDB()

	return &Storage{}, nil
}

func (s *Storage) SetValue(metricType, name, value string) error {
	err := checkWrongType(metricType)
	if err != nil {
		return err
	}

	metric := s.GetMetric(metricType, name)
	if metric == nil {
		metric = &metrics.Metrics{
			MType: metricType,
			ID:    name,
		}
		s.AddMetric(metric)
	}

	metric.SetValue(value)
	s.UpdateMetric(metric)
	return nil
}

func (s *Storage) GetValue(metricType, name string) (string, error) {
	err := checkWrongType(metricType)
	if err != nil {
		return "", err
	}

	metric := s.GetMetric(metricType, name)
	if metric == nil {
		return "", errors.New("not found metric")
	}

	return metric.GetValue(), nil
}

func (s *Storage) GetMetrics() []*metrics.Metrics {
	items := []*metrics.Metrics{}
	rows, err := database.Query("SELECT type,name,value,delta FROM metrics")
	if err != nil {
		logger.Get().Info(
			"get metrics error",
			zap.String("error", err.Error()),
		)
	}

	defer rows.Close()
	for rows.Next() {
		var metric metrics.Metrics
		err := rows.Scan(&metric.MType, &metric.ID, &metric.Value, &metric.Delta)
		if err != nil {
			logger.Get().Info(
				"parse metric error",
				zap.String("error", err.Error()),
			)
		}
		items = append(items, &metric)
	}

	if rows.Err() != nil {
		logger.Get().Info(
			"rows error",
			zap.String("error", rows.Err().Error()),
		)
	}

	return items
}

func (s *Storage) GetMetric(metricType, name string) *metrics.Metrics {
	row := database.QueryRow("SELECT type,name,value,delta FROM metrics WHERE type = $1 AND name = $2", metricType, name)

	if row.Err() != nil {
		fmt.Println(row.Err())
	}
	var metric metrics.Metrics
	row.Scan(&metric.MType, &metric.ID, &metric.Value, &metric.Delta)

	if metric.ID == "" {
		return nil
	}

	return &metric
}

func initDB(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("base init fail: %s", err)
	}
	database = db
	return nil
}

func prepareDB() {
	sql := `CREATE TABLE IF NOT EXISTS metrics(
    	name    varchar(255) PRIMARY KEY,
    	type    varchar(40),
    	value    double precision default null,
    	delta    int default null
	);`

	database.Exec(sql)
}

func checkWrongType(metricType string) error {
	if metricType != metrics.MetricTypeGauge &&
		metricType != metrics.MetricTypeCounter {
		return errors.New("wrong type")
	}

	return nil
}

func (s Storage) UpdateMetric(metric *metrics.Metrics) error {
	result, err := database.Exec("UPDATE metrics SET value = $1, delta = $2 WHERE type = $3 AND name = $4", metric.Value, metric.Delta, metric.MType, metric.ID)
	if err != nil {
		return err
	}
	fmt.Println(result.RowsAffected())
	return nil
}

func (s Storage) AddMetric(metric *metrics.Metrics) error {
	result, err := database.Exec("INSERT INTO metrics (type, name, value, delta) VALUES ($1,$2,$3,$4)", metric.MType, metric.ID, metric.Value, metric.Delta)
	if err != nil {
		return err
	}
	fmt.Println(result.RowsAffected())
	return nil
}

func (s *Storage) SetMetrics(metrics []metrics.Metrics) error {

	return nil
}
