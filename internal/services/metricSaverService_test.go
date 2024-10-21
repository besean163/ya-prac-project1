package services

import (
	"errors"
	"testing"
	"ya-prac-project1/internal/metrics"
	mock "ya-prac-project1/internal/services/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mock.NewMockSaveStorage(ctrl)

	ms := []metrics.Metrics{
		metrics.NewMetric("test_1", metrics.MetricTypeGauge, "1.5"),
		metrics.NewMetric("test_2", metrics.MetricTypeGauge, "2.5"),
	}

	store.EXPECT().GetMetrics().Return(ms).AnyTimes()
	s := NewMetricSaverService(store)

	assert.Equal(t, ms, s.GetMetrics())
}

func TestGetMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mock.NewMockSaveStorage(ctrl)

	ms := []metrics.Metrics{
		metrics.NewMetric("test_1", metrics.MetricTypeGauge, "1.5"),
		metrics.NewMetric("test_2", metrics.MetricTypeGauge, "2.5"),
	}

	store.EXPECT().GetMetrics().Return(ms).AnyTimes()
	s := NewMetricSaverService(store)

	actual, _ := s.GetMetric(metrics.MetricTypeGauge, "test_1")
	expect := metrics.NewMetric("test_1", metrics.MetricTypeGauge, "1.5")
	assert.Equal(t, expect, actual)

	_, err := s.GetMetric(metrics.MetricTypeGauge, "test_3")
	eErr := errors.New("metric not found")
	assert.Equal(t, eErr, err)
}

func TestSaveMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mock.NewMockSaveStorage(ctrl)

	ms := []metrics.Metrics{
		metrics.NewMetric("test_1", metrics.MetricTypeGauge, "1.5"),
		metrics.NewMetric("test_2", metrics.MetricTypeGauge, "2.5"),
	}

	cms := []metrics.Metrics{
		metrics.NewMetric("test_3", metrics.MetricTypeGauge, "3.5"),
	}

	ums := []metrics.Metrics{
		metrics.NewMetric("test_1", metrics.MetricTypeGauge, "10.5"),
		metrics.NewMetric("test_2", metrics.MetricTypeGauge, "2.5"),
	}

	store.EXPECT().GetMetrics().Return(ms).AnyTimes()
	store.EXPECT().CreateMetrics(cms).Return(nil).AnyTimes()
	store.EXPECT().UpdateMetrics(ums).Return(nil).AnyTimes()
	store.EXPECT().CreateMetrics(gomock.Any()).Return(errors.New("wrong create collection")).AnyTimes()
	store.EXPECT().UpdateMetrics(gomock.Any()).Return(errors.New("wrong update collection")).AnyTimes()
	s := NewMetricSaverService(store)

	err := s.SaveMetrics([]metrics.Metrics{
		metrics.NewMetric("test_1", metrics.MetricTypeGauge, "10.5"),
		metrics.NewMetric("test_2", metrics.MetricTypeGauge, "2.5"),
		metrics.NewMetric("test_3", metrics.MetricTypeGauge, "3.5"),
	})

	assert.Equal(t, err, nil)
}

func TestSaveMetrics_wrong(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mock.NewMockSaveStorage(ctrl)

	ms := []metrics.Metrics{
		metrics.NewMetric("test_1", metrics.MetricTypeGauge, "1.5"),
		metrics.NewMetric("test_2", metrics.MetricTypeGauge, "2.5"),
	}

	store.EXPECT().GetMetrics().Return(ms).AnyTimes()
	s := NewMetricSaverService(store)

	err := s.SaveMetrics([]metrics.Metrics{
		metrics.NewMetric("test_1", "wrong_type", "10.5"),
	})

	assert.Equal(t, err, errors.New("wrong metric type"))
}

func TestSaveMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mock.NewMockSaveStorage(ctrl)

	ms := []metrics.Metrics{
		metrics.NewMetric("test_1", metrics.MetricTypeGauge, "1.5"),
		metrics.NewMetric("test_2", metrics.MetricTypeGauge, "2.5"),
	}

	cms := []metrics.Metrics{
		metrics.NewMetric("test_3", metrics.MetricTypeGauge, "3.5"),
	}

	ums := []metrics.Metrics{
		metrics.NewMetric("test_1", metrics.MetricTypeGauge, "10.5"),
	}

	store.EXPECT().GetMetrics().Return(ms).AnyTimes()
	store.EXPECT().CreateMetrics(cms).Return(nil).AnyTimes()
	store.EXPECT().UpdateMetrics(ums).Return(nil).AnyTimes()
	store.EXPECT().CreateMetrics(gomock.Any()).Return(errors.New("wrong create collection")).AnyTimes()
	store.EXPECT().UpdateMetrics(gomock.Any()).Return(errors.New("wrong update collection")).AnyTimes()
	s := NewMetricSaverService(store)

	err := s.SaveMetric(metrics.NewMetric("test_1", metrics.MetricTypeGauge, "10.5"))
	assert.ErrorIs(t, err, nil)

	err = s.SaveMetric(metrics.NewMetric("test_3", metrics.MetricTypeGauge, "3.5"))
	assert.ErrorIs(t, err, nil)
}

func TestSaveMetric_wrong(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mock.NewMockSaveStorage(ctrl)

	ms := []metrics.Metrics{
		metrics.NewMetric("test_1", metrics.MetricTypeGauge, "1.5"),
		metrics.NewMetric("test_2", metrics.MetricTypeGauge, "2.5"),
	}

	store.EXPECT().GetMetrics().Return(ms).AnyTimes()
	s := NewMetricSaverService(store)

	err := s.SaveMetric(metrics.NewMetric("test_1", "wrong_type", "10.5"))
	assert.Equal(t, err, errors.New("wrong metric type"))
}

func TestGetMetricsKeyMap(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mock.NewMockSaveStorage(ctrl)

	ms := []metrics.Metrics{
		metrics.NewMetric("test_1", metrics.MetricTypeGauge, "1.5"),
		metrics.NewMetric("test_2", metrics.MetricTypeGauge, "2.5"),
	}

	store.EXPECT().GetMetrics().Return(ms).AnyTimes()
	s := NewMetricSaverService(store)

	actual := s.getMetricsKeyMap()
	expect := map[string]metrics.Metrics{
		"gauge_test_1": metrics.NewMetric("test_1", metrics.MetricTypeGauge, "1.5"),
		"gauge_test_2": metrics.NewMetric("test_2", metrics.MetricTypeGauge, "2.5"),
	}
	assert.Equal(t, expect, actual)
}

func BenchmarkGetMetric(b *testing.B) {
	ctrl := gomock.NewController(b)
	store := mock.NewMockSaveStorage(ctrl)

	ms := []metrics.Metrics{
		metrics.NewMetric("test_1", metrics.MetricTypeGauge, "1.5"),
		metrics.NewMetric("test_2", metrics.MetricTypeGauge, "2.5"),
		metrics.NewMetric("test_3", metrics.MetricTypeGauge, "3.5"),
		metrics.NewMetric("test_4", metrics.MetricTypeGauge, "3.5"),
		metrics.NewMetric("test_5", metrics.MetricTypeGauge, "3.5"),
		metrics.NewMetric("test_6", metrics.MetricTypeGauge, "3.5"),
		metrics.NewMetric("test_7", metrics.MetricTypeGauge, "3.5"),
		metrics.NewMetric("test_8", metrics.MetricTypeGauge, "3.5"),
		metrics.NewMetric("test_9", metrics.MetricTypeGauge, "3.5"),
		metrics.NewMetric("test_10", metrics.MetricTypeGauge, "3.5"),
	}

	store.EXPECT().GetMetrics().Return(ms).AnyTimes()
	s := NewMetricSaverService(store)

	for i := 0; i < b.N; i++ {
		s.GetMetric(metrics.MetricTypeGauge, "test_10")
	}
}
