package metrics

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetValue(t *testing.T) {
	m := NewMetric("test", MetricTypeGauge, "20")
	assert.Equal(t, "20", m.GetValue())
}

func TestSetValue(t *testing.T) {
	m := NewMetric("test", MetricTypeGauge, "20")
	assert.Equal(t, "20", m.GetValue())
	m.SetValue("30")
	assert.Equal(t, "30", m.GetValue())

	m = NewMetric("test", MetricTypeCounter, "20")
	assert.Equal(t, "20", m.GetValue())
	m.SetValue("30")
	assert.Equal(t, "50", m.GetValue())

	m = NewMetric("test", "wrong", "20")
	assert.Equal(t, "", m.GetValue())
	m.SetValue("")
	assert.Equal(t, "", m.GetValue())
}

func TestGetInfo(t *testing.T) {
	m := NewMetric("test", MetricTypeGauge, "20")
	expect := fmt.Sprintf("%s - %s - %s", m.ID, m.MType, m.GetValue())
	assert.Equal(t, expect, m.GetInfo())
}

func TestGetKey(t *testing.T) {
	m := NewMetric("test", MetricTypeGauge, "20")
	expect := fmt.Sprintf("%s_%s", m.MType, m.ID)
	assert.Equal(t, expect, m.GetKey())
}

func TestValidate(t *testing.T) {
	m := NewMetric("test", MetricTypeGauge, "20")
	err := m.Validate()
	assert.ErrorIs(t, err, nil)

	m = NewMetric("test", "wrong", "20")
	err = m.Validate()
	assert.Error(t, err, errors.New("wrong metric type"))
}

func TestSetValue_Gauge(t *testing.T) {
	t.Run("Set valid gauge value", func(t *testing.T) {
		metric := Metrics{
			ID:    "test_gauge",
			MType: MetricTypeGauge,
		}

		err := metric.SetValue("123.45")
		assert.NoError(t, err)
		assert.NotNil(t, metric.Value)
		assert.Equal(t, 123.45, *metric.Value)
	})

	t.Run("Set invalid gauge value", func(t *testing.T) {
		metric := Metrics{
			ID:    "test_gauge",
			MType: MetricTypeGauge,
		}

		err := metric.SetValue("invalid")
		assert.Error(t, err)
		assert.Nil(t, metric.Value)
	})
}

func TestSetValue_Counter(t *testing.T) {
	t.Run("Set valid counter value", func(t *testing.T) {
		metric := Metrics{
			ID:    "test_counter",
			MType: MetricTypeCounter,
		}

		err := metric.SetValue("10")
		assert.NoError(t, err)
		assert.NotNil(t, metric.Delta)
		assert.Equal(t, int64(10), *metric.Delta)
	})

	t.Run("Update existing counter value", func(t *testing.T) {
		existingDelta := int64(15)
		metric := Metrics{
			ID:    "test_counter",
			MType: MetricTypeCounter,
			Delta: &existingDelta,
		}

		err := metric.SetValue("5")
		assert.NoError(t, err)
		assert.NotNil(t, metric.Delta)
		assert.Equal(t, int64(20), *metric.Delta)
	})

	t.Run("Set invalid counter value", func(t *testing.T) {
		metric := Metrics{
			ID:    "test_counter",
			MType: MetricTypeCounter,
		}

		err := metric.SetValue("invalid")
		assert.Error(t, err)
		assert.Nil(t, metric.Delta)
	})
}

func TestSetValue_UnknownType(t *testing.T) {
	t.Run("Unknown metric type", func(t *testing.T) {
		metric := Metrics{
			ID:    "test_unknown",
			MType: "unknown_type",
		}

		err := metric.SetValue("123")
		assert.NoError(t, err)
		assert.Nil(t, metric.Value)
		assert.Nil(t, metric.Delta)
	})
}
