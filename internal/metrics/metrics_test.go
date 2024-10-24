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
