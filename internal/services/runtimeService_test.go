package services

import (
	"bytes"
	"context"
	"net/http"
	"testing"
	"time"
	"ya-prac-project1/internal/logger"
	"ya-prac-project1/internal/metrics"
	mock "ya-prac-project1/internal/services/mocks"

	"github.com/golang/mock/gomock"
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

func TestUpdateRuntimeMetrics(t *testing.T) {
	logger.Set()
	ctrl := gomock.NewController(t)
	store := mock.NewMockStorage(ctrl)

	store.EXPECT().SetMetrics(gomock.Any()).AnyTimes()

	ctx, stop := context.WithTimeout(context.Background(), 1*time.Second)
	s := NewRuntimeService(store)
	s.Run(ctx, 10)
	<-ctx.Done()
	stop()
}

func TestUpdateRuntimeMetrics2(t *testing.T) {
	logger.Set()
	ctrl := gomock.NewController(t)
	store := mock.NewMockStorage(ctrl)

	store.EXPECT().GetMetrics().Return([]metrics.Metrics{}).AnyTimes()
	store.EXPECT().SetMetrics(gomock.Any()).AnyTimes()
	ctx, stop := context.WithTimeout(context.Background(), 2*time.Second)

	s := NewRuntimeService(store)
	s.updateRuntimeMetrics(ctx, 1)
	<-ctx.Done()
	stop()
}

func TestRunSendRequest(t *testing.T) {
	logger.Set()
	ctrl := gomock.NewController(t)
	store := mock.NewMockStorage(ctrl)

	store.EXPECT().GetMetrics().Return([]metrics.Metrics{}).AnyTimes()

	s := NewRuntimeService(store)
	rCh := make(chan *http.Request, 1)
	s.RunSendRequest(rCh, "", "", "")
	<-rCh
}

func TestRunSendRequest2(t *testing.T) {
	logger.Set()
	ctrl := gomock.NewController(t)
	store := mock.NewMockStorage(ctrl)

	store.EXPECT().GetMetrics().Return([]metrics.Metrics{}).AnyTimes()

	s := NewRuntimeService(store)
	rCh := make(chan *http.Request, 1)
	s.RunSendRequest(rCh, "", "test", "")
	<-rCh
}

func TestRunSendgRPCRequest_1(t *testing.T) {
	logger.Set()
	ctrl := gomock.NewController(t)
	store := mock.NewMockStorage(ctrl)

	store.EXPECT().GetMetrics().Return([]metrics.Metrics{}).AnyTimes()

	s := NewRuntimeService(store)
	s.RunSendgRPCRequest("")
}

func TestEncryptMessage(t *testing.T) {
	encryptMessage(*bytes.NewBuffer([]byte{}), "")
}
func TestEncryptMessage2(t *testing.T) {
	encryptMessage(*bytes.NewBuffer([]byte{}), "testdata/public.pem")
}

func TestEncryptMessage3(t *testing.T) {
	encryptMessage(*bytes.NewBuffer([]byte{}), "testdata/wrong_public.pem")
}
