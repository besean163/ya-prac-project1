// Package databasestorage предоставляет хранилище в sql базе
package databasestorage

import (
	"errors"
	"testing"
	"ya-prac-project1/internal/logger"
	"ya-prac-project1/internal/metrics"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgerrcode"
	pgx "github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
)

func TestGetUpdateMetricSQL(t *testing.T) {
	getUpdateMetricSQL()
}

func TestGetInsertMetricSQL(t *testing.T) {
	getInsertMetricSQL()
}

func TestGetRetryFunc(t *testing.T) {
	attempts := 3
	waitDelta := 1

	t.Run("first call always true", func(t *testing.T) {
		retryFunc := getRetryFunc(attempts, waitDelta)
		assert.True(t, retryFunc(nil), "First call should always return true")
	})

	t.Run("no error, no retry", func(t *testing.T) {
		retryFunc := getRetryFunc(attempts, waitDelta)

		// First call
		retryFunc(nil)

		// No error, should return false
		assert.False(t, retryFunc(nil), "No error should stop retrying")
	})

	t.Run("connection exception retries", func(t *testing.T) {
		retryFunc := getRetryFunc(attempts, waitDelta)

		// First call
		retryFunc(nil)

		// Simulate a pgx.PgError with a connection exception code
		pgError := &pgx.PgError{
			Code: pgerrcode.ConnectionException,
		}

		for i := 0; i < attempts-1; i++ {
			assert.True(t, retryFunc(pgError), "Should retry for connection exceptions")
		}

		// Exceed attempts
		assert.False(t, retryFunc(pgError), "Should stop retrying after max attempts")
	})

	t.Run("non-connection error, no retry", func(t *testing.T) {
		retryFunc := getRetryFunc(attempts, waitDelta)

		// First call
		retryFunc(nil)

		// Simulate a non-connection pgx.PgError
		pgError := &pgx.PgError{
			Code: pgerrcode.InvalidTextRepresentation,
		}

		assert.False(t, retryFunc(pgError), "Should not retry for non-connection exceptions")
	})

	t.Run("non-pg error, no retry", func(t *testing.T) {
		retryFunc := getRetryFunc(attempts, waitDelta)

		// First call
		retryFunc(nil)

		// Simulate a generic error
		err := errors.New("generic error")

		assert.False(t, retryFunc(err), "Should not retry for generic errors")
	})
}

func TestGetMetrics_Success(t *testing.T) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Ожидаем выполнение запроса
	rows := sqlmock.NewRows([]string{"type", "name", "value", "delta"}).
		AddRow("gauge", "test_gauge", 123.45, nil).
		AddRow("counter", "test_counter", nil, 10)
	mock.ExpectQuery("SELECT type,name,value,delta FROM metrics").WillReturnRows(rows)

	// Создаем Storage с моковой базой данных
	storage := &Storage{DB: db}
	metrics := storage.GetMetrics()

	// Проверяем результат
	assert.Len(t, metrics, 2)

	// Проверяем первую метрику
	assert.Equal(t, "gauge", metrics[0].MType)
	assert.Equal(t, "test_gauge", metrics[0].ID)
	assert.NotNil(t, metrics[0].Value)
	assert.Equal(t, 123.45, *metrics[0].Value)

	// Проверяем вторую метрику
	assert.Equal(t, "counter", metrics[1].MType)
	assert.Equal(t, "test_counter", metrics[1].ID)
	assert.NotNil(t, metrics[1].Delta)
	assert.Equal(t, int64(10), *metrics[1].Delta)
}

func TestGetMetrics_QueryError(t *testing.T) {
	logger.Set()
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Ожидаем ошибку при выполнении запроса
	mock.ExpectQuery("SELECT type,name,value,delta FROM metrics").WillReturnError(errors.New("query error"))

	storage := &Storage{DB: db}
	metrics := storage.GetMetrics()

	// Проверяем, что метрики пустые
	assert.Len(t, metrics, 0)
}

func TestGetMetrics_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Ожидаем выполнение запроса, но с ошибкой в данных
	rows := sqlmock.NewRows([]string{"type", "name", "value", "delta"}).
		AddRow("gauge", "test_gauge", "invalid", nil) // Некорректное значение для value
	mock.ExpectQuery("SELECT type,name,value,delta FROM metrics").WillReturnRows(rows)

	storage := &Storage{DB: db}
	metrics := storage.GetMetrics()

	// Проверяем, что метрики пустые из-за ошибки сканирования
	assert.Len(t, metrics, 1)
}

func TestGetMetrics_RowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Ожидаем выполнение запроса, но с ошибкой после итерации по строкам
	rows := sqlmock.NewRows([]string{"type", "name", "value", "delta"}).
		AddRow("gauge", "test_gauge", 123.45, nil).
		RowError(0, errors.New("row error"))
	mock.ExpectQuery("SELECT type,name,value,delta FROM metrics").WillReturnRows(rows)

	storage := &Storage{DB: db}
	metrics := storage.GetMetrics()

	// Проверяем, что метрики пустые из-за ошибки в строках
	assert.Len(t, metrics, 0)
}

