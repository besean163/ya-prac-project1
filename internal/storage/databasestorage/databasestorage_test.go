// Package databasestorage предоставляет хранилище в sql базе
package databasestorage

import (
	"testing"

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
	f := getRetryFunc(2, 0)
	assert.True(t, f(nil))
	assert.False(t, f(nil))
}
