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

func TestGetMetrics(t *testing.T) {
	ms := []metrics.Metrics{
		metrics.NewMetric("test_1", metrics.MetricTypeGauge, "1.5"),
		metrics.NewMetric("test_2", metrics.MetricTypeGauge, "2.5"),
	}
	s := NewStorage()
	s.SetMetrics(ms)

	assert.Equal(t, ms, s.GetMetrics())
}

func TestUpdateMetrics(t *testing.T) {
	ms := []metrics.Metrics{
		metrics.NewMetric("test_1", metrics.MetricTypeGauge, "1.5"),
		metrics.NewMetric("test_2", metrics.MetricTypeGauge, "2.5"),
	}

	ums := []metrics.Metrics{
		metrics.NewMetric("test_1", metrics.MetricTypeGauge, "10.5"),
	}

	s := NewStorage()
	s.SetMetrics(ms)

	s.UpdateMetrics(ums)

	expect := []metrics.Metrics{
		metrics.NewMetric("test_1", metrics.MetricTypeGauge, "10.5"),
		metrics.NewMetric("test_2", metrics.MetricTypeGauge, "2.5"),
	}
	assert.Equal(t, expect, s.GetMetrics())
}

func TestCreateMetrics(t *testing.T) {
	ms := []metrics.Metrics{
		metrics.NewMetric("test_1", metrics.MetricTypeGauge, "1.5"),
		metrics.NewMetric("test_2", metrics.MetricTypeGauge, "2.5"),
	}

	cms := []metrics.Metrics{
		metrics.NewMetric("test_3", metrics.MetricTypeGauge, "3.5"),
	}

	s := NewStorage()
	s.SetMetrics(ms)

	s.CreateMetrics(cms)

	expect := []metrics.Metrics{
		metrics.NewMetric("test_1", metrics.MetricTypeGauge, "1.5"),
		metrics.NewMetric("test_2", metrics.MetricTypeGauge, "2.5"),
		metrics.NewMetric("test_3", metrics.MetricTypeGauge, "3.5"),
	}
	assert.Equal(t, expect, s.GetMetrics())
}