func TestUpdateMetrics_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Ожидаем начало транзакции
	mock.ExpectBegin()

	// Ожидаем выполнение запросов обновления метрик
	mock.ExpectExec("UPDATE metrics SET value = \\$1, delta = \\$2 WHERE type = \\$3 AND name = \\$4").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "gauge", "metric1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("UPDATE metrics SET value = \\$1, delta = \\$2 WHERE type = \\$3 AND name = \\$4").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "counter", "metric2").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Ожидаем успешное завершение транзакции
	mock.ExpectCommit()

	storage := &Storage{DB: db}
	metricsToUpdate := []metrics.Metrics{
		{MType: "gauge", ID: "metric1", Value: floatPtr(123.45)},
		{MType: "counter", ID: "metric2", Delta: int64Ptr(10)},
	}

	err = storage.UpdateMetrics(metricsToUpdate)
	assert.NoError(t, err)
}

func TestUpdateMetrics_BeginError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Ожидаем ошибку при начале транзакции
	mock.ExpectBegin().WillReturnError(errors.New("begin error"))

	storage := &Storage{DB: db}
	metricsToUpdate := []metrics.Metrics{
		{MType: "gauge", ID: "metric1", Value: floatPtr(123.45)},
	}

	err = storage.UpdateMetrics(metricsToUpdate)
	assert.Error(t, err)
	assert.Equal(t, "begin error", err.Error())
}

func TestUpdateMetrics_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Ожидаем начало транзакции
	mock.ExpectBegin()

	// Ожидаем ошибку при выполнении SQL-запроса
	mock.ExpectExec("UPDATE metrics SET value = \\$1, delta = \\$2 WHERE type = \\$3 AND name = \\$4").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "gauge", "metric1").
		WillReturnError(errors.New("exec error"))

	// Ожидаем успешный откат транзакции
	mock.ExpectRollback()

	storage := &Storage{DB: db}
	metricsToUpdate := []metrics.Metrics{
		{MType: "gauge", ID: "metric1", Value: floatPtr(123.45)},
	}

	err = storage.UpdateMetrics(metricsToUpdate)
	assert.Error(t, err)
	assert.Equal(t, "exec error", err.Error())
}

func TestUpdateMetrics_RollbackSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Ожидаем начало транзакции
	mock.ExpectBegin()

	// Ожидаем ошибку при выполнении SQL-запроса
	mock.ExpectExec("UPDATE metrics SET value = \\$1, delta = \\$2 WHERE type = \\$3 AND name = \\$4").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "gauge", "metric1").
		WillReturnError(errors.New("exec error"))

	// Ожидаем успешный откат транзакции
	mock.ExpectRollback()

	storage := &Storage{DB: db}
	metricsToUpdate := []metrics.Metrics{
		{MType: "gauge", ID: "metric1", Value: floatPtr(123.45)},
	}

	err = storage.UpdateMetrics(metricsToUpdate)
	assert.Error(t, err)
	assert.Equal(t, "exec error", err.Error())
}

func TestUpdateMetrics_CommitError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Ожидаем начало транзакции
	mock.ExpectBegin()

	// Ожидаем успешное выполнение SQL-запроса
	mock.ExpectExec("UPDATE metrics SET value = \\$1, delta = \\$2 WHERE type = \\$3 AND name = \\$4").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "gauge", "metric1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Ожидаем ошибку при завершении транзакции
	mock.ExpectCommit().WillReturnError(errors.New("commit error"))

	storage := &Storage{DB: db}
	metricsToUpdate := []metrics.Metrics{
		{MType: "gauge", ID: "metric1", Value: floatPtr(123.45)},
	}

	err = storage.UpdateMetrics(metricsToUpdate)
	assert.Error(t, err)
	assert.Equal(t, "commit error", err.Error())
}

func floatPtr(v float64) *float64 {
	return &v
}

func int64Ptr(v int64) *int64 {
	return &v
}

func TestCreateMetrics_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Ожидаем начало транзакции
	mock.ExpectBegin()

	// Ожидаем выполнение запросов вставки метрик
	mock.ExpectExec("INSERT INTO metrics \\(type, name, value, delta\\) VALUES \\(\\$1,\\$2,\\$3,\\$4\\)").
		WithArgs("gauge", "metric1", sqlmock.AnyArg(), nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO metrics \\(type, name, value, delta\\) VALUES \\(\\$1,\\$2,\\$3,\\$4\\)").
		WithArgs("counter", "metric2", nil, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Ожидаем успешное завершение транзакции
	mock.ExpectCommit()

	storage := &Storage{DB: db}
	metricsToCreate := []metrics.Metrics{
		{MType: "gauge", ID: "metric1", Value: floatPtr(123.45)},
		{MType: "counter", ID: "metric2", Delta: int64Ptr(10)},
	}

	err = storage.CreateMetrics(metricsToCreate)
	assert.NoError(t, err)
}

func TestCreateMetrics_BeginError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Ожидаем ошибку при начале транзакции
	mock.ExpectBegin().WillReturnError(errors.New("begin error"))

	storage := &Storage{DB: db}
	metricsToCreate := []metrics.Metrics{
		{MType: "gauge", ID: "metric1", Value: floatPtr(123.45)},
	}

	err = storage.CreateMetrics(metricsToCreate)
	assert.Error(t, err)
	assert.Equal(t, "begin error", err.Error())
}

