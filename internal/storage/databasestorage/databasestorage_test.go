// Package databasestorage предоставляет хранилище в sql базе
package databasestorage

import (
	"errors"
	"testing"

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
