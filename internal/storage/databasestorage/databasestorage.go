// Модуль databasestorage предоставляет хранилище в sql базе
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

// Storage структура представляющая репозиторий
type Storage struct {
	DB *sql.DB
}

// NewStorage создает репозиторий
func NewStorage(db *sql.DB) (*Storage, error) {
	var err error
	if db == nil {
		err = errors.New("no database connection")
	}
	if err != nil {
		return nil, err
	}

	storage := Storage{
		DB: db,
	}

	err = storage.prepareDB()
	if err != nil {
		return nil, err
	}

	return &storage, nil
}

// GetMetrics возвращает все метрики в репозитории
func (s *Storage) GetMetrics() []metrics.Metrics {
	items := []metrics.Metrics{}

	retry := getRetryFunc(3, 2)

	var err error
	var rows *sql.Rows
	for retry(err) {
		rows, err = s.DB.Query("SELECT type,name,value,delta FROM metrics")
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

// CreateMetrics добавляет полученные метрики в репозиторий
func (s *Storage) CreateMetrics(ms []metrics.Metrics) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}

	for _, metric := range ms {
		m := metric
		_, err := tx.Exec(getInsertMetricSQL(), m.MType, m.ID, m.Value, m.Delta)
		if err != nil {
			logger.Get().Info("tx insert metric error", zap.String("error", err.Error()))
			tx.Rollback()
			return err
		}

	}

	return tx.Commit()
}

// UpdateMetrics обновляет полученные метрики в репозитории
func (s *Storage) UpdateMetrics(ms []metrics.Metrics) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}

	for _, metric := range ms {
		m := metric
		_, err := tx.Exec(getUpdateMetricSQL(), m.Value, m.Delta, m.MType, m.ID)
		if err != nil {
			logger.Get().Info("tx update metric error", zap.String("error", err.Error()))
			tx.Rollback()
			return err
		}

	}

	return tx.Commit()
}

func (s *Storage) prepareDB() error {
	sql := `CREATE TABLE IF NOT EXISTS metrics(
    	name    varchar(255) PRIMARY KEY,
    	type    varchar(40),
    	value    double precision default null,
    	delta    bigint default null
	);`

	_, err := s.DB.Exec(sql)
	if err != nil {
		return err
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

func getUpdateMetricSQL() string {
	return "UPDATE metrics SET value = $1, delta = $2 WHERE type = $3 AND name = $4"
}

func getInsertMetricSQL() string {
	return "INSERT INTO metrics (type, name, value, delta) VALUES ($1,$2,$3,$4)"
}
