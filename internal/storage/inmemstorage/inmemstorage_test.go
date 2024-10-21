package inmemstorage

import (
	"testing"
	"ya-prac-project1/internal/metrics"

	"github.com/stretchr/testify/assert"
)

func TestSetMetrics(t *testing.T) {
	s := NewStorage()

	assert.Equal(t, s.Metrics, []metrics.Metrics{})

	s.SetMetrics([]metrics.Metrics{
		metrics.NewMetric("id", "type", "20"),
	})
	assert.Equal(t, s.Metrics, []metrics.Metrics{
		metrics.NewMetric("id", "type", "20"),
	})

	s.SetMetrics([]metrics.Metrics{
		metrics.NewMetric("id1", "type1", "30"),
	})
	assert.Equal(t, s.Metrics, []metrics.Metrics{
		metrics.NewMetric("id1", "type1", "30"),
	})

}