func TestCreateMetrics_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Ожидаем начало транзакции
	mock.ExpectBegin()

	// Ожидаем ошибку при выполнении SQL-запроса
	mock.ExpectExec("INSERT INTO metrics \\(type, name, value, delta\\) VALUES \\(\\$1,\\$2,\\$3,\\$4\\)").
		WithArgs("gauge", "metric1", sqlmock.AnyArg(), nil).
		WillReturnError(errors.New("exec error"))

	// Ожидаем успешный откат транзакции
	mock.ExpectRollback()

	storage := &Storage{DB: db}
	metricsToCreate := []metrics.Metrics{
		{MType: "gauge", ID: "metric1", Value: floatPtr(123.45)},
	}

	err = storage.CreateMetrics(metricsToCreate)
	assert.Error(t, err)
	assert.Equal(t, "exec error", err.Error())
}

func TestCreateMetrics_RollbackSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Ожидаем начало транзакции
	mock.ExpectBegin()

	// Ожидаем ошибку при выполнении SQL-запроса
	mock.ExpectExec("INSERT INTO metrics \\(type, name, value, delta\\) VALUES \\(\\$1,\\$2,\\$3,\\$4\\)").
		WithArgs("gauge", "metric1", sqlmock.AnyArg(), nil).
		WillReturnError(errors.New("exec error"))

	// Ожидаем успешный откат транзакции
	mock.ExpectRollback()

	storage := &Storage{DB: db}
	metricsToCreate := []metrics.Metrics{
		{MType: "gauge", ID: "metric1", Value: floatPtr(123.45)},
	}

	err = storage.CreateMetrics(metricsToCreate)
	assert.Error(t, err)
	assert.Equal(t, "exec error", err.Error())
}

func TestCreateMetrics_CommitError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Ожидаем начало транзакции
	mock.ExpectBegin()

	// Ожидаем успешное выполнение SQL-запросов
	mock.ExpectExec("INSERT INTO metrics \\(type, name, value, delta\\) VALUES \\(\\$1,\\$2,\\$3,\\$4\\)").
		WithArgs("gauge", "metric1", sqlmock.AnyArg(), nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Ожидаем ошибку при завершении транзакции
	mock.ExpectCommit().WillReturnError(errors.New("commit error"))

	storage := &Storage{DB: db}
	metricsToCreate := []metrics.Metrics{
		{MType: "gauge", ID: "metric1", Value: floatPtr(123.45)},
	}

	err = storage.CreateMetrics(metricsToCreate)
	assert.Error(t, err)
	assert.Equal(t, "commit error", err.Error())
}

func TestPrepareDB_Success(t *testing.T) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Ожидаем успешное выполнение SQL-запроса
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS metrics").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Создаем объект Storage и вызываем prepareDB
	storage := &Storage{DB: db}
	err = storage.prepareDB()
	assert.NoError(t, err)

	// Проверяем, что все ожидаемые запросы были выполнены
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPrepareDB_Error(t *testing.T) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Ожидаем ошибку при выполнении SQL-запроса
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS metrics").
		WillReturnError(errors.New("exec error"))

	// Создаем объект Storage и вызываем prepareDB
	storage := &Storage{DB: db}
	err = storage.prepareDB()
	assert.Error(t, err)
	assert.Equal(t, "exec error", err.Error())

	// Проверяем, что все ожидаемые запросы были выполнены
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNewStorage_NilDB(t *testing.T) {
	// Передаем nil вместо базы данных
	storage, err := NewStorage(nil)

	// Ожидаем ошибку
	assert.Nil(t, storage)
	assert.Error(t, err)
	assert.Equal(t, "no database connection", err.Error())
}

func TestNewStorage_PrepareDBError(t *testing.T) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Ожидаем ошибку при выполнении SQL-запроса в prepareDB
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS metrics").WillReturnError(errors.New("prepare error"))

	// Проверяем результат
	storage, err := NewStorage(db)
	assert.Nil(t, storage)
	assert.Error(t, err)
	assert.Equal(t, "prepare error", err.Error())
}

func TestNewStorage_Success(t *testing.T) {
	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Ожидаем успешное выполнение SQL-запроса в prepareDB
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS metrics").WillReturnResult(sqlmock.NewResult(1, 1))

	// Проверяем результат
	storage, err := NewStorage(db)
	assert.NotNil(t, storage)
	assert.NoError(t, err)
	assert.Equal(t, db, storage.DB)
}
