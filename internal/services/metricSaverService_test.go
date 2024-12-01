package services

import (
	"context"
	"errors"
	"testing"
	"ya-prac-project1/internal/logger"
	"ya-prac-project1/internal/metrics"
	mock "ya-prac-project1/internal/services/mocks"

	pb "ya-prac-project1/internal/services/proto"

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

func TestUpdateMetrics(t *testing.T) {
	logger.Set()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем мок для SaveStorage
	mockStorage := mock.NewMockSaveStorage(ctrl)

	// Инициализируем сервис через конструктор
	service := NewMetricSaverService(mockStorage)

	// Пример входных данных для тестов
	in := &pb.SaveMetricsRequest{
		Metrics: []*pb.Metric{
			{
				Id:    "metric1",
				Type:  "counter",
				Value: new(float64), // Указатель на float64
			},
			{
				Id:    "metric2",
				Type:  "gauge",
				Delta: new(int64), // Указатель на int64
			},
		},
	}

	t.Run("successful update", func(t *testing.T) {
		// Ожидания для вызовов зависимостей
		mockStorage.EXPECT().GetMetrics().Return([]metrics.Metrics{}).Times(1)
		mockStorage.EXPECT().CreateMetrics(gomock.Any()).Return(nil).Times(1)

		// Выполняем тест
		resp, err := service.UpdateMetrics(context.Background(), in)

		// Проверяем результат
		assert.NoError(t, err)
		assert.Empty(t, resp.Error)
	})

	t.Run("create metrics error", func(t *testing.T) {
		// Ожидаем ошибку при создании метрик
		mockStorage.EXPECT().GetMetrics().Return([]metrics.Metrics{}).Times(1)
		mockStorage.EXPECT().CreateMetrics(gomock.Any()).Return(assert.AnError).Times(1)

		// Выполняем тест
		resp, err := service.UpdateMetrics(context.Background(), in)

		// Проверяем, что ошибка вернулась
		assert.Error(t, err)
		assert.Equal(t, assert.AnError.Error(), resp.Error)
	})

	t.Run("update metrics error", func(t *testing.T) {
		// Подготовка существующих метрик
		existingMetrics := []metrics.Metrics{
			{ID: "metric1", MType: "counter"},
		}
		mockStorage.EXPECT().GetMetrics().Return(existingMetrics).Times(1)

		// Выполняем тест
		_, err := service.UpdateMetrics(context.Background(), in)

		// Проверяем, что ошибка вернулась
		assert.Error(t, err)
	})

	t.Run("validation error", func(t *testing.T) {
		// Создаем некорректный запрос
		invalidRequest := &pb.SaveMetricsRequest{
			Metrics: []*pb.Metric{
				{
					Id:   "invalid_metric",
					Type: "unknown_type", // Некорректный тип метрики
				},
			},
		}

		// Не ожидается вызов CreateMetrics или UpdateMetrics
		mockStorage.EXPECT().GetMetrics().Return([]metrics.Metrics{}).Times(1)

		// Выполняем тест
		_, err := service.UpdateMetrics(context.Background(), invalidRequest)

		// Проверяем результат
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "wrong metric type") // Убедитесь, что валидация отрабатывает
	})
}

func TestPBMetricReset(t *testing.T) {
	metric := pb.Metric{}
	metric.Reset()
}

func TestPBMetricString(t *testing.T) {
	metric := pb.Metric{}
	_ = metric.String()
}

func TestPBMetricProtoMessage(t *testing.T) {
	metric := pb.Metric{}
	metric.ProtoMessage()
}

func TestPBMetricGetDelta(t *testing.T) {
	metric := pb.Metric{}
	_ = metric.GetDelta()
}
func TestPBMetricGetValue(t *testing.T) {
	metric := pb.Metric{}
	_ = metric.GetValue()
}

func TestPBMetricGetId(t *testing.T) {
	var metric pb.Metric
	_ = metric.GetId()
	metric = pb.Metric{}
	_ = metric.GetId()
}

func TestPBMetricGetType(t *testing.T) {
	metric := pb.Metric{}
	_ = metric.GetType()
}

func TestPBSaveMetricsRequestReset(t *testing.T) {
	r := pb.SaveMetricsRequest{}
	r.Reset()
}

func TestPBSaveMetricsRequestString(t *testing.T) {
	r := pb.SaveMetricsRequest{}
	_ = r.String()
}

// func TestPBSaveMetricsRequestProtoMessage(t *testing.T) {
// 	r := pb.SaveMetricsRequest{}
// 	r.ProtoMessage()
// }

func TestPBSaveMetricsRequestGetMetrics(t *testing.T) {
	r := pb.SaveMetricsRequest{}
	_ = r.GetMetrics()
}

func SaveMetricsResponseReset(t *testing.T) {
	r := &pb.SaveMetricsResponse{}
	r.Reset()
}

func SaveMetricsResponseString(t *testing.T) {
	r := pb.SaveMetricsResponse{}
	_ = r.String()
}

func SaveMetricsResponseProtoMessage(t *testing.T) {
	r := pb.SaveMetricsResponse{}
	r.ProtoMessage()
}
