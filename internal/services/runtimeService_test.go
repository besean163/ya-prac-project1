package services

import (
	"testing"
	"ya-prac-project1/internal/metrics"

	"github.com/stretchr/testify/assert"
)

func TestGetRuntimeMetrics(t *testing.T) {
	ms := getRuntimeMetrics()

	assert.Equal(t, 30, len(ms))
}

func TestGetPoolCountMetric(t *testing.T) {
	m, _ := getPoolCountMetric()

	expect := metrics.NewMetric("PollCount", metrics.MetricTypeCounter, "1")
	assert.Equal(t, expect, m)
}

func TestGetRandomValueMetirc(t *testing.T) {
	m, _ := getRandomValueMetirc()

	assert.NotEqual(t, "", m.GetValue())
}
