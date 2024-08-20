package databasestorage

import (
	"database/sql"
	"errors"
	"time"
	"ya-prac-project1/internal/logger"
	"ya-prac-project1/internal/metrics"

	"github.com/jackc/pgerrcode"
	pgx "github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

var database *sql.DB

type Storage struct {
}

func NewStorage(db *sql.DB) (*Storage, error) {
	var err error
	if db == nil {
		err = errors.New("no database connection")
	}
	if err != nil {
		return nil, err
	}
	database = db
	err = prepareDB()
	if err != nil {
		return nil, err
	}

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

func (s *Storage) GetMetrics() []metrics.Metrics {
	items := []metrics.Metrics{}

	retry := getRetryFunc(3, 2)

	var err error
	var rows *sql.Rows
	for retry(err) {
		rows, err = database.Query("SELECT type,name,value,delta FROM metrics")
	}

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
		items = append(items, metric)
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
		logger.Get().Debug("get metric error", zap.String("error", row.Err().Error()))
	}
	var metric metrics.Metrics
	err := row.Scan(&metric.MType, &metric.ID, &metric.Value, &metric.Delta)
	if err != nil {
		logger.Get().Debug("get metric scan error:" + err.Error())
	}

	if metric.ID == "" {
		return nil
	}

	return &metric
}

func prepareDB() error {
	sql := `CREATE TABLE IF NOT EXISTS metrics(
    	name    varchar(255) PRIMARY KEY,
    	type    varchar(40),
    	value    double precision default null,
    	delta    bigint default null
	);`

	_, err := database.Exec(sql)
	if err != nil {
		return err
	}
	return nil
}

func checkWrongType(metricType string) error {
	if metricType != metrics.MetricTypeGauge &&
		metricType != metrics.MetricTypeCounter {
		return errors.New("wrong type")
	}

	return nil
}

func (s Storage) UpdateMetric(metric *metrics.Metrics) error {
	retry := getRetryFunc(3, 2)

	var err error
	for retry(err) {
		_, err = database.Exec("UPDATE metrics SET value = $1, delta = $2 WHERE type = $3 AND name = $4", metric.Value, metric.Delta, metric.MType, metric.ID)
	}

	if err != nil {
		return err
	}

	return nil
}

func (s Storage) AddMetric(metric *metrics.Metrics) error {

	retry := getRetryFunc(3, 2)

	var err error
	for retry(err) {
		_, err = database.Exec("INSERT INTO metrics (type, name, value, delta) VALUES ($1,$2,$3,$4)", metric.MType, metric.ID, metric.Value, metric.Delta)
	}

	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) SetMetrics(metrics []metrics.Metrics) error {
	existMetrics := s.GetMetrics()

	found := false
	for _, metric := range metrics {
		found = false
		for _, eMetric := range existMetrics {
			if metric.MType == eMetric.MType && metric.ID == eMetric.ID {
				e := eMetric
				value := new(float64)
				if metric.Value != nil {
					*value = *metric.Value
				}
				delta := new(int64)
				if metric.Delta != nil {
					*delta = *metric.Delta + *e.Delta
				}
				e.Value = value
				e.Delta = delta
				err := s.UpdateMetric(&e)
				if err != nil {
					logger.Get().Debug("update metric error", zap.String("error", err.Error()))
				}
				found = true
				break
			}
		}
		if !found {
			m := metric
			s.AddMetric(&m)
			existMetrics = append(existMetrics, m)
		}
	}

	return nil
}

func getRetryFunc(attempts, waitDelta int) func(err error) bool {
	secDelta := 0
	attempt := 0
	return func(err error) bool {
		attempt++

		// первый запуск
		if attempt == 1 && attempt <= attempts {
			return true
		}

		// если ошибки нет то не нужно нового трая
		if err == nil {
			return false
		}

		pgError := pgx.PgError{}
		if errors.Is(err, &pgError) && pgerrcode.IsConnectionException(pgError.Code) {
			time.Sleep(time.Duration(1+secDelta) * time.Second)
			secDelta += waitDelta
			return attempt <= attempts
		}

		// если дошли сюда, то попытки закончились
		return false
	}
}
